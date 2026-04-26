package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// OrderAftersaleCalculation 自动计算退款金额
func OrderAftersaleCalculation(orderDetailID uint, number int) int64 {
	var item model.OrderItem
	global.DB.First(&item, orderDetailID)
	if item.ID == 0 {
		return 0
	}
	if number <= 0 {
		number = item.Quantity
	}
	return item.Price * int64(number)
}

// OrderAftersaleChoiceTypeList 可选售后类型
func OrderAftersaleChoiceTypeList(orderID uint) []map[string]interface{} {
	var order model.Order
	global.DB.First(&order, orderID)
	list := []map[string]interface{}{}
	switch order.Status {
	case model.OrderStatusPaid:
		list = append(list, map[string]interface{}{"type": model.AftersaleTypeRefundOnly, "name": "仅退款"})
	case model.OrderStatusShipped, model.OrderStatusCompleted:
		list = append(list, map[string]interface{}{"type": model.AftersaleTypeRefundOnly, "name": "仅退款"})
		list = append(list, map[string]interface{}{"type": model.AftersaleTypeReturn, "name": "退货退款"})
	}
	return list
}

// OrderAftersaleReturnGoodsAddress 退货地址（优先仓库地址）
func OrderAftersaleReturnGoodsAddress(orderID uint) map[string]string {
	useWarehouse := GetConfig("order_aftersale_use_warehouse_address") == "1"
	if useWarehouse {
		var item model.OrderItem
		global.DB.Where("order_id = ?", orderID).First(&item)
		var wg model.WarehouseGoods
		global.DB.Where("goods_id = ?", item.GoodsID).First(&wg)
		if wg.ID > 0 {
			var w model.Warehouse
			global.DB.First(&w, wg.WarehouseID)
			if w.ID > 0 {
				return map[string]string{"name": w.ContactsName, "tel": w.ContactsTel, "address": w.Province + w.City + w.County + w.Address}
			}
		}
	}
	return map[string]string{
		"name": GetConfig("aftersale_contact_name"), "tel": GetConfig("aftersale_contact_tel"),
		"address": GetConfig("aftersale_address"),
	}
}

// OrderAftersaleStepData 售后进度
func OrderAftersaleStepData(asID uint) []model.AftersaleHistory {
	var list []model.AftersaleHistory
	global.DB.Where("aftersale_id = ?", asID).Order("id ASC").Find(&list)
	return list
}

// OrderAftersaleTipsMsg 售后提示信息
func OrderAftersaleTipsMsg(status int8) string {
	m := map[int8]string{
		model.AftersaleStatusPending:   "您的售后申请已提交，请等待商家处理",
		model.AftersaleStatusShipping:  "商家已确认，请尽快退货并填写物流信息",
		model.AftersaleStatusAudit:     "商家已收到退货，正在审核中",
		model.AftersaleStatusDone:      "售后已完成",
		model.AftersaleStatusRefused:   "商家已拒绝您的售后申请",
		model.AftersaleStatusCancelled: "售后已取消",
	}
	return m[status]
}

// OrderAftersaleTotal 售后总数
func OrderAftersaleTotal(userID uint, status *int8) int64 {
	db := global.DB.Model(&model.OrderAftersale{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var c int64
	db.Count(&c)
	return c
}

// OrderIsCanLaunchAftersale 判断订单是否可发起售后
func OrderIsCanLaunchAftersale(orderID, userID uint) bool {
	var order model.Order
	global.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order)
	if order.ID == 0 {
		return false
	}
	return order.Status == model.OrderStatusPaid || order.Status == model.OrderStatusShipped || order.Status == model.OrderStatusCompleted
}
