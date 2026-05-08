# GoShop 可观测性与 SLO

> 状态：初版。适用于单体 API 阶段；上游业务分支（订单/支付/分销等）有更精细的业务指标需求时再迭代。

## 1. 指标清单

GoShop API 通过 `server.metrics_path`（例：`/internal/metrics`）暴露 Prometheus 指标：

### 自定义（`internal/middleware/prometheus.go`）

| 指标 | 类型 | 标签 | 含义 |
|---|---|---|---|
| `goshop_http_in_flight` | Gauge | — | 当前正在处理的 HTTP 请求数（饱和度） |
| `goshop_http_requests_total` | Counter | `method`, `route`, `status` | HTTP 请求累计（按 Gin `FullPath()` 聚合，未命中路由=`unknown`） |
| `goshop_http_request_duration_seconds` | Histogram | `method`, `route`, `status` | 请求耗时（默认 bucket） |

### 自动注册（Go 运行时 + process）

`promhttp.Handler()` 自带：`go_goroutines`, `go_gc_duration_seconds_*`, `go_memstats_*`, `process_resident_memory_bytes`, `process_cpu_seconds_total`, `up` 等。

> **缺口**：业务指标（支付成功/失败、退款失败、佣金结算、cron leader 状态等）尚未加入。与 HANDOVER P3 "测试纵深/观测" 保持一致；新增需伴随 wiring 而非空架子。

## 2. SLO 定义

| SLO | 目标 | 观测窗口 | 错误预算 |
|---|---|---|---|
| **可用性** | 非 5xx 比例 ≥ **99.9%** | 30 天滚动 | 0.1% ≈ **43.2 min / 30d** |
| **延迟（非 admin 接口）** | P99 < **1s** | 30 天滚动的 95% 时间 | 5% 时间可越过阈值 |

备注：
- `admin` 接口（`/api/admin/*`）因包含重 CRUD / 导出 / 统计，暂不进延迟 SLO 约束；仍监控但不烧错误预算。
- SLO 是**团队承诺**，不是单点告警。告警阈值另定（见下）。

## 3. 告警策略（对应 `deploy/prometheus/rules.yml`）

分级以 **消费路径** 为准：

| severity | 响应 | 示例 |
|---|---|---|
| `page` | 立即分页（电话 / 钉钉值班） | 实例掉线、快烧（1h 5xx>2%）、P99>3s |
| `ticket` | 工作时间内处理 | 慢烧（24h 5xx>0.1%）、goroutine>10k、单路由 5xx>5% |
| `info` | 只上 dashboard，不主动告警 | GC pause、内存增长趋势 |

关键告警：

- **`GoshopSLOAvailabilityBurnFast`**：1h 错误率 > 2% 持续 5m → page。按此速率 1 天内耗尽整月预算。
- **`GoshopSLOAvailabilityBurn24h`**：24h 错误率 > 0.1% 持续 30m → ticket。SLO 慢烧。
- **`GoshopLatencyP99Critical`**：非 admin P99 > 3s 持续 5m → page。
- **`GoshopInstanceDown`**：`up == 0` 持续 1m → page。
- **`GoshopHTTPSaturation`** / **`GoshopGoroutineLeak`**：饱和度与泄漏趋势 → ticket。

完整表达式见 rules.yml。

## 4. 常用 PromQL

```promql
# 全站 QPS
sum(rate(goshop_http_requests_total[1m]))

# Top 5 最慢路由 (P99)
topk(5, goshop:http_latency_by_route:p99_5m)

# Top 5 错误最多的路由（5xx 绝对速率）
topk(5, goshop:http_errors_5xx:rate5m)

# 30d 可用性
1 - (
  sum(increase(goshop_http_requests_total{status=~"5.."}[30d]))
  /
  sum(increase(goshop_http_requests_total[30d]))
)

# 错误预算剩余（30d，目标 99.9%）
1 - (
  (
    sum(increase(goshop_http_requests_total{status=~"5.."}[30d]))
    /
    sum(increase(goshop_http_requests_total[30d]))
  ) / 0.001
)

# in-flight 饱和度
goshop_http_in_flight

# 按方法分布
sum by (method) (rate(goshop_http_requests_total[5m]))
```

## 5. Runbook（常见告警处置）

### 5.1 `availability-burn`（错误预算烧尽）

1. 在 Prometheus / Grafana 上查 `topk(5, goshop:http_errors_5xx:rate5m)`，锁定故障路由。
2. 读该路由对应 handler 的 error 日志；按 `request_id` / `trace_id` 去 DB / 第三方日志关联（目前 GoShop 日志是 `slog`，结构化字段 `route` / `admin_id` / `err` 等可直接 grep）。
3. 若故障集中在支付类路径（`/api/pay/*`、`/api/pay/notify` 等），优先确认：
   - `CronRefundReconcile`、`CronPayReconcile` 日志是否有大量 failed
   - 微信 / 支付宝侧 HTTP 返回码（走 `pkg/httpx.Client`，timeout=10s）
4. 必要时降级：
   - 临时关闭单一支付驱动：管理后台 `支付方式 → 停用`
   - 关闭限速中间件的严格分母（见 `internal/middleware/ratelimit.go`）

### 5.2 `latency-P99-critical`

1. 看 `topk(5, goshop:http_latency_by_route:p99_5m)`，定位慢接口。
2. 确认 DB：`show processlist`，是否有长事务或未索引查询。
3. 确认缓存：Redis 是否可达（Incr/Set/SetNX 的限速、验证码、cron leader 都依赖）。
4. 确认下游：微信/支付宝/快递 100 是否挂（`httpx.Client` 会在 10s 失败，但未抖动的前提下可能整体变慢）。
5. 如果只影响单一路由，可考虑 `ratelimit` 针对该路由收紧限速，避免拖挂整机。

### 5.3 `goroutine-leak`

1. 确认 `GOSHOP_PPROF=1` 已启用（默认关闭）。取 pprof：`curl -s $HOST/internal/pprof/goroutine?debug=2 > /tmp/g.txt`
2. 在 `g.txt` 中 `grep -c 'goroutine'` 得总数，再 `sort | uniq -c | sort -rn | head` 找到最常见的栈顶。
3. 高发嫌疑区：
   - `internal/service/platform_integral.go` 的微信 access_token 路径（已接 httpx，但仍建议确认）
   - `app.SafeGo` 的调用者是否漏了 context/超时（SafeGo 只管 panic，不管长住）
4. 确认 cron leader 是否被 Redis 故障导致多实例同时抢主——看 `slog` 中 `cron` 相关行。

### 5.4 `instance-down`

1. `kubectl describe pod / docker logs`，看退出原因。
2. 如果是 OOMKilled：看 `GoshopMemoryGrowth` 是否先期触发。
3. 如果是 panic：slog 会打 stack（`app.SafeGo` 对 goroutine panic 有 recover，但主路径 panic 仍 crash 进程）。
4. 回滚策略：tag 降级到 `v1.6.0`（Release 页可直接下 binary）。

## 6. 下一步（未覆盖）

- **业务指标**：加 `goshop_pay_success_total{provider}` / `goshop_refund_total{status}` / `goshop_commission_settled_total` 等，便于财务对账直接上 dashboard。
- **trace**：接入 OpenTelemetry Gin instrumentation，跨 DB / 第三方 HTTP 传播 traceparent；目前只有 `reqid` 中间件生成请求 ID。
- **日志统一**：所有 slog 字段命名规范化（route / err / order_id / request_id / admin_id）；现在约 80% 已统一。
- **Grafana dashboard**：本文的 PromQL 建议固化为 Grafana JSON；暂由接入方自行组装。

