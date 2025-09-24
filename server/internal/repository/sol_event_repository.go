package repository

import (
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

// SolEventRepository Solana事件仓库接口
type SolEventRepository interface {
	Create(ctx context.Context, event *models.SolEvent) error
	CreateBatch(ctx context.Context, events []*models.SolEvent) error
	GetByTxID(ctx context.Context, txID string) ([]*models.SolEvent, error)
	GetBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*models.SolEvent, int64, error)
	GetByAddress(ctx context.Context, address string, page, pageSize int) ([]*models.SolEvent, int64, error)
	GetByProgramID(ctx context.Context, programID string, page, pageSize int) ([]*models.SolEvent, int64, error)
	GetByEventType(ctx context.Context, eventType string, page, pageSize int) ([]*models.SolEvent, int64, error)
	GetTransferEvents(ctx context.Context, fromAddr, toAddr string, page, pageSize int) ([]*models.SolEvent, int64, error)
	List(ctx context.Context, page, pageSize int) ([]*models.SolEvent, int64, error)
	Delete(ctx context.Context, id uint) error
	DeleteByTxID(ctx context.Context, txID string) error
	GetSlotRange(ctx context.Context, startSlot, endSlot uint64) ([]*models.SolEvent, error)
}

// solEventRepository Solana事件仓库实现
type solEventRepository struct {
	db *gorm.DB
}

// NewSolEventRepository 创建Solana事件仓库
func NewSolEventRepository(db *gorm.DB) SolEventRepository {
	return &solEventRepository{db: db}
}

// Create 创建事件
func (r *solEventRepository) Create(ctx context.Context, event *models.SolEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// CreateBatch 批量创建事件
func (r *solEventRepository) CreateBatch(ctx context.Context, events []*models.SolEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&events).Error
}

// GetByTxID 根据交易ID获取事件列表
func (r *solEventRepository) GetByTxID(ctx context.Context, txID string) ([]*models.SolEvent, error) {
	var events []*models.SolEvent
	err := r.db.WithContext(ctx).Where("tx_id = ?", txID).Order("event_index").Find(&events).Error
	return events, err
}

// GetBySlot 根据slot获取事件列表（分页）
func (r *solEventRepository) GetBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*models.SolEvent, int64, error) {
	var events []*models.SolEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolEvent{}).Where("slot = ?", slot)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&events).Error
	return events, total, err
}

// GetByAddress 根据地址获取事件列表（分页）
func (r *solEventRepository) GetByAddress(ctx context.Context, address string, page, pageSize int) ([]*models.SolEvent, int64, error) {
	var events []*models.SolEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolEvent{}).
		Where("from_address = ? OR to_address = ?", address, address)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&events).Error
	return events, total, err
}

// GetByProgramID 根据程序ID获取事件列表（分页）
func (r *solEventRepository) GetByProgramID(ctx context.Context, programID string, page, pageSize int) ([]*models.SolEvent, int64, error) {
	var events []*models.SolEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolEvent{}).Where("program_id = ?", programID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&events).Error
	return events, total, err
}

// GetByEventType 根据事件类型获取事件列表（分页）
func (r *solEventRepository) GetByEventType(ctx context.Context, eventType string, page, pageSize int) ([]*models.SolEvent, int64, error) {
	var events []*models.SolEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolEvent{}).Where("event_type = ?", eventType)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&events).Error
	return events, total, err
}

// GetTransferEvents 获取转账事件（分页）
func (r *solEventRepository) GetTransferEvents(ctx context.Context, fromAddr, toAddr string, page, pageSize int) ([]*models.SolEvent, int64, error) {
	var events []*models.SolEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolEvent{}).
		Where("event_type IN (?)", []string{"system_transfer", "spl_transfer"})

	// 如果指定了地址过滤
	if fromAddr != "" && toAddr != "" {
		query = query.Where("from_address = ? AND to_address = ?", fromAddr, toAddr)
	} else if fromAddr != "" {
		query = query.Where("from_address = ?", fromAddr)
	} else if toAddr != "" {
		query = query.Where("to_address = ?", toAddr)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&events).Error
	return events, total, err
}

// List 获取事件列表（分页）
func (r *solEventRepository) List(ctx context.Context, page, pageSize int) ([]*models.SolEvent, int64, error) {
	var events []*models.SolEvent
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.SolEvent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("id DESC").Find(&events).Error
	return events, total, err
}

// Delete 删除事件
func (r *solEventRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.SolEvent{}, id).Error
}

// DeleteByTxID 根据交易ID删除事件
func (r *solEventRepository) DeleteByTxID(ctx context.Context, txID string) error {
	return r.db.WithContext(ctx).Where("tx_id = ?", txID).Delete(&models.SolEvent{}).Error
}

// GetSlotRange 获取指定slot范围内的事件
func (r *solEventRepository) GetSlotRange(ctx context.Context, startSlot, endSlot uint64) ([]*models.SolEvent, error) {
	var events []*models.SolEvent
	err := r.db.WithContext(ctx).
		Where("slot >= ? AND slot <= ?", startSlot, endSlot).
		Order("slot ASC, event_index ASC").
		Find(&events).Error
	return events, err
}
