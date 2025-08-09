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

// BitcoinScanner 比特币扫块器
type BitcoinScanner struct {
	config     *config.ChainConfig
	httpClient *http.Client
}

// NewBitcoinScanner 创建新的比特币扫块器
func NewBitcoinScanner(cfg *config.ChainConfig) *BitcoinScanner {
	return &BitcoinScanner{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetLatestBlockHeight 获取最新区块高度
func (bs *BitcoinScanner) GetLatestBlockHeight() (uint64, error) {
	// 使用Blockstream API获取最新区块高度
	url := fmt.Sprintf("%s/blocks/tip/height", bs.config.ExplorerAPIURL)

	resp, err := bs.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block height: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var heightStr string
	if err := json.NewDecoder(resp.Body).Decode(&heightStr); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse height: %w", err)
	}

	return height, nil
}

// GetBlockByHeight 根据高度获取区块
func (bs *BitcoinScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	// 使用Blockstream API获取区块信息
	url := fmt.Sprintf("%s/block-height/%d", bs.config.ExplorerAPIURL, height)

	resp, err := bs.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var blockData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&blockData); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	// 解析区块数据
	block := &models.Block{
		Chain: "btc",
	}

	// 设置基本字段
	if hash, ok := blockData["id"].(string); ok {
		block.Hash = hash
	}
	if heightVal, ok := blockData["height"].(float64); ok {
		block.Height = uint64(heightVal)
	}
	if version, ok := blockData["version"].(float64); ok {
		block.Version = int(version)
	}
	if timestamp, ok := blockData["timestamp"].(float64); ok {
		block.Timestamp = time.Unix(int64(timestamp), 0)
	}
	if size, ok := blockData["size"].(float64); ok {
		block.Size = uint64(size)
	}
	if weight, ok := blockData["weight"].(float64); ok {
		block.Weight = uint64(weight)
	}
	if txCount, ok := blockData["tx_count"].(float64); ok {
		block.TransactionCount = int(txCount)
	}
	if difficulty, ok := blockData["difficulty"].(float64); ok {
		block.Difficulty = difficulty
	}
	if nonce, ok := blockData["nonce"].(float64); ok {
		block.Nonce = uint64(nonce)
	}
	if bits, ok := blockData["bits"].(string); ok {
		block.Bits = bits
	}
	if fee, ok := blockData["fee"].(float64); ok {
		block.Fee = fee
	}

	// 设置其他字段
	if prevHash, ok := blockData["previousblockhash"].(string); ok {
		block.PreviousHash = prevHash
	}
	if merkleRoot, ok := blockData["merkle_root"].(string); ok {
		block.MerkleRoot = merkleRoot
	}

	return block, nil
}

// ValidateBlock 验证区块
func (bs *BitcoinScanner) ValidateBlock(block *models.Block) error {
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

	// 验证哈希格式（64位十六进制）
	if len(block.Hash) != 64 {
		return fmt.Errorf("invalid hash length: %d", len(block.Hash))
	}

	// 验证哈希字符（十六进制）
	for _, c := range block.Hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("invalid hash characters")
		}
	}

	return nil
}

// GetBlockTransactions 获取区块交易
func (bs *BitcoinScanner) GetBlockTransactions(blockHash string) ([]map[string]interface{}, error) {
	// 使用Blockstream API获取区块交易
	url := fmt.Sprintf("%s/block/%s/txs", bs.config.ExplorerAPIURL, blockHash)

	resp, err := bs.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var transactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}

	return transactions, nil
}

// CalculateBlockStats 计算区块统计信息
func (bs *BitcoinScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	// 计算总金额
	var totalAmount float64
	for _, tx := range transactions {
		if value, ok := tx["value"].(float64); ok {
			totalAmount += value
		}
	}
	block.TotalAmount = totalAmount

	// 计算确认数（这里简化处理，实际应该从网络获取）
	block.Confirmations = 1
}
