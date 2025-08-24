package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// ParserConfigHandler 解析配置处理器
type ParserConfigHandler struct {
	parserConfigService services.ParserConfigService
}

// NewParserConfigHandler 创建解析配置处理器
func NewParserConfigHandler(parserConfigService services.ParserConfigService) *ParserConfigHandler {
	return &ParserConfigHandler{
		parserConfigService: parserConfigService,
	}
}

// CreateParserConfig 创建解析配置
func (h *ParserConfigHandler) CreateParserConfig(c *gin.Context) {
	var req dto.CreateParserConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	config := req.ToModel()
	err := h.parserConfigService.CreateParserConfig(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建解析配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "解析配置创建成功",
		"data":    dto.NewParserConfigResponse(config),
	})
}

// GetParserConfigByID 根据ID获取解析配置
func (h *ParserConfigHandler) GetParserConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID参数",
		})
		return
	}

	config, err := h.parserConfigService.GetParserConfigByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "解析配置不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.NewParserConfigResponse(config),
	})
}

// GetParserConfigsByContract 根据合约地址获取解析配置列表
func (h *ParserConfigHandler) GetParserConfigsByContract(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "合约地址不能为空",
		})
		return
	}

	configs, err := h.parserConfigService.GetParserConfigsByContract(c.Request.Context(), contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取解析配置失败",
			"details": err.Error(),
		})
		return
	}

	responses := make([]*dto.ParserConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = dto.NewParserConfigResponse(config)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"contract_address": contractAddress,
			"parser_configs":   responses,
			"total_count":      len(responses),
		},
	})
}

// GetParserConfigBySignature 根据合约地址和函数签名获取解析配置
func (h *ParserConfigHandler) GetParserConfigBySignature(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	signature := c.Param("signature")

	if contractAddress == "" || signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "合约地址和函数签名不能为空",
		})
		return
	}

	config, err := h.parserConfigService.GetParserConfigBySignature(c.Request.Context(), contractAddress, signature)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "解析配置不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.NewParserConfigResponse(config),
	})
}

// ListParserConfigs 分页获取解析配置列表
func (h *ParserConfigHandler) ListParserConfigs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	contractAddress := c.Query("contract_address")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	configs, total, err := h.parserConfigService.ListParserConfigs(c.Request.Context(), page, pageSize, contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取解析配置列表失败",
			"details": err.Error(),
		})
		return
	}

	response := dto.NewParserConfigListResponse(configs, total, page, pageSize)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateParserConfig 更新解析配置
func (h *ParserConfigHandler) UpdateParserConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID参数",
		})
		return
	}

	var req dto.UpdateParserConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	config, err := h.parserConfigService.GetParserConfigByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "解析配置不存在",
			"details": err.Error(),
		})
		return
	}

	req.ApplyToModel(config)
	err = h.parserConfigService.UpdateParserConfig(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新解析配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "解析配置更新成功",
		"data":    dto.NewParserConfigResponse(config),
	})
}

// DeleteParserConfig 删除解析配置
func (h *ParserConfigHandler) DeleteParserConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID参数",
		})
		return
	}

	err = h.parserConfigService.DeleteParserConfig(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除解析配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "解析配置删除成功",
	})
}

// GetContractParserInfo 获取合约的完整解析信息（三表联查）
func (h *ParserConfigHandler) GetContractParserInfo(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "合约地址不能为空",
		})
		return
	}

	info, err := h.parserConfigService.GetContractParserInfo(c.Request.Context(), contractAddress)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "获取合约解析信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.NewContractParserInfoResponse(info),
	})
}

// ParseInputData 解析交易输入数据
func (h *ParserConfigHandler) ParseInputData(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "合约地址不能为空",
		})
		return
	}

	var req struct {
		InputData string `json:"input_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	result, err := h.parserConfigService.ParseInputData(c.Request.Context(), contractAddress, req.InputData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "解析输入数据失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
