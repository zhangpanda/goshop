package model

import "time"

const (
	AftersaleTypeRefundOnly  int8 = 0 // 仅退款
	AftersaleTypeReturn      int8 = 1 // 退货退款
	AftersaleStatusPending   int8 = 0 // 待确认
	AftersaleStatusShipping  int8 = 1 // 待退货
	AftersaleStatusAudit     int8 = 2 // 待审核
	AftersaleStatusDone      int8 = 3 // 已完成
	AftersaleStatusRefused   int8 = 4 // 已拒绝
	AftersaleStatusCancelled int8 = 5 // 已取消
)

// OrderAftersale 售后申请表
type OrderAftersale struct {
	ID            uint               `json:"id" gorm:"primaryKey;comment:售后ID"`
	OrderID       uint               `json:"order_id" gorm:"index;not null;comment:订单ID"`
	OrderDetailID uint               `json:"order_detail_id" gorm:"index;not null;comment:订单明细ID"`
	UserID        uint               `json:"user_id" gorm:"index;not null;comment:用户ID"`
	GoodsID       uint               `json:"goods_id" gorm:"comment:商品ID"`
	Status        int8               `json:"status" gorm:"default:0;comment:状态:0待确认1待退货2待审核3已完成4已拒绝5已取消"`
	Type          int8               `json:"type" gorm:"default:0;comment:类型:0仅退款1退货退款"`
	Reason        string             `json:"reason" gorm:"size:255;comment:申请原因"`
	Price         int64              `json:"price" gorm:"not null;comment:退款金额(分)"`
	Number        int                `json:"number" gorm:"default:0;comment:退货数量"`
	Msg           string             `json:"msg" gorm:"type:text;comment:补充说明"`
	Images        string             `json:"images" gorm:"type:text;comment:凭证图片JSON"`
	RefuseReason  string             `json:"refuse_reason" gorm:"size:255;comment:拒绝原因"`
	ExpressName   string             `json:"express_name" gorm:"size:64;comment:退货快递公司"`
	ExpressNo     string             `json:"express_no" gorm:"size:64;comment:退货快递单号"`
	Histories     []AftersaleHistory `json:"histories,omitempty" gorm:"foreignKey:AftersaleID"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// AftersaleHistory 售后状态变更历史表
type AftersaleHistory struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	AftersaleID uint      `json:"aftersale_id" gorm:"index;not null;comment:售后ID"`
	Status      int8      `json:"status" gorm:"comment:状态"`
	Msg         string    `json:"msg" gorm:"size:255;comment:说明"`
	Creator     string    `json:"creator" gorm:"size:64;comment:操作人"`
	CreatedAt   time.Time `json:"created_at"`
}
