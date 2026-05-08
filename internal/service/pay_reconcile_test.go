package service

import (
	"context"
	"testing"

	"github.com/zhangpanda/goshop/internal/app"
)

func TestReconcileWechatPayments_noWxClient(t *testing.T) {
	oldWx := app.Must().WxPay
	app.Must().WxPay = nil
	t.Cleanup(func() { app.Must().WxPay = oldWx })
	pl, od := ReconcileWechatPayments(context.Background(), app.Must())
	if pl != 0 || od != 0 {
		t.Fatalf("expected 0,0 without WxPay, got %d,%d", pl, od)
	}
}
