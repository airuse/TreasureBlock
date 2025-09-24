package services

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
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
	lastFeeData map[string]*FeeLevels // key: chain (eth, btc, bsc), value: 费率数据
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

	// ETH每10秒更新一次，BTC每30秒更新一次，BSC每5秒更新一次，SOL每10秒更新一次
	ethTicker := time.NewTicker(10 * time.Second)
	btcTicker := time.NewTicker(30 * time.Second)
	bscTicker := time.NewTicker(5 * time.Second)
	solTicker := time.NewTicker(3 * time.Second)
	defer ethTicker.Stop()
	defer btcTicker.Stop()
	defer bscTicker.Stop()
	defer solTicker.Stop()

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
		case <-bscTicker.C:
			// 更新BSC费率
			fs.updateBSCFeeData(ctx)
		case <-solTicker.C:
			// 更新SOL费率
			fs.updateSOLFeeData(ctx)
		}
	}
}

// updateFeeData 更新费率数据
func (fs *FeeScheduler) updateFeeData(ctx context.Context) {
	// 更新ETH费率
	fs.updateETHFeeData(ctx)
	// 更新BTC费率
	fs.updateBTCFeeData(ctx)
	// 更新BSC费率
	fs.updateBSCFeeData(ctx)
	// 更新SOL费率
	fs.updateSOLFeeData(ctx)
}

// updateSOLFeeData 更新SOL费率数据
func (fs *FeeScheduler) updateSOLFeeData(ctx context.Context) {
	// 获取SOL费率数据
	feeData, err := fs.getSOLFeeData(ctx)
	if err != nil {
		fs.logger.WithField("chain", "sol").WithError(err).Error("获取SOL费率数据失败，使用默认值")
		feeData = fs.getDefaultSOLFeeData()
	}

	// 广播费率更新
	fs.broadcastFeeData("sol", feeData)

	fs.logger.WithField("chain", "sol").Debug("SOL费率数据已更新")
}

