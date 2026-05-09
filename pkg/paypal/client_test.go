package paypal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newStub 用 httptest 模拟 PayPal REST 服务；返回的 Client 已注入 stub.URL 作为 base，
// 并强制将 token 有效期设置得够长（因此测试内多个请求只会触发一次 /v1/oauth2/token）。
func newStub(t *testing.T, routes map[string]http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	c := NewClient("sandbox", "id", "sec")
	c.base = srv.URL
	return srv, c
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func TestAccessToken_Cached(t *testing.T) {
	var calls int
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			calls++
			if u, p, ok := r.BasicAuth(); !ok || u != "id" || p != "sec" {
				t.Errorf("basic auth missing or wrong: %q/%q", u, p)
			}
			writeJSON(w, 200, map[string]any{"access_token": "tok-123", "expires_in": 3600, "token_type": "Bearer"})
		},
	})
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		tok, err := c.AccessToken(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if tok != "tok-123" {
			t.Fatalf("tok=%q", tok)
		}
	}
	if calls != 1 {
		t.Errorf("oauth hit %d times; expected 1 (cache broken)", calls)
	}
}

func TestAccessToken_Expired(t *testing.T) {
	var calls int
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			calls++
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 30, "token_type": "Bearer"})
		},
	})
	ctx := context.Background()
	if _, err := c.AccessToken(ctx); err != nil {
		t.Fatal(err)
	}
	// 手动让过期时间越过 tokenSkew=60s 的缓冲
	c.tokenExp = time.Now().Add(10 * time.Second)
	if _, err := c.AccessToken(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("expected re-fetch after skew expiry; got calls=%d", calls)
	}
}

func TestAccessToken_BadCredentials(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 401, map[string]any{"error": "invalid_client"})
		},
	})
	if _, err := c.AccessToken(context.Background()); err == nil {
		t.Fatal("expected error on 401")
	} else if !strings.Contains(err.Error(), "401") {
		t.Errorf("err = %v; want to contain 401", err)
	}
}

func TestCreateOrder_ApprovalURL(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 3600})
		},
		"POST /v2/checkout/orders": func(w http.ResponseWriter, r *http.Request) {
			// 校验入参
			body, _ := io.ReadAll(r.Body)
			var in map[string]any
			if err := json.Unmarshal(body, &in); err != nil {
				t.Fatal(err)
			}
			if in["intent"] != "CAPTURE" {
				t.Errorf("intent = %v", in["intent"])
			}
			units := in["purchase_units"].([]any)[0].(map[string]any)
			if units["reference_id"] != "GO-1" || units["invoice_id"] != "GO-1" {
				t.Errorf("missing order_no on reference/invoice: %v", units)
			}
			amt := units["amount"].(map[string]any)
			if amt["value"] != "12.34" {
				t.Errorf("amount = %v; want 12.34", amt["value"])
			}

			writeJSON(w, 201, map[string]any{
				"id":     "PP-ORDER-1",
				"status": "CREATED",
				"links": []map[string]any{
					{"rel": "self", "href": "/v2/x", "method": "GET"},
					{"rel": "approve", "href": "https://www.sandbox.paypal.com/checkoutnow?token=PP-ORDER-1", "method": "GET"},
				},
			})
		},
	})
	out, err := c.CreateOrder(context.Background(), CreateOrderReq{
		OrderNo: "GO-1", Description: "t", Amount: 1234, Currency: "USD",
		ReturnURL: "https://x/ok", CancelURL: "https://x/no",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "PP-ORDER-1" || out.Status != "CREATED" {
		t.Errorf("unexpected resp: %+v", out)
	}
	if !strings.Contains(out.ApprovalURL(), "PP-ORDER-1") {
		t.Errorf("approval url = %q", out.ApprovalURL())
	}
}

func TestCaptureOrder(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 3600})
		},
		"POST /v2/checkout/orders/PP-2/capture": func(w http.ResponseWriter, r *http.Request) {
			// 关键：模拟 PayPal 真实响应——invoice_id 位于 captures[] 内部，不是 purchase_unit 外层
			writeJSON(w, 201, map[string]any{
				"id":     "PP-2",
				"status": "COMPLETED",
				"purchase_units": []map[string]any{{
					"reference_id": "default",
					// 注意：这里不设 invoice_id，验证我们从 captures[] 里读
					"payments": map[string]any{
						"captures": []map[string]any{{
							"id":         "CAP-XYZ",
							"status":     "COMPLETED",
							"invoice_id": "GO-CAP-2", // ← 真实 PayPal 返回的位置
							"amount":     map[string]any{"value": "9.00", "currency_code": "USD"},
						}},
					},
				}},
			})
		},
	})
	out, err := c.CaptureOrder(context.Background(), "PP-2")
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "COMPLETED" {
		t.Fatalf("status=%s", out.Status)
	}
	caps := out.PurchaseUnits[0].Payments.Captures
	if len(caps) != 1 || caps[0].ID != "CAP-XYZ" {
		t.Fatalf("captures=%+v", caps)
	}
	if caps[0].InvoiceID != "GO-CAP-2" {
		t.Errorf("invoice_id on capture = %q; want GO-CAP-2", caps[0].InvoiceID)
	}
}

