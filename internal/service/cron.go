package service

import (
	"log/slog"
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
		tx := global.DB.Begin()
		tx.Model(&o).Update("status", model.OrderStatusCancelled)
		// 恢复库存
		var items []model.OrderItem
		tx.Where("order_id = ?", o.ID).Find(&items)
		for _, item := range items {
			tx.Model(&model.GoodsSKU{}).Where("id = ?", item.SKUID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity))
		}
		if err := tx.Commit().Error; err != nil {
			fail++
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

	for _, o := range orders {
		now := time.Now()
		global.DB.Model(&o).Updates(map[string]interface{}{
			"status": model.OrderStatusCompleted, "completed_at": &now,
		})
		AddOrderStatusHistory(o.ID, model.OrderStatusShipped, model.OrderStatusCompleted, "超时自动确认收货", "系统")
		SendMessage(o.UserID, "订单已完成", "您的订单已自动确认收货", "order", o.ID)
		// 订单完成奖励积分
		OrderRewardPoints(o.UserID, o.ID, o.PayAmount)
		// 分销佣金结算
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

// StartCronJobs 启动所有定时任务
func StartCronJobs() {
	// 订单自动关闭 - 每分钟检查
	go func() {
		for {
			time.Sleep(1 * time.Minute)
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
			s, f := CronGoodsGiveIntegral()
			if s > 0 || f > 0 {
				slog.Info("cron", "job", "goods_integral", "success", s, "fail", f)
			}
		}
	}()

	// 锁定积分释放 - 每小时检查
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			s, _ := CronIntegralRelease(21600) // 15天
			if s > 0 {
				slog.Info("cron", "job", "integral_release", "success", s)
			}
		}
	}()

	slog.Info("cron jobs started")
}
