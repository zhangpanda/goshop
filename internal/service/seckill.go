package service

import (
	"errors"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
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
	err := RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		if err := tx.Create(&promo).Error; err != nil {
			return err
		}
		for _, item := range req.Items {
			pi := model.PromotionItem{
				PromotionID: promo.ID, GoodsID: item.GoodsID, SKUID: item.SKUID,
				PromoPrice: item.PromoPrice, PromoStock: item.PromoStock, PerLimit: item.PerLimit,
			}
			if err := tx.Create(&pi).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	app.Must().DB.Preload("Items").First(&promo, promo.ID)
	return &promo, nil
}

func GetSeckillList(page, pageSize int) (int64, []model.Promotion, error) {
	var total int64
	app.Must().DB.Model(&model.Promotion{}).Where("type = ?", "seckill").Count(&total)
	var list []model.Promotion
	err := app.Must().DB.Where("type = ?", "seckill").
		Preload("Items").Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func GetActiveSeckills() ([]model.Promotion, error) {
	now := time.Now()
	var list []model.Promotion
	err := app.Must().DB.Where("type = ? AND status = 1 AND start_time <= ? AND end_time > ?", "seckill", now, now).
		Preload("Items").Find(&list).Error
	return list, err
}

// SeckillBuy 秒杀下单校验（库存+限购）
func SeckillBuy(userID, itemID uint) error {
	var item model.PromotionItem
	if err := app.Must().DB.First(&item, itemID).Error; err != nil {
		return errors.New("秒杀商品不存在")
	}
	var promo model.Promotion
	if err := app.Must().DB.First(&promo, item.PromotionID).Error; err != nil || !promo.IsActive() || promo.Type != "seckill" {
		return errors.New("秒杀活动不在进行中")
	}
	if item.Sold >= item.PromoStock {
		return errors.New("已售罄")
	}
	if item.PerLimit > 0 {
		var bought int64
		// 只计入已支付/已发货/已完成订单；未支付(可能被取消)不占名额，避免恶意占位导致其他用户买不到。
		app.Must().DB.Model(&model.OrderItem{}).
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.user_id = ? AND order_items.sku_id = ? AND orders.status IN ?",
				userID, item.SKUID,
				[]int8{model.OrderStatusPaid, model.OrderStatusShipped, model.OrderStatusCompleted}).
			Count(&bought)
		if int(bought) >= item.PerLimit {
			return errors.New("已达限购数量")
		}
	}
	// 扣减秒杀库存（乐观锁）
	result := app.Must().DB.Model(&model.PromotionItem{}).
		Where("id = ? AND sold < promo_stock", item.ID).
		Update("sold", gorm.Expr("sold + 1"))
	if result.RowsAffected == 0 {
		return errors.New("已售罄")
	}
	return nil
}
