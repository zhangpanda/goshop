package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type SpecBaseReq struct {
	Price         int64   `json:"price" binding:"required"`
	OriginalPrice int64   `json:"original_price"`
	Inventory     int     `json:"inventory"`
	BuyMinNumber  int     `json:"buy_min_number"`
	BuyMaxNumber  int     `json:"buy_max_number"`
	Weight        float64 `json:"weight"`
	Volume        float64 `json:"volume"`
	Coding        string  `json:"coding"`
	Barcode       string  `json:"barcode"`
	SpecValues    string  `json:"spec_values" binding:"required"`
}

func SaveGoodsSpecBase(goodsID uint, specs []SpecBaseReq) error {
	tx := global.DB.Begin()
	tx.Where("goods_id = ?", goodsID).Delete(&model.GoodsSpecBase{})
	for _, s := range specs {
		tx.Create(&model.GoodsSpecBase{
			GoodsID: goodsID, Price: s.Price, OriginalPrice: s.OriginalPrice,
			Inventory: s.Inventory, BuyMinNumber: s.BuyMinNumber, BuyMaxNumber: s.BuyMaxNumber,
			Weight: s.Weight, Volume: s.Volume, Coding: s.Coding, Barcode: s.Barcode,
			SpecValues: s.SpecValues,
		})
	}
	tx.Commit()
	return nil
}

func GetGoodsSpecBase(goodsID uint) ([]model.GoodsSpecBase, error) {
	var list []model.GoodsSpecBase
	return list, global.DB.Where("goods_id = ?", goodsID).Find(&list).Error
}

// SaveGoodsPhotos 保存商品相册
func SaveGoodsPhotos(goodsID uint, images []string) error {
	tx := global.DB.Begin()
	tx.Where("goods_id = ?", goodsID).Delete(&model.GoodsPhoto{})
	for i, img := range images {
		tx.Create(&model.GoodsPhoto{GoodsID: goodsID, Image: img, Sort: i, IsShow: 1})
	}
	tx.Commit()
	return nil
}

func GetGoodsPhotos(goodsID uint) ([]model.GoodsPhoto, error) {
	var list []model.GoodsPhoto
	return list, global.DB.Where("goods_id = ? AND is_show = 1", goodsID).Order("sort").Find(&list).Error
}
