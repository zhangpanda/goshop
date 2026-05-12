package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// InjectTraceID 将 trace_id 写入 gin.Context 的 key，供 logger middleware 使用。
func InjectTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().HasTraceID() {
			c.Set("trace_id", span.SpanContext().TraceID().String())
		}
		c.Next()
	}
}
