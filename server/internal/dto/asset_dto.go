package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"

	"github.com/shopspring/decimal"
)

// CreateAssetRequest 创建资产请求DTO
type CreateAssetRequest struct {
	Symbol     string          `json:"symbol" validate:"required,max=10"`
	Amount     decimal.Decimal `json:"amount" validate:"gte=0"`
	Locked     decimal.Decimal `json:"locked" validate:"gte=0"`
	LastTxid   string          `json:"last_txid" validate:"required,max=1024"`
	LastHeight uint64          `json:"last_height" validate:"gte=0"`
	Type       uint16          `json:"type" validate:"required"`
	Chain      string          `json:"chain" validate:"required,max=20"`
}

// UpdateAssetRequest 更新资产请求DTO
type UpdateAssetRequest struct {
	Amount     *decimal.Decimal `json:"amount,omitempty" validate:"omitempty,gte=0"`
	Locked     *decimal.Decimal `json:"locked,omitempty" validate:"omitempty,gte=0"`
	LastTxid   *string          `json:"last_txid,omitempty" validate:"omitempty,max=1024"`
	LastHeight *uint64          `json:"last_height,omitempty" validate:"omitempty,gte=0"`
}

// AssetResponse 资产响应DTO
type AssetResponse struct {
	ID         uint            `json:"id"`
	Address    string          `json:"address"`
	Symbol     string          `json:"symbol"`
	Amount     decimal.Decimal `json:"amount"`
	Locked     decimal.Decimal `json:"locked"`
	LastTxid   string          `json:"last_txid"`
	LastHeight uint64          `json:"last_height"`
	Type       uint16          `json:"type"`
	Chain      string          `json:"chain"`
	Ctime      time.Time       `json:"ctime"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// AssetSummaryResponse 资产摘要响应DTO
type AssetSummaryResponse struct {
	Address string          `json:"address"`
	Symbol  string          `json:"symbol"`
	Amount  decimal.Decimal `json:"amount"`
	Chain   string          `json:"chain"`
}

// ToModel 将CreateAssetRequest转换为Asset模型
func (req *CreateAssetRequest) ToModel(address string) *models.Asset {
	return &models.Asset{
		Address:    address,
		Symbol:     req.Symbol,
		Amount:     req.Amount,
		Locked:     req.Locked,
		LastTxid:   req.LastTxid,
		LastHeight: req.LastHeight,
		Type:       req.Type,
		Chain:      req.Chain,
		Ctime:      time.Now(),
	}
}

// ApplyToModel 将UpdateAssetRequest应用到Asset模型
func (req *UpdateAssetRequest) ApplyToModel(asset *models.Asset) {
	if req.Amount != nil {
		asset.Amount = *req.Amount
	}
	if req.Locked != nil {
		asset.Locked = *req.Locked
	}
	if req.LastTxid != nil {
		asset.LastTxid = *req.LastTxid
	}
	if req.LastHeight != nil {
		asset.LastHeight = *req.LastHeight
	}
}

// FromModel 将Asset模型转换为AssetResponse
func (resp *AssetResponse) FromModel(asset *models.Asset) {
	resp.ID = asset.ID
	resp.Address = asset.Address
	resp.Symbol = asset.Symbol
	resp.Amount = asset.Amount
	resp.Locked = asset.Locked
	resp.LastTxid = asset.LastTxid
	resp.LastHeight = asset.LastHeight
	resp.Type = asset.Type
	resp.Chain = asset.Chain
	resp.Ctime = asset.Ctime
	resp.CreatedAt = asset.CreatedAt
	resp.UpdatedAt = asset.UpdatedAt
}

// NewAssetResponse 创建AssetResponse
func NewAssetResponse(asset *models.Asset) *AssetResponse {
	resp := &AssetResponse{}
	resp.FromModel(asset)
	return resp
}

// NewAssetSummaryResponse 创建AssetSummaryResponse
func NewAssetSummaryResponse(asset *models.Asset) *AssetSummaryResponse {
	return &AssetSummaryResponse{
		Address: asset.Address,
		Symbol:  asset.Symbol,
		Amount:  asset.Amount,
		Chain:   asset.Chain,
	}
}
