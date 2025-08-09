package handlers

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CoinConfigHandler struct {
	coinConfigService services.CoinConfigService
}

/*
创建币种配置处理器
@param coinConfigService services.CoinConfigService
@return *CoinConfigHandler
*/
func NewCoinConfigHandler(coinConfigService services.CoinConfigService) *CoinConfigHandler {
	return &CoinConfigHandler{
		coinConfigService: coinConfigService,
	}
}

/*
创建币种配置
@param c *gin.Context
@return
*/
func (h *CoinConfigHandler) CreateCoinConfig(c *gin.Context) {
	// 1. 从URL路径获取符号参数
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "symbol parameter is required",
		})
		return
	}

	// 2. 从请求体获取币种配置信息并验证
	var req dto.CreateCoinConfigRequest
	if err := utils.ValidateAndBind(c, &req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// 3. 将DTO转换为模型
	coinConfig := req.ToModel(symbol)

	// 4. 调用服务层处理业务逻辑
	err := h.coinConfigService.CreateCoinConfig(c.Request.Context(), coinConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 5. 返回响应DTO
	response := dto.NewCoinConfigResponse(coinConfig)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "coin config created successfully",
	})
}

/*
根据符号获取币种配置
@param c *gin.Context
@return
*/
func (h *CoinConfigHandler) GetCoinConfigBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "symbol parameter is required",
		})
		return
	}
	coinConfig, err := h.coinConfigService.GetCoinConfigBySymbol(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	response := dto.NewCoinConfigResponse(coinConfig)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
