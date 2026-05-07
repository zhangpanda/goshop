package initialize

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/cache"
	"github.com/zhangpanda/goshop/pkg/wechat"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() error {
	cfg := global.Cfg.DB
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(log.New(os.Stdout, "\n", log.LstdFlags), logger.Config{
			SlowThreshold: 5 * time.Second,
			LogLevel:      logger.Warn,
		}),
	})
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.DB = db
	return nil
}

func InitRedis() error {
	cfg := global.Cfg.Redis
	if cfg.Host == "" {
		slog.Info("cache", "backend", "memory", "reason", "redis not configured")
		global.Cache = cache.NewMemoryCache()
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("cache", "backend", "memory", "reason", "redis connect failed", "error", err)
		global.Cache = cache.NewMemoryCache()
		return nil
	}

	global.Cache = cache.NewRedisCache(rdb)
	slog.Info("cache", "backend", "redis")
	return nil
}

func InitWechatPay() error {
	cfg := global.Cfg.Wechat
	if cfg.AppID == "" || cfg.MchID == "" {
		return nil // 未配置则跳过
	}
	client, err := wechat.NewClient(cfg.AppID, cfg.MchID, cfg.MchAPIKey, cfg.SerialNo, cfg.PrivateKey, cfg.NotifyURL)
	if err != nil {
		return fmt.Errorf("init wechat pay: %w", err)
	}
	global.WxPay = client
	return nil
}

func InitDefaultAdmin() {
	if os.Getenv("GOSHOP_SKIP_DEFAULT_ADMIN") == "true" {
		return
	}
	var count int64
	global.DB.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return
	}
	role := model.Role{Name: "超级管理员", Desc: "拥有所有权限", Powers: `["*"]`, Status: 1}
	global.DB.Create(&role)
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	global.DB.Create(&model.Admin{Username: "admin", Password: string(hash), Nickname: "管理员", RoleID: role.ID, Status: 1})
	slog.Warn("seed", "action", "default_admin_created", "username", "admin", "password", "admin123", "msg", "CHANGE THIS PASSWORD BEFORE EXPOSING TO PUBLIC NETWORK")
}

