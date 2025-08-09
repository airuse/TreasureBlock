package handlers

import (
	"blockChainBrowser/server/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScannerHandler struct {
	baseConfigService services.BaseConfigService
}

func NewScannerHandler(baseConfigService services.BaseConfigService) *ScannerHandler {
	return &ScannerHandler{
		baseConfigService: baseConfigService,
	}
}

/*
获取配置
@param c *gin.Context
@return
*/
func (h *ScannerHandler) GetScannerConfig(c *gin.Context) {

	configTypeStr := c.Query("configType")
	configGroup := c.Query("configGroup")
	configKey := c.Query("configKey")

	configTypeUint64, err := strconv.ParseUint(configTypeStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid config_type",
		})
		return
	}
	configType := uint8(configTypeUint64)

	config, err := h.baseConfigService.GetByConfigKey(c.Request.Context(), configKey, configType, configGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}
