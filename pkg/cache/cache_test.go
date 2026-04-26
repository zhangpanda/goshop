package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_SetGet(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	if err := c.Set(ctx, "k1", "v1", time.Minute); err != nil {
		t.Fatal(err)
	}
	val, err := c.Get(ctx, "k1")
	if err != nil || val != "v1" {
		t.Fatalf("Get = %q, %v; want v1, nil", val, err)
	}
}

func TestMemoryCache_GetMiss(t *testing.T) {
	c := NewMemoryCache()
	_, err := c.Get(context.Background(), "nonexistent")
	if err != ErrNil {
		t.Fatalf("Get miss err = %v; want ErrNil", err)
	}
}

func TestMemoryCache_Expiry(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	c.Set(ctx, "exp", "val", 50*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	_, err := c.Get(ctx, "exp")
	if err != ErrNil {
		t.Fatal("expected key to be expired")
	}
}

func TestMemoryCache_Del(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	c.Set(ctx, "d1", "v", time.Minute)
	c.Set(ctx, "d2", "v", time.Minute)
	c.Del(ctx, "d1", "d2")
	if _, err := c.Get(ctx, "d1"); err != ErrNil {
		t.Fatal("d1 should be deleted")
	}
}

func TestMemoryCache_FlushDB(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	c.Set(ctx, "a", "1", time.Minute)
	c.Set(ctx, "b", "2", time.Minute)
	c.FlushDB(ctx)
	n, _ := c.DBSize(ctx)
	if n != 0 {
		t.Fatalf("DBSize after flush = %d; want 0", n)
	}
}

func TestMemoryCache_DBSize(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	c.Set(ctx, "x", "1", time.Minute)
	c.Set(ctx, "y", "2", time.Minute)
	n, _ := c.DBSize(ctx)
	if n != 2 {
		t.Fatalf("DBSize = %d; want 2", n)
	}
}

func TestMemoryCache_Keys(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	c.Set(ctx, "prefix:a", "1", time.Minute)
	c.Set(ctx, "prefix:b", "2", time.Minute)
	c.Set(ctx, "other:c", "3", time.Minute)
	keys, _ := c.Keys(ctx, "prefix:*")
	if len(keys) != 2 {
		t.Fatalf("Keys(prefix:*) = %d; want 2", len(keys))
	}
}

func TestMemoryCache_NoExpiry(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()
	c.Set(ctx, "perm", "val", 0)
	val, err := c.Get(ctx, "perm")
	if err != nil || val != "val" {
		t.Fatalf("permanent key: Get = %q, %v", val, err)
	}
}
