package dto

import "time"

// CreateUserTransactionRequest 创建用户交易请求
type CreateUserTransactionRequest struct {
	Chain       string  `json:"chain" binding:"required,oneof=btc eth" validate:"required,oneof=btc eth"`
	Symbol      string  `json:"symbol" binding:"required" validate:"required"`
	FromAddress string  `json:"from_address" binding:"omitempty" validate:"omitempty"`
	ToAddress   string  `json:"to_address" binding:"omitempty" validate:"omitempty"`
	Amount      string  `json:"amount" binding:"omitempty" validate:"omitempty"`
	Fee         string  `json:"fee" binding:"omitempty" validate:"omitempty"`
	GasLimit    *uint   `json:"gas_limit" validate:"omitempty,min=1"`
	GasPrice    *string `json:"gas_price" validate:"omitempty"`
	Nonce       *uint64 `json:"nonce" validate:"omitempty,min=0"`
	Remark      string  `json:"remark" validate:"omitempty,max=500"`

	// 代币交易相关字段
	TransactionType       string `json:"transaction_type" binding:"omitempty" validate:"omitempty,oneof=coin token"`
	ContractOperationType string `json:"contract_operation_type" binding:"omitempty" validate:"omitempty,oneof=transfer approve transferFrom balanceOf"`
	TokenContractAddress  string `json:"token_contract_address" binding:"omitempty" validate:"omitempty"`
}

// UpdateUserTransactionRequest 更新用户交易请求
type UpdateUserTransactionRequest struct {
	Status        *string `json:"status" validate:"omitempty,oneof=draft unsigned unsent in_progress packed confirmed failed"`
	TxHash        *string `json:"tx_hash" validate:"omitempty"`
	UnsignedTx    *string `json:"unsigned_tx" validate:"omitempty"`
	SignedTx      *string `json:"signed_tx" validate:"omitempty"`
	BlockHeight   *uint64 `json:"block_height" validate:"omitempty"`
	Confirmations *uint   `json:"confirmations" validate:"omitempty"`
	ErrorMsg      *string `json:"error_msg" validate:"omitempty"`
	Remark        *string `json:"remark" validate:"omitempty,max=500"`
}

// UserTransactionResponse 用户交易响应
type UserTransactionResponse struct {
	ID            uint      `json:"id"`
	UserID        uint64    `json:"user_id"`
	Chain         string    `json:"chain"`
	Symbol        string    `json:"symbol"`
	FromAddress   string    `json:"from_address"`
	ToAddress     string    `json:"to_address"`
	Amount        string    `json:"amount"`
	Fee           string    `json:"fee"`
	GasLimit      *uint     `json:"gas_limit"`
	GasPrice      *string   `json:"gas_price"`
	Nonce         *uint64   `json:"nonce"`
	Status        string    `json:"status"`
	TxHash        *string   `json:"tx_hash"`
	BlockHeight   *uint64   `json:"block_height"`
	Confirmations *uint     `json:"confirmations"`
	ErrorMsg      *string   `json:"error_msg"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// ERC-20合约操作相关字段
	TransactionType       string `json:"transaction_type,omitempty"`
	ContractOperationType string `json:"contract_operation_type,omitempty"`
	TokenContractAddress  string `json:"token_contract_address,omitempty"`
}

// UserTransactionListResponse 用户交易列表响应
type UserTransactionListResponse struct {
	Transactions []UserTransactionResponse `json:"transactions"`
	Total        int64                     `json:"total"`
	Page         int                       `json:"page"`
	PageSize     int                       `json:"page_size"`
	HasMore      bool                      `json:"has_more"`
}

// ExportTransactionRequest 导出交易请求
type ExportTransactionRequest struct {
	ID uint `json:"id" binding:"required" validate:"required"`
}

// ExportTransactionResponse 导出交易响应
type ExportTransactionResponse struct {
	UnsignedTx  string  `json:"unsigned_tx"`
	Chain       string  `json:"chain"`
	Symbol      string  `json:"symbol"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      string  `json:"amount"`
	Fee         string  `json:"fee"`
	GasLimit    *uint   `json:"gas_limit"`
	GasPrice    *string `json:"gas_price"`
	Nonce       *uint64 `json:"nonce"`
}

// ImportSignatureRequest 导入签名请求
type ImportSignatureRequest struct {
	ID       uint   `json:"id" binding:"required" validate:"required"`
	SignedTx string `json:"signed_tx" binding:"required" validate:"required"`
}

// SendTransactionRequest 发送交易请求
type SendTransactionRequest struct {
	ID uint `json:"id" binding:"required" validate:"required"`
}

// UserTransactionStatsResponse 用户交易统计响应
type UserTransactionStatsResponse struct {
	TotalTransactions int64 `json:"total_transactions"`
	DraftCount        int64 `json:"draft_count"`
	UnsignedCount     int64 `json:"unsigned_count"`
	UnsentCount       int64 `json:"unsent_count"`
	InProgressCount   int64 `json:"in_progress_count"`
	PackedCount       int64 `json:"packed_count"`
	ConfirmedCount    int64 `json:"confirmed_count"`
	FailedCount       int64 `json:"failed_count"`
}
