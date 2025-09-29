package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	// 交易预取滑动窗口（每链）
	txPrefetchers map[string]*txPrefetchState
	// 最新高度缓存（每链）
	latestHeightCache map[string]uint64
	latestMu          sync.RWMutex
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
		config:            cfg,
		scanners:          make(map[string]Scanner),
		stopChan:          make(chan struct{}),
		txPrefetchers:     make(map[string]*txPrefetchState),
		latestHeightCache: make(map[string]uint64),
	}

	// 初始化各种链的扫块器
	scanner.initializeScanners()

	// 启动每链最新高度刷新器
	scanner.startLatestHeightRefreshers()

	return scanner
}

// txPrefetchState 维护一个异步滑动窗口缓存：区块高度 -> (区块, 交易列表)
type txPrefetchState struct {
	mu         sync.RWMutex
	windowSize int
	sem        chan struct{} // 限制预取并发
	blocks     map[uint64]*models.Block
	txs        map[uint64][]map[string]interface{}
}

func newTxPrefetchState(windowSize, concurrency int) *txPrefetchState {
	if windowSize <= 0 {
		windowSize = 5
	}
	if concurrency <= 0 {
		concurrency = 5
	}
	return &txPrefetchState{
		windowSize: windowSize,
		sem:        make(chan struct{}, concurrency),
		blocks:     make(map[uint64]*models.Block),
		txs:        make(map[uint64][]map[string]interface{}),
	}
}

func (ps *txPrefetchState) get(height uint64) (*models.Block, []map[string]interface{}, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	b, okb := ps.blocks[height]
	t, okt := ps.txs[height]
	if okb && okt {
		return b, t, true
	}
	return nil, nil, false
}

func (ps *txPrefetchState) set(height uint64, b *models.Block, t []map[string]interface{}) {
	ps.mu.Lock()
	ps.blocks[height] = b
	ps.txs[height] = t
	ps.mu.Unlock()
}

func (ps *txPrefetchState) has(height uint64) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	_, okb := ps.blocks[height]
	_, okt := ps.txs[height]
	return okb && okt
}

