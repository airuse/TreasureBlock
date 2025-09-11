package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"

	"github.com/sirupsen/logrus"
)

// EarningsService 收益服务接口
type EarningsService interface {
	// 处理区块验证收益
	ProcessBlockVerificationEarnings(ctx context.Context, userID uint64, blockInfo *dto.BlockEarningsInfo) error
	// 获取用户收益记录列表
	GetUserEarningsRecords(ctx context.Context, userID uint64, req *dto.EarningsRecordListRequest) ([]*dto.EarningsRecordResponse, int64, error)
	// 获取用户余额
	GetUserBalance(ctx context.Context, userID uint64) (*dto.UserBalanceResponse, error)
	// 获取指定链的用户余额
	GetUserBalanceByChain(ctx context.Context, userID uint64, sourceChain string) (*dto.UserBalanceResponse, error)
	// 获取用户收益统计
	GetUserEarningsStats(ctx context.Context, userID uint64) (*dto.EarningsStatsResponse, error)
	// 获取收益趋势数据
	GetEarningsTrend(ctx context.Context, userID uint64, hours int) ([]*dto.EarningsTrendPoint, error)
	// 转账T币
	TransferTCoins(ctx context.Context, fromUserID, toUserID uint64, amount int64, description string, sourceChain string) (*dto.TransferResponse, error)
	// 消耗T币（业务消耗）
	ConsumeCoins(ctx context.Context, userID uint64, amount int64, source, description string, sourceChain string) error
	// 计算区块收益金额
	CalculateBlockEarnings(transactionCount int64) int64
}

type earningsService struct {
	earningsRepo    repository.EarningsRepository
	userBalanceRepo repository.UserBalanceRepository
	logger          *logrus.Logger
}

// NewEarningsService 创建收益服务
func NewEarningsService(
	earningsRepo repository.EarningsRepository,
	userBalanceRepo repository.UserBalanceRepository,
) EarningsService {
	return &earningsService{
		earningsRepo:    earningsRepo,
		userBalanceRepo: userBalanceRepo,
		logger:          logrus.New(),
	}
}

// GetEarningsTrend 获取收益趋势数据
func (s *earningsService) GetEarningsTrend(ctx context.Context, userID uint64, hours int) ([]*dto.EarningsTrendPoint, error) {
	if hours <= 0 {
		hours = 2 // 默认2小时
	}

	// 计算时间范围
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	// 获取指定时间范围内的收益记录
	records, err := s.earningsRepo.GetEarningsRecordsByDateRange(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get earnings records for trend: %w", err)
	}

	// 过滤只包含收益增加的记录
	var trendPoints []*dto.EarningsTrendPoint
	for _, record := range records {
		if record.Type == "add" && record.Source == "block_verification" {
			trendPoint := &dto.EarningsTrendPoint{
				Timestamp:        record.CreatedAt.Format("15:04"),
				Amount:           record.Amount,
				BlockHeight:      *record.BlockHeight,
				TransactionCount: *record.TransactionCount,
				SourceChain:      record.SourceChain,
			}
			trendPoints = append(trendPoints, trendPoint)
		}
	}

	// 按时间排序（从早到晚）
	sort.Slice(trendPoints, func(i, j int) bool {
		timeI, _ := time.Parse("15:04", trendPoints[i].Timestamp)
		timeJ, _ := time.Parse("15:04", trendPoints[j].Timestamp)
		return timeI.Before(timeJ)
	})

	return trendPoints, nil
}

// CalculateBlockEarnings 计算区块收益金额
func (s *earningsService) CalculateBlockEarnings(transactionCount int64) int64 {
	// 1个交易对应1个T币
	return transactionCount
}

// ProcessBlockVerificationEarnings 处理区块验证收益
func (s *earningsService) ProcessBlockVerificationEarnings(ctx context.Context, userID uint64, blockInfo *dto.BlockEarningsInfo) error {
	s.logger.Infof("Processing block verification earnings for user %d, block %d, transactions: %d",
		userID, blockInfo.BlockID, blockInfo.TransactionCount)

	// 计算收益金额
	earningsAmount := s.CalculateBlockEarnings(blockInfo.TransactionCount)
	if earningsAmount <= 0 {
		s.logger.Warnf("No earnings for block %d, transaction count: %d", blockInfo.BlockID, blockInfo.TransactionCount)
		return nil
	}

	// 获取用户当前余额
	currentBalance, err := s.userBalanceRepo.GetUserBalanceByChain(ctx, userID, blockInfo.Chain)
	if err != nil {
		s.logger.Errorf("Failed to get user balance: %v", err)
		return fmt.Errorf("failed to get user balance: %w", err)
	}

	// 增加用户余额
	newBalance, err := s.userBalanceRepo.IncrementUserBalanceByChain(ctx, userID, earningsAmount, true, blockInfo.Chain)
	if err != nil {
		s.logger.Errorf("Failed to increment user balance: %v", err)
		return fmt.Errorf("failed to increment user balance: %w", err)
	}

	// 创建收益记录
	earningsRecord := &models.EarningsRecord{
		UserID:           userID,
		Amount:           earningsAmount,
		Type:             "add",
		Source:           "block_verification",
		SourceID:         &blockInfo.BlockID,
		SourceChain:      blockInfo.Chain,
		BlockHeight:      &blockInfo.BlockHeight,
		TransactionCount: &blockInfo.TransactionCount,
		Description:      fmt.Sprintf("扫块收益 - 区块高度: %d, 交易数量: %d", blockInfo.BlockHeight, blockInfo.TransactionCount),
		BalanceBefore:    currentBalance.Balance,
		BalanceAfter:     newBalance.Balance,
	}

	if err := s.earningsRepo.CreateEarningsRecord(ctx, earningsRecord); err != nil {
		s.logger.Errorf("Failed to create earnings record: %v", err)
		return fmt.Errorf("failed to create earnings record: %w", err)
	}

	s.logger.Infof("Successfully processed block verification earnings: user %d earned %d T-coins for block %d",
		userID, earningsAmount, blockInfo.BlockID)

	return nil
}

