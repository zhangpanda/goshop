package model

import "time"

// Review 商品评价表
type Review struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:评价ID"`
	UserID      uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	OrderID     uint      `json:"order_id" gorm:"index;not null;comment:订单ID"`
	OrderItemID uint      `json:"order_item_id" gorm:"index;not null;comment:订单明细ID"`
	GoodsID     uint      `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	SKUID       uint      `json:"sku_id" gorm:"column:sku_id;comment:SKU ID"`
	Rating      int8      `json:"rating" gorm:"not null;comment:评分1-5星"`
	Content     string    `json:"content" gorm:"type:text;comment:评价内容"`
	Images      string    `json:"images" gorm:"type:text;comment:评价图片JSON数组"`
	Reply       string    `json:"reply" gorm:"type:text;comment:商家回复"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time `json:"created_at"`
}
