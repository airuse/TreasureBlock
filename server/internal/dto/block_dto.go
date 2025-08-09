package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"
)

// CreateBlockRequest 创建区块请求DTO
type CreateBlockRequest struct {
	Hash             string    `json:"hash" validate:"required,len=64"`
	Height           uint64    `json:"height" validate:"required,gt=0"`
	PreviousHash     string    `json:"previous_hash" validate:"omitempty,len=64"`
	MerkleRoot       string    `json:"merkle_root" validate:"omitempty,len=64"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Difficulty       float64   `json:"difficulty" validate:"gte=0"`
	Nonce            uint64    `json:"nonce" validate:"gte=0"`
	Size             uint64    `json:"size" validate:"gte=0"`
	TransactionCount int       `json:"transaction_count" validate:"gte=0"`
	TotalAmount      float64   `json:"total_amount" validate:"gte=0"`
	Fee              float64   `json:"fee" validate:"gte=0"`
	Confirmations    uint64    `json:"confirmations" validate:"gte=0"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain" validate:"required,oneof=btc eth"`
}

// UpdateBlockRequest 更新区块请求DTO
type UpdateBlockRequest struct {
	Difficulty       *float64 `json:"difficulty,omitempty" validate:"omitempty,gte=0"`
	Nonce            *uint64  `json:"nonce,omitempty" validate:"omitempty,gte=0"`
	Size             *uint64  `json:"size,omitempty" validate:"omitempty,gte=0"`
	TransactionCount *int     `json:"transaction_count,omitempty" validate:"omitempty,gte=0"`
	TotalAmount      *float64 `json:"total_amount,omitempty" validate:"omitempty,gte=0"`
	Fee              *float64 `json:"fee,omitempty" validate:"omitempty,gte=0"`
	Confirmations    *uint64  `json:"confirmations,omitempty" validate:"omitempty,gte=0"`
	IsOrphan         *bool    `json:"is_orphan,omitempty"`
}

// BlockResponse 区块响应DTO
type BlockResponse struct {
	ID               uint      `json:"id"`
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
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
		MerkleRoot:       req.MerkleRoot,
		Timestamp:        req.Timestamp,
		Difficulty:       req.Difficulty,
		Nonce:            req.Nonce,
		Size:             req.Size,
		TransactionCount: req.TransactionCount,
		TotalAmount:      req.TotalAmount,
		Fee:              req.Fee,
		Confirmations:    req.Confirmations,
		IsOrphan:         req.IsOrphan,
		Chain:            req.Chain,
	}
}

// ApplyToModel 将UpdateBlockRequest应用到Block模型
func (req *UpdateBlockRequest) ApplyToModel(block *models.Block) {
	if req.Difficulty != nil {
		block.Difficulty = *req.Difficulty
	}
	if req.Nonce != nil {
		block.Nonce = *req.Nonce
	}
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
}

// FromModel 将Block模型转换为BlockResponse
func (resp *BlockResponse) FromModel(block *models.Block) {
	resp.ID = block.ID
	resp.Hash = block.Hash
	resp.Height = block.Height
	resp.PreviousHash = block.PreviousHash
	resp.MerkleRoot = block.MerkleRoot
	resp.Timestamp = block.Timestamp
	resp.Difficulty = block.Difficulty
	resp.Nonce = block.Nonce
	resp.Size = block.Size
	resp.TransactionCount = block.TransactionCount
	resp.TotalAmount = block.TotalAmount
	resp.Fee = block.Fee
	resp.Confirmations = block.Confirmations
	resp.IsOrphan = block.IsOrphan
	resp.Chain = block.Chain
	resp.CreatedAt = block.CreatedAt
	resp.UpdatedAt = block.UpdatedAt
}

// NewBlockResponse 创建BlockResponse
func NewBlockResponse(block *models.Block) *BlockResponse {
	resp := &BlockResponse{}
	resp.FromModel(block)
	return resp
}

// NewBlockListResponse 创建BlockListResponse
func NewBlockListResponse(blocks []*models.Block, total int64, page, pageSize int) *BlockListResponse {
	blockResponses := make([]*BlockResponse, len(blocks))
	for i, block := range blocks {
		blockResponses[i] = NewBlockResponse(block)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &BlockListResponse{
		Blocks:     blockResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
