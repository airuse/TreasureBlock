package dto

import (
	"time"

	"blockChainBrowser/server/internal/models"
)

// CreateCoinConfigRequest 创建币种配置请求DTO
type CreateCoinConfigRequest struct {
	ChainName     string `json:"chain_name" validate:"required,max=20"`
	Symbol        string `json:"symbol" validate:"required,max=20"`
	CoinType      uint8  `json:"coin_type" validate:"gte=0,lte=4"`
	ContractAddr  string `json:"contract_addr" validate:"max=120"`
	Precision     uint   `json:"precision" validate:"gte=0,lte=18"`
	Decimals      uint   `json:"decimals" validate:"gte=0,lte=18"`
	Name          string `json:"name" validate:"required,max=100"`
	LogoURL       string `json:"logo_url" validate:"max=255"`
	WebsiteURL    string `json:"website_url" validate:"max=255"`
	ExplorerURL   string `json:"explorer_url" validate:"max=255"`
	Description   string `json:"description"`
	MarketCapRank uint   `json:"market_cap_rank"`
	IsStablecoin  bool   `json:"is_stablecoin"`
	IsVerified    bool   `json:"is_verified"`
	Status        int8   `json:"status" validate:"oneof=0 1"`
}

// UpdateCoinConfigRequest 更新币种配置请求DTO
type UpdateCoinConfigRequest struct {
	ChainName     *string `json:"chain_name,omitempty" validate:"omitempty,max=20"`
	Symbol        *string `json:"symbol,omitempty" validate:"omitempty,max=20"`
	CoinType      *uint8  `json:"coin_type,omitempty" validate:"omitempty,gte=0,lte=4"`
	ContractAddr  *string `json:"contract_addr,omitempty" validate:"omitempty,max=120"`
	Precision     *uint   `json:"precision,omitempty" validate:"omitempty,gte=0,lte=18"`
	Decimals      *uint   `json:"decimals,omitempty" validate:"omitempty,gte=0,lte=18"`
	Name          *string `json:"name,omitempty" validate:"omitempty,max=100"`
	LogoURL       *string `json:"logo_url,omitempty" validate:"omitempty,max=255"`
	WebsiteURL    *string `json:"website_url,omitempty" validate:"omitempty,max=255"`
	ExplorerURL   *string `json:"explorer_url,omitempty" validate:"omitempty,max=255"`
	Description   *string `json:"description,omitempty"`
	MarketCapRank *uint   `json:"market_cap_rank,omitempty"`
	IsStablecoin  *bool   `json:"is_stablecoin,omitempty"`
	IsVerified    *bool   `json:"is_verified,omitempty"`
	Status        *int8   `json:"status,omitempty" validate:"omitempty,oneof=0 1"`
}

// CoinConfigResponse 币种配置响应DTO
type CoinConfigResponse struct {
	ID            uint      `json:"id"`
	ChainName     string    `json:"chain_name"`
	Symbol        string    `json:"symbol"`
	CoinType      uint8     `json:"coin_type"`
	ContractAddr  string    `json:"contract_addr"`
	Precision     uint      `json:"precision"`
	Decimals      uint      `json:"decimals"`
	Name          string    `json:"name"`
	LogoURL       string    `json:"logo_url"`
	WebsiteURL    string    `json:"website_url"`
	ExplorerURL   string    `json:"explorer_url"`
	Description   string    `json:"description"`
	MarketCapRank uint      `json:"market_cap_rank"`
	IsStablecoin  bool      `json:"is_stablecoin"`
	IsVerified    bool      `json:"is_verified"`
	Status        int8      `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CoinConfigListResponse 币种配置列表响应DTO
type CoinConfigListResponse struct {
	CoinConfigs []*CoinConfigResponse `json:"coin_configs"`
	Total       int64                 `json:"total"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
	TotalPages  int                   `json:"total_pages"`
}

