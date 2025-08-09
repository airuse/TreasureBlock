package scanner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"blockChainBrowser/client/scanner/config"
	"blockChainBrowser/client/scanner/internal/models"
	"blockChainBrowser/client/scanner/internal/scanners"

	"github.com/sirupsen/logrus"
)

// BlockScanner 主扫块器
type BlockScanner struct {
	config        *config.Config
	scanners      map[string]interface{}
	progress      map[string]*models.ScanProgress
	progressMutex sync.RWMutex
	stopChan      chan struct{}
	running       bool
	runningMutex  sync.RWMutex
	httpClient    *http.Client
}

// Scanner 扫块器接口
type Scanner interface {
	GetLatestBlockHeight() (uint64, error)
	GetBlockByHeight(height uint64) (*models.Block, error)
	ValidateBlock(block *models.Block) error
	GetBlockTransactions(blockHash string) ([]map[string]interface{}, error)
	CalculateBlockStats(block *models.Block, transactions []map[string]interface{})
}

// NewBlockScanner 创建新的主扫块器
func NewBlockScanner(cfg *config.Config) *BlockScanner {
	scanner := &BlockScanner{
		config:   cfg,
		scanners: make(map[string]interface{}),
		progress: make(map[string]*models.ScanProgress),
		stopChan: make(chan struct{}),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 初始化各种链的扫块器
	scanner.initializeScanners()

	return scanner
}

// initializeScanners 初始化各种链的扫块器
func (bs *BlockScanner) initializeScanners() {
	for chainName, chainConfig := range bs.config.Blockchain.Chains {
		if !chainConfig.Enabled {
			continue
		}

		var scanner Scanner
		switch chainName {
		case "btc":
			scanner = scanners.NewBitcoinScanner(&chainConfig)
		case "eth":
			scanner = scanners.NewEthereumScanner(&chainConfig)
		default:
			logrus.Warnf("Unsupported chain: %s", chainName)
			continue
		}

		bs.scanners[chainName] = scanner

		// 初始化进度
		bs.progress[chainName] = &models.ScanProgress{
			Chain:           chainName,
			CurrentHeight:   bs.config.Scan.StartBlockHeight,
			ProcessedBlocks: 0,
			FailedBlocks:    0,
			StartTime:       time.Now(),
			LastUpdateTime:  time.Now(),
			Status:          "stopped",
		}
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

	// 更新所有链的状态为停止
	bs.progressMutex.Lock()
	for _, progress := range bs.progress {
		progress.Status = "stopped"
		progress.LastUpdateTime = time.Now()
	}
	bs.progressMutex.Unlock()
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
		}(chainName, scanner.(Scanner))
	}

	wg.Wait()
}

// scanChain 扫描指定链 - 持续扫描模式
func (bs *BlockScanner) scanChain(chainName string, scanner Scanner) {
	// 初始化时已经确保所有启用的链都有进度记录
	progress := bs.getProgress(chainName)

	// 更新状态为运行中
	bs.updateProgress(chainName, func(p *models.ScanProgress) {
		p.Status = "running"
		p.LastUpdateTime = time.Now()
	})

	logrus.Infof("Starting continuous scanning for chain %s", chainName)

	// 持续扫描循环 - 永不停止直到程序关闭
	for bs.IsRunning() {
		// 获取最新区块高度
		latestHeight, err := scanner.GetLatestBlockHeight()
		if err != nil {
			logrus.Errorf("Failed to get latest block height for %s: %v", chainName, err)
			// 出错时等待一下再重试
			time.Sleep(bs.config.Scan.RetryDelay)
			continue
		}

		// 更新最新高度到进度中
		bs.updateProgress(chainName, func(p *models.ScanProgress) {
			p.LatestHeight = latestHeight
		})

		// 确定当前需要扫描的起始高度
		currentHeight := progress.CurrentHeight
		if currentHeight == 0 {
			// 如果是第一次运行，从配置的起始高度开始
			if bs.config.Scan.StartBlockHeight > 0 {
				currentHeight = bs.config.Scan.StartBlockHeight
			} else {
				currentHeight = 1
			}
		}

		// 计算确认后的安全高度
		safeHeight := latestHeight
		if latestHeight > uint64(bs.config.Scan.Confirmations) {
			safeHeight = latestHeight - uint64(bs.config.Scan.Confirmations)
		}

		// 如果当前高度已经达到安全高度，等待新区块
		if currentHeight > safeHeight {
			logrus.Debugf("Chain %s: waiting for new blocks (current: %d, safe: %d, latest: %d)",
				chainName, currentHeight, safeHeight, latestHeight)
			time.Sleep(5 * time.Second) // 等待5秒再检查
			continue
		}

		// 计算本轮扫描的结束高度
		batchEnd := currentHeight + uint64(bs.config.Scan.BatchSize) - 1
		if batchEnd > safeHeight {
			batchEnd = safeHeight
		}

		logrus.Infof("Chain %s: scanning blocks %d to %d (latest: %d)",
			chainName, currentHeight, batchEnd, latestHeight)

		// 扫描这一批区块
		bs.scanBlockBatch(chainName, scanner, currentHeight, batchEnd)

		// 更新当前高度到下一个区块
		bs.updateProgress(chainName, func(p *models.ScanProgress) {
			p.CurrentHeight = batchEnd + 1
			p.LastUpdateTime = time.Now()
		})

		// 重新获取更新后的进度
		progress = bs.getProgress(chainName)

		// 短暂休息，避免过于频繁的请求
		time.Sleep(1 * time.Second)
	}

	// 更新状态为停止
	bs.updateProgress(chainName, func(p *models.ScanProgress) {
		p.Status = "stopped"
		p.LastUpdateTime = time.Now()
	})

	logrus.Infof("Stopped continuous scanning for chain %s", chainName)
}

