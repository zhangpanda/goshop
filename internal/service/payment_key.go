package service

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/zhangpanda/goshop/internal/model"
)

// InferPaymentKeyFromPaymentName 根据支付方式展示名称猜测内部 payment_key（与 ShopXO PHP 类名无关）。
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

// PaymentDriverKeyFromPayment 从配置 JSON 读取 payment_key；不含 ShopXO 的 payment 类名字段。
// /api.php 与 PHP 迁移数据请用 compat/shopxo.PaymentDriverKeyFromPayment。
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
		}
	}
	return InferPaymentKeyFromPaymentName(p.Name), nil
}
