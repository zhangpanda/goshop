package global

import (
	"github.com/redis/go-redis/v9"
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/pkg/cache"
	"github.com/zhangpanda/goshop/pkg/wechat"
	"gorm.io/gorm"
)

var (
	Cfg   *config.Config
	DB    *gorm.DB
	RDB   *redis.Client // Redis 原生客户端；缓存走内存时为 nil（限流/分布式锁等检测此字段）
	Cache cache.Cache
	WxPay *wechat.Client
)
