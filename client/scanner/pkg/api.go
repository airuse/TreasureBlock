package pkg

import (
	"fmt"
	"strconv"
	"strings"

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

// UploadTransactionsBatch 批量上传交易
func (api *ScannerAPI) UploadTransactionsBatch(transactions []map[string]interface{}) error {
	if len(transactions) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"transactions": transactions,
	}

	if err := api.client.POST("/api/v1/transactions/create/batch", payload, nil); err != nil {
		return fmt.Errorf("batch upload transactions failed: %w", err)
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

// UploadContractInfoToContractAPI 上传合约信息到合约API
func (api *ScannerAPI) UploadContractInfoToContractAPI(contractInfo *ContractInfo) error {
	if err := api.client.POST("/api/v1/contracts", contractInfo, nil); err != nil {
		return fmt.Errorf("upload contract info failed: %w", err)
	}
	return nil
}

// GetLastVerifiedBlockHeight 获取最后一个验证通过的区块高度
func (api *ScannerAPI) GetLastVerifiedBlockHeight(chain string) (uint64, error) {
	var data LastVerifiedBlockHeightData

	endpoint := fmt.Sprintf("/api/v1/blocks/verification/last-verified?chain=%s", chain)
	if err := api.client.GET(endpoint, &data); err != nil {
		// 将 404 视为“没有记录”，返回高度 0 且无错误，避免噪声
		if strings.Contains(err.Error(), "HTTP 404") || strings.Contains(err.Error(), "record not found") {
			api.logger.Infof("[%s] No last verified block yet (treating as height 0)", chain)
			return 0, nil
		}
		return 0, fmt.Errorf("get last verified block height failed: %w", err)
	}

	// 支持既可能是字符串也可能是数字
	i, err := data.Height.Int64()
	if err != nil {
		return 0, fmt.Errorf("parse last verified block height failed: %w", err)
	}
	if i < 0 {
		return 0, fmt.Errorf("invalid height: %d", i)
	}
	return uint64(i), nil
}

// VerifyBlock 验证区块
func (api *ScannerAPI) VerifyBlock(blockID uint64) error {

	endpoint := fmt.Sprintf("/api/v1/blocks/%d/verify", blockID)
	if err := api.client.POST(endpoint, nil, nil); err != nil {
		return fmt.Errorf("verify block failed: %w", err)
	}

	return nil
}

// UpdateBlockStats 更新区块统计字段（供扫块端调用）
func (api *ScannerAPI) UpdateBlockStats(hash string, payload map[string]interface{}) error {
	// 统一由hash定位，后端UpdateBlock支持按hash查找
	body := map[string]interface{}{
		"hash": hash,
	}
	for k, v := range payload {
		body[k] = v
	}
	if err := api.client.POST("/api/v1/blocks/update", body, nil); err != nil {
		return fmt.Errorf("update block stats failed: %w", err)
	}
	return nil
}

// ================== 扩展：转账事件 & Solana 明细 ==================

// UploadTransferEventsBatch 批量上传跨链转账事件
func (api *ScannerAPI) UploadTransferEventsBatch(events []map[string]interface{}) error {
	if len(events) == 0 {
		return nil
	}
	body := map[string]interface{}{"events": events}
	if err := api.client.POST("/api/v1/transfers/create/batch", body, nil); err != nil {
		return fmt.Errorf("upload transfer events failed: %w", err)
	}
	return nil
}

// UploadSolTxDetail 上传单笔Sol交易明细及指令
func (api *ScannerAPI) UploadSolTxDetail(requestBody map[string]interface{}) error {
	if err := api.client.POST("/api/v1/sol/tx/detail", requestBody, nil); err != nil {
		return fmt.Errorf("upload sol tx detail failed: %w", err)
	}
	return nil
}

// UploadSolTxDetailBatch 批量上传Sol交易明细及指令
func (api *ScannerAPI) UploadSolTxDetailBatch(requestBody map[string]interface{}) error {
	if err := api.client.POST("/api/v1/sol/tx/detail/batch", requestBody, nil); err != nil {
		return fmt.Errorf("upload sol tx detail batch failed: %w", err)
	}
	return nil
}
