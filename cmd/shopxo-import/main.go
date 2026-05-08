package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/zhangpanda/goshop/internal/shopxomigrate"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.SetFlags(0)

	host := flag.String("host", envOr("GOSHOP_MYSQL_HOST", "127.0.0.1"), "MySQL 主机（或环境变量 GOSHOP_MYSQL_HOST）")
	port := flag.Int("port", envInt("GOSHOP_MYSQL_PORT", 3306), "MySQL 端口（或 GOSHOP_MYSQL_PORT）")
	user := flag.String("user", envOr("GOSHOP_MYSQL_USER", "root"), "MySQL 用户（或 GOSHOP_MYSQL_USER）")
	password := flag.String("password", envOr("GOSHOP_MYSQL_PASSWORD", ""), "MySQL 密码（或 GOSHOP_MYSQL_PASSWORD）")
	from := flag.String("from", "", "ShopXO 源库名（必须）")
	to := flag.String("to", "", "GoShop 目标库名（必须，需已 AutoMigrate 建表）")
	tablePrefix := flag.String("table-prefix", "sxo_", "ShopXO 表前缀（默认 sxo_）")
	wipe := flag.Bool("wipe-target-tables", false, "导入前 TRUNCATE 目标库中相关表（危险，确认无重要数据）")
	dryRun := flag.Bool("dry-run", false, "仅打印将执行的 SQL 字符数并退出，不连库")
	printSQL := flag.Bool("print-sql", false, "将完整 SQL 打印到 stdout 并退出（不连库）")
	adminReset := flag.String("reset-admin-password", "", "导入后把指定管理员 id 的 password 设为 bcrypt(明文)；常与 -admin-id 联用")
	adminID := flag.Uint("admin-id", 1, "配合 -reset-admin-password 的管理员主键")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `从 ShopXO MySQL 库向已建表的 GoShop 库导入核心业务数据（用户/分类/商品/SKU/订单/地址/品牌/文章/管理员）。

用法:
  %s -from shopxo_prod -to goshop_prod [选项]

示例:
  export GOSHOP_MYSQL_PASSWORD='secret'
  %s -from shopxo -to goshop -wipe-target-tables \
    -reset-admin-password 'YourNewAdminPass' -admin-id 1

注意:
  - 先在同一 MySQL 实例上准备好两库；目标库需已运行过 GoShop 迁移（空表或允许 -wipe 清空）。
  - C 端用户密码 ShopXO 为 md5+salt，GoShop 为 bcrypt；导入后 users.password 为空，需运营引导重置或自研兼容登录。
  - 上传文件与 URL 替换、支付回调域名仍需按 docs/migration-from-shopxo.md 手工完成。
  - admins.role_id 须在目标库 roles 表有对应行，否则管理端会 403；本程序不自动插入 roles（见文档「管理员与 roles」）。

`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *from == "" || *to == "" {
		flag.Usage()
		os.Exit(2)
	}

	sqlText := shopxomigrate.RenderSQL(*from, *to, *tablePrefix)
	if *printSQL {
		fmt.Print(sqlText)
		return
	}
	if *dryRun {
		fmt.Fprintf(os.Stderr, "dry-run: SQL length=%d bytes (from=%q to=%q prefix=%q)\n", len(sqlText), *from, *to, *tablePrefix)
		return
	}

	dsn := buildDSN(*user, *password, *host, *port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("mysql open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("mysql ping: %v", err)
	}

	if *wipe {
		wipeSQL := shopxomigrate.WipeImportedTablesSQL(*to)
		if _, err := db.Exec(wipeSQL); err != nil {
			log.Fatalf("wipe target tables: %v", err)
		}
		log.Printf("已清空目标库 %q 中导入涉及的表", *to)
	}

	if _, err := db.Exec(sqlText); err != nil {
		log.Fatalf("import exec: %v", err)
	}
	log.Printf("导入完成: %q -> %q", *from, *to)

	if *adminReset != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*adminReset), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("bcrypt: %v", err)
		}
		q := fmt.Sprintf("UPDATE %s.`admins` SET `password` = ? WHERE `id` = ?",
			quoteDBIdent(*to))
		res, err := db.Exec(q, string(hash), *adminID)
		if err != nil {
			log.Fatalf("reset admin password: %v", err)
		}
		n, _ := res.RowsAffected()
		log.Printf("已更新管理员 id=%d 密码（影响行数 %d）", *adminID, n)
	}
}

func buildDSN(user, password, host string, port int) string {
	cfg := mysqlDriver.NewConfig()
	cfg.User = user
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.Params = map[string]string{
		"charset":         "utf8mb4",
		"parseTime":       "true",
		"multiStatements": "true",
	}
	return cfg.FormatDSN()
}

func quoteDBIdent(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return def
	}
	return n
}
