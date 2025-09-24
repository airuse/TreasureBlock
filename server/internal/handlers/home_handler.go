package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// HomeHandler 首页处理器
type HomeHandler struct {
	blockService       services.BlockService
	transactionService services.TransactionService
	statsService       services.StatsService
	solService         services.SolService
}

// NewHomeHandler 创建首页处理器
func NewHomeHandler(
	blockService services.BlockService,
	transactionService services.TransactionService,
	statsService services.StatsService,
	solService services.SolService,
) *HomeHandler {
	return &HomeHandler{
		blockService:       blockService,
		transactionService: transactionService,
		statsService:       statsService,
		solService:         solService,
	}
}

// GetHomeStats 获取首页统计数据
func (h *HomeHandler) GetHomeStats(c *gin.Context) {
	chain := c.Query("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "链类型参数缺失",
		})
		return
	}

	// 验证链类型
	if chain != "eth" && chain != "btc" && chain != "bsc" && chain != "sol" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不支持的链类型，仅支持 eth、btc、bsc 和 sol",
		})
		return
	}

	ctx := c.Request.Context()

	// 获取概览数据
	overview, err := h.getOverviewStats(ctx, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取概览数据失败: " + err.Error(),
		})
		return
	}

	// 获取最新区块
	latestBlocks, err := h.getLatestBlocks(ctx, chain, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取最新区块失败: " + err.Error(),
		})
		return
	}

	// 获取最新交易
	latestTransactions, err := h.getLatestTransactions(ctx, chain, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取最新交易失败: " + err.Error(),
		})
		return
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"data": gin.H{
			"overview":           overview,
			"latestBlocks":       latestBlocks,
			"latestTransactions": latestTransactions,
		},
		"message": "成功获取首页统计数据",
	}

	c.JSON(http.StatusOK, response)
}

// GetBtcHomeStats 获取比特币首页统计数据（固定链为 btc）
func (h *HomeHandler) GetBtcHomeStats(c *gin.Context) {
	chain := "btc"

	ctx := c.Request.Context()

	// 获取概览数据
	overview, err := h.getOverviewStats(ctx, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取概览数据失败: " + err.Error(),
		})
		return
	}

	// 获取最新区块
	latestBlocks, err := h.getLatestBlocks(ctx, chain, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取最新区块失败: " + err.Error(),
		})
		return
	}

	// 获取最新交易
	latestTransactions, err := h.getLatestTransactions(ctx, chain, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取最新交易失败: " + err.Error(),
		})
		return
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"data": gin.H{
			"overview":           overview,
			"latestBlocks":       latestBlocks,
			"latestTransactions": latestTransactions,
		},
		"message": "成功获取比特币首页统计数据",
	}

	c.JSON(http.StatusOK, response)
}

