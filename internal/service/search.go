package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

func AddSearchHistory(userID uint, keyword string) {
	if keyword == "" {
		return
	}
	var h model.SearchHistory
	global.DB.Where("user_id = ? AND keyword = ?", userID, keyword).Find(&h)
	if h.ID > 0 {
		global.DB.Save(&h)
		return
	}
	global.DB.Create(&model.SearchHistory{UserID: userID, Keyword: keyword})
}

func GetSearchHistory(userID uint) ([]model.SearchHistory, error) {
	var list []model.SearchHistory
	return list, global.DB.Where("user_id = ?", userID).Order("id DESC").Limit(20).Find(&list).Error
}

func ClearSearchHistory(userID uint) error {
	return global.DB.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}

func GetHotKeywords(limit int) ([]string, error) {
	var results []struct{ Keyword string }
	err := global.DB.Model(&model.SearchHistory{}).Select("keyword, COUNT(*) as cnt").
		Group("keyword").Order("cnt DESC").Limit(limit).Find(&results).Error
	var keywords []string
	for _, r := range results {
		keywords = append(keywords, r.Keyword)
	}
	return keywords, err
}

func GetScreeningPrices() ([]model.ScreeningPrice, error) {
	var list []model.ScreeningPrice
	return list, global.DB.Order("sort").Find(&list).Error
}
