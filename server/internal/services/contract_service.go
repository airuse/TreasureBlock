package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// ContractService 合约服务接口
type ContractService interface {
	CreateOrUpdateContract(ctx context.Context, contractInfo *models.Contract) (*models.Contract, error)
	GetContractByAddress(ctx context.Context, address string) (*models.Contract, error)
	GetContractByAddressAndChain(ctx context.Context, address string, chainName string) (*models.Contract, error)
	GetContractsByChain(ctx context.Context, chainName string) ([]*models.Contract, error)
	GetContractsByType(ctx context.Context, contractType string) ([]*models.Contract, error)
	GetERC20Tokens(ctx context.Context) ([]*models.Contract, error)
	UpdateContractStatus(ctx context.Context, address string, status int8) error
	VerifyContract(ctx context.Context, address string) error
	GetAllContracts(ctx context.Context) ([]*models.Contract, error)
	GetContractsWithFilters(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Contract, int64, error)
	DeleteContract(ctx context.Context, address string) error
}

// contractService 合约服务实现
type contractService struct {
	contractRepo   repository.ContractRepository
	coinConfigRepo repository.CoinConfigRepository
}

// NewContractService 创建合约服务
func NewContractService(contractRepo repository.ContractRepository, coinConfigRepo repository.CoinConfigRepository) ContractService {

	return &contractService{
		contractRepo:   contractRepo,
		coinConfigRepo: coinConfigRepo,
	}
}

// CreateOrUpdateContract 创建或更新合约
func (s *contractService) CreateOrUpdateContract(ctx context.Context, contractInfo *models.Contract) (*models.Contract, error) {
	// 优先按地址+链名检查是否已存在（避免跨链混淆）
	existingContract, err := s.contractRepo.GetByAddressAndChain(ctx, contractInfo.Address, contractInfo.ChainName)
	if err != nil || existingContract == nil {
		// 回退到仅按地址检查，兼容旧数据
		existingContract, err = s.contractRepo.GetByAddress(ctx, contractInfo.Address)
	}
	if err == nil && existingContract != nil {
		// 合约已存在，更新信息
		return s.updateExistingContract(ctx, existingContract, contractInfo)
	}

	// 创建新合约
	return s.createNewContract(ctx, contractInfo)
}

// createNewContract 创建新合约
func (s *contractService) createNewContract(ctx context.Context, contractInfo *models.Contract) (*models.Contract, error) {
	contract := &models.Contract{
		Address:       contractInfo.Address,
		ChainName:     contractInfo.ChainName,
		ProgramID:     contractInfo.ProgramID,
		ContractType:  contractInfo.ContractType,
		Name:          contractInfo.Name,
		Symbol:        contractInfo.Symbol,
		Decimals:      contractInfo.Decimals,
		TotalSupply:   contractInfo.TotalSupply,
		IsERC20:       contractInfo.IsERC20,
		Status:        contractInfo.Status,        // 使用传入的状态，而不是硬编码
		Verified:      contractInfo.Verified,      // 使用传入的验证状态
		Creator:       contractInfo.Creator,       // 设置创建者地址
		CreationTx:    contractInfo.CreationTx,    // 设置创建交易哈希
		CreationBlock: contractInfo.CreationBlock, // 设置创建区块
		ContractLogo:  contractInfo.ContractLogo,  // 设置合约Logo
		CTime:         time.Now(),
		MTime:         time.Now(),
	}

	// 设置JSON字段
	if err := s.setJSONFields(contract, contractInfo); err != nil {
		return nil, fmt.Errorf("failed to set JSON fields: %w", err)
	}

	// 保存到数据库
	if err := s.contractRepo.Create(ctx, contract); err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}
	fmt.Printf("创建了一个新的合约\n")
	// 如果是ERC-20合约，自动创建币种配置
	if contract.IsERC20 {
		fmt.Printf("创建了一个新的ERC-20 币种\n")
		if err := s.createCoinConfigForERC20(ctx, contract); err != nil {
			// 记录错误但不影响合约创建
			// fmt.Printf("Warning: Failed to create coin config for ERC-20 contract %s: %v\n", contract.Address, err)
		}
	}

	return contract, nil
}

// updateExistingContract 更新现有合约
func (s *contractService) updateExistingContract(ctx context.Context, existing *models.Contract, contractInfo *models.Contract) (*models.Contract, error) {
	// 更新所有基本信息字段
	existing.ChainName = contractInfo.ChainName
	existing.ProgramID = contractInfo.ProgramID
	existing.Name = contractInfo.Name
	existing.Symbol = contractInfo.Symbol
	existing.Decimals = contractInfo.Decimals
	existing.TotalSupply = contractInfo.TotalSupply
	existing.IsERC20 = contractInfo.IsERC20
	existing.ContractType = contractInfo.ContractType
	existing.Status = contractInfo.Status
	existing.Verified = contractInfo.Verified
	existing.Creator = contractInfo.Creator
	existing.CreationTx = contractInfo.CreationTx
	existing.CreationBlock = contractInfo.CreationBlock
	existing.ContractLogo = contractInfo.ContractLogo
	existing.MTime = time.Now()

	// 更新JSON字段
	if err := s.setJSONFields(existing, contractInfo); err != nil {
		return nil, fmt.Errorf("failed to set JSON fields: %w", err)
	}

	// 保存更新
	if err := s.contractRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update contract: %w", err)
	}

	return existing, nil
}

