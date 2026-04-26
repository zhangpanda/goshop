package model

import "time"

// Shipment 发货/物流表
type Shipment struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:发货ID"`
	OrderID        uint      `json:"order_id" gorm:"uniqueIndex;not null;comment:订单ID"`
	ExpressCompany string    `json:"express_company" gorm:"size:32;not null;comment:快递公司"`
	ExpressNo      string    `json:"express_no" gorm:"index;size:64;not null;comment:快递单号"`
	CreatedAt      time.Time `json:"created_at"`
}
