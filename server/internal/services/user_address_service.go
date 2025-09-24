package services

import (
	"blockChainBrowser/server/config"
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"
	"context"
	"encoding/json"
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
	GetUserAddresses(userID uint, chain string) ([]dto.UserAddressResponse, error)
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
	GetAddressUTXOs(userID uint, address string) ([]*models.BTCUTXO, error)
	GetUserAddressesByPending(userID uint, chain string) ([]dto.UserAddressPendingResponse, error)
}

// userAddressService 用户地址服务实现
type userAddressService struct {
	userAddressRepo repository.UserAddressRepository
	blockRepo       repository.BlockRepository
	transactionRepo repository.TransactionRepository
	contractRepo    repository.ContractRepository
	contractCall    ContractCallService
	btcUtxoService  BTCUTXOService
	baseConfigRepo  repository.BaseConfigRepository
	userTxRepo      repository.UserTransactionRepository
}

// NewUserAddressService 创建用户地址服务
func NewUserAddressService(userAddressRepo repository.UserAddressRepository, blockRepo repository.BlockRepository, contractRepo repository.ContractRepository, contractCall ContractCallService, btcUtxoService BTCUTXOService, baseConfigRepo repository.BaseConfigRepository, userTxRepo repository.UserTransactionRepository) UserAddressService {
	return &userAddressService{
		userAddressRepo: userAddressRepo,
		blockRepo:       blockRepo,
		transactionRepo: repository.NewTransactionRepository(),
		contractRepo:    contractRepo,
		contractCall:    contractCall,
		btcUtxoService:  btcUtxoService,
		baseConfigRepo:  baseConfigRepo,
		userTxRepo:      userTxRepo,
	}
}

