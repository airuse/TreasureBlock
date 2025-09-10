package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction 交易流水表 - 支持BTC和ETH的通用模型
type Transaction struct {
	ID           uint    `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID         string  `json:"tx_id" gorm:"type:varchar(120);not null;column:tx_id;index;comment:txid"`
	TxType       uint8   `json:"tx_type" gorm:"type:tinyint(3) unsigned;not null;default:0;column:tx_type;comment:0:常规充值（100_1） 1:常规提现(0_100) 2:用户A提现到用户B充值地址(0_1) 3:打手续费（2_1） 4:归集(1_0) 5:系统地址充值、冷转热(100_0) 6:转冷(0_101) 7:外部向手续费充值 (100_2) 8:用户提现到手续费地址、特殊的手续费地址充值(0_2) 9: 系统地址充值手续费,提现token,无eth时触发(2_0)"`
	Confirm      uint    `json:"confirm" gorm:"type:int(11) unsigned;not null;column:confirm;comment:确认数"`
	Status       uint8   `json:"status" gorm:"type:tinyint(3) unsigned;not null;default:0;column:status;comment:0:未知，1:成功，2:失败,3：失败后已处理"`
	SendStatus   uint8   `json:"send_status" gorm:"type:tinyint(3) unsigned;not null;default:0;column:send_status;comment:广播状态，0：未广播，1：已广播"`
	Balance      string  `json:"balance" gorm:"type:decimal(65,0);not null;default:0;column:balance;comment:出币前余额"`
	Amount       string  `json:"amount" gorm:"type:decimal(65,0) unsigned;not null;default:0;column:amount;comment:交易额"`
	TransID      uint    `json:"trans_id" gorm:"type:int(11) unsigned;not null;column:trans_id;comment:提现ID"`
	Height       uint64  `json:"height" gorm:"type:bigint(20) unsigned;not null;column:height;comment:块高度"`
	ContractAddr string  `json:"contract_addr" gorm:"type:varchar(120);not null;column:contract_addr;comment:合约地址"`
	Hex          *string `json:"hex" gorm:"type:longtext;column:hex"`
	TxScene      string  `json:"tx_scene" gorm:"type:varchar(20);not null;default:0;column:tx_scene;comment:交易场景"`
	Remark       string  `json:"remark" gorm:"type:varchar(256);not null;default:'';column:remark;comment:备注"`

	// 链相关字段
	Chain  string `json:"chain" gorm:"type:varchar(20);not null;index;comment:链类型(btc,eth)"`
	Symbol string `json:"symbol" gorm:"type:varchar(20);not null;column:symbol;comment:币种"`

	// 地址字段
	AddressFrom string `json:"address_from" gorm:"type:varchar(120);not null;index;column:address_from;comment:发货人地址"`
	AddressTo   string `json:"address_to" gorm:"type:varchar(120);not null;index;column:address_to;comment:收货人地址"`

	// Gas相关字段（ETH特有，BTC可为空）
	GasLimit uint   `json:"gas_limit" gorm:"type:int(11) unsigned;not null;column:gas_limit;comment:燃油限制"`
	GasPrice string `json:"gas_price" gorm:"type:decimal(65,0) unsigned;not null;column:gas_price;comment:燃油价格"`
	GasUsed  uint   `json:"gas_used" gorm:"type:int(11) unsigned;not null;column:gas_used;comment:实际使用燃油价格"`

	// EIP-1559 相关字段（ETH特有）
	MaxFeePerGas         string `json:"max_fee_per_gas" gorm:"type:varchar(100);column:max_fee_per_gas;comment:最高费用(MaxFee)"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas" gorm:"type:varchar(100);column:max_priority_fee_per_gas;comment:最高小费(MaxPriorityFee)"`
	EffectiveGasPrice    string `json:"effective_gas_price" gorm:"type:varchar(100);column:effective_gas_price;comment:有效Gas价格"`

	// 手续费字段
	Fee     string  `json:"fee" gorm:"type:decimal(36,18) unsigned;not null;default:0.000000000000000000;column:fee;comment:预留手续费"`
	UsedFee *string `json:"used_fee" gorm:"type:decimal(36,18);column:used_fee;comment:真实手续费"`

	// 排序相关字段
	Nonce      uint64  `json:"nonce" gorm:"type:bigint(20) unsigned;not null;default:0;column:nonce;comment:交易序号（ETH）或输入索引（BTC）"`
	BlockIndex uint    `json:"block_index" gorm:"type:int(11) unsigned;not null;default:0;column:block_index;comment:交易在区块中的索引位置"`
	BlockID    *uint64 `json:"block_id" gorm:"type:bigint(20) unsigned;column:block_id;index;comment:关联的区块ID"`

	// 日志数据字段
	Logs string `json:"logs" gorm:"type:longtext;column:logs;comment:交易日志数据(JSON格式)"`

	// 代币标识字段（非数据库字段，仅用于API响应）
	IsToken            bool   `json:"is_token" gorm:"-"`
	TokenName          string `json:"token_name,omitempty" gorm:"-"`            // 代币全名（非数据库字段）
	TokenSymbol        string `json:"token_symbol,omitempty" gorm:"-"`          // 代币符号（非数据库字段）
	TokenDecimals      uint8  `json:"token_decimals,omitempty" gorm:"-"`        // 代币精度（非数据库字段）
	TokenDescription   string `json:"token_description,omitempty" gorm:"-"`     // 代币描述（非数据库字段）
	TokenWebsite       string `json:"token_website,omitempty" gorm:"-"`         // 代币官网（非数据库字段）
	TokenExplorer      string `json:"token_explorer,omitempty" gorm:"-"`        // 代币浏览器链接（非数据库字段）
	TokenLogo          string `json:"token_logo,omitempty" gorm:"-"`            // 代币Logo（非数据库字段）
	TokenMarketCapRank *int   `json:"token_market_cap_rank,omitempty" gorm:"-"` // 市值排名（非数据库字段）
	TokenIsStablecoin  bool   `json:"token_is_stablecoin,omitempty" gorm:"-"`   // 是否为稳定币（非数据库字段）
	TokenIsVerified    bool   `json:"token_is_verified,omitempty" gorm:"-"`     // 是否已验证（非数据库字段）

	// 时间字段
	Ctime     time.Time      `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime;index;comment:入库时间"`
	Mtime     time.Time      `json:"mtime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:mtime;comment:更改时间"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transaction"
}

// IsBTC 检查是否为BTC交易
func (t *Transaction) IsBTC() bool {
	return t.Chain == "btc"
}

// IsETH 检查是否为ETH交易
func (t *Transaction) IsETH() bool {
	return t.Chain == "eth"
}

// GetChainType 获取链类型
func (t *Transaction) GetChainType() string {
	return t.Chain
}
