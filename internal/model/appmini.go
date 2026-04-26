package model

import "time"

// AppMini 小程序配置表
type AppMini struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:小程序ID"`
	Platform  string    `json:"platform" gorm:"size:30;uniqueIndex;not null;comment:平台:weixin/alipay/baidu/toutiao/qq/kuaishou"`
	Title     string    `json:"title" gorm:"size:64;comment:小程序名称"`
	Describe  string    `json:"describe" gorm:"size:255;comment:小程序描述"`
	AppID     string    `json:"app_id" gorm:"size:64;comment:AppID"`
	AppSecret string    `json:"-" gorm:"size:128;comment:AppSecret"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WalletLog 钱包变动日志表
type WalletLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	Amount    int64     `json:"amount" gorm:"not null;comment:变动金额(分)正加负减"`
	Balance   int64     `json:"balance" gorm:"not null;comment:变动后余额(分)"`
	Type      string    `json:"type" gorm:"size:32;comment:类型:recharge/pay/refund/admin"`
	RefID     uint      `json:"ref_id" gorm:"comment:关联ID"`
	Remark    string    `json:"remark" gorm:"size:128;comment:备注"`
	CreatedAt time.Time `json:"created_at"`
}
