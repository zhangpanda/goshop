package shopxo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/wechat"
)

// 本文件：/api.php 兼容层用的订单/支付 JSON 形状（对照 shopxo-uniapp 常见字段；默认值防前端空指针）。

var shopxoOrderStatusName = map[int8]string{
	model.OrderStatusPending:   "待付款",
	model.OrderStatusPaid:      "待发货",
	model.OrderStatusShipped:   "待收货",
	model.OrderStatusCompleted: "已完成",
	model.OrderStatusCancelled: "已取消",
	model.OrderStatusRefunded:  "已退款",
	model.OrderStatusBooking:   "预约待确认",
}

// 订单状态 ID 映射（对照 ShopXO v6.8 常见 common_order_status 编号，便于 uni-app 展示）
var shopxoStatusIDByInternal = map[int8]int{
	model.OrderStatusBooking:   0,
	model.OrderStatusPending:   1,
	model.OrderStatusPaid:      2,
	model.OrderStatusShipped:   3,
	model.OrderStatusCompleted: 4,
	model.OrderStatusCancelled: 5,
	model.OrderStatusRefunded:  4,
}

var shopxoStatusNameByShopXOStatusID = map[int]string{
	0: "待确认",
	1: "待付款",
	2: "待发货",
	3: "待收货",
	4: "已完成",
	5: "已取消",
	6: "已关闭",
}

var shopxoOrderModelName = map[int8]string{
	model.OrderModelExpress: "快递",
	model.OrderModelLocal:   "同城",
	model.OrderModelPickup:  "自提",
	model.OrderModelVirtual: "虚拟",
}

func shopxoFmtYuanFen(fen int64) string {
	return fmt.Sprintf("%.2f", float64(fen)/100)
}

func shopxoShopXOStatusMeta(internal int8) (sxID int, name string) {
	sxID = shopxoStatusIDByInternal[internal]
	if sxID == 0 && internal != model.OrderStatusBooking {
		sxID = 1
	}
	name = shopxoStatusNameByShopXOStatusID[sxID]
	if name == "" {
		name = shopxoOrderStatusName[internal]
	}
	if name == "" {
		name = "未知"
	}
	return sxID, name
}

func shopxoPayStatusMeta(o *model.Order) (payStatus int, payName string) {
	if o == nil {
		return 0, "未支付"
	}
	if o.Status == model.OrderStatusRefunded {
		return 2, "已退款"
	}
	if o.Status == model.OrderStatusPending || o.Status == model.OrderStatusBooking {
		return 0, "未支付"
	}
	if o.PaidAt != nil || o.Status >= model.OrderStatusPaid {
		return 1, "已支付"
	}
	return 0, "未支付"
}

func shopxoAddressDataFromOrder(o *model.Order) map[string]interface{} {
	if o == nil {
		return nil
	}
	raw := strings.TrimSpace(o.Address)
	if raw == "" || raw == "{}" {
		return nil
	}
	var addr model.Address
	if json.Unmarshal([]byte(raw), &addr) != nil {
		return nil
	}
	return map[string]interface{}{
		"name":          addr.Name,
		"tel":           addr.Phone,
		"province_name": addr.Province,
		"city_name":     addr.City,
		"county_name":   addr.District,
		"address":       addr.Detail,
		"lng":           addr.Lng,
		"lat":           addr.Lat,
	}
}

func shopxoPaymentNameByID(id uint) string {
	if global.DB == nil || id == 0 {
		return ""
	}
	var p model.Payment
	if err := global.DB.First(&p, id).Error; err != nil {
		return ""
	}
	return p.Name
}

// shopxoOrderDisplayPaymentID 列表/详情：优先订单已存 payment_id，否则默认支付方式。
func shopxoOrderDisplayPaymentID(o *model.Order) uint {
	if o != nil && o.PaymentID > 0 {
		return o.PaymentID
	}
	return DefaultPaymentIDForShopXO()
}

