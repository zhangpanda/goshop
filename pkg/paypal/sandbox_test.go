package paypal

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestSandbox_CreateOrderRoundTrip 真实打 PayPal Sandbox：OAuth → Create Order → 断言 approval link。
//
// 默认 Skip，仅在 GOSHOP_TEST_PAYPAL_SANDBOX_CLIENT_ID / _SECRET 配置时触发。
// 不触发 Capture（需要浏览器人工确认），也不触发 Refund（需要先 Capture）。
// 主要目的：验证 client 与真实 PayPal 的协议一致性（URL path、请求/响应字段）。
func TestSandbox_CreateOrderRoundTrip(t *testing.T) {
	id := os.Getenv("GOSHOP_TEST_PAYPAL_SANDBOX_CLIENT_ID")
	sec := os.Getenv("GOSHOP_TEST_PAYPAL_SANDBOX_SECRET")
	if id == "" || sec == "" {
		t.Skip("未配置 GOSHOP_TEST_PAYPAL_SANDBOX_CLIENT_ID / _SECRET；跳过真实 sandbox 联调")
	}

	c := NewClient("sandbox", id, sec)
	ctx := context.Background()

	tok, err := c.AccessToken(ctx)
	if err != nil {
		t.Fatalf("oauth: %v", err)
	}
	if len(tok) < 32 {
		t.Fatalf("token looks short: %d chars", len(tok))
	}

	orderNo := "GO-SANDBOX-" + randString(8)
	out, err := c.CreateOrder(ctx, CreateOrderReq{
		OrderNo:     orderNo,
		Description: "goshop sandbox smoke",
		Amount:      1234, // 分 → 12.34 USD
		Currency:    "USD",
		ReturnURL:   "https://example.com/ok",
		CancelURL:   "https://example.com/no",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if out.ID == "" {
		t.Fatal("empty order id")
	}
	if out.Status != "CREATED" && out.Status != "PAYER_ACTION_REQUIRED" {
		t.Fatalf("unexpected status: %s", out.Status)
	}
	approval := out.ApprovalURL()
	if approval == "" || !strings.Contains(approval, "paypal.com") {
		t.Fatalf("approval url looks wrong: %q", approval)
	}

	// 再查一次 GetOrder，确认 invoice_id/reference_id 回写正确
	got, err := c.GetOrder(ctx, out.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if len(got.PurchaseUnits) == 0 || got.PurchaseUnits[0].InvoiceID != orderNo {
		t.Fatalf("invoice_id 未回写: %+v", got)
	}

	t.Logf("sandbox OK: order_id=%s approval=%s (手工访问可完成付款)", out.ID, approval)
}

func randString(n int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[int(os.Getpid())%len(alphabet)]
		// 低熵即可，仅为防止测试间 OrderNo 碰撞
	}
	return string(b)
}
