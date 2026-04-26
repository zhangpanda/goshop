package global

import (
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/pkg/wechat"
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
)

var (
	Cfg    *config.Config
	DB     *gorm.DB
	RDB    *redis.Client
	WxPay  *wechat.Client
)
