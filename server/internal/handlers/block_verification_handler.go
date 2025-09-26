package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/middleware"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// BlockVerificationHandler 区块验证处理器
type BlockVerificationHandler struct {
	verificationService  services.BlockVerificationService
	earningsService      services.EarningsService
	contractParseService services.ContractParseService
	btcUTXOService       services.BTCUTXOService
	transactionService   services.TransactionService
	solService           services.SolService
	// 缓存每条链最后验证通过的区块高度
	heightCache      map[string]uint64
	heightCacheMutex sync.RWMutex
	// 可选：用于定期触发后台刷新/清理
	cacheUpdatedAt map[string]time.Time
}

// NewBlockVerificationHandler 创建区块验证处理器
func NewBlockVerificationHandler(verificationService services.BlockVerificationService, earningsService services.EarningsService, contractParseService services.ContractParseService, btcUTXOService services.BTCUTXOService, transactionService services.TransactionService, solService services.SolService) *BlockVerificationHandler {
	return &BlockVerificationHandler{
		verificationService:  verificationService,
		earningsService:      earningsService,
		contractParseService: contractParseService,
		btcUTXOService:       btcUTXOService,
		transactionService:   transactionService,
		solService:           solService,
		heightCache:          make(map[string]uint64),
		cacheUpdatedAt:       make(map[string]time.Time),
	}
}

// VerifyBlock 验证区块接口
func (h *BlockVerificationHandler) VerifyBlock(c *gin.Context) {
	blockIDStr := c.Param("blockID")
	if blockIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "block ID is required",
		})
		return
	}

	blockID, err := strconv.ParseUint(blockIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid block ID format",
		})
		return
	}

	// 执行区块验证
	result, err := h.verificationService.VerifyBlock(c.Request.Context(), blockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "BLOCK_VERIFICATION_FAILED",
			"details": err.Error(),
		})
		return
	}

	if !result.IsValid {
		// 如果验证不通过，则返回错误，让客户端结束
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "BLOCK_VERIFICATION_FAILED",
			"details": result.Details,
		})
		return
	}
	// 获取区块，非BTC直接跳过
	block, err := h.verificationService.GetBlockByID(c.Request.Context(), blockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "BLOCK_VERIFICATION_FAILED",
			"details": err.Error(),
		})
		return
	}
	if block != nil && block.Chain == "btc" {
		if err := h.ParseBTCBlockTransactions(c.Request.Context(), blockID); err != nil {
			logrus.Errorf("ParseBTCBlockTransactions failed for block %d: %v", blockID, err)
		}
	} else if block != nil && block.Chain == "eth" {
		h.contractParseService.ParseBlockTransactions(c.Request.Context(), blockID)
	} else if block != nil && block.Chain == "sol" {
		// 解析 Sol 交易：根据维护的 Program 规则解析，保存事件/指令，不兼容数据入扩展表

		// details, err := h.solService.GetDetailsByBlockID(c.Request.Context(), blockID)
		// if err != nil {
		// 	logrus.Errorf("GetDetailsByBlockID failed for block %d: %v", blockID, err)
		// } else if len(details) > 0 {
		// 	programs, err := h.solService.GetAllPrograms(c.Request.Context())
		// 	if err != nil {
		// 		logrus.Errorf("GetAllPrograms failed: %v", err)
		// 	} else {
		// 		// 构建映射
		// 		programMap := make(map[string]*models.SolProgram, len(programs))
		// 		for _, p := range programs {
		// 			programMap[p.ProgramID] = p
		// 		}
		// 		for _, d := range details {
		// 			if d == nil {
		// 				continue
		// 			}
		// 			req, extra, err := h.solService.ParseUsingProgramRules(c.Request.Context(), d, programMap)
		// 			if err != nil {
		// 				logrus.Errorf("ParseUsingProgramRules failed tx %s: %v", d.TxID, err)
		// 				continue
		// 			}
		// 			// 转换额外数据
		// 			var extras []models.SolParsedExtra
		// 			if len(extra) > 0 {
		// 				for k, v := range extra {
		// 					bs, _ := json.Marshal(v)
		// 					extras = append(extras, models.SolParsedExtra{TxID: d.TxID, ProgramID: k, IsInner: false, Data: models.JSONText(string(bs))})
		// 				}
		// 			}
		// 			if err := h.solService.SaveArtifacts(c.Request.Context(), d.TxID, d.BlockID, d.Slot, req.Events, req.Instructions, extras); err != nil {
		// 				logrus.Errorf("SaveArtifacts failed tx %s: %v", d.TxID, err)
		// 			}
		// 		}
		// 	}
		// }
	}

	// 验证通过需要吧数据库 block 表的 verification_status 更新为 1
	h.verificationService.UpdateBlockVerificationStatus(c.Request.Context(), blockID, true, "验证通过")

	// 同步更新本地缓存的最后验证高度（由后端掌控）
	if block != nil {
		h.heightCacheMutex.Lock()
		if cur, ok := h.heightCache[block.Chain]; !ok || block.Height > cur {
			h.heightCache[block.Chain] = block.Height
			h.cacheUpdatedAt[block.Chain] = time.Now()
		}
		h.heightCacheMutex.Unlock()
	}

	// 获取用户ID
	userID, exists := middleware.GetUserIDFromContext(c)

	if !exists {
		logrus.Errorf("Failed to get user ID from context for block verification earnings")
	} else {
		// 获取区块信息用于计算收益
		block, err := h.verificationService.GetBlockByID(c.Request.Context(), blockID)
		if err != nil {
			logrus.Errorf("Failed to get block info for earnings calculation: %v", err)
		} else {
			blockInfo := &dto.BlockEarningsInfo{
				BlockID:          blockID,
				BlockHeight:      block.Height,
				Chain:            block.Chain,
				TransactionCount: int64(block.TransactionCount),
				EarningsAmount:   int64(block.TransactionCount), // 1个交易对应1个T币
			}

			if err := h.earningsService.ProcessBlockVerificationEarnings(c.Request.Context(), uint64(userID), blockInfo); err != nil {
				logrus.Errorf("Failed to process block verification earnings: %v", err)
				// 收益处理失败不影响区块验证成功的响应
			} else {
				logrus.Infof("Successfully processed earnings for user %d, block %d, earned %d T-coins",
					userID, blockID, result.Transactions)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"block_id":                    result.BlockID,
			"is_valid":                    result.IsValid,
			"reason":                      result.Reason,
			"details":                     result.Details,
			"transactions":                result.Transactions,
			"receipts":                    result.Receipts,
			"verification_status_updated": true,
			"earnings_processed":          exists, // 表示是否处理了收益
		},
		"message": "区块验证完成",
	})
}

