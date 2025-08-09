package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type CoinConfigRepository interface {
	Create(ctx context.Context, coinConfig *models.CoinConfig) error
	GetBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error)
}

type coinConfigRepository struct {
	db *gorm.DB
}

func NewCoinConfigRepository() CoinConfigRepository {
	return &coinConfigRepository{
		db: database.GetDB(),
	}
}

func (r *coinConfigRepository) Create(ctx context.Context, coinConfig *models.CoinConfig) error {
	return r.db.WithContext(ctx).Create(coinConfig).Error
}

func (r *coinConfigRepository) GetBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error) {
	var coinConfig models.CoinConfig
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).First(&coinConfig).Error
	if err != nil {
		return nil, err
	}
	return &coinConfig, nil
}
