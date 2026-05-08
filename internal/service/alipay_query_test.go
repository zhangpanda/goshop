package service

import (
	"context"
	"testing"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
)

func TestAlipayQueryTrade_notConfigured(t *testing.T) {
	old := app.Must().Cfg
	app.Must().Cfg = &config.Config{}
	t.Cleanup(func() { app.Must().Cfg = old })
	_, _, err := AlipayQueryTrade(context.Background(), "TEST_OUT_NO")
	if err == nil {
		t.Fatal("expected error when alipay not configured")
	}
}

func TestReconcileAlipayPayments_notConfigured(t *testing.T) {
	old := app.Must().Cfg
	app.Must().Cfg = &config.Config{}
	t.Cleanup(func() { app.Must().Cfg = old })
	pl, od := ReconcileAlipayPayments(context.Background(), app.Must())
	if pl != 0 || od != 0 {
		t.Fatalf("expected 0,0, got %d,%d", pl, od)
	}
}

func TestAlipayTradePaid(t *testing.T) {
	if !AlipayTradePaid("TRADE_SUCCESS") || !AlipayTradePaid("TRADE_FINISHED") {
		t.Fatal("expected paid")
	}
	if AlipayTradePaid("WAIT_BUYER_PAY") || AlipayTradePaid("") {
		t.Fatal("expected not paid")
	}
}
