package handlers

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	assetService services.AssetService
}

func NewAssetHandler(assetService services.AssetService) *AssetHandler {
	return &AssetHandler{
		assetService: assetService,
	}
}

func (h *AssetHandler) CreateAsset(c *gin.Context) {
	// 1. 从URL路径获取地址参数
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "address parameter is required",
		})
		return
	}

	// 2. 从请求体获取资产信息并验证
	var req dto.CreateAssetRequest
	if err := utils.ValidateAndBind(c, &req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// 3. 将DTO转换为模型
	asset := req.ToModel(address)

	// 4. 调用服务层处理业务逻辑
	err := h.assetService.CreateAsset(c.Request.Context(), asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 5. 返回响应DTO
	response := dto.NewAssetResponse(asset)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "asset created successfully",
	})
}
