package service

import (
	"testing"
	"time"

	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// withFreshDB 在每个测试中替换全局 Deps 的 DB 为一个独立的 in-memory SQLite，
// 用完自动恢复，避免用例之间状态污染。
func withFreshDB(t *testing.T, migrate []any) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(migrate...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	orig := app.Must().DB
	app.Must().DB = db
	// 兜底恢复 Cfg（Settle 会读取 distribution_rate_level1/2，空配置走默认值）
	if app.Must().Cfg == nil {
		app.Must().Cfg = &config.Config{}
	}
	t.Cleanup(func() { app.Must().DB = orig })
	return db
}

// TestSettleCommission_Idempotent 保证同一订单多次调用 SettleCommission
// 只产生一次佣金（代码级 SELECT COUNT + 物理 UNIQUE(idem_key) 双重保障）。
func TestSettleCommission_Idempotent(t *testing.T) {
	db := withFreshDB(t, []any{
		&model.Order{}, &model.Distributor{}, &model.CommissionLog{},
	})
	// 购买者(user=100) 的上级是 user=200（分销商）
	if err := db.Create(&model.Distributor{UserID: 200, ParentID: 0, Level: 1, Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Distributor{UserID: 100, ParentID: 200, Level: 1, Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	order := model.Order{
		OrderNo: "T1", UserID: 100, PayAmount: 10000, Status: model.OrderStatusCompleted,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	// 调用三次——应只有 1 条 type=order 的 CommissionLog（一级分销，二级父 user 不存在）
	for i := 0; i < 3; i++ {
		SettleCommission(order.ID)
	}

	var logs []model.CommissionLog
	if err := db.Where("order_id = ? AND type = ?", order.ID, "order").Find(&logs).Error; err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 commission log, got %d", len(logs))
	}
	if logs[0].IdemKey == nil || *logs[0].IdemKey == "" {
		t.Errorf("IdemKey should be set for type=order; got %v", logs[0].IdemKey)
	}

	// 一级佣金比例默认 10%，10000 分 → 1000 分
	var d model.Distributor
	db.Where("user_id = ?", 200).First(&d)
	if d.Balance != 1000 {
		t.Errorf("expected distributor balance 1000, got %d", d.Balance)
	}
}

// TestCronRefundReconcile_Finalizes 模拟 "第三方已退款、本地事务未完成" 场景：
// 存在 RefundLog(status=0) 且订单仍为 Paid。cron 应推进订单到 Refunded 并置 RefundLog=1，
// 同时回库存。
func TestCronRefundReconcile_Finalizes(t *testing.T) {
	db := withFreshDB(t, []any{
		&model.Order{}, &model.OrderItem{}, &model.GoodsSKU{}, &model.RefundLog{},
	})
	sku := model.GoodsSKU{Stock: 5, Price: 100}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatal(err)
	}
	order := model.Order{OrderNo: "T2", UserID: 1, PayAmount: 200, Status: model.OrderStatusPaid, TransactionID: "tx_ok"}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.OrderItem{OrderID: order.ID, SKUID: sku.ID, Quantity: 2}).Error; err != nil {
		t.Fatal(err)
	}
	// stuck refund log: 创建时间 >5min（让 cron 识别）
	stuck := model.RefundLog{
		OrderID: order.ID, UserID: 1,
		RefundNo: "R-stuck", RefundPrice: 200, Status: 0,
	}
	if err := db.Create(&stuck).Error; err != nil {
		t.Fatal(err)
	}
	// 把 created_at 往回拨 10 分钟（cron 阈值 5 分钟）
	db.Exec(`UPDATE refund_logs SET created_at = ? WHERE id = ?`, time.Now().Add(-10*time.Minute), stuck.ID)

	sucs, fail := CronRefundReconcile(app.Must())
	if sucs != 1 || fail != 0 {
		t.Fatalf("expected sucs=1 fail=0, got sucs=%d fail=%d", sucs, fail)
	}

	var got model.Order
	db.First(&got, order.ID)
	if got.Status != model.OrderStatusRefunded {
		t.Errorf("order status = %d, want Refunded(%d)", got.Status, model.OrderStatusRefunded)
	}
	var gotSKU model.GoodsSKU
	db.First(&gotSKU, sku.ID)
	if gotSKU.Stock != 7 {
		t.Errorf("stock after refund = %d, want 7", gotSKU.Stock)
	}
	var gotRL model.RefundLog
	db.First(&gotRL, stuck.ID)
	if gotRL.Status != 1 {
		t.Errorf("refund log status = %d, want 1", gotRL.Status)
	}
}

// TestCronRefundReconcile_AlreadyRefunded 订单已被其他路径推进到 Refunded，cron 仅补刷 RefundLog 状态。
func TestCronRefundReconcile_AlreadyRefunded(t *testing.T) {
	db := withFreshDB(t, []any{
		&model.Order{}, &model.OrderItem{}, &model.GoodsSKU{}, &model.RefundLog{},
	})
	order := model.Order{OrderNo: "T3", UserID: 1, PayAmount: 100, Status: model.OrderStatusRefunded, TransactionID: "tx_done"}
	db.Create(&order)
	rl := model.RefundLog{OrderID: order.ID, UserID: 1, RefundNo: "R-dangling", RefundPrice: 100, Status: 0}
	db.Create(&rl)
	db.Exec(`UPDATE refund_logs SET created_at = ? WHERE id = ?`, time.Now().Add(-10*time.Minute), rl.ID)

	sucs, fail := CronRefundReconcile(app.Must())
	if sucs != 1 || fail != 0 {
		t.Fatalf("expected sucs=1 fail=0, got sucs=%d fail=%d", sucs, fail)
	}
	var gotRL model.RefundLog
	db.First(&gotRL, rl.ID)
	if gotRL.Status != 1 {
		t.Errorf("refund log status = %d, want 1", gotRL.Status)
	}
	if gotRL.TradeNo != "tx_done" {
		t.Errorf("trade_no = %q, want %q", gotRL.TradeNo, "tx_done")
	}
}
