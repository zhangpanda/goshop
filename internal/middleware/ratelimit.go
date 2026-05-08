package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/pkg/response"
)

// slidingWindow 进程内滑动窗口（单机限流）
type slidingWindow struct {
	mu      sync.Mutex
	windows map[string][]time.Time
	limit   int
	window  time.Duration
}

func newSlidingWindow(limit int, window time.Duration) *slidingWindow {
	sw := &slidingWindow{
		windows: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
	go sw.cleanup()
	return sw
}

func (sw *slidingWindow) allow(key string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	timestamps := sw.windows[key]
	start := 0
	for start < len(timestamps) && timestamps[start].Before(cutoff) {
		start++
	}
	timestamps = timestamps[start:]

	if len(timestamps) >= sw.limit {
		sw.windows[key] = timestamps
		return false
	}

	sw.windows[key] = append(timestamps, now)
	return true
}

func (sw *slidingWindow) cleanup() {
	ticker := time.NewTicker(sw.window)
	defer ticker.Stop()
	for range ticker.C {
		sw.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-sw.window)
		for k, ts := range sw.windows {
			start := 0
			for start < len(ts) && ts[start].Before(cutoff) {
				start++
			}
			if start == len(ts) {
				delete(sw.windows, k)
			} else {
				sw.windows[k] = ts[start:]
			}
		}
		sw.mu.Unlock()
	}
}

type memLimiter struct {
	sw *slidingWindow
}

func (m *memLimiter) allow(_ context.Context, key string) (bool, error) {
	return m.sw.allow(key), nil
}

// redisLuaSliding ZSET 滑动窗口，多实例共享
var redisLuaSliding = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_ns = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]
local min = now - window_ns
redis.call('ZREMRANGEBYSCORE', key, '0', tostring(min))
local n = redis.call('ZCARD', key)
if n < limit then
  redis.call('ZADD', key, now, member)
  local ttl_ms = math.floor(window_ns / 1000000) + 2000
  redis.call('PEXPIRE', key, ttl_ms)
  return 1
end
return 0
`)

type redisLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func (r *redisLimiter) allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	member := fmt.Sprintf("%d-%s", now, randHex(8))
	v, err := redisLuaSliding.Run(ctx, r.rdb, []string{"goshop:ratelimit:" + key},
		strconv.FormatInt(now, 10),
		strconv.FormatInt(r.window.Nanoseconds(), 10),
		strconv.Itoa(r.limit),
		member).Int()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

func randHex(n int) string {
	b := make([]byte, n/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func rateLimitBackendMode() string {
	if app.Must().Cfg == nil {
		return "auto"
	}
	b := app.Must().Cfg.Server.RateLimitBackend
	if b == "" {
		return "auto"
	}
	return b
}

func newLimiter(limit int, window time.Duration) interface {
	allow(ctx context.Context, key string) (bool, error)
} {
	mode := rateLimitBackendMode()
	useRedis := app.Must().RDB != nil && (mode == "redis" || mode == "auto")
	if useRedis {
		return &redisLimiter{rdb: app.Must().RDB, limit: limit, window: window}
	}
	if mode == "redis" && app.Must().RDB == nil {
		slog.Warn("ratelimit", "backend", "memory", "reason", "rate_limit_backend=redis but Redis unavailable")
	}
	return &memLimiter{sw: newSlidingWindow(limit, window)}
}

// RateLimit 滑动窗口限流：Redis 可用且后端为 auto/redis 时用集群限流，否则进程内限流。
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	lim := newLimiter(limit, window)
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if uid, exists := c.Get("user_id"); exists {
			userKey := "uid:" + formatUID(uid)
			ok, err := lim.allow(ctx, userKey)
			if err != nil {
				slog.Warn("ratelimit", "key", "user", "err", err.Error())
				response.Fail(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
				c.Abort()
				return
			}
			if !ok {
				response.Fail(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
				c.Abort()
				return
			}
		}
		ipKey := "ip:" + c.ClientIP()
		ok, err := lim.allow(ctx, ipKey)
		if err != nil {
			slog.Warn("ratelimit", "key", "ip", "err", err.Error())
			response.Fail(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		if !ok {
			response.Fail(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

func formatUID(v interface{}) string {
	switch id := v.(type) {
	case uint:
		return uintToStr(id)
	case int:
		return uintToStr(uint(id))
	default:
		return "0"
	}
}

func uintToStr(n uint) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
