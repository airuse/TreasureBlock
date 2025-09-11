package models

import "time"

// BTCTransaction 比特币原生交易表
type BTCTransaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关键标识
	TxID        string `gorm:"size:80;index" json:"tx_id"`
	BlockHash   string `gorm:"size:80;index" json:"block_hash"`
	BlockHeight uint64 `gorm:"index" json:"block_height"`

	// 常用字段
	From     string `gorm:"size:120;index" json:"from"`
	To       string `gorm:"size:120;index" json:"to"`
	Amount   string `gorm:"size:64" json:"amount"` // 金额（BTC，字符串表示）
	Fee      string `gorm:"size:64" json:"fee"`    // 费用（BTC，字符串表示）
	Size     uint   `json:"size"`
	Weight   uint   `json:"weight"`
	LockTime uint32 `json:"lock_time"`

	// 原始数据
	Hex      string `gorm:"type:longtext" json:"hex"`
	VinJSON  string `gorm:"type:longtext" json:"vin_json"`
	VoutJSON string `gorm:"type:longtext" json:"vout_json"`
}

func (BTCTransaction) TableName() string {
	return "btc_transaction"
}
