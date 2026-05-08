package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/pkg/cache"
)

var cronLocalLeaderWarnOnce sync.Once

func cronLockToken() string {
	h, _ := os.Hostname()
	return fmt.Sprintf("%d:%s", os.Getpid(), h)
}

// tryAcquireCronTick 多实例下使用 Redis SETNX 抢本轮执行权；仅内存缓存时打一次告警并仍执行（单实例或本地开发）。
func tryAcquireCronTick(ctx context.Context, job string, lease time.Duration) bool {
	if lease <= 0 {
		lease = 30 * time.Second
	}
	switch c := app.Must().Cache.(type) {
	case *cache.RedisCache:
		ok, err := c.SetNX(ctx, "goshop:cron:tick:"+job, cronLockToken(), lease)
		if err != nil {
			slog.Warn("cron", "leader", "redis", "job", job, "err", err.Error())
			return false
		}
		return ok
	default:
		cronLocalLeaderWarnOnce.Do(func() {
			slog.Warn("cron", "leader", "memory_only", "hint", "多实例请配置 Redis，或对纯 API 副本设置 GOSHOP_CRON_ENABLED=false")
		})
		return true
	}
}
