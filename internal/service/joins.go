package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

func CreateBrandCategoryRecord(name string, sort int) *model.BrandCategory {
	bc := model.BrandCategory{Name: name, Sort: sort}
	global.DB.Create(&bc)
	return &bc
}

func GetBrandCategoryListRecords() []model.BrandCategory {
	var list []model.BrandCategory
	global.DB.Order("sort DESC").Find(&list)
	return list
}

func SaveGoodsCategoryJoinRecords(goodsID uint, categoryIDs []uint) {
	tx := global.DB.Begin()
	tx.Where("goods_id = ?", goodsID).Delete(&model.GoodsCategoryJoin{})
	for _, cid := range categoryIDs {
		tx.Create(&model.GoodsCategoryJoin{GoodsID: goodsID, CategoryID: cid})
	}
	tx.Commit() //nolint:errcheck // fire-and-forget helper
}

// GoodsSalesCountInc 商品销量增加
func GoodsSalesCountInc(goodsID uint, count int) {
	global.DB.Model(&model.Goods{}).Where("id = ?", goodsID).
		Update("sales_count", global.DB.Raw("sales_count + ?", count))
}

// GoodsAccessCountInc 商品访问量增加
func GoodsAccessCountInc(goodsID uint) {
	global.DB.Model(&model.Goods{}).Where("id = ?", goodsID).
		Update("access_count", global.DB.Raw("access_count + 1"))
}
