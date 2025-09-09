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

	"github.com/ethereum/go-ethereum/common"
)

// UserAddressService 用户地址服务接口
type UserAddressService interface {
	CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error)
	GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error)
	UpdateAddress(userID uint, addressID uint, req *dto.UpdateUserAddressRequest) (*dto.UserAddressResponse, error)
	DeleteAddress(userID uint, addressID uint) error
	GetAddressByID(userID uint, addressID uint) (*dto.UserAddressResponse, error)
	GetAddressTransactions(userID uint, address string, page, pageSize int, chain string) (*dto.AddressTransactionsResponse, error)
	GetAllWalletAddresses() ([]dto.UserAddressResponse, error)
	GetAllWalletAddressModels() ([]*models.UserAddress, error)
	UpdateAddWalletBalance(ID uint, amount uint64) error
	UpdateReduceWalletBalance(ID uint, amount uint64) error
	GetAddressesByAuthorizedAddress(authorizedAddr string) ([]dto.UserAddressResponse, error)
	RefreshAddressBalances(userID uint, addressID uint) (*dto.UserAddressResponse, error)
}

// userAddressService 用户地址服务实现
type userAddressService struct {
	userAddressRepo repository.UserAddressRepository
	blockRepo       repository.BlockRepository
	transactionRepo repository.TransactionRepository
	contractRepo    repository.ContractRepository
	contractCall    ContractCallService
}

// NewUserAddressService 创建用户地址服务
func NewUserAddressService(userAddressRepo repository.UserAddressRepository, blockRepo repository.BlockRepository, contractRepo repository.ContractRepository, contractCall ContractCallService) UserAddressService {
	return &userAddressService{
		userAddressRepo: userAddressRepo,
		blockRepo:       blockRepo,
		transactionRepo: repository.NewTransactionRepository(),
		contractRepo:    contractRepo,
		contractCall:    contractCall,
	}
}

