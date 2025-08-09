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
	BatchSize        int           `json:"batch_size"`
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

// BlockUploadRequest 区块上传请求
type BlockUploadRequest struct {
	Hash             string    `json:"hash"`
	Height           uint64    `json:"height"`
	PreviousHash     string    `json:"previous_hash"`
	MerkleRoot       string    `json:"merkle_root"`
	Timestamp        time.Time `json:"timestamp"`
	Difficulty       float64   `json:"difficulty"`
	Nonce            uint64    `json:"nonce"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain"`
}

// BlockResponse 区块响应
type BlockResponse struct {
	ID               uint      `json:"id"`
	Hash             string    `json:"hash"`
	Height           uint64    `json:"height"`
	PreviousHash     string    `json:"previous_hash"`
	MerkleRoot       string    `json:"merkle_root"`
	Timestamp        time.Time `json:"timestamp"`
	Difficulty       float64   `json:"difficulty"`
	Nonce            uint64    `json:"nonce"`
	Size             uint64    `json:"size"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"`
	Fee              float64   `json:"fee"`
	Confirmations    uint64    `json:"confirmations"`
	IsOrphan         bool      `json:"is_orphan"`
	Chain            string    `json:"chain"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
