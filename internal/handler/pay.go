package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/paypal"
	"github.com/zhangpanda/goshop/pkg/response"
)

func PayOrder(c *gin.Context) {
	var req service.PayOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := service.PayOrder(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

// UnifiedPay 统一支付入口（支持多支付方式）
func UnifiedPay(c *gin.Context) {
	var req service.UnifiedPayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ClientIP = c.ClientIP()
	resp, err := service.UnifiedPay(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

// AlipayNotify 支付宝异步回调（含RSA2验签）
func AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	params := make(map[string]string, len(c.Request.PostForm))
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if !service.AlipayVerifySign(params, app.Must().Cfg.Alipay.PublicKey) {
		c.String(http.StatusOK, "fail")
		return
	}
	status := params["trade_status"]
	if status == "TRADE_SUCCESS" || status == "TRADE_FINISHED" {
		if err := service.HandlePayNotify(params["out_trade_no"], params["trade_no"]); err != nil {
			c.String(http.StatusOK, "fail")
			return
		}
	}
	c.String(http.StatusOK, "success")
}

func PayNotify(c *gin.Context) {
	handler, err := notify.NewRSANotifyHandler(app.Must().Cfg.Wechat.MchAPIKey, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "初始化失败"})
		return
	}

	var transaction payments.Transaction
	_, err = handler.ParseNotifyRequest(c.Request.Context(), c.Request, &transaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "验签失败"})
		return
	}

	if transaction.OutTradeNo == nil || transaction.TransactionId == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "缺少交易信息"})
		return
	}

	if err := service.HandlePayNotify(*transaction.OutTradeNo, *transaction.TransactionId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// SandboxCallback 沙盒支付回调（模拟第三方回调，仅 payment.sandbox=true 时可用）
func SandboxCallback(c *gin.Context) {
	if !app.Must().Cfg.Payment.Sandbox {
		response.Fail(c, http.StatusForbidden, "沙盒模式未开启")
		return
	}
	orderNo := c.Query("order_no")
	tradeNo := c.Query("trade_no")
	if orderNo == "" {
		response.Fail(c, http.StatusBadRequest, "缺少 order_no")
		return
	}
	if tradeNo == "" {
		tradeNo = "SANDBOX_" + orderNo
	}
	if err := service.HandlePayNotify(orderNo, tradeNo); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{
		"message":  "沙盒支付成功",
		"order_no": orderNo,
		"trade_no": tradeNo,
	})
}

func RefundOrder(c *gin.Context) {
	var req service.RefundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.RefundOrder(c.GetUint("user_id"), &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

// PayPalNotify PayPal webhook 入口（/api/pay/paypal/notify）。
//
// 行为：
//  1. 读原始请求体 + transmission 相关 header
//  2. 若 paypal.webhook_id 已配置 → 调 PayPal VerifyWebhook 同步验签；失败返回 400。
//     若未配置 → 跳过验签（仅限 dev/联调，生产务必配置 webhook_id）。
//  3. 只处理 PAYMENT.CAPTURE.COMPLETED：从 resource 提取 invoice_id（= 本地 OrderNo）
//     与 capture id（第三方交易号），调 service.HandlePayNotify 幂等更新订单 + 触发副作用。
//  4. 其他事件（APPROVED / DENIED / REFUNDED 等）当前仅 200 应答，不处理；后续按需扩展。
func PayPalNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "read body: %v", err)
		return
	}

	cfg := app.Must().Cfg.PayPal
	if cfg.ClientID == "" || cfg.Secret == "" {
		c.String(http.StatusServiceUnavailable, "paypal not configured")
		return
	}

	// 如果配置了 webhook_id 就严格验签；否则放行并打 warn（dev 模式）
	if cfg.WebhookID != "" {
		client := paypal.NewClient(cfg.Mode, cfg.ClientID, cfg.Secret)
		ok, verr := client.VerifyWebhook(c.Request.Context(), paypal.VerifyWebhookReq{
			AuthAlgo:         c.GetHeader("PAYPAL-AUTH-ALGO"),
			CertURL:          c.GetHeader("PAYPAL-CERT-URL"),
			TransmissionID:   c.GetHeader("PAYPAL-TRANSMISSION-ID"),
			TransmissionSig:  c.GetHeader("PAYPAL-TRANSMISSION-SIG"),
			TransmissionTime: c.GetHeader("PAYPAL-TRANSMISSION-TIME"),
			WebhookID:        cfg.WebhookID,
			WebhookEvent:     body,
		})
		if verr != nil || !ok {
			slog.Warn("paypal webhook verify failed", "err", verr, "ok", ok)
			c.String(http.StatusBadRequest, "verify failed")
			return
		}
	} else {
		slog.Warn("paypal webhook received without webhook_id configured; verification skipped (dev only)")
	}

	// 事件解析
	var ev struct {
		ID           string          `json:"id"`
		EventType    string          `json:"event_type"`
		ResourceType string          `json:"resource_type"`
		Resource     json.RawMessage `json:"resource"`
	}
	if err := json.Unmarshal(body, &ev); err != nil {
		c.String(http.StatusBadRequest, "decode: %v", err)
		return
	}

	switch ev.EventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		var r struct {
			ID        string `json:"id"`         // capture id
			InvoiceID string `json:"invoice_id"` // = 本地 OrderNo
			Status    string `json:"status"`     // COMPLETED
		}
		if err := json.Unmarshal(ev.Resource, &r); err != nil {
			c.String(http.StatusBadRequest, "resource decode: %v", err)
			return
		}
		if r.InvoiceID == "" || r.ID == "" {
			slog.Warn("paypal capture completed missing fields", "event_id", ev.ID, "resource", string(ev.Resource))
			c.String(http.StatusOK, "ignored")
			return
		}
		if err := service.HandlePayNotify(r.InvoiceID, r.ID); err != nil {
			slog.Warn("paypal HandlePayNotify", "invoice", r.InvoiceID, "err", err)
			c.String(http.StatusInternalServerError, "handle: %v", err)
			return
		}
		c.String(http.StatusOK, "ok")
	default:
		// 未处理的事件类型直接 200 应答，避免 PayPal 重试
		c.String(http.StatusOK, "ignored event: %s", ev.EventType)
	}
}

