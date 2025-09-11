package models

import (
	"time"

	"gorm.io/gorm"
)

// EarningsRecord 收益流水记录表
type EarningsRecord struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64         `gorm:"not null;index" json:"user_id"`      // 用户ID
	Amount           int64          `gorm:"not null" json:"amount"`             // 收益金额（T币数量，单位：个）
	Type             string         `gorm:"not null;size:20;index" json:"type"` // 类型：add（增加）、decrease（减少）
	Source           string         `gorm:"not null;size:50" json:"source"`     // 来源：block_verification（扫块验证）、transfer_out（转出）、business_consume（业务消耗）等
	SourceID         *uint64        `gorm:"index" json:"source_id"`             // 来源ID（如区块ID、交易ID等）
	SourceChain      string         `gorm:"size:50;index" json:"source_chain"`  // 来源链名称
	BlockHeight      *uint64        `gorm:"index" json:"block_height"`          // 相关区块高度
	TransactionCount *int64         `gorm:"default:0" json:"transaction_count"` // 相关交易数量（扫块时记录）
	Description      string         `gorm:"size:255" json:"description"`        // 描述信息
	BalanceBefore    int64          `gorm:"not null" json:"balance_before"`     // 操作前余额
	BalanceAfter     int64          `gorm:"not null" json:"balance_after"`      // 操作后余额
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (EarningsRecord) TableName() string {
	return "earnings_records"
}

// UserBalance 用户T币余额表
type UserBalance struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64         `gorm:"not null;uniqueIndex:idx_user_chain" json:"user_id"`                          // 用户ID，唯一索引（与链组合）
	SourceChain     string         `gorm:"size:50;not null;default:all;uniqueIndex:idx_user_chain" json:"source_chain"` // 余额归属链（如 btc/eth/all）
	Balance         int64          `gorm:"not null;default:0" json:"balance"`                                           // 当前余额（T币数量，单位：个）
	TotalEarned     int64          `gorm:"not null;default:0" json:"total_earned"`                                      // 累计获得的T币数量
	TotalSpent      int64          `gorm:"not null;default:0" json:"total_spent"`                                       // 累计消耗的T币数量
	LastEarningTime *time.Time     `json:"last_earning_time"`                                                           // 最后一次获得收益时间
	LastSpendTime   *time.Time     `json:"last_spend_time"`                                                             // 最后一次消耗时间
	Version         int64          `gorm:"not null;default:0" json:"version"`                                           // 版本号，用于乐观锁
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (UserBalance) TableName() string {
	return "user_balances"
}

// EarningsStats 收益统计（用于聚合查询）
type EarningsStats struct {
	UserID           uint64 `json:"user_id"`
	TotalEarnings    int64  `json:"total_earnings"`    // 总收益
	TotalSpendings   int64  `json:"total_spendings"`   // 总支出
	CurrentBalance   int64  `json:"current_balance"`   // 当前余额
	BlockCount       int64  `json:"block_count"`       // 扫块数量
	TransactionCount int64  `json:"transaction_count"` // 交易数量
}
