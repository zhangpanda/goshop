package service

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	net_http "net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/wechat"
)

// PaymentDriver 支付驱动接口
type PaymentDriver interface {
	Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error)
	Refund(ctx context.Context, req *RefundDriverReq) error
}

type PayDriverReq struct {
	OrderNo     string
	Description string
	Amount      int64 // 分
	ClientIP    string
	OpenID      string // 微信支付需要
	ReturnURL   string // PC支付回跳
}

type PayDriverResp struct {
	PayURL     string                 `json:"pay_url,omitempty"`     // 跳转支付URL
	PrepayData map[string]interface{} `json:"prepay_data,omitempty"` // 小程序/APP调起参数
	QRCode     string                 `json:"qr_code,omitempty"`     // 二维码链接
	TradeNo    string                 `json:"trade_no,omitempty"`
}

type RefundDriverReq struct {
	OrderNo  string
	RefundNo string
	Total    int64
	Refund   int64
	Reason   string
}

// ========== 微信JSAPI（已有，包装一下） ==========

type WechatJSAPIDriver struct{}

func (d *WechatJSAPIDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	if global.WxPay == nil {
		return nil, errors.New("微信支付未配置")
	}
	resp, err := global.WxPay.Prepay(ctx, &wechat.PrepayRequest{OrderNo: req.OrderNo, Description: req.Description, Amount: req.Amount, OpenID: req.OpenID})
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"appId":     resp.Appid,
		"timeStamp": resp.TimeStamp,
		"nonceStr":  resp.NonceStr,
		"package":   resp.Package,
		"signType":  resp.SignType,
		"paySign":   resp.PaySign,
	}
	return &PayDriverResp{PrepayData: data}, nil
}

func (d *WechatJSAPIDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	if global.WxPay == nil {
		return errors.New("微信支付未配置")
	}
	_, err := global.WxPay.Refund(ctx, &wechat.RefundRequest{OrderNo: req.OrderNo, RefundNo: req.RefundNo, Total: req.Total, Refund: req.Refund, Reason: req.Reason})
	return err
}

// 类型别名避免循环引用
type wechatPrepayReq = struct {
	OrderNo     string
	Description string
	Amount      int64
	OpenID      string
}
type wechatRefundReq = struct {
	OrderNo  string
	RefundNo string
	Total    int64
	Refund   int64
	Reason   string
}

// ========== 微信H5支付 ==========

type WechatH5Driver struct{}

func (d *WechatH5Driver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	if global.WxPay == nil {
		return nil, errors.New("微信支付未配置")
	}
	h5URL, err := global.WxPay.PrepayH5(ctx, &wechat.PrepayRequest{
		OrderNo: req.OrderNo, Description: req.Description, Amount: req.Amount,
	}, req.ClientIP)
	if err != nil {
		return nil, fmt.Errorf("微信H5预下单失败: %w", err)
	}
	return &PayDriverResp{PayURL: h5URL}, nil
}

func (d *WechatH5Driver) Refund(ctx context.Context, req *RefundDriverReq) error {
	return (&WechatJSAPIDriver{}).Refund(ctx, req)
}

// ========== 支付宝支付 ==========

type AlipayDriver struct {
	ClientType string // pc, h5, app
}

func (d *AlipayDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	cfg := global.Cfg.Alipay
	if cfg.AppID == "" {
		return nil, errors.New("支付宝未配置")
	}
	params := map[string]string{
		"app_id":      cfg.AppID,
		"method":      d.method(),
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  cfg.NotifyURL,
		"biz_content": fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s","product_code":"%s"}`, req.OrderNo, float64(req.Amount)/100, req.Description, d.productCode()),
	}
	if req.ReturnURL != "" {
		params["return_url"] = req.ReturnURL
	}
	params["sign"] = alipaySign(params, cfg.PrivateKey)
	payURL := "https://openapi.alipay.com/gateway.do?" + alipayEncode(params)
	return &PayDriverResp{PayURL: payURL}, nil
}

func (d *AlipayDriver) method() string {
	switch d.ClientType {
	case "h5":
		return "alipay.trade.wap.pay"
	case "app":
		return "alipay.trade.app.pay"
	default:
		return "alipay.trade.page.pay"
	}
}

func (d *AlipayDriver) productCode() string {
	switch d.ClientType {
	case "h5":
		return "QUICK_WAP_WAY"
	case "app":
		return "QUICK_MSECURITY_PAY"
	default:
		return "FAST_INSTANT_TRADE_PAY"
	}
}

