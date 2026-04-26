package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/wechat"
)

type PayOrderReq struct {
	OrderID uint   `json:"order_id" binding:"required"`
	OpenID  string `json:"openid" binding:"required"`
}

// PayOrder 发起支付，返回小程序调起支付的参数
func PayOrder(userID uint, req *PayOrderReq) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	if global.WxPay == nil {
		return nil, errors.New("微信支付未配置")
	}

	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("订单状态不允许支付")
	}

	resp, err := global.WxPay.Prepay(context.Background(), &wechat.PrepayRequest{
		OrderNo:     order.OrderNo,
		Description: "商城订单",
		Amount:      order.PayAmount,
		OpenID:      req.OpenID,
	})
	if err != nil {
		return nil, fmt.Errorf("预下单失败: %w", err)
	}
	return resp, nil
}

// HandlePayNotify 处理支付回调：先按订单号单笔；否则按 PayLog.pay_no（合并支付）
func HandlePayNotify(outTradeNo string, transactionID string) error {
	var order model.Order
	if err := global.DB.Where("order_no = ?", outTradeNo).First(&order).Error; err == nil {
		if order.Status != model.OrderStatusPending {
			return nil
		}
		now := time.Now()
		return global.DB.Model(&order).Updates(map[string]interface{}{
			"status":  model.OrderStatusPaid,
			"paid_at": &now,
		}).Error
	}
	return PayLogSuccess(outTradeNo, transactionID)
}

type RefundReq struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Reason  string `json:"reason"`
}

func RefundOrder(userID uint, req *RefundReq) error {
	if global.WxPay == nil {
		return errors.New("微信支付未配置")
	}

	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped {
		return errors.New("订单状态不允许退款")
	}

	refundNo := fmt.Sprintf("R%d", time.Now().UnixNano())
	_, err := global.WxPay.Refund(context.Background(), &wechat.RefundRequest{
		OrderNo:  order.OrderNo,
		RefundNo: refundNo,
		Total:    order.PayAmount,
		Refund:   order.PayAmount,
		Reason:   req.Reason,
	})
	if err != nil {
		return fmt.Errorf("退款失败: %w", err)
	}

	// 恢复库存
	tx := global.DB.Begin()
	tx.Model(&order).Update("status", model.OrderStatusRefunded)
	var items []model.OrderItem
	tx.Where("order_id = ?", order.ID).Find(&items)
	for _, item := range items {
		tx.Model(&model.GoodsSKU{}).Where("id = ?", item.SKUID).
			Update("stock", global.DB.Raw("stock + ?", item.Quantity))
	}
	tx.Commit()

	return nil
}
