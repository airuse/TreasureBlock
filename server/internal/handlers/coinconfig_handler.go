package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/services"

	"github.com/gin-gonic/gin"
)

// CoinConfigHandler 币种配置处理器
type CoinConfigHandler struct {
	coinConfigService   services.CoinConfigService
	parserConfigService services.ParserConfigService
}

// NewCoinConfigHandler 创建币种配置处理器
func NewCoinConfigHandler(coinConfigService services.CoinConfigService, parserConfigService services.ParserConfigService) *CoinConfigHandler {
	return &CoinConfigHandler{
		coinConfigService:   coinConfigService,
		parserConfigService: parserConfigService,
	}
}

// CreateCoinConfig 创建币种配置
func (h *CoinConfigHandler) CreateCoinConfig(c *gin.Context) {
	var req dto.CreateCoinConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	coinConfig := req.ToModel()
	err := h.coinConfigService.CreateCoinConfig(c.Request.Context(), coinConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建币种配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "币种配置创建成功",
		"data":    dto.NewCoinConfigResponse(coinConfig),
	})
}

// GetCoinConfigByID 根据ID获取币种配置
func (h *CoinConfigHandler) GetCoinConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID参数",
		})
		return
	}

	coinConfig, err := h.coinConfigService.GetCoinConfigByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "币种配置不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.NewCoinConfigResponse(coinConfig),
	})
}

// GetCoinConfigBySymbol 根据符号获取币种配置
func (h *CoinConfigHandler) GetCoinConfigBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "币种符号不能为空",
		})
		return
	}

	coinConfig, err := h.coinConfigService.GetCoinConfigBySymbol(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "币种配置不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.NewCoinConfigResponse(coinConfig),
	})
}

// GetCoinConfigByContractAddress 根据合约地址获取币种配置
func (h *CoinConfigHandler) GetCoinConfigByContractAddress(c *gin.Context) {
	contractAddr := c.Param("contractAddress")
	if contractAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "合约地址不能为空",
		})
		return
	}

	coinConfig, err := h.coinConfigService.GetCoinConfigByContractAddress(c.Request.Context(), contractAddr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "币种配置不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.NewCoinConfigResponse(coinConfig),
	})
}

// GetCoinConfigsByChain 根据链名称获取币种配置列表
func (h *CoinConfigHandler) GetCoinConfigsByChain(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "链名称不能为空",
		})
		return
	}

	coinConfigs, err := h.coinConfigService.GetCoinConfigsByChain(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取币种配置失败",
			"details": err.Error(),
		})
		return
	}

	responses := make([]*dto.CoinConfigResponse, len(coinConfigs))
	for i, coinConfig := range coinConfigs {
		responses[i] = dto.NewCoinConfigResponse(coinConfig)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chain_name":   chain,
			"coin_configs": responses,
			"total_count":  len(responses),
		},
	})
}

// GetStablecoins 获取指定链的稳定币列表
func (h *CoinConfigHandler) GetStablecoins(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "链名称不能为空",
		})
		return
	}

	coinConfigs, err := h.coinConfigService.GetStablecoins(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取稳定币列表失败",
			"details": err.Error(),
		})
		return
	}

	responses := make([]*dto.CoinConfigResponse, len(coinConfigs))
	for i, coinConfig := range coinConfigs {
		responses[i] = dto.NewCoinConfigResponse(coinConfig)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chain_name":  chain,
			"stablecoins": responses,
			"total_count": len(responses),
		},
	})
}

// GetVerifiedTokens 获取指定链的已验证代币列表
func (h *CoinConfigHandler) GetVerifiedTokens(c *gin.Context) {
	chain := c.Param("chain")
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "链名称不能为空",
		})
		return
	}

	coinConfigs, err := h.coinConfigService.GetVerifiedTokens(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取已验证代币列表失败",
			"details": err.Error(),
		})
		return
	}

	responses := make([]*dto.CoinConfigResponse, len(coinConfigs))
	for i, coinConfig := range coinConfigs {
		responses[i] = dto.NewCoinConfigResponse(coinConfig)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chain_name":      chain,
			"verified_tokens": responses,
			"total_count":     len(responses),
		},
	})
}

// ListCoinConfigs 分页获取币种配置列表
func (h *CoinConfigHandler) ListCoinConfigs(c *gin.Context) {
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

	coinConfigs, total, err := h.coinConfigService.ListCoinConfigs(c.Request.Context(), page, pageSize, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取币种配置列表失败",
			"details": err.Error(),
		})
		return
	}

	response := dto.NewCoinConfigListResponse(coinConfigs, total, page, pageSize)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateCoinConfig 更新币种配置
