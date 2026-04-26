package model

import "time"

// Express 快递公司表
type Express struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:快递ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:快递公司名称"`
	Code      string    `json:"code" gorm:"size:32;uniqueIndex;not null;comment:快递编码(如sf/yt/yd)"`
	Icon      string    `json:"icon" gorm:"size:255;comment:快递公司图标"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// InventoryLog 库存变动日志表
type InventoryLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	OrderID   uint      `json:"order_id" gorm:"index;comment:订单ID"`
	GoodsID   uint      `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	SKUID     uint      `json:"sku_id" gorm:"column:sku_id;index;comment:SKU ID"`
	Quantity  int       `json:"quantity" gorm:"not null;comment:数量(正入库负出库)"`
	Type      string    `json:"type" gorm:"size:32;comment:类型:order/cancel/refund/admin"`
	Remark    string    `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt time.Time `json:"created_at"`
}

// GoodsGiveIntegralLog 商品赠送积分日志表
type GoodsGiveIntegralLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	OrderID   uint      `json:"order_id" gorm:"index;comment:订单ID"`
	UserID    uint      `json:"user_id" gorm:"index;not null;comment:用户ID"`
	GoodsID   uint      `json:"goods_id" gorm:"index;comment:商品ID"`
	Integral  int       `json:"integral" gorm:"not null;comment:赠送积分"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0待赠送1已赠送"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