// ShopXOOrderDetailView 构造 uni-app 订单详情 data.data（状态码形状与上表映射一致）。
func ShopXOOrderDetailView(o *model.Order) map[string]interface{} {
	if o == nil {
		return map[string]interface{}{}
	}
	sxStatus, statusName := shopxoShopXOStatusMeta(o.Status)
	paySt, payStName := shopxoPayStatusMeta(o)
	op := service.OrderOperateButtons(o)
	oper := shopxoOperateDataInt(op)
	pid := shopxoOrderDisplayPaymentID(o)
	omName := shopxoOrderModelName[o.OrderModel]
	if omName == "" {
		omName = "快递"
	}
	sym := shopxoCurrencySymbol()
	items := make([]map[string]interface{}, 0, len(o.Items))
	for _, it := range o.Items {
		spec := []map[string]string{}
		if it.SkuName != "" {
			spec = append(spec, map[string]string{"name": "规格", "value": it.SkuName})
		}
		items = append(items, map[string]interface{}{
			"id":                      it.ID,
			"goods_id":                it.GoodsID,
			"goods_url":               "",
			"images":                  it.Image,
			"title":                   it.Title,
			"spec":                    spec,
			"price":                   shopxoFmtYuanFen(it.Price),
			"buy_number":              it.Quantity,
			"orderaftersale_btn_text": nil,
		})
	}
	out := map[string]interface{}{
		"id":                                    o.ID,
		"order_no":                              o.OrderNo,
		"status":                                sxStatus,
		"status_name":                           statusName,
		"pay_status":                            paySt,
		"pay_status_name":                       payStName,
		"warehouse_name":                        "",
		"warehouse_url":                         "",
		"warehouse_icon":                        nil,
		"order_model":                           int(o.OrderModel),
		"order_model_name":                      omName,
		"price":                                 shopxoFmtYuanFen(o.TotalAmount),
		"total_price":                           shopxoFmtYuanFen(o.TotalAmount),
		"preferential_price":                    "0.00",
		"increase_price":                        "0.00",
		"pay_price":                             shopxoFmtYuanFen(o.PayAmount),
		"payment_id":                            pid,
		"payment_name":                          shopxoPaymentNameByID(pid),
		"is_under_line_text":                    nil,
		"user_note":                             o.Remark,
		"currency_data":                         map[string]string{"currency_symbol": sym},
		"items":                                 items,
		"operate_data":                          oper,
		"extension_data":                        []interface{}{},
		"is_can_launch_aftersale":               0,
		"plugins_express_data":                  0,
		"express_data":                          nil,
		"plugins_delivery_data":                 0,
		"plugins_is_order_allot_button":         0,
		"plugins_is_order_batch_button":         0,
		"plugins_is_order_frequencycard_button": 0,
		"plugins_ordergoodsform_data":           0,
		"plugins_orderresources_data":           0,
		"plugins_is_orderfeed_button":           0,
		"plugins_intellectstools_data":          nil,
		"add_time":                              o.CreatedAt.Format("2006-01-02 15:04:05"),
		"confirm_time":                          "",
		"pay_time":                              "",
		"delivery_time":                         "",
		"collect_time":                          "",
		"cancel_time":                           "",
		"close_time":                            "",
	}
	if o.PaidAt != nil {
		out["pay_time"] = o.PaidAt.Format("2006-01-02 15:04:05")
	}
	if o.ShippedAt != nil {
		out["delivery_time"] = o.ShippedAt.Format("2006-01-02 15:04:05")
	}
	if o.CompletedAt != nil {
		out["collect_time"] = o.CompletedAt.Format("2006-01-02 15:04:05")
	}
	if addr := shopxoAddressDataFromOrder(o); addr != nil {
		out["address_data"] = addr
	}
	if o.OrderModel == model.OrderModelPickup && strings.TrimSpace(o.ExtractionCode) != "" {
		out["extraction_data"] = map[string]interface{}{
			"code": o.ExtractionCode,
		}
	}
	return out
}

