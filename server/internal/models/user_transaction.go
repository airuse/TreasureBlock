package models

import (
	"time"

	"gorm.io/gorm"
)

// UserTransaction 用户交易表 - 存储用户创建的待签名交易
type UserTransaction struct {
	ID     uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID uint64 `json:"user_id" gorm:"not null;index;comment:用户ID"`
	Chain  string `json:"chain" gorm:"type:varchar(20);not null;index;comment:链类型(btc,eth)"`
	Symbol string `json:"symbol" gorm:"type:varchar(20);not null;comment:币种"`

	// 交易基本信息
	FromAddress string `json:"from_address" gorm:"type:varchar(120);not null;index;comment:发送地址"`
	ToAddress   string `json:"to_address" gorm:"type:varchar(120);not null;comment:接收地址"`
	Amount      string `json:"amount" gorm:"type:decimal(65,0);not null;comment:交易金额"`
	Fee         string `json:"fee" gorm:"type:decimal(65,0);not null;default:0;comment:手续费"`

	// ETH特有字段
	GasLimit *uint   `json:"gas_limit" gorm:"comment:Gas限制"`
	GasPrice *string `json:"gas_price" gorm:"type:varchar(100);comment:Gas价格"`
	Nonce    *uint64 `json:"nonce" gorm:"comment:交易序号"`

	// EIP-1559费率字段
	MaxPriorityFeePerGas *string `json:"max_priority_fee_per_gas" gorm:"type:varchar(100);comment:最大优先费用(Gwei)"`
	MaxFeePerGas         *string `json:"max_fee_per_gas" gorm:"type:varchar(100);comment:最大费用(Gwei)"`

	// 交易状态
	Status string  `json:"status" gorm:"type:varchar(20);not null;default:'draft';index;comment:状态:draft,unsigned,in_progress,packed,confirmed,failed"`
	TxHash *string `json:"tx_hash" gorm:"type:varchar(120);comment:交易哈希"`

	// 签名相关
	UnsignedTx *string `json:"unsigned_tx" gorm:"type:longtext;comment:未签名交易数据"`
	SignedTx   *string `json:"signed_tx" gorm:"type:longtext;comment:已签名交易数据"`

	// QR码导出相关字段
	ChainID    *string `json:"chain_id" gorm:"type:varchar(10);comment:链ID"`
	TxData     *string `json:"tx_data" gorm:"type:longtext;comment:交易数据(十六进制)"`
	AccessList *string `json:"access_list" gorm:"type:longtext;comment:访问列表(JSON格式)"`

	// 签名组件
	V *string `json:"v" gorm:"type:varchar(100);comment:签名V组件"`
	R *string `json:"r" gorm:"type:varchar(100);comment:签名R组件"`
	S *string `json:"s" gorm:"type:varchar(100);comment:签名S组件"`

	// 交易结果
	BlockHeight   *uint64 `json:"block_height" gorm:"comment:区块高度"`
	Confirmations *uint   `json:"confirmations" gorm:"default:0;comment:确认数"`
	ErrorMsg      *string `json:"error_msg" gorm:"type:text;comment:错误信息"`

	// 备注
	Remark string `json:"remark" gorm:"type:varchar(500);comment:备注"`

	// 代币交易相关字段
	TransactionType       string `json:"transaction_type" gorm:"type:varchar(20);default:'coin';comment:交易类型(coin,token)"`
	ContractOperationType string `json:"contract_operation_type" gorm:"type:varchar(20);comment:合约操作类型(transfer,approve,transferFrom,balanceOf)"`
	TokenContractAddress  string `json:"token_contract_address" gorm:"type:varchar(120);comment:代币合约地址"`
	AllowanceAddress      string `json:"allowance_address" gorm:"type:varchar(120);comment:授权地址（代币持有者地址）"`

	// 代币信息字段（非数据库字段，仅用于API响应）
	TokenName     string `json:"token_name,omitempty" gorm:"-"`     // 代币全名（非数据库字段）
	TokenDecimals *uint8 `json:"token_decimals,omitempty" gorm:"-"` // 代币精度（非数据库字段）

	// 时间字段
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (UserTransaction) TableName() string {
	return "user_transactions"
}

// IsDraft 检查是否为草稿状态
func (t *UserTransaction) IsDraft() bool {
	return t.Status == "draft"
}

// IsUnsigned 检查是否为未签名状态
func (t *UserTransaction) IsUnsigned() bool {
	return t.Status == "unsigned"
}

// IsUnsent 检查是否为未发送状态
func (t *UserTransaction) IsUnsent() bool {
	return t.Status == "in_progress"
}

// IsInProgress 检查是否为在途状态
func (t *UserTransaction) IsInProgress() bool {
	return t.Status == "in_progress"
}

// IsPacked 检查是否为已打包状态
func (t *UserTransaction) IsPacked() bool {
	return t.Status == "packed"
}

// IsConfirmed 检查是否为已确认状态
func (t *UserTransaction) IsConfirmed() bool {
	return t.Status == "confirmed"
}

// IsFailed 检查是否为失败状态
func (t *UserTransaction) IsFailed() bool {
	return t.Status == "failed"
}

// CanExport 检查是否可以导出
func (t *UserTransaction) CanExport() bool {
	return t.IsDraft() || t.IsUnsigned()
}

// CanSend 检查是否可以发送
func (t *UserTransaction) CanSend() bool {
	return t.IsUnsent()
}

// CanView 检查是否可以查看详情
func (t *UserTransaction) CanView() bool {
	return t.IsInProgress() || t.IsPacked() || t.IsConfirmed() || t.IsFailed()
}
