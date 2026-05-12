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
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/initialize"
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
	deps := &app.Deps{Cfg: cfg}
	app.Register(deps)

	if p := os.Getenv("GOSHOP_METRICS_PATH"); p != "" {
		cfg.Server.MetricsPath = p
	}
	if b := os.Getenv("GOSHOP_RATE_LIMIT_BACKEND"); b != "" {
		cfg.Server.RateLimitBackend = b
	}

	// 结构化日志：release 模式用 JSON，debug 用 Text
	if cfg.Server.Mode == "release" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	// 初始化数据库
	if err := initialize.InitDB(); err != nil {
		log.Fatalf("init db: %v", err)
	}
	if os.Getenv("GOSHOP_AUTO_MIGRATE") != "false" {
		if err := initialize.RunAllSchemaMigrations(app.Must().DB); err != nil {
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
	repository.Init(app.Must().DB)

	// 初始化 OrderService 单例
	service.InitOrderService(service.NewOrderService(
		app.Must().DB, repository.Repos.Order, repository.Repos.Cart,
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

	cronCtx, cronCancel := context.WithCancel(context.Background())
	service.RegisterPayEventListeners()
	go service.StartCronJobs(cronCtx, deps)

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

	cronCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	if err := deps.Close(); err != nil {
		slog.Warn("shutdown", "deps_close", err.Error())
	}
	slog.Info("server exited")
}
