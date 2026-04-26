package global

import (
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/pkg/cache"
	"github.com/zhangpanda/goshop/pkg/wechat"
	"gorm.io/gorm"
)

var (
	Cfg   *config.Config
	DB    *gorm.DB
	Cache cache.Cache
	WxPay *wechat.Client
)
