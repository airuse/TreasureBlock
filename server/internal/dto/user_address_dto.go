package dto

// CreateUserAddressRequest 创建用户地址请求
type CreateUserAddressRequest struct {
	Address             string   `json:"address" binding:"required"`
	Chain               string   `json:"chain" binding:"required,oneof=eth btc bsc sol other"`
	Label               string   `json:"label"`
	Type                string   `json:"type" binding:"required,oneof=wallet contract authorized_contract exchange other"`
	ContractID          *uint    `json:"contract_id"`          // 关联的合约ID，仅当type为contract时有效
	AuthorizedAddresses []string `json:"authorized_addresses"` // 授权地址列表，仅当type为contract时有效
	Notes               string   `json:"notes"`                // 备注信息
}

// UpdateUserAddressRequest 更新用户地址请求
type UpdateUserAddressRequest struct {
	Label                 *string   `json:"label"`
	Chain                 *string   `json:"chain" binding:"omitempty,oneof=eth btc bsc sol other"`
	Type                  *string   `json:"type" binding:"omitempty,oneof=wallet contract authorized_contract exchange other"`
	ContractID            *uint     `json:"contract_id"`
	AuthorizedAddresses   *[]string `json:"authorized_addresses"`    // 授权地址列表
	ContractBalance       *string   `json:"contract_balance"`        // 合约余额
	ContractBalanceHeight *uint64   `json:"contract_balance_height"` // 合约余额更新时的区块高度（兼容旧入参，内部忽略）
	Notes                 *string   `json:"notes"`                   // 备注信息
	IsActive              *bool     `json:"is_active"`
}

// UserAddressResponse 用户地址响应
type UserAddressResponse struct {
	ID                  uint                         `json:"id"`
	Address             string                       `json:"address"`
	Chain               string                       `json:"chain"`
	Label               string                       `json:"label"`
	Type                string                       `json:"type"`
	ContractID          *uint                        `json:"contract_id"`          // 关联的合约ID
	AuthorizedAddresses map[string]map[string]string `json:"authorized_addresses"` // {address:{allowance:"..."}}
	Notes               string                       `json:"notes"`                // 备注信息
	Balance             *string                      `json:"balance"`              // 地址余额
	ContractBalance     *string                      `json:"contract_balance"`     // 合约余额
	TransactionCount    int64                        `json:"transaction_count"`
	UTXOCount           int64                        `json:"utxo_count"` // UTXO数量（仅BTC使用）
	IsActive            bool                         `json:"is_active"`
	BalanceHeight       uint64                       `json:"balance_height"`
	CreatedAt           string                       `json:"created_at"`
	UpdatedAt           string                       `json:"updated_at"`
}

type UserAddressPendingResponse struct {
	ID        uint   `json:"id"`
	Address   string `json:"address"`
	Amount    string `json:"amount"`
	Status    string `json:"status"`
	Fee       string `json:"fee"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
