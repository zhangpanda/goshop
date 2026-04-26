package router

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/handler"
	"github.com/zhangpanda/goshop/internal/middleware"
)

func Setup(r *gin.Engine) {
	r.Use(middleware.Cors(), middleware.Logger(), gin.Recovery())

	// 静态文件
	r.Static("/uploads", "./uploads")
	r.Static("/static/diy", "./static/diy")
	r.Static("/static/form_input", "./static/form_input")
	r.StaticFile("/diy.html", "./static/diy.html")
	r.StaticFile("/form.html", "./static/form.html")

	// ShopXO uni-app兼容路由
	handler.SetupShopXOCompat(r)

	// DIY/Form 编辑器兼容路由
	handler.SetupDiyApiCompat(r)

	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/wx/login", handler.WxLogin)
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

	// 公开安全接口
	api.POST("/verify-code", handler.SendVerifyCode)
	api.POST("/forget-password", handler.ForgetPassword)

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
	api.POST("/admin/login", handler.AdminLoginHandler)
	api.GET("/admin/captcha", handler.AdminCaptcha)

	// 后台管理（使用 AdminAuth 中间件 + 操作日志）
	admin := api.Group("/admin").Use(middleware.AdminAuth(), middleware.AdminOperationLog())
	{
		admin.GET("/dashboard", handler.AdminDashboard)

		// 管理员管理
		admin.POST("/admins", handler.CreateAdminHandler)
		admin.GET("/admins", handler.GetAdminListHandler)
		admin.PUT("/admins/:id/status", handler.UpdateAdminStatusHandler)

		// 角色管理
		admin.POST("/roles", handler.CreateRoleHandler)
		admin.GET("/roles", handler.GetRoleListHandler)
		admin.PUT("/roles/:id", handler.UpdateRoleHandler)
		admin.DELETE("/roles/:id", handler.DeleteRoleHandler)

		// 商品管理
		admin.POST("/goods", handler.CreateGoods)
		admin.PUT("/goods/:id", handler.AdminUpdateGoods)
		admin.DELETE("/goods/:id", handler.AdminDeleteGoods)
		admin.PUT("/goods/:id/status", handler.AdminToggleGoodsStatus)

		// 分类管理
		admin.POST("/categories", handler.CreateCategory)
		admin.PUT("/categories/:id", handler.AdminUpdateCategory)
		admin.DELETE("/categories/:id", handler.AdminDeleteCategory)

		// 订单管理
		admin.GET("/orders", handler.AdminGetOrders)
		admin.PUT("/orders/:id/remark", handler.AdminUpdateOrderRemark)
		admin.POST("/orders/ship", handler.ShipOrder)

		// 用户管理
		admin.GET("/users", handler.AdminGetUsers)
		admin.PUT("/users/:id/status", handler.AdminUpdateUserStatus)

		// 优惠券管理
		admin.POST("/coupons", handler.CreateCoupon)
		admin.PUT("/coupons/:id", handler.CouponUpdateHandler)
		admin.DELETE("/coupons/:id", handler.CouponDeleteHandler)

		// 促销管理
		admin.POST("/promotions", handler.CreatePromotion)
		admin.PUT("/promotions/:id", handler.PromotionUpdateHandler)
		admin.DELETE("/promotions/:id", handler.PromotionDeleteHandler)

		// 评价管理
		admin.PUT("/reviews/:id/reply", handler.ReplyReview)

		// 售后管理
		admin.GET("/aftersale", handler.AdminAftersaleList)
		admin.PUT("/aftersale/:id/confirm", handler.AdminAftersaleConfirm)
		admin.PUT("/aftersale/:id/audit", handler.AdminAftersaleAudit)
		admin.PUT("/aftersale/:id/refuse", handler.AdminAftersaleRefuse)

		// 品牌管理
		admin.POST("/brands", handler.CreateBrand)

		// 文章管理
		admin.POST("/article-categories", handler.CreateArticleCategory)
		admin.POST("/articles", handler.CreateArticle)

		// 规格模板
		admin.POST("/spec-templates", handler.CreateSpecTemplate)
		admin.GET("/spec-templates", handler.GetSpecTemplateList)

		// 参数模板
		admin.POST("/params-templates", handler.CreateParamsTemplate)
		admin.GET("/params-templates", handler.GetParamsTemplateList)
		admin.PUT("/goods/:id/params", handler.SaveGoodsParams)
		admin.PUT("/goods/:id/specs", handler.SaveGoodsSpecBase)
		admin.PUT("/goods/:id/photos", handler.SaveGoodsPhotos)

		// 系统配置
		admin.GET("/config", handler.GetConfigGroup)
		admin.POST("/config", handler.SetConfigHandler)

		// 幻灯片/导航/链接
		admin.POST("/slides", handler.CreateSlideHandler)
		admin.POST("/navigations", handler.CreateNavigationHandler)
		admin.POST("/links", handler.CreateLinkHandler)

		// 支付方式
		admin.POST("/payments", handler.CreatePaymentHandler)

		// 附件管理
		admin.GET("/attachments", handler.GetAttachmentList)
		admin.GET("/attachment-categories", handler.GetAttachmentCategoryList)
		admin.POST("/attachment-categories", handler.CreateAttachmentCategoryHandler)

		// 错误日志
		admin.GET("/error-logs", handler.GetErrorLogList)

		// 退款日志
		admin.GET("/refund-logs", handler.AdminGetRefundLogList)

		// 仓库管理
		admin.POST("/warehouses", handler.CreateWarehouse)
		admin.GET("/warehouses", handler.GetWarehouseList)
		admin.PUT("/warehouses/:id", handler.UpdateWarehouse)
		admin.DELETE("/warehouses/:id", handler.DeleteWarehouse)
		admin.POST("/warehouse-goods", handler.WarehouseGoodsAdd)
		admin.GET("/warehouses/:id/goods", handler.WarehouseGoodsList)
		admin.POST("/warehouse-goods-spec", handler.WarehouseGoodsSpecSave)

		// 快递公司
		admin.POST("/express", handler.CreateExpressHandler)
		admin.GET("/express", handler.GetExpressList)

		// 库存日志
		admin.GET("/inventory-logs", handler.GetInventoryLogList)

		// 权限树
		admin.POST("/powers", handler.CreatePowerHandler)
		admin.GET("/powers", handler.GetPowerTree)
		admin.PUT("/roles/:id/powers", handler.SaveRolePowers)
		admin.GET("/roles/:id/powers", handler.GetRolePowersHandler)

		// 品牌分类
		admin.POST("/brand-categories", handler.CreateBrandCategory)

		// 商品多分类
		admin.PUT("/goods/:id/categories", handler.SaveGoodsCategoryJoin)

		// 插件
		admin.GET("/plugins", handler.PluginList)
		admin.POST("/plugins", handler.PluginInstall)
		admin.PUT("/plugins/:id/uninstall", handler.PluginUninstall)

		// DIY页面
		admin.GET("/diy", handler.DiyListHandler)
		admin.POST("/diy", handler.DiyCreateHandler)
		admin.PUT("/diy/:id", handler.DiyUpdateHandler)
		admin.DELETE("/diy/:id", handler.DiyDeleteHandler)

		// 自定义页面
		admin.POST("/custom-views", handler.CustomViewCreateHandler)

		// 主题
		admin.GET("/themes", handler.ThemeListHandler)
		admin.POST("/themes", handler.ThemeCreateHandler)
		admin.POST("/themes/upload", handler.ThemeUploadHandler)
		admin.POST("/upload", handler.Upload)

		// 表单
		admin.GET("/forms", handler.FormInputListHandler)
		admin.POST("/forms", handler.FormInputCreateHandler)
		admin.DELETE("/forms/:id", handler.FormInputDeleteHandler)
		admin.GET("/form-data", handler.FormInputDataListHandler)

		// APP导航
		admin.POST("/app/home-nav", handler.AppHomeNavCreateHandler)
		admin.POST("/app/center-nav", handler.AppCenterNavCreateHandler)
		admin.PUT("/app/tabbar", handler.AppTabbarSaveHandler)

		// 快捷菜单
		admin.GET("/shortcut-menus", handler.ShortcutMenuListHandler)
		admin.PUT("/shortcut-menus", handler.ShortcutMenuSaveHandler)

		// 协议
		admin.POST("/agreement", handler.AgreementSaveHandler)

		// SEO
		admin.GET("/seo", handler.SeoGetHandler)
		admin.POST("/seo", handler.SeoSaveHandler)

		// Design
		admin.GET("/designs", handler.DesignListHandler)
		admin.POST("/designs", handler.DesignCreateHandler)
		admin.PUT("/designs/:id", handler.DesignUpdateHandler)

		// Layout
		admin.GET("/layouts", handler.LayoutListHandler)
		admin.POST("/layouts", handler.LayoutSaveHandler)

		// APP商品详情
		admin.PUT("/goods/:id/content-app", handler.SaveGoodsContentAppHandler)

		// 订单客服
		admin.GET("/order-service", handler.AdminOrderServiceList)
		admin.PUT("/order-service/:id/reply", handler.AdminReplyOrderService)

		// 快捷导航
		admin.POST("/quick-nav", handler.QuickNavCreateHandler)
		admin.GET("/quick-nav", handler.QuickNavListHandler)

		// 插件配置
		admin.GET("/plugin-config", handler.PluginConfigGetHandler)
		admin.POST("/plugin-config", handler.PluginConfigSetHandler)

		// 角色插件权限
		admin.PUT("/roles/:id/plugins", handler.SaveRolePluginsHandler)

		// 表单字段
		admin.PUT("/forms/:id/fields", handler.SaveFormFieldsHandler)
		admin.GET("/forms/:id/fields", handler.GetFormFieldsHandler)

		// 问答留言管理
		admin.GET("/answers", handler.GetAnswerList)
		admin.PUT("/answers/:id/reply", handler.AdminReplyAnswer)
		admin.DELETE("/answers/:id", handler.AdminDeleteAnswer)

		// 订单预约确认
		admin.PUT("/orders/booking-confirm", handler.AdminBookingConfirm)

		// 数据导出
		admin.POST("/export", handler.ExportData)

		// 缓存管理
		admin.POST("/cache/clear", handler.ClearCache)
		admin.GET("/cache/stats", handler.GetCacheStats)

		// 多语言配置
		admin.POST("/multilingual", handler.SetMultilingualConfig)

		// 货币配置
		admin.POST("/currency", handler.SetCurrencyConfig)

		// 完整统计
		admin.GET("/statistical", handler.AdminStatistical)

		// 小程序管理
		admin.POST("/app-mini", handler.SaveAppMini)
		admin.GET("/app-mini", handler.GetAppMiniList)
		admin.DELETE("/app-mini/:id", handler.DeleteAppMini)

		// 站点配置
		admin.GET("/site-config", handler.GetSiteConfigHandler)
		admin.POST("/site-config", handler.SaveSiteConfigHandler)
		admin.GET("/self-extraction-address", handler.GetSelfExtractionAddress)
		admin.POST("/self-extraction-address", handler.SaveSelfExtractionAddress)

		// 动态表格通用查询
		admin.POST("/form-table", handler.FormTableQueryHandler)

		// SQL控制台
		admin.POST("/sql-console", handler.SqlConsoleExecute)

		// 系统信息
		admin.GET("/system-info", handler.GetSystemInfo)

		// 公开列表接口的admin镜像（供管理后台CrudPage使用）
		admin.GET("/article-categories", handler.GetArticleCategoryList)
		admin.GET("/brand-categories", handler.GetBrandCategoryList)
		admin.GET("/links", handler.GetLinkList)
		admin.GET("/screening-prices", handler.GetScreeningPrices)
		admin.GET("/slides", handler.GetSlideList)
		admin.GET("/navigations", handler.GetNavigationList)
		admin.GET("/brands", handler.GetBrandList)
		admin.GET("/coupons", handler.GetCouponList)
		admin.GET("/custom-views", handler.CustomViewListHandler)
		admin.GET("/promotions", handler.GetActivePromotions)
		admin.GET("/articles", handler.GetArticleList)
		admin.GET("/app/home-nav", handler.AppHomeNavListHandler)
		admin.GET("/app/center-nav", handler.AppCenterNavListHandler)

		// 管理员确认线下收款
		admin.PUT("/orders/pay-underline", handler.AdminOrderPayUnderLineHandler)

		// 短信日志
		admin.GET("/sms-logs", handler.SmsLogListHandler)
		admin.DELETE("/sms-logs", handler.SmsLogDeleteHandler)

		// 邮件日志
		admin.GET("/email-logs", handler.EmailLogListHandler)
		admin.DELETE("/email-logs", handler.EmailLogDeleteHandler)

		// ========== 补全的CRUD操作 ==========

		// 管理员补全
		admin.GET("/admins/:id", handler.AdminDetailHandler)
		admin.DELETE("/admins/:id", handler.AdminDeleteHandler)

		// 角色补全
		admin.GET("/roles/:id", handler.RoleDetailHandler)
		admin.PUT("/roles/:id/status", handler.RoleStatusUpdateHandler)

		// 分类状态
		admin.PUT("/categories/:id/status", handler.CategoryStatusUpdate)

		// 评论补全
		admin.DELETE("/reviews/:id", handler.ReviewDeleteHandler)

		// 订单补全
		admin.PUT("/orders/:id/cancel", handler.AdminCancelOrder)
		admin.PUT("/orders/:id/confirm", handler.AdminConfirmReceive)
		admin.DELETE("/orders/:id", handler.AdminDeleteOrder)

		// 售后补全
		admin.DELETE("/aftersale/:id", handler.AdminAftersaleDelete)
		admin.PUT("/aftersale/:id/cancel", handler.AdminAftersaleCancel)

		// 品牌补全
		admin.PUT("/brands/:id", handler.BrandUpdateHandler)
		admin.DELETE("/brands/:id", handler.BrandDeleteHandler)
		admin.PUT("/brands/:id/status", handler.BrandStatusUpdateHandler)
		admin.GET("/brands/:id", handler.BrandDetailHandler)

		// 文章补全
		admin.PUT("/articles/:id", handler.ArticleUpdateHandler)
		admin.DELETE("/articles/:id", handler.ArticleDeleteHandler)
		admin.PUT("/articles/:id/status", handler.ArticleStatusUpdateHandler)
		admin.GET("/articles/:id", handler.ArticleDetailHandler)
		admin.DELETE("/article-categories/:id", handler.ArticleCategoryDeleteHandler)

		// 幻灯片补全
		admin.PUT("/slides/:id", handler.SlideUpdateHandler)
		admin.DELETE("/slides/:id", handler.SlideDeleteHandler)
		admin.PUT("/slides/:id/status", handler.SlideStatusUpdateHandler)

		// 导航补全
		admin.PUT("/navigations/:id", handler.NavigationUpdateHandler)
		admin.DELETE("/navigations/:id", handler.NavigationDeleteHandler)
		admin.PUT("/navigations/:id/status", handler.NavigationStatusUpdateHandler)

		// 友情链接补全
		admin.PUT("/links/:id", handler.LinkUpdateHandler)
		admin.DELETE("/links/:id", handler.LinkDeleteHandler)
		admin.PUT("/links/:id/status", handler.LinkStatusUpdateHandler)

		// 快递补全
		admin.PUT("/express/:id", handler.ExpressUpdateHandler)
		admin.DELETE("/express/:id", handler.ExpressDeleteHandler)

		// 支付方式补全
		admin.PUT("/payments/:id", handler.PaymentUpdateHandler)
		admin.DELETE("/payments/:id", handler.PaymentDeleteHandler)
		admin.PUT("/payments/:id/status", handler.PaymentStatusUpdateHandler)

		// 自定义页面补全
		admin.PUT("/custom-views/:id", handler.CustomViewUpdateHandler)
		admin.DELETE("/custom-views/:id", handler.CustomViewDeleteHandler)
		admin.PUT("/custom-views/:id/status", handler.CustomViewStatusUpdateHandler)

		// 快捷导航补全
		admin.PUT("/quick-nav/:id", handler.QuickNavUpdateHandler)
		admin.DELETE("/quick-nav/:id", handler.QuickNavDeleteHandler)
		admin.PUT("/quick-nav/:id/status", handler.QuickNavStatusUpdateHandler)

		// 附件补全
		admin.DELETE("/attachments/:id", handler.AttachmentDeleteHandler)
		admin.DELETE("/attachment-categories/:id", handler.AttachmentCategoryDeleteHandler)

		// 搜索记录
		admin.GET("/search-history", handler.AdminSearchHistoryList)
		admin.DELETE("/search-history/:id", handler.SearchHistoryDeleteHandler)
		admin.DELETE("/search-history-all", handler.SearchHistoryAllDeleteHandler)

		// 错误日志补全
		admin.DELETE("/error-logs/:id", handler.ErrorLogDeleteHandler)
		admin.DELETE("/error-logs-all", handler.ErrorLogAllDeleteHandler)

		// 消息补全
		admin.GET("/messages", handler.AdminMessageList)
		admin.DELETE("/messages/:id", handler.MessageDeleteHandler)

		// 支付日志补全
		admin.GET("/pay-logs", handler.AdminPayLogList)
		admin.GET("/pay-logs/:id", handler.PayLogDetailHandler)
		admin.PUT("/pay-logs/:id/close", handler.PayLogCloseHandler)

		// 积分日志
		admin.GET("/integral-logs", handler.AdminIntegralLogList)

		// 支付请求日志
		admin.GET("/pay-request-logs", handler.AdminPayRequestLogList)

		// 地区补全
		admin.POST("/regions", handler.RegionSaveHandler)
		admin.DELETE("/regions/:id", handler.RegionDeleteHandler)

		// 筛选价格补全
		admin.DELETE("/screening-prices/:id", handler.ScreeningPriceDeleteHandler)

		// 用户地址管理
		admin.GET("/user-address", handler.UserAddressListHandler)
		admin.GET("/user-address/:id", handler.UserAddressDetailHandler)
		admin.PUT("/user-address/:id", handler.UserAddressSaveHandler)
		admin.DELETE("/user-address/:id", handler.UserAddressDeleteHandler)

		// 仓库商品补全
		admin.DELETE("/warehouse-goods/:id", handler.WarehouseGoodsDeleteHandler)
		admin.PUT("/warehouse-goods/:id/status", handler.WarehouseGoodsStatusUpdateHandler)

		// 商品浏览/收藏/购物车管理
		admin.GET("/goods-browse", handler.AdminGoodsBrowseList)
		admin.DELETE("/goods-browse/:id", handler.AdminGoodsBrowseDelete)
		admin.GET("/goods-favor", handler.AdminGoodsFavorList)
		admin.DELETE("/goods-favor/:id", handler.AdminGoodsFavorDelete)
		admin.GET("/goods-cart", handler.AdminGoodsCartList)
		admin.DELETE("/goods-cart/:id", handler.AdminGoodsCartDelete)

		// 邮件测试
		admin.POST("/email-test", handler.EmailTestHandler)

		// 操作审计日志
		admin.GET("/operation-logs", handler.AdminOperationLogList)

		// 物流轨迹查询
		admin.GET("/orders/:id/logistics", handler.LogisticsTrackHandler)
	}
}
