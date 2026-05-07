package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 缓存抽象接口
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	FlushDB(ctx context.Context) error
	DBSize(ctx context.Context) (int64, error)
	Info(ctx context.Context) (string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// ErrNil 表示 key 不存在
var ErrNil = redis.Nil

// ========== Redis 实现 ==========

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// SetNX 仅在 key 不存在时设置值并指定 TTL，用于短时分布式互斥（如多实例定时任务）。
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

func (r *RedisCache) DBSize(ctx context.Context) (int64, error) {
	return r.client.DBSize(ctx).Result()
}

func (r *RedisCache) Info(ctx context.Context) (string, error) {
	return r.client.Info(ctx, "memory", "keyspace").Result()
}

func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}

// ========== 内存实现 ==========

type memItem struct {
	value  string
	expiry time.Time // zero means no expiry
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]memItem
	stop  chan struct{}
}

func NewMemoryCache() *MemoryCache {
	m := &MemoryCache{items: make(map[string]memItem), stop: make(chan struct{})}
	go m.gc()
	return m
}

func (m *MemoryCache) Close() { close(m.stop) }

func (m *MemoryCache) gc() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			m.mu.Lock()
			for k, v := range m.items {
				if !v.expiry.IsZero() && now.After(v.expiry) {
					delete(m.items, k)
				}
			}
			m.mu.Unlock()
		case <-m.stop:
			return
		}
	}
}

func (m *MemoryCache) Get(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()
	if !ok {
		return "", ErrNil
	}
	if !item.expiry.IsZero() && time.Now().After(item.expiry) {
		m.mu.Lock()
		delete(m.items, key)
		m.mu.Unlock()
		return "", ErrNil
	}
	return item.value, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	item := memItem{value: fmt.Sprintf("%v", value)}
	if expiration > 0 {
		item.expiry = time.Now().Add(expiration)
	}
	m.mu.Lock()
	m.items[key] = item
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, k := range keys {
		delete(m.items, k)
	}
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) FlushDB(_ context.Context) error {
	m.mu.Lock()
	m.items = make(map[string]memItem)
	m.mu.Unlock()
	return nil
}

func (m *MemoryCache) DBSize(_ context.Context) (int64, error) {
	m.mu.RLock()
	n := len(m.items)
	m.mu.RUnlock()
	return int64(n), nil
}

func (m *MemoryCache) Info(_ context.Context) (string, error) {
	m.mu.RLock()
	n := len(m.items)
	m.mu.RUnlock()
	return fmt.Sprintf("memory_cache_keys:%d", n), nil
}

func (m *MemoryCache) Keys(_ context.Context, pattern string) ([]string, error) {
	prefix := strings.TrimSuffix(pattern, "*")
	m.mu.RLock()
	defer m.mu.RUnlock()
	var keys []string
	now := time.Now()
	for k, v := range m.items {
		if !v.expiry.IsZero() && now.After(v.expiry) {
			continue
		}
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
