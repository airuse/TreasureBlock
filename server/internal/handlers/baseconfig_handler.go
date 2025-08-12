package handlers

import (
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseConfigHandler 基础配置处理器
type BaseConfigHandler struct {
	baseConfigService services.BaseConfigService
}

// NewBaseConfigHandler 创建基础配置处理器
func NewBaseConfigHandler(baseConfigService services.BaseConfigService) *BaseConfigHandler {
	return &BaseConfigHandler{
		baseConfigService: baseConfigService,
	}
}

// GetConfigsByGroup 根据分组获取配置列表
// @Summary 根据分组获取配置列表
// @Description 根据配置分组获取相关的配置项
// @Tags 基础配置
// @Produce json
// @Param group path string true "配置分组"
// @Success 200 {object} utils.Response{data=[]models.BaseConfig}
// @Router /api/base-configs/group/{group} [get]
func (h *BaseConfigHandler) GetConfigsByGroup(c *gin.Context) {
	group := c.Param("group")
	if group == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "分组参数不能为空")
		return
	}

	configs, err := h.baseConfigService.GetByConfigGroup(c.Request.Context(), group)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取配置失败")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取配置成功", configs)
}

// GetConfigsByType 根据类型获取配置列表
// @Summary 根据类型获取配置列表
// @Description 根据配置类型获取相关的配置项
// @Tags 基础配置
// @Produce json
// @Param type path int true "配置类型"
// @Success 200 {object} utils.Response{data=[]models.BaseConfig}
// @Router /api/base-configs/type/{type} [get]
func (h *BaseConfigHandler) GetConfigsByType(c *gin.Context) {
	configTypeStr := c.Param("type")
	if configTypeStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "类型参数不能为空")
		return
	}

	// 这里可以添加类型转换逻辑，暂时返回错误
	utils.ErrorResponse(c, http.StatusBadRequest, "类型参数必须是数字")
}

// GetPermissionTypes 获取权限类型列表（专门用于API密钥权限）
// @Summary 获取权限类型列表
// @Description 获取所有可用的API权限类型
// @Tags 权限管理
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.BaseConfig}
// @Router /api/permissions [get]
func (h *BaseConfigHandler) GetPermissionTypes(c *gin.Context) {
	configs, err := h.baseConfigService.GetByConfigGroup(c.Request.Context(), "api_permissions")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取权限类型失败")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取权限类型成功", configs)
}
