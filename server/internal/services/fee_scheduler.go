package services

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"blockChainBrowser/server/internal/interfaces"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// FeeScheduler 费率调度器
type FeeScheduler struct {
	rpcManager *utils.RPCClientManager
	baseCfgSvc BaseConfigService
	logger     *logrus.Logger
	wsHandler  interfaces.WebSocketBroadcaster // WebSocket广播接口

	// 缓存上一次推送的费率信息
	lastFeeData map[string]*FeeLevels // key: chain (eth, btc), value: 费率数据
}

// FeeData 费率数据结构
type FeeData struct {
	Chain             string `json:"chain"`
	BaseFee           string `json:"base_fee"`           // Base Fee (Gwei)
	MaxPriorityFee    string `json:"max_priority_fee"`   // Max Priority Fee (Gwei)
	MaxFee            string `json:"max_fee"`            // Max Fee (Gwei)
	GasPrice          string `json:"gas_price"`          // Legacy Gas Price (Gwei)
	LastUpdated       int64  `json:"last_updated"`       // 最后更新时间戳
	BlockNumber       uint64 `json:"block_number"`       // 当前区块号
	NetworkCongestion string `json:"network_congestion"` // 网络拥堵状态
}

// FeeLevels 费率等级
type FeeLevels struct {
	Slow   FeeData `json:"slow"`
	Normal FeeData `json:"normal"`
	Fast   FeeData `json:"fast"`
}

// NewFeeScheduler 创建费率调度器
func NewFeeScheduler() *FeeScheduler {
	return &FeeScheduler{
		rpcManager:  utils.NewRPCClientManager(),
		baseCfgSvc:  NewBaseConfigService(repository.NewBaseConfigRepository()),
		logger:      logrus.New(),
		lastFeeData: make(map[string]*FeeLevels), // 初始化缓存map
	}
}

// SetWebSocketHandler 设置WebSocket处理器
func (fs *FeeScheduler) SetWebSocketHandler(handler interfaces.WebSocketBroadcaster) {
	fs.wsHandler = handler
}

// Start 启动费率调度器
func (fs *FeeScheduler) Start(ctx context.Context) {
	fs.logger.Info("费率调度器已启动")

	// ETH每10秒更新一次，BTC每30秒更新一次（Mempool数据变化较快）
	ethTicker := time.NewTicker(10 * time.Second)
	btcTicker := time.NewTicker(30 * time.Second)
	defer ethTicker.Stop()
	defer btcTicker.Stop()

	// 立即执行一次
	fs.updateFeeData(ctx)

	for {
		select {
		case <-ctx.Done():
			fs.logger.Info("费率调度器已停止")
			return
		case <-ethTicker.C:
			// 更新ETH费率
			fs.updateETHFeeData(ctx)
		case <-btcTicker.C:
			// 更新BTC费率
			fs.updateBTCFeeData(ctx)
		}
	}
}

// updateFeeData 更新费率数据
func (fs *FeeScheduler) updateFeeData(ctx context.Context) {
	// 更新ETH费率
	fs.updateETHFeeData(ctx)
	// 更新BTC费率
	fs.updateBTCFeeData(ctx)
}

// updateETHFeeData 更新ETH费率数据
func (fs *FeeScheduler) updateETHFeeData(ctx context.Context) {
	ethFeeData, err := fs.getETHFeeData(ctx)
	if err != nil {
		fs.logger.Errorf("获取ETH费率失败: %v", err)
	} else {
		fs.broadcastFeeData("eth", ethFeeData)
	}
}

// updateBTCFeeData 更新BTC费率数据
func (fs *FeeScheduler) updateBTCFeeData(ctx context.Context) {
	btcFeeData, err := fs.getBTCFeeData(ctx)
	if err != nil {
		fs.logger.Errorf("获取BTC费率失败: %v", err)
	} else {
		fs.broadcastFeeData("btc", btcFeeData)
	}
}

