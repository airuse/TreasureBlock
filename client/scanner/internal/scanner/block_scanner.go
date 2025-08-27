package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/models"
	"blockChainBrowser/client/scanner/internal/scanners"
	"blockChainBrowser/client/scanner/pkg"

	"github.com/sirupsen/logrus"
)

// BlockScanner 主扫块器
type BlockScanner struct {
	config       *config.Config
	scanners     map[string]Scanner
	stopChan     chan struct{}
	running      bool
	runningMutex sync.RWMutex
}

// Scanner 扫块器接口
type Scanner interface {
	GetLatestBlockHeight() (uint64, error)
	GetBlockByHeight(height uint64) (*models.Block, error)
	ValidateBlock(block *models.Block) error
	GetBlockTransactionsFromBlock(block *models.Block) ([]map[string]interface{}, error)
	CalculateBlockStats(block *models.Block, transactions []map[string]interface{})
}

// NewBlockScanner 创建新的主扫块器
func NewBlockScanner(cfg *config.Config) *BlockScanner {
	scanner := &BlockScanner{
		config:   cfg,
		scanners: make(map[string]Scanner),
		stopChan: make(chan struct{}),
	}

	// 初始化各种链的扫块器
	scanner.initializeScanners()

	return scanner
}

// initializeScanners 初始化各种链的扫块器
func (bs *BlockScanner) initializeScanners() {
	for chainName, chainConfig := range bs.config.Blockchain.Chains {
		if !chainConfig.Enabled || !chainConfig.Scan.Enabled {
			continue
		}

		var scanner Scanner
		switch chainName {
		case "btc":
			btcconfig := &chainConfig
			scanner = scanners.NewBitcoinScanner(btcconfig)
		case "eth":
			ethconfig := &chainConfig
			scanner = scanners.NewEthereumScanner(ethconfig)
		default:
			logrus.Warnf("Unsupported chain: %s", chainName)
			continue
		}

		bs.scanners[chainName] = scanner

		logrus.Infof("Initialized scanner for chain: %s", chainName)
	}
}

// Start 开始扫描
func (bs *BlockScanner) Start() error {
	bs.runningMutex.Lock()
	defer bs.runningMutex.Unlock()

	if bs.running {
		return fmt.Errorf("scanner is already running")
	}

	bs.running = true
	logrus.Info("Starting block scanner...")

	// 启动扫描协程
	go bs.scanLoop()

	return nil
}

// Stop 停止扫描
func (bs *BlockScanner) Stop() {
	bs.runningMutex.Lock()
	defer bs.runningMutex.Unlock()

	if !bs.running {
		return
	}

	logrus.Info("Stopping block scanner...")
	close(bs.stopChan)
	bs.running = false
}

// IsRunning 检查是否正在运行
func (bs *BlockScanner) IsRunning() bool {
	bs.runningMutex.RLock()
	defer bs.runningMutex.RUnlock()
	return bs.running
}

// scanLoop 扫描循环 - 启动所有链的持续扫描
func (bs *BlockScanner) scanLoop() {
	logrus.Info("Starting continuous scanning for all chains...")

	// 启动所有链的持续扫描（每个链都在自己的goroutine中持续运行）
	bs.scanAllChains()

	// 等待停止信号
	<-bs.stopChan
	logrus.Info("Block scanner stopped")
}

// scanAllChains 扫描所有链
func (bs *BlockScanner) scanAllChains() {
	var wg sync.WaitGroup

	for chainName, scanner := range bs.scanners {
		wg.Add(1)
		go func(chain string, s Scanner) {
			defer wg.Done()
			bs.scanChain(chain, s)
		}(chainName, scanner)
	}

	wg.Wait()
}

