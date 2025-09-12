package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// HomeHandler 首页处理器
type HomeHandler struct {
	blockService       services.BlockService
	transactionService services.TransactionService
	statsService       services.StatsService
}

// NewHomeHandler 创建首页处理器
func NewHomeHandler(
	blockService services.BlockService,
	transactionService services.TransactionService,
	statsService services.StatsService,
) *HomeHandler {
	return &HomeHandler{
		blockService:       blockService,
		transactionService: transactionService,
		statsService:       statsService,
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
	if chain != "eth" && chain != "btc" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不支持的链类型，仅支持 eth 和 btc",
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
		if count, err := h.statsService.GetTotalBlockCount(ctx, chain); err == nil {
			resultChan <- result{key: "totalBlocks", value: count}
		} else {
			resultChan <- result{key: "totalBlocks", value: int64(0), err: err}
		}
	}()

	go func() {
		// 2. 总交易数
		if count, err := h.statsService.GetTotalTransactionCount(ctx, chain); err == nil {
			resultChan <- result{key: "totalTransactions", value: count}
		} else {
			resultChan <- result{key: "totalTransactions", value: int64(0), err: err}
		}
	}()

	go func() {
		// 3. 最新区块的Base Fee（仅ETH）
		if chain == "eth" {
			if baseFee, err := h.statsService.GetLatestBaseFee(ctx, chain); err == nil {
				resultChan <- result{key: "baseFee", value: baseFee}
			} else {
				resultChan <- result{key: "baseFee", value: int64(0), err: err}
			}
		} else {
			resultChan <- result{key: "baseFee", value: int64(0)}
		}
	}()

	go func() {
		// 4. 10分钟内的交易量（节省计算资源）
		if volume, err := h.statsService.GetDailyVolume(ctx, chain, 10*time.Minute); err == nil {
			resultChan <- result{key: "dailyVolume", value: volume}
		} else {
			resultChan <- result{key: "dailyVolume", value: float64(0), err: err}
		}
	}()

	go func() {
		// 5. 10分钟内的平均Gas价格（仅ETH）
		if chain == "eth" {
			if gasPrice, err := h.statsService.GetAverageGasPrice(ctx, chain, 10*time.Minute); err == nil {
				resultChan <- result{key: "avgGasPrice", value: gasPrice}
			} else {
				resultChan <- result{key: "avgGasPrice", value: int64(0), err: err}
			}
		} else {
			resultChan <- result{key: "avgGasPrice", value: int64(0)}
		}
	}()

	go func() {
		// 6. 10分钟内的平均出块时间（节省计算资源）
		if blockTime, err := h.statsService.GetAverageBlockTime(ctx, chain, 10*time.Minute); err == nil {
			resultChan <- result{key: "avgBlockTime", value: blockTime}
		} else {
			resultChan <- result{key: "avgBlockTime", value: float64(0), err: err}
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

// getLatestTransactions 获取最新交易
func (h *HomeHandler) getLatestTransactions(ctx context.Context, chain string, limit int) ([]gin.H, error) {
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
