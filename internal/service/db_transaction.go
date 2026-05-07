package service

import "gorm.io/gorm"

/**
 * RunInDBTx 在 (*gorm.DB).Transaction 内执行 fn：fn 返回非 nil 时自动回滚，nil 时提交。
 * 全库多语句写库请优先用此，避免手写 Begin/Commit/Rollback。
 */
func RunInDBTx(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