// scanChain 扫描指定链 - 持续扫描模式
func (bs *BlockScanner) scanChain(chainName string, scanner Scanner) {
	chainConfig := bs.getChainConfig(chainName)

	if chainConfig == nil {
		logrus.Errorf("[%s] Failed to get config for chain", chainName)
		return
	}

	logrus.Infof("[%s] Starting continuous scanning for chain %s", chainName, chainName)

	// 创建定时器
	ticker := time.NewTicker(chainConfig.Scan.Interval)
	defer ticker.Stop()

	// 持续扫描循环 - 使用定时器确保精确的扫描间隔
	for {
		select {
		case <-bs.stopChan:
			logrus.Infof("[%s] Stopped continuous scanning", chainName)
			return
		case <-ticker.C:
			// 每次循环都重新获取最新状态并扫描
			bs.processScanCycle(chainName, scanner, chainConfig)
		}
	}
}

// processScanCycle 处理单个扫描周期 - 每次循环都重新获取最新状态
func (bs *BlockScanner) processScanCycle(chainName string, scanner Scanner, chainConfig *config.ChainConfig) {
	// 1. 获取最新区块高度
	latestHeight, err := scanner.GetLatestBlockHeight()
	if err != nil {
		logrus.Errorf("[%s] Failed to get latest block height: %v", chainName, err)
		return
	}

	// 2. 计算确认后的安全高度
	safeHeight := latestHeight
	if latestHeight > uint64(chainConfig.Scan.Confirmations) {
		safeHeight = latestHeight - uint64(chainConfig.Scan.Confirmations)
	}

	// 3. 如果安全高度为0，说明还没有足够的确认，等待新区块
	if safeHeight == 0 {
		logrus.Debugf("[%s] Safe height is 0, waiting for more confirmations (latest: %d, confirmations: %d)",
			chainName, latestHeight, chainConfig.Scan.Confirmations)
		return
	}

	// 4. 获取最后一个验证通过的区块高度
	lastVerifiedHeight, err := bs.getLastVerifiedBlockHeight(chainName)

	if err != nil {
		logrus.Errorf("[%s] Failed to get last verified block height: %v", chainName, err)
		// 如果获取失败，使用配置的起始高度
		logrus.Infof("[%s] Using configured start height: %d", chainName, lastVerifiedHeight)
		return
	} else {
		logrus.Infof("[%s] Last verified height: %d", chainName, lastVerifiedHeight)
	}

	// 5. 判断是否需要扫描
	// 只有当安全高度 > 最后验证高度时，才需要扫描
	if safeHeight <= lastVerifiedHeight {
		return
	}

	// 6. 计算要扫描的区块高度
	scanHeight := lastVerifiedHeight + 1

	logrus.Infof("[%s] Scanning block at height %d (latest: %d, safe: %d, last verified: %d)", chainName, scanHeight, latestHeight, safeHeight, lastVerifiedHeight)

	// 7. 扫描单个区块
	select {
	case <-bs.stopChan:
		logrus.Infof("[%s] Stopped scanning due to stop signal", chainName)
		return // 如果收到停止信号，立即退出
	default:
		// 扫描单个区块
		bs.scanSingleBlock(chainName, scanner, scanHeight, chainConfig)
	}

	logrus.Infof("[%s] Completed scan cycle for height %d", chainName, scanHeight)
}

// getLastVerifiedBlockHeight 获取最后一个验证通过的区块高度
func (bs *BlockScanner) getLastVerifiedBlockHeight(chainName string) (uint64, error) {
	api := config.GetScannerAPI()
	if api == nil {
		return 0, fmt.Errorf("scanner API not initialized")
	}

	height, err := api.GetLastVerifiedBlockHeight(chainName)
	if err != nil {
		return 0, fmt.Errorf("failed to get last verified block height: %w", err)
	}

	return height, nil
}