// getSOLFeeData 从RPC获取SOL费率数据
func (fs *FeeScheduler) getSOLFeeData(ctx context.Context) (*FeeLevels, error) {
	if fs.rpcManager == nil {
		return nil, fmt.Errorf("RPC管理器未初始化")
	}

	// 获取SOL RPC客户端
	solClient, err := fs.rpcManager.GetSolanaClient("sol")
	if err != nil {
		return nil, fmt.Errorf("获取SOL RPC客户端失败: %w", err)
	}

	// 获取最近的优先费用数据
	prioritizationFees, err := solClient.GetRecentPrioritizationFees(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取优先费用失败: %w", err)
	}

	// 计算不同速度档位的费用
	slowFee, normalFee, fastFee := fs.calculateSOLFeeLevels(prioritizationFees)

	// 获取网络拥堵状态
	congestion := fs.calculateSOLCongestion(prioritizationFees)

	return &FeeLevels{
		Slow: FeeData{
			Chain:             "sol",
			BaseFee:           "5000", // SOL基础费用 5000 lamports
			MaxPriorityFee:    slowFee,
			MaxFee:            slowFee,
			GasPrice:          slowFee,
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "sol",
			BaseFee:           "5000",
			MaxPriorityFee:    normalFee,
			MaxFee:            normalFee,
			GasPrice:          normalFee,
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "sol",
			BaseFee:           "5000",
			MaxPriorityFee:    fastFee,
			MaxFee:            fastFee,
			GasPrice:          fastFee,
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
	}, nil
}

// calculateSOLFeeLevels 计算SOL不同速度档位的费用
func (fs *FeeScheduler) calculateSOLFeeLevels(fees []utils.PrioritizationFeeItem) (slow, normal, fast string) {
	if len(fees) == 0 {
		// 默认费用
		return "0", "0", "1000"
	}

	// 统计0优先费的占比
	zeroFeeCount := 0
	var nonZeroFees []uint64

	for _, fee := range fees {
		if fee.PrioritizationFee == 0 {
			zeroFeeCount++
		} else {
			nonZeroFees = append(nonZeroFees, fee.PrioritizationFee)
		}
	}

	zeroFeeRatio := float64(zeroFeeCount) / float64(len(fees))

	// 如果0优先费占比很高，大部分交易都能用0优先费成功
	if zeroFeeRatio >= 0.8 {
		// 大部分都是0，只有少数需要优先费
		if len(nonZeroFees) > 0 {
			sort.Slice(nonZeroFees, func(i, j int) bool { return nonZeroFees[i] < nonZeroFees[j] })
			// 取非零费用的25%分位数作为fast
			fastIdx := int(float64(len(nonZeroFees)) * 0.25)
			if fastIdx >= len(nonZeroFees) {
				fastIdx = len(nonZeroFees) - 1
			}
			return "0", "0", strconv.FormatUint(nonZeroFees[fastIdx], 10)
		}
		return "0", "0", "1000"
	}

	// 如果0优先费占比在50%-80%之间，slow和normal可以用0，fast需要一些优先费
	if zeroFeeRatio >= 0.5 {
		if len(nonZeroFees) > 0 {
			sort.Slice(nonZeroFees, func(i, j int) bool { return nonZeroFees[i] < nonZeroFees[j] })
			// 取非零费用的50%分位数作为fast
			fastIdx := int(float64(len(nonZeroFees)) * 0.5)
			if fastIdx >= len(nonZeroFees) {
				fastIdx = len(nonZeroFees) - 1
			}
			return "0", "0", strconv.FormatUint(nonZeroFees[fastIdx], 10)
		}
		return "0", "0", "5000"
	}

	// 如果0优先费占比低于50%，需要根据实际分布计算
	// 排序所有费用
	sort.Slice(fees, func(i, j int) bool { return fees[i].PrioritizationFee < fees[j].PrioritizationFee })

	// 计算百分位数
	slowIdx := int(float64(len(fees)) * 0.25)   // 25th percentile
	normalIdx := int(float64(len(fees)) * 0.50) // 50th percentile
	fastIdx := int(float64(len(fees)) * 0.75)   // 75th percentile

	if slowIdx >= len(fees) {
		slowIdx = len(fees) - 1
	}
	if normalIdx >= len(fees) {
		normalIdx = len(fees) - 1
	}
	if fastIdx >= len(fees) {
		fastIdx = len(fees) - 1
	}

	slow = strconv.FormatUint(fees[slowIdx].PrioritizationFee, 10)
	normal = strconv.FormatUint(fees[normalIdx].PrioritizationFee, 10)
	fast = strconv.FormatUint(fees[fastIdx].PrioritizationFee, 10)

	return slow, normal, fast
}

// calculateSOLCongestion 计算SOL网络拥堵状态
func (fs *FeeScheduler) calculateSOLCongestion(fees []utils.PrioritizationFeeItem) string {
	if len(fees) == 0 {
		return "normal"
	}

	// 统计0优先费的占比
	zeroFeeCount := 0
	var nonZeroFees []uint64

	for _, fee := range fees {
		if fee.PrioritizationFee == 0 {
			zeroFeeCount++
		} else {
			nonZeroFees = append(nonZeroFees, fee.PrioritizationFee)
		}
	}

	zeroFeeRatio := float64(zeroFeeCount) / float64(len(fees))

	// 如果0优先费占比超过70%，说明网络不拥堵，大部分交易都能用0优先费成功
	if zeroFeeRatio >= 0.7 {
		return "low"
	}

	// 如果0优先费占比在30%-70%之间，需要结合非零费用的分布来判断
	if zeroFeeRatio >= 0.3 {
		// 计算非零费用的中位数
		if len(nonZeroFees) == 0 {
			return "low"
		}

		// 排序非零费用
		sort.Slice(nonZeroFees, func(i, j int) bool { return nonZeroFees[i] < nonZeroFees[j] })
		medianIdx := len(nonZeroFees) / 2
		medianFee := nonZeroFees[medianIdx]

		// 根据中位数判断拥堵状态
		if medianFee < 5000 {
			return "low"
		} else if medianFee < 20000 {
			return "normal"
		} else {
			return "high"
		}
	}

	// 如果0优先费占比低于30%，说明网络确实拥堵，大部分交易都需要支付优先费
	// 计算所有费用的中位数
	allFees := make([]uint64, len(fees))
	for i, fee := range fees {
		allFees[i] = fee.PrioritizationFee
	}
	sort.Slice(allFees, func(i, j int) bool { return allFees[i] < allFees[j] })
	medianIdx := len(allFees) / 2
	medianFee := allFees[medianIdx]

	if medianFee < 10000 {
		return "normal"
	} else {
		return "high"
	}
}

// getDefaultSOLFeeData 获取默认SOL费率数据（占位）
func (fs *FeeScheduler) getDefaultSOLFeeData() *FeeLevels {
	congestion := "normal"
	return &FeeLevels{
		Slow: FeeData{
			Chain:             "sol",
			BaseFee:           "0",
			MaxPriorityFee:    "0",
			MaxFee:            "0",
			GasPrice:          "0",
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "sol",
			BaseFee:           "0",
			MaxPriorityFee:    "0",
			MaxFee:            "0",
			GasPrice:          "0",
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "sol",
			BaseFee:           "0",
			MaxPriorityFee:    "0",
			MaxFee:            "0",
			GasPrice:          "0",
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
	}
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

// updateBSCFeeData 更新BSC费率数据
func (fs *FeeScheduler) updateBSCFeeData(ctx context.Context) {
	bscFeeData, err := fs.getBSCFeeData(ctx)
	if err != nil {
		fs.logger.Errorf("获取BSC费率失败: %v", err)
	} else {
		fs.broadcastFeeData("bsc", bscFeeData)
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
		fs.logger.Warnf("获取Base Fee失败: %v", err)
		baseFee = "0" // 20 Gwei in wei
	}

	// 获取历史区块的矿工费数据来计算合理的max priority fee
	maxPriorityFee, err := fs.calculateMaxPriorityFeeFromHistory(ctx, "eth", latestBlock.Number)
	if err != nil {
		fs.logger.Warnf("计算Max Priority Fee失败: %v", err)
		maxPriorityFee = "0"
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

// getBSCFeeData 获取BSC费率数据
func (fs *FeeScheduler) getBSCFeeData(ctx context.Context) (*FeeLevels, error) {
	// 获取最新区块信息
	latestBlock, err := fs.getLatestBlock(ctx, "bsc")
	if err != nil {
		fs.logger.Errorf("获取最新BSC区块失败: %v", err)
		return fs.getDefaultBSCFeeData(), nil
	}

	// 获取base fee
	baseFee, err := fs.getBaseFeeFromBlock(ctx, latestBlock)
	if err != nil {
		fs.logger.Warnf("获取BSC Base Fee失败: %v", err)
		baseFee = "0" // 5 Gwei in wei (BSC通常比ETH便宜)
	}

	// 获取历史区块的矿工费数据来计算合理的max priority fee
	maxPriorityFee, err := fs.calculateMaxPriorityFeeFromHistory(ctx, "bsc", latestBlock.Number)
	if err != nil {
		fs.logger.Warnf("计算BSC Max Priority Fee失败: %v", err)
		maxPriorityFee = "0"
	}

	// 计算max fee
	maxFee := fs.calculateMaxFee(baseFee, maxPriorityFee)

	// 计算网络拥堵状态
	congestion := fs.calculateNetworkCongestion(baseFee)

	// 创建费率等级
	feeLevels := &FeeLevels{
		Slow: FeeData{
			Chain:             "bsc",
			BaseFee:           baseFee,
			MaxPriorityFee:    fs.calculateSlowPriorityFee(maxPriorityFee),
			MaxFee:            fs.calculateMaxFee(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "bsc",
			BaseFee:           baseFee,
			MaxPriorityFee:    maxPriorityFee,
			MaxFee:            maxFee,
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, maxPriorityFee),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       latestBlock.Number,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "bsc",
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

	if chain == "bsc" {
		block, ok := blockInterface.(*types.Block)
		if !ok {
			return nil, fmt.Errorf("无效的BSC区块数据格式")
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
func (fs *FeeScheduler) calculateMaxPriorityFeeFromHistory(ctx context.Context, chain string, currentBlockNumber uint64) (string, error) {
	// 只分析上一个区块，获取最新的费率信息
	if currentBlockNumber == 0 {
		return "0", nil
	}

	previousBlockNumber := currentBlockNumber - 1
	blockInterface, err := fs.rpcManager.GetBlockByNumber(ctx, chain, big.NewInt(int64(previousBlockNumber)))
	if err != nil {
		fs.logger.Warnf("获取上一个区块失败，使用默认费率: %v", err)
		return "0", nil
	}

	// 将interface{}转换为ETH区块类型
	block, ok := blockInterface.(*types.Block)
	if !ok {
		fs.logger.Warnf("无效的%s区块数据格式，使用默认费率", chain)
		return "0", nil
	}

	// 收集所有交易的priority fee（EIP-1559直接取GasTipCap，Legacy取GasPrice - BaseFee）
	var allPriorityFees []*big.Int
	var totalTxs, eip1559Txs, legacyTxs int

	// 获取当前区块的base fee
	baseFee := block.BaseFee()
	if baseFee == nil {
		baseFee = big.NewInt(0)
	}

	for _, tx := range block.Transactions() {
		totalTxs++
		if tx.Type() == 2 { // EIP-1559 交易
			eip1559Txs++
			if tx.GasTipCap() != nil && tx.GasTipCap().Cmp(big.NewInt(0)) > 0 {
				allPriorityFees = append(allPriorityFees, tx.GasTipCap())
			}
		} else { // Legacy 交易
			legacyTxs++
			if tx.GasPrice() != nil && tx.GasPrice().Cmp(big.NewInt(0)) > 0 {
				// Legacy的priority fee = gasPrice - baseFee
				priorityFee := new(big.Int).Sub(tx.GasPrice(), baseFee)
				if priorityFee.Cmp(big.NewInt(0)) > 0 {
					allPriorityFees = append(allPriorityFees, priorityFee)
				}
			}
		}
	}

	// 计算所有priority fee的中位数
	var medianPriorityFee string
	if len(allPriorityFees) > 0 {
		median := fs.calculateMedian(allPriorityFees)
		medianPriorityFee = median.String()
	} else {
		medianPriorityFee = "0"
	}

	return medianPriorityFee, nil
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

// getDefaultBSCFeeData 获取默认BSC费率数据
func (fs *FeeScheduler) getDefaultBSCFeeData() *FeeLevels {
	baseFee := "5000000000"        // 5 Gwei in wei (BSC通常比ETH便宜)
	maxPriorityFee := "1000000000" // 1 Gwei in wei
	maxFee := fs.calculateMaxFee(baseFee, maxPriorityFee)
	congestion := "normal"

	return &FeeLevels{
		Slow: FeeData{
			Chain:             "bsc",
			BaseFee:           baseFee,
			MaxPriorityFee:    fs.calculateSlowPriorityFee(maxPriorityFee),
			MaxFee:            fs.calculateMaxFee(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, fs.calculateSlowPriorityFee(maxPriorityFee)),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Normal: FeeData{
			Chain:             "bsc",
			BaseFee:           baseFee,
			MaxPriorityFee:    maxPriorityFee,
			MaxFee:            maxFee,
			GasPrice:          fs.calculateLegacyGasPrice(baseFee, maxPriorityFee),
			LastUpdated:       time.Now().Unix(),
			BlockNumber:       0,
			NetworkCongestion: congestion,
		},
		Fast: FeeData{
			Chain:             "bsc",
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

// Stop 停止费率调度器
func (fs *FeeScheduler) Stop() {
	if fs.rpcManager != nil {
		fs.rpcManager.Close()
	}
	fs.logger.Info("费率调度器已关闭")
}
