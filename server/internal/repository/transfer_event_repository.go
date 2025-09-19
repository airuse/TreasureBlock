package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
)

type TransferEventRepository interface {
	CreateBatch(events []*models.TransferEvent) error
}

type transferEventRepository struct{}

func NewTransferEventRepository() TransferEventRepository {
	return &transferEventRepository{}
}

func (r *transferEventRepository) CreateBatch(events []*models.TransferEvent) error {
	if len(events) == 0 {
		return nil
	}
	return database.GetDB().Create(&events).Error
}