// scanSingleBlock 扫描单个区块
func (bs *BlockScanner) scanSingleBlock(chainName string, scanner Scanner, height uint64, chainConfig *config.ChainConfig) {
	startTime := time.Now()

	block, err := scanner.GetBlockByHeight(height)
	if err != nil {
		logrus.Errorf("[%s] Failed to get block %d: %v", chainName, height, err)
		return
	}

	// 验证区块
	if err := scanner.ValidateBlock(block); err != nil {
		logrus.Errorf("[%s] Block validation failed for block %d: %v", chainName, height, err)
		return
	}

	// 提交区块到服务器，获取区块ID
	blockID, err := bs.submitBlockToServer(block)
	if err != nil {
		logrus.Errorf("[%s] Failed to submit block %d to server: %v", chainName, height, err)
		return
	}

	// 获取交易信息 - 直接从区块获取，避免哈希不一致问题
	transactions, err := scanner.GetBlockTransactionsFromBlock(block)
	if err != nil {
		logrus.Warnf("[%s] Failed to get transactions for block %d: %v", chainName, height, err)
	} else {
		scanner.CalculateBlockStats(block, transactions)

		// 上传交易信息到服务器，传入区块ID
		if err := bs.submitTransactionsToServer(chainName, block, transactions, blockID); err != nil {
			logrus.Warnf("[%s] Failed to submit transactions for block %d: %v", chainName, height, err)
			return
		}
	}

	// 验证区块
	if err := bs.verifyBlock(blockID); err != nil {
		logrus.Errorf("[%s] Block verification failed for block %d: %v", chainName, height, err)
		return
	}
	// 由于区块最早提交的时候没有 调用 CalculateBlockStats 方法给 block.BurnedEth、block.MinerTipEth、block.TotalAmount、block.Fee、block.Confirmations 等字段赋值
	// 所以需要在这里调用 更新这些字段 到后端API
	bs.updateBlockStatsToServer(block, transactions, blockID)

	// 保存到文件（如果启用）
	if chainConfig.Scan.SaveToFile {
		if err := bs.saveBlockToFile(block, chainConfig.Scan.OutputDir); err != nil {
			logrus.Warnf("[%s] Failed to save block %d to file: %v", chainName, height, err)
		}
	}

	processTime := time.Since(startTime).Milliseconds()
	logrus.Infof("[%s] Successfully processed and verified block %d (hash: %s, %d tx, %dms)",
		chainName, height, block.Hash[:16]+"...", block.TransactionCount, processTime)
}

// updateBlockStatsToServer 计算并将区块统计字段更新到后端
func (bs *BlockScanner) updateBlockStatsToServer(block *models.Block, transactions []map[string]interface{}, blockID uint64) {
	api := config.GetScannerAPI()
	if api == nil {
		logrus.Warn("scanner API not initialized; skip UpdateBlockStats")
		return
	}

	// 组装需要更新的字段（与后端 UpdateBlockRequest 对齐，使用可选字段键名）
	payload := map[string]interface{}{}

	// 这些值应由各链的 CalculateBlockStats 写入到 block 上，这里只做透传
	if block.TotalAmount > 0 {
		payload["total_amount"] = block.TotalAmount
	}
	if block.Fee > 0 {
		payload["fee"] = block.Fee
	}
	if block.Confirmations > 0 {
		payload["confirmations"] = block.Confirmations
	}

	// ETH London 相关字段（字符串）
	if block.BurnedEth != nil {
		payload["burned_eth"] = block.BurnedEth.Text('f', 18)
	}
	if block.MinerTipEth != nil {
		payload["miner_tip_eth"] = block.MinerTipEth.Text('f', 18)
	}
	if block.BaseFee != nil {
		payload["base_fee"] = block.BaseFee.String()
	}

	// 可选：大小、交易数（若统计有修正）
	if block.Size > 0 {
		payload["size"] = block.Size
	}
	if block.TransactionCount > 0 {
		payload["transaction_count"] = block.TransactionCount
	}

	if len(payload) == 0 {
		return
	}

	if err := api.UpdateBlockStats(block.Hash, payload); err != nil {
		logrus.Warnf("Failed to update block stats (hash=%s): %v", block.Hash, err)
	}
}