// getETHFeeData 获取ETH费率数据
func (fs *FeeScheduler) getETHFeeData(ctx context.Context) (*FeeLevels, error) {
	// 获取最新区块信息
	latestBlock, err := fs.getLatestBlock(ctx, "eth")
	if err != nil {
		fs.logger.Errorf("获取最新区块失败: %v", err)
		return fs.getDefaultETHFeeData(), nil
	}

	// 获取base fee
	baseFee, err := fs.getBaseFeeFromBlock(ctx, latestBlock)
	if err != nil {
		fs.logger.Warnf("获取Base Fee失败，使用默认值: %v", err)
		baseFee = "20000000000" // 20 Gwei in wei
	}

	// 获取历史区块的矿工费数据来计算合理的max priority fee
	maxPriorityFee, err := fs.calculateMaxPriorityFeeFromHistory(ctx, latestBlock.Number)
	if err != nil {
		fs.logger.Warnf("计算Max Priority Fee失败，使用默认值: %v", err)
		maxPriorityFee = "2000000000" // 2 Gwei in wei
	}

	// 计算max fee
	maxFee := fs.calculateMaxFee(baseFee, maxPriorityFee)

	// 计算网络拥堵状态
	congestion := fs.calculateNetworkCongestion(baseFee)

	// 创建费率等级
	feeLevels := &FeeLevels{
		Slow: FeeData{
			Chain:             "eth",
			BaseFee:           baseFee,
			MaxPriorityFee:    fs.calculateSlowPriorityFee(maxPriorityFee),
			MaxFee:            fs.calculateMaxFee(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "eth",
			BaseFee:           baseFee,
			MaxPriorityFee:    maxPriorityFee,
			MaxFee:            maxFee,
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, maxPriorityFee),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "eth",
			BaseFee:           baseFee,
			MaxPriorityFee:    fs.calculateFastPriorityFee(maxPriorityFee),
			MaxFee:            fs.calculateMaxFee(baseFee, fs.calculateFastPriorityFee(maxPriorityFee)),
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, fs.calculateFastPriorityFee(maxPriorityFee)),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
	}

	return feeLevels, nil
}

