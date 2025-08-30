package handlers

import (
	"net/http"
	"strconv"

	"blockChainBrowser/server/internal/dto"
	"blockChainBrowser/server/internal/middleware"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 201 {object} utils.Response{data=models.User} "注册成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 409 {object} utils.Response "用户名或邮箱已存在"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "用户名已存在" || err.Error() == "邮箱已存在" {
			statusCode = http.StatusConflict
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "注册成功", user)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} utils.Response{data=dto.LoginResponse} "登录成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "用户名或密码错误"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "用户名或密码错误" {
			statusCode = 400
		}
		if err.Error() == "用户账户已被禁用" {
			statusCode = 400
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "登录成功", response)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.UserProfileResponse} "获取成功"
// @Failure 401 {object} utils.Response "未认证"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	profile, err := h.authService.GetUserProfile(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取用户资料失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取成功", profile)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的登录密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "密码修改信息"
// @Success 200 {object} utils.Response "修改成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "当前密码错误"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	err := h.authService.ChangePassword(userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "当前密码错误" {
			statusCode = http.StatusUnauthorized
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "密码修改成功", nil)
}

// CreateAPIKey 创建API密钥
// @Summary 创建API密钥
// @Description 为当前用户创建新的API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAPIKeyRequest true "API密钥创建信息"
// @Success 201 {object} utils.Response{data=dto.CreateAPIKeyResponse} "创建成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "未认证"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/api-keys [post]
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.RateLimit == 0 {
		req.RateLimit = 1000 // 默认每小时1000次请求
	}

	response, err := h.authService.CreateAPIKey(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建API密钥失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "API密钥创建成功", response)
}

// GetAPIKeys 获取API密钥列表
// @Summary 获取API密钥列表
// @Description 获取当前用户的所有API密钥
// @Tags API密钥
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.APIKeyResponse} "获取成功"
// @Failure 401 {object} utils.Response "未认证"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/api-keys [get]
func (h *AuthHandler) GetAPIKeys(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	apiKeys, err := h.authService.GetAPIKeys(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取API密钥列表失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取成功", apiKeys)
}

// UpdateAPIKey 更新API密钥
// @Summary 更新API密钥
// @Description 更新指定的API密钥信息
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Param request body dto.UpdateAPIKeyRequest true "API密钥更新信息"
// @Success 200 {object} utils.Response{data=dto.APIKeyResponse} "更新成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "未认证"
// @Failure 403 {object} utils.Response "无权限"
// @Failure 404 {object} utils.Response "API密钥不存在"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/api-keys/{id} [put]
func (h *AuthHandler) UpdateAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的API密钥ID", err.Error())
		return
	}

	var req dto.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	response, err := h.authService.UpdateAPIKey(userID, uint(keyID), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "无权限修改此API密钥" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "API密钥不存在" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "API密钥更新成功", response)
}

// DeleteAPIKey 删除API密钥
// @Summary 删除API密钥
// @Description 删除指定的API密钥
// @Tags API密钥
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} utils.Response "删除成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "未认证"
// @Failure 403 {object} utils.Response "无权限"
// @Failure 404 {object} utils.Response "API密钥不存在"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/api-keys/{id} [delete]
func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的API密钥ID", err.Error())
		return
	}

	err = h.authService.DeleteAPIKey(userID, uint(keyID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "无权限删除此API密钥" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "API密钥不存在" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "API密钥删除成功", nil)
}

// GetAccessToken 获取访问令牌
// @Summary 获取访问令牌
// @Description 使用API密钥和Secret密钥获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.GetAccessTokenRequest true "获取访问令牌请求"
// @Success 200 {object} utils.Response{data=dto.GetAccessTokenResponse} "获取成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "API密钥或Secret密钥错误"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/auth/token [post]
func (h *AuthHandler) GetAccessToken(c *gin.Context) {
	var req dto.GetAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	response, err := h.authService.GetAccessToken(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "API密钥或Secret密钥错误" || err.Error() == "API密钥无效或已过期" {
			statusCode = http.StatusUnauthorized
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "访问令牌获取成功", response)
}

// GetUsageStats 获取API使用统计
// @Summary 获取API使用统计
// @Description 获取指定API密钥的使用统计信息
// @Tags API密钥
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} utils.Response{data=dto.APIUsageStatsResponse} "获取成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "未认证"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/user/api-keys/{id}/stats [get]
func (h *AuthHandler) GetUsageStats(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的API密钥ID", err.Error())
		return
	}

	stats, err := h.authService.GetUsageStats(userID, uint(keyID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取使用统计失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取成功", stats)
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 刷新当前的JWT令牌，获取新的访问令牌
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.LoginResponse} "刷新成功"
// @Failure 401 {object} utils.Response "令牌无效"
// @Failure 500 {object} utils.Response "服务器错误"
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	// 调用服务层刷新令牌
	response, err := h.authService.RefreshToken(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "令牌刷新失败", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "令牌刷新成功", response)
}
