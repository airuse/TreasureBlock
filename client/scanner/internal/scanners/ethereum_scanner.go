package scanners

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/failover"
	"blockChainBrowser/client/scanner/internal/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumScanner 以太坊扫块器 - 使用官方go-ethereum包
type EthereumScanner struct {
	config *config.ChainConfig
	// 客户端连接池
	localClient      *ethclient.Client
	externalClients  []*ethclient.Client
	currentNodeIndex int // 当前使用的外部节点索引
	// 故障转移管理器
	failoverManager *failover.FailoverManager
}

// NewEthereumScanner 创建新的以太坊扫块器
func NewEthereumScanner(cfg *config.ChainConfig) *EthereumScanner {
	scanner := &EthereumScanner{
		config:           cfg,
		externalClients:  make([]*ethclient.Client, 0),
		currentNodeIndex: 0,
	}

	// 初始化本地节点客户端
	if cfg.RPCURL != "" {
		if client, err := ethclient.Dial(cfg.RPCURL); err == nil {
			scanner.localClient = client
		}
	}

	// 初始化多个外部API客户端
	if len(cfg.ExplorerAPIURLs) > 0 {
		for _, apiURL := range cfg.ExplorerAPIURLs {
			if client, err := ethclient.Dial(apiURL); err == nil {
				scanner.externalClients = append(scanner.externalClients, client)
				fmt.Printf("[ETH Scanner] Successfully connected to external API: %s\n", apiURL)
			} else {
				fmt.Printf("[ETH Scanner] Warning: Failed to connect to external API %s: %v\n", apiURL, err)
			}
		}
	}

	// 创建故障转移管理器
	scanner.failoverManager = failover.NewFailoverManager(scanner.localClient, scanner.externalClients)

	return scanner
}

// GetLatestBlockHeight 获取最新区块高度
func (es *EthereumScanner) GetLatestBlockHeight() (uint64, error) {
	result, err := es.failoverManager.CallWithFailoverUint64("get latest block height", func(client *ethclient.Client) (uint64, error) {
		return client.BlockNumber(context.Background())
	})

	if err == nil {
		fmt.Printf("[ETH Scanner] Latest block height: %d\n", result)
	}
	return result, err
}

// GetBlockByHeight 根据高度获取区块
func (es *EthereumScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	fmt.Printf("[ETH Scanner] Scanning block at height: %d\n", height)

	result, err := es.failoverManager.CallWithFailoverRawBlock("get block by height", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(height)))
	})

	if err != nil {
		return nil, err
	}

	// 解析区块数据
	block := es.parseBlock(result)

	fmt.Printf("[ETH Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
		block.Height, block.Hash[:16]+"...", block.TransactionCount)
	return block, nil
}

