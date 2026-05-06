package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type ParamsTemplateReq struct {
	Name    string             `json:"name" binding:"required"`
	Configs []ParamsConfigItem `json:"configs"`
}
type ParamsConfigItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func CreateParamsTemplate(req *ParamsTemplateReq) (*model.GoodsParamsTemplate, error) {
	t := model.GoodsParamsTemplate{Name: req.Name}
	tx := global.DB.Begin()
	tx.Create(&t)
	for i, c := range req.Configs {
		tx.Create(&model.GoodsParamsConfig{TemplateID: t.ID, Name: c.Name, Value: c.Value, Sort: i})
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	global.DB.Preload("Configs").First(&t, t.ID)
	return &t, nil
}

func GetParamsTemplateList() ([]model.GoodsParamsTemplate, error) {
	var list []model.GoodsParamsTemplate
	return list, global.DB.Preload("Configs").Find(&list).Error
}

func SaveGoodsParams(goodsID uint, params []ParamsConfigItem) error {
	global.DB.Where("goods_id = ?", goodsID).Delete(&model.GoodsParams{})
	for i, p := range params {
		global.DB.Create(&model.GoodsParams{GoodsID: goodsID, Name: p.Name, Value: p.Value, Sort: i})
	}
	return nil
}

func GetGoodsParams(goodsID uint) ([]model.GoodsParams, error) {
	var list []model.GoodsParams
	return list, global.DB.Where("goods_id = ?", goodsID).Order("sort").Find(&list).Error
}
