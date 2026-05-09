// Package paypal 提供 PayPal REST API v2 客户端封装：
//   - OAuth2 Client Credentials 令牌获取（带内存缓存，提前 60s 过期避免边界抖动）
//   - Orders v2: Create / Capture（替代已废弃的 Payments v1）
//   - Payments v2: Refund
//   - Webhooks: 同步验签（用于异步通知回调）
//
// 所有 HTTP 请求走 pkg/httpx.Client（10s 超时）。日志打点只记录必要字段，不泄露 secret。
//
// 使用示例：
//
//	c := paypal.NewClient("sandbox", clientID, secret)
//	order, err := c.CreateOrder(ctx, paypal.CreateOrderReq{...})
//	cap, err := c.CaptureOrder(ctx, order.ID)
//	err = c.RefundCapture(ctx, cap.PurchaseUnits[0].Payments.Captures[0].ID, paypal.RefundReq{...})
package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zhangpanda/goshop/pkg/httpx"
)

const (
	BaseSandbox = "https://api-m.sandbox.paypal.com"
	BaseLive    = "https://api-m.paypal.com"
)

// Client 单例线程安全；Token 在 mu 保护下缓存。
type Client struct {
	base     string
	clientID string
	secret   string

	mu        sync.Mutex
	token     string
	tokenExp  time.Time
	tokenSkew time.Duration // 提前过期缓冲，默认 60s
}

// NewClient 根据 mode 选择 sandbox 或 live 基地址。未识别的 mode 一律按 sandbox 处理。
func NewClient(mode, clientID, secret string) *Client {
	base := BaseSandbox
	if mode == "live" {
		base = BaseLive
	}
	return &Client{
		base:      base,
		clientID:  clientID,
		secret:    secret,
		tokenSkew: 60 * time.Second,
	}
}

// BaseURL 导出供外部构造 approval link 等调试场景使用。
func (c *Client) BaseURL() string { return c.base }

// AccessToken 返回有效访问令牌；若缓存未命中则调用 /v1/oauth2/token。
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.tokenExp.Add(-c.tokenSkew)) {
		return c.token, nil
	}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.base+"/v1/oauth2/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.secret)

	resp, err := httpx.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("paypal oauth: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("paypal oauth %d: %s", resp.StatusCode, snippet(body))
	}
	var ok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &ok); err != nil {
		return "", fmt.Errorf("paypal oauth decode: %w", err)
	}
	if ok.AccessToken == "" {
		return "", errors.New("paypal oauth: empty access_token")
	}
	c.token = ok.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(ok.ExpiresIn) * time.Second)
	return c.token, nil
}

// do 为 PayPal REST 统一发请求；自动附加 Bearer 与 JSON header，4xx/5xx 时读出错误体。
// extraHeaders 覆盖/追加请求头，用于 webhook 验签等特殊场景（此处主要 Create/Capture）。
func (c *Client) do(ctx context.Context, method, path string, body any, extraHeaders map[string]string) ([]byte, error) {
	token, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := httpx.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("paypal %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return raw, fmt.Errorf("paypal %s %s: HTTP %d %s", method, path, resp.StatusCode, snippet(raw))
	}
	return raw, nil
}

// ========== Orders v2 ==========

// CreateOrderReq 主力字段；完整 spec 详见 https://developer.paypal.com/docs/api/orders/v2/
type CreateOrderReq struct {
	OrderNo     string // 用作 purchase_units[].invoice_id + reference_id，回调时反查本地订单
	Description string // 商品描述
	Amount      int64  // 金额，分（本地 int64）
	Currency    string // ISO-4217，如 "USD"/"CNY"；空则 USD
	ReturnURL   string
	CancelURL   string
}

type OrderLink struct {
	HREF   string `json:"href"`
	Rel    string `json:"rel"`    // "approve" 是让用户付款的关键链接
	Method string `json:"method"` // "GET" / "POST"
}

type CreateOrderResp struct {
	ID     string      `json:"id"`
	Status string      `json:"status"` // "CREATED"
	Links  []OrderLink `json:"links"`
}

// ApprovalURL 从 Links 中挑出 rel=approve 的 href，用户跳转该 URL 确认付款。
func (r *CreateOrderResp) ApprovalURL() string {
	for _, l := range r.Links {
		if l.Rel == "approve" {
			return l.HREF
		}
	}
	return ""
}

func (c *Client) CreateOrder(ctx context.Context, req CreateOrderReq) (*CreateOrderResp, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	body := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{{
			"reference_id": req.OrderNo,
			"invoice_id":   req.OrderNo,
			"description":  truncate(req.Description, 127),
			"amount": map[string]any{
				"currency_code": currency,
				"value":         fmt.Sprintf("%.2f", float64(req.Amount)/100),
			},
		}},
		"application_context": map[string]any{
			"return_url":  req.ReturnURL,
			"cancel_url":  req.CancelURL,
			"user_action": "PAY_NOW",
		},
	}
	raw, err := c.do(ctx, http.MethodPost, "/v2/checkout/orders", body, nil)
	if err != nil {
		return nil, err
	}
	var out CreateOrderResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("paypal create order decode: %w", err)
	}
	if out.ID == "" {
		return nil, fmt.Errorf("paypal create order: empty id, raw=%s", snippet(raw))
	}
	return &out, nil
}

