package services

import (
	"context"
	"fmt"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// TransactionService 交易服务接口
type TransactionService interface {
	GetTransactionByHash(ctx context.Context, hash string) (*models.Transaction, error)
	GetTransactionsByAddress(ctx context.Context, address string, page, pageSize int) ([]*models.Transaction, int64, error)
	GetTransactionsByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error)
	ListTransactions(ctx context.Context, page, pageSize int, chain string) ([]*models.Transaction, int64, error)
	CreateTransaction(ctx context.Context, tx *models.Transaction) error
	UpdateTransaction(ctx context.Context, tx *models.Transaction) error
	DeleteTransaction(ctx context.Context, hash string) error
}

// transactionService 交易服务实现
type transactionService struct {
	txRepo repository.TransactionRepository
}

// NewTransactionService 创建交易服务实例
func NewTransactionService(txRepo repository.TransactionRepository) TransactionService {
	return &transactionService{
		txRepo: txRepo,
	}
}

// GetTransactionByHash 根据哈希获取交易
func (s *transactionService) GetTransactionByHash(ctx context.Context, hash string) (*models.Transaction, error) {
	if hash == "" {
		return nil, fmt.Errorf("transaction hash cannot be empty")
	}

	tx, err := s.txRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash: %w", err)
	}

	return tx, nil
}

// GetTransactionsByAddress 根据地址获取交易列表
func (s *transactionService) GetTransactionsByAddress(ctx context.Context, address string, page, pageSize int) ([]*models.Transaction, int64, error) {
	if address == "" {
		return nil, 0, fmt.Errorf("address cannot be empty")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	txs, total, err := s.txRepo.GetByAddress(ctx, address, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions by address: %w", err)
	}

	return txs, total, nil
}

// GetTransactionsByBlockHash 根据区块哈希获取交易列表
func (s *transactionService) GetTransactionsByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error) {
	if blockHash == "" {
		return nil, fmt.Errorf("block hash cannot be empty")
	}

	txs, err := s.txRepo.GetByBlockHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by block hash: %w", err)
	}

	return txs, nil
}

// ListTransactions 分页查询交易列表
func (s *transactionService) ListTransactions(ctx context.Context, page, pageSize int, chain string) ([]*models.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	txs, total, err := s.txRepo.List(ctx, offset, pageSize, chain)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return txs, total, nil
}

// CreateTransaction 创建交易
func (s *transactionService) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if tx.TxID == "" {
		return fmt.Errorf("transaction hash cannot be empty")
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// UpdateTransaction 更新交易
func (s *transactionService) UpdateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if err := s.txRepo.Update(ctx, tx); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// DeleteTransaction 删除交易
func (s *transactionService) DeleteTransaction(ctx context.Context, hash string) error {
	if hash == "" {
		return fmt.Errorf("transaction hash cannot be empty")
	}

	if err := s.txRepo.Delete(ctx, hash); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}