// shopxoCurrencySymbol 未初始化 DB 时回退 ¥（单测、脚本）。
func shopxoCurrencySymbol() string {
	if global.DB == nil {
		return "¥"
	}
	return service.GetCurrencyConfig().Symbol
}

func shopxoOperateDataInt(op *service.OrderOperate) map[string]int {
	if op == nil {
		return map[string]int{
			"is_cancel": 0, "is_pay": 0, "is_collect": 0, "is_comments": 0, "is_delete": 0,
		}
	}
	b2i := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	return map[string]int{
		"is_cancel":   b2i(op.CanCancel),
		"is_pay":      b2i(op.CanPay),
		"is_collect":  b2i(op.CanReceive),
		"is_comments": b2i(op.CanReview),
		"is_delete":   b2i(op.CanDelete),
	}
}

// ShopXOOrderListRow 单条订单 → uni-app 订单列表最小字段集。
func ShopXOOrderListRow(o *model.Order) map[string]interface{} {
	if o == nil {
		return map[string]interface{}{}
	}
	sym := shopxoCurrencySymbol()
	items := make([]map[string]interface{}, 0, len(o.Items))
	var buyCount int
	for _, it := range o.Items {
		buyCount += it.Quantity
		spec := []map[string]string{}
		if it.SkuName != "" {
			spec = append(spec, map[string]string{"name": "规格", "value": it.SkuName})
		}
		items = append(items, map[string]interface{}{
			"id":                      it.ID,
			"images":                  it.Image,
			"title":                   it.Title,
			"spec":                    spec,
			"price":                   shopxoFmtYuanFen(it.Price),
			"buy_number":              it.Quantity,
			"orderaftersale_btn_text": nil,
		})
	}
	sxSt, stName := shopxoShopXOStatusMeta(o.Status)
	op := service.OrderOperateButtons(o)
	return map[string]interface{}{
		"id":                                    o.ID,
		"status":                                sxSt,
		"status_name":                           stName,
		"warehouse_name":                        "",
		"warehouse_url":                         "",
		"warehouse_icon":                        nil,
		"is_under_line_text":                    nil,
		"payment_id":                            shopxoOrderDisplayPaymentID(o),
		"total_price":                           shopxoFmtYuanFen(o.PayAmount),
		"buy_number_count":                      buyCount,
		"currency_data":                         map[string]string{"currency_symbol": sym},
		"items":                                 items,
		"operate_data":                          shopxoOperateDataInt(op),
		"is_can_launch_aftersale":               0,
		"order_model":                           int(o.OrderModel),
		"weixin_collect_data":                   "",
		"plugins_express_data":                  0,
		"express_data":                          nil,
		"plugins_delivery_data":                 0,
		"plugins_is_order_allot_button":         0,
		"plugins_is_order_batch_button":         0,
		"plugins_is_order_frequencycard_button": 0,
		"plugins_ordergoodsform_data":           0,
		"plugins_orderresources_data":           0,
		"plugins_is_orderfeed_button":           0,
		"plugins_intellectstools_data":          nil,
	}
}

// DefaultPaymentIDForShopXO 列表/详情用默认支付方式（订单无 payment_id 时回退）。
func DefaultPaymentIDForShopXO() uint {
	if global.DB == nil {
		return 0
	}
	if id := service.BuyDefaultPayment("common"); id > 0 {
		return id
	}
	if id := service.BuyDefaultPayment(""); id > 0 {
		return id
	}
	var p model.Payment
	if err := global.DB.Where("status = 1").Order("sort DESC, id").First(&p).Error; err == nil {
		return p.ID
	}
	return 0
}

// ShopXOUserPaymentRows 用户端可选支付方式（启用中的配置行）。
func ShopXOUserPaymentRows() ([]map[string]interface{}, error) {
	if global.DB == nil {
		return []map[string]interface{}{}, nil
	}
	var list []model.Payment
	if err := global.DB.Where("status = 1").Order("sort DESC, id").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		p := &list[i]
		key, _ := PaymentDriverKeyFromPayment(p)
		out = append(out, map[string]interface{}{
			"id":        p.ID,
			"name":      p.Name,
			"logo":      p.Logo,
			"payment":   ShopXOPluginNameFromDriverKey(key),
			"tips":      nil,
			"config":    p.Config,
			"sort":      p.Sort,
			"is_enable": p.Status,
		})
	}
	return out, nil
}