// PayPalCapture 同步捕获 PayPal order（用户 approval 后前端 return_url 携带 ?token=ORDER_ID 回跳时调用）。
//
// 与 Webhook 路径双保险：两者都最终调 HandlePayNotify，后者带 status=Pending 的幂等保护，
// 重复触发不会二次扣款/重复发通知。
//
// 失败路径（用户取消、金额不符等）不改订单本地状态，仍是 Pending，等超时 cron 关单或用户重新发起。
func PayPalCapture(c *gin.Context) {
	paypalOrderID := c.Query("token")
	if paypalOrderID == "" {
		paypalOrderID = c.Query("order_id")
	}
	if paypalOrderID == "" {
		response.Fail(c, http.StatusBadRequest, "missing token/order_id")
		return
	}
	cfg := app.Must().Cfg.PayPal
	if cfg.ClientID == "" || cfg.Secret == "" {
		response.Fail(c, http.StatusServiceUnavailable, "paypal not configured")
		return
	}
	client := paypal.NewClient(cfg.Mode, cfg.ClientID, cfg.Secret)
	out, err := client.CaptureOrder(c.Request.Context(), paypalOrderID)
	// PayPal 的幂等性：若订单已被捕获（例如重复点击 return_url），CaptureOrder 会返回
	// 422 ORDER_ALREADY_CAPTURED。此时降级到 GetOrder 拿现有 capture 信息，再走同样的
	// HandlePayNotify（它本身幂等，重复调用不会二次扣款/重复通知）。
	if err != nil && strings.Contains(err.Error(), "ORDER_ALREADY_CAPTURED") {
		got, gerr := client.GetOrder(c.Request.Context(), paypalOrderID)
		if gerr != nil {
			response.Fail(c, http.StatusBadGateway, "capture already done, but GetOrder failed: "+gerr.Error())
			return
		}
		// GetOrder 的 invoice_id 在 purchase_unit 外层，但没有 capture id —— 需要再查 captures 接口。
		// 简化处理：如果 GetOrder 响应里有 invoice_id，就用它 + paypalOrderID 作 trade_no。
		if len(got.PurchaseUnits) > 0 && got.PurchaseUnits[0].InvoiceID != "" {
			if err := service.HandlePayNotify(got.PurchaseUnits[0].InvoiceID, paypalOrderID); err != nil {
				response.Fail(c, http.StatusInternalServerError, err.Error())
				return
			}
			response.OK(c, gin.H{
				"order_no":   got.PurchaseUnits[0].InvoiceID,
				"capture_id": "",
				"note":       "已捕获；capture_id 未从 GetOrder 取到，退款前请到 PayPal Dashboard 查 capture 记录",
			})
			return
		}
		response.Fail(c, http.StatusInternalServerError, "capture already done; cannot resolve invoice_id")
		return
	}
	if err != nil {
		slog.Warn("paypal capture", "order_id", paypalOrderID, "err", err)
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	if out.Status != "COMPLETED" {
		response.Fail(c, http.StatusBadRequest, "capture status="+out.Status)
		return
	}
	var invoice, captureID string
	if len(out.PurchaseUnits) > 0 {
		pu := out.PurchaseUnits[0]
		if caps := pu.Payments.Captures; len(caps) > 0 {
			captureID = caps[0].ID
			invoice = caps[0].InvoiceID // 首选：PP capture 响应里 invoice_id 的真实位置
		}
		if invoice == "" {
			invoice = pu.InvoiceID // 兜底：某些响应可能也放在 purchase_unit 外层
		}
	}
	if invoice == "" || captureID == "" {
		response.Fail(c, http.StatusBadRequest, "capture 响应缺 invoice_id 或 capture_id")
		return
	}
	if err := service.HandlePayNotify(invoice, captureID); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"order_no":   invoice,
		"capture_id": captureID,
	})
}
