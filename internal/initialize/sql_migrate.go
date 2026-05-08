package initialize

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/migratefs"
	"gorm.io/gorm"
)

/**
 * mysqlMigrateDatabaseURL 构造 golang-migrate 的 mysql:// URL。
 *
 * 必须使用独立连接池：若用 mysqldb.WithInstance(gorm 的 *sql.DB) 并在 m.Close() 时释放，
 * mysql 驱动会关闭整个共享连接池，导致后续 GORM AutoMigrate 报 「database is closed」。
 */
func mysqlMigrateDatabaseURL() (string, error) {
	cfg := app.Must().Cfg.DB
	if cfg.Host == "" || cfg.DBName == "" {
		return "", fmt.Errorf("migrate: db host or dbname empty")
	}
	port := cfg.Port
	if port == 0 {
		port = 3306
	}
	mc := mysqlDriver.NewConfig()
	mc.User = cfg.User
	mc.Passwd = cfg.Password
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%d", cfg.Host, port)
	mc.DBName = cfg.DBName
	mc.Params = map[string]string{"charset": "utf8mb4"}
	mc.ParseTime = true
	mc.Loc = time.Local
	mc.MultiStatements = true
	return "mysql://" + mc.FormatDSN(), nil
}

// RunEmbeddedSQLMigrations 执行内嵌的 golang-migrate SQL 版本（当前含 baseline 占位）。
func RunEmbeddedSQLMigrations() error {
	migrateURL, err := mysqlMigrateDatabaseURL()
	if err != nil {
		return err
	}
	src, err := iofs.New(migratefs.FS, ".")
	if err != nil {
		return fmt.Errorf("migrate iofs: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// RunAllSchemaMigrations 嵌入式 SQL 升级 →（默认）GORM AutoMigrate；仅 SQL 时可设 GOSHOP_DISABLE_AUTOMIGRATE=true。
func RunAllSchemaMigrations(db *gorm.DB) error {
	if err := RunEmbeddedSQLMigrations(); err != nil {
		return err
	}
	if os.Getenv("GOSHOP_DISABLE_AUTOMIGRATE") == "true" {
		slog.Info("migrate", "automigrate", "skipped", "env", "GOSHOP_DISABLE_AUTOMIGRATE=true")
		return nil
	}
	return RunSchemaAutoMigrate(db)
}
