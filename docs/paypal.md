## PayPal 对接（Orders v2 REST API）

### 1. 开发者账号与凭据

1. 打开 https://developer.paypal.com → **Log in to Dashboard**（用现有 PayPal 账号即可，无需额外注册开发者账号）。
2. **Apps & Credentials** → 默认 tab **Sandbox** → **Default Application** 里拿到：
   - `Client ID`（80 字符，`A...` 开头）
   - `Secret`（80 字符，`E...` 开头）
3. Sandbox 下 **Sandbox Accounts** 有 PayPal 预置的两个测试账号（business + personal），用 personal 去完成测试付款。

正式上线（`mode: live`）需要 PayPal **Business 账号**，不能用 Personal。Personal 账号可创建 Sandbox 应用但不能签约 Live 收款。

### 2. config.yaml

```yaml
paypal:
  mode: sandbox       # live 切换为 live
  client_id: "Aeog..."
  secret: "EHh3..."   # 生产请用环境变量/KMS，不要入 git
  webhook_id: "WH-..."     # 可选；下文
  currency: "USD"          # 国内账号通常不支持 CNY
  return_url: "https://yourdomain.com/api/pay/paypal/capture"
  cancel_url: "https://yourdomain.com/order/cancel"
```

### 3. 支付流程

```
[用户下单]
    │
    ▼
UnifiedPay(payment_key=paypal)
    │   POST /v2/checkout/orders (invoice_id = OrderNo, intent=CAPTURE)
    ▼
GoShop 返回 approval URL（PayPal 跳转页）
    │
    ▼
[用户在 PayPal 页面登录+批准付款]
    │
    ▼
PayPal → 用户 → return_url (?token=PP_ORDER_ID&PayerID=...)
    │
    ▼
GET /api/pay/paypal/capture?token=PP_ORDER_ID
    │   POST /v2/checkout/orders/{id}/capture
    ▼
HandlePayNotify(invoice_id=OrderNo, trade_no=capture_id)
    │
    ▼
订单 status: Pending → Paid，触发 postPaidHook（订单历史 + 通知）
```

同时 PayPal 会**异步**推 `PAYMENT.CAPTURE.COMPLETED` 到 `POST /api/pay/paypal/notify`，与 capture 路径**双保险**。两者都最终落到 `HandlePayNotify`，后者带 `WHERE status=pending` 的幂等保护，重复触发不会二次扣款。

### 4. Webhook 配置

1. Dashboard → **Webhooks** → **Add Webhook**
2. Webhook URL: `https://yourdomain.com/api/pay/paypal/notify`（必须公网 HTTPS）
3. 勾选事件（最少）：
   - `PAYMENT.CAPTURE.COMPLETED` — 核心，捕获完成即触发本地订单 Paid
   - `PAYMENT.CAPTURE.DENIED` / `PAYMENT.CAPTURE.REFUNDED`（可选，后续扩展）
4. 保存后页面显示 **Webhook ID**，填到 config `paypal.webhook_id`
5. `webhook_id` 配置后，`/api/pay/paypal/notify` 会调 `/v1/notifications/verify-webhook-signature` 同步验签。**生产必须配置**，dev 联调可以留空临时跳过验签

### 5. 退款

`pay.go` 中的 `RefundOrder` 按 `order.PaymentID → payment_key = paypal` 调 `PayPalDriver.Refund`。前提：
- `RefundLog.TradeNo` 必须是 PayPal **capture_id**（webhook/capture 路径已自动写入）
- 不能直接用 PayPal 的 `order_id` 去退款，必须先取 `capture_id`

### 6. 本地联调

```bash
# 启动后端
go build -o bin/goshop ./cmd/server && ./bin/goshop

# 用 sandbox 下单（假设已登录拿到 token $JWT）
curl -X POST localhost:8080/api/pay/unified \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"order_id": 1, "payment_key": "paypal"}'
# → 返回 {"pay_url":"https://www.sandbox.paypal.com/checkoutnow?token=..."}

# 浏览器访问 pay_url → 用 sandbox personal account 确认付款
# → 跳转回 return_url，GoShop 触发 capture
# → 订单状态 Pending → Paid
```

Webhook 本地联调建议用 [smee.io](https://smee.io) 或 ngrok 把公网 URL 映射到本机 8080。

### 7. 货币与金额

PayPal 金额按 **主货币单位**（元/美元），本项目订单 `pay_amount` 以**分**存储。`PayPalDriver.Pay` 内部做 `amount / 100.0` 转换，保留两位小数。注意：
- 不同国家账号可用币种不同。USD 是最通用的。
- 中国 PayPal 账号大部分场景不能以 CNY 收款，需落 USD 或与 PayPal 商议。
- **不要在 Config 里写 CNY 又期望 PayPal 直接结算**，先在 Dashboard 确认商户账号的币种能力。

### 8. 从 sandbox 切 live

1. `paypal.mode`: `sandbox` → `live`
2. 凭据换为 Live App 的 Client ID / Secret（Apps & Credentials → **Live** tab）
3. Webhook 新建一个 Live 环境的订阅，拿到新的 `webhook_id` 覆盖配置
4. 确认商户账号已完成 Business 验证 + 银行绑定 + 税号填写（PayPal 合规要求）
