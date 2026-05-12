package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 业务指标：下单/支付/退款计数器。
// 在关键路径上 Inc，Prometheus 抓取后可用于告警和 Dashboard。
var (
	MetricOrderCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goshop_order_created_total",
		Help: "Total number of orders created",
	})
	MetricPaySuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goshop_pay_success_total",
		Help: "Total number of successful payments",
	})
	MetricPayTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goshop_pay_total",
		Help: "Total number of payment attempts (notify received)",
	})
	MetricRefundSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goshop_refund_success_total",
		Help: "Total number of successful refunds",
	})
)
