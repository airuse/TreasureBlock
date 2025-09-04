package pkg

import "encoding/json"

// TransactionData QR码中的交易数据结构
type TransactionData struct {
	ID         int              `json:"id"`                   // 交易ID
	ChainID    string           `json:"chainId"`              // 链ID
	Type       string           `json:"type"`                 // 链类型：eth 或 btc
	Nonce      uint64           `json:"nonce"`                // 交易序号
	From       string           `json:"from"`                 // 发送地址
	To         string           `json:"to"`                   // 接收地址
	Value      string           `json:"value"`                // 交易金额（十六进制）
	Data       string           `json:"data"`                 // 交易数据（十六进制）
	AccessList []AccessListItem `json:"accessList,omitempty"` // 访问列表（ETH EIP-2930）

	// EIP-1559费率字段
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"` // 最大优先费用（Gwei）
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`         // 最大费用（Gwei）
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
	if transaction.ChainID == "" {
		return nil, &ValidationError{Field: "chainId", Message: "链ID不能为空"}
	}
	if transaction.From == "" {
		return nil, &ValidationError{Field: "from", Message: "发送地址不能为空"}
	}
	if transaction.To == "" {
		return nil, &ValidationError{Field: "to", Message: "接收地址不能为空"}
	}

	return &transaction, nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

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
func (t *TransactionData) IsBTC() bool {
	return t.Type == "btc" || t.ChainID == "btc"
}

// GetChainName 获取链名称
func (t *TransactionData) GetChainName() string {
	switch t.ChainID {
	case "1", "eth":
		return "Ethereum"
	case "btc":
		return "Bitcoin"
	default:
		return "Unknown"
	}
}
