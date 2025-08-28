package models

import (
	"time"
)

// UserAddress 用户地址模型
type UserAddress struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"not null;index"`
	Address          string    `json:"address" gorm:"type:varchar(120);not null"`
	Label            string    `json:"label" gorm:"type:varchar(100)"`
	Type             string    `json:"type" gorm:"type:varchar(20);not null;default:'wallet'"` // wallet, contract, exchange, other
	Balance          float64   `json:"balance" gorm:"type:decimal(20,8);default:0"`
	TransactionCount int64     `json:"transaction_count" gorm:"default:0"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	CreatedHeight    uint64    `json:"created_height" gorm:"default:0"` // 创建时的区块高度
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (UserAddress) TableName() string {
	return "user_addresses"
}
