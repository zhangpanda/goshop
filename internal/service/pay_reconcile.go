package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

const wechatTradeStateSuccess = "SUCCESS"

// ReconcileWechatPayments 主动查询微信订单状态，补「回调丢失」场景：合并支付 PayLog + 单笔订单号。
func ReconcileWechatPayments(ctx context.Context, deps *app.Deps) (payLogsFixed, ordersFixed int) {
	if deps.WxPay == nil {
		return 0, 0
	}
	cutoff := time.Now().Add(-10 * time.Minute)

	var logs []model.PayLog
	deps.DB.Where("status = ? AND created_at < ?", 0, cutoff).Limit(40).Find(&logs)
	for i := range logs {
		pl := &logs[i]
		tx, err := deps.WxPay.QueryOrderByOutTradeNo(ctx, pl.PayNo)
		if err != nil {
			continue
		}
		if tx == nil || tx.TradeState == nil || *tx.TradeState != wechatTradeStateSuccess {
			continue
		}
		tid := ""
		if tx.TransactionId != nil {
			tid = *tx.TransactionId
		}
		if err := PayLogSuccess(pl.PayNo, tid); err != nil {
			slog.Warn("pay_reconcile", "kind", "paylog", "pay_no", pl.PayNo, "err", err.Error())
			continue
		}
		payLogsFixed++
	}

	var orders []model.Order
	deps.DB.Where("status = ? AND created_at < ?", model.OrderStatusPending, cutoff).Limit(30).Find(&orders)
	for i := range orders {
		o := &orders[i]
		tx, err := deps.WxPay.QueryOrderByOutTradeNo(ctx, o.OrderNo)
		if err != nil {
			continue
		}
		if tx == nil || tx.TradeState == nil || *tx.TradeState != wechatTradeStateSuccess {
			continue
		}
		tid := ""
		if tx.TransactionId != nil {
			tid = *tx.TransactionId
		}
		if err := HandlePayNotify(o.OrderNo, tid); err != nil {
			slog.Warn("pay_reconcile", "kind", "order", "order_no", o.OrderNo, "err", err.Error())
			continue
		}
		ordersFixed++
	}
	return payLogsFixed, ordersFixed
}

// ReconcileAlipayPayments 主动查询支付宝订单状态，补回调丢失（合并 PayLog.out_trade_no = pay_no，单笔为 order_no）。
func ReconcileAlipayPayments(ctx context.Context, deps *app.Deps) (payLogsFixed, ordersFixed int) {
	if deps.Cfg == nil {
		return 0, 0
	}
	cfg := deps.Cfg.Alipay
	if cfg.AppID == "" || cfg.PrivateKey == "" || cfg.PublicKey == "" {
		return 0, 0
	}
	cutoff := time.Now().Add(-10 * time.Minute)

	var logs []model.PayLog
	deps.DB.Where("status = ? AND created_at < ?", 0, cutoff).Limit(40).Find(&logs)
	for i := range logs {
		pl := &logs[i]
		st, tradeNo, err := AlipayQueryTrade(ctx, pl.PayNo)
		if err != nil {
			continue
		}
		if !AlipayTradePaid(st) {
			continue
		}
		if err := PayLogSuccess(pl.PayNo, tradeNo); err != nil {
			slog.Warn("pay_reconcile", "channel", "alipay", "kind", "paylog", "pay_no", pl.PayNo, "err", err.Error())
			continue
		}
		payLogsFixed++
	}

	var orders []model.Order
	deps.DB.Where("status = ? AND created_at < ?", model.OrderStatusPending, cutoff).Limit(30).Find(&orders)
	for i := range orders {
		o := &orders[i]
		st, tradeNo, err := AlipayQueryTrade(ctx, o.OrderNo)
		if err != nil {
			continue
		}
		if !AlipayTradePaid(st) {
			continue
		}
		if err := HandlePayNotify(o.OrderNo, tradeNo); err != nil {
			slog.Warn("pay_reconcile", "channel", "alipay", "kind", "order", "order_no", o.OrderNo, "err", err.Error())
			continue
		}
		ordersFixed++
	}
	return payLogsFixed, ordersFixed
}
