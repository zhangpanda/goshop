package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

func ToggleFavorite(userID, goodsID uint) (bool, error) {
	var fav model.Favorite
	app.Must().DB.Where("user_id = ? AND goods_id = ?", userID, goodsID).Find(&fav)
	if fav.ID > 0 {
		app.Must().DB.Delete(&fav)
		return false, nil
	}
	fav = model.Favorite{UserID: userID, GoodsID: goodsID}
	return true, app.Must().DB.Create(&fav).Error
}

func GetFavorites(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	var total int64
	app.Must().DB.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&total)
	var list []model.Favorite
	err := app.Must().DB.Where("user_id = ?", userID).Preload("Goods").
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func AddBrowseHistory(userID, goodsID uint) {
	var h model.BrowseHistory
	app.Must().DB.Where("user_id = ? AND goods_id = ?", userID, goodsID).Find(&h)
	if h.ID > 0 {
		app.Must().DB.Save(&h)
		return
	}
	app.Must().DB.Create(&model.BrowseHistory{UserID: userID, GoodsID: goodsID})
}

func GetBrowseHistory(userID uint, page, pageSize int) ([]model.BrowseHistory, int64, error) {
	var total int64
	app.Must().DB.Model(&model.BrowseHistory{}).Where("user_id = ?", userID).Count(&total)
	var list []model.BrowseHistory
	err := app.Must().DB.Where("user_id = ?", userID).Preload("Goods").
		Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func ClearBrowseHistory(userID uint) error {
	return app.Must().DB.Where("user_id = ?", userID).Delete(&model.BrowseHistory{}).Error
}