// getBTCFeeData 获取BTC费率数据
func (fs *FeeScheduler) getBTCFeeData(ctx context.Context) (*FeeLevels, error) {
	// 获取最新区块信息
	latestBlock, err := fs.getLatestBTCBlock(ctx)
	if err != nil {
		fs.logger.Errorf("获取最新BTC区块失败: %v", err)
		return fs.getDefaultBTCFeeData(), nil
	}

	// 基于最近3个区块计算费率
	feeRates, err := fs.calculateBTCFeeRatesFromHistory(ctx)
	if err != nil {
		fs.logger.Warnf("计算BTC费率失败，使用默认值: %v", err)
		return fs.getDefaultBTCFeeData(), nil
	}

	// 计算网络拥堵状态
	congestion := fs.calculateBTCNetworkCongestion(feeRates.Normal)

	// 创建费率等级
	feeLevels := &FeeLevels{
		Slow: FeeData{
			Chain:             "btc",
			BaseFee:           feeRates.Slow,
			MaxPriorityFee:    feeRates.Slow,
			MaxFee:            feeRates.Slow,
			GasPrice:          feeRates.Slow,
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "btc",
			BaseFee:           feeRates.Normal,
			MaxPriorityFee:    feeRates.Normal,
			MaxFee:            feeRates.Normal,
			GasPrice:          feeRates.Normal,
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "btc",
			BaseFee:           feeRates.Fast,
			MaxPriorityFee:    feeRates.Fast,
			MaxFee:            feeRates.Fast,
			GasPrice:          feeRates.Fast,
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
	}

	return feeLevels, nil
}

// BlockInfo 区块信息结构
type BlockInfo struct {
	Number    uint64
	BaseFee   *big.Int
	GasUsed   uint64
	GasLimit  uint64
	Timestamp uint64
}

// BTCBlockInfo BTC区块信息结构
type BTCBlockInfo struct {
	Number    uint64
	Hash      string
	Timestamp uint64
	TxCount   int
}

// BTCFeeRates BTC费率结构
type BTCFeeRates struct {
	Slow   string // sat/vB
	Normal string // sat/vB
	Fast   string // sat/vB
}

// BTCTransaction BTC交易结构
type BTCTransaction struct {
	TxID       string
	Fee        int64 // 交易费 (satoshi)
	Size       int   // 交易大小 (bytes)
	FeeRate    int64 // 费率 (satoshi per vbyte)
	IsCoinbase bool  // 是否为coinbase交易
}

// BTCMempoolTransaction Mempool中的BTC交易结构
type BTCMempoolTransaction struct {
	TxID    string
	Fee     int64 // 交易费 (satoshi)
	Size    int   // 交易大小 (bytes)
	FeeRate int64 // 费率 (satoshi per vbyte)
	Time    int64 // 进入Mempool的时间戳
}

// getLatestBlock 获取最新区块
func (fs *FeeScheduler) getLatestBlock(ctx context.Context, chain string) (*BlockInfo, error) {
	// 使用RPCClientManager获取最新区块号
	blockNumber, err := fs.rpcManager.GetBlockNumber(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("获取区块号失败: %v", err)
	}

	// 使用RPCClientManager获取区块详情
	blockInterface, err := fs.rpcManager.GetBlockByNumber(ctx, chain, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("获取区块详情失败: %v", err)
	}

	// 根据链类型解析区块数据
	if chain == "eth" {
		block, ok := blockInterface.(*types.Block)
		if !ok {
			return nil, fmt.Errorf("无效的ETH区块数据格式")
		}

		return &BlockInfo{
			Number:    blockNumber,
			BaseFee:   block.BaseFee(),
			GasUsed:   block.GasUsed(),
			GasLimit:  block.GasLimit(),
			Timestamp: block.Time(),
		}, nil
	}

	// 对于其他链，返回默认值
	return &BlockInfo{
		Number:    blockNumber,
		BaseFee:   big.NewInt(20000000000), // 默认20 Gwei
		GasUsed:   0,
		GasLimit:  0,
		Timestamp: 0,
	}, nil
}

// getBaseFeeFromBlock 从区块获取Base Fee
func (fs *FeeScheduler) getBaseFeeFromBlock(ctx context.Context, block *BlockInfo) (string, error) {
	if block.BaseFee == nil {
		return "20000000000", nil // 默认20 Gwei
	}
	return block.BaseFee.String(), nil
}

// calculateMaxPriorityFeeFromHistory 从上一个区块计算Max Priority Fee
func (fs *FeeScheduler) calculateMaxPriorityFeeFromHistory(ctx context.Context, currentBlockNumber uint64) (string, error) {
	// 只分析上一个区块，获取最新的费率信息
	if currentBlockNumber == 0 {
		return "2000000000", nil // 默认2 Gwei
	}

	previousBlockNumber := currentBlockNumber - 1
	blockInterface, err := fs.rpcManager.GetBlockByNumber(ctx, "eth", big.NewInt(int64(previousBlockNumber)))
	if err != nil {
		fs.logger.Warnf("获取上一个区块失败，使用默认费率: %v", err)
		return "2000000000", nil // 默认2 Gwei
	}

	// 将interface{}转换为ETH区块类型
	block, ok := blockInterface.(*types.Block)
	if !ok {
		fs.logger.Warnf("无效的ETH区块数据格式，使用默认费率")
		return "2000000000", nil // 默认2 Gwei
	}

	// 收集上一个区块中所有EIP-1559交易的矿工费
	var priorityFees []*big.Int
	var totalTxs, eip1559Txs int

	for _, tx := range block.Transactions() {
		totalTxs++
		if tx.Type() == 2 { // EIP-1559 交易
			eip1559Txs++
			if tx.GasTipCap() != nil && tx.GasTipCap().Cmp(big.NewInt(0)) > 0 {
				priorityFees = append(priorityFees, tx.GasTipCap())
			}
		}
	}

	// fs.logger.Infof("区块 %d 统计: 总交易数=%d, EIP-1559交易数=%d, 有效矿工费交易数=%d",
	// 	previousBlockNumber, totalTxs, eip1559Txs, len(priorityFees))

	if len(priorityFees) == 0 {
		fs.logger.Warn("上一个区块中没有有效的EIP-1559矿工费，使用默认费率")
		return "2000000000", nil // 默认2 Gwei
	}

	// 计算中位数
	medianPriorityFee := fs.calculateMedian(priorityFees)
	// fs.logger.Infof("基于上一个区块(%d)的%d笔EIP-1559交易，计算得出Max Priority Fee: %s Gwei",
	// 	previousBlockNumber, len(priorityFees), medianPriorityFee.String())

	return medianPriorityFee.String(), nil
}

// calculateMedian 计算中位数
func (fs *FeeScheduler) calculateMedian(values []*big.Int) *big.Int {
	if len(values) == 0 {
		return big.NewInt(2000000000) // 默认2 Gwei
	}

	// 排序
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i].Cmp(values[j]) > 0 {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	// 返回中位数
	mid := len(values) / 2
	if len(values)%2 == 0 {
		// 偶数个元素，返回中间两个的平均值
		sum := new(big.Int).Add(values[mid-1], values[mid])
		return sum.Div(sum, big.NewInt(2))
	}
	return values[mid]
}

// getDefaultETHFeeData 获取默认ETH费率数据
func (fs *FeeScheduler) getDefaultETHFeeData() *FeeLevels {
	baseFee := "20000000000"       // 20 Gwei in wei
	maxPriorityFee := "2000000000" // 2 Gwei in wei
	maxFee := fs.calculateMaxFee(baseFee, maxPriorityFee)
	congestion := "normal"

	return &FeeLevels{
		Slow: FeeData{
			Chain:             "eth",
			BaseFee:           baseFee,
			MaxPriorityFee:    fs.calculateSlowPriorityFee(maxPriorityFee),
			MaxFee:            fs.calculateMaxFee(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "eth",
			BaseFee:           baseFee,
			MaxPriorityFee:    maxPriorityFee,
			MaxFee:            maxFee,
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, maxPriorityFee),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "eth",
			BaseFee:           baseFee,
			MaxPriorityFee:    fs.calculateFastPriorityFee(maxPriorityFee),
			MaxFee:            fs.calculateMaxFee(baseFee, fs.calculateFastPriorityFee(maxPriorityFee)),
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, fs.calculateFastPriorityFee(maxPriorityFee)),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
	}
}

// getBaseFee 获取Base Fee (简化实现)
func (fs *FeeScheduler) getBaseFee(ctx context.Context, chain string) (string, error) {
	if chain != "eth" {
		return "0", nil
	}
	// 使用默认值
	return "20000000000", nil // 20 Gwei in wei
}

// calculateMaxPriorityFee 计算Max Priority Fee
func (fs *FeeScheduler) calculateMaxPriorityFee(ctx context.Context, chain string) (string, error) {
	if chain != "eth" {
		return "2000000000", nil // 默认2 Gwei
	}

	// 获取最近几个区块的矿工费数据
	// 这里简化实现，实际应该分析历史区块的矿工费
	// 返回一个合理的默认值
	return "2000000000", nil // 2 Gwei
}

// calculateMaxFee 计算Max Fee
func (fs *FeeScheduler) calculateMaxFee(baseFee, maxPriorityFee string) string {
	baseFeeBig, _ := new(big.Int).SetString(baseFee, 10)
	priorityFeeBig, _ := new(big.Int).SetString(maxPriorityFee, 10)

	// Max Fee = (Base Fee + Max Priority Fee) * 1.1 (冗余10%)
	sum := new(big.Int).Add(baseFeeBig, priorityFeeBig)
	redundancy := new(big.Int).Mul(sum, big.NewInt(11)) // 1.1 * 10 = 11
	redundancy.Div(redundancy, big.NewInt(10))          // 除以10得到1.1倍

	return redundancy.String()
}

// calculateSlowPriorityFee 计算慢速优先费
func (fs *FeeScheduler) calculateSlowPriorityFee(normalPriorityFee string) string {
	priorityFeeBig, _ := new(big.Int).SetString(normalPriorityFee, 10)
	// 慢速：正常优先费的50%
	slow := new(big.Int).Mul(priorityFeeBig, big.NewInt(50))
	slow.Div(slow, big.NewInt(100))
	return slow.String()
}

// calculateFastPriorityFee 计算快速优先费
func (fs *FeeScheduler) calculateFastPriorityFee(normalPriorityFee string) string {
	priorityFeeBig, _ := new(big.Int).SetString(normalPriorityFee, 10)
	// 快速：正常优先费的200%
	fast := new(big.Int).Mul(priorityFeeBig, big.NewInt(200))
	fast.Div(fast, big.NewInt(100))
	return fast.String()
}

// calculateLegacyGasPrice 计算Legacy Gas Price
func (fs *FeeScheduler) calculateLegacyGasPrice(baseFee, maxPriorityFee string) string {
	baseFeeBig, _ := new(big.Int).SetString(baseFee, 10)
	priorityFeeBig, _ := new(big.Int).SetString(maxPriorityFee, 10)

	// Legacy Gas Price = Base Fee + Max Priority Fee
	sum := new(big.Int).Add(baseFeeBig, priorityFeeBig)
	return sum.String()
}

// calculateNetworkCongestion 计算网络拥堵状态
func (fs *FeeScheduler) calculateNetworkCongestion(baseFee string) string {
	baseFeeBig, _ := new(big.Int).SetString(baseFee, 10)

	// 转换为Gwei进行比较
	baseFeeGwei := new(big.Int).Div(baseFeeBig, big.NewInt(1000000000)) // 除以10^9

	// 根据base fee判断拥堵状态
	if baseFeeGwei.Cmp(big.NewInt(50)) > 0 {
		return "high" // 高拥堵
	} else if baseFeeGwei.Cmp(big.NewInt(20)) > 0 {
		return "medium" // 中等拥堵
	} else {
		return "low" // 低拥堵
	}
}

// broadcastFeeData 广播费率数据
func (fs *FeeScheduler) broadcastFeeData(chain string, feeLevels *FeeLevels) {
	// 更新缓存
	fs.lastFeeData[chain] = feeLevels

	if fs.wsHandler == nil {
		return
	}

	// 通过接口调用WebSocket广播方法
	fs.wsHandler.BroadcastFeeEvent(chain, feeLevels)
}

// formatFeeForLog 格式化费率用于日志
func (fs *FeeScheduler) formatFeeForLog(feeWei string) string {
	feeBig, ok := new(big.Int).SetString(feeWei, 10)
	if !ok {
		return "0"
	}

	// 转换为Gwei
	gwei := new(big.Int).Div(feeBig, big.NewInt(1000000000))
	return gwei.String() + " Gwei"
}

// GetETHFeeData 公开方法，用于测试
func (fs *FeeScheduler) GetETHFeeData(ctx context.Context) (*FeeLevels, error) {
	return fs.getETHFeeData(ctx)
}

// GetLastFeeData 获取上一次推送的费率数据
func (fs *FeeScheduler) GetLastFeeData(chain string) *FeeLevels {
	if feeData, exists := fs.lastFeeData[chain]; exists {
		return feeData
	}
	return nil
}

// GetAllLastFeeData 获取所有链的上一次推送的费率数据
func (fs *FeeScheduler) GetAllLastFeeData() map[string]*FeeLevels {
	return fs.lastFeeData
}

// getLatestBTCBlock 获取最新BTC区块
func (fs *FeeScheduler) getLatestBTCBlock(ctx context.Context) (*BTCBlockInfo, error) {
	// 使用RPCClientManager获取最新区块号
	blockNumber, err := fs.rpcManager.GetBlockNumber(ctx, "btc")
	if err != nil {
		return nil, fmt.Errorf("获取BTC区块号失败: %v", err)
	}

	// 获取区块详情
	blockInterface, err := fs.rpcManager.GetBlockByNumber(ctx, "btc", big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("获取BTC区块详情失败: %v", err)
	}

	// 将interface{}转换为map[string]interface{}
	block, ok := blockInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的区块数据格式")
	}

	// 解析区块信息
	hash, _ := block["hash"].(string)
	timestamp, _ := block["time"].(float64)
	txs, _ := block["tx"].([]interface{})

	return &BTCBlockInfo{
		Number:    blockNumber,
		Hash:      hash,
		Timestamp: uint64(timestamp),
		TxCount:   len(txs),
	}, nil
}

