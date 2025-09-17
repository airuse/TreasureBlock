package repository

import (
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/models"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// StatsRepository 统计仓库接口
type StatsRepository interface {
	GetActiveAddressesCount(ctx context.Context, chain string, duration time.Duration) (int64, error)
	GetNetworkHashrate(ctx context.Context, chain string) (int64, error)
	GetDailyVolume(ctx context.Context, chain string, duration time.Duration) (float64, error)
	GetAverageGasPrice(ctx context.Context, chain string, duration time.Duration) (int64, error)
	GetAverageBlockTime(ctx context.Context, chain string, duration time.Duration) (float64, error)
	GetTotalBlockCount(ctx context.Context, chain string) (int64, error)
	GetTotalTransactionCount(ctx context.Context, chain string) (int64, error)
	GetLatestBaseFee(ctx context.Context, chain string) (int64, error)
	GetCurrentDifficulty(ctx context.Context, chain string) (int64, error)
}

// statsRepository 统计仓储实现
type statsRepository struct {
	db *gorm.DB
}

// NewStatsRepository 创建统计仓储实例
func NewStatsRepository() StatsRepository {
	return &statsRepository{
		db: database.GetDB(),
	}
}

// GetTotalBlockCount 获取总区块数
func (r *statsRepository) GetTotalBlockCount(ctx context.Context, chain string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("chain = ?", chain).
		Count(&count).Error
	return count, err
}

// GetTotalTransactionCount 获取总交易数
func (r *statsRepository) GetTotalTransactionCount(ctx context.Context, chain string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Where("chain = ?", chain).
		Count(&count).Error
	return count, err
}

// GetLatestBaseFee 获取最新区块的Base Fee
func (r *statsRepository) GetLatestBaseFee(ctx context.Context, chain string) (int64, error) {
	// 只对ETH和BSC获取Base Fee
	if chain != "eth" && chain != "bsc" {
		return 0, nil
	}

	var baseFee string
	err := r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("chain = ? AND base_fee IS NOT NULL AND base_fee != ''", chain).
		Order("height DESC").
		Limit(1).
		Select("base_fee").
		Scan(&baseFee).Error

	if err != nil || baseFee == "" {
		return 20000000000, nil // 20 Gwei 默认值
	}

	// 尝试将字符串转换为int64
	var result int64
	_, err = fmt.Sscanf(baseFee, "%d", &result)
	if err != nil {
		// 如果转换失败，返回默认值
		return 20000000000, nil // 20 Gwei 默认值
	}

	return result, nil
}

// GetActiveAddressesCount 获取活跃地址数（指定时间范围内有交易的地址）
func (r *statsRepository) GetActiveAddressesCount(ctx context.Context, chain string, duration time.Duration) (int64, error) {
	// 暂时返回0，因为查询太慢
	// TODO: 后续可以通过缓存或异步计算来优化
	return 0, nil
}

// GetNetworkHashrate 获取网络状态指标
func (r *statsRepository) GetNetworkHashrate(ctx context.Context, chain string) (int64, error) {
	// 暂时返回0，因为验证者数据拿不到
	// TODO: 后续可以接入真实的验证者API
	return 0, nil
}

// GetDailyVolume 获取指定时间范围内的交易量
func (r *statsRepository) GetDailyVolume(ctx context.Context, chain string, duration time.Duration) (float64, error) {
	// 计算指定时间范围前的时间
	startTime := time.Now().Add(-duration)

	// 使用GORM查询，配合索引提升性能
	var totalVolume float64
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Where("chain = ? AND ctime >= ? AND amount IS NOT NULL AND amount != ''", chain, startTime).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(65,18))), 0)").
		Scan(&totalVolume).Error

	return totalVolume, err
}

// GetAverageGasPrice 获取指定时间范围内的平均Gas价格
func (r *statsRepository) GetAverageGasPrice(ctx context.Context, chain string, duration time.Duration) (int64, error) {
	// 只对ETH和BSC计算Gas价格
	if chain != "eth" && chain != "bsc" {
		return 0, nil
	}

	// 计算指定时间范围前的时间
	startTime := time.Now().Add(-duration)

	// 使用GORM查询，配合索引提升性能
	var avgGasPrice float64
	err := r.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Where("chain = ? AND ctime >= ? AND gas_price IS NOT NULL AND gas_price != '' AND gas_price != '0'", chain, startTime).
		Select("AVG(CAST(gas_price AS DECIMAL(65,0)))").
		Scan(&avgGasPrice).Error

	if err != nil {
		return 0, err
	}

	// 如果查询结果为空或为0，返回默认值
	if avgGasPrice == 0 {
		return 20000000000, nil // 20 Gwei 默认值
	}

	return int64(avgGasPrice), err
}

// GetAverageBlockTime 获取指定时间范围内的平均出块时间
func (r *statsRepository) GetAverageBlockTime(ctx context.Context, chain string, duration time.Duration) (float64, error) {
	// 计算指定时间范围前的时间
	startTime := time.Now().Add(-duration)

	// 使用GORM查询，只获取验证通过的区块
	var blocks []models.Block
	err := r.db.WithContext(ctx).
		Where("chain = ? AND created_at >= ? AND is_verified = 1", chain, startTime).
		Order("height DESC").
		Select("height, timestamp").
		Find(&blocks).Error

	if err != nil || len(blocks) < 2 {
		return 0, err
	}

	// 计算平均出块时间
	var totalTime float64
	blockCount := 0

	for i := 0; i < len(blocks)-1; i++ {
		// Timestamp字段已经是time.Time类型，直接使用
		timeDiff := blocks[i].Timestamp.Sub(blocks[i+1].Timestamp)
		totalTime += timeDiff.Seconds()
		blockCount++
	}

	if blockCount == 0 {
		// 如果没有计算出块时间，返回默认值
		if chain == "eth" {
			return 12.0, nil // ETH默认出块时间12秒
		}
		return 0, nil
	}

	return totalTime / float64(blockCount), nil
}

// GetCurrentDifficulty 获取当前难度
func (r *statsRepository) GetCurrentDifficulty(ctx context.Context, chain string) (int64, error) {
	// 获取最新区块的难度
	var difficulty string
	err := r.db.WithContext(ctx).
		Model(&models.Block{}).
		Where("chain = ? AND difficulty IS NOT NULL AND difficulty != ''", chain).
		Order("height DESC").
		Limit(1).
		Select("difficulty").
		Scan(&difficulty).Error

	if err != nil || difficulty == "" {
		return 0, err
	}

	// 尝试将字符串转换为int64
	var result int64
	_, err = fmt.Sscanf(difficulty, "%d", &result)
	if err != nil {
		// 如果转换失败，尝试科学计数法
		if strings.Contains(difficulty, "e") || strings.Contains(difficulty, "E") {
			// 处理科学计数法，如 1.25e+24
			var floatVal float64
			_, err = fmt.Sscanf(difficulty, "%e", &floatVal)
			if err == nil {
				result = int64(floatVal)
			}
		}
	}

	return result, nil
}
