package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

// UserTransactionRepository 用户交易仓储接口
type UserTransactionRepository interface {
	Create(ctx context.Context, tx *models.UserTransaction) error
	GetByID(ctx context.Context, id uint, userID uint64) (*models.UserTransaction, error)
	GetByUserID(ctx context.Context, userID uint64, page, pageSize int, status string) ([]*models.UserTransaction, int64, error)
	Update(ctx context.Context, tx *models.UserTransaction) error
	Delete(ctx context.Context, id uint, userID uint64) error
	GetStatsByUserID(ctx context.Context, userID uint64) (*models.UserTransaction, error)
	GetByStatus(ctx context.Context, userID uint64, status string) ([]*models.UserTransaction, error)
	GetByChain(ctx context.Context, userID uint64, chain string, page, pageSize int) ([]*models.UserTransaction, int64, error)
	GetByChainExcludingPending(ctx context.Context, chain string) ([]*models.UserTransaction, error)
	GetByUserIDExcludingPending(ctx context.Context, userID uint) ([]*models.UserTransaction, error)
}

// userTransactionRepository 用户交易仓储实现
type userTransactionRepository struct {
	db *gorm.DB
}

// NewUserTransactionRepository 创建用户交易仓储实例
func NewUserTransactionRepository() UserTransactionRepository {
	return &userTransactionRepository{
		db: database.GetDB(),
	}
}

// Create 创建用户交易
func (r *userTransactionRepository) Create(ctx context.Context, tx *models.UserTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetByID 根据ID获取用户交易
func (r *userTransactionRepository) GetByID(ctx context.Context, id uint, userID uint64) (*models.UserTransaction, error) {
	var tx models.UserTransaction
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("交易不存在或无权限访问")
		}
		return nil, err
	}
	return &tx, nil
}

// GetByUserID 根据用户ID获取交易列表
func (r *userTransactionRepository) GetByUserID(ctx context.Context, userID uint64, page, pageSize int, status string) ([]*models.UserTransaction, int64, error) {
	var transactions []*models.UserTransaction
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// 如果指定了状态，添加状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Model(&models.UserTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按创建时间倒序
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// Update 更新用户交易
func (r *userTransactionRepository) Update(ctx context.Context, tx *models.UserTransaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

// Delete 删除用户交易
func (r *userTransactionRepository) Delete(ctx context.Context, id uint, userID uint64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.UserTransaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("交易不存在或无权限删除")
	}
	return nil
}

// GetStatsByUserID 获取用户交易统计
func (r *userTransactionRepository) GetStatsByUserID(ctx context.Context, userID uint64) (*models.UserTransaction, error) {
	// 这里返回一个空的统计对象，实际统计在Service层计算
	return &models.UserTransaction{}, nil
}

// GetByStatus 根据状态获取用户交易
func (r *userTransactionRepository) GetByStatus(ctx context.Context, userID uint64, status string) ([]*models.UserTransaction, error) {
	var txs []*models.UserTransaction
	q := r.db.WithContext(ctx).Model(&models.UserTransaction{}).Where("status = ?", status)
	if userID != 0 {
		q = q.Where("user_id = ?", userID)
	}
	err := q.Find(&txs).Error
	return txs, err
}

// GetByChain 根据链类型获取用户交易
func (r *userTransactionRepository) GetByChain(ctx context.Context, userID uint64, chain string, page, pageSize int) ([]*models.UserTransaction, int64, error) {
	var transactions []*models.UserTransaction
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ? AND chain = ?", userID, chain)

	// 获取总数
	if err := query.Model(&models.UserTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetByChainExcludingPending 根据链类型获取用户交易，排除被pending交易使用的
func (r *userTransactionRepository) GetByChainExcludingPending(ctx context.Context, chain string) ([]*models.UserTransaction, error) {
	var txs []*models.UserTransaction
	q := r.db.WithContext(ctx).Model(&models.UserTransaction{}).Where("chain = ? AND status IN (?, ?)", chain, "in_progress", "packed")
	err := q.Find(&txs).Error
	return txs, err
}

// GetByUserIDExcludingPending 根据用户ID获取用户交易，排除被pending交易使用的
func (r *userTransactionRepository) GetByUserIDExcludingPending(ctx context.Context, userID uint) ([]*models.UserTransaction, error) {
	var txs []*models.UserTransaction
	q := r.db.WithContext(ctx).Model(&models.UserTransaction{}).Where("user_id = ? AND status IN (?, ?)", userID, "in_progress", "packed")
	err := q.Find(&txs).Error
	return txs, err
}
