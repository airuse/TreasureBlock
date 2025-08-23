package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
)

type CoinConfigService interface {
	CreateCoinConfig(ctx context.Context, coinConfig *models.CoinConfig) error
	GetCoinConfigByID(ctx context.Context, id uint) (*models.CoinConfig, error)
	GetCoinConfigBySymbol(ctx context.Context, symbol string) (*models.CoinConfig, error)
	GetCoinConfigByContractAddress(ctx context.Context, contractAddress string) (*models.CoinConfig, error)
	GetCoinConfigsByChain(ctx context.Context, chain string) ([]*models.CoinConfig, error)
	GetStablecoins(ctx context.Context, chain string) ([]*models.CoinConfig, error)
	GetVerifiedTokens(ctx context.Context, chain string) ([]*models.CoinConfig, error)
	ListCoinConfigs(ctx context.Context, page, pageSize int, chain string) ([]*models.CoinConfig, int64, error)
	UpdateCoinConfig(ctx context.Context, coinConfig *models.CoinConfig) error
	DeleteCoinConfig(ctx context.Context, id uint) error
	GetAllCoinConfigs(ctx context.Context) ([]*models.CoinConfig, error)
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
	if coinConfig.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return s.coinConfigRepo.Create(ctx, coinConfig)
}

/*
根据ID获取币种配置
@param ctx context.Context
@param id uint
@return *models.CoinConfig, error
*/
func (s *coinConfigService) GetCoinConfigByID(ctx context.Context, id uint) (*models.CoinConfig, error) {
	if id == 0 {
		return nil, fmt.Errorf("id cannot be zero")
	}
	coinConfig, err := s.coinConfigRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get coinConfig by id: %w", err)
	}
	return coinConfig, nil
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

/*
根据合约地址获取币种配置
@param ctx context.Context
@param contractAddress string
@return *models.CoinConfig, error
*/
func (s *coinConfigService) GetCoinConfigByContractAddress(ctx context.Context, contractAddress string) (*models.CoinConfig, error) {
	if contractAddress == "" {
		return nil, fmt.Errorf("contract address cannot be empty")
	}
	coinConfig, err := s.coinConfigRepo.GetByContractAddress(ctx, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get coinConfig by contract address: %w", err)
	}
	return coinConfig, nil
}

/*
根据链名称获取币种配置列表
@param ctx context.Context
@param chain string
@return []*models.CoinConfig, error
*/
func (s *coinConfigService) GetCoinConfigsByChain(ctx context.Context, chain string) ([]*models.CoinConfig, error) {
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}
	coinConfigs, err := s.coinConfigRepo.GetByChain(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get coinConfigs by chain: %w", err)
	}
	return coinConfigs, nil
}

/*
获取指定链的稳定币列表
@param ctx context.Context
@param chain string
@return []*models.CoinConfig, error
*/
func (s *coinConfigService) GetStablecoins(ctx context.Context, chain string) ([]*models.CoinConfig, error) {
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}
	coinConfigs, err := s.coinConfigRepo.GetStablecoins(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get stablecoins by chain: %w", err)
	}
	return coinConfigs, nil
}

/*
获取指定链的已验证代币列表
@param ctx context.Context
@param chain string
@return []*models.CoinConfig, error
*/
func (s *coinConfigService) GetVerifiedTokens(ctx context.Context, chain string) ([]*models.CoinConfig, error) {
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}
	coinConfigs, err := s.coinConfigRepo.GetVerifiedTokens(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get verified tokens by chain: %w", err)
	}
	return coinConfigs, nil
}

/*
分页获取币种配置列表
@param ctx context.Context
@param page int
@param pageSize int
@param chain string
@return []*models.CoinConfig, int64, error
*/
func (s *coinConfigService) ListCoinConfigs(ctx context.Context, page, pageSize int, chain string) ([]*models.CoinConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	coinConfigs, total, err := s.coinConfigRepo.List(ctx, offset, pageSize, chain)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list coin configs: %w", err)
	}

	return coinConfigs, total, nil
}

/*
更新币种配置
@param ctx context.Context
@param coinConfig *models.CoinConfig
@return error
*/
func (s *coinConfigService) UpdateCoinConfig(ctx context.Context, coinConfig *models.CoinConfig) error {
	if coinConfig == nil {
		return fmt.Errorf("coinConfig cannot be nil")
	}
	if coinConfig.ID == 0 {
		return fmt.Errorf("coinConfig id cannot be zero")
	}
	return s.coinConfigRepo.Update(ctx, coinConfig)
}

/*
删除币种配置
@param ctx context.Context
@param id uint
@return error
*/
func (s *coinConfigService) DeleteCoinConfig(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("id cannot be zero")
	}
	return s.coinConfigRepo.Delete(ctx, id)
}

/*
获取所有币种配置
@param ctx context.Context
@return []*models.CoinConfig, error
*/
func (s *coinConfigService) GetAllCoinConfigs(ctx context.Context) ([]*models.CoinConfig, error) {
	coinConfigs, err := s.coinConfigRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all coin configs: %w", err)
	}
	return coinConfigs, nil
}
