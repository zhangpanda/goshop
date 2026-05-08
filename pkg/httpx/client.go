// Package httpx 提供带超时的共享 *http.Client，用于调用第三方服务
// （微信、支付宝、快递 100 等），避免裸 http.Get/Post 默认无限等。
//
// 使用：
//
//	resp, err := httpx.Client.Get(url)
//	resp, err := httpx.Client.Post(url, "application/json", body)
//
// 如需自定义超时，调用方可自行构造 *http.Client；此包只提供默认值。
package httpx

import (
	"net/http"
	"time"
)

// Client 默认 HTTP 客户端，10 秒整体超时（含连接+TLS+读取）。
var Client = &http.Client{Timeout: 10 * time.Second}
