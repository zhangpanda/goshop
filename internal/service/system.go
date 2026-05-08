package service

import (
	"strings"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// ---- 系统配置 ----
func GetConfig(key string) string {
	var c model.Config
	app.Must().DB.Where("`key` = ?", key).Find(&c)
	return c.Value
}

func SetConfig(key, value, group, desc string) error {
	var c model.Config
	app.Must().DB.Where("`key` = ?", key).Find(&c)
	if c.ID > 0 {
		return app.Must().DB.Model(&c).Updates(map[string]interface{}{"value": value, "desc": desc}).Error
	}
	return app.Must().DB.Create(&model.Config{Group: group, Key: key, Value: value, Desc: desc}).Error
}

func GetConfigGroup(group string) ([]model.Config, error) {
	var list []model.Config
	return list, app.Must().DB.Where("`group` = ?", group).Find(&list).Error
}

/**
 * CustomerServiceTel 读取客服电话，与 InitDefaultConfig 中 common_app_customer_service_tel 一致。
 * 若为空则回退 app_customer_service_tel（历史错误 key，兼容旧数据）。
 */
func CustomerServiceTel() string {
	if v := strings.TrimSpace(GetConfig("common_app_customer_service_tel")); v != "" {
		return v
	}
	return strings.TrimSpace(GetConfig("app_customer_service_tel"))
}

/**
 * CustomerServiceCustom 读取客服自定义文案/链接等扩展配置。
 * 优先 common_app_customer_service_custom，回退 app_customer_service_custom。
 */
func CustomerServiceCustom() string {
	if v := strings.TrimSpace(GetConfig("common_app_customer_service_custom")); v != "" {
		return v
	}
	return strings.TrimSpace(GetConfig("app_customer_service_custom"))
}

// ---- 地区 ----
func GetRegionList(parentID uint) ([]model.Region, error) {
	var list []model.Region
	return list, app.Must().DB.Where("parent_id = ?", parentID).Order("sort, id").Find(&list).Error
}

// ---- 幻灯片/导航/链接 通用CRUD ----
func SlideList() ([]model.Slide, error) {
	var l []model.Slide
	return l, app.Must().DB.Where("status=1").Order("sort DESC").Find(&l).Error
}
func CreateSlide(s *model.Slide) error { return app.Must().DB.Create(s).Error }

func NavigationList(typ string) ([]model.Navigation, error) {
	var l []model.Navigation
	db := app.Must().DB.Where("status=1")
	if typ != "" {
		db = db.Where("type=?", typ)
	}
	return l, db.Order("sort DESC").Find(&l).Error
}
func CreateNavigation(n *model.Navigation) error { return app.Must().DB.Create(n).Error }

func LinkList() ([]model.Link, error) {
	var l []model.Link
	return l, app.Must().DB.Where("status=1").Order("sort DESC").Find(&l).Error
}
func CreateLink(l *model.Link) error { return app.Must().DB.Create(l).Error }

// ---- 支付方式 ----
func PaymentList() ([]model.Payment, error) {
	var l []model.Payment
	return l, app.Must().DB.Order("sort DESC").Find(&l).Error
}
func CreatePayment(p *model.Payment) error { return app.Must().DB.Create(p).Error }

// ---- 附件 ----
func AttachmentList(categoryID uint, page, pageSize int) ([]model.Attachment, int64, error) {
	var total int64
	db := app.Must().DB.Model(&model.Attachment{})
	if categoryID > 0 {
		db = db.Where("category_id=?", categoryID)
	}
	db.Count(&total)
	var list []model.Attachment
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
func DeleteAttachment(id uint) error { return app.Must().DB.Delete(&model.Attachment{}, id).Error }
func AttachmentCategoryList() ([]model.AttachmentCategory, error) {
	var l []model.AttachmentCategory
	return l, app.Must().DB.Find(&l).Error
}
func CreateAttachmentCategory(name string) error {
	return app.Must().DB.Create(&model.AttachmentCategory{Name: name}).Error
}

// ---- 错误日志 ----
func AddErrorLog(typ, content, url, ip string) {
	app.Must().DB.Create(&model.ErrorLog{Type: typ, Content: content, URL: url, IP: ip})
}
func GetErrorLogList(page, pageSize int) ([]model.ErrorLog, int64, error) {
	var total int64
	app.Must().DB.Model(&model.ErrorLog{}).Count(&total)
	var list []model.ErrorLog
	err := app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ---- 订单状态历史 ----
func AddOrderStatusHistory(orderID uint, oldStatus, newStatus int8, msg, creator string) {
	app.Must().DB.Create(&model.OrderStatusHistory{OrderID: orderID, OriginalStatus: oldStatus, NewStatus: newStatus, Msg: msg, Creator: creator})
}
func GetOrderStatusHistory(orderID uint) ([]model.OrderStatusHistory, error) {
	var list []model.OrderStatusHistory
	return list, app.Must().DB.Where("order_id = ?", orderID).Order("id").Find(&list).Error
}
