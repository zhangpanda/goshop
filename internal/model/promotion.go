package model

import "time"

// Promotion 营销活动表（限时促销/秒杀/拼团）
type Promotion struct {
	ID        uint            `json:"id" gorm:"primaryKey;comment:活动ID"`
	Name      string          `json:"name" gorm:"size:64;not null;comment:活动名称"`
	Type      string          `json:"type" gorm:"size:16;default:promo;comment:类型:promo/seckill/group"`
	StartTime time.Time       `json:"start_time" gorm:"comment:开始时间"`
	EndTime   time.Time       `json:"end_time" gorm:"comment:结束时间"`
	Status    int8            `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	GroupSize int             `json:"group_size,omitempty" gorm:"default:0;comment:拼团人数(拼团专用)"`
	GroupTime int             `json:"group_time,omitempty" gorm:"default:0;comment:拼团时限(分钟)"`
	Items     []PromotionItem `json:"items,omitempty" gorm:"foreignKey:PromotionID"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// PromotionItem 活动商品表
type PromotionItem struct {
	ID          uint  `json:"id" gorm:"primaryKey;comment:记录ID"`
	PromotionID uint  `json:"promotion_id" gorm:"index;not null;comment:活动ID"`
	GoodsID     uint  `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	SKUID       uint  `json:"sku_id" gorm:"column:sku_id;index;not null;comment:SKU ID"`
	PromoPrice  int64 `json:"promo_price" gorm:"not null;comment:促销价(分)"`
	PromoStock  int   `json:"promo_stock" gorm:"not null;comment:促销库存"`
	Sold        int   `json:"sold" gorm:"default:0;comment:已售数量"`
	PerLimit    int   `json:"per_limit" gorm:"default:0;comment:每人限购(0不限)"`
}

func (p *Promotion) IsActive() bool {
	now := time.Now()
	return p.Status == 1 && now.After(p.StartTime) && now.Before(p.EndTime)
}
