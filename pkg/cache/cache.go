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
	// Incr 原子自增，常用于计数器/限速窗口。Key 不存在时从 0 开始并返回 1。
	// Incr 本身不设置 TTL；首次调用后请配合 Expire 建立过期窗口。
	Incr(ctx context.Context, key string) (int64, error)
	// Expire 为已存在的 key 设置 TTL；key 不存在时返回 nil（不是错误）。
	Expire(ctx context.Context, key string, ttl time.Duration) error
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

// Incr 原子自增；key 不存在时返回 1。Incr 不会重置 TTL。
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Expire 设置 key 的 TTL；key 不存在时 Redis 返回 false，此处仅在网络错误时上报。
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
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

// Incr 原子自增；过期后再 Incr 视为从 0 开始。
// 注意：当前实现不保留 TTL（与 Redis INCR 的语义一致：INCR 不改过期时间）；
// 首次调用后请显式 Expire 建立窗口。
func (m *MemoryCache) Incr(_ context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[key]
	var n int64
	if ok && (item.expiry.IsZero() || time.Now().Before(item.expiry)) {
		_, _ = fmt.Sscanf(item.value, "%d", &n)
	}
	n++
	item.value = fmt.Sprintf("%d", n)
	m.items[key] = item // 保留原有 expiry；若 item 为零值则 expiry 为 zero（无限期）
	return n, nil
}

// Expire 为已存在的 key 设置 TTL；key 不存在返回 nil（与 Redis Expire 命令对齐，仅以 error 为异常）。
func (m *MemoryCache) Expire(_ context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if it, ok := m.items[key]; ok {
		it.expiry = time.Now().Add(ttl)
		m.items[key] = it
	}
	return nil
}
