package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type ContractParseResultRepository interface {
	Create(ctx context.Context, r *models.ContractParseResult) error
	CreateBatch(ctx context.Context, results []*models.ContractParseResult) error
	GetByTxHash(ctx context.Context, txHash string) ([]*models.ContractParseResult, error)
	GetByTxHashAndLogIndex(ctx context.Context, txHash string, logIndex uint) (*models.ContractParseResult, error)
}

type contractParseResultRepository struct {
	db *gorm.DB
}

func NewContractParseResultRepository() ContractParseResultRepository {
	return &contractParseResultRepository{db: database.GetDB()}
}

func (r *contractParseResultRepository) Create(ctx context.Context, res *models.ContractParseResult) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *contractParseResultRepository) CreateBatch(ctx context.Context, results []*models.ContractParseResult) error {
	if len(results) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&results).Error
}

func (r *contractParseResultRepository) GetByTxHash(ctx context.Context, txHash string) ([]*models.ContractParseResult, error) {
	var out []*models.ContractParseResult
	if err := r.db.WithContext(ctx).Where("tx_hash = ?", txHash).Order("log_index ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *contractParseResultRepository) GetByTxHashAndLogIndex(ctx context.Context, txHash string, logIndex uint) (*models.ContractParseResult, error) {
	var result models.ContractParseResult
	err := r.db.WithContext(ctx).Where("tx_hash = ? AND log_index = ?", txHash, logIndex).First(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
