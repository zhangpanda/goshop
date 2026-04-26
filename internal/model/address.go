package model

import "time"

// Address 收货地址表
type Address struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:地址ID"`
	UserID      uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	Name        string    `json:"name" gorm:"size:32;not null;comment:收货人姓名"`
	Phone       string    `json:"phone" gorm:"size:20;not null;comment:收货人手机号"`
	Province    string    `json:"province" gorm:"size:32;comment:省"`
	City        string    `json:"city" gorm:"size:32;comment:市"`
	District    string    `json:"district" gorm:"size:32;comment:区/县"`
	Detail      string    `json:"detail" gorm:"size:255;not null;comment:详细地址"`
	Lng         float64   `json:"lng" gorm:"default:0;comment:经度"`
	Lat         float64   `json:"lat" gorm:"default:0;comment:纬度"`
	IDCard      string    `json:"id_card" gorm:"size:20;comment:身份证号"`
	IDCardFront string    `json:"id_card_front" gorm:"size:255;comment:身份证正面"`
	IDCardBack  string    `json:"id_card_back" gorm:"size:255;comment:身份证反面"`
	IsDefault   bool      `json:"is_default" gorm:"default:false;comment:是否默认地址"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
