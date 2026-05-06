package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/reqid"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := reqid.New()
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)

		start := time.Now()
		c.Next()

		slog.Info("request",
			"request_id", id,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}
