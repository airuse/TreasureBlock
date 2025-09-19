package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

type TransferEventService interface {
	CreateEvents(events []*models.TransferEvent) error
}

type transferEventService struct {
	repo repository.TransferEventRepository
}

func NewTransferEventService(repo repository.TransferEventRepository) TransferEventService {
	return &transferEventService{repo: repo}
}

func (s *transferEventService) CreateEvents(events []*models.TransferEvent) error {
	return s.repo.CreateBatch(events)
}
