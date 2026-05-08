package shopxo

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
)

// ShopxoPHPClassToDriverKey 将 ShopXO / shopxo-uniapp 常见 PHP 支付类名映射为内部 payment_key。
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
		return service.InferPaymentKeyFromPaymentName(class)
	}
}

// PaymentDriverKeyFromPayment 解析支付方式：payment_key → ShopXO 的 payment 类名 → 名称启发式（对标 PHP 后台/迁移数据）。
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
	return service.InferPaymentKeyFromPaymentName(p.Name), nil
}