// getOverviewStats 获取概览统计数据
func (h *HomeHandler) getOverviewStats(ctx context.Context, chain string) (gin.H, error) {
	// 使用并发查询优化性能
	type result struct {
		key   string
		value interface{}
		err   error
	}

	resultChan := make(chan result, 7) // 缓冲通道，增加到7个查询

	// 并发执行所有统计查询
	go func() {
		// 1. 总区块数
		if chain == "sol" {
			// Solana使用slot数作为区块数
			if count, err := h.statsService.GetTotalSolanaSlotCount(ctx); err == nil {
				resultChan <- result{key: "totalBlocks", value: count}
			} else {
				resultChan <- result{key: "totalBlocks", value: int64(0), err: err}
			}
		} else {
			if count, err := h.statsService.GetTotalBlockCount(ctx, chain); err == nil {
				resultChan <- result{key: "totalBlocks", value: count}
			} else {
				resultChan <- result{key: "totalBlocks", value: int64(0), err: err}
			}
		}
	}()

	go func() {
		// 2. 总交易数
		if chain == "sol" {
			// Solana使用专门的交易数统计
			if count, err := h.statsService.GetTotalSolanaTransactionCount(ctx); err == nil {
				resultChan <- result{key: "totalTransactions", value: count}
			} else {
				resultChan <- result{key: "totalTransactions", value: int64(0), err: err}
			}
		} else {
			if count, err := h.statsService.GetTotalTransactionCount(ctx, chain); err == nil {
				resultChan <- result{key: "totalTransactions", value: count}
			} else {
				resultChan <- result{key: "totalTransactions", value: int64(0), err: err}
			}
		}
	}()

	go func() {
		// 3. 最新区块的Base Fee（ETH和BSC都支持）
		if chain == "eth" || chain == "bsc" {
			if baseFee, err := h.statsService.GetLatestBaseFee(ctx, chain); err == nil {
				resultChan <- result{key: "baseFee", value: baseFee}
			} else {
				resultChan <- result{key: "baseFee", value: int64(0), err: err}
			}
		} else if chain == "sol" {
			// Solana使用平均费用作为Base Fee
			if avgFee, err := h.statsService.GetSolanaAverageFee(ctx, 10*time.Minute); err == nil {
				resultChan <- result{key: "baseFee", value: avgFee}
			} else {
				resultChan <- result{key: "baseFee", value: int64(5000), err: err} // 默认5000 lamports
			}
		} else {
			resultChan <- result{key: "baseFee", value: int64(0)}
		}
	}()

	go func() {
		// 4. 10分钟内的交易量（节省计算资源）
		if chain == "sol" {
			// Solana使用专门的交易量统计
			if volume, err := h.statsService.GetSolanaDailyVolume(ctx, 10*time.Minute); err == nil {
				resultChan <- result{key: "dailyVolume", value: volume}
			} else {
				resultChan <- result{key: "dailyVolume", value: float64(0), err: err}
			}
		} else {
			if volume, err := h.statsService.GetDailyVolume(ctx, chain, 10*time.Minute); err == nil {
				resultChan <- result{key: "dailyVolume", value: volume}
			} else {
				resultChan <- result{key: "dailyVolume", value: float64(0), err: err}
			}
		}
	}()

	go func() {
		// 5. 10分钟内的平均Gas价格（ETH和BSC都支持）
		if chain == "eth" || chain == "bsc" {
			if gasPrice, err := h.statsService.GetAverageGasPrice(ctx, chain, 10*time.Minute); err == nil {
				resultChan <- result{key: "avgGasPrice", value: gasPrice}
			} else {
				resultChan <- result{key: "avgGasPrice", value: int64(0), err: err}
			}
		} else if chain == "sol" {
			// Solana使用平均费用作为Gas价格
			if avgFee, err := h.statsService.GetSolanaAverageFee(ctx, 10*time.Minute); err == nil {
				resultChan <- result{key: "avgGasPrice", value: avgFee}
			} else {
				resultChan <- result{key: "avgGasPrice", value: int64(5000), err: err} // 默认5000 lamports
			}
		} else {
			resultChan <- result{key: "avgGasPrice", value: int64(0)}
		}
	}()

	go func() {
		// 6. 10分钟内的平均出块时间（节省计算资源）
		if chain == "sol" {
			// Solana使用slot时间
			if slotTime, err := h.statsService.GetSolanaAverageSlotTime(ctx, 10*time.Minute); err == nil {
				resultChan <- result{key: "avgBlockTime", value: slotTime}
			} else {
				resultChan <- result{key: "avgBlockTime", value: float64(0.4), err: err} // 默认400ms
			}
		} else {
			if blockTime, err := h.statsService.GetAverageBlockTime(ctx, chain, 10*time.Minute); err == nil {
				resultChan <- result{key: "avgBlockTime", value: blockTime}
			} else {
				resultChan <- result{key: "avgBlockTime", value: float64(0), err: err}
			}
		}
	}()

	go func() {
		// 7. 当前难度（两条链均可）
		if difficulty, err := h.statsService.GetCurrentDifficulty(ctx, chain); err == nil {
			resultChan <- result{key: "difficulty", value: difficulty}
		} else {
			resultChan <- result{key: "difficulty", value: int64(0), err: err}
		}
	}()

	// 收集所有结果
	results := make(map[string]interface{})
	var errors []error

	// 等待所有查询完成，设置超时时间
	timeout := time.After(5 * time.Second) // 5秒超时

	for i := 0; i < 7; i++ { // 增加到7个查询
		select {
		case result := <-resultChan:
			if result.err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", result.key, result.err))
			}
			results[result.key] = result.value
		case <-timeout:
			// 超时处理
			return nil, fmt.Errorf("查询超时，部分数据可能不完整")
		}
	}

	// 如果有错误，记录日志但不中断
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("Warning: %v", err)
		}
	}

	return gin.H{
		"totalBlocks":       results["totalBlocks"],
		"totalTransactions": results["totalTransactions"],
		"baseFee":           results["baseFee"],
		"activeAddresses":   0,
		"networkHashrate":   0,
		"dailyVolume":       results["dailyVolume"],
		"avgGasPrice":       results["avgGasPrice"],
		"avgBlockTime":      results["avgBlockTime"],
		"difficulty":        results["difficulty"],
		"avgFee":            0,
	}, nil
}

// getLatestBlocks 获取最新区块
func (h *HomeHandler) getLatestBlocks(ctx context.Context, chain string, limit int) ([]gin.H, error) {
	if chain == "sol" {
		// Solana使用slot作为区块
		return h.getLatestSolanaSlots(ctx, limit)
	}

	blocks, _, err := h.blockService.ListBlocks(ctx, 1, limit, chain)
	if err != nil {
		return nil, err
	}

	var result []gin.H
	for _, block := range blocks {
		result = append(result, gin.H{
			"height":             block.Height,
			"hash":               block.Hash,
			"timestamp":          block.Timestamp.UnixMilli(),
			"transactions_count": block.TransactionCount,
			"size":               block.Size,
			"miner":              block.Miner,
			"chain":              block.Chain,
		})
	}

	return result, nil
}

