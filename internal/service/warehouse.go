package service

import (
	"errors"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type WarehouseReq struct {
	Name         string `json:"name" binding:"required"`
	Alias        string `json:"alias"`
	Level        int    `json:"level"`
	ContactsName string `json:"contacts_name"`
	ContactsTel  string `json:"contacts_tel"`
	Province     string `json:"province"`
	City         string `json:"city"`
	County       string `json:"county"`
	Address      string `json:"address"`
}

func CreateWarehouse(req *WarehouseReq) (*model.Warehouse, error) {
	w := model.Warehouse{Name: req.Name, Alias: req.Alias, Level: req.Level, IsEnable: 1,
		ContactsName: req.ContactsName, ContactsTel: req.ContactsTel,
		Province: req.Province, City: req.City, County: req.County, Address: req.Address}
	return &w, global.DB.Create(&w).Error
}

func GetWarehouseList() ([]model.Warehouse, error) {
	var list []model.Warehouse
	return list, global.DB.Order("level DESC, id ASC").Find(&list).Error
}

func UpdateWarehouse(id uint, req *WarehouseReq) error {
	return global.DB.Model(&model.Warehouse{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "level": req.Level,
		"contacts_name": req.ContactsName, "contacts_tel": req.ContactsTel,
		"province": req.Province, "city": req.City, "county": req.County, "address": req.Address,
	}).Error
}

func DeleteWarehouse(id uint) error {
	var count int64
	global.DB.Model(&model.WarehouseGoods{}).Where("warehouse_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("仓库下有商品，无法删除")
	}
	return global.DB.Delete(&model.Warehouse{}, id).Error
}

func UpdateWarehouseStatus(id uint, status int8) error {
	return global.DB.Model(&model.Warehouse{}).Where("id = ?", id).Update("is_enable", status).Error
}

// 仓库商品管理
func WarehouseGoodsAdd(warehouseID, goodsID uint, inventory int) error {
	var wg model.WarehouseGoods
	global.DB.Where("warehouse_id = ? AND goods_id = ?", warehouseID, goodsID).Find(&wg)
	if wg.ID > 0 {
		return global.DB.Model(&wg).Update("inventory", inventory).Error
	}
	return global.DB.Create(&model.WarehouseGoods{WarehouseID: warehouseID, GoodsID: goodsID, Inventory: inventory, IsEnable: 1}).Error
}

func WarehouseGoodsList(warehouseID uint) ([]model.WarehouseGoods, error) {
	var list []model.WarehouseGoods
	return list, global.DB.Where("warehouse_id = ?", warehouseID).Find(&list).Error
}

func WarehouseGoodsSpecSave(warehouseID, goodsID, skuID uint, inventory int, specValues string) error {
	var ws model.WarehouseGoodsSpec
	global.DB.Where("warehouse_id = ? AND goods_id = ? AND sku_id = ?", warehouseID, goodsID, skuID).Find(&ws)
	if ws.ID > 0 {
		return global.DB.Model(&ws).Update("inventory", inventory).Error
	}
	return global.DB.Create(&model.WarehouseGoodsSpec{WarehouseID: warehouseID, GoodsID: goodsID, SKUID: skuID, Inventory: inventory, SpecValues: specValues}).Error
}

// WarehouseGoodsInventory 查询某商品在所有仓库的总库存
func WarehouseGoodsInventory(goodsID uint) int {
	var total int64
	global.DB.Model(&model.WarehouseGoods{}).Where("goods_id = ? AND is_enable = 1", goodsID).
		Select("COALESCE(SUM(inventory),0)").Scan(&total)
	return int(total)
}
