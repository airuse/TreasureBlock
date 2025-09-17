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
	"github.com/ethereum/go-ethereum/rpc"
)

// EVMScanner EVM兼容链扫块器基类 - 供ETH、BSC等EVM兼容链共享
type EVMScanner struct {
	config    *config.ChainConfig
	chainName string
	chainID   *big.Int
	// 客户端连接池
	localClient      *ethclient.Client
	externalClients  []*ethclient.Client
	currentNodeIndex int // 当前使用的外部节点索引
}

// NewEVMSanner 创建新的EVM扫块器
func NewEVMSanner(cfg *config.ChainConfig, chainName string) *EVMScanner {
	scanner := &EVMScanner{
		config:           cfg,
		chainName:        chainName,
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
				fmt.Printf("[%s Scanner] Successfully connected to external API: %s\n", strings.ToUpper(chainName), apiURL)
			} else {
				fmt.Printf("[%s Scanner] Warning: Failed to connect to external API %s: %v\n", strings.ToUpper(chainName), apiURL, err)
			}
		}
	}

	// 获取网络链ID
	failoverManager := failover.NewFailoverManager(scanner.localClient, scanner.externalClients)
	chainID, err := failoverManager.CallWithFailoverNetworkID("get network id", func(client *ethclient.Client) (*big.Int, error) {
		return client.NetworkID(context.Background())
	})
	if err != nil {
		fmt.Printf("[%s Scanner] Warning: Failed to detect chain ID: %v\n", strings.ToUpper(chainName), err)
		// 使用配置中的chain_id作为回退
		chainID = big.NewInt(int64(cfg.ChainID))
	}
	scanner.chainID = chainID
	return scanner
}

// GetLatestBlockHeight 获取最新区块高度
func (es *EVMScanner) GetLatestBlockHeight() (uint64, error) {
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	result, err := failoverManager.CallWithFailoverUint64("get latest block height", func(client *ethclient.Client) (uint64, error) {
		return client.BlockNumber(context.Background())
	})

	if err == nil {
		fmt.Printf("[%s Scanner] Latest block height: %d\n", strings.ToUpper(es.chainName), result)
	}
	return result, err
}

// GetBlockByHeight 根据高度获取区块
func (es *EVMScanner) GetBlockByHeight(height uint64) (*models.Block, error) {
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	result, err := failoverManager.CallWithFailoverRawBlock("get block by height", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(height)))
	})

	if err != nil {
		return nil, err
	}
	if err := es.ValidateEVMBlock(result); err != nil {
		return nil, err
	}
	// 解析区块数据
	block := es.parseBlock(result)

	fmt.Printf("[%s Scanner] Successfully scanned block %d (hash: %s) with %d transactions\n",
		strings.ToUpper(es.chainName), block.Height, block.Hash[:16]+"...", block.TransactionCount)
	return block, nil
}

// ValidateBlock 验证区块
func (es *EVMScanner) ValidateBlock(block *models.Block) error {
	// 基本验证
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	if block.Hash == "" {
		return fmt.Errorf("block hash is empty")
	}
	if block.Height == 0 {
		return fmt.Errorf("block height is 0")
	}
	return nil
}

// ValidateEVMBlock 验证EVM区块
func (es *EVMScanner) ValidateEVMBlock(block *types.Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	// 按 ETH 实现校验
	if block.Header() == nil {
		return fmt.Errorf("block header is nil")
	}
	if block.Hash() == (common.Hash{}) {
		return fmt.Errorf("block hash is empty")
	}
	if block.NumberU64() == 0 {
		return fmt.Errorf("block height is zero")
	}
	if block.Time() == 0 {
		return fmt.Errorf("block timestamp is zero")
	}
	return nil
}

