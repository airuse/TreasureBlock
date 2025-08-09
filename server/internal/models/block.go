package models

import (
	"time"

	"gorm.io/gorm"
)

// Block 区块模型 - 纯数据模型，不包含验证逻辑
type Block struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Hash             string         `json:"hash" gorm:"type:char(64);uniqueIndex;not null"`
	Height           uint64         `json:"height" gorm:"uniqueIndex;not null"`
	PreviousHash     string         `json:"previous_hash" gorm:"type:char(64);index"`
	MerkleRoot       string         `json:"merkle_root" gorm:"type:char(64)"`
	Timestamp        time.Time      `json:"timestamp"`
	Difficulty       float64        `json:"difficulty"`
	Nonce            uint64         `json:"nonce"`
	Size             uint64         `json:"size"`
	TransactionCount int            `json:"transaction_count"`
	TotalAmount      float64        `json:"total_amount"`
	Fee              float64        `json:"fee"`
	Confirmations    uint64         `json:"confirmations"`
	IsOrphan         bool           `json:"is_orphan" gorm:"default:false"`
	Chain            string         `json:"chain" gorm:"type:varchar(20);index"` // btc, eth, etc.
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (Block) TableName() string {
	return "blocks"
}
