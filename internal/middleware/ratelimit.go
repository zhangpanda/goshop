package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/pkg/response"
)

// slidingWindow tracks request timestamps within a window.
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

	// Trim expired entries
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

// RateLimit returns a sliding-window rate limiter keyed by IP + authenticated UserID.
// limit: max requests within window per key.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	sw := newSlidingWindow(limit, window)
	return func(c *gin.Context) {
		// Primary key: IP
		key := "ip:" + c.ClientIP()

		// If user is authenticated, also enforce per-user limit
		if uid, exists := c.Get("user_id"); exists {
			userKey := "uid:" + formatUID(uid)
			if !sw.allow(userKey) {
				response.Fail(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
				c.Abort()
				return
			}
		}

		if !sw.allow(key) {
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