// parseBlock 解析以太坊区块数据
func (es *EthereumScanner) parseBlock(block *types.Block) *models.Block {
	return &models.Block{
		Chain:            "eth",
		Hash:             block.Hash().Hex(),
		Height:           block.NumberU64(),
		Version:          0, // 以太坊区块没有Version字段，设为0
		Timestamp:        time.Unix(int64(block.Time()), 0),
		Size:             uint64(block.Size()),
		Weight:           block.GasLimit(),
		StrippedSize:     block.GasUsed(),
		TransactionCount: len(block.Transactions()),
		Difficulty:       float64(block.Difficulty().Uint64()),
		Nonce:            block.Nonce(),
		PreviousHash:     block.ParentHash().Hex(),
		MerkleRoot:       block.Root().Hex(),
		Confirmations:    1,                      // 简化处理
		Miner:            block.Coinbase().Hex(), // 获取矿工地址
	}
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

// extractTransactionsFromBlock 直接从区块中提取交易信息
func (es *EthereumScanner) extractTransactionsFromBlock(block *types.Block) []map[string]interface{} {
	transactions := make([]map[string]interface{}, len(block.Transactions()))

	for i, tx := range block.Transactions() {
		// 获取交易签名
		v, r, s := tx.RawSignatureValues()

		// 处理EIP-1559交易
		var gasPriceStr, maxFeePerGas, maxPriorityFeePerGas, effectiveGasPrice string
		var txType uint8

		if tx.Type() == 2 { // EIP-1559 交易
			txType = 2
			// EIP-1559 交易使用 MaxFeePerGas 和 MaxPriorityFeePerGas
			maxFeePerGas = tx.GasFeeCap().String()
			maxPriorityFeePerGas = tx.GasTipCap().String()

			// 计算有效gas价格 (base fee + priority fee)
			if block.BaseFee() != nil {
				baseFee := block.BaseFee()
				priorityFee := tx.GasTipCap()
				if priorityFee.Cmp(baseFee) > 0 {
					effectiveGasPrice = new(big.Int).Add(baseFee, priorityFee).String()
				} else {
					effectiveGasPrice = new(big.Int).Mul(baseFee, big.NewInt(2)).String()
				}
			} else {
				effectiveGasPrice = maxFeePerGas // 如果无法获取base fee，使用max fee
			}

			// 为了兼容性，设置gasPrice为effectiveGasPrice
			gasPriceStr = effectiveGasPrice
		} else { // Legacy 交易
			txType = 0
			gasPriceStr = tx.GasPrice().String()
			maxFeePerGas = "0"
			maxPriorityFeePerGas = "0"
			effectiveGasPrice = gasPriceStr
		}

		// 安全地获取 To 地址，合约创建交易可能为 nil
		var toAddress string
		if tx.To() != nil {
			toAddress = tx.To().Hex()
		} else {
			toAddress = "" // 合约创建交易
		}

		// 获取 From 地址 - 使用简单稳定的方法
		var fromAddress string
		// 使用LatestSignerForChainID，它会自动选择合适的签名器
		signer := types.LatestSignerForChainID(tx.ChainId())
		if sender, err := signer.Sender(tx); err == nil {
			fromAddress = sender.Hex()
		} else {
			fmt.Printf("[ETH Scanner] Warning: Failed to recover sender for tx %s: %v\n", tx.Hash().Hex(), err)
			fromAddress = ""
		}

		txData := map[string]interface{}{
			"hash":                 tx.Hash().Hex(),
			"nonce":                tx.Nonce(),
			"type":                 txType,
			"from":                 fromAddress, // 发送者地址
			"to":                   toAddress,
			"value":                tx.Value().String(),
			"gasPrice":             gasPriceStr,
			"maxFeePerGas":         maxFeePerGas,
			"maxPriorityFeePerGas": maxPriorityFeePerGas,
			"effectiveGasPrice":    effectiveGasPrice,
			"gasLimit":             tx.Gas(),                     // 原始gas limit
			"gasUsed":              tx.Gas(),                     // 暂时使用gas limit，后续可以通过receipt获取实际值
			"data":                 fmt.Sprintf("%x", tx.Data()), // 保存原始交易数据为hex字符串
			"raw_data":             tx.Data(),                    // 保存原始字节数据
			"v":                    v.String(),
			"r":                    r.String(),
			"s":                    s.String(),
			"block_index":          i, // 交易在区块中的索引位置
		}

		// 简化合约交易检测：仅检查是否为配置的代币地址
		if toAddress != "" && es.isConfiguredTokenAddress(toAddress) && len(tx.Data()) > 0 {
			txData["is_contract_tx"] = true
		} else {
			txData["is_contract_tx"] = false
		}

		transactions[i] = txData
	}

	return transactions
}

// isConfiguredTokenAddress 检查地址是否为配置的币种地址
func (es *EthereumScanner) isConfiguredTokenAddress(address string) bool {
	if address == "" {
		return false
	}

	// 检查地址是否在配置的币种地址列表中（包含从API获取的地址）
	for _, tokenAddr := range es.config.TokenAddresses {
		if strings.EqualFold(address, tokenAddr) {
			return true
		}
	}

	return false
}

// enrichTransactionsWithContractInfo 获取所有交易回执（并发处理）
func (es *EthereumScanner) enrichTransactionsWithContractInfo(transactions []map[string]interface{}) error {
	if len(transactions) == 0 {
		return nil
	}

	// 收集所有交易哈希
	var txHashes []string
	for _, tx := range transactions {
		if hash, ok := tx["hash"].(string); ok {
			txHashes = append(txHashes, hash)
		}
	}

	// 并发获取所有交易回执
	if len(txHashes) > 0 {
		if err := es.batchGetTransactionReceipts(transactions, txHashes); err != nil {
			fmt.Printf("[ETH Scanner] Warning: Failed to batch get transaction receipts: %v\n", err)
		}
	}

	return nil
}

// batchGetTransactionReceipts 高效并发获取所有交易回执
func (es *EthereumScanner) batchGetTransactionReceipts(transactions []map[string]interface{}, txHashes []string) error {
	if len(txHashes) == 0 {
		return nil
	}

	startTime := time.Now()
	fmt.Printf("[ETH Scanner] 🚀 Starting parallel fetch of %d transaction receipts...\n", len(txHashes))

	// 创建哈希到交易的映射
	hashToTxMap := make(map[string]int)
	for i, tx := range transactions {
		if hash, ok := tx["hash"].(string); ok {
			hashToTxMap[hash] = i
		}
	}

	// 并发结果结构
	type receiptResult struct {
		hash    string
		receipt *types.Receipt
		err     error
		index   int
	}

	// 动态调整并发数：小批量用更高并发，大批量适当降低
	maxConcurrency := 20
	if len(txHashes) > 500 {
		maxConcurrency = 15
	} else if len(txHashes) < 50 {
		maxConcurrency = len(txHashes)
	}

	fmt.Printf("[ETH Scanner] Using %d concurrent workers for %d receipts\n", maxConcurrency, len(txHashes))

	// 创建工作池
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(chan receiptResult, len(txHashes))

	// 启动所有并发获取任务
	for i, txHash := range txHashes {
		go func(hash string, idx int) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 使用智能负载均衡获取回执
			var receipt *types.Receipt
			var err error

			err = es.failoverManager.CallWithFailover("get transaction receipt", func(client *ethclient.Client) error {
				var receiptErr error
				receipt, receiptErr = client.TransactionReceipt(context.Background(), common.HexToHash(hash))
				return receiptErr
			})

			results <- receiptResult{
				hash:    hash,
				receipt: receipt,
				err:     err,
				index:   idx,
			}
		}(txHash, i)
	}

	// 收集所有结果
	successCount := 0
	failureCount := 0
	logCount := 0
	processedCount := 0

	for i := 0; i < len(txHashes); i++ {
		result := <-results
		processedCount++

		if result.err != nil {
			fmt.Printf("[ETH Scanner] ❌ Failed to get receipt for tx %s: %v\n", result.hash, result.err)
			failureCount++
			continue
		}

		// 更新交易信息
		if index, exists := hashToTxMap[result.hash]; exists && index < len(transactions) {
			tx := transactions[index]

			// 设置交易状态
			if result.receipt.Status == 1 {
				tx["status"] = "success"
			} else {
				tx["status"] = "failed"
			}

			// 设置实际使用的gas
			tx["gasUsed"] = result.receipt.GasUsed

			// 解析所有交易的日志（不仅仅是合约交易）
			if len(result.receipt.Logs) > 0 {
				es.parseContractLogs(tx, result.receipt)
				logCount += len(result.receipt.Logs)
			}

			successCount++
		}

		// 显示进度（每50个）
		if processedCount%50 == 0 {
			elapsed := time.Since(startTime)
			fmt.Printf("[ETH Scanner] 📈 Progress: %d/%d receipts processed (%.1f%%) in %v\n",
				processedCount, len(txHashes), float64(processedCount)/float64(len(txHashes))*100, elapsed)
		}
	}

	elapsed := time.Since(startTime)
	avgTime := float64(elapsed.Milliseconds()) / float64(len(txHashes))

	fmt.Printf("[ETH Scanner] 📊 Parallel Receipt Fetch Complete:\n")
	fmt.Printf("  ✅ Success: %d/%d (%.1f%%)\n", successCount, len(txHashes), float64(successCount)/float64(len(txHashes))*100)
	fmt.Printf("  ❌ Failed: %d/%d (%.1f%%)\n", failureCount, len(txHashes), float64(failureCount)/float64(len(txHashes))*100)
	fmt.Printf("  📋 Total logs parsed: %d\n", logCount)
	fmt.Printf("  ⏱️  Total time: %v (parallel with %d workers)\n", elapsed, maxConcurrency)
	fmt.Printf("  📈 Average: %.2fms per receipt\n", avgTime)
	fmt.Printf("  🚀 Rate: %.1f receipts/second\n", float64(len(txHashes))/elapsed.Seconds())
	fmt.Printf("  ⚡ Speedup vs serial: ~%.1fx faster\n", float64(maxConcurrency)*0.7) // 估算加速比

	return nil
}

