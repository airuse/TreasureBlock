package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	// 保存一下上次扫描的高度，避免重复扫描
	var lastScanHeight uint64
	var heightMutex sync.RWMutex

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
			// 使用协程处理单个扫描周期，避免阻塞定时器
			go bs.processScanCycle(chainName, scanner, chainConfig, &lastScanHeight, &heightMutex)
		}
	}
}

// processScanCycle 处理单个扫描周期
func (bs *BlockScanner) processScanCycle(chainName string, scanner Scanner, chainConfig *config.ChainConfig, lastScanHeight *uint64, heightMutex *sync.RWMutex) {
	// 获取最新区块高度
	latestHeight, err := scanner.GetLatestBlockHeight()
	if err != nil {
		logrus.Errorf("[%s] Failed to get latest block height: %v", chainName, err)
		return
	}

	// 计算确认后的安全高度
	safeHeight := latestHeight
	if latestHeight > uint64(chainConfig.Scan.Confirmations) {
		safeHeight = latestHeight - uint64(chainConfig.Scan.Confirmations)
	}

	// 线程安全地检查是否需要扫描
	heightMutex.RLock()
	lastHeight := *lastScanHeight
	heightMutex.RUnlock()

	if lastHeight > 0 && lastHeight >= safeHeight {
		return
	}

	// 如果安全高度为0，说明还没有足够的确认，等待新区块
	if safeHeight == 0 {
		logrus.Debugf("[%s] Safe height is 0, waiting for more confirmations (latest: %d, confirmations: %d)",
			chainName, latestHeight, chainConfig.Scan.Confirmations)
		return
	}

	// 检查是否有遗漏的区块需要补扫
	startHeight := lastHeight + 1
	if startHeight == 1 { // 第一次扫描
		startHeight = safeHeight
	}

	// 扫描从startHeight到safeHeight的所有区块
	for height := startHeight; height <= safeHeight; height++ {
		select {
		case <-bs.stopChan:
			return // 如果收到停止信号，立即退出
		default:
			logrus.Infof("[%s] Scanning block at height %d (latest: %d, safe: %d)", chainName, height, latestHeight, safeHeight)

			// 扫描单个区块
			go bs.scanSingleBlock(chainName, scanner, height, chainConfig)

			// 线程安全地更新最后扫描高度
			heightMutex.Lock()
			*lastScanHeight = height
			heightMutex.Unlock()
		}
	}
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

	// 获取交易信息 - 直接从区块获取，避免哈希不一致问题
	transactions, err := scanner.GetBlockTransactionsFromBlock(block)
	if err != nil {
		logrus.Warnf("[%s] Failed to get transactions for block %d: %v", chainName, height, err)
	} else {
		scanner.CalculateBlockStats(block, transactions)
		// 提取到服务器
	}

	// 上传交易信息到服务器
	if len(transactions) > 0 {
		if err := bs.submitTransactionsToServer(chainName, block, transactions); err != nil {
			logrus.Warnf("[%s] Failed to submit transactions for block %d: %v", chainName, height, err)
		}
	}

	// 提交到服务器
	if err := bs.submitBlockToServer(block); err != nil {
		logrus.Errorf("[%s] Failed to submit block %d to server: %v", chainName, height, err)
		return
	}

	// 保存到文件（如果启用）
	if chainConfig.Scan.SaveToFile {
		if err := bs.saveBlockToFile(block, chainConfig.Scan.OutputDir); err != nil {
			logrus.Warnf("[%s] Failed to save block %d to file: %v", chainName, height, err)
		}
	}

	processTime := time.Since(startTime).Milliseconds()
	logrus.Infof("[%s] Successfully processed block %d (hash: %s, %d tx, %dms)",
		chainName, height, block.Hash[:16]+"...", block.TransactionCount, processTime)
}

// submitBlockToServer 提交区块到服务器
func (bs *BlockScanner) submitBlockToServer(block *models.Block) error {
	// 获取API实例
	api := config.GetScannerAPI()
	if api == nil {
		return fmt.Errorf("scanner API not initialized")
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
	}

	// 使用API上传区块
	_, err := api.UploadBlock(blockRequest)
	if err != nil {
		return fmt.Errorf("failed to upload block via API: %w", err)
	}

	return nil
}

// submitTransactionsToServer 将交易信息提交到服务器
func (bs *BlockScanner) submitTransactionsToServer(chainName string, block *models.Block, transactions []map[string]interface{}) error {
	// 获取API实例
	api := config.GetScannerAPI()
	if api == nil {
		return fmt.Errorf("scanner API instance not available")
	}

	// 批量上传交易
	for _, tx := range transactions {
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
			"tx_id":         tx["hash"],
			"tx_type":       0,   // 默认类型
			"confirm":       1,   // 默认确认数
			"status":        1,   // 成功状态
			"send_status":   1,   // 已广播
			"balance":       "0", // 暂时设为0
			"amount":        tx["value"],
			"trans_id":      0, // 暂时设为0
			"chain":         chainName,
			"symbol":        chainName,
			"address_from":  fromAddress,
			"address_to":    toAddress,
			"gas_limit":     tx["gasLimit"],
			"gas_price":     tx["gasPrice"],
			"gas_used":      tx["gasUsed"],
			"fee":           "0", // 暂时设为0，后续计算
			"used_fee":      nil,
			"height":        block.Height,
			"contract_addr": "", // 暂时设为空
			"hex":           nil,
			"tx_scene":      "blockchain_scan",
			"remark":        "Scanned from blockchain",
			"block_index":   tx["block_index"],
			"logs":          tx["logs"],
			"receipt":       tx["receipt"],
		}

		// 上传交易
		if err := api.UploadTransaction(txRequest); err != nil {
			logrus.Warnf("[%s] Failed to upload transaction %s: %v", chainName, tx["hash"], err)
			continue // 继续上传其他交易
		}
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
