package service

import (
	"log/slog"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// postPaidHook 统一的"支付成功后"副作用入口，所有支付路径（微信/支付宝/钱包/线下/合并/沙盒）
// 在订单已事务化落库为 Paid 后调用此函数。
//
// 注意：
//  1. 调用方必须先确保订单状态已从 Pending 原子变更为 Paid（例如 UPDATE 返回 RowsAffected=1），
//     postPaidHook 本身不做 status 校验，避免重复触发。
//  2. 当前副作用：
//     - 订单状态历史（AddOrderStatusHistory）
//     - 用户通知（站内信 + 微信模板消息）
//  3. 所有子副作用内部自行处理错误；本函数只记录 warning，不向上传播。
func postPaidHook(orderID uint, note, creator string) {
	var order model.Order
	if err := app.Must().DB.First(&order, orderID).Error; err != nil {
		slog.Warn("postPaidHook load order", "order_id", orderID, "err", err)
		return
	}
	AddOrderStatusHistory(orderID, model.OrderStatusPending, model.OrderStatusPaid, note, creator)
	NotifyOrderStatus(order.UserID, orderID, order.OrderNo, "paid")
}
