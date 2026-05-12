// cmd/alipay-reconcile-once/main.go
//
// 一次性对所有 Pending 订单执行 alipay.trade.query 并推进订单到 Paid。
// 不受 cron 的 10min cutoff 限制，用于本地 sandbox 联调即时对账。
//
// 用法：
//
//	./alipay-reconcile-once config.yaml
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: alipay-reconcile-once <config.yaml>")
		os.Exit(2)
	}
	cfg, err := config.Load(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "gorm open:", err)
		os.Exit(1)
	}
	app.Register(&app.Deps{Cfg: cfg, DB: db})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 对所有 Pending 订单 query + notify（不受 cron 10min cutoff 限制）
	var orders []model.Order
	db.Where("status = ?", model.OrderStatusPending).Find(&orders)
	var fixed int
	for i := range orders {
		o := &orders[i]
		st, tradeNo, err := service.AlipayQueryTrade(ctx, o.OrderNo)
		if err != nil {
			fmt.Printf("  query %s: err=%v\n", o.OrderNo, err)
			continue
		}
		fmt.Printf("  order %s (id=%d): trade_status=%s trade_no=%s\n", o.OrderNo, o.ID, st, tradeNo)
		if !service.AlipayTradePaid(st) {
			continue
		}
		if err := service.HandlePayNotify(o.OrderNo, tradeNo); err != nil {
			fmt.Printf("    HandlePayNotify err: %v\n", err)
			continue
		}
		fixed++
	}
	fmt.Printf("orders_fixed=%d (of %d pending)\n", fixed, len(orders))
}