// NewCoinConfigResponse 创建币种配置响应
func NewCoinConfigResponse(coinConfig *models.CoinConfig) *CoinConfigResponse {
	return &CoinConfigResponse{
		ID:            coinConfig.ID,
		ChainName:     coinConfig.ChainName,
		Symbol:        coinConfig.Symbol,
		CoinType:      coinConfig.CoinType,
		ContractAddr:  coinConfig.ContractAddr,
		Precision:     coinConfig.Precision,
		Decimals:      coinConfig.Decimals,
		Name:          coinConfig.Name,
		LogoURL:       coinConfig.LogoURL,
		WebsiteURL:    coinConfig.WebsiteURL,
		ExplorerURL:   coinConfig.ExplorerURL,
		Description:   coinConfig.Description,
		MarketCapRank: coinConfig.MarketCapRank,
		IsStablecoin:  coinConfig.IsStablecoin,
		IsVerified:    coinConfig.IsVerified,
		Status:        coinConfig.Status,
		CreatedAt:     coinConfig.CreatedAt,
		UpdatedAt:     coinConfig.UpdatedAt,
	}
}

// NewCoinConfigListResponse 创建币种配置列表响应
func NewCoinConfigListResponse(coinConfigs []*models.CoinConfig, total int64, page, pageSize int) *CoinConfigListResponse {
	coinConfigResponses := make([]*CoinConfigResponse, len(coinConfigs))
	for i, coinConfig := range coinConfigs {
		coinConfigResponses[i] = NewCoinConfigResponse(coinConfig)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &CoinConfigListResponse{
		CoinConfigs: coinConfigResponses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}
}

// ToModel 将CreateCoinConfigRequest转换为CoinConfig模型
func (req *CreateCoinConfigRequest) ToModel() *models.CoinConfig {
	return &models.CoinConfig{
		ChainName:     req.ChainName,
		Symbol:        req.Symbol,
		CoinType:      req.CoinType,
		ContractAddr:  req.ContractAddr,
		Precision:     req.Precision,
		Decimals:      req.Decimals,
		Name:          req.Name,
		LogoURL:       req.LogoURL,
		WebsiteURL:    req.WebsiteURL,
		ExplorerURL:   req.ExplorerURL,
		Description:   req.Description,
		MarketCapRank: req.MarketCapRank,
		IsStablecoin:  req.IsStablecoin,
		IsVerified:    req.IsVerified,
		Status:        req.Status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ApplyToModel 将UpdateCoinConfigRequest应用到CoinConfig模型
func (req *UpdateCoinConfigRequest) ApplyToModel(coinConfig *models.CoinConfig) {
	if req.ChainName != nil {
		coinConfig.ChainName = *req.ChainName
	}
	if req.Symbol != nil {
		coinConfig.Symbol = *req.Symbol
	}
	if req.CoinType != nil {
		coinConfig.CoinType = *req.CoinType
	}
	if req.ContractAddr != nil {
		coinConfig.ContractAddr = *req.ContractAddr
	}
	if req.Precision != nil {
		coinConfig.Precision = *req.Precision
	}
	if req.Decimals != nil {
		coinConfig.Decimals = *req.Decimals
	}
	if req.Name != nil {
		coinConfig.Name = *req.Name
	}
	if req.LogoURL != nil {
		coinConfig.LogoURL = *req.LogoURL
	}
	if req.WebsiteURL != nil {
		coinConfig.WebsiteURL = *req.WebsiteURL
	}
	if req.ExplorerURL != nil {
		coinConfig.ExplorerURL = *req.ExplorerURL
	}
	if req.Description != nil {
		coinConfig.Description = *req.Description
	}
	if req.MarketCapRank != nil {
		coinConfig.MarketCapRank = *req.MarketCapRank
	}
	if req.IsStablecoin != nil {
		coinConfig.IsStablecoin = *req.IsStablecoin
	}
	if req.IsVerified != nil {
		coinConfig.IsVerified = *req.IsVerified
	}
	if req.Status != nil {
		coinConfig.Status = *req.Status
	}
	coinConfig.UpdatedAt = time.Now()
}