// parseContractLogs 保存合约交易的原始日志数据
func (es *EthereumScanner) parseContractLogs(tx map[string]interface{}, receipt *types.Receipt) {
	if receipt == nil || len(receipt.Logs) == 0 {
		return
	}

	// 保存所有日志的原始数据，供后续手动解析使用
	var logs []map[string]interface{}
	for i, log := range receipt.Logs {
		logData := map[string]interface{}{
			"index":    i,
			"address":  log.Address.Hex(),
			"topics":   make([]string, len(log.Topics)),
			"data":     fmt.Sprintf("%x", log.Data),
			"raw_data": log.Data,
		}

		// 保存所有topics
		for j, topic := range log.Topics {
			logData["topics"].([]string)[j] = topic.Hex()
		}

		logs = append(logs, logData)
	}

	// 保存日志到交易数据中
	tx["logs"] = logs
	tx["log_count"] = len(logs)

	fmt.Printf("[ETH Scanner] Saved %d logs for transaction %s\n", len(logs), tx["hash"])
}

// GetBlockTransactionsFromBlock 直接从区块中获取交易信息（避免哈希不一致问题）
func (es *EthereumScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	// 这里我们需要通过区块高度重新获取完整的区块数据
	// 因为 models.Block 中只包含基本信息，不包含完整的交易数据
	ethBlock, err := es.failoverManager.CallWithFailoverRawBlock("get block by height for transactions", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(block.Height)))
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get block by height for transactions: %w", err)
	}

	// 直接从区块中提取交易信息
	transactions := es.extractTransactionsFromBlock(ethBlock)

	// 增强交易信息：检查合约代码、获取回执、解析日志
	if err := es.enrichTransactionsWithContractInfo(transactions); err != nil {
		fmt.Printf("[ETH Scanner] Warning: Failed to enrich transactions with contract info: %v\n", err)
		// 不返回错误，继续处理
	}

	return transactions, nil
}