func TestRefundCapture_Partial(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 3600})
		},
		"POST /v2/payments/captures/CAP-1/refund": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var in map[string]any
			_ = json.Unmarshal(body, &in)
			if in["invoice_id"] != "R-1" {
				t.Errorf("invoice_id = %v", in["invoice_id"])
			}
			amt, ok := in["amount"].(map[string]any)
			if !ok || amt["value"] != "5.00" {
				t.Errorf("amount body = %v", in["amount"])
			}
			writeJSON(w, 201, map[string]any{"id": "RF-1", "status": "COMPLETED"})
		},
	})
	out, err := c.RefundCapture(context.Background(), "CAP-1", RefundReq{
		InvoiceID: "R-1", Amount: 500, Currency: "USD", NoteToPayer: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "RF-1" || out.Status != "COMPLETED" {
		t.Errorf("unexpected resp: %+v", out)
	}
}

func TestRefundCapture_Full_OmitsAmount(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 3600})
		},
		"POST /v2/payments/captures/CAP-2/refund": func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var in map[string]any
			_ = json.Unmarshal(body, &in)
			if _, has := in["amount"]; has {
				t.Errorf("full refund should NOT include amount; got %v", in["amount"])
			}
			writeJSON(w, 201, map[string]any{"id": "RF-2", "status": "COMPLETED"})
		},
	})
	if _, err := c.RefundCapture(context.Background(), "CAP-2", RefundReq{InvoiceID: "R-2"}); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWebhook_Success(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 3600})
		},
		"POST /v1/notifications/verify-webhook-signature": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"verification_status": "SUCCESS"})
		},
	})
	ok, err := c.VerifyWebhook(context.Background(), VerifyWebhookReq{
		WebhookID: "WH-1", WebhookEvent: json.RawMessage(`{}`),
	})
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestVerifyWebhook_Failure(t *testing.T) {
	_, c := newStub(t, map[string]http.HandlerFunc{
		"POST /v1/oauth2/token": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"access_token": "t", "expires_in": 3600})
		},
		"POST /v1/notifications/verify-webhook-signature": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{"verification_status": "FAILURE"})
		},
	})
	ok, err := c.VerifyWebhook(context.Background(), VerifyWebhookReq{WebhookID: "WH-1", WebhookEvent: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected ok=false")
	}
}

func TestVerifyWebhook_MissingID(t *testing.T) {
	c := NewClient("sandbox", "id", "sec")
	if _, err := c.VerifyWebhook(context.Background(), VerifyWebhookReq{}); err == nil {
		t.Fatal("expected error when webhook_id empty")
	}
}
