package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

/**
 * ShopXO uni-app 兼容：订单列表单条、支付方式解析等（仅用于兼容层，保持字段有默认值避免前端空指针）。
 */

var shopxoOrderStatusName = map[int8]string{
	model.OrderStatusPending:   "待付款",
	model.OrderStatusPaid:      "待发货",
	model.OrderStatusShipped:   "待收货",
	model.OrderStatusCompleted: "已完成",
	model.OrderStatusCancelled: "已取消",
	model.OrderStatusRefunded:  "已退款",
	model.OrderStatusBooking:   "预约待确认",
}

// ShopXO 订单状态 ID（与 ShopXO ConstService common_order_status 一致：0待确认…6已关闭）
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

/**
 * shopxoOrderDisplayPaymentID 列表/详情展示：订单已记录的支付方式优先，否则回退默认。
 */
func shopxoOrderDisplayPaymentID(o *model.Order) uint {
	if o != nil && o.PaymentID > 0 {
		return o.PaymentID
	}
	return DefaultPaymentIDForShopXO()
}

/**
 * ShopXOOrderDetailView 构造 uni-app 订单详情页 data.data 所需字段（与 ShopXO 状态码对齐）。
 */
func ShopXOOrderDetailView(o *model.Order) map[string]interface{} {
	if o == nil {
		return map[string]interface{}{}
	}
	sxStatus, statusName := shopxoShopXOStatusMeta(o.Status)
	paySt, payStName := shopxoPayStatusMeta(o)
	op := OrderOperateButtons(o)
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
		"id":                       o.ID,
		"order_no":                 o.OrderNo,
		"status":                   sxStatus,
		"status_name":              statusName,
		"pay_status":               paySt,
		"pay_status_name":          payStName,
		"warehouse_name":         "",
		"warehouse_url":            "",
		"warehouse_icon":           nil,
		"order_model":              int(o.OrderModel),
		"order_model_name":         omName,
		"price":                    shopxoFmtYuanFen(o.TotalAmount),
		"total_price":              shopxoFmtYuanFen(o.TotalAmount),
		"preferential_price":       "0.00",
		"increase_price":           "0.00",
		"pay_price":                shopxoFmtYuanFen(o.PayAmount),
		"payment_id":               pid,
		"payment_name":             shopxoPaymentNameByID(pid),
		"is_under_line_text":       nil,
		"user_note":                o.Remark,
		"currency_data":            map[string]string{"currency_symbol": sym},
		"items":                    items,
		"operate_data":             oper,
		"extension_data":           []interface{}{},
		"is_can_launch_aftersale":  0,
		"plugins_express_data":     0,
		"express_data":             nil,
		"plugins_delivery_data":    0,
		"plugins_is_order_allot_button":         0,
		"plugins_is_order_batch_button":         0,
		"plugins_is_order_frequencycard_button": 0,
		"plugins_ordergoodsform_data":           0,
		"plugins_orderresources_data":           0,
		"plugins_is_orderfeed_button":           0,
		"plugins_intellectstools_data":          nil,
		"add_time":                 o.CreatedAt.Format("2006-01-02 15:04:05"),
		"confirm_time":             "",
		"pay_time":                 "",
		"delivery_time":            "",
		"collect_time":             "",
		"cancel_time":              "",
		"close_time":               "",
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

/**
 * shopxoCurrencySymbol 兼容未初始化 DB 的场景（单测、脚本）。
 */
func shopxoCurrencySymbol() string {
	if global.DB == nil {
		return "¥"
	}
	return GetCurrencyConfig().Symbol
}

func shopxoOperateDataInt(op *OrderOperate) map[string]int {
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

/**
 * ShopXOOrderListRow 将单条订单转为 uni-app 列表所需的最小字段集。
 */
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
			"id":          it.ID,
			"images":      it.Image,
			"title":       it.Title,
			"spec":        spec,
			"price":       shopxoFmtYuanFen(it.Price),
			"buy_number":  it.Quantity,
			"orderaftersale_btn_text": nil,
		})
	}
	sxSt, stName := shopxoShopXOStatusMeta(o.Status)
	op := OrderOperateButtons(o)
	return map[string]interface{}{
		"id":                      o.ID,
		"status":                  sxSt,
		"status_name":             stName,
		"warehouse_name":          "",
		"warehouse_url":           "",
		"warehouse_icon":          nil,
		"is_under_line_text":      nil,
		"payment_id":              shopxoOrderDisplayPaymentID(o),
		"total_price":             shopxoFmtYuanFen(o.PayAmount),
		"buy_number_count":        buyCount,
		"currency_data":           map[string]string{"currency_symbol": sym},
		"items":                   items,
		"operate_data":            shopxoOperateDataInt(op),
		"is_can_launch_aftersale": 0,
		"order_model":             int(o.OrderModel),
		"weixin_collect_data":     "",
		"plugins_express_data":    0,
		"express_data":            nil,
		"plugins_delivery_data":   0,
		"plugins_is_order_allot_button":        0,
		"plugins_is_order_batch_button":        0,
		"plugins_is_order_frequencycard_button": 0,
		"plugins_ordergoodsform_data":          0,
		"plugins_orderresources_data":          0,
		"plugins_is_orderfeed_button":          0,
		"plugins_intellectstools_data":         nil,
	}
}