// scanBlockBatch 批量扫描区块
func (bs *BlockScanner) scanBlockBatch(chainName string, scanner Scanner, startHeight, endHeight uint64) {
	for height := startHeight; height <= endHeight; height++ {
		startTime := time.Now()
		result := &models.BlockScanResult{
			Chain:      chainName,
			Height:     height,
			Status:     "success",
			RetryCount: 0,
		}

		// 重试机制
		for retry := 0; retry <= bs.config.Scan.MaxRetries; retry++ {
			block, err := scanner.GetBlockByHeight(height)
			if err != nil {
				result.RetryCount = retry
				if retry < bs.config.Scan.MaxRetries {
					logrus.Warnf("Failed to get block %d for chain %s (retry %d/%d): %v",
						height, chainName, retry+1, bs.config.Scan.MaxRetries, err)
					time.Sleep(bs.config.Scan.RetryDelay)
					continue
				} else {
					result.Status = "failed"
					result.Error = err.Error()
					logrus.Errorf("Failed to get block %d for chain %s after %d retries: %v",
						height, chainName, bs.config.Scan.MaxRetries, err)
					break
				}
			}

			// 验证区块
			if err := scanner.ValidateBlock(block); err != nil {
				result.Status = "failed"
				result.Error = fmt.Sprintf("block validation failed: %v", err)
				logrus.Errorf("Block validation failed for block %d on chain %s: %v", height, chainName, err)
				break
			}

			// 获取交易信息
			transactions, err := scanner.GetBlockTransactions(block.Hash)
			if err != nil {
				logrus.Warnf("Failed to get transactions for block %d on chain %s: %v", height, chainName, err)
			} else {
				scanner.CalculateBlockStats(block, transactions)
			}

			// 提交到服务器
			if err := bs.submitBlockToServer(block); err != nil {
				result.Status = "failed"
				result.Error = fmt.Sprintf("failed to submit to server: %v", err)
				logrus.Errorf("Failed to submit block %d to server for chain %s: %v", height, chainName, err)
				break
			}

			// 保存到文件（如果启用）
			if bs.config.Scan.SaveToFile {
				if err := bs.saveBlockToFile(block); err != nil {
					logrus.Warnf("Failed to save block %d to file for chain %s: %v", height, chainName, err)
				}
			}

			result.Hash = block.Hash
			result.Timestamp = block.Timestamp
			result.ProcessTime = time.Since(startTime).Milliseconds()

			logrus.Infof("Successfully processed block %d for chain %s", height, chainName)
			break
		}

	}
}

// submitBlockToServer 提交区块到服务器
func (bs *BlockScanner) submitBlockToServer(block *models.Block) error {
	serverURL := fmt.Sprintf("%s://%s:%d/api/v1/blocks",
		bs.config.Server.Protocol,
		bs.config.Server.Host,
		bs.config.Server.Port)

	_, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	req, err := http.NewRequest("POST", serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if bs.config.Server.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+bs.config.Server.APIKey)
	}

	resp, err := bs.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit block: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// saveBlockToFile 保存区块到文件
func (bs *BlockScanner) saveBlockToFile(block *models.Block) error {
	// 确保输出目录存在
	if err := os.MkdirAll(bs.config.Scan.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 创建链目录
	chainDir := filepath.Join(bs.config.Scan.OutputDir, block.Chain)
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

// getProgress 获取扫描进度
func (bs *BlockScanner) getProgress(chainName string) *models.ScanProgress {
	bs.progressMutex.RLock()
	defer bs.progressMutex.RUnlock()
	return bs.progress[chainName]
}

// updateProgress 更新扫描进度
func (bs *BlockScanner) updateProgress(chainName string, updater func(*models.ScanProgress)) {
	bs.progressMutex.Lock()
	defer bs.progressMutex.Unlock()

	// 初始化时已经确保所有启用的链都有进度记录，无需检查exists
	updater(bs.progress[chainName])
}

// GetProgress 获取所有链的扫描进度
func (bs *BlockScanner) GetProgress() map[string]*models.ScanProgress {
	bs.progressMutex.RLock()
	defer bs.progressMutex.RUnlock()

	result := make(map[string]*models.ScanProgress)
	for chain, progress := range bs.progress {
		// 创建副本避免并发访问问题
		progressCopy := *progress
		result[chain] = &progressCopy
	}

	return result
}
