package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
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
	tx := global.DB.Begin()
	tx.Create(&t)
	for i, st := range req.Types {
		typ := model.SpecType{TemplateID: t.ID, Name: st.Name, Sort: i}
		tx.Create(&typ)
		for j, v := range st.Values {
			tx.Create(&model.SpecValue{TypeID: typ.ID, Value: v, Sort: j})
		}
	}
	tx.Commit()
	global.DB.Preload("Types.Values").First(&t, t.ID)
	return &t, nil
}

func GetSpecTemplateList() ([]model.SpecTemplate, error) {
	var list []model.SpecTemplate
	return list, global.DB.Preload("Types.Values").Find(&list).Error
}

func DeleteSpecTemplate(id uint) error {
	tx := global.DB.Begin()
	var types []model.SpecType
	tx.Where("template_id = ?", id).Find(&types)
	for _, t := range types {
		tx.Where("type_id = ?", t.ID).Delete(&model.SpecValue{})
	}
	tx.Where("template_id = ?", id).Delete(&model.SpecType{})
	tx.Delete(&model.SpecTemplate{}, id)
	tx.Commit()
	return nil
}
