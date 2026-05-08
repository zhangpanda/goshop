package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
// 成功单笔更新后调用 postPaidHook 触发订单历史与通知（与合并支付路径 PayLogSuccess 行为一致）。
func HandlePayNotify(outTradeNo string, transactionID string) error {
	var order model.Order
	if err := app.Must().DB.Where("order_no = ?", outTradeNo).First(&order).Error; err == nil {
		if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusPending {
			// 已取消/已退款/已完成等：无须处理
			return nil
		}
		if order.Status == model.OrderStatusPaid {
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
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil // 并发竞争，另一个 goroutine 已处理
		}
		postPaidHook(order.ID, "支付成功", "系统")
		return nil
	}
	return PayLogSuccess(outTradeNo, transactionID)
}

type RefundReq struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Reason  string `json:"reason"`
}

// RefundOrder 申请退款：分三步
//  1. 事务内幂等预检 + 插入 RefundLog(status=0)。
//  2. 调支付驱动退款（支付驱动由 order.PaymentID 解析，兼容微信/支付宝/钱包/线下/PayPal）。
//  3. 第三方成功后事务化更新订单状态为 Refunded + 回库存 + 标记 RefundLog=1。
//
// 幂等：同一订单已有 status IN (0, 1) 的 RefundLog 时拒绝重复申请。
// 事务失败后 RefundLog 会留在 status=0（处理中），可由 reconcile cron 或人工处理。
func RefundOrder(userID uint, req *RefundReq) error {
	var order model.Order
	if err := app.Must().DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped {
		return errors.New("订单状态不允许退款")
	}

	// 幂等预检 + 插入处理中的 RefundLog
	refundNo := fmt.Sprintf("R%d-%d", order.ID, time.Now().UnixNano())
	rl := model.RefundLog{
		OrderID:     order.ID,
		UserID:      userID,
		RefundNo:    refundNo,
		RefundPrice: order.PayAmount,
		Reason:      req.Reason,
		Status:      0,
	}
	if err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&model.RefundLog{}).
			Where("order_id = ? AND status IN ?", order.ID, []int8{0, 1}).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errors.New("已有退款申请在处理中或已完成")
		}
		return tx.Create(&rl).Error
	}); err != nil {
		return err
	}

	// 解析支付驱动并调退款
	driverKey := resolveOrderPaymentKey(&order)
	driver, err := GetPaymentDriver(driverKey)
	if err != nil {
		markRefundFailed(&rl, err.Error())
		return err
	}
	if err := driver.Refund(context.Background(), &RefundDriverReq{
		OrderNo:  order.OrderNo,
		RefundNo: refundNo,
		Total:    order.PayAmount,
		Refund:   order.PayAmount,
		Reason:   req.Reason,
	}); err != nil {
		markRefundFailed(&rl, err.Error())
		return fmt.Errorf("退款失败: %w", err)
	}

	// 第三方成功：事务化更新订单 + 回库存 + 标记 RefundLog 成功
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		r := tx.Model(&model.Order{}).
			Where("id = ? AND status IN ?", order.ID, []int8{model.OrderStatusPaid, model.OrderStatusShipped}).
			Update("status", model.OrderStatusRefunded)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			// 第三方已退款但本地状态已变更（如并发）——记录并返回 nil，不让调用方重试
			slog.Warn("refund tx: order status changed", "order_id", order.ID, "refund_no", refundNo)
			return tx.Model(&rl).Updates(map[string]interface{}{"status": 1, "trade_no": order.TransactionID}).Error
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
		return tx.Model(&rl).Updates(map[string]interface{}{"status": 1, "trade_no": order.TransactionID}).Error
	})
}

// resolveOrderPaymentKey 从订单反查支付方式：order.PaymentID → Payment.Config.payment_key。
// 未绑定或解析失败时兜底微信 JSAPI（与仓库历史行为一致）。
func resolveOrderPaymentKey(order *model.Order) string {
	if order != nil && order.PaymentID > 0 {
		var p model.Payment
		if err := app.Must().DB.First(&p, order.PaymentID).Error; err == nil {
			if k, err := PaymentDriverKeyFromPayment(&p); err == nil {
				return k
			}
		}
	}
	return "wechat_jsapi"
}

// markRefundFailed 标记退款失败；出于幂等考虑允许多次调用。
func markRefundFailed(rl *model.RefundLog, reason string) {
	app.Must().DB.Model(rl).Updates(map[string]interface{}{
		"status": 2,
		"reason": truncateRefundReason(rl.Reason, reason),
	})
}

func truncateRefundReason(original, appendErr string) string {
	combined := original
	if appendErr != "" {
		if combined != "" {
			combined += " | "
		}
		combined += "ERR: " + appendErr
	}
	if len(combined) > 240 {
		return combined[:240]
	}
	return combined
}