// submitBlockToServer 提交区块到服务器，返回区块ID
func (bs *BlockScanner) submitBlockToServer(block *models.Block) (uint64, error) {
	// 获取API实例
	api := config.GetScannerAPI()
	if api == nil {
		return 0, fmt.Errorf("scanner API not initialized")
	}

	// 构建区块上传请求
	blockRequest := &pkg.BlockUploadRequest{
		Hash:             block.Hash,
		Height:           block.Height,
		PreviousHash:     block.PreviousHash,
		Timestamp:        block.Timestamp,
		Size:             block.Size,
		TransactionCount: block.TransactionCount,
		TotalAmount:      block.TotalAmount,
		Fee:              block.Fee,
		Confirmations:    block.Confirmations,
		IsOrphan:         block.IsOrphan,
		Chain:            block.Chain, // BTC特有
		MerkleRoot:       block.MerkleRoot,
		Bits:             block.Bits,
		Version:          uint32(block.Version),
		Weight:           block.Weight,       // ETH特有
		GasLimit:         block.Weight,       // 以太坊解析中Weight映射自GasLimit
		GasUsed:          block.StrippedSize, // 以太坊解析中StrippedSize映射自GasUsed
		Miner:            block.Miner,
		ParentHash:       block.PreviousHash, // 近似映射
		Nonce:            fmt.Sprintf("%d", block.Nonce),
		Difficulty:       fmt.Sprintf("%f", block.Difficulty),
		BaseFee: func() string {
			if block.BaseFee != nil {
				return block.BaseFee.String()
			}
			return ""
		}(),
		BurnedEth: func() string {
			if block.BurnedEth != nil {
				return block.BurnedEth.Text('f', 18)
			}
			return ""
		}(),
		MinerTipEth: func() string {
			if block.MinerTipEth != nil {
				return block.MinerTipEth.Text('f', 18)
			}
			return ""
		}(),
		// ETH状态根字段
		StateRoot:        block.StateRoot,
		TransactionsRoot: block.TransactionsRoot,
		ReceiptsRoot:     block.ReceiptsRoot,
	}

	// 使用API上传区块
	response, err := api.UploadBlock(blockRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to upload block via API: %w", err)
	}

	// 从响应中提取区块ID
	if response != nil {
		return uint64(response.ID), nil
	}

	return 0, fmt.Errorf("failed to get block ID from response")
}

// submitTransactionsToServer 将交易信息提交到服务器
func (bs *BlockScanner) submitTransactionsToServer(chainName string, block *models.Block, transactions []map[string]interface{}, blockID uint64) error {
	// 获取API实例
	api := config.GetScannerAPI()
	if api == nil {
		return fmt.Errorf("scanner API instance not available")
	}

	if len(transactions) == 0 {
		logrus.Infof("[%s] No transactions to upload for block %d", chainName, block.Height)
		return nil
	}

	// 获取链配置
	chainConfig := bs.getChainConfig(chainName)
	if chainConfig == nil {
		return fmt.Errorf("failed to get chain config for %s", chainName)
	}

	// 根据配置选择上传方式
	if chainConfig.Scan.BatchUpload {
		return bs.submitTransactionsBatch(chainName, block, transactions, blockID, chainConfig, api)
	} else {
		return bs.submitTransactionsIndividually(chainName, block, transactions, blockID, chainConfig, api)
	}
}

