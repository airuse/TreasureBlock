package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// TransactionHandler 交易处理器
type TransactionHandler struct {
	txService                services.TransactionService
	receiptService           services.TransactionReceiptService
	parserConfigRepo         repository.ParserConfigRepository
	blockVerificationService services.BlockVerificationService
	contractParseService     services.ContractParseService
	coinConfigService        services.CoinConfigService
	userAddressService       services.UserAddressService
}

// NewTransactionHandler 创建交易处理器
func NewTransactionHandler(
	txService services.TransactionService,
	receiptService services.TransactionReceiptService,
	parserConfigRepo repository.ParserConfigRepository,
	blockVerificationService services.BlockVerificationService,
	contractParseService services.ContractParseService,
	coinConfigService services.CoinConfigService,
	userAddressService services.UserAddressService,
) *TransactionHandler {
	return &TransactionHandler{
		txService:                txService,
		receiptService:           receiptService,
		parserConfigRepo:         parserConfigRepo,
		blockVerificationService: blockVerificationService,
		contractParseService:     contractParseService,
		coinConfigService:        coinConfigService,
		userAddressService:       userAddressService,
	}
}

// GetTransactionByHash 根据哈希获取交易
func (h *TransactionHandler) GetTransactionByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transaction hash is required",
		})
		return
	}

	tx, err := h.txService.GetTransactionByHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewTransactionResponse(tx)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetTransactionsByAddress 根据地址获取交易列表
func (h *TransactionHandler) GetTransactionsByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address is required",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取链类型参数
	chain := c.Query("chain")

	txs, total, err := h.txService.GetTransactionsByAddress(c.Request.Context(), address, page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewTransactionListResponse(txs, total, page, pageSize)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetTransactionsByBlockHash 根据区块哈希获取交易列表
func (h *TransactionHandler) GetTransactionsByBlockHash(c *gin.Context) {
	blockHash := c.Param("blockHash")
	if blockHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "block hash is required",
		})
		return
	}

	txs, err := h.txService.GetTransactionsByBlockHash(c.Request.Context(), blockHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	txResponses := make([]*dto.TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = dto.NewTransactionResponse(tx)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transactions": txResponses,
			"count":        len(txs),
		},
	})
}

// ListTransactions 分页查询交易列表
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	chain := c.Query("chain")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	txs, total, err := h.txService.ListTransactions(c.Request.Context(), page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewTransactionListResponse(txs, total, page, pageSize)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListTransactionsPublic 公开查询交易列表（支持分页，但限制数量，防止洪水攻击）
func (h *TransactionHandler) ListTransactionsPublic(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")
	chain := c.Query("chain")

	// 解析分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 50
	}

	// 限制每页最大数量
	const maxPageSize = 1000
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制最大页码（防止深度翻页）
	const maxPage = 20
	if page > maxPage {
		page = maxPage
	}

	txs, total, err := h.txService.ListTransactions(c.Request.Context(), page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取交易信息失败",
		})
		return
	}

	response := dto.NewTransactionListResponse(txs, total, page, pageSize)

	// 添加限制说明
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "公开查询支持分页，但有限制",
		"limits": gin.H{
			"max_page_size": maxPageSize,
			"max_page":      maxPage,
			"note":          "如需更多数据，请登录后使用完整API",
		},
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  (total + int64(pageSize) - 1) / int64(pageSize),
			"total_count":  total,
		},
	})
}

