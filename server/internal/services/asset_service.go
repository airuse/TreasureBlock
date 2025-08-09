package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
)

type AssetService interface {
	CreateAsset(ctx context.Context, asset *models.Asset) error
	GetAssetByAddress(ctx context.Context, address string) (*models.Asset, error)
	GetAssetBySymbol(ctx context.Context, symbol string) (*models.Asset, error)
}

type assetService struct {
	assetRepo repository.AssetRepository
}

func NewAssetService(assetRepo repository.AssetRepository) AssetService {
	return &assetService{
		assetRepo: assetRepo,
	}
}

/*
创建资产
@param ctx context.Context
@param asset *models.Asset
@return error
*/
func (s *assetService) CreateAsset(ctx context.Context, asset *models.Asset) error {
	if asset == nil {
		return fmt.Errorf("asset cannot be nil")
	}
	if asset.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if asset.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	return s.assetRepo.Create(ctx, asset)
}

/*
根据地址获取资产
@param ctx context.Context
@param address string
@return *models.Asset, error
*/
func (s *assetService) GetAssetByAddress(ctx context.Context, address string) (*models.Asset, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	asset, err := s.assetRepo.GetByAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset by address: %w", err)
	}
	return asset, nil
}

/*
根据符号获取资产
@param ctx context.Context
@param symbol string
@return *models.Asset, error
*/
func (s *assetService) GetAssetBySymbol(ctx context.Context, symbol string) (*models.Asset, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	asset, err := s.assetRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset by symbol: %w", err)
	}
	return asset, nil
}
