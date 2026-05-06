package repository

import (
	"fmt"

	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// ---- Cart ----

type gormCartRepo struct{ db *gorm.DB }

func NewCartRepo(db *gorm.DB) CartRepo { return &gormCartRepo{db: db} }

func (r *gormCartRepo) FindByIDsAndUser(ids []uint, userID uint) ([]model.Cart, error) {
	var carts []model.Cart
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).
		Preload("Goods").Preload("SKU").Find(&carts).Error
	return carts, err
}

func (r *gormCartRepo) DeleteByIDsAndUser(tx *gorm.DB, ids []uint, userID uint) error {
	return tx.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.Cart{}).Error
}

// ---- Address ----

type gormAddressRepo struct{ db *gorm.DB }

func NewAddressRepo(db *gorm.DB) AddressRepo { return &gormAddressRepo{db: db} }

func (r *gormAddressRepo) GetByIDAndUser(id, userID uint) (*model.Address, error) {
	var addr model.Address
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&addr).Error
	return &addr, err
}

// ---- SKU ----

type gormSKURepo struct{ db *gorm.DB }

func NewSKURepo(db *gorm.DB) SKURepo { return &gormSKURepo{db: db} }

func (r *gormSKURepo) DeductStock(tx *gorm.DB, skuID uint, quantity int) error {
	result := tx.Model(&model.GoodsSKU{}).Where("id = ? AND stock >= ?", skuID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))
	if result.RowsAffected == 0 {
		return fmt.Errorf("库存不足")
	}
	return nil
}

func (r *gormSKURepo) RestoreStock(tx *gorm.DB, skuID uint, quantity int) error {
	return tx.Model(&model.GoodsSKU{}).Where("id = ?", skuID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}
