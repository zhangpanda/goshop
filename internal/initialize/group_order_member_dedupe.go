package initialize

import (
	"fmt"
	"log/slog"

	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// DedupeGroupOrderMembersBeforeUniqueIndex 在 AutoMigrate 创建 uniq_group_order_member 之前删除重复行；
// 每个 (group_order_id, user_id) 仅保留 id 最小的一行。表不存在时无操作。
func DedupeGroupOrderMembersBeforeUniqueIndex(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if !db.Migrator().HasTable(&model.GroupOrderMember{}) {
		return nil
	}

	var stmt string
	switch db.Dialector.Name() {
	case "mysql":
		stmt = `DELETE g1 FROM group_order_members g1
INNER JOIN group_order_members g2
  ON g1.group_order_id = g2.group_order_id AND g1.user_id = g2.user_id AND g1.id > g2.id`
	case "sqlite":
		stmt = `DELETE FROM group_order_members
WHERE id IN (
  SELECT id FROM (
    SELECT g1.id AS id FROM group_order_members g1
    INNER JOIN group_order_members g2
      ON g1.group_order_id = g2.group_order_id AND g1.user_id = g2.user_id AND g1.id > g2.id
  ) AS gom_dup_ids
)`
	default:
		return fmt.Errorf("dedupe group_order_members: unsupported dialect %q", db.Dialector.Name())
	}

	res := db.Exec(stmt)
	if res.Error != nil {
		return fmt.Errorf("dedupe group_order_members: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		slog.Info("migrate", "action", "dedupe_group_order_members", "rows_deleted", res.RowsAffected)
	}
	return nil
}
