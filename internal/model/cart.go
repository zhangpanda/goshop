package model

import "time"

// Cart 购物车表
type Cart struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:购物车ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	GoodsID   uint      `json:"goods_id" gorm:"not null;comment:商品ID"`
	SKUID     uint      `json:"sku_id" gorm:"column:sku_id;not null;comment:SKU ID"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1;comment:数量"`
	Selected  bool      `json:"selected" gorm:"default:true;comment:是否选中"`
	Goods     *Goods    `json:"goods,omitempty" gorm:"foreignKey:GoodsID"`
	SKU       *GoodsSKU `json:"sku,omitempty" gorm:"foreignKey:SKUID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
