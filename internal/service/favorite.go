package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

func ToggleFavorite(userID, goodsID uint) (bool, error) {
	var fav model.Favorite
	global.DB.Where("user_id = ? AND goods_id = ?", userID, goodsID).Find(&fav)
	if fav.ID > 0 {
		global.DB.Delete(&fav)
		return false, nil
	}
	fav = model.Favorite{UserID: userID, GoodsID: goodsID}
	return true, global.DB.Create(&fav).Error
}

func GetFavorites(userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	var total int64
	global.DB.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&total)
	var list []model.Favorite
	err := global.DB.Where("user_id = ?", userID).Preload("Goods").
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func AddBrowseHistory(userID, goodsID uint) {
	var h model.BrowseHistory
	global.DB.Where("user_id = ? AND goods_id = ?", userID, goodsID).Find(&h)
	if h.ID > 0 {
		global.DB.Save(&h)
		return
	}
	global.DB.Create(&model.BrowseHistory{UserID: userID, GoodsID: goodsID})
}

func GetBrowseHistory(userID uint, page, pageSize int) ([]model.BrowseHistory, int64, error) {
	var total int64
	global.DB.Model(&model.BrowseHistory{}).Where("user_id = ?", userID).Count(&total)
	var list []model.BrowseHistory
	err := global.DB.Where("user_id = ?", userID).Preload("Goods").
		Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func ClearBrowseHistory(userID uint) error {
	return global.DB.Where("user_id = ?", userID).Delete(&model.BrowseHistory{}).Error
}
