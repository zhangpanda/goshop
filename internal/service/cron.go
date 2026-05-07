package service

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// CronOrderClose 自动关闭超时未支付订单（默认30分钟）
func CronOrderClose(minutes int) (sucs, fail int) {
	if minutes <= 0 {
		minutes = 30
	}
	deadline := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var orders []model.Order
	global.DB.Where("status = ? AND created_at < ?", model.OrderStatusPending, deadline).Find(&orders)

	for _, o := range orders {
		o := o
		var changed bool
		err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
			res := tx.Model(&model.Order{}).Where("id = ? AND status = ?", o.ID, model.OrderStatusPending).
				Update("status", model.OrderStatusCancelled)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			changed = true
			var items []model.OrderItem
			if err := tx.Where("order_id = ?", o.ID).Find(&items).Error; err != nil {
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
		if err != nil {
			fail++
			continue
		}
		if !changed {
			continue
		}
		AddOrderStatusHistory(o.ID, model.OrderStatusPending, model.OrderStatusCancelled, "超时未支付自动关闭", "系统")
		SendMessage(o.UserID, "订单已关闭", "您的订单因超时未支付已自动关闭", "order", o.ID)
		sucs++
	}
	return
}

// CronOrderAutoReceive 自动确认收货（默认15天）
func CronOrderAutoReceive(days int) (sucs, fail int) {
	if days <= 0 {
		days = 15
	}
	deadline := time.Now().AddDate(0, 0, -days)
	var orders []model.Order
	global.DB.Where("status = ? AND shipped_at < ?", model.OrderStatusShipped, deadline).Find(&orders)

	for i := range orders {
		o := &orders[i]
		var updated bool
		err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
			now := time.Now()
			res := tx.Model(&model.Order{}).
				Where("id = ? AND status = ? AND shipped_at IS NOT NULL AND shipped_at < ?", o.ID, model.OrderStatusShipped, deadline).
				Updates(map[string]interface{}{"status": model.OrderStatusCompleted, "completed_at": &now})
			if res.Error != nil {
				return res.Error
			}
			updated = res.RowsAffected > 0
			return nil
		})
		if err != nil {
			fail++
			continue
		}
		if !updated {
			continue
		}
		AddOrderStatusHistory(o.ID, model.OrderStatusShipped, model.OrderStatusCompleted, "超时自动确认收货", "系统")
		SendMessage(o.UserID, "订单已完成", "您的订单已自动确认收货", "order", o.ID)
		if err := OrderRewardPoints(o.UserID, o.ID, o.PayAmount); err != nil {
			slog.Warn("cron", "job", "order_receive", "reason", "reward_points", "order_id", o.ID, "err", err.Error())
		}
		SettleCommission(o.ID)
		sucs++
	}
	return
}

// CronGoodsGiveIntegral 商品赠送积分（订单完成后延迟赠送）
func CronGoodsGiveIntegral() (sucs, fail int) {
	// 查找已完成但未赠送积分的订单（完成超过24小时）
	deadline := time.Now().Add(-24 * time.Hour)
	var orders []model.Order
	global.DB.Where("status = ? AND completed_at < ? AND completed_at IS NOT NULL", model.OrderStatusCompleted, deadline).
		Limit(100).Find(&orders)

	for _, o := range orders {
		// 检查是否已赠送（通过积分日志判断）
		var count int64
		global.DB.Model(&model.PointsLog{}).Where("user_id = ? AND type = ? AND ref_id = ?", o.UserID, "goods_integral", o.ID).Count(&count)
		if count > 0 {
			continue
		}
		// 查商品赠送积分（简化：每消费1元赠1积分）
		points := int(o.PayAmount / 100)
		if points > 0 {
			ChangePoints(o.UserID, points, "goods_integral", o.ID, "商品赠送积分")
			sucs++
		}
	}
	return
}

// StartCronJobs 启动所有定时任务（多实例且使用 Redis 缓存时，每轮通过 SETNX 仅一台执行）。
func StartCronJobs() {
	if os.Getenv("GOSHOP_CRON_ENABLED") == "false" {
		slog.Info("cron", "skipped", true, "env", "GOSHOP_CRON_ENABLED=false")
		return
	}

	// 订单自动关闭 - 每分钟检查
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			ok := tryAcquireCronTick(ctx, "order_close", 50*time.Second)
			cancel()
			if !ok {
				continue
			}
			s, f := CronOrderClose(30)
			if s > 0 || f > 0 {
				slog.Info("cron", "job", "order_close", "success", s, "fail", f)
			}
		}
	}()

	// 自动确认收货 - 每小时检查
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			ok := tryAcquireCronTick(ctx, "order_receive", 50*time.Minute)
			cancel()
			if !ok {
				continue
			}
			s, f := CronOrderAutoReceive(15)
			if s > 0 || f > 0 {
				slog.Info("cron", "job", "order_receive", "success", s, "fail", f)
			}
		}
	}()

	// 商品积分赠送 - 每小时检查
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			ok := tryAcquireCronTick(ctx, "goods_integral", 50*time.Minute)
			cancel()
			if !ok {
				continue
			}
			s, f := CronGoodsGiveIntegral()
			if s > 0 || f > 0 {
				slog.Info("cron", "job", "goods_integral", "success", s, "fail", f)
			}
		}
	}()

	// 长时间未完成的合并支付 PayLog：关单并解除待付子单的 payment_id（与预下单失败回滚互补）
	go func() {
		for {
			time.Sleep(15 * time.Minute)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			ok := tryAcquireCronTick(ctx, "stale_paylog_close", 12*time.Minute)
			cancel()
			if !ok {
				continue
			}
			n := StalePendingPayLogsCleanup(120)
			if n > 0 {
				slog.Info("cron", "job", "stale_paylog_close", "closed", n)
			}
		}
	}()

	// 微信订单主动查单补单（回调丢失补偿）；需配置 WxPay
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			ctxLock, cancelLock := context.WithTimeout(context.Background(), 10*time.Second)
			ok := tryAcquireCronTick(ctxLock, "wechat_reconcile", 4*time.Minute)
			cancelLock()
			if !ok {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			pl, od := ReconcileWechatPayments(ctx)
			cancel()
			if pl > 0 || od > 0 {
				slog.Info("cron", "job", "wechat_reconcile", "paylogs", pl, "orders", od)
			}
		}
	}()

	// 锁定积分释放 - 每小时检查
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			ok := tryAcquireCronTick(ctx, "integral_release", 50*time.Minute)
			cancel()
			if !ok {
				continue
			}
			s, _ := CronIntegralRelease(21600) // 15天
			if s > 0 {
				slog.Info("cron", "job", "integral_release", "success", s)
			}
		}
	}()

	slog.Info("cron jobs started")
}
