package repository

import (
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

type SolParsedExtraRepository interface {
	CreateBatch(ctx context.Context, extras []*models.SolParsedExtra) error
	GetByTxID(ctx context.Context, txID string) ([]*models.SolParsedExtra, error)
}

type solParsedExtraRepository struct {
	db *gorm.DB
}

func NewSolParsedExtraRepository(db *gorm.DB) SolParsedExtraRepository {
	return &solParsedExtraRepository{db: db}
}

func (r *solParsedExtraRepository) CreateBatch(ctx context.Context, extras []*models.SolParsedExtra) error {
	if len(extras) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&extras).Error
}

func (r *solParsedExtraRepository) GetByTxID(ctx context.Context, txID string) ([]*models.SolParsedExtra, error) {
	var extras []*models.SolParsedExtra
	err := r.db.WithContext(ctx).Where("tx_id = ?", txID).Order("id ASC").Find(&extras).Error
	return extras, err
}