// CreateAddress 创建用户地址
func (s *userAddressService) CreateAddress(userID uint, req *dto.CreateUserAddressRequest) (*dto.UserAddressResponse, error) {
	// 验证地址格式（按链类型）

	// 自动获取当前区块高度和地址余额（仅对ETH执行）
	createdHeight := uint64(0)
	balance := "0"
	if strings.ToLower(req.Chain) == "eth" {
		var err error
		createdHeight, balance, err = s.getCurrentBlockHeightAndBalance(req.Address)
		if err != nil {
			// 如果获取失败，记录错误日志，使用默认值，但不影响地址创建
			createdHeight = 0
			balance = "0"
		}
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
		Chain:               strings.ToLower(req.Chain),
		Label:               req.Label,
		Type:                req.Type,
		ContractID:          req.ContractID,
		AuthorizedAddresses: authMap,
		Notes:               req.Notes,
		Balance:             &balance,
		TransactionCount:    0,
		IsActive:            true,
		BalanceHeight:       createdHeight,
		AtaOwnerAddress:     req.AtaOwnerAddress,
		AtaMintAddress:      req.AtaMintAddress,
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
			// 按链调用，避免跨链拿错余额
			chain := strings.ToLower(req.Chain)
			bal := (*big.Int)(nil)
			var err3 error
			switch chain {
			case "eth":
				bal, err3 = s.contractCall.CallBalanceOfOnChain(ctx, "eth", contract.Address, req.Address, blockNum)
			case "bsc":
				bal, err3 = s.contractCall.CallBalanceOfOnChain(ctx, "bsc", contract.Address, req.Address, blockNum)
			default:
				bal, err3 = s.contractCall.CallBalanceOf(ctx, contract.Address, req.Address, blockNum)
			}
			if err3 == nil {
				balStr := bal.String()
				userAddress.ContractBalance = &balStr
			}
			// 授权额度（owner=当前地址, spender=每个授权地址）
			if len(authMap) > 0 {
				for spender := range authMap {
					// authMap 存的是授权者地址，spender 是授权者地址，req.Address 是当前地址
					var allowance *big.Int
					var err4 error
					switch chain {
					case "eth":
						allowance, err4 = s.contractCall.CallAllowanceOnChain(ctx, "eth", contract.Address, spender, req.Address, blockNum)
					case "bsc":
						allowance, err4 = s.contractCall.CallAllowanceOnChain(ctx, "bsc", contract.Address, spender, req.Address, blockNum)
					default:
						allowance, err4 = s.contractCall.CallAllowance(ctx, contract.Address, spender, req.Address, blockNum)
					}
					if err4 == nil {
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
func (s *userAddressService) GetUserAddresses(userID uint, chain string) ([]dto.UserAddressResponse, error) {
	addresses, err := s.userAddressRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.UserAddressResponse
	for _, addr := range addresses {
		if !strings.EqualFold(addr.Chain, chain) {
			continue
		}
		// 直接返回存储的余额，不进行动态计算
		// 余额现在由交易处理和合约解析自动维护
		responses = append(responses, *s.convertToResponse(&addr))
	}

	return responses, nil
}

// GetUserAddressesByPending 获取用户所有在途交易地址
func (s *userAddressService) GetUserAddressesByPending(userID uint, chain string) ([]dto.UserAddressPendingResponse, error) {
	ctx := context.Background()

	// 获取用户在途交易
	userTxs, err := s.userTxRepo.GetByUserIDExcludingPending(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户交易失败: %w", err)
	}

	// 过滤指定链的交易
	var filteredTxs []*models.UserTransaction
	for _, tx := range userTxs {
		if strings.EqualFold(tx.Chain, chain) {
			filteredTxs = append(filteredTxs, tx)
		}
	}

	// 转换为响应格式
	var responses []dto.UserAddressPendingResponse
	for _, tx := range filteredTxs {
		response := dto.UserAddressPendingResponse{
			ID:        tx.ID,
			Address:   tx.FromAddress,
			Amount:    tx.Amount,
			Status:    tx.Status,
			Fee:       tx.Fee,
			CreatedAt: tx.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: tx.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
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
	// 更新 SOL-ATA 关联冗余
	if req.AtaOwnerAddress != nil {
		address.AtaOwnerAddress = *req.AtaOwnerAddress
	}
	if req.AtaMintAddress != nil {
		address.AtaMintAddress = *req.AtaMintAddress
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
		Chain:               address.Chain,
		Label:               address.Label,
		Type:                address.Type,
		ContractID:          address.ContractID,
		AuthorizedAddresses: auth,
		Notes:               address.Notes,
		Balance:             address.Balance,
		ContractBalance:     address.ContractBalance,
		TransactionCount:    address.TransactionCount,
		UTXOCount:           address.UTXOCount,
		IsActive:            address.IsActive,
		BalanceHeight:       address.BalanceHeight,
		CreatedAt:           address.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           address.UpdatedAt.Format("2006-01-02 15:04:05"),
		// SOL-ATA 冗余字段
		AtaOwnerAddress: address.AtaOwnerAddress,
		AtaMintAddress:  address.AtaMintAddress,
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

// getEVMCurrentBlockHeightAndBalance 获取指定EVM链当前区块高度和地址余额（返回Wei字符串）
func (s *userAddressService) getEVMCurrentBlockHeightAndBalance(chain string, address string) (uint64, string, error) {
	ctx := context.Background()

	chainLower := strings.ToLower(chain)

	// 检查RPC配置
	chainConfig, exists := config.AppConfig.Blockchain.Chains[chainLower]
	if !exists || (chainConfig.RPCURL == "" && len(chainConfig.RPCURLs) == 0) {
		return 0, "0", fmt.Errorf("未配置%s RPC URL", strings.ToUpper(chainLower))
	}

	// 故障转移
	fo, err := utils.NewEthFailoverFromChain(chainLower)
	if err != nil {
		return 0, "0", fmt.Errorf("初始化%s故障转移失败: %w", strings.ToUpper(chainLower), err)
	}
	defer fo.Close()

	// 最新高度
	blockNumber, err := fo.BlockNumber(ctx)
	if err != nil {
		return 0, "0", fmt.Errorf("获取当前区块高度失败: %w", err)
	}

	// 余额（按该高度）
	balance, err := fo.BalanceAt(ctx, common.HexToAddress(address), big.NewInt(int64(blockNumber)))
	if err != nil {
		return blockNumber, "0", fmt.Errorf("获取地址余额失败: %w", err)
	}

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
	transactions, total, err := s.transactionRepo.GetByAddress(context.Background(), address, offset, pageSize, chain)
	if err != nil {
		return nil, fmt.Errorf("获取交易列表失败: %w", err)
	}

	// 转换为响应DTO
	var responses []dto.AddressTransactionResponse
	for _, tx := range transactions {
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

	// 根据链类型刷新钱包余额
	height := uint64(0)
	bal := "0"

	switch strings.ToLower(addr.Chain) {
	case "eth":
		var err error
		height, bal, err = s.getCurrentBlockHeightAndBalance(addr.Address)
		if err == nil {
			addr.Balance = &bal
			addr.BalanceHeight = height
		}
	case "bsc":
		var err error
		height, bal, err = s.getEVMCurrentBlockHeightAndBalance("bsc", addr.Address)
		if err == nil {
			addr.Balance = &bal
			addr.BalanceHeight = height
		}
	case "btc":
		// 对于BTC，从UTXO计算余额
		var err error
		height, bal, err = s.getBTCBalanceFromUTXO(addr.Address)
		if err == nil {
			addr.Balance = &bal
			addr.BalanceHeight = height
		}
		// 同时更新UTXO数量
		if s.btcUtxoService != nil {
			ctx := context.Background()
			utxoCount, utxoErr := s.btcUtxoService.GetUTXOCountByAddress(ctx, "btc", addr.Address)
			if utxoErr == nil {
				addr.UTXOCount = utxoCount
			}
		}
	case "sol":
		// 使用 Sol RPC 获取余额（lamports）与最新 slot 作为高度
		fo, err := utils.NewSolFailoverFromChain("sol")
		if err == nil {
			ctx := context.Background()
			slot, lamports, gerr := fo.GetAccountBalance(ctx, addr.Address)
			if gerr == nil {
				bal = strconv.FormatUint(lamports, 10)
				height = slot
				addr.Balance = &bal
				addr.BalanceHeight = height
			}
			fo.Close()
		}
	default:
		// 其他链类型暂时不处理
		height = 0
		bal = "0"
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
			// 指定链调用，避免跨链取错余额
			var bal2 *big.Int
			var err3 error
			switch strings.ToLower(addr.Chain) {
			case "eth":
				bal2, err3 = s.contractCall.CallBalanceOfOnChain(ctx, "eth", contract.Address, addr.Address, blockNum)
			case "bsc":
				bal2, err3 = s.contractCall.CallBalanceOfOnChain(ctx, "bsc", contract.Address, addr.Address, blockNum)
			default:
				bal2, err3 = s.contractCall.CallBalanceOf(ctx, contract.Address, addr.Address, blockNum)
			}
			if err3 == nil {
				balStr := bal2.String()
				addr.ContractBalance = &balStr

			}

			// 授权额度（owner=授权地址，spender=当前地址）
			if addr.AuthorizedAddresses != nil {
				updated := make(models.AuthorizedAddressesJSON, len(addr.AuthorizedAddresses))
				for owner := range addr.AuthorizedAddresses {
					var alw *big.Int
					var err4 error
					switch strings.ToLower(addr.Chain) {
					case "eth":
						alw, err4 = s.contractCall.CallAllowanceOnChain(ctx, "eth", contract.Address, owner, addr.Address, blockNum)
					case "bsc":
						alw, err4 = s.contractCall.CallAllowanceOnChain(ctx, "bsc", contract.Address, owner, addr.Address, blockNum)
					default:
						alw, err4 = s.contractCall.CallAllowance(ctx, contract.Address, owner, addr.Address, blockNum)
					}
					if err4 == nil {
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

// getBTCBalanceFromUTXO 从外部API获取BTC余额
func (s *userAddressService) getBTCBalanceFromUTXO(address string) (uint64, string, error) {
	// 直接从外部API获取余额
	btcFailover, err := utils.NewBTCFailoverFromChain("btc")
	if err != nil {
		return 0, "0", fmt.Errorf("创建BTC故障转移管理器失败: %w", err)
	}

	ctx := context.Background()

	// 获取最新高度
	fmt.Println("开始获取最新高度")
	height, err := btcFailover.GetLatestBlockHeight(ctx)
	if err != nil {
		return 0, "0", fmt.Errorf("获取最新高度失败: %w", err)
	}
	balanceHeightOffset, err := s.baseConfigRepo.GetByConfigKey(ctx, "confirmations_btc", 1, "scan")

	if err != nil {
		return 0, "0", fmt.Errorf("获取平衡高度偏移失败: %w", err)
	}

	// 解析配置值
	offset, err := strconv.ParseUint(balanceHeightOffset.ConfigValue, 10, 64)
	if err != nil {
		return 0, "0", fmt.Errorf("解析高度偏移配置失败: %w", err)
	}

	// 高度等于最新高度- 安全高度偏移
	if height > offset {
		height = height - offset
	} else {
		height = 0
	}

	fmt.Printf("获取到的安全高度是: %d\n", height)

	// 拉取地址UTXO并入库（作为余额的结构化来源）
	if s.btcUtxoService != nil {
		addrUtxos, err := btcFailover.GetAddressUTXOs(ctx, address)
		if err == nil && len(addrUtxos) > 0 {
			// 预拉取并缓存相关交易以获得脚本与类型
			outputs := make([]*models.BTCUTXO, 0, len(addrUtxos))
			txCache := make(map[string]*utils.BTCTx)
			for _, u := range addrUtxos {
				var tx *utils.BTCTx
				if cached, ok := txCache[u.TxID]; ok {
					tx = cached
				} else {
					if t, terr := btcFailover.GetTransaction(ctx, u.TxID); terr == nil {
						tx = t
						txCache[u.TxID] = t
					}
				}

				// 仅保留已确认的UTXO（过滤交易池中的未确认输出）
				// 若区块高度为0或缺失，视为未确认，跳过
				if u.Status.BlockHeight <= 0 {
					continue
				}

				var scriptHex, scriptType string
				if tx != nil {
					// 保护性检查索引
					if u.Vout >= 0 && u.Vout < len(tx.Vout) {
						vo := tx.Vout[u.Vout]
						scriptHex = vo.ScriptPubKey
						scriptType = convertBTCScriptTypeRESTToRPC(vo.ScriptPubKeyType)
					}
				}
				// 如果区块高度大于最新高度，则跳过
				if u.Status.BlockHeight > height {
					continue
				}

				// 备注：coinbase成熟度判断依赖交易输入信息，当前REST结构未提供Vin，保持默认false
				isCoinbase := false

				out := &models.BTCUTXO{
					Chain:        "btc",
					TxID:         u.TxID,
					VoutIndex:    uint32(u.Vout),
					BlockHeight:  u.Status.BlockHeight,
					Address:      address,
					ScriptPubKey: scriptHex,
					ScriptType:   scriptType,
					IsCoinbase:   isCoinbase,
					ValueSatoshi: u.Value,
				}
				// 如果没有高度，使用前面计算的安全高度
				outputs = append(outputs, out)
			}
			// 批量Upsert到 btc_utxo 表
			if upErr := s.btcUtxoService.UpsertOutputs(ctx, outputs); upErr != nil {
				fmt.Printf("UTXO入库失败: %v\n", upErr)
			}
		}
	}

	// 通过数据库 btcUtxoService 获取余额
	utxos, err := s.btcUtxoService.GetUTXOsByAddress(ctx, "btc", address)
	if err != nil {
		return 0, "0", fmt.Errorf("获取BTC余额失败: %w", err)
	}
	balance := int64(0)
	for _, u := range utxos {
		balance += u.ValueSatoshi
	}

	return height, strconv.FormatInt(balance, 10), nil
}

// convertBTCScriptTypeRESTToRPC 将 REST/Esplora 风格脚本类型转换为与扫块程序一致的类型命名
// 统一目标：
// - p2pkh            -> pubkeyhash
// - p2sh             -> scripthash
// - v0_p2wpkh        -> witness_v0_keyhash
// - v0_p2wsh         -> witness_v0_scripthash
// - v1_p2tr          -> witness_v1_taproot
// - nulldata         -> nulldata
// - nonstandard      -> nonstandard
// - multisig         -> multisig
// - 其他/unknown     -> witness_unknown（与扫描端保持类别一致）
func convertBTCScriptTypeRESTToRPC(t string) string {
	tt := strings.ToLower(strings.TrimSpace(t))
	switch tt {
	case "p2pkh":
		return "pubkeyhash"
	case "p2sh":
		return "scripthash"
	case "v0_p2wpkh":
		return "witness_v0_keyhash"
	case "v0_p2wsh":
		return "witness_v0_scripthash"
	case "v1_p2tr":
		return "witness_v1_taproot"
	case "nulldata":
		return "nulldata"
	case "nonstandard":
		return "nonstandard"
	case "multisig":
		return "multisig"
	case "witness_unknown":
		return "witness_unknown"
	case "unknown":
		return "witness_unknown"
	default:
		// 兜底：部分实现可能返回简化/变体命名，这里做一些宽松匹配
		if strings.Contains(tt, "wpkh") {
			return "witness_v0_keyhash"
		}
		if strings.Contains(tt, "wsh") {
			return "witness_v0_scripthash"
		}
		if strings.Contains(tt, "taproot") || strings.Contains(tt, "p2tr") {
			return "witness_v1_taproot"
		}
		if strings.Contains(tt, "pkh") {
			return "pubkeyhash"
		}
		if strings.Contains(tt, "sh") {
			return "scripthash"
		}
		if strings.HasPrefix(tt, "v0_") {
			// 未知v0见证，归类为unknown见证
			return "witness_unknown"
		}
		return "witness_unknown"
	}
}

// GetAddressUTXOs 获取地址的UTXO列表
func (s *userAddressService) GetAddressUTXOs(userID uint, address string) ([]*models.BTCUTXO, error) {
	ctx := context.Background()

	// 首先验证地址是否存在
	userAddress, err := s.userAddressRepo.GetByAddress(address)
	if err != nil {
		return nil, fmt.Errorf("地址不存在: %w", err)
	}

	// 验证地址是否属于该用户
	if userAddress.UserID != userID {
		return nil, fmt.Errorf("地址不属于当前用户")
	}

	// 只允许BTC地址获取UTXO
	if userAddress.Chain != "btc" {
		return nil, fmt.Errorf("只有BTC地址支持UTXO查询")
	}

	// 获取UTXO列表
	utxos, err := s.btcUtxoService.GetUTXOsByAddress(ctx, "btc", address)
	if err != nil {
		return nil, fmt.Errorf("获取UTXO列表失败: %w", err)
	}

	txs, err := s.userTxRepo.GetByChainExcludingPending(ctx, "btc")
	if err != nil {
		return nil, fmt.Errorf("获取pending交易失败: %w", err)
	}

	for _, tx := range txs {
		if tx.BTCTxInJSON == nil || *tx.BTCTxInJSON == "" {
			continue
		}

		// 解析BTCTxInJSON数组
		var txInArray []map[string]interface{}
		err := json.Unmarshal([]byte(*tx.BTCTxInJSON), &txInArray)
		if err != nil {
			return nil, fmt.Errorf("解析BTCTxInJSON失败: %w", err)
		}

		// 遍历每个输入
		for vinIndex, txIn := range txInArray {
			txid, ok1 := txIn["txid"].(string)
			vout, ok2 := txIn["vout"].(float64) // JSON数字默认解析为float64
			if !ok1 || !ok2 {
				continue
			}

			// 查找匹配的UTXO并标记为已花费
			for _, utxo := range utxos {
				if utxo.TxID == txid && utxo.VoutIndex == uint32(vout) {
					utxo.SpentTxID = *tx.TxHash
					vinIndexUint32 := uint32(vinIndex)
					utxo.SpentVinIndex = &vinIndexUint32
					utxo.SpentHeight = tx.BlockHeight
					utxo.SpentAt = &tx.CreatedAt
					utxo.Status = "spent"
				}
			}
		}
	}

	return utxos, nil
}
