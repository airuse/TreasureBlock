package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

// BlockRepository 区块仓储接口
type BlockRepository interface {
	Create(ctx context.Context, block *models.Block) error
	GetByHash(ctx context.Context, hash string) (*models.Block, error)
	GetByHeight(ctx context.Context, height uint64) (*models.Block, error)
	GetLatest(ctx context.Context, chain string) (*models.Block, error)
	List(ctx context.Context, offset, limit int, chain string) ([]*models.Block, int64, error)
	Update(ctx context.Context, block *models.Block) error
	Delete(ctx context.Context, hash string) error
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

// Delete 删除区块
func (r *blockRepository) Delete(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Where("hash = ?", hash).Delete(&models.Block{}).Error
}
