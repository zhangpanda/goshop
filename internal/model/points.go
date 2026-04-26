package model

import "time"

// PointsLog 积分变动记录表
type PointsLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	Points    int       `json:"points" gorm:"not null;comment:变动积分(正加负减)"`
	Balance   int       `json:"balance" gorm:"not null;comment:变动后余额"`
	Type      string    `json:"type" gorm:"size:32;not null;comment:类型:order_reward/sign_in/exchange/admin"`
	RefID     uint      `json:"ref_id" gorm:"comment:关联ID"`
	Remark    string    `json:"remark" gorm:"size:128;comment:备注"`
	CreatedAt time.Time `json:"created_at"`
}
