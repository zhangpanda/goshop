package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

func CreateBrand(name, logo, desc string, sort int) (*model.Brand, error) {
	b := model.Brand{Name: name, Logo: logo, Desc: desc, Sort: sort, Status: 1}
	return &b, global.DB.Create(&b).Error
}

func GetBrandList() ([]model.Brand, error) {
	var list []model.Brand
	return list, global.DB.Where("status = 1").Order("sort DESC, id DESC").Find(&list).Error
}

func UpdateBrand(id uint, name, logo, desc string, sort int) error {
	return global.DB.Model(&model.Brand{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": name, "logo": logo, "desc": desc, "sort": sort,
	}).Error
}

func DeleteBrand(id uint) error { return global.DB.Delete(&model.Brand{}, id).Error }
