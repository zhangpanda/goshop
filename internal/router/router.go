package router

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/compat/shopxo"
	"github.com/zhangpanda/goshop/internal/handler"
	"github.com/zhangpanda/goshop/internal/middleware"
)

func Setup(r *gin.Engine) {
	r.MaxMultipartMemory = 8 << 20 // 8MB for file uploads

	// 探活：尽量轻量，不挂 Prometheus/CORS/日志中间件（便于编排健康检查）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/ready", handler.ReadinessCheck)

	if app.Must().Cfg != nil && app.Must().Cfg.Server.MetricsPath != "" {
		r.Use(middleware.PrometheusHTTP())
	}
	r.Use(middleware.OtelGin("goshop"), middleware.InjectTraceID())
	r.Use(middleware.Cors(), middleware.Logger(), gin.Recovery())

	if app.Must().Cfg != nil && app.Must().Cfg.Server.MetricsPath != "" {
		r.GET(app.Must().Cfg.Server.MetricsPath, gin.WrapH(promhttp.Handler()))
	}

	// pprof：默认关闭；GOSHOP_PPROF=1 时挂在 /internal/pprof/*。
	// 线上请务必在反代/安全组上把 /internal/* 只放给内网。
	if os.Getenv("GOSHOP_PPROF") == "1" {
		mountPprof(r)
	}

	// 静态文件
	r.StaticFS("/uploads", gin.Dir("./uploads", false))
	r.StaticFS("/static/diy", gin.Dir("./static/diy", false))
	r.StaticFS("/static/form_input", gin.Dir("./static/form_input", false))
	r.StaticFile("/diy.html", "./static/diy.html")
	r.StaticFile("/form.html", "./static/form.html")

	// /api.php 兼容入口（shopxo-uniapp 等常用请求形态，非 ShopXO 官方实现）
	shopxo.SetupShopXOCompat(r)

	// DIY/Form 编辑器兼容路由
	shopxo.SetupDiyApiCompat(r)

	authRL := middleware.RateLimit(10, time.Minute)  // 登录/注册: 10次/分钟
	apiRL := middleware.RateLimit(120, time.Minute)  // 公开 API: 120次/分钟/IP

	api := r.Group("/api", apiRL)
	{
		api.POST("/register", authRL, handler.Register)
		api.POST("/login", authRL, handler.Login)
		api.POST("/wx/login", authRL, handler.WxLogin)
	}

	// 公开接口
	api.GET("/categories", handler.GetCategoryTree)
	api.GET("/goods", handler.GetGoodsList)
	api.GET("/goods/:id", handler.GetGoodsDetail)
	api.GET("/goods/:id/reviews", handler.GetGoodsReviews)
	api.GET("/goods/:id/params", handler.GetGoodsParamsHandler)
	api.GET("/goods/:id/specs", handler.GetGoodsSpecBase)
	api.GET("/goods/:id/photos", handler.GetGoodsPhotos)
	api.GET("/coupons", handler.GetCouponList)
	api.GET("/promotions", handler.GetActivePromotions)
	api.GET("/seckills", handler.GetActiveSeckills)
	api.GET("/group-buys", handler.GetActiveGroupBuys)
	api.GET("/group-orders/:id", handler.GetGroupOrderDetail)
	api.GET("/brands", handler.GetBrandList)
	api.GET("/articles", handler.GetArticleList)
	api.GET("/articles/:id", handler.GetArticleDetail)
	api.GET("/article-categories", handler.GetArticleCategoryList)
	api.GET("/search/hot", handler.GetHotKeywords)
	api.GET("/search/prices", handler.GetScreeningPrices)
	api.GET("/regions", handler.GetRegionList)
	api.GET("/slides", handler.GetSlideList)
	api.GET("/navigations", handler.GetNavigationList)
	api.GET("/links", handler.GetLinkList)
	api.GET("/payments", handler.GetPaymentList)

	api.GET("/express", handler.GetExpressList)
	api.GET("/brand-categories", handler.GetBrandCategoryList)
	api.GET("/custom-views", handler.CustomViewListHandler)
	api.GET("/agreement", handler.AgreementGetHandler)
	api.GET("/app/home-nav", handler.AppHomeNavListHandler)
	api.GET("/app/center-nav", handler.AppCenterNavListHandler)
	api.GET("/app/tabbar", handler.AppTabbarListHandler)
	api.GET("/quick-nav", handler.QuickNavListHandler)
	api.GET("/goods/:id/content-app", handler.GetGoodsContentAppHandler)
	api.GET("/answers", handler.GetAnswerList)

	// DiyApi公开数据源
	api.GET("/diyapi/goods", handler.DiyApiGoodsAutoData)
	api.GET("/diyapi/articles", handler.DiyApiArticleAutoData)
	api.GET("/diyapi/brands", handler.DiyApiBrandAutoData)

	// 站点配置公开
	api.GET("/site-config", handler.GetSiteConfigHandler)
	api.GET("/self-extraction-address", handler.GetSelfExtractionAddress)

	// 二维码
	api.GET("/qrcode", handler.GenerateQRCode)

	// 商品补充公开接口
	api.GET("/goods/:id/stock", handler.GoodsStockHandler)
	api.GET("/goods/:id/spec-detail", handler.GoodsSpecDetailHandler)
	api.GET("/goods/:id/score", handler.GoodsScoreHandler)
	api.GET("/home/floors", handler.HomeFloorListHandler)
	api.GET("/guess-you-like", handler.GuessYouLikeHandler)

	// 多平台登录
	api.POST("/platform/login", handler.PlatformLogin)

	// 支付回调
	api.POST("/pay/notify", handler.PayNotify)
	api.POST("/pay/alipay-notify", handler.AlipayNotify)
	api.POST("/pay/paypal/notify", handler.PayPalNotify)
	api.GET("/pay/paypal/capture", handler.PayPalCapture)  // 前端 return_url 回跳
	api.POST("/pay/paypal/capture", handler.PayPalCapture) // REST 显式调用
	api.GET("/pay/sandbox/callback", handler.SandboxCallback)

	// 公开安全接口
	smsRL := middleware.RateLimit(5, time.Minute) // 验证码: 5次/分钟
	api.POST("/verify-code", smsRL, handler.SendVerifyCode)
	api.POST("/forget-password", smsRL, handler.ForgetPassword)

	// 多语言/货币公开接口
	api.GET("/multilingual", handler.GetMultilingualConfig)
	api.GET("/lang-pack", handler.GetLangPack)
	api.GET("/currency", handler.GetCurrencyConfig)

	auth := api.Group("").Use(middleware.JWTAuth())
	{
		auth.GET("/user/profile", handler.GetProfile)
		auth.PUT("/user/password", handler.UpdatePassword)
		auth.POST("/user/bind-mobile", handler.BindMobile)
		auth.GET("/user/platforms", handler.GetUserPlatforms)

		// 购物车
		auth.POST("/cart", handler.AddCart)
		auth.GET("/cart", handler.GetCartList)
		auth.PUT("/cart/:id", handler.UpdateCart)
		auth.DELETE("/cart", handler.DeleteCart)
		auth.PUT("/cart/select-all", handler.SelectAllCart)

		// 收货地址
		auth.POST("/address", handler.CreateAddress)
		auth.GET("/address", handler.GetAddressList)
		auth.PUT("/address/:id", handler.UpdateAddress)
		auth.DELETE("/address/:id", handler.DeleteAddress)

		// 订单
		auth.POST("/orders", handler.CreateOrder)
		auth.GET("/orders", handler.GetOrderList)
		auth.GET("/orders/:id", handler.GetOrderDetail)
		auth.DELETE("/orders/:id", handler.DeleteOrderHandler)
		auth.PUT("/orders/:id/cancel", handler.CancelOrder)
		auth.PUT("/orders/:id/receive", handler.ConfirmReceive)
		auth.GET("/orders/:id/shipment", handler.GetShipment)

		// 支付
		auth.POST("/pay", handler.PayOrder)
		auth.POST("/pay/unified", handler.UnifiedPay)
		auth.POST("/pay/refund", handler.RefundOrder)
		auth.POST("/pay/log", handler.CreatePayLog)
		auth.GET("/pay/log", handler.GetPayLogList)

		// 优惠券
		auth.POST("/coupons/:id/receive", handler.ReceiveCoupon)
		auth.GET("/my/coupons", handler.GetMyCoupons)

		// 秒杀
		auth.POST("/seckill/:item_id/buy", handler.SeckillBuy)

		// 拼团
		auth.POST("/group/:id/open", handler.OpenGroup)
		auth.POST("/group/:id/join", handler.JoinGroup)

		// 分销
		auth.POST("/distribution/apply", handler.ApplyDistributor)
		auth.GET("/distribution/me", handler.GetMyDistributor)
		auth.GET("/distribution/team", handler.GetMyTeam)
		auth.GET("/distribution/commission-logs", handler.GetMyCommissionLogs)
		auth.POST("/distribution/withdraw", handler.RequestWithdraw)

		// 收藏
		auth.POST("/favorites/:id", handler.ToggleFavorite)
		auth.GET("/favorites", handler.GetFavorites)

		// 浏览记录
		auth.GET("/history", handler.GetBrowseHistory)
		auth.DELETE("/history", handler.ClearBrowseHistory)

		// 评价
		auth.POST("/reviews", handler.CreateReview)

		// 积分
		auth.POST("/points/sign", handler.SignIn)
		auth.GET("/points/log", handler.GetPointsLog)

		// 消息
		auth.GET("/messages", handler.GetMessages)
		auth.PUT("/messages/:id/read", handler.ReadMessage)
		auth.PUT("/messages/read-all", handler.ReadAllMessages)

		// 售后
		auth.POST("/aftersale", handler.AftersaleCreate)
		auth.GET("/aftersale", handler.GetAftersaleList)
		auth.GET("/aftersale/:id", handler.GetAftersaleDetail)
		auth.PUT("/aftersale/:id/delivery", handler.AftersaleDelivery)
		auth.PUT("/aftersale/:id/cancel", handler.AftersaleCancel)

		// 搜索历史
		auth.GET("/search/history", handler.GetSearchHistory)
		auth.DELETE("/search/history", handler.ClearSearchHistory)

		// 订单客服
		auth.POST("/order-service", handler.CreateOrderServiceHandler)
		auth.GET("/orders/:id/service", handler.GetOrderServiceList)

		// 表单提交
		auth.POST("/form-data", handler.FormInputDataSubmitHandler)

		// 订单状态历史
		auth.GET("/orders/:id/history", handler.GetOrderStatusHistory)

		// 订单状态分组统计
		auth.GET("/orders/status-total", handler.OrderStatusGroupTotalHandler)

		// 订单操作按钮
		auth.GET("/orders/:id/operate", handler.OrderOperateHandler)

		// 文件上传
		auth.POST("/upload", handler.Upload)

		// 问答留言
		auth.POST("/answers", handler.CreateAnswer)

		// 账号注销
		auth.POST("/user/logout", handler.UserLogout)

		// DiyApi需登录的数据源
		auth.GET("/diyapi/favor", handler.DiyApiGoodsFavorAutoData)
		auth.GET("/diyapi/browse", handler.DiyApiGoodsBrowseAutoData)
	}

	// 管理员登录（无需鉴权）
	api.POST("/admin/login", authRL, handler.AdminLoginHandler)
	api.GET("/admin/captcha", handler.AdminCaptcha)

	// 后台管理（使用 AdminAuth 中间件 + 操作日志）
	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuth(), middleware.AdminOperationLog())
	{
		admin.GET("/dashboard", handler.AdminDashboard)
		admin.GET("/statistical", handler.AdminStatistical)
		admin.GET("/system-info", handler.GetSystemInfo)

		// ========== 权限管理 (Power.Index) ==========
		rbac := admin.Group("").Use(middleware.AdminPower("Power.Index"))
		{
			rbac.POST("/admins", handler.CreateAdminHandler)
			rbac.GET("/admins", handler.GetAdminListHandler)
			rbac.GET("/admins/:id", handler.AdminDetailHandler)
			rbac.PUT("/admins/:id/status", handler.UpdateAdminStatusHandler)
			rbac.DELETE("/admins/:id", handler.AdminDeleteHandler)
			rbac.POST("/roles", handler.CreateRoleHandler)
			rbac.GET("/roles", handler.GetRoleListHandler)
			rbac.GET("/roles/:id", handler.RoleDetailHandler)
			rbac.PUT("/roles/:id", handler.UpdateRoleHandler)
			rbac.DELETE("/roles/:id", handler.DeleteRoleHandler)
			rbac.PUT("/roles/:id/status", handler.RoleStatusUpdateHandler)
			rbac.POST("/powers", handler.CreatePowerHandler)
			rbac.GET("/powers", handler.GetPowerTree)
			rbac.PUT("/roles/:id/powers", handler.SaveRolePowers)
			rbac.GET("/roles/:id/powers", handler.GetRolePowersHandler)
			rbac.GET("/roles/:id/plugins", handler.GetRolePluginsHandler)
			rbac.PUT("/roles/:id/plugins", handler.SaveRolePluginsHandler)
		}

		// ========== 商品管理 (Goods.Index) ==========
		goods := admin.Group("").Use(middleware.AdminPower("Goods.Index"))
		{
			goods.POST("/goods", handler.CreateGoods)
			goods.PUT("/goods/:id", handler.AdminUpdateGoods)
			goods.DELETE("/goods/:id", handler.AdminDeleteGoods)
			goods.PUT("/goods/:id/status", handler.AdminToggleGoodsStatus)
			goods.PUT("/goods/:id/params", handler.SaveGoodsParams)
			goods.PUT("/goods/:id/specs", handler.SaveGoodsSpecBase)
			goods.PUT("/goods/:id/photos", handler.SaveGoodsPhotos)
			goods.PUT("/goods/:id/categories", handler.SaveGoodsCategoryJoin)
			goods.PUT("/goods/:id/content-app", handler.SaveGoodsContentAppHandler)
			goods.POST("/categories", handler.CreateCategory)
			goods.PUT("/categories/:id", handler.AdminUpdateCategory)
			goods.DELETE("/categories/:id", handler.AdminDeleteCategory)
			goods.PUT("/categories/:id/status", handler.CategoryStatusUpdate)
			goods.PUT("/reviews/:id/reply", handler.ReplyReview)
			goods.DELETE("/reviews/:id", handler.ReviewDeleteHandler)
			goods.POST("/spec-templates", handler.CreateSpecTemplate)
			goods.GET("/spec-templates", handler.GetSpecTemplateList)
			goods.POST("/params-templates", handler.CreateParamsTemplate)
			goods.GET("/params-templates", handler.GetParamsTemplateList)
			goods.POST("/promotions", handler.CreatePromotion)
			goods.PUT("/promotions/:id", handler.PromotionUpdateHandler)
			goods.DELETE("/promotions/:id", handler.PromotionDeleteHandler)
			goods.GET("/seckills", handler.GetSeckillList)
			goods.POST("/seckills", handler.CreateSeckill)
			goods.GET("/group-buys", handler.GetGroupBuyList)
			goods.POST("/group-buys", handler.CreateGroupBuy)
			goods.POST("/coupons", handler.CreateCoupon)
			goods.PUT("/coupons/:id", handler.CouponUpdateHandler)
			goods.DELETE("/coupons/:id", handler.CouponDeleteHandler)
		}

		// ========== 订单管理 (Order.Index) ==========
		order := admin.Group("").Use(middleware.AdminPower("Order.Index"))
		{
			order.GET("/orders", handler.AdminGetOrders)
			order.PUT("/orders/:id/remark", handler.AdminUpdateOrderRemark)
			order.POST("/orders/ship", handler.ShipOrder)
			order.PUT("/orders/:id/cancel", handler.AdminCancelOrder)
			order.PUT("/orders/:id/confirm", handler.AdminConfirmReceive)
			order.DELETE("/orders/:id", handler.AdminDeleteOrder)
			order.PUT("/orders/booking-confirm", handler.AdminBookingConfirm)
			order.PUT("/orders/pay-underline", handler.AdminOrderPayUnderLineHandler)
			order.GET("/orders/:id/logistics", handler.LogisticsTrackHandler)
			order.GET("/order-service", handler.AdminOrderServiceList)
			order.PUT("/order-service/:id/reply", handler.AdminReplyOrderService)
			order.GET("/aftersale", handler.AdminAftersaleList)
			order.PUT("/aftersale/:id/confirm", handler.AdminAftersaleConfirm)
			order.PUT("/aftersale/:id/audit", handler.AdminAftersaleAudit)
			order.PUT("/aftersale/:id/refuse", handler.AdminAftersaleRefuse)
			order.DELETE("/aftersale/:id", handler.AdminAftersaleDelete)
			order.PUT("/aftersale/:id/cancel", handler.AdminAftersaleCancel)
		}

		// ========== 用户管理 (User.Index) ==========
		user := admin.Group("").Use(middleware.AdminPower("User.Index"))
		{
			user.GET("/users", handler.AdminGetUsers)
			user.PUT("/users/:id/status", handler.AdminUpdateUserStatus)
			user.DELETE("/users/:id", handler.AdminDeleteUserHandler)
			user.GET("/user-address", handler.UserAddressListHandler)
			user.GET("/user-address/:id", handler.UserAddressDetailHandler)
			user.PUT("/user-address/:id", handler.UserAddressSaveHandler)
			user.DELETE("/user-address/:id", handler.UserAddressDeleteHandler)
		}

		// ========== 网站管理 (WebSiteAdmin.Index) ==========
		website := admin.Group("").Use(middleware.AdminPower("WebSiteAdmin.Index"))
		{
			website.POST("/slides", handler.CreateSlideHandler)
			website.PUT("/slides/:id", handler.SlideUpdateHandler)
			website.DELETE("/slides/:id", handler.SlideDeleteHandler)
			website.PUT("/slides/:id/status", handler.SlideStatusUpdateHandler)
			website.POST("/navigations", handler.CreateNavigationHandler)
			website.PUT("/navigations/:id", handler.NavigationUpdateHandler)
			website.DELETE("/navigations/:id", handler.NavigationDeleteHandler)
			website.PUT("/navigations/:id/status", handler.NavigationStatusUpdateHandler)
			website.POST("/links", handler.CreateLinkHandler)
			website.PUT("/links/:id", handler.LinkUpdateHandler)
			website.DELETE("/links/:id", handler.LinkDeleteHandler)
			website.PUT("/links/:id/status", handler.LinkStatusUpdateHandler)
			website.POST("/payments", handler.CreatePaymentHandler)
			website.PUT("/payments/:id", handler.PaymentUpdateHandler)
			website.DELETE("/payments/:id", handler.PaymentDeleteHandler)
			website.PUT("/payments/:id/status", handler.PaymentStatusUpdateHandler)
			website.POST("/custom-views", handler.CustomViewCreateHandler)
			website.PUT("/custom-views/:id", handler.CustomViewUpdateHandler)
			website.DELETE("/custom-views/:id", handler.CustomViewDeleteHandler)
			website.PUT("/custom-views/:id/status", handler.CustomViewStatusUpdateHandler)
			website.POST("/quick-nav", handler.QuickNavCreateHandler)
			website.GET("/quick-nav", handler.QuickNavListHandler)
			website.PUT("/quick-nav/:id", handler.QuickNavUpdateHandler)
			website.DELETE("/quick-nav/:id", handler.QuickNavDeleteHandler)
			website.PUT("/quick-nav/:id/status", handler.QuickNavStatusUpdateHandler)
			website.GET("/shortcut-menus", handler.ShortcutMenuListHandler)
			website.PUT("/shortcut-menus", handler.ShortcutMenuSaveHandler)
			website.POST("/agreement", handler.AgreementSaveHandler)
			website.GET("/attachments", handler.GetAttachmentList)
			website.DELETE("/attachments/:id", handler.AttachmentDeleteHandler)
			website.GET("/attachment-categories", handler.GetAttachmentCategoryList)
			website.POST("/attachment-categories", handler.CreateAttachmentCategoryHandler)
			website.DELETE("/attachment-categories/:id", handler.AttachmentCategoryDeleteHandler)
			website.POST("/upload", handler.Upload)
			website.POST("/express", handler.CreateExpressHandler)
			website.GET("/express", handler.GetExpressList)
			website.PUT("/express/:id", handler.ExpressUpdateHandler)
			website.DELETE("/express/:id", handler.ExpressDeleteHandler)
			website.POST("/regions", handler.RegionSaveHandler)
			website.DELETE("/regions/:id", handler.RegionDeleteHandler)
			website.DELETE("/screening-prices/:id", handler.ScreeningPriceDeleteHandler)
			website.GET("/diy", handler.DiyListHandler)
			website.POST("/diy", handler.DiyCreateHandler)
			website.PUT("/diy/:id", handler.DiyUpdateHandler)
			website.DELETE("/diy/:id", handler.DiyDeleteHandler)
			website.GET("/designs", handler.DesignListHandler)
			website.POST("/designs", handler.DesignCreateHandler)
			website.PUT("/designs/:id", handler.DesignUpdateHandler)
			website.GET("/layouts", handler.LayoutListHandler)
			website.POST("/layouts", handler.LayoutSaveHandler)
			website.GET("/themes", handler.ThemeListHandler)
			website.POST("/themes", handler.ThemeCreateHandler)
			website.POST("/themes/upload", handler.ThemeUploadHandler)
			website.GET("/forms", handler.FormInputListHandler)
			website.POST("/forms", handler.FormInputCreateHandler)
			website.DELETE("/forms/:id", handler.FormInputDeleteHandler)
			website.GET("/form-data", handler.FormInputDataListHandler)
			website.PUT("/forms/:id/fields", handler.SaveFormFieldsHandler)
			website.GET("/forms/:id/fields", handler.GetFormFieldsHandler)
		}

		// ========== 品牌管理 (Brand.Index) ==========
		brand := admin.Group("").Use(middleware.AdminPower("Brand.Index"))
		{
			brand.POST("/brands", handler.CreateBrand)
			brand.PUT("/brands/:id", handler.BrandUpdateHandler)
			brand.DELETE("/brands/:id", handler.BrandDeleteHandler)
			brand.PUT("/brands/:id/status", handler.BrandStatusUpdateHandler)
			brand.GET("/brands/:id", handler.BrandDetailHandler)
			brand.POST("/brand-categories", handler.CreateBrandCategory)
		}

		// ========== 仓库管理 (Warehouse.Index) ==========
		wh := admin.Group("").Use(middleware.AdminPower("Warehouse.Index"))
		{
			wh.POST("/warehouses", handler.CreateWarehouse)
			wh.GET("/warehouses", handler.GetWarehouseList)
			wh.PUT("/warehouses/:id", handler.UpdateWarehouse)
			wh.DELETE("/warehouses/:id", handler.DeleteWarehouse)
			wh.POST("/warehouse-goods", handler.WarehouseGoodsAdd)
			wh.GET("/warehouses/:id/goods", handler.WarehouseGoodsList)
			wh.POST("/warehouse-goods-spec", handler.WarehouseGoodsSpecSave)
			wh.DELETE("/warehouse-goods/:id", handler.WarehouseGoodsDeleteHandler)
			wh.PUT("/warehouse-goods/:id/status", handler.WarehouseGoodsStatusUpdateHandler)
			wh.GET("/inventory-logs", handler.GetInventoryLogList)
		}

		// ========== 文章管理 (Article.Index) ==========
		article := admin.Group("").Use(middleware.AdminPower("Article.Index"))
		{
			article.POST("/article-categories", handler.CreateArticleCategory)
			article.DELETE("/article-categories/:id", handler.ArticleCategoryDeleteHandler)
			article.POST("/articles", handler.CreateArticle)
			article.PUT("/articles/:id", handler.ArticleUpdateHandler)
			article.DELETE("/articles/:id", handler.ArticleDeleteHandler)
			article.PUT("/articles/:id/status", handler.ArticleStatusUpdateHandler)
			article.GET("/articles/:id", handler.ArticleDetailHandler)
		}

		// ========== APP管理 (App.Index) ==========
		app := admin.Group("").Use(middleware.AdminPower("App.Index"))
		{
			app.POST("/app/home-nav", handler.AppHomeNavCreateHandler)
			app.POST("/app/center-nav", handler.AppCenterNavCreateHandler)
			app.PUT("/app/tabbar", handler.AppTabbarSaveHandler)
			app.POST("/app-mini", handler.SaveAppMini)
			app.GET("/app-mini", handler.GetAppMiniList)
			app.DELETE("/app-mini/:id", handler.DeleteAppMini)
		}

		// ========== 数据/日志 (Data.Index) ==========
		data := admin.Group("").Use(middleware.AdminPower("Data.Index"))
		{
			data.GET("/messages", handler.AdminMessageList)
			data.DELETE("/messages/:id", handler.MessageDeleteHandler)
			data.GET("/pay-logs", handler.AdminPayLogList)
			data.GET("/pay-logs/:id", handler.PayLogDetailHandler)
			data.PUT("/pay-logs/:id/close", handler.PayLogCloseHandler)
			data.GET("/pay-request-logs", handler.AdminPayRequestLogList)
			data.GET("/integral-logs", handler.AdminIntegralLogList)
			data.GET("/refund-logs", handler.AdminGetRefundLogList)
			data.GET("/error-logs", handler.GetErrorLogList)
			data.DELETE("/error-logs/:id", handler.ErrorLogDeleteHandler)
			data.DELETE("/error-logs-all", handler.ErrorLogAllDeleteHandler)
			data.GET("/sms-logs", handler.SmsLogListHandler)
			data.DELETE("/sms-logs", handler.SmsLogDeleteHandler)
			data.GET("/email-logs", handler.EmailLogListHandler)
			data.DELETE("/email-logs", handler.EmailLogDeleteHandler)
			data.GET("/search-history", handler.AdminSearchHistoryList)
			data.DELETE("/search-history/:id", handler.SearchHistoryDeleteHandler)
			data.DELETE("/search-history-all", handler.SearchHistoryAllDeleteHandler)
			data.GET("/operation-logs", handler.AdminOperationLogList)
			data.GET("/goods-browse", handler.AdminGoodsBrowseList)
			data.DELETE("/goods-browse/:id", handler.AdminGoodsBrowseDelete)
			data.GET("/goods-favor", handler.AdminGoodsFavorList)
			data.DELETE("/goods-favor/:id", handler.AdminGoodsFavorDelete)
			data.GET("/goods-cart", handler.AdminGoodsCartList)
			data.DELETE("/goods-cart/:id", handler.AdminGoodsCartDelete)
		}

		// ========== 应用/插件 (Store.Index) ==========
		store := admin.Group("").Use(middleware.AdminPower("Store.Index"))
		{
			store.GET("/plugins", handler.PluginList)
			store.POST("/plugins", handler.PluginInstall)
			store.PUT("/plugins/:id/uninstall", handler.PluginUninstall)
			store.GET("/plugin-config", handler.PluginConfigGetHandler)
			store.POST("/plugin-config", handler.PluginConfigSetHandler)
		}

		// ========== 系统配置 (Config.Index) ==========
		cfg := admin.Group("").Use(middleware.AdminPower("Config.Index"))
		{
			cfg.GET("/config", handler.GetConfigGroup)
			cfg.POST("/config", handler.SetConfigHandler)
			cfg.GET("/multilingual", handler.GetMultilingualConfig)
			cfg.POST("/multilingual", handler.SetMultilingualConfig)
			cfg.GET("/currency", handler.GetCurrencyConfig)
			cfg.POST("/currency", handler.SetCurrencyConfig)
			cfg.GET("/site-config", handler.GetSiteConfigHandler)
			cfg.POST("/site-config", handler.SaveSiteConfigHandler)
			cfg.GET("/self-extraction-address", handler.GetSelfExtractionAddress)
			cfg.POST("/self-extraction-address", handler.SaveSelfExtractionAddress)
		}

		// ========== 站点设置 (Site.Index) ==========
		site := admin.Group("").Use(middleware.AdminPower("Site.Index"))
		{
			site.GET("/seo", handler.SeoGetHandler)
			site.POST("/seo", handler.SeoSaveHandler)
			site.POST("/email-test", handler.EmailTestHandler)
		}

		// ========== 工具 (Tool.Index) ==========
		tool := admin.Group("").Use(middleware.AdminPower("Tool.Index"))
		{
			tool.POST("/cache/clear", handler.ClearCache)
			tool.GET("/cache/stats", handler.GetCacheStats)
			tool.POST("/sql-console", handler.SqlConsoleExecute)
		}

		// ========== 问答管理 ==========
		admin.GET("/answers", handler.GetAnswerList)
		admin.PUT("/answers/:id/reply", handler.AdminReplyAnswer)
		admin.DELETE("/answers/:id", handler.AdminDeleteAnswer)

		// ========== 分销管理 ==========
		admin.GET("/distributors", handler.AdminDistributorList)
		admin.POST("/distributors", handler.AdminCreateDistributor)
		admin.GET("/withdraws", handler.AdminWithdrawList)
		admin.PUT("/withdraws/:id/audit", handler.AdminAuditWithdraw)

		// ========== 数据导出 ==========
		admin.POST("/export", handler.ExportData)
		admin.POST("/form-table", handler.FormTableQueryHandler)

		// 公开列表接口的admin镜像（供管理后台CrudPage下拉选择用，仅需登录）
		admin.GET("/article-categories", handler.GetArticleCategoryList)
		admin.GET("/brand-categories", handler.GetBrandCategoryList)
		admin.GET("/links", handler.GetLinkList)
		admin.GET("/screening-prices", handler.GetScreeningPrices)
		admin.GET("/slides", handler.GetSlideList)
		admin.GET("/navigations", handler.GetNavigationList)
		admin.GET("/brands", handler.GetBrandList)
		admin.GET("/coupons", handler.GetCouponList)
		admin.GET("/custom-views", handler.CustomViewListHandler)
		admin.GET("/promotions", handler.AdminPromotionListHandler)
		admin.GET("/articles", handler.GetArticleList)
		admin.GET("/app/home-nav", handler.AppHomeNavListHandler)
		admin.GET("/app/center-nav", handler.AppCenterNavListHandler)
	}
}
