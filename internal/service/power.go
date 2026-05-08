package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

func CreatePower(parentID uint, name, control string, sort int) (*model.Power, error) {
	p := model.Power{ParentID: parentID, Name: name, Control: control, Sort: sort, Status: 1}
	return &p, app.Must().DB.Create(&p).Error
}

func GetPowerTree() ([]model.Power, error) {
	var list []model.Power
	app.Must().DB.Where("parent_id = 0 AND status = 1").Order("sort DESC").Find(&list)
	for i := range list {
		app.Must().DB.Where("parent_id = ? AND status = 1", list[i].ID).Order("sort DESC").Find(&list[i].Children)
	}
	return list, nil
}

func DeletePower(id uint) error { return app.Must().DB.Delete(&model.Power{}, id).Error }

func SaveRolePowers(roleID uint, powerIDs []uint) error {
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePower{}).Error; err != nil {
			return err
		}
		for _, pid := range powerIDs {
			if err := tx.Create(&model.RolePower{RoleID: roleID, PowerID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func GetRolePowers(roleID uint) ([]uint, error) {
	var rps []model.RolePower
	app.Must().DB.Where("role_id = ?", roleID).Find(&rps)
	ids := make([]uint, len(rps))
	for i, rp := range rps {
		ids[i] = rp.PowerID
	}
	return ids, nil
}
