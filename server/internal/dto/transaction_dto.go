package dto

import (
	"encoding/json"
	"time"

	"blockChainBrowser/server/internal/models"
)

// TransactionReceiptData 交易凭证数据（对齐go-ethereum Receipt的JSON字段）
type TransactionReceiptData struct {
	Type              interface{}              `json:"type,omitempty"`
	Root              *string                  `json:"root,omitempty"`
	Status            interface{}              `json:"status,omitempty"`
	ReceiptType       interface{}              `json:"receiptType,omitempty"`
	PostState         *string                  `json:"postState,omitempty"`
	CumulativeGasUsed interface{}              `json:"cumulativeGasUsed,omitempty"`
	LogsBloom         *string                  `json:"logsBloom,omitempty"`
	Logs              []map[string]interface{} `json:"logs,omitempty"`
	TransactionHash   *string                  `json:"transactionHash,omitempty"`
	ContractAddress   *string                  `json:"contractAddress,omitempty"`
	GasUsed           interface{}              `json:"gasUsed,omitempty"`
	EffectiveGasPrice interface{}              `json:"effectiveGasPrice,omitempty"`
	BlobGasUsed       interface{}              `json:"blobGasUsed,omitempty"`
	BlobGasPrice      interface{}              `json:"blobGasPrice,omitempty"`
	BlockHash         *string                  `json:"blockHash,omitempty"`
	BlockNumber       interface{}              `json:"blockNumber,omitempty"`
	TransactionIndex  interface{}              `json:"transactionIndex,omitempty"`
}

// BTCTransactionTurnsData BTC原生交易冗余数据（存入btc_transaction表）
type BTCTransactionTurnsData struct {
	TxID        string                   `json:"tx_id"`
	BlockHash   string                   `json:"block_hash"`
	BlockHeight uint64                   `json:"block_height"`
	From        string                   `json:"from"`
	To          string                   `json:"to"`
	Amount      string                   `json:"amount"`
	Fee         string                   `json:"fee"`
	Size        uint                     `json:"size"`
	Weight      uint                     `json:"weight"`
	LockTime    uint32                   `json:"locktime"`
	Hex         string                   `json:"hex"`
	Vin         []map[string]interface{} `json:"vin"`
	Vout        []map[string]interface{} `json:"vout"`
}

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
	Height       uint64  `json:"height" validate:"required,gt=0"`
	ContractAddr string  `json:"contract_addr" validate:"required"`
	Hex          *string `json:"hex,omitempty"`
	TxScene      string  `json:"tx_scene" validate:"required"`
	Remark       string  `json:"remark" validate:"max=256"`

	// 链相关字段（加入 bsc 和 sol）
	Chain  string `json:"chain" validate:"required,oneof=btc eth bsc sol"`
	Symbol string `json:"symbol" validate:"required,oneof=eth btc bsc sol"`

	// 地址字段
	AddressFrom  string  `json:"address_from" validate:"required"`
	AddressTo    string  `json:"address_to" validate:"required"`
	AddressFroms *string `json:"address_froms,omitempty"`
	AddressTos   *string `json:"address_tos,omitempty"`

	// Gas相关字段（EVM链）
	GasLimit uint64 `json:"gas_limit" validate:"required"`
	GasPrice string `json:"gas_price" validate:"required"`
	GasUsed  uint64 `json:"gas_used" validate:"gte=0"`

	// EIP-1559 相关字段（ETH/BSC）
	MaxFeePerGas         string `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas,omitempty"`
	EffectiveGasPrice    string `json:"effective_gas_price,omitempty"`

	// 手续费字段
	Fee        string  `json:"fee" validate:"required"`
	UsedFee    *string `json:"used_fee,omitempty"`
	BlockIndex uint    `json:"block_index" validate:"required"`
	Nonce      uint64  `json:"nonce" validate:"required"`

	// 新增字段
	Logs     []map[string]interface{} `json:"logs,omitempty"`
	Receipt  *TransactionReceiptData  `json:"receipt,omitempty"`
	BlockID  *uint64                  `json:"block_id,omitempty"`
	TurnsOut *BTCTransactionTurnsData `json:"turns_out,omitempty"`

	// BTC 原始交易数据字段
	Vin  *string `json:"vin,omitempty"`
	Vout *string `json:"vout,omitempty"`
}

