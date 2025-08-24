package repository

import (
	"context"
	"fmt"

	"blockChainBrowser/server/internal/models"

	"gorm.io/gorm"
)

// ParserConfigRepository 解析配置仓储接口
type ParserConfigRepository interface {
	CreateParserConfig(ctx context.Context, config *models.ParserConfig) error
	GetParserConfigByID(ctx context.Context, id uint) (*models.ParserConfig, error)
	GetParserConfigsByContract(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error)
	GetParserConfigBySignature(ctx context.Context, contractAddress, signature string) (*models.ParserConfig, error)
	ListParserConfigs(ctx context.Context, page, pageSize int, contractAddress string) ([]*models.ParserConfig, int64, error)
	UpdateParserConfig(ctx context.Context, config *models.ParserConfig) error
	DeleteParserConfig(ctx context.Context, id uint) error
	GetActiveParserConfigs(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error)
	GetContractParserInfo(ctx context.Context, contractAddress string) (*models.ContractParserInfo, error)
}

// parserConfigRepository 解析配置仓储实现
type parserConfigRepository struct {
	db *gorm.DB
}

// NewParserConfigRepository 创建解析配置仓储
func NewParserConfigRepository(db *gorm.DB) ParserConfigRepository {
	return &parserConfigRepository{db: db}
}

// CreateParserConfig 创建解析配置
func (r *parserConfigRepository) CreateParserConfig(ctx context.Context, config *models.ParserConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetParserConfigByID 根据ID获取解析配置
func (r *parserConfigRepository) GetParserConfigByID(ctx context.Context, id uint) (*models.ParserConfig, error) {
	var config models.ParserConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetParserConfigsByContract 根据合约地址获取所有解析配置
func (r *parserConfigRepository) GetParserConfigsByContract(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error) {
	var configs []*models.ParserConfig
	err := r.db.WithContext(ctx).
		Where("contract_address = ?", contractAddress).
		Order("priority DESC, created_at ASC").
		Find(&configs).Error
	return configs, err
}

// GetParserConfigBySignature 根据合约地址和函数签名获取解析配置
func (r *parserConfigRepository) GetParserConfigBySignature(ctx context.Context, contractAddress, signature string) (*models.ParserConfig, error) {
	var config models.ParserConfig
	err := r.db.WithContext(ctx).
		Where("contract_address = ? AND function_signature = ? AND is_active = true", contractAddress, signature).
		Order("priority DESC").
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ListParserConfigs 分页获取解析配置列表
func (r *parserConfigRepository) ListParserConfigs(ctx context.Context, page, pageSize int, contractAddress string) ([]*models.ParserConfig, int64, error) {
	var configs []*models.ParserConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ParserConfig{})

	if contractAddress != "" {
		query = query.Where("contract_address = ?", contractAddress)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Order("priority DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&configs).Error

	return configs, total, err
}

// UpdateParserConfig 更新解析配置
func (r *parserConfigRepository) UpdateParserConfig(ctx context.Context, config *models.ParserConfig) error {
	// 明确指定要更新的字段，排除时间字段和ID
	return r.db.WithContext(ctx).
		Select("contract_address", "parser_type", "function_signature", "function_name",
			"function_description", "param_config", "parser_rules", "display_format",
			"is_active", "priority", "logs_parser_type", "event_signature", "event_name",
			"event_description", "logs_param_config", "logs_parser_rules", "logs_display_format").
		Where("id = ?", config.ID).
		Updates(config).Error
}

// DeleteParserConfig 删除解析配置
func (r *parserConfigRepository) DeleteParserConfig(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ParserConfig{}, id).Error
}

// GetActiveParserConfigs 获取指定合约的活跃解析配置
func (r *parserConfigRepository) GetActiveParserConfigs(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error) {
	var configs []*models.ParserConfig
	err := r.db.WithContext(ctx).
		Where("contract_address = ? AND is_active = true", contractAddress).
		Order("priority DESC, created_at ASC").
		Find(&configs).Error
	return configs, err
}

// GetContractParserInfo 获取合约的完整解析信息（三表联查）
func (r *parserConfigRepository) GetContractParserInfo(ctx context.Context, contractAddress string) (*models.ContractParserInfo, error) {
	info := &models.ContractParserInfo{}

	// 查询合约信息
	var contract models.Contract
	if err := r.db.WithContext(ctx).Where("address = ?", contractAddress).First(&contract).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("contract not found: %s", contractAddress)
		}
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}
	info.Contract = &contract

	// 查询币种配置信息
	var coinConfig models.CoinConfig
	if err := r.db.WithContext(ctx).Where("contract_addr = ?", contractAddress).First(&coinConfig).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to get coin config: %w", err)
		}
		// 币种配置不存在不是错误，可能是纯合约
	} else {
		info.CoinConfig = &coinConfig
	}

	// 查询解析配置
	parserConfigs, err := r.GetActiveParserConfigs(ctx, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get parser configs: %w", err)
	}
	info.ParserConfigs = parserConfigs

	return info, nil
}
