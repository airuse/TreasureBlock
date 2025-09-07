package pkg

import (
	"encoding/json"
	"time"
)

// ================== 通用响应 ==================

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// GetLastVerifiedBlockHeightResponse 获取最后一个验证通过区块高度的响应
type GetLastVerifiedBlockHeightResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Chain  string `json:"chain"`
		Height string `json:"height"`
	} `json:"data"`
	Message string `json:"message"`
}

// LastVerifiedBlockHeightData 仅 data 部分的DTO（用于客户端通用解包）
type LastVerifiedBlockHeightData struct {
	Chain  string      `json:"chain"`
	Height json.Number `json:"height"`
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
	Confirmations int `json:"confirmations"`
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
	ChainID          int       `json:"chain_id"`

	// BTC特有字段
	MerkleRoot string `json:"merkle_root,omitempty"`
	Bits       string `json:"bits,omitempty"`
	Version    uint32 `json:"version,omitempty"`
	Weight     uint64 `json:"weight,omitempty"`

	// ETH特有字段
	GasLimit    uint64 `json:"gas_limit,omitempty"`
	GasUsed     uint64 `json:"gas_used,omitempty"`
	Miner       string `json:"miner,omitempty"`
	ParentHash  string `json:"parent_hash,omitempty"`
	Nonce       string `json:"nonce,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
	BaseFee     string `json:"base_fee,omitempty"`
	BurnedEth   string `json:"burned_eth,omitempty"`
	MinerTipEth string `json:"miner_tip_eth,omitempty"`

	// ETH状态根字段
	StateRoot        string `json:"state_root,omitempty"`
	TransactionsRoot string `json:"transactions_root,omitempty"`
	ReceiptsRoot     string `json:"receipts_root,omitempty"`
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
	GasLimit    uint64 `json:"gas_limit,omitempty"`
	GasUsed     uint64 `json:"gas_used,omitempty"`
	Miner       string `json:"miner,omitempty"`
	ParentHash  string `json:"parent_hash,omitempty"`
	Nonce       string `json:"nonce,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
	BaseFee     string `json:"base_fee,omitempty"`
	BurnedEth   string `json:"burned_eth,omitempty"`
	MinerTipEth string `json:"miner_tip_eth,omitempty"`

	// 时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CoinConfigData 币种配置数据结构
type CoinConfigData struct {
	ID            uint   `json:"id"`
	ChainName     string `json:"chain_name"`
	Symbol        string `json:"symbol"`
	CoinType      uint8  `json:"coin_type"`
	ContractAddr  string `json:"contract_addr"`
	Precision     uint   `json:"precision"`
	Decimals      uint   `json:"decimals"`
	Name          string `json:"name"`
	LogoURL       string `json:"logo_url"`
	WebsiteURL    string `json:"website_url"`
	ExplorerURL   string `json:"explorer_url"`
	Description   string `json:"description"`
	MarketCapRank uint   `json:"market_cap_rank"`
	IsStablecoin  bool   `json:"is_stablecoin"`
	IsVerified    bool   `json:"is_verified"`
	Status        int8   `json:"status"`
}

// CreateCoinConfigRequest 创建币种配置请求
type CreateCoinConfigRequest struct {
	ChainName     string `json:"chain_name"`
	Symbol        string `json:"symbol"`
	CoinType      uint8  `json:"coin_type"`
	ContractAddr  string `json:"contract_addr"`
	Precision     uint   `json:"precision"`
	Decimals      uint   `json:"decimals"`
	Name          string `json:"name"`
	LogoURL       string `json:"logo_url"`
	WebsiteURL    string `json:"website_url"`
	ExplorerURL   string `json:"explorer_url"`
	Description   string `json:"description"`
	MarketCapRank uint   `json:"market_cap_rank"`
	IsStablecoin  bool   `json:"is_stablecoin"`
	IsVerified    bool   `json:"is_verified"`
	Status        int8   `json:"status"`
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
