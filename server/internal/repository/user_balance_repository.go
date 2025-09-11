package repository

import (
	"context"
	"errors"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// UserBalanceRepository 用户余额仓库接口
type UserBalanceRepository interface {
	// 获取用户余额
	GetUserBalance(ctx context.Context, userID uint64) (*models.UserBalance, error)
	// 创建用户余额记录
	CreateUserBalance(ctx context.Context, balance *models.UserBalance) error
	// 更新用户余额（使用乐观锁）
	UpdateUserBalance(ctx context.Context, balance *models.UserBalance) error
	// 增加用户余额（原子操作）
	IncrementUserBalance(ctx context.Context, userID uint64, amount int64, updateEarned bool) (*models.UserBalance, error)
	// 减少用户余额（原子操作）
	DecrementUserBalance(ctx context.Context, userID uint64, amount int64, updateSpent bool) (*models.UserBalance, error)
	// 获取用户余额列表（管理用）
	GetUserBalancesList(ctx context.Context, page, pageSize int) ([]*models.UserBalance, int64, error)
	// 检查用户余额是否足够
	CheckSufficientBalance(ctx context.Context, userID uint64, amount int64) (bool, error)

	// 链感知方法（不破坏现有调用）
	GetUserBalanceByChain(ctx context.Context, userID uint64, sourceChain string) (*models.UserBalance, error)
	IncrementUserBalanceByChain(ctx context.Context, userID uint64, amount int64, updateEarned bool, sourceChain string) (*models.UserBalance, error)
	DecrementUserBalanceByChain(ctx context.Context, userID uint64, amount int64, updateSpent bool, sourceChain string) (*models.UserBalance, error)
	CheckSufficientBalanceByChain(ctx context.Context, userID uint64, amount int64, sourceChain string) (bool, error)
}

type userBalanceRepository struct {
	db *gorm.DB
}

// NewUserBalanceRepository 创建用户余额仓库
func NewUserBalanceRepository(db *gorm.DB) UserBalanceRepository {
	return &userBalanceRepository{db: db}
}

// GetUserBalance 获取用户余额
func (r *userBalanceRepository) GetUserBalance(ctx context.Context, userID uint64) (*models.UserBalance, error) {
	var balance models.UserBalance
	err := r.db.WithContext(ctx).Where("user_id = ? AND source_chain = ?", userID, "all").First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果记录不存在，创建一个新的余额记录
			newBalance := &models.UserBalance{
				UserID:      userID,
				SourceChain: "all",
				Balance:     0,
				TotalEarned: 0,
				TotalSpent:  0,
				Version:     0,
			}
			createErr := r.db.WithContext(ctx).Create(newBalance).Error
			if createErr != nil {
				return nil, createErr
			}
			return newBalance, nil
		}
		return nil, err
	}
	return &balance, nil
}

// CreateUserBalance 创建用户余额记录
func (r *userBalanceRepository) CreateUserBalance(ctx context.Context, balance *models.UserBalance) error {
	return r.db.WithContext(ctx).Create(balance).Error
}

// UpdateUserBalance 更新用户余额（使用乐观锁）
func (r *userBalanceRepository) UpdateUserBalance(ctx context.Context, balance *models.UserBalance) error {
	// 使用乐观锁更新
	result := r.db.WithContext(ctx).
		Model(balance).
		Where("id = ? AND version = ?", balance.ID, balance.Version).
		Updates(map[string]interface{}{
			"balance":           balance.Balance,
			"total_earned":      balance.TotalEarned,
			"total_spent":       balance.TotalSpent,
			"last_earning_time": balance.LastEarningTime,
			"last_spend_time":   balance.LastSpendTime,
			"version":           gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("optimistic lock failed: balance was modified by another transaction")
	}

	// 更新版本号
	balance.Version++
	return nil
}

// IncrementUserBalance 增加用户余额（原子操作）
func (r *userBalanceRepository) IncrementUserBalance(ctx context.Context, userID uint64, amount int64, updateEarned bool) (*models.UserBalance, error) {
	var balance models.UserBalance

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取或创建用户余额记录
		err := tx.Where("user_id = ? AND source_chain = ?", userID, "all").First(&balance).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新记录
				balance = models.UserBalance{
					UserID:      userID,
					SourceChain: "all",
					Balance:     amount,
					TotalEarned: 0,
					TotalSpent:  0,
					Version:     0,
				}
				if updateEarned {
					balance.TotalEarned = amount
				}
				return tx.Create(&balance).Error
			}
			return err
		}

		// 更新余额
		updates := map[string]interface{}{
			"balance": gorm.Expr("balance + ?", amount),
			"version": gorm.Expr("version + 1"),
		}
		if updateEarned {
			updates["total_earned"] = gorm.Expr("total_earned + ?", amount)
		}

		result := tx.Model(&balance).Where("id = ? AND version = ?", balance.ID, balance.Version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("optimistic lock failed: balance was modified by another transaction")
		}

		// 重新获取更新后的记录
		return tx.Where("user_id = ? AND source_chain = ?", userID, "all").First(&balance).Error
	})

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// DecrementUserBalance 减少用户余额（原子操作）
func (r *userBalanceRepository) DecrementUserBalance(ctx context.Context, userID uint64, amount int64, updateSpent bool) (*models.UserBalance, error) {
	var balance models.UserBalance

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取用户余额记录
		err := tx.Where("user_id = ? AND source_chain = ?", userID, "all").First(&balance).Error
		if err != nil {
			return err
		}

		// 检查余额是否足够
		if balance.Balance < amount {
			return errors.New("insufficient balance")
		}

		// 更新余额
		updates := map[string]interface{}{
			"balance": gorm.Expr("balance - ?", amount),
			"version": gorm.Expr("version + 1"),
		}
		if updateSpent {
			updates["total_spent"] = gorm.Expr("total_spent + ?", amount)
		}

		result := tx.Model(&balance).Where("id = ? AND version = ?", balance.ID, balance.Version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("optimistic lock failed: balance was modified by another transaction")
		}

		// 重新获取更新后的记录
		return tx.Where("user_id = ? AND source_chain = ?", userID, "all").First(&balance).Error
	})

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// GetUserBalancesList 获取用户余额列表（管理用）
func (r *userBalanceRepository) GetUserBalancesList(ctx context.Context, page, pageSize int) ([]*models.UserBalance, int64, error) {
	var balances []*models.UserBalance
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.UserBalance{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("balance DESC").Offset(offset).Limit(pageSize).Find(&balances).Error; err != nil {
		return nil, 0, err
	}

	return balances, total, nil
}

// CheckSufficientBalance 检查用户余额是否足够
func (r *userBalanceRepository) CheckSufficientBalance(ctx context.Context, userID uint64, amount int64) (bool, error) {
	var balance models.UserBalance
	err := r.db.WithContext(ctx).Select("balance").Where("user_id = ? AND source_chain = ?", userID, "all").First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // 用户没有余额记录，余额为0
		}
		return false, err
	}

	return balance.Balance >= amount, nil
}