// submitTransactionsBatch 批量上传交易
func (bs *BlockScanner) submitTransactionsBatch(chainName string, block *models.Block, transactions []map[string]interface{}, blockID uint64, chainConfig *config.ChainConfig, api *pkg.ScannerAPI) error {
	// 构建批量交易请求数据
	var batchTransactions []map[string]interface{}

	for _, tx := range transactions {
		txRequest := bs.buildTransactionRequest(chainName, block, tx, blockID)
		batchTransactions = append(batchTransactions, txRequest)
	}

	// 使用批量上传API
	logrus.Infof("[%s] Uploading %d transactions in batch for block %d", chainName, len(batchTransactions), block.Height)

	if err := api.UploadTransactionsBatch(batchTransactions); err != nil {
		// 检查是否为致命错误，如果是就直接退出程序
		if strings.Contains(strings.ToUpper(err.Error()), "BLOCK_TRANSACTION_FAILED") {
			pkg.HandleFatalError(err, "交易批量上传失败")
		}

		return fmt.Errorf("batch upload transactions failed: %w", err)
	}

	logrus.Infof("[%s] Successfully uploaded %d transactions in batch for block %d", chainName, len(batchTransactions), block.Height)
	return nil
}

// submitTransactionsIndividually 单个上传交易（保持原有逻辑作为备选）
func (bs *BlockScanner) submitTransactionsIndividually(chainName string, block *models.Block, transactions []map[string]interface{}, blockID uint64, chainConfig *config.ChainConfig, api *pkg.ScannerAPI) error {
	// 并发上限：使用链配置的 MaxConcurrent，缺省为 10
	concurrency := 10
	if chainConfig.Scan.MaxConcurrent > 0 {
		concurrency = chainConfig.Scan.MaxConcurrent
	}

	// 并发上传交易，全部完成后再返回
	var wg sync.WaitGroup
	errCh := make(chan error, len(transactions))
	sem := make(chan struct{}, concurrency)

	for i := range transactions {
		// 避免闭包捕获同一变量，按索引取当前交易
		tx := transactions[i]

		// 在每个 goroutine 中构建请求并上传
		wg.Add(1)
		go func(tx map[string]interface{}) {
			defer wg.Done()
			// 限流：获取并发令牌
			sem <- struct{}{}
			defer func() { <-sem }()

			// 构建交易请求数据
			txRequest := bs.buildTransactionRequest(chainName, block, tx, blockID)

			// 上传交易
			if err := api.UploadTransaction(txRequest); err != nil {
				// 检查是否为致命错误，如果是就直接退出程序
				if strings.Contains(strings.ToUpper(err.Error()), "BLOCK_TRANSACTION_FAILED") {
					pkg.HandleFatalError(err, "交易上传失败")
				}

				logrus.Warnf("[%s] Failed to upload transaction %v: %v", chainName, tx["hash"], err)
				// 收集非致命错误，等待全部完成后再统一返回
				errCh <- err
				return
			}
		}(tx)
	}

	// 等待全部上传完成
	wg.Wait()
	close(errCh)

	// 汇总错误
	failed := 0
	for range errCh {
		failed++
	}
	if failed > 0 {
		return fmt.Errorf("%d transactions failed to upload", failed)
	}

	logrus.Infof("[%s] Successfully uploaded %d transactions individually for block %d", chainName, len(transactions), block.Height)
	return nil
}

