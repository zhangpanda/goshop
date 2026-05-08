package service

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// CronOrderClose 自动关闭超时未支付订单（默认30分钟）
func CronOrderClose(deps *app.Deps, minutes int) (sucs, fail int) {
	if minutes <= 0 {
		minutes = 30
	}
	deadline := time.Now().Add(-time.Duration(minutes) * time.Minute)
	var orders []model.Order
	deps.DB.Where("status = ? AND created_at < ?", model.OrderStatusPending, deadline).Find(&orders)

	for _, o := range orders {
		o := o
		var changed bool
		err := RunInDBTx(deps.DB, func(tx *gorm.DB) error {
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
func CronOrderAutoReceive(deps *app.Deps, days int) (sucs, fail int) {
	if days <= 0 {
		days = 15
	}
	deadline := time.Now().AddDate(0, 0, -days)
	var orders []model.Order
	deps.DB.Where("status = ? AND shipped_at < ?", model.OrderStatusShipped, deadline).Find(&orders)

	for i := range orders {
		o := &orders[i]
		var updated bool
		err := RunInDBTx(deps.DB, func(tx *gorm.DB) error {
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
func CronGoodsGiveIntegral(deps *app.Deps) (sucs, fail int) {
	deadline := time.Now().Add(-24 * time.Hour)
	var orders []model.Order
	deps.DB.Where("status = ? AND completed_at < ? AND completed_at IS NOT NULL", model.OrderStatusCompleted, deadline).
		Limit(100).Find(&orders)

	for _, o := range orders {
		var count int64
		deps.DB.Model(&model.PointsLog{}).Where("user_id = ? AND type = ? AND ref_id = ?", o.UserID, "goods_integral", o.ID).Count(&count)
		if count > 0 {
			continue
		}
		points := int(o.PayAmount / 100)
		if points > 0 {
			ChangePoints(o.UserID, points, "goods_integral", o.ID, "商品赠送积分")
			sucs++
		}
	}
	return
}

// StartCronJobs 启动所有定时任务；ctx 取消时各循环退出（与 HTTP Shutdown 协同）。
func StartCronJobs(ctx context.Context, deps *app.Deps) {
	if deps == nil {
		deps = app.Must()
	}
	if os.Getenv("GOSHOP_CRON_ENABLED") == "false" {
		slog.Info("cron", "skipped", true, "env", "GOSHOP_CRON_ENABLED=false")
		return
	}

	go cronLoop(ctx, "order_close", time.Minute, 5*time.Second, func(lockCtx context.Context) bool {
		return tryAcquireCronTick(lockCtx, "order_close", 50*time.Second)
	}, func() {
		s, f := CronOrderClose(deps, 30)
		if s > 0 || f > 0 {
			slog.Info("cron", "job", "order_close", "success", s, "fail", f)
		}
	})

	go cronLoop(ctx, "order_receive", time.Hour, 10*time.Second, func(lockCtx context.Context) bool {
		return tryAcquireCronTick(lockCtx, "order_receive", 50*time.Minute)
	}, func() {
		s, f := CronOrderAutoReceive(deps, 15)
		if s > 0 || f > 0 {
			slog.Info("cron", "job", "order_receive", "success", s, "fail", f)
		}
	})

	go cronLoop(ctx, "goods_integral", time.Hour, 10*time.Second, func(lockCtx context.Context) bool {
		return tryAcquireCronTick(lockCtx, "goods_integral", 50*time.Minute)
	}, func() {
		s, f := CronGoodsGiveIntegral(deps)
		if s > 0 || f > 0 {
			slog.Info("cron", "job", "goods_integral", "success", s, "fail", f)
		}
	})

	go cronLoop(ctx, "stale_paylog_close", 15*time.Minute, 10*time.Second, func(lockCtx context.Context) bool {
		return tryAcquireCronTick(lockCtx, "stale_paylog_close", 12*time.Minute)
	}, func() {
		n := StalePendingPayLogsCleanup(deps, 120)
		if n > 0 {
			slog.Info("cron", "job", "stale_paylog_close", "closed", n)
		}
	})

	go cronLoop(ctx, "pay_reconcile", 5*time.Minute, 10*time.Second, func(lockCtx context.Context) bool {
		return tryAcquireCronTick(lockCtx, "pay_reconcile", 4*time.Minute)
	}, func() {
		rctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		wPl, wOd := ReconcileWechatPayments(rctx, deps)
		aPl, aOd := ReconcileAlipayPayments(rctx, deps)
		if wPl > 0 || wOd > 0 || aPl > 0 || aOd > 0 {
			slog.Info("cron", "job", "pay_reconcile", "wechat_paylogs", wPl, "wechat_orders", wOd, "alipay_paylogs", aPl, "alipay_orders", aOd)
		}
	})

	go cronLoop(ctx, "integral_release", time.Hour, 10*time.Second, func(lockCtx context.Context) bool {
		return tryAcquireCronTick(lockCtx, "integral_release", 50*time.Minute)
	}, func() {
		s, _ := CronIntegralRelease(deps, 21600)
		if s > 0 {
			slog.Info("cron", "job", "integral_release", "success", s)
		}
	})

	slog.Info("cron jobs started")
}

func cronLoop(
	ctx context.Context,
	jobName string,
	period time.Duration,
	lockTimeout time.Duration,
	tryLock func(lockCtx context.Context) bool,
	work func(),
) {
	t := time.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("cron", "stopped", true, "job", jobName)
			return
		case <-t.C:
			lockCtx, cancel := context.WithTimeout(context.Background(), lockTimeout)
			ok := tryLock(lockCtx)
			cancel()
			if !ok {
				continue
			}
			work()
		}
	}
}
