package repository

import (
	"blockChainBrowser/server/internal/models"
	"context"

	"gorm.io/gorm"
)

// SolInstructionRepository Solana指令仓库接口
type SolInstructionRepository interface {
	Create(ctx context.Context, instruction *models.SolInstruction) error
	CreateBatch(ctx context.Context, instructions []*models.SolInstruction) error
	GetByTxID(ctx context.Context, txID string) ([]*models.SolInstruction, error)
	GetBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*models.SolInstruction, int64, error)
	GetByProgramID(ctx context.Context, programID string, page, pageSize int) ([]*models.SolInstruction, int64, error)
	GetByInstructionType(ctx context.Context, instructionType string, page, pageSize int) ([]*models.SolInstruction, int64, error)
	GetOuterInstructions(ctx context.Context, txID string) ([]*models.SolInstruction, error)
	GetInnerInstructions(ctx context.Context, txID string) ([]*models.SolInstruction, error)
	List(ctx context.Context, page, pageSize int) ([]*models.SolInstruction, int64, error)
	Delete(ctx context.Context, id uint) error
	DeleteByTxID(ctx context.Context, txID string) error
	GetSlotRange(ctx context.Context, startSlot, endSlot uint64) ([]*models.SolInstruction, error)
}

// solInstructionRepository Solana指令仓库实现
type solInstructionRepository struct {
	db *gorm.DB
}

// NewSolInstructionRepository 创建Solana指令仓库
func NewSolInstructionRepository(db *gorm.DB) SolInstructionRepository {
	return &solInstructionRepository{db: db}
}

// Create 创建指令
func (r *solInstructionRepository) Create(ctx context.Context, instruction *models.SolInstruction) error {
	return r.db.WithContext(ctx).Create(instruction).Error
}

// CreateBatch 批量创建指令
func (r *solInstructionRepository) CreateBatch(ctx context.Context, instructions []*models.SolInstruction) error {
	if len(instructions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&instructions).Error
}

// GetByTxID 根据交易ID获取指令列表
func (r *solInstructionRepository) GetByTxID(ctx context.Context, txID string) ([]*models.SolInstruction, error) {
	var instructions []*models.SolInstruction
	err := r.db.WithContext(ctx).Where("tx_id = ?", txID).Order("instruction_index").Find(&instructions).Error
	return instructions, err
}

// GetBySlot 根据slot获取指令列表（分页）
func (r *solInstructionRepository) GetBySlot(ctx context.Context, slot uint64, page, pageSize int) ([]*models.SolInstruction, int64, error) {
	var instructions []*models.SolInstruction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolInstruction{}).Where("slot = ?", slot)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&instructions).Error
	return instructions, total, err
}

// GetByProgramID 根据程序ID获取指令列表（分页）
func (r *solInstructionRepository) GetByProgramID(ctx context.Context, programID string, page, pageSize int) ([]*models.SolInstruction, int64, error) {
	var instructions []*models.SolInstruction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolInstruction{}).Where("program_id = ?", programID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&instructions).Error
	return instructions, total, err
}

// GetByInstructionType 根据指令类型获取指令列表（分页）
func (r *solInstructionRepository) GetByInstructionType(ctx context.Context, instructionType string, page, pageSize int) ([]*models.SolInstruction, int64, error) {
	var instructions []*models.SolInstruction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.SolInstruction{}).Where("instruction_type = ?", instructionType)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&instructions).Error
	return instructions, total, err
}

// GetOuterInstructions 获取外层指令
func (r *solInstructionRepository) GetOuterInstructions(ctx context.Context, txID string) ([]*models.SolInstruction, error) {
	var instructions []*models.SolInstruction
	err := r.db.WithContext(ctx).
		Where("tx_id = ? AND is_inner = ?", txID, false).
		Order("instruction_index").
		Find(&instructions).Error
	return instructions, err
}

// GetInnerInstructions 获取内层指令
func (r *solInstructionRepository) GetInnerInstructions(ctx context.Context, txID string) ([]*models.SolInstruction, error) {
	var instructions []*models.SolInstruction
	err := r.db.WithContext(ctx).
		Where("tx_id = ? AND is_inner = ?", txID, true).
		Order("instruction_index").
		Find(&instructions).Error
	return instructions, err
}

// List 获取指令列表（分页）
func (r *solInstructionRepository) List(ctx context.Context, page, pageSize int) ([]*models.SolInstruction, int64, error) {
	var instructions []*models.SolInstruction
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.SolInstruction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("id DESC").Find(&instructions).Error
	return instructions, total, err
}

// Delete 删除指令
func (r *solInstructionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.SolInstruction{}, id).Error
}

// DeleteByTxID 根据交易ID删除指令
func (r *solInstructionRepository) DeleteByTxID(ctx context.Context, txID string) error {
	return r.db.WithContext(ctx).Where("tx_id = ?", txID).Delete(&models.SolInstruction{}).Error
}

// GetSlotRange 获取指定slot范围内的指令
func (r *solInstructionRepository) GetSlotRange(ctx context.Context, startSlot, endSlot uint64) ([]*models.SolInstruction, error) {
	var instructions []*models.SolInstruction
	err := r.db.WithContext(ctx).
		Where("slot >= ? AND slot <= ?", startSlot, endSlot).
		Order("slot ASC, instruction_index ASC").
		Find(&instructions).Error
	return instructions, err
}
