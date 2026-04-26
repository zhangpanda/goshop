package model

import "time"

// Distributor 分销商
type Distributor struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id" gorm:"uniqueIndex;not null;comment:用户ID"`
	ParentID        uint      `json:"parent_id" gorm:"index;default:0;comment:上级分销商用户ID"`
	Level           int8      `json:"level" gorm:"default:1;comment:等级(1/2/3)"`
	TotalCommission int64     `json:"total_commission" gorm:"default:0;comment:累计佣金(分)"`
	Balance         int64     `json:"balance" gorm:"default:0;comment:可提现余额(分)"`
	OrderCount      int       `json:"order_count" gorm:"default:0;comment:推广订单数"`
	Status          int8      `json:"status" gorm:"default:1;comment:0冻结1正常"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CommissionLog 佣金记录
type CommissionLog struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	DistributorID uint      `json:"distributor_id" gorm:"index;not null"`
	OrderID       uint      `json:"order_id" gorm:"index;comment:来源订单ID"`
	Amount        int64     `json:"amount" gorm:"not null;comment:佣金金额(分)"`
	Type          string    `json:"type" gorm:"size:16;comment:类型:order/withdraw/adjust"`
	Remark        string    `json:"remark" gorm:"size:255"`
	CreatedAt     time.Time `json:"created_at"`
}

// WithdrawRequest 提现申请
type WithdrawRequest struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	DistributorID uint       `json:"distributor_id" gorm:"index;not null"`
	UserID        uint       `json:"user_id" gorm:"index;not null"`
	Amount        int64      `json:"amount" gorm:"not null;comment:提现金额(分)"`
	Status        int8       `json:"status" gorm:"default:0;comment:0待审核1已通过2已拒绝3已打款"`
	AccountType   string     `json:"account_type" gorm:"size:16;comment:alipay/wechat/bank"`
	AccountNo     string     `json:"account_no" gorm:"size:64;comment:收款账号"`
	AccountName   string     `json:"account_name" gorm:"size:32;comment:收款人姓名"`
	RejectReason  string     `json:"reject_reason" gorm:"size:255"`
	AuditAt       *time.Time `json:"audit_at"`
	CreatedAt     time.Time  `json:"created_at"`
}
