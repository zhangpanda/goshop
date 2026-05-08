package app

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/pkg/cache"
	"github.com/zhangpanda/goshop/pkg/wechat"
	"gorm.io/gorm"
)

// Deps 聚合进程级运行时依赖，替代散落的 package global 可变包变量。
//
// 约定：在任意业务包调用 Must() 之前，必须由 cmd/*/main 或 TestMain 执行 Register（非 nil）。
// 仅 Cfg 可先就绪，DB/Redis 等在 initialize 阶段写入同一 Deps 指针。
type Deps struct {
	Cfg   *config.Config
	DB    *gorm.DB
	RDB   *redis.Client
	Cache cache.Cache
	WxPay *wechat.Client
}

var reg atomic.Pointer[Deps]

// Register 注册（或替换）全局 Deps，供 Must 使用。d 不得为 nil。
func Register(d *Deps) {
	if d == nil {
		panic("app: Register(nil) 非法，请传入 &app.Deps{Cfg: cfg}（或测试用桩）")
	}
	reg.Store(d)
}

// Must 返回已注册的 Deps。若从未 Register 或 Register(nil)，将 panic（fail-fast，避免静默空指针）。
func Must() *Deps {
	p := reg.Load()
	if p == nil {
		panic("app: Deps 未注册。请在进程入口（cmd/*/main）或 testing.TestMain 中先调用 app.Register(&app.Deps{...})，再执行 initialize 与各包逻辑")
	}
	return p
}

// Close 释放 *gorm.DB 底层连接池与 Redis 客户端（若已创建）。在 HTTP Server Shutdown 完成后调用，便于优雅退出。
func (d *Deps) Close() error {
	if d == nil {
		return nil
	}
	var errs []error
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("gorm sql db: %w", err))
		} else if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("sql.Close: %w", err))
		}
	}
	if d.RDB != nil {
		if err := d.RDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis.Close: %w", err))
		}
	}
	return errors.Join(errs...)
}

// Registered 是否已 Register（测试辅助）。
func Registered() bool {
	return reg.Load() != nil
}

// Clear 仅用于测试：取消注册，避免用例间泄漏。
func Clear() {
	reg.Store(nil)
}

// String 供调试（不打印密钥）。
func (d *Deps) String() string {
	if d == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Deps(cfg=%t db=%t rdb=%t cache=%t wxpay=%t)",
		d.Cfg != nil, d.DB != nil, d.RDB != nil, d.Cache != nil, d.WxPay != nil)
}
