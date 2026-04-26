package model

import "time"

// SearchHistory 搜索历史表
type SearchHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	Keyword   string    `json:"keyword" gorm:"size:128;not null;comment:搜索关键词"`
	CreatedAt time.Time `json:"created_at"`
}

// ScreeningPrice 价格筛选区间表
type ScreeningPrice struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:筛选ID"`
	Name     string `json:"name" gorm:"size:64;not null;comment:筛选名称"`
	MinPrice int64  `json:"min_price" gorm:"not null;comment:最低价(分)"`
	MaxPrice int64  `json:"max_price" gorm:"not null;comment:最高价(分)"`
	Sort     int    `json:"sort" gorm:"default:0;comment:排序"`
}