// GetUserBalanceByChain 获取指定链的用户余额
func (r *userBalanceRepository) GetUserBalanceByChain(ctx context.Context, userID uint64, sourceChain string) (*models.UserBalance, error) {
	var balance models.UserBalance
	err := r.db.WithContext(ctx).Where("user_id = ? AND source_chain = ?", userID, sourceChain).First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newBalance := &models.UserBalance{
				UserID:      userID,
				SourceChain: sourceChain,
				Balance:     0,
				TotalEarned: 0,
				TotalSpent:  0,
				Version:     0,
			}
			createErr := r.db.WithContext(ctx).Create(newBalance).Error
			if createErr != nil {
				return nil, createErr
			}
			return newBalance, nil
		}
		return nil, err
	}
	return &balance, nil
}

// IncrementUserBalanceByChain 增加指定链的用户余额（原子操作）
func (r *userBalanceRepository) IncrementUserBalanceByChain(ctx context.Context, userID uint64, amount int64, updateEarned bool, sourceChain string) (*models.UserBalance, error) {
	var balance models.UserBalance

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取或创建用户余额记录
		err := tx.Where("user_id = ? AND source_chain = ?", userID, sourceChain).First(&balance).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新记录
				balance = models.UserBalance{
					UserID:      userID,
					SourceChain: sourceChain,
					Balance:     amount,
					TotalEarned: 0,
					TotalSpent:  0,
					Version:     0,
				}
				if updateEarned {
					balance.TotalEarned = amount
				}
				return tx.Create(&balance).Error
			}
			return err
		}

		// 更新余额
		updates := map[string]interface{}{
			"balance": gorm.Expr("balance + ?", amount),
			"version": gorm.Expr("version + 1"),
		}
		if updateEarned {
			updates["total_earned"] = gorm.Expr("total_earned + ?", amount)
		}

		result := tx.Model(&balance).Where("id = ? AND version = ?", balance.ID, balance.Version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("optimistic lock failed: balance was modified by another transaction")
		}

		// 重新获取更新后的记录
		return tx.Where("user_id = ? AND source_chain = ?", userID, sourceChain).First(&balance).Error
	})

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// DecrementUserBalanceByChain 减少指定链的用户余额（原子操作）
func (r *userBalanceRepository) DecrementUserBalanceByChain(ctx context.Context, userID uint64, amount int64, updateSpent bool, sourceChain string) (*models.UserBalance, error) {
	var balance models.UserBalance

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取用户余额记录
		err := tx.Where("user_id = ? AND source_chain = ?", userID, sourceChain).First(&balance).Error
		if err != nil {
			return err
		}

		// 检查余额是否足够
		if balance.Balance < amount {
			return errors.New("insufficient balance")
		}

		// 更新余额
		updates := map[string]interface{}{
			"balance": gorm.Expr("balance - ?", amount),
			"version": gorm.Expr("version + 1"),
		}
		if updateSpent {
			updates["total_spent"] = gorm.Expr("total_spent + ?", amount)
		}

		result := tx.Model(&balance).Where("id = ? AND version = ?", balance.ID, balance.Version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("optimistic lock failed: balance was modified by another transaction")
		}

		// 重新获取更新后的记录
		return tx.Where("user_id = ? AND source_chain = ?", userID, sourceChain).First(&balance).Error
	})

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// CheckSufficientBalanceByChain 检查指定链的用户余额是否足够
func (r *userBalanceRepository) CheckSufficientBalanceByChain(ctx context.Context, userID uint64, amount int64, sourceChain string) (bool, error) {
	var balance models.UserBalance
	err := r.db.WithContext(ctx).Select("balance").Where("user_id = ? AND source_chain = ?", userID, sourceChain).First(&balance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return balance.Balance >= amount, nil
}
