package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

type DiyApiParams struct {
	CategoryID uint   `form:"category_id" json:"category_id"`
	BrandID    uint   `form:"brand_id" json:"brand_id"`
	OrderBy    string `form:"order_by" json:"order_by"`
	Limit      int    `form:"limit" json:"limit"`
}

func DiyApiGoodsAutoData(p *DiyApiParams) ([]model.Goods, error) {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	db := app.Must().DB.Where("status = 1")
	if p.CategoryID > 0 {
		db = db.Where("category_id = ?", p.CategoryID)
	}
	if p.BrandID > 0 {
		db = db.Where("brand_id = ?", p.BrandID)
	}
	order := "sort DESC, id DESC"
	switch p.OrderBy {
	case "sales":
		order = "sales_count DESC"
	case "new":
		order = "id DESC"
	}
	var list []model.Goods
	err := db.Preload("SKUs").Order(order).Limit(p.Limit).Find(&list).Error
	return list, err
}

func DiyApiArticleAutoData(categoryID uint, limit int) ([]model.Article, error) {
	if limit <= 0 {
		limit = 10
	}
	db := app.Must().DB.Where("status = 1")
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	var list []model.Article
	return list, db.Order("sort DESC, id DESC").Limit(limit).Find(&list).Error
}

func DiyApiBrandAutoData(limit int) ([]model.Brand, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []model.Brand
	return list, app.Must().DB.Where("status = 1").Order("sort DESC").Limit(limit).Find(&list).Error
}

func DiyApiGoodsFavorAutoData(userID uint, limit int) ([]model.Favorite, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.Favorite
	return list, app.Must().DB.Where("user_id = ?", userID).Preload("Goods").Order("id DESC").Limit(limit).Find(&list).Error
}

func DiyApiGoodsBrowseAutoData(userID uint, limit int) ([]model.BrowseHistory, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.BrowseHistory
	return list, app.Must().DB.Where("user_id = ?", userID).Preload("Goods").Order("updated_at DESC").Limit(limit).Find(&list).Error
}
