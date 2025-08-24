package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"
)

// CreateParserConfigRequest 创建解析配置请求
type CreateParserConfigRequest struct {
	ContractAddress     string               `json:"contract_address" binding:"required"`
	ParserType          models.ParserType    `json:"parser_type" binding:"required,oneof=input_data event_log"`
	FunctionSignature   string               `json:"function_signature" binding:"required"`
	FunctionName        string               `json:"function_name" binding:"required"`
	FunctionDescription string               `json:"function_description" binding:"required"`
	ParamConfig         []models.ParamConfig `json:"param_config"`
	ParserRules         models.ParserRules   `json:"parser_rules"`
	DisplayFormat       string               `json:"display_format"`
	IsActive            *bool                `json:"is_active"`
	Priority            *int                 `json:"priority"`
}

// ToModel 转换为模型
func (r *CreateParserConfigRequest) ToModel() *models.ParserConfig {
	config := &models.ParserConfig{
		ContractAddress:     r.ContractAddress,
		ParserType:          r.ParserType,
		FunctionSignature:   r.FunctionSignature,
		FunctionName:        r.FunctionName,
		FunctionDescription: r.FunctionDescription,
		ParamConfig:         models.ParamConfigs(r.ParamConfig),
		ParserRules:         models.ParserRulesJSON(r.ParserRules),
		DisplayFormat:       r.DisplayFormat,
	}

	if r.IsActive != nil {
		config.IsActive = *r.IsActive
	} else {
		config.IsActive = true
	}

	if r.Priority != nil {
		config.Priority = *r.Priority
	} else {
		config.Priority = 0
	}

	return config
}

// UpdateParserConfigRequest 更新解析配置请求
type UpdateParserConfigRequest struct {
	ContractAddress     *string               `json:"contract_address,omitempty"`
	ParserType          *models.ParserType    `json:"parser_type,omitempty"`
	FunctionSignature   *string               `json:"function_signature,omitempty"`
	FunctionName        *string               `json:"function_name,omitempty"`
	FunctionDescription *string               `json:"function_description,omitempty"`
	ParamConfig         *[]models.ParamConfig `json:"param_config,omitempty"`
	ParserRules         *models.ParserRules   `json:"parser_rules,omitempty"`
	DisplayFormat       *string               `json:"display_format,omitempty"`
	IsActive            *bool                 `json:"is_active,omitempty"`
	Priority            *int                  `json:"priority,omitempty"`
	// 日志解析配置字段
	LogsParserType    *string                  `json:"logs_parser_type,omitempty"`
	EventSignature    *string                  `json:"event_signature,omitempty"`
	EventName         *string                  `json:"event_name,omitempty"`
	EventDescription  *string                  `json:"event_description,omitempty"`
	LogsParamConfig   *models.LogsParamConfigs `json:"logs_param_config,omitempty"`
	LogsParserRules   *models.LogsParserRules  `json:"logs_parser_rules,omitempty"`
	LogsDisplayFormat *string                  `json:"logs_display_format,omitempty"`
}

// ApplyToModel 应用更新到模型
func (r *UpdateParserConfigRequest) ApplyToModel(config *models.ParserConfig) {
	if r.ContractAddress != nil {
		config.ContractAddress = *r.ContractAddress
	}
	if r.ParserType != nil {
		config.ParserType = *r.ParserType
	}
	if r.FunctionSignature != nil {
		config.FunctionSignature = *r.FunctionSignature
	}
	if r.FunctionName != nil {
		config.FunctionName = *r.FunctionName
	}
	if r.FunctionDescription != nil {
		config.FunctionDescription = *r.FunctionDescription
	}
	if r.ParamConfig != nil {
		config.ParamConfig = models.ParamConfigs(*r.ParamConfig)
	}
	if r.ParserRules != nil {
		config.ParserRules = models.ParserRulesJSON(*r.ParserRules)
	}
	if r.DisplayFormat != nil {
		config.DisplayFormat = *r.DisplayFormat
	}
	if r.IsActive != nil {
		config.IsActive = *r.IsActive
	}
	if r.Priority != nil {
		config.Priority = *r.Priority
	}
	// 日志解析配置字段
	if r.LogsParserType != nil {
		config.LogsParserType = *r.LogsParserType
	}
	if r.EventSignature != nil {
		config.EventSignature = *r.EventSignature
	}
	if r.EventName != nil {
		config.EventName = *r.EventName
	}
	if r.EventDescription != nil {
		config.EventDescription = *r.EventDescription
	}
	if r.LogsParamConfig != nil {
		config.LogsParamConfig = *r.LogsParamConfig
	}
	if r.LogsParserRules != nil {
		config.LogsParserRules = models.LogsParserRulesJSON(*r.LogsParserRules)
	}
	if r.LogsDisplayFormat != nil {
		config.LogsDisplayFormat = *r.LogsDisplayFormat
	}
}

