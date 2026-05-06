package service

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/zhangpanda/goshop/internal/model"
)

// PaymentDriverKeyFromPayment 从支付方式配置解析 driver key。
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
				return ShopxoPHPClassToDriverKey(strings.TrimSpace(s)), nil
			}
		}
	}
	return InferPaymentKeyFromPaymentName(p.Name), nil
}

// ShopxoPHPClassToDriverKey maps ShopXO PHP payment class names to driver keys.
func ShopxoPHPClassToDriverKey(class string) string {
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
		return InferPaymentKeyFromPaymentName(class)
	}
}

// InferPaymentKeyFromPaymentName guesses driver key from payment name.
func InferPaymentKeyFromPaymentName(name string) string {
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
