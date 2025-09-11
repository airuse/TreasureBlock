package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"
)

// CreateBlockRequest 创建区块请求DTO
type CreateBlockRequest struct {
	Hash             string    `json:"hash" validate:"required"`
	Height           uint64    `json:"height" validate:"required,gt=0"`
	PreviousHash     string    `json:"previous_hash" validate:"omitempty"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Size             uint64    `json:"size" validate:"gte=0"`
	TransactionCount int       `json:"transaction_count" validate:"gte=0"`
	TotalAmount      float64   `json:"total_amount" validate:"gte=0"`
	Fee              float64   `json:"fee" validate:"gte=0"`
	Confirmations    uint64    `json:"confirmations" validate:"gte=0"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain" validate:"required,oneof=btc eth"`
	ChainID          int       `json:"chain_id" validate:"required,gt=0"`

	// BTC特有字段
	MerkleRoot string `json:"merkle_root,omitempty" validate:"omitempty"`
	Bits       string `json:"bits,omitempty" validate:"omitempty,max=20"`
	Version    uint32 `json:"version,omitempty"`
	Weight     uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit    uint64 `json:"gas_limit,omitempty"`
	GasUsed     uint64 `json:"gas_used,omitempty"`
	Miner       string `json:"miner,omitempty" validate:"omitempty,max=120"`
	ParentHash  string `json:"parent_hash,omitempty" validate:"omitempty"`
	Nonce       string `json:"nonce,omitempty" validate:"omitempty,max=20"`
	Difficulty  string `json:"difficulty,omitempty" validate:"omitempty,max=50"`
	BaseFee     string `json:"base_fee,omitempty" validate:"omitempty,max=100"`
	BurnedEth   string `json:"burned_eth,omitempty" validate:"omitempty,max=100"`
	MinerTipEth string `json:"miner_tip_eth,omitempty" validate:"omitempty,max=100"`

	// ETH状态根字段
	StateRoot        string `json:"state_root,omitempty" validate:"omitempty"`
	TransactionsRoot string `json:"transactions_root,omitempty" validate:"omitempty"`
	ReceiptsRoot     string `json:"receipts_root,omitempty" validate:"omitempty"`
}

// UpdateBlockRequest 更新区块请求DTO
type UpdateBlockRequest struct {
	Size             *uint64  `json:"size,omitempty" validate:"omitempty,gte=0"`
	TransactionCount *int     `json:"transaction_count,omitempty" validate:"omitempty,gte=0"`
	TotalAmount      *float64 `json:"total_amount,omitempty" validate:"omitempty,gte=0"`
	Fee              *float64 `json:"fee,omitempty" validate:"omitempty,gte=0"`
	Confirmations    *uint64  `json:"confirmations,omitempty" validate:"omitempty,gte=0"`
	IsOrphan         *bool    `json:"is_orphan,omitempty"`

	// BTC特有字段
	MerkleRoot *string `json:"merkle_root,omitempty" validate:"omitempty"`
	Bits       *string `json:"bits,omitempty" validate:"omitempty,max=20"`
	Version    *uint32 `json:"version,omitempty"`
	Weight     *uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit    *uint64 `json:"gas_limit,omitempty"`
	GasUsed     *uint64 `json:"gas_used,omitempty"`
	Miner       *string `json:"miner,omitempty" validate:"omitempty,max=120"`
	ParentHash  *string `json:"parent_hash,omitempty" validate:"omitempty"`
	Nonce       *string `json:"nonce,omitempty" validate:"omitempty,max=20"`
	Difficulty  *string `json:"difficulty,omitempty" validate:"omitempty,max=50"`
	BaseFee     *string `json:"base_fee,omitempty" validate:"omitempty,max=100"`
	BurnedEth   *string `json:"burned_eth,omitempty" validate:"omitempty,max=100"`
	MinerTipEth *string `json:"miner_tip_eth,omitempty" validate:"omitempty,max=100"`

	// ETH状态根字段
	StateRoot        *string `json:"state_root,omitempty" validate:"omitempty"`
	TransactionsRoot *string `json:"transactions_root,omitempty" validate:"omitempty"`
	ReceiptsRoot     *string `json:"receipts_root,omitempty" validate:"omitempty"`
}