// ParserConfigResponse 解析配置响应
type ParserConfigResponse struct {
	ID                  uint                 `json:"id"`
	ContractAddress     string               `json:"contract_address"`
	ParserType          models.ParserType    `json:"parser_type"`
	FunctionSignature   string               `json:"function_signature"`
	FunctionName        string               `json:"function_name"`
	FunctionDescription string               `json:"function_description"`
	ParamConfig         []models.ParamConfig `json:"param_config"`
	ParserRules         models.ParserRules   `json:"parser_rules"`
	DisplayFormat       string               `json:"display_format"`
	IsActive            bool                 `json:"is_active"`
	Priority            int                  `json:"priority"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

// NewParserConfigResponse 创建解析配置响应
func NewParserConfigResponse(config *models.ParserConfig) *ParserConfigResponse {
	return &ParserConfigResponse{
		ID:                  config.ID,
		ContractAddress:     config.ContractAddress,
		ParserType:          config.ParserType,
		FunctionSignature:   config.FunctionSignature,
		FunctionName:        config.FunctionName,
		FunctionDescription: config.FunctionDescription,
		ParamConfig:         []models.ParamConfig(config.ParamConfig),
		ParserRules:         models.ParserRules(config.ParserRules),
		DisplayFormat:       config.DisplayFormat,
		IsActive:            config.IsActive,
		Priority:            config.Priority,
		CreatedAt:           config.CreatedAt,
		UpdatedAt:           config.UpdatedAt,
	}
}

// PaginationResponse 分页信息响应
type PaginationResponse struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalPages  int   `json:"total_pages"`
	TotalCount  int64 `json:"total_count"`
}

// NewPaginationResponse 创建分页信息响应
func NewPaginationResponse(total int64, page, pageSize int) *PaginationResponse {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &PaginationResponse{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalCount:  total,
	}
}

// ParserConfigListResponse 解析配置列表响应
type ParserConfigListResponse struct {
	ParserConfigs []*ParserConfigResponse `json:"parser_configs"`
	Pagination    *PaginationResponse     `json:"pagination"`
}

// NewParserConfigListResponse 创建解析配置列表响应
func NewParserConfigListResponse(configs []*models.ParserConfig, total int64, page, pageSize int) *ParserConfigListResponse {
	responses := make([]*ParserConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = NewParserConfigResponse(config)
	}

	return &ParserConfigListResponse{
		ParserConfigs: responses,
		Pagination:    NewPaginationResponse(total, page, pageSize),
	}
}

// ContractParserInfoResponse 合约解析信息响应
type ContractParserInfoResponse struct {
	Contract      interface{}             `json:"contract"` // 使用interface{}避免循环依赖
	CoinConfig    *CoinConfigResponse     `json:"coin_config"`
	ParserConfigs []*ParserConfigResponse `json:"parser_configs"`
}

// NewContractParserInfoResponse 创建合约解析信息响应
func NewContractParserInfoResponse(info *models.ContractParserInfo) *ContractParserInfoResponse {
	response := &ContractParserInfoResponse{}

	if info.Contract != nil {
		// 直接转换为map避免循环依赖
		response.Contract = map[string]interface{}{
			"id":             info.Contract.ID,
			"address":        info.Contract.Address,
			"chain_name":     info.Contract.ChainName,
			"contract_type":  info.Contract.ContractType,
			"name":           info.Contract.Name,
			"symbol":         info.Contract.Symbol,
			"decimals":       info.Contract.Decimals,
			"total_supply":   info.Contract.TotalSupply,
			"is_erc20":       info.Contract.IsERC20,
			"status":         info.Contract.Status,
			"verified":       info.Contract.Verified,
			"creator":        info.Contract.Creator,
			"creation_tx":    info.Contract.CreationTx,
			"creation_block": info.Contract.CreationBlock,
			"created_at":     info.Contract.CTime,
			"updated_at":     info.Contract.MTime,
		}
	}

	if info.CoinConfig != nil {
		response.CoinConfig = NewCoinConfigResponse(info.CoinConfig)
	}

	if len(info.ParserConfigs) > 0 {
		response.ParserConfigs = make([]*ParserConfigResponse, len(info.ParserConfigs))
		for i, config := range info.ParserConfigs {
			response.ParserConfigs[i] = NewParserConfigResponse(config)
		}
	}

	return response
}
