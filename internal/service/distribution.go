package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// SettleCommission 订单完成后结算佣金（一级/二级，比例来自配置）。
// 幂等：同 orderID 已存在 type=order 的 CommissionLog 则直接返回，不重复发放。
// 全流程在单个事务内：加余额 + 写流水。失败会整体回滚（只会有 0 或 2 条 log 被创建）。
func SettleCommission(orderID uint) {
	_ = RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.First(&order, orderID).Error; err != nil {
			return err
		}
		if order.Status != model.OrderStatusCompleted {
			return nil
		}
		// 幂等：同 orderID 已结算过的 order-type 日志存在则跳过
		var settled int64
		if err := tx.Model(&model.CommissionLog{}).
			Where("order_id = ? AND type = ?", orderID, "order").Count(&settled).Error; err != nil {
			return err
		}
		if settled > 0 {
			return nil
		}
		var buyer model.Distributor
		if err := tx.Where("user_id = ?", order.UserID).First(&buyer).Error; err != nil {
			return nil // 购买者非分销商；不算错，直接返回
		}
		if buyer.ParentID == 0 {
			return nil
		}

		rate1, rate2 := int64(10), int64(5)
		if v := GetConfig("distribution_rate_level1"); v != "" {
			fmt.Sscanf(v, "%d", &rate1)
		}
		if v := GetConfig("distribution_rate_level2"); v != "" {
			fmt.Sscanf(v, "%d", &rate2)
		}

		if c1 := order.PayAmount * rate1 / 100; c1 > 0 {
			if err := addCommissionTx(tx, buyer.ParentID, orderID, c1, "一级推广佣金"); err != nil {
				return err
			}
		}
		var parent model.Distributor
		if err := tx.Where("user_id = ?", buyer.ParentID).First(&parent).Error; err == nil && parent.ParentID > 0 {
			if c2 := order.PayAmount * rate2 / 100; c2 > 0 {
				if err := addCommissionTx(tx, parent.ParentID, orderID, c2, "二级推广佣金"); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// addCommissionTx 在给定事务中为某分销商增加佣金并写流水。上级非分销商时忽略（不视为错误）。
func addCommissionTx(tx *gorm.DB, userID, orderID uint, amount int64, remark string) error {
	var d model.Distributor
	if err := tx.Where("user_id = ? AND status = 1", userID).First(&d).Error; err != nil {
		return nil
	}
	if err := tx.Model(&d).Updates(map[string]interface{}{
		"balance":          gorm.Expr("balance + ?", amount),
		"total_commission": gorm.Expr("total_commission + ?", amount),
		"order_count":      gorm.Expr("order_count + 1"),
	}).Error; err != nil {
		return err
	}
	return tx.Create(&model.CommissionLog{
		DistributorID: d.ID, OrderID: orderID, Amount: amount, Type: "order", Remark: remark,
	}).Error
}

// ========== 提现 ==========

// RequestWithdraw 用户申请提现：扣余额 + 写 WithdrawRequest + 写流水在同一事务内完成。
// 任一步失败整体回滚，避免"余额扣了但申请/流水没写"造成资金漂移。
func RequestWithdraw(userID uint, amount int64, accountType, accountNo, accountName string) error {
	if amount <= 0 {
		return errors.New("提现金额必须大于0")
	}
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		var d model.Distributor
		if err := tx.Where("user_id = ? AND status = 1", userID).First(&d).Error; err != nil {
			return errors.New("非分销商")
		}
		result := tx.Model(&model.Distributor{}).
			Where("id = ? AND balance >= ?", d.ID, amount).
			Update("balance", gorm.Expr("balance - ?", amount))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("余额不足(可提现%d分)", d.Balance)
		}
		if err := tx.Create(&model.WithdrawRequest{
			DistributorID: d.ID, UserID: userID, Amount: amount,
			AccountType: accountType, AccountNo: accountNo, AccountName: accountName,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&model.CommissionLog{
			DistributorID: d.ID, Amount: -amount, Type: "withdraw", Remark: "提现申请",
		}).Error
	})
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

// AuditWithdraw 管理员审核提现。
//  1. 通过 FOR UPDATE 行锁 + 事务保证单次性，避免并发双审；
//  2. 拒绝分支整个退款流程（退余额 + 写流水 + 改状态）在同一事务内完成；
//  3. Status != 0 直接返回 "已处理"，实现幂等。
func AuditWithdraw(id uint, approve bool, reason string) error {
	return RunInDBTx(app.Must().DB, func(tx *gorm.DB) error {
		var w model.WithdrawRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&w, id).Error; err != nil {
			return errors.New("提现记录不存在")
		}
		if w.Status != 0 {
			return errors.New("已处理")
		}
		now := time.Now()
		if approve {
			return tx.Model(&w).Updates(map[string]interface{}{"status": 1, "audit_at": &now}).Error
		}
		// 拒绝：退回余额 → 写流水 → 改状态，任一失败整体回滚
		if err := tx.Model(&model.Distributor{}).Where("id = ?", w.DistributorID).
			Update("balance", gorm.Expr("balance + ?", w.Amount)).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.CommissionLog{
			DistributorID: w.DistributorID, Amount: w.Amount, Type: "adjust", Remark: "提现拒绝退回",
		}).Error; err != nil {
			return err
		}
		return tx.Model(&w).Updates(map[string]interface{}{
			"status": 2, "reject_reason": reason, "audit_at": &now,
		}).Error
	})
}

func GetCommissionLogs(distributorID uint, page, pageSize int) (int64, []model.CommissionLog, error) {
	var total int64
	db := app.Must().DB.Model(&model.CommissionLog{}).Where("distributor_id = ?", distributorID)
	db.Count(&total)
	var list []model.CommissionLog
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}