// GetUserEarningsRecords 获取用户收益记录列表
func (s *earningsService) GetUserEarningsRecords(ctx context.Context, userID uint64, req *dto.EarningsRecordListRequest) ([]*dto.EarningsRecordResponse, int64, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.Type != "" {
		filters["type"] = req.Type
	}
	if req.Source != "" {
		filters["source"] = req.Source
	}
	if req.Chain != "" {
		filters["chain"] = req.Chain
	}
	if req.StartDate != "" {
		filters["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		filters["end_date"] = req.EndDate
	}

	// 获取记录
	records, total, err := s.earningsRepo.GetEarningsRecordsByUserID(ctx, userID, req.Page, req.PageSize, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get earnings records: %w", err)
	}

	// 转换为DTO
	var response []*dto.EarningsRecordResponse
	for _, record := range records {
		response = append(response, &dto.EarningsRecordResponse{
			ID:               record.ID,
			UserID:           record.UserID,
			Amount:           record.Amount,
			Type:             record.Type,
			Source:           record.Source,
			SourceID:         record.SourceID,
			SourceChain:      record.SourceChain,
			BlockHeight:      record.BlockHeight,
			TransactionCount: record.TransactionCount,
			Description:      record.Description,
			BalanceBefore:    record.BalanceBefore,
			BalanceAfter:     record.BalanceAfter,
			CreatedAt:        record.CreatedAt,
			UpdatedAt:        record.UpdatedAt,
		})
	}

	return response, total, nil
}

// GetUserBalance 获取用户余额
func (s *earningsService) GetUserBalance(ctx context.Context, userID uint64) (*dto.UserBalanceResponse, error) {
	balance, err := s.userBalanceRepo.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return &dto.UserBalanceResponse{
		ID:              balance.ID,
		UserID:          balance.UserID,
		Balance:         balance.Balance,
		TotalEarned:     balance.TotalEarned,
		TotalSpent:      balance.TotalSpent,
		LastEarningTime: balance.LastEarningTime,
		LastSpendTime:   balance.LastSpendTime,
		CreatedAt:       balance.CreatedAt,
		UpdatedAt:       balance.UpdatedAt,
	}, nil
}

// GetUserBalanceByChain 获取指定链的用户余额
func (s *earningsService) GetUserBalanceByChain(ctx context.Context, userID uint64, sourceChain string) (*dto.UserBalanceResponse, error) {
	balance, err := s.userBalanceRepo.GetUserBalanceByChain(ctx, userID, sourceChain)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance by chain: %w", err)
	}

	return &dto.UserBalanceResponse{
		ID:              balance.ID,
		UserID:          balance.UserID,
		Balance:         balance.Balance,
		TotalEarned:     balance.TotalEarned,
		TotalSpent:      balance.TotalSpent,
		LastEarningTime: balance.LastEarningTime,
		LastSpendTime:   balance.LastSpendTime,
		CreatedAt:       balance.CreatedAt,
		UpdatedAt:       balance.UpdatedAt,
	}, nil
}

// GetUserEarningsStats 获取用户收益统计
func (s *earningsService) GetUserEarningsStats(ctx context.Context, userID uint64) (*dto.EarningsStatsResponse, error) {
	stats, err := s.earningsRepo.GetEarningsStatsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get earnings stats: %w", err)
	}

	return &dto.EarningsStatsResponse{
		UserID:           stats.UserID,
		TotalEarnings:    stats.TotalEarnings,
		TotalSpendings:   stats.TotalSpendings,
		CurrentBalance:   stats.CurrentBalance,
		BlockCount:       stats.BlockCount,
		TransactionCount: stats.TransactionCount,
	}, nil
}

