package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// AuthorizedAddressesJSON 授权地址JSON类型（实现JSON存储）
type AuthorizedAddressesJSON []string

func (a AuthorizedAddressesJSON) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *AuthorizedAddressesJSON) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into AuthorizedAddressesJSON", value)
	}

	return json.Unmarshal(bytes, a)
}

// UserAddress 用户地址模型
type UserAddress struct {
	ID                          uint                    `json:"id" gorm:"primaryKey"`
	UserID                      uint                    `json:"user_id" gorm:"not null;index"`
	Address                     string                  `json:"address" gorm:"type:varchar(120);not null"`
	Label                       string                  `json:"label" gorm:"type:varchar(100)"`
	Type                        string                  `json:"type" gorm:"type:varchar(20);not null;default:'wallet'"` // wallet, contract, exchange, other
	ContractID                  *uint                   `json:"contract_id" gorm:"index"`                               // 关联的合约ID，仅当type为contract时有效
	AuthorizedAddresses         AuthorizedAddressesJSON `json:"authorized_addresses" gorm:"type:json"`                  // 授权地址列表，JSON数组格式，仅当type为contract时有效
	Notes                       string                  `json:"notes" gorm:"type:text"`                                 // 备注信息
	Balance                     *string                 `json:"balance" gorm:"type:decimal(65,0);default:0"`            // 地址余额，以最小单位存储（如wei）
	ContractBalance             *string                 `json:"contract_balance" gorm:"type:decimal(65,0)"`             // 合约余额，以最小单位存储（如wei）
	ContractBalanceHeight       *uint64                 `json:"contract_balance_height"`                                // 合约余额最后更新的区块高度
	ContractBalanceReceiptIndex *uint64                 `json:"contract_balance_in_tx_index"`                           // 合约余额更新时的交易哈希
	TransactionCount            int64                   `json:"transaction_count" gorm:"default:0"`
	IsActive                    bool                    `json:"is_active" gorm:"default:true"`
	BalanceHeight               uint64                  `json:"balance_height" gorm:"default:0"` // 地址余额对应的区块高度
	CreatedAt                   time.Time               `json:"created_at"`
	UpdatedAt                   time.Time               `json:"updated_at"`

	// 关联关系
	User     User      `json:"user" gorm:"foreignKey:UserID"`
	Contract *Contract `json:"contract,omitempty" gorm:"foreignKey:ContractID"`
}

// TableName 指定表名
func (UserAddress) TableName() string {
	return "user_addresses"
}
