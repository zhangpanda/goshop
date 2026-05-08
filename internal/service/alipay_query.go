package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
)

// AlipayTradePaid 判断支付宝 trade.query 返回的 trade_status 是否表示已支付。
func AlipayTradePaid(status string) bool {
	switch status {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return true
	default:
		return false
	}
}

// AlipayQueryTrade 调用官方 alipay.trade.query（商户订单号 out_trade_no）。同步响应经 RSA2 验签；网关地址见 alipay.gateway_url。
func AlipayQueryTrade(ctx context.Context, outTradeNo string) (tradeStatus, alipayTradeNo string, err error) {
	cfg := app.Must().Cfg.Alipay
	if cfg.AppID == "" || cfg.PrivateKey == "" || cfg.PublicKey == "" {
		return "", "", errors.New("支付宝未配置")
	}
	if strings.TrimSpace(outTradeNo) == "" {
		return "", "", errors.New("out_trade_no 为空")
	}
	biz, err := json.Marshal(map[string]string{"out_trade_no": outTradeNo})
	if err != nil {
		return "", "", err
	}
	params := map[string]string{
		"app_id":      cfg.AppID,
		"method":      "alipay.trade.query",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(biz),
	}
	body, err := alipayPostGateway(ctx, params)
	if err != nil {
		return "", "", err
	}
	key, rawResp, err := alipayVerifyGatewayJSON(body, cfg.PublicKey)
	if err != nil {
		return "", "", err
	}
	if key == "error_response" {
		var er struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			SubCode string `json:"sub_code"`
			SubMsg  string `json:"sub_msg"`
		}
		_ = json.Unmarshal(rawResp, &er)
		return "", "", fmt.Errorf("alipay: code=%s msg=%s %s %s", er.Code, er.Msg, er.SubCode, er.SubMsg)
	}
	if key != "alipay_trade_query_response" {
		return "", "", fmt.Errorf("alipay trade.query 非预期响应: %s", key)
	}
	var q struct {
		Code        string `json:"code"`
		Msg         string `json:"msg"`
		SubCode     string `json:"sub_code"`
		SubMsg      string `json:"sub_msg"`
		TradeStatus string `json:"trade_status"`
		TradeNo     string `json:"trade_no"`
	}
	if err := json.Unmarshal(rawResp, &q); err != nil {
		return "", "", fmt.Errorf("alipay trade.query body: %w", err)
	}
	if q.Code != "10000" {
		return "", "", fmt.Errorf("alipay trade.query: code=%s msg=%s %s %s", q.Code, q.Msg, q.SubCode, q.SubMsg)
	}
	return q.TradeStatus, q.TradeNo, nil
}
