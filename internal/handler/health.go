package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
)

// ReadinessCheck 检查数据库连接是否正常
func ReadinessCheck(c *gin.Context) {
	db, err := global.DB.DB()
	if err != nil || db.Ping() != nil {
		c.JSON(503, gin.H{"status": "unavailable", "reason": "database"})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}
