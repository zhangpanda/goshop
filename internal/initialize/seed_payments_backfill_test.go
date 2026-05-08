package initialize

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	app.Register(&app.Deps{})
	os.Exit(m.Run())
}

var sqliteMemSeq atomic.Uint64

// testSQLiteDB 使用独立内存 SQLite（避免 file::memory:?cache=shared 在进程内复用同库），仅迁移 payments 表。
func testSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := fmt.Sprintf("initpay_%d_%s", sqliteMemSeq.Add(1), strings.ReplaceAll(t.Name(), "/", "_"))
	dsn := "file:" + name + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Payment{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestEnsureDefaultPayments_FullAndIdempotent(t *testing.T) {
	prev := app.Must().DB
	t.Cleanup(func() { app.Must().DB = prev })

	app.Must().DB = testSQLiteDB(t)
	EnsureDefaultPayments()

	var n int64
	app.Must().DB.Model(&model.Payment{}).Count(&n)
	if n != 12 {
		t.Fatalf("首次补全应 12 条，得到 %d", n)
	}

	EnsureDefaultPayments()
	app.Must().DB.Model(&model.Payment{}).Count(&n)
	if n != 12 {
		t.Fatalf("幂等：仍应为 12 条，得到 %d", n)
	}
}

func TestEnsureDefaultPayments_PartialOnlyOffline(t *testing.T) {
	prev := app.Must().DB
	t.Cleanup(func() { app.Must().DB = prev })

	app.Must().DB = testSQLiteDB(t)
	app.Must().DB.Create(&model.Payment{
		Name:   "线下支付",
		Config: `{"payment_key":"offline"}`,
		Sort:   100,
		Status: 1,
	})
	EnsureDefaultPayments()

	var n int64
	app.Must().DB.Model(&model.Payment{}).Count(&n)
	if n != 12 {
		t.Fatalf("仅有 offline 时应补 11 条共 12，得到 %d", n)
	}
}

func TestEnsureDefaultPayments_OldStyleWechatName(t *testing.T) {
	prev := app.Must().DB
	t.Cleanup(func() { app.Must().DB = prev })

	app.Must().DB = testSQLiteDB(t)
	app.Must().DB.Create(&model.Payment{
		Name:   "微信支付",
		Config: "{}",
		Status: 1,
	})
	EnsureDefaultPayments()

	var rows []model.Payment
	app.Must().DB.Find(&rows)
	if len(rows) != 12 {
		t.Fatalf("名称推断 wechat_jsapi 后应补其余 11 条共 12，得到 %d", len(rows))
	}
	var jsapi int
	for i := range rows {
		k, _ := service.PaymentDriverKeyFromPayment(&rows[i])
		if k == "wechat_jsapi" {
			jsapi++
		}
	}
	if jsapi != 1 {
		t.Fatalf("wechat_jsapi 应只对应一行，得到 %d", jsapi)
	}
}
