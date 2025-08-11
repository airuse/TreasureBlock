package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// TransactionHandler 交易处理器
type TransactionHandler struct {
	txService services.TransactionService
}

// NewTransactionHandler 创建交易处理器
func NewTransactionHandler(txService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txService: txService,
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
