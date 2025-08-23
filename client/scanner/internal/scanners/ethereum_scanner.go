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
	currentNodeIndex int      // 当前使用的外部节点索引
	chainID          *big.Int // 缓存的网络链ID（作为回退）
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

	failoverManager := failover.NewFailoverManager(scanner.localClient, scanner.externalClients)
	chainID, err := failoverManager.CallWithFailoverNetworkID("get network id", func(client *ethclient.Client) (*big.Int, error) {
		return client.NetworkID(context.Background())
	})
	if err != nil {
		fmt.Printf("[ETH Scanner] Warning: Failed to detect chain ID: %v\n", err)
	}
	scanner.chainID = chainID
	return scanner
}

// GetLatestBlockHeight 获取最新区块高度
func (es *EthereumScanner) GetLatestBlockHeight() (uint64, error) {
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	result, err := failoverManager.CallWithFailoverUint64("get latest block height", func(client *ethclient.Client) (uint64, error) {
		return client.BlockNumber(context.Background())
	})

	if err == nil {
		fmt.Printf("[ETH Scanner] Latest block height: %d\n", result)
	}
	return result, err
}

// GetBlockByHeight 根据高度获取区块
func (es *EthereumScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	// fmt.Printf("[ETH Scanner] Scanning block at height: %d\n", height)
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	result, err := failoverManager.CallWithFailoverRawBlock("get block by height", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(height)))
	})

	if err != nil {
		return nil, err
	}
	if err := es.ValidateEthereumBlock(result); err != nil {
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
	parsedBlock := &models.Block{
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
		BaseFee:          block.BaseFee(),
	}
	return parsedBlock
}

// ValidateBlock 验证区块 - 使用go-ethereum内置验证
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

	// 注意：go-ethereum包已经内置了完整的区块验证逻辑
	// 包括哈希格式、区块头、交易等验证
	// 这里只做最基本的业务逻辑验证

	return nil
}

