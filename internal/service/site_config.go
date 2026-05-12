package service

import (
	"encoding/json"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

func GetSiteConfig() map[string]string {
	var configs []model.Config
	app.Must().DB.Where("`group` IN ('site','base','seo','app') OR `key` LIKE 'home_site%' OR `key` LIKE 'home_seo%' OR `key` LIKE 'home_footer%'").Find(&configs)
	result := make(map[string]string, len(configs))
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result
}

func SaveSiteConfig(configs map[string]string) {
	for k, v := range configs {
		SetConfig(k, v, "site", "")
	}
}

func GetSelfExtractionAddressList() []map[string]interface{} {
	raw := GetConfig("site_self_extraction_address")
	if raw == "" {
		return nil
	}
	var list []map[string]interface{}
	json.Unmarshal([]byte(raw), &list)
	return list
}

func SaveSelfExtractionAddress(list []map[string]interface{}) {
	data, _ := json.Marshal(list)
	SetConfig("site_self_extraction_address", string(data), "site", "自提点地址")
}

type AppMiniReq struct {
	Platform  string `json:"platform" binding:"required"`
	Title     string `json:"title"`
	Describe  string `json:"describe"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	Status    int8   `json:"status"`
}

func SaveAppMini(req *AppMiniReq) error {
	var m model.AppMini
	app.Must().DB.Where("platform = ?", req.Platform).First(&m)
	if m.ID > 0 {
		return app.Must().DB.Model(&m).Updates(map[string]interface{}{
			"title": req.Title, "describe": req.Describe,
			"app_id": req.AppID, "app_secret": req.AppSecret, "status": req.Status,
		}).Error
	}
	return app.Must().DB.Create(&model.AppMini{
		Platform: req.Platform, Title: req.Title, Describe: req.Describe,
		AppID: req.AppID, AppSecret: req.AppSecret, Status: req.Status,
	}).Error
}

func GetAppMiniList() ([]model.AppMini, error) {
	var list []model.AppMini
	return list, app.Must().DB.Order("id ASC").Find(&list).Error
}

func DeleteAppMini(id uint) error { return app.Must().DB.Delete(&model.AppMini{}, id).Error }