// CreateAddress 创建用户地址
func (s *userAddressService) CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error) {
	// 验证地址格式
	if !s.isValidAddress(req.Address) {
		return nil, errors.New("无效的地址格式")
	}

	// 自动获取当前区块高度和地址余额
	createdHeight, balance, err := s.getCurrentBlockHeightAndBalance(req.Address)
	if err != nil {
		// 如果获取失败，记录错误日志，使用默认值，但不影响地址创建
		// fmt.Printf("警告：获取区块高度和余额失败: %v，使用默认值\n", err)
		createdHeight = 0
		balance = "0"
	}

	// 处理授权地址，确保空数组正确处理
	// build authorized map
	authMap := make(models.AuthorizedAddressesJSON)
	for _, addr := range req.AuthorizedAddresses {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		authMap[addr] = models.AuthorizedInfo{Allowance: "0"}
	}

	// 创建新地址
	userAddress := &models.UserAddress{
		UserID:              userID,
		Address:             req.Address,
		Label:               req.Label,
		Type:                req.Type,
		ContractID:          req.ContractID,
		AuthorizedAddresses: authMap,
		Notes:               req.Notes,
		Balance:             &balance,
		TransactionCount:    0,
		IsActive:            true,
		BalanceHeight:       createdHeight,
	}

	// 如果是合约地址，预先查询合约余额和被授权地址余额
	if (strings.ToLower(req.Type) == "contract" || strings.ToLower(req.Type) == "authorized_contract") && req.ContractID != nil && s.contractRepo != nil && s.contractCall != nil {
		ctx := context.Background()
		if contract, err2 := s.contractRepo.GetByID(ctx, *req.ContractID); err2 == nil && contract != nil && contract.Address != "" {
			var blockNum *big.Int
			if createdHeight > 0 {
				blockNum = new(big.Int).SetUint64(createdHeight)
			}
			// 合约余额（当前地址在该ERC-20合约下的余额）
			if bal, err3 := s.contractCall.CallBalanceOf(ctx, contract.Address, req.Address, blockNum); err3 == nil {
				balStr := bal.String()
				userAddress.ContractBalance = &balStr
			}
			// 授权额度（owner=当前地址, spender=每个授权地址）
			if len(authMap) > 0 {
				for spender := range authMap {
					// authMap 存的是授权者地址，spender 是授权者地址，req.Address 是当前地址
					if allowance, err4 := s.contractCall.CallAllowance(ctx, contract.Address, spender, req.Address, blockNum); err4 == nil {
						info := authMap[spender]
						info.Allowance = allowance.String()
						authMap[spender] = info
					}
				}
				userAddress.AuthorizedAddresses = authMap
			}
		}
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
		// 直接返回存储的余额，不进行动态计算
		// 余额现在由交易处理和合约解析自动维护
		responses = append(responses, *s.convertToResponse(&addr))
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
		// 如果类型不是合约，清空合约ID
		if *req.Type != "contract" {
			address.ContractID = nil
		}
	}
	if req.ContractID != nil {
		address.ContractID = req.ContractID
	}
	if req.AuthorizedAddresses != nil {
		if address.AuthorizedAddresses == nil {
			address.AuthorizedAddresses = make(models.AuthorizedAddressesJSON)
		}
		newMap := make(models.AuthorizedAddressesJSON)
		for _, addr := range *req.AuthorizedAddresses {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				continue
			}
			if old, ok := address.AuthorizedAddresses[addr]; ok {
				newMap[addr] = old
			} else {
				newMap[addr] = models.AuthorizedInfo{Allowance: "0"}
			}
		}
		address.AuthorizedAddresses = newMap
	}
	if req.ContractBalance != nil {
		address.ContractBalance = req.ContractBalance
	}
	if req.Notes != nil {
		address.Notes = *req.Notes
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
	var auth map[string]map[string]string
	if address.AuthorizedAddresses != nil {
		auth = make(map[string]map[string]string, len(address.AuthorizedAddresses))
		for k, v := range address.AuthorizedAddresses {
			auth[k] = map[string]string{"allowance": v.Allowance}
		}
	}
	return &dto.UserAddressResponse{
		ID:                  address.ID,
		Address:             address.Address,
		Label:               address.Label,
		Type:                address.Type,
		ContractID:          address.ContractID,
		AuthorizedAddresses: auth,
		Notes:               address.Notes,
		Balance:             address.Balance,
		ContractBalance:     address.ContractBalance,
		TransactionCount:    address.TransactionCount,
		IsActive:            address.IsActive,
		BalanceHeight:       address.BalanceHeight,
		CreatedAt:           address.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           address.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// getCurrentBlockHeightAndBalance 获取当前区块高度和地址余额
func (s *userAddressService) getCurrentBlockHeightAndBalance(address string) (uint64, string, error) {
	ctx := context.Background()

	blockNumber := uint64(0)
	// 从配置文件获取ETH RPC URL
	chainConfig, exists := config.AppConfig.Blockchain.Chains["eth"]
	if !exists || (chainConfig.RPCURL == "" && len(chainConfig.RPCURLs) == 0) {
		return blockNumber, "0", fmt.Errorf("未配置ETH RPC URL")
	}

	// 使用故障转移管理器
	fo, err := utils.NewEthFailoverFromChain("eth")
	if err != nil {
		return blockNumber, "0", fmt.Errorf("初始化ETH故障转移失败: %w", err)
	}
	defer fo.Close()

	blockNumber, err = fo.BlockNumber(ctx)
	if err != nil {
		return blockNumber, "0", fmt.Errorf("获取当前区块高度失败: %w", err)
	}

	// 获取地址余额 (使用数据库中的最新区块高度)
	balance, err := fo.BalanceAt(ctx, common.HexToAddress(address), big.NewInt(int64(blockNumber)))
	if err != nil {
		return blockNumber, "0", fmt.Errorf("获取地址余额失败: %w", err)
	}

	// 直接返回Wei值，不进行单位转换，保持精度
	return blockNumber, balance.String(), nil
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

// GetAddressesByAuthorizedAddress 根据授权地址查询地址列表（高效JSON查询）
func (s *userAddressService) GetAddressesByAuthorizedAddress(authorizedAddr string) ([]dto.UserAddressResponse, error) {
	// 使用原生SQL进行JSON查询，性能更好
	query := `
		SELECT * FROM user_addresses 
		WHERE JSON_CONTAINS(authorized_addresses, ?) 
		AND type = 'contract' 
		AND is_active = true
	`

	addresses, err := s.userAddressRepo.GetByJSONQuery(query, fmt.Sprintf(`"%s"`, authorizedAddr))
	if err != nil {
		return nil, fmt.Errorf("查询授权地址失败: %w", err)
	}

	var responses []dto.UserAddressResponse
	for _, addr := range addresses {
		responses = append(responses, *s.convertToResponse(&addr))
	}

	return responses, nil
}

// IsAddressAuthorized 检查地址是否为指定合约的授权地址（高效查询）
func (s *userAddressService) IsAddressAuthorized(contractAddress string, authorizedAddr string) (bool, error) {
	query := `
		SELECT COUNT(*) as count 
		FROM user_addresses 
		WHERE address = ? 
		AND type = 'contract' 
		AND JSON_CONTAINS(authorized_addresses, ?)
		AND is_active = true
	`

	count, err := s.userAddressRepo.CountByJSONQuery(query, contractAddress, fmt.Sprintf(`"%s"`, authorizedAddr))
	if err != nil {
		return false, fmt.Errorf("检查授权地址失败: %w", err)
	}

	return count > 0, nil
}
func (s *userAddressService) GetAllWalletAddresses() ([]dto.UserAddressResponse, error) {
	addresses, err := s.userAddressRepo.GetAllWalletAddresses()
	if err != nil {
		return nil, err
	}
	var responses []dto.UserAddressResponse
	for _, addr := range addresses {
		responses = append(responses, *s.convertToResponse(&addr))
	}
	return responses, nil
}

// GetAllWalletAddressModels 获取所有钱包类型的地址模型（用于内部处理）
func (s *userAddressService) GetAllWalletAddressModels() ([]*models.UserAddress, error) {
	return s.userAddressRepo.GetByType("wallet")
}
func (s *userAddressService) UpdateAddWalletBalance(ID uint, amount uint64) error {
	address, err := s.userAddressRepo.GetByID(ID)
	if err != nil {
		return err
	}
	balance, err := strconv.ParseInt(*address.Balance, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %w", err)
	}
	newBalance := balance + int64(amount)
	balanceStr := strconv.FormatInt(newBalance, 10)
	address.Balance = &balanceStr
	return s.userAddressRepo.Update(address)
}

func (s *userAddressService) UpdateReduceWalletBalance(ID uint, amount uint64) error {
	address, err := s.userAddressRepo.GetByID(ID)
	if err != nil {
		return err
	}
	balance, err := strconv.ParseInt(*address.Balance, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %w", err)
	}
	newBalance := balance - int64(amount)
	balanceStr := strconv.FormatInt(newBalance, 10)
	address.Balance = &balanceStr
	return s.userAddressRepo.Update(address)
}

// RefreshAddressBalances 刷新地址余额、合约余额以及授权额度
func (s *userAddressService) RefreshAddressBalances(userID uint, addressID uint) (*dto.UserAddressResponse, error) {
	// 加载地址
	addr, err := s.userAddressRepo.GetByID(addressID)
	if err != nil {
		return nil, fmt.Errorf("地址不存在")
	}
	if addr.UserID != userID {
		return nil, fmt.Errorf("无权限操作此地址")
	}

	// 刷新钱包余额
	height, bal, err := s.getCurrentBlockHeightAndBalance(addr.Address)
	if err == nil {
		addr.Balance = &bal
		addr.BalanceHeight = height
	}

	// 刷新合约余额与授权额度（仅当为合约/被授权合约且具备依赖）
	if (strings.ToLower(addr.Type) == "contract" || strings.ToLower(addr.Type) == "authorized_contract") && addr.ContractID != nil && s.contractRepo != nil && s.contractCall != nil {
		ctx := context.Background()
		if contract, err2 := s.contractRepo.GetByID(ctx, *addr.ContractID); err2 == nil && contract != nil && contract.Address != "" {
			// 使用最新高度（如可用）
			var blockNum *big.Int
			if height > 0 {
				blockNum = new(big.Int).SetUint64(height)
			}
			// 合约余额（address 在该合约下的余额）
			if bal2, err3 := s.contractCall.CallBalanceOf(ctx, contract.Address, addr.Address, blockNum); err3 == nil {
				balStr := bal2.String()
				addr.ContractBalance = &balStr

			}

			// 授权额度（owner=授权地址，spender=当前地址）
			if addr.AuthorizedAddresses != nil {
				updated := make(models.AuthorizedAddressesJSON, len(addr.AuthorizedAddresses))
				for owner := range addr.AuthorizedAddresses {
					if alw, err4 := s.contractCall.CallAllowance(ctx, contract.Address, owner, addr.Address, blockNum); err4 == nil {
						updated[owner] = models.AuthorizedInfo{Allowance: alw.String()}
					} else {
						// 保留原值
						updated[owner] = addr.AuthorizedAddresses[owner]
					}
				}
				addr.AuthorizedAddresses = updated
			}
		}
	}

	if err := s.userAddressRepo.Update(addr); err != nil {
		return nil, err
	}
	return s.convertToResponse(addr), nil
}
