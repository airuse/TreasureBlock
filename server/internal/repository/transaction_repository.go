package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"
	"sort"

	"gorm.io/gorm"
)

// TransactionRepository 交易仓储接口
type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	CreateBatch(ctx context.Context, txs []*models.Transaction) error
	GetByHash(ctx context.Context, hash string) (*models.Transaction, error)
	GetByAddress(ctx context.Context, address string, offset, limit int) ([]*models.Transaction, int64, error)
	GetByBlockHash(ctx context.Context, blockHash string) ([]*models.Transaction, error)
	GetByBlockHeight(ctx context.Context, blockHeight uint64, offset, limit int, chain string) ([]*models.Transaction, int64, error)
	GetByBlockID(ctx context.Context, blockID uint64) ([]*models.Transaction, error)
	List(ctx context.Context, offset, limit int, chain string) ([]*models.Transaction, int64, error)
	ListBTCTransactions(ctx context.Context, offset, limit int, chain string) ([]*models.Transaction, int64, error)
	GetLatestBlockHeight(ctx context.Context, chain string) (uint64, error)
	GetLatestTransactionsByBlockIndex(ctx context.Context, chain string, blockHeight uint64, limit int) ([]*models.Transaction, error)
	Update(ctx context.Context, tx *models.Transaction) error
	Delete(ctx context.Context, hash string) error
	LogicalDeleteByBlockID(ctx context.Context, blockID uint64) error
	// ComputeEthSumsWei 统计自指定高度以来的ETH转入/转出总额（Wei）
	ComputeEthSumsWei(ctx context.Context, address string, fromHeight uint64) (string, string, error)
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

// CreateBatch 批量创建交易
func (r *transactionRepository) CreateBatch(ctx context.Context, txs []*models.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(txs, 1000).Error
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

	// 使用 UNION 查询替代 OR 条件，提高查询性能
	// 分别查询作为发送方和接收方的交易，然后合并
	var txsFrom []*models.Transaction
	var txsTo []*models.Transaction

	// 查询作为发送方的交易
	if err := r.db.WithContext(ctx).
		Where("address_from = ?", address).
		Order("height DESC").
		Offset(offset).
		Limit(limit).
		Find(&txsFrom).Error; err != nil {
		return nil, 0, err
	}

	// 查询作为接收方的交易
	if err := r.db.WithContext(ctx).
		Where("address_to = ?", address).
		Order("height DESC").
		Offset(offset).
		Limit(limit).
		Find(&txsTo).Error; err != nil {
		return nil, 0, err
	}

	// 合并结果并按高度排序
	txs = append(txsFrom, txsTo...)

	// 按高度降序排序
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Height > txs[j].Height
	})

	// 获取总数 - 分别统计然后相加
	var totalFrom, totalTo int64
	if err := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("address_from = ?", address).Count(&totalFrom).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.WithContext(ctx).Model(&models.Transaction{}).Where("address_to = ?", address).Count(&totalTo).Error; err != nil {
		return nil, 0, err
	}
	total = totalFrom + totalTo

	return txs, total, nil
}

// ComputeEthSumsWei 统计自指定高度以来该地址的ETH转入/转出总额（Wei）
func (r *transactionRepository) ComputeEthSumsWei(ctx context.Context, address string, fromHeight uint64) (string, string, error) {
	type sumRow struct{ Sum string }

	var inSum sumRow
	var outSum sumRow

	// 仅统计 ETH 主币（symbol = 'ETH' 或空代表主币，视表字段定义而定）且成功的交易
	// 转入：address_to = address
	if err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), '0') as sum").
		Where("chain = ? AND symbol = ? AND status = 1 AND height > ? AND address_to = ?", "eth", "ETH", fromHeight, address).
		Scan(&inSum).Error; err != nil {
		return "0", "0", err
	}

	// 转出：address_from = address
	if err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), '0') as sum").
		Where("chain = ? AND symbol = ? AND status = 1 AND height > ? AND address_from = ?", "eth", "ETH", fromHeight, address).
		Scan(&outSum).Error; err != nil {
		return "0", "0", err
	}

	return inSum.Sum, outSum.Sum, nil
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

// ListBTCTransactions 分页查询BTC交易列表
func (r *transactionRepository) ListBTCTransactions(ctx context.Context, offset, limit int, chain string) ([]*models.Transaction, int64, error) {
	var txs []*models.Transaction
	var total int64

	query := r.db.WithContext(ctx)
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}

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