// ShopXOOrderIndexPayload 构造 order/index 的 data（user-order 列表所需最小字段集）。
func ShopXOOrderIndexPayload(userID uint, req *service.OrderListReq) (map[string]interface{}, error) {
	if global.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	listReq := *req
	listReq.PageSize = pageSize
	resp, err := service.GetOrderList(userID, &listReq)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, 0, len(resp.List))
	for i := range resp.List {
		rows = append(rows, ShopXOOrderListRow(&resp.List[i]))
	}
	payRows, _ := ShopXOUserPaymentRows()
	if payRows == nil {
		payRows = []map[string]interface{}{}
	}
	pageTotal := int(math.Ceil(float64(resp.Total) / float64(pageSize)))
	if pageTotal == 0 && resp.Total > 0 {
		pageTotal = 1
	}
	return map[string]interface{}{
		"total":              resp.Total,
		"page_total":         pageTotal,
		"data":               rows,
		"payment_list":       payRows,
		"default_payment_id": DefaultPaymentIDForShopXO(),
	}, nil
}

// PaymentDriverKeyFromPayment 从支付方式配置解析 UnifiedPay 的 payment_key。
func PaymentDriverKeyFromPayment(p *model.Payment) (string, error) {
	return service.PaymentDriverKeyFromPayment(p)
}

// paymentShopXOIsWeixinAppMini 是否走「APP 拉起小程序收银台」模式（无 openid 时 order/pay 建 PayLog + weixinapp://）。
// 配置：`"payment":"WeixinAppMini"` 或名称含「APP小程序」等（约定来自常见 shopxo 系配置习惯）。
func paymentShopXOIsWeixinAppMini(p *model.Payment) bool {
	if p == nil {
		return false
	}
	n := p.Name
	if strings.Contains(n, "APP小程序") || strings.Contains(n, "小程序收银") {
		return true
	}
	raw := strings.TrimSpace(p.Config)
	if raw == "" {
		return false
	}
	var cfg map[string]interface{}
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		return false
	}
	s, ok := cfg["payment"].(string)
	return ok && strings.TrimSpace(s) == "WeixinAppMini"
}

// shopxoCashierMiniPath 小程序收银台 path（默认 pages/cashier/cashier，与常见 WeixinAppMini 配置一致）。
func shopxoCashierMiniPath(p *model.Payment) string {
	const def = "pages/cashier/cashier"
	if p == nil {
		return def
	}
	var cfg map[string]interface{}
	if json.Unmarshal([]byte(strings.TrimSpace(p.Config)), &cfg) != nil {
		return def
	}
	if s, ok := cfg["path"].(string); ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	return def
}

func shopxoPHPClassToDriverKey(class string) string {
	return service.ShopxoPHPClassToDriverKey(class)
}

func inferPaymentKeyFromPaymentName(name string) string {
	return service.InferPaymentKeyFromPaymentName(name)
}

// ShopXOPluginNameFromDriverKey 映射为 uni-app payment 组件期望的 data.payment.payment 字符串。
func ShopXOPluginNameFromDriverKey(driverKey string) string {
	switch {
	case strings.HasPrefix(driverKey, "wechat"):
		return "Weixin"
	case strings.HasPrefix(driverKey, "alipay"):
		return "Alipay"
	case driverKey == "wallet":
		return "WalletPay"
	case driverKey == "offline":
		return "CashPayment"
	case driverKey == "paypal":
		return "PayPal"
	default:
		return "Weixin"
	}
}

