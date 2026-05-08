package service

import (
	"os"
	"testing"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
)

// TestMain 为整个 service 测试包注册默认 app.Deps（单测可再覆盖 Cfg/Cache 等字段）。
func TestMain(m *testing.M) {
	app.Register(&app.Deps{Cfg: &config.Config{}})
	os.Exit(m.Run())
}
