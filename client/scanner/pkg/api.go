package pkg

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// ScannerAPI 扫块器API - 所有业务接口都在这里
type ScannerAPI struct {
	client *Client
	logger *logrus.Logger
}

// NewScannerAPI 创建扫块器API实例
func NewScannerAPI(baseURL, apiKey, secretKey, environment string, logger *logrus.Logger) *ScannerAPI {
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &ScannerAPI{
		client: NewClient(baseURL, apiKey, secretKey, environment),
		logger: logger,
	}
}

// ================== 健康检查 ==================

// HealthCheck 健康检查
func (api *ScannerAPI) HealthCheck() error {
	return api.client.GET("/health", nil)
}

// ================== 扫块配置 ==================

// GetScannerConfig 获取扫块配置
func (api *ScannerAPI) GetScannerConfig(configType uint8, configGroup, configKey string) (*ScannerConfigResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/scanner/getconfig?configType=%d&configGroup=%s&configKey=%s",
		configType, configGroup, configKey)

	var result ScannerConfigResponse
	if err := api.client.GET(endpoint, &result); err != nil {
		return nil, fmt.Errorf("get scanner config failed: %w", err)
	}

	return &result, nil
}

// GetScanConfig 获取完整的扫块配置
func (api *ScannerAPI) GetScanConfig(chain string) (*ScanConfig, error) {
	config := &ScanConfig{}

	// 扫块间隔
	if interval, err := api.getConfigValue(1, "scan", "interval_"+chain); err == nil {
		if duration, err := time.ParseDuration(interval); err == nil {
			config.ScanInterval = duration
		} else {
			config.ScanInterval = 10 * time.Second
		}
	} else {
		config.ScanInterval = 10 * time.Second
	}

	// 确认数
	if confirmations, err := api.getConfigValue(1, "scan", "confirmations_"+chain); err == nil {
		if conf, err := strconv.Atoi(confirmations); err == nil {
			config.Confirmations = conf
		} else {
			config.Confirmations = 6
		}
	} else {
		config.Confirmations = 6
	}

	// 起始块高度
	if startHeight, err := api.getConfigValue(1, "scan", "start_height_"+chain); err == nil {
		if height, err := strconv.ParseUint(startHeight, 10, 64); err == nil {
			config.StartBlockHeight = height
		} else {
			config.StartBlockHeight = 0
		}
	} else {
		config.StartBlockHeight = 0
	}

	// 最大重试次数
	if maxRetries, err := api.getConfigValue(1, "scan", "max_retries_"+chain); err == nil {
		if retries, err := strconv.Atoi(maxRetries); err == nil {
			config.MaxRetries = retries
		} else {
			config.MaxRetries = 3
		}
	} else {
		config.MaxRetries = 3
	}

	// 重试延迟
	if retryDelay, err := api.getConfigValue(1, "scan", "retry_delay_"+chain); err == nil {
		if delay, err := time.ParseDuration(retryDelay); err == nil {
			config.RetryDelay = delay
		} else {
			config.RetryDelay = 5 * time.Second
		}
	} else {
		config.RetryDelay = 5 * time.Second
	}

	// 安全高度
	if safetyHeight, err := api.getConfigValue(1, "scan", "safety_height_"+chain); err == nil {
		if height, err := strconv.ParseUint(safetyHeight, 10, 64); err == nil {
			config.SafetyHeight = height
		} else {
			config.SafetyHeight = 12
		}
	} else {
		config.SafetyHeight = 12
	}

	api.logger.Infof("Loaded scan config for chain %s", chain)
	return config, nil
}

// GetRPCConfig 获取RPC配置
func (api *ScannerAPI) GetRPCConfig(chain string) (*RPCConfig, error) {
	config := &RPCConfig{}

	// RPC URL
	if rpcURL, err := api.getConfigValue(2, "rpc", "url_"+chain); err == nil {
		config.URL = rpcURL
	} else {
		return nil, fmt.Errorf("failed to get RPC URL for chain %s: %w", chain, err)
	}

	// API Key (可选)
	if apiKey, err := api.getConfigValue(2, "rpc", "api_key_"+chain); err == nil {
		config.APIKey = apiKey
	}

	// Username (可选)
	if username, err := api.getConfigValue(2, "rpc", "username_"+chain); err == nil {
		config.Username = username
	}

	// Password (可选)
	if password, err := api.getConfigValue(2, "rpc", "password_"+chain); err == nil {
		config.Password = password
	}

	return config, nil
}

// ================== 区块相关 ==================

// UploadBlock 上传区块
func (api *ScannerAPI) UploadBlock(block *BlockUploadRequest) (*BlockResponse, error) {
	var result BlockResponse
	if err := api.client.POST("/api/v1/blocks/create", block, &result); err != nil {
		return nil, fmt.Errorf("upload block failed: %w", err)
	}
	return &result, nil
}

// GetLatestBlock 获取最新区块
func (api *ScannerAPI) GetLatestBlock(chain string) (*BlockResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/blocks/latest?chain=%s", chain)
	var result BlockResponse
	if err := api.client.GET(endpoint, &result); err != nil {
		return nil, fmt.Errorf("get latest block failed: %w", err)
	}
	return &result, nil
}

// GetBlockByHash 根据哈希获取区块
func (api *ScannerAPI) GetBlockByHash(hash string) (*BlockResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/blocks/hash/%s", hash)
	var result BlockResponse
	if err := api.client.GET(endpoint, &result); err != nil {
		return nil, fmt.Errorf("get block by hash failed: %w", err)
	}
	return &result, nil
}

