package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/initialize"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/repository"
	"github.com/zhangpanda/goshop/internal/router"
	"github.com/zhangpanda/goshop/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	global.Cfg = cfg

	// 结构化日志：release 模式用 JSON，debug 用 Text
	if cfg.Server.Mode == "release" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	// 初始化数据库
	if err := initialize.InitDB(); err != nil {
		log.Fatalf("init db: %v", err)
	}
	if os.Getenv("GOSHOP_AUTO_MIGRATE") != "false" {
		// 自动建表前：拼团成员历史重复数据去重，避免 uniq_group_order_member 迁移失败
		if err := initialize.DedupeGroupOrderMembersBeforeUniqueIndex(global.DB); err != nil {
			log.Fatalf("pre automigrate dedupe group_order_members: %v", err)
		}
		// 自动建表
		if err := global.DB.AutoMigrate(
			&model.User{}, &model.Category{}, &model.Goods{}, &model.GoodsSKU{},
			&model.Cart{}, &model.Address{}, &model.Order{}, &model.OrderItem{}, &model.OrderStatusHistory{},
			&model.Coupon{}, &model.UserCoupon{}, &model.Promotion{}, &model.PromotionItem{},
			&model.Favorite{}, &model.BrowseHistory{}, &model.Review{},
			&model.Shipment{}, &model.PointsLog{}, &model.Message{},
			&model.Admin{}, &model.Role{},
			&model.OrderAftersale{}, &model.AftersaleHistory{},
			&model.Brand{}, &model.Article{}, &model.ArticleCategory{},
			&model.SpecTemplate{}, &model.SpecType{}, &model.SpecValue{},
			&model.GoodsParamsTemplate{}, &model.GoodsParamsConfig{}, &model.GoodsParams{},
			&model.SearchHistory{}, &model.ScreeningPrice{},
			&model.Config{}, &model.Region{}, &model.Slide{}, &model.Navigation{}, &model.Link{},
			&model.Payment{}, &model.SmsLog{}, &model.EmailLog{},
			&model.Attachment{}, &model.AttachmentCategory{}, &model.ErrorLog{},
			&model.UserPlatform{}, &model.VerifyCode{},
			&model.PayLog{}, &model.PayRequestLog{}, &model.RefundLog{},
			&model.GoodsSpecBase{}, &model.GoodsPhoto{},
			&model.Warehouse{}, &model.WarehouseGoods{}, &model.WarehouseGoodsSpec{},
			&model.Express{}, &model.InventoryLog{}, &model.GoodsGiveIntegralLog{},
			&model.BrandCategory{}, &model.BrandCategoryJoin{}, &model.GoodsCategoryJoin{},
			&model.Power{}, &model.RolePower{},
			&model.Plugin{}, &model.PluginCategory{},
			&model.Diy{}, &model.CustomView{}, &model.ThemeData{},
			&model.FormInput{}, &model.FormInputData{},
			&model.AppHomeNav{}, &model.AppCenterNav{}, &model.AppTabbar{},
			&model.ShortcutMenu{}, &model.Agreement{},
			&model.OrderTraceSource{}, &model.OrderCurrency{},
			&model.Design{}, &model.FormTableUserFields{}, &model.GoodsContentApp{},
			&model.Layout{}, &model.OrderService{}, &model.PayLogValue{},
			&model.PluginsDataConfig{}, &model.QuickNav{}, &model.RolePlugins{},
			&model.Answer{},
			&model.AppMini{}, &model.WalletLog{},
			&model.AdminOperationLog{},
			&model.GroupOrder{}, &model.GroupOrderMember{},
			&model.Distributor{}, &model.CommissionLog{}, &model.WithdrawRequest{},
		); err != nil {
			log.Fatalf("auto migrate: %v", err)
		}
	} else {
		slog.Info("migrate", "skipped", true, "env", "GOSHOP_AUTO_MIGRATE=false")
	}

	// 初始化缓存（Redis 或内存）
	if err := initialize.InitRedis(); err != nil {
		slog.Warn("init", "component", "cache", "error", err.Error())
	}

	// 初始化 Repository 层
	repository.Init(global.DB)

	// 初始化 OrderService 单例
	service.InitOrderService(service.NewOrderService(
		global.DB, repository.Repos.Order, repository.Repos.Cart,
		repository.Repos.Address, repository.Repos.SKU,
	))

	// 初始化默认管理员
	initialize.InitDefaultAdmin()

	// 初始化默认配置
	initialize.InitDefaultConfig()

	// 初始化默认权限节点
	initialize.InitDefaultPowers()

	// 初始化默认导航菜单
	initialize.InitDefaultNavigation()

	// 初始化展示数据
	initialize.InitDefaultSeedData()
	initialize.EnsureDefaultPayments()

	// 初始化微信支付
	if err := initialize.InitWechatPay(); err != nil {
		log.Fatalf("init wechat pay: %v", err)
	}

	// 启动 HTTP 服务
	gin.SetMode(cfg.Server.Mode)

	// 启动定时任务
	go service.StartCronJobs()
	r := gin.New()
	if len(cfg.Server.TrustedProxies) > 0 {
		r.SetTrustedProxies(cfg.Server.TrustedProxies)
	} else {
		r.SetTrustedProxies(nil)
	}
	router.Setup(r)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		slog.Info("server started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// 优雅关闭：等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	slog.Info("server exited")
}
