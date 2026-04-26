package model

import "time"

// GroupOrder 拼团订单
type GroupOrder struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	PromotionID uint       `json:"promotion_id" gorm:"index;not null;comment:活动ID"`
	ItemID      uint       `json:"item_id" gorm:"index;not null;comment:活动商品ID"`
	LeaderID    uint       `json:"leader_id" gorm:"not null;comment:团长用户ID"`
	NeedCount   int        `json:"need_count" gorm:"not null;comment:成团人数"`
	JoinCount   int        `json:"join_count" gorm:"default:1;comment:已参团人数"`
	Status      int8       `json:"status" gorm:"default:0;comment:0拼团中1已成团2已失败"`
	ExpireAt    time.Time  `json:"expire_at" gorm:"comment:过期时间"`
	FinishedAt  *time.Time `json:"finished_at,omitempty" gorm:"comment:成团时间"`
	CreatedAt   time.Time  `json:"created_at"`
}

// GroupOrderMember 拼团成员
type GroupOrderMember struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	GroupOrderID uint      `json:"group_order_id" gorm:"index;not null"`
	UserID       uint      `json:"user_id" gorm:"index;not null"`
	OrderID      uint      `json:"order_id" gorm:"comment:关联订单ID"`
	CreatedAt    time.Time `json:"created_at"`
}
