package middleware

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracer 初始化 OTEL TracerProvider。
// 需要环境变量 OTEL_EXPORTER_OTLP_ENDPOINT（如 localhost:4317）才会启用；
// 未配置时返回 noop，不影响正常运行。
func InitTracer(ctx context.Context, serviceName string) (shutdown func(context.Context) error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return func(context.Context) error { return nil }
	}

	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		slog.Warn("otel", "action", "init_exporter", "err", err.Error())
		return func(context.Context) error { return nil }
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

// OtelGin 返回 otelgin 中间件（自动创建 span + 注入 trace context）。
func OtelGin(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// TraceIDFromContext 从 gin.Context 提取当前 trace ID（用于日志注入）。
func TraceIDFromContext(c *gin.Context) string {
	span := trace.SpanFromContext(c.Request.Context())
	if !span.SpanContext().HasTraceID() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}
