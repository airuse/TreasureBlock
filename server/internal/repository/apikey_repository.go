package repository

import (
	"errors"
	"time"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

type APIKeyRepository interface {
	Create(apiKey *models.APIKey) error
	GetByID(id uint) (*models.APIKey, error)
	GetByAPIKey(apiKey string) (*models.APIKey, error)
	GetByUserID(userID uint) ([]*models.APIKey, error)
	Update(apiKey *models.APIKey) error
	Delete(id uint) error
	UpdateLastUsed(id uint, lastUsedAt time.Time) error
	IncrementUsageCount(id uint) error
	GetActiveByAPIKey(apiKey string) (*models.APIKey, error)
	IsAPIKeyExists(apiKey string) (bool, error)
}

type apiKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *apiKeyRepository) GetByID(id uint) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Preload("User").First(&apiKey, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API密钥不存在")
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) GetByAPIKey(apiKey string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("User").Where("api_key = ?", apiKey).First(&key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API密钥不存在")
		}
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) GetByUserID(userID uint) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&apiKeys).Error
	return apiKeys, err
}

func (r *apiKeyRepository) Update(apiKey *models.APIKey) error {
	return r.db.Save(apiKey).Error
}

func (r *apiKeyRepository) Delete(id uint) error {
	return r.db.Delete(&models.APIKey{}, id).Error
}

func (r *apiKeyRepository) UpdateLastUsed(id uint, lastUsedAt time.Time) error {
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", lastUsedAt).Error
}

func (r *apiKeyRepository) IncrementUsageCount(id uint) error {
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

func (r *apiKeyRepository) GetActiveByAPIKey(apiKey string) (*models.APIKey, error) {
	var key models.APIKey
	query := r.db.Preload("User").Where("api_key = ? AND is_active = ?", apiKey, true)
	
	// 检查是否过期
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
	
	err := query.First(&key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API密钥无效或已过期")
		}
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) IsAPIKeyExists(apiKey string) (bool, error) {
	var count int64
	err := r.db.Model(&models.APIKey{}).Where("api_key = ?", apiKey).Count(&count).Error
	return count > 0, err
}
