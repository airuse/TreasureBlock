package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
)

type CoinConfigService interface {
	CreateCoinConfig(ctx context.Context, coinConfig *models.CoinConfig) error
	GetCoinConfigBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error)
}

type coinConfigService struct {
	coinConfigRepo repository.CoinConfigRepository
}

func NewCoinConfigService(coinConfigRepo repository.CoinConfigRepository) CoinConfigService {
	return &coinConfigService{
		coinConfigRepo: coinConfigRepo,
	}
}

/*
创建币种配置
@param ctx context.Context
@param coinConfig *models.CoinConfig
@return error
*/
func (s *coinConfigService) CreateCoinConfig(ctx context.Context, coinConfig *models.CoinConfig) error {
	if coinConfig == nil {
		return fmt.Errorf("coinConfig cannot be nil")
	}
	if coinConfig.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	return s.coinConfigRepo.Create(ctx, coinConfig)
}

/*
根据符号获取币种配置
@param ctx context.Context
@param symbol string
@return *models.CoinConfig, error
*/
func (s *coinConfigService) GetCoinConfigBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	coinConfig, err := s.coinConfigRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get coinConfig by symbol: %w", err)
	}
	return coinConfig, nil
}
