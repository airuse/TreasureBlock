package services

import (
	"context"
	"fmt"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
)

// ParserConfigService 解析配置服务接口
type ParserConfigService interface {
	CreateParserConfig(ctx context.Context, config *models.ParserConfig) error
	GetParserConfigByID(ctx context.Context, id uint) (*models.ParserConfig, error)
	GetParserConfigsByContract(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error)
	GetParserConfigBySignature(ctx context.Context, contractAddress, signature string) (*models.ParserConfig, error)
	ListParserConfigs(ctx context.Context, page, pageSize int, contractAddress string) ([]*models.ParserConfig, int64, error)
	UpdateParserConfig(ctx context.Context, config *models.ParserConfig) error
	DeleteParserConfig(ctx context.Context, id uint) error
	GetActiveParserConfigs(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error)
	GetContractParserInfo(ctx context.Context, contractAddress string) (*models.ContractParserInfo, error)
	ParseInputData(ctx context.Context, contractAddress, inputData string) (*ParsedInputData, error)
}

// ParsedInputData 解析后的输入数据
type ParsedInputData struct {
	FunctionName        string                 `json:"function_name"`
	FunctionDescription string                 `json:"function_description"`
	DisplayFormat       string                 `json:"display_format"`
	ParsedParams        map[string]interface{} `json:"parsed_params"`
	ToAddress           string                 `json:"to_address,omitempty"`
	Amount              string                 `json:"amount,omitempty"`
	AmountUnit          string                 `json:"amount_unit,omitempty"`
}

// parserConfigService 解析配置服务实现
type parserConfigService struct {
	parserConfigRepo repository.ParserConfigRepository
}

// NewParserConfigService 创建解析配置服务
func NewParserConfigService(parserConfigRepo repository.ParserConfigRepository) ParserConfigService {
	return &parserConfigService{
		parserConfigRepo: parserConfigRepo,
	}
}

// CreateParserConfig 创建解析配置
func (s *parserConfigService) CreateParserConfig(ctx context.Context, config *models.ParserConfig) error {
	// 验证合约地址格式
	if len(config.ContractAddress) != 42 || config.ContractAddress[:2] != "0x" {
		return fmt.Errorf("invalid contract address format: %s", config.ContractAddress)
	}

	// 验证函数签名格式
	if len(config.FunctionSignature) != 10 || config.FunctionSignature[:2] != "0x" {
		return fmt.Errorf("invalid function signature format: %s", config.FunctionSignature)
	}

	return s.parserConfigRepo.CreateParserConfig(ctx, config)
}

// GetParserConfigByID 根据ID获取解析配置
func (s *parserConfigService) GetParserConfigByID(ctx context.Context, id uint) (*models.ParserConfig, error) {
	return s.parserConfigRepo.GetParserConfigByID(ctx, id)
}

// GetParserConfigsByContract 根据合约地址获取所有解析配置
func (s *parserConfigService) GetParserConfigsByContract(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error) {
	return s.parserConfigRepo.GetParserConfigsByContract(ctx, contractAddress)
}

// GetParserConfigBySignature 根据合约地址和函数签名获取解析配置
func (s *parserConfigService) GetParserConfigBySignature(ctx context.Context, contractAddress, signature string) (*models.ParserConfig, error) {
	return s.parserConfigRepo.GetParserConfigBySignature(ctx, contractAddress, signature)
}

// ListParserConfigs 分页获取解析配置列表
func (s *parserConfigService) ListParserConfigs(ctx context.Context, page, pageSize int, contractAddress string) ([]*models.ParserConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.parserConfigRepo.ListParserConfigs(ctx, page, pageSize, contractAddress)
}

// UpdateParserConfig 更新解析配置
func (s *parserConfigService) UpdateParserConfig(ctx context.Context, config *models.ParserConfig) error {
	return s.parserConfigRepo.UpdateParserConfig(ctx, config)
}

// DeleteParserConfig 删除解析配置
func (s *parserConfigService) DeleteParserConfig(ctx context.Context, id uint) error {
	return s.parserConfigRepo.DeleteParserConfig(ctx, id)
}

