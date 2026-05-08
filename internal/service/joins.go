package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

func CreateBrandCategoryRecord(name string, sort int) *model.BrandCategory {
	bc := model.BrandCategory{Name: name, Sort: sort}
	app.Must().DB.Create(&bc)
	return &bc
}

func GetBrandCategoryListRecords() []model.BrandCategory {
	var list []model.BrandCategory
	app.Must().DB.Order("sort DESC").Find(&list)
	return list
}

func SaveGoodsCategoryJoinRecords(goodsID uint, categoryIDs []uint) {
	_ = RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Where("goods_id = ?", goodsID).Delete(&model.GoodsCategoryJoin{}).Error; err != nil {
			return err
		}
		for _, cid := range categoryIDs {
			if err := tx.Create(&model.GoodsCategoryJoin{GoodsID: goodsID, CategoryID: cid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GoodsSalesCountInc 商品销量增加
func GoodsSalesCountInc(goodsID uint, count int) {
	app.Must().DB.Model(&model.Goods{}).Where("id = ?", goodsID).
		Update("sales_count", gorm.Expr("sales_count + ?", count))
}

// GoodsAccessCountInc 商品访问量增加
func GoodsAccessCountInc(goodsID uint) {
	app.Must().DB.Model(&model.Goods{}).Where("id = ?", goodsID).
		Update("access_count", gorm.Expr("access_count + 1"))
}
