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

func shopxoFmtYuanFen(fen int64) string {
	return fmt.Sprintf("%.2f", float64(fen)/100)
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
	stName := shopxoOrderStatusName[o.Status]
	if stName == "" {
		stName = "未知"
	}
	op := OrderOperateButtons(o)
	return map[string]interface{}{
		"id":                      o.ID,
		"status":                  int(o.Status),
		"status_name":             stName,
		"warehouse_name":          "",
		"warehouse_url":           "",
		"warehouse_icon":          nil,
		"is_under_line_text":      nil,
		"payment_id":              DefaultPaymentIDForShopXO(),
		"total_price":             shopxoFmtYuanFen(o.PayAmount),
		"buy_number_count":        buyCount,
		"currency_data":           map[string]string{"currency_symbol": sym},
		"items":                   items,
		"operate_data":            shopxoOperateDataInt(op),
		"is_can_launch_aftersale": 0,
		"order_model":             o.OrderModel,
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
 * ShopXOCompatUnifiedPay 供 uni-app /api.php?s=order/pay：单笔订单 + 支付方式表主键。
 * outerMsg 非空时 HTTP 外层使用自定义 msg（与线下支付弹窗文案对齐）。
 */
func ShopXOCompatUnifiedPay(userID, orderID, paymentID uint, openID, returnURL, clientIP string) (map[string]interface{}, *string, error) {
	if global.DB == nil {
		return nil, nil, fmt.Errorf("数据库未初始化")
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
	resp, err := UnifiedPay(userID, &UnifiedPayReq{
		OrderID:    orderID,
		PaymentKey: driverKey,
		OpenID:     openID,
		ReturnURL:  returnURL,
		ClientIP:   clientIP,
	})
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
