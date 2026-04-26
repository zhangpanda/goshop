package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/auth"
	"github.com/zhangpanda/goshop/pkg/response"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(strings.TrimPrefix(token, "Bearer "), global.Cfg.JWT.Secret)
		if err != nil || !claims.IsAdmin {
			response.Fail(c, http.StatusUnauthorized, "token无效")
			c.Abort()
			return
		}
		var admin model.Admin
		if err := global.DB.First(&admin, claims.UserID).Error; err != nil || admin.Status == 0 {
			response.Fail(c, http.StatusUnauthorized, "账号已禁用或不存在")
			c.Abort()
			return
		}
		c.Set("admin_id", admin.ID)
		c.Set("admin_role_id", admin.RoleID)
		c.Next()
	}
}

func AdminPower(power string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.GetUint("admin_role_id")
		if !service.CheckPower(roleID, power) {
			response.Fail(c, http.StatusForbidden, "无权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