// UpdateTransactionRequest 更新交易请求DTO
type UpdateTransactionRequest struct {
	Confirm    *uint   `json:"confirm,omitempty" validate:"omitempty,gte=0"`
	Status     *uint8  `json:"status,omitempty" validate:"omitempty,lte=3"`
	SendStatus *uint8  `json:"send_status,omitempty" validate:"omitempty,oneof=0 1"`
	GasUsed    *uint64 `json:"gas_used,omitempty" validate:"omitempty,gte=0"`
	UsedFee    *string `json:"used_fee,omitempty"`
	Hex        *string `json:"hex,omitempty"`
	Remark     *string `json:"remark,omitempty" validate:"omitempty,max=256"`
}

// TransactionResponse 交易响应DTO
type TransactionResponse struct {
	ID                   uint              `json:"id"`
	TxID                 string            `json:"tx_id"`
	TxType               uint8             `json:"tx_type"`
	Confirm              uint              `json:"confirm"`
	Status               uint8             `json:"status"`
	SendStatus           uint8             `json:"send_status"`
	Balance              string            `json:"balance"`
	Amount               string            `json:"amount"`
	TransID              uint              `json:"trans_id"`
	Symbol               string            `json:"symbol"`
	AddressFrom          string            `json:"address_from"`
	AddressTo            string            `json:"address_to"`
	GasLimit             uint64            `json:"gas_limit"`
	GasPrice             string            `json:"gas_price"`
	GasUsed              uint64            `json:"gas_used"`
	MaxFeePerGas         string            `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas string            `json:"max_priority_fee_per_gas,omitempty"`
	EffectiveGasPrice    string            `json:"effective_gas_price,omitempty"`
	Fee                  string            `json:"fee"`
	UsedFee              *string           `json:"used_fee"`
	Height               uint64            `json:"height"`
	ContractAddr         string            `json:"contract_addr"`
	Hex                  *string           `json:"hex"`
	TxScene              string            `json:"tx_scene"`
	Remark               string            `json:"remark"`
	IsToken              bool              `json:"is_token"`
	TokenName            string            `json:"token_name,omitempty"`
	TokenSymbol          string            `json:"token_symbol,omitempty"`
	TokenDecimals        uint8             `json:"token_decimals,omitempty"`
	TokenDescription     string            `json:"token_description,omitempty"`
	TokenWebsite         string            `json:"token_website,omitempty"`
	TokenExplorer        string            `json:"token_explorer,omitempty"`
	TokenLogo            string            `json:"token_logo,omitempty"`
	TokenMarketCapRank   *int              `json:"token_market_cap_rank,omitempty"`
	TokenIsStablecoin    bool              `json:"token_is_stablecoin,omitempty"`
	TokenIsVerified      bool              `json:"token_is_verified,omitempty"`
	Nonce                uint64            `json:"nonce"`
	Vin                  *string           `json:"vin,omitempty"`
	Vout                 *string           `json:"vout,omitempty"`
	BTCUTXOs             []*models.BTCUTXO `json:"btc_utxos,omitempty"`
	Ctime                time.Time         `json:"ctime"`
	Mtime                time.Time         `json:"mtime"`
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
	// 处理日志数据
	var logsJSON string
	if req.Logs != nil {
		if logsBytes, err := json.Marshal(req.Logs); err == nil {
			logsJSON = string(logsBytes)
		}
	}

	return &models.Transaction{
		TxID:         req.TxID,
		TxType:       req.TxType,
		Confirm:      req.Confirm,
		Status:       req.Status,
		SendStatus:   req.SendStatus,
		Balance:      req.Balance,
		Amount:       req.Amount,
		TransID:      req.TransID,
		Height:       req.Height,
		ContractAddr: req.ContractAddr,
		Hex:          req.Hex,
		TxScene:      req.TxScene,
		Remark:       req.Remark,
		Chain:        req.Chain,
		Symbol:       req.Symbol,
		AddressFrom:  req.AddressFrom,
		AddressTo:    req.AddressTo,
		AddressFroms: func() string {
			if req.AddressFroms != nil {
				return *req.AddressFroms
			}
			return ""
		}(),
		AddressTos: func() string {
			if req.AddressTos != nil {
				return *req.AddressTos
			}
			return ""
		}(),
		GasLimit:             req.GasLimit,
		GasPrice:             req.GasPrice,
		GasUsed:              req.GasUsed,
		MaxFeePerGas:         req.MaxFeePerGas,
		MaxPriorityFeePerGas: req.MaxPriorityFeePerGas,
		EffectiveGasPrice:    req.EffectiveGasPrice,
		Fee:                  req.Fee,
		UsedFee:              req.UsedFee,
		BlockIndex:           req.BlockIndex,
		BlockID:              req.BlockID,
		Nonce:                req.Nonce,
		Logs:                 logsJSON,
		Vin: func() string {
			if req.Vin != nil {
				return *req.Vin
			}
			return ""
		}(),
		Vout: func() string {
			if req.Vout != nil {
				return *req.Vout
			}
			return ""
		}(),
		Ctime: time.Now(),
		Mtime: time.Now(),
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
	resp.MaxFeePerGas = tx.MaxFeePerGas
	resp.MaxPriorityFeePerGas = tx.MaxPriorityFeePerGas
	resp.EffectiveGasPrice = tx.EffectiveGasPrice
	resp.Fee = tx.Fee
	resp.UsedFee = tx.UsedFee
	resp.Height = tx.Height
	resp.ContractAddr = tx.ContractAddr
	resp.Hex = tx.Hex
	resp.TxScene = tx.TxScene
	resp.Remark = tx.Remark
	resp.IsToken = tx.IsToken
	resp.TokenName = tx.TokenName                   // 添加代币名称
	resp.TokenSymbol = tx.TokenSymbol               // 添加代币符号
	resp.TokenDecimals = tx.TokenDecimals           // 添加代币精度
	resp.TokenDescription = tx.TokenDescription     // 添加代币描述
	resp.TokenWebsite = tx.TokenWebsite             // 添加代币官网
	resp.TokenExplorer = tx.TokenExplorer           // 添加代币浏览器链接
	resp.TokenLogo = tx.TokenLogo                   // 添加代币Logo
	resp.TokenMarketCapRank = tx.TokenMarketCapRank // 添加市值排名
	resp.TokenIsStablecoin = tx.TokenIsStablecoin   // 添加是否为稳定币
	resp.TokenIsVerified = tx.TokenIsVerified       // 添加是否已验证
	resp.Nonce = tx.Nonce
	resp.Vin = func() *string {
		if tx.Vin != "" {
			return &tx.Vin
		}
		return nil
	}()
	resp.Vout = func() *string {
		if tx.Vout != "" {
			return &tx.Vout
		}
		return nil
	}()
	resp.Ctime = tx.Ctime
	resp.Mtime = tx.Mtime
}

// NewTransactionResponse 创建交易响应
func NewTransactionResponse(tx *models.Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:                   tx.ID,
		TxID:                 tx.TxID,
		TxType:               tx.TxType,
		Confirm:              tx.Confirm,
		Status:               tx.Status,
		SendStatus:           tx.SendStatus,
		Balance:              tx.Balance,
		Amount:               tx.Amount,
		TransID:              tx.TransID,
		Symbol:               tx.Symbol,
		AddressFrom:          tx.AddressFrom,
		AddressTo:            tx.AddressTo,
		GasLimit:             tx.GasLimit,
		GasPrice:             tx.GasPrice,
		GasUsed:              tx.GasUsed,
		MaxFeePerGas:         tx.MaxFeePerGas,
		MaxPriorityFeePerGas: tx.MaxPriorityFeePerGas,
		EffectiveGasPrice:    tx.EffectiveGasPrice,
		Fee:                  tx.Fee,
		UsedFee:              tx.UsedFee,
		Height:               tx.Height,
		ContractAddr:         tx.ContractAddr,
		Hex:                  tx.Hex,
		TxScene:              tx.TxScene,
		Remark:               tx.Remark,
		IsToken:              tx.IsToken,
		TokenName:            tx.TokenName,          // 添加代币名称
		TokenSymbol:          tx.TokenSymbol,        // 添加代币符号
		TokenDecimals:        tx.TokenDecimals,      // 添加代币精度
		TokenDescription:     tx.TokenDescription,   // 添加代币描述
		TokenWebsite:         tx.TokenWebsite,       // 添加代币官网
		TokenExplorer:        tx.TokenExplorer,      // 添加代币浏览器链接
		TokenLogo:            tx.TokenLogo,          // 添加代币Logo
		TokenMarketCapRank:   tx.TokenMarketCapRank, // 添加市值排名
		TokenIsStablecoin:    tx.TokenIsStablecoin,  // 添加是否为稳定币
		TokenIsVerified:      tx.TokenIsVerified,    // 添加是否已验证
		Nonce:                tx.Nonce,
		Vin: func() *string {
			if tx.Vin != "" {
				return &tx.Vin
			}
			return nil
		}(),
		Vout: func() *string {
			if tx.Vout != "" {
				return &tx.Vout
			}
			return nil
		}(),
		Ctime:    tx.Ctime,
		Mtime:    tx.Mtime,
		BTCUTXOs: tx.BTCUTXOs,
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

// AddressTransactionResponse 地址交易响应
type AddressTransactionResponse struct {
	ID                   uint   `json:"id"`
	TxID                 string `json:"tx_id"`
	Height               uint64 `json:"height"`                   // 区块高度
	BlockIndex           uint   `json:"block_index"`              // 所在区块位置
	AddressFrom          string `json:"address_from"`             // 发送方地址
	AddressTo            string `json:"address_to"`               // 接收方地址
	Amount               string `json:"amount"`                   // 交易金额
	GasLimit             uint64 `json:"gas_limit"`                // Gas限制
	GasPrice             string `json:"gas_price"`                // Gas价格
	GasUsed              uint64 `json:"gas_used"`                 // 实际使用的Gas
	MaxFeePerGas         string `json:"max_fee_per_gas"`          // 最高费用
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"` // 最高小费
	EffectiveGasPrice    string `json:"effective_gas_price"`      // 有效Gas价格
	Fee                  string `json:"fee"`                      // 手续费
	Status               uint8  `json:"status"`                   // 交易状态
	Confirm              uint   `json:"confirm"`                  // 确认数
	Chain                string `json:"chain"`                    // 链类型
	Symbol               string `json:"symbol"`                   // 币种
	ContractAddr         string `json:"contract_addr"`            // 合约地址
	Ctime                string `json:"ctime"`                    // 创建时间
	Mtime                string `json:"mtime"`                    // 修改时间
}

// AddressTransactionsResponse 地址交易列表响应
type AddressTransactionsResponse struct {
	Transactions []AddressTransactionResponse `json:"transactions"`
	Total        int64                        `json:"total"`
	Page         int                          `json:"page"`
	PageSize     int                          `json:"page_size"`
	HasMore      bool                         `json:"has_more"`
}

// GetAddressTransactionsRequest 获取地址交易请求
type GetAddressTransactionsRequest struct {
	Address  string `form:"address" binding:"required"`        // 地址
	Page     int    `form:"page" binding:"min=1"`              // 页码，从1开始
	PageSize int    `form:"page_size" binding:"min=1,max=100"` // 每页大小，最大100
	Chain    string `form:"chain"`                             // 链类型（可选）
}
