package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

// BlockRepository 区块仓储接口
type BlockRepository interface {
	Create(ctx context.Context, block *models.Block) error
	GetByHash(ctx context.Context, hash string) (*models.Block, error)
	GetByHeight(ctx context.Context, height uint64) (*models.Block, error)
	GetByID(ctx context.Context, id uint64) (*models.Block, error)
	GetLatest(ctx context.Context, chain string) (*models.Block, error)
	GetLastVerifiedBlock(ctx context.Context, chain string) (*models.Block, error)
	List(ctx context.Context, offset, limit int, chain string) ([]*models.Block, int64, error)
	Update(ctx context.Context, block *models.Block) error
	UpdateFieldsByHash(ctx context.Context, hash string, fields map[string]interface{}) error
	UpdateVerificationStatus(ctx context.Context, blockID uint64, status uint8, reason string) error
	Delete(ctx context.Context, hash string) error
	GetTimeoutBlocks(ctx context.Context, chain string, height uint64) ([]*models.Block, error)
	UpdateBlockHash(ctx context.Context, blockID uint64, newHash string) error
	LogicalDelete(ctx context.Context, blockID uint64) error
	UpdateVerificationDeadline(ctx context.Context, blockID uint64, deadline *time.Time) error
	GetFailedBlockCountByHeight(ctx context.Context, height uint64, chain string) (int, error)
}

// blockRepository 区块仓储实现
type blockRepository struct {
	db *gorm.DB
}

// NewBlockRepository 创建区块仓储实例
func NewBlockRepository() BlockRepository {
	return &blockRepository{
		db: database.GetDB(),
	}
}

// Create 创建区块
func (r *blockRepository) Create(ctx context.Context, block *models.Block) error {
	return r.db.WithContext(ctx).Create(block).Error
}

// GetByHash 根据哈希获取区块
func (r *blockRepository) GetByHash(ctx context.Context, hash string) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetByHeight 根据高度获取区块
func (r *blockRepository) GetByHeight(ctx context.Context, height uint64) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).Where("height = ?", height).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetByID 根据ID获取区块
func (r *blockRepository) GetByID(ctx context.Context, id uint64) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetLatest 获取最新区块
func (r *blockRepository) GetLatest(ctx context.Context, chain string) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).
		Where("chain = ?", chain).
		Order("height DESC").
		First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetLastVerifiedBlock 获取最后一个验证通过的区块
func (r *blockRepository) GetLastVerifiedBlock(ctx context.Context, chain string) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).
		Where("chain = ? AND is_verified = ?", chain, 1).
		Order("height DESC").
		First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// List 分页查询区块列表
func (r *blockRepository) List(ctx context.Context, offset, limit int, chain string) ([]*models.Block, int64, error) {
	var blocks []*models.Block
	var total int64

	query := r.db.WithContext(ctx)
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}

	// 获取总数
	if err := query.Model(&models.Block{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("height DESC").
		Offset(offset).
		Limit(limit).
		Find(&blocks).Error

	return blocks, total, err
}

// Update 更新区块
func (r *blockRepository) Update(ctx context.Context, block *models.Block) error {
	return r.db.WithContext(ctx).Save(block).Error
}

// UpdateFieldsByHash 按哈希更新指定字段（局部更新，避免零值覆盖）
func (r *blockRepository) UpdateFieldsByHash(ctx context.Context, hash string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	fields["updated_at"] = gorm.Expr("NOW()")
	return r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("hash = ?", hash).
		Updates(fields).Error
}

// UpdateVerificationStatus 更新区块验证状态
func (r *blockRepository) UpdateVerificationStatus(ctx context.Context, blockID uint64, status uint8, reason string) error {
	return r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("id = ?", blockID).
		Updates(map[string]interface{}{
			"is_verified": status,
			"updated_at":  gorm.Expr("NOW()"),
		}).Error
}

// Delete 删除区块
func (r *blockRepository) Delete(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Where("hash = ?", hash).Delete(&models.Block{}).Error
}

// GetTimeoutBlocks 获取超时的区块（未验证且超过验证截止时间）
func (r *blockRepository) GetTimeoutBlocks(ctx context.Context, chain string, height uint64) ([]*models.Block, error) {
	var blocks []*models.Block
	err := r.db.WithContext(ctx).
		Where("chain = ? AND is_verified != ? AND verification_deadline < NOW() AND height = ?", chain, 1, height).
		Find(&blocks).Error
	return blocks, err
}

// UpdateBlockHash 更新区块哈希
func (r *blockRepository) UpdateBlockHash(ctx context.Context, blockID uint64, newHash string) error {

	return r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("id = ?", blockID).
		Update("hash", newHash).Error
}

// LogicalDelete 逻辑删除区块（设置deleted_at）
func (r *blockRepository) LogicalDelete(ctx context.Context, blockID uint64) error {
	return r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("id = ?", blockID).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// UpdateVerificationDeadline 更新区块验证截止时间
func (r *blockRepository) UpdateVerificationDeadline(ctx context.Context, blockID uint64, deadline *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("id = ?", blockID).
		Update("verification_deadline", deadline).Error
}

// GetFailedBlockCountByHeight 获取指定高度下验证失败的区块数量（包括逻辑删除的）
func (r *blockRepository) GetFailedBlockCountByHeight(ctx context.Context, height uint64, chain string) (int, error) {
	var count int64
	// 使用原生SQL绕过GORM的软删除过滤，统计所有失败区块
	err := r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM blocks WHERE height = ? AND chain = ? AND is_verified != ?", height, chain, 1).
		Scan(&count).Error
	return int(count), err
}
