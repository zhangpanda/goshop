package service

import (
	"testing"
	"time"

	"github.com/zhangpanda/goshop/internal/model"
)

func TestOrderOperateButtons_Pending(t *testing.T) {
	order := &model.Order{Status: model.OrderStatusPending}
	op := OrderOperateButtons(order)
	if !op.CanCancel {
		t.Error("Pending order should be cancellable")
	}
	if !op.CanPay {
		t.Error("Pending order should be payable")
	}
	if op.CanReceive {
		t.Error("Pending order should not be receivable")
	}
}

func TestOrderOperateButtons_Shipped(t *testing.T) {
	order := &model.Order{Status: model.OrderStatusShipped}
	op := OrderOperateButtons(order)
	if !op.CanReceive {
		t.Error("Shipped order should be receivable")
	}
	if !op.CanAftersale {
		t.Error("Shipped order should allow aftersale")
	}
	if op.CanPay {
		t.Error("Shipped order should not be payable")
	}
}

func TestOrderOperateButtons_Completed(t *testing.T) {
	order := &model.Order{Status: model.OrderStatusCompleted}
	op := OrderOperateButtons(order)
	if !op.CanReview {
		t.Error("Completed order should be reviewable")
	}
	if !op.CanDelete {
		t.Error("Completed order should be deletable")
	}
	if op.CanCancel {
		t.Error("Completed order should not be cancellable")
	}
}

func TestOrderOperateButtons_Cancelled(t *testing.T) {
	order := &model.Order{Status: model.OrderStatusCancelled}
	op := OrderOperateButtons(order)
	if !op.CanDelete {
		t.Error("Cancelled order should be deletable")
	}
	if op.CanPay || op.CanCancel || op.CanReceive {
		t.Error("Cancelled order should have no active operations")
	}
}

func TestOrderStepData_NewOrder(t *testing.T) {
	order := &model.Order{Status: model.OrderStatusPending}
	order.CreatedAt = time.Now()
	steps := OrderStepData(order)
	if len(steps) != 5 {
		t.Fatalf("steps count = %d, want 5", len(steps))
	}
	if steps[0].Status != 2 {
		t.Error("step 0 (提交订单) should be completed")
	}
	if steps[1].Status != 1 {
		t.Error("step 1 (付款) should be current")
	}
	if steps[2].Status != 0 {
		t.Error("step 2 (发货) should be pending")
	}
}

func TestOrderStepData_PaidOrder(t *testing.T) {
	now := time.Now()
	order := &model.Order{Status: model.OrderStatusPaid, PaidAt: &now}
	order.CreatedAt = now
	steps := OrderStepData(order)
	if steps[1].Status != 2 {
		t.Error("step 1 (付款) should be completed")
	}
	if steps[2].Status != 1 {
		t.Error("step 2 (发货) should be current")
	}
}

func TestOrderStepData_CompletedOrder(t *testing.T) {
	now := time.Now()
	order := &model.Order{Status: model.OrderStatusCompleted, PaidAt: &now, ShippedAt: &now, CompletedAt: &now}
	order.CreatedAt = now
	steps := OrderStepData(order)
	for i, s := range steps {
		if s.Status != 2 {
			t.Errorf("step %d (%s) should be completed, got status=%d", i, s.Name, s.Status)
		}
	}
}

func TestCalcCouponDiscount_FullReduction(t *testing.T) {
	coupon := &model.Coupon{Type: model.CouponTypeFull, MinAmount: 10000, Value: 2000}
	d, err := CalcCouponDiscount(coupon, 15000)
	if err != nil {
		t.Fatal(err)
	}
	if d != 2000 {
		t.Errorf("discount = %d, want 2000", d)
	}
}

func TestCalcCouponDiscount_BelowMin(t *testing.T) {
	coupon := &model.Coupon{Type: model.CouponTypeFull, MinAmount: 10000, Value: 2000}
	_, err := CalcCouponDiscount(coupon, 5000)
	if err == nil {
		t.Error("expected error for below min amount")
	}
}

func TestCalcCouponDiscount_Discount85(t *testing.T) {
	coupon := &model.Coupon{Type: model.CouponTypeDiscount, MinAmount: 0, Value: 85}
	d, err := CalcCouponDiscount(coupon, 10000)
	if err != nil {
		t.Fatal(err)
	}
	// 8.5折 = 10000 - 10000*85/100 = 10000 - 8500 = 1500
	if d != 1500 {
		t.Errorf("discount = %d, want 1500", d)
	}
}

func TestCalcCouponDiscount_ValueExceedsOrder(t *testing.T) {
	coupon := &model.Coupon{Type: model.CouponTypeNoLimit, MinAmount: 0, Value: 5000}
	d, err := CalcCouponDiscount(coupon, 3000)
	if err != nil {
		t.Fatal(err)
	}
	// 至少付1分
	if d != 2999 {
		t.Errorf("discount = %d, want 2999 (order-1)", d)
	}
}

func TestGenerateOrderNo_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		no := generateOrderNo()
		if seen[no] {
			t.Fatalf("duplicate order no: %s", no)
		}
		seen[no] = true
	}
}

func TestGenerateOrderNo_Length(t *testing.T) {
	no := generateOrderNo()
	// 14 (timestamp) + 8 (hex) = 22
	if len(no) != 22 {
		t.Errorf("order no length = %d, want 22, got %q", len(no), no)
	}
}
