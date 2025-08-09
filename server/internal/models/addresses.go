package models

import (
	"time"

	"gorm.io/gorm"
)

// Address 地址模型 - 纯数据模型，不包含验证逻辑
type Address struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Address   string         `json:"address" gorm:"type:varchar(120);uniqueIndex;not null"`
	Type      uint16         `json:"type"`
	Nonce     string         `json:"nonce"`
	Hash      string         `json:"hash" gorm:"type:varchar(1024);not null"`
	Mtime     time.Time      `json:"mtime"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (Address) TableName() string {
	return "address"
}
