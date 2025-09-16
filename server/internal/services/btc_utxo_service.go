package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
)

type BTCUTXOService interface {
	UpsertOutputs(ctx context.Context, outputs []*models.BTCUTXO) error
	MarkSpent(ctx context.Context, chain string, prevTxID string, voutIndex uint32, spentTxID string, vinIndex uint32, spentHeight uint64) error
	GetOutputs(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error)
	GetSpent(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error)
	GetUTXOCountByAddress(ctx context.Context, chain string, address string) (int64, error)
	GetUTXOsByAddress(ctx context.Context, chain string, address string) ([]*models.BTCUTXO, error)
	GetUTXOsByAddressExcludingPending(ctx context.Context, chain string, address string) ([]*models.BTCUTXO, error)
}

type btcUtxoService struct {
	repo repository.BTCUTXORepository
}

func NewBTCUTXOService(repo repository.BTCUTXORepository) BTCUTXOService {
	return &btcUtxoService{repo: repo}
}

func (s *btcUtxoService) UpsertOutputs(ctx context.Context, outputs []*models.BTCUTXO) error {
	if len(outputs) == 0 {
		return nil
	}

	return s.repo.UpsertOutputs(ctx, outputs)
}

func (s *btcUtxoService) MarkSpent(ctx context.Context, chain string, prevTxID string, voutIndex uint32, spentTxID string, vinIndex uint32, spentHeight uint64) error {
	return s.repo.MarkSpent(ctx, chain, prevTxID, voutIndex, spentTxID, vinIndex, spentHeight)
}

func (s *btcUtxoService) GetOutputs(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error) {
	if txID == "" {
		return nil, fmt.Errorf("txID cannot be empty")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}

	return s.repo.GetOutputs(ctx, chain, txID)
}

func (s *btcUtxoService) GetSpent(ctx context.Context, chain string, txID string) ([]*models.BTCUTXO, error) {
	if txID == "" {
		return nil, fmt.Errorf("txID cannot be empty")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}
	return s.repo.GetSpent(ctx, chain, txID)
}

func (s *btcUtxoService) GetUTXOCountByAddress(ctx context.Context, chain string, address string) (int64, error) {
	if address == "" {
		return 0, fmt.Errorf("address cannot be empty")
	}
	if chain == "" {
		return 0, fmt.Errorf("chain cannot be empty")
	}
	return s.repo.GetUTXOCountByAddress(ctx, chain, address)
}

func (s *btcUtxoService) GetUTXOsByAddress(ctx context.Context, chain string, address string) ([]*models.BTCUTXO, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}
	return s.repo.GetUTXOsByAddress(ctx, chain, address)
}

func (s *btcUtxoService) GetUTXOsByAddressExcludingPending(ctx context.Context, chain string, address string) ([]*models.BTCUTXO, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}
	return s.repo.GetUTXOsByAddressExcludingPending(ctx, chain, address)
}
