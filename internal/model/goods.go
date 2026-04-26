package model

import "time"

// Goods 商品SPU表
type Goods struct {
	ID           uint       `json:"id" gorm:"primaryKey;comment:商品ID"`
	CategoryID   uint       `json:"category_id" gorm:"index;not null;comment:分类ID"`
	Title        string     `json:"title" gorm:"size:255;not null;comment:商品标题"`
	Subtitle     string     `json:"subtitle" gorm:"size:255;comment:副标题"`
	MainImage    string     `json:"main_image" gorm:"size:255;comment:主图URL"`
	Images       string     `json:"images" gorm:"type:text;comment:多图JSON数组"`
	Detail       string     `json:"detail" gorm:"type:longtext;comment:富文本详情"`
	BrandID      uint       `json:"brand_id" gorm:"index;comment:品牌ID"`
	Status       int8       `json:"status" gorm:"default:0;comment:状态:0下架1上架"`
	Sort         int        `json:"sort" gorm:"default:0;comment:排序(越大越前)"`
	SalesCount   int        `json:"sales_count" gorm:"default:0;comment:销量"`
	AccessCount  int        `json:"access_count" gorm:"default:0;comment:浏览量"`
	GiveIntegral int        `json:"give_integral" gorm:"default:0;comment:赠送积分"`
	SKUs         []GoodsSKU `json:"skus,omitempty" gorm:"foreignKey:GoodsID"`
	Category     *Category  `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// GoodsSKU 商品SKU表
type GoodsSKU struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:SKU ID"`
	GoodsID   uint      `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	Name      string    `json:"name" gorm:"size:128;not null;comment:规格名称(如红色/XL)"`
	Price     int64     `json:"price" gorm:"not null;comment:价格(分)"`
	Stock     int       `json:"stock" gorm:"default:0;comment:库存"`
	Image     string    `json:"image" gorm:"size:255;comment:SKU图片"`
	Specs     string    `json:"specs" gorm:"type:text;comment:规格JSON"`
	Coding    string    `json:"coding" gorm:"size:80;comment:SKU编码"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:1启用0禁用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
