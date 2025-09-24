package models

import (
	"time"
)

// Contract 合约模型
type Contract struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Address       string    `json:"address" gorm:"uniqueIndex;not null;type:varchar(120)"` // 合约地址
	ChainName     string    `json:"chain_name" gorm:"not null;type:varchar(20)"`           // 链名称
	ProgramID     string    `json:"program_id" gorm:"type:varchar(120)"`                   // 关联的程序ID（如Sol Program ID）
	ContractType  string    `json:"contract_type" gorm:"not null;type:varchar(50)"`        // 合约类型
	Name          string    `json:"name" gorm:"type:varchar(100)"`                         // 合约名称
	Symbol        string    `json:"symbol" gorm:"type:varchar(20)"`                        // 合约符号
	Decimals      uint8     `json:"decimals" gorm:"default:0"`                             // 小数位数（ERC-20）
	TotalSupply   string    `json:"total_supply" gorm:"type:varchar(100)"`                 // 总供应量
	IsERC20       bool      `json:"is_erc20" gorm:"default:false"`                         // 是否为ERC-20代币
	Interfaces    string    `json:"interfaces" gorm:"type:text"`                           // 支持的接口（JSON格式）
	Methods       string    `json:"methods" gorm:"type:text"`                              // 可调用的方法（JSON格式）
	Events        string    `json:"events" gorm:"type:text"`                               // 支持的事件（JSON格式）
	Metadata      string    `json:"metadata" gorm:"type:text"`                             // 其他元数据（JSON格式）
	Status        int8      `json:"status" gorm:"default:1"`                               // 状态：1-启用，0-禁用
	Verified      bool      `json:"verified" gorm:"default:false"`                         // 是否已验证
	Creator       string    `json:"creator" gorm:"type:varchar(120)"`                      // 创建者地址
	CreationTx    string    `json:"creation_tx" gorm:"type:varchar(120)"`                  // 创建交易哈希
	CreationBlock uint64    `json:"creation_block" gorm:"default:0"`                       // 创建区块高度
	ContractLogo  string    `json:"contract_logo" gorm:"type:longtext"`                    // 合约Logo图片(Base64编码)
	CTime         time.Time `json:"ctime" gorm:"autoCreateTime"`                           // 创建时间
	MTime         time.Time `json:"mtime" gorm:"autoUpdateTime"`                           // 更新时间
}

// TableName 指定表名
func (Contract) TableName() string {
	return "contracts"
}

// ContractType 合约类型常量
const (
	ContractTypeUnknown    = "unknown"
	ContractTypeERC20      = "ERC-20"
	ContractTypeERC721     = "ERC-721"
	ContractTypeERC1155    = "ERC-1155"
	ContractTypeProxy      = "Proxy"
	ContractTypeFactory    = "Factory"
	ContractTypeRouter     = "Router"
	ContractTypePool       = "Pool"
	ContractTypeVault      = "Vault"
	ContractTypeDAO        = "DAO"
	ContractTypeGame       = "Game"
	ContractTypeBridge     = "Bridge"
	ContractTypeDeFi       = "DeFi"
	ContractTypeStaking    = "Staking"
	ContractTypeMintable   = "Mintable"
	ContractTypeGovernance = "Governance"
)

// ContractStatus 合约状态常量
const (
	ContractStatusDisabled  = 0
	ContractStatusEnabled   = 1
	ContractStatusPaused    = 2
	ContractStatusUpgrading = 3
)
