package middleware

import (
	"net/http"

	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
)

type PermissionMiddleware struct {
	permissionService *services.PermissionService
}

func NewPermissionMiddleware(permissionService *services.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{permissionService: permissionService}
}

func (m *PermissionMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
			c.Abort()
			return
		}
		if !m.permissionService.HasPermission(userID, resource, action) {
			utils.ErrorResponse(c, http.StatusForbidden, "权限不足", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
