package service

import (
	"testing"

	"github.com/zhangpanda/goshop/internal/model"
)

func TestPaymentDriverKeyFromPayment_IgnoresShopXOPHPClassField(t *testing.T) {
	p := &model.Payment{Name: "支付宝扫码", Config: `{"payment":"Weixin"}`}
	k, err := PaymentDriverKeyFromPayment(p)
	if err != nil {
		t.Fatal(err)
	}
	if k != "alipay_h5" {
		t.Fatalf("service 层应忽略 PHP payment 类名，按名称推断 alipay_h5，got %q", k)
	}
}

func TestPaymentDriverKeyFromPayment_paymentKeyWins(t *testing.T) {
	p := &model.Payment{Config: `{"payment_key":"alipay_pc","payment":"Weixin"}`}
	k, err := PaymentDriverKeyFromPayment(p)
	if err != nil || k != "alipay_pc" {
		t.Fatalf("got %q err=%v", k, err)
	}
}
