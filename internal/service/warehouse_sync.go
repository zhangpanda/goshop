package service

import (
	"fmt"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// GoodsSpecInventorySync 仓库库存同步到商品SKU库存（所有启用仓库的该商品库存求和）
func GoodsSpecInventorySync(goodsID uint) error {
	// 检查该商品是否有仓库管理记录（无论仓库启用与否）
	var warehouseCount int64
	app.Must().DB.Model(&model.WarehouseGoodsSpec{}).Where("goods_id = ?", goodsID).Count(&warehouseCount)
	if warehouseCount == 0 {
		return nil // 该商品不使用仓库管理，跳过
	}

	// 获取所有启用仓库的规格库存
	type specSum struct {
		SKUID     uint `gorm:"column:sku_id"`
		Inventory int
	}
	var sums []specSum
	app.Must().DB.Model(&model.WarehouseGoodsSpec{}).
		Joins("JOIN warehouses ON warehouses.id = warehouse_goods_specs.warehouse_id AND warehouses.is_enable = 1").
		Where("warehouse_goods_specs.goods_id = ?", goodsID).
		Select("warehouse_goods_specs.sku_id, SUM(warehouse_goods_specs.inventory) as inventory").
		Group("warehouse_goods_specs.sku_id").Find(&sums)

	// 该商品有仓库管理但所有仓库都已禁用时，将SKU库存置零
	if len(sums) == 0 {
		app.Must().DB.Model(&model.GoodsSKU{}).Where("goods_id = ?", goodsID).Update("stock", 0)
	} else {
		// 先将所有 SKU 库存置零，再用仓库汇总数据覆盖（防止被移出仓库的 SKU 残留库存）
		app.Must().DB.Model(&model.GoodsSKU{}).Where("goods_id = ?", goodsID).Update("stock", 0)
		for _, s := range sums {
			if s.SKUID > 0 {
				app.Must().DB.Model(&model.GoodsSKU{}).Where("id = ?", s.SKUID).Update("stock", s.Inventory)
			}
		}
	}
	// 同步GoodsSpecBase
	var baseSums []struct {
		SpecValues string
		Inventory  int
	}
	app.Must().DB.Model(&model.WarehouseGoodsSpec{}).
		Joins("JOIN warehouses ON warehouses.id = warehouse_goods_specs.warehouse_id AND warehouses.is_enable = 1").
		Where("warehouse_goods_specs.goods_id = ?", goodsID).
		Select("warehouse_goods_specs.spec_values, SUM(warehouse_goods_specs.inventory) as inventory").
		Group("warehouse_goods_specs.spec_values").Find(&baseSums)
	for _, s := range baseSums {
		if s.SpecValues != "" {
			app.Must().DB.Model(&model.GoodsSpecBase{}).Where("goods_id = ? AND spec_values = ?", goodsID, s.SpecValues).Update("inventory", s.Inventory)
		}
	}
	return nil
}

// GoodsSpecChangeInventorySync 商品规格变更后同步仓库库存结构
func GoodsSpecChangeInventorySync(goodsID uint) error {
	return GoodsSpecInventorySync(goodsID)
}

// WarehouseGoodsInventoryDeduct 仓库库存扣减（下单时）
func WarehouseGoodsInventoryDeduct(orderID, goodsID uint, skuID uint, quantity int) error {
	var ws model.WarehouseGoodsSpec
	app.Must().DB.Where("goods_id = ? AND sku_id = ? AND inventory >= ?", goodsID, skuID, quantity).
		Joins("JOIN warehouses ON warehouses.id = warehouse_goods_specs.warehouse_id AND warehouses.is_enable = 1").
		Order("warehouses.level DESC").First(&ws)
	if ws.ID == 0 {
		return nil // 无仓库管理则跳过
	}
	result := app.Must().DB.Model(&model.WarehouseGoodsSpec{}).Where("id = ? AND inventory >= ?", ws.ID, quantity).
		Update("inventory", gorm.Expr("inventory - ?", quantity))
	if result.RowsAffected == 0 {
		return fmt.Errorf("仓库库存不足")
	}
	// 同步仓库商品总库存
	app.Must().DB.Model(&model.WarehouseGoods{}).Where("warehouse_id = ? AND goods_id = ?", ws.WarehouseID, goodsID).
		Update("inventory", gorm.Expr("inventory - ?", quantity))
	// 记录库存日志
	AddInventoryLog(orderID, goodsID, skuID, -quantity, "order", "订单扣库存")
	return nil
}

// WarehouseGoodsInventoryRollback 仓库库存回滚（取消/退货时）
func WarehouseGoodsInventoryRollback(orderID, goodsID uint, skuID uint, quantity int) error {
	// 找到最近扣减的仓库
	var log model.InventoryLog
	app.Must().DB.Where("order_id = ? AND goods_id = ? AND sku_id = ? AND type = 'order'", orderID, goodsID, skuID).
		Order("id DESC").First(&log)
	// 回滚到默认仓库
	var ws model.WarehouseGoodsSpec
	app.Must().DB.Where("goods_id = ? AND sku_id = ?", goodsID, skuID).First(&ws)
	if ws.ID > 0 {
		app.Must().DB.Model(&ws).Update("inventory", gorm.Expr("inventory + ?", quantity))
		app.Must().DB.Model(&model.WarehouseGoods{}).Where("warehouse_id = ? AND goods_id = ?", ws.WarehouseID, goodsID).
			Update("inventory", gorm.Expr("inventory + ?", quantity))
	}
	AddInventoryLog(orderID, goodsID, skuID, quantity, "rollback", "库存回滚")
	return nil
}
