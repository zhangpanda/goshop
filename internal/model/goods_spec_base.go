package model

import "time"

// GoodsSpecBase 商品多维规格组合表
type GoodsSpecBase struct {
	ID            uint    `json:"id" gorm:"primaryKey;comment:规格组合ID"`
	GoodsID       uint    `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	Price         int64   `json:"price" gorm:"not null;comment:销售价(分)"`
	OriginalPrice int64   `json:"original_price" gorm:"default:0;comment:原价(分)"`
	Inventory     int     `json:"inventory" gorm:"default:0;comment:库存"`
	BuyMinNumber  int     `json:"buy_min_number" gorm:"default:0;comment:最低购买数"`
	BuyMaxNumber  int     `json:"buy_max_number" gorm:"default:0;comment:最高购买数"`
	Weight        float64 `json:"weight" gorm:"default:0;comment:重量(kg)"`
	Volume        float64 `json:"volume" gorm:"default:0;comment:体积(m³)"`
	Coding        string  `json:"coding" gorm:"size:80;comment:商品编码"`
	Barcode       string  `json:"barcode" gorm:"size:80;comment:条形码"`
	SpecValues    string  `json:"spec_values" gorm:"size:255;comment:规格值组合(如红色,256GB)"`
	CreatedAt     time.Time `json:"created_at"`
}

// GoodsPhoto 商品相册表
type GoodsPhoto struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:相册ID"`
	GoodsID   uint      `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	Image     string    `json:"image" gorm:"size:255;not null;comment:图片URL"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	IsShow    int8      `json:"is_show" gorm:"default:1;comment:是否显示"`
	CreatedAt time.Time `json:"created_at"`
}
