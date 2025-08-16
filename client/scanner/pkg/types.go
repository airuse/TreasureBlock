package pkg

import "time"

// ================== 通用响应 ==================

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ================== 扫块配置 ==================

// ScannerConfigResponse 扫块配置响应
type ScannerConfigResponse struct {
	ID          uint      `json:"id"`
	ConfigType  uint8     `json:"config_type"`
	ConfigGroup string    `json:"config_group"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	Status      uint8     `json:"status"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ScanConfig 客户端扫块配置
type ScanConfig struct {
	ScanInterval     time.Duration `json:"scan_interval"`
	Confirmations    int           `json:"confirmations"`
	StartBlockHeight uint64        `json:"start_block_height"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	SafetyHeight     uint64        `json:"safety_height"`
}

// RPCConfig RPC配置
type RPCConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ================== 区块相关 ==================

// BlockUploadRequest 区块上传请求（与后端DTO对齐）
type BlockUploadRequest struct {
	// 通用字段
	Hash             string    `json:"hash"`
	Height           uint64    `json:"height"`
	PreviousHash     string    `json:"previous_hash"`
	Timestamp        time.Time `json:"timestamp"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain"`

	// BTC特有字段
	MerkleRoot string `json:"merkle_root,omitempty"`
	Bits       string `json:"bits,omitempty"`
	Version    uint32 `json:"version,omitempty"`
	Weight     uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit   uint64 `json:"gas_limit,omitempty"`
	GasUsed    uint64 `json:"gas_used,omitempty"`
	Miner      string `json:"miner,omitempty"`
	ParentHash string `json:"parent_hash,omitempty"`
	Nonce      string `json:"nonce,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
}

// BlockResponse 区块响应（与后端DTO对齐）
type BlockResponse struct {
	ID               uint      `json:"id"`
	Hash             string    `json:"hash"`
	Height           uint64    `json:"height"`
	PreviousHash     string    `json:"previous_hash"`
	Timestamp        time.Time `json:"timestamp"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain"`

	// BTC特有字段
	MerkleRoot string `json:"merkle_root,omitempty"`
	Bits       string `json:"bits,omitempty"`
	Version    uint32 `json:"version,omitempty"`
	Weight     uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit   uint64 `json:"gas_limit,omitempty"`
	GasUsed    uint64 `json:"gas_used,omitempty"`
	Miner      string `json:"miner,omitempty"`
	ParentHash string `json:"parent_hash,omitempty"`
	Nonce      string `json:"nonce,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`

	// 时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CoinConfigData 币种配置数据结构
type CoinConfigData struct {
	Symbol       string `json:"symbol"`
	ChainName    string `json:"chain_name"`
	ContractAddr string `json:"contract_addr"`
	Status       int8   `json:"status"`
}

// CreateCoinConfigRequest 创建币种配置请求
type CreateCoinConfigRequest struct {
	ChainName        string  `json:"chain_name"`
	CoinType         uint8   `json:"coin_type"`
	ContractAddr     string  `json:"contract_addr"`
	Precision        uint    `json:"precision"`
	ColdAddress      string  `json:"cold_address"`
	ColdAddressHash  string  `json:"cold_address_hash"`
	MaxStock         float64 `json:"max_stock"`
	MaxBalance       float64 `json:"max_balance"`
	MinBalance       float64 `json:"min_balance"`
	CollectLimit     float64 `json:"collect_limit"`
	CollectLeft      float64 `json:"collect_left"`
	InternalGasLimit uint    `json:"internal_gas_limit"`
	OnceMinFee       float64 `json:"once_min_fee"`
	SymbolID         string  `json:"symbol_id"`
	Status           int8    `json:"status"`
}

// ContractInfo 合约信息
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

// ContractInfoRequest 合约信息请求（用于上传到合约API）
type ContractInfoRequest struct {
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
}
