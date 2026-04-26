package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type ShipOrderReq struct {
	OrderID        uint   `json:"order_id" binding:"required"`
	ExpressCompany string `json:"express_company" binding:"required"`
	ExpressNo      string `json:"express_no" binding:"required"`
}

func ShipOrder(req *ShipOrderReq) error {
	var order model.Order
	if err := global.DB.First(&order, req.OrderID).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid {
		return errors.New("订单状态不允许发货")
	}

	tx := global.DB.Begin()
	now := time.Now()
	tx.Model(&order).Updates(map[string]interface{}{"status": model.OrderStatusShipped, "shipped_at": &now})
	tx.Create(&model.Shipment{
		OrderID:        req.OrderID,
		ExpressCompany: req.ExpressCompany,
		ExpressNo:      req.ExpressNo,
	})
	tx.Commit()
	return nil
}

func ConfirmReceive(userID, orderID uint) error {
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusShipped {
		return errors.New("订单状态不允许确认收货")
	}
	now := time.Now()
	if err := global.DB.Model(&order).Updates(map[string]interface{}{
		"status": model.OrderStatusCompleted, "completed_at": &now,
	}).Error; err != nil {
		return err
	}
	// 分销佣金结算
	go SettleCommission(orderID)
	return nil
}

func GetShipment(orderID uint) (*model.Shipment, error) {
	var s model.Shipment
	global.DB.Where("order_id = ?", orderID).Find(&s)
	if s.ID == 0 {
		return nil, errors.New("暂无物流信息")
	}
	return &s, nil
}

// TrackInfo 物流轨迹
type TrackInfo struct {
	Time    string `json:"time"`
	Context string `json:"context"`
}

// QueryExpress 查询物流轨迹（快递100 API）
func QueryExpress(company, no string) ([]TrackInfo, error) {
	url := fmt.Sprintf("https://www.kuaidi100.com/query?type=%s&postid=%s", company, no)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []TrackInfo `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data, nil
}
