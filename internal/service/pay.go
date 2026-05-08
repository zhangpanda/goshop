package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/wechat"
	"gorm.io/gorm"
)

type PayOrderReq struct {
	OrderID uint   `json:"order_id" binding:"required"`
	OpenID  string `json:"openid" binding:"required"`
}

// PayOrder 发起支付，返回小程序调起支付的参数
func PayOrder(userID uint, req *PayOrderReq) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	if app.Must().WxPay == nil {
		return nil, errors.New("微信支付未配置")
	}

	var order model.Order
	if err := app.Must().DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return nil, errors.New("订单状态不允许支付")
	}

	resp, err := app.Must().WxPay.Prepay(context.Background(), &wechat.PrepayRequest{
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

// HandlePayNotify 处理支付回调：先按订单号单笔；否则按 PayLog.pay_no（合并支付）。
// 幂等保护：通过 WHERE status = pending 条件更新，并发重复回调只有一个能成功。
func HandlePayNotify(outTradeNo string, transactionID string) error {
	var order model.Order
	if err := app.Must().DB.Where("order_no = ?", outTradeNo).First(&order).Error; err == nil {
		if order.Status != model.OrderStatusPending {
			return nil // 已处理，幂等返回
		}
		now := time.Now()
		result := app.Must().DB.Model(&model.Order{}).
			Where("id = ? AND status = ?", order.ID, model.OrderStatusPending).
			Updates(map[string]interface{}{
				"status":         model.OrderStatusPaid,
				"paid_at":        &now,
				"transaction_id": transactionID,
			})
		if result.RowsAffected == 0 {
			return nil // 并发竞争，另一个 goroutine 已处理
		}
		return result.Error
	}
	return PayLogSuccess(outTradeNo, transactionID)
}

type RefundReq struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Reason  string `json:"reason"`
}

func RefundOrder(userID uint, req *RefundReq) error {
	if app.Must().WxPay == nil {
		return errors.New("微信支付未配置")
	}

	var order model.Order
	if err := app.Must().DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped {
		return errors.New("订单状态不允许退款")
	}

	refundNo := fmt.Sprintf("R%d", time.Now().UnixNano())
	_, err := app.Must().WxPay.Refund(context.Background(), &wechat.RefundRequest{
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
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Model(&order).Update("status", model.OrderStatusRefunded).Error; err != nil {
			return err
		}
		var items []model.OrderItem
		if err := tx.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			if err := tx.Model(&model.GoodsSKU{}).Where("id = ?", item.SKUID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
