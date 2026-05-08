package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// ========== 分销商 ==========

func ApplyDistributor(userID, parentID uint) (*model.Distributor, error) {
	var exists model.Distributor
	if app.Must().DB.Where("user_id = ?", userID).First(&exists).Error == nil {
		return &exists, nil // 已是分销商
	}
	d := model.Distributor{UserID: userID, ParentID: parentID, Level: 1, Status: 1}
	return &d, app.Must().DB.Create(&d).Error
}

func GetDistributorByUser(userID uint) (*model.Distributor, error) {
	var d model.Distributor
	if err := app.Must().DB.Where("user_id = ?", userID).First(&d).Error; err != nil {
		return nil, errors.New("非分销商")
	}
	return &d, nil
}

func GetDistributorList(page, pageSize int) (int64, []model.Distributor, error) {
	var total int64
	app.Must().DB.Model(&model.Distributor{}).Count(&total)
	var list []model.Distributor
	err := app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func GetSubDistributors(userID uint) ([]model.Distributor, error) {
	var list []model.Distributor
	return list, app.Must().DB.Where("parent_id = ?", userID).Find(&list).Error
}

// ========== 佣金计算 ==========

// SettleCommission 订单完成后结算佣金（一级10%，二级5%）
func SettleCommission(orderID uint) {
	var order model.Order
	if app.Must().DB.First(&order, orderID).Error != nil || order.Status != model.OrderStatusCompleted {
		return
	}
	// 下单用户的上级
	var buyer model.Distributor
	if app.Must().DB.Where("user_id = ?", order.UserID).First(&buyer).Error != nil {
		return
	}
	if buyer.ParentID == 0 {
		return
	}
	rate1 := int64(10) // 一级佣金比例%（默认值）
	rate2 := int64(5)  // 二级佣金比例%（默认值）
	if v := GetConfig("distribution_rate_level1"); v != "" {
		fmt.Sscanf(v, "%d", &rate1)
	}
	if v := GetConfig("distribution_rate_level2"); v != "" {
		fmt.Sscanf(v, "%d", &rate2)
	}

	// 一级佣金
	commission1 := order.PayAmount * rate1 / 100
	if commission1 > 0 {
		addCommission(buyer.ParentID, orderID, commission1, "一级推广佣金")
	}
	// 二级佣金
	var parent model.Distributor
	if app.Must().DB.Where("user_id = ?", buyer.ParentID).First(&parent).Error == nil && parent.ParentID > 0 {
		commission2 := order.PayAmount * rate2 / 100
		if commission2 > 0 {
			addCommission(parent.ParentID, orderID, commission2, "二级推广佣金")
		}
	}
}

func addCommission(userID, orderID uint, amount int64, remark string) {
	var d model.Distributor
	if app.Must().DB.Where("user_id = ? AND status = 1", userID).First(&d).Error != nil {
		return
	}
	app.Must().DB.Model(&d).Updates(map[string]interface{}{
		"balance":          gorm.Expr("balance + ?", amount),
		"total_commission": gorm.Expr("total_commission + ?", amount),
		"order_count":      gorm.Expr("order_count + 1"),
	})
	app.Must().DB.Create(&model.CommissionLog{
		DistributorID: d.ID, OrderID: orderID, Amount: amount, Type: "order", Remark: remark,
	})
}

// ========== 提现 ==========

func RequestWithdraw(userID uint, amount int64, accountType, accountNo, accountName string) error {
	if amount <= 0 {
		return errors.New("提现金额必须大于0")
	}
	var d model.Distributor
	if err := app.Must().DB.Where("user_id = ? AND status = 1", userID).First(&d).Error; err != nil {
		return errors.New("非分销商")
	}
	// 原子扣减余额
	result := app.Must().DB.Model(&d).Where("balance >= ?", amount).
		Update("balance", gorm.Expr("balance - ?", amount))
	if result.RowsAffected == 0 {
		return fmt.Errorf("余额不足(可提现%d分)", d.Balance)
	}
	app.Must().DB.Create(&model.WithdrawRequest{
		DistributorID: d.ID, UserID: userID, Amount: amount,
		AccountType: accountType, AccountNo: accountNo, AccountName: accountName,
	})
	app.Must().DB.Create(&model.CommissionLog{
		DistributorID: d.ID, Amount: -amount, Type: "withdraw", Remark: "提现申请",
	})
	return nil
}

func GetWithdrawList(page, pageSize int, status *int8) (int64, []model.WithdrawRequest, error) {
	var total int64
	db := app.Must().DB.Model(&model.WithdrawRequest{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	db.Count(&total)
	var list []model.WithdrawRequest
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func AuditWithdraw(id uint, approve bool, reason string) error {
	var w model.WithdrawRequest
	if err := app.Must().DB.First(&w, id).Error; err != nil {
		return errors.New("提现记录不存在")
	}
	if w.Status != 0 {
		return errors.New("已处理")
	}
	now := time.Now()
	if approve {
		app.Must().DB.Model(&w).Updates(map[string]interface{}{"status": 1, "audit_at": &now})
	} else {
		// 退回余额
		app.Must().DB.Model(&model.Distributor{}).Where("id = ?", w.DistributorID).
			Update("balance", gorm.Expr("balance + ?", w.Amount))
		app.Must().DB.Create(&model.CommissionLog{
			DistributorID: w.DistributorID, Amount: w.Amount, Type: "adjust", Remark: "提现拒绝退回",
		})
		app.Must().DB.Model(&w).Updates(map[string]interface{}{"status": 2, "reject_reason": reason, "audit_at": &now})
	}
	return nil
}

func GetCommissionLogs(distributorID uint, page, pageSize int) (int64, []model.CommissionLog, error) {
	var total int64
	db := app.Must().DB.Model(&model.CommissionLog{}).Where("distributor_id = ?", distributorID)
	db.Count(&total)
	var list []model.CommissionLog
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}
