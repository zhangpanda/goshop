// Package event 提供进程内事件总线，用于解耦业务副作用（通知、佣金、积分等）。
// 未来可替换为消息队列实现而不影响发布方。
package event

import (
	"log/slog"
	"sync"
)

// Handler 事件处理函数。payload 类型由事件名约定。
type Handler func(payload any)

var (
	mu       sync.RWMutex
	handlers = map[string][]Handler{}
	wg       sync.WaitGroup
)

// On 注册事件监听器。同一事件可注册多个 handler，按注册顺序执行。
func On(name string, fn Handler) {
	mu.Lock()
	handlers[name] = append(handlers[name], fn)
	mu.Unlock()
}

// Emit 触发事件。每个 handler 在独立 goroutine 中执行，panic 被恢复并记录日志。
func Emit(name string, payload any) {
	mu.RLock()
	hs := handlers[name]
	mu.RUnlock()
	for _, fn := range hs {
		wg.Add(1)
		go func(f Handler) {
			defer wg.Done()
			safeCall(name, f, payload)
		}(fn)
	}
}

// EmitSync 同步触发事件（测试用）。
func EmitSync(name string, payload any) {
	mu.RLock()
	hs := handlers[name]
	mu.RUnlock()
	for _, fn := range hs {
		safeCall(name, fn, payload)
	}
}

// Drain 等待所有已触发的异步 handler 完成。应在 HTTP server Shutdown 之后调用。
func Drain() {
	wg.Wait()
}

func safeCall(event string, fn Handler, payload any) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("event handler panic", "event", event, "recover", r)
		}
	}()
	fn(payload)
}

// Reset 清空所有注册的 handler（仅测试用）。
func Reset() {
	mu.Lock()
	handlers = map[string][]Handler{}
	mu.Unlock()
}
