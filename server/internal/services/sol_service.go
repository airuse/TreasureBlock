package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

type SolService interface {
	SaveTxDetail(detail *models.SolTxDetail, instructions []*models.SolInstruction) error
}

type solService struct{ repo repository.SolRepository }

func NewSolService(repo repository.SolRepository) SolService { return &solService{repo: repo} }

func (s *solService) SaveTxDetail(detail *models.SolTxDetail, instructions []*models.SolInstruction) error {
	if err := s.repo.CreateTxDetail(detail); err != nil {
		return err
	}
	return s.repo.CreateInstructions(instructions)
}
