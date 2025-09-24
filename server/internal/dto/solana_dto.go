package dto

// SolTxDetailRequest Solana交易明细请求DTO（简化版本）
type SolTxDetailRequest struct {
	TxID              string  `json:"tx_id" validate:"required"`
	Slot              uint64  `json:"slot" validate:"required"`
	BlockID           *uint64 `json:"block_id"`
	Blockhash         string  `json:"blockhash" validate:"required"`
	RecentBlockhash   string  `json:"recent_blockhash"`
	Version           string  `json:"version"`
	Fee               uint64  `json:"fee"`
	ComputeUnits      uint64  `json:"compute_units"`
	Status            string  `json:"status"`
	AccountKeys       string  `json:"account_keys"`
	PreBalances       string  `json:"pre_balances"`
	PostBalances      string  `json:"post_balances"`
	PreTokenBalances  string  `json:"pre_token_balances"`
	PostTokenBalances string  `json:"post_token_balances"`
	Logs              string  `json:"logs"`
	Instructions      string  `json:"instructions"`
	InnerInstructions string  `json:"inner_instructions"`
	LoadedAddresses   string  `json:"loaded_addresses"`
	Rewards           string  `json:"rewards"`
	Events            string  `json:"events"`
	RawTransaction    string  `json:"raw_transaction"`
	RawMeta           string  `json:"raw_meta"`
}

// SolEventRequest Solana事件请求DTO
type SolEventRequest struct {
	TxID        string  `json:"tx_id" validate:"required"`
	BlockID     *uint64 `json:"block_id"`
	Slot        uint64  `json:"slot" validate:"required"`
	EventIndex  int     `json:"event_index"`
	EventType   string  `json:"event_type"`
	ProgramID   string  `json:"program_id"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      string  `json:"amount"`
	Mint        string  `json:"mint"`
	Decimals    int     `json:"decimals"`
	IsInner     bool    `json:"is_inner"`
	AssetType   string  `json:"asset_type"`
	ExtraData   string  `json:"extra_data"`
}

// SolInstructionRequest Solana指令请求DTO（简化版本）
type SolInstructionRequest struct {
	TxID             string  `json:"tx_id" validate:"required"`
	BlockID          *uint64 `json:"block_id"`
	Slot             uint64  `json:"slot" validate:"required"`
	InstructionIndex int     `json:"instruction_index"`
	ProgramID        string  `json:"program_id"`
	Accounts         string  `json:"accounts"`
	Data             string  `json:"data"`
	ParsedData       string  `json:"parsed_data"`
	InstructionType  string  `json:"instruction_type"`
	IsInner          bool    `json:"is_inner"`
	StackHeight      int     `json:"stack_height"`
}

// SolTxDetailCreateRequest 创建Solana交易明细请求DTO
type SolTxDetailCreateRequest struct {
	Detail       SolTxDetailRequest      `json:"detail" validate:"required"`
	Events       []SolEventRequest       `json:"events"`
	Instructions []SolInstructionRequest `json:"instructions"`
}

// SolTxDetailResponse Solana交易明细响应DTO
type SolTxDetailResponse struct {
	ID                uint    `json:"id"`
	TxID              string  `json:"tx_id"`
	Slot              uint64  `json:"slot"`
	BlockID           *uint64 `json:"block_id"`
	Blockhash         string  `json:"blockhash"`
	RecentBlockhash   string  `json:"recent_blockhash"`
	Version           string  `json:"version"`
	Fee               uint64  `json:"fee"`
	ComputeUnits      uint64  `json:"compute_units"`
	Status            string  `json:"status"`
	AccountKeys       string  `json:"account_keys"`
	PreBalances       string  `json:"pre_balances"`
	PostBalances      string  `json:"post_balances"`
	PreTokenBalances  string  `json:"pre_token_balances"`
	PostTokenBalances string  `json:"post_token_balances"`
	Logs              string  `json:"logs"`
	Instructions      string  `json:"instructions"`
	InnerInstructions string  `json:"inner_instructions"`
	LoadedAddresses   string  `json:"loaded_addresses"`
	Rewards           string  `json:"rewards"`
	Events            string  `json:"events"`
	RawTransaction    string  `json:"raw_transaction"`
	RawMeta           string  `json:"raw_meta"`
	Ctime             string  `json:"ctime"`
	Mtime             string  `json:"mtime"`
}

// SolEventResponse Solana事件响应DTO
type SolEventResponse struct {
	ID          uint    `json:"id"`
	TxID        string  `json:"tx_id"`
	BlockID     *uint64 `json:"block_id"`
	Slot        uint64  `json:"slot"`
	EventIndex  int     `json:"event_index"`
	EventType   string  `json:"event_type"`
	ProgramID   string  `json:"program_id"`
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      string  `json:"amount"`
	Mint        string  `json:"mint"`
	Decimals    int     `json:"decimals"`
	IsInner     bool    `json:"is_inner"`
	AssetType   string  `json:"asset_type"`
	ExtraData   string  `json:"extra_data"`
	Ctime       string  `json:"ctime"`
}

// SolInstructionResponse Solana指令响应DTO
type SolInstructionResponse struct {
	ID               uint    `json:"id"`
	TxID             string  `json:"tx_id"`
	BlockID          *uint64 `json:"block_id"`
	Slot             uint64  `json:"slot"`
	InstructionIndex int     `json:"instruction_index"`
	ProgramID        string  `json:"program_id"`
	Accounts         string  `json:"accounts"`
	Data             string  `json:"data"`
	ParsedData       string  `json:"parsed_data"`
	InstructionType  string  `json:"instruction_type"`
	IsInner          bool    `json:"is_inner"`
	StackHeight      int     `json:"stack_height"`
	Ctime            string  `json:"ctime"`
}

// BatchSolDataRequest 批量处理Solana数据请求DTO
type BatchSolDataRequest struct {
	Transactions []SolTxDetailCreateRequest `json:"transactions" validate:"required"`
}

// BatchSolDataResponse 批量处理Solana数据响应DTO
type BatchSolDataResponse struct {
	Success   bool     `json:"success"`
	Processed int      `json:"processed"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
	Message   string   `json:"message"`
}

