package service

import (
	"testing"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/testutil"
)

func TestCreateOrder_Integration(t *testing.T) {
	testutil.SetupTestDB()

	// Seed: goods + SKU + cart + address
	goods := model.Goods{Title: "测试商品", Status: 1, CategoryID: 1}
	global.DB.Create(&goods)

	sku := model.GoodsSKU{GoodsID: goods.ID, Name: "默认", Price: 9900, Stock: 10, Status: 1}
	global.DB.Create(&sku)

	addr := model.Address{UserID: 1, Name: "张三", Phone: "13800000000", Province: "北京", City: "北京", District: "朝阳", Detail: "xx路"}
	global.DB.Create(&addr)

	cart := model.Cart{UserID: 1, GoodsID: goods.ID, SKUID: sku.ID, Quantity: 2}
	global.DB.Create(&cart)

	// Create order
	addrID := addr.ID
	req := &CreateOrderReq{
		AddressID:  &addrID,
		CartIDs:    []uint{cart.ID},
		OrderModel: model.OrderModelExpress,
	}
	order, err := CreateOrder(1, req)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order.TotalAmount != 19800 {
		t.Errorf("TotalAmount = %d, want 19800", order.TotalAmount)
	}
	if order.Status != model.OrderStatusPending {
		t.Errorf("Status = %d, want Pending(0)", order.Status)
	}
	if len(order.Items) != 1 {
		t.Fatalf("Items count = %d, want 1", len(order.Items))
	}

	// Verify stock deducted
	var updatedSKU model.GoodsSKU
	global.DB.First(&updatedSKU, sku.ID)
	if updatedSKU.Stock != 8 {
		t.Errorf("Stock = %d, want 8", updatedSKU.Stock)
	}

	// Verify cart cleared
	var cartCount int64
	global.DB.Model(&model.Cart{}).Where("user_id = 1").Count(&cartCount)
	if cartCount != 0 {
		t.Errorf("Cart count = %d, want 0", cartCount)
	}
}

func TestCreateOrder_EmptyCart(t *testing.T) {
	testutil.SetupTestDB()

	addrID := uint(1)
	req := &CreateOrderReq{
		AddressID:  &addrID,
		CartIDs:    []uint{999},
		OrderModel: model.OrderModelExpress,
	}
	_, err := CreateOrder(1, req)
	if err == nil {
		t.Fatal("expected error for empty cart")
	}
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	testutil.SetupTestDB()

	goods := model.Goods{Title: "库存不足商品", Status: 1, CategoryID: 1}
	global.DB.Create(&goods)

	sku := model.GoodsSKU{GoodsID: goods.ID, Name: "默认", Price: 100, Stock: 1, Status: 1}
	global.DB.Create(&sku)

	addr := model.Address{UserID: 1, Name: "张三", Phone: "13800000000", Province: "北京", City: "北京", District: "朝阳", Detail: "xx路"}
	global.DB.Create(&addr)

	cart := model.Cart{UserID: 1, GoodsID: goods.ID, SKUID: sku.ID, Quantity: 5}
	global.DB.Create(&cart)

	addrID := addr.ID
	req := &CreateOrderReq{
		AddressID:  &addrID,
		CartIDs:    []uint{cart.ID},
		OrderModel: model.OrderModelExpress,
	}
	_, err := CreateOrder(1, req)
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}

func TestCancelOrder_Integration(t *testing.T) {
	testutil.SetupTestDB()

	goods := model.Goods{Title: "取消测试", Status: 1, CategoryID: 1}
	global.DB.Create(&goods)

	sku := model.GoodsSKU{GoodsID: goods.ID, Name: "默认", Price: 5000, Stock: 10, Status: 1}
	global.DB.Create(&sku)

	addr := model.Address{UserID: 1, Name: "张三", Phone: "13800000000", Province: "北京", City: "北京", District: "朝阳", Detail: "xx路"}
	global.DB.Create(&addr)

	cart := model.Cart{UserID: 1, GoodsID: goods.ID, SKUID: sku.ID, Quantity: 3}
	global.DB.Create(&cart)

	addrID := addr.ID
	order, _ := CreateOrder(1, &CreateOrderReq{
		AddressID: &addrID, CartIDs: []uint{cart.ID}, OrderModel: model.OrderModelExpress,
	})

	// Cancel
	err := CancelOrder(1, order.ID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	// Verify status
	var o model.Order
	global.DB.First(&o, order.ID)
	if o.Status != model.OrderStatusCancelled {
		t.Errorf("Status = %d, want Cancelled(4)", o.Status)
	}

	// Verify stock restored
	var updatedSKU model.GoodsSKU
	global.DB.First(&updatedSKU, sku.ID)
	if updatedSKU.Stock != 10 {
		t.Errorf("Stock = %d, want 10 (restored)", updatedSKU.Stock)
	}
}
