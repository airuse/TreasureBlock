package dto

import (
	"blockChainBrowser/server/internal/models"
	"time"
)

// CreateTransactionRequest 创建交易请求DTO - 匹配现有Transaction表结构
type CreateTransactionRequest struct {
	TxID         string  `json:"tx_id" validate:"required,len=66"`
	TxType       uint8   `json:"tx_type" validate:"gte=0,lte=9"`
	Confirm      uint    `json:"confirm" validate:"gte=0"`
	Status       uint8   `json:"status" validate:"gte=0,lte=3"`
	SendStatus   uint8   `json:"send_status" validate:"oneof=0 1"`
	Balance      string  `json:"balance" validate:"required"`
	Amount       string  `json:"amount" validate:"required"`
	TransID      uint    `json:"trans_id" validate:"required"`
	Symbol       string  `json:"symbol" validate:"required,oneof=eth btc"`
	AddressFrom  string  `json:"address_from" validate:"required"`
	AddressTo    string  `json:"address_to" validate:"required"`
	GasLimit     uint    `json:"gas_limit" validate:"required"`
	GasPrice     string  `json:"gas_price" validate:"required"`
	GasUsed      uint    `json:"gas_used" validate:"gte=0"`
	Fee          string  `json:"fee" validate:"required"`
	UsedFee      *string `json:"used_fee,omitempty"`
	Height       uint64  `json:"height" validate:"required,gt=0"`
	ContractAddr string  `json:"contract_addr" validate:"required"`
	Hex          *string `json:"hex,omitempty"`
	TxScene      string  `json:"tx_scene" validate:"required"`
	Remark       string  `json:"remark" validate:"max=256"`
}

// UpdateTransactionRequest 更新交易请求DTO
type UpdateTransactionRequest struct {
	Confirm    *uint   `json:"confirm,omitempty" validate:"omitempty,gte=0"`
	Status     *uint8  `json:"status,omitempty" validate:"omitempty,lte=3"`
	SendStatus *uint8  `json:"send_status,omitempty" validate:"omitempty,oneof=0 1"`
	GasUsed    *uint   `json:"gas_used,omitempty" validate:"omitempty,gte=0"`
	UsedFee    *string `json:"used_fee,omitempty"`
	Hex        *string `json:"hex,omitempty"`
	Remark     *string `json:"remark,omitempty" validate:"omitempty,max=256"`
}

// TransactionResponse 交易响应DTO
type TransactionResponse struct {
	ID           uint      `json:"id"`
	TxID         string    `json:"tx_id"`
	TxType       uint8     `json:"tx_type"`
	Confirm      uint      `json:"confirm"`
	Status       uint8     `json:"status"`
	SendStatus   uint8     `json:"send_status"`
	Balance      string    `json:"balance"`
	Amount       string    `json:"amount"`
	TransID      uint      `json:"trans_id"`
	Symbol       string    `json:"symbol"`
	AddressFrom  string    `json:"address_from"`
	AddressTo    string    `json:"address_to"`
	GasLimit     uint      `json:"gas_limit"`
	GasPrice     string    `json:"gas_price"`
	GasUsed      uint      `json:"gas_used"`
	Fee          string    `json:"fee"`
	UsedFee      *string   `json:"used_fee"`
	Height       uint64    `json:"height"`
	ContractAddr string    `json:"contract_addr"`
	Hex          *string   `json:"hex"`
	TxScene      string    `json:"tx_scene"`
	Remark       string    `json:"remark"`
	Ctime        time.Time `json:"ctime"`
	Mtime        time.Time `json:"mtime"`
}

// TransactionSummaryResponse 交易摘要响应DTO
type TransactionSummaryResponse struct {
	TxID        string `json:"tx_id"`
	Symbol      string `json:"symbol"`
	Amount      string `json:"amount"`
	AddressFrom string `json:"address_from"`
	AddressTo   string `json:"address_to"`
	Status      uint8  `json:"status"`
	Height      uint64 `json:"height"`
}

// TransactionListResponse 交易列表响应DTO
type TransactionListResponse struct {
	Transactions []*TransactionResponse `json:"transactions"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"page_size"`
	TotalPages   int                    `json:"total_pages"`
}

