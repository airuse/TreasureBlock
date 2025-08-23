package models

import (
	"time"

	"gorm.io/gorm"
)

type CoinConfig struct {
	ID            uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ChainName     string         `gorm:"type:varchar(20);not null;column:chain_name;comment:链名称" json:"chain_name"`
	Symbol        string         `gorm:"type:varchar(20);not null;column:symbol;comment:币种符号" json:"symbol"`
	CoinType      uint8          `gorm:"type:tinyint(3) unsigned;not null;default:1;column:coin_type;comment:币种类型: 0=原生币, 1=ERC20, 2=ERC223, 3=ERC721, 4=ERC1155" json:"coin_type"`
	ContractAddr  string         `gorm:"type:varchar(120);not null;default:'';column:contract_addr;comment:合约地址(原生币为空)" json:"contract_addr"`
	Precision     uint           `gorm:"type:tinyint(3) unsigned;not null;default:18;column:precision;comment:精度(小数位数)" json:"precision"`
	Decimals      uint           `gorm:"type:tinyint(3) unsigned;not null;default:18;column:decimals;comment:精度别名(兼容性)" json:"decimals"`
	Name          string         `gorm:"type:varchar(100);not null;column:name;comment:币种全名" json:"name"`
	LogoURL       string         `gorm:"type:varchar(255);default:'';column:logo_url;comment:币种Logo URL" json:"logo_url"`
	WebsiteURL    string         `gorm:"type:varchar(255);default:'';column:website_url;comment:官方网站" json:"website_url"`
	ExplorerURL   string         `gorm:"type:varchar(255);default:'';column:explorer_url;comment:区块浏览器地址" json:"explorer_url"`
	Description   string         `gorm:"type:text;column:description;comment:币种描述" json:"description"`
	MarketCapRank uint           `gorm:"type:int(11) unsigned;default:0;column:market_cap_rank;comment:市值排名" json:"market_cap_rank"`
	IsStablecoin  bool           `gorm:"type:boolean;not null;default:false;column:is_stablecoin;comment:是否为稳定币" json:"is_stablecoin"`
	IsVerified    bool           `gorm:"type:boolean;not null;default:true;column:is_verified;comment:是否已验证" json:"is_verified"`
	Status        int8           `gorm:"type:tinyint(3);not null;default:1;column:status;comment:状态: 0=禁用, 1=启用" json:"status"`
	CreatedAt     time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:updated_at;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (CoinConfig) TableName() string {
	return "coin_config"
}

// IsNative 检查是否为原生币
func (c *CoinConfig) IsNative() bool {
	return c.CoinType == 0
}

// IsERC20 检查是否为ERC-20代币
func (c *CoinConfig) IsERC20() bool {
	return c.CoinType == 1
}

// IsERC721 检查是否为ERC-721代币
func (c *CoinConfig) IsERC721() bool {
	return c.CoinType == 3
}

// IsERC1155 检查是否为ERC-1155代币
func (c *CoinConfig) IsERC1155() bool {
	return c.CoinType == 4
}

// GetDisplayName 获取显示名称
func (c *CoinConfig) GetDisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Symbol
}

// GetLogoURL 获取Logo URL，如果没有则返回默认图标
func (c *CoinConfig) GetLogoURL() string {
	if c.LogoURL != "" {
		return c.LogoURL
	}
	// 返回默认图标
	return "https://cryptologos.cc/logos/question-mark.png"
}
