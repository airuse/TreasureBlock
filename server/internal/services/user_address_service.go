package services

import (
	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// UserAddressService 用户地址服务接口
type UserAddressService interface {
	CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error)
	GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error)
	UpdateAddress(userID uint, addressID uint, req *dto.UpdateUserAddressRequest) (*dto.UserAddressResponse, error)
	DeleteAddress(userID uint, addressID uint) error
	GetAddressByID(userID uint, addressID uint) (*dto.UserAddressResponse, error)
	GetAddressTransactions(userID uint, address string, page, pageSize int, chain string) (*dto.AddressTransactionsResponse, error)
}

// userAddressService 用户地址服务实现
type userAddressService struct {
	userAddressRepo repository.UserAddressRepository
	blockRepo       repository.BlockRepository
	transactionRepo repository.TransactionRepository
}

// NewUserAddressService 创建用户地址服务
func NewUserAddressService(userAddressRepo repository.UserAddressRepository, blockRepo repository.BlockRepository) UserAddressService {
	return &userAddressService{
		userAddressRepo: userAddressRepo,
		blockRepo:       blockRepo,
		transactionRepo: repository.NewTransactionRepository(),
	}
}

// CreateAddress 创建用户地址
func (s *userAddressService) CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error) {
	// 验证地址格式
	if !s.isValidAddress(req.Address) {
		return nil, errors.New("无效的地址格式")
	}

	// 检查地址是否已存在
	existingAddress, err := s.userAddressRepo.GetByAddress(req.Address)
	if err == nil && existingAddress != nil {
		return nil, errors.New("该地址已存在")
	}

	// 自动获取当前区块高度和地址余额
	createdHeight, balance, err := s.getCurrentBlockHeightAndBalance(req.Address)
	if err != nil {
		// 如果获取失败，记录错误日志，使用默认值，但不影响地址创建
		fmt.Printf("警告：获取区块高度和余额失败: %v，使用默认值\n", err)
		createdHeight = 0
		balance = 0
	}

	// 创建新地址
	userAddress := &models.UserAddress{
		UserID:           userID,
		Address:          req.Address,
		Label:            req.Label,
		Type:             req.Type,
		Balance:          balance,
		TransactionCount: 0,
		IsActive:         true,
		CreatedHeight:    createdHeight,
	}

	if err := s.userAddressRepo.Create(userAddress); err != nil {
		return nil, err
	}

	return s.convertToResponse(userAddress), nil
}

// GetUserAddresses 获取用户的所有地址
func (s *userAddressService) GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error) {
	addresses, err := s.userAddressRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.UserAddressResponse
	for _, addr := range addresses {
		if addr.Type == "wallet" {
			// 动态计算余额：从创建高度起统计交易收支
			inWei, outWei, sumErr := s.transactionRepo.ComputeEthSumsWei(context.Background(), addr.Address, addr.CreatedHeight)
			if sumErr == nil {
				dynamicBalance := s.computeEthDynamicBalance(addr.Balance, inWei, outWei)
				resp := s.convertToResponse(&addr)
				resp.Balance = dynamicBalance
				responses = append(responses, *resp)
			}
		} else {
			responses = append(responses, *s.convertToResponse(&addr))
		}
	}

	return responses, nil
}

// UpdateAddress 更新用户地址
func (s *userAddressService) UpdateAddress(userID uint, addressID uint, req *dto.UpdateUserAddressRequest) (*dto.UserAddressResponse, error) {
	// 获取地址
	address, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return nil, errors.New("地址不存在")
	}

	// 验证权限
	if address.UserID != userID {
		return nil, errors.New("无权限操作此地址")
	}

	// 更新字段
	if req.Label != nil {
		address.Label = *req.Label
	}
	if req.Type != nil {
		address.Type = *req.Type
	}
	if req.IsActive != nil {
		address.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.userAddressRepo.Update(address); err != nil {
		return nil, err
	}

	return s.convertToResponse(address), nil
}

// DeleteAddress 删除用户地址
func (s *userAddressService) DeleteAddress(userID uint, addressID uint) error {
	// 获取地址
	address, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return errors.New("地址不存在")
	}

	// 验证权限
	if address.UserID != userID {
		return errors.New("无权限操作此地址")
	}

	return s.userAddressRepo.Delete(addressID)
}

// GetAddressByID 根据ID获取地址
func (s *userAddressService) GetAddressByID(userID uint, addressID uint) (*dto.UserAddressResponse, error) {
	address, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return nil, errors.New("地址不存在")
	}

	// 验证权限
	if address.UserID != userID {
		return nil, errors.New("无权限查看此地址")
	}

	return s.convertToResponse(address), nil
}

// isValidAddress 验证地址格式
func (s *userAddressService) isValidAddress(address string) bool {
	// 简单的以太坊地址验证
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	if len(address) != 42 {
		return false
	}
	// 可以添加更多验证逻辑
	return true
}

