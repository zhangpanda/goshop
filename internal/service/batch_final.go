package service

import (
	jsonPkg "encoding/json"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// ==================== 批次7: 统计细分 ====================

type StatisticalExt struct {
	OrderProfitTotal    int64              `json:"order_profit_total" form:"order_profit_total"`
	PayTypeTotal        []PayTypeStat      `json:"pay_type_total" form:"pay_type_total"`
	BuyUserTotal        int64              `json:"buy_user_total" form:"buy_user_total"`
	OrderRegionTotal    []RegionStat       `json:"order_region_total" form:"order_region_total"`
	NewUserYesterday    int64              `json:"new_user_yesterday" form:"new_user_yesterday"`
	NewUserToday        int64              `json:"new_user_today" form:"new_user_today"`
	OrderCompleteToday  int64              `json:"order_complete_today" form:"order_complete_today"`
	OrderCompleteYesterday int64           `json:"order_complete_yesterday" form:"order_complete_yesterday"`
}
type PayTypeStat struct {
	PaymentKey string `json:"payment_key" form:"payment_key"`
	Count      int64  `json:"count" form:"count"`
	Amount     int64  `json:"amount" form:"amount"`
}
type RegionStat struct {
	Province string `json:"province" form:"province"`
	Count    int64  `json:"count" form:"count"`
}

func GetStatisticalExt() *StatisticalExt {
	d := &StatisticalExt{}
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	// 利润（简化：实付-退款）
	global.DB.Model(&model.Order{}).Where("status IN ?", []int8{1, 2, 3}).Select("COALESCE(SUM(pay_amount),0)").Scan(&d.OrderProfitTotal)
	var refund int64
	global.DB.Model(&model.RefundLog{}).Where("status = 1").Select("COALESCE(SUM(refund_price),0)").Scan(&refund)
	d.OrderProfitTotal -= refund
	// 支付方式统计
	global.DB.Model(&model.PayLog{}).Where("status = 1").Select("client_type as payment_key, COUNT(*) as count, SUM(total_price) as amount").Group("client_type").Find(&d.PayTypeTotal)
	// 付费用户数
	global.DB.Model(&model.Order{}).Where("status > 0").Distinct("user_id").Count(&d.BuyUserTotal)
	// 新用户
	global.DB.Model(&model.User{}).Where("DATE(created_at) = ?", today).Count(&d.NewUserToday)
	global.DB.Model(&model.User{}).Where("DATE(created_at) = ?", yesterday).Count(&d.NewUserYesterday)
	// 完成订单
	global.DB.Model(&model.Order{}).Where("status = 3 AND DATE(completed_at) = ?", today).Count(&d.OrderCompleteToday)
	global.DB.Model(&model.Order{}).Where("status = 3 AND DATE(completed_at) = ?", yesterday).Count(&d.OrderCompleteYesterday)
	return d
}

// ==================== 批次8: 用户模块补全 ====================

func UserTotal() int64 { return totalCount(&model.User{}) }

func UserSave(userID uint, nickname, avatar, phone string) error {
	updates := map[string]interface{}{}
	if nickname != "" { updates["nickname"] = nickname }
	if avatar != "" { updates["avatar"] = avatar }
	if phone != "" { updates["phone"] = phone }
	if len(updates) == 0 { return nil }
	return global.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func UserAvatarUpload(userID uint, avatarURL string) error {
	return global.DB.Model(&model.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error
}

func UserLoginRecord(userID uint, ip, platform string) {
	// 记录到消息或日志
	SendMessage(userID, "登录通知", fmt.Sprintf("您于%s在%s登录", time.Now().Format(time.DateTime), platform), "system", 0)
}

func UserListService(page, pageSize int, keyword string, status *int8) ([]model.User, int64, error) {
	var total int64
	db := global.DB.Model(&model.User{})
	if keyword != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil { db = db.Where("status = ?", *status) }
	db.Count(&total)
	var list []model.User
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func UserDelete(userID uint) error {
	return global.DB.Delete(&model.User{}, userID).Error
}

// ==================== 批次9: 订单模块补全 ====================

// OrderDirectSuccess 虚拟商品直接完成
func OrderDirectSuccess(orderID uint) error {
	now := time.Now()
	global.DB.Model(&model.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status": model.OrderStatusCompleted, "paid_at": &now, "completed_at": &now,
	})
	AddOrderStatusHistory(orderID, model.OrderStatusPending, model.OrderStatusCompleted, "虚拟商品自动完成", "系统")
	return nil
}

// OrderPayCheck 支付前校验
func OrderPayCheck(userID, orderID uint) (*model.Order, error) {
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, fmt.Errorf("订单状态不允许支付")
	}
	return &order, nil
}

// OrderExpressData 订单物流信息
func OrderExpressData(orderID uint) (*model.Shipment, []TrackInfo, error) {
	s, err := GetShipment(orderID)
	if err != nil { return nil, nil, err }
	tracks, _ := QueryExpress(s.ExpressCompany, s.ExpressNo)
	return s, tracks, nil
}

// OrderExtractionData 自提订单信息
type ExtractionInfo struct {
	Code    string `json:"code" form:"code"`
	Address string `json:"address" form:"address"`
}

func OrderExtractionData(orderID uint) *ExtractionInfo {
	var order model.Order
	global.DB.Select("extraction_code, address").First(&order, orderID)
	return &ExtractionInfo{Code: order.ExtractionCode, Address: order.Address}
}

// OrderStepData 订单进度步骤
type OrderStep struct {
	Name   string `json:"name" form:"name"`
	Status int8   `json:"status"` // 0未完成 1当前 2已完成
	Time   string `json:"time" form:"time"`
}

func OrderStepData(order *model.Order) []OrderStep {
	steps := []OrderStep{
		{Name: "提交订单", Status: 2, Time: order.CreatedAt.Format(time.DateTime)},
		{Name: "付款", Status: 0},
		{Name: "发货", Status: 0},
		{Name: "收货", Status: 0},
		{Name: "完成", Status: 0},
	}
	if order.PaidAt != nil { steps[1] = OrderStep{Name: "付款", Status: 2, Time: order.PaidAt.Format(time.DateTime)} }
	if order.ShippedAt != nil { steps[2] = OrderStep{Name: "发货", Status: 2, Time: order.ShippedAt.Format(time.DateTime)} }
	if order.CompletedAt != nil {
		steps[3] = OrderStep{Name: "收货", Status: 2, Time: order.CompletedAt.Format(time.DateTime)}
		steps[4] = OrderStep{Name: "完成", Status: 2, Time: order.CompletedAt.Format(time.DateTime)}
	}
	// 标记当前步骤
	for i := range steps {
		if steps[i].Status == 0 { steps[i].Status = 1; break }
	}
	return steps
}

// OrderItemList 独立获取订单商品列表
func OrderItemList(orderID uint) []model.OrderItem {
	var list []model.OrderItem
	global.DB.Where("order_id = ?", orderID).Find(&list)
	return list
}

// ==================== 批次10: 零散模块补全 ====================

// PayLogClose 关闭支付日志
func PayLogClose(payNo string) error {
	now := time.Now()
	return global.DB.Model(&model.PayLog{}).Where("pay_no = ? AND status = 0", payNo).
		Updates(map[string]interface{}{"status": 2, "closed_at": &now}).Error
}

// PayLogTypeList 支付日志业务类型列表
func PayLogTypeList() []map[string]string {
	return []map[string]string{
		{"value": "system-order", "name": "系统订单"},
	}
}

// ConfigList 获取所有配置
func ConfigList() ([]model.Config, error) {
	var list []model.Config
	return list, global.DB.Order("`group`, `key`").Find(&list).Error
}

// SmsTemplateValue 获取短信模板
func SmsTemplateValue(typ string) string {
	return GetConfig("sms_" + typ + "_template")
}

// MultilingualSetUserValue 设置用户语言偏好
func MultilingualSetUserValue(userID uint, lang string) {
	SetConfig(fmt.Sprintf("user_%d_lang", userID), lang, "multilingual", "用户语言偏好")
}

func MultilingualGetUserValue(userID uint) string {
	v := GetConfig(fmt.Sprintf("user_%d_lang", userID))
	if v == "" { return GetMultilingualConfig().DefaultLang }
	return v
}

// ==================== 批次11: DiyApi/AppMiniUser/FormInputApi ====================

// DiyApiCustomInit DIY自定义组件初始化
func DiyApiCustomInit(diyID uint) (map[string]interface{}, error) {
	var diy model.Diy
	if err := global.DB.First(&diy, diyID).Error; err != nil { return nil, err }
	DiyAccessCountInc(diyID)
	var data map[string]interface{}
	if diy.Data != "" {
		jsonUnmarshal([]byte(diy.Data), &data)
	}
	return data, nil
}

func jsonUnmarshal(b []byte, v interface{}) { _ = jsonPkg.Unmarshal(b, v) }

// DiyApiUserHeadData 用户头部数据（消息数+购物车数+收藏数）
func DiyApiUserHeadData(userID uint) map[string]int64 {
	return map[string]int64{
		"message_total":  UnreadCount(userID),
		"cart_total":     GoodsCartTotal(userID),
		"favor_total":    GoodsFavorTotal(userID),
	}
}

// AppMiniUserInfo 获取各平台用户信息
func AppMiniUserInfo(userID uint, platform string) (*model.UserPlatform, error) {
	var p model.UserPlatform
	if err := global.DB.Where("user_id = ? AND platform = ?", userID, platform).First(&p).Error; err != nil {
		return nil, fmt.Errorf("未绑定%s", platform)
	}
	return &p, nil
}

// AppMiniOnekeyUserMobileBind 一键手机号绑定（小程序获取手机号）
func AppMiniOnekeyUserMobileBind(userID uint, mobile string) error {
	return global.DB.Model(&model.User{}).Where("id = ?", userID).Update("phone", mobile).Error
}

// FormInputApiList Form表单API列表
func FormInputApiList(page, pageSize int) ([]model.FormInput, int64, error) {
	var total int64
	global.DB.Model(&model.FormInput{}).Where("status = 1").Count(&total)
	var list []model.FormInput
	err := global.DB.Where("status = 1").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// FormInputApiDetail Form表单API详情
func FormInputApiDetail(id uint) (*model.FormInput, error) {
	var f model.FormInput
	if err := global.DB.First(&f, id).Error; err != nil { return nil, err }
	return &f, nil
}

// FormInputApiSave Form表单API保存数据
func FormInputApiSave(formID, userID uint, data string) error {
	return FormInputDataSubmit(formID, userID, data)
}

// FormInputApiDelete Form表单API删除数据
func FormInputApiDelete(id, userID uint) error {
	return global.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.FormInputData{}).Error
}
