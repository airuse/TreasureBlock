package dto

import (
	"blockChainBrowser/server/internal/models"
)

// TransactionReceiptResponse 交易凭证响应DTO
type TransactionReceiptResponse struct {
	// 交易基本信息
	TxHash            string `json:"tx_hash"`
	TxType            uint8  `json:"tx_type"`
	Status            uint64 `json:"status"`
	GasUsed           uint64 `json:"gas_used"`
	EffectiveGasPrice string `json:"effective_gas_price"`
	BlobGasUsed       uint64 `json:"blob_gas_used"`
	BlobGasPrice      string `json:"blob_gas_price"`
	BlockHash         string `json:"block_hash"`
	BlockNumber       uint64 `json:"block_number"`
	TransactionIndex  uint   `json:"transaction_index"`
	Chain             string `json:"chain"`

	// 交易输入数据（用于前端解析）
	InputData string `json:"input_data,omitempty"`

	// 日志数据（用于前端解析）
	LogsData string `json:"logs_data,omitempty"`

	// 合约地址
	ContractAddress string `json:"contract_address,omitempty"`

	// 时间信息
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	// 解析配置数据（用于前端解析）
	ParserConfigs []ParserConfigInfo `json:"parser_configs,omitempty"`
}

// ParserConfigInfo 解析配置信息
type ParserConfigInfo struct {
	FunctionSignature   string                 `json:"function_signature"`
	FunctionName        string                 `json:"function_name"`
	FunctionDescription string                 `json:"function_description"`
	DisplayFormat       string                 `json:"display_format"`
	ParamConfig         []ParamConfigInfo      `json:"param_config,omitempty"`
	ParserRules         map[string]interface{} `json:"parser_rules,omitempty"`

	// 日志解析相关字段
	LogsParserType    string                 `json:"logs_parser_type,omitempty"`
	EventSignature    string                 `json:"event_signature,omitempty"`
	EventName         string                 `json:"event_name,omitempty"`
	EventDescription  string                 `json:"event_description,omitempty"`
	LogsParamConfig   []LogsParamConfigInfo  `json:"logs_param_config,omitempty"`
	LogsParserRules   map[string]interface{} `json:"logs_parser_rules,omitempty"`
	LogsDisplayFormat string                 `json:"logs_display_format,omitempty"`
}

// ParamConfigInfo 参数配置信息
type ParamConfigInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Offset      int    `json:"offset"`
	Length      int    `json:"length"`
	Description string `json:"description"`
}

// 日志参数配置信息
type LogsParamConfigInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TopicIndex  *int   `json:"topic_index,omitempty"` // 在topics中的索引
	DataIndex   *int   `json:"data_index,omitempty"`  // 在data中的索引
	Description string `json:"description"`
}

// NewTransactionReceiptResponse 创建交易凭证响应
func NewTransactionReceiptResponse(receipt *models.TransactionReceipt, tx *models.Transaction, parserConfigs []*models.ParserConfig) *TransactionReceiptResponse {
	response := &TransactionReceiptResponse{
		TxHash:            receipt.TxHash,
		TxType:            receipt.TxType,
		Status:            receipt.Status,
		GasUsed:           receipt.GasUsed,
		EffectiveGasPrice: receipt.EffectiveGasPrice,
		BlobGasUsed:       receipt.BlobGasUsed,
		BlobGasPrice:      receipt.BlobGasPrice,
		BlockHash:         receipt.BlockHash,
		BlockNumber:       receipt.BlockNumber,
		TransactionIndex:  receipt.TransactionIndex,
		Chain:             receipt.Chain,
		LogsData:          receipt.LogsData,
		ContractAddress:   receipt.ContractAddress,
		CreatedAt:         receipt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         receipt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 如果有关联的交易信息，添加输入数据
	if tx != nil && tx.Hex != nil {
		response.InputData = *tx.Hex
	}

	// 添加解析配置信息
	if len(parserConfigs) > 0 {
		response.ParserConfigs = make([]ParserConfigInfo, len(parserConfigs))
		for i, config := range parserConfigs {
			response.ParserConfigs[i] = ParserConfigInfo{
				FunctionSignature:   config.FunctionSignature,
				FunctionName:        config.FunctionName,
				FunctionDescription: config.FunctionDescription,
				DisplayFormat:       config.DisplayFormat,
				ParamConfig:         convertParamConfigs(config.ParamConfig),
				ParserRules:         convertParserRules(config.ParserRules),

				// 日志解析相关字段
				LogsParserType:    config.LogsParserType,
				EventSignature:    config.EventSignature,
				EventName:         config.EventName,
				EventDescription:  config.EventDescription,
				LogsParamConfig:   convertLogsParamConfigs(config.LogsParamConfig),
				LogsParserRules:   convertLogsParserRules(config.LogsParserRules),
				LogsDisplayFormat: config.LogsDisplayFormat,
			}
		}
	}

	return response
}

// convertParamConfigs 转换参数配置
func convertParamConfigs(paramConfigs []models.ParamConfig) []ParamConfigInfo {
	if len(paramConfigs) == 0 {
		return nil
	}

	result := make([]ParamConfigInfo, len(paramConfigs))
	for i, config := range paramConfigs {
		result[i] = ParamConfigInfo{
			Name:        config.Name,
			Type:        config.Type,
			Offset:      config.Offset,
			Length:      config.Length,
			Description: config.Description,
		}
	}
	return result
}

// convertParserRules 转换解析规则
func convertParserRules(parserRules models.ParserRulesJSON) map[string]interface{} {
	// 将 ParserRulesJSON 转换为 map[string]interface{}
	rules := map[string]interface{}{
		"extract_to_address": parserRules.ExtractToAddress,
		"extract_amount":     parserRules.ExtractAmount,
		"amount_unit":        parserRules.AmountUnit,
		"extract_data":       parserRules.ExtractData,
	}
	return rules
}

// convertLogsParamConfigs 转换日志参数配置
func convertLogsParamConfigs(logsParamConfigs []models.LogsParamConfig) []LogsParamConfigInfo {
	if len(logsParamConfigs) == 0 {
		return nil
	}

	result := make([]LogsParamConfigInfo, len(logsParamConfigs))
	for i, config := range logsParamConfigs {
		result[i] = LogsParamConfigInfo{
			Name:        config.Name,
			Type:        config.Type,
			TopicIndex:  config.TopicIndex,
			DataIndex:   config.DataIndex,
			Description: config.Description,
		}
	}
	return result
}

// convertLogsParserRules 转换日志解析规则
func convertLogsParserRules(logsParserRules models.LogsParserRulesJSON) map[string]interface{} {
	// 将 LogsParserRulesJSON 转换为 map[string]interface{}
	rules := map[string]interface{}{
		"extract_from_address":    logsParserRules.ExtractFromAddress,
		"extract_to_address":      logsParserRules.ExtractToAddress,
		"extract_amount":          logsParserRules.ExtractAmount,
		"amount_unit":             logsParserRules.AmountUnit,
		"extract_owner_address":   logsParserRules.ExtractOwnerAddress,
		"extract_spender_address": logsParserRules.ExtractSpenderAddress,
	}
	return rules
}
