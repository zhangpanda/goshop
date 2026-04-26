package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type TrackResult struct {
	ExpressName string      `json:"express_name"`
	ExpressNo   string      `json:"express_no"`
	Status      string      `json:"status"`
	Traces      []TrackInfo `json:"traces"`
}

// GetLogisticsTrack 查询物流轨迹
func GetLogisticsTrack(orderID uint) (*TrackResult, error) {
	var shipment model.Shipment
	if err := global.DB.Where("order_id = ?", orderID).First(&shipment).Error; err != nil {
		return nil, fmt.Errorf("未找到发货信息")
	}

	// 查快递编码
	var express model.Express
	global.DB.Where("name = ?", shipment.ExpressCompany).First(&express)
	code := express.Code
	if code == "" {
		code = "auto"
	}

	// 尝试从快递100查询（需要配置key）
	apiKey := GetConfig("logistics_api_key")
	if apiKey != "" {
		return queryKuaidi100(apiKey, code, shipment.ExpressNo, shipment.ExpressCompany)
	}

	// 无API key时返回基础信息
	return &TrackResult{
		ExpressName: shipment.ExpressCompany,
		ExpressNo:   shipment.ExpressNo,
		Status:      "暂无轨迹（请配置物流查询API Key）",
		Traces: []TrackInfo{
			{Time: shipment.CreatedAt.Format(time.DateTime), Context: fmt.Sprintf("已发货 %s %s", shipment.ExpressCompany, shipment.ExpressNo)},
		},
	}, nil
}

func queryKuaidi100(key, com, num, name string) (*TrackResult, error) {
	customer := GetConfig("logistics_api_customer")
	if customer == "" {
		customer = key
	}
	// 快递100 实时查询 API: POST https://poll.kuaidi100.com/poll/query.do
	paramObj, _ := json.Marshal(map[string]string{"com": com, "num": num})
	param := string(paramObj)
	sign := md5Sign(param + key + customer)
	form := url.Values{
		"customer": {customer},
		"sign":     {sign},
		"param":    {param},
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://poll.kuaidi100.com/poll/query.do", form)
	if err != nil {
		return &TrackResult{ExpressName: name, ExpressNo: num, Status: "查询失败"}, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		State   string `json:"state"`
		Message string `json:"message"`
		Data    []struct {
			Time    string `json:"time"`
			Context string `json:"context"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)

	traces := make([]TrackInfo, len(result.Data))
	for i, d := range result.Data {
		traces[i] = TrackInfo{Time: d.Time, Context: d.Context}
	}

	states := map[string]string{"0": "在途", "1": "揽收", "2": "疑难", "3": "签收", "4": "退签", "5": "派件", "6": "退回"}
	status := states[result.State]
	if status == "" {
		status = "查询中"
	}

	return &TrackResult{ExpressName: name, ExpressNo: num, Status: status, Traces: traces}, nil
}

func md5Sign(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}