// convertToResponse 转换为响应DTO
func (s *userAddressService) convertToResponse(address *models.UserAddress) *dto.UserAddressResponse {
	return &dto.UserAddressResponse{
		ID:               address.ID,
		Address:          address.Address,
		Label:            address.Label,
		Type:             address.Type,
		Balance:          address.Balance,
		TransactionCount: address.TransactionCount,
		IsActive:         address.IsActive,
		CreatedHeight:    address.CreatedHeight,
		CreatedAt:        address.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        address.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// getCurrentBlockHeightAndBalance 获取当前区块高度和地址余额
// 优化：使用数据库中的最新区块高度，节省一次RPC调用
func (s *userAddressService) getCurrentBlockHeightAndBalance(address string) (uint64, float64, error) {
	ctx := context.Background()

	// 从数据库获取ETH最新区块高度（节省一次RPC调用）
	latestBlock, err := s.blockRepo.GetLatest(ctx, "eth")
	if err != nil {
		return 0, 0, fmt.Errorf("获取最新区块高度失败: %w", err)
	}

	blockNumber := latestBlock.Height

	// 从配置文件获取ETH RPC URL
	chainConfig, exists := config.AppConfig.Blockchain.Chains["eth"]
	if !exists || chainConfig.RPCURL == "" {
		return blockNumber, 0, fmt.Errorf("未配置ETH RPC URL")
	}

	// 连接ETH客户端
	client, err := ethclient.Dial(chainConfig.RPCURL)
	if err != nil {
		return blockNumber, 0, fmt.Errorf("连接ETH RPC失败: %w", err)
	}
	defer client.Close()

	// 获取地址余额 (使用数据库中的最新区块高度)
	balance, err := client.BalanceAt(ctx, common.HexToAddress(address), big.NewInt(int64(blockNumber)))
	if err != nil {
		return blockNumber, 0, fmt.Errorf("获取地址余额失败: %w", err)
	}

	// 将Wei转换为ETH (1 ETH = 10^18 Wei)
	balanceFloat := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))

	// 转换为float64
	balanceFloat64, _ := balanceFloat.Float64()

	return blockNumber, balanceFloat64, nil
}

// GetAddressTransactions 获取地址相关的交易列表
func (s *userAddressService) GetAddressTransactions(userID uint, address string, page, pageSize int, chain string) (*dto.AddressTransactionsResponse, error) {
	// 验证用户是否有权限查看该地址
	userAddress, err := s.userAddressRepo.GetByAddress(address)
	if err != nil {
		return nil, fmt.Errorf("地址不存在")
	}

	// 验证权限：只有地址所有者才能查看交易
	if userAddress.UserID != userID {
		return nil, fmt.Errorf("无权限查看此地址的交易")
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取交易列表
	transactions, total, err := s.transactionRepo.GetByAddress(context.Background(), address, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取交易列表失败: %w", err)
	}

	// 转换为响应DTO
	var responses []dto.AddressTransactionResponse
	for _, tx := range transactions {
		// 如果指定了链类型，只返回对应链的交易
		if chain != "" && tx.Chain != chain {
			continue
		}

		response := dto.AddressTransactionResponse{
			ID:                   tx.ID,
			TxID:                 tx.TxID,
			Height:               tx.Height,
			BlockIndex:           tx.BlockIndex,
			AddressFrom:          tx.AddressFrom,
			AddressTo:            tx.AddressTo,
			Amount:               tx.Amount,
			GasLimit:             tx.GasLimit,
			GasPrice:             tx.GasPrice,
			GasUsed:              tx.GasUsed,
			MaxFeePerGas:         tx.MaxFeePerGas,
			MaxPriorityFeePerGas: tx.MaxPriorityFeePerGas,
			EffectiveGasPrice:    tx.EffectiveGasPrice,
			Fee:                  tx.Fee,
			Status:               tx.Status,
			Confirm:              tx.Confirm,
			Chain:                tx.Chain,
			Symbol:               tx.Symbol,
			ContractAddr:         tx.ContractAddr,
			Ctime:                tx.Ctime.Format("2006-01-02 15:04:05"),
			Mtime:                tx.Mtime.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	// 计算是否有更多数据
	hasMore := int64(offset+pageSize) < total

	return &dto.AddressTransactionsResponse{
		Transactions: responses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		HasMore:      hasMore,
	}, nil
}

// computeEthDynamicBalance 基于创建高度余额和自创建高度以来的收支（Wei）计算当前余额（ETH）
func (s *userAddressService) computeEthDynamicBalance(createdBalanceETH float64, inWei string, outWei string) float64 {
	// 将 Wei 字符串转为大整数
	inBigInt := new(big.Int)
	outBigInt := new(big.Int)
	if _, ok := inBigInt.SetString(inWei, 10); !ok {
		inBigInt = big.NewInt(0)
	}
	if _, ok := outBigInt.SetString(outWei, 10); !ok {
		outBigInt = big.NewInt(0)
	}

	// 计算净变化（Wei）
	netWei := new(big.Int).Sub(inBigInt, outBigInt)

	// 将 createdBalanceETH 转为 Wei：createdBalanceETH * 1e18
	baseWei := new(big.Float).Mul(new(big.Float).SetFloat64(createdBalanceETH), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))

	// 将 netWei 转为 big.Float
	netWeiFloat := new(big.Float).SetInt(netWei)

	// 当前余额 Wei = baseWei + netWei
	currentWei := new(big.Float).Add(baseWei, netWeiFloat)

	// 转为 ETH：/ 1e18
	currentETHFloat := new(big.Float).Quo(currentWei, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))

	// 转为 float64 返回
	currentETH, _ := currentETHFloat.Float64()
	return currentETH
}
