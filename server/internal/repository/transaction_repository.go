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
	GetByBlockHeight(ctx context.Context, blockHeight uint64, offset, limit int, chain string) ([]*models.Transaction, int64, error)
	GetByBlockID(ctx context.Context, blockID uint64) ([]*models.Transaction, error)
	List(ctx context.Context, offset, limit int, chain string) ([]*models.Transaction, int64, error)
	GetLatestBlockHeight(ctx context.Context, chain string) (uint64, error)
	GetLatestTransactionsByBlockIndex(ctx context.Context, chain string, blockHeight uint64, limit int) ([]*models.Transaction, error)
	Update(ctx context.Context, tx *models.Transaction) error
	Delete(ctx context.Context, hash string) error
	LogicalDeleteByBlockID(ctx context.Context, blockID uint64) error
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
	err := r.db.WithContext(ctx).Where("`tx_id` = ?", hash).First(&tx).Error
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
	err := query.Order("block_index DESC").
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
		Order("block_index ASC").
		Find(&txs).Error
	return txs, err
}

// GetByBlockHeight 根据区块高度获取交易列表
func (r *transactionRepository) GetByBlockHeight(ctx context.Context, blockHeight uint64, offset, limit int, chain string) ([]*models.Transaction, int64, error) {
	var txs []*models.Transaction
	var total int64

	query := r.db.WithContext(ctx).Where("height = ?", blockHeight)
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}

	// 获取总数
	if err := query.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("block_index ASC").
		Offset(offset).
		Limit(limit).
		Find(&txs).Error

	return txs, total, err
}

// GetByBlockID 根据区块ID获取交易列表
func (r *transactionRepository) GetByBlockID(ctx context.Context, blockID uint64) ([]*models.Transaction, error) {
	var txs []*models.Transaction
	err := r.db.WithContext(ctx).
		Where("block_id = ?", blockID).
		Order("block_index ASC").
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
	err := query.Order("block_index DESC").
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

// LogicalDeleteByBlockID 根据区块ID逻辑删除所有交易
func (r *transactionRepository) LogicalDeleteByBlockID(ctx context.Context, blockID uint64) error {
	// 直接使用 GORM 的 Delete 方法进行软删除
	result := r.db.WithContext(ctx).Where("block_id = ?", blockID).Delete(&models.Transaction{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetLatestBlockHeight 获取最新区块高度
func (r *transactionRepository) GetLatestBlockHeight(ctx context.Context, chain string) (uint64, error) {
	var block models.Block
	err := r.db.WithContext(ctx).
		Where("chain = ? AND is_verified = 1", chain).
		Order("height DESC").
		Limit(1).
		Select("height").
		First(&block).Error

	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

// GetLatestTransactionsByBlockIndex 获取指定区块高度的前几条交易
func (r *transactionRepository) GetLatestTransactionsByBlockIndex(ctx context.Context, chain string, blockHeight uint64, limit int) ([]*models.Transaction, error) {
	var txs []*models.Transaction

	// 使用GORM查询，配合索引提升性能
	err := r.db.WithContext(ctx).
		Where("chain = ? AND height = ?", chain, blockHeight).
		Order("block_index DESC").
		Limit(limit).
		Find(&txs).Error

	return txs, err
}
