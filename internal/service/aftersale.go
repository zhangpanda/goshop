package service

import (
	"errors"
	"fmt"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

type AftersaleCreateReq struct {
	OrderDetailID uint   `json:"order_detail_id" form:"order_detail_id" binding:"required"`
	Type          int8   `json:"type" form:"type" binding:"oneof=0 1"`
	Reason        string `json:"reason" form:"reason" binding:"required"`
	Price         int64  `json:"price" form:"price" binding:"required,min=1"`
	Number        int    `json:"number"`
	Msg           string `json:"msg"`
	Images        string `json:"images"`
}

func AftersaleCreate(userID uint, req *AftersaleCreateReq) (*model.OrderAftersale, error) {
	var item model.OrderItem
	if err := app.Must().DB.First(&item, req.OrderDetailID).Error; err != nil {
		return nil, errors.New("订单明细不存在")
	}
	var order model.Order
	if err := app.Must().DB.Where("id = ? AND user_id = ?", item.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped && order.Status != model.OrderStatusCompleted {
		return nil, errors.New("当前订单状态不支持售后")
	}
	// 检查是否已有进行中的售后
	var count int64
	app.Must().DB.Model(&model.OrderAftersale{}).Where("order_detail_id = ? AND status IN ?", req.OrderDetailID, []int8{0, 1, 2}).Count(&count)
	if count > 0 {
		return nil, errors.New("该商品已有进行中的售后")
	}

	as := model.OrderAftersale{
		OrderID: item.OrderID, OrderDetailID: req.OrderDetailID,
		UserID: userID, GoodsID: item.GoodsID,
		Type: req.Type, Reason: req.Reason, Price: req.Price,
		Number: req.Number, Msg: req.Msg, Images: req.Images,
	}
	// 校验退款金额不超过明细实付
	maxPrice := item.Price * int64(item.Quantity)
	if as.Price > maxPrice {
		return nil, fmt.Errorf("退款金额不能超过 %d", maxPrice)
	}
	if err := app.Must().DB.Create(&as).Error; err != nil {
		return nil, err
	}
	addAftersaleHistory(as.ID, model.AftersaleStatusPending, "用户提交售后申请", "用户")
	return &as, nil
}

// AftersaleDelivery 用户填写退货物流
type AftersaleDeliveryReq struct {
	ExpressName string `json:"express_name" form:"express_name" binding:"required"`
	ExpressNo   string `json:"express_no" form:"express_no" binding:"required"`
}

func AftersaleDelivery(userID, asID uint, req *AftersaleDeliveryReq) error {
	var as model.OrderAftersale
	if err := app.Must().DB.Where("id = ? AND user_id = ?", asID, userID).First(&as).Error; err != nil {
		return errors.New("售后单不存在")
	}
	if as.Status != model.AftersaleStatusShipping {
		return errors.New("当前状态不需要填写物流")
	}
	app.Must().DB.Model(&as).Updates(map[string]interface{}{
		"express_name": req.ExpressName, "express_no": req.ExpressNo, "status": model.AftersaleStatusAudit,
	})
	addAftersaleHistory(asID, model.AftersaleStatusAudit, fmt.Sprintf("用户已发货 %s %s", req.ExpressName, req.ExpressNo), "用户")
	return nil
}

func AftersaleCancel(userID, asID uint) error {
	var as model.OrderAftersale
	if err := app.Must().DB.Where("id = ? AND user_id = ?", asID, userID).First(&as).Error; err != nil {
		return errors.New("售后单不存在")
	}
	if as.Status == model.AftersaleStatusDone || as.Status == model.AftersaleStatusCancelled {
		return errors.New("当前状态不可取消")
	}
	app.Must().DB.Model(&as).Update("status", model.AftersaleStatusCancelled)
	addAftersaleHistory(asID, model.AftersaleStatusCancelled, "用户取消售后", "用户")
	return nil
}

// 管理员确认（仅退款直接退，退货退款等用户发货）
func AftersaleConfirm(asID uint) error {
	var as model.OrderAftersale
	if err := app.Must().DB.First(&as, asID).Error; err != nil {
		return errors.New("售后单不存在")
	}
	if as.Status != model.AftersaleStatusPending {
		return errors.New("当前状态不可确认")
	}
	if as.Type == model.AftersaleTypeRefundOnly {
		return doRefund(&as)
	}
	// 退货退款 → 等用户发货
	app.Must().DB.Model(&as).Update("status", model.AftersaleStatusShipping)
	addAftersaleHistory(asID, model.AftersaleStatusShipping, "商家已确认，请退货", "管理员")
	return nil
}

// 管理员审核通过（收到退货后）
func AftersaleAudit(asID uint) error {
	var as model.OrderAftersale
	if err := app.Must().DB.First(&as, asID).Error; err != nil {
		return errors.New("售后单不存在")
	}
	if as.Status != model.AftersaleStatusAudit {
		return errors.New("当前状态不可审核")
	}
	return doRefund(&as)
}

func AftersaleRefuse(asID uint, reason string) error {
	var as model.OrderAftersale
	if err := app.Must().DB.First(&as, asID).Error; err != nil {
		return errors.New("售后单不存在")
	}
	if as.Status == model.AftersaleStatusDone || as.Status == model.AftersaleStatusCancelled {
		return errors.New("当前状态不可拒绝")
	}
	app.Must().DB.Model(&as).Updates(map[string]interface{}{"status": model.AftersaleStatusRefused, "refuse_reason": reason})
	addAftersaleHistory(asID, model.AftersaleStatusRefused, "商家拒绝: "+reason, "管理员")
	return nil
}

func doRefund(as *model.OrderAftersale) error {
	err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		res := tx.Model(as).Update("status", model.AftersaleStatusDone)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("售后单不存在或状态已变更")
		}
		if as.Number > 0 {
			var item model.OrderItem
			if err := tx.First(&item, as.OrderDetailID).Error; err != nil {
				return err
			}
			res := tx.Model(&model.GoodsSKU{}).Where("id = ?", item.SKUID).
				Update("stock", gorm.Expr("stock + ?", as.Number))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return errors.New("SKU不存在，无法恢复库存")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	addAftersaleHistory(as.ID, model.AftersaleStatusDone, fmt.Sprintf("退款完成 %d分", as.Price), "系统")
	return nil
}

func GetAftersaleList(userID uint, page, pageSize int) ([]model.OrderAftersale, int64, error) {
	var total int64
	app.Must().DB.Model(&model.OrderAftersale{}).Where("user_id = ?", userID).Count(&total)
	var list []model.OrderAftersale
	err := app.Must().DB.Where("user_id = ?", userID).Preload("Histories").
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func GetAftersaleDetail(userID, asID uint) (*model.OrderAftersale, error) {
	var as model.OrderAftersale
	if err := app.Must().DB.Where("id = ? AND user_id = ?", asID, userID).Preload("Histories").First(&as).Error; err != nil {
		return nil, errors.New("售后单不存在")
	}
	return &as, nil
}

func AdminGetAftersaleList(page, pageSize int, status *int8) ([]model.OrderAftersale, int64, error) {
	var total int64
	db := app.Must().DB.Model(&model.OrderAftersale{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	db.Count(&total)
	var list []model.OrderAftersale
	err := db.Preload("Histories").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func addAftersaleHistory(asID uint, status int8, msg, creator string) {
	app.Must().DB.Create(&model.AftersaleHistory{AftersaleID: asID, Status: status, Msg: msg, Creator: creator})
}
