package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goshop",
		Name:      "http_in_flight",
		Help:      "当前正在处理的 HTTP 请求数",
	})
	httpReqTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goshop",
		Name:      "http_requests_total",
		Help:      "HTTP 请求总数（按方法、路由模板、状态码）",
	}, []string{"method", "route", "status"})
	httpReqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goshop",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP 请求耗时（秒）",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "route", "status"})
)

// PrometheusHTTP 记录请求计数与直方图（依赖 route 在 c.Next() 之后由 Gin 写入 FullPath）。
func PrometheusHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		httpInFlight.Inc()
		start := time.Now()
		c.Next()
		httpInFlight.Dec()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())
		lbl := []string{c.Request.Method, route, status}
		httpReqTotal.WithLabelValues(lbl...).Inc()
		httpReqDuration.WithLabelValues(lbl...).Observe(time.Since(start).Seconds())
	}
}