// ShopXOPayPayloadFromDriver 内部支付结果 → uni-app payment 组件可消费结构。
func ShopXOPayPayloadFromDriver(driverKey string, p *model.Payment, prep *service.PayDriverResp, offlineUserMsg string) map[string]interface{} {
	plugin := ShopXOPluginNameFromDriverKey(driverKey)
	payRow := map[string]interface{}{"payment": plugin}
	if p != nil {
		payRow["id"] = p.ID
		payRow["name"] = p.Name
	}
	if driverKey == "wallet" {
		return map[string]interface{}{
			"is_success":      1,
			"is_payment_type": 2,
			"payment":         payRow,
		}
	}
	if driverKey == "offline" {
		msg := offlineUserMsg
		if msg == "" {
			msg = "请按线下说明完成付款，付款后请等待商家确认。"
		}
		return map[string]interface{}{
			"is_success":      0,
			"is_payment_type": 1,
			"payment":         payRow,
			"msg":             msg,
		}
	}
	inner := interface{}(nil)
	if prep != nil {
		if prep.PayURL != "" {
			inner = prep.PayURL
		} else if len(prep.PrepayData) > 0 {
			inner = prep.PrepayData
		} else if prep.QRCode != "" {
			inner = map[string]interface{}{
				"qrcode_url": prep.QRCode,
				"name":       p.Name,
				"order_no":   prep.TradeNo,
			}
		}
	}
	return map[string]interface{}{
		"is_success":      0,
		"is_payment_type": 0,
		"data":            inner,
		"payment":         payRow,
	}
}

// ShopXOCompatUnifiedPay 供 /api.php?s=order/pay：单笔或多笔订单（ids）+ 支付方式主键。
// outerMsg 非空时外层 HTTP msg 用自定义文案（如线下支付提示）。
func ShopXOCompatUnifiedPay(userID uint, orderIDs []uint, paymentID uint, openID, returnURL, clientIP string) (map[string]interface{}, *string, error) {
	if global.DB == nil {
		return nil, nil, fmt.Errorf("数据库未初始化")
	}
	if len(orderIDs) == 0 {
		return nil, nil, fmt.Errorf("请选择订单")
	}
	var pay model.Payment
	if err := global.DB.First(&pay, paymentID).Error; err != nil {
		return nil, nil, fmt.Errorf("支付方式不存在")
	}
	if pay.Status != 1 {
		return nil, nil, fmt.Errorf("支付方式已禁用")
	}
	driverKey, keyErr := PaymentDriverKeyFromPayment(&pay)
	if keyErr != nil || strings.TrimSpace(driverKey) == "" {
		driverKey = inferPaymentKeyFromPaymentName(pay.Name)
	}

	// APP 拉起小程序收银台：先建 PayLog，返回 weixinapp://；小程序内 cashier/paydata 完成 JSAPI 预下单（对照常见 WeixinAppMini 流程）。
	if strings.TrimSpace(openID) == "" && driverKey == "wechat_jsapi" && paymentShopXOIsWeixinAppMini(&pay) {
		// clientType "shopxo"：兼容层来源标记，见 CreatePayLog 注释
		pl, err := service.CreatePayLog(userID, orderIDs, paymentID, "shopxo")
		if err != nil {
			return nil, nil, err
		}
		path := shopxoCashierMiniPath(&pay)
		payURL := fmt.Sprintf("weixinapp://%s?order_no=%s", path, url.QueryEscape(pl.PayNo))
		payRow := map[string]interface{}{
			"payment": ShopXOPluginNameFromDriverKey(driverKey),
			"id":      pay.ID,
			"name":    pay.Name,
		}
		return map[string]interface{}{
			"is_success":      0,
			"is_payment_type": 0,
			"data":            payURL,
			"order_id":        pl.ID,
			"order_no":        pl.PayNo,
			"payment":         payRow,
		}, nil, nil
	}

	var resp *service.PayDriverResp
	var err error
	if len(orderIDs) == 1 {
		resp, err = service.UnifiedPay(userID, &service.UnifiedPayReq{
			OrderID:         orderIDs[0],
			PaymentKey:      driverKey,
			OpenID:          openID,
			ReturnURL:       returnURL,
			ClientIP:        clientIP,
			PaymentRecordID: paymentID,
		})
	} else {
		resp, err = service.MultiOrderUnifiedPay(userID, orderIDs, driverKey, paymentID, openID, returnURL, clientIP)
	}
	if err != nil {
		return nil, nil, err
	}
	if driverKey == "offline" {
		offlineCopy := "请按线下说明完成付款，订单已标记为已支付，如有疑问请联系商家。"
		p := ShopXOPayPayloadFromDriver(driverKey, &pay, resp, offlineCopy)
		msg := offlineCopy
		if m, ok := p["msg"].(string); ok && m != "" {
			msg = m
		}
		return p, &msg, nil
	}
	return ShopXOPayPayloadFromDriver(driverKey, &pay, resp, ""), nil, nil
}

