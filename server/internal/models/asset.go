package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Asset 资产模型 - 纯数据模型，不包含验证逻辑
type Asset struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	Address    string          `json:"address" gorm:"type:char(64);uniqueIndex;not null"`
	Symbol     string          `json:"symbol" gorm:"type:varchar(10);not null"`
	Amount     decimal.Decimal `json:"amount" gorm:"not null;default:0.00000000000000000000;comment:资产数量"`
	Locked     decimal.Decimal `json:"locked" gorm:"not null;default:0.00000000000000000000;comment:锁定数量"`
	LastTxid   string          `json:"last_txid" gorm:"type:varchar(1024);not null;comment:最新交易ID"`
	LastHeight uint64          `json:"last_height" gorm:"not null;comment:最新区块高度"`
	Type       uint16          `json:"type" gorm:"not null;comment:资产类型"`
	Chain      string          `json:"chain" gorm:"type:varchar(20);index"` // btc, eth, etc.
	Ctime      time.Time       `json:"ctime" gorm:"not null;comment:创建时间"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
}

func (Asset) TableName() string {
	return "asset"
}
