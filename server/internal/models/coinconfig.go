package models

import (
	"time"

	"gorm.io/gorm"
)

type CoinConfig struct {
	ID               uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ChainName        string         `gorm:"type:varchar(50);not null;column:chain_name;comment:链名称" json:"chain_name"`
	Symbol           string         `gorm:"type:varchar(50);not null;column:symbol;comment:币种" json:"symbol"`
	CoinType         uint8          `gorm:"type:tinyint(3) unsigned;not null;column:coin_type;comment:0：eth,1：ERC20，2：ERC223，3：ERC773,4:TRC10, 5:TRC20" json:"coin_type"`
	ContractAddr     string         `gorm:"type:varchar(120);not null;column:contract_addr;comment:合约地址" json:"contract_addr"`
	Precision        uint           `gorm:"type:int(11) unsigned;not null;column:precision;comment:精度，正整数<=18" json:"precision"`
	ColdAddress      string         `gorm:"type:varchar(200);not null;column:cold_address;comment:冷钱包地址" json:"cold_address"`
	ColdAddressHash  string         `gorm:"type:varchar(120);not null;column:cold_address_hash;comment:冷钱包地址的哈希，防篡改" json:"cold_address_hash"`
	MaxStock         float64        `gorm:"type:decimal(32,8);not null;default:0.00000000;column:max_stock;comment:最大库存，触发热转冷（单位：整数/个）" json:"max_stock"`
	MaxBalance       float64        `gorm:"type:decimal(32,8);not null;default:0.00000000;column:max_balance;comment:触发热转冷时，留下多少（单位：整数/个）" json:"max_balance"`
	MinBalance       float64        `gorm:"type:decimal(16,8);not null;default:0.00000000;column:min_balance;comment:报警值（单位：整数/个）" json:"min_balance"`
	CollectLimit     float64        `gorm:"type:decimal(16,8);not null;default:0.00000000;column:collect_limit;comment:归集的阈值（大于多少执行归集）（单位：整数/个）" json:"collect_limit"`
	CollectLeft      float64        `gorm:"type:decimal(16,8) unsigned;not null;column:collect_left;comment:归集留多少（单位：整数/个）" json:"collect_left"`
	InternalGasLimit uint           `gorm:"type:int(11) unsigned;not null;column:internal_gas_limit;comment:内部归集时的limit" json:"internal_gas_limit"`
	OnceMinFee       float64        `gorm:"type:float;default:0.003;column:once_min_fee;comment:归集预打手续费" json:"once_min_fee"`
	CTime            time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime;comment:入库时间" json:"ctime"`
	MTime            time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:mtime;comment:更改时间" json:"mtime"`
	SymbolID         string         `gorm:"type:varchar(100);default:'';column:symbol_id;comment:币种id" json:"symbol_id"`
	Status           int8           `gorm:"type:tinyint(3);not null;default:0;column:status;comment:0禁用 1启用" json:"status"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (CoinConfig) TableName() string {
	return "coin_config"
}
