package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/testutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TestCommissionIdemKey_MySQLUnique 验证 CommissionLog.IdemKey 作为
// `uniqueIndex:uk_commission_idem` 在真 MySQL 下的行为：
//  1. type=order 且同 (order_id, distributor_id) 只允许一条；
//  2. 其他类型 IdemKey 为 NULL 时可多条（MySQL UNIQUE 对 NULL 不比较）。
//
// SQLite 虽然也有同样语义，但我们要的是在持续部署目标（MySQL 8）下做一次兜底。
// 本测试默认 Skip；CI 以 GOSHOP_TEST_MYSQL_DSN 触发。
func TestCommissionIdemKey_MySQLUnique(t *testing.T) {
	db, teardown := testutil.SetupMySQLAppDeps(t)
	if db == nil {
		return
	}
	defer teardown()

	// 清空可能的遗留数据
	db.Where("1=1").Delete(&model.CommissionLog{})
	db.Where("1=1").Delete(&model.Distributor{})

	// 准备分销商
	d := model.Distributor{UserID: 500, Status: 1}
	if err := db.Create(&d).Error; err != nil {
		t.Fatal(err)
	}

	k := fmt.Sprintf("order:7001:%d", d.ID)
	first := model.CommissionLog{DistributorID: d.ID, OrderID: 7001, Amount: 100, Type: "order", IdemKey: &k}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("first insert: %v", err)
	}

	// 同 IdemKey 第二次插入必须被 MySQL UNIQUE 拒绝
	dup := k
	second := model.CommissionLog{DistributorID: d.ID, OrderID: 7001, Amount: 100, Type: "order", IdemKey: &dup}
	if err := db.Create(&second).Error; err == nil {
		t.Fatal("expected UNIQUE violation on duplicate IdemKey, got nil")
	} else if !isDuplicateKeyError(err) {
		t.Fatalf("expected duplicate-key error, got: %v", err)
	}

	// 非 order 类型 IdemKey=NULL 可多条并存
	for i := 0; i < 3; i++ {
		if err := db.Create(&model.CommissionLog{DistributorID: d.ID, Amount: -50, Type: "withdraw"}).Error; err != nil {
			t.Fatalf("withdraw insert %d: %v", i, err)
		}
	}

	var total int64
	db.Model(&model.CommissionLog{}).Count(&total)
	if total != 4 {
		t.Errorf("expected 4 rows (1 order + 3 withdraws), got %d", total)
	}
}

// TestSettleCommission_ConcurrentMySQL 在真 MySQL 下并发调用 SettleCommission，
// 应仅有一条 order-type CommissionLog 被写入；其他并发协程要么被 SELECT COUNT
// 拦掉，要么被 IdemKey 唯一键拦掉。这验证了 P0 修复的防并发资损能力。
func TestSettleCommission_ConcurrentMySQL(t *testing.T) {
	db, teardown := testutil.SetupMySQLAppDeps(t)
	if db == nil {
		return
	}
	defer teardown()

	// 清空
	db.Where("1=1").Delete(&model.CommissionLog{})
	db.Where("1=1").Delete(&model.Distributor{})
	db.Where("1=1").Delete(&model.Order{})

	// user=600 的上级 = user=700（分销商）
	if err := db.Create(&model.Distributor{UserID: 700, Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Distributor{UserID: 600, ParentID: 700, Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	order := model.Order{
		OrderNo: "CONC-1", UserID: 600, PayAmount: 50000, Status: model.OrderStatusCompleted,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	// 10 并发同时结算同一订单
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			SettleCommission(order.ID)
		}()
	}
	wg.Wait()

	var orderLogCount int64
	db.Model(&model.CommissionLog{}).Where("order_id = ? AND type = ?", order.ID, "order").Count(&orderLogCount)
	if orderLogCount != 1 {
		t.Errorf("expected exactly 1 commission log for order, got %d (idempotency broken)", orderLogCount)
	}
}

// TestRefundRowLock_FORUPDATE 验证 clause.Locking FOR UPDATE 在 MySQL 下真的
// 阻塞另一事务读取同一行，从而避免 AuditWithdraw 的重复退款。
// SQLite 不支持行级锁，不能验证这条。
func TestRefundRowLock_FORUPDATE(t *testing.T) {
	db, teardown := testutil.SetupMySQLAppDeps(t)
	if db == nil {
		return
	}
	defer teardown()

	db.Where("1=1").Delete(&model.WithdrawRequest{})
	db.Where("1=1").Delete(&model.Distributor{})
	if err := db.Create(&model.Distributor{UserID: 800, Status: 1, Balance: 0}).Error; err != nil {
		t.Fatal(err)
	}
	w := model.WithdrawRequest{DistributorID: 1, UserID: 800, Amount: 1000, Status: 0}
	if err := db.Create(&w).Error; err != nil {
		t.Fatal(err)
	}

	// 并发两次 AuditWithdraw(false) — 只有一个能成功退回余额 + 改 status=2，
	// 另一个应返回 "已处理"。
	var (
		wg       sync.WaitGroup
		errs     [2]error
		okCount  int
		okCountM sync.Mutex
	)
	for i := 0; i < 2; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = AuditWithdraw(w.ID, false, "并发测试")
			if errs[i] == nil {
				okCountM.Lock()
				okCount++
				okCountM.Unlock()
			}
		}()
	}
	wg.Wait()

	if okCount != 1 {
		t.Errorf("expected exactly 1 successful audit, got %d (errs=%v)", okCount, errs)
	}

	// 余额只退回一次
	var d model.Distributor
	db.Where("user_id = ?", 800).First(&d)
	if d.Balance != 1000 {
		t.Errorf("distributor balance = %d, want 1000 (refund happened exactly once)", d.Balance)
	}

	// status 最终 = 2（已拒绝）
	var got model.WithdrawRequest
	db.First(&got, w.ID)
	if got.Status != 2 {
		t.Errorf("withdraw status = %d, want 2", got.Status)
	}

	// 通过 clause.Locking 显式一次行锁读，确认与 AuditWithdraw 内部一致
	if err := db.Transaction(func(tx *gorm.DB) error {
		var locked model.WithdrawRequest
		return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, w.ID).Error
	}); err != nil {
		t.Errorf("FOR UPDATE read failed: %v", err)
	}
}
