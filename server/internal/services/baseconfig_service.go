package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
)

type BaseConfigService interface {
	GetByConfigType(ctx context.Context, configType uint8) ([]*models.BaseConfig, error)
	GetByConfigGroup(ctx context.Context, configGroup string) ([]*models.BaseConfig, error)
	GetByConfigKey(ctx context.Context, configKey string, configType uint8, configGroup string) (*models.BaseConfig, error)
}

type baseConfigService struct {
	baseConfigRepo repository.BaseConfigRepository
}

func NewBaseConfigService(baseConfigRepo repository.BaseConfigRepository) BaseConfigService {
	return &baseConfigService{
		baseConfigRepo: baseConfigRepo,
	}
}

func (s *baseConfigService) GetByConfigType(ctx context.Context, configType uint8) ([]*models.BaseConfig, error) {
	if configType == 0 {
		return nil, fmt.Errorf("configType cannot be zero")
	}
	return s.baseConfigRepo.GetByConfigType(ctx, configType)
}

func (s *baseConfigService) GetByConfigGroup(ctx context.Context, configGroup string) ([]*models.BaseConfig, error) {
	if configGroup == "" {
		return nil, fmt.Errorf("configGroup cannot be empty")
	}
	return s.baseConfigRepo.GetByConfigGroup(ctx, configGroup)
}

func (s *baseConfigService) GetByConfigKey(ctx context.Context, configKey string, configType uint8, configGroup string) (*models.BaseConfig, error) {
	if configKey == "" {
		return nil, fmt.Errorf("configKey cannot be empty")
	}
	if configType == 0 {
		return nil, fmt.Errorf("configType cannot be zero")
	}
	if configGroup == "" {
		return nil, fmt.Errorf("configGroup cannot be empty")
	}
	return s.baseConfigRepo.GetByConfigKey(ctx, configKey, configType, configGroup)
}
