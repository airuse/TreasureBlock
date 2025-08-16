package dto

// ContractInfo 合约信息（从客户端传入）
type ContractInfo struct {
	Address      string            `json:"address"`
	Name         string            `json:"name"`
	Symbol       string            `json:"symbol"`
	Decimals     uint8             `json:"decimals"`
	TotalSupply  string            `json:"total_supply"`
	ChainName    string            `json:"chain_name"`
	IsERC20      bool              `json:"is_erc20"`
	ContractType string            `json:"contract_type"`
	Interfaces   []string          `json:"interfaces"`
	Methods      []string          `json:"methods"`
	Events       []string          `json:"events"`
	Metadata     map[string]string `json:"metadata"`
	// 新增字段
	Status        int8   `json:"status"`         // 合约状态
	Verified      bool   `json:"verified"`       // 是否已验证
	Creator       string `json:"creator"`        // 创建者地址
	CreationTx    string `json:"creation_tx"`    // 创建交易哈希
	CreationBlock uint64 `json:"creation_block"` // 创建区块高度
}

// ContractResponse 合约响应（返回给客户端）
type ContractResponse struct {
	ID            uint              `json:"id"`
	Address       string            `json:"address"`
	Name          string            `json:"name"`
	Symbol        string            `json:"symbol"`
	Decimals      uint8             `json:"decimals"`
	TotalSupply   string            `json:"total_supply"`
	ChainName     string            `json:"chain_name"`
	IsERC20       bool              `json:"is_erc20"`
	ContractType  string            `json:"contract_type"`
	Interfaces    []string          `json:"interfaces"`
	Methods       []string          `json:"methods"`
	Events        []string          `json:"events"`
	Metadata      map[string]string `json:"metadata"`
	Status        int8              `json:"status"`
	Verified      bool              `json:"verified"`
	Creator       string            `json:"creator"`
	CreationTx    string            `json:"creation_tx"`
	CreationBlock uint64            `json:"creation_block"`
	CTime         string            `json:"ctime"` // 创建时间
	MTime         string            `json:"mtime"` // 更新时间
}
