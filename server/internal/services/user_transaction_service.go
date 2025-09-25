package services

import (
	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/database"
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	solcommon "github.com/blocto/solana-go-sdk/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

// UserTransactionService 用户交易服务接口
type UserTransactionService interface {
	CreateTransaction(ctx context.Context, userID uint64, req *dto.CreateUserTransactionRequest) (*dto.UserTransactionResponse, error)
	GetTransactionByID(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
	GetUserTransactions(ctx context.Context, userID uint64, page, pageSize int, status string) (*dto.UserTransactionListResponse, error)
	GetUserTransactionsByChain(ctx context.Context, userID uint64, chain string, page, pageSize int, status string) (*dto.UserTransactionListResponse, error)
	UpdateTransaction(ctx context.Context, id uint, userID uint64, req *dto.UpdateUserTransactionRequest) (*dto.UserTransactionResponse, error)
	DeleteTransaction(ctx context.Context, id uint, userID uint64) error
	ExportTransaction(ctx context.Context, id uint, userID uint64, req *dto.ExportTransactionRequest) (*dto.ExportTransactionResponse, error)
	ImportSignature(ctx context.Context, id uint, userID uint64, req *dto.ImportSignatureRequest) (*dto.UserTransactionResponse, error)
	SendTransaction(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
	GetUserTransactionStats(ctx context.Context, userID uint64, chain string) (*dto.UserTransactionStatsResponse, error)

	// SOL 专用接口
	ExportSolUnsigned(ctx context.Context, id uint, userID uint64) (*dto.SolExportTransactionResponse, error)
	ImportSolSignature(ctx context.Context, id uint, userID uint64, req *dto.SolImportSignatureRequest) (*dto.UserTransactionResponse, error)
	SendSolTransaction(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error)
}

// userTransactionService 用户交易服务实现
type userTransactionService struct {
	userTxRepo       repository.UserTransactionRepository
	coinConfigRepo   repository.CoinConfigRepository
	parserConfigRepo repository.ParserConfigRepository
	logger           *logrus.Logger
	contractRepo     repository.ContractRepository
	userAddressRepo  repository.UserAddressRepository
}

// NewUserTransactionService 创建用户交易服务实例
func NewUserTransactionService() UserTransactionService {
	return &userTransactionService{
		userTxRepo:       repository.NewUserTransactionRepository(),
		coinConfigRepo:   repository.NewCoinConfigRepository(),
		parserConfigRepo: repository.NewParserConfigRepository(database.GetDB()),
		contractRepo:     repository.NewContractRepository(database.GetDB()),
		userAddressRepo:  repository.NewUserAddressRepository(database.GetDB()),
		logger:           logrus.New(),
	}
}

// ExportSolUnsigned 导出SOL未签名交易载荷
func (s *userTransactionService) ExportSolUnsigned(ctx context.Context, id uint, userID uint64) (*dto.SolExportTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(userTx.Chain, "sol") {
		return nil, fmt.Errorf("交易非SOL链")
	}
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()
	solClient, err := rpcManager.GetSolanaClient("sol")
	if err != nil {
		return nil, fmt.Errorf("获取Solana客户端失败: %w", err)
	}
	blockhash, err := solClient.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块哈希失败: %w", err)
	}
	resp := &dto.SolExportTransactionResponse{
		ID:              userTx.ID,
		Chain:           userTx.Chain,
		RecentBlockhash: blockhash,
		FeePayer:        userTx.FromAddress,
		Version:         "legacy",
		Instructions:    []map[string]any{},
		Context:         map[string]any{},
	}
	if strings.EqualFold(userTx.TransactionType, "token") && userTx.TokenContractAddress != "" {
		// 获取代币精度信息
		decimals := uint8(6) // 默认精度
		if contract, err := s.contractRepo.GetByAddress(ctx, userTx.TokenContractAddress); err == nil && contract != nil {
			decimals = contract.Decimals
		}

		// 处理发送者ATA地址
		_, _, _, err := s.processSenderATA(ctx, userTx.FromAddress, userTx.TokenContractAddress, userTx.Amount, decimals)
		if err != nil {
			return nil, err
		}

		// 根据合约操作类型生成不同的指令
		switch strings.ToLower(userTx.ContractOperationType) {
		case "approve":
			// 授权操作：授权 ToAddress 可以花费 FromAddress 的代币
			// 需要获取授权者的钱包地址（FromAddress 对应的钱包）
			authority := userTx.FromAddress
			if fromAddr, err := s.userAddressRepo.GetByAddress(userTx.FromAddress); err == nil && fromAddr != nil {
				if fromAddr.Type == "ata" && fromAddr.AtaOwnerAddress != "" {
					authority = fromAddr.AtaOwnerAddress
				}
			}

			// 被授权者地址（ToAddress）
			delegate := userTx.ToAddress

			resp.Instructions = append(resp.Instructions, map[string]any{
				"program_id": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
				"type":       "spl_approve",
				"params": map[string]any{
					"mint":      userTx.TokenContractAddress,
					"authority": authority, // 授权者钱包地址
					"delegate":  delegate,  // 被授权者地址
					"amount":    userTx.Amount,
					"decimals":  decimals,
				},
			})

		case "transferfrom":
			// 授权转账操作：使用授权额度进行转账
			// 发送者(资金来源)必须是持币人A：优先 allowance_address，其次 FromAddress 若为 ATA 则取其 AtaOwner，否则回退 FromAddress
			sourceOwner := userTx.AllowanceAddress
			if sourceOwner == "" {
				if fromAddr, err := s.userAddressRepo.GetByAddress(userTx.FromAddress); err == nil && fromAddr != nil && fromAddr.Type == "ata" && fromAddr.AtaOwnerAddress != "" {
					sourceOwner = fromAddr.AtaOwnerAddress
				} else {
					sourceOwner = userTx.FromAddress
				}
			}
			fromOwner, _, _, err := s.processSenderATA(ctx, sourceOwner, userTx.TokenContractAddress, userTx.Amount, decimals)
			if err != nil {
				return nil, err
			}

			// 处理接收者地址
			toOwner := userTx.ToAddress

			// 检查ToAddress是否是ATA地址，如果是则获取对应的钱包地址
			if toAddr, err := s.userAddressRepo.GetByAddress(userTx.ToAddress); err == nil && toAddr != nil {
				if toAddr.Type == "ata" && toAddr.AtaOwnerAddress != "" {
					toOwner = toAddr.AtaOwnerAddress
				}
			}

			// 检查接收者的ATA账户是否存在于链上
			needCreateATA := false
			if fromOwner != toOwner {
				// 直接使用数据库中的ToAddress（如果ToAddress是ATA地址）
				toATAAddress := userTx.ToAddress

				// 检查ToAddress是否是ATA类型
				if toAddr, err := s.userAddressRepo.GetByAddress(userTx.ToAddress); err == nil && toAddr != nil {
					if toAddr.Type == "ata" {
						// 通过RPC检查ATA账户是否存在于链上
						exists, err := s.checkATAExistsOnChain(ctx, toATAAddress)
						if err != nil {
							s.logger.Errorf("检查ATA账户存在性失败: %v", err)
							// 如果检查失败，为了安全起见，假设需要创建
							needCreateATA = true
						} else if !exists {
							// ATA账户在链上不存在，需要创建
							needCreateATA = true
							s.logger.Infof("ATA账户 %s 在链上不存在，需要创建", toATAAddress)
						} else {
							s.logger.Infof("ATA账户 %s 在链上已存在，无需创建", toATAAddress)
						}
					} else {
						// ToAddress不是ATA类型，不需要创建ATA账户
						s.logger.Infof("ToAddress %s 不是ATA类型，无需创建ATA账户", toATAAddress)
					}
				} else {
					// 无法查询到ToAddress信息，为了安全起见，假设需要创建
					s.logger.Warnf("无法查询ToAddress %s 的信息，假设需要创建ATA账户", toATAAddress)
					needCreateATA = true
				}
			}

			// 构建指令参数
			instructionParams := map[string]any{
				"mint":       userTx.TokenContractAddress, // Mint地址（代币合约地址）
				"from_owner": fromOwner,                   // 发送者钱包地址（持币人A）
				"to_owner":   toOwner,                     // 接收者钱包地址
				"amount":     userTx.Amount,
				"decimals":   decimals, // 添加代币精度信息
			}

			// 被授权者B作为authority（本次签名者/fee_payer），使用 from 字段
			instructionParams["delegate_auth"] = userTx.FromAddress

			// 如果需要创建接收者ATA账户，添加标记
			if needCreateATA {
				instructionParams["create"] = 1
			}

			resp.Instructions = append(resp.Instructions, map[string]any{
				"program_id": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
				"type":       "spl_transfer_checked",
				"params":     instructionParams,
			})

		case "transfer":
		default:
			// 普通转账操作：直接转账
			// 处理发送者ATA地址
			fromOwner, _, _, err := s.processSenderATA(ctx, userTx.FromAddress, userTx.TokenContractAddress, userTx.Amount, decimals)
			if err != nil {
				return nil, err
			}

			// 处理接收者地址
			toOwner := userTx.ToAddress

			// 检查ToAddress是否是ATA地址，如果是则获取对应的钱包地址
			if toAddr, err := s.userAddressRepo.GetByAddress(userTx.ToAddress); err == nil && toAddr != nil {
				if toAddr.Type == "ata" && toAddr.AtaOwnerAddress != "" {
					toOwner = toAddr.AtaOwnerAddress
				}
			}

			// 检查接收者的ATA账户是否存在于链上
			needCreateATA := false
			if fromOwner != toOwner {
				// 直接使用数据库中的ToAddress（如果ToAddress是ATA地址）
				toATAAddress := userTx.ToAddress

				// 检查ToAddress是否是ATA类型
				if toAddr, err := s.userAddressRepo.GetByAddress(userTx.ToAddress); err == nil && toAddr != nil {
					if toAddr.Type == "ata" {
						// 通过RPC检查ATA账户是否存在于链上
						exists, err := s.checkATAExistsOnChain(ctx, toATAAddress)
						if err != nil {
							s.logger.Errorf("检查ATA账户存在性失败: %v", err)
							// 如果检查失败，为了安全起见，假设需要创建
							needCreateATA = true
						} else if !exists {
							// ATA账户在链上不存在，需要创建
							needCreateATA = true
							s.logger.Infof("ATA账户 %s 在链上不存在，需要创建", toATAAddress)
						} else {
							s.logger.Infof("ATA账户 %s 在链上已存在，无需创建", toATAAddress)
						}
					} else {
						// ToAddress不是ATA类型，不需要创建ATA账户
						s.logger.Infof("ToAddress %s 不是ATA类型，无需创建ATA账户", toATAAddress)
					}
				} else {
					// 无法查询到ToAddress信息，为了安全起见，假设需要创建
					s.logger.Warnf("无法查询ToAddress %s 的信息，假设需要创建ATA账户", toATAAddress)
					needCreateATA = true
				}
			}

			// 构建指令参数
			instructionParams := map[string]any{
				"mint":       userTx.TokenContractAddress, // Mint地址（代币合约地址）
				"from_owner": fromOwner,                   // 发送者钱包地址
				"to_owner":   toOwner,                     // 接收者钱包地址
				"amount":     userTx.Amount,
				"decimals":   decimals, // 添加代币精度信息
			}

			// 普通转账无delegate

			// 如果需要创建接收者ATA账户，添加标记
			if needCreateATA {
				instructionParams["create"] = 1
			}

			resp.Instructions = append(resp.Instructions, map[string]any{
				"program_id": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
				"type":       "spl_transfer_checked",
				"params":     instructionParams,
			})
		}
	} else {
		resp.Instructions = append(resp.Instructions, map[string]any{
			"program_id": "11111111111111111111111111111111",
			"type":       "system_transfer",
			"params": map[string]any{
				"from":     userTx.FromAddress,
				"to":       userTx.ToAddress,
				"lamports": userTx.Amount,
			},
		})
	}
	// 将交易状态更新为 unsigned，表示已导出待签名
	userTx.Status = "unsigned"
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("更新交易状态失败: %w", err)
	}
	return resp, nil
}

// ImportSolSignature 导入SOL签名（base64）
func (s *userTransactionService) ImportSolSignature(ctx context.Context, id uint, userID uint64, req *dto.SolImportSignatureRequest) (*dto.UserTransactionResponse, error) {
	fmt.Printf("DEBUG: ImportSolSignature 服务层 - 交易ID: %d, 用户ID: %d, 签名数据长度: %d\n",
		id, userID, len(req.SignedBase))

	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		fmt.Printf("DEBUG: 获取交易失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: 交易链类型: %s\n", userTx.Chain)
	if !strings.EqualFold(userTx.Chain, "sol") {
		return nil, fmt.Errorf("交易非SOL链")
	}

	userTx.SignedTx = &req.SignedBase
	userTx.Status = "in_progress"

	fmt.Printf("DEBUG: 准备更新交易状态\n")
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		fmt.Printf("DEBUG: 更新交易失败: %v\n", err)
		return nil, fmt.Errorf("保存签名失败: %w", err)
	}

	fmt.Printf("DEBUG: 交易更新成功，准备发送交易\n")
	// 可选：立即发送
	return s.SendSolTransaction(ctx, id, userID)
}

