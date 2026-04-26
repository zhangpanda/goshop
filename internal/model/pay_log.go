package model

import "time"

// PayLog 支付日志表
type PayLog struct {
	ID         uint       `json:"id" gorm:"primaryKey;comment:支付日志ID"`
	PayNo      string     `json:"pay_no" gorm:"uniqueIndex;size:64;not null;comment:支付单号"`
	OrderIDs   string     `json:"order_ids" gorm:"size:255;comment:合并支付的订单ID(逗号分隔)"`
	UserID     uint       `json:"user_id" gorm:"index;not null;comment:用户ID"`
	PaymentID  uint       `json:"payment_id" gorm:"index;comment:支付方式ID"`
	TotalPrice int64      `json:"total_price" gorm:"not null;comment:支付金额(分)"`
	TradeNo    string     `json:"trade_no" gorm:"size:128;comment:第三方交易号"`
	Status     int8       `json:"status" gorm:"default:0;comment:状态:0待支付1已支付2已关闭"`
	ClientType string     `json:"client_type" gorm:"size:16;comment:客户端:pc/h5/weixin/alipay"`
	PaidAt     *time.Time `json:"paid_at" gorm:"comment:支付时间"`
	ClosedAt   *time.Time `json:"closed_at" gorm:"comment:关闭时间"`
	CreatedAt  time.Time  `json:"created_at"`
}

// PayRequestLog 支付请求日志表
type PayRequestLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	PayLogID  uint      `json:"pay_log_id" gorm:"index;comment:支付日志ID"`
	Request   string    `json:"request" gorm:"type:text;comment:请求内容"`
	Response  string    `json:"response" gorm:"type:text;comment:响应内容"`
	Business  string    `json:"business" gorm:"size:32;comment:业务类型:pay/notify/refund"`
	CreatedAt time.Time `json:"created_at"`
}

// RefundLog 退款日志表
type RefundLog struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:退款ID"`
	OrderID     uint      `json:"order_id" gorm:"index;not null;comment:订单ID"`
	PayLogID    uint      `json:"pay_log_id" gorm:"index;comment:支付日志ID"`
	UserID      uint      `json:"user_id" gorm:"index;comment:用户ID"`
	RefundNo    string    `json:"refund_no" gorm:"uniqueIndex;size:64;comment:退款单号"`
	TradeNo     string    `json:"trade_no" gorm:"size:128;comment:第三方交易号"`
	RefundPrice int64     `json:"refund_price" gorm:"not null;comment:退款金额(分)"`
	Reason      string    `json:"reason" gorm:"size:255;comment:退款原因"`
	Status      int8      `json:"status" gorm:"default:0;comment:状态:0处理中1成功2失败"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