func (d *AlipayDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	// 支付宝退款需要调用 alipay.trade.refund 接口
	cfg := global.Cfg.Alipay
	if cfg.AppID == "" {
		return errors.New("支付宝未配置")
	}
	params := map[string]string{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.refund",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": fmt.Sprintf(`{"out_trade_no":"%s","refund_amount":"%.2f","out_request_no":"%s","refund_reason":"%s"}`, req.OrderNo, float64(req.Refund)/100, req.RefundNo, req.Reason),
	}
	params["sign"] = alipaySign(params, cfg.PrivateKey)
	// 发起退款请求
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	_, err := (&net_http.Client{Timeout: 30 * time.Second}).PostForm("https://openapi.alipay.com/gateway.do", data)
	return err
}

// ========== 线下支付 ==========

type OfflineDriver struct{}

func (d *OfflineDriver) Pay(_ context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	return &PayDriverResp{TradeNo: "OFFLINE_" + req.OrderNo}, nil
}

func (d *OfflineDriver) Refund(_ context.Context, _ *RefundDriverReq) error {
	return nil // 线下退款由管理员手动处理
}

// ========== 驱动注册 ==========

var paymentDrivers = map[string]PaymentDriver{
	"wechat_jsapi":  &WechatJSAPIDriver{},
	"wechat_h5":     &WechatH5Driver{},
	"wechat_app":    &WechatAppDriver{},
	"wechat_native": &WechatNativeDriver{},
	"alipay_pc":     &AlipayDriver{ClientType: "pc"},
	"alipay_h5":     &AlipayDriver{ClientType: "h5"},
	"alipay_app":    &AlipayDriver{ClientType: "app"},
	"alipay_mini":   &AlipayMiniDriver{},
	"alipay_face":   &AlipayFaceDriver{},
	"paypal":        &PayPalDriver{},
	"wallet":        &WalletPayDriver{},
	"offline":       &OfflineDriver{},
}

// ========== 微信APP支付 ==========

type WechatAppDriver struct{}

func (d *WechatAppDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	if global.WxPay == nil {
		return nil, errors.New("微信支付未配置")
	}
	resp, err := global.WxPay.PrepayApp(ctx, &wechat.PrepayRequest{
		OrderNo: req.OrderNo, Description: req.Description, Amount: req.Amount,
	})
	if err != nil {
		return nil, fmt.Errorf("微信APP预下单失败: %w", err)
	}
	return &PayDriverResp{PrepayData: map[string]interface{}{
		"partnerid": resp.PartnerId, "prepayid": resp.PrepayId,
		"package": resp.Package, "noncestr": resp.NonceStr, "timestamp": resp.TimeStamp, "sign": resp.Sign,
	}}, nil
}
func (d *WechatAppDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	return (&WechatJSAPIDriver{}).Refund(ctx, req)
}

// ========== 微信扫码(Native)支付 ==========

type WechatNativeDriver struct{}

func (d *WechatNativeDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	if global.WxPay == nil {
		return nil, errors.New("微信支付未配置")
	}
	codeURL, err := global.WxPay.PrepayNative(ctx, &wechat.PrepayRequest{
		OrderNo: req.OrderNo, Description: req.Description, Amount: req.Amount,
	})
	if err != nil {
		return nil, fmt.Errorf("微信Native预下单失败: %w", err)
	}
	return &PayDriverResp{QRCode: codeURL}, nil
}
func (d *WechatNativeDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	return (&WechatJSAPIDriver{}).Refund(ctx, req)
}

// ========== 支付宝小程序 ==========

type AlipayMiniDriver struct{}

func (d *AlipayMiniDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	cfg := global.Cfg.Alipay
	if cfg.AppID == "" {
		return nil, errors.New("支付宝未配置")
	}
	return &PayDriverResp{PrepayData: map[string]interface{}{
		"tradeNO": req.OrderNo, "orderStr": fmt.Sprintf("alipay_sdk=goshop&app_id=%s&total_amount=%.2f", cfg.AppID, float64(req.Amount)/100),
	}}, nil
}
func (d *AlipayMiniDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	return (&AlipayDriver{}).Refund(ctx, req)
}

// ========== 支付宝面对面(当面付) ==========

type AlipayFaceDriver struct{}

func (d *AlipayFaceDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	cfg := global.Cfg.Alipay
	if cfg.AppID == "" {
		return nil, errors.New("支付宝未配置")
	}
	return &PayDriverResp{QRCode: fmt.Sprintf("https://qr.alipay.com/%s", req.OrderNo)}, nil
}
func (d *AlipayFaceDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	return (&AlipayDriver{}).Refund(ctx, req)
}

