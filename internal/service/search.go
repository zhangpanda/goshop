package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

func AddSearchHistory(userID uint, keyword string) {
	if keyword == "" {
		return
	}
	var h model.SearchHistory
	app.Must().DB.Where("user_id = ? AND keyword = ?", userID, keyword).Find(&h)
	if h.ID > 0 {
		app.Must().DB.Save(&h)
		return
	}
	app.Must().DB.Create(&model.SearchHistory{UserID: userID, Keyword: keyword})
}

func GetSearchHistory(userID uint) ([]model.SearchHistory, error) {
	var list []model.SearchHistory
	return list, app.Must().DB.Where("user_id = ?", userID).Order("id DESC").Limit(20).Find(&list).Error
}

func ClearSearchHistory(userID uint) error {
	return app.Must().DB.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}

func GetHotKeywords(limit int) ([]string, error) {
	var results []struct{ Keyword string }
	err := app.Must().DB.Model(&model.SearchHistory{}).Select("keyword, COUNT(*) as cnt").
		Group("keyword").Order("cnt DESC").Limit(limit).Find(&results).Error
	var keywords []string
	for _, r := range results {
		keywords = append(keywords, r.Keyword)
	}
	return keywords, err
}

func GetScreeningPrices() ([]model.ScreeningPrice, error) {
	var list []model.ScreeningPrice
	return list, app.Must().DB.Order("sort").Find(&list).Error
}
