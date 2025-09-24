package pkg

import "encoding/json"

// TransactionData QR码中的交易数据结构
type TransactionData struct {
	ID         int              `json:"id"`                   // 交易ID
	ChainID    string           `json:"chainId"`              // 链ID
	Type       string           `json:"type"`                 // 链类型：eth | btc | sol
	Nonce      uint64           `json:"nonce"`                // 交易序号
	From       string           `json:"from"`                 // 发送地址
	To         string           `json:"to"`                   // 接收地址
	Value      string           `json:"value"`                // 交易金额（十六进制）
	Data       string           `json:"data"`                 // 交易数据（十六进制）
	AccessList []AccessListItem `json:"accessList,omitempty"` // 访问列表（ETH EIP-2930）

	// Gas 限制（由后端在导出时估算并下发，签名时直接使用）
	Gas uint64 `json:"gas,omitempty"`

	// EIP-1559费率字段
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"` // 最大优先费用（Gwei）
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`         // 最大费用（Gwei）

	// BTC 特定字段
	Address string `json:"address,omitempty"` // BTC发送地址（用于地址匹配）
	MsgTx   *MsgTx `json:"MsgTx,omitempty"`   // BTC交易结构
}

// MsgTx BTC交易结构
type MsgTx struct {
	Version  int32     `json:"Version"`            // 交易版本
	TxIn     []TxIn    `json:"TxIn"`               // 交易输入
	TxOut    []TxOut   `json:"TxOut"`              // 交易输出
	LockTime uint32    `json:"LockTime"`           // 锁定时间
	PrevOuts []PrevOut `json:"PrevOuts,omitempty"` // 前置输出（用于segwit签名）
}

// TxIn BTC交易输入
type TxIn struct {
	Txid     string `json:"txid"`     // 前一个交易的哈希
	Vout     int    `json:"vout"`     // 前一个交易的输出索引
	Sequence uint32 `json:"sequence"` // 序列号
}

// PrevOut 前置输出（供segwit签名使用）
type PrevOut struct {
	Txid            string `json:"txid"`
	Vout            int    `json:"vout"`
	ValueSatoshi    int64  `json:"value_satoshi"`
	ScriptPubKeyHex string `json:"script_pubkey_hex"`
}

// TxOut BTC交易输出
type TxOut struct {
	ValueSatoshi int64  `json:"value_satoshi"` // 输出金额（satoshi）
	Address      string `json:"address"`       // 输出地址
}

// AccessListItem 访问列表项
type AccessListItem struct {
	Address     string   `json:"address"`     // 合约地址
	StorageKeys []string `json:"storageKeys"` // 存储键
}

// SignatureResult 签名结果
type SignatureResult struct {
	SignedTx string `json:"signed_tx"` // 完整的签名交易
	V        string `json:"v"`         // 签名V组件
	R        string `json:"r"`         // 签名R组件
	S        string `json:"s"`         // 签名S组件
}

// ParseQRCodeData 解析QR码数据
func ParseQRCodeData(qrData string) (*TransactionData, error) {
	var transaction TransactionData
	err := json.Unmarshal([]byte(qrData), &transaction)
	if err != nil {
		return nil, err
	}

	// 验证必需字段
	if transaction.ID == 0 {
		return nil, &ValidationError{Field: "id", Message: "交易ID不能为空"}
	}
	if transaction.Type == "" {
		return nil, &ValidationError{Field: "type", Message: "链类型不能为空"}
	}

	// 根据链类型进行不同的验证
	if transaction.IsBTC() {
		// BTC交易验证
		if transaction.Address == "" {
			return nil, &ValidationError{Field: "address", Message: "BTC发送地址不能为空"}
		}
		if transaction.MsgTx == nil {
			return nil, &ValidationError{Field: "MsgTx", Message: "BTC交易结构不能为空"}
		}
		if len(transaction.MsgTx.TxIn) == 0 {
			return nil, &ValidationError{Field: "TxIn", Message: "BTC交易输入不能为空"}
		}
		if len(transaction.MsgTx.TxOut) == 0 {
			return nil, &ValidationError{Field: "TxOut", Message: "BTC交易输出不能为空"}
		}
		// 设置From字段为Address字段的值，用于兼容现有代码
		transaction.From = transaction.Address
	} else if transaction.IsEVM() {
		// EVM交易验证（ETH/BSC等）
		if transaction.ChainID == "" {
			return nil, &ValidationError{Field: "chainId", Message: "链ID不能为空"}
		}
		if transaction.From == "" {
			return nil, &ValidationError{Field: "from", Message: "发送地址不能为空"}
		}
		if transaction.To == "" {
			return nil, &ValidationError{Field: "to", Message: "接收地址不能为空"}
		}
	} else if transaction.IsSOL() {
		// 对于SOL，这里仅做基本通过，实际签名数据从外层 SolanaUnsigned 解析
		// 保持与现有流程兼容
	} else {
		return nil, &ValidationError{Field: "type", Message: "不支持的链类型"}
	}

	return &transaction, nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// ToJSON 将交易数据转换为JSON字符串
func (t *TransactionData) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsETH 判断是否为ETH交易
func (t *TransactionData) IsETH() bool {
	return t.Type == "eth" || t.ChainID == "1" || t.ChainID == "eth"
}

// IsBTC 判断是否为BTC交易
func (t *TransactionData) IsBTC() bool { return t.Type == "btc" }

// IsBSC 判断是否为BSC交易
func (t *TransactionData) IsBSC() bool {
	// BSC 主网: 56, 测试网: 97
	return t.Type == "bsc" || t.ChainID == "56" || t.ChainID == "97"
}

// IsEVM 判断是否为EVM兼容链（ETH/BSC等）
func (t *TransactionData) IsEVM() bool {
	return t.IsETH() || t.IsBSC()
}

// ===== SOL =====

// SolanaUnsigned 本地组装未签名交易所需数据
type SolanaUnsigned struct {
	ID              int                      `json:"id"`
	Chain           string                   `json:"chain"`
	Type            string                   `json:"type"`
	Version         string                   `json:"version"`
	RecentBlockhash string                   `json:"recent_blockhash"`
	FeePayer        string                   `json:"fee_payer"`
	Instructions    []map[string]interface{} `json:"instructions"`
	Context         map[string]interface{}   `json:"context,omitempty"`
}

func (t *TransactionData) IsSOL() bool { return t.Type == "sol" }

// GetChainName 获取链名称
func (t *TransactionData) GetChainName() string {
	switch t.Type {
	case "eth":
		return "Ethereum"
	case "bsc":
		return "BNB Smart Chain"
	case "btc":
		return "Bitcoin"
	default:
		return "Unknown"
	}
}
