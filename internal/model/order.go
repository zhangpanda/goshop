package model

import "time"

const (
	OrderStatusPending   int8 = 0 // 待付款
	OrderStatusPaid      int8 = 1 // 待发货
	OrderStatusShipped   int8 = 2 // 待收货
	OrderStatusCompleted int8 = 3 // 已完成
	OrderStatusCancelled int8 = 4 // 已取消
	OrderStatusRefunded  int8 = 5 // 已退款
	OrderStatusBooking   int8 = 6 // 预约待确认
)

const (
	OrderModelExpress int8 = 0 // 快递
	OrderModelLocal   int8 = 1 // 同城
	OrderModelPickup  int8 = 2 // 自提
	OrderModelVirtual int8 = 3 // 虚拟
)

// Order 订单主表
type Order struct {
	ID              uint        `json:"id" gorm:"primaryKey;comment:订单ID"`
	OrderNo         string      `json:"order_no" gorm:"uniqueIndex;size:32;not null;comment:订单编号"`
	UserID          uint        `json:"user_id" gorm:"index:idx_order_user_status;not null;comment:用户ID"`
	TotalAmount     int64       `json:"total_amount" gorm:"not null;comment:总金额(分)"`
	PayAmount       int64       `json:"pay_amount" gorm:"not null;comment:实付金额(分)"`
	Status          int8        `json:"status" gorm:"default:0;index:idx_order_user_status;index:idx_order_status_created;comment:状态:0待付款1待发货2待收货3已完成4已取消5已退款"`
	OrderModel      int8        `json:"order_model" gorm:"default:0;comment:订单模式:0快递1同城2自提3虚拟"`
	ExtractionCode  string      `json:"extraction_code" gorm:"size:6;comment:自提码"`
	FictitiousValue string      `json:"fictitious_value" gorm:"type:text;comment:虚拟商品信息"`
	Remark          string      `json:"remark" gorm:"size:255;comment:订单备注"`
	PaymentID       uint        `json:"payment_id" gorm:"index;default:0;comment:支付方式ID(用户选用)"`
	Address         string      `json:"address" gorm:"type:text;comment:收货地址快照JSON"`
	Items           []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	PaidAt          *time.Time  `json:"paid_at" gorm:"comment:支付时间"`
	TransactionID   string      `json:"transaction_id" gorm:"size:64;comment:第三方支付流水号"`
	ShippedAt       *time.Time  `json:"shipped_at" gorm:"comment:发货时间"`
	CompletedAt     *time.Time  `json:"completed_at" gorm:"comment:完成时间"`
	CreatedAt       time.Time   `json:"created_at" gorm:"index:idx_order_status_created"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// OrderItem 订单商品明细表
type OrderItem struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:明细ID"`
	OrderID   uint      `json:"order_id" gorm:"index;not null;comment:订单ID"`
	GoodsID   uint      `json:"goods_id" gorm:"not null;comment:商品ID"`
	SKUID     uint      `json:"sku_id" gorm:"column:sku_id;not null;comment:SKU ID"`
	Title     string    `json:"title" gorm:"size:255;comment:商品标题快照"`
	Image     string    `json:"image" gorm:"size:255;comment:商品图片快照"`
	SkuName   string    `json:"sku_name" gorm:"size:128;comment:SKU名称快照"`
	Price     int64     `json:"price" gorm:"not null;comment:下单时价格(分)"`
	Quantity  int       `json:"quantity" gorm:"not null;comment:购买数量"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderStatusHistory 订单状态变更历史表
type OrderStatusHistory struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	OrderID        uint      `json:"order_id" gorm:"index;not null;comment:订单ID"`
	OriginalStatus int8      `json:"original_status" gorm:"comment:原状态"`
	NewStatus      int8      `json:"new_status" gorm:"comment:新状态"`
	Msg            string    `json:"msg" gorm:"size:255;comment:变更说明"`
	Creator        string    `json:"creator" gorm:"size:64;comment:操作人"`
	CreatedAt      time.Time `json:"created_at"`
}
