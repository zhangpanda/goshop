package service

import (
	"context"
	crypto_rand "crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// PayLogSuccess 内部：PayLog 已更新但子单均未处于待支付时的可恢复分支（整单事务将回滚 PayLog 更新）。
var errPayLogNoPendingOrdersUpdated = errors.New("goshop: merge pay no pending order rows updated")

// PayLogSuccess 内部：OrderIDs 解析后无有效订单 ID。
var errPayLogNoValidOrderIDs = errors.New("goshop: pay_log has no valid order ids")

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

// applyMultiOrderPaymentID 合并支付时将支付方式写入子单；应在同一事务内与后续扣款/建表一起调用。
func applyMultiOrderPaymentID(tx *gorm.DB, userID uint, ids []uint, paymentRecordID uint) error {
	if paymentRecordID == 0 {
		return nil
	}
	if err := tx.Model(&model.Order{}).Where("id IN ? AND user_id = ? AND status = ?", ids, userID, model.OrderStatusPending).
		Update("payment_id", paymentRecordID).Error; err != nil {
		return err
	}
	var n int64
	if err := tx.Model(&model.Order{}).Where("id IN ? AND user_id = ? AND status = ? AND payment_id = ?", ids, userID, model.OrderStatusPending, paymentRecordID).
		Count(&n).Error; err != nil {
		return err
	}
	if int(n) != len(ids) {
		return fmt.Errorf("合并支付更新支付方式失败，请刷新后重试")
	}
	return nil
}

// parsePayLogOrderIDList 解析 PayLog.order_ids 逗号分隔的订单 ID。
func parsePayLogOrderIDList(csv string) []uint {
	var out []uint
	for _, idStr := range strings.Split(csv, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		var oid uint
		_, _ = fmt.Sscanf(idStr, "%d", &oid)
		if oid > 0 {
			out = append(out, oid)
		}
	}
	return out
}

// revertMergePayThirdPartyPrep 合并支付在第三方预下单失败时：关闭待支付 PayLog，并清零仍为待付款子单的 payment_id，便于用户重试。
func revertMergePayThirdPartyPrep(pl *model.PayLog, userID uint, orderIDs []uint) {
	if pl == nil || pl.ID == 0 {
		return
	}
	err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&model.PayLog{}).Where("id = ? AND user_id = ? AND status = ?", pl.ID, userID, 0).
			Updates(map[string]interface{}{"status": 2, "closed_at": &now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		for _, oid := range orderIDs {
			if err := tx.Model(&model.Order{}).Where("id = ? AND user_id = ? AND status = ?", oid, userID, model.OrderStatusPending).
				Update("payment_id", 0).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Warn("pay", "action", "revert_merge_prep_failed", "pay_log_id", pl.ID, "err", err.Error())
	}
}

// StalePendingPayLogsCleanup 关闭长时间仍处于待支付的 PayLog，并解除待付款子单的 payment_id（合并支付半开态兜底）。
func StalePendingPayLogsCleanup(deps *app.Deps, maxAgeMinutes int) (closed int) {
	if maxAgeMinutes <= 0 {
		maxAgeMinutes = 120
	}
	deadline := time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute)
	var logs []model.PayLog
	if err := deps.DB.Where("status = ? AND created_at < ?", 0, deadline).Find(&logs).Error; err != nil {
		slog.Warn("pay", "action", "stale_paylog_query", "err", err.Error())
		return 0
	}
	for i := range logs {
		pl := &logs[i]
		ids := parsePayLogOrderIDList(pl.OrderIDs)
		var rowsAff int64
		err := RunInDBTx(deps.DB, func(tx *gorm.DB) error {
			now := time.Now()
			res := tx.Model(&model.PayLog{}).Where("id = ? AND status = ?", pl.ID, 0).
				Updates(map[string]interface{}{"status": 2, "closed_at": &now})
			if res.Error != nil {
				return res.Error
			}
			rowsAff = res.RowsAffected
			if rowsAff == 0 {
				return nil
			}
			for _, oid := range ids {
				if err := tx.Model(&model.Order{}).Where("id = ? AND user_id = ? AND status = ?", oid, pl.UserID, model.OrderStatusPending).
					Update("payment_id", 0).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			slog.Warn("pay", "action", "stale_paylog_cleanup_fail", "pay_log_id", pl.ID, "err", err.Error())
			continue
		}
		if rowsAff > 0 {
			closed++
		}
	}
	return closed
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
	if err := app.Must().DB.Where("id IN ? AND user_id = ?", ids, userID).Order("id ASC").Find(&orders).Error; err != nil {
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

	driver, err := GetPaymentDriver(paymentKey)
	if err != nil {
		return nil, err
	}

	if paymentKey == "offline" {
		var paidIDs []uint
		err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
			if err := applyMultiOrderPaymentID(tx, userID, ids, paymentRecordID); err != nil {
				return err
			}
			now := time.Now()
			upd := map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now}
			for _, o := range orders {
				r := tx.Model(&model.Order{}).Where("id = ? AND status = ?", o.ID, model.OrderStatusPending).Updates(upd)
				if r.Error != nil {
					return r.Error
				}
				if r.RowsAffected > 0 {
					paidIDs = append(paidIDs, o.ID)
				}
			}
			if len(paidIDs) != len(orders) {
				return fmt.Errorf("部分订单状态已变更，请刷新后重试")
			}
			return nil
		})
		if err != nil {
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
		var paidIDs []uint
		err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
			if err := applyMultiOrderPaymentID(tx, userID, ids, paymentRecordID); err != nil {
				return err
			}
			result := tx.Model(&model.User{}).Where("id = ? AND wallet_balance >= ?", userID, total).
				Update("wallet_balance", gorm.Expr("wallet_balance - ?", total))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("钱包余额不足")
			}
			var user model.User
			if err := tx.First(&user, userID).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WalletLog{
				UserID: userID, Amount: -total, Balance: user.WalletBalance, Type: "pay",
				RefID: 0, Remark: fmt.Sprintf("合并订单支付(%d笔)", len(orders)),
			}).Error; err != nil {
				return err
			}
			now := time.Now()
			wupd := map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now}
			for _, o := range orders {
				r := tx.Model(&model.Order{}).Where("id = ? AND status = ?", o.ID, model.OrderStatusPending).Updates(wupd)
				if r.Error != nil {
					return r.Error
				}
				if r.RowsAffected > 0 {
					paidIDs = append(paidIDs, o.ID)
				}
			}
			if len(paidIDs) != len(orders) {
				return fmt.Errorf("部分订单状态已变更，请刷新后重试")
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("支付失败: %w", err)
		}
		for _, oid := range paidIDs {
			AddOrderStatusHistory(oid, model.OrderStatusPending, model.OrderStatusPaid, "钱包支付", "系统")
		}
		return &PayDriverResp{TradeNo: fmt.Sprintf("WALLET_MULTI_%d", time.Now().UnixNano())}, nil
	}

	// clientType 写入 PayLog；兼容层用 "shopxo" 仅作来源标记，非第三方商标用法
	var pl *model.PayLog
	err = RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := applyMultiOrderPaymentID(tx, userID, ids, paymentRecordID); err != nil {
			return err
		}
		var cerr error
		pl, cerr = createPayLogWithDB(tx, userID, ids, paymentRecordID, "shopxo")
		return cerr
	})
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
	if err != nil {
		revertMergePayThirdPartyPrep(pl, userID, ids)
		return nil, err
	}
	return resp, nil
}

