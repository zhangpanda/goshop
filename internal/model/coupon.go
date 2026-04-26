package model

import "time"

const (
	CouponTypeFull     int8 = 1 // 满减
	CouponTypeDiscount int8 = 2 // 折扣
	CouponTypeNoLimit  int8 = 3 // 无门槛
)

// Coupon 优惠券模板表
type Coupon struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:优惠券ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:优惠券名称"`
	Type      int8      `json:"type" gorm:"not null;comment:类型:1满减2折扣3无门槛"`
	MinAmount int64     `json:"min_amount" gorm:"default:0;comment:最低消费(分)"`
	Value     int64     `json:"value" gorm:"not null;comment:面值(分)或折扣(85=8.5折)"`
	Total     int       `json:"total" gorm:"not null;comment:发行总量"`
	Received  int       `json:"received" gorm:"default:0;comment:已领取数"`
	PerLimit  int       `json:"per_limit" gorm:"default:1;comment:每人限领"`
	StartTime time.Time `json:"start_time" gorm:"comment:生效时间"`
	EndTime   time.Time `json:"end_time" gorm:"comment:失效时间"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCoupon 用户优惠券表
type UserCoupon struct {
	ID        uint       `json:"id" gorm:"primaryKey;comment:记录ID"`
	UserID    uint       `json:"user_id" gorm:"index;not null;comment:用户ID"`
	CouponID  uint       `json:"coupon_id" gorm:"index;not null;comment:优惠券ID"`
	OrderID   *uint      `json:"order_id" gorm:"index;comment:使用的订单ID"`
	Status    int8       `json:"status" gorm:"default:0;comment:状态:0未使用1已使用2已过期"`
	Coupon    *Coupon    `json:"coupon,omitempty" gorm:"foreignKey:CouponID"`
	UsedAt    *time.Time `json:"used_at" gorm:"comment:使用时间"`
	CreatedAt time.Time  `json:"created_at"`
}
