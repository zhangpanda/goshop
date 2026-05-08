package app

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// SafeGo 启动后台 goroutine 并捕获 panic，防止未处理异常导致整个进程崩溃。
//
// 用法：
//
//	app.SafeGo("search_history", func() { ... })
//
// name 用于日志定位。fn 内部仍应自行处理业务错误；SafeGo 只兜底 panic。
func SafeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("goroutine panic",
					"name", name,
					"recover", fmt.Sprint(r),
					"stack", string(debug.Stack()),
				)
			}
		}()
		fn()
	}()
}
