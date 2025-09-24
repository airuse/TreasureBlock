package repository

import (
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

// SolTxDetailRepository Solana交易明细仓库接口
type SolTxDetailRepository interface {
	Create(ctx context.Context, detail *models.SolTxDetail) error
	CreateBatch(ctx context.Context, details []*models.SolTxDetail) error
	GetByTxID(ctx context.Context, txID string) (*models.SolTxDetail, error)
	GetBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*models.SolTxDetail, int64, error)
	GetByBlockID(ctx context.Context, blockID uint64) ([]*models.SolTxDetail, error)
	Update(ctx context.Context, detail *models.SolTxDetail) error
	Delete(ctx context.Context, txID string) error
	List(ctx context.Context, page, pageSize int) ([]*models.SolTxDetail, int64, error)
	GetByStatus(ctx context.Context, status string, page, pageSize int) ([]*models.SolTxDetail, int64, error)
	GetSlotRange(ctx context.Context, startSlot, endSlot uint64) ([]*models.SolTxDetail, error)
}

// solTxDetailRepository Solana交易明细仓库实现
type solTxDetailRepository struct {
	db *gorm.DB
}

// NewSolTxDetailRepository 创建Solana交易明细仓库
func NewSolTxDetailRepository(db *gorm.DB) SolTxDetailRepository {
	return &solTxDetailRepository{db: db}
}

// Create 创建交易明细
func (r *solTxDetailRepository) Create(ctx context.Context, detail *models.SolTxDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

// CreateBatch 批量创建交易明细
func (r *solTxDetailRepository) CreateBatch(ctx context.Context, details []*models.SolTxDetail) error {
	if len(details) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&details).Error
}

// GetByTxID 根据交易ID获取明细
func (r *solTxDetailRepository) GetByTxID(ctx context.Context, txID string) (*models.SolTxDetail, error) {
	var detail models.SolTxDetail
	err := r.db.WithContext(ctx).Where("tx_id = ?", txID).First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// GetBySlot 根据slot获取交易列表（分页）
func (r *solTxDetailRepository) GetBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*models.SolTxDetail, int64, error) {
	var details []*models.SolTxDetail
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolTxDetail{}).Where("slot = ?", slot)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&details).Error
	return details, total, err
}

// GetByBlockID 根据区块ID获取交易列表
func (r *solTxDetailRepository) GetByBlockID(ctx context.Context, blockID uint64) ([]*models.SolTxDetail, error) {
	var details []*models.SolTxDetail
	err := r.db.WithContext(ctx).Where("block_id = ?", blockID).Find(&details).Error
	return details, err
}

// Update 更新交易明细
func (r *solTxDetailRepository) Update(ctx context.Context, detail *models.SolTxDetail) error {
	return r.db.WithContext(ctx).Save(detail).Error
}

// Delete 删除交易明细
func (r *solTxDetailRepository) Delete(ctx context.Context, txID string) error {
	return r.db.WithContext(ctx).Where("tx_id = ?", txID).Delete(&models.SolTxDetail{}).Error
}

// List 获取交易明细列表（分页）
func (r *solTxDetailRepository) List(ctx context.Context, page, pageSize int) ([]*models.SolTxDetail, int64, error) {
	var details []*models.SolTxDetail
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.SolTxDetail{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("id DESC").Find(&details).Error
	return details, total, err
}

// GetByStatus 根据状态获取交易列表（分页）
func (r *solTxDetailRepository) GetByStatus(ctx context.Context, status string, page, pageSize int) ([]*models.SolTxDetail, int64, error) {
	var details []*models.SolTxDetail
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolTxDetail{}).Where("status = ?", status)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&details).Error
	return details, total, err
}

// GetSlotRange 获取指定slot范围内的交易
func (r *solTxDetailRepository) GetSlotRange(ctx context.Context, startSlot, endSlot uint64) ([]*models.SolTxDetail, error) {
	var details []*models.SolTxDetail
	err := r.db.WithContext(ctx).
		Where("slot >= ? AND slot <= ?", startSlot, endSlot).
		Order("slot ASC, id ASC").
		Find(&details).Error
	return details, err
}