// getLatestSolanaSlots 获取最新Solana slot
func (h *HomeHandler) getLatestSolanaSlots(ctx context.Context, limit int) ([]gin.H, error) {
	// 获取最新的交易详情，按slot分组
	txDetails, _, err := h.solService.ListTxDetails(ctx, nil, 1, limit*10) // 获取更多数据用于分组
	if err != nil {
		return nil, err
	}

	// 如果没有交易数据，返回空数组
	if len(txDetails) == 0 {
		return []gin.H{}, nil
	}

	// 按slot分组，获取最新的几个slot
	slotMap := make(map[uint64]*struct {
		Slot      uint64
		TxCount   int
		Timestamp time.Time
		Blockhash string
	})

	for _, tx := range txDetails {
		if slotMap[tx.Slot] == nil {
			// 解析时间字符串
			ctime, _ := time.Parse(time.RFC3339, tx.Ctime)
			slotMap[tx.Slot] = &struct {
				Slot      uint64
				TxCount   int
				Timestamp time.Time
				Blockhash string
			}{
				Slot:      tx.Slot,
				TxCount:   0,
				Timestamp: ctime,
				Blockhash: tx.Blockhash,
			}
		}
		slotMap[tx.Slot].TxCount++
	}

	// 转换为切片并排序
	var slots []*struct {
		Slot      uint64
		TxCount   int
		Timestamp time.Time
		Blockhash string
	}
	for _, slot := range slotMap {
		slots = append(slots, slot)
	}

	// 按slot降序排序
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].Slot > slots[j].Slot
	})

	// 限制返回数量
	if len(slots) > limit {
		slots = slots[:limit]
	}

	var result []gin.H
	for _, slot := range slots {
		result = append(result, gin.H{
			"height":             slot.Slot,
			"hash":               slot.Blockhash,
			"timestamp":          slot.Timestamp.UnixMilli(),
			"transactions_count": slot.TxCount,
			"size":               0,  // Solana没有区块大小概念
			"miner":              "", // Solana没有矿工概念
			"chain":              "sol",
		})
	}

	return result, nil
}

// getLatestTransactions 获取最新交易
func (h *HomeHandler) getLatestTransactions(ctx context.Context, chain string, limit int) ([]gin.H, error) {
	if chain == "sol" {
		// Solana使用专门的交易详情
		return h.getLatestSolanaTransactions(ctx, limit)
	}

	// 使用新方法获取最新区块的前几条交易
	transactions, err := h.transactionService.GetLatestTransactions(ctx, chain, limit)
	if err != nil {
		return nil, err
	}

	var result []gin.H
	for _, tx := range transactions {
		result = append(result, gin.H{
			"hash":      tx.TxID,              // 使用TxID字段
			"timestamp": tx.Ctime.UnixMilli(), // 使用Ctime字段
			"amount":    tx.Amount,
			"from":      tx.AddressFrom,
			"to":        tx.AddressTo,
			"gas_price": tx.GasPrice,
			"gas_used":  tx.GasUsed,
			"chain":     tx.Chain,
			"height":    tx.Height, // 添加区块高度
		})
	}

	return result, nil
}

// getLatestSolanaTransactions 获取最新Solana交易
func (h *HomeHandler) getLatestSolanaTransactions(ctx context.Context, limit int) ([]gin.H, error) {
	// 获取最新的交易详情
	txDetails, _, err := h.solService.ListTxDetails(ctx, nil, 1, limit)
	if err != nil {
		return nil, err
	}

	// 如果没有交易数据，返回空数组
	if len(txDetails) == 0 {
		return []gin.H{}, nil
	}

	var result []gin.H
	for _, tx := range txDetails {
		// 解析时间字符串
		ctime, _ := time.Parse(time.RFC3339, tx.Ctime)
		result = append(result, gin.H{
			"hash":      tx.TxID,
			"timestamp": ctime.UnixMilli(),
			"amount":    fmt.Sprintf("%d", tx.Fee), // 使用费用作为金额
			"from":      "",                        // Solana交易没有简单的from/to概念
			"to":        "",
			"gas_price": tx.Fee, // 使用费用作为gas价格
			"gas_used":  tx.ComputeUnits,
			"chain":     "sol",
			"height":    tx.Slot, // 使用slot作为高度
		})
	}

	return result, nil
}

// GetSolHomeStats 获取Solana首页统计数据（固定链为 sol）
func (h *HomeHandler) GetSolHomeStats(c *gin.Context) {
	chain := "sol"

	ctx := c.Request.Context()

	// 获取概览数据
	overview, err := h.getOverviewStats(ctx, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取概览数据失败: " + err.Error(),
		})
		return
	}

	// 获取最新区块
	latestBlocks, err := h.getLatestBlocks(ctx, chain, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取最新区块失败: " + err.Error(),
		})
		return
	}

	// 获取最新交易
	latestTransactions, err := h.getLatestTransactions(ctx, chain, 3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取最新交易失败: " + err.Error(),
		})
		return
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"data": gin.H{
			"overview":           overview,
			"latestBlocks":       latestBlocks,
			"latestTransactions": latestTransactions,
		},
		"message": "成功获取Solana首页统计数据",
	}

	c.JSON(http.StatusOK, response)
}