// SlotStatsResponse Slot统计响应DTO
type SlotStatsResponse struct {
	Slot             uint64 `json:"slot"`
	TransactionCount int    `json:"transaction_count"`
	TotalFees        string `json:"total_fees"`
	SuccessfulTxs    int    `json:"successful_txs"`
	FailedTxs        int    `json:"failed_txs"`
	ComputeUnitsUsed uint64 `json:"compute_units_used"`
}

// ================= SOL 专用交易导出/导入/发送 DTO =================

// SolExportTransactionResponse 用于导出未签名的SOL交易数据
type SolExportTransactionResponse struct {
	ID              uint             `json:"id"`
	Chain           string           `json:"chain"`
	RecentBlockhash string           `json:"recent_blockhash"`
	FeePayer        string           `json:"fee_payer"`
	Version         string           `json:"version"`           // legacy 或 v0
	Instructions    []map[string]any `json:"instructions"`      // 指令计划（由前端/签名器构造最终指令）
	Context         map[string]any   `json:"context,omitempty"` // 额外上下文，例如 token decimals 等
}

// SolImportSignatureRequest 导入SOL签名（base64原始交易）
type SolImportSignatureRequest struct {
	ID         uint   `json:"id" binding:"required" validate:"required"`
	SignedBase string `json:"signed_base64" binding:"required" validate:"required"`
}

// SolSendTransactionRequest 发送SOL交易
type SolSendTransactionRequest struct {
	ID uint `json:"id" binding:"required" validate:"required"`
}

// SolSendTransactionResponse 发送SOL交易响应
type SolSendTransactionResponse struct {
	Success bool   `json:"success"`
	TxHash  string `json:"tx_hash"`
	Message string `json:"message"`
}
