package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// 快递公司
func CreateExpress(name, code, icon string, sort int) (*model.Express, error) {
	e := model.Express{Name: name, Code: code, Icon: icon, Sort: sort, Status: 1}
	return &e, global.DB.Create(&e).Error
}
func GetExpressList() ([]model.Express, error) {
	var list []model.Express
	return list, global.DB.Where("status = 1").Order("sort DESC").Find(&list).Error
}
func DeleteExpress(id uint) error { return global.DB.Delete(&model.Express{}, id).Error }

// 库存变动日志
func AddInventoryLog(orderID, goodsID, skuID uint, quantity int, typ, remark string) {
	global.DB.Create(&model.InventoryLog{OrderID: orderID, GoodsID: goodsID, SKUID: skuID, Quantity: quantity, Type: typ, Remark: remark})
}
func GetInventoryLogList(goodsID uint, page, pageSize int) ([]model.InventoryLog, int64, error) {
	var total int64
	db := global.DB.Model(&model.InventoryLog{})
	if goodsID > 0 {
		db = db.Where("goods_id = ?", goodsID)
	}
	db.Count(&total)
	var list []model.InventoryLog
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// 订单软删除
func DeleteOrder(userID, orderID uint) error {
	return global.DB.Where("id = ? AND user_id = ? AND status IN ?", orderID, userID,
		[]int8{model.OrderStatusCompleted, model.OrderStatusCancelled}).
		Delete(&model.Order{}).Error
}
