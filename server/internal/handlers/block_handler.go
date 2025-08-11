package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// BlockHandler 区块处理器
type BlockHandler struct {
	blockService services.BlockService
}

// NewBlockHandler 创建区块处理器
func NewBlockHandler(blockService services.BlockService) *BlockHandler {
	return &BlockHandler{
		blockService: blockService,
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

// ListBlocksPublic 公开查询区块列表（支持分页，但限制数量，防止洪水攻击）
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
	const maxPageSize = 20
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// 限制最大页码（防止深度翻页）
	const maxPage = 10
	if page > maxPage {
		page = maxPage
	}

	blocks, total, err := h.blockService.ListBlocks(c.Request.Context(), page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取区块信息失败",
		})
		return
	}

	response := dto.NewBlockListResponse(blocks, total, page, pageSize)

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
}