// BlockResponse 区块响应DTO
type BlockResponse struct {
	ID               uint      `json:"id"`
	Hash             string    `json:"hash"`
	Height           uint64    `json:"height"`
	PreviousHash     string    `json:"previous_hash"`
	Timestamp        time.Time `json:"timestamp"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain"`

	// BTC特有字段
	MerkleRoot string `json:"merkle_root,omitempty"`
	Bits       string `json:"bits,omitempty"`
	Version    uint32 `json:"version,omitempty"`
	Weight     uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit    uint64 `json:"gas_limit,omitempty"`
	GasUsed     uint64 `json:"gas_used,omitempty"`
	Miner       string `json:"miner,omitempty"`
	ParentHash  string `json:"parent_hash,omitempty"`
	Nonce       string `json:"nonce,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
	BaseFee     string `json:"base_fee,omitempty"`
	BurnedEth   string `json:"burned_eth,omitempty"`
	MinerTipEth string `json:"miner_tip_eth,omitempty"`

	// ETH状态根字段
	StateRoot        string `json:"state_root,omitempty"`
	TransactionsRoot string `json:"transactions_root,omitempty"`
	ReceiptsRoot     string `json:"receipts_root,omitempty"`

	// 时间字段
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BlockListResponse 区块列表响应DTO
type BlockListResponse struct {
	Blocks     []*BlockResponse `json:"blocks"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// ToModel 将CreateBlockRequest转换为Block模型
func (req *CreateBlockRequest) ToModel() *models.Block {
	return &models.Block{
		Hash:             req.Hash,
		Height:           req.Height,
		PreviousHash:     req.PreviousHash,
		Timestamp:        req.Timestamp,
		Size:             req.Size,
		TransactionCount: req.TransactionCount,
		TotalAmount:      req.TotalAmount,
		Fee:              req.Fee,
		Confirmations:    req.Confirmations,
		IsOrphan:         req.IsOrphan,
		Chain:            req.Chain,
		ChainID:          req.ChainID,
		// BTC特有字段
		MerkleRoot: req.MerkleRoot,
		Bits:       req.Bits,
		Version:    req.Version,
		Weight:     req.Weight,

		// ETH特有字段
		GasLimit:    req.GasLimit,
		GasUsed:     req.GasUsed,
		Miner:       req.Miner,
		ParentHash:  req.ParentHash,
		Nonce:       req.Nonce,
		Difficulty:  req.Difficulty,
		BaseFee:     req.BaseFee,
		BurnedEth:   req.BurnedEth,
		MinerTipEth: req.MinerTipEth,

		// ETH状态根字段
		StateRoot:        req.StateRoot,
		TransactionsRoot: req.TransactionsRoot,
		ReceiptsRoot:     req.ReceiptsRoot,
	}
}

// ApplyToModel 将UpdateBlockRequest应用到Block模型
func (req *UpdateBlockRequest) ApplyToModel(block *models.Block) {
	if req.Size != nil {
		block.Size = *req.Size
	}
	if req.TransactionCount != nil {
		block.TransactionCount = *req.TransactionCount
	}
	if req.TotalAmount != nil {
		block.TotalAmount = *req.TotalAmount
	}
	if req.Fee != nil {
		block.Fee = *req.Fee
	}
	if req.Confirmations != nil {
		block.Confirmations = *req.Confirmations
	}
	if req.IsOrphan != nil {
		block.IsOrphan = *req.IsOrphan
	}

	// BTC特有字段
	if req.MerkleRoot != nil {
		block.MerkleRoot = *req.MerkleRoot
	}
	if req.Bits != nil {
		block.Bits = *req.Bits
	}
	if req.Version != nil {
		block.Version = *req.Version
	}
	if req.Weight != nil {
		block.Weight = *req.Weight
	}

	// ETH特有字段
	if req.GasLimit != nil {
		block.GasLimit = *req.GasLimit
	}
	if req.GasUsed != nil {
		block.GasUsed = *req.GasUsed
	}
	if req.Miner != nil {
		block.Miner = *req.Miner
	}
	if req.ParentHash != nil {
		block.ParentHash = *req.ParentHash
	}
	if req.Nonce != nil {
		block.Nonce = *req.Nonce
	}
	if req.Difficulty != nil {
		block.Difficulty = *req.Difficulty
	}
	if req.BaseFee != nil {
		block.BaseFee = *req.BaseFee
	}
	if req.BurnedEth != nil {
		block.BurnedEth = *req.BurnedEth
	}
	if req.MinerTipEth != nil {
		block.MinerTipEth = *req.MinerTipEth
	}

	// ETH状态根字段
	if req.StateRoot != nil {
		block.StateRoot = *req.StateRoot
	}
	if req.TransactionsRoot != nil {
		block.TransactionsRoot = *req.TransactionsRoot
	}
	if req.ReceiptsRoot != nil {
		block.ReceiptsRoot = *req.ReceiptsRoot
	}
}

// NewBlockResponse 创建BlockResponse
func NewBlockResponse(block *models.Block) *BlockResponse {
	if block == nil {
		return nil
	}
	return &BlockResponse{
		ID:               block.ID,
		Hash:             block.Hash,
		Height:           block.Height,
		PreviousHash:     block.PreviousHash,
		Timestamp:        block.Timestamp,
		Size:             block.Size,
		TransactionCount: block.TransactionCount,
		TotalAmount:      block.TotalAmount,
		Fee:              block.Fee,
		Confirmations:    block.Confirmations,
		IsOrphan:         block.IsOrphan,
		Chain:            block.Chain,
		// BTC特有
		MerkleRoot: block.MerkleRoot,
		Bits:       block.Bits,
		Version:    block.Version,
		Weight:     block.Weight,
		// ETH特有
		GasLimit:    block.GasLimit,
		GasUsed:     block.GasUsed,
		Miner:       block.Miner,
		ParentHash:  block.ParentHash,
		Nonce:       block.Nonce,
		Difficulty:  block.Difficulty,
		BaseFee:     block.BaseFee,
		BurnedEth:   block.BurnedEth,
		MinerTipEth: block.MinerTipEth,

		// ETH状态根字段
		StateRoot:        block.StateRoot,
		TransactionsRoot: block.TransactionsRoot,
		ReceiptsRoot:     block.ReceiptsRoot,
		// 时间
		CreatedAt: block.CreatedAt,
		UpdatedAt: block.UpdatedAt,
	}
}

// NewBlockListResponse 创建BlockListResponse
func NewBlockListResponse(blocks []*models.Block, total int64, page, pageSize int) *BlockListResponse {
	respBlocks := make([]*BlockResponse, 0, len(blocks))
	for _, b := range blocks {
		respBlocks = append(respBlocks, NewBlockResponse(b))
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &BlockListResponse{
		Blocks:     respBlocks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
