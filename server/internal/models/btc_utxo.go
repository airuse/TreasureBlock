package models

import (
	"time"

	"gorm.io/gorm"
)

// BTCUTXO 存储BTC的UTXO记录（包含来源与花费状态）
type BTCUTXO struct {
	ID           uint64  `json:"id" gorm:"primaryKey;column:id;autoIncrement;comment:主键ID"`
	Chain        string  `json:"chain" gorm:"type:varchar(20);not null;uniqueIndex:ux_chain_txid_vout,priority:1;index;comment:链类型(btc)"`
	TxID         string  `json:"tx_id" gorm:"type:varchar(120);not null;uniqueIndex:ux_chain_txid_vout,priority:2;index;comment:创建此输出的txid"`
	VoutIndex    uint32  `json:"vout_index" gorm:"type:int(11) unsigned;not null;uniqueIndex:ux_chain_txid_vout,priority:3;comment:输出索引n"`
	BlockHeight  uint64  `json:"block_height" gorm:"type:bigint(20) unsigned;not null;index;comment:创建所在块高度"`
	BlockID      *uint64 `json:"block_id" gorm:"type:bigint(20) unsigned;index;comment:关联区块ID"`
	Address      string  `json:"address" gorm:"type:varchar(120);not null;index;comment:锁定地址(如有)"`
	ScriptPubKey string  `json:"script_pub_key" gorm:"type:text;comment:脚本hex"`
	ScriptType   string  `json:"script_type" gorm:"type:varchar(32);index;comment:脚本类型p2pkh/p2sh/p2wpkh/p2wsh/p2tr/nulldata/unknown"`
	IsCoinbase   bool    `json:"is_coinbase" gorm:"type:tinyint(1);not null;default:0;index;comment:是否coinbase产生"`
	ValueSatoshi int64   `json:"value_satoshi" gorm:"type:bigint(20);not null;comment:金额(聪)"`

	// 花费信息
	SpentTxID     string     `json:"spent_tx_id" gorm:"type:varchar(120);index;comment:花费该UTXO的txid"`
	SpentVinIndex *uint32    `json:"spent_vin_index" gorm:"type:int(11) unsigned;comment:在花费交易中的vin索引"`
	SpentHeight   *uint64    `json:"spent_height" gorm:"type:bigint(20) unsigned;comment:花费发生的高度"`
	SpentAt       *time.Time `json:"spent_at" gorm:"type:timestamp;comment:花费时间"`

	// 时间与软删除
	Ctime     time.Time      `json:"ctime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:ctime;index;comment:入库时间"`
	Mtime     time.Time      `json:"mtime" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:mtime;comment:更改时间"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (BTCUTXO) TableName() string {
	return "btc_utxo"
}
