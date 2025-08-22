package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// TransactionHandler 交易处理器
type TransactionHandler struct {
	txService      services.TransactionService
	receiptService services.TransactionReceiptService
}

// NewTransactionHandler 创建交易处理器
func NewTransactionHandler(txService services.TransactionService, receiptService services.TransactionReceiptService) *TransactionHandler {
	return &TransactionHandler{
		txService:      txService,
		receiptService: receiptService,
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

	txs, total, err := h.txService.GetTransactionsByAddress(c.Request.Context(), address, page, pageSize)
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

	if req.Receipt != nil {
		receiptData := map[string]interface{}{"tx_hash": req.TxID, "chain": req.Chain}

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
		if req.Receipt.CumulativeGasUsed != nil {
			if v := convertToUint64(req.Receipt.CumulativeGasUsed); v != nil {
				receiptData["cumulative_gas_used"] = *v
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

	receipt, err := h.receiptService.GetTransactionReceiptByHash(c.Request.Context(), Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    receipt,
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
