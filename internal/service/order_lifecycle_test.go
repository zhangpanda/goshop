package service

import (
	"testing"
	"time"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/repository"
	"github.com/zhangpanda/goshop/pkg/cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupLifecycleDB 创建 shared-cache SQLite 内存库，支持多 goroutine 访问同一数据。
func setupLifecycleDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(model.AllModels()...)
	app.Register(&app.Deps{
		Cfg:   &config.Config{},
		DB:    db,
		Cache: cache.NewMemoryCache(),
	})
	repository.Init(db)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		// 等 SafeGo goroutine 完成
		time.Sleep(50 * time.Millisecond)
	})
}

// TestOrderLifecycle 核心链路集成测试：下单→支付→发货→收货→佣金结算。
// 使用 SQLite 内存库验证逻辑正确性。
func TestOrderLifecycle(t *testing.T) {
	setupLifecycleDB(t)
	db := app.Must().DB

	// === Seed ===
	user := model.User{Username: "buyer", Status: 1}
	db.Create(&user)

	parentUser := model.User{Username: "parent", Status: 1}
	db.Create(&parentUser)

	// 分销商关系：parentUser 是 user 的上级
	parentDist := model.Distributor{UserID: parentUser.ID, Status: 1, Level: 1}
	db.Create(&parentDist)
	buyerDist := model.Distributor{UserID: user.ID, ParentID: parentUser.ID, Status: 1, Level: 1}
	db.Create(&buyerDist)

	// 配置佣金比例
	db.Create(&model.Config{Key: "distribution_rate_level1", Value: "10", Group: "distribution"})

	goods := model.Goods{Title: "测试商品", Status: 1, CategoryID: 1}
	db.Create(&goods)

	sku := model.GoodsSKU{GoodsID: goods.ID, Name: "默认规格", Price: 10000, Stock: 5, Status: 1}
	db.Create(&sku)

	addr := model.Address{UserID: user.ID, Name: "张三", Phone: "13800000000", Province: "北京", City: "北京", District: "朝阳", Detail: "测试路1号"}
	db.Create(&addr)

	cart := model.Cart{UserID: user.ID, GoodsID: goods.ID, SKUID: sku.ID, Quantity: 1}
	db.Create(&cart)

	// === 1. 下单 ===
	addrID := addr.ID
	order, err := CreateOrder(user.ID, &CreateOrderReq{
		AddressID:  &addrID,
		CartIDs:    []uint{cart.ID},
		OrderModel: model.OrderModelExpress,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if order.Status != model.OrderStatusPending {
		t.Fatalf("order status = %d, want Pending(0)", order.Status)
	}
	if order.PayAmount != 10000 {
		t.Fatalf("pay_amount = %d, want 10000", order.PayAmount)
	}

	// 验证库存扣减
	var skuAfter model.GoodsSKU
	db.First(&skuAfter, sku.ID)
	if skuAfter.Stock != 4 {
		t.Fatalf("stock = %d, want 4", skuAfter.Stock)
	}

	// === 2. 支付回调 ===
	if err := HandlePayNotify(order.OrderNo, "TX_123456"); err != nil {
		t.Fatalf("HandlePayNotify: %v", err)
	}

	var paidOrder model.Order
	db.First(&paidOrder, order.ID)
	if paidOrder.Status != model.OrderStatusPaid {
		t.Fatalf("after pay: status = %d, want Paid(1)", paidOrder.Status)
	}
	if paidOrder.TransactionID != "TX_123456" {
		t.Fatalf("transaction_id = %q, want TX_123456", paidOrder.TransactionID)
	}

	// 幂等：重复回调不报错
	if err := HandlePayNotify(order.OrderNo, "TX_123456"); err != nil {
		t.Fatalf("HandlePayNotify idempotent: %v", err)
	}

	// === 3. 发货 ===
	if err := ShipOrder(&ShipOrderReq{
		OrderID:        order.ID,
		ExpressCompany: "顺丰",
		ExpressNo:      "SF123456",
	}); err != nil {
		t.Fatalf("ShipOrder: %v", err)
	}

	var shippedOrder model.Order
	db.First(&shippedOrder, order.ID)
	if shippedOrder.Status != model.OrderStatusShipped {
		t.Fatalf("after ship: status = %d, want Shipped(2)", shippedOrder.Status)
	}

	// === 4. 确认收货 + 佣金结算 ===
	// 注意：ConfirmReceive 内部用 SafeGo 异步调 SettleCommission，
	// SQLite 内存库不支持跨 goroutine，所以这里手动同步调用。
	if err := ConfirmReceive(user.ID, order.ID); err != nil {
		t.Fatalf("ConfirmReceive: %v", err)
	}

	// 等 SafeGo goroutine 里的 SettleCommission 完成（SQLite 不支持并发写）
	time.Sleep(200 * time.Millisecond)

	var completedOrder model.Order
	db.First(&completedOrder, order.ID)
	if completedOrder.Status != model.OrderStatusCompleted {
		t.Fatalf("after receive: status = %d, want Completed(3)", completedOrder.Status)
	}

	// 同步调用佣金结算（SafeGo 在 SQLite 内存库下可能跨 goroutine 失败）
	SettleCommission(order.ID)

	var commLog model.CommissionLog
	db.Where("order_id = ? AND type = ?", order.ID, "order").First(&commLog)
	if commLog.ID == 0 {
		t.Fatalf("commission log not found")
	}
	expectedCommission := int64(10000 * 10 / 100) // 10%
	if commLog.Amount != expectedCommission {
		t.Fatalf("commission = %d, want %d", commLog.Amount, expectedCommission)
	}

	// 验证分销商余额增加
	var dist model.Distributor
	db.First(&dist, parentDist.ID)
	if dist.Balance != expectedCommission {
		t.Fatalf("distributor balance = %d, want %d", dist.Balance, expectedCommission)
	}

	// 幂等：再次结算不重复发放
	SettleCommission(order.ID)
	var logCount int64
	db.Model(&model.CommissionLog{}).Where("order_id = ? AND type = ?", order.ID, "order").Count(&logCount)
	if logCount != 1 {
		t.Fatalf("commission log count = %d, want 1 (idempotent)", logCount)
	}
}

// TestOrderRefund 退款链路测试：下单→支付→发货→退款→库存回滚→幂等。
func TestOrderRefund(t *testing.T) {
	setupLifecycleDB(t)
	db := app.Must().DB

	// 开启沙盒模式，退款不走真实第三方
	app.Must().Cfg.Payment.Sandbox = true

	user := model.User{Username: "buyer2", Status: 1}
	db.Create(&user)

	goods := model.Goods{Title: "退款商品", Status: 1, CategoryID: 1}
	db.Create(&goods)

	sku := model.GoodsSKU{GoodsID: goods.ID, Name: "默认", Price: 5000, Stock: 10, Status: 1}
	db.Create(&sku)

	addr := model.Address{UserID: user.ID, Name: "李四", Phone: "13900000000", Province: "上海", City: "上海", District: "浦东", Detail: "测试路2号"}
	db.Create(&addr)

	cart := model.Cart{UserID: user.ID, GoodsID: goods.ID, SKUID: sku.ID, Quantity: 2}
	db.Create(&cart)

	// 下单
	addrID := addr.ID
	order, err := CreateOrder(user.ID, &CreateOrderReq{
		AddressID:  &addrID,
		CartIDs:    []uint{cart.ID},
		OrderModel: model.OrderModelExpress,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	// 支付
	HandlePayNotify(order.OrderNo, "TX_REFUND_1")

	// 发货
	ShipOrder(&ShipOrderReq{OrderID: order.ID, ExpressCompany: "圆通", ExpressNo: "YT999"})

	// 退款（Shipped 状态允许退款）
	if err := RefundOrder(user.ID, &RefundReq{OrderID: order.ID, Reason: "不想要了"}); err != nil {
		t.Fatalf("RefundOrder: %v", err)
	}

	var refundedOrder model.Order
	db.First(&refundedOrder, order.ID)
	if refundedOrder.Status != model.OrderStatusRefunded {
		t.Fatalf("after refund: status = %d, want Refunded(5)", refundedOrder.Status)
	}

	// 验证库存回滚
	var skuFinal model.GoodsSKU
	db.First(&skuFinal, sku.ID)
	if skuFinal.Stock != 10 {
		t.Fatalf("stock after refund = %d, want 10 (restored)", skuFinal.Stock)
	}

	// 验证退款幂等
	err = RefundOrder(user.ID, &RefundReq{OrderID: order.ID, Reason: "再退一次"})
	if err == nil {
		t.Fatalf("duplicate refund should fail")
	}
}
