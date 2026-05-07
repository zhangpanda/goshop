package service

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
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
	"gorm.io/gorm"
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
	bizContent, _ := json.Marshal(map[string]interface{}{
		"out_trade_no": req.OrderNo,
		"total_amount": fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"subject":      req.Description,
		"product_code": d.productCode(),
	})
	params := map[string]string{
		"app_id":      cfg.AppID,
		"method":      d.method(),
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  cfg.NotifyURL,
		"biz_content": string(bizContent),
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
	refundBiz, _ := json.Marshal(map[string]interface{}{
		"out_trade_no":   req.OrderNo,
		"refund_amount":  fmt.Sprintf("%.2f", float64(req.Refund)/100),
		"out_request_no": req.RefundNo,
		"refund_reason":  req.Reason,
	})
	params := map[string]string{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.refund",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(refundBiz),
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
	if global.Cfg != nil && global.Cfg.Payment.Sandbox {
		return &SandboxDriver{Name: name, Real: d}, nil
	}
	return d, nil
}

// ========== 沙盒驱动 ==========

// SandboxDriver 包装真实驱动：走完参数构造和签名，但不依赖第三方返回
type SandboxDriver struct {
	Name string
	Real PaymentDriver
}

func (d *SandboxDriver) Pay(ctx context.Context, req *PayDriverReq) (*PayDriverResp, error) {
	tradeNo := fmt.Sprintf("SANDBOX_%s_%d", req.OrderNo, time.Now().UnixMilli())
	callbackURL := fmt.Sprintf("/api/pay/sandbox/callback?order_no=%s&trade_no=%s", req.OrderNo, tradeNo)

	// 尝试调用真实驱动的 Pay（验证参数构造+签名逻辑）
	// 密钥未配置等情况会报错或 panic，沙盒模式下均忽略
	realResp := d.tryRealPay(ctx, req)

	resp := &PayDriverResp{
		TradeNo: tradeNo,
		PayURL:  callbackURL,
	}
	if realResp != nil {
		resp.PrepayData = realResp.PrepayData
		resp.QRCode = realResp.QRCode
	}
	return resp, nil
}

func (d *SandboxDriver) tryRealPay(ctx context.Context, req *PayDriverReq) (resp *PayDriverResp) {
	defer func() { recover() }() //nolint // 沙盒容忍 panic（如 nil config）
	resp, _ = d.Real.Pay(ctx, req)
	return
}

func (d *SandboxDriver) Refund(ctx context.Context, req *RefundDriverReq) error {
	defer func() { recover() }() //nolint
	d.Real.Refund(ctx, req)      //nolint:errcheck
	return nil
}

// ========== 统一支付入口 ==========

type UnifiedPayReq struct {
	OrderID         uint   `json:"order_id" binding:"required"`
	PaymentKey      string `json:"payment_key" binding:"required"` // wechat_jsapi, alipay_pc, offline 等
	OpenID          string `json:"openid"`
	ReturnURL       string `json:"return_url"`
	ClientIP        string `json:"-"`
	PaymentRecordID uint   `json:"payment_id"` // 支付方式表主键，回写订单
}

func UnifiedPay(userID uint, req *UnifiedPayReq) (*PayDriverResp, error) {
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("订单状态不允许支付")
	}
	if req.PaymentRecordID > 0 {
		global.DB.Model(&order).Update("payment_id", req.PaymentRecordID)
	}

	driver, err := GetPaymentDriver(req.PaymentKey)
	if err != nil {
		return nil, err
	}

	// 线下支付：单笔订单与钱包一致，事务内乐观更新 pending→已付
	if req.PaymentKey == "offline" {
		now := time.Now()
		upd := map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now}
		if req.PaymentRecordID > 0 {
			upd["payment_id"] = req.PaymentRecordID
		}
		if err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
			r := tx.Model(&model.Order{}).Where("id = ? AND user_id = ? AND status = ?", order.ID, userID, model.OrderStatusPending).Updates(upd)
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected == 0 {
				return fmt.Errorf("订单状态已变更，请刷新后重试")
			}
			return nil
		}); err != nil {
			return nil, err
		}
		AddOrderStatusHistory(order.ID, model.OrderStatusPending, model.OrderStatusPaid, "线下支付", "系统")
		return &PayDriverResp{TradeNo: "OFFLINE_" + order.OrderNo}, nil
	}

	// 钱包支付：原子扣余额
	if req.PaymentKey == "wallet" {
		err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
			result := tx.Model(&model.User{}).Where("id = ? AND wallet_balance >= ?", userID, order.PayAmount).
				Update("wallet_balance", gorm.Expr("wallet_balance - ?", order.PayAmount))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("钱包余额不足")
			}
			var user model.User
			if err := tx.First(&user, userID).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WalletLog{UserID: userID, Amount: -order.PayAmount, Balance: user.WalletBalance, Type: "pay", RefID: order.ID, Remark: "订单支付"}).Error; err != nil {
				return err
			}
			now := time.Now()
			wupd := map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now}
			if req.PaymentRecordID > 0 {
				wupd["payment_id"] = req.PaymentRecordID
			}
			r := tx.Model(&model.Order{}).Where("id = ? AND user_id = ? AND status = ?", order.ID, userID, model.OrderStatusPending).Updates(wupd)
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected == 0 {
				return fmt.Errorf("订单状态已变更，请刷新后重试")
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("支付失败: %w", err)
		}
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

// AlipayVerifySign 验证支付宝回调签名（RSA2-SHA256）
func AlipayVerifySign(params map[string]string, publicKeyPEM string) bool {
	if publicKeyPEM == "" {
		return false
	}
	sign, ok := params["sign"]
	if !ok || sign == "" {
		return false
	}
	// 构造待验签字符串（排除 sign 和 sign_type）
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
	// 解码签名
	sigBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	// 解析公钥
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		publicKeyPEM = "-----BEGIN PUBLIC KEY-----\n" + publicKeyPEM + "\n-----END PUBLIC KEY-----"
		block, _ = pem.Decode([]byte(publicKeyPEM))
	}
	if block == nil {
		return false
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false
	}
	h := sha256.New()
	h.Write([]byte(buf.String()))
	return rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, h.Sum(nil), sigBytes) == nil
}
