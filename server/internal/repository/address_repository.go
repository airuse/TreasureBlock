package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type AddressRepository interface {
	Create(ctx context.Context, address *models.Address) error
	GetByAddress(ctx context.Context, address string) (*models.Address, error)
}

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository() AddressRepository {
	return &addressRepository{
		db: database.GetDB(),
	}
}

func (r *addressRepository) Create(ctx context.Context, address *models.Address) error {
	return r.db.WithContext(ctx).Create(address).Error
}
func (r *addressRepository) GetByAddress(ctx context.Context, address string) (*models.Address, error) {
	var addr models.Address
	err := r.db.WithContext(ctx).Where("address = ?", address).First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}