// ToModel 将CreateTransactionRequest转换为Transaction模型
func (req *CreateTransactionRequest) ToModel() *models.Transaction {
	return &models.Transaction{
		TxID:         req.TxID,
		TxType:       req.TxType,
		Confirm:      req.Confirm,
		Status:       req.Status,
		SendStatus:   req.SendStatus,
		Balance:      req.Balance,
		Amount:       req.Amount,
		TransID:      req.TransID,
		Symbol:       req.Symbol,
		AddressFrom:  req.AddressFrom,
		AddressTo:    req.AddressTo,
		GasLimit:     req.GasLimit,
		GasPrice:     req.GasPrice,
		GasUsed:      req.GasUsed,
		Fee:          req.Fee,
		UsedFee:      req.UsedFee,
		Height:       req.Height,
		ContractAddr: req.ContractAddr,
		Hex:          req.Hex,
		TxScene:      req.TxScene,
		Remark:       req.Remark,
		Ctime:        time.Now(),
		Mtime:        time.Now(),
	}
}

// ApplyToModel 将UpdateTransactionRequest应用到Transaction模型
func (req *UpdateTransactionRequest) ApplyToModel(tx *models.Transaction) {
	if req.Confirm != nil {
		tx.Confirm = *req.Confirm
	}
	if req.Status != nil {
		tx.Status = *req.Status
	}
	if req.SendStatus != nil {
		tx.SendStatus = *req.SendStatus
	}
	if req.GasUsed != nil {
		tx.GasUsed = *req.GasUsed
	}
	if req.UsedFee != nil {
		tx.UsedFee = req.UsedFee
	}
	if req.Hex != nil {
		tx.Hex = req.Hex
	}
	if req.Remark != nil {
		tx.Remark = *req.Remark
	}
	tx.Mtime = time.Now()
}

// FromModel 将Transaction模型转换为TransactionResponse
func (resp *TransactionResponse) FromModel(tx *models.Transaction) {
	resp.ID = tx.ID
	resp.TxID = tx.TxID
	resp.TxType = tx.TxType
	resp.Confirm = tx.Confirm
	resp.Status = tx.Status
	resp.SendStatus = tx.SendStatus
	resp.Balance = tx.Balance
	resp.Amount = tx.Amount
	resp.TransID = tx.TransID
	resp.Symbol = tx.Symbol
	resp.AddressFrom = tx.AddressFrom
	resp.AddressTo = tx.AddressTo
	resp.GasLimit = tx.GasLimit
	resp.GasPrice = tx.GasPrice
	resp.GasUsed = tx.GasUsed
	resp.Fee = tx.Fee
	resp.UsedFee = tx.UsedFee
	resp.Height = tx.Height
	resp.ContractAddr = tx.ContractAddr
	resp.Hex = tx.Hex
	resp.TxScene = tx.TxScene
	resp.Remark = tx.Remark
	resp.Ctime = tx.Ctime
	resp.Mtime = tx.Mtime
}

// NewTransactionResponse 创建交易响应
func NewTransactionResponse(tx *models.Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:           tx.ID,
		TxID:         tx.TxID,
		TxType:       tx.TxType,
		Confirm:      tx.Confirm,
		Status:       tx.Status,
		SendStatus:   tx.SendStatus,
		Balance:      tx.Balance,
		Amount:       tx.Amount,
		TransID:      tx.TransID,
		Symbol:       tx.Symbol,
		AddressFrom:  tx.AddressFrom,
		AddressTo:    tx.AddressTo,
		GasLimit:     tx.GasLimit,
		GasPrice:     tx.GasPrice,
		GasUsed:      tx.GasUsed,
		Fee:          tx.Fee,
		UsedFee:      tx.UsedFee,
		Height:       tx.Height,
		ContractAddr: tx.ContractAddr,
		Hex:          tx.Hex,
		TxScene:      tx.TxScene,
		Remark:       tx.Remark,
		Ctime:        tx.Ctime,
		Mtime:        tx.Mtime,
	}
}

// NewTransactionSummaryResponse 创建TransactionSummaryResponse
func NewTransactionSummaryResponse(tx *models.Transaction) *TransactionSummaryResponse {
	return &TransactionSummaryResponse{
		TxID:        tx.TxID,
		Symbol:      tx.Symbol,
		Amount:      tx.Amount,
		AddressFrom: tx.AddressFrom,
		AddressTo:   tx.AddressTo,
		Status:      tx.Status,
		Height:      tx.Height,
	}
}

// NewTransactionListResponse 创建TransactionListResponse
func NewTransactionListResponse(txs []*models.Transaction, total int64, page, pageSize int) *TransactionListResponse {
	txResponses := make([]*TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = NewTransactionResponse(tx)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &TransactionListResponse{
		Transactions: txResponses,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}
}
