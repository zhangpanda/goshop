package initialize

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	mysqldb "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/zhangpanda/goshop/internal/migratefs"
	"gorm.io/gorm"
)

// RunEmbeddedSQLMigrations 执行内嵌的 golang-migrate SQL 版本（当前含 baseline 占位）。
func RunEmbeddedSQLMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	driver, err := mysqldb.WithInstance(sqlDB, &mysqldb.Config{})
	if err != nil {
		return fmt.Errorf("migrate mysql driver: %w", err)
	}
	src, err := iofs.New(migratefs.FS, ".")
	if err != nil {
		return fmt.Errorf("migrate iofs: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
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
	if err := RunEmbeddedSQLMigrations(db); err != nil {
		return err
	}
	if os.Getenv("GOSHOP_DISABLE_AUTOMIGRATE") == "true" {
		slog.Info("migrate", "automigrate", "skipped", "env", "GOSHOP_DISABLE_AUTOMIGRATE=true")
		return nil
	}
	return RunSchemaAutoMigrate(db)
}
