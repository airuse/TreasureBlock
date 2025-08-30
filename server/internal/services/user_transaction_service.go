package services

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"errors"
	"fmt"
	"strconv"
)

// UserTransactionService 用户交易服务接口
type UserTransactionService interface {
	CreateTransaction(ctx context.Context, userID uint64, req *dto.CreateUserTransactionRequest) (*dto.UserTransactionResponse, error)
	GetTransactionByID(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
	GetUserTransactions(ctx context.Context, userID uint64, page, pageSize int, status string) (*dto.UserTransactionListResponse, error)
	UpdateTransaction(ctx context.Context, id uint, userID uint64, req *dto.UpdateUserTransactionRequest) (*dto.UserTransactionResponse, error)
	DeleteTransaction(ctx context.Context, id uint, userID uint64) error
	ExportTransaction(ctx context.Context, id uint, userID uint64) (*dto.ExportTransactionResponse, error)
	ImportSignature(ctx context.Context, id uint, userID uint64, req *dto.ImportSignatureRequest) (*dto.UserTransactionResponse, error)
	SendTransaction(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
	GetUserTransactionStats(ctx context.Context, userID uint64) (*dto.UserTransactionStatsResponse, error)
}

// userTransactionService 用户交易服务实现
type userTransactionService struct {
	userTxRepo repository.UserTransactionRepository
}

// NewUserTransactionService 创建用户交易服务实例
func NewUserTransactionService() UserTransactionService {
	return &userTransactionService{
		userTxRepo: repository.NewUserTransactionRepository(),
	}
}

// CreateTransaction 创建用户交易
func (s *userTransactionService) CreateTransaction(ctx context.Context, userID uint64, req *dto.CreateUserTransactionRequest) (*dto.UserTransactionResponse, error) {
	// 创建用户交易模型
	userTx := &models.UserTransaction{
		UserID:      userID,
		Chain:       req.Chain,
		Symbol:      req.Symbol,
		FromAddress: req.FromAddress,
		ToAddress:   req.ToAddress,
		Amount:      req.Amount,
		Fee:         req.Fee,
		GasLimit:    req.GasLimit,
		GasPrice:    req.GasPrice,
		Nonce:       req.Nonce,
		Status:      "draft", // 初始状态为草稿
		Remark:      req.Remark,

		// ERC-20相关字段
		TransactionType:       req.TransactionType,
		ContractOperationType: req.ContractOperationType,
		TokenContractAddress:  req.TokenContractAddress,
	}

	// 保存到数据库
	if err := s.userTxRepo.Create(ctx, userTx); err != nil {
		return nil, fmt.Errorf("创建交易失败: %w", err)
	}

	// 转换为响应DTO
	return s.convertToResponse(userTx), nil
}

// GetTransactionByID 根据ID获取用户交易
func (s *userTransactionService) GetTransactionByID(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(userTx), nil
}

// GetUserTransactions 获取用户交易列表
func (s *userTransactionService) GetUserTransactions(ctx context.Context, userID uint64, page, pageSize int, status string) (*dto.UserTransactionListResponse, error) {
	transactions, total, err := s.userTxRepo.GetByUserID(ctx, userID, page, pageSize, status)
	if err != nil {
		return nil, fmt.Errorf("获取交易列表失败: %w", err)
	}

	// 转换为响应DTO
	var responses []dto.UserTransactionResponse
	for _, tx := range transactions {
		responses = append(responses, *s.convertToResponse(tx))
	}

	// 计算是否有更多数据
	hasMore := int64(page*pageSize) < total

	return &dto.UserTransactionListResponse{
		Transactions: responses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		HasMore:      hasMore,
	}, nil
}

// UpdateTransaction 更新用户交易
func (s *userTransactionService) UpdateTransaction(ctx context.Context, id uint, userID uint64, req *dto.UpdateUserTransactionRequest) (*dto.UserTransactionResponse, error) {
	// 获取现有交易
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Status != nil {
		userTx.Status = *req.Status
	}
	if req.TxHash != nil {
		userTx.TxHash = req.TxHash
	}
	if req.UnsignedTx != nil {
		userTx.UnsignedTx = req.UnsignedTx
	}
	if req.SignedTx != nil {
		userTx.SignedTx = req.SignedTx
	}
	if req.BlockHeight != nil {
		userTx.BlockHeight = req.BlockHeight
	}
	if req.Confirmations != nil {
		userTx.Confirmations = req.Confirmations
	}
	if req.ErrorMsg != nil {
		userTx.ErrorMsg = req.ErrorMsg
	}
	if req.Remark != nil {
		userTx.Remark = *req.Remark
	}

	// 保存更新
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("更新交易失败: %w", err)
	}

	return s.convertToResponse(userTx), nil
}

// DeleteTransaction 删除用户交易
func (s *userTransactionService) DeleteTransaction(ctx context.Context, id uint, userID uint64) error {
	return s.userTxRepo.Delete(ctx, id, userID)
}

// ExportTransaction 导出交易
func (s *userTransactionService) ExportTransaction(ctx context.Context, id uint, userID uint64) (*dto.ExportTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 检查是否可以导出
	if !userTx.CanExport() {
		return nil, errors.New("当前状态的交易无法导出")
	}

	// 生成未签名交易数据（这里简化处理，实际应该调用区块链SDK）
	unsignedTx := s.generateUnsignedTx(userTx)

	// 更新交易状态为未签名
	userTx.Status = "unsigned"
	userTx.UnsignedTx = &unsignedTx
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("更新交易状态失败: %w", err)
	}

	return &dto.ExportTransactionResponse{
		UnsignedTx:  unsignedTx,
		Chain:       userTx.Chain,
		Symbol:      userTx.Symbol,
		FromAddress: userTx.FromAddress,
		ToAddress:   userTx.ToAddress,
		Amount:      userTx.Amount,
		Fee:         userTx.Fee,
		GasLimit:    userTx.GasLimit,
		GasPrice:    userTx.GasPrice,
		Nonce:       userTx.Nonce,
	}, nil
}

