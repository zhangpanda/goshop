package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/initialize"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	d := &app.Deps{Cfg: cfg}
	app.Register(d)
	defer func() {
		if err := d.Close(); err != nil {
			slog.Warn("migrate", "deps_close", err.Error())
		}
	}()

	if err := initialize.InitDB(); err != nil {
		log.Fatalf("init db: %v", err)
	}
	if err := initialize.RunAllSchemaMigrations(app.Must().DB); err != nil {
		log.Fatalf("schema migrate: %v", err)
	}
	slog.Info("migrate", "status", "ok", "hint", "GOSHOP_AUTO_MIGRATE=false 时于流水线执行本命令；GOSHOP_DISABLE_AUTOMIGRATE=true 可仅跑嵌入式 SQL 版本")
	os.Exit(0)
}
