package handlers

import (
	"net/http"

	"blockChainBrowser/server/internal/middleware"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService *services.PermissionService
}

func NewPermissionHandler(permissionService *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	resource := c.Query("resource")
	action := c.Query("action")
	if resource == "" || action == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "资源名称和操作类型不能为空", nil)
		return
	}
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}
	allowed := h.permissionService.HasPermission(userID, resource, action)
	utils.SuccessResponse(c, http.StatusOK, "ok", gin.H{"allowed": allowed})
}

func (h *PermissionHandler) UserPermissions(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}
	perms, err := h.permissionService.GetUserPermissions(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取权限失败", err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "ok", perms)
}
