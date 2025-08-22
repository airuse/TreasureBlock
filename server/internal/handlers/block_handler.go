package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// BlockHandler 区块处理器
type BlockHandler struct {
	blockService services.BlockService
	wsHandler    *WebSocketHandler
}

// NewBlockHandler 创建区块处理器
func NewBlockHandler(blockService services.BlockService, wsHandler *WebSocketHandler) *BlockHandler {
	return &BlockHandler{
		blockService: blockService,
		wsHandler:    wsHandler,
	}
}

// GetBlockByHash 根据哈希获取区块
func (h *BlockHandler) GetBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "block hash is required",
		})
		return
	}

	block, err := h.blockService.GetBlockByHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewBlockResponse(block)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetBlockByHeight 根据高度获取区块
func (h *BlockHandler) GetBlockByHeight(c *gin.Context) {
	heightStr := c.Param("height")
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid block height",
		})
		return
	}

	block, err := h.blockService.GetBlockByHeight(c.Request.Context(), height)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewBlockResponse(block)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetLatestBlock 获取最新区块
func (h *BlockHandler) GetLatestBlock(c *gin.Context) {
	chain := c.Query("chain")
	if chain == "" {
		chain = "btc" // 默认BTC链
	}

	block, err := h.blockService.GetLatestBlock(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewBlockResponse(block)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListBlocks 分页查询区块列表
func (h *BlockHandler) ListBlocks(c *gin.Context) {
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

	blocks, total, err := h.blockService.ListBlocks(c.Request.Context(), page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := dto.NewBlockListResponse(blocks, total, page, pageSize)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListBlocksPublic 公开查询区块列表（限制总数据量为100条，支持分页）
func (h *BlockHandler) ListBlocksPublic(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	chain := c.Query("chain")

	// 解析分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 限制每页最大数量
	const maxPageSize = 50
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制总数据量为100条（不是分页后的100条）
	const maxTotalBlocks = 100

	// 计算实际的总页数
	actualTotalPages := (maxTotalBlocks + pageSize - 1) / pageSize

	// 限制最大页码（基于实际数据量）
	if page > actualTotalPages {
		page = actualTotalPages
	}

	// 调用服务层，但限制总数量为100条
	blocks, total, err := h.blockService.ListBlocks(c.Request.Context(), page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取区块信息失败",
		})
		return
	}

	// 限制返回的总数为100条
	if total > maxTotalBlocks {
		total = maxTotalBlocks
	}

	response := dto.NewBlockListResponse(blocks, total, page, pageSize)

	// 添加限制说明
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "公开查询限制：最多显示100条数据",
		"limits": gin.H{
			"max_total_blocks": maxTotalBlocks,
			"max_page_size":    maxPageSize,
			"max_pages":        actualTotalPages,
			"note":             "如需更多数据，请登录后使用完整API",
		},
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  actualTotalPages,
			"total_count":  total,
		},
	})
}

// GetBlockByHeightPublic 公开查询区块详情（根据高度）
func (h *BlockHandler) GetBlockByHeightPublic(c *gin.Context) {
	heightStr := c.Param("height")
	chain := c.Query("chain")

	if chain == "" {
		chain = "eth" // 默认ETH链
	}

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的区块高度",
		})
		return
	}

	// 限制查询范围（防止查询过高的区块）
	const maxHeight = 999999999
	if height > maxHeight {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "区块高度超出限制",
			"limits": gin.H{
				"max_height": maxHeight,
				"note":       "如需查询更高区块，请登录后使用完整API",
			},
		})
		return
	}

	block, err := h.blockService.GetBlockByHeight(c.Request.Context(), height)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "区块不存在或查询失败",
		})
		return
	}

	// 验证区块链匹配
	if block.Chain != chain {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "区块链不匹配",
		})
		return
	}

	response := dto.NewBlockResponse(block)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "公开查询成功，但功能有限制",
		"limits": gin.H{
			"note": "如需更多功能，请登录后使用完整API",
		},
	})
}

