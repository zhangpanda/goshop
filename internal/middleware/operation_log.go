package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// AdminOperationLog 记录管理员操作日志。
// 关键：goroutine 中不能使用 *gin.Context（handler 返回后 c 会被 gin pool 复用）。
// 因此所有字段必须在主协程中抽取成普通值，再交给 app.SafeGo 异步写库。
// SafeGo 还会兜底 panic，避免 DB 抖动时异步任务拖挂整个进程。
func AdminOperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == "GET" {
			return
		}
		adminID := c.GetUint("admin_id")
		if adminID == 0 {
			return
		}
		entry := model.AdminOperationLog{
			AdminID:   adminID,
			Username:  c.GetString("admin_username"),
			Action:    c.Request.Method + " " + c.Request.URL.Path,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		app.SafeGo("operation_log", func() {
			if err := app.Must().DB.Create(&entry).Error; err != nil {
				slog.Warn("operation_log write", "admin_id", entry.AdminID, "err", err)
			}
		})
	}
}