/**
 * DefaultPaymentIDForShopXO 列表/详情展示用默认支付方式（订单表无 payment_id 时的回退）。
 */
func DefaultPaymentIDForShopXO() uint {
	if global.DB == nil {
		return 0
	}
	if id := BuyDefaultPayment("common"); id > 0 {
		return id
	}
	if id := BuyDefaultPayment(""); id > 0 {
		return id
	}
	var p model.Payment
	if err := global.DB.Where("status = 1").Order("sort DESC, id").First(&p).Error; err == nil {
		return p.ID
	}
	return 0
}

/**
 * ShopXOUserPaymentRows 用户端可选支付方式（启用中的配置行）。
 */
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
			"id":         p.ID,
			"name":       p.Name,
			"logo":       p.Logo,
			"payment":    ShopXOPluginNameFromDriverKey(key),
			"tips":       nil,
			"config":     p.Config,
			"sort":       p.Sort,
			"is_enable":  p.Status,
		})
	}
	return out, nil
}

/**
 * ShopXOOrderIndexPayload 构造 order/index 接口 data 字段（与 uni-app user-order 一致的最小形状）。
 */
func ShopXOOrderIndexPayload(userID uint, req *OrderListReq) (map[string]interface{}, error) {
	if global.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	listReq := *req
	listReq.PageSize = pageSize
	resp, err := GetOrderList(userID, &listReq)
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
		"total":               resp.Total,
		"page_total":          pageTotal,
		"data":                rows,
		"payment_list":        payRows,
		"default_payment_id":  DefaultPaymentIDForShopXO(),
	}, nil
}

/**
 * PaymentDriverKeyFromPayment 从支付方式配置解析 UnifiedPay 使用的 payment_key。
 */
func PaymentDriverKeyFromPayment(p *model.Payment) (string, error) {
	if p == nil {
		return "", errors.New("支付方式不存在")
	}
	raw := strings.TrimSpace(p.Config)
	if raw != "" {
		var cfg map[string]interface{}
		if json.Unmarshal([]byte(raw), &cfg) == nil {
			if s, ok := cfg["payment_key"].(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s), nil
			}
			if s, ok := cfg["payment"].(string); ok && strings.TrimSpace(s) != "" {
				return shopxoPHPClassToDriverKey(strings.TrimSpace(s)), nil
			}
		}
	}
	return inferPaymentKeyFromPaymentName(p.Name), nil
}

func shopxoPHPClassToDriverKey(class string) string {
	switch class {
	case "Weixin", "WeixinAppMini", "WeixinScanQrcode", "WeixinH5":
		return "wechat_jsapi"
	case "Alipay", "AlipayMini", "AlipayScanQrcode", "AlipayH5", "AlipayCert":
		return "alipay_h5"
	case "WalletPay":
		return "wallet"
	case "CashPayment", "DeliveryPayment":
		return "offline"
	default:
		return inferPaymentKeyFromPaymentName(class)
	}
}

func inferPaymentKeyFromPaymentName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return "wechat_jsapi"
	}
	if strings.Contains(n, "微信") || strings.Contains(n, "wechat") || strings.Contains(n, "weixin") {
		return "wechat_jsapi"
	}
	if strings.Contains(n, "支付宝") || strings.Contains(n, "alipay") {
		return "alipay_h5"
	}
	if strings.Contains(n, "钱包") || strings.Contains(n, "wallet") {
		return "wallet"
	}
	if strings.Contains(n, "线下") || strings.Contains(n, "货到付款") || strings.Contains(n, "现金") {
		return "offline"
	}
	return "wechat_jsapi"
}

/**
 * ShopXOPluginNameFromDriverKey 与 uni-app payment 组件中 data.payment.payment 对齐。
 */
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

/**
 * ShopXOPayPayloadFromDriver 将内部支付结果转为 uni-app payment 组件可消费的结构。
 */
func ShopXOPayPayloadFromDriver(driverKey string, p *model.Payment, prep *PayDriverResp, offlineUserMsg string) map[string]interface{} {
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

/**
 * ShopXOCompatUnifiedPay 供 uni-app /api.php?s=order/pay：支持单笔或多笔订单（ids）+ 支付方式表主键。
 * outerMsg 非空时 HTTP 外层使用自定义 msg（与线下支付弹窗文案对齐）。
 */
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
	var resp *PayDriverResp
	var err error
	if len(orderIDs) == 1 {
		resp, err = UnifiedPay(userID, &UnifiedPayReq{
			OrderID:         orderIDs[0],
			PaymentKey:      driverKey,
			OpenID:          openID,
			ReturnURL:       returnURL,
			ClientIP:        clientIP,
			PaymentRecordID: paymentID,
		})
	} else {
		resp, err = MultiOrderUnifiedPay(userID, orderIDs, driverKey, paymentID, openID, returnURL, clientIP)
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
