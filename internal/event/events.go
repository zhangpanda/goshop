package event

// 事件名常量
const (
	OrderPaid      = "order.paid"      // payload: OrderPaidEvent
	OrderCompleted = "order.completed" // payload: OrderCompletedEvent
	OrderRefunded  = "order.refunded"  // payload: OrderRefundedEvent
)

type OrderPaidEvent struct {
	OrderID uint
	UserID  uint
	OrderNo string
}

type OrderCompletedEvent struct {
	OrderID   uint
	UserID    uint
	PayAmount int64
}

type OrderRefundedEvent struct {
	OrderID uint
	UserID  uint
}
