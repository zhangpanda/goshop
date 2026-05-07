package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/initialize"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	global.Cfg = cfg

	if err := initialize.InitDB(); err != nil {
		log.Fatalf("init db: %v", err)
	}
	if err := initialize.RunAllSchemaMigrations(global.DB); err != nil {
		log.Fatalf("schema migrate: %v", err)
	}
	slog.Info("migrate", "status", "ok", "hint", "GOSHOP_AUTO_MIGRATE=false 时于流水线执行本命令；GOSHOP_DISABLE_AUTOMIGRATE=true 可仅跑嵌入式 SQL 版本")
	os.Exit(0)
}