func InitDefaultConfig() {
	var count int64
	global.DB.Model(&model.Config{}).Count(&count)
	if count > 0 {
		return
	}
	configs := []model.Config{
		// base
		{Group: "base", Key: "home_site_name", Value: "GoShop", Desc: "站点名称"},
		{Group: "base", Key: "home_site_logo", Value: "", Desc: "电脑端Logo"},
		{Group: "base", Key: "home_site_logo_wap", Value: "", Desc: "手机端Logo"},
		{Group: "base", Key: "home_site_logo_square", Value: "", Desc: "正方形Logo"},
		{Group: "base", Key: "common_timezone", Value: "Asia/Shanghai", Desc: "默认时区"},
		{Group: "base", Key: "common_shop_notice", Value: "", Desc: "商城公告"},
		{Group: "base", Key: "home_footer_info", Value: "", Desc: "底部信息"},
		{Group: "base", Key: "home_statistics_code", Value: "", Desc: "底部统计代码"},
		{Group: "base", Key: "home_content_max_width", Value: "1200", Desc: "页面最大宽度(px)"},
		{Group: "base", Key: "common_cdn_attachment_host", Value: "", Desc: "附件CDN域名"},
		// site
		{Group: "site", Key: "home_site_web_state", Value: "1", Desc: "Web端站点状态(0关1开)"},
		{Group: "site", Key: "home_site_close_reason", Value: "升级中...", Desc: "站点关闭原因"},
		{Group: "site", Key: "home_user_reg_type", Value: "username,sms,email", Desc: "注册方式"},
		{Group: "site", Key: "home_user_login_type", Value: "username,email,sms", Desc: "登录方式"},
		{Group: "site", Key: "common_register_is_enable_audit", Value: "0", Desc: "注册开启审核"},
		{Group: "site", Key: "common_verify_interval_time", Value: "60", Desc: "验证码间隔(秒)"},
		{Group: "site", Key: "common_verify_expire_time", Value: "600", Desc: "验证码有效时间(秒)"},
		// seo
		{Group: "seo", Key: "home_seo_site_title", Value: "GoShop - 品质电商", Desc: "SEO标题"},
		{Group: "seo", Key: "home_seo_site_keywords", Value: "电商,购物,GoShop", Desc: "SEO关键词"},
		{Group: "seo", Key: "home_seo_site_description", Value: "GoShop品质电商平台", Desc: "SEO描述"},
		// email
		{Group: "email", Key: "common_email_smtp_host", Value: "", Desc: "SMTP服务器"},
		{Group: "email", Key: "common_email_smtp_port", Value: "465", Desc: "SMTP端口"},
		{Group: "email", Key: "common_email_smtp_account", Value: "", Desc: "发信人邮件地址"},
		{Group: "email", Key: "common_email_smtp_name", Value: "", Desc: "SMTP用户名"},
		{Group: "email", Key: "common_email_smtp_pwd", Value: "", Desc: "SMTP密码"},
		// sms
		{Group: "sms", Key: "common_sms_type", Value: "aliyun", Desc: "短信平台"},
		{Group: "sms", Key: "common_sms_sign", Value: "", Desc: "短信签名"},
		{Group: "sms", Key: "common_sms_apikey", Value: "", Desc: "AccessKey"},
		{Group: "sms", Key: "common_sms_secret", Value: "", Desc: "AccessSecret"},
		// order
		{Group: "order", Key: "common_order_close_limit_time", Value: "30", Desc: "未付款自动关闭(分钟)"},
		{Group: "order", Key: "common_order_success_limit_time", Value: "15", Desc: "自动确认收货(天)"},
		{Group: "order", Key: "home_order_aftersale_return_launch_day", Value: "7", Desc: "售后发起期限(天)"},
		// attachment
		{Group: "attachment", Key: "home_max_limit_image", Value: "10240000", Desc: "图片最大限制(字节)"},
		{Group: "attachment", Key: "home_max_limit_file", Value: "51200000", Desc: "文件最大限制(字节)"},
		// cache
		{Group: "cache", Key: "common_cache_data_redis_host", Value: "127.0.0.1", Desc: "Redis地址"},
		{Group: "cache", Key: "common_cache_data_redis_port", Value: "6379", Desc: "Redis端口"},
		// admin
		{Group: "admin", Key: "admin_theme_site_name", Value: "GoShop管理后台", Desc: "后台站点名称"},
		{Group: "admin", Key: "common_page_size", Value: "20", Desc: "分页数量"},
		// app
		{Group: "app", Key: "common_app_is_enable_search", Value: "1", Desc: "启用搜索"},
		{Group: "app", Key: "common_app_customer_service_tel", Value: "", Desc: "客服电话"},
		{Group: "app", Key: "common_app_customer_service_email", Value: "hi@zhangpanda.com", Desc: "客服邮箱"},
		{Group: "app", Key: "common_app_customer_service_hours", Value: "周一至周五 9:00-18:00", Desc: "工作时间"},
		{Group: "app", Key: "common_app_h5_url", Value: "", Desc: "手机端H5地址"},
		// weixin
		{Group: "weixin", Key: "common_app_mini_weixin_appid", Value: "", Desc: "微信小程序AppID"},
		{Group: "weixin", Key: "common_app_mini_weixin_appsecret", Value: "", Desc: "微信小程序AppSecret"},
		// alipay
		{Group: "alipay", Key: "common_app_mini_alipay_appid", Value: "", Desc: "支付宝小程序AppID"},
		// logistics
		{Group: "logistics", Key: "logistics_api_key", Value: "", Desc: "快递100 API Key"},
		{Group: "logistics", Key: "logistics_api_customer", Value: "", Desc: "快递100 Customer"},
		// distribution
		{Group: "distribution", Key: "distribution_rate_level1", Value: "10", Desc: "一级分销佣金比例(%)"},
		{Group: "distribution", Key: "distribution_rate_level2", Value: "5", Desc: "二级分销佣金比例(%)"},
	}
	global.DB.Create(&configs)
	slog.Info("seed", "action", "config", "count", len(configs))
}

