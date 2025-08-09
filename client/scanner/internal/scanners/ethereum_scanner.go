package scanners

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/models"
)

// EthereumScanner 以太坊扫块器
type EthereumScanner struct {
	config     *config.ChainConfig
	httpClient *http.Client
}

// NewEthereumScanner 创建新的以太坊扫块器
func NewEthereumScanner(cfg *config.ChainConfig) *EthereumScanner {
	return &EthereumScanner{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetLatestBlockHeight 获取最新区块高度
func (es *EthereumScanner) GetLatestBlockHeight() (uint64, error) {
	// 使用Etherscan API获取最新区块高度
	url := fmt.Sprintf("%s?module=proxy&action=eth_blockNumber&apikey=%s",
		es.config.ExplorerAPIURL, es.config.APIKey)

	resp, err := es.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block height: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if result, ok := response["result"].(string); ok {
		// 移除0x前缀并转换为十进制
		if len(result) > 2 && result[:2] == "0x" {
			result = result[2:]
		}
		height, err := strconv.ParseUint(result, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse height: %w", err)
		}
		return height, nil
	}

	return 0, fmt.Errorf("invalid response format")
}

// GetBlockByHeight 根据高度获取区块
func (es *EthereumScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	// 使用Etherscan API获取区块信息
	hexHeight := fmt.Sprintf("0x%x", height)
	url := fmt.Sprintf("%s?module=proxy&action=eth_getBlockByNumber&tag=%s&boolean=true&apikey=%s",
		es.config.ExplorerAPIURL, hexHeight, es.config.APIKey)

	resp, err := es.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		block := &models.Block{
			Chain: "eth",
		}

		// 设置基本字段
		if hash, ok := result["hash"].(string); ok {
			block.Hash = hash
		}
		if number, ok := result["number"].(string); ok {
			if len(number) > 2 && number[:2] == "0x" {
				number = number[2:]
			}
			if heightVal, err := strconv.ParseUint(number, 16, 64); err == nil {
				block.Height = heightVal
			}
		}
		if timestamp, ok := result["timestamp"].(string); ok {
			if len(timestamp) > 2 && timestamp[:2] == "0x" {
				timestamp = timestamp[2:]
			}
			if ts, err := strconv.ParseUint(timestamp, 16, 64); err == nil {
				block.Timestamp = time.Unix(int64(ts), 0)
			}
		}
		if size, ok := result["size"].(string); ok {
			if len(size) > 2 && size[:2] == "0x" {
				size = size[2:]
			}
			if sizeVal, err := strconv.ParseUint(size, 16, 64); err == nil {
				block.Size = sizeVal
			}
		}
		if nonce, ok := result["nonce"].(string); ok {
			if len(nonce) > 2 && nonce[:2] == "0x" {
				nonce = nonce[2:]
			}
			if nonceVal, err := strconv.ParseUint(nonce, 16, 64); err == nil {
				block.Nonce = nonceVal
			}
		}
		if difficulty, ok := result["difficulty"].(string); ok {
			if len(difficulty) > 2 && difficulty[:2] == "0x" {
				difficulty = difficulty[2:]
			}
			if diffVal, err := strconv.ParseUint(difficulty, 16, 64); err == nil {
				block.Difficulty = float64(diffVal)
			}
		}
		if gasLimit, ok := result["gasLimit"].(string); ok {
			if len(gasLimit) > 2 && gasLimit[:2] == "0x" {
				gasLimit = gasLimit[2:]
			}
			if gasLimitVal, err := strconv.ParseUint(gasLimit, 16, 64); err == nil {
				block.Weight = gasLimitVal // 使用Weight字段存储gasLimit
			}
		}
		if gasUsed, ok := result["gasUsed"].(string); ok {
			if len(gasUsed) > 2 && gasUsed[:2] == "0x" {
				gasUsed = gasUsed[2:]
			}
			if gasUsedVal, err := strconv.ParseUint(gasUsed, 16, 64); err == nil {
				block.StrippedSize = gasUsedVal // 使用StrippedSize字段存储gasUsed
			}
		}
		if parentHash, ok := result["parentHash"].(string); ok {
			block.PreviousHash = parentHash
		}
		if transactions, ok := result["transactions"].([]interface{}); ok {
			block.TransactionCount = len(transactions)
		}

		return block, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

// ValidateBlock 验证区块
func (es *EthereumScanner) ValidateBlock(block *models.Block) error {
	// 基本验证
	if block.Hash == "" {
		return fmt.Errorf("block hash is empty")
	}

	if block.Height == 0 {
		return fmt.Errorf("block height is zero")
	}

	if block.Timestamp.IsZero() {
		return fmt.Errorf("block timestamp is zero")
	}

	// 验证哈希格式（66位，包含0x前缀）
	if len(block.Hash) != 66 || block.Hash[:2] != "0x" {
		return fmt.Errorf("invalid hash format: %s", block.Hash)
	}

	// 验证哈希字符（十六进制）
	for _, c := range block.Hash[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("invalid hash characters")
		}
	}

	return nil
}

// GetBlockTransactions 获取区块交易
func (es *EthereumScanner) GetBlockTransactions(blockHash string) ([]map[string]interface{}, error) {
	// 使用Etherscan API获取区块交易
	url := fmt.Sprintf("%s?module=proxy&action=eth_getBlockByHash&tag=%s&boolean=true&apikey=%s",
		es.config.ExplorerAPIURL, blockHash, es.config.APIKey)

	resp, err := es.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if transactions, ok := result["transactions"].([]interface{}); ok {
			var txs []map[string]interface{}
			for _, tx := range transactions {
				if txMap, ok := tx.(map[string]interface{}); ok {
					txs = append(txs, txMap)
				}
			}
			return txs, nil
		}
	}

	return nil, fmt.Errorf("invalid response format")
}

// CalculateBlockStats 计算区块统计信息
func (es *EthereumScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	// 计算总金额（以太坊中通常通过gas费用计算）
	var totalGasUsed uint64
	for _, tx := range transactions {
		if gasUsed, ok := tx["gasUsed"].(string); ok {
			if len(gasUsed) > 2 && gasUsed[:2] == "0x" {
				gasUsed = gasUsed[2:]
			}
			if gasUsedVal, err := strconv.ParseUint(gasUsed, 16, 64); err == nil {
				totalGasUsed += gasUsedVal
			}
		}
	}

	// 计算总费用（简化计算）
	block.TotalAmount = float64(totalGasUsed) * 20e9 // 假设gas price为20 Gwei
	block.Confirmations = 1
}
