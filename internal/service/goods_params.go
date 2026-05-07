package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
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
	if err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Create(&t).Error; err != nil {
			return err
		}
		for i, c := range req.Configs {
			if err := tx.Create(&model.GoodsParamsConfig{TemplateID: t.ID, Name: c.Name, Value: c.Value, Sort: i}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
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
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Where("goods_id = ?", goodsID).Delete(&model.GoodsParams{}).Error; err != nil {
			return err
		}
		for i, p := range params {
			if err := tx.Create(&model.GoodsParams{GoodsID: goodsID, Name: p.Name, Value: p.Value, Sort: i}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func GetGoodsParams(goodsID uint) ([]model.GoodsParams, error) {
	var list []model.GoodsParams
	return list, global.DB.Where("goods_id = ?", goodsID).Order("sort").Find(&list).Error
}
