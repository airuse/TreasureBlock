package models

import (
	"time"

	"gorm.io/gorm"
)

// TransactionReceipt 交易凭证模型
type TransactionReceipt struct {
	ID                uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	TxHash            string         `json:"tx_hash" gorm:"uniqueIndex;size:66;not null;comment:交易哈希"`
	TxType            uint8          `json:"tx_type" gorm:"comment:交易类型"`
	PostState         string         `json:"post_state" gorm:"size:66;comment:执行后状态根"`
	Status            uint64         `json:"status" gorm:"comment:交易状态 1=成功 0=失败"`
	CumulativeGasUsed uint64         `json:"cumulative_gas_used" gorm:"comment:累计Gas使用量"`
	Bloom             string         `json:"bloom" gorm:"type:text;comment:布隆过滤器"`
	LogsData          string         `json:"logs_data" gorm:"type:longtext;comment:日志数据JSON"`
	ContractAddress   string         `json:"contract_address" gorm:"size:42;comment:合约地址(如果是合约创建交易)"`
	GasUsed           uint64         `json:"gas_used" gorm:"comment:实际Gas使用量"`
	BlockHash         string         `json:"block_hash" gorm:"size:66;index;comment:区块哈希"`
	BlockNumber       uint64         `json:"block_number" gorm:"index;comment:区块号"`
	TransactionIndex  uint           `json:"transaction_index" gorm:"comment:交易在区块中的索引"`
	Chain             string         `json:"chain" gorm:"size:10;index;comment:链名称"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (TransactionReceipt) TableName() string {
	return "transaction_receipts"
}
