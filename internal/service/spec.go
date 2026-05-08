package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

type SpecTemplateReq struct {
	Name  string        `json:"name" binding:"required"`
	Types []SpecTypeReq `json:"types" binding:"required,min=1"`
}
type SpecTypeReq struct {
	Name   string   `json:"name" binding:"required"`
	Values []string `json:"values" binding:"required,min=1"`
}

func CreateSpecTemplate(req *SpecTemplateReq) (*model.SpecTemplate, error) {
	t := model.SpecTemplate{Name: req.Name}
	err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Create(&t).Error; err != nil {
			return err
		}
		for i, st := range req.Types {
			typ := model.SpecType{TemplateID: t.ID, Name: st.Name, Sort: i}
			if err := tx.Create(&typ).Error; err != nil {
				return err
			}
			for j, v := range st.Values {
				if err := tx.Create(&model.SpecValue{TypeID: typ.ID, Value: v, Sort: j}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	app.Must().DB.Preload("Types.Values").First(&t, t.ID)
	return &t, nil
}

func GetSpecTemplateList() ([]model.SpecTemplate, error) {
	var list []model.SpecTemplate
	return list, app.Must().DB.Preload("Types.Values").Find(&list).Error
}

func DeleteSpecTemplate(id uint) error {
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		var types []model.SpecType
		if err := tx.Where("template_id = ?", id).Find(&types).Error; err != nil {
			return err
		}
		for _, t := range types {
			if err := tx.Where("type_id = ?", t.ID).Delete(&model.SpecValue{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("template_id = ?", id).Delete(&model.SpecType{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.SpecTemplate{}, id).Error
	})
}
