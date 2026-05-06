package repository

import "gorm.io/gorm"

// Repos holds all repository instances. Initialized once at startup.
var Repos *Registry

type Registry struct {
	Order   OrderRepo
	Cart    CartRepo
	Address AddressRepo
	SKU     SKURepo
}

// Init creates all repository instances from the given DB.
func Init(db *gorm.DB) {
	Repos = &Registry{
		Order:   NewOrderRepo(db),
		Cart:    NewCartRepo(db),
		Address: NewAddressRepo(db),
		SKU:     NewSKURepo(db),
	}
}
