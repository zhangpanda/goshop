package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	net_http "net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
)

const alipayGatewayDoDefault = "https://openapi.alipay.com/gateway.do"

// AlipayGatewayDoURL 返回支付宝 OpenAPI 网关完整地址（以 /gateway.do 结尾）。
// 配置 alipay.gateway_url 可为沙箱（如 https://openapi.alipaydev.com/gateway.do）或自定义网关；仅填 origin 时自动补全 /gateway.do。
func AlipayGatewayDoURL() string {
	if app.Must().Cfg == nil {
		return alipayGatewayDoDefault
	}
	u := strings.TrimSpace(app.Must().Cfg.Alipay.GatewayURL)
	if u == "" {
		return alipayGatewayDoDefault
	}
	u = strings.TrimRight(u, "/")
	if strings.HasSuffix(u, "/gateway.do") {
		return u
	}
	return u + "/gateway.do"
}

// alipayPostGateway 对参数 RSA2 签名并以 application/x-www-form-urlencoded POST 到网关，返回响应体原文。
func alipayPostGateway(ctx context.Context, params map[string]string) ([]byte, error) {
	cfg := app.Must().Cfg.Alipay
	params["sign"] = alipaySign(params, cfg.PrivateKey)
	if params["sign"] == "" {
		return nil, errors.New("支付宝签名失败")
	}
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := net_http.NewRequestWithContext(ctx, net_http.MethodPost, AlipayGatewayDoURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&net_http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != net_http.StatusOK {
		return nil, fmt.Errorf("alipay http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

// alipayVerifyGatewayJSON 解析网关同步 JSON，校验根级 sign（RSA2），返回唯一的 *_response 键名与原文 RawMessage。
func alipayVerifyGatewayJSON(body []byte, publicKeyPEM string) (respKey string, rawResp json.RawMessage, err error) {
	if publicKeyPEM == "" {
		return "", nil, errors.New("支付宝公钥未配置")
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return "", nil, fmt.Errorf("alipay 响应解析失败: %w", err)
	}
	if st, ok := root["sign_type"]; ok {
		var stStr string
		_ = json.Unmarshal(st, &stStr)
		if stStr != "" && !strings.EqualFold(stStr, "RSA2") {
			return "", nil, fmt.Errorf("alipay 不支持的 sign_type: %s", stStr)
		}
	}
	var signStr string
	rawSign, ok := root["sign"]
	if !ok {
		return "", nil, errors.New("alipay 响应缺少 sign")
	}
	if err := json.Unmarshal(rawSign, &signStr); err != nil {
		return "", nil, fmt.Errorf("alipay sign 解析失败: %w", err)
	}
	if signStr == "" {
		return "", nil, errors.New("alipay 响应 sign 为空")
	}
	var respKeys []string
	for k := range root {
		if k == "sign" || k == "sign_type" {
			continue
		}
		if strings.HasSuffix(k, "_response") {
			respKeys = append(respKeys, k)
		}
	}
	sort.Strings(respKeys)
	if len(respKeys) != 1 {
		return "", nil, fmt.Errorf("alipay 响应需唯一 *_response，实际: %v", respKeys)
	}
	respKey = respKeys[0]
	rawResp = root[respKey]
	if !AlipayVerifyGatewaySyncSign(string(rawResp), signStr, publicKeyPEM) {
		return "", nil, errors.New("alipay 同步响应验签失败")
	}
	return respKey, rawResp, nil
}

// alipayOAuthExchange alipay.system.oauth.token（授权码换 user_id），用于小程序登录。
func alipayOAuthExchange(appID, authCode string) (userID, unionID string, err error) {
	cfg := app.Must().Cfg.Alipay
	if cfg.PrivateKey == "" || cfg.PublicKey == "" {
		return "", "", errors.New("支付宝未配置私钥或公钥")
	}
	if appID == "" || authCode == "" {
		return "", "", errors.New("支付宝 AppID 或授权码为空")
	}
	biz, err := json.Marshal(map[string]string{
		"grant_type": "authorization_code",
		"code":       authCode,
	})
	if err != nil {
		return "", "", err
	}
	params := map[string]string{
		"app_id":      appID,
		"method":      "alipay.system.oauth.token",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(biz),
	}
	body, err := alipayPostGateway(context.Background(), params)
	if err != nil {
		return "", "", err
	}
	key, raw, err := alipayVerifyGatewayJSON(body, cfg.PublicKey)
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
		_ = json.Unmarshal(raw, &er)
		return "", "", fmt.Errorf("alipay oauth: code=%s msg=%s %s %s", er.Code, er.Msg, er.SubCode, er.SubMsg)
	}
	if key != "alipay_system_oauth_token_response" {
		return "", "", fmt.Errorf("alipay oauth 非预期响应: %s", key)
	}
	var tr struct {
		UserID  string `json:"user_id"`
		OpenID  string `json:"open_id"`
		UnionID string `json:"unionid"`
		SubCode string `json:"sub_code"`
		SubMsg  string `json:"sub_msg"`
	}
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", "", fmt.Errorf("alipay oauth body: %w", err)
	}
	uid := tr.UserID
	if uid == "" {
		uid = tr.OpenID
	}
	if uid == "" {
		if tr.SubCode != "" {
			return "", "", fmt.Errorf("alipay oauth: %s %s", tr.SubCode, tr.SubMsg)
		}
		return "", "", errors.New("alipay oauth 未返回 user_id")
	}
	return uid, tr.UnionID, nil
}