// getBTCMempoolFeeRates 获取BTC Mempool中的费率数据
func (fs *FeeScheduler) getBTCMempoolFeeRates(ctx context.Context) ([]float64, error) {
	// 获取BTC故障转移管理器
	fo, exists := fs.rpcManager.GetBTCFailover("btc")
	if !exists {
		return nil, fmt.Errorf("未找到BTC故障转移管理器")
	}

	// 获取Mempool中的交易
	mempoolTxs, err := fo.GetMempoolTransactions(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取Mempool交易失败: %w", err)
	}

	var feeRates []float64
	for _, tx := range mempoolTxs {
		if tx.FeeRate > 0 {
			feeRates = append(feeRates, tx.FeeRate)
		}
	}

	fs.logger.Infof("从Mempool获取到%d笔交易的费率数据", len(feeRates))
	return feeRates, nil
}

// calculateBTCFeeRatesFromMempool 基于Mempool数据计算费率
func (fs *FeeScheduler) calculateBTCFeeRatesFromMempool(feeRates []float64) (*BTCFeeRates, error) {
	if len(feeRates) == 0 {
		return &BTCFeeRates{
			Slow:   "10.0",
			Normal: "20.0",
			Fast:   "50.0",
		}, nil
	}

	// 计算不同等级的费率（一位小数）
	slowRate := fs.calculateBTCPercentileFloat(feeRates, 20)
	normalRate := fs.calculateBTCPercentileFloat(feeRates, 50)
	fastRate := fs.calculateBTCPercentileFloat(feeRates, 80)

	fs.logger.Infof("基于Mempool计算BTC费率: 慢速=%.1f sat/vB，普通=%.1f sat/vB，快速=%.1f sat/vB",
		slowRate, normalRate, fastRate)

	return &BTCFeeRates{
		Slow:   fmt.Sprintf("%.1f", slowRate),
		Normal: fmt.Sprintf("%.1f", normalRate),
		Fast:   fmt.Sprintf("%.1f", fastRate),
	}, nil
}

