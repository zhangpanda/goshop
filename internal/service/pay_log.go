package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

func generatePayNo() string {
	return fmt.Sprintf("P%d", time.Now().UnixNano())
}

// CreatePayLog 创建支付日志（支持合并支付多个订单）
func CreatePayLog(userID uint, orderIDs []uint, paymentID uint, clientType string) (*model.PayLog, error) {
	var totalPrice int64
	ids := make([]string, len(orderIDs))
	for i, oid := range orderIDs {
		var order model.Order
		if err := global.DB.First(&order, oid).Error; err != nil {
			return nil, fmt.Errorf("订单%d不存在", oid)
		}
		if order.UserID != userID || order.Status != model.OrderStatusPending {
			return nil, fmt.Errorf("订单%d状态不允许支付", oid)
		}
		totalPrice += order.PayAmount
		ids[i] = fmt.Sprintf("%d", oid)
	}

	pl := model.PayLog{
		PayNo:      generatePayNo(),
		OrderIDs:   strings.Join(ids, ","),
		UserID:     userID,
		PaymentID:  paymentID,
		TotalPrice: totalPrice,
		ClientType: clientType,
	}
	global.DB.Create(&pl)
	return &pl, nil
}

// PayLogSuccess 支付成功回调处理
func PayLogSuccess(payNo, tradeNo string) error {
	var pl model.PayLog
	global.DB.Where("pay_no = ?", payNo).Find(&pl)
	if pl.ID == 0 {
		return fmt.Errorf("支付日志不存在")
	}
	if pl.Status != 0 {
		return nil // 幂等
	}

	now := time.Now()
	global.DB.Model(&pl).Updates(map[string]interface{}{"status": 1, "trade_no": tradeNo, "paid_at": &now})

	// 更新关联订单状态
	for _, idStr := range strings.Split(pl.OrderIDs, ",") {
		var oid uint
		fmt.Sscanf(idStr, "%d", &oid)
		if oid > 0 {
			global.DB.Model(&model.Order{}).Where("id = ? AND status = ?", oid, model.OrderStatusPending).
				Updates(map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now})
			AddOrderStatusHistory(oid, model.OrderStatusPending, model.OrderStatusPaid, "支付成功", "系统")
			// 发消息
			var order model.Order
			global.DB.First(&order, oid)
			NotifyOrderStatus(order.UserID, oid, order.OrderNo, "paid")
		}
	}
	return nil
}

// CreateRefundLog 创建退款日志
func CreateRefundLog(orderID, payLogID, userID uint, refundPrice int64, reason string) *model.RefundLog {
	rl := model.RefundLog{
		OrderID:     orderID,
		PayLogID:    payLogID,
		UserID:      userID,
		RefundNo:    fmt.Sprintf("R%d", time.Now().UnixNano()),
		RefundPrice: refundPrice,
		Reason:      reason,
	}
	global.DB.Create(&rl)
	return &rl
}

func UpdateRefundLog(refundNo string, status int8, tradeNo string) {
	global.DB.Model(&model.RefundLog{}).Where("refund_no = ?", refundNo).
		Updates(map[string]interface{}{"status": status, "trade_no": tradeNo})
}

func GetPayLogList(userID uint, page, pageSize int) ([]model.PayLog, int64, error) {
	var total int64
	global.DB.Model(&model.PayLog{}).Where("user_id = ?", userID).Count(&total)
	var list []model.PayLog
	err := global.DB.Where("user_id = ?", userID).Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func GetRefundLogList(page, pageSize int) ([]model.RefundLog, int64, error) {
	var total int64
	global.DB.Model(&model.RefundLog{}).Count(&total)
	var list []model.RefundLog
	err := global.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func AddPayRequestLog(payLogID uint, request, response, business string) {
	global.DB.Create(&model.PayRequestLog{PayLogID: payLogID, Request: request, Response: response, Business: business})
}
