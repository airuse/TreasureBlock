package models

import (
	"time"
)

// Block 区块模型
type Block struct {
	Hash             string    `json:"hash"`
	Height           uint64    `json:"height"`
	PreviousHash     string    `json:"previous_hash"`
	MerkleRoot       string    `json:"merkle_root"`
	Timestamp        time.Time `json:"timestamp"`
	Difficulty       float64   `json:"difficulty"`
	Nonce            uint64    `json:"nonce"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain"`
	Version          int       `json:"version"`
	Bits             string    `json:"bits"`
	Weight           uint64    `json:"weight"`
	StrippedSize     uint64    `json:"stripped_size"`
	Miner            string    `json:"miner,omitempty"` // 矿工地址
}

// BlockResponse 区块响应
type BlockResponse struct {
	Success bool   `json:"success"`
	Data    Block  `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BlockListResponse 区块列表响应
type BlockListResponse struct {
	Success bool    `json:"success"`
	Data    []Block `json:"data,omitempty"`
	Error   string  `json:"error,omitempty"`
	Total   int64   `json:"total,omitempty"`
	Page    int     `json:"page,omitempty"`
	Limit   int     `json:"limit,omitempty"`
}

// ScanProgress 扫描进度
type ScanProgress struct {
	Chain           string    `json:"chain"`
	CurrentHeight   uint64    `json:"current_height"`
	LatestHeight    uint64    `json:"latest_height"`
	ProcessedBlocks int64     `json:"processed_blocks"`
	FailedBlocks    int64     `json:"failed_blocks"`
	StartTime       time.Time `json:"start_time"`
	LastUpdateTime  time.Time `json:"last_update_time"`
	Status          string    `json:"status"` // running, paused, stopped, completed
}
