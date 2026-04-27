package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// AdminOperationLog 记录管理员操作日志
func AdminOperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// 只记录写操作
		if c.Request.Method == "GET" {
			return
		}
		adminID := c.GetUint("admin_id")
		if adminID == 0 {
			return
		}
		go func() {
			global.DB.Create(&model.AdminOperationLog{
				AdminID:   adminID,
				Username:  c.GetString("admin_username"),
				Action:    c.Request.Method + " " + c.Request.URL.Path,
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
			})
		}()
	}
}
