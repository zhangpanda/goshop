package initialize

import (
	"errors"

	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// RunSchemaAutoMigrate 与 HTTP 服务启动时一致的 Schema 迁移：先拼团成员去重，再 AutoMigrate 全部注册模型。
// 可在 Job/CI 中独立调用（见 cmd/migrate）。
func RunSchemaAutoMigrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if err := DedupeGroupOrderMembersBeforeUniqueIndex(db); err != nil {
		return err
	}
	return db.AutoMigrate(autoMigrateModelList()...)
}

// autoMigrateModelList 复用 model.AllModels（单一真值）。保留本包内封装函数以便今后
// 在 initialize 阶段追加子包特有的模型而不影响 testutil 等其他调用方。
func autoMigrateModelList() []any {
	return model.AllModels()
}
