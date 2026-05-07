package service

import "gorm.io/gorm"

/**
 * RunInDBTx 在 (*gorm.DB).Transaction 内执行 fn：fn 返回非 nil 时自动回滚，nil 时提交。
 * 优先用此替代手写 Begin + Commit + Rollback，避免「Commit 失败后又 Rollback」等一致性问题。
 */
func RunInDBTx(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
