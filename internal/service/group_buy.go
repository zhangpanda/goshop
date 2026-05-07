package service

import (
	"errors"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// isDuplicateKeyError 检测 MySQL/SQLite 等下的唯一约束冲突（用于参团防重兜底）。
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") ||
		strings.Contains(s, "unique constraint") ||
		strings.Contains(s, "constraint failed")
}

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
	err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Create(&promo).Error; err != nil {
			return err
		}
		for _, item := range req.Items {
			pi := model.PromotionItem{
				PromotionID: promo.ID, GoodsID: item.GoodsID, SKUID: item.SKUID,
				PromoPrice: item.PromoPrice, PromoStock: item.PromoStock,
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

// OpenGroup 开团：扣活动库存、写入团单与团长成员在同一事务内完成。
func OpenGroup(userID, itemID uint) (*model.GroupOrder, error) {
	var out *model.GroupOrder
	err := RunInDBTx(global.DB, func(tx *gorm.DB) error {
		var item model.PromotionItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, itemID).Error; err != nil {
			return errors.New("拼团商品不存在")
		}
		var promo model.Promotion
		if err := tx.First(&promo, item.PromotionID).Error; err != nil {
			return errors.New("拼团活动不存在")
		}
		if !promo.IsActive() || promo.Type != "group" {
			return errors.New("拼团活动不在进行中")
		}
		res := tx.Model(&model.PromotionItem{}).Where("id = ? AND sold < promo_stock", item.ID).
			Update("sold", gorm.Expr("sold + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("已售罄")
		}
		grp := model.GroupOrder{
			PromotionID: promo.ID, ItemID: item.ID, LeaderID: userID,
			NeedCount: promo.GroupSize, JoinCount: 1,
			ExpireAt: time.Now().Add(time.Duration(promo.GroupTime) * time.Minute),
		}
		if err := tx.Create(&grp).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.GroupOrderMember{GroupOrderID: grp.ID, UserID: userID}).Error; err != nil {
			if isDuplicateKeyError(err) {
				return errors.New("请勿重复提交开团请求")
			}
			return err
		}
		out = &grp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JoinGroup 参团：锁团单行、扣库存、插入成员、递增人数与成团判定在同一事务内完成。
func JoinGroup(userID, groupOrderID uint) error {
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		var g model.GroupOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&g, groupOrderID).Error; err != nil {
			return errors.New("拼团不存在")
		}
		if g.Status != 0 {
			return errors.New("拼团已结束")
		}
		if time.Now().After(g.ExpireAt) {
			if err := tx.Model(&g).Update("status", 2).Error; err != nil {
				return err
			}
			return errors.New("拼团已过期")
		}
		var count int64
		if err := tx.Model(&model.GroupOrderMember{}).Where("group_order_id = ? AND user_id = ?", groupOrderID, userID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("已参加此拼团")
		}
		var item model.PromotionItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, g.ItemID).Error; err != nil {
			return errors.New("拼团商品不存在")
		}
		res := tx.Model(&model.PromotionItem{}).Where("id = ? AND sold < promo_stock", item.ID).
			Update("sold", gorm.Expr("sold + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("已售罄")
		}
		if err := tx.Create(&model.GroupOrderMember{GroupOrderID: groupOrderID, UserID: userID}).Error; err != nil {
			if isDuplicateKeyError(err) {
				return errors.New("已参加此拼团")
			}
			return err
		}
		if err := tx.Model(&model.GroupOrder{}).Where("id = ? AND status = ?", groupOrderID, 0).
			Update("join_count", gorm.Expr("join_count + 1")).Error; err != nil {
			return err
		}
		if err := tx.First(&g, groupOrderID).Error; err != nil {
			return err
		}
		if g.JoinCount >= g.NeedCount && g.Status == 0 {
			now := time.Now()
			if err := tx.Model(&model.GroupOrder{}).
				Where("id = ? AND status = 0 AND join_count >= need_count", groupOrderID).
				Updates(map[string]interface{}{"status": 1, "finished_at": &now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
