package model

import "time"

// Favorite 商品收藏表
type Favorite struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:收藏ID"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex:idx_user_goods;not null;comment:用户ID"`
	GoodsID   uint      `json:"goods_id" gorm:"uniqueIndex:idx_user_goods;not null;comment:商品ID"`
	Goods     *Goods    `json:"goods,omitempty" gorm:"foreignKey:GoodsID"`
	CreatedAt time.Time `json:"created_at"`
}

// BrowseHistory 浏览记录表
type BrowseHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	GoodsID   uint      `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	Goods     *Goods    `json:"goods,omitempty" gorm:"foreignKey:GoodsID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
