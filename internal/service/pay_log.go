package service

import (
	"context"
	crypto_rand "crypto/rand"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

/**
 * uniqueUints 对订单 ID 去重并保持首次出现顺序。
 */
func uniqueUints(orderIDs []uint) []uint {
	seen := make(map[uint]struct{}, len(orderIDs))
	out := make([]uint, 0, len(orderIDs))
	for _, id := range orderIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

/**
 * MultiOrderUnifiedPay 多笔待支付订单统一支付：线下/钱包在事务内批量更新；微信/支付宝先 CreatePayLog，out_trade_no 为 PayLog.pay_no。
 */
func MultiOrderUnifiedPay(userID uint, orderIDs []uint, paymentKey string, paymentRecordID uint, openID, returnURL, clientIP string) (*PayDriverResp, error) {
	ids := uniqueUints(orderIDs)
	if len(ids) == 0 {
		return nil, fmt.Errorf("请选择订单")
	}
	var orders []model.Order
	if err := global.DB.Where("id IN ? AND user_id = ?", ids, userID).Order("id ASC").Find(&orders).Error; err != nil {
		return nil, err
	}
	if len(orders) != len(ids) {
		return nil, fmt.Errorf("部分订单不存在")
	}
	for i := range orders {
		if orders[i].Status != model.OrderStatusPending {
			return nil, fmt.Errorf("订单状态不允许支付")
		}
	}

	if paymentRecordID > 0 {
		global.DB.Model(&model.Order{}).Where("id IN ?", ids).Update("payment_id", paymentRecordID)
	}

	driver, err := GetPaymentDriver(paymentKey)
	if err != nil {
		return nil, err
	}

	if paymentKey == "offline" {
		now := time.Now()
		upd := map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now}
		tx := global.DB.Begin()
		if err := tx.Error; err != nil {
			return nil, err
		}
		var paidIDs []uint
		for _, o := range orders {
			r := tx.Model(&model.Order{}).Where("id = ? AND status = ?", o.ID, model.OrderStatusPending).Updates(upd)
			if r.Error != nil {
				tx.Rollback()
				return nil, r.Error
			}
			if r.RowsAffected > 0 {
				paidIDs = append(paidIDs, o.ID)
			}
		}
		if len(paidIDs) != len(orders) {
			tx.Rollback()
			return nil, fmt.Errorf("部分订单状态已变更，请刷新后重试")
		}
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		for _, oid := range paidIDs {
			AddOrderStatusHistory(oid, model.OrderStatusPending, model.OrderStatusPaid, "线下支付", "系统")
		}
		return &PayDriverResp{TradeNo: fmt.Sprintf("OFFLINE_MULTI_%d", ids[0])}, nil
	}

	if paymentKey == "wallet" {
		var total int64
		for _, o := range orders {
			total += o.PayAmount
		}
		tx := global.DB.Begin()
		if err := tx.Error; err != nil {
			return nil, err
		}
		result := tx.Model(&model.User{}).Where("id = ? AND wallet_balance >= ?", userID, total).
			Update("wallet_balance", gorm.Expr("wallet_balance - ?", total))
		if result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("钱包余额不足")
		}
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Create(&model.WalletLog{
			UserID: userID, Amount: -total, Balance: user.WalletBalance, Type: "pay",
			RefID: 0, Remark: fmt.Sprintf("合并订单支付(%d笔)", len(orders)),
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		now := time.Now()
		wupd := map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now}
		var paidIDs []uint
		for _, o := range orders {
			r := tx.Model(&model.Order{}).Where("id = ? AND status = ?", o.ID, model.OrderStatusPending).Updates(wupd)
			if r.Error != nil {
				tx.Rollback()
				return nil, r.Error
			}
			if r.RowsAffected > 0 {
				paidIDs = append(paidIDs, o.ID)
			}
		}
		if len(paidIDs) != len(orders) {
			tx.Rollback()
			return nil, fmt.Errorf("部分订单状态已变更，请刷新后重试")
		}
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("支付失败: %w", err)
		}
		for _, oid := range paidIDs {
			AddOrderStatusHistory(oid, model.OrderStatusPending, model.OrderStatusPaid, "钱包支付", "系统")
		}
		return &PayDriverResp{TradeNo: fmt.Sprintf("WALLET_MULTI_%d", time.Now().UnixNano())}, nil
	}

	// clientType 写入 PayLog；兼容层用 "shopxo" 仅作来源标记，非第三方商标用法
	pl, err := CreatePayLog(userID, ids, paymentRecordID, "shopxo")
	if err != nil {
		return nil, err
	}
	resp, err := driver.Pay(context.Background(), &PayDriverReq{
		OrderNo:     pl.PayNo,
		Description: "商城订单",
		Amount:      pl.TotalPrice,
		OpenID:      openID,
		ReturnURL:   returnURL,
		ClientIP:    clientIP,
	})
	return resp, err
}

