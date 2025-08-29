package models

import (
	"time"

	"gorm.io/gorm"
)

// ContractParseResult 保存交易日志解析后的关键信息
type ContractParseResult struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	TxHash          string         `json:"tx_hash" gorm:"size:66;index;not null;comment:交易哈希"`
	ContractAddress string         `json:"contract_address" gorm:"size:66;index;comment:合约地址"`
	Chain           string         `json:"chain" gorm:"size:16;index;comment:链"`
	BlockNumber     uint64         `json:"block_number" gorm:"index;comment:区块号"`
	LogIndex        uint           `json:"log_index" gorm:"index;comment:日志索引"`
	EventSignature  string         `json:"event_signature" gorm:"size:66;index;comment:事件签名topics[0]"`
	EventName       string         `json:"event_name" gorm:"size:64;comment:事件名(可选)"`
	FromAddress     string         `json:"from_address" gorm:"size:66;index;comment:解析出的from地址"`
	ToAddress       string         `json:"to_address" gorm:"size:66;index;comment:解析出的to地址"`
	AmountWei       string         `json:"amount_wei" gorm:"type:varchar(100);default:'0';comment:解析出的amount(wei)十进制字符串"`
	TokenDecimals   uint16         `json:"token_decimals" gorm:"default:0;comment:代币精度(可选)"`
	TokenSymbol     string         `json:"token_symbol" gorm:"size:32;comment:代币符号(可选)"`
	RawLogsHash     string         `json:"raw_logs_hash" gorm:"size:66;comment:原始logs的hash,用于幂等"`
	ParsedJSON      string         `json:"parsed_json" gorm:"type:longtext;comment:解析出的原始JSON快照"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (ContractParseResult) TableName() string {
	return "contract_parse_result"
}
