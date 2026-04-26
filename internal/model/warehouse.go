package model

import "time"

// Warehouse 仓库表
type Warehouse struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:仓库ID"`
	Name         string    `json:"name" gorm:"size:64;not null;comment:仓库名称"`
	Alias        string    `json:"alias" gorm:"size:64;comment:仓库别名"`
	Level        int       `json:"level" gorm:"default:0;comment:权重"`
	IsEnable     int8      `json:"is_enable" gorm:"default:1;comment:是否启用"`
	ContactsName string    `json:"contacts_name" gorm:"size:64;comment:联系人"`
	ContactsTel  string    `json:"contacts_tel" gorm:"size:20;comment:联系电话"`
	Province     string    `json:"province" gorm:"size:32;comment:省"`
	City         string    `json:"city" gorm:"size:32;comment:市"`
	County       string    `json:"county" gorm:"size:32;comment:区/县"`
	Address      string    `json:"address" gorm:"size:255;comment:详细地址"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WarehouseGoods 仓库商品关联表
type WarehouseGoods struct {
	ID          uint `json:"id" gorm:"primaryKey;comment:记录ID"`
	WarehouseID uint `json:"warehouse_id" gorm:"uniqueIndex:idx_wg;not null;comment:仓库ID"`
	GoodsID     uint `json:"goods_id" gorm:"uniqueIndex:idx_wg;not null;comment:商品ID"`
	Inventory   int  `json:"inventory" gorm:"default:0;comment:库存"`
	IsEnable    int8 `json:"is_enable" gorm:"default:1;comment:是否启用"`
}

// WarehouseGoodsSpec 仓库商品规格库存表
type WarehouseGoodsSpec struct {
	ID          uint   `json:"id" gorm:"primaryKey;comment:记录ID"`
	WarehouseID uint   `json:"warehouse_id" gorm:"index;not null;comment:仓库ID"`
	GoodsID     uint   `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	SKUID       uint   `json:"sku_id" gorm:"column:sku_id;index;not null;comment:SKU ID"`
	Inventory   int    `json:"inventory" gorm:"default:0;comment:库存"`
	SpecValues  string `json:"spec_values" gorm:"size:255;comment:规格值组合"`
}
