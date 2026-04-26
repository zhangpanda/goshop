package service

import (
	"errors"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

type SeckillReq struct {
	Name      string             `json:"name" binding:"required"`
	StartTime time.Time          `json:"start_time" binding:"required"`
	EndTime   time.Time          `json:"end_time" binding:"required"`
	Items     []PromotionItemReq `json:"items" binding:"required,min=1"`
}

func CreateSeckill(req *SeckillReq) (*model.Promotion, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, errors.New("结束时间必须晚于开始时间")
	}
	promo := model.Promotion{
		Name: req.Name, Type: "seckill",
		StartTime: req.StartTime, EndTime: req.EndTime, Status: 1,
	}
	tx := global.DB.Begin()
	if err := tx.Create(&promo).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, item := range req.Items {
		pi := model.PromotionItem{
			PromotionID: promo.ID, GoodsID: item.GoodsID, SKUID: item.SKUID,
			PromoPrice: item.PromoPrice, PromoStock: item.PromoStock, PerLimit: item.PerLimit,
		}
		if err := tx.Create(&pi).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()
	global.DB.Preload("Items").First(&promo, promo.ID)
	return &promo, nil
}

func GetSeckillList(page, pageSize int) (int64, []model.Promotion, error) {
	var total int64
	global.DB.Model(&model.Promotion{}).Where("type = ?", "seckill").Count(&total)
	var list []model.Promotion
	err := global.DB.Where("type = ?", "seckill").
		Preload("Items").Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func GetActiveSeckills() ([]model.Promotion, error) {
	now := time.Now()
	var list []model.Promotion
	err := global.DB.Where("type = ? AND status = 1 AND start_time <= ? AND end_time > ?", "seckill", now, now).
		Preload("Items").Find(&list).Error
	return list, err
}

// SeckillBuy 秒杀下单校验（库存+限购）
func SeckillBuy(userID, itemID uint) error {
	var item model.PromotionItem
	if err := global.DB.First(&item, itemID).Error; err != nil {
		return errors.New("秒杀商品不存在")
	}
	var promo model.Promotion
	if err := global.DB.First(&promo, item.PromotionID).Error; err != nil || !promo.IsActive() || promo.Type != "seckill" {
		return errors.New("秒杀活动不在进行中")
	}
	if item.Sold >= item.PromoStock {
		return errors.New("已售罄")
	}
	if item.PerLimit > 0 {
		var bought int64
		global.DB.Model(&model.OrderItem{}).
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.user_id = ? AND order_items.sku_id = ? AND orders.status != ?",
				userID, item.SKUID, model.OrderStatusCancelled).
			Count(&bought)
		if int(bought) >= item.PerLimit {
			return errors.New("已达限购数量")
		}
	}
	// 扣减秒杀库存（乐观锁）
	result := global.DB.Model(&model.PromotionItem{}).
		Where("id = ? AND sold < promo_stock", item.ID).
		Update("sold", gorm.Expr("sold + 1"))
	if result.RowsAffected == 0 {
		return errors.New("已售罄")
	}
	return nil
}
