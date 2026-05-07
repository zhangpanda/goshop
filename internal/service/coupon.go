package service

import (
	"errors"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

type CreateCouponReq struct {
	Name      string    `json:"name" binding:"required"`
	Type      int8      `json:"type" binding:"required,oneof=1 2 3"`
	MinAmount int64     `json:"min_amount"`
	Value     int64     `json:"value" binding:"required,min=1"`
	Total     int       `json:"total" binding:"required,min=1"`
	PerLimit  int       `json:"per_limit" binding:"min=1"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

func CreateCoupon(req *CreateCouponReq) (*model.Coupon, error) {
	coupon := model.Coupon{
		Name:      req.Name,
		Type:      req.Type,
		MinAmount: req.MinAmount,
		Value:     req.Value,
		Total:     req.Total,
		PerLimit:  req.PerLimit,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    1,
	}
	if err := global.DB.Create(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func GetCouponList() ([]model.Coupon, error) {
	var list []model.Coupon
	err := global.DB.Where("status = 1 AND end_time > ?", time.Now()).
		Order("id DESC").Find(&list).Error
	return list, err
}

func ReceiveCoupon(userID, couponID uint) error {
	var coupon model.Coupon
	if err := global.DB.First(&coupon, couponID).Error; err != nil {
		return errors.New("优惠券不存在")
	}
	if coupon.Status != 1 {
		return errors.New("优惠券已下架")
	}
	if time.Now().After(coupon.EndTime) {
		return errors.New("优惠券已过期")
	}
	if coupon.Received >= coupon.Total {
		return errors.New("优惠券已领完")
	}

	// 检查领取数量
	var count int64
	global.DB.Model(&model.UserCoupon{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count)
	if int(count) >= coupon.PerLimit {
		return errors.New("已达领取上限")
	}

	tx := global.DB.Begin()
	result := tx.Model(&model.Coupon{}).Where("id = ? AND received < total", couponID).
		Update("received", gorm.Expr("received + 1"))
	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("领取失败")
	}

	uc := model.UserCoupon{UserID: userID, CouponID: couponID, Status: 0}
	if err := tx.Create(&uc).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func GetMyCoupons(userID uint, status *int8) ([]model.UserCoupon, error) {
	var list []model.UserCoupon
	db := global.DB.Where("user_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	err := db.Preload("Coupon").Order("id DESC").Find(&list).Error
	return list, err
}

// CalcCouponDiscount 计算优惠券抵扣金额
func CalcCouponDiscount(coupon *model.Coupon, orderAmount int64) (int64, error) {
	if orderAmount < coupon.MinAmount {
		return 0, errors.New("未达到最低消费")
	}
	switch coupon.Type {
	case model.CouponTypeFull, model.CouponTypeNoLimit:
		if coupon.Value >= orderAmount {
			return orderAmount - 1, nil // 至少付1分
		}
		return coupon.Value, nil
	case model.CouponTypeDiscount:
		// value=85 表示 8.5 折
		discount := orderAmount - orderAmount*coupon.Value/100
		return discount, nil
	}
	return 0, errors.New("未知优惠券类型")
}