// calculateBTCFeeRatesFromHistory 基于Mempool和历史数据计算BTC费率
func (fs *FeeScheduler) calculateBTCFeeRatesFromHistory(ctx context.Context) (*BTCFeeRates, error) {
	// 使用Mempool数据（实时性最好）
	mempoolRates, err := fs.getBTCMempoolFeeRates(ctx)
	if err != nil {
		fs.logger.Warnf("获取Mempool费率失败: %v", err)
	}
	// 使用Mempool数据计算费率
	return fs.calculateBTCFeeRatesFromMempool(mempoolRates)
}

// calculateBTCPercentileFloat 计算百分位（浮点）
func (fs *FeeScheduler) calculateBTCPercentileFloat(feeRates []float64, percentile int) float64 {
	if len(feeRates) == 0 {
		return 20.0
	}
	// 简单排序
	for i := 0; i < len(feeRates)-1; i++ {
		for j := i + 1; j < len(feeRates); j++ {
			if feeRates[i] > feeRates[j] {
				feeRates[i], feeRates[j] = feeRates[j], feeRates[i]
			}
		}
	}
	idx := (len(feeRates) * percentile) / 100
	if idx >= len(feeRates) {
		idx = len(feeRates) - 1
	}
	// 一位小数
	v := feeRates[idx]
	return math.Round(v*10) / 10
}