/**
 * cashierOpenIDBelongsToUser 校验小程序 openid 是否属于该支付单所属用户（UserPlatform 或 users.open_id）。
 */
func cashierOpenIDBelongsToUser(userID uint, openID string) bool {
	if userID == 0 || strings.TrimSpace(openID) == "" {
		return false
	}
	var n int64
	global.DB.Model(&model.UserPlatform{}).Where("user_id = ? AND openid = ?", userID, openID).Count(&n)
	if n > 0 {
		return true
	}
	var u model.User
	if err := global.DB.First(&u, userID).Error; err != nil {
		return false
	}
	return u.OpenID == openID
}

// ShopXOCashierPayData 对应 api.php?s=cashier/paydata：authcode 换 openid 后对 PayLog 发起微信 JSAPI 预下单。
func ShopXOCashierPayData(authCode, payNo string) (map[string]interface{}, error) {
	authCode = strings.TrimSpace(authCode)
	payNo = strings.TrimSpace(payNo)
	if authCode == "" {
		return nil, errors.New("authcode 不能为空")
	}
	if payNo == "" {
		return nil, errors.New("order_no 不能为空")
	}
	if global.Cfg == nil || strings.TrimSpace(global.Cfg.Wechat.AppID) == "" {
		return nil, errors.New("微信小程序未配置")
	}
	sess, err := wechat.Code2Session(global.Cfg.Wechat.AppID, global.Cfg.Wechat.AppSecret, authCode)
	if err != nil {
		return nil, err
	}
	openID := strings.TrimSpace(sess.OpenID)
	if openID == "" {
		return nil, errors.New("未获取到 openid")
	}
	var pl model.PayLog
	if err := global.DB.Where("pay_no = ?", payNo).First(&pl).Error; err != nil {
		return nil, errors.New("支付单不存在")
	}
	if pl.Status != 0 {
		return nil, errors.New("支付单已处理")
	}
	if !cashierOpenIDBelongsToUser(pl.UserID, openID) {
		return nil, errors.New("当前微信与订单用户不一致，请使用下单账号对应小程序登录")
	}
	var pay model.Payment
	if err := global.DB.First(&pay, pl.PaymentID).Error; err != nil {
		return nil, errors.New("支付方式不存在")
	}
	driverKey, _ := PaymentDriverKeyFromPayment(&pay)
	if strings.TrimSpace(driverKey) == "" {
		driverKey = inferPaymentKeyFromPaymentName(pay.Name)
	}
	driver, err := service.GetPaymentDriver(driverKey)
	if err != nil {
		return nil, err
	}
	resp, err := driver.Pay(context.Background(), &service.PayDriverReq{
		OrderNo:     pl.PayNo,
		Description: "商城订单",
		Amount:      pl.TotalPrice,
		OpenID:      openID,
		ClientIP:    "",
	})
	if err != nil {
		return nil, err
	}
	out := ShopXOPayPayloadFromDriver(driverKey, &pay, resp, "")
	out["order_id"] = pl.ID
	out["order_no"] = pl.PayNo
	return out, nil
}