// CreateTransaction 创建交易记录
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请求参数验证失败", "details": err.Error()})
		return
	}

	var req dto.CreateTransactionRequest
	if err := mapToStruct(rawData, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "交易数据转换失败", "details": err.Error()})
		return
	}

	// 验证超时检查
	if req.BlockID != nil {
		if err := h.checkBlockVerificationTimeout(c.Request.Context(), *req.BlockID); err != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success":  false,
				"error":    "BLOCK_TRANSACTION_FAILED",
				"details":  err.Error(),
				"block_id": *req.BlockID,
			})
			return
		}
	}

	if req.Receipt != nil {
		receiptData := map[string]interface{}{"tx_hash": req.TxID, "chain": req.Chain}

		// 设置 block_id 字段
		if req.BlockID != nil {
			receiptData["block_id"] = *req.BlockID
		}

		// type -> tx_type (convert interface{} to uint8)
		if req.Receipt.Type != nil {
			if v := convertToUint64(req.Receipt.Type); v != nil {
				if *v <= 255 {
					receiptData["tx_type"] = uint8(*v)
				}
			}
		}
		if req.Receipt.Root != nil {
			receiptData["post_state"] = *req.Receipt.Root
		}
		if req.Receipt.Status != nil {
			if v := convertToUint64(req.Receipt.Status); v != nil {
				receiptData["status"] = *v
			}
		}
		// 处理新增字段：EffectiveGasPrice, BlobGasUsed, BlobGasPrice
		if req.Receipt.EffectiveGasPrice != nil {
			if v, ok := req.Receipt.EffectiveGasPrice.(string); ok {
				receiptData["effective_gas_price"] = v
			}
		}
		if req.Receipt.BlobGasUsed != nil {
			if v := convertToUint64(req.Receipt.BlobGasUsed); v != nil {
				receiptData["blob_gas_used"] = *v
			}
		}
		if req.Receipt.BlobGasPrice != nil {
			if v, ok := req.Receipt.BlobGasPrice.(string); ok {
				receiptData["blob_gas_price"] = v
			}
		}
		if req.Receipt.LogsBloom != nil {
			receiptData["bloom"] = *req.Receipt.LogsBloom
		}
		if req.Receipt.Logs != nil {
			receiptData["logs_data"] = req.Receipt.Logs
		}
		if req.Receipt.ContractAddress != nil {
			receiptData["contract_address"] = *req.Receipt.ContractAddress
		}
		if req.Receipt.GasUsed != nil {
			if v := convertToUint64(req.Receipt.GasUsed); v != nil {
				receiptData["gas_used"] = *v
			}
		}
		if req.Receipt.BlockHash != nil {
			receiptData["block_hash"] = *req.Receipt.BlockHash
		}
		if req.Receipt.BlockNumber != nil {
			if v := convertToUint64(req.Receipt.BlockNumber); v != nil {
				receiptData["block_number"] = *v
			}
		}
		if req.Receipt.TransactionIndex != nil {
			if v := convertToUint(req.Receipt.TransactionIndex); v != nil {
				receiptData["transaction_index"] = *v
			}
		}
		if req.Receipt.CumulativeGasUsed != nil {
			if v := convertToUint64(req.Receipt.CumulativeGasUsed); v != nil {
				receiptData["cumulative_gas_used"] = *v
			}
		}
		if req.Receipt.ReceiptType != nil {
			if v := convertToUint64(req.Receipt.ReceiptType); v != nil {
				receiptData["receipt_type"] = *v
			}
		}

		if err := h.receiptService.CreateTransactionReceipt(c.Request.Context(), receiptData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建交易凭证失败", "details": err.Error()})
			return
		}

	}

	tx := req.ToModel()
	if err := h.txService.CreateTransaction(c.Request.Context(), tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建交易失败", "details": err.Error()})
		return
	}

	response := dto.NewTransactionResponse(tx)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": response, "message": "交易和凭证创建成功"})
}

