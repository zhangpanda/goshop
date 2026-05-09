package model

// AllModels 返回本项目所有持久化模型的零值指针集合，供 gorm.AutoMigrate / testutil 共用。
//
// 单一入口避免 initialize 和 testutil 两边维护两份同步漂移的清单；同时本文件仅
// 依赖同包模型，不引入 service/compat 等子包，避免 testutil 用它时产生循环依赖。
//
// 新增业务模型请在本函数末尾追加。
func AllModels() []any {
	return []any{
		&User{}, &Category{}, &Goods{}, &GoodsSKU{},
		&Cart{}, &Address{}, &Order{}, &OrderItem{}, &OrderStatusHistory{},
		&Coupon{}, &UserCoupon{}, &Promotion{}, &PromotionItem{},
		&Favorite{}, &BrowseHistory{}, &Review{},
		&Shipment{}, &PointsLog{}, &Message{},
		&Admin{}, &Role{},
		&OrderAftersale{}, &AftersaleHistory{},
		&Brand{}, &Article{}, &ArticleCategory{},
		&SpecTemplate{}, &SpecType{}, &SpecValue{},
		&GoodsParamsTemplate{}, &GoodsParamsConfig{}, &GoodsParams{},
		&SearchHistory{}, &ScreeningPrice{},
		&Config{}, &Region{}, &Slide{}, &Navigation{}, &Link{},
		&Payment{}, &SmsLog{}, &EmailLog{},
		&Attachment{}, &AttachmentCategory{}, &ErrorLog{},
		&UserPlatform{}, &VerifyCode{},
		&PayLog{}, &PayRequestLog{}, &RefundLog{},
		&GoodsSpecBase{}, &GoodsPhoto{},
		&Warehouse{}, &WarehouseGoods{}, &WarehouseGoodsSpec{},
		&Express{}, &InventoryLog{}, &GoodsGiveIntegralLog{},
		&BrandCategory{}, &BrandCategoryJoin{}, &GoodsCategoryJoin{},
		&Power{}, &RolePower{},
		&Plugin{}, &PluginCategory{},
		&Diy{}, &CustomView{}, &ThemeData{},
		&FormInput{}, &FormInputData{},
		&AppHomeNav{}, &AppCenterNav{}, &AppTabbar{},
		&ShortcutMenu{}, &Agreement{},
		&OrderTraceSource{}, &OrderCurrency{},
		&Design{}, &FormTableUserFields{}, &GoodsContentApp{},
		&Layout{}, &OrderService{}, &PayLogValue{},
		&PluginsDataConfig{}, &QuickNav{}, &RolePlugins{},
		&Answer{},
		&AppMini{}, &WalletLog{},
		&AdminOperationLog{},
		&GroupOrder{}, &GroupOrderMember{},
		&Distributor{}, &CommissionLog{}, &WithdrawRequest{},
	}
}
