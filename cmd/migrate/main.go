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
	if err := initialize.RunSchemaAutoMigrate(global.DB); err != nil {
		log.Fatalf("schema migrate: %v", err)
	}
	slog.Info("migrate", "status", "ok", "hint", "可与 GOSHOP_AUTO_MIGRATE=false 配合在发布流水线中单独执行")
	os.Exit(0)
}
