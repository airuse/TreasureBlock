package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"
)

// CreateAddressRequest 创建地址请求DTO
type CreateAddressRequest struct {
	Chain string `json:"chain" validate:"required,oneof=btc eth bsc"`
	Type  uint16 `json:"type" validate:"gte=0"`
	Nonce string `json:"nonce" validate:"required,min=1"`
	Hash  string `json:"hash" validate:"required,min=1,max=1024"`

	// BTC特有字段
	UTXOCount uint64 `json:"utxo_count,omitempty"`
	Balance   string `json:"balance,omitempty"`

	// ETH特有字段
	TransactionCount uint64 `json:"transaction_count,omitempty"`
	ContractCount    uint64 `json:"contract_count,omitempty"`
}

// UpdateAddressRequest 更新地址请求DTO
type UpdateAddressRequest struct {
	Type  *uint16 `json:"type,omitempty" validate:"omitempty,gte=0"`
	Nonce *string `json:"nonce,omitempty" validate:"omitempty,min=1"`
	Hash  *string `json:"hash,omitempty" validate:"omitempty,min=1,max=1024"`

	// BTC特有字段
	UTXOCount *uint64 `json:"utxo_count,omitempty"`
	Balance   *string `json:"balance,omitempty"`

	// ETH特有字段
	TransactionCount *uint64 `json:"transaction_count,omitempty"`
	ContractCount    *uint64 `json:"contract_count,omitempty"`
}

// AddressResponse 地址响应DTO
type AddressResponse struct {
	ID      uint   `json:"id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
	Type    uint16 `json:"type"`
	Nonce   string `json:"nonce"`
	Hash    string `json:"hash"`

	// BTC特有字段
	UTXOCount uint64 `json:"utxo_count,omitempty"`
	Balance   string `json:"balance,omitempty"`

	// ETH特有字段
	TransactionCount uint64 `json:"transaction_count,omitempty"`
	ContractCount    uint64 `json:"contract_count,omitempty"`

	// 时间字段
	Mtime     time.Time `json:"mtime"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToModel 将CreateAddressRequest转换为Address模型
func (req *CreateAddressRequest) ToModel(address string) *models.Address {
	return &models.Address{
		Address: address,
		Chain:   req.Chain,
		Type:    req.Type,
		Nonce:   req.Nonce,
		Hash:    req.Hash,

		// BTC特有字段
		UTXOCount: req.UTXOCount,
		Balance:   req.Balance,

		// ETH特有字段
		TransactionCount: req.TransactionCount,
		ContractCount:    req.ContractCount,

		Mtime: time.Now(),
	}
}

// ApplyToModel 将UpdateAddressRequest应用到Address模型
func (req *UpdateAddressRequest) ApplyToModel(addr *models.Address) {
	if req.Type != nil {
		addr.Type = *req.Type
	}
	if req.Nonce != nil {
		addr.Nonce = *req.Nonce
	}
	if req.Hash != nil {
		addr.Hash = *req.Hash
	}

	// BTC特有字段
	if req.UTXOCount != nil {
		addr.UTXOCount = *req.UTXOCount
	}
	if req.Balance != nil {
		addr.Balance = *req.Balance
	}

	// ETH特有字段
	if req.TransactionCount != nil {
		addr.TransactionCount = *req.TransactionCount
	}
	if req.ContractCount != nil {
		addr.ContractCount = *req.ContractCount
	}

	addr.Mtime = time.Now()
}

// FromModel 将Address模型转换为AddressResponse
func (resp *AddressResponse) FromModel(addr *models.Address) {
	resp.ID = addr.ID
	resp.Address = addr.Address
	resp.Type = addr.Type
	resp.Nonce = addr.Nonce
	resp.Hash = addr.Hash
	resp.Mtime = addr.Mtime
	resp.CreatedAt = addr.CreatedAt
	resp.UpdatedAt = addr.UpdatedAt
}

// NewAddressResponse 创建AddressResponse
func NewAddressResponse(addr *models.Address) *AddressResponse {
	resp := &AddressResponse{}
	resp.FromModel(addr)
	return resp
}