func generatePayNo() string {
	b := make([]byte, 4)
	crypto_rand.Read(b)
	return fmt.Sprintf("P%s%08x", time.Now().Format("20060102150405"), b)
}

// CreatePayLog 创建支付日志（支持合并支付多个订单）。clientType 为渠道/来源标识（如 api、shopxo）。
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
	if err := global.DB.Create(&pl).Error; err != nil {
		return nil, err
	}
	return &pl, nil
}

// PayLogSuccess 支付成功回调处理
func PayLogSuccess(payNo, tradeNo string) error {
	var pl model.PayLog
	if err := global.DB.Where("pay_no = ?", payNo).First(&pl).Error; err != nil {
		return fmt.Errorf("支付日志不存在")
	}
	if pl.Status != 0 {
		return nil // 幂等
	}

	now := time.Now()

	// 整段事务：PayLog 标记 + 所有子订单更新，要么全成功要么全回滚
	tx := global.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	res := tx.Model(&model.PayLog{}).Where("id = ? AND status = ?", pl.ID, 0).
		Updates(map[string]interface{}{"status": 1, "trade_no": tradeNo, "paid_at": &now})
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil // 并发下另一路已标记为已支付
	}

	// 更新关联订单状态
	var paidIDs []uint
	orderIDCount := 0
	for _, idStr := range strings.Split(pl.OrderIDs, ",") {
		var oid uint
		fmt.Sscanf(idStr, "%d", &oid)
		if oid == 0 {
			continue
		}
		orderIDCount++
		upd := map[string]interface{}{
			"status": model.OrderStatusPaid, "paid_at": &now,
			"transaction_id": tradeNo,
		}
		if pl.PaymentID > 0 {
			upd["payment_id"] = pl.PaymentID
		}
		ores := tx.Model(&model.Order{}).Where("id = ? AND status = ?", oid, model.OrderStatusPending).Updates(upd)
		if ores.Error != nil {
			tx.Rollback()
			return ores.Error
		}
		if ores.RowsAffected > 0 {
			paidIDs = append(paidIDs, oid)
		}
	}

	if orderIDCount == 0 {
		tx.Rollback()
		slog.Warn("pay callback", "reason", "pay_log_no_valid_order_ids", "pay_no", payNo)
		return nil
	}
	if len(paidIDs) == 0 {
		tx.Rollback()
		// 渠道重复通知或订单已通过其他路径变为已支付：若子单已全部已付，仅补 PayLog；否则告警并返回成功以免第三方无限重试
		var oids []uint
		for _, idStr := range strings.Split(pl.OrderIDs, ",") {
			var oid uint
			fmt.Sscanf(idStr, "%d", &oid)
			if oid > 0 {
				oids = append(oids, oid)
			}
		}
		allPaid := true
		for _, oid := range oids {
			var o model.Order
			if err := global.DB.First(&o, oid).Error; err != nil || o.Status != model.OrderStatusPaid {
				allPaid = false
				break
			}
		}
		if allPaid {
			res2 := global.DB.Model(&model.PayLog{}).Where("id = ? AND status = ?", pl.ID, 0).
				Updates(map[string]interface{}{"status": 1, "trade_no": tradeNo, "paid_at": &now})
			return res2.Error
		}
		slog.Warn("pay callback", "reason", "merge_pay_no_pending_and_not_all_paid", "pay_no", payNo)
		return nil
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 事务提交后再写历史和通知（非关键路径，失败不影响支付结果）
	for _, oid := range paidIDs {
		AddOrderStatusHistory(oid, model.OrderStatusPending, model.OrderStatusPaid, "支付成功", "系统")
		var order model.Order
		if err := global.DB.First(&order, oid).Error; err != nil {
			slog.Warn("pay callback", "reason", "notify_skip_order_load", "order_id", oid, "err", err)
			continue
		}
		NotifyOrderStatus(order.UserID, oid, order.OrderNo, "paid")
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
