package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// SendMessage 发送站内信
func SendMessage(userID uint, title, content, typ string, refID uint) error {
	return app.Must().DB.Create(&model.Message{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    typ,
		RefID:   refID,
	}).Error
}

func GetMessages(userID uint, page, pageSize int) ([]model.Message, int64, error) {
	var total int64
	app.Must().DB.Model(&model.Message{}).Where("user_id = ?", userID).Count(&total)
	var list []model.Message
	err := app.Must().DB.Where("user_id = ?", userID).
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func ReadMessage(userID, msgID uint) error {
	return app.Must().DB.Model(&model.Message{}).Where("id = ? AND user_id = ?", msgID, userID).
		Update("is_read", true).Error
}

func ReadAllMessages(userID uint) error {
	return app.Must().DB.Model(&model.Message{}).Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func UnreadCount(userID uint) int64 {
	var count int64
	app.Must().DB.Model(&model.Message{}).Where("user_id = ? AND is_read = false", userID).Count(&count)
	return count
}

// SendWxTemplateMsg 发送微信模板消息
func SendWxTemplateMsg(openID, templateID string, data map[string]interface{}, page string) error {
	cfg := app.Must().Cfg.Wechat
	if cfg.AppID == "" {
		return nil
	}

	// 获取 access_token
	tokenURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		cfg.AppID, cfg.AppSecret)
	tokenResp, err := http.Get(tokenURL)
	if err != nil {
		return err
	}
	defer tokenResp.Body.Close()
	var tokenResult struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(tokenResp.Body).Decode(&tokenResult)
	if tokenResult.AccessToken == "" {
		return fmt.Errorf("获取access_token失败")
	}

	// 发送模板消息
	body := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"page":        page,
		"data":        data,
	}
	bodyJSON, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", tokenResult.AccessToken)
	_, err = http.Post(url, "application/json", bytes.NewReader(bodyJSON))
	return err
}

// NotifyOrderStatus 订单状态变更通知（站内信+微信）
func NotifyOrderStatus(userID, orderID uint, orderNo, status string) {
	titles := map[string]string{
		"paid":      "订单支付成功",
		"shipped":   "订单已发货",
		"completed": "订单已完成",
		"refunded":  "退款成功",
	}
	title := titles[status]
	if title == "" {
		title = "订单状态更新"
	}
	content := fmt.Sprintf("您的订单 %s %s", orderNo, title)
	SendMessage(userID, title, content, "order", orderID)
}