// ImportSignature 导入签名
func (s *userTransactionService) ImportSignature(ctx context.Context, id uint, userID uint64, req *dto.ImportSignatureRequest) (*dto.UserTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 检查是否可以导入签名
	if userTx.Status != "unsigned" {
		return nil, errors.New("只有未签名状态的交易才能导入签名")
	}

	// 更新签名数据
	userTx.SignedTx = &req.SignedTx
	userTx.Status = "unsent" // 状态变为未发送

	// 保存更新
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("导入签名失败: %w", err)
	}

	return s.convertToResponse(userTx), nil
}

// SendTransaction 发送交易
func (s *userTransactionService) SendTransaction(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 检查是否可以发送
	if !userTx.CanSend() {
		return nil, errors.New("只有未发送状态的交易才能发送")
	}

	// 这里应该调用区块链SDK发送交易
	// 简化处理，直接更新状态
	userTx.Status = "in_progress"

	// 保存更新
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("发送交易失败: %w", err)
	}

	return s.convertToResponse(userTx), nil
}

// GetUserTransactionStats 获取用户交易统计
func (s *userTransactionService) GetUserTransactionStats(ctx context.Context, userID uint64) (*dto.UserTransactionStatsResponse, error) {
	// 获取各种状态的交易数量
	statuses := []string{"draft", "unsigned", "unsent", "in_progress", "packed", "confirmed", "failed"}

	stats := &dto.UserTransactionStatsResponse{}

	for _, status := range statuses {
		transactions, err := s.userTxRepo.GetByStatus(ctx, userID, status)
		if err != nil {
			continue
		}

		count := int64(len(transactions))
		stats.TotalTransactions += count

		switch status {
		case "draft":
			stats.DraftCount = count
		case "unsigned":
			stats.UnsignedCount = count
		case "unsent":
			stats.UnsentCount = count
		case "in_progress":
			stats.InProgressCount = count
		case "packed":
			stats.PackedCount = count
		case "confirmed":
			stats.ConfirmedCount = count
		case "failed":
			stats.FailedCount = count
		}
	}

	return stats, nil
}

// convertToResponse 转换为响应DTO
func (s *userTransactionService) convertToResponse(userTx *models.UserTransaction) *dto.UserTransactionResponse {
	return &dto.UserTransactionResponse{
		ID:            userTx.ID,
		UserID:        userTx.UserID,
		Chain:         userTx.Chain,
		Symbol:        userTx.Symbol,
		FromAddress:   userTx.FromAddress,
		ToAddress:     userTx.ToAddress,
		Amount:        userTx.Amount,
		Fee:           userTx.Fee,
		GasLimit:      userTx.GasLimit,
		GasPrice:      userTx.GasPrice,
		Nonce:         userTx.Nonce,
		Status:        userTx.Status,
		TxHash:        userTx.TxHash,
		BlockHeight:   userTx.BlockHeight,
		Confirmations: userTx.Confirmations,
		ErrorMsg:      userTx.ErrorMsg,
		Remark:        userTx.Remark,
		CreatedAt:     userTx.CreatedAt,
		UpdatedAt:     userTx.UpdatedAt,

		// ERC-20相关字段
		TransactionType:       userTx.TransactionType,
		ContractOperationType: userTx.ContractOperationType,
		TokenContractAddress:  userTx.TokenContractAddress,
	}
}

// generateUnsignedTx 生成未签名交易数据（简化实现）
func (s *userTransactionService) generateUnsignedTx(userTx *models.UserTransaction) string {
	// 这里应该调用区块链SDK生成未签名交易
	// 简化处理，返回JSON格式的交易数据
	unsignedTx := fmt.Sprintf(`{
		"chain": "%s",
		"symbol": "%s",
		"from": "%s",
		"to": "%s",
		"amount": "%s",
		"fee": "%s",
		"gasLimit": %s,
		"gasPrice": "%s",
		"nonce": %s
	}`,
		userTx.Chain,
		userTx.Symbol,
		userTx.FromAddress,
		userTx.ToAddress,
		userTx.Amount,
		userTx.Fee,
		s.uintToString(userTx.GasLimit),
		s.stringToString(userTx.GasPrice),
		s.uint64ToString(userTx.Nonce),
	)

	return unsignedTx
}

// 辅助方法
func (s *userTransactionService) uintToString(u *uint) string {
	if u == nil {
		return "null"
	}
	return strconv.FormatUint(uint64(*u), 10)
}

func (s *userTransactionService) stringToString(str *string) string {
	if str == nil {
		return "null"
	}
	return *str
}

func (s *userTransactionService) uint64ToString(u *uint64) string {
	if u == nil {
		return "null"
	}
	return strconv.FormatUint(*u, 10)
}
