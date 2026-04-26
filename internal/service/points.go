package service

import (
	"errors"
	"fmt"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// ChangePoints 变动积分（通用方法）
func ChangePoints(userID uint, points int, typ string, refID uint, remark string) error {
	tx := global.DB.Begin()

	var user model.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return errors.New("用户不存在")
	}

	newBalance := user.Points + points
	if newBalance < 0 {
		tx.Rollback()
		return errors.New("积分不足")
	}

	tx.Model(&user).Update("points", newBalance)
	tx.Create(&model.PointsLog{
		UserID:  userID,
		Points:  points,
		Balance: newBalance,
		Type:    typ,
		RefID:   refID,
		Remark:  remark,
	})
	tx.Commit()
	return nil
}

// OrderRewardPoints 订单完成奖励积分（每消费1元=1积分）
func OrderRewardPoints(userID, orderID uint, payAmount int64) error {
	points := int(payAmount / 100) // 分转元
	if points <= 0 {
		return nil
	}
	return ChangePoints(userID, points, "order_reward", orderID, fmt.Sprintf("订单奖励%d积分", points))
}

// SignIn 签到奖励
func SignIn(userID uint) (int, error) {
	points := 10 // 每日签到10积分
	err := ChangePoints(userID, points, "sign_in", 0, "每日签到")
	return points, err
}

// ExchangePoints 积分抵扣（100积分=1元）
func ExchangePoints(userID uint, points int) (int64, error) {
	if points < 100 {
		return 0, errors.New("最少使用100积分")
	}
	discount := int64(points / 100) * 100 // 抵扣金额(分)
	usePoints := int(discount / 100) * 100 // 实际使用积分（100的整数倍）
	err := ChangePoints(userID, -usePoints, "exchange", 0, fmt.Sprintf("积分抵扣%d积分", usePoints))
	return discount, err
}

func GetPointsLog(userID uint, page, pageSize int) ([]model.PointsLog, int64, error) {
	var total int64
	global.DB.Model(&model.PointsLog{}).Where("user_id = ?", userID).Count(&total)
	var list []model.PointsLog
	err := global.DB.Where("user_id = ?", userID).
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
