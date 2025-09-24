package services

import (
	"context"
	"time"

	"blockChainBrowser/server/internal/repository"
)

// StatsService 统计服务接口
type StatsService interface {
	GetActiveAddressesCount(ctx context.Context, chain string, duration time.Duration) (int64, error)
	GetNetworkHashrate(ctx context.Context, chain string) (int64, error)
	GetDailyVolume(ctx context.Context, chain string, duration time.Duration) (float64, error)
	GetAverageGasPrice(ctx context.Context, chain string, duration time.Duration) (int64, error)
	GetAverageBlockTime(ctx context.Context, chain string, duration time.Duration) (float64, error)
	GetTotalBlockCount(ctx context.Context, chain string) (int64, error)
	GetTotalTransactionCount(ctx context.Context, chain string) (int64, error)
	GetLatestBaseFee(ctx context.Context, chain string) (int64, error)
	GetCurrentDifficulty(ctx context.Context, chain string) (int64, error)
	// Solana specific methods
	GetTotalSolanaSlotCount(ctx context.Context) (int64, error)
	GetTotalSolanaTransactionCount(ctx context.Context) (int64, error)
	GetSolanaDailyVolume(ctx context.Context, duration time.Duration) (float64, error)
	GetSolanaAverageFee(ctx context.Context, duration time.Duration) (int64, error)
	GetSolanaAverageSlotTime(ctx context.Context, duration time.Duration) (float64, error)
}

// statsService 统计服务实现
type statsService struct {
	statsRepo repository.StatsRepository
}

// NewStatsService 创建统计服务实例
func NewStatsService(statsRepo repository.StatsRepository) StatsService {
	return &statsService{
		statsRepo: statsRepo,
	}
}

// GetActiveAddressesCount 获取活跃地址数量
func (s *statsService) GetActiveAddressesCount(ctx context.Context, chain string, duration time.Duration) (int64, error) {
	return s.statsRepo.GetActiveAddressesCount(ctx, chain, duration)
}

// GetNetworkHashrate 获取网络算力
func (s *statsService) GetNetworkHashrate(ctx context.Context, chain string) (int64, error) {
	return s.statsRepo.GetNetworkHashrate(ctx, chain)
}

// GetDailyVolume 获取指定时间范围内的交易量
func (s *statsService) GetDailyVolume(ctx context.Context, chain string, duration time.Duration) (float64, error) {
	return s.statsRepo.GetDailyVolume(ctx, chain, duration)
}

// GetAverageGasPrice 获取指定时间范围内的平均Gas价格
func (s *statsService) GetAverageGasPrice(ctx context.Context, chain string, duration time.Duration) (int64, error) {
	return s.statsRepo.GetAverageGasPrice(ctx, chain, duration)
}

// GetAverageBlockTime 获取指定时间范围内的平均出块时间
func (s *statsService) GetAverageBlockTime(ctx context.Context, chain string, duration time.Duration) (float64, error) {
	return s.statsRepo.GetAverageBlockTime(ctx, chain, duration)
}

// GetTotalBlockCount 获取总区块数
func (s *statsService) GetTotalBlockCount(ctx context.Context, chain string) (int64, error) {
	return s.statsRepo.GetTotalBlockCount(ctx, chain)
}

// GetTotalTransactionCount 获取总交易数
func (s *statsService) GetTotalTransactionCount(ctx context.Context, chain string) (int64, error) {
	return s.statsRepo.GetTotalTransactionCount(ctx, chain)
}

// GetLatestBaseFee 获取最新区块的Base Fee
func (s *statsService) GetLatestBaseFee(ctx context.Context, chain string) (int64, error) {
	return s.statsRepo.GetLatestBaseFee(ctx, chain)
}

// GetCurrentDifficulty 获取当前难度
func (s *statsService) GetCurrentDifficulty(ctx context.Context, chain string) (int64, error) {
	return s.statsRepo.GetCurrentDifficulty(ctx, chain)
}

// GetTotalSolanaSlotCount 获取Solana总slot数
func (s *statsService) GetTotalSolanaSlotCount(ctx context.Context) (int64, error) {
	return s.statsRepo.GetTotalSolanaSlotCount(ctx)
}

// GetTotalSolanaTransactionCount 获取Solana总交易数
func (s *statsService) GetTotalSolanaTransactionCount(ctx context.Context) (int64, error) {
	return s.statsRepo.GetTotalSolanaTransactionCount(ctx)
}

// GetSolanaDailyVolume 获取Solana指定时间范围内的交易量
func (s *statsService) GetSolanaDailyVolume(ctx context.Context, duration time.Duration) (float64, error) {
	return s.statsRepo.GetSolanaDailyVolume(ctx, duration)
}

// GetSolanaAverageFee 获取Solana指定时间范围内的平均费用
func (s *statsService) GetSolanaAverageFee(ctx context.Context, duration time.Duration) (int64, error) {
	return s.statsRepo.GetSolanaAverageFee(ctx, duration)
}

// GetSolanaAverageSlotTime 获取Solana指定时间范围内的平均出块时间
func (s *statsService) GetSolanaAverageSlotTime(ctx context.Context, duration time.Duration) (float64, error) {
	return s.statsRepo.GetSolanaAverageSlotTime(ctx, duration)
}
