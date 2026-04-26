package model

// BrandCategory 品牌分类表
type BrandCategory struct {
	ID   uint   `json:"id" gorm:"primaryKey;comment:分类ID"`
	Name string `json:"name" gorm:"size:64;not null;comment:分类名称"`
	Sort int    `json:"sort" gorm:"default:0;comment:排序"`
}

// BrandCategoryJoin 品牌分类关联表
type BrandCategoryJoin struct {
	ID              uint `json:"id" gorm:"primaryKey;comment:记录ID"`
	BrandID         uint `json:"brand_id" gorm:"uniqueIndex:idx_bcj;not null;comment:品牌ID"`
	BrandCategoryID uint `json:"brand_category_id" gorm:"uniqueIndex:idx_bcj;not null;comment:品牌分类ID"`
}

// GoodsCategoryJoin 商品多分类关联表
type GoodsCategoryJoin struct {
	ID         uint `json:"id" gorm:"primaryKey;comment:记录ID"`
	GoodsID    uint `json:"goods_id" gorm:"uniqueIndex:idx_gcj;not null;comment:商品ID"`
	CategoryID uint `json:"category_id" gorm:"uniqueIndex:idx_gcj;not null;comment:分类ID"`
}
