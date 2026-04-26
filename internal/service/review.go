package service

import (
	"errors"

	"github.com/zhangpanda/goshop/global"
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
	if err := global.DB.First(&item, req.OrderItemID).Error; err != nil {
		return nil, errors.New("订单明细不存在")
	}
	var order model.Order
	if err := global.DB.Where("id = ? AND user_id = ?", item.OrderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != model.OrderStatusCompleted {
		return nil, errors.New("订单未完成，不能评价")
	}
	// 检查是否已评价
	var count int64
	global.DB.Model(&model.Review{}).Where("order_item_id = ?", req.OrderItemID).Count(&count)
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
	if err := global.DB.Create(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func GetGoodsReviews(goodsID uint, page, pageSize int) ([]model.Review, int64, error) {
	var total int64
	global.DB.Model(&model.Review{}).Where("goods_id = ?", goodsID).Count(&total)
	var list []model.Review
	err := global.DB.Where("goods_id = ?", goodsID).Preload("User").
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func ReplyReview(reviewID uint, reply string) error {
	return global.DB.Model(&model.Review{}).Where("id = ?", reviewID).Update("reply", reply).Error
}
