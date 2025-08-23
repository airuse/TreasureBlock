package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type CoinConfigRepository interface {
	Create(ctx context.Context, coinConfig *models.CoinConfig) error
	GetByID(ctx context.Context, id uint) (*models.CoinConfig, error)
	GetBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error)
	GetByContractAddress(ctx context.Context, contractAddress string) (*models.CoinConfig, error)
	GetByChain(ctx context.Context, chain string) ([]*models.CoinConfig, error)
	GetByChainAndStatus(ctx context.Context, chain string, status int8) ([]*models.CoinConfig, error)
	GetStablecoins(ctx context.Context, chain string) ([]*models.CoinConfig, error)
	GetVerifiedTokens(ctx context.Context, chain string) ([]*models.CoinConfig, error)
	List(ctx context.Context, offset, limit int, chain string) ([]*models.CoinConfig, int64, error)
	Update(ctx context.Context, coinConfig *models.CoinConfig) error
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context) ([]*models.CoinConfig, error)
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

func (r *coinConfigRepository) GetByID(ctx context.Context, id uint) (*models.CoinConfig, error) {
	var coinConfig models.CoinConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&coinConfig).Error
	if err != nil {
		return nil, err
	}
	return &coinConfig, nil
}

func (r *coinConfigRepository) GetBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error) {
	var coinConfig models.CoinConfig
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).First(&coinConfig).Error
	if err != nil {
		return nil, err
	}
	return &coinConfig, nil
}

func (r *coinConfigRepository) GetByContractAddress(ctx context.Context, contractAddress string) (*models.CoinConfig, error) {
	var coinConfig models.CoinConfig
	err := r.db.WithContext(ctx).Where("contract_addr = ?", contractAddress).First(&coinConfig).Error
	if err != nil {
		return nil, err
	}
	return &coinConfig, nil
}

func (r *coinConfigRepository) GetByChain(ctx context.Context, chain string) ([]*models.CoinConfig, error) {
	var coinConfigs []*models.CoinConfig
	err := r.db.WithContext(ctx).Where("chain_name = ? AND status = ?", chain, 1).Find(&coinConfigs).Error
	if err != nil {
		return nil, err
	}
	return coinConfigs, nil
}

func (r *coinConfigRepository) GetByChainAndStatus(ctx context.Context, chain string, status int8) ([]*models.CoinConfig, error) {
	var coinConfigs []*models.CoinConfig
	err := r.db.WithContext(ctx).Where("chain_name = ? AND status = ?", chain, status).Find(&coinConfigs).Error
	if err != nil {
		return nil, err
	}
	return coinConfigs, nil
}

func (r *coinConfigRepository) GetStablecoins(ctx context.Context, chain string) ([]*models.CoinConfig, error) {
	var coinConfigs []*models.CoinConfig
	err := r.db.WithContext(ctx).Where("chain_name = ? AND is_stablecoin = ? AND status = ?", chain, true, 1).Find(&coinConfigs).Error
	if err != nil {
		return nil, err
	}
	return coinConfigs, nil
}

func (r *coinConfigRepository) GetVerifiedTokens(ctx context.Context, chain string) ([]*models.CoinConfig, error) {
	var coinConfigs []*models.CoinConfig
	err := r.db.WithContext(ctx).Where("chain_name = ? AND is_verified = ? AND status = ?", chain, true, 1).Find(&coinConfigs).Error
	if err != nil {
		return nil, err
	}
	return coinConfigs, nil
}

func (r *coinConfigRepository) List(ctx context.Context, offset, limit int, chain string) ([]*models.CoinConfig, int64, error) {
	var coinConfigs []*models.CoinConfig
	var total int64

	query := r.db.WithContext(ctx)
	if chain != "" {
		query = query.Where("chain_name = ?", chain)
	}

	// 获取总数
	err := query.Model(&models.CoinConfig{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Offset(offset).Limit(limit).Order("market_cap_rank ASC, symbol ASC").Find(&coinConfigs).Error
	if err != nil {
		return nil, 0, err
	}

	return coinConfigs, total, nil
}

func (r *coinConfigRepository) Update(ctx context.Context, coinConfig *models.CoinConfig) error {
	return r.db.WithContext(ctx).Save(coinConfig).Error
}

func (r *coinConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CoinConfig{}, id).Error
}

func (r *coinConfigRepository) GetAll(ctx context.Context) ([]*models.CoinConfig, error) {
	var coinConfigs []*models.CoinConfig
	err := r.db.WithContext(ctx).Find(&coinConfigs).Error
	if err != nil {
		return nil, err
	}
	return coinConfigs, nil
}
