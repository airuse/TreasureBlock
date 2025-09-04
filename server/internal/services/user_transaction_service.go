package services

import (
	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// UserTransactionService 用户交易服务接口
type UserTransactionService interface {
	CreateTransaction(ctx context.Context, userID uint64, req *dto.CreateUserTransactionRequest) (*dto.UserTransactionResponse, error)
	GetTransactionByID(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
	GetUserTransactions(ctx context.Context, userID uint64, page, pageSize int, status string) (*dto.UserTransactionListResponse, error)
	UpdateTransaction(ctx context.Context, id uint, userID uint64, req *dto.UpdateUserTransactionRequest) (*dto.UserTransactionResponse, error)
	DeleteTransaction(ctx context.Context, id uint, userID uint64) error
	ExportTransaction(ctx context.Context, id uint, userID uint64, req *dto.ExportTransactionRequest) (*dto.ExportTransactionResponse, error)
	ImportSignature(ctx context.Context, id uint, userID uint64, req *dto.ImportSignatureRequest) (*dto.UserTransactionResponse, error)
	SendTransaction(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
	GetUserTransactionStats(ctx context.Context, userID uint64) (*dto.UserTransactionStatsResponse, error)
}

// userTransactionService 用户交易服务实现
type userTransactionService struct {
	userTxRepo repository.UserTransactionRepository
	logger     *logrus.Logger
}

// NewUserTransactionService 创建用户交易服务实例
func NewUserTransactionService() UserTransactionService {
	return &userTransactionService{
		userTxRepo: repository.NewUserTransactionRepository(),
		logger:     logrus.New(),
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
func (s *userTransactionService) ExportTransaction(ctx context.Context, id uint, userID uint64, req *dto.ExportTransactionRequest) (*dto.ExportTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 检查是否可以导出
	if !userTx.CanExport() {
		return nil, errors.New("当前状态的交易无法导出")
	}

	// 获取发送地址的当前nonce（如果交易中没有设置nonce）
	currentNonce := userTx.Nonce
	if currentNonce == nil {
		// 使用ethclient获取地址的当前nonce
		nonce, err := s.getAddressNonce(ctx, userTx.FromAddress)
		if err != nil {
			// 如果获取nonce失败，使用默认值0
			fmt.Printf("获取地址nonce失败: %v，使用默认值0\n", err)
			defaultNonce := uint64(0)
			currentNonce = &defaultNonce
		} else {
			currentNonce = &nonce
		}

		// 更新交易记录中的nonce
		userTx.Nonce = currentNonce
	}

	// 生成未签名交易数据（这里简化处理，实际应该调用区块链SDK）
	unsignedTx := s.generateUnsignedTx(userTx)

	// 生成QR码数据（使用配置中的链ID）
	chainID := ""
	if chainCfg, ok := config.AppConfig.Blockchain.Chains[strings.ToLower(userTx.Chain)]; ok {
		chainID = strconv.Itoa(chainCfg.ChainID)
	} else {
		if strings.ToLower(userTx.Chain) == "eth" {
			chainID = "1"
		} else {
			chainID = userTx.Chain
		}
	}

	// 生成交易数据（这里需要调用前端相同的逻辑，暂时使用占位符）
	txData := s.generateTxData(userTx)

	// 生成AccessList（如果是代币交易）
	accessList := s.generateAccessList(userTx)

	// 处理费率设置
	if req.MaxPriorityFeePerGas != nil {
		userTx.MaxPriorityFeePerGas = req.MaxPriorityFeePerGas
	} else if userTx.MaxPriorityFeePerGas == nil {
		// 如果没有设置费率，使用默认值
		defaultTip := "2" // 2 Gwei
		userTx.MaxPriorityFeePerGas = &defaultTip
	}

	if req.MaxFeePerGas != nil {
		userTx.MaxFeePerGas = req.MaxFeePerGas
	} else if userTx.MaxFeePerGas == nil {
		// 如果没有设置费率，使用默认值
		defaultFee := "30" // 30 Gwei
		userTx.MaxFeePerGas = &defaultFee
	}

	// 更新交易状态为未签名，并保存QR码数据
	userTx.Status = "unsigned"
	userTx.UnsignedTx = &unsignedTx
	userTx.ChainID = &chainID
	userTx.TxData = &txData
	userTx.AccessList = &accessList

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
		Nonce:       currentNonce, // 使用获取到的nonce
		ChainID:     &chainID,
		TxData:      &txData,
		AccessList:  &accessList,
		// 添加费率字段
		MaxPriorityFeePerGas: userTx.MaxPriorityFeePerGas,
		MaxFeePerGas:         userTx.MaxFeePerGas,
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
	userTx.Status = "in_progress" // 直接设置为在途状态，因为会自动发送

	// 保存签名组件
	if req.V != nil {
		userTx.V = req.V
	}
	if req.R != nil {
		userTx.R = req.R
	}
	if req.S != nil {
		userTx.S = req.S
	}

	// 保存更新
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("导入签名失败: %w", err)
	}

	// 自动发送交易
	sendResp, err := s.SendTransaction(ctx, id, userID)
	if err != nil {
		// 发送失败，更新状态为失败
		userTx.Status = "failed"
		errorMsg := err.Error()
		userTx.ErrorMsg = &errorMsg
		s.userTxRepo.Update(ctx, userTx)
		return nil, fmt.Errorf("自动发送交易失败: %w", err)
	}

	return sendResp, nil
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

	// 检查是否有已签名的交易数据
	if userTx.SignedTx == nil || *userTx.SignedTx == "" {
		return nil, errors.New("交易尚未签名，无法发送")
	}

	// 创建RPC客户端管理器
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()

	// 准备发送交易请求
	sendReq := &utils.SendTransactionRequest{
		Chain:       userTx.Chain,
		SignedTx:    *userTx.SignedTx,
		FromAddress: userTx.FromAddress,
		ToAddress:   userTx.ToAddress,
		Amount:      userTx.Amount,
		Fee:         userTx.Fee,
	}

	// 调用RPC发送交易
	sendResp, err := rpcManager.SendTransaction(ctx, sendReq)
	if err != nil {
		s.logger.Errorf("发送交易失败: %v", err)
		return nil, fmt.Errorf("发送交易失败: %w", err)
	}

	if !sendResp.Success {
		// 发送失败，更新状态为失败
		userTx.Status = "failed"
		errorMsg := sendResp.Message
		userTx.ErrorMsg = &errorMsg

		if err := s.userTxRepo.Update(ctx, userTx); err != nil {
			s.logger.Errorf("更新交易状态失败: %v", err)
		}

		return nil, fmt.Errorf("发送交易失败: %s", sendResp.Message)
	}

	// 发送成功，更新交易状态
	userTx.Status = "in_progress"
	userTx.TxHash = &sendResp.TxHash
	userTx.ErrorMsg = nil // 清除之前的错误信息

	// 保存更新
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("更新交易状态失败: %w", err)
	}

	s.logger.Infof("交易发送成功: ID=%d, TxHash=%s", userTx.ID, sendResp.TxHash)

	// 异步更新交易状态（从区块链查询最新状态）
	go s.updateTransactionStatusAsync(context.Background(), userTx.ID, userID)

	return s.convertToResponse(userTx), nil
}

// updateTransactionStatusAsync 异步更新交易状态（从区块链查询）
func (s *userTransactionService) updateTransactionStatusAsync(ctx context.Context, id uint, userID uint64) {
	// 等待一段时间让交易在区块链上确认
	time.Sleep(5 * time.Second)

	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		s.logger.Errorf("获取交易失败: %v", err)
		return
	}

	// 只有已发送的交易才需要查询状态
	if userTx.Status != "in_progress" && userTx.Status != "packed" {
		return
	}

	// 检查是否有交易哈希
	if userTx.TxHash == nil || *userTx.TxHash == "" {
		return
	}

	// 创建RPC客户端管理器
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()

	// 查询交易状态
	txStatus, err := rpcManager.GetTransactionStatus(ctx, userTx.Chain, *userTx.TxHash)
	if err != nil {
		s.logger.Errorf("查询交易状态失败: %v", err)
		return
	}

	// 根据查询结果更新状态
	oldStatus := userTx.Status
	switch txStatus.Status {
	case "pending":
		userTx.Status = "in_progress"
	case "confirmed":
		userTx.Status = "confirmed"
		if txStatus.BlockHeight > 0 {
			userTx.BlockHeight = &txStatus.BlockHeight
		}
		if txStatus.Confirmations > 0 {
			confirmations := uint(txStatus.Confirmations)
			userTx.Confirmations = &confirmations
		}
	case "failed":
		userTx.Status = "failed"
		errorMsg := "交易在区块链上失败"
		userTx.ErrorMsg = &errorMsg
	}

	// 如果状态有变化，保存更新
	if userTx.Status != oldStatus {
		if err := s.userTxRepo.Update(ctx, userTx); err != nil {
			s.logger.Errorf("更新交易状态失败: %v", err)
		} else {
			s.logger.Infof("交易状态已更新: ID=%d, 从 %s 到 %s", userTx.ID, oldStatus, userTx.Status)
		}
	}
}

// updateTransactionStatus 更新交易状态（内部方法）
func (s *userTransactionService) updateTransactionStatus(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 只有已发送的交易才需要查询状态
	if userTx.Status != "in_progress" && userTx.Status != "packed" {
		return s.convertToResponse(userTx), nil
	}

	// 检查是否有交易哈希
	if userTx.TxHash == nil || *userTx.TxHash == "" {
		return s.convertToResponse(userTx), nil
	}

	// 创建RPC客户端管理器
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()

	// 查询交易状态
	txStatus, err := rpcManager.GetTransactionStatus(ctx, userTx.Chain, *userTx.TxHash)
	if err != nil {
		s.logger.Errorf("查询交易状态失败: %v", err)
		// 查询失败不影响返回，只是不更新状态
		return s.convertToResponse(userTx), nil
	}

	// 根据查询结果更新状态
	oldStatus := userTx.Status
	switch txStatus.Status {
	case "pending":
		userTx.Status = "in_progress"
	case "confirmed":
		userTx.Status = "confirmed"
		if txStatus.BlockHeight > 0 {
			userTx.BlockHeight = &txStatus.BlockHeight
		}
		if txStatus.Confirmations > 0 {
			confirmations := uint(txStatus.Confirmations)
			userTx.Confirmations = &confirmations
		}
	case "failed":
		userTx.Status = "failed"
		errorMsg := "交易在区块链上失败"
		userTx.ErrorMsg = &errorMsg
	}

	// 如果状态有变化，保存更新
	if userTx.Status != oldStatus {
		if err := s.userTxRepo.Update(ctx, userTx); err != nil {
			s.logger.Errorf("更新交易状态失败: %v", err)
		} else {
			s.logger.Infof("交易状态已更新: ID=%d, 从 %s 到 %s", userTx.ID, oldStatus, userTx.Status)
		}
	}

	return s.convertToResponse(userTx), nil
}

// GetUserTransactionStats 获取用户交易统计
func (s *userTransactionService) GetUserTransactionStats(ctx context.Context, userID uint64) (*dto.UserTransactionStatsResponse, error) {
	// 获取各种状态的交易数量
	statuses := []string{"draft", "unsigned", "in_progress", "packed", "confirmed", "failed"}

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

		// QR码导出相关字段
		ChainID:    userTx.ChainID,
		TxData:     userTx.TxData,
		AccessList: userTx.AccessList,

		// 签名组件
		V: userTx.V,
		R: userTx.R,
		S: userTx.S,
	}
}

// generateUnsignedTx 生成未签名交易数据（简化实现）
func (s *userTransactionService) generateUnsignedTx(userTx *models.UserTransaction) string {
	// 这里应该调用区块链SDK生成未签名交易
	// 简化处理，返回JSON格式的交易数据，包含EIP-1559费率
	unsignedTx := fmt.Sprintf(`{
		"chain": "%s",
		"symbol": "%s",
		"from": "%s",
		"to": "%s",
		"amount": "%s",
		"fee": "%s",
		"gasLimit": %s,
		"gasPrice": "%s",
		"nonce": %s,
		"maxPriorityFeePerGas": "%s",
		"maxFeePerGas": "%s"
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
		s.stringToString(userTx.MaxPriorityFeePerGas),
		s.stringToString(userTx.MaxFeePerGas),
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

// getAddressNonce 获取地址的当前nonce
func (s *userTransactionService) getAddressNonce(ctx context.Context, address string) (uint64, error) {
	// 从配置文件获取ETH RPC URL
	chainConfig, exists := config.AppConfig.Blockchain.Chains["eth"]
	if !exists || (chainConfig.RPCURL == "" && len(chainConfig.RPCURLs) == 0) {
		return 0, fmt.Errorf("未配置ETH RPC URL")
	}

	// 使用故障转移管理器
	fo, err := utils.NewEthFailoverFromChain("eth")
	if err != nil {
		return 0, fmt.Errorf("初始化ETH故障转移失败: %w", err)
	}
	defer fo.Close()

	// 获取地址的当前nonce（故障转移）
	nonce, err := fo.NonceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return 0, fmt.Errorf("获取地址nonce失败: %w", err)
	}

	return nonce, nil
}

// generateTxData 生成交易数据（十六进制格式）
func (s *userTransactionService) generateTxData(userTx *models.UserTransaction) string {
	// 根据交易类型生成不同的数据
	if userTx.TransactionType == "token" && userTx.TokenContractAddress != "" {
		switch userTx.ContractOperationType {
		case "balanceOf":
			// balanceOf(address) 函数选择器: 0x70a08231
			return fmt.Sprintf("0x70a08231%s", s.padAddress(userTx.FromAddress))
		case "transfer":
			// transfer(address,uint256) 函数选择器: 0xa9059cbb
			amountHex := s.convertAmountToHex(userTx.Amount)
			return fmt.Sprintf("0xa9059cbb%s%s", s.padAddress(userTx.ToAddress), amountHex)
		case "approve":
			// approve(address,uint256) 函数选择器: 0x095ea7b3
			amountHex := s.convertAmountToHex(userTx.Amount)
			return fmt.Sprintf("0x095ea7b3%s%s", s.padAddress(userTx.ToAddress), amountHex)
		case "transferFrom":
			// transferFrom(address,address,uint256) 函数选择器: 0x23b872dd
			amountHex := s.convertAmountToHex(userTx.Amount)
			return fmt.Sprintf("0x23b872dd%s%s%s", s.padAddress(userTx.FromAddress), s.padAddress(userTx.ToAddress), amountHex)
		}
	}

	// ETH转账，data为空
	return "0x"
}

// generateAccessList 生成AccessList
func (s *userTransactionService) generateAccessList(userTx *models.UserTransaction) string {
	// 如果是代币交易，生成AccessList
	if userTx.TransactionType == "token" && userTx.TokenContractAddress != "" {
		// 简化处理，返回空的AccessList
		// TODO: 实现完整的AccessList生成逻辑
		return "[]"
	}

	return "[]"
}

// padAddress 将地址填充为32字节
func (s *userTransactionService) padAddress(address string) string {
	// 移除0x前缀并填充到64个字符（32字节）
	cleanAddr := address
	if len(address) > 2 && address[:2] == "0x" {
		cleanAddr = address[2:]
	}
	return fmt.Sprintf("%064s", cleanAddr)
}

// convertAmountToHex 将金额转换为十六进制格式
func (s *userTransactionService) convertAmountToHex(amount string) string {
	// 将字符串转换为大整数，然后转换为十六进制
	// 数据库中存储的是整数格式的金额
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		// 如果转换失败，返回0
		return "0x0"
	}

	// 转换为十六进制并添加0x前缀
	hexStr := fmt.Sprintf("0x%s", amountBig.Text(16))
	return hexStr
}