// TransferTCoins 转账T币
func (s *earningsService) TransferTCoins(ctx context.Context, fromUserID, toUserID uint64, amount int64, description string, sourceChain string) (*dto.TransferResponse, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid transfer amount: %d", amount)
	}

	if fromUserID == toUserID {
		return nil, fmt.Errorf("cannot transfer to yourself")
	}

	// 检查发送方余额是否足够
	sufficient, err := s.userBalanceRepo.CheckSufficientBalanceByChain(ctx, fromUserID, amount, sourceChain)
	if err != nil {
		return nil, fmt.Errorf("failed to check balance: %w", err)
	}
	if !sufficient {
		return nil, fmt.Errorf("insufficient balance")
	}

	// 获取发送方和接收方的余额信息
	fromBalance, err := s.userBalanceRepo.GetUserBalanceByChain(ctx, fromUserID, sourceChain)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender balance: %w", err)
	}

	toBalance, err := s.userBalanceRepo.GetUserBalanceByChain(ctx, toUserID, sourceChain)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver balance: %w", err)
	}

	// 执行转账操作
	// 扣除发送方余额
	newFromBalance, err := s.userBalanceRepo.DecrementUserBalanceByChain(ctx, fromUserID, amount, false, sourceChain)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct sender balance: %w", err)
	}

	// 增加接收方余额
	newToBalance, err := s.userBalanceRepo.IncrementUserBalanceByChain(ctx, toUserID, amount, false, sourceChain)
	if err != nil {
		// 如果接收方操作失败，需要回滚发送方操作
		// 这里简化处理，实际应该使用事务
		s.logger.Errorf("Failed to increment receiver balance, need manual intervention: %v", err)
		return nil, fmt.Errorf("failed to increment receiver balance: %w", err)
	}

	// 创建转账记录
	transferTime := time.Now()
	transferDescription := description
	if transferDescription == "" {
		transferDescription = fmt.Sprintf("转账给用户 %d", toUserID)
	}

	// 创建发送方记录
	fromRecord := &models.EarningsRecord{
		UserID:        fromUserID,
		Amount:        amount,
		Type:          "decrease",
		Source:        "transfer_out",
		Description:   transferDescription,
		BalanceBefore: fromBalance.Balance,
		BalanceAfter:  newFromBalance.Balance,
		CreatedAt:     transferTime,
	}

	// 创建接收方记录
	toRecord := &models.EarningsRecord{
		UserID:        toUserID,
		Amount:        amount,
		Type:          "add",
		Source:        "transfer_in",
		Description:   fmt.Sprintf("来自用户 %d 的转账", fromUserID),
		BalanceBefore: toBalance.Balance,
		BalanceAfter:  newToBalance.Balance,
		CreatedAt:     transferTime,
	}

	// 批量创建记录
	if err := s.earningsRepo.CreateEarningsRecordsBatch(ctx, []*models.EarningsRecord{fromRecord, toRecord}); err != nil {
		s.logger.Errorf("Failed to create transfer records: %v", err)
		return nil, fmt.Errorf("failed to create transfer records: %w", err)
	}

	return &dto.TransferResponse{
		FromUserID:        fromUserID,
		ToUserID:          toUserID,
		Amount:            amount,
		Description:       transferDescription,
		FromBalanceBefore: fromBalance.Balance,
		FromBalanceAfter:  newFromBalance.Balance,
		ToBalanceBefore:   toBalance.Balance,
		ToBalanceAfter:    newToBalance.Balance,
		TransferTime:      transferTime,
	}, nil
}

// ConsumeCoins 消耗T币（业务消耗）
func (s *earningsService) ConsumeCoins(ctx context.Context, userID uint64, amount int64, source, description string, sourceChain string) error {
	if amount <= 0 {
		return fmt.Errorf("invalid consume amount: %d", amount)
	}

	// 检查余额是否足够
	sufficient, err := s.userBalanceRepo.CheckSufficientBalance(ctx, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}
	if !sufficient {
		return fmt.Errorf("insufficient balance")
	}

	// 获取当前余额
	currentBalance, err := s.userBalanceRepo.GetUserBalanceByChain(ctx, userID, sourceChain)
	if err != nil {
		return fmt.Errorf("failed to get user balance: %w", err)
	}

	// 减少用户余额
	newBalance, err := s.userBalanceRepo.DecrementUserBalance(ctx, userID, amount, true)
	if err != nil {
		return fmt.Errorf("failed to decrement user balance: %w", err)
	}

	// 创建消耗记录
	consumeRecord := &models.EarningsRecord{
		UserID:        userID,
		Amount:        amount,
		Type:          "decrease",
		Source:        source,
		Description:   description,
		BalanceBefore: currentBalance.Balance,
		BalanceAfter:  newBalance.Balance,
	}

	if err := s.earningsRepo.CreateEarningsRecord(ctx, consumeRecord); err != nil {
		s.logger.Errorf("Failed to create consume record: %v", err)
		return fmt.Errorf("failed to create consume record: %w", err)
	}

	s.logger.Infof("Successfully consumed %d T-coins for user %d, source: %s", amount, userID, source)

	return nil
}