func generatePayNo() string {
	b := make([]byte, 4)
	crypto_rand.Read(b)
	return fmt.Sprintf("P%s%08x", time.Now().Format("20060102150405"), b)
}

// CreatePayLog 创建支付日志（支持合并支付多个订单）。clientType 为渠道/来源标识（如 api、shopxo）。
func CreatePayLog(userID uint, orderIDs []uint, paymentID uint, clientType string) (*model.PayLog, error) {
	return createPayLogWithDB(app.Must().DB, userID, orderIDs, paymentID, clientType)
}

// createPayLogWithDB 使用指定 DB/事务句柄创建 PayLog，读单与写入在同一连接上便于与合并支付事务组合。
func createPayLogWithDB(db *gorm.DB, userID uint, orderIDs []uint, paymentID uint, clientType string) (*model.PayLog, error) {
	var totalPrice int64
	ids := make([]string, len(orderIDs))
	for i, oid := range orderIDs {
		var order model.Order
		if err := db.First(&order, oid).Error; err != nil {
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
	if err := db.Create(&pl).Error; err != nil {
		return nil, err
	}
	return &pl, nil
}

// PayLogSuccess 支付成功回调处理
func PayLogSuccess(payNo, tradeNo string) error {
	var pl model.PayLog
	if err := app.Must().DB.Where("pay_no = ?", payNo).First(&pl).Error; err != nil {
		return fmt.Errorf("支付日志不存在")
	}
	if pl.Status != 0 {
		return nil // 幂等
	}

	now := time.Now()

	var paidIDs []uint
	err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		res := tx.Model(&model.PayLog{}).Where("id = ? AND status = ?", pl.ID, 0).
			Updates(map[string]interface{}{"status": 1, "trade_no": tradeNo, "paid_at": &now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}

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
				return ores.Error
			}
			if ores.RowsAffected > 0 {
				paidIDs = append(paidIDs, oid)
			}
		}
		if orderIDCount == 0 {
			return errPayLogNoValidOrderIDs
		}
		if len(paidIDs) == 0 {
			return errPayLogNoPendingOrdersUpdated
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errPayLogNoValidOrderIDs) {
			slog.Warn("pay callback", "reason", "pay_log_no_valid_order_ids", "pay_no", payNo)
			return nil
		}
		if errors.Is(err, errPayLogNoPendingOrdersUpdated) {
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
				if err := app.Must().DB.First(&o, oid).Error; err != nil || o.Status != model.OrderStatusPaid {
					allPaid = false
					break
				}
			}
			if allPaid {
				res2 := app.Must().DB.Model(&model.PayLog{}).Where("id = ? AND status = ?", pl.ID, 0).
					Updates(map[string]interface{}{"status": 1, "trade_no": tradeNo, "paid_at": &now})
				return res2.Error
			}
			slog.Warn("pay callback", "reason", "merge_pay_no_pending_and_not_all_paid", "pay_no", payNo)
			return nil
		}
		return err
	}
	if len(paidIDs) == 0 {
		return nil
	}

	// 事务提交后再写历史和通知（非关键路径，失败不影响支付结果）
	for _, oid := range paidIDs {
		AddOrderStatusHistory(oid, model.OrderStatusPending, model.OrderStatusPaid, "支付成功", "系统")
		var order model.Order
		if err := app.Must().DB.First(&order, oid).Error; err != nil {
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
	app.Must().DB.Create(&rl)
	return &rl
}

func UpdateRefundLog(refundNo string, status int8, tradeNo string) {
	app.Must().DB.Model(&model.RefundLog{}).Where("refund_no = ?", refundNo).
		Updates(map[string]interface{}{"status": status, "trade_no": tradeNo})
}

func GetPayLogList(userID uint, page, pageSize int) ([]model.PayLog, int64, error) {
	var total int64
	app.Must().DB.Model(&model.PayLog{}).Where("user_id = ?", userID).Count(&total)
	var list []model.PayLog
	err := app.Must().DB.Where("user_id = ?", userID).Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// PayLogMenuRowsShopXO shopxo-uniapp paylog/index：name + url 列表。
func PayLogMenuRowsShopXO(userID uint) []map[string]string {
	list, _, _ := GetPayLogList(userID, 1, 40)
	if len(list) == 0 {
		return []map[string]string{
			{"name": "我的订单", "url": "/pages/user-order/user-order"},
		}
	}
	out := make([]map[string]string, 0, len(list))
	for i := range list {
		pl := &list[i]
		st := "待支付"
		if pl.Status == 1 {
			st = "已支付"
		} else if pl.Status == 2 {
			st = "已关闭"
		}
		out = append(out, map[string]string{
			"name": fmt.Sprintf("%s · %s", pl.PayNo, st),
			"url":  fmt.Sprintf("/pages/paylog-detail/paylog-detail?id=%d", pl.ID),
		})
	}
	return out
}

// PayLogOrderRowsShopXO shopxo-uniapp paylog/detail：订单号与详情链接。
func PayLogOrderRowsShopXO(userID, payLogID uint) ([]map[string]string, error) {
	var pl model.PayLog
	if err := app.Must().DB.Where("id = ? AND user_id = ?", payLogID, userID).First(&pl).Error; err != nil {
		return nil, errors.New("支付记录不存在")
	}
	ids := parsePayLogOrderIDList(pl.OrderIDs)
	if len(ids) == 0 {
		return []map[string]string{{"order_no": pl.PayNo, "url": "/pages/user-order/user-order"}}, nil
	}
	var orders []model.Order
	app.Must().DB.Where("id IN ? AND user_id = ?", ids, userID).Find(&orders)
	out := make([]map[string]string, 0, len(orders))
	for i := range orders {
		o := orders[i]
		out = append(out, map[string]string{
			"order_no": o.OrderNo,
			"url":      fmt.Sprintf("/pages/user-order-detail/user-order-detail?id=%d", o.ID),
		})
	}
	if len(out) == 0 {
		return []map[string]string{{"order_no": pl.PayNo, "url": "/pages/user-order/user-order"}}, nil
	}
	return out, nil
}

func GetRefundLogList(page, pageSize int) ([]model.RefundLog, int64, error) {
	var total int64
	app.Must().DB.Model(&model.RefundLog{}).Count(&total)
	var list []model.RefundLog
	err := app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func AddPayRequestLog(payLogID uint, request, response, business string) {
	app.Must().DB.Create(&model.PayRequestLog{PayLogID: payLogID, Request: request, Response: response, Business: business})
}
