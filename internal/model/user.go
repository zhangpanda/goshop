package model

import "time"

type User struct {
	ID              uint      `json:"id" gorm:"primaryKey;comment:用户ID"`
	Username        string    `json:"username" gorm:"uniqueIndex;size:64;comment:用户名"`
	Password        string    `json:"-" gorm:"size:128;comment:密码(bcrypt)"`
	Nickname        string    `json:"nickname" gorm:"size:64;comment:昵称"`
	Phone           string    `json:"phone" gorm:"index;size:20;comment:手机号"`
	Avatar          string    `json:"avatar" gorm:"size:255;comment:头像URL"`
	OpenID          string    `json:"-" gorm:"index;size:64;comment:微信OpenID"`
	UnionID         string    `json:"-" gorm:"index;size:64;comment:微信UnionID"`
	Points          int       `json:"points" gorm:"default:0;comment:积分余额"`
	WalletBalance   int64     `json:"wallet_balance" gorm:"default:0;comment:钱包余额(分)"`
	LockingIntegral int       `json:"locking_integral" gorm:"default:0;comment:锁定积分(待赠送)"`
	Status          int8      `json:"status" gorm:"default:1;comment:状态:0禁用1正常"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
