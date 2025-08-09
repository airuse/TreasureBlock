package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"
)

// CreateCoinConfigRequest 创建币种配置请求DTO
type CreateCoinConfigRequest struct {
	ChainName        string  `json:"chain_name" validate:"required,max=50"`
	CoinType         uint8   `json:"coin_type" validate:"required,lte=5"` // 0：eth,1：ERC20，2：ERC223，3：ERC773,4:TRC10, 5:TRC20
	ContractAddr     string  `json:"contract_addr" validate:"required,max=120"`
	Precision        uint    `json:"precision" validate:"required,lte=18"`
	ColdAddress      string  `json:"cold_address" validate:"required,max=200"`
	ColdAddressHash  string  `json:"cold_address_hash" validate:"required,max=120"`
	MaxStock         float64 `json:"max_stock" validate:"gte=0"`
	MaxBalance       float64 `json:"max_balance" validate:"gte=0"`
	MinBalance       float64 `json:"min_balance" validate:"gte=0"`
	CollectLimit     float64 `json:"collect_limit" validate:"gte=0"`
	CollectLeft      float64 `json:"collect_left" validate:"gte=0"`
	InternalGasLimit uint    `json:"internal_gas_limit" validate:"required"`
	OnceMinFee       float64 `json:"once_min_fee" validate:"gte=0"`
	SymbolID         string  `json:"symbol_id" validate:"max=100"`
	Status           int8    `json:"status" validate:"oneof=0 1"`
}

// UpdateCoinConfigRequest 更新币种配置请求DTO
type UpdateCoinConfigRequest struct {
	ChainName        *string  `json:"chain_name,omitempty" validate:"omitempty,max=50"`
	CoinType         *uint8   `json:"coin_type,omitempty" validate:"omitempty,lte=5"`
	ContractAddr     *string  `json:"contract_addr,omitempty" validate:"omitempty,max=120"`
	Precision        *uint    `json:"precision,omitempty" validate:"omitempty,lte=18"`
	ColdAddress      *string  `json:"cold_address,omitempty" validate:"omitempty,max=200"`
	ColdAddressHash  *string  `json:"cold_address_hash,omitempty" validate:"omitempty,max=120"`
	MaxStock         *float64 `json:"max_stock,omitempty" validate:"omitempty,gte=0"`
	MaxBalance       *float64 `json:"max_balance,omitempty" validate:"omitempty,gte=0"`
	MinBalance       *float64 `json:"min_balance,omitempty" validate:"omitempty,gte=0"`
	CollectLimit     *float64 `json:"collect_limit,omitempty" validate:"omitempty,gte=0"`
	CollectLeft      *float64 `json:"collect_left,omitempty" validate:"omitempty,gte=0"`
	InternalGasLimit *uint    `json:"internal_gas_limit,omitempty"`
	OnceMinFee       *float64 `json:"once_min_fee,omitempty" validate:"omitempty,gte=0"`
	SymbolID         *string  `json:"symbol_id,omitempty" validate:"omitempty,max=100"`
	Status           *int8    `json:"status,omitempty" validate:"omitempty,oneof=0 1"`
}

// CoinConfigResponse 币种配置响应DTO
type CoinConfigResponse struct {
	ID               uint      `json:"id"`
	ChainName        string    `json:"chain_name"`
	Symbol           string    `json:"symbol"`
	CoinType         uint8     `json:"coin_type"`
	ContractAddr     string    `json:"contract_addr"`
	Precision        uint      `json:"precision"`
	ColdAddress      string    `json:"cold_address"`
	ColdAddressHash  string    `json:"cold_address_hash"`
	MaxStock         float64   `json:"max_stock"`
	MaxBalance       float64   `json:"max_balance"`
	MinBalance       float64   `json:"min_balance"`
	CollectLimit     float64   `json:"collect_limit"`
	CollectLeft      float64   `json:"collect_left"`
	InternalGasLimit uint      `json:"internal_gas_limit"`
	OnceMinFee       float64   `json:"once_min_fee"`
	CTime            time.Time `json:"ctime"`
	MTime            time.Time `json:"mtime"`
	SymbolID         string    `json:"symbol_id"`
	Status           int8      `json:"status"`
}

// CoinConfigSummaryResponse 币种配置摘要响应DTO
type CoinConfigSummaryResponse struct {
	Symbol   string `json:"symbol"`
	CoinType uint8  `json:"coin_type"`
	Chain    string `json:"chain_name"`
	Status   int8   `json:"status"`
}

