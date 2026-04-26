package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/pkg/auth"
	"github.com/zhangpanda/goshop/pkg/response"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(strings.TrimPrefix(token, "Bearer "), global.Cfg.JWT.Secret)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "token无效")
			c.Abort()
			return
		}
		if claims.IsAdmin {
			response.Fail(c, http.StatusUnauthorized, "token类型错误")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
