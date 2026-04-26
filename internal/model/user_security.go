package model

import "time"

// UserPlatform 用户多平台绑定表
type UserPlatform struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	Platform  string    `json:"platform" gorm:"size:30;index;comment:平台:weixin/alipay/baidu/toutiao/qq"`
	OpenID    string    `json:"openid" gorm:"size:64;index;comment:平台OpenID"`
	UnionID   string    `json:"unionid" gorm:"size:64;index;comment:平台UnionID"`
	Token     string    `json:"-" gorm:"size:64;comment:平台Token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VerifyCode 验证码表
type VerifyCode struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	Account   string    `json:"account" gorm:"size:128;index;not null;comment:手机号或邮箱"`
	Code      string    `json:"-" gorm:"size:10;not null;comment:验证码"`
	Type      string    `json:"type" gorm:"size:32;comment:类型:register/login/forget/bind"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0未使用1已使用"`
	ExpireAt  time.Time `json:"expire_at" gorm:"comment:过期时间"`
	CreatedAt time.Time `json:"created_at"`
}
