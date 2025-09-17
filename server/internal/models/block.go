package models

import (
	"time"

	"gorm.io/gorm"
)

// Block 区块模型 - 支持BTC和ETH的通用模型
type Block struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Hash             string    `json:"hash" gorm:"type:varchar(255);index;not null"`
	Height           uint64    `json:"height" gorm:"index:idx_block_verified_chain_height,priority:3;not null"`
	PreviousHash     string    `json:"previous_hash" gorm:"type:char(66);index"`
	Timestamp        time.Time `json:"timestamp" gorm:"index:idx_block_time_chain_timestamp,priority:2"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan" gorm:"default:false"`
	Chain            string    `json:"chain" gorm:"type:varchar(20);index:idx_block_time_chain_timestamp,priority:1;index:idx_block_chain_deleted,priority:1;index:idx_block_verified_chain_height,priority:1;index:idx_block_base_fee_chain_height,priority:1;not null"` // btc, eth
	ChainID          int       `json:"chain_id" gorm:"type:int;index;not null"`
	// BTC特有字段
	MerkleRoot string `json:"merkle_root,omitempty" gorm:"type:char(66)"`
	Bits       string `json:"bits,omitempty" gorm:"type:varchar(20)"`
	Version    uint32 `json:"version,omitempty"`
	Weight     uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit   uint64 `json:"gas_limit,omitempty"`
	GasUsed    uint64 `json:"gas_used,omitempty"`
	Miner      string `json:"miner,omitempty" gorm:"type:varchar(120)"`
	ParentHash string `json:"parent_hash,omitempty" gorm:"type:char(66)"`
	Nonce      string `json:"nonce,omitempty" gorm:"type:varchar(20)"`
	Difficulty string `json:"difficulty,omitempty" gorm:"type:varchar(50)"`

	// ETH状态根字段
	StateRoot        string `json:"state_root,omitempty" gorm:"type:char(66);comment:状态根哈希"`
	TransactionsRoot string `json:"transactions_root,omitempty" gorm:"type:char(66);comment:交易根哈希"`
	ReceiptsRoot     string `json:"receipts_root,omitempty" gorm:"type:char(66);comment:收据根哈希"`

	// ETH London 升级相关字段
	BaseFee     string `json:"base_fee,omitempty" gorm:"type:varchar(100);index:idx_block_base_fee_chain_height,priority:2;comment:基础费"` // wei，字符串存储
	BurnedEth   string `json:"burned_eth,omitempty" gorm:"type:varchar(100);comment:燃烧费"`                                                // ETH数量字符串
	MinerTipEth string `json:"miner_tip_eth,omitempty" gorm:"type:varchar(100);comment:款工收益"`                                            // ETH数量字符串

	// 验证相关字段
	VerificationDeadline *time.Time `json:"verification_deadline" gorm:"type:timestamp;column:verification_deadline;comment:最晚验证时间"`
	IsVerified           uint8      `json:"is_verified" gorm:"type:tinyint(1);not null;default:0;index:idx_block_verified_chain_height,priority:2;column:is_verified;comment:验证是否通过 0:未验证 1:验证通过 2:验证失败"`

	// 通用时间字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index:idx_block_chain_deleted,priority:2"`
}

// TableName 指定表名
func (Block) TableName() string {
	return "blocks"
}

// IsBTC 检查是否为BTC区块
func (b *Block) IsBTC() bool {
	return b.Chain == "btc"
}

// IsETH 检查是否为ETH区块
func (b *Block) IsETH() bool {
	return b.Chain == "eth"
}

// GetChainType 获取链类型
func (b *Block) GetChainType() string {
	return b.Chain
}
