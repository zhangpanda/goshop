package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
)

func GenerateQRCodeURL(content string) string {
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=%s", content)
}

var forbiddenSQLKeywords = []string{
	"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "TRUNCATE", "CREATE",
	"GRANT", "REVOKE", "INTO OUTFILE", "INTO DUMPFILE", "LOAD_FILE",
	"SLEEP", "BENCHMARK", "EXEC", "EXECUTE",
}

func ExecuteSQL(sqlStr string) ([]map[string]interface{}, error) {
	trimmed := strings.TrimSpace(sqlStr)
	upper := strings.ToUpper(trimmed)
	if strings.Contains(trimmed, ";") && strings.TrimRight(trimmed, "; \t\n") != strings.TrimRight(strings.SplitN(trimmed, ";", 2)[0], " \t\n") {
		return nil, errors.New("禁止多语句执行")
	}
	allowed := false
	for _, prefix := range []string{"SELECT", "SHOW", "DESC", "EXPLAIN"} {
		if strings.HasPrefix(upper, prefix) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("仅允许 SELECT/SHOW/DESC/EXPLAIN 查询")
	}
	for _, kw := range forbiddenSQLKeywords {
		if strings.Contains(upper, kw) {
			return nil, fmt.Errorf("SQL 包含禁止关键字: %s", kw)
		}
	}
	for _, blocked := range []string{"INFORMATION_SCHEMA", "PERFORMANCE_SCHEMA", "MYSQL.", "SYS."} {
		if strings.Contains(upper, blocked) {
			return nil, errors.New("禁止查询系统库或元数据")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var results []map[string]interface{}
	err := app.Must().DB.WithContext(ctx).Raw(trimmed).Limit(1000).Find(&results).Error
	return results, err
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	AppVersion   string `json:"app_version"`
	DBVersion    string `json:"db_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	StartTime    string `json:"start_time"`
}

var appStartTime = time.Now()

// AppVersion 由构建时 -ldflags 注入，默认 dev。
var AppVersion = "dev"

func GetSystemInfo() *SystemInfo {
	var dbVer string
	app.Must().DB.Raw("SELECT VERSION()").Scan(&dbVer)
	return &SystemInfo{
		GoVersion:    runtime.Version(),
		AppVersion:   AppVersion,
		DBVersion:    dbVer,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		StartTime:    appStartTime.Format(time.DateTime),
	}
}
