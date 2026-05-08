package service

import (
	"encoding/json"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// ---- 插件 ----
func PluginList() ([]model.Plugin, error) {
	var l []model.Plugin
	return l, app.Must().DB.Find(&l).Error
}
func PluginInstall(name, title, desc, author, version string) error {
	return app.Must().DB.Create(&model.Plugin{Name: name, Title: title, Desc: desc, Author: author, Version: version, Status: 1}).Error
}
func PluginUninstall(id uint) error {
	return app.Must().DB.Model(&model.Plugin{}).Where("id = ?", id).Update("status", 0).Error
}

// ---- DIY ----
func DiyList() ([]model.Diy, error) { var l []model.Diy; return l, app.Must().DB.Find(&l).Error }
func DiyCreate(name, data string) (*model.Diy, error) {
	d := model.Diy{Name: name, Data: data}
	return &d, app.Must().DB.Create(&d).Error
}
func DiyUpdate(id uint, data string) error {
	return app.Must().DB.Model(&model.Diy{}).Where("id = ?", id).Update("data", data).Error
}
func DiyDelete(id uint) error { return app.Must().DB.Delete(&model.Diy{}, id).Error }

// ---- 自定义页面 ----
func CustomViewList() ([]model.CustomView, error) {
	var l []model.CustomView
	return l, app.Must().DB.Where("status=1").Find(&l).Error
}
func CustomViewCreate(title, content string) (*model.CustomView, error) {
	v := model.CustomView{Title: title, Content: content}
	return &v, app.Must().DB.Create(&v).Error
}

// ---- 主题 ----
func ThemeList() ([]model.ThemeData, error) {
	var l []model.ThemeData
	return l, app.Must().DB.Find(&l).Error
}
func ThemeCreate(name, data string) error {
	return app.Must().DB.Create(&model.ThemeData{Name: name, Data: data}).Error
}

// ---- 表单 ----
func FormInputList() ([]model.FormInput, error) {
	var l []model.FormInput
	return l, app.Must().DB.Find(&l).Error
}
func FormInputCreate(name, config string) error {
	return app.Must().DB.Create(&model.FormInput{Name: name, Config: config, Status: 1}).Error
}
func FormInputDataSubmit(formID, userID uint, data string) error {
	return app.Must().DB.Create(&model.FormInputData{FormID: formID, UserID: userID, Data: data}).Error
}
func FormInputDataList(formID uint, page, pageSize int) ([]model.FormInputData, int64, error) {
	var total int64
	app.Must().DB.Model(&model.FormInputData{}).Where("form_id = ?", formID).Count(&total)
	var l []model.FormInputData
	err := app.Must().DB.Where("form_id = ?", formID).Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&l).Error
	return l, total, err
}

// FormInputDataDetailForUser shopxo-uniapp forminputdata/detail：单条提交记录 + 字段定义。
func FormInputDataDetailForUser(userID, rowID uint) (map[string]interface{}, error) {
	var row model.FormInputData
	if err := app.Must().DB.Where("id = ? AND user_id = ?", rowID, userID).First(&row).Error; err != nil {
		return nil, err
	}
	fields, _ := GetFormFields(row.FormID)
	var payload interface{}
	_ = json.Unmarshal([]byte(row.Data), &payload)
	return map[string]interface{}{
		"data":       payload,
		"field_list": fields,
	}, nil
}

// ---- APP导航 ----
func AppHomeNavList() ([]model.AppHomeNav, error) {
	var l []model.AppHomeNav
	return l, app.Must().DB.Where("status=1").Order("sort DESC").Find(&l).Error
}
func AppHomeNavCreate(n *model.AppHomeNav) error { return app.Must().DB.Create(n).Error }
func AppCenterNavList() ([]model.AppCenterNav, error) {
	var l []model.AppCenterNav
	return l, app.Must().DB.Where("status=1").Order("sort DESC").Find(&l).Error
}
func AppCenterNavCreate(n *model.AppCenterNav) error { return app.Must().DB.Create(n).Error }
func AppTabbarList() ([]model.AppTabbar, error) {
	var l []model.AppTabbar
	return l, app.Must().DB.Order("sort").Find(&l).Error
}
func AppTabbarSave(items []model.AppTabbar) error {
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Where("1=1").Delete(&model.AppTabbar{}).Error; err != nil {
			return err
		}
		for i := range items {
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- 快捷菜单 ----
func ShortcutMenuList() ([]model.ShortcutMenu, error) {
	var l []model.ShortcutMenu
	return l, app.Must().DB.Order("sort").Find(&l).Error
}
func ShortcutMenuSave(items []model.ShortcutMenu) error {
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Where("1=1").Delete(&model.ShortcutMenu{}).Error; err != nil {
			return err
		}
		for i := range items {
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- 协议 ----
func AgreementGet(name string) *model.Agreement {
	var a model.Agreement
	app.Must().DB.Where("name = ?", name).Find(&a)
	return &a
}
func AgreementSave(name, content string) error {
	var a model.Agreement
	app.Must().DB.Where("name = ?", name).Find(&a)
	if a.ID > 0 {
		return app.Must().DB.Model(&a).Update("content", content).Error
	}
	return app.Must().DB.Create(&model.Agreement{Name: name, Content: content}).Error
}

// ---- 订单溯源 ----
func AddOrderTraceSource(orderID, userID uint, source, params string) {
	app.Must().DB.Create(&model.OrderTraceSource{OrderID: orderID, UserID: userID, Source: source, Params: params})
}
func GetOrderTraceSource(orderID uint) ([]model.OrderTraceSource, error) {
	var l []model.OrderTraceSource
	return l, app.Must().DB.Where("order_id = ?", orderID).Find(&l).Error
}
