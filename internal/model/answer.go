package model

import "time"

// Answer 问答留言表
type Answer struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:问答ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	GoodsID   uint      `json:"goods_id" gorm:"index;comment:商品ID"`
	Title     string    `json:"title" gorm:"size:255;not null;comment:问题标题"`
	Content   string    `json:"content" gorm:"type:text;comment:问题内容"`
	Reply     string    `json:"reply" gorm:"type:text;comment:回复内容"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0待回复1已回复"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
