package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ParserType 解析类型
type ParserType string

const (
	ParserTypeInputData ParserType = "input_data"
	ParserTypeEventLog  ParserType = "event_log"
)

// ParamConfig 参数配置结构
type ParamConfig struct {
	Name        string `json:"name"`        // 参数名称
	Type        string `json:"type"`        // 参数类型 (address, uint256, bytes等)
	Offset      int    `json:"offset"`      // 在input data中的偏移量
	Length      int    `json:"length"`      // 参数长度
	Description string `json:"description"` // 参数描述
}

// ParserRules 解析规则结构
type ParserRules struct {
	ExtractToAddress string `json:"extract_to_address,omitempty"` // 提取收款地址的规则
	ExtractAmount    string `json:"extract_amount,omitempty"`     // 提取金额的规则
	AmountUnit       string `json:"amount_unit,omitempty"`        // 金额单位
	ExtractData      string `json:"extract_data,omitempty"`       // 提取其他数据的规则
}

// ParamConfigs 参数配置数组类型（实现JSON存储）
type ParamConfigs []ParamConfig

func (p ParamConfigs) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.Marshal(p)
}

func (p *ParamConfigs) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ParamConfigs", value)
	}

	return json.Unmarshal(bytes, p)
}

// ParserRulesJSON 解析规则JSON类型（实现JSON存储）
type ParserRulesJSON ParserRules

func (p ParserRulesJSON) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ParserRulesJSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ParserRulesJSON", value)
	}

	return json.Unmarshal(bytes, p)
}

// LogsParamConfig 日志参数配置结构
type LogsParamConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TopicIndex  *int   `json:"topic_index,omitempty"`
	DataIndex   *int   `json:"data_index,omitempty"`
	Description string `json:"description"`
}

// LogsParamConfigs 日志参数配置数组类型（实现JSON存储）
type LogsParamConfigs []LogsParamConfig

func (p LogsParamConfigs) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.Marshal(p)
}

func (p *LogsParamConfigs) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into LogsParamConfigs", value)
	}

	return json.Unmarshal(bytes, p)
}

// LogsParserRules 日志解析规则结构
type LogsParserRules struct {
	ExtractFromAddress    string `json:"extract_from_address,omitempty"`
	ExtractToAddress      string `json:"extract_to_address,omitempty"`
	ExtractAmount         string `json:"extract_amount,omitempty"`
	AmountUnit            string `json:"amount_unit,omitempty"`
	ExtractOwnerAddress   string `json:"extract_owner_address,omitempty"`
	ExtractSpenderAddress string `json:"extract_spender_address,omitempty"`
}

// LogsParserRulesJSON 日志解析规则JSON类型（实现JSON存储）
type LogsParserRulesJSON LogsParserRules

func (p LogsParserRulesJSON) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *LogsParserRulesJSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into LogsParserRulesJSON", value)
	}

	return json.Unmarshal(bytes, p)
}

// ParserConfig 解析配置模型
type ParserConfig struct {
	ID                  uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	ContractAddress     string          `json:"contract_address" gorm:"column:contract_address;size:42;not null;index:idx_contract_sig"`
	ParserType          ParserType      `json:"parser_type" gorm:"column:parser_type;type:enum('input_data','event_log');not null;index:idx_parser_type"`
	FunctionSignature   string          `json:"function_signature" gorm:"column:function_signature;size:10;not null;index:idx_contract_sig,idx_function_sig"`
	FunctionName        string          `json:"function_name" gorm:"column:function_name;size:100;not null"`
	FunctionDescription string          `json:"function_description" gorm:"column:function_description;size:500;not null"`
	ParamConfig         ParamConfigs    `json:"param_config" gorm:"column:param_config;type:json"`
	ParserRules         ParserRulesJSON `json:"parser_rules" gorm:"column:parser_rules;type:json"`
	DisplayFormat       string          `json:"display_format" gorm:"column:display_format;size:500"`
	IsActive            bool            `json:"is_active" gorm:"column:is_active;default:true;index:idx_active_priority"`
	Priority            int             `json:"priority" gorm:"column:priority;default:0;index:idx_active_priority"`
	CreatedAt           time.Time       `json:"created_at" gorm:"autoCreateTime;column:created_at"`
	UpdatedAt           time.Time       `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`

	// 日志解析相关字段
	LogsParserType    string              `json:"logs_parser_type,omitempty" gorm:"column:logs_parser_type;type:enum('input_data','event_log','both');default:'input_data'"`
	EventSignature    string              `json:"event_signature,omitempty" gorm:"column:event_signature;size:66"`
	EventName         string              `json:"event_name,omitempty" gorm:"column:event_name;size:100"`
	EventDescription  string              `json:"event_description,omitempty" gorm:"column:event_description;size:255"`
	LogsParamConfig   LogsParamConfigs    `json:"logs_param_config,omitempty" gorm:"column:logs_param_config;type:json"`
	LogsParserRules   LogsParserRulesJSON `json:"logs_parser_rules,omitempty" gorm:"column:logs_parser_rules;type:json"`
	LogsDisplayFormat string              `json:"logs_display_format,omitempty" gorm:"column:logs_display_format;size:255"`
}

// TableName 指定表名
func (ParserConfig) TableName() string {
	return "parser_config"
}

// BeforeCreate GORM钩子：创建前处理
func (p *ParserConfig) BeforeCreate(tx *gorm.DB) error {
	// 验证必填字段
	if p.ContractAddress == "" {
		return fmt.Errorf("contract_address is required")
	}
	if p.FunctionSignature == "" {
		return fmt.Errorf("function_signature is required")
	}
	if p.FunctionName == "" {
		return fmt.Errorf("function_name is required")
	}
	return nil
}

// BeforeUpdate GORM钩子：更新前处理
func (p *ParserConfig) BeforeUpdate(tx *gorm.DB) error {
	// 只验证必填字段，不修改时间字段
	if p.ContractAddress == "" {
		return fmt.Errorf("contract_address is required")
	}
	if p.FunctionSignature == "" {
		return fmt.Errorf("function_signature is required")
	}
	if p.FunctionName == "" {
		return fmt.Errorf("function_name is required")
	}
	return nil
}

// ContractParserInfo 合约解析信息（用于获取完整的合约解析配置）
type ContractParserInfo struct {
	Contract      *Contract       `json:"contract"`
	CoinConfig    *CoinConfig     `json:"coin_config"`
	ParserConfigs []*ParserConfig `json:"parser_configs"`
}
