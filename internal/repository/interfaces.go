package repository

import (
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// OrderRepo defines data access operations for orders.
type OrderRepo interface {
	GetByID(id uint) (*model.Order, error)
	GetByUserAndID(userID, id uint) (*model.Order, error)
	List(userID uint, status *int8, offset, limit int) ([]model.Order, int64, error)
	Create(tx *gorm.DB, order *model.Order) error
	UpdateStatus(tx *gorm.DB, id uint, status int8) error
}

// CartRepo defines data access operations for cart.
type CartRepo interface {
	FindByIDsAndUser(ids []uint, userID uint) ([]model.Cart, error)
	DeleteByIDsAndUser(tx *gorm.DB, ids []uint, userID uint) error
}

// AddressRepo defines data access operations for addresses.
type AddressRepo interface {
	GetByIDAndUser(id, userID uint) (*model.Address, error)
}

// SKURepo defines data access operations for goods SKU.
type SKURepo interface {
	DeductStock(tx *gorm.DB, skuID uint, quantity int) error
	RestoreStock(tx *gorm.DB, skuID uint, quantity int) error
}
