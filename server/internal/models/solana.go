package models

import "time"

// SolTxDetail 存储 Solana 交易的明细（头部与meta快照）
type SolTxDetail struct {
	ID                uint      `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID              string    `json:"tx_id" gorm:"type:varchar(120);not null;uniqueIndex;comment:签名"`
	Slot              uint64    `json:"slot" gorm:"type:bigint(20) unsigned;not null;index;comment:slot高度"`
	Blockhash         string    `json:"blockhash" gorm:"type:varchar(120);not null;default:''"`
	RecentBlockhash   string    `json:"recent_blockhash" gorm:"type:varchar(120);not null;default:''"`
	Version           string    `json:"version" gorm:"type:varchar(20);not null;default:'legacy'"`
	Fee               uint64    `json:"fee" gorm:"type:bigint(20) unsigned;not null;default:0"`
	ComputeUnits      uint64    `json:"compute_units" gorm:"type:bigint(20) unsigned;not null;default:0"`
	CostUnits         uint64    `json:"cost_units" gorm:"type:bigint(20) unsigned;not null;default:0"`
	StatusJSON        string    `json:"status_json" gorm:"type:longtext;not null;comment:status结构JSON"`
	AccountKeys       string    `json:"account_keys" gorm:"type:longtext;not null;comment:accountKeys+loaded地址集合JSON"`
	LoadedWritable    string    `json:"loaded_writable" gorm:"type:longtext;not null;comment:loaded writable地址JSON"`
	LoadedReadonly    string    `json:"loaded_readonly" gorm:"type:longtext;not null;comment:loaded readonly地址JSON"`
	PreBalances       string    `json:"pre_balances" gorm:"type:longtext;not null;comment:preBalances JSON"`
	PostBalances      string    `json:"post_balances" gorm:"type:longtext;not null;comment:postBalances JSON"`
	PreTokenBalances  string    `json:"pre_token_balances" gorm:"type:longtext;not null;comment:preTokenBalances JSON"`
	PostTokenBalances string    `json:"post_token_balances" gorm:"type:longtext;not null;comment:postTokenBalances JSON"`
	RewardsJSON       string    `json:"rewards_json" gorm:"type:longtext;not null"`
	LogsJSON          string    `json:"logs_json" gorm:"type:longtext;not null"`
	RawJSON           string    `json:"raw_json" gorm:"type:longtext;not null;comment:完整RPC返回或交易序列化JSON"`
	Ctime             time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime"`
	Mtime             time.Time `json:"mtime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:mtime"`
}

func (SolTxDetail) TableName() string { return "sol_tx_detail" }

// SolInstruction 保存外层/内层指令（统一一张表，outer_index=-1 表示外层索引未知）
type SolInstruction struct {
	ID           uint   `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	TxID         string `json:"tx_id" gorm:"type:varchar(120);not null;index:idx_solins_txid,priority:1"`
	OuterIndex   int    `json:"outer_index" gorm:"type:int(11);not null;default:-1;index:idx_solins_txid,priority:2;comment:外层指令索引;内层同外层索引"`
	InnerIndex   int    `json:"inner_index" gorm:"type:int(11);not null;default:-1;index:idx_solins_txid,priority:3;comment:内层指令在该外层内的索引;外层为-1"`
	ProgramID    string `json:"program_id" gorm:"type:varchar(120);not null;default:''"`
	AccountsJSON string `json:"accounts_json" gorm:"type:longtext;not null;comment:索引或展开后的账户列表JSON"`
	DataBase58   string `json:"data_b58" gorm:"type:longtext;not null"`
	IsInner      bool   `json:"is_inner" gorm:"type:tinyint(1);not null;default:0"`
}

func (SolInstruction) TableName() string { return "sol_instruction" }
