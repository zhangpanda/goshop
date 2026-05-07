package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
)

// ReadinessCheck 检查数据库（及已配置的 Redis）是否可用。
func ReadinessCheck(c *gin.Context) {
	db, err := global.DB.DB()
	if err != nil || db.Ping() != nil {
		c.JSON(503, gin.H{"status": "unavailable", "reason": "database"})
		return
	}
	if global.RDB != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := global.RDB.Ping(ctx).Err(); err != nil {
			c.JSON(503, gin.H{"status": "unavailable", "reason": "redis"})
			return
		}
	}
	c.JSON(200, gin.H{"status": "ok"})
}