// ========== PayPal ==========

type PayPalDriver struct{}

func (d *PayPalDriver) Pay(_ context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	// PayPal Checkout URL占位，实际需对接PayPal REST API
	return &PayDriverResp{PayURL: fmt.Sprintf("https://www.paypal.com/checkoutnow?token=%s", req.OrderNo)}, nil
}
func (d *PayPalDriver) Refund(_ context.Context, _ *RefundDriverReq) error { return nil }

// ========== 钱包余额支付 ==========

type WalletPayDriver struct{}

func (d *WalletPayDriver) Pay(_ context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	// 钱包支付在UnifiedPay中特殊处理
	return &PayDriverResp{TradeNo: "WALLET_" + req.OrderNo}, nil
}
func (d *WalletPayDriver) Refund(_ context.Context, _ *RefundDriverReq) error { return nil }

func GetPaymentDriver(name string) (PaymentDriver, error) {
	d, ok := paymentDrivers[name]
	if !ok {
		return nil, fmt.Errorf("不支持的支付方式: %s", name)
	}
	return d, nil
}

// ========== 统一支付入口 ==========

type UnifiedPayReq struct {
	OrderID    uint   `json:"order_id" binding:"required"`
	PaymentKey string `json:"payment_key" binding:"required"` // wechat_jsapi, alipay_pc, offline 等
	OpenID     string `json:"openid"`
	ReturnURL  string `json:"return_url"`
	ClientIP   string `json:"-"`
}

func UnifiedPay(userID uint, req *UnifiedPayReq) (*PayDriverResp, error) {
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("订单状态不允许支付")
	}

	driver, err := GetPaymentDriver(req.PaymentKey)
	if err != nil {
		return nil, err
	}

	// 线下支付：直接标记为待确认（管理员确认后变为已支付）
	if req.PaymentKey == "offline" {
		now := time.Now()
		global.DB.Model(&order).Updates(map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now})
		AddOrderStatusHistory(order.ID, model.OrderStatusPending, model.OrderStatusPaid, "线下支付", "系统")
		return &PayDriverResp{TradeNo: "OFFLINE_" + order.OrderNo}, nil
	}

	// 钱包支付：扣余额
	if req.PaymentKey == "wallet" {
		var user model.User
		global.DB.First(&user, userID)
		if user.WalletBalance < order.PayAmount {
			return nil, fmt.Errorf("钱包余额不足(余额%d分)", user.WalletBalance)
		}
		tx := global.DB.Begin()
		tx.Model(&user).Update("wallet_balance", user.WalletBalance-order.PayAmount)
		tx.Create(&model.WalletLog{UserID: userID, Amount: -order.PayAmount, Balance: user.WalletBalance - order.PayAmount, Type: "pay", RefID: order.ID, Remark: "订单支付"})
		now := time.Now()
		tx.Model(&order).Updates(map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now})
		tx.Commit()
		AddOrderStatusHistory(order.ID, model.OrderStatusPending, model.OrderStatusPaid, "钱包支付", "系统")
		return &PayDriverResp{TradeNo: "WALLET_" + order.OrderNo}, nil
	}

	resp, err := driver.Pay(context.Background(), &PayDriverReq{
		OrderNo:     order.OrderNo,
		Description: "商城订单",
		Amount:      order.PayAmount,
		OpenID:      req.OpenID,
		ReturnURL:   req.ReturnURL,
		ClientIP:    req.ClientIP,
	})
	return resp, err
}

// ========== 支付宝签名辅助 ==========

func alipaySign(params map[string]string, privateKeyPEM string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k + "=" + params[k])
	}
	// RSA2-SHA256签名
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		// 尝试裸key包装
		privateKeyPEM = "-----BEGIN RSA PRIVATE KEY-----\n" + privateKeyPEM + "\n-----END RSA PRIVATE KEY-----"
		block, _ = pem.Decode([]byte(privateKeyPEM))
	}
	if block == nil {
		return ""
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS1
		key2, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return ""
		}
		key = key2
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(buf.String()))
	sig, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(sig)
}

func alipayEncode(params map[string]string) string {
	var buf strings.Builder
	first := true
	for k, v := range params {
		if !first {
			buf.WriteByte('&')
		}
		buf.WriteString(k + "=" + url.QueryEscape(v))
		first = false
	}
	return buf.String()
}