// buildTransactionRequest 构建交易请求数据
func (bs *BlockScanner) buildTransactionRequest(chainName string, block *models.Block, tx map[string]interface{}, blockID uint64) map[string]interface{} {
	// 提取/推导地址
	fromAddress := ""
	toAddress := ""
	if v, ok := tx["from"].(string); ok {
		fromAddress = v
	}
	if v, ok := tx["to"].(string); ok {
		toAddress = v
	}

	// 针对BTC从vin/vout推导
	if chainName == "btc" {
		if fromAddress == "" {
			if vin, ok := tx["vin"].([]interface{}); ok && len(vin) > 0 {
				if in0, ok := vin[0].(map[string]interface{}); ok {
					if prevout, ok := in0["prevout"].(map[string]interface{}); ok {
						if spk, ok := prevout["scriptPubKey"].(map[string]interface{}); ok {
							if addrs, ok := spk["addresses"].([]interface{}); ok && len(addrs) > 0 {
								if a0, ok := addrs[0].(string); ok {
									fromAddress = a0
								}
							}
						}
					}
				}
			}
		}
		if toAddress == "" {
			if vout, ok := tx["vout"].([]interface{}); ok && len(vout) > 0 {
				if out0, ok := vout[0].(map[string]interface{}); ok {
					if spk, ok := out0["scriptPubKey"].(map[string]interface{}); ok {
						if addrs, ok := spk["addresses"].([]interface{}); ok && len(addrs) > 0 {
							if a0, ok := addrs[0].(string); ok {
								toAddress = a0
							}
						}
					}
				}
			}
		}
	}

	// 构建交易请求数据
	txRequest := map[string]interface{}{
		"tx_id": tx["hash"],
		"tx_type": func() uint8 {
			if txType, ok := tx["type"].(uint8); ok {
				return txType
			}
			return 0
		}(),
		"confirm":      1,
		"status":       1,
		"send_status":  1,
		"balance":      "0",
		"amount":       tx["value"],
		"trans_id":     0,
		"chain":        chainName,
		"symbol":       chainName,
		"address_from": fromAddress,
		"address_to":   toAddress,
		"gas_limit":    tx["gasLimit"],
		"gas_price":    tx["gasPrice"],
		"gas_used":     tx["gasUsed"],
		"fee":          "0",
		"used_fee":     nil,
		"height":       block.Height,
		"block_id":     blockID,
		"contract_addr": func() string {
			if contractAddr, ok := tx["contract_address"].(string); ok {
				return contractAddr
			}
			return ""
		}(),
		"hex":         tx["data"],
		"tx_scene":    "blockchain_scan",
		"remark":      "Scanned from blockchain",
		"block_index": tx["block_index"],
		"nonce":       tx["nonce"],
		"logs":        tx["logs"],
		"receipt":     tx["receipt"],
	}

	// 添加EIP-1559相关字段（如果存在）
	if maxFeePerGas, ok := tx["maxFeePerGas"]; ok {
		txRequest["max_fee_per_gas"] = maxFeePerGas
	}
	if maxPriorityFeePerGas, ok := tx["maxPriorityFeePerGas"]; ok {
		txRequest["max_priority_fee_per_gas"] = maxPriorityFeePerGas
	}
	if effectiveGasPrice, ok := tx["effectiveGasPrice"]; ok {
		txRequest["effective_gas_price"] = effectiveGasPrice
	}

	return txRequest
}

// verifyBlock 验证区块
func (bs *BlockScanner) verifyBlock(blockID uint64) error {
	api := config.GetScannerAPI()
	if api == nil {
		return fmt.Errorf("scanner API not initialized")
	}

	if err := api.VerifyBlock(blockID); err != nil {
		// 检查是否为致命错误，如果是就直接退出程序
		if strings.Contains(strings.ToUpper(err.Error()), "BLOCK_VERIFICATION_FAILED") {
			pkg.HandleFatalError(err, "区块验证失败")
		}

		return fmt.Errorf("block verification failed: %w", err)
	}

	return nil
}

// saveBlockToFile 保存区块到文件
func (bs *BlockScanner) saveBlockToFile(block *models.Block, outputDir string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 创建链目录
	chainDir := filepath.Join(outputDir, block.Chain)
	if err := os.MkdirAll(chainDir, 0755); err != nil {
		return fmt.Errorf("failed to create chain directory: %w", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("block_%d_%s.json", block.Height, block.Hash[:8])
	filepath := filepath.Join(chainDir, filename)

	// 序列化区块数据
	jsonData, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// getChainConfig 获取指定链的配置
func (bs *BlockScanner) getChainConfig(chainName string) *config.ChainConfig {
	if chainConfig, exists := bs.config.Blockchain.Chains[chainName]; exists {
		return &chainConfig
	}
	return nil
}
