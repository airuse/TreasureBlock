package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type BaseConfigRepository interface {
	GetByConfigType(ctx context.Context, configType uint8) ([]*models.BaseConfig, error)
	GetByConfigGroup(ctx context.Context, configGroup string) ([]*models.BaseConfig, error)
	GetByConfigKey(ctx context.Context, configKey string, configType uint8, configGroup string) (*models.BaseConfig, error)
}
type baseConfigRepository struct {
	db *gorm.DB
}

func NewBaseConfigRepository() BaseConfigRepository {
	return &baseConfigRepository{
		db: database.GetDB(),
	}
}

func (r *baseConfigRepository) GetByConfigType(ctx context.Context, configType uint8) ([]*models.BaseConfig, error) {
	var baseConfigs []*models.BaseConfig
	err := r.db.WithContext(ctx).Where("config_type = ?", configType).Find(&baseConfigs).Error
	if err != nil {
		return nil, err
	}
	return baseConfigs, nil
}

func (r *baseConfigRepository) GetByConfigGroup(ctx context.Context, configGroup string) ([]*models.BaseConfig, error) {
	var baseConfigs []*models.BaseConfig
	err := r.db.WithContext(ctx).Where("`group` = ?", configGroup).Find(&baseConfigs).Error
	if err != nil {
		return nil, err
	}
	return baseConfigs, nil
}

func (r *baseConfigRepository) GetByConfigKey(ctx context.Context, configKey string, configType uint8, configGroup string) (*models.BaseConfig, error) {
	var baseConfig models.BaseConfig
	err := r.db.WithContext(ctx).Where("config_key = ? AND config_type = ? AND `group` = ?", configKey, configType, configGroup).First(&baseConfig).Error
	if err != nil {
		return nil, err
	}
	return &baseConfig, nil
}
