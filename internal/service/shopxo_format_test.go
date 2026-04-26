package service

import (
	"testing"

	"github.com/zhangpanda/goshop/internal/model"
)

func TestPaymentDriverKeyFromPayment_JSON(t *testing.T) {
	p := &model.Payment{Config: `{"payment_key":"alipay_h5"}`}
	k, err := PaymentDriverKeyFromPayment(p)
	if err != nil {
		t.Fatal(err)
	}
	if k != "alipay_h5" {
		t.Fatalf("got %q", k)
	}
}

func TestPaymentDriverKeyFromPayment_PHPClass(t *testing.T) {
	p := &model.Payment{Config: `{"payment":"Weixin"}`}
	k, err := PaymentDriverKeyFromPayment(p)
	if err != nil {
		t.Fatal(err)
	}
	if k != "wechat_jsapi" {
		t.Fatalf("got %q", k)
	}
}

func TestShopXOPluginNameFromDriverKey(t *testing.T) {
	if ShopXOPluginNameFromDriverKey("wechat_jsapi") != "Weixin" {
		t.Fatal()
	}
	if ShopXOPluginNameFromDriverKey("wallet") != "WalletPay" {
		t.Fatal()
	}
}

func TestShopXOOrderListRow_Operate(t *testing.T) {
	o := &model.Order{
		ID:        1,
		Status:    model.OrderStatusPending,
		PayAmount: 1999,
		Items: []model.OrderItem{
			{ID: 10, Title: "x", Image: "/a.jpg", SkuName: "红", Price: 1999, Quantity: 1},
		},
	}
	row := ShopXOOrderListRow(o)
	op, ok := row["operate_data"].(map[string]int)
	if !ok {
		t.Fatal("operate_data type")
	}
	if op["is_pay"] != 1 || op["is_cancel"] != 1 {
		t.Fatalf("%v", op)
	}
}
