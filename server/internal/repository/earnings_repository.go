package repository

import (
	"context"
	"time"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// EarningsRepository 收益记录仓库接口
type EarningsRepository interface {
	// 创建收益记录
	CreateEarningsRecord(ctx context.Context, record *models.EarningsRecord) error
	// 批量创建收益记录
	CreateEarningsRecordsBatch(ctx context.Context, records []*models.EarningsRecord) error
	// 根据用户ID获取收益记录列表
	GetEarningsRecordsByUserID(ctx context.Context, userID uint64, page, pageSize int, filters map[string]interface{}) ([]*models.EarningsRecord, int64, error)
	// 根据ID获取收益记录
	GetEarningsRecordByID(ctx context.Context, id uint64) (*models.EarningsRecord, error)
	// 获取用户收益统计
	GetEarningsStatsByUserID(ctx context.Context, userID uint64) (*models.EarningsStats, error)
	// 获取用户指定时间范围内的收益记录
	GetEarningsRecordsByDateRange(ctx context.Context, userID uint64, startDate, endDate time.Time) ([]*models.EarningsRecord, error)
}

type earningsRepository struct {
	db *gorm.DB
}

// NewEarningsRepository 创建收益记录仓库
func NewEarningsRepository(db *gorm.DB) EarningsRepository {
	return &earningsRepository{db: db}
}

// CreateEarningsRecord 创建收益记录
func (r *earningsRepository) CreateEarningsRecord(ctx context.Context, record *models.EarningsRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// CreateEarningsRecordsBatch 批量创建收益记录
func (r *earningsRepository) CreateEarningsRecordsBatch(ctx context.Context, records []*models.EarningsRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

// GetEarningsRecordsByUserID 根据用户ID获取收益记录列表
func (r *earningsRepository) GetEarningsRecordsByUserID(ctx context.Context, userID uint64, page, pageSize int, filters map[string]interface{}) ([]*models.EarningsRecord, int64, error) {
	var records []*models.EarningsRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&models.EarningsRecord{}).Where("user_id = ?", userID)

	// 应用过滤条件
	if recordType, ok := filters["type"]; ok && recordType != "" {
		query = query.Where("type = ?", recordType)
	}
	if source, ok := filters["source"]; ok && source != "" {
		query = query.Where("source = ?", source)
	}
	if chain, ok := filters["chain"]; ok && chain != "" {
		query = query.Where("source_chain = ?", chain)
	}
	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetEarningsRecordByID 根据ID获取收益记录
func (r *earningsRepository) GetEarningsRecordByID(ctx context.Context, id uint64) (*models.EarningsRecord, error) {
	var record models.EarningsRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetEarningsStatsByUserID 获取用户收益统计
func (r *earningsRepository) GetEarningsStatsByUserID(ctx context.Context, userID uint64) (*models.EarningsStats, error) {
	var stats models.EarningsStats

	// 使用原生SQL查询获取统计数据
	query := `
		SELECT 
			? as user_id,
			COALESCE(SUM(CASE WHEN type = 'add' THEN amount ELSE 0 END), 0) as total_earnings,
			COALESCE(SUM(CASE WHEN type = 'decrease' THEN amount ELSE 0 END), 0) as total_spendings,
			COALESCE(SUM(CASE WHEN type = 'add' THEN amount ELSE -amount END), 0) as current_balance,
			COALESCE(COUNT(DISTINCT CASE WHEN type = 'add' AND source = 'block_verification' THEN source_id END), 0) as block_count,
			COALESCE(SUM(CASE WHEN type = 'add' AND source = 'block_verification' THEN transaction_count ELSE 0 END), 0) as transaction_count
		FROM earnings_records 
		WHERE user_id = ? AND deleted_at IS NULL
	`

	err := r.db.WithContext(ctx).Raw(query, userID, userID).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetEarningsRecordsByDateRange 获取用户指定时间范围内的收益记录
func (r *earningsRepository) GetEarningsRecordsByDateRange(ctx context.Context, userID uint64, startDate, endDate time.Time) ([]*models.EarningsRecord, error) {
	var records []*models.EarningsRecord

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	return records, nil
}