// SendSolTransaction 发送SOL交易
func (s *userTransactionService) SendSolTransaction(ctx context.Context, id uint, userID uint64) (*dto.UserTransactionResponse, error) {
	userTx, err := s.userTxRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(userTx.Chain, "sol") {
		return nil, fmt.Errorf("交易非SOL链")
	}
	if userTx.SignedTx == nil || *userTx.SignedTx == "" {
		return nil, fmt.Errorf("缺少已签名交易base64")
	}
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()
	sendResp, err := rpcManager.SendTransaction(ctx, &utils.SendTransactionRequest{
		Chain:       userTx.Chain,
		SignedTx:    *userTx.SignedTx,
		FromAddress: userTx.FromAddress,
		ToAddress:   userTx.ToAddress,
		Amount:      userTx.Amount,
		Fee:         userTx.Fee,
	})
	if err != nil {
		return nil, fmt.Errorf("发送交易失败: %w", err)
	}
	if !sendResp.Success {
		return nil, fmt.Errorf("发送交易失败: %s", sendResp.Message)
	}
	userTx.Status = "in_progress"
	userTx.TxHash = &sendResp.TxHash
	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		return nil, fmt.Errorf("更新交易状态失败: %w", err)
	}
	return s.convertToResponse(userTx), nil
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
		AllowanceAddress:      req.AllowanceAddress,
	}

	// 如果为BTC，持久化原始交易关键字段
	if strings.ToLower(req.Chain) == "btc" {
		if req.BTCVersion != nil {
			userTx.BTCVersion = req.BTCVersion
		}
		if req.BTCLockTime != nil {
			userTx.BTCLockTime = req.BTCLockTime
		}
		if len(req.BTCTxIn) > 0 {
			if b, err := json.Marshal(req.BTCTxIn); err == nil {
				s := string(b)
				userTx.BTCTxInJSON = &s
			}
		}
		if len(req.BTCTxOut) > 0 {
			if b, err := json.Marshal(req.BTCTxOut); err == nil {
				s := string(b)
				userTx.BTCTxOutJSON = &s
			}
		}
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

	// 获取代币配置信息，用于填充代币精度
	tokenConfigs, err := s.getTokenConfigs(ctx)
	if err != nil {
		// 如果获取代币配置失败，记录错误但不影响交易列表返回
		// fmt.Printf("Warning: Failed to get token configs: %v\n", err)
	}

	// 转换为响应DTO
	var responses []dto.UserTransactionResponse
	for _, tx := range transactions {
		// 如果是代币交易，尝试获取代币精度信息
		if tx.TransactionType == "token" && tx.TokenContractAddress != "" {
			if config, exists := tokenConfigs[strings.ToLower(tx.TokenContractAddress)]; exists {
				tx.TokenName = config.Name
				// 转换类型：*uint -> *uint8
				decimals := uint8(config.Decimals)
				tx.TokenDecimals = &decimals
			}
		} else if tx.TransactionType == "coin" && tx.Symbol == "SOL" {
			// 对于SOL原生代币交易，设置正确的精度
			decimals := uint8(9) // SOL使用9位精度
			tx.TokenDecimals = &decimals
		}
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

// GetUserTransactionsByChain 根据链类型获取用户交易列表
func (s *userTransactionService) GetUserTransactionsByChain(ctx context.Context, userID uint64, chain string, page, pageSize int, status string) (*dto.UserTransactionListResponse, error) {
	transactions, total, err := s.userTxRepo.GetByChain(ctx, userID, chain, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取交易列表失败: %w", err)
	}

	// 如果指定了状态，进行过滤
	if status != "" {
		var filteredTransactions []*models.UserTransaction
		for _, tx := range transactions {
			if tx.Status == status {
				filteredTransactions = append(filteredTransactions, tx)
			}
		}
		transactions = filteredTransactions
		// 重新计算总数（这里简化处理，实际应该修改repository层支持状态过滤）
		total = int64(len(transactions))
	}

	// 获取代币配置信息，用于填充代币精度
	tokenConfigs, err := s.getTokenConfigs(ctx)
	if err != nil {
		// 如果获取代币配置失败，记录错误但不影响交易列表返回
		// fmt.Printf("Warning: Failed to get token configs: %v\n", err)
	}

	// 转换为响应DTO
	var responses []dto.UserTransactionResponse
	for _, tx := range transactions {
		// 如果是代币交易，尝试获取代币精度信息
		if tx.TransactionType == "token" && tx.TokenContractAddress != "" {
			if config, exists := tokenConfigs[strings.ToLower(tx.TokenContractAddress)]; exists {
				tx.TokenName = config.Name
				// 转换类型：*uint -> *uint8
				decimals := uint8(config.Decimals)
				tx.TokenDecimals = &decimals
			}
		} else if tx.TransactionType == "coin" && tx.Symbol == "SOL" {
			// 对于SOL原生代币交易，设置正确的精度
			decimals := uint8(9) // SOL使用9位精度
			tx.TokenDecimals = &decimals
		}
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

	// 更新基础字段
	if req.FromAddress != nil {
		userTx.FromAddress = *req.FromAddress
	}
	if req.ToAddress != nil {
		userTx.ToAddress = *req.ToAddress
	}
	if req.Amount != nil {
		userTx.Amount = *req.Amount
	}
	if req.Fee != nil {
		userTx.Fee = *req.Fee
	}
	if req.Remark != nil {
		userTx.Remark = *req.Remark
	}

	// 更新ETH相关字段
	if req.GasLimit != nil {
		userTx.GasLimit = req.GasLimit
	}
	if req.GasPrice != nil {
		userTx.GasPrice = req.GasPrice
	}
	if req.Nonce != nil {
		userTx.Nonce = req.Nonce
	}

	// 更新EIP-1559费率字段
	if req.MaxPriorityFeePerGas != nil {
		userTx.MaxPriorityFeePerGas = req.MaxPriorityFeePerGas
	}
	if req.MaxFeePerGas != nil {
		userTx.MaxFeePerGas = req.MaxFeePerGas
	}

	// 更新代币交易相关字段
	if req.TransactionType != nil {
		userTx.TransactionType = *req.TransactionType
	}
	if req.ContractOperationType != nil {
		userTx.ContractOperationType = *req.ContractOperationType
	}
	if req.TokenContractAddress != nil {
		userTx.TokenContractAddress = *req.TokenContractAddress
	}
	if req.AllowanceAddress != nil {
		userTx.AllowanceAddress = *req.AllowanceAddress
	}

	// 更新BTC特有字段
	if req.BTCVersion != nil {
		userTx.BTCVersion = req.BTCVersion
	}
	if req.BTCLockTime != nil {
		userTx.BTCLockTime = req.BTCLockTime
	}
	if len(req.BTCTxIn) > 0 {
		if b, err := json.Marshal(req.BTCTxIn); err == nil {
			s := string(b)
			userTx.BTCTxInJSON = &s
		}
	}
	if len(req.BTCTxOut) > 0 {
		if b, err := json.Marshal(req.BTCTxOut); err == nil {
			s := string(b)
			userTx.BTCTxOutJSON = &s
		}
	}

	// 更新状态相关字段
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

	// SOL 专用导出：返回未签名消息结构（用于前端/离线签名器组装与签名）
	if strings.EqualFold(userTx.Chain, "sol") {
		rpcManager := utils.NewRPCClientManager()
		defer rpcManager.Close()

		solClient, err := rpcManager.GetSolanaClient("sol")
		if err != nil {
			return nil, fmt.Errorf("获取Solana客户端失败: %w", err)
		}
		blockhash, err := solClient.GetLatestBlockhash(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取最新区块哈希失败: %w", err)
		}

		// 构造通用未签名交易负载（供签名器消费）
		// 注意：对于SPL代币，将使用 TokenContractAddress 作为 mint 传递
		solPayload := map[string]interface{}{
			"type":             "sol_unsigned",
			"version":          "legacy",
			"recent_blockhash": blockhash,
			"fee_payer":        userTx.FromAddress,
		}

		// 指令集合（由签名器据此生成真实指令并签名）
		if strings.EqualFold(userTx.TransactionType, "token") && userTx.TokenContractAddress != "" {
			// SPL Token 转账计划
			// 需要签名器根据 mint + from_owner + to_owner 计算 ATA
			instr := map[string]interface{}{
				"program_id": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
				"type":       "spl_transfer_checked",
				"params": map[string]interface{}{
					"mint":       userTx.TokenContractAddress,
					"from_owner": userTx.FromAddress,
					"to_owner":   userTx.ToAddress,
					"amount":     userTx.Amount, // 基于最小单位（按mint decimals）
					"decimals":   0,             // 由签名器查询/填充具体精度
				},
			}
			solPayload["instructions"] = []interface{}{instr}
		} else {
			// 原生 SOL 转账计划（lamports）
			instr := map[string]interface{}{
				"program_id": "11111111111111111111111111111111",
				"type":       "system_transfer",
				"params": map[string]interface{}{
					"from":     userTx.FromAddress,
					"to":       userTx.ToAddress,
					"lamports": userTx.Amount,
				},
			}
			solPayload["instructions"] = []interface{}{instr}
		}

		// 序列化为字符串，放入 UnsignedTx 字段，其他EVM特定字段置空
		b, _ := json.Marshal(solPayload)
		unsigned := string(b)

		userTx.Status = "unsigned"
		userTx.UnsignedTx = &unsigned
		if err := s.userTxRepo.Update(ctx, userTx); err != nil {
			return nil, fmt.Errorf("更新交易状态失败: %w", err)
		}

		return &dto.ExportTransactionResponse{
			UnsignedTx:           unsigned,
			Chain:                userTx.Chain,
			Symbol:               userTx.Symbol,
			FromAddress:          userTx.FromAddress,
			ToAddress:            userTx.ToAddress,
			Amount:               userTx.Amount,
			Fee:                  userTx.Fee,
			GasLimit:             nil,
			GasPrice:             nil,
			Nonce:                nil,
			MaxPriorityFeePerGas: nil,
			MaxFeePerGas:         nil,
			ChainID:              nil,
			TxData:               nil,
			AccessList:           nil,
		}, nil
	}

	// 检查是否可以导出
	if !userTx.CanExport() {
		return nil, errors.New("当前状态的交易无法导出")
	}

	// 如果交易已有hash，说明已经在途，需要检查是否已打包
	if userTx.TxHash != nil && *userTx.TxHash != "" {
		// 调用RPC检查交易是否已打包
		isPacked, err := s.checkTransactionPacked(ctx, userTx.Chain, *userTx.TxHash)
		if err != nil {
			return nil, fmt.Errorf("检查交易状态失败: %w", err)
		}

		if isPacked {
			// 交易已打包，更新数据库状态
			userTx.Status = "packed"
			userTx.UpdatedAt = time.Now()
			if err := s.userTxRepo.Update(ctx, userTx); err != nil {
				// fmt.Printf("更新交易状态失败: %v\n", err)
			}
			return nil, errors.New("此交易已经被打包上线，不能替换！")
		}
	}

	// 获取发送地址的当前nonce（如果交易中没有设置nonce）
	currentNonce := userTx.Nonce
	if currentNonce == nil {
		// 使用对应链的 pending nonce，避免与内存池未上链交易冲突
		nonce, err := s.getAddressNonceByChain(ctx, userTx.Chain, userTx.FromAddress)
		fmt.Printf("获取地址nonce: %v\n", nonce)
		if err != nil {
			// 如果获取nonce失败，使用默认值0
			// fmt.Printf("获取地址nonce失败: %v，使用默认值0\n", err)
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
	// fmt.Printf("🔍 费率设置调试信息:\n")
	// fmt.Printf("  req.MaxPriorityFeePerGas: %v\n", req.MaxPriorityFeePerGas)
	// fmt.Printf("  req.MaxFeePerGas: %v\n", req.MaxFeePerGas)
	// fmt.Printf("  userTx.MaxPriorityFeePerGas (before): %v\n", userTx.MaxPriorityFeePerGas)
	// fmt.Printf("  userTx.MaxFeePerGas (before): %v\n", userTx.MaxFeePerGas)

	if req.MaxPriorityFeePerGas != nil {
		// 前端传递的已经是Wei单位，直接使用
		userTx.MaxPriorityFeePerGas = req.MaxPriorityFeePerGas
		// fmt.Printf("  ✅ 使用请求中的 MaxPriorityFeePerGas: %s wei\n", *req.MaxPriorityFeePerGas)
	} else if userTx.MaxPriorityFeePerGas == nil {
		// 如果没有设置费率，使用默认值 2 Gwei = 2,000,000,000 wei
		defaultTip := "2000000000" // 2 Gwei in wei
		userTx.MaxPriorityFeePerGas = &defaultTip
		// fmt.Printf("  ⚠️ 使用默认 MaxPriorityFeePerGas: 2 Gwei -> %s wei\n", defaultTip)
	} else {
		// 数据库中已存在的值，检查是否需要从Gwei转换为Wei
		if s.isGweiValue(*userTx.MaxPriorityFeePerGas) {
			priorityFeeWei, err := s.convertGweiToWei(*userTx.MaxPriorityFeePerGas)
			if err == nil {
				userTx.MaxPriorityFeePerGas = &priorityFeeWei
				// fmt.Printf("  🔄 转换数据库中的 MaxPriorityFeePerGas: %s Gwei -> %s wei\n", *userTx.MaxPriorityFeePerGas, priorityFeeWei)
			}
		}
	}

	if req.MaxFeePerGas != nil {
		// 前端传递的已经是Wei单位，直接使用
		userTx.MaxFeePerGas = req.MaxFeePerGas
		// fmt.Printf("  ✅ 使用请求中的 MaxFeePerGas: %s wei\n", *req.MaxFeePerGas)
	} else if userTx.MaxFeePerGas == nil {
		// 如果没有设置费率，使用默认值 30 Gwei = 30,000,000,000 wei
		defaultFee := "30000000000" // 30 Gwei in wei
		userTx.MaxFeePerGas = &defaultFee
		// fmt.Printf("  ⚠️ 使用默认 MaxFeePerGas: 30 Gwei -> %s wei\n", defaultFee)
	} else {
		// 数据库中已存在的值，检查是否需要从Gwei转换为Wei
		if s.isGweiValue(*userTx.MaxFeePerGas) {
			maxFeeWei, err := s.convertGweiToWei(*userTx.MaxFeePerGas)
			if err == nil {
				userTx.MaxFeePerGas = &maxFeeWei
				// fmt.Printf("  🔄 转换数据库中的 MaxFeePerGas: %s Gwei -> %s wei\n", *userTx.MaxFeePerGas, maxFeeWei)
			}
		}
	}

	// fmt.Printf("  userTx.MaxPriorityFeePerGas (after): %v\n", userTx.MaxPriorityFeePerGas)
	// fmt.Printf("  userTx.MaxFeePerGas (after): %v\n", userTx.MaxFeePerGas)
	// fmt.Printf("开始进行估算GasLimit")
	// fmt.Printf("参数 查验 userTx.Chain = %s,userTx.GasLimit = %v \n", userTx.Chain, userTx.GasLimit)
	// 估算GasLimit（未设置时；EVM链；合约调用或代币交易）
	chainLower := strings.ToLower(userTx.Chain)
	if chainLower == "eth" || chainLower == "bsc" || chainLower == "binance" {
		// fmt.Printf("参数 查验 userTx.TransactionType %s\n", userTx.TransactionType)
		// ETH + token/合约调用 -> 估算；ETH 原生 -> 固定21000
		if userTx.TransactionType == "token" {
			rpcManager := utils.NewRPCClientManager()
			defer rpcManager.Close()

			value := big.NewInt(0)
			var dataBytes []byte
			if txData != "" && txData != "0x" {
				hexStr := strings.TrimPrefix(txData, "0x")
				if b, err := hex.DecodeString(hexStr); err == nil {
					dataBytes = b
				}
			}

			toForGas := userTx.ToAddress
			if userTx.TokenContractAddress != "" { // 代币调用时 To 是合约
				toForGas = userTx.TokenContractAddress
			}

			// fmt.Printf("🔍 估算Gas  txData: %+v\n", txData)

			var gas uint64
			var err error
			if chainLower == "eth" {
				gas, err = rpcManager.EstimateEthGas(ctx, userTx.FromAddress, toForGas, value, dataBytes)
			} else {
				gas, err = rpcManager.EstimateBscGas(ctx, userTx.FromAddress, toForGas, value, dataBytes)
			}
			if err == nil {
				gasWithBuffer := gas + gas/5
				gasU := uint(gasWithBuffer)
				userTx.GasLimit = &gasU
				// fmt.Printf("Gas估算成功: %d\n", gasU)
			} else {
				s.logger.Warnf("Gas估算失败，保持原值: %v", err)
			}
		} else {
			g := uint(21000)
			userTx.GasLimit = &g
			// fmt.Printf("Gas估算失败，保持原值: %d type=%s txData=%s\n", g, userTx.TransactionType, txData)
		}
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
	if userTx.Status != "unsigned" && userTx.Status != "in_progress" {
		return nil, errors.New("只有未签名或在途状态的交易才能导入签名")
	}

	// 如果交易已有hash，说明已经在途，需要检查是否已打包
	if userTx.TxHash != nil && *userTx.TxHash != "" {
		// 调用RPC检查交易是否已打包
		isPacked, err := s.checkTransactionPacked(ctx, userTx.Chain, *userTx.TxHash)
		if err != nil {
			return nil, fmt.Errorf("检查交易状态失败: %w", err)
		}

		if isPacked {
			// 交易已打包，更新数据库状态
			userTx.Status = "packed"
			userTx.UpdatedAt = time.Now()
			if err := s.userTxRepo.Update(ctx, userTx); err != nil {
				// fmt.Printf("更新交易状态失败: %v\n", err)
			}
			return nil, errors.New("此交易已经被打包上线，不能替换！")
		}
	}

	// 更新签名数据（对于SOL，SignedTx 预期为 base64 原始交易）
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
		// 发送失败，保存错误到数据库
		errorMsg := fmt.Sprintf("自动发送交易失败: %v", err)
		s.saveTransactionError(ctx, userTx, errorMsg)
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

	// 对于ETH交易，检查账户余额是否足够
	if strings.ToLower(userTx.Chain) == "eth" && userTx.TransactionType == "coin" {
		if err := s.validateEthBalance(ctx, userTx); err != nil {
			// 余额验证失败，保存错误到数据库
			errorMsg := fmt.Sprintf("余额验证失败: %v", err)
			s.saveTransactionError(ctx, userTx, errorMsg)
			return nil, fmt.Errorf("余额验证失败: %w", err)
		}
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
		Amount:      userTx.Amount, //代币交易时，Amount为0
		Fee:         userTx.Fee,
	}
	if strings.ToLower(userTx.Chain) == "eth" && userTx.TransactionType == "token" {
		sendReq.Amount = "0x0"
	}

	// 调用RPC发送交易
	sendResp, err := rpcManager.SendTransaction(ctx, sendReq)
	if err != nil {
		// RPC调用失败，保存错误到数据库
		errorMsg := fmt.Sprintf("RPC调用失败: %v", err)
		s.saveTransactionError(ctx, userTx, errorMsg)
		return nil, fmt.Errorf("发送交易失败: %w", err)
	}

	if !sendResp.Success {
		// 发送失败，保存错误到数据库
		errorMsg := fmt.Sprintf("交易发送失败: %s", sendResp.Message)
		s.saveTransactionError(ctx, userTx, errorMsg)
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
		// 查询失败，保存错误信息到数据库
		errorMsg := fmt.Sprintf("查询交易状态失败: %v", err)
		s.saveTransactionError(ctx, userTx, errorMsg)
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
		// 查询失败，保存错误信息到数据库
		errorMsg := fmt.Sprintf("查询交易状态失败: %v", err)
		s.saveTransactionError(ctx, userTx, errorMsg)
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
func (s *userTransactionService) GetUserTransactionStats(ctx context.Context, userID uint64, chain string) (*dto.UserTransactionStatsResponse, error) {
	// 获取各种状态的交易数量
	statuses := []string{"draft", "unsigned", "in_progress", "packed", "confirmed", "failed"}

	stats := &dto.UserTransactionStatsResponse{}

	for _, status := range statuses {
		transactions, err := s.userTxRepo.GetByStatus(ctx, userID, status)
		if err != nil {
			continue
		}

		// 如果指定了链类型，过滤出对应链的交易
		var filteredTransactions []*models.UserTransaction
		if chain != "" {
			for _, tx := range transactions {
				if strings.EqualFold(tx.Chain, chain) {
					filteredTransactions = append(filteredTransactions, tx)
				}
			}
		} else {
			filteredTransactions = transactions
		}

		count := int64(len(filteredTransactions))
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
		AllowanceAddress:      userTx.AllowanceAddress,
		TokenName:             userTx.TokenName,
		TokenDecimals:         userTx.TokenDecimals,

		// QR码导出相关字段
		ChainID:    userTx.ChainID,
		TxData:     userTx.TxData,
		AccessList: userTx.AccessList,

		// 签名组件
		V: userTx.V,
		R: userTx.R,
		S: userTx.S,

		// BTC特有字段
		BTCVersion:   userTx.BTCVersion,
		BTCLockTime:  userTx.BTCLockTime,
		BTCTxInJSON:  userTx.BTCTxInJSON,
		BTCTxOutJSON: userTx.BTCTxOutJSON,
	}
}

// getTokenConfigs 获取代币配置信息
func (s *userTransactionService) getTokenConfigs(ctx context.Context) (map[string]*models.CoinConfig, error) {
	// 获取所有代币配置
	configs, err := s.coinConfigRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取代币配置失败: %w", err)
	}

	// 构建代币地址到配置的映射
	tokenMap := make(map[string]*models.CoinConfig)
	for _, config := range configs {
		if config.ContractAddr != "" {
			// 使用小写地址作为key，确保匹配
			address := strings.ToLower(config.ContractAddr)
			tokenMap[address] = config
		}
	}

	return tokenMap, nil
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
func (s *userTransactionService) getAddressNonceByChain(ctx context.Context, chain string, address string) (uint64, error) {
	chainLower := strings.ToLower(chain)
	switch chainLower {
	case "eth", "ethereum":
		// ETH：使用 pending nonce，便于连续发多笔
		if _, ok := config.AppConfig.Blockchain.Chains[chainLower]; !ok {
			return 0, fmt.Errorf("未配置%s RPC URL", chainLower)
		}
		fo, err := utils.NewEthFailoverFromChain(chainLower)
		if err != nil {
			return 0, fmt.Errorf("初始化%s故障转移失败: %w", strings.ToUpper(chainLower), err)
		}
		defer fo.Close()
		nonce, err := fo.PendingNonceAt(ctx, common.HexToAddress(address))
		if err != nil {
			return 0, fmt.Errorf("获取地址pending nonce失败: %w", err)
		}
		return nonce, nil
	case "bsc", "binance":
		// BSC：优先使用 latest（避免历史未清理的 pending 抬高nonce）
		if _, ok := config.AppConfig.Blockchain.Chains[chainLower]; !ok {
			return 0, fmt.Errorf("未配置%s RPC URL", chainLower)
		}
		fo, err := utils.NewEthFailoverFromChain(chainLower)
		if err != nil {
			return 0, fmt.Errorf("初始化%s故障转移失败: %w", strings.ToUpper(chainLower), err)
		}
		defer fo.Close()
		nonce, err := fo.NonceAt(ctx, common.HexToAddress(address), nil)
		if err != nil {
			return 0, fmt.Errorf("获取地址latest nonce失败: %w", err)
		}
		return nonce, nil
	default:
		return 0, fmt.Errorf("不支持的链获取nonce: %s", chain)
	}
}

// generateTxData 生成交易数据（十六进制格式）
func (s *userTransactionService) generateTxData(userTx *models.UserTransaction) string {
	// 根据交易类型生成不同的数据
	if userTx.TransactionType == "token" && userTx.TokenContractAddress != "" {
		// 使用parser_config表的配置动态生成交易数据
		return s.generateContractCallData(userTx)
	}

	// ETH转账，data为空
	return "0x"
}

// generateContractCallData 根据parser_config配置生成合约调用数据
func (s *userTransactionService) generateContractCallData(userTx *models.UserTransaction) string {
	// 获取parser_config配置
	config, err := s.getParserConfigByOperation(context.Background(), userTx.ContractOperationType)
	if err != nil {
		s.logger.Errorf("获取parser_config失败: %v", err)
		return "0x"
	}

	// 构建交易数据
	data := config.FunctionSignature // 函数选择器

	// 根据参数配置添加参数
	for _, param := range config.ParamConfig {
		var paramValue string
		switch param.Name {
		case "to":
			paramValue = s.padAddress(userTx.ToAddress)
		case "from":
			// 对于transferFrom操作，from参数应该是代币持有者地址（allowance_address）
			if userTx.ContractOperationType == "transferFrom" && userTx.AllowanceAddress != "" {
				paramValue = s.padAddress(userTx.AllowanceAddress)
			} else {
				paramValue = s.padAddress(userTx.FromAddress)
			}
		case "owner":
			paramValue = s.padAddress(userTx.FromAddress)
		case "spender":
			paramValue = s.padAddress(userTx.ToAddress)
		case "value":
			paramValue = s.convertAmountToHex(userTx.Amount)
			// 去掉0x前缀
			paramValue = strings.TrimPrefix(paramValue, "0x")
		case "wad":
			paramValue = s.convertAmountToHex(userTx.Amount)
			// 去掉0x前缀
			paramValue = strings.TrimPrefix(paramValue, "0x")
		default:
			s.logger.Warnf("未知参数名: %s", param.Name)
			continue
		}

		// 确保参数长度正确
		if len(paramValue) < param.Length*2 { // 每个字节2个十六进制字符
			paramValue = strings.Repeat("0", param.Length*2-len(paramValue)) + paramValue
		}

		data += paramValue
	}

	return data
}

// getParserConfigByOperation 根据操作类型获取parser_config配置
func (s *userTransactionService) getParserConfigByOperation(ctx context.Context, operationType string) (*models.ParserConfig, error) {
	// 从数据库查询parser_config配置
	config, err := s.parserConfigRepo.GetByFunctionName(ctx, operationType)
	if err != nil {
		return nil, fmt.Errorf("查询parser_config失败: %w", err)
	}

	if config == nil {
		return nil, fmt.Errorf("未找到操作类型 %s 的parser_config配置", operationType)
	}

	return config, nil
}

// generateAccessList 生成AccessList
func (s *userTransactionService) generateAccessList(userTx *models.UserTransaction) string {
	// 如果是代币交易，生成AccessList
	if userTx.TransactionType == "token" && userTx.TokenContractAddress != "" {
		accessList := s.generateAccessListForTokenTransfer(userTx)
		if len(accessList) == 0 {
			return "[]"
		}

		// 转换为JSON字符串
		jsonData, err := json.Marshal(accessList)
		if err != nil {
			s.logger.Errorf("序列化AccessList失败: %v", err)
			return "[]"
		}

		return string(jsonData)
	}

	return "[]"
}

// generateAccessListForTokenTransfer 为代币转账生成AccessList
func (s *userTransactionService) generateAccessListForTokenTransfer(userTx *models.UserTransaction) []map[string]interface{} {
	if userTx.TokenContractAddress == "" {
		return nil
	}

	accessList := []map[string]interface{}{}

	// 根据合约操作类型生成不同的AccessList
	switch userTx.ContractOperationType {
	case "transfer":
		// 标准transfer操作，通常只需要访问余额存储槽
		accessList = append(accessList, map[string]interface{}{
			"address": userTx.TokenContractAddress,
			"storageKeys": []string{
				// 发送者余额存储槽 (keccak256(abi.encodePacked(sender, balanceOf_slot)))
				s.calculateStorageSlot(userTx.FromAddress, "0x0000000000000000000000000000000000000000000000000000000000000002"),
				// 接收者余额存储槽
				s.calculateStorageSlot(userTx.ToAddress, "0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
		})

	case "approve":
		// approve操作，需要访问allowance存储槽
		accessList = append(accessList, map[string]interface{}{
			"address": userTx.TokenContractAddress,
			"storageKeys": []string{
				// allowance存储槽 (keccak256(abi.encodePacked(owner, spender, allowance_slot)))
				s.calculateAllowanceStorageSlot(userTx.FromAddress, userTx.ToAddress, "0x0000000000000000000000000000000000000000000000000000000000000003"),
			},
		})

	case "transferFrom":
		// transferFrom操作，需要访问发送者、接收者余额和allowance
		accessList = append(accessList, map[string]interface{}{
			"address": userTx.TokenContractAddress,
			"storageKeys": []string{
				// 发送者余额
				s.calculateStorageSlot(userTx.FromAddress, "0x0000000000000000000000000000000000000000000000000000000000000002"),
				// 接收者余额
				s.calculateStorageSlot(userTx.ToAddress, "0x0000000000000000000000000000000000000000000000000000000000000002"),
				// allowance
				s.calculateAllowanceStorageSlot(userTx.FromAddress, userTx.ToAddress, "0x0000000000000000000000000000000000000000000000000000000000000003"),
			},
		})

	default:
		// 其他操作类型，不添加AccessList
		return nil
	}

	return accessList
}

// calculateStorageSlot 计算存储槽位置
func (s *userTransactionService) calculateStorageSlot(address, slot string) string {
	// 移除0x前缀
	cleanAddr := strings.TrimPrefix(address, "0x")
	cleanSlot := strings.TrimPrefix(slot, "0x")

	// 填充地址到64个字符
	paddedAddr := fmt.Sprintf("%064s", cleanAddr)

	// 拼接地址和槽位
	combined := paddedAddr + cleanSlot

	// 计算keccak256哈希
	hashBytes := crypto.Keccak256([]byte(combined))

	return "0x" + hex.EncodeToString(hashBytes)
}

// calculateAllowanceStorageSlot 计算allowance存储槽位置
func (s *userTransactionService) calculateAllowanceStorageSlot(owner, spender, slot string) string {
	// 移除0x前缀
	cleanOwner := strings.TrimPrefix(owner, "0x")
	cleanSpender := strings.TrimPrefix(spender, "0x")
	cleanSlot := strings.TrimPrefix(slot, "0x")

	// 填充地址到64个字符
	paddedOwner := fmt.Sprintf("%064s", cleanOwner)
	paddedSpender := fmt.Sprintf("%064s", cleanSpender)

	// 拼接owner、spender和槽位
	combined := paddedOwner + paddedSpender + cleanSlot

	// 计算keccak256哈希
	hashBytes := crypto.Keccak256([]byte(combined))

	return "0x" + hex.EncodeToString(hashBytes)
}

// padAddress 将地址填充为32字节
func (s *userTransactionService) padAddress(address string) string {
	// 移除0x前缀并填充到64个字符（32字节）
	cleanAddr := strings.TrimPrefix(address, "0x")
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

// convertGweiToWei 将Gwei转换为Wei
func (s *userTransactionService) convertGweiToWei(gweiStr string) (string, error) {
	// 解析Gwei值
	gweiBig, ok := new(big.Int).SetString(gweiStr, 10)
	if !ok {
		return "", fmt.Errorf("无效的Gwei值: %s", gweiStr)
	}

	// 1 Gwei = 10^9 Wei
	weiMultiplier := big.NewInt(1000000000) // 10^9
	weiBig := new(big.Int).Mul(gweiBig, weiMultiplier)

	return weiBig.String(), nil
}

// isGweiValue 判断值是否为Gwei单位（小于10^9的值通常是Gwei）
func (s *userTransactionService) isGweiValue(valueStr string) bool {
	valueBig, ok := new(big.Int).SetString(valueStr, 10)
	if !ok {
		return false
	}

	// 如果值小于10^9，很可能是Gwei单位
	// 典型的Gwei值范围：1-1000 Gwei
	gweiThreshold := big.NewInt(1000000000) // 10^9
	return valueBig.Cmp(gweiThreshold) < 0
}

// saveTransactionError 保存交易错误到数据库
func (s *userTransactionService) saveTransactionError(ctx context.Context, userTx *models.UserTransaction, errorMsg string) {
	userTx.Status = "failed"
	userTx.ErrorMsg = &errorMsg

	if err := s.userTxRepo.Update(ctx, userTx); err != nil {
		s.logger.Errorf("更新交易状态失败: %v", err)
	}

	s.logger.Errorf("交易失败: %s", errorMsg)
}

// validateEthBalance 验证ETH账户余额是否足够支付交易
func (s *userTransactionService) validateEthBalance(ctx context.Context, userTx *models.UserTransaction) error {
	// 获取账户余额
	fo, err := utils.NewEthFailoverFromChain("eth")
	if err != nil {
		return fmt.Errorf("初始化ETH故障转移失败: %w", err)
	}
	defer fo.Close()

	balance, err := fo.BalanceAt(ctx, common.HexToAddress(userTx.FromAddress), nil)
	if err != nil {
		return fmt.Errorf("获取账户余额失败: %w", err)
	}

	// 计算交易金额
	amountBig, ok := new(big.Int).SetString(userTx.Amount, 10)
	if !ok {
		return fmt.Errorf("无效的交易金额: %s", userTx.Amount)
	}

	// 计算Gas费用
	var gasCost *big.Int
	if userTx.GasLimit != nil && userTx.MaxFeePerGas != nil {
		// EIP-1559交易：使用MaxFeePerGas
		maxFeeBig, ok := new(big.Int).SetString(*userTx.MaxFeePerGas, 10)
		if !ok {
			return fmt.Errorf("无效的MaxFeePerGas: %s", *userTx.MaxFeePerGas)
		}
		gasCost = new(big.Int).Mul(maxFeeBig, big.NewInt(int64(*userTx.GasLimit)))
	} else if userTx.GasLimit != nil && userTx.GasPrice != nil {
		// Legacy交易：使用GasPrice
		gasPriceBig, ok := new(big.Int).SetString(*userTx.GasPrice, 10)
		if !ok {
			return fmt.Errorf("无效的GasPrice: %s", *userTx.GasPrice)
		}
		gasCost = new(big.Int).Mul(gasPriceBig, big.NewInt(int64(*userTx.GasLimit)))
	} else {
		return fmt.Errorf("缺少Gas费用信息")
	}

	// 计算总成本：交易金额 + Gas费用
	totalCost := new(big.Int).Add(amountBig, gasCost)

	// 检查余额是否足够
	if balance.Cmp(totalCost) < 0 {
		// 转换wei到ETH用于显示
		balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
		totalCostEth := new(big.Float).Quo(new(big.Float).SetInt(totalCost), big.NewFloat(1e18))
		shortfall := new(big.Int).Sub(totalCost, balance)
		shortfallEth := new(big.Float).Quo(new(big.Float).SetInt(shortfall), big.NewFloat(1e18))

		return fmt.Errorf("余额不足: 当前余额 %.6f ETH, 需要 %.6f ETH, 缺少 %.6f ETH",
			balanceEth, totalCostEth, shortfallEth)
	}

	s.logger.Infof("余额验证通过: 余额=%s wei, 交易金额=%s wei, Gas费用=%s wei, 总成本=%s wei",
		balance.String(), amountBig.String(), gasCost.String(), totalCost.String())

	return nil
}

// calculateATAAddress 计算Associated Token Account地址
func (s *userTransactionService) calculateATAAddress(ownerAddress, mintAddress string) string {
	// 使用Solana SDK计算ATA地址
	ownerPubKey := solcommon.PublicKeyFromString(ownerAddress)
	mintPubKey := solcommon.PublicKeyFromString(mintAddress)

	ataAddress, _, err := solcommon.FindAssociatedTokenAddress(ownerPubKey, mintPubKey)
	if err != nil {
		s.logger.Errorf("计算ATA地址失败: %v", err)
		return ""
	}

	return ataAddress.String()
}

// checkATAExistsOnChain 通过RPC检查ATA账户是否存在于链上
func (s *userTransactionService) checkATAExistsOnChain(ctx context.Context, ataAddress string) (bool, error) {
	// 创建RPC客户端管理器
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()

	solClient, err := rpcManager.GetSolanaClient("sol")
	if err != nil {
		return false, fmt.Errorf("获取Solana客户端失败: %w", err)
	}

	// 使用getAccountInfo检查账户是否存在
	accountInfo, err := solClient.GetAccountInfo(ctx, ataAddress)
	if err != nil {
		return false, fmt.Errorf("查询账户信息失败: %w", err)
	}

	// 如果accountInfo不为nil，说明账户存在
	return accountInfo != nil, nil
}

// getTokenBalance 获取代币账户余额
func (s *userTransactionService) getTokenBalance(ctx context.Context, ataAddress string) (*big.Int, error) {
	// 创建RPC客户端管理器
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()

	solClient, err := rpcManager.GetSolanaClient("sol")
	if err != nil {
		return nil, fmt.Errorf("获取Solana客户端失败: %w", err)
	}

	// 获取账户信息
	accountInfo, err := solClient.GetAccountInfo(ctx, ataAddress)
	if err != nil {
		return nil, fmt.Errorf("查询代币账户信息失败: %w", err)
	}

	if accountInfo == nil {
		return big.NewInt(0), nil // 账户不存在，余额为0
	}

	// 解析代币账户数据获取余额
	// 代币账户数据格式：mint(32) + owner(32) + amount(8) + delegate(32) + state(1) + ...
	if len(accountInfo.Data) == 0 {
		return nil, fmt.Errorf("代币账户数据为空")
	}

	// 解码base64数据
	dataBytes, err := base64.StdEncoding.DecodeString(accountInfo.Data[0])
	if err != nil {
		return nil, fmt.Errorf("解码代币账户数据失败: %w", err)
	}

	if len(dataBytes) < 72 {
		return nil, fmt.Errorf("代币账户数据格式错误，长度不足")
	}

	// 提取余额（第64-72字节，8字节的uint64）
	amountBytes := dataBytes[64:72]
	amount := binary.LittleEndian.Uint64(amountBytes)

	return big.NewInt(int64(amount)), nil
}

// processSenderATA 处理发送者ATA地址逻辑
func (s *userTransactionService) processSenderATA(ctx context.Context, fromAddress, mintAddress, amount string, decimals uint8) (fromOwner, fromATAAddress string, needCreateFromATA bool, err error) {
	// 检查发送者地址类型和ATA账户状态
	fromOwner = fromAddress
	fromATAAddress = fromAddress
	needCreateFromATA = false

	// 查询发送者地址信息
	fromAddr, err := s.userAddressRepo.GetByAddress(fromAddress)
	if err != nil {
		s.logger.Errorf("查询发送者地址信息失败: %v", err)
		return "", "", false, fmt.Errorf("查询发送者地址信息失败: %w", err)
	}

	if fromAddr != nil {
		if fromAddr.Type == "ata" {
			// 如果FromAddress是ATA地址，直接使用（后续仍做链上检查，但在错误时保守处理为存在）
			fromATAAddress = fromAddress
			fromOwner = fromAddr.AtaOwnerAddress
			s.logger.Infof("发送者地址是ATA地址: %s, 所属钱包: %s", fromATAAddress, fromOwner)
		} else {
			// 如果FromAddress是钱包地址，计算对应的ATA地址
			fromOwner = fromAddress
			fromATAAddress = s.calculateATAAddress(fromOwner, mintAddress)
			if fromATAAddress == "" {
				return "", "", false, fmt.Errorf("计算发送者ATA地址失败")
			}
			s.logger.Infof("发送者地址是钱包地址: %s, 计算ATA地址: %s", fromOwner, fromATAAddress)
		}
	} else {
		// 如果数据库中没有记录，假设FromAddress是钱包地址
		fromOwner = fromAddress
		fromATAAddress = s.calculateATAAddress(fromOwner, mintAddress)
		if fromATAAddress == "" {
			return "", "", false, fmt.Errorf("计算发送者ATA地址失败")
		}
		s.logger.Infof("数据库中没有发送者地址记录，假设是钱包地址: %s, 计算ATA地址: %s", fromOwner, fromATAAddress)
	}

	// 检查发送者ATA账户是否存在
	fromATAExists, err := s.checkATAExistsOnChain(ctx, fromATAAddress)
	if err != nil {
		// 保守策略：RPC 查询失败时，不创建（避免对已存在账户重复创建）
		s.logger.Warnf("检查发送者ATA账户存在性失败(保守假定存在): %v", err)
		fromATAExists = true
	}
	needCreateFromATA = !fromATAExists
	if needCreateFromATA {
		s.logger.Infof("发送者ATA账户不存在，将在交易中创建: %s", fromATAAddress)
	} else {
		s.logger.Infof("发送者ATA账户已存在: %s", fromATAAddress)
	}

	// 如果发送者ATA账户存在，检查代币余额
	if fromATAExists {
		balance, err := s.getTokenBalance(ctx, fromATAAddress)
		if err != nil {
			s.logger.Errorf("获取发送者代币余额失败: %v", err)
			return "", "", false, fmt.Errorf("获取发送者代币余额失败: %w", err)
		}

		// 解析转账金额
		amountBig, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return "", "", false, fmt.Errorf("无效的转账金额: %s", amount)
		}

		// 检查余额是否足够
		if balance.Cmp(amountBig) < 0 {
			// 转换余额和金额用于显示
			balanceFloat := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(math.Pow10(int(decimals))))
			amountFloat := new(big.Float).Quo(new(big.Float).SetInt(amountBig), big.NewFloat(math.Pow10(int(decimals))))
			return "", "", false, fmt.Errorf("代币余额不足: 当前余额 %s, 需要 %s", balanceFloat.Text('f', int(decimals)), amountFloat.Text('f', int(decimals)))
		}
	}

	return fromOwner, fromATAAddress, needCreateFromATA, nil
}

// checkTransactionPacked 检查交易是否已打包
func (s *userTransactionService) checkTransactionPacked(ctx context.Context, chain, txHash string) (bool, error) {
	// 创建RPC客户端管理器
	rpcManager := utils.NewRPCClientManager()
	defer rpcManager.Close()

	// 调用RPC获取交易状态
	statusResp, err := rpcManager.GetTransactionStatus(ctx, chain, txHash)
	if err != nil {
		return false, fmt.Errorf("获取交易状态失败: %w", err)
	}

	// 检查交易是否已确认（有区块高度说明已打包）
	return statusResp.BlockHeight > 0, nil
}
