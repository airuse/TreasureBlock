package repository

import (
	"context"
	"fmt"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// TransactionReceiptRepository 交易凭证仓储接口
type TransactionReceiptRepository interface {
	Create(ctx context.Context, receipt *models.TransactionReceipt) error
	GetByTxHash(ctx context.Context, txHash string) (*models.TransactionReceipt, error)
	GetByBlockHash(ctx context.Context, blockHash string) ([]*models.TransactionReceipt, error)
	GetByBlockNumber(ctx context.Context, blockNumber uint64, chain string) ([]*models.TransactionReceipt, error)
	Update(ctx context.Context, receipt *models.TransactionReceipt) error
	Delete(ctx context.Context, id uint) error
	Exists(ctx context.Context, txHash string) (bool, error)
}

// transactionReceiptRepository 交易凭证仓储实现
type transactionReceiptRepository struct {
	db *gorm.DB
}

// NewTransactionReceiptRepository 创建新的交易凭证仓储
func NewTransactionReceiptRepository(db *gorm.DB) TransactionReceiptRepository {
	return &transactionReceiptRepository{
		db: db,
	}
}

// Create 创建交易凭证
func (r *transactionReceiptRepository) Create(ctx context.Context, receipt *models.TransactionReceipt) error {
	if err := r.db.WithContext(ctx).Create(receipt).Error; err != nil {
		return fmt.Errorf("failed to create transaction receipt: %w", err)
	}
	return nil
}

// GetByTxHash 根据交易哈希获取凭证
func (r *transactionReceiptRepository) GetByTxHash(ctx context.Context, txHash string) (*models.TransactionReceipt, error) {
	var receipt models.TransactionReceipt
	if err := r.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&receipt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction receipt not found for hash: %s", txHash)
		}
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	return &receipt, nil
}

// GetByBlockHash 根据区块哈希获取所有凭证
func (r *transactionReceiptRepository) GetByBlockHash(ctx context.Context, blockHash string) ([]*models.TransactionReceipt, error) {
	var receipts []*models.TransactionReceipt
	if err := r.db.WithContext(ctx).Where("block_hash = ?", blockHash).Find(&receipts).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction receipts by block hash: %w", err)
	}
	return receipts, nil
}

// GetByBlockNumber 根据区块号获取所有凭证
func (r *transactionReceiptRepository) GetByBlockNumber(ctx context.Context, blockNumber uint64, chain string) ([]*models.TransactionReceipt, error) {
	var receipts []*models.TransactionReceipt
	query := r.db.WithContext(ctx).Where("block_number = ?", blockNumber)
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}
	if err := query.Find(&receipts).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction receipts by block number: %w", err)
	}
	return receipts, nil
}

// Update 更新交易凭证
func (r *transactionReceiptRepository) Update(ctx context.Context, receipt *models.TransactionReceipt) error {
	if err := r.db.WithContext(ctx).Save(receipt).Error; err != nil {
		return fmt.Errorf("failed to update transaction receipt: %w", err)
	}
	return nil
}

// Delete 删除交易凭证
func (r *transactionReceiptRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.TransactionReceipt{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete transaction receipt: %w", err)
	}
	return nil
}

// Exists 检查交易凭证是否存在
func (r *transactionReceiptRepository) Exists(ctx context.Context, txHash string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.TransactionReceipt{}).Where("tx_hash = ?", txHash).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check transaction receipt existence: %w", err)
	}
	return count > 0, nil
}
