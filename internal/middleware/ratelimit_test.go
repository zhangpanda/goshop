package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/global"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	if global.Cfg == nil {
		global.Cfg = &config.Config{}
	}
	os.Exit(m.Run())
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", RateLimit(5, time.Minute), func(c *gin.Context) {
		c.String(200, "ok")
	})

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("request %d: got %d, want 200", i, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", RateLimit(3, time.Minute), func(c *gin.Context) {
		c.String(200, "ok")
	})

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)
		if i < 3 && w.Code != 200 {
			t.Fatalf("request %d: got %d, want 200", i, w.Code)
		}
		if i >= 3 && w.Code != 429 {
			t.Fatalf("request %d: got %d, want 429", i, w.Code)
		}
	}
}

func TestRateLimit_SlidingWindowExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", RateLimit(2, 100*time.Millisecond), func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Use up the limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		r.ServeHTTP(w, req)
	}

	// Should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	r.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	// Wait for window to expire
	time.Sleep(120 * time.Millisecond)

	// Should be allowed again
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("after window expiry: got %d, want 200", w.Code)
	}
}

func TestRateLimit_DifferentIPsIndependent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", RateLimit(1, time.Minute), func(c *gin.Context) {
		c.String(200, "ok")
	})

	// IP A uses its quota
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("IP A first request should pass")
	}

	// IP B should still be allowed
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("IP B should be independent from IP A")
	}
}