// GetLastVerifiedBlockHeight 获取最后一个验证通过的区块高度
func (h *BlockVerificationHandler) GetLastVerifiedBlockHeight(c *gin.Context) {
	chain := c.Query("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "chain parameter is required",
		})
		return
	}

	// 先尝试命中本地缓存（后端掌控缓存）
	h.heightCacheMutex.RLock()
	cachedHeight, ok := h.heightCache[chain]
	h.heightCacheMutex.RUnlock()

	if ok {
		// 异步触发超时处理，尽快返回响应
		go func(ch string, last uint64) {
			if err := h.verificationService.HandleTimeoutBlocks(context.Background(), ch, last+1); err != nil {
				logrus.Errorf("HandleTimeoutBlocks error: %v", err)
			}
		}(chain, cachedHeight)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"chain":  chain,
				"height": cachedHeight,
			},
			"message": "获取成功(缓存)",
		})
		return
	}

	// 未命中缓存则查询服务层并回填缓存
	height, err := h.verificationService.GetLastVerifiedBlockHeight(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到验证通过的区块",
			"details": err.Error(),
		})
		return
	}

	h.heightCacheMutex.Lock()
	h.heightCache[chain] = height
	h.cacheUpdatedAt[chain] = time.Now()
	h.heightCacheMutex.Unlock()

	// 异步执行超时处理
	go func(ch string, last uint64) {
		if err := h.verificationService.HandleTimeoutBlocks(context.Background(), ch, last+1); err != nil {
			logrus.Errorf("HandleTimeoutBlocks error: %v", err)
		}
	}(chain, height)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chain":  chain,
			"height": height,
		},
		"message": "获取成功",
	})
}