// CreateTransactionsBatch 批量创建交易记录
func (h *TransactionHandler) CreateTransactionsBatch(c *gin.Context) {
	var rawData map[string]interface{}
	if err := c.ShouldBindJSON(&rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请求参数验证失败", "details": err.Error()})
		return
	}

	// 解析批量请求
	transactionsData, ok := rawData["transactions"].([]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "transactions字段格式错误，应为数组"})
		return
	}

	if len(transactionsData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "transactions数组不能为空"})
		return
	}
	// fmt.Printf("批量数据一共为: %d\n", len(transactionsData))

	// 限制批量数量，防止请求过大
	const maxBatchSize = 4000
	if len(transactionsData) > maxBatchSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": fmt.Sprintf("批量数量超过限制，最大支持%d条", maxBatchSize)})
		return
	}

	// 批量处理交易
	var successCount int
	var failedCount int
	var errors []string
	txModels := make([]*models.Transaction, 0)
	receiptsData := make([]map[string]interface{}, 0)

	coinConfigs, err := h.coinConfigService.GetAllCoinConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "获取币种配置失败", "details": err.Error()})
		return
	}
	coinConfigMap := make(map[string]*models.CoinConfig)
	for _, cc := range coinConfigs {
		coinConfigMap[cc.ContractAddr] = cc
	}
	for i, txData := range transactionsData {
		txMap, ok := txData.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("第%d条交易数据格式错误", i+1))
			failedCount++
			continue
		}

		// 转换为CreateTransactionRequest
		var req dto.CreateTransactionRequest
		if err := mapToStruct(txMap, &req); err != nil {
			errors = append(errors, fmt.Sprintf("第%d条交易数据转换失败: %v", i+1, err))
			failedCount++
			continue
		}

		// 验证超时检查
		if req.BlockID != nil {
			if err := h.checkBlockVerificationTimeout(c.Request.Context(), *req.BlockID); err != nil {
				errors = append(errors, fmt.Sprintf("第%d条交易区块验证超时: %v", i+1, err))
				failedCount++
				continue
			}
		}

		// 处理交易凭证
		if req.Receipt != nil {
			receiptData := map[string]interface{}{"tx_hash": req.TxID, "chain": req.Chain}

			// 设置 block_id 字段
			if req.BlockID != nil {
				receiptData["block_id"] = *req.BlockID
			}

			// type -> tx_type (convert interface{} to uint8)
			if req.Receipt.Type != nil {
				if v := convertToUint64(req.Receipt.Type); v != nil {
					if *v <= 255 {
						receiptData["tx_type"] = uint8(*v)
					}
				}
			}
			if req.Receipt.Root != nil {
				receiptData["post_state"] = *req.Receipt.Root
			}
			if req.Receipt.Status != nil {
				if v := convertToUint64(req.Receipt.Status); v != nil {
					receiptData["status"] = *v
				}
			}
			// 处理新增字段：EffectiveGasPrice, BlobGasUsed, BlobGasPrice
			if req.Receipt.EffectiveGasPrice != nil {
				if v, ok := req.Receipt.EffectiveGasPrice.(string); ok {
					receiptData["effective_gas_price"] = v
				}
			}
			if req.Receipt.BlobGasUsed != nil {
				if v := convertToUint64(req.Receipt.BlobGasUsed); v != nil {
					receiptData["blob_gas_used"] = *v
				}
			}
			if req.Receipt.BlobGasPrice != nil {
				if v, ok := req.Receipt.BlobGasPrice.(string); ok {
					receiptData["blob_gas_price"] = v
				}
			}
			if req.Receipt.LogsBloom != nil {
				receiptData["bloom"] = *req.Receipt.LogsBloom
			}
			if req.Receipt.Logs != nil {
				receiptData["logs_data"] = req.Receipt.Logs
			}

			if coinConfigMap[req.AddressTo] != nil {
				receiptData["contract_address"] = req.AddressTo
			}
			if req.Receipt.GasUsed != nil {
				if v := convertToUint64(req.Receipt.GasUsed); v != nil {
					receiptData["gas_used"] = *v
				}
			}
			if req.Receipt.BlockHash != nil {
				receiptData["block_hash"] = *req.Receipt.BlockHash
			}
			if req.Receipt.BlockNumber != nil {
				if v := convertToUint64(req.Receipt.BlockNumber); v != nil {
					receiptData["block_number"] = *v
				}
			}
			if req.Receipt.TransactionIndex != nil {
				if v := convertToUint(req.Receipt.TransactionIndex); v != nil {
					receiptData["transaction_index"] = *v
				}
			}
			if req.Receipt.CumulativeGasUsed != nil {
				if v := convertToUint64(req.Receipt.CumulativeGasUsed); v != nil {
					receiptData["cumulative_gas_used"] = *v
				}
			}
			if req.Receipt.ReceiptType != nil {
				if v := convertToUint64(req.Receipt.ReceiptType); v != nil {
					receiptData["receipt_type"] = *v
				}
			}

			// 累积凭证，后续统一批量入库
			receiptsData = append(receiptsData, receiptData)
		}

		// 累积待批量创建的交易记录
		tx := req.ToModel()
		txModels = append(txModels, tx)
	}

	// 真正批量插入交易
	if len(txModels) > 0 {
		if err := h.txService.CreateTransactionsBatch(c.Request.Context(), txModels); err != nil {
			errors = append(errors, fmt.Sprintf("批量创建交易失败: %v", err))
			failedCount += len(txModels)
		} else {
			successCount += len(txModels)
		}
	}

	// 真正批量插入交易凭证
	if len(receiptsData) > 0 {
		if err := h.receiptService.CreateTransactionReceiptsBatch(c.Request.Context(), receiptsData); err != nil {
			errors = append(errors, fmt.Sprintf("批量创建交易凭证失败: %v", err))
		}
	}

	// 构建响应
	response := gin.H{
		"success":       true,
		"total_count":   len(transactionsData),
		"success_count": successCount,
		"failed_count":  failedCount,
	}

	if failedCount > 0 {
		response["errors"] = errors
		response["message"] = fmt.Sprintf("批量创建完成，成功%d条，失败%d条", successCount, failedCount)
	} else {
		response["message"] = fmt.Sprintf("批量创建成功，共%d条交易", successCount)
	}

	c.JSON(http.StatusCreated, response)

}

