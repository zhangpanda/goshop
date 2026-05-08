package service

import (
	"errors"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

type CreateReviewReq struct {
	OrderItemID uint   `json:"order_item_id" form:"order_item_id" binding:"required"`
	Rating      int8   `json:"rating" form:"rating" binding:"required,min=1,max=5"`
	Content     string `json:"content"`
	Images      string `json:"images"`
}

func CreateReview(userID uint, req *CreateReviewReq) (*model.Review, error) {
	// 查订单明细
	var item model.OrderItem
	if err := app.Must().DB.First(&item, req.OrderItemID).Error; err != nil {
		return nil, errors.New("订单明细不存在")
	}
	var order model.Order
	if err := app.Must().DB.Where("id = ? AND user_id = ?", item.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped && order.Status != model.OrderStatusCompleted {
		return nil, errors.New("订单状态不允许评价")
	}
	// 检查是否已评价
	var count int64
	app.Must().DB.Model(&model.Review{}).Where("order_item_id = ?", req.OrderItemID).Count(&count)
	if count > 0 {
		return nil, errors.New("已评价")
	}

	review := model.Review{
		UserID:      userID,
		OrderID:     item.OrderID,
		OrderItemID: req.OrderItemID,
		GoodsID:     item.GoodsID,
		SKUID:       item.SKUID,
		Rating:      req.Rating,
		Content:     req.Content,
		Images:      req.Images,
	}
	if err := app.Must().DB.Create(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// CreateOrderReviewsShopXO 对应 shopxo-uniapp order/commentssave：按订单批量评价多商品。
func CreateOrderReviewsShopXO(userID, orderID uint, goodsIDs []uint, ratings []int, contents, imageJSONs []string) error {
	if len(goodsIDs) == 0 {
		return errors.New("无评价商品")
	}
	if len(ratings) != len(goodsIDs) || len(contents) != len(goodsIDs) {
		return errors.New("评价参数数量不一致")
	}
	var order model.Order
	if err := app.Must().DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped && order.Status != model.OrderStatusCompleted {
		return errors.New("订单状态不允许评价")
	}
	for i := range goodsIDs {
		var item model.OrderItem
		if err := app.Must().DB.Where("order_id = ? AND goods_id = ?", orderID, goodsIDs[i]).First(&item).Error; err != nil {
			return errors.New("订单商品不匹配")
		}
		img := ""
		if i < len(imageJSONs) {
			img = imageJSONs[i]
		}
		if _, err := CreateReview(userID, &CreateReviewReq{
			OrderItemID: item.ID,
			Rating:      int8(ratings[i]),
			Content:     contents[i],
			Images:      img,
		}); err != nil {
			return err
		}
	}
	return nil
}

func GetGoodsReviews(goodsID uint, page, pageSize int) ([]model.Review, int64, error) {
	var total int64
	app.Must().DB.Model(&model.Review{}).Where("goods_id = ?", goodsID).Count(&total)
	var list []model.Review
	err := app.Must().DB.Where("goods_id = ?", goodsID).Preload("User").
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func ReplyReview(reviewID uint, reply string) error {
	return app.Must().DB.Model(&model.Review{}).Where("id = ?", reviewID).Update("reply", reply).Error
}
