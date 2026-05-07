package service

import (
	"context"
	"testing"

	"github.com/zhangpanda/goshop/global"
)

func TestReconcileWechatPayments_noWxClient(t *testing.T) {
	global.WxPay = nil
	pl, od := ReconcileWechatPayments(context.Background())
	if pl != 0 || od != 0 {
		t.Fatalf("expected 0,0 without WxPay, got %d,%d", pl, od)
	}
}
