package migratefs

import "embed"

// FS 嵌入式 SQL 迁移（golang-migrate）。新增版本请追加 000002_xxx.up.sql / .down.sql。
//
//go:embed *.sql
var FS embed.FS