// setJSONFields 设置JSON字段
func (s *contractService) setJSONFields(contract *models.Contract, contractInfo *models.Contract) error {
	// 设置接口信息 - 即使为空也要设置，避免保留旧数据
	if contractInfo.Interfaces != "" {
		interfacesData, err := json.Marshal(contractInfo.Interfaces)
		if err != nil {
			return fmt.Errorf("failed to marshal interfaces: %w", err)
		}
		contract.Interfaces = string(interfacesData)
	} else {
		contract.Interfaces = "" // 清空字段
	}

	// 设置方法信息
	if contractInfo.Methods != "" {
		methodsData, err := json.Marshal(contractInfo.Methods)
		if err != nil {
			return fmt.Errorf("failed to marshal methods: %w", err)
		}
		contract.Methods = string(methodsData)
	} else {
		contract.Methods = "" // 清空字段
	}

	// 设置事件信息
	if contractInfo.Events != "" {
		eventsData, err := json.Marshal(contractInfo.Events)
		if err != nil {
			return fmt.Errorf("failed to marshal events: %w", err)
		}
		contract.Events = string(eventsData)
	} else {
		contract.Events = "" // 清空字段
	}

	// 设置元数据
	if contractInfo.Metadata != "" {
		metadataData, err := json.Marshal(contractInfo.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		contract.Metadata = string(metadataData)
	} else {
		contract.Metadata = "" // 清空字段
	}

	return nil
}

// GetContractsWithFilters 根据过滤条件获取合约列表
func (s *contractService) GetContractsWithFilters(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Contract, int64, error) {
	return s.contractRepo.GetWithFilters(ctx, filters, page, pageSize)
}

// createCoinConfigForERC20 为ERC-20合约创建币种配置
func (s *contractService) createCoinConfigForERC20(ctx context.Context, contract *models.Contract) error {
	// 检查是否已存在币种配置
	existingCoin, err := s.coinConfigRepo.GetByContractAddress(ctx, contract.Address)
	if err == nil && existingCoin != nil {
		// 币种配置已存在，跳过创建
		return nil
	}

	// 创建新的币种配置
	coinConfig := &models.CoinConfig{
		ChainName:    contract.ChainName,
		Symbol:       contract.Symbol,
		CoinType:     1, // ERC-20
		ContractAddr: contract.Address,
		Precision:    uint(contract.Decimals),
		Decimals:     uint(contract.Decimals),
		Name:         contract.Name,
		LogoURL:      contract.ContractLogo,
		IsVerified:   contract.Verified,
		Status:       1, // 启用状态
	}

	// 保存币种配置
	return s.coinConfigRepo.Create(ctx, coinConfig)
}

// GetContractByAddress 根据地址获取合约
func (s *contractService) GetContractByAddress(ctx context.Context, address string) (*models.Contract, error) {
	return s.contractRepo.GetByAddress(ctx, address)
}

// GetContractByAddressAndChain 根据地址和链名称获取合约
func (s *contractService) GetContractByAddressAndChain(ctx context.Context, address string, chainName string) (*models.Contract, error) {
	return s.contractRepo.GetByAddressAndChain(ctx, address, chainName)
}

// GetContractsByChain 根据链名称获取合约列表
func (s *contractService) GetContractsByChain(ctx context.Context, chainName string) ([]*models.Contract, error) {
	return s.contractRepo.GetByChainName(ctx, chainName)
}

// GetContractsByType 根据合约类型获取合约列表
func (s *contractService) GetContractsByType(ctx context.Context, contractType string) ([]*models.Contract, error) {
	return s.contractRepo.GetByType(ctx, contractType)
}

// GetERC20Tokens 获取所有ERC-20代币合约
func (s *contractService) GetERC20Tokens(ctx context.Context) ([]*models.Contract, error) {
	return s.contractRepo.GetERC20Tokens(ctx)
}

// UpdateContractStatus 更新合约状态
func (s *contractService) UpdateContractStatus(ctx context.Context, address string, status int8) error {
	contract, err := s.contractRepo.GetByAddress(ctx, address)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	contract.Status = status
	contract.MTime = time.Now()

	return s.contractRepo.Update(ctx, contract)
}

// VerifyContract 验证合约
func (s *contractService) VerifyContract(ctx context.Context, address string) error {
	contract, err := s.contractRepo.GetByAddress(ctx, address)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	contract.Verified = true
	contract.MTime = time.Now()

	return s.contractRepo.Update(ctx, contract)
}

// GetAllContracts 获取所有合约
func (s *contractService) GetAllContracts(ctx context.Context) ([]*models.Contract, error) {
	return s.contractRepo.GetAll(ctx)
}

// DeleteContract 删除合约
func (s *contractService) DeleteContract(ctx context.Context, address string) error {
	return s.contractRepo.Delete(ctx, address)
}
