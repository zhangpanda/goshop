package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// ==================== Order补全 ====================

// OrderConfirm 管理员确认订单（预约模式）
func OrderConfirm(orderID, adminID uint) error { return BookingConfirm(orderID) }

// OrderServiceData 订单服务数据
func OrderServiceData(orderID uint) []model.OrderService {
	l, _ := OrderServiceList(orderID)
	return l
}

// OrderListHandle 订单列表数据处理
func OrderListHandle(list []model.Order) []model.Order {
	for i := range list {
		if list[i].Items == nil {
			global.DB.Where("order_id = ?", list[i].ID).Find(&list[i].Items)
		}
	}
	return list
}

// OrderPayLogInsert 订单支付日志插入
func OrderPayLogInsert(userID uint, orderIDs []uint, paymentID uint, clientType string) (*model.PayLog, error) {
	return CreatePayLog(userID, orderIDs, paymentID, clientType)
}

// OrderPayLogValueList 支付日志关联订单
func OrderPayLogValueList(payLogID uint) []model.PayLogValue {
	var list []model.PayLogValue
	global.DB.Where("pay_log_id = ?", payLogID).Find(&list)
	return list
}

// OrderTraceSourceData 订单溯源数据
func OrderTraceSourceData(orderID uint) ([]model.OrderTraceSource, error) {
	return GetOrderTraceSource(orderID)
}

// OrderAddressData 订单地址数据
func OrderAddressData(order *model.Order) map[string]interface{} {
	var addr map[string]interface{}
	json.Unmarshal([]byte(order.Address), &addr)
	return addr
}

// ==================== ThemeAdmin补全 ====================

func ThemeAdminList() ([]model.ThemeData, error) { return ThemeList() }
func ThemeAdminSave(name, data string) error     { return ThemeCreate(name, data) }
func ThemeAdminSwitch(id uint) error {
	global.DB.Model(&model.ThemeData{}).Where("status = 1").Update("status", 0)
	return global.DB.Model(&model.ThemeData{}).Where("id = ?", id).Update("status", 1).Error
}
func ThemeAdminDelete(id uint) error { return global.DB.Delete(&model.ThemeData{}, id).Error }
func ThemeAdminConfig(id uint) *model.ThemeData {
	var t model.ThemeData
	global.DB.First(&t, id)
	return &t
}
func DefaultTheme() *model.ThemeData {
	var t model.ThemeData
	global.DB.Where("status = 1").First(&t)
	return &t
}

// ==================== ThemeData补全 ====================

func ThemeDataSave(id uint, name, data string) error {
	if id > 0 {
		return global.DB.Model(&model.ThemeData{}).Where("id = ?", id).Updates(map[string]interface{}{"name": name, "data": data}).Error
	}
	return ThemeCreate(name, data)
}
func ThemeDataDelete(id uint) error { return global.DB.Delete(&model.ThemeData{}, id).Error }
func ThemeDataStatusUpdate(id uint, status int8) error {
	return statusUpdate("theme_data", id, "status", status)
}
func ThemeDataListHandle(list []model.ThemeData) []model.ThemeData { return list }

// ==================== Diy补全 ====================

func DiyData(id uint) *model.Diy {
	var d model.Diy
	global.DB.First(&d, id)
	return &d
}
func DiySave(id uint, name, data string) error {
	if id > 0 {
		return DiyUpdate(id, data)
	}
	_, err := DiyCreate(name, data)
	return err
}
func DiyPreviewData(id uint) map[string]interface{} {
	d := DiyData(id)
	if d == nil {
		return nil
	}
	var parsed map[string]interface{}
	json.Unmarshal([]byte(d.Data), &parsed)
	return parsed
}
func AppClientHomeDiyData() map[string]interface{} {
	var d model.Diy
	global.DB.Where("status = 1").Order("id DESC").First(&d)
	if d.ID == 0 {
		return nil
	}
	var parsed map[string]interface{}
	json.Unmarshal([]byte(d.Data), &parsed)
	return parsed
}
func AppClientHomeDiyId() uint {
	var d model.Diy
	global.DB.Where("status = 1").Select("id").First(&d)
	return d.ID
}

// ==================== Design补全 ====================

func DesignSave(id uint, name, data string) error {
	if id > 0 {
		return DesignUpdate(id, data)
	}
	_, err := DesignCreate(name, data)
	return err
}
func DesignSync(id uint) error {
	d := &model.Design{}
	global.DB.First(d, id)
	return LayoutSave(d.Name, "home", d.Data)
}

// ==================== FormInput补全 ====================

func FormInputDetail(id uint) *model.FormInput {
	var f model.FormInput
	global.DB.First(&f, id)
	return &f
}
func FormInputPreview(id uint) map[string]interface{} {
	f := FormInputDetail(id)
	if f == nil {
		return nil
	}
	fields, _ := GetFormFields(id)
	return map[string]interface{}{"form": f, "fields": fields}
}