// SearchBlocksPublic 公开搜索区块（支持高度和哈希搜索）
func (h *BlockHandler) SearchBlocksPublic(c *gin.Context) {
	query := c.Query("query") // 搜索关键词（与前端保持一致）
	chain := c.Query("chain") // 区块链
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "搜索关键词不能为空",
		})
		return
	}

	if chain == "" {
		chain = "eth" // 默认ETH链
	}

	// 解析分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 限制每页最大数量
	const maxPageSize = 50
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制搜索结果总数
	const maxSearchResults = 100

	// 判断搜索类型
	var blocks []*models.Block
	var total int64

	// 尝试解析为区块高度
	if height, err := strconv.ParseUint(query, 10, 64); err == nil {
		// 搜索特定高度的区块
		block, err := h.blockService.GetBlockByHeight(c.Request.Context(), height)
		if err == nil && block.Chain == chain {
			blocks = []*models.Block{block}
			total = 1
		} else {
			blocks = []*models.Block{}
			total = 0
		}
	} else {
		// 搜索区块哈希
		block, err := h.blockService.GetBlockByHash(c.Request.Context(), query)
		if err == nil && block.Chain == chain {
			blocks = []*models.Block{block}
			total = 1
		} else {
			blocks = []*models.Block{}
			total = 0
		}
	}

	// 限制返回结果数量
	if total > maxSearchResults {
		total = maxSearchResults
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	if page > int(totalPages) {
		page = int(totalPages)
	}

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(blocks) {
		blocks = []*models.Block{}
	} else if end > len(blocks) {
		blocks = blocks[start:]
	} else {
		blocks = blocks[start:end]
	}

	response := dto.NewBlockListResponse(blocks, total, page, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "搜索完成",
		"search": gin.H{
			"query":       query,
			"chain":       chain,
			"total_found": total,
		},
		"limits": gin.H{
			"max_search_results": maxSearchResults,
			"max_page_size":      maxPageSize,
			"note":               "如需更多搜索结果，请登录后使用完整API",
		},
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
	})
}

// SearchBlocks 认证用户搜索区块（支持高度和哈希搜索，无限制）
func (h *BlockHandler) SearchBlocks(c *gin.Context) {
	query := c.Query("query") // 搜索关键词（与前端保持一致）
	chain := c.Query("chain") // 区块链
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "搜索关键词不能为空",
		})
		return
	}

	if chain == "" {
		chain = "eth" // 默认ETH链
	}

	// 解析分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 认证用户无分页大小限制
	const maxPageSize = 1000
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 判断搜索类型
	var blocks []*models.Block
	var total int64

	// 尝试解析为区块高度
	if height, err := strconv.ParseUint(query, 10, 64); err == nil {
		// 搜索特定高度的区块
		block, err := h.blockService.GetBlockByHeight(c.Request.Context(), height)
		if err == nil && block.Chain == chain {
			blocks = []*models.Block{block}
			total = 1
		} else {
			blocks = []*models.Block{}
			total = 0
		}
	} else {
		// 搜索区块哈希
		block, err := h.blockService.GetBlockByHash(c.Request.Context(), query)
		if err == nil && block.Chain == chain {
			blocks = []*models.Block{block}
			total = 1
		} else {
			blocks = []*models.Block{}
			total = 0
		}
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	if page > int(totalPages) {
		page = int(totalPages)
	}

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(blocks) {
		blocks = []*models.Block{}
	} else if end > len(blocks) {
		blocks = blocks[start:]
	} else {
		blocks = blocks[start:end]
	}

	response := dto.NewBlockListResponse(blocks, total, page, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "搜索完成",
		"search": gin.H{
			"query":       query,
			"chain":       chain,
			"total_found": total,
		},
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
	})
}

func (h *BlockHandler) CreateBlock(c *gin.Context) {
	// 1. 从请求体获取区块数据并验证
	var req dto.CreateBlockRequest
	if err := utils.ValidateAndBind(c, &req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// 2. 将DTO转换为模型
	block := req.ToModel()

	// 3. 调用服务层处理业务逻辑
	err := h.blockService.CreateBlock(c.Request.Context(), block)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 4. 返回响应DTO
	response := dto.NewBlockResponse(block)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "block created successfully",
	})

	// 5. 使用新的WebSocket广播方法，发送区块事件
	// 根据链类型选择对应的ChainType
	var chainType ChainType
	switch block.Chain {
	case "eth":
		chainType = ChainTypeETH
	case "btc":
		chainType = ChainTypeBTC
	default:
		chainType = ChainTypeETH // 默认以太坊
	}

	// 广播区块事件
	h.wsHandler.BroadcastBlockEvent(chainType, response)
}