func (ps *txPrefetchState) pruneBelow(minHeight uint64) {
	ps.mu.Lock()
	for h := range ps.blocks {
		if h < minHeight {
			delete(ps.blocks, h)
			delete(ps.txs, h)
		}
	}
	ps.mu.Unlock()
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
			btcconfig := &chainConfig
			scanner = scanners.NewBitcoinScanner(btcconfig)
		case "eth":
			ethconfig := &chainConfig
			scanner = scanners.NewEthereumScanner(ethconfig)
		case "bsc":
			bscConfig := &chainConfig
			scanner = scanners.NewBSCScanner(bscConfig)
		case "sol":
			solConfig := &chainConfig
			scanner = scanners.NewSolanaScanner(solConfig) // 使用改进的原生实现
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

	// 发送启动成功通知
	bs.sendDingTalkInfo(
		"扫块器启动成功",
		"扫块器已成功启动，开始扫描区块链数据",
	)

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

	// 发送停止通知
	bs.sendDingTalkInfo(
		"扫块器停止",
		"扫块器已停止运行",
	)

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

// 周期性刷新各链的最新高度缓存，降低RPC压力
func (bs *BlockScanner) startLatestHeightRefreshers() {
	for chainName, sc := range bs.scanners {
		go func(chain string, scanner Scanner) {
			// 默认300ms刷新一次
			ticker := time.NewTicker(300 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-bs.stopChan:
					return
				case <-ticker.C:
					if h, err := scanner.GetLatestBlockHeight(); err == nil {
						bs.latestMu.Lock()
						bs.latestHeightCache[chain] = h
						bs.latestMu.Unlock()
					}
				}
			}
		}(chainName, sc)
	}
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
	cycleStart := time.Now()
	var durLatest, durGetLast, durPrefetch, durScan time.Duration
	// 1. 获取最新区块高度（优先使用缓存）
	stepStart := time.Now()
	var latestHeight uint64
	bs.latestMu.RLock()
	cached, ok := bs.latestHeightCache[chainName]
	bs.latestMu.RUnlock()
	if ok && cached > 0 {
		latestHeight = cached
	} else {
		// 缓存未命中时才调用RPC
		latest, err := scanner.GetLatestBlockHeight()
		if err != nil {
			logrus.Errorf("[%s] Failed to get latest block height: %v", chainName, err)
			return
		}
		latestHeight = latest
		bs.latestMu.Lock()
		bs.latestHeightCache[chainName] = latestHeight
		bs.latestMu.Unlock()
	}
	durLatest = time.Since(stepStart)

	// 2. 计算确认后的安全高度
	safeHeight := latestHeight
	if latestHeight > uint64(chainConfig.Confirmations) {
		safeHeight = latestHeight - uint64(chainConfig.Confirmations)
	}

	// 3. 如果安全高度为0，说明还没有足够的确认，等待新区块
	if safeHeight == 0 {
		logrus.Debugf("[%s] Safe height is 0, waiting for more confirmations (latest: %d, confirmations: %d)",
			chainName, latestHeight, chainConfig.Confirmations)
		return
	}

	// 4. 获取最后一个验证通过的区块高度
	stepStart = time.Now()
	lastVerifiedHeight, err := bs.getLastVerifiedBlockHeight(chainName)
	durGetLast = time.Since(stepStart)

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

	// 6. 计算要扫描的区块高度范围
	startHeight := lastVerifiedHeight + 1
	k := chainConfig.Scan.HeightsPerCycle
	if k <= 0 {
		k = 5
	}
	endHeight := startHeight + uint64(k) - 1
	if endHeight > safeHeight {
		endHeight = safeHeight
	}

	if endHeight < startHeight {
		return
	}

	logrus.Infof("[%s] Scanning blocks %d..%d (latest: %d, safe: %d, last verified: %d)", chainName, startHeight, endHeight, latestHeight, safeHeight, lastVerifiedHeight)

	// 6.1 预取窗口推进至 endHeight
	stepStart = time.Now()
	bs.ensureTxPrefetch(chainName, scanner, chainConfig, startHeight, safeHeight)
	durPrefetch = time.Since(stepStart)

	// 7. 顺序扫描多个区块（保持提交顺序，与预取不冲突）
	for h := startHeight; h <= endHeight; h++ {
		select {
		case <-bs.stopChan:
			logrus.Infof("[%s] Stopped scanning due to stop signal", chainName)
			return
		default:
		}
		step := time.Now()
		bs.scanSingleBlock(chainName, scanner, h, chainConfig)
		d := time.Since(step)
		durScan += d // 记录本轮所有区块总耗时
	}

	logrus.Infof("[%s] Completed scan cycle for heights %d..%d", chainName, startHeight, endHeight)

	// 输出本次周期的耗时统计
	cycleTotal := time.Since(cycleStart)
	logrus.Infof("[%s] Cycle timing (range=%d..%d): total=%dms latest=%dms last=%dms prefetch=%dms scanSum=%dms",
		chainName,
		startHeight,
		endHeight,
		cycleTotal.Milliseconds(),
		durLatest.Milliseconds(),
		durGetLast.Milliseconds(),
		durPrefetch.Milliseconds(),
		durScan.Milliseconds(),
	)
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

	// 优先从预取缓存取出区块与交易
	var block *models.Block
	var transactions []map[string]interface{}
	if ps, ok := bs.txPrefetchers[chainName]; ok {
		if b, txs, ok2 := ps.get(height); ok2 {
			block = b
			transactions = txs
		}
	}
	if block == nil {
		blk, getErr := scanner.GetBlockByHeight(height)
		if getErr != nil {
			logrus.Errorf("[%s] Failed to get block %d: %v", chainName, height, getErr)
			// 发送钉钉警告通知
			bs.sendDingTalkWarning(
				fmt.Sprintf("扫块器获取区块失败 - %s", chainName),
				fmt.Sprintf("链: %s\n区块高度: %d\n错误: %v", chainName, height, getErr),
			)
			return
		}
		block = blk
	}

	// 获取交易信息 - 若未预取，则从区块获取
	var txErr error
	if transactions == nil {
		transactions, txErr = scanner.GetBlockTransactionsFromBlock(block)
	}

	if txErr != nil {
		logrus.Warnf("[%s] Failed to get transactions for block %d: %v", chainName, height, txErr)
		// 如果获取交易失败，仍然需要更新验证高度
		bs.updateLastVerifiedHeight(chainName, height, block.Hash)
		return
	}

	// 检查区块是否包含目标地址的交易（对 Solana 和 BSC 链）
	if chainName == "sol" || chainName == "bsc" {
		if !bs.containsTargetAddresses(chainName, transactions) {
			// 更新验证高度并跳过区块创建
			bs.updateLastVerifiedHeight(chainName, height, block.Hash)
			return
		}
	}

	// 验证区块
	if err := scanner.ValidateBlock(block); err != nil {
		logrus.Errorf("[%s] Block validation failed for block %d: %v", chainName, height, err)
		// 发送钉钉警告通知
		bs.sendDingTalkWarning(
			fmt.Sprintf("区块验证失败 - %s", chainName),
			fmt.Sprintf("链: %s\n区块高度: %d\n区块哈希: %s\n错误: %v", chainName, height, block.Hash, err),
		)
		return
	}

	// 提交区块到服务器，获取区块ID
	blockID, err := bs.submitBlockToServer(block)
	if err != nil {
		logrus.Errorf("[%s] Failed to submit block %d to server: %v", chainName, height, err)
		// 发送钉钉警告通知
		bs.sendDingTalkWarning(
			fmt.Sprintf("区块提交失败 - %s", chainName),
			fmt.Sprintf("链: %s\n区块高度: %d\n区块哈希: %s\n错误: %v", chainName, height, block.Hash, err),
		)
		return
	}

	// 计算区块统计信息
	scanner.CalculateBlockStats(block, transactions)

	// 上传交易信息到服务器，传入区块ID
	if err := bs.submitTransactionsToServer(chainName, block, transactions, blockID); err != nil {
		logrus.Warnf("[%s] Failed to submit transactions for block %d: %v", chainName, height, err)
		return
	}

	bs.updateBlockStatsToServer(block, transactions, blockID)

	// 验证区块
	if verr := bs.verifyBlock(blockID); verr != nil {
		logrus.Errorf("[%s] Block verification failed for block %d: %v", chainName, height, verr)
		return
	}

	// 保存到文件（如果启用）
	if chainConfig.Scan.SaveToFile {
		if ferr := bs.saveBlockToFile(block, chainConfig.Scan.OutputDir); ferr != nil {
			logrus.Warnf("[%s] Failed to save block %d to file: %v", chainName, height, ferr)
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
	if block.Miner != "" {
		payload["miner"] = block.Miner
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
		ChainID:          block.ChainID,
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
	if chainName != "sol" {

		if err := api.UploadTransactionsBatch(batchTransactions); err != nil {
			// 检查是否为致命错误，如果是就直接退出程序
			if strings.Contains(strings.ToUpper(err.Error()), "BLOCK_TRANSACTION_FAILED") {
				pkg.HandleFatalError(err, "交易批量上传失败")
			}

			return fmt.Errorf("batch upload transactions failed: %w", err)
		}
	}

	logrus.Infof("[%s] Successfully uploaded %d transactions in batch for block %d", chainName, len(batchTransactions), block.Height)

	// 如果是 Sol 链，额外上传转账事件与交易明细（最佳努力）
	if chainName == "sol" {
		bs.processSolanaArtifacts(transactions, block, blockID)
	}
	return nil
}

// submitTransactionsIndividually 单个上传交易（保持原有逻辑作为备选）
func (bs *BlockScanner) submitTransactionsIndividually(chainName string, block *models.Block, transactions []map[string]interface{}, blockID uint64, chainConfig *config.ChainConfig, api *pkg.ScannerAPI) error {
	// 并发上限：使用链配置的 MaxConcurrentUpload，缺省为 10
	concurrency := 10
	if chainConfig.Scan.MaxConcurrentUpload > 0 {
		concurrency = chainConfig.Scan.MaxConcurrentUpload
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

			if chainName != "sol" {

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
	if chainName == "sol" {
		bs.processSolanaArtifacts(transactions, block, blockID)
	}
	return nil
}

// processSolanaArtifacts 处理 Solana 特定的工件（委托给 SolanaScanner）
func (bs *BlockScanner) processSolanaArtifacts(transactions []map[string]interface{}, block *models.Block, blockID uint64) {
	// 获取 Solana 扫描器实例
	if solanaScanner, exists := bs.scanners["sol"]; exists {
		if solScanner, ok := solanaScanner.(*scanners.SolanaScanner); ok {
			// 转换 block 为通用格式
			blockData := map[string]interface{}{
				"height": block.Height,
				"hash":   block.Hash,
			}
			// 委托给 Solana 扫描器处理
			solScanner.ProcessSolanaArtifacts(transactions, blockData, blockID)
		}
	}
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

	// 构建交易请求数据
	txRequest := map[string]interface{}{
		"tx_id": tx["hash"],
		"tx_type": func() uint8 {
			if txType, ok := tx["type"].(uint8); ok {
				return txType
			}
			return 0
		}(),
		"confirm":     1,
		"status":      1,
		"send_status": 1,
		"balance":     "0",
		"amount": func() string {
			if amount, ok := tx["value"].(float64); ok {
				return strconv.FormatFloat(amount, 'f', -1, 64)
			} else if amount, ok := tx["value"].(uint64); ok {
				return strconv.FormatUint(amount, 10)
			} else if amount, ok := tx["value"].(int); ok {
				if amount < 0 {
					return "0"
				}
				return strconv.FormatInt(int64(amount), 10)
			} else if amount, ok := tx["value"].(string); ok {
				return amount
			}
			return "0"
		}(),
		"trans_id":     0,
		"chain":        chainName,
		"symbol":       chainName,
		"address_from": fromAddress,
		"address_to":   toAddress,
		"gas_limit":    tx["gasLimit"],
		"gas_price":    tx["gasPrice"],
		"gas_used":     tx["gasUsed"],
		"fee": func() string {
			if fee, ok := tx["fee"].(float64); ok {
				return strconv.FormatFloat(fee, 'f', -1, 64)
			} else if fee, ok := tx["fee"].(uint64); ok {
				return strconv.FormatUint(fee, 10)
			} else if fee, ok := tx["fee"].(int); ok {
				if fee < 0 {
					return "0"
				}
				return strconv.FormatInt(int64(fee), 10)
			} else if fee, ok := tx["fee"].(string); ok {
				return fee
			}
			return "0"
		}(),
		"used_fee": nil,
		"height":   block.Height,
		"block_id": blockID,
		"contract_addr": func() string {
			if contractAddr, ok := tx["contract_address"].(string); ok {
				return contractAddr
			}
			return ""
		}(),
		"hex":           tx["data"],
		"tx_scene":      "blockchain_scan",
		"remark":        "Scanned from blockchain",
		"block_index":   tx["block_index"],
		"nonce":         tx["nonce"],
		"logs":          tx["logs"],
		"receipt":       tx["receipt"],
		"vin":           tx["vin"],
		"vout":          tx["vout"],
		"address_froms": tx["address_froms"],
		"address_tos":   tx["address_tos"],
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

// ensureTxPrefetch 启动或推进交易预取滑动窗口
func (bs *BlockScanner) ensureTxPrefetch(chainName string, scanner Scanner, chainConfig *config.ChainConfig, nextHeight uint64, safeHeight uint64) {
	ps := bs.txPrefetchers[chainName]
	if ps == nil {
		window := chainConfig.Scan.PrefetchWindow
		if window <= 0 {
			window = 5
		}
		concurrency := chainConfig.Scan.MaxConcurrentPrefetch
		ps = newTxPrefetchState(window, concurrency)
		bs.txPrefetchers[chainName] = ps
	}
	// 目标上界
	var upper uint64
	if nextHeight+uint64(ps.windowSize)-1 <= safeHeight {
		upper = nextHeight + uint64(ps.windowSize) - 1
	} else {
		upper = safeHeight
	}
	for h := nextHeight; h <= upper; h++ {
		if ps.has(h) {
			continue
		}
		select {
		case ps.sem <- struct{}{}:
			go func(height uint64) {
				defer func() { <-ps.sem }()
				// 拉取区块
				blk, err := scanner.GetBlockByHeight(height)
				if err != nil {
					logrus.Warnf("[%s] Prefetch block %d failed: %v", chainName, height, err)
					return
				}
				// 拉取交易
				txs, err := scanner.GetBlockTransactionsFromBlock(blk)
				if err != nil {
					logrus.Warnf("[%s] Prefetch txs for block %d failed: %v", chainName, height, err)
					return
				}
				ps.set(height, blk, txs)
			}(h)
		default:
			// 并发已满，后续周期再补
			return
		}
	}
	// 修剪窗口以下的数据，避免内存膨胀
	ps.pruneBelow(nextHeight)
}

// sendDingTalkWarning 发送钉钉警告通知
func (bs *BlockScanner) sendDingTalkWarning(title, message string) {
	notifier := config.GetDingTalkNotifier()
	if notifier != nil {
		// 异步发送通知，避免阻塞主程序
		go func() {
			if err := notifier.SendWarning(title, message); err != nil {
				logrus.Errorf("Failed to send DingTalk warning notification: %v", err)
			}
		}()
	}
}

// sendDingTalkInfo 发送钉钉信息通知
func (bs *BlockScanner) sendDingTalkInfo(title, message string) {
	notifier := config.GetDingTalkNotifier()
	if notifier != nil {
		// 异步发送通知，避免阻塞主程序
		go func() {
			if err := notifier.SendInfo(title, message); err != nil {
				logrus.Errorf("Failed to send DingTalk info notification: %v", err)
			}
		}()
	}
}

// containsTargetAddresses 检查交易列表中是否包含目标地址
func (bs *BlockScanner) containsTargetAddresses(chainName string, transactions []map[string]interface{}) bool {
	// 获取目标地址列表
	targetAddresses := bs.getTargetAddresses(chainName)
	if len(targetAddresses) == 0 {
		// 如果没有配置目标地址，则处理所有交易
		return true
	}

	// 创建地址映射以提高查找效率
	addressMap := make(map[string]bool)
	for _, addr := range targetAddresses {
		addressMap[strings.ToLower(addr)] = true
	}

	// 检查每个交易是否包含目标地址
	for _, tx := range transactions {
		if bs.transactionContainsTargetAddress(tx, addressMap) {
			return true
		}
	}

	return false
}

// getTargetAddresses 获取指定链的目标地址列表
func (bs *BlockScanner) getTargetAddresses(chainName string) []string {
	var addresses []string

	// 从配置中获取钱包地址
	for _, wallet := range bs.config.Blockchain.WalletAddresses {
		if wallet.Chain == chainName {
			addresses = append(addresses, wallet.Address)
		}
	}

	return addresses
}

// transactionContainsTargetAddress 检查单个交易是否包含目标地址
func (bs *BlockScanner) transactionContainsTargetAddress(tx map[string]interface{}, addressMap map[string]bool) bool {
	// 检查 from 地址
	if from, ok := tx["from"].(string); ok {
		if addressMap[strings.ToLower(from)] {
			return true
		}
	}

	// 检查 to 地址
	if to, ok := tx["to"].(string); ok {
		if addressMap[strings.ToLower(to)] {
			return true
		}
	}

	// 检查 Solana 特有的地址字段
	if accountKeys, ok := tx["account_keys"].([]interface{}); ok {
		for _, key := range accountKeys {
			if keyStr, ok := key.(string); ok {
				if addressMap[strings.ToLower(keyStr)] {
					return true
				}
			}
		}
	}

	// 检查 Solana 事件中的地址
	if solEvents, ok := tx["sol_events"].([]map[string]interface{}); ok {
		for _, event := range solEvents {
			if from, ok := event["from"].(string); ok {
				if addressMap[strings.ToLower(from)] {
					return true
				}
			}
			if to, ok := event["to"].(string); ok {
				if addressMap[strings.ToLower(to)] {
					return true
				}
			}
			if fromAccount, ok := event["from_account"].(string); ok {
				if addressMap[strings.ToLower(fromAccount)] {
					return true
				}
			}
			if toAccount, ok := event["to_account"].(string); ok {
				if addressMap[strings.ToLower(toAccount)] {
					return true
				}
			}
		}
	}

	return false
}

// updateLastVerifiedHeight 更新最后验证的区块高度
func (bs *BlockScanner) updateLastVerifiedHeight(chainName string, height uint64, hash string) {
	api := config.GetScannerAPI()
	if api == nil {
		logrus.Warnf("[%s] Scanner API not available, cannot update last verified height", chainName)
		return
	}

	if err := api.UpdateLastVerifiedBlockHeight(chainName, height, hash); err != nil {
		logrus.Errorf("[%s] Failed to update last verified height %d: %v", chainName, height, err)
		return
	}

	logrus.Debugf("[%s] Updated last verified height to %d (hash: %s)", chainName, height, hash)
}