// checkBlockVerificationTimeout 检查区块验证是否超时
func (h *TransactionHandler) checkBlockVerificationTimeout(ctx context.Context, blockID uint64) error {
	// 获取区块信息
	block, err := h.blockVerificationService.GetBlockByID(ctx, blockID)
	if err != nil {
		return fmt.Errorf("failed to get block: %w", err)
	}

	// 检查区块是否已经超时
	if block.VerificationDeadline != nil && time.Now().After(*block.VerificationDeadline) {
		// 区块已超时，执行超时处理逻辑
		if err := h.blockVerificationService.HandleTimeoutBlocks(ctx, block.Chain, block.Height); err != nil {
			return fmt.Errorf("failed to handle timeout blocks: %w", err)
		}

		// 返回客户端需要的错误格式
		return fmt.Errorf("BLOCK_TRANSACTION_FAILED")
	}

	// 检查区块是否已经被标记为验证失败
	if block.IsVerified == 2 {
		return fmt.Errorf("BLOCK_TRANSACTION_FAILED")
	}

	return nil
}

// GetTransactionsByBlockHeight 根据区块高度获取交易列表（支持分页）
func (h *TransactionHandler) GetTransactionsByBlockHeight(c *gin.Context) {
	blockHeightStr := c.Param("blockHeight")
	if blockHeightStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "block height is required",
		})
		return
	}

	// 解析区块高度
	blockHeight, err := strconv.ParseUint(blockHeightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid block height format",
		})
		return
	}

	// 解析分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	chain := c.Query("chain")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 调用服务获取交易
	txs, total, err := h.txService.GetTransactionsByBlockHeight(c.Request.Context(), blockHeight, page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取区块交易失败",
			"details": err.Error(),
		})
		return
	}

	// 构建响应
	txResponses := make([]*dto.TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = dto.NewTransactionResponse(tx)
	}

	// 计算分页信息
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	response := gin.H{
		"transactions": txResponses,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
		"block_height": blockHeight,
		"chain":        chain,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取区块交易成功",
	})
}

// GetBTCTransactionsByBlockHeight 根据区块高度获取BTC交易列表
func (h *TransactionHandler) GetBTCTransactionsByBlockHeight(c *gin.Context) {
	blockHeightStr := c.Param("blockHeight")
	if blockHeightStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "block height is required",
		})
		return
	}

	// 解析区块高度
	blockHeight, err := strconv.ParseUint(blockHeightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid block height format",
		})
		return
	}

	// 解析分页参数（公开接口限制更严格）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")
	chain := c.Query("chain")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 50
	}

	// 限制每页最大数量
	const maxPageSize = 100
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制最大页码（防止深度翻页）
	const maxPage = 10
	if page > maxPage {
		page = maxPage
	}

	// 调用服务获取交易
	txs, total, err := h.txService.GetBTCTransactionsByBlockHeight(c.Request.Context(), blockHeight, page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取区块交易失败",
			"details": err.Error(),
		})
		return
	}

	// 构建响应
	txResponses := make([]*dto.TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = dto.NewTransactionResponse(tx)
	}

	// 计算分页信息
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	response := gin.H{
		"transactions": txResponses,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
		"block_height": blockHeight,
		"chain":        chain,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取区块交易成功（公开接口）",
		"limits": gin.H{
			"max_page_size": maxPageSize,
			"max_page":      maxPage,
			"note":          "如需更多数据，请登录后使用完整API",
		},
	})
}

// GetTransactionsByBlockHeightPublic 根据区块高度获取交易列表（公开接口，有限制）
func (h *TransactionHandler) GetTransactionsByBlockHeightPublic(c *gin.Context) {
	blockHeightStr := c.Param("blockHeight")
	if blockHeightStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "block height is required",
		})
		return
	}

	// 解析区块高度
	blockHeight, err := strconv.ParseUint(blockHeightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid block height format",
		})
		return
	}

	// 解析分页参数（公开接口限制更严格）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")
	chain := c.Query("chain")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 50
	}

	// 限制每页最大数量
	const maxPageSize = 100
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制最大页码（防止深度翻页）
	const maxPage = 10
	if page > maxPage {
		page = maxPage
	}

	// 调用服务获取交易
	txs, total, err := h.txService.GetTransactionsByBlockHeight(c.Request.Context(), blockHeight, page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取区块交易失败",
			"details": err.Error(),
		})
		return
	}

	// 构建响应
	txResponses := make([]*dto.TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = dto.NewTransactionResponse(tx)
	}

	// 计算分页信息
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	response := gin.H{
		"transactions": txResponses,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
		"block_height": blockHeight,
		"chain":        chain,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取区块交易成功（公开接口）",
		"limits": gin.H{
			"max_page_size": maxPageSize,
			"max_page":      maxPage,
			"note":          "如需更多数据，请登录后使用完整API",
		},
	})
}

