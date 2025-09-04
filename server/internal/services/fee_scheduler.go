package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"blockChainBrowser/server/internal/interfaces"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/utils"

	"github.com/sirupsen/logrus"
)

// FeeScheduler 费率调度器
type FeeScheduler struct {
	rpcManager *utils.RPCClientManager
	baseCfgSvc BaseConfigService
	logger     *logrus.Logger
	wsHandler  interfaces.WebSocketBroadcaster // WebSocket广播接口
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
		rpcManager: utils.NewRPCClientManager(),
		baseCfgSvc: NewBaseConfigService(repository.NewBaseConfigRepository()),
		logger:     logrus.New(),
	}
}

// SetWebSocketHandler 设置WebSocket处理器
func (fs *FeeScheduler) SetWebSocketHandler(handler interfaces.WebSocketBroadcaster) {
	fs.wsHandler = handler
}

// Start 启动费率调度器
func (fs *FeeScheduler) Start(ctx context.Context) {
	fs.logger.Info("费率调度器已启动")

	// 每10秒更新一次费率
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	fs.updateFeeData(ctx)

	for {
		select {
		case <-ctx.Done():
			fs.logger.Info("费率调度器已停止")
			return
		case <-ticker.C:
			fs.updateFeeData(ctx)
		}
	}
}

// updateFeeData 更新费率数据
func (fs *FeeScheduler) updateFeeData(ctx context.Context) {
	// 更新ETH费率
	ethFeeData, err := fs.getETHFeeData(ctx)
	if err != nil {
		fs.logger.Errorf("获取ETH费率失败: %v", err)
	} else {
		fs.broadcastFeeData("eth", ethFeeData)
	}

	// 更新BTC费率（如果需要）
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
	// BTC费率计算逻辑
	// 这里可以根据需要实现BTC的费率计算
	// 暂时返回默认值
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
	}, nil
}

// BlockInfo 区块信息结构
type BlockInfo struct {
	Number    uint64
	BaseFee   *big.Int
	GasUsed   uint64
	GasLimit  uint64
	Timestamp uint64
}

// getLatestBlock 获取最新区块
func (fs *FeeScheduler) getLatestBlock(ctx context.Context, chain string) (*BlockInfo, error) {
	// 使用RPCClientManager获取最新区块号
	blockNumber, err := fs.rpcManager.GetBlockNumber(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("获取区块号失败: %v", err)
	}

	// 使用RPCClientManager获取区块详情
	block, err := fs.rpcManager.GetBlockByNumber(ctx, chain, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("获取区块详情失败: %v", err)
	}

	return &BlockInfo{
		Number:    blockNumber,
		BaseFee:   block.BaseFee(),
		GasUsed:   block.GasUsed(),
		GasLimit:  block.GasLimit(),
		Timestamp: block.Time(),
	}, nil
}

// getBaseFeeFromBlock 从区块获取Base Fee
func (fs *FeeScheduler) getBaseFeeFromBlock(ctx context.Context, block *BlockInfo) (string, error) {
	if block.BaseFee == nil {
		return "20000000000", nil // 默认20 Gwei
	}
	return block.BaseFee.String(), nil
}

// calculateMaxPriorityFeeFromHistory 从历史区块计算Max Priority Fee
func (fs *FeeScheduler) calculateMaxPriorityFeeFromHistory(ctx context.Context, currentBlockNumber uint64) (string, error) {
	// 获取最近几个区块的矿工费数据
	var priorityFees []*big.Int
	blockCount := 5 // 分析最近5个区块

	for i := uint64(0); i < uint64(blockCount) && currentBlockNumber > i; i++ {
		blockNumber := currentBlockNumber - i
		block, err := fs.rpcManager.GetBlockByNumber(ctx, "eth", big.NewInt(int64(blockNumber)))
		if err != nil {
			continue
		}

		// 分析区块中的交易，计算平均矿工费
		for _, tx := range block.Transactions() {
			if tx.Type() == 2 { // EIP-1559 交易
				if tx.GasTipCap() != nil {
					priorityFees = append(priorityFees, tx.GasTipCap())
				}
			}
		}
	}

	if len(priorityFees) == 0 {
		return "2000000000", nil // 默认2 Gwei
	}

	// 计算中位数
	medianPriorityFee := fs.calculateMedian(priorityFees)
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
	if fs.wsHandler == nil {
		return
	}

	fs.logger.Infof("广播%s费率数据: Slow=%s, Normal=%s, Fast=%s",
		chain,
		fs.formatFeeForLog(feeLevels.Slow.MaxFee),
		fs.formatFeeForLog(feeLevels.Normal.MaxFee),
		fs.formatFeeForLog(feeLevels.Fast.MaxFee))

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

// Stop 停止费率调度器
func (fs *FeeScheduler) Stop() {
	if fs.rpcManager != nil {
		fs.rpcManager.Close()
	}
	fs.logger.Info("费率调度器已关闭")
}
