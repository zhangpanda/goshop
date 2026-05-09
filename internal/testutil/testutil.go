package testutil

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/repository"
	"github.com/zhangpanda/goshop/pkg/cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB 将内存 SQLite 与内存缓存注册到 app.Deps（TestMain 或单测开头调用）。
func SetupTestDB() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to open test db: " + err.Error())
	}
	db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Goods{}, &model.GoodsSKU{},
		&model.Cart{}, &model.Address{}, &model.Order{}, &model.OrderItem{},
		&model.Coupon{}, &model.UserCoupon{}, &model.Favorite{},
		&model.PointsLog{}, &model.Message{},
		&model.Shipment{}, &model.OrderAftersale{},
		&model.GoodsCategoryJoin{}, &model.GoodsContentApp{},
		&model.GoodsSpecBase{}, &model.GoodsPhoto{}, &model.GoodsParams{},
		&model.OrderStatusHistory{}, &model.OrderCurrency{},
		&model.Promotion{}, &model.PromotionItem{},
		&model.Config{},
		// handler/service 测试常用到的只读元数据
		&model.Article{}, &model.ArticleCategory{},
		&model.Slide{}, &model.Navigation{}, &model.Payment{},
		&model.Review{}, &model.Brand{},
	)
	app.Register(&app.Deps{
		DB:    db,
		Cache: cache.NewMemoryCache(),
	})
	repository.Init(db)
}
