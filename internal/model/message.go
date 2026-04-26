package model

import "time"

// Message 站内消息表
type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:消息ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	Title     string    `json:"title" gorm:"size:128;not null;comment:消息标题"`
	Content   string    `json:"content" gorm:"type:text;comment:消息内容"`
	Type      string    `json:"type" gorm:"size:32;comment:消息类型:order/system/promotion"`
	RefID     uint      `json:"ref_id" gorm:"comment:关联ID"`
	IsRead    bool      `json:"is_read" gorm:"default:false;comment:是否已读"`
	CreatedAt time.Time `json:"created_at"`
}
