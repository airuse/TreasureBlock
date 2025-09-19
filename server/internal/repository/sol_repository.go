package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
)

type SolRepository interface {
	CreateTxDetail(detail *models.SolTxDetail) error
	CreateInstructions(instructions []*models.SolInstruction) error
}

type solRepository struct{}

func NewSolRepository() SolRepository { return &solRepository{} }

func (r *solRepository) CreateTxDetail(detail *models.SolTxDetail) error {
	return database.GetDB().Create(detail).Error
}

func (r *solRepository) CreateInstructions(instructions []*models.SolInstruction) error {
	if len(instructions) == 0 {
		return nil
	}
	return database.GetDB().Create(&instructions).Error
}