// ValidateEthereumBlock 验证以太坊区块 - 使用go-ethereum内置验证器
func (es *EthereumScanner) ValidateEthereumBlock(ethBlock *types.Block) error {
	if ethBlock == nil {
		return fmt.Errorf("ethereum block is nil")
	}

	// 验证区块头
	if ethBlock.Header() == nil {
		return fmt.Errorf("block header is nil")
	}

	// 验证区块哈希
	if ethBlock.Hash() == (common.Hash{}) {
		return fmt.Errorf("block hash is zero")
	}

	// 验证区块高度
	if ethBlock.NumberU64() == 0 {
		return fmt.Errorf("block height is zero")
	}

	// 验证时间戳
	if ethBlock.Time() == 0 {
		return fmt.Errorf("block timestamp is zero")
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
			// EIP-1559 使用 fee cap 与 tip cap
			feeCap := tx.GasFeeCap()
			tipCap := tx.GasTipCap()
			maxFeePerGas = feeCap.String()
			maxPriorityFeePerGas = tipCap.String()

			// 注意：这里的effectiveGasPrice只是预估，真实值需要从交易回执中获取
			// 在enrichTransactionsWithContractInfo中会用真实的EffectiveGasPrice覆盖
			var effective *big.Int
			if block.BaseFee() != nil && block.BaseFee().Sign() > 0 {
				// 有BaseFee的情况（Ethereum主网或支持EIP-1559的信太链）
				basePlusTip := new(big.Int).Add(block.BaseFee(), tipCap)
				if basePlusTip.Cmp(feeCap) < 0 {
					effective = basePlusTip
				} else {
					effective = feeCap
				}
			} else {
				// 没有BaseFee的情况（可能是信太链或旧链）
				if es.isL2Chain() {
					// 信太链可能有不同的费用机制，使用tipCap作为基础
					effective = tipCap
					fmt.Printf("[ETH Scanner] L2 chain detected, using tipCap as effective gas price: %s\n", effective.String())
				} else {
					// 旧链或未暴露 baseFee 时，退化为上限
					effective = feeCap
				}
			}
			effectiveGasPrice = effective.String()
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

		// 获取 From 地址 - 兼容处理不同签名方案，避免链ID为0导致的panic
		var fromAddress string
		var signer types.Signer
		if es.chainID != nil && es.chainID.Sign() > 0 {
			signer = types.LatestSignerForChainID(es.chainID)
		} else {
			// 如果链ID无效，使用 Homestead 签名器
			signer = types.HomesteadSigner{}
		}

		if sender, err := types.Sender(signer, tx); err == nil {
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
			"gasLimit":             tx.Gas(),                       // 原始gas limit
			"gasUsed":              tx.Gas(),                       // 暂时使用gas limit，后续可以通过receipt获取实际值
			"data":                 fmt.Sprintf("0x%x", tx.Data()), // 保存Input data为标准的0x前缀hex字符串
			"raw_data":             tx.Data(),                      // 保存原始字节数据（用于调试）
			"v":                    v.String(),
			"r":                    r.String(),
			"s":                    s.String(),
			"block_index":          i, // 交易在区块中的索引位置
		}

		// 简化合约交易检测：仅检查是否为配置的代币地址
		if toAddress != "" && es.isConfiguredTokenAddress(toAddress) && len(tx.Data()) > 0 {
			txData["is_contract_tx"] = true
			// 为代币交易设置合约地址
			txData["contract_address"] = toAddress
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
func (es *EthereumScanner) enrichTransactionsWithContractInfo(block *models.Block, transactions []map[string]interface{}) error {
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
		if err := es.batchGetTransactionReceipts(block, transactions, txHashes); err != nil {
			fmt.Printf("[ETH Scanner] Warning: Failed to batch get transaction receipts: %v\n", err)
		}
	}

	return nil
}

// batchGetTransactionReceipts 高效并发获取所有交易回执
func (es *EthereumScanner) batchGetTransactionReceipts(block *models.Block, transactions []map[string]interface{}, txHashes []string) error {
	if len(txHashes) == 0 {
		return nil
	}

	startTime := time.Now()
	// fmt.Printf("[ETH Scanner] 🚀 Starting parallel fetch of %d transaction receipts...\n", len(txHashes))

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

	// 从配置文件获取固定并发数
	maxConcurrency := es.config.Scan.MaxConcurrent
	if maxConcurrency <= 0 {
		maxConcurrency = 20 // 默认值
	}

	// fmt.Printf("[ETH Scanner] Using %d concurrent workers for %d receipts\n", maxConcurrency, len(txHashes))

	// 创建工作池
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(chan receiptResult, len(txHashes))
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	// 启动所有并发获取任务
	for i, txHash := range txHashes {
		go func(hash string, idx int) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 使用智能负载均衡获取回执
			var receipt *types.Receipt

			err := failoverManager.CallWithFailover("get transaction receipt", func(client *ethclient.Client) error {
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

			// 关键修复：使用交易回执中的真实EffectiveGasPrice
			if result.receipt.EffectiveGasPrice != nil {
				// 更新gas价格为真实值
				tx["effectiveGasPrice"] = result.receipt.EffectiveGasPrice.String()
				tx["gasPrice"] = result.receipt.EffectiveGasPrice.String()

				// 计算真实的交易费用
				realFee := new(big.Int).Mul(result.receipt.EffectiveGasPrice, big.NewInt(int64(result.receipt.GasUsed)))
				tx["realFee"] = realFee.String()

				// fmt.Printf("[ETH Scanner] Tx %s: Updated gas price to %s wei, real fee: %s wei\n",
				// 	result.hash, result.receipt.EffectiveGasPrice.String(), realFee.String())
			}

			// 关键修复：为所有交易保存完整的回执信息，不管有没有日志
			es.parseContractLogs(tx, result.receipt)

			tx["receipt"] = result.receipt
			successCount++
		}
	}

	elapsed := time.Since(startTime)

	stats := failoverManager.GetStats()
	fmt.Printf("[ETH Scanner] %d 📊 Parallel Receipt Fetch Complete:\n", block.Height)
	fmt.Printf("  ✅ Total Nmuber: %d\n", len(txHashes))
	fmt.Printf("  ⏱️ Total time: %v (parallel with %d workers)\n", elapsed, maxConcurrency)
	fmt.Printf("  📉 Stats: %+v\n", stats)

	return nil
}

// parseContractLogs 保存合约交易的原始日志数据
func (es *EthereumScanner) parseContractLogs(tx map[string]interface{}, receipt *types.Receipt) {
	if receipt == nil {
		return
	}

	// 关键修复：设置合约地址
	// 如果交易回执中有合约地址，说明这是合约创建交易
	if receipt.ContractAddress != (common.Address{}) {
		tx["contract_address"] = receipt.ContractAddress.Hex()
	} else if len(receipt.Logs) > 0 {
		// 如果有日志，说明这是合约调用交易，使用第一个日志的地址作为合约地址
		tx["contract_address"] = receipt.Logs[0].Address.Hex()
	}

	// 保存所有日志的原始数据，供后续手动解析使用
	var logs []map[string]interface{}
	if len(receipt.Logs) > 0 {
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
	}

	// 关键修复：保存完整的交易回执信息（不管有没有日志都要保存）
	if receipt.EffectiveGasPrice != nil {
		tx["effective_gas_price"] = receipt.EffectiveGasPrice.String()
	}

	if receipt.BlobGasUsed > 0 {
		tx["blob_gas_used"] = receipt.BlobGasUsed
	}

	if receipt.BlobGasPrice != nil {
		tx["blob_gas_price"] = receipt.BlobGasPrice.String()
	}

	if receipt.BlockNumber != nil {
		tx["block_number"] = receipt.BlockNumber.Uint64()
	}

	// fmt.Printf("[ETH Scanner] Saved %d logs for transaction %s\n", len(logs), tx["hash"])
}

// GetBlockTransactionsFromBlock 直接从区块中获取交易信息（避免哈希不一致问题）
func (es *EthereumScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	// 这里我们需要通过区块高度重新获取完整的区块数据
	// 因为 models.Block 中只包含基本信息，不包含完整的交易数据
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	ethBlock, err := failoverManager.CallWithFailoverRawBlock("get block by height for transactions", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(block.Height)))
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get block by height for transactions: %w", err)
	}

	// 直接从区块中提取交易信息
	transactions := es.extractTransactionsFromBlock(ethBlock)

	// 增强交易信息：检查合约代码、获取回执、解析日志
	if err := es.enrichTransactionsWithContractInfo(block, transactions); err != nil {
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
		// 获取实际的gas使用量 - 优先使用回执中的真实值
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

		// 获取交易价值
		if valueStr, ok := tx["value"].(string); ok {
			if value, ok := new(big.Int).SetString(valueStr, 10); ok {
				totalValue.Add(totalValue, value)
			}
		}
	}

	// 验证我们累加的 totalGasUsed 与区块实际 gasUsed 是否一致
	actualGasUsed := block.StrippedSize // 在 parseBlock 中我们把 block.GasUsed() 存到了 StrippedSize
	if totalGasUsed != actualGasUsed {
		fmt.Printf("[ETH Scanner] Warning: Block %d gas used mismatch: calculated=%d, actual=%d\n",
			block.Height, totalGasUsed, actualGasUsed)
	}

	// 重新计算真实的总费用 - 从交易回执中获取
	totalFee = big.NewInt(0)
	for _, tx := range transactions {
		// 从交易回执中获取真实的gas使用量和gas价格
		if receipt, ok := tx["receipt"].(*types.Receipt); ok && receipt != nil {
			// 获取真实使用的gas
			gasUsed := receipt.GasUsed

			// 获取真实的gas价格 - 优先使用回执中的EffectiveGasPrice
			var gasPrice *big.Int
			if receipt.EffectiveGasPrice != nil {
				// 优先使用回执中的EffectiveGasPrice（这是真实的价格）
				gasPrice = receipt.EffectiveGasPrice
			} else {
				// 回退到交易中的价格
				if txType, _ := tx["type"].(uint8); txType == 2 {
					// EIP-1559交易，使用effectiveGasPrice
					if effectiveGasPriceStr, ok := tx["effectiveGasPrice"].(string); ok && effectiveGasPriceStr != "" {
						if price, ok := new(big.Int).SetString(effectiveGasPriceStr, 10); ok {
							gasPrice = price
						}
					}
				} else {
					// Legacy交易，使用gasPrice
					if gasPriceStr, ok := tx["gasPrice"].(string); ok {
						if price, ok := new(big.Int).SetString(gasPriceStr, 10); ok {
							gasPrice = price
						}
					}
				}
			}

			if gasPrice != nil {
				// 计算这笔交易的费用：gasUsed * gasPrice
				txFee := new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)
				totalFee.Add(totalFee, txFee)

				// 记录每笔交易的费用信息
				// fmt.Printf("[ETH Scanner] Block %d Tx %s: GasUsed=%d, GasPrice=%s wei, Fee=%s wei\n",
				// 	block.Height, tx["hash"], gasUsed, gasPrice.String(), txFee.String())
			}
		}
	}

	// 计算矿工小费与燃烧 - 修正逻辑
	// 关键修正：只要区块有BaseFee，所有交易都会燃烧，不管交易类型
	if block.BaseFee != nil && block.BaseFee.Sign() > 0 {
		// 燃烧费 = baseFee * 区块实际gasUsed（这是协议规定的，对所有交易都适用）
		burnedWei := new(big.Int).Mul(new(big.Int).SetUint64(actualGasUsed), block.BaseFee)

		// 矿工小费 = 总费用 - 燃烧费
		minerTipWei := new(big.Int).Sub(totalFee, burnedWei)
		if minerTipWei.Sign() < 0 {
			// 如果小费不够，矿工可能亏损，但不会为负
			minerTipWei.SetInt64(0)
			fmt.Printf("[ETH Scanner] Warning: Block %d miner tip is negative, setting to 0\n", block.Height)
		}

		// 转换为ETH单位
		block.BurnedEth = new(big.Float).Quo(new(big.Float).SetInt(burnedWei), new(big.Float).SetInt(big.NewInt(1e18)))
		block.MinerTipEth = new(big.Float).Quo(new(big.Float).SetInt(minerTipWei), new(big.Float).SetInt(big.NewInt(1e18)))

		fmt.Printf("[ETH Scanner] Block %d: BaseFee=%s wei, ActualGasUsed=%d, TotalFee=%s wei\n",
			block.Height, block.BaseFee.String(), actualGasUsed, totalFee.String())
		fmt.Printf("[ETH Scanner] Block %d: BurnedWei=%s, MinerTipWei=%s, BurnedETH=%s, MinerTipETH=%s\n",
			block.Height, burnedWei.String(), minerTipWei.String(),
			block.BurnedEth.Text('f', 18), block.MinerTipEth.Text('f', 18))
	} else {

		// 没有燃烧，全部费用归矿工
		block.BurnedEth = new(big.Float).SetInt(big.NewInt(0))
		block.MinerTipEth = new(big.Float).Quo(new(big.Float).SetInt(totalFee), new(big.Float).SetInt(big.NewInt(1e18)))

		fmt.Printf("[ETH Scanner] Block %d: No burning (BaseFee=%v), TotalFee=%s wei, all fees to miner: %s ETH\n",
			block.Height, block.BaseFee, totalFee.String(), block.MinerTipEth.Text('f', 18))
	}

	// 转换为ETH单位
	ethFee := new(big.Float).Quo(new(big.Float).SetInt(totalFee), new(big.Float).SetInt(big.NewInt(1e18)))
	ethValue := new(big.Float).Quo(new(big.Float).SetInt(totalValue), new(big.Float).SetInt(big.NewInt(1e18)))

	// 设置区块统计信息
	block.TotalAmount, _ = ethValue.Float64()
	block.Fee, _ = ethFee.Float64()
	block.Confirmations = 1

	fmt.Printf("[ETH Scanner] Block %d: Legacy TXs=%d, EIP-1559 TXs=%d, Total Value=%s ETH, Total Fee=%s ETH\n",
		block.Height, legacyTxCount, eip1559TxCount, ethValue.Text('f', 18), ethFee.Text('f', 18))
}

// isL2Chain 检查是否为L2信太链
func (es *EthereumScanner) isL2Chain() bool {
	if es.chainID == nil {
		return false
	}

	// 常见的L2链ID
	l2ChainIDs := map[int64]string{
		137:   "Polygon",       // Polygon Mainnet
		80001: "Polygon Test",  // Polygon Mumbai
		42161: "Arbitrum One",  // Arbitrum One
		42170: "Arbitrum Nova", // Arbitrum Nova
		10:    "Optimism",      // Optimism
		8453:  "Base",          // Base
		1101:  "Polygon zkEVM", // Polygon zkEVM
		324:   "zkSync Era",    // zkSync Era
	}

	if chainName, exists := l2ChainIDs[es.chainID.Int64()]; exists {
		fmt.Printf("[ETH Scanner] Detected L2 chain: %s (ChainID: %d)\n", chainName, es.chainID.Int64())
		return true
	}

	return false
}
