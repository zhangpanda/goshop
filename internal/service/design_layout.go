package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// Design
func DesignList() ([]model.Design, error) { var l []model.Design; return l, global.DB.Find(&l).Error }
func DesignCreate(name, data string) (*model.Design, error) {
	d := model.Design{Name: name, Data: data}
	return &d, global.DB.Create(&d).Error
}
func DesignUpdate(id uint, data string) error {
	return global.DB.Model(&model.Design{}).Where("id = ?", id).Update("data", data).Error
}

// Layout
func LayoutList() ([]model.Layout, error) { var l []model.Layout; return l, global.DB.Find(&l).Error }
func LayoutSave(name, typ, data string) error {
	var l model.Layout
	global.DB.Where("type = ?", typ).Find(&l)
	if l.ID > 0 {
		return global.DB.Model(&l).Updates(map[string]interface{}{"name": name, "data": data}).Error
	}
	return global.DB.Create(&model.Layout{Name: name, Type: typ, Data: data}).Error
}

// GoodsContentApp
func SaveGoodsContentApp(goodsID uint, content string) error {
	var c model.GoodsContentApp
	global.DB.Where("goods_id = ?", goodsID).Find(&c)
	if c.ID > 0 {
		return global.DB.Model(&c).Update("content", content).Error
	}
	return global.DB.Create(&model.GoodsContentApp{GoodsID: goodsID, Content: content}).Error
}
func GetGoodsContentApp(goodsID uint) string {
	var c model.GoodsContentApp
	global.DB.Where("goods_id = ?", goodsID).Find(&c)
	return c.Content
}

// OrderService
func CreateOrderService(orderID, userID uint, typ, content string) (*model.OrderService, error) {
	s := model.OrderService{OrderID: orderID, UserID: userID, Type: typ, Content: content}
	return &s, global.DB.Create(&s).Error
}
func ReplyOrderService(id, adminID uint, reply string) error {
	return global.DB.Model(&model.OrderService{}).Where("id = ?", id).
		Updates(map[string]interface{}{"admin_id": adminID, "reply": reply, "status": 1}).Error
}
func OrderServiceList(orderID uint) ([]model.OrderService, error) {
	var l []model.OrderService
	return l, global.DB.Where("order_id = ?", orderID).Order("id DESC").Find(&l).Error
}
func AdminOrderServiceList(status *int8, page, pageSize int) ([]model.OrderService, int64, error) {
	var total int64
	db := global.DB.Model(&model.OrderService{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	db.Count(&total)
	var l []model.OrderService
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&l).Error
	return l, total, err
}

// QuickNav
func QuickNavList() ([]model.QuickNav, error) {
	var l []model.QuickNav
	return l, global.DB.Where("status=1").Order("sort DESC").Find(&l).Error
}
func QuickNavCreate(n *model.QuickNav) error { return global.DB.Create(n).Error }

// PluginsDataConfig
func PluginConfigGet(pluginID uint, key string) string {
	var c model.PluginsDataConfig
	global.DB.Where("plugin_id = ? AND `key` = ?", pluginID, key).Find(&c)
	return c.Value
}
func PluginConfigSet(pluginID uint, key, value string) error {
	var c model.PluginsDataConfig
	global.DB.Where("plugin_id = ? AND `key` = ?", pluginID, key).Find(&c)
	if c.ID > 0 {
		return global.DB.Model(&c).Update("value", value).Error
	}
	return global.DB.Create(&model.PluginsDataConfig{PluginID: pluginID, Key: key, Value: value}).Error
}

// RolePlugins
func SaveRolePlugins(roleID uint, pluginIDs []uint) error {
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePlugins{}).Error; err != nil {
			return err
		}
		for _, pid := range pluginIDs {
			if err := tx.Create(&model.RolePlugins{RoleID: roleID, PluginID: pid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetRolePluginIDs 返回角色已绑定的插件 ID 列表。
func GetRolePluginIDs(roleID uint) ([]uint, error) {
	var rows []model.RolePlugins
	if err := global.DB.Where("role_id = ?", roleID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.PluginID)
	}
	return out, nil
}

// FormTableUserFields
func SaveFormFields(formID uint, fields []model.FormTableUserFields) error {
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Where("form_id = ?", formID).Delete(&model.FormTableUserFields{}).Error; err != nil {
			return err
		}
		for i := range fields {
			fields[i].FormID = formID
			fields[i].Sort = i
			if err := tx.Create(&fields[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func GetFormFields(formID uint) ([]model.FormTableUserFields, error) {
	var l []model.FormTableUserFields
	return l, global.DB.Where("form_id = ?", formID).Order("sort").Find(&l).Error
}
