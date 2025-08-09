package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type AssetRepository interface {
	Create(ctx context.Context, asset *models.Asset) error
	GetByAddress(ctx context.Context, address string) (*models.Asset, error)
	GetBySymbol(ctx context.Context, symbol string) (*models.Asset, error)
}

type assetRepository struct {
	db *gorm.DB
}

func NewAssetRepository() AssetRepository {
	return &assetRepository{
		db: database.GetDB(),
	}
}

func (r *assetRepository) Create(ctx context.Context, asset *models.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *assetRepository) GetByAddress(ctx context.Context, address string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.WithContext(ctx).Where("address = ?", address).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *assetRepository) GetBySymbol(ctx context.Context, symbol string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}
