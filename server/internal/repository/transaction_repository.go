package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

// TransactionRepository 交易仓储接口
type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetByHash(ctx context.Context, hash string) (*models.Transaction, error)
	GetByAddress(ctx context.Context, address string, offset, limit int) ([]*models.Transaction, int64, error)
	GetByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error)
	List(ctx context.Context, offset, limit int, chain string) ([]*models.Transaction, int64, error)
	Update(ctx context.Context, tx *models.Transaction) error
	Delete(ctx context.Context, hash string) error
}

// transactionRepository 交易仓储实现
type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository 创建交易仓储实例
func NewTransactionRepository() TransactionRepository {
	return &transactionRepository{
		db: database.GetDB(),
	}
}

// Create 创建交易
func (r *transactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetByHash 根据哈希获取交易
func (r *transactionRepository) GetByHash(ctx context.Context, hash string) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetByAddress 根据地址获取交易列表
func (r *transactionRepository) GetByAddress(ctx context.Context, address string, offset, limit int) ([]*models.Transaction, int64, error) {
	var txs []*models.Transaction
	var total int64

	query := r.db.WithContext(ctx).Where("from_address = ? OR to_address = ?", address, address)

	// 获取总数
	if err := query.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&txs).Error

	return txs, total, err
}

// GetByBlockHash 根据区块哈希获取交易列表
func (r *transactionRepository) GetByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error) {
	var txs []*models.Transaction
	err := r.db.WithContext(ctx).
		Where("block_hash = ?", blockHash).
		Order("timestamp ASC").
		Find(&txs).Error
	return txs, err
}

// List 分页查询交易列表
func (r *transactionRepository) List(ctx context.Context, offset, limit int, chain string) ([]*models.Transaction, int64, error) {
	var txs []*models.Transaction
	var total int64

	query := r.db.WithContext(ctx)
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}

	// 获取总数
	if err := query.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&txs).Error

	return txs, total, err
}

// Update 更新交易
func (r *transactionRepository) Update(ctx context.Context, tx *models.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

// Delete 删除交易
func (r *transactionRepository) Delete(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Where("hash = ?", hash).Delete(&models.Transaction{}).Error
}