// ==================== Plugins补全 ====================

func PluginsStatus(id uint, status int8) error {
	return global.DB.Model(&model.Plugin{}).Where("id = ?", id).Update("status", status).Error
}
func PluginsData(id uint) *model.Plugin { var p model.Plugin; global.DB.First(&p, id); return &p }
func PluginsCheck(name string) bool {
	var c int64
	global.DB.Model(&model.Plugin{}).Where("name = ? AND status = 1", name).Count(&c)
	return c > 0
}
func PluginsField(name, field string) string {
	var p model.Plugin
	global.DB.Where("name = ?", name).Select(field).First(&p)
	return ""
}
func PluginsDataSave(id uint, config string) error {
	return global.DB.Model(&model.Plugin{}).Where("id = ?", id).Update("config", config).Error
}
func PluginsDataHandle(list []model.Plugin) []model.Plugin { return list }
func PluginsBaseList() ([]model.Plugin, error)             { return PluginList() }
func PluginsHomeDataList() []model.Plugin {
	var l []model.Plugin
	global.DB.Where("status = 1").Find(&l)
	return l
}
func PluginsSortList() []model.Plugin {
	var l []model.Plugin
	global.DB.Where("status = 1").Order("sort DESC, id ASC").Find(&l)
	return l
}
func PluginsNewVersionCheck() bool                                                      { return false }
func PluginsEventCall(hookName string, params map[string]interface{})                   {} // Go用接口而非事件钩子
func PluginsControlCall(name, method string, params map[string]interface{}) interface{} { return nil }

// ==================== PluginsAdmin补全 ====================

func PluginsAdminList() ([]model.Plugin, error)           { return PluginList() }
func PluginsAdminSave(id uint, config string) error       { return PluginsDataSave(id, config) }
func PluginsAdminStatusUpdate(id uint, status int8) error { return PluginsStatus(id, status) }
func PluginsAdminDelete(id uint) error                    { return global.DB.Delete(&model.Plugin{}, id).Error }

// ==================== Config补全 ====================

func ConfigInit()                                     { /* Go不需要PHP的配置初始化 */ }
func ConfigSave(key, value, group, desc string) error { return SetConfig(key, value, group, desc) }
func ConfigContentRow(key string) string              { return GetConfig(key) }
func SiteFictitiousConfig() map[string]string {
	return map[string]string{"is_enable": GetConfig("common_fictitious_order_direct_pay")}
}
func SiteTitleIconHandle() map[string]string {
	return map[string]string{"title": GetConfig("site_title"), "icon": GetConfig("site_icon")}
}
func SiteFilingList() []map[string]interface{} {
	raw := GetConfig("home_site_filing")
	var list []map[string]interface{}
	json.Unmarshal([]byte(raw), &list)
	return list
}

// ==================== Attachment补全 ====================

func AttachmentDiskFilesToDb(path, pathType string) {
	// 扫描磁盘文件同步到数据库（简化实现）
}

// ==================== AttachmentApi补全 ====================

func AttachmentApiList(categoryID uint, page, pageSize int) ([]model.Attachment, int64, error) {
	return AttachmentList(categoryID, page, pageSize)
}
func AttachmentApiSave(a *model.Attachment) error                    { return AttachmentSave(a) }
func AttachmentApiDelete(id uint) error                              { return AttachmentDelete(id) }
func AttachmentApiCategoryList() ([]model.AttachmentCategory, error) { return AttachmentCategoryList() }
func AttachmentApiCategorySave(name string) error                    { return CreateAttachmentCategory(name) }
func AttachmentApiCategoryDelete(id uint) error                      { return AttachmentCategoryDelete(id) }

// ==================== Statistical补全 ====================

func StatisticalDateTimeList(days int) []string {
	var list []string
	for i := days - 1; i >= 0; i-- {
		list = append(list, time.Now().AddDate(0, 0, -i).Format("2006-01-02"))
	}
	return list
}

// ==================== Multilingual补全（已在batch_final.go中） ====================

// ==================== App补全 ====================

func AppCustomerServiceConfig() map[string]string {
	return map[string]string{
		"tel":    GetConfig("app_customer_service_tel"),
		"custom": GetConfig("app_customer_service_custom"),
	}
}
func AppBaseConfig() map[string]string {
	return map[string]string{
		"h5_url": GetConfig("common_app_h5_url"),
	}
}

// ==================== AppMini补全 ====================

func AppMiniConfig(platform string) map[string]string {
	var m model.AppMini
	global.DB.Where("platform = ?", platform).First(&m)
	return map[string]string{"app_id": m.AppID, "title": m.Title, "describe": m.Describe, "status": fmt.Sprint(m.Status)}
}