// GetBTCTransactionsByBlockHeightPublic 根据区块高度获取BTC 交易列表（公开接口，有限制）
func (h *TransactionHandler) GetBTCTransactionsByBlockHeightPublic(c *gin.Context) {
	blockHeightStr := c.Param("blockHeight")
	if blockHeightStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "block height is required",
		})
		return
	}

	// 解析区块高度
	blockHeight, err := strconv.ParseUint(blockHeightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid block height format",
		})
		return
	}

	// 解析分页参数（公开接口限制更严格）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")
	chain := c.Query("chain")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 50
	}

	// 限制每页最大数量
	const maxPageSize = 100
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制最大页码（防止深度翻页）
	const maxPage = 10
	if page > maxPage {
		page = maxPage
	}

	// 调用服务获取交易
	txs, total, err := h.txService.GetBTCTransactionsByBlockHeight(c.Request.Context(), blockHeight, page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取区块交易失败",
			"details": err.Error(),
		})
		return
	}

	// 构建响应
	txResponses := make([]*dto.TransactionResponse, len(txs))
	for i, tx := range txs {
		txResponses[i] = dto.NewTransactionResponse(tx)
	}

	// 计算分页信息
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	response := gin.H{
		"transactions": txResponses,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
		"block_height": blockHeight,
		"chain":        chain,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取区块交易成功（公开接口）",
		"limits": gin.H{
			"max_page_size": maxPageSize,
			"max_page":      maxPage,
			"note":          "如需更多数据，请登录后使用完整API",
		},
	})
}

// GetTransactionReceiptByHash 根据交易哈希获取凭证
func (h *TransactionHandler) GetTransactionReceiptByHash(c *gin.Context) {
	Hash := c.Param("hash")
	if Hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "transaction hash is required",
		})
		return
	}

	// 获取交易凭证
	receipt, err := h.receiptService.GetTransactionReceiptByHash(c.Request.Context(), Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 获取关联的交易信息（用于获取输入数据）
	var tx *models.Transaction
	var parserConfigs []*models.ParserConfig
	if receipt != nil {
		// 获取交易
		tx, err = h.txService.GetTransactionByHash(c.Request.Context(), Hash)
		if err != nil {
			// 如果获取交易失败，记录警告但不影响凭证返回
			// fmt.Printf("Warning: Failed to get transaction for hash %s: %v\n", Hash, err)
		}

		// 如果获取到交易信息，尝试获取解析配置
		if tx != nil && tx.AddressTo != "" {
			// 直接使用parserConfigRepo查找解析配置
			configs, err := h.parserConfigRepo.GetParserConfigsByContract(c.Request.Context(), tx.AddressTo)
			if err != nil {
				// 如果获取解析配置失败，记录警告但不影响返回
				// fmt.Printf("Warning: Failed to get parser configs for contract %s: %v\n", tx.ContractAddr, err)
			} else {
				if len(configs) > 0 {
					parserConfigs = configs
				} else {
					configs, err = h.parserConfigRepo.GetParserConfigsByContract(c.Request.Context(), "*")
					if err != nil {
						// fmt.Printf("Warning: Failed to get parser configs for contract %s: %v\n", tx.AddressFrom, err)
					} else {
						parserConfigs = configs
					}
				}
			}
		}
	}

	// 使用新的DTO返回完整信息
	response := dto.NewTransactionReceiptResponse(receipt, tx, parserConfigs)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// mapToStruct 将map转换为struct
func mapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// convertToUint64 智能转换各种类型到uint64
func convertToUint64(value interface{}) *uint64 {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		if len(v) > 2 && v[:2] == "0x" {
			if num, err := strconv.ParseUint(v[2:], 16, 64); err == nil {
				return &num
			}
		}
		if num, err := strconv.ParseUint(v, 10, 64); err == nil {
			return &num
		}
	case float64:
		if v >= 0 && v <= float64(^uint64(0)) {
			num := uint64(v)
			return &num
		}
	case int:
		if v >= 0 {
			num := uint64(v)
			return &num
		}
	case int64:
		if v >= 0 {
			num := uint64(v)
			return &num
		}
	case uint64:
		return &v
	}
	return nil
}

// convertToUint 智能转换各种类型到uint
func convertToUint(value interface{}) *uint {
	if value == nil {
		return nil
	}
	if uint64Value := convertToUint64(value); uint64Value != nil {
		if *uint64Value <= uint64(^uint(0)) {
			u := uint(*uint64Value)
			return &u
		}
	}
	return nil
}
