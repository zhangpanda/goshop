package service

import (
	"errors"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

type GroupBuyReq struct {
	Name      string             `json:"name" binding:"required"`
	StartTime time.Time          `json:"start_time" binding:"required"`
	EndTime   time.Time          `json:"end_time" binding:"required"`
	GroupSize int                `json:"group_size" binding:"required,min=2"`
	GroupTime int                `json:"group_time" binding:"required,min=1"` // 分钟
	Items     []PromotionItemReq `json:"items" binding:"required,min=1"`
}

func CreateGroupBuy(req *GroupBuyReq) (*model.Promotion, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, errors.New("结束时间必须晚于开始时间")
	}
	promo := model.Promotion{
		Name: req.Name, Type: "group",
		StartTime: req.StartTime, EndTime: req.EndTime,
		GroupSize: req.GroupSize, GroupTime: req.GroupTime, Status: 1,
	}
	tx := global.DB.Begin()
	if err := tx.Create(&promo).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, item := range req.Items {
		pi := model.PromotionItem{
			PromotionID: promo.ID, GoodsID: item.GoodsID, SKUID: item.SKUID,
			PromoPrice: item.PromoPrice, PromoStock: item.PromoStock,
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

func GetGroupBuyList(page, pageSize int) (int64, []model.Promotion, error) {
	var total int64
	global.DB.Model(&model.Promotion{}).Where("type = ?", "group").Count(&total)
	var list []model.Promotion
	err := global.DB.Where("type = ?", "group").
		Preload("Items").Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func GetActiveGroupBuys() ([]model.Promotion, error) {
	now := time.Now()
	var list []model.Promotion
	err := global.DB.Where("type = ? AND status = 1 AND start_time <= ? AND end_time > ?", "group", now, now).
		Preload("Items").Find(&list).Error
	return list, err
}

// OpenGroup 开团
func OpenGroup(userID, itemID uint) (*model.GroupOrder, error) {
	var item model.PromotionItem
	if err := global.DB.First(&item, itemID).Error; err != nil {
		return nil, errors.New("拼团商品不存在")
	}
	var promo model.Promotion
	if err := global.DB.First(&promo, item.PromotionID).Error; err != nil || !promo.IsActive() || promo.Type != "group" {
		return nil, errors.New("拼团活动不在进行中")
	}
	// 扣库存（原子操作）
	result := global.DB.Model(&model.PromotionItem{}).
		Where("id = ? AND sold < promo_stock", item.ID).
		Update("sold", gorm.Expr("sold + 1"))
	if result.RowsAffected == 0 {
		return nil, errors.New("已售罄")
	}
	grp := &model.GroupOrder{
		PromotionID: promo.ID, ItemID: item.ID, LeaderID: userID,
		NeedCount: promo.GroupSize, JoinCount: 1,
		ExpireAt: time.Now().Add(time.Duration(promo.GroupTime) * time.Minute),
	}
	global.DB.Create(grp)
	global.DB.Create(&model.GroupOrderMember{GroupOrderID: grp.ID, UserID: userID})
	return grp, nil
}

// JoinGroup 参团
func JoinGroup(userID, groupOrderID uint) error {
	var g model.GroupOrder
	if err := global.DB.First(&g, groupOrderID).Error; err != nil {
		return errors.New("拼团不存在")
	}
	if g.Status != 0 {
		return errors.New("拼团已结束")
	}
	if time.Now().After(g.ExpireAt) {
		global.DB.Model(&g).Update("status", 2)
		return errors.New("拼团已过期")
	}
	// 检查是否已参团
	var count int64
	global.DB.Model(&model.GroupOrderMember{}).Where("group_order_id = ? AND user_id = ?", groupOrderID, userID).Count(&count)
	if count > 0 {
		return errors.New("已参加此拼团")
	}
	// 扣库存（原子操作）
	var item model.PromotionItem
	global.DB.First(&item, g.ItemID)
	result := global.DB.Model(&model.PromotionItem{}).
		Where("id = ? AND sold < promo_stock", item.ID).
		Update("sold", gorm.Expr("sold + 1"))
	if result.RowsAffected == 0 {
		return errors.New("已售罄")
	}
	global.DB.Create(&model.GroupOrderMember{GroupOrderID: groupOrderID, UserID: userID})
	// 原子递增 join_count，DB 层判断是否成团
	global.DB.Model(&model.GroupOrder{}).
		Where("id = ? AND status = 0", groupOrderID).
		Update("join_count", gorm.Expr("join_count + 1"))
	// 检查是否成团（重新读取）
	global.DB.First(&g, groupOrderID)
	if g.JoinCount >= g.NeedCount && g.Status == 0 {
		now := time.Now()
		global.DB.Model(&model.GroupOrder{}).
			Where("id = ? AND status = 0 AND join_count >= need_count", groupOrderID).
			Updates(map[string]interface{}{"status": 1, "finished_at": &now})
	}
	return nil
}

// GetGroupOrderDetail 获取拼团详情
func GetGroupOrderDetail(id uint) (*model.GroupOrder, []model.GroupOrderMember, error) {
	var g model.GroupOrder
	if err := global.DB.First(&g, id).Error; err != nil {
		return nil, nil, errors.New("拼团不存在")
	}
	var members []model.GroupOrderMember
	global.DB.Where("group_order_id = ?", id).Find(&members)
	return &g, members, nil
}