// GetBlockTransactionsFromBlock 从区块获取交易信息
func (es *EVMScanner) GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error) {
	// 与 ETH 一致：获取区块 -> 提取交易 -> 并发补全回执
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	evnBlock, err := failoverManager.CallWithFailoverRawBlock("get block by height for transactions", func(client *ethclient.Client) (*types.Block, error) {
		return client.BlockByNumber(context.Background(), big.NewInt(int64(block.Height)))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height for transactions: %w", err)
	}

	transactions := es.extractTransactionsFromBlock(evnBlock)
	if err := es.enrichTransactionsWithContractInfo(block, transactions); err != nil {
		fmt.Printf("[%s Scanner] Warning: Failed to enrich transactions with contract info: %v\n", strings.ToUpper(es.chainName), err)
	}
	return transactions, nil
}

// CalculateBlockStats 计算区块统计信息
func (es *EVMScanner) CalculateBlockStats(block *models.Block, transactions []map[string]interface{}) {
	var totalGasUsed uint64
	totalFee := big.NewInt(0)
	totalValue := big.NewInt(0)
	legacyTxCount := 0
	eip1559TxCount := 0

	for _, tx := range transactions {
		if gasUsed, ok := tx["gasUsed"].(uint); ok {
			totalGasUsed += uint64(gasUsed)
		} else if gasLimit, ok := tx["gasLimit"].(uint); ok {
			totalGasUsed += uint64(gasLimit)
			fmt.Printf("[%s Scanner] Warning: Using gasLimit %d instead of gasUsed for tx\n", strings.ToUpper(es.chainName), gasLimit)
		}
		if txType, _ := tx["type"].(uint8); txType == 2 {
			eip1559TxCount++
		} else {
			legacyTxCount++
		}
		if valueStr, ok := tx["value"].(string); ok {
			if v, ok := new(big.Int).SetString(valueStr, 10); ok {
				totalValue.Add(totalValue, v)
			}
		}
	}

	actualGasUsed := block.StrippedSize
	if totalGasUsed != actualGasUsed {
		fmt.Printf("[%s Scanner] Warning: Block %d gas used mismatch: calculated=%d, actual=%d\n", strings.ToUpper(es.chainName), block.Height, totalGasUsed, actualGasUsed)
	}

	totalFee = big.NewInt(0)
	for _, tx := range transactions {
		if receipt, ok := tx["receipt"].(*types.Receipt); ok && receipt != nil {
			gasUsed := uint(receipt.GasUsed)
			var gasPrice *big.Int
			if receipt.EffectiveGasPrice != nil {
				gasPrice = receipt.EffectiveGasPrice
			} else {
				if txType, _ := tx["type"].(uint8); txType == 2 {
					if effectiveGasPriceStr, ok := tx["effectiveGasPrice"].(string); ok && effectiveGasPriceStr != "" {
						if p, ok := new(big.Int).SetString(effectiveGasPriceStr, 10); ok {
							gasPrice = p
						}
					}
				} else {
					if gasPriceStr, ok := tx["gasPrice"].(string); ok {
						if p, ok := new(big.Int).SetString(gasPriceStr, 10); ok {
							gasPrice = p
						}
					}
				}
			}
			if gasPrice != nil {
				txFee := new(big.Int).Mul(new(big.Int).SetUint64(uint64(gasUsed)), gasPrice)
				totalFee.Add(totalFee, txFee)
			}
		}
	}

	if block.BaseFee != nil && block.BaseFee.Sign() > 0 {
		burnedWei := new(big.Int).Mul(new(big.Int).SetUint64(actualGasUsed), block.BaseFee)
		minerTipWei := new(big.Int).Sub(totalFee, burnedWei)
		if minerTipWei.Sign() < 0 {
			minerTipWei.SetInt64(0)
			fmt.Printf("[%s Scanner] Warning: Block %d miner tip is negative, setting to 0\n", strings.ToUpper(es.chainName), block.Height)
		}
		block.BurnedEth = new(big.Float).Quo(new(big.Float).SetInt(burnedWei), new(big.Float).SetInt(big.NewInt(1e18)))
		block.MinerTipEth = new(big.Float).Quo(new(big.Float).SetInt(minerTipWei), new(big.Float).SetInt(big.NewInt(1e18)))
	} else {
		block.BurnedEth = new(big.Float).SetInt(big.NewInt(0))
		block.MinerTipEth = new(big.Float).Quo(new(big.Float).SetInt(totalFee), new(big.Float).SetInt(big.NewInt(1e18)))
	}

	ethFee := new(big.Float).Quo(new(big.Float).SetInt(totalFee), new(big.Float).SetInt(big.NewInt(1e18)))
	ethValue := new(big.Float).Quo(new(big.Float).SetInt(totalValue), new(big.Float).SetInt(big.NewInt(1e18)))

	block.TotalAmount, _ = ethValue.Float64()
	block.Fee, _ = ethFee.Float64()
	block.TransactionCount = len(transactions)
	block.Confirmations = 1

	fmt.Printf("[%s Scanner] Block %d: Legacy TXs=%d, EIP-1559 TXs=%d, Total Value=%s ETH, Total Fee=%s ETH\n",
		strings.ToUpper(es.chainName), block.Height, legacyTxCount, eip1559TxCount, ethValue.Text('f', 18), ethFee.Text('f', 18))
}

// extractTransactionsFromBlock 直接从区块中提取交易信息（与 ETH 逻辑一致）
func (es *EVMScanner) extractTransactionsFromBlock(block *types.Block) []map[string]interface{} {
	transactions := make([]map[string]interface{}, len(block.Transactions()))

	for i, tx := range block.Transactions() {
		v, r, s := tx.RawSignatureValues()

		var gasPriceStr, maxFeePerGas, maxPriorityFeePerGas, effectiveGasPrice string
		var txType uint8

		if tx.Type() == 2 { // EIP-1559
			txType = 2
			feeCap := tx.GasFeeCap()
			tipCap := tx.GasTipCap()
			maxFeePerGas = feeCap.String()
			maxPriorityFeePerGas = tipCap.String()

			var effective *big.Int
			if block.BaseFee() != nil && block.BaseFee().Sign() > 0 {
				basePlusTip := new(big.Int).Add(block.BaseFee(), tipCap)
				if basePlusTip.Cmp(feeCap) < 0 {
					effective = basePlusTip
				} else {
					effective = feeCap
				}
			} else {
				if es.isL2Chain() {
					effective = tipCap
					fmt.Printf("[%s Scanner] L2 chain detected, using tipCap as effective gas price: %s\n", strings.ToUpper(es.chainName), effective.String())
				} else {
					effective = feeCap
				}
			}
			effectiveGasPrice = effective.String()
			gasPriceStr = effectiveGasPrice
		} else {
			txType = 0
			gasPriceStr = tx.GasPrice().String()
			maxFeePerGas = "0"
			maxPriorityFeePerGas = "0"
			effectiveGasPrice = gasPriceStr
		}

		var toAddress string
		if tx.To() != nil {
			toAddress = tx.To().Hex()
		} else {
			toAddress = ""
		}

		var fromAddress string
		var signer types.Signer
		if es.chainID != nil && es.chainID.Sign() > 0 {
			signer = types.LatestSignerForChainID(es.chainID)
		} else {
			signer = types.HomesteadSigner{}
		}
		if sender, err := types.Sender(signer, tx); err == nil {
			fromAddress = sender.Hex()
		} else {
			fmt.Printf("[%s Scanner] Warning: Failed to recover sender for tx %s: %v\n", strings.ToUpper(es.chainName), tx.Hash().Hex(), err)
			fromAddress = ""
		}

		txData := map[string]interface{}{
			"hash":                 tx.Hash().Hex(),
			"nonce":                tx.Nonce(),
			"type":                 txType,
			"from":                 fromAddress,
			"to":                   toAddress,
			"value":                tx.Value().String(),
			"gasPrice":             gasPriceStr,
			"maxFeePerGas":         maxFeePerGas,
			"maxPriorityFeePerGas": maxPriorityFeePerGas,
			"effectiveGasPrice":    effectiveGasPrice,
			"gasLimit":             uint(tx.Gas()),
			"gasUsed":              uint(tx.Gas()), // 将由回执覆盖
			"data":                 fmt.Sprintf("0x%x", tx.Data()),
			"raw_data":             tx.Data(),
			"v":                    v.String(),
			"r":                    r.String(),
			"s":                    s.String(),
			"block_index":          uint(i),
		}

		if toAddress != "" && es.isConfiguredTokenAddress(toAddress) && len(tx.Data()) > 0 {
			txData["is_contract_tx"] = true
			txData["contract_address"] = toAddress
		} else {
			txData["is_contract_tx"] = false
		}

		transactions[i] = txData
	}

	return transactions
}

// isConfiguredTokenAddress 检查地址是否为配置的币种地址
func (es *EVMScanner) isConfiguredTokenAddress(address string) bool {
	if address == "" {
		return false
	}
	for _, tokenAddr := range es.config.TokenAddresses {
		if strings.EqualFold(address, tokenAddr) {
			return true
		}
	}
	return false
}

// enrichTransactionsWithContractInfo 获取所有交易回执（并发处理）
func (es *EVMScanner) enrichTransactionsWithContractInfo(block *models.Block, transactions []map[string]interface{}) error {
	if len(transactions) == 0 {
		return nil
	}
	var txHashes []string
	for _, tx := range transactions {
		if hash, ok := tx["hash"].(string); ok {
			txHashes = append(txHashes, hash)
		}
	}
	if len(txHashes) > 0 {
		if err := es.batchGetTransactionReceipts(block, transactions, txHashes); err != nil {
			fmt.Printf("[%s Scanner] Warning: Failed to batch get transaction receipts: %v\n", strings.ToUpper(es.chainName), err)
		}
	}
	return nil
}

// batchGetTransactionReceipts 高效并发获取所有交易回执
func (es *EVMScanner) batchGetTransactionReceipts(block *models.Block, transactions []map[string]interface{}, txHashes []string) error {
	if len(txHashes) == 0 {
		return nil
	}
	if receipts, err := es.tryBlockReceipts(block.Height); err == nil {
		return es.processBlockReceipts(block, transactions, receipts)
	}
	return es.fallbackToIndividualReceipts(block, transactions, txHashes)
}

// tryBlockReceipts 尝试使用 BlockReceipts 获取整个区块的回执
func (es *EVMScanner) tryBlockReceipts(blockHeight uint64) ([]*types.Receipt, error) {
	startTime := time.Now()
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	receipts, err := failoverManager.CallWithFailoverReceipts("get block receipts", func(client *ethclient.Client) ([]*types.Receipt, error) {
		blockNum := rpc.BlockNumber(blockHeight)
		return client.BlockReceipts(context.Background(), rpc.BlockNumberOrHash{BlockNumber: &blockNum})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get block receipts: %w", err)
	}
	elapsed := time.Since(startTime)
	stats := failoverManager.GetStats()
	fmt.Printf("[%s Scanner] %d 📊 BlockReceipts Fetch Complete: total=%d, time=%v, stats=%+v\n",
		strings.ToUpper(es.chainName), blockHeight, len(receipts), elapsed, stats)
	return receipts, nil
}

// processBlockReceipts 处理 BlockReceipts 的回执
func (es *EVMScanner) processBlockReceipts(block *models.Block, transactions []map[string]interface{}, receipts []*types.Receipt) error {
	if len(receipts) == 0 {
		fmt.Printf("[%s Scanner] Warning: BlockReceipts returned empty receipts for block %d\n", strings.ToUpper(es.chainName), block.Height)
		return nil
	}
	hashToTxMap := make(map[string]int)
	for i, tx := range transactions {
		if hash, ok := tx["hash"].(string); ok {
			hashToTxMap[hash] = i
		}
	}
	for _, receipt := range receipts {
		if receipt == nil {
			continue
		}
		txHash := receipt.TxHash.Hex()
		if index, exists := hashToTxMap[txHash]; exists && index < len(transactions) {
			tx := transactions[index]
			if receipt.Status == 1 {
				tx["status"] = "success"
			} else {
				tx["status"] = "failed"
			}
			tx["gasUsed"] = uint(receipt.GasUsed)
			if receipt.EffectiveGasPrice != nil {
				tx["effectiveGasPrice"] = receipt.EffectiveGasPrice.String()
				tx["gasPrice"] = receipt.EffectiveGasPrice.String()
				realFee := new(big.Int).Mul(receipt.EffectiveGasPrice, new(big.Int).SetUint64(uint64(receipt.GasUsed)))
				tx["realFee"] = realFee.String()
			}
			es.parseContractLogs(tx, receipt)
			tx["receipt"] = receipt
		}
	}
	return nil
}

// fallbackToIndividualReceipts 回退到逐个获取 TransactionReceipt
func (es *EVMScanner) fallbackToIndividualReceipts(block *models.Block, transactions []map[string]interface{}, txHashes []string) error {
	startTime := time.Now()
	hashToTxMap := make(map[string]int)
	for i, tx := range transactions {
		if hash, ok := tx["hash"].(string); ok {
			hashToTxMap[hash] = i
		}
	}
	type receiptResult struct {
		hash    string
		receipt *types.Receipt
		err     error
		index   int
	}
	maxConcurrency := es.config.Scan.MaxConcurrent
	if maxConcurrency <= 0 {
		maxConcurrency = 20
	}
	fmt.Printf("[%s Scanner] 🔄 Using %d concurrent workers for individual receipt fetching\n", strings.ToUpper(es.chainName), maxConcurrency)
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(chan receiptResult, len(txHashes))
	failoverManager := failover.NewFailoverManager(es.localClient, es.externalClients)
	for i, txHash := range txHashes {
		go func(hash string, idx int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			var receipt *types.Receipt
			err := failoverManager.CallWithFailover("get transaction receipt", func(client *ethclient.Client) error {
				var receiptErr error
				receipt, receiptErr = client.TransactionReceipt(context.Background(), common.HexToHash(hash))
				return receiptErr
			})
			results <- receiptResult{hash: hash, receipt: receipt, err: err, index: idx}
		}(txHash, i)
	}
	successCount := 0
	failureCount := 0
	for i := 0; i < len(txHashes); i++ {
		result := <-results
		if result.err != nil {
			fmt.Printf("[%s Scanner] ❌ Failed to get receipt for tx %s: %v\n", strings.ToUpper(es.chainName), result.hash, result.err)
			failureCount++
			continue
		}
		if index, exists := hashToTxMap[result.hash]; exists && index < len(transactions) {
			tx := transactions[index]
			if result.receipt.Status == 1 {
				tx["status"] = "success"
			} else {
				tx["status"] = "failed"
			}
			tx["gasUsed"] = uint(result.receipt.GasUsed)
			if result.receipt.EffectiveGasPrice != nil {
				tx["effectiveGasPrice"] = result.receipt.EffectiveGasPrice.String()
				tx["gasPrice"] = result.receipt.EffectiveGasPrice.String()
				realFee := new(big.Int).Mul(result.receipt.EffectiveGasPrice, new(big.Int).SetUint64(uint64(result.receipt.GasUsed)))
				tx["realFee"] = realFee.String()
			}
			es.parseContractLogs(tx, result.receipt)
			tx["receipt"] = result.receipt
			successCount++
		}
	}
	elapsed := time.Since(startTime)
	stats := failoverManager.GetStats()
	fmt.Printf("[%s Scanner] %d 📊 TransactionReceipt Fetch Complete: success=%d failed=%d time=%v workers=%d stats=%+v\n",
		strings.ToUpper(es.chainName), block.Height, successCount, failureCount, elapsed, es.config.Scan.MaxConcurrent, stats)
	return nil
}

// parseContractLogs 保存合约交易的原始日志数据
func (es *EVMScanner) parseContractLogs(tx map[string]interface{}, receipt *types.Receipt) {
	if receipt == nil {
		return
	}
	if receipt.ContractAddress != (common.Address{}) {
		tx["contract_address"] = receipt.ContractAddress.Hex()
	} else if len(receipt.Logs) > 0 {
		tx["contract_address"] = receipt.Logs[0].Address.Hex()
	}
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
			for j, topic := range log.Topics {
				logData["topics"].([]string)[j] = topic.Hex()
			}
			logs = append(logs, logData)
		}
		tx["logs"] = logs
		tx["log_count"] = len(logs)
	}
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
}

