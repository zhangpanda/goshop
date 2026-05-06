package testutil

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/repository"
	"github.com/zhangpanda/goshop/pkg/cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB initializes global.DB with an in-memory SQLite database
// and global.Cache with an in-memory cache. Call in TestMain or individual tests.
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
	)
	global.DB = db
	global.Cache = cache.NewMemoryCache()
	repository.Init(db)
}
