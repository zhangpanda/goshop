package initialize

import (
	"testing"
	"time"

	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDedupeGroupOrderMembersBeforeUniqueIndex_sqlite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// 模拟升级前旧库：无唯一约束，才能插入重复 (group_order_id, user_id)
	if err := db.Exec(`
CREATE TABLE group_order_members (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  group_order_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  order_id INTEGER,
  created_at DATETIME
)`).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := db.Exec(`INSERT INTO group_order_members (group_order_id,user_id,order_id,created_at) VALUES (1,1,10,?)`, now).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO group_order_members (group_order_id,user_id,order_id,created_at) VALUES (1,1,20,?)`, now).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO group_order_members (group_order_id,user_id,order_id,created_at) VALUES (1,2,30,?)`, now).Error; err != nil {
		t.Fatal(err)
	}

	if err := DedupeGroupOrderMembersBeforeUniqueIndex(db); err != nil {
		t.Fatal(err)
	}

	var n int64
	if err := db.Model(&model.GroupOrderMember{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 rows after dedupe, got %d", n)
	}

	var one model.GroupOrderMember
	if err := db.Where("group_order_id = ? AND user_id = ?", 1, 1).First(&one).Error; err != nil {
		t.Fatal(err)
	}
	if one.OrderID != 10 {
		t.Fatalf("expected to keep smallest id row (order_id=10), got order_id=%d", one.OrderID)
	}
}
