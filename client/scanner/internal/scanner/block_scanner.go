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
	lastScanHeight := uint64(0)

	// 持续扫描循环 - 永不停止直到程序关闭
	for bs.IsRunning() {
		// 获取最新区块高度
		latestHeight, err := scanner.GetLatestBlockHeight()
		if err != nil {
			logrus.Errorf("[%s] Failed to get latest block height: %v", chainName, err)
			time.Sleep(chainConfig.Scan.RetryDelay)
			continue
		}

		// 计算确认后的安全高度
		safeHeight := latestHeight
		if latestHeight > uint64(chainConfig.Scan.Confirmations) {
			safeHeight = latestHeight - uint64(chainConfig.Scan.Confirmations)
		}

		if lastScanHeight > 0 && lastScanHeight >= safeHeight {
			time.Sleep(chainConfig.Scan.Interval)
			continue
		}

		// 如果安全高度为0，说明还没有足够的确认，等待新区块
		if safeHeight == 0 {
			logrus.Debugf("[%s] Safe height is 0, waiting for more confirmations (latest: %d, confirmations: %d)",
				chainName, latestHeight, chainConfig.Scan.Confirmations)
			time.Sleep(chainConfig.Scan.Interval)
			continue
		}

		// 扫描当前高度的区块
		logrus.Infof("[%s] Scanning block at height %d (latest: %d, safe: %d)", chainName, safeHeight, latestHeight, safeHeight)

		// 扫描单个区块
		bs.scanSingleBlock(chainName, scanner, safeHeight, chainConfig)

		lastScanHeight = safeHeight
		// 按照配置的间隔休息
		time.Sleep(chainConfig.Scan.Interval)
	}

	logrus.Infof("[%s] Stopped continuous scanning", chainName)
}

// scanSingleBlock 扫描单个区块
func (bs *BlockScanner) scanSingleBlock(chainName string, scanner Scanner, height uint64, chainConfig *config.ChainConfig) {
	startTime := time.Now()

	// 重试机制
	for retry := 0; retry <= chainConfig.Scan.MaxRetries; retry++ {
		block, err := scanner.GetBlockByHeight(height)
		if err != nil {
			if retry < chainConfig.Scan.MaxRetries {
				logrus.Warnf("[%s] Failed to get block %d (retry %d/%d): %v",
					chainName, height, retry+1, chainConfig.Scan.MaxRetries, err)
				time.Sleep(chainConfig.Scan.RetryDelay)
				continue
			} else {
				logrus.Errorf("[%s] Failed to get block %d after %d retries: %v",
					chainName, height, chainConfig.Scan.MaxRetries, err)
				return
			}
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
			} else {
				logrus.Infof("[%s] Successfully submitted %d transactions for block %d", chainName, len(transactions), height)
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
		return // 单个区块扫描成功后退出
	}
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
		MerkleRoot:       block.MerkleRoot,
		Timestamp:        block.Timestamp,
		Difficulty:       block.Difficulty,
		Nonce:            block.Nonce,
		Size:             block.Size,
		TransactionCount: block.TransactionCount,
		TotalAmount:      block.TotalAmount,
		Fee:              block.Fee,
		Confirmations:    block.Confirmations,
		IsOrphan:         block.IsOrphan,
		Chain:            block.Chain,
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
			"symbol":        chainName,
			"address_from":  tx["from"],
			"address_to":    tx["to"],
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
