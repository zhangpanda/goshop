package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
)

func Cors() gin.HandlerFunc {
	allowedOrigins := global.Cfg.Server.CorsOrigins

	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 配置了生产域名则按白名单匹配
			if len(allowedOrigins) > 0 {
				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}
				return false
			}
			// 未配置则仅允许开发环境 localhost
			for _, prefix := range []string{"http://localhost:", "http://127.0.0.1:"} {
				if strings.HasPrefix(origin, prefix) {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
	})
}
