package shopxomigrate

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed data.sql
var dataSQL string

// RenderSQL 将嵌入的 ShopXO→GoShop 导入脚本中的占位符替换为实际库名与表前缀。
// shopxoDB：源库（含 sxo_* 业务表）；goshopDB：目标库（已完成 GoShop 建表）。
// tablePrefix：ShopXO 表前缀，默认 sxo_；与官方不一致时传入实际前缀（如 pre_）。
func RenderSQL(shopxoDB, goshopDB, tablePrefix string) string {
	if tablePrefix == "" {
		tablePrefix = "sxo_"
	}
	s := dataSQL
	s = strings.ReplaceAll(s, "__SHOPXO_DB__", quoteDB(shopxoDB))
	s = strings.ReplaceAll(s, "__GOSHOP_DB__", quoteDB(goshopDB))
	if tablePrefix != "sxo_" {
		s = strings.ReplaceAll(s, "sxo_", tablePrefix)
	}
	return s
}

// WipeImportedTablesSQL 生成在目标库中清空本导入涉及的表所需的语句（外键检查已关闭）。
func WipeImportedTablesSQL(goshopDB string) string {
	db := quoteDB(goshopDB)
	tables := []string{
		"users", "categories", "goods", "goods_skus", "orders",
		"order_items", "addresses", "brands", "articles", "admins",
	}
	var b strings.Builder
	b.WriteString("SET FOREIGN_KEY_CHECKS = 0;\n")
	for _, t := range tables {
		b.WriteString(fmt.Sprintf("TRUNCATE TABLE %s.%s;\n", db, quoteDB(t)))
	}
	b.WriteString("SET FOREIGN_KEY_CHECKS = 1;\n")
	return b.String()
}

func quoteDB(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}
