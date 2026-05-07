package initialize

import (
	"errors"

	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// RunSchemaAutoMigrate 与 HTTP 服务启动时一致的 Schema 迁移：先拼团成员去重，再 AutoMigrate 全部注册模型。
// 可在 Job/CI 中独立调用（见 cmd/migrate）。
func RunSchemaAutoMigrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if err := DedupeGroupOrderMembersBeforeUniqueIndex(db); err != nil {
		return err
	}
	return db.AutoMigrate(autoMigrateModelList()...)
}

func autoMigrateModelList() []any {
	return []any{
		&model.User{}, &model.Category{}, &model.Goods{}, &model.GoodsSKU{},
		&model.Cart{}, &model.Address{}, &model.Order{}, &model.OrderItem{}, &model.OrderStatusHistory{},
		&model.Coupon{}, &model.UserCoupon{}, &model.Promotion{}, &model.PromotionItem{},
		&model.Favorite{}, &model.BrowseHistory{}, &model.Review{},
		&model.Shipment{}, &model.PointsLog{}, &model.Message{},
		&model.Admin{}, &model.Role{},
		&model.OrderAftersale{}, &model.AftersaleHistory{},
		&model.Brand{}, &model.Article{}, &model.ArticleCategory{},
		&model.SpecTemplate{}, &model.SpecType{}, &model.SpecValue{},
		&model.GoodsParamsTemplate{}, &model.GoodsParamsConfig{}, &model.GoodsParams{},
		&model.SearchHistory{}, &model.ScreeningPrice{},
		&model.Config{}, &model.Region{}, &model.Slide{}, &model.Navigation{}, &model.Link{},
		&model.Payment{}, &model.SmsLog{}, &model.EmailLog{},
		&model.Attachment{}, &model.AttachmentCategory{}, &model.ErrorLog{},
		&model.UserPlatform{}, &model.VerifyCode{},
		&model.PayLog{}, &model.PayRequestLog{}, &model.RefundLog{},
		&model.GoodsSpecBase{}, &model.GoodsPhoto{},
		&model.Warehouse{}, &model.WarehouseGoods{}, &model.WarehouseGoodsSpec{},
		&model.Express{}, &model.InventoryLog{}, &model.GoodsGiveIntegralLog{},
		&model.BrandCategory{}, &model.BrandCategoryJoin{}, &model.GoodsCategoryJoin{},
		&model.Power{}, &model.RolePower{},
		&model.Plugin{}, &model.PluginCategory{},
		&model.Diy{}, &model.CustomView{}, &model.ThemeData{},
		&model.FormInput{}, &model.FormInputData{},
		&model.AppHomeNav{}, &model.AppCenterNav{}, &model.AppTabbar{},
		&model.ShortcutMenu{}, &model.Agreement{},
		&model.OrderTraceSource{}, &model.OrderCurrency{},
		&model.Design{}, &model.FormTableUserFields{}, &model.GoodsContentApp{},
		&model.Layout{}, &model.OrderService{}, &model.PayLogValue{},
		&model.PluginsDataConfig{}, &model.QuickNav{}, &model.RolePlugins{},
		&model.Answer{},
		&model.AppMini{}, &model.WalletLog{},
		&model.AdminOperationLog{},
		&model.GroupOrder{}, &model.GroupOrderMember{},
		&model.Distributor{}, &model.CommissionLog{}, &model.WithdrawRequest{},
	}
}