// calculateBTCNetworkCongestion 计算BTC网络拥堵状态
func (fs *FeeScheduler) calculateBTCNetworkCongestion(normalRate string) string {
	rate, err := fmt.Sscanf(normalRate, "%f", new(float64))
	if err != nil {
		return "normal"
	}

	// 根据费率判断拥堵状态
	if rate > 100 {
		return "high" // 高拥堵
	} else if rate > 50 {
		return "medium" // 中等拥堵
	} else {
		return "low" // 低拥堵
	}
}

// getDefaultBTCFeeData 获取默认BTC费率数据
func (fs *FeeScheduler) getDefaultBTCFeeData() *FeeLevels {
	return &FeeLevels{
		Slow: FeeData{
			Chain:             "btc",
			BaseFee:           "10", // 10 sat/vB
			MaxPriorityFee:    "10",
			MaxFee:            "10",
			GasPrice:          "10",
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: "normal",
		},
		Normal: FeeData{
			Chain:             "btc",
			BaseFee:           "20", // 20 sat/vB
			MaxPriorityFee:    "20",
			MaxFee:            "20",
			GasPrice:          "20",
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: "normal",
		},
		Fast: FeeData{
			Chain:             "btc",
			BaseFee:           "50", // 50 sat/vB
			MaxPriorityFee:    "50",
			MaxFee:            "50",
			GasPrice:          "50",
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: "normal",
		},
	}
}

// Stop 停止费率调度器
func (fs *FeeScheduler) Stop() {
	if fs.rpcManager != nil {
		fs.rpcManager.Close()
	}
	fs.logger.Info("费率调度器已关闭")
}
