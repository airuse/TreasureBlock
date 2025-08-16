package models

import (
	"time"

	"gorm.io/gorm"
)

// Address 地址模型 - 支持BTC和ETH的通用模型
type Address struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Address string `json:"address" gorm:"type:varchar(120);uniqueIndex;not null"`
	Chain   string `json:"chain" gorm:"type:varchar(20);not null;index;comment:链类型(btc,eth)"`
	Type    uint16 `json:"type"`
	Nonce   string `json:"nonce"`
	Hash    string `json:"hash" gorm:"type:varchar(1024);not null"`

	// BTC特有字段
	UTXOCount uint64 `json:"utxo_count,omitempty" gorm:"default:0;comment:UTXO数量"`
	Balance   string `json:"balance,omitempty" gorm:"type:decimal(65,18);default:0;comment:余额"`

	// ETH特有字段
	TransactionCount uint64 `json:"transaction_count,omitempty" gorm:"default:0;comment:交易数量"`
	ContractCount    uint64 `json:"contract_count,omitempty" gorm:"default:0;comment:合约数量"`

	// 通用字段
	Mtime     time.Time      `json:"mtime"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Address) TableName() string {
	return "address"
}

// IsBTC 检查是否为BTC地址
func (a *Address) IsBTC() bool {
	return a.Chain == "btc"
}

// IsETH 检查是否为ETH地址
func (a *Address) IsETH() bool {
	return a.Chain == "eth"
}

// GetChainType 获取链类型
func (a *Address) GetChainType() string {
	return a.Chain
}
