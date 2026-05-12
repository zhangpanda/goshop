package service

import (
	"errors"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// SplitOrderByWarehouse 按仓库拆分订单：查询每个购物车商品的可用仓库，按仓库分组后分别下单。
// 任一商品查不到可用仓库时放弃拆单（返回 nil, nil），由调用方走默认单订单逻辑。
func SplitOrderByWarehouse(userID uint, req *CreateOrderReq) ([]*model.Order, error) {
	var carts []model.Cart
	if err := app.Must().DB.Where("id IN ? AND user_id = ?", req.CartIDs, userID).
		Preload("Goods").Preload("SKU").Find(&carts).Error; err != nil || len(carts) == 0 {
		return nil, errors.New("购物车为空")
	}
	type groupKey struct{ WarehouseID uint }
	groups := map[groupKey][]model.Cart{}
	for _, c := range carts {
		var ws model.WarehouseGoodsSpec
		err := app.Must().DB.Where("warehouse_goods_specs.goods_id = ? AND warehouse_goods_specs.sku_id = ? AND warehouse_goods_specs.inventory > 0", c.GoodsID, c.SKUID).
			Joins("JOIN warehouse_goods ON warehouse_goods.warehouse_id = warehouse_goods_specs.warehouse_id AND warehouse_goods.goods_id = warehouse_goods_specs.goods_id AND warehouse_goods.is_enable = 1").
			Joins("JOIN warehouses ON warehouses.id = warehouse_goods_specs.warehouse_id AND warehouses.is_enable = 1").
			Order("warehouses.level DESC").First(&ws).Error
		if err != nil || ws.WarehouseID == 0 {
			return nil, nil
		}
		key := groupKey{WarehouseID: ws.WarehouseID}
		groups[key] = append(groups[key], c)
	}
	if len(groups) <= 1 {
		return nil, nil
	}
	var orders []*model.Order
	for _, groupCarts := range groups {
		ids := make([]uint, len(groupCarts))
		for i, c := range groupCarts {
			ids[i] = c.ID
		}
		subReq := *req
		subReq.CartIDs = ids
		order, err := CreateOrder(userID, &subReq)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