// parseBlock 解析区块数据（与 ETH 实现一致）
func (es *EVMScanner) parseBlock(block *types.Block) *models.Block {
	// 获取矿工地址
	miner := ""
	if block.Coinbase() != (common.Address{}) {
		miner = block.Coinbase().Hex()
	}
	parsed := &models.Block{
		Chain:            es.chainName,
		ChainID:          es.config.ChainID,
		Hash:             block.Hash().Hex(),
		Height:           block.NumberU64(),
		Version:          0,
		Timestamp:        time.Unix(int64(block.Time()), 0),
		Size:             uint64(block.Size()),
		Weight:           block.GasLimit(),
		StrippedSize:     block.GasUsed(),
		TransactionCount: len(block.Transactions()),
		Difficulty:       float64(block.Difficulty().Uint64()),
		Nonce:            block.Nonce(),
		PreviousHash:     block.ParentHash().Hex(),
		MerkleRoot:       block.Root().Hex(),
		Confirmations:    1,
		Miner:            miner,
		BaseFee:          block.BaseFee(),
		StateRoot:        block.Root().Hex(),
		TransactionsRoot: block.TxHash().Hex(),
		ReceiptsRoot:     block.ReceiptHash().Hex(),
	}
	return parsed
}

// isL2Chain 检查是否为常见 L2 链
func (es *EVMScanner) isL2Chain() bool {
	if es.chainID == nil {
		return false
	}
	l2 := map[int64]struct{}{
		137:   {}, // Polygon
		80001: {},
		42161: {}, // Arbitrum One
		42170: {}, // Arbitrum Nova
		10:    {}, // Optimism
		8453:  {}, // Base
		1101:  {}, // Polygon zkEVM
		324:   {}, // zkSync Era
	}
	_, ok := l2[es.chainID.Int64()]
	return ok
}
