package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

func CreatePower(parentID uint, name, control string, sort int) (*model.Power, error) {
	p := model.Power{ParentID: parentID, Name: name, Control: control, Sort: sort, Status: 1}
	return &p, global.DB.Create(&p).Error
}

func GetPowerTree() ([]model.Power, error) {
	var list []model.Power
	global.DB.Where("parent_id = 0 AND status = 1").Order("sort DESC").Find(&list)
	for i := range list {
		global.DB.Where("parent_id = ? AND status = 1", list[i].ID).Order("sort DESC").Find(&list[i].Children)
	}
	return list, nil
}

func DeletePower(id uint) error { return global.DB.Delete(&model.Power{}, id).Error }

func SaveRolePowers(roleID uint, powerIDs []uint) error {
	tx := global.DB.Begin()
	tx.Where("role_id = ?", roleID).Delete(&model.RolePower{})
	for _, pid := range powerIDs {
		tx.Create(&model.RolePower{RoleID: roleID, PowerID: pid})
	}
	return tx.Commit().Error
}

func GetRolePowers(roleID uint) ([]uint, error) {
	var rps []model.RolePower
	global.DB.Where("role_id = ?", roleID).Find(&rps)
	ids := make([]uint, len(rps))
	for i, rp := range rps {
		ids[i] = rp.PowerID
	}
	return ids, nil
}
