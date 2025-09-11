package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
)

type BTCTransactionRepository interface {
	Create(tx *models.BTCTransaction) error
	CreateBatch(txs []*models.BTCTransaction) error
}

type btcTransactionRepository struct{}

func NewBTCTransactionRepository() BTCTransactionRepository {
	return &btcTransactionRepository{}
}

func (r *btcTransactionRepository) Create(tx *models.BTCTransaction) error {
	return database.GetDB().Create(tx).Error
}

func (r *btcTransactionRepository) CreateBatch(txs []*models.BTCTransaction) error {
	if len(txs) == 0 {
		return nil
	}
	return database.GetDB().Create(&txs).Error
}