// ParseBTCBlockTransactions 解析BTC区块内交易的 vin/vout，维护本地UTXO表
func (h *BlockVerificationHandler) ParseBTCBlockTransactions(ctx context.Context, blockID uint64) error {

	txs, err := h.transactionService.GetTransactionsByBlockID(ctx, blockID)
	if err != nil {
		return err
	}
	if len(txs) == 0 {
		return nil
	}

	// 定义解析结构
	type scriptPubKey struct {
		Address   string   `json:"address"`
		Addresses []string `json:"addresses"`
		ASM       string   `json:"asm"`
		Desc      string   `json:"desc"`
		Hex       string   `json:"hex"`
		Type      string   `json:"type"`
	}
	type parsedVout struct {
		N            int          `json:"n"`
		ScriptPubKey scriptPubKey `json:"scriptPubKey"`
		Value        float64      `json:"value"`
	}
	type prevoutInfo struct {
		Generated    bool         `json:"generated"`
		Height       uint64       `json:"height"`
		ScriptPubKey scriptPubKey `json:"scriptPubKey"`
		Value        float64      `json:"value"`
	}
	type parsedVin struct {
		Coinbase    string       `json:"coinbase"`
		Txid        string       `json:"txid"`
		Vout        int          `json:"vout"`
		Prevout     *prevoutInfo `json:"prevout"`
		Sequence    uint32       `json:"sequence"`
		Txinwitness []string     `json:"txinwitness"`
	}

	// 辅助：BTC -> satoshi（避免负数，做四舍五入）
	toSatoshi := func(v float64) int64 {
		if v <= 0 {
			return 0
		}
		return int64(v*1e8 + 0.5)
	}

	var utxoBatch []*models.BTCUTXO

	for _, tx := range txs {
		if tx == nil || tx.Vout == "" {
			continue
		}

		// 解析 vout 生成 UTXO
		var vouts []parsedVout
		if err := json.Unmarshal([]byte(tx.Vout), &vouts); err != nil {
			logrus.Errorf("unmarshal vout failed for tx %s: %v", tx.TxID, err)
		} else {
			for _, o := range vouts {
				address := o.ScriptPubKey.Address
				if address == "" && len(o.ScriptPubKey.Addresses) > 0 {
					address = o.ScriptPubKey.Addresses[0]
				}
				utxo := &models.BTCUTXO{
					Chain:        "btc",
					TxID:         tx.TxID,
					VoutIndex:    uint32(o.N),
					BlockHeight:  tx.Height,
					BlockID:      tx.BlockID,
					Address:      address,
					ScriptPubKey: o.ScriptPubKey.Hex,
					ScriptType:   o.ScriptPubKey.Type,
					IsCoinbase:   false, // 输出是否来自coinbase由vin判断；保留false，这里不强制
					ValueSatoshi: toSatoshi(o.Value),
				}
				utxoBatch = append(utxoBatch, utxo)
			}
		}

		// 解析 vin 标记已花费的 UTXO
		if tx.Vin != "" {
			var vins []parsedVin
			if err := json.Unmarshal([]byte(tx.Vin), &vins); err != nil {
				logrus.Errorf("unmarshal vin failed for tx %s: %v", tx.TxID, err)
			} else {
				for vinIdx, in := range vins {
					// coinbase 输入不花费任何UTXO
					if in.Coinbase != "" || (in.Prevout != nil && in.Prevout.Generated) {
						continue
					}
					if in.Txid == "" {
						continue
					}
					if err := h.btcUTXOService.MarkSpent(ctx, "btc", in.Txid, uint32(in.Vout), tx.TxID, uint32(vinIdx), tx.Height); err != nil {
						logrus.Errorf("mark spent failed prev %s:%d by %s vin %d: %v", in.Txid, in.Vout, tx.TxID, vinIdx, err)
					}
				}
			}
		}
	}

	if len(utxoBatch) > 0 {
		if err := h.btcUTXOService.UpsertOutputs(ctx, utxoBatch); err != nil {
			return err
		}
	}

	return nil
}
