package models

import "time"

// TransferEvent 统一的跨链转账事件表
// 适用于 SOL System/SPL 转账、EVM ERC20/原生转账、BTC 亦可映射（可选）
type TransferEvent struct {
	ID         uint   `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	Chain      string `json:"chain" gorm:"type:varchar(20);not null;index:idx_transfer_chain_height,priority:1;index:idx_transfer_chain_txid,priority:1;index:idx_transfer_chain_from,priority:1;index:idx_transfer_chain_to,priority:1;comment:链名称"`
	TxID       string `json:"tx_id" gorm:"type:varchar(120);not null;index:idx_transfer_chain_txid,priority:2;comment:交易哈希或签名"`
	Height     uint64 `json:"height" gorm:"type:bigint(20) unsigned;not null;index:idx_transfer_chain_height,priority:2;comment:块高度/slot"`
	BlockIndex uint   `json:"block_index" gorm:"type:int(11) unsigned;not null;default:0;comment:交易在区块内索引"`
	EventIndex uint   `json:"event_index" gorm:"type:int(11) unsigned;not null;default:0;comment:事件在交易内索引(含inner)"`

	ProgramID      string `json:"program_id" gorm:"type:varchar(120);not null;default:'';comment:触发该转账的程序/合约地址"`
	AssetType      string `json:"asset_type" gorm:"type:varchar(20);not null;default:'NATIVE';comment:NATIVE|SPL|ERC20|ERC721|TOKEN2022等"`
	MintOrContract string `json:"mint_or_contract" gorm:"type:varchar(120);not null;default:'';comment:资产标识(SPL mint/合约地址/或SOL)"`

	FromAddress string `json:"from_address" gorm:"type:varchar(120);not null;index:idx_transfer_chain_from,priority:2;comment:发送地址(持有人)"`
	ToAddress   string `json:"to_address" gorm:"type:varchar(120);not null;index:idx_transfer_chain_to,priority:2;comment:接收地址(持有人)"`

	AmountRaw string `json:"amount_raw" gorm:"type:decimal(65,0);not null;default:0;comment:最小单位字符串"`
	Decimals  uint8  `json:"decimals" gorm:"type:tinyint(3) unsigned;not null;default:0;comment:精度"`
	AmountUI  string `json:"amount_ui" gorm:"type:varchar(100);not null;default:'0';comment:按精度换算后的字符串"`

	IsInner bool  `json:"is_inner" gorm:"type:tinyint(1);not null;default:0;comment:是否由内联指令产生"`
	Status  uint8 `json:"status" gorm:"type:tinyint(3) unsigned;not null;default:1;comment:1成功 2失败"`

	Ctime time.Time `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime;comment:创建时间"`
	Mtime time.Time `json:"mtime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:mtime;comment:更新时间"`
}

func (TransferEvent) TableName() string { return "transfer_event" }
