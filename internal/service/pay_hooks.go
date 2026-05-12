package service

import (
	"log/slog"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/event"
	"github.com/zhangpanda/goshop/internal/model"
)

// RegisterPayEventListeners 注册支付成功后的事件监听器。应在 main 启动时调用一次。
func RegisterPayEventListeners() {
	event.On(event.OrderPaid, func(payload any) {
		e, ok := payload.(event.OrderPaidEvent)
		if !ok {
			return
		}
		AddOrderStatusHistory(e.OrderID, model.OrderStatusPending, model.OrderStatusPaid, "支付成功", "系统")
		NotifyOrderStatus(e.UserID, e.OrderID, e.OrderNo, "paid")
	})

	event.On(event.OrderCompleted, func(payload any) {
		e, ok := payload.(event.OrderCompletedEvent)
		if !ok {
			return
		}
		SettleCommission(e.OrderID)
		_ = OrderRewardPoints(e.UserID, e.OrderID, e.PayAmount)
	})
}

// postPaidHook 统一的"支付成功后"副作用入口。通过事件总线解耦具体处理逻辑。
func postPaidHook(orderID uint, note, creator string) {
	var order model.Order
	if err := app.Must().DB.First(&order, orderID).Error; err != nil {
		slog.Warn("postPaidHook load order", "order_id", orderID, "err", err)
		return
	}
	event.Emit(event.OrderPaid, event.OrderPaidEvent{
		OrderID: orderID,
		UserID:  order.UserID,
		OrderNo: order.OrderNo,
	})
}