// GetBlockByHeight 根据高度获取区块
func (api *ScannerAPI) GetBlockByHeight(height uint64) (*BlockResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/blocks/height/%d", height)
	var result BlockResponse
	if err := api.client.GET(endpoint, &result); err != nil {
		return nil, fmt.Errorf("get block by height failed: %w", err)
	}
	return &result, nil
}

// ================== 私有方法 ==================

// getConfigValue 获取配置值
func (api *ScannerAPI) getConfigValue(configType uint8, configGroup, configKey string) (string, error) {
	config, err := api.GetScannerConfig(configType, configGroup, configKey)
	if err != nil {
		return "", fmt.Errorf("failed to get config %s.%s: %w", configGroup, configKey, err)
	}

	if config.Status != 1 {
		return "", fmt.Errorf("config %s.%s is disabled", configGroup, configKey)
	}

	return config.ConfigValue, nil
}

// ================== 币种配置相关 ==================

// GetAllCoinConfigs 获取所有币种配置
func (api *ScannerAPI) GetAllCoinConfigs() ([]CoinConfigData, error) {
	var result struct {
		Success bool             `json:"success"`
		Data    []CoinConfigData `json:"data"`
		Message string           `json:"message,omitempty"`
		Error   string           `json:"error,omitempty"`
	}

	if err := api.client.GET("/api/v1/coin-configs/scanner", &result); err != nil {
		return nil, fmt.Errorf("get all coin configs failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API returned error: %s", result.Error)
	}

	return result.Data, nil
}

// GetCoinConfigsByChain 根据链名称获取币种配置
func (api *ScannerAPI) GetCoinConfigsByChain(chain string) ([]CoinConfigData, error) {
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ChainName   string           `json:"chain_name"`
			CoinConfigs []CoinConfigData `json:"coin_configs"`
			TotalCount  int              `json:"total_count"`
		} `json:"data"`
		Message string `json:"message,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	url := fmt.Sprintf("/api/v1/coin-configs/chain/%s", chain)
	if err := api.client.GET(url, &result); err != nil {
		return nil, fmt.Errorf("get coin configs by chain failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API returned error: %s", result.Error)
	}

	return result.Data.CoinConfigs, nil
}

// GetAllContracts 获取所有合约
func (api *ScannerAPI) GetAllContracts() ([]ContractInfo, error) {
	var result struct {
		Success bool           `json:"success"`
		Data    []ContractInfo `json:"data"`
		Message string         `json:"message"`
		Error   string         `json:"error,omitempty"`
	}

	if err := api.client.GET("/api/v1/contracts", &result); err != nil {
		return nil, fmt.Errorf("get all contracts failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API returned error: %s", result.Error)
	}

	return result.Data, nil
}

// UploadTransaction 上传交易
func (api *ScannerAPI) UploadTransaction(tx map[string]interface{}) error {
	if err := api.client.POST("/api/v1/transactions/create", tx, nil); err != nil {
		return fmt.Errorf("upload transaction failed: %w", err)
	}
	return nil
}

// CreateCoinConfig 创建币种配置
func (api *ScannerAPI) CreateCoinConfig(config *CreateCoinConfigRequest) error {
	if err := api.client.POST("/api/v1/coin-configs", config, nil); err != nil {
		return fmt.Errorf("create coin config failed: %w", err)
	}
	return nil
}

// UploadContractInfoToContractAPI 上传合约信息到合约API（新的合约表）
func (api *ScannerAPI) UploadContractInfoToContractAPI(contractInfo *ContractInfo) error {
	if contractInfo == nil {
		return fmt.Errorf("contract info cannot be nil")
	}

	// 构造合约信息请求
	contractReq := &ContractInfoRequest{
		Address:      contractInfo.Address,
		Name:         contractInfo.Name,
		Symbol:       contractInfo.Symbol,
		Decimals:     contractInfo.Decimals,
		TotalSupply:  contractInfo.TotalSupply,
		ChainName:    contractInfo.ChainName,
		IsERC20:      contractInfo.IsERC20,
		ContractType: contractInfo.ContractType,
		Interfaces:   contractInfo.Interfaces,
		Methods:      contractInfo.Methods,
		Events:       contractInfo.Events,
		Metadata:     contractInfo.Metadata,
	}

	// 上传到合约API
	if err := api.client.POST("/api/v1/contracts", contractReq, nil); err != nil {
		return fmt.Errorf("failed to upload contract info: %w", err)
	}

	api.logger.Infof("Successfully uploaded contract info to contract API: %s (%s)",
		contractInfo.Symbol, contractInfo.Address)

	return nil
}

// BatchUploadContractsToContractAPI 批量上传合约信息到合约API
func (api *ScannerAPI) BatchUploadContractsToContractAPI(contractInfos []*ContractInfo) error {
	var successCount, failureCount int

	for _, contractInfo := range contractInfos {
		if err := api.UploadContractInfoToContractAPI(contractInfo); err != nil {
			fmt.Printf("[API] Warning: Failed to upload contract %s to contract API: %v\n", contractInfo.Address, err)
			failureCount++
			continue
		}
		successCount++
	}

	fmt.Printf("[API] Batch upload to contract API completed: %d success, %d failure\n", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("some contracts failed to upload to contract API: %d failures", failureCount)
	}

	return nil
}