// ToModel 将CreateCoinConfigRequest转换为CoinConfig模型
func (req *CreateCoinConfigRequest) ToModel(symbol string) *models.CoinConfig {
	return &models.CoinConfig{
		ChainName:        req.ChainName,
		Symbol:           symbol,
		CoinType:         req.CoinType,
		ContractAddr:     req.ContractAddr,
		Precision:        req.Precision,
		ColdAddress:      req.ColdAddress,
		ColdAddressHash:  req.ColdAddressHash,
		MaxStock:         req.MaxStock,
		MaxBalance:       req.MaxBalance,
		MinBalance:       req.MinBalance,
		CollectLimit:     req.CollectLimit,
		CollectLeft:      req.CollectLeft,
		InternalGasLimit: req.InternalGasLimit,
		OnceMinFee:       req.OnceMinFee,
		SymbolID:         req.SymbolID,
		Status:           req.Status,
		CTime:            time.Now(),
		MTime:            time.Now(),
	}
}

// ApplyToModel 将UpdateCoinConfigRequest应用到CoinConfig模型
func (req *UpdateCoinConfigRequest) ApplyToModel(config *models.CoinConfig) {
	if req.ChainName != nil {
		config.ChainName = *req.ChainName
	}
	if req.CoinType != nil {
		config.CoinType = *req.CoinType
	}
	if req.ContractAddr != nil {
		config.ContractAddr = *req.ContractAddr
	}
	if req.Precision != nil {
		config.Precision = *req.Precision
	}
	if req.ColdAddress != nil {
		config.ColdAddress = *req.ColdAddress
	}
	if req.ColdAddressHash != nil {
		config.ColdAddressHash = *req.ColdAddressHash
	}
	if req.MaxStock != nil {
		config.MaxStock = *req.MaxStock
	}
	if req.MaxBalance != nil {
		config.MaxBalance = *req.MaxBalance
	}
	if req.MinBalance != nil {
		config.MinBalance = *req.MinBalance
	}
	if req.CollectLimit != nil {
		config.CollectLimit = *req.CollectLimit
	}
	if req.CollectLeft != nil {
		config.CollectLeft = *req.CollectLeft
	}
	if req.InternalGasLimit != nil {
		config.InternalGasLimit = *req.InternalGasLimit
	}
	if req.OnceMinFee != nil {
		config.OnceMinFee = *req.OnceMinFee
	}
	if req.SymbolID != nil {
		config.SymbolID = *req.SymbolID
	}
	if req.Status != nil {
		config.Status = *req.Status
	}
	config.MTime = time.Now()
}

// FromModel 将CoinConfig模型转换为CoinConfigResponse
func (resp *CoinConfigResponse) FromModel(config *models.CoinConfig) {
	resp.ID = config.ID
	resp.ChainName = config.ChainName
	resp.Symbol = config.Symbol
	resp.CoinType = config.CoinType
	resp.ContractAddr = config.ContractAddr
	resp.Precision = config.Precision
	resp.ColdAddress = config.ColdAddress
	resp.ColdAddressHash = config.ColdAddressHash
	resp.MaxStock = config.MaxStock
	resp.MaxBalance = config.MaxBalance
	resp.MinBalance = config.MinBalance
	resp.CollectLimit = config.CollectLimit
	resp.CollectLeft = config.CollectLeft
	resp.InternalGasLimit = config.InternalGasLimit
	resp.OnceMinFee = config.OnceMinFee
	resp.CTime = config.CTime
	resp.MTime = config.MTime
	resp.SymbolID = config.SymbolID
	resp.Status = config.Status
}

// NewCoinConfigResponse 创建CoinConfigResponse
func NewCoinConfigResponse(config *models.CoinConfig) *CoinConfigResponse {
	resp := &CoinConfigResponse{}
	resp.FromModel(config)
	return resp
}

// NewCoinConfigSummaryResponse 创建CoinConfigSummaryResponse
func NewCoinConfigSummaryResponse(config *models.CoinConfig) *CoinConfigSummaryResponse {
	return &CoinConfigSummaryResponse{
		Symbol:   config.Symbol,
		CoinType: config.CoinType,
		Chain:    config.ChainName,
		Status:   config.Status,
	}
}