func (h *CoinConfigHandler) UpdateCoinConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID参数",
		})
		return
	}

	var req dto.UpdateCoinConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	coinConfig, err := h.coinConfigService.GetCoinConfigByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "币种配置不存在",
			"details": err.Error(),
		})
		return
	}

	req.ApplyToModel(coinConfig)
	err = h.coinConfigService.UpdateCoinConfig(c.Request.Context(), coinConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新币种配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "币种配置更新成功",
		"data":    dto.NewCoinConfigResponse(coinConfig),
	})
}

// DeleteCoinConfig 删除币种配置
func (h *CoinConfigHandler) DeleteCoinConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的ID参数",
		})
		return
	}

	err = h.coinConfigService.DeleteCoinConfig(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除币种配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "币种配置删除成功",
	})
}

// GetAllCoinConfigs 获取所有币种配置
func (h *CoinConfigHandler) GetAllCoinConfigs(c *gin.Context) {
	coinConfigs, err := h.coinConfigService.GetAllCoinConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取所有币种配置失败",
			"details": err.Error(),
		})
		return
	}

	responses := make([]*dto.CoinConfigResponse, len(coinConfigs))
	for i, coinConfig := range coinConfigs {
		responses[i] = dto.NewCoinConfigResponse(coinConfig)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"coin_configs": responses,
			"total_count":  len(responses),
		},
	})
}

// GetCoinConfigsForScanner 获取币种配置（扫块程序专用，简化版本）
func (h *CoinConfigHandler) GetCoinConfigsForScanner(c *gin.Context) {
	chain := c.Query("chain") // 可选参数，支持按链筛选

	var coinConfigs []*models.CoinConfig
	var err error

	if chain != "" {
		coinConfigs, err = h.coinConfigService.GetCoinConfigsByChain(c.Request.Context(), chain)
	} else {
		coinConfigs, err = h.coinConfigService.GetAllCoinConfigs(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取币种配置失败",
			"details": err.Error(),
		})
		return
	}

	// 转换为扫块程序需要的简化格式
	scannerData := make([]gin.H, len(coinConfigs))
	for i, config := range coinConfigs {
		scannerData[i] = gin.H{
			"id":            config.ID,
			"chain_name":    config.ChainName,
			"symbol":        config.Symbol,
			"coin_type":     config.CoinType,
			"contract_addr": config.ContractAddr,
			"precision":     config.Precision,
			"decimals":      config.Decimals,
			"name":          config.Name,
			"status":        config.Status,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    scannerData,
		"message": "获取币种配置成功",
	})
}

// GetCoinConfigForMaintenance 获取币种配置信息（维护用，包含解析配置）
func (h *CoinConfigHandler) GetCoinConfigForMaintenance(c *gin.Context) {
	// 添加调试日志
	// fmt.Printf("GetCoinConfigForMaintenance called with contract_address: %s\n", c.Param("contractAddress"))

	contractAddress := c.Param("contractAddress")
	if contractAddress == "" {
		// fmt.Println("Error: contract_address is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "合约地址不能为空",
		})
		return
	}

	// 获取币种配置
	// fmt.Printf("Getting coin config for contract: %s\n", contractAddress)
	coinConfig, err := h.coinConfigService.GetCoinConfigByContractAddress(c.Request.Context(), contractAddress)
	if err != nil {
		// fmt.Printf("No coin config found for contract %s: %v\n", contractAddress, err)
		// 如果不存在，返回空数据供创建
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"coin_config":      nil,
				"parser_configs":   []gin.H{},
				"contract_address": contractAddress,
			},
		})
		return
	}

	// 获取解析配置
	parserConfigs, err := h.parserConfigService.GetParserConfigsByContract(c.Request.Context(), contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取解析配置失败",
			"details": err.Error(),
		})
		return
	}

	// 转换为响应格式
	parserConfigResponses := make([]gin.H, len(parserConfigs))
	for i, config := range parserConfigs {
		parserConfigResponses[i] = gin.H{
			"id":                   config.ID,
			"function_name":        config.FunctionName,
			"function_signature":   config.FunctionSignature,
			"function_description": config.FunctionDescription,
			"display_format":       config.DisplayFormat,
			"param_config":         config.ParamConfig,
			"parser_rules":         config.ParserRules,
			"priority":             config.Priority,
			"is_active":            config.IsActive,
			// 新增日志解析配置字段
			"logs_parser_type":    config.LogsParserType,
			"event_signature":     config.EventSignature,
			"event_name":          config.EventName,
			"event_description":   config.EventDescription,
			"logs_param_config":   config.LogsParamConfig,
			"logs_parser_rules":   config.LogsParserRules,
			"logs_display_format": config.LogsDisplayFormat,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"coin_config":      dto.NewCoinConfigResponse(coinConfig),
			"parser_configs":   parserConfigResponses,
			"contract_address": contractAddress,
		},
	})
}
