package router

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// mountPprof 把 net/http/pprof 标准 handler 挂到 /internal/pprof/*。
// 仅当 GOSHOP_PPROF=1 时启用；部署侧须保证该路径不对公网暴露。
//
// 常用端点：
//
//	/internal/pprof/           索引
//	/internal/pprof/heap       堆快照
//	/internal/pprof/goroutine  goroutine 栈
//	/internal/pprof/profile    CPU profile（?seconds=30）
func mountPprof(r *gin.Engine) {
	g := r.Group("/internal/pprof")
	g.GET("/", gin.WrapF(pprof.Index))
	g.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	g.GET("/profile", gin.WrapF(pprof.Profile))
	g.POST("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/trace", gin.WrapF(pprof.Trace))
	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		g.GET("/"+name, gin.WrapH(pprof.Handler(name)))
	}
}
