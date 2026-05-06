package service

import (
	"errors"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

type CreatePromotionReq struct {
	Name      string             `json:"name" binding:"required"`
	StartTime time.Time          `json:"start_time" binding:"required"`
	EndTime   time.Time          `json:"end_time" binding:"required"`
	Items     []PromotionItemReq `json:"items" binding:"required,min=1"`
}

type PromotionItemReq struct {
	GoodsID    uint  `json:"goods_id" binding:"required"`
	SKUID      uint  `json:"sku_id" binding:"required"`
	PromoPrice int64 `json:"promo_price" binding:"required,min=1"`
	PromoStock int   `json:"promo_stock" binding:"required,min=1"`
	PerLimit   int   `json:"per_limit"`
}

func CreatePromotion(req *CreatePromotionReq) (*model.Promotion, error) {
	promo := model.Promotion{
		Name:      req.Name,
		Type:      "promo",
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    1,
	}

	tx := global.DB.Begin()
	if err := tx.Create(&promo).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, item := range req.Items {
		pi := model.PromotionItem{
			PromotionID: promo.ID,
			GoodsID:     item.GoodsID,
			SKUID:       item.SKUID,
			PromoPrice:  item.PromoPrice,
			PromoStock:  item.PromoStock,
		}
		if err := tx.Create(&pi).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	global.DB.Preload("Items").First(&promo, promo.ID)
	return &promo, nil
}

func GetActivePromotions() ([]model.Promotion, error) {
	var list []model.Promotion
	now := time.Now()
	err := global.DB.Where("type = ? AND status = 1 AND start_time <= ? AND end_time > ?", "promo", now, now).
		Preload("Items").Find(&list).Error
	return list, err
}

/**
 * GetPromotionAdminList 管理后台：仅普通促销 type=promo，分页。
 */
func GetPromotionAdminList(page, pageSize int, keyword string) (int64, []model.Promotion, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := global.DB.Model(&model.Promotion{}).Where("type = ?", "promo")
	if kw := strings.TrimSpace(keyword); kw != "" {
		q = q.Where("name LIKE ?", "%"+kw+"%")
	}
	var total int64
	q.Count(&total)
	var list []model.Promotion
	listQ := global.DB.Where("type = ?", "promo")
	if kw := strings.TrimSpace(keyword); kw != "" {
		listQ = listQ.Where("name LIKE ?", "%"+kw+"%")
	}
	err := listQ.Preload("Items").Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

// GetPromoPrice 获取 SKU 的促销价，无促销返回 0
func GetPromoPrice(skuID uint) (int64, error) {
	var item model.PromotionItem
	now := time.Now()
	global.DB.Joins("JOIN promotions ON promotions.id = promotion_items.promotion_id").
		Where("promotion_items.sku_id = ? AND promotions.status = 1 AND promotions.start_time <= ? AND promotions.end_time > ? AND promotion_items.sold < promotion_items.promo_stock",
			skuID, now, now).
		Find(&item)
	if item.ID == 0 {
		return 0, errors.New("无促销")
	}
	return item.PromoPrice, nil
}
