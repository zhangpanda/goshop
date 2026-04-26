package service

import (
	"context"
	"testing"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/pkg/cache"
)

func setupTestCache() {
	global.Cache = cache.NewMemoryCache()
}

func TestBuyDataStorage_ReadDelete(t *testing.T) {
	setupTestCache()
	data := map[string]interface{}{"goods_id": float64(1), "quantity": float64(2)}
	if err := BuyDataStorage(100, data); err != nil {
		t.Fatal(err)
	}
	got, err := BuyDataRead(100)
	if err != nil {
		t.Fatal(err)
	}
	if got["goods_id"] != float64(1) {
		t.Errorf("goods_id = %v; want 1", got["goods_id"])
	}
	BuyDataDelete(100)
	got, err = BuyDataRead(100)
	if err == nil && got != nil {
		t.Error("expected nil after delete")
	}
}

func TestBuyDataRead_Miss(t *testing.T) {
	setupTestCache()
	_, err := BuyDataRead(999)
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestBuyGoodsCheck_NoGoods(t *testing.T) {
	// Without DB, this should return an error
	// Skip if global.DB is nil
	if global.DB == nil {
		t.Skip("requires database")
	}
}

func TestClearCache(t *testing.T) {
	setupTestCache()
	global.Cache.Set(context.Background(), "test:a", "1", time.Minute)
	global.Cache.Set(context.Background(), "test:b", "2", time.Minute)
	global.Cache.Set(context.Background(), "other:c", "3", time.Minute)
	if err := ClearCache("test:"); err != nil {
		t.Fatal(err)
	}
	// test:* should be gone
	if _, err := global.Cache.Get(context.Background(), "test:a"); err == nil {
		t.Error("test:a should be cleared")
	}
	// other:c should remain
	if v, err := global.Cache.Get(context.Background(), "other:c"); err != nil || v != "3" {
		t.Error("other:c should remain")
	}
}

func TestClearCache_All(t *testing.T) {
	setupTestCache()
	global.Cache.Set(context.Background(), "x", "1", time.Minute)
	if err := ClearCache("all"); err != nil {
		t.Fatal(err)
	}
	n, _ := global.Cache.DBSize(context.Background())
	if n != 0 {
		t.Fatalf("DBSize after flush all = %d; want 0", n)
	}
}

func TestGetCacheStats(t *testing.T) {
	setupTestCache()
	stats := GetCacheStats()
	if stats["db_size"] == nil {
		t.Error("stats should have db_size")
	}
}
