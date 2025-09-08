package repository

import (
	"context"
	"encoding/json"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// ContractRepository 合约仓库接口
type ContractRepository interface {
	Create(ctx context.Context, contract *models.Contract) error
	GetByAddress(ctx context.Context, address string) (*models.Contract, error)
	GetByChainName(ctx context.Context, chainName string) ([]*models.Contract, error)
	GetByType(ctx context.Context, contractType string) ([]*models.Contract, error)
	GetERC20Tokens(ctx context.Context) ([]*models.Contract, error)
	Update(ctx context.Context, contract *models.Contract) error
	Delete(ctx context.Context, address string) error
	Exists(ctx context.Context, address string) (bool, error)
	GetAll(ctx context.Context) ([]*models.Contract, error)
	GetVerifiedContracts(ctx context.Context) ([]*models.Contract, error)
	GetContractsByStatus(ctx context.Context, status int8) ([]*models.Contract, error)
	GetWithFilters(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Contract, int64, error)
	GetByID(ctx context.Context, id uint) (*models.Contract, error)
}

// contractRepository 合约仓库实现
type contractRepository struct {
	db *gorm.DB
}

// NewContractRepository 创建合约仓库
func NewContractRepository(db *gorm.DB) ContractRepository {
	return &contractRepository{db: db}
}

// Create 创建合约
func (r *contractRepository) Create(ctx context.Context, contract *models.Contract) error {
	return r.db.WithContext(ctx).Create(contract).Error
}

// GetByID 根据ID获取合约
func (r *contractRepository) GetByID(ctx context.Context, id uint) (*models.Contract, error) {
	var contract models.Contract
	if err := r.db.WithContext(ctx).First(&contract, id).Error; err != nil {
		return nil, err
	}
	return &contract, nil
}

// GetByAddress 根据地址获取合约
func (r *contractRepository) GetByAddress(ctx context.Context, address string) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.WithContext(ctx).Where("address = ?", address).First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

// GetByChainName 根据链名称获取合约列表
func (r *contractRepository) GetByChainName(ctx context.Context, chainName string) ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.WithContext(ctx).Where("chain_name = ?", chainName).Find(&contracts).Error
	return contracts, err
}

// GetByType 根据合约类型获取合约列表
func (r *contractRepository) GetByType(ctx context.Context, contractType string) ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.WithContext(ctx).Where("contract_type = ?", contractType).Find(&contracts).Error
	return contracts, err
}

// GetERC20Tokens 获取所有ERC-20代币合约
func (r *contractRepository) GetERC20Tokens(ctx context.Context) ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.WithContext(ctx).Where("is_erc20 = ?", true).Find(&contracts).Error
	return contracts, err
}

// Update 更新合约
func (r *contractRepository) Update(ctx context.Context, contract *models.Contract) error {
	return r.db.WithContext(ctx).Save(contract).Error
}

// Delete 删除合约
func (r *contractRepository) Delete(ctx context.Context, address string) error {
	return r.db.WithContext(ctx).Where("address = ?", address).Delete(&models.Contract{}).Error
}

// Exists 检查合约是否存在
func (r *contractRepository) Exists(ctx context.Context, address string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Contract{}).Where("address = ?", address).Count(&count).Error
	return count > 0, err
}

// GetAll 获取所有合约
func (r *contractRepository) GetAll(ctx context.Context) ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.WithContext(ctx).Find(&contracts).Error
	return contracts, err
}

// GetVerifiedContracts 获取已验证的合约
func (r *contractRepository) GetVerifiedContracts(ctx context.Context) ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.WithContext(ctx).Where("verified = ?", true).Find(&contracts).Error
	return contracts, err
}

// GetContractsByStatus 根据状态获取合约
func (r *contractRepository) GetContractsByStatus(ctx context.Context, status int8) ([]*models.Contract, error) {
	var contracts []*models.Contract
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&contracts).Error
	return contracts, err
}

// Helper functions for JSON fields
func (r *contractRepository) SetInterfaces(contract *models.Contract, interfaces []string) error {
	data, err := json.Marshal(interfaces)
	if err != nil {
		return err
	}
	contract.Interfaces = string(data)
	return nil
}

func (r *contractRepository) SetMethods(contract *models.Contract, methods []string) error {
	data, err := json.Marshal(methods)
	if err != nil {
		return err
	}
	contract.Methods = string(data)
	return nil
}

func (r *contractRepository) SetEvents(contract *models.Contract, events []string) error {
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}
	contract.Events = string(data)
	return nil
}

func (r *contractRepository) SetMetadata(contract *models.Contract, metadata map[string]string) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	contract.Metadata = string(data)
	return nil
}

// GetWithFilters 根据过滤条件获取合约列表
func (r *contractRepository) GetWithFilters(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Contract, int64, error) {
	var contracts []*models.Contract
	var total int64

	// 构建查询
	query := r.db.WithContext(ctx).Model(&models.Contract{})

	// 应用过滤条件
	if chainName, ok := filters["chainName"].(string); ok && chainName != "" {
		query = query.Where("chain_name = ?", chainName)
	}

	if contractType, ok := filters["contractType"].(string); ok && contractType != "" {
		query = query.Where("contract_type = ?", contractType)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		// 转换状态字符串为数字
		var statusNum int8
		switch status {
		case "active":
			statusNum = 1
		case "inactive":
			statusNum = 0
		case "paused":
			statusNum = 2
		default:
			statusNum = 1 // 默认显示活跃状态
		}
		query = query.Where("status = ?", statusNum)
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		// 搜索合约地址、名称、符号
		query = query.Where(
			"address LIKE ? OR name LIKE ? OR symbol LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&contracts).Error

	return contracts, total, err
}
