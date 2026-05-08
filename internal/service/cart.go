package service

import (
	"errors"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

type AddCartReq struct {
	GoodsID  uint `json:"goods_id" form:"goods_id" binding:"required"`
	SKUID    uint `json:"sku_id" form:"sku_id" binding:"required"`
	Quantity int  `json:"quantity" form:"quantity" binding:"required,min=1"`
}

type UpdateCartReq struct {
	Quantity *int  `json:"quantity" form:"quantity" binding:"omitempty,min=1"`
	Selected *bool `json:"selected" form:"selected"`
}

func AddCart(userID uint, req *AddCartReq) (*model.Cart, error) {
	// 校验 SKU 是否存在且有库存
	var sku model.GoodsSKU
	if err := app.Must().DB.First(&sku, req.SKUID).Error; err != nil {
		return nil, errors.New("SKU不存在")
	}
	if sku.Stock < req.Quantity {
		return nil, errors.New("库存不足")
	}

	// 已存在则累加数量
	var cart model.Cart
	app.Must().DB.Where("user_id = ? AND sku_id = ?", userID, req.SKUID).Find(&cart)
	if cart.ID > 0 {
		cart.Quantity += req.Quantity
		app.Must().DB.Save(&cart)
	} else {
		cart = model.Cart{
			UserID:   userID,
			GoodsID:  req.GoodsID,
			SKUID:    req.SKUID,
			Quantity: req.Quantity,
			Selected: true,
		}
		app.Must().DB.Create(&cart)
	}
	return &cart, nil
}

func GetCartList(userID uint) ([]model.Cart, error) {
	var list []model.Cart
	err := app.Must().DB.Where("user_id = ?", userID).
		Preload("Goods").Preload("SKU").
		Find(&list).Error
	return list, err
}

func UpdateCart(userID, cartID uint, req *UpdateCartReq) error {
	var cart model.Cart
	if err := app.Must().DB.Where("id = ? AND user_id = ?", cartID, userID).First(&cart).Error; err != nil {
		return errors.New("购物车记录不存在")
	}
	if req.Quantity != nil {
		cart.Quantity = *req.Quantity
	}
	if req.Selected != nil {
		cart.Selected = *req.Selected
	}
	return app.Must().DB.Save(&cart).Error
}

func DeleteCart(userID uint, ids []uint) error {
	return app.Must().DB.Where("user_id = ? AND id IN ?", userID, ids).Delete(&model.Cart{}).Error
}

func SelectAllCart(userID uint, selected bool) error {
	return app.Must().DB.Model(&model.Cart{}).Where("user_id = ?", userID).Update("selected", selected).Error
}
