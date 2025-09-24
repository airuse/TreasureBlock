package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// JSONText 是用于存储 JSON 的通用字符串类型，支持数据库 Value/Scan
type JSONText string

func (j JSONText) Value() (driver.Value, error) {
	if j == "" {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONText) Scan(value interface{}) error {
	if value == nil {
		*j = ""
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = JSONText(string(v))
		return nil
	case string:
		*j = JSONText(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into JSONText", value)
	}
}

// SolTxDetail 存储 Solana 交易的明细（基于 jsonParsed 数据简化）
type SolTxDetail struct {
	ID                uint      `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID              string    `json:"tx_id" gorm:"type:varchar(120);not null;uniqueIndex;comment:交易签名"`
	Slot              uint64    `json:"slot" gorm:"type:bigint(20) unsigned;not null;index;comment:区块高度"`
	BlockID           *uint64   `json:"block_id" gorm:"type:bigint(20) unsigned;index;comment:关联 blocks 表的ID"`
	Blockhash         string    `json:"blockhash" gorm:"type:varchar(120);not null;default:'';comment:区块哈希"`
	RecentBlockhash   string    `json:"recent_blockhash" gorm:"type:varchar(120);not null;default:'';comment:最近区块哈希"`
	Version           string    `json:"version" gorm:"type:varchar(20);not null;default:'legacy';comment:交易版本"`
	Fee               uint64    `json:"fee" gorm:"type:bigint(20) unsigned;not null;default:0;comment:交易费用(lamports)"`
	ComputeUnits      uint64    `json:"compute_units" gorm:"type:bigint(20) unsigned;not null;default:0;comment:消耗的计算单元"`
	Status            string    `json:"status" gorm:"type:varchar(20);not null;default:'success';comment:交易状态"`
	AccountKeys       JSONText  `json:"account_keys" gorm:"type:json;comment:账户密钥数组JSON"`
	PreBalances       JSONText  `json:"pre_balances" gorm:"type:json;comment:执行前余额JSON"`
	PostBalances      JSONText  `json:"post_balances" gorm:"type:json;comment:执行后余额JSON"`
	PreTokenBalances  JSONText  `json:"pre_token_balances" gorm:"type:json;comment:执行前代币余额JSON"`
	PostTokenBalances JSONText  `json:"post_token_balances" gorm:"type:json;comment:执行后代币余额JSON"`
	Logs              JSONText  `json:"logs" gorm:"type:json;comment:日志消息JSON"`
	Instructions      JSONText  `json:"instructions" gorm:"type:json;comment:指令数组JSON"`
	InnerInstructions JSONText  `json:"inner_instructions" gorm:"type:json;comment:内部指令JSON"`
	LoadedAddresses   JSONText  `json:"loaded_addresses" gorm:"type:json;comment:加载的地址JSON"`
	Rewards           JSONText  `json:"rewards" gorm:"type:json;comment:奖励信息JSON"`
	Events            JSONText  `json:"events" gorm:"type:json;comment:解析出的事件JSON"`
	RawTransaction    JSONText  `json:"raw_transaction" gorm:"type:json;comment:原始交易数据JSON"`
	RawMeta           JSONText  `json:"raw_meta" gorm:"type:json;comment:原始元数据JSON"`
	Ctime             time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime"`
	Mtime             time.Time `json:"mtime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;column:mtime"`
}

func (SolTxDetail) TableName() string { return "sol_tx_detail" }

// SolEvent 存储 Solana 交易事件（从指令中解析出的转账等事件）
type SolEvent struct {
	ID          uint      `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID        string    `json:"tx_id" gorm:"type:varchar(120);not null;index;comment:交易签名"`
	BlockID     *uint64   `json:"block_id" gorm:"type:bigint(20) unsigned;index;comment:关联 blocks 表的ID"`
	Slot        uint64    `json:"slot" gorm:"type:bigint(20) unsigned;not null;index;comment:区块高度"`
	EventIndex  int       `json:"event_index" gorm:"type:int(11);not null;default:0;comment:事件在交易中的索引"`
	EventType   string    `json:"event_type" gorm:"type:varchar(50);not null;default:'';comment:事件类型"`
	ProgramID   string    `json:"program_id" gorm:"type:varchar(120);not null;default:'';comment:程序ID"`
	FromAddress string    `json:"from_address" gorm:"type:varchar(120);not null;default:'';comment:发送地址"`
	ToAddress   string    `json:"to_address" gorm:"type:varchar(120);not null;default:'';comment:接收地址"`
	Amount      string    `json:"amount" gorm:"type:varchar(50);not null;default:'0';comment:金额"`
	Mint        string    `json:"mint" gorm:"type:varchar(120);not null;default:'';comment:代币mint地址"`
	Decimals    int       `json:"decimals" gorm:"type:int(11);not null;default:9;comment:精度"`
	IsInner     bool      `json:"is_inner" gorm:"type:tinyint(1);not null;default:0;comment:是否为内部指令"`
	AssetType   string    `json:"asset_type" gorm:"type:varchar(20);not null;default:'NATIVE';comment:资产类型"`
	ExtraData   JSONText  `json:"extra_data" gorm:"type:json;comment:额外数据JSON"`
	Ctime       time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime"`
}

func (SolEvent) TableName() string { return "sol_event" }

// SolInstruction 存储 Solana 指令详情（简化版本）
type SolInstruction struct {
	ID               uint      `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID             string    `json:"tx_id" gorm:"type:varchar(120);not null;index;comment:交易签名"`
	BlockID          *uint64   `json:"block_id" gorm:"type:bigint(20) unsigned;index;comment:关联 blocks 表的ID"`
	Slot             uint64    `json:"slot" gorm:"type:bigint(20) unsigned;not null;index;comment:区块高度"`
	InstructionIndex int       `json:"instruction_index" gorm:"type:int(11);not null;default:0;comment:指令索引"`
	ProgramID        string    `json:"program_id" gorm:"type:varchar(120);not null;default:'';comment:程序ID"`
	Accounts         JSONText  `json:"accounts" gorm:"type:json;comment:账户数组JSON"`
	Data             string    `json:"data" gorm:"type:text;not null;comment:指令数据"`
	ParsedData       JSONText  `json:"parsed_data" gorm:"type:json;comment:解析后的数据JSON"`
	InstructionType  string    `json:"instruction_type" gorm:"type:varchar(50);not null;default:'';comment:指令类型"`
	IsInner          bool      `json:"is_inner" gorm:"type:tinyint(1);not null;default:0;comment:是否为内部指令"`
	StackHeight      int       `json:"stack_height" gorm:"type:int(11);not null;default:1;comment:堆栈高度"`
	Ctime            time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime"`
}

func (SolInstruction) TableName() string { return "sol_instruction" }

// SolProgram 维护 Solana 程序及其解析规则
type SolProgram struct {
	ID               uint      `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	ProgramID        string    `json:"program_id" gorm:"type:varchar(120);not null;uniqueIndex;comment:Sol 程序ID(地址)"`
	Name             string    `json:"name" gorm:"type:varchar(120);not null;default:'';comment:程序名称"`
	Alias            string    `json:"alias" gorm:"type:varchar(120);not null;default:'';comment:别名"`
	Category         string    `json:"category" gorm:"type:varchar(50);not null;default:'';comment:分类(系统/外部/DeFi/NFT等)"`
	Type             string    `json:"type" gorm:"type:varchar(50);not null;default:'';comment:类型(系统/Token等)"`
	IsSystem         bool      `json:"is_system" gorm:"type:tinyint(1);not null;default:0;comment:是否系统程序"`
	Version          string    `json:"version" gorm:"type:varchar(20);not null;default:'';comment:版本"`
	Status           string    `json:"status" gorm:"type:varchar(20);not null;default:'active';comment:状态(active/inactive)"`
	Description      string    `json:"description" gorm:"type:text;comment:描述"`
	InstructionRules JSONText  `json:"instruction_rules" gorm:"type:json;comment:指令解析规则JSON"`
	EventRules       JSONText  `json:"event_rules" gorm:"type:json;comment:事件解析规则JSON"`
	SampleData       JSONText  `json:"sample_data" gorm:"type:json;comment:示例数据(JSONParsed)"`
	Ctime            time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime"`
	Mtime            time.Time `json:"mtime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;column:mtime"`
}

func (SolProgram) TableName() string { return "sol_program" }

// SolParsedExtra 无法归档为事件/指令的解析额外数据
type SolParsedExtra struct {
	ID        uint      `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID      string    `json:"tx_id" gorm:"type:varchar(120);not null;index;comment:交易签名"`
	BlockID   *uint64   `json:"block_id" gorm:"type:bigint(20) unsigned;index;comment:关联 blocks 表的ID"`
	Slot      uint64    `json:"slot" gorm:"type:bigint(20) unsigned;not null;index;comment:区块高度"`
	ProgramID string    `json:"program_id" gorm:"type:varchar(120);not null;default:'';comment:程序ID"`
	IsInner   bool      `json:"is_inner" gorm:"type:tinyint(1);not null;default:0;comment:是否为内部指令"`
	Data      JSONText  `json:"data" gorm:"type:json;comment:无法标准化的数据JSON"`
	Ctime     time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime"`
}

func (SolParsedExtra) TableName() string { return "sol_parsed_extra" }