func InitDefaultPowers() {
	var count int64
	global.DB.Model(&model.Power{}).Count(&count)
	if count > 0 {
		return
	}
	// 预置14个一级菜单权限节点
	powers := []model.Power{
		{ID: 41, ParentID: 0, Name: "系统", Control: "Config.Index", Sort: 1, Status: 1},
		{ID: 81, ParentID: 0, Name: "站点", Control: "Site.Index", Sort: 2, Status: 1},
		{ID: 1, ParentID: 0, Name: "权限", Control: "Power.Index", Sort: 3, Status: 1},
		{ID: 126, ParentID: 0, Name: "用户", Control: "User.Index", Sort: 4, Status: 1},
		{ID: 38, ParentID: 0, Name: "商品", Control: "Goods.Index", Sort: 5, Status: 1},
		{ID: 177, ParentID: 0, Name: "订单", Control: "Order.Index", Sort: 6, Status: 1},
		{ID: 222, ParentID: 0, Name: "网站", Control: "WebSiteAdmin.Index", Sort: 7, Status: 1},
		{ID: 252, ParentID: 0, Name: "品牌", Control: "Brand.Index", Sort: 8, Status: 1},
		{ID: 438, ParentID: 0, Name: "仓库", Control: "Warehouse.Index", Sort: 9, Status: 1},
		{ID: 319, ParentID: 0, Name: "手机", Control: "App.Index", Sort: 20, Status: 1},
		{ID: 204, ParentID: 0, Name: "文章", Control: "Article.Index", Sort: 21, Status: 1},
		{ID: 182, ParentID: 0, Name: "数据", Control: "Data.Index", Sort: 22, Status: 1},
		{ID: 340, ParentID: 0, Name: "应用", Control: "Store.Index", Sort: 30, Status: 1},
		{ID: 118, ParentID: 0, Name: "工具", Control: "Tool.Index", Sort: 50, Status: 1},
		// 二级菜单（核心）
		{ID: 22, ParentID: 1, Name: "管理员列表", Control: "Admin.Index", Sort: 1, Status: 1},
		{ID: 4, ParentID: 1, Name: "角色管理", Control: "Role.Index", Sort: 20, Status: 1},
		{ID: 13, ParentID: 1, Name: "权限分配", Control: "Power.Index", Sort: 30, Status: 1},
		{ID: 39, ParentID: 38, Name: "商品管理", Control: "Goods.Index", Sort: 1, Status: 1},
		{ID: 201, ParentID: 38, Name: "商品分类", Control: "GoodsCategory.Index", Sort: 10, Status: 1},
		{ID: 356, ParentID: 38, Name: "商品评论", Control: "Goodscomments.Index", Sort: 20, Status: 1},
		{ID: 178, ParentID: 177, Name: "订单管理", Control: "Order.Index", Sort: 1, Status: 1},
		{ID: 364, ParentID: 177, Name: "订单售后", Control: "Orderaftersale.Index", Sort: 10, Status: 1},
		{ID: 127, ParentID: 126, Name: "用户列表", Control: "User.Index", Sort: 0, Status: 1},
		{ID: 339, ParentID: 41, Name: "系统配置", Control: "Config.Index", Sort: 0, Status: 1},
		{ID: 103, ParentID: 81, Name: "站点设置", Control: "Site.Index", Sort: 0, Status: 1},
		{ID: 199, ParentID: 81, Name: "SEO设置", Control: "Seo.Index", Sort: 30, Status: 1},
		{ID: 119, ParentID: 118, Name: "缓存管理", Control: "Cache.Index", Sort: 1, Status: 1},
		{ID: 349, ParentID: 118, Name: "SQL控制台", Control: "Sqlconsole.Index", Sort: 10, Status: 1},
	}
	global.DB.Create(&powers)
	slog.Info("seed", "action", "powers", "count", len(powers))
}

func InitDefaultNavigation() {
	var count int64
	global.DB.Model(&model.Navigation{}).Count(&count)
	if count > 0 {
		return
	}
	navs := []model.Navigation{
		// header
		{Name: "产品", URL: "/products", Sort: 30, Status: 1, Type: "header"},
		{Name: "文章", URL: "/articles", Sort: 20, Status: 1, Type: "header"},
		{Name: "支持", URL: "/support", Sort: 10, Status: 1, Type: "header"},
		// footer
		{Name: "全部产品", URL: "/products", Sort: 60, Status: 1, Type: "footer"},
		{Name: "文章资讯", URL: "/articles", Sort: 50, Status: 1, Type: "footer"},
		{Name: "品牌故事", URL: "/story", Sort: 40, Status: 1, Type: "footer"},
		{Name: "联系我们", URL: "/support", Sort: 30, Status: 1, Type: "footer"},
		{Name: "订单查询", URL: "/account/orders", Sort: 20, Status: 1, Type: "footer"},
		{Name: "售后服务", URL: "/account/aftersale", Sort: 10, Status: 1, Type: "footer"},
	}
	global.DB.Create(&navs)
	slog.Info("seed", "action", "navigation", "count", len(navs))
}