// CalculateBlockStats 计算区块统计信息
func (es *EthereumScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	// 计算总gas使用量和总费用
	var totalGasUsed uint64
	totalFee := big.NewInt(0)
	totalValue := big.NewInt(0)
	legacyTxCount := 0
	eip1559TxCount := 0

	for _, tx := range transactions {
		// 获取实际的gas使用量
		if gasUsed, ok := tx["gasUsed"].(uint64); ok {
			totalGasUsed += gasUsed
		} else {
			// 如果没有gasUsed，回退到gasLimit
			if gasLimit, ok := tx["gasLimit"].(uint64); ok {
				totalGasUsed += gasLimit
				fmt.Printf("[ETH Scanner] Warning: Using gasLimit %d instead of gasUsed for tx\n", gasLimit)
			}
		}

		// 获取交易类型
		txType, _ := tx["type"].(uint8)
		if txType == 2 {
			eip1559TxCount++
		} else {
			legacyTxCount++
		}

		// 计算费用 - 优先使用effectiveGasPrice（EIP-1559兼容）
		var gasPrice *big.Int
		if effectiveGasPriceStr, ok := tx["effectiveGasPrice"].(string); ok && effectiveGasPriceStr != "" {
			if price, ok := new(big.Int).SetString(effectiveGasPriceStr, 10); ok {
				gasPrice = price
			}
		} else if gasPriceStr, ok := tx["gasPrice"].(string); ok {
			if price, ok := new(big.Int).SetString(gasPriceStr, 10); ok {
				gasPrice = price
			}
		}

		if gasPrice != nil {
			// 优先使用gasUsed，如果没有则使用gasLimit
			var gasForFee uint64
			if gasUsed, ok := tx["gasUsed"].(uint64); ok {
				gasForFee = gasUsed
			} else if gasLimit, ok := tx["gasLimit"].(uint64); ok {
				gasForFee = gasLimit
			} else {
				continue // 跳过这笔交易
			}

			// 计算这笔交易的费用：gasUsed * effectiveGasPrice
			txFee := new(big.Int).Mul(big.NewInt(int64(gasForFee)), gasPrice)
			totalFee.Add(totalFee, txFee)
		}

		// 获取交易价值
		if valueStr, ok := tx["value"].(string); ok {
			if value, ok := new(big.Int).SetString(valueStr, 10); ok {
				totalValue.Add(totalValue, value)
			}
		}
	}

	// 转换为ETH单位
	ethFee := new(big.Float).Quo(new(big.Float).SetInt(totalFee), new(big.Float).SetInt(big.NewInt(1e18)))
	ethValue := new(big.Float).Quo(new(big.Float).SetInt(totalValue), new(big.Float).SetInt(big.NewInt(1e18)))

	// 设置区块统计信息
	block.TotalAmount, _ = ethValue.Float64()
	block.Fee, _ = ethFee.Float64()
	block.Confirmations = 1

	// 记录详细的统计信息
	fmt.Printf("[ETH Scanner] Block %d stats: Gas used: %d, Total fee: %s ETH, Total value: %s ETH\n",
		block.Height, totalGasUsed, ethFee.Text('f', 18), ethValue.Text('f', 18))
	fmt.Printf("[ETH Scanner] Transaction types: Legacy: %d, EIP-1559: %d\n", legacyTxCount, eip1559TxCount)
}
