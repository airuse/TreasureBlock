package handlers

import (
	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddressHandler struct {
	addressService services.AddressService
}

func NewAddressHandler(addressService services.AddressService) *AddressHandler {
	return &AddressHandler{
		addressService: addressService,
	}
}

func (h *AddressHandler) CreateAddress(c *gin.Context) {
	// 1. 从URL路径获取地址参数
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "address parameter is required",
		})
		return
	}

	// 2. 从请求体获取地址详细信息并验证
	var req dto.CreateAddressRequest
	if err := utils.ValidateAndBind(c, &req); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// 3. 将DTO转换为模型
	addr := req.ToModel(address)

	// 4. 调用服务层处理业务逻辑
	err := h.addressService.CreateAddress(c.Request.Context(), addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 5. 返回响应DTO
	response := dto.NewAddressResponse(addr)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "address created successfully",
	})
}
