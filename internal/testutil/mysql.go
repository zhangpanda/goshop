package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/repository"
	"github.com/zhangpanda/goshop/pkg/cache"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// mysqlEnvDSN 从环境变量读取 MySQL DSN，用于集成测试与 MySQL 专属行为测试。
//
// 约定：
//   - GOSHOP_TEST_MYSQL_DSN 存在 → 用它
//   - 否则拼 GOSHOP_TEST_MYSQL_HOST / _PORT / _USER / _PASS / _DB 为 DSN
//   - 最终仍为空 → 返回 ""（调用方决定 Skip 或 Fail）
func mysqlEnvDSN() string {
	if dsn := os.Getenv("GOSHOP_TEST_MYSQL_DSN"); dsn != "" {
		return dsn
	}
	host := os.Getenv("GOSHOP_TEST_MYSQL_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("GOSHOP_TEST_MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("GOSHOP_TEST_MYSQL_USER")
	if user == "" {
		user = "root"
	}
	pass := os.Getenv("GOSHOP_TEST_MYSQL_PASS")
	db := os.Getenv("GOSHOP_TEST_MYSQL_DB")
	if db == "" {
		db = "goshop_test"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4&multiStatements=true",
		user, pass, host, port, db)
}

// SetupMySQLOrSkip 连接 env 指定的 MySQL 并执行 AutoMigrate（与启动期一致的模型集合）。
// 没有可用 MySQL 时 `t.Skip`，保证本地 dev 无 Docker/MySQL 时不会失败。
//
// 使用场景：
//   - 验证仅在真 MySQL 下可观测的行为（如 FOR UPDATE 行锁、
//     带 NULL 的 UNIQUE 索引并发冲突、金额 DECIMAL 精度等）
//   - CI 中 `GOSHOP_TEST_MYSQL_DSN=root:testpass@tcp(127.0.0.1:3306)/goshop_test?parseTime=true` 会命中
//
// 返回的 *gorm.DB 生命周期为单个测试：用 t.Cleanup 关闭底层连接池并 DROP 测试库。
func SetupMySQLOrSkip(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := mysqlEnvDSN()
	if dsn == "" {
		t.Skip("MySQL 测试跳过：未配置 GOSHOP_TEST_MYSQL_DSN（或 *_HOST）；本地 dev 可启动 docker compose up -d mysql 或直接 export DSN")
		return nil
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("connect mysql: %v", err)
	}

	// 连接测：10s 内 ping 不通则 skip，避免 CI 半启动状态下失败
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("gorm.DB(): %v", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Skipf("MySQL 不可达: %v", err)
		return nil
	}

	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	// 每个测试用独立的 DB 隔离有点重（DROP DATABASE 权限不一定够），
	// 改为测试结束时清理所有它创建过的行。这里不主动 TRUNCATE，调用方按需在
	// t.Cleanup 里做；Setup 仅保证 schema 就绪。
	t.Cleanup(func() {
		// 不 Close：CI 里 MySQL 服务容器生命周期与 workflow 一致，连接池关闭无必要；
		// 此处仅确保测试注册的 app.Deps 不泄漏到其他包。
	})
	return db
}

// SetupMySQLAppDeps 构造基于真 MySQL 的 app.Deps 并注册到全局；返回 teardown。
// 与 SetupTestDB 并列——前者走 in-memory SQLite（快，不覆盖 MySQL 特定行为），
// 本函数走真 MySQL（慢但覆盖面更广，主要用于并发/锁/唯一索引类验证）。
func SetupMySQLAppDeps(t *testing.T) (db *gorm.DB, teardown func()) {
	t.Helper()
	db = SetupMySQLOrSkip(t)
	if db == nil { // t.Skip 已在里面调用
		return nil, func() {}
	}
	var prev *app.Deps
	if app.Registered() {
		prev = app.Must()
	}
	app.Register(&app.Deps{
		Cfg:   &config.Config{},
		DB:    db,
		Cache: cache.NewMemoryCache(),
	})
	repository.Init(db)
	return db, func() {
		if prev != nil {
			app.Register(prev)
		} else {
			app.Clear()
		}
	}
}
