package scanners

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/models"

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

	return scanner
}

// callWithFailoverUint64 通用的故障转移调用方法（返回uint64）
func (es *EthereumScanner) callWithFailoverUint64(operation string, clientCall func(*ethclient.Client) (uint64, error)) (uint64, error) {
	// 首先尝试本地节点
	if es.localClient != nil {
		result, err := clientCall(es.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[ETH Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(es.externalClients) > 0 {
		// 从当前节点开始尝试
		startIndex := es.currentNodeIndex
		for i := 0; i < len(es.externalClients); i++ {
			currentIndex := (startIndex + i) % len(es.externalClients)
			client := es.externalClients[currentIndex]

			result, err := clientCall(client)
			if err == nil {
				// 成功获取，更新当前节点索引
				es.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[ETH Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return 0, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverBlock 通用的故障转移调用方法（返回*models.Block）
func (es *EthereumScanner) callWithFailoverBlock(operation string, clientCall func(*ethclient.Client) (*models.Block, error)) (*models.Block, error) {
	// 首先尝试本地节点
	if es.localClient != nil {
		result, err := clientCall(es.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[ETH Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(es.externalClients) > 0 {
		// 从当前节点开始尝试
		startIndex := es.currentNodeIndex
		for i := 0; i < len(es.externalClients); i++ {
			currentIndex := (startIndex + i) % len(es.externalClients)
			client := es.externalClients[currentIndex]

			result, err := clientCall(client)
			if err == nil {
				// 成功获取，更新当前节点索引
				es.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[ETH Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverTransactions 通用的故障转移调用方法（返回[]map[string]interface{}）
func (es *EthereumScanner) callWithFailoverTransactions(operation string, clientCall func(*ethclient.Client) ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	// 首先尝试本地节点
	if es.localClient != nil {
		result, err := clientCall(es.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[ETH Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(es.externalClients) > 0 {
		// 从当前节点开始尝试
		startIndex := es.currentNodeIndex
		for i := 0; i < len(es.externalClients); i++ {
			currentIndex := (startIndex + i) % len(es.externalClients)
			client := es.externalClients[currentIndex]

			result, err := clientCall(client)
			if err == nil {
				// 成功获取，更新当前节点索引
				es.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[ETH Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverRawBlock 通用的故障转移调用方法（返回*types.Block）
func (es *EthereumScanner) callWithFailoverRawBlock(operation string, clientCall func(*ethclient.Client) (*types.Block, error)) (*types.Block, error) {
	// 首先尝试本地节点
	if es.localClient != nil {
		result, err := clientCall(es.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[ETH Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(es.externalClients) > 0 {
		// 从当前节点开始尝试
		startIndex := es.currentNodeIndex
		for i := 0; i < len(es.externalClients); i++ {
			currentIndex := (startIndex + i) % len(es.externalClients)
			client := es.externalClients[currentIndex]

			result, err := clientCall(client)
			if err == nil {
				// 成功获取，更新当前节点索引
				es.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[ETH Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// callWithFailoverReceipt 通用的故障转移调用方法（返回*types.Receipt）
func (es *EthereumScanner) callWithFailoverReceipt(operation string, clientCall func(*ethclient.Client) (*types.Receipt, error)) (*types.Receipt, error) {
	// 首先尝试本地节点
	if es.localClient != nil {
		result, err := clientCall(es.localClient)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[ETH Scanner] Local node failed for %s: %v, trying external APIs\n", operation, err)
	}

	// 如果本地节点失败或不存在，尝试外部节点
	if len(es.externalClients) > 0 {
		// 从当前节点开始尝试
		startIndex := es.currentNodeIndex
		for i := 0; i < len(es.externalClients); i++ {
			currentIndex := (startIndex + i) % len(es.externalClients)
			client := es.externalClients[currentIndex]

			result, err := clientCall(client)
			if err == nil {
				// 成功获取，更新当前节点索引
				es.currentNodeIndex = currentIndex
				return result, nil
			}

			fmt.Printf("[ETH Scanner] External API node %d failed for %s: %v\n", currentIndex, operation, err)
		}
	}

	return nil, fmt.Errorf("failed to %s: all nodes failed", operation)
}

// GetLatestBlockHeight 获取最新区块高度
func (es *EthereumScanner) GetLatestBlockHeight() (uint64, error) {
	result, err := es.callWithFailoverUint64("get latest block height", func(client *ethclient.Client) (uint64, error) {
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

	result, err := es.callWithFailoverBlock("get block by height", func(client *ethclient.Client) (*models.Block, error) {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(height)))
		if err != nil {
			return nil, err
		}
		return es.parseBlock(block), nil
	})

	if err == nil {
		fmt.Printf("[ETH Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
			result.Height, result.Hash[:16]+"...", result.TransactionCount)
	}
	return result, err
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
		Confirmations:    1, // 简化处理
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

		txData := map[string]interface{}{
			"hash":                 tx.Hash().Hex(),
			"nonce":                tx.Nonce(),
			"type":                 txType,
			"gasPrice":             gasPriceStr,
			"maxFeePerGas":         maxFeePerGas,
			"maxPriorityFeePerGas": maxPriorityFeePerGas,
			"effectiveGasPrice":    effectiveGasPrice,
			"gasLimit":             tx.Gas(), // 原始gas limit
			"gasUsed":              tx.Gas(), // 暂时使用gas limit，后续可以通过receipt获取实际值
			"to":                   toAddress,
			"value":                tx.Value().String(),
			"data":                 tx.Data(),
			"v":                    v.String(),
			"r":                    r.String(),
			"s":                    s.String(),
		}
		transactions[i] = txData
	}

	return transactions
}

// GetBlockTransactionsFromBlock 直接从区块中获取交易信息（避免哈希不一致问题）
func (es *EthereumScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	// 这里我们需要通过区块高度重新获取完整的区块数据
	// 因为 models.Block 中只包含基本信息，不包含完整的交易数据
	ethBlock, err := es.callWithFailoverRawBlock("get block by height for transactions", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(block.Height)))
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get block by height for transactions: %w", err)
	}

	// 直接从区块中提取交易信息
	return es.extractTransactionsFromBlock(ethBlock), nil
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
