package repository

import (
	"errors"

	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

type gormOrderRepo struct{ db *gorm.DB }

func NewOrderRepo(db *gorm.DB) OrderRepo { return &gormOrderRepo{db: db} }

func (r *gormOrderRepo) GetByID(id uint) (*model.Order, error) {
	var o model.Order
	err := r.db.Preload("Items").First(&o, id).Error
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	return &o, nil
}

func (r *gormOrderRepo) GetByUserAndID(userID, id uint) (*model.Order, error) {
	var o model.Order
	err := r.db.Where("id = ? AND user_id = ?", id, userID).Preload("Items").First(&o).Error
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	return &o, nil
}

func (r *gormOrderRepo) List(userID uint, status *int8, offset, limit int) ([]model.Order, int64, error) {
	db := r.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var total int64
	db.Count(&total)
	var list []model.Order
	err := db.Preload("Items").Order("id DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *gormOrderRepo) Create(tx *gorm.DB, order *model.Order) error {
	return tx.Create(order).Error
}

func (r *gormOrderRepo) UpdateStatus(tx *gorm.DB, id uint, status int8) error {
	return tx.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