// CaptureOrderResp 仅抓取关键字段；完整响应字段较多。
// 注意：PayPal capture 响应中 invoice_id 实际在 captures[] 元素内部，
// 而不是在 purchase_units[] 外层。两个位置都保留字段以兼容不同场景
// （GetOrder 返回的 PurchaseUnit 有 invoice_id，CaptureOrder 的不一定有）。
type CaptureOrderResp struct {
	ID            string `json:"id"`
	Status        string `json:"status"` // COMPLETED/APPROVED/FAILED
	PurchaseUnits []struct {
		ReferenceID string `json:"reference_id"`
		InvoiceID   string `json:"invoice_id"` // 仅 GetOrder 响应可靠返回
		Payments    struct {
			Captures []struct {
				ID        string `json:"id"`         // 捕获 ID，退款时要用
				Status    string `json:"status"`     // COMPLETED
				InvoiceID string `json:"invoice_id"` // 真实位置；CreateOrder 时传的 invoice_id 会在这里回显
				Amount    struct {
					Value        string `json:"value"`
					CurrencyCode string `json:"currency_code"`
				} `json:"amount"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

func (c *Client) CaptureOrder(ctx context.Context, orderID string) (*CaptureOrderResp, error) {
	raw, err := c.do(ctx, http.MethodPost,
		"/v2/checkout/orders/"+url.PathEscape(orderID)+"/capture", nil, nil)
	if err != nil {
		return nil, err
	}
	var out CaptureOrderResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("paypal capture decode: %w", err)
	}
	return &out, nil
}

// GetOrderResp 只关心核心字段；可用于查单对账。
type GetOrderResp struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		ReferenceID string `json:"reference_id"`
		InvoiceID   string `json:"invoice_id"`
	} `json:"purchase_units"`
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*GetOrderResp, error) {
	raw, err := c.do(ctx, http.MethodGet,
		"/v2/checkout/orders/"+url.PathEscape(orderID), nil, nil)
	if err != nil {
		return nil, err
	}
	var out GetOrderResp
	return &out, json.Unmarshal(raw, &out)
}

// ========== Payments v2: Refund ==========

type RefundReq struct {
	InvoiceID   string // 商家退款单号（回写本地 RefundLog.refund_no）
	Amount      int64  // 退款金额（分）；0 表示全额退
	Currency    string // 空则与原订单保持
	NoteToPayer string // 可选
}

type RefundResp struct {
	ID     string `json:"id"`
	Status string `json:"status"` // COMPLETED/PENDING/FAILED
}

// RefundCapture 对指定 capture 发起退款；全额时 amount 传空（PayPal 要求不带 amount 字段）。
func (c *Client) RefundCapture(ctx context.Context, captureID string, req RefundReq) (*RefundResp, error) {
	body := map[string]any{
		"invoice_id":    req.InvoiceID,
		"note_to_payer": req.NoteToPayer,
	}
	if req.Amount > 0 {
		currency := req.Currency
		if currency == "" {
			currency = "USD"
		}
		body["amount"] = map[string]any{
			"currency_code": currency,
			"value":         fmt.Sprintf("%.2f", float64(req.Amount)/100),
		}
	}
	raw, err := c.do(ctx, http.MethodPost,
		"/v2/payments/captures/"+url.PathEscape(captureID)+"/refund", body, nil)
	if err != nil {
		return nil, err
	}
	var out RefundResp
	return &out, json.Unmarshal(raw, &out)
}

// ========== Webhook 验签 ==========

// VerifyWebhookReq 对应 PayPal /v1/notifications/verify-webhook-signature 入参。
type VerifyWebhookReq struct {
	AuthAlgo         string          `json:"auth_algo"`
	CertURL          string          `json:"cert_url"`
	TransmissionID   string          `json:"transmission_id"`
	TransmissionSig  string          `json:"transmission_sig"`
	TransmissionTime string          `json:"transmission_time"`
	WebhookID        string          `json:"webhook_id"`
	WebhookEvent     json.RawMessage `json:"webhook_event"` // 原始事件 JSON
}

// VerifyWebhook 返回 true 表示签名成功；失败（含验签失败、网络错误）一律 false 并上报原始错误。
func (c *Client) VerifyWebhook(ctx context.Context, req VerifyWebhookReq) (bool, error) {
	if req.WebhookID == "" {
		return false, errors.New("webhook_id empty (set paypal.webhook_id in config)")
	}
	raw, err := c.do(ctx, http.MethodPost, "/v1/notifications/verify-webhook-signature", req, nil)
	if err != nil {
		return false, err
	}
	var ok struct {
		VerificationStatus string `json:"verification_status"` // SUCCESS / FAILURE
	}
	if err := json.Unmarshal(raw, &ok); err != nil {
		return false, fmt.Errorf("paypal verify decode: %w", err)
	}
	return ok.VerificationStatus == "SUCCESS", nil
}

// ========== utils ==========

func snippet(b []byte) string {
	if len(b) > 240 {
		return string(b[:240]) + "…"
	}
	return string(b)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