// GetActiveParserConfigs 获取指定合约的活跃解析配置
func (s *parserConfigService) GetActiveParserConfigs(ctx context.Context, contractAddress string) ([]*models.ParserConfig, error) {
	return s.parserConfigRepo.GetActiveParserConfigs(ctx, contractAddress)
}

// GetContractParserInfo 获取合约的完整解析信息（三表联查）
func (s *parserConfigService) GetContractParserInfo(ctx context.Context, contractAddress string) (*models.ContractParserInfo, error) {
	return s.parserConfigRepo.GetContractParserInfo(ctx, contractAddress)
}

// ParseInputData 解析交易输入数据
func (s *parserConfigService) ParseInputData(ctx context.Context, contractAddress, inputData string) (*ParsedInputData, error) {
	if inputData == "" || inputData == "0x" {
		return nil, fmt.Errorf("empty input data")
	}

	// 提取函数签名（前10个字符：0x + 8位十六进制）
	if len(inputData) < 10 {
		return nil, fmt.Errorf("input data too short")
	}

	functionSignature := inputData[:10]

	// 查找匹配的解析配置
	config, err := s.GetParserConfigBySignature(ctx, contractAddress, functionSignature)
	if err != nil {
		return nil, fmt.Errorf("no parser config found for signature %s: %w", functionSignature, err)
	}

	// 解析参数
	parsedParams, err := s.parseParameters(inputData[10:], config.ParamConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// 构建解析结果
	result := &ParsedInputData{
		FunctionName:        config.FunctionName,
		FunctionDescription: config.FunctionDescription,
		DisplayFormat:       config.DisplayFormat,
		ParsedParams:        parsedParams,
	}

	// 根据解析规则提取关键信息
	rules := models.ParserRules(config.ParserRules)
	if rules.ExtractToAddress != "" {
		if toAddr, ok := s.extractValueByRule(rules.ExtractToAddress, parsedParams, inputData).(string); ok {
			result.ToAddress = toAddr
		}
	}

	if rules.ExtractAmount != "" {
		if amount, ok := s.extractValueByRule(rules.ExtractAmount, parsedParams, inputData).(string); ok {
			result.Amount = amount
			result.AmountUnit = rules.AmountUnit
		}
	}

	return result, nil
}

// parseParameters 解析参数
func (s *parserConfigService) parseParameters(paramData string, paramConfigs models.ParamConfigs) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, param := range paramConfigs {
		value, err := s.parseParameter(paramData, param)
		if err != nil {
			return nil, fmt.Errorf("failed to parse parameter %s: %w", param.Name, err)
		}
		result[param.Name] = value
	}

	return result, nil
}

// parseParameter 解析单个参数
func (s *parserConfigService) parseParameter(data string, param models.ParamConfig) (interface{}, error) {
	// 计算实际位置（paramData已经去掉了前面的函数签名）
	start := param.Offset - 10 // 减去函数签名的长度
	if start < 0 {
		start = 0
	}
	end := start + param.Length*2 // 十六进制，每字节2个字符

	if len(data) < end {
		return nil, fmt.Errorf("data too short for parameter %s", param.Name)
	}

	hexValue := data[start:end]

	switch param.Type {
	case "address":
		// 地址类型，取后40位（20字节）
		if len(hexValue) >= 40 {
			return "0x" + hexValue[len(hexValue)-40:], nil
		}
		return "0x" + hexValue, nil

	case "uint256":
		// 大整数，保持十六进制格式或转换为字符串
		return "0x" + hexValue, nil

	case "bytes":
		return "0x" + hexValue, nil

	default:
		return hexValue, nil
	}
}

// extractValueByRule 根据规则提取值
func (s *parserConfigService) extractValueByRule(rule string, parsedParams map[string]interface{}, inputData string) interface{} {
	switch rule {
	case "params.to":
		return parsedParams["to"]
	case "params.value":
		return parsedParams["value"]
	case "params.amount":
		return parsedParams["amount"]
	case "params.wad":
		return parsedParams["wad"]
	case "transaction.value":
		// 这个需要从交易本身获取，这里暂时返回空
		return ""
	default:
		// 尝试从parsedParams中获取
		if value, ok := parsedParams[rule]; ok {
			return value
		}
		return nil
	}
}
