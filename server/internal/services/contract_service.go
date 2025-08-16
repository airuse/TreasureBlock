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
	GetContractsByChain(ctx context.Context, chainName string) ([]*models.Contract, error)
	GetContractsByType(ctx context.Context, contractType string) ([]*models.Contract, error)
	GetERC20Tokens(ctx context.Context) ([]*models.Contract, error)
	UpdateContractStatus(ctx context.Context, address string, status int8) error
	VerifyContract(ctx context.Context, address string) error
	GetAllContracts(ctx context.Context) ([]*models.Contract, error)
	DeleteContract(ctx context.Context, address string) error
}

// contractService 合约服务实现
type contractService struct {
	contractRepo repository.ContractRepository
}

// NewContractService 创建合约服务
func NewContractService(contractRepo repository.ContractRepository) ContractService {

	return &contractService{
		contractRepo: contractRepo,
	}
}

// CreateOrUpdateContract 创建或更新合约
func (s *contractService) CreateOrUpdateContract(ctx context.Context, contractInfo *models.Contract) (*models.Contract, error) {
	// 检查合约是否已存在
	existingContract, err := s.contractRepo.GetByAddress(ctx, contractInfo.Address)
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
		Address:      contractInfo.Address,
		ChainName:    contractInfo.ChainName,
		ContractType: contractInfo.ContractType,
		Name:         contractInfo.Name,
		Symbol:       contractInfo.Symbol,
		Decimals:     contractInfo.Decimals,
		TotalSupply:  contractInfo.TotalSupply,
		IsERC20:      contractInfo.IsERC20,
		Status:       models.ContractStatusEnabled,
		Verified:     false,
		CTime:        time.Now(),
		MTime:        time.Now(),
	}

	// 设置JSON字段
	if err := s.setJSONFields(contract, contractInfo); err != nil {
		return nil, fmt.Errorf("failed to set JSON fields: %w", err)
	}

	// 保存到数据库
	if err := s.contractRepo.Create(ctx, contract); err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	return contract, nil
}

// updateExistingContract 更新现有合约
func (s *contractService) updateExistingContract(ctx context.Context, existing *models.Contract, contractInfo *models.Contract) (*models.Contract, error) {
	// 更新基本信息
	existing.Name = contractInfo.Name
	existing.Symbol = contractInfo.Symbol
	existing.Decimals = contractInfo.Decimals
	existing.TotalSupply = contractInfo.TotalSupply
	existing.IsERC20 = contractInfo.IsERC20
	existing.ContractType = contractInfo.ContractType
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
	// 设置接口信息
	if len(contractInfo.Interfaces) > 0 {
		interfacesData, err := json.Marshal(contractInfo.Interfaces)
		if err != nil {
			return fmt.Errorf("failed to marshal interfaces: %w", err)
		}
		contract.Interfaces = string(interfacesData)
	}

	// 设置方法信息
	if len(contractInfo.Methods) > 0 {
		methodsData, err := json.Marshal(contractInfo.Methods)
		if err != nil {
			return fmt.Errorf("failed to marshal methods: %w", err)
		}
		contract.Methods = string(methodsData)
	}

	// 设置事件信息
	if len(contractInfo.Events) > 0 {
		eventsData, err := json.Marshal(contractInfo.Events)
		if err != nil {
			return fmt.Errorf("failed to marshal events: %w", err)
		}
		contract.Events = string(eventsData)
	}

	// 设置元数据
	if len(contractInfo.Metadata) > 0 {
		metadataData, err := json.Marshal(contractInfo.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		contract.Metadata = string(metadataData)
	}

	return nil
}

// GetContractByAddress 根据地址获取合约
func (s *contractService) GetContractByAddress(ctx context.Context, address string) (*models.Contract, error) {
	return s.contractRepo.GetByAddress(ctx, address)
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
