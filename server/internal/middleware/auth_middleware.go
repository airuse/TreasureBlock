package middleware

import (
	"net/http"
	"strings"
	"time"

	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"blockChainBrowser/server/internal/services"
	"blockChainBrowser/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "缺少认证令牌", nil)
			c.Abort()
			return
		}

		// 检查Bearer格式
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌格式错误", nil)
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 验证JWT令牌
		token, err := authService.ValidateAccessToken(tokenString)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "认证令牌无效", err.Error())
			c.Abort()
			return
		}

		// 提取用户信息
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "令牌声明无效", nil)
			c.Abort()
			return
		}

		userID, err := services.ExtractUserIDFromToken(token)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "无效的用户ID", err.Error())
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", userID)
		c.Set("username", claims["username"])
		c.Set("token_type", claims["type"])

		// 如果是access_token类型，还需要存储api_key_id
		if tokenType, ok := claims["type"].(string); ok && tokenType == "access_token" {
			apiKeyID, err := services.ExtractAPIKeyIDFromToken(token)
			if err != nil {
				utils.ErrorResponse(c, http.StatusUnauthorized, "无效的API密钥ID", err.Error())
				c.Abort()
				return
			}
			c.Set("api_key_id", apiKeyID)
		}

		c.Next()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(
	apiKeyRepo repository.APIKeyRepository,
	requestLogRepo repository.RequestLogRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对使用access_token的请求进行限流
		tokenType, exists := c.Get("token_type")
		if !exists || tokenType != "access_token" {
			c.Next()
			return
		}

		apiKeyID, exists := c.Get("api_key_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusInternalServerError, "无法获取API密钥ID", nil)
			c.Abort()
			return
		}

		apiKeyIDUint, ok := apiKeyID.(uint)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "API密钥ID类型错误", nil)
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusInternalServerError, "无法获取用户ID", nil)
			c.Abort()
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误", nil)
			c.Abort()
			return
		}

		// 获取API密钥信息
		apiKey, err := apiKeyRepo.GetByID(apiKeyIDUint)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "API密钥无效", err.Error())
			c.Abort()
			return
		}

		// 检查当前小时的请求数量
		hourAgo := time.Now().Add(-time.Hour)
		hourlyCount, err := requestLogRepo.GetHourlyRequestCount(userIDUint, apiKeyIDUint, hourAgo)
		if err != nil {
			// 限流检查失败，允许请求通过但记录错误
			// 在生产环境中可能需要更严格的处理
		} else if hourlyCount >= int64(apiKey.RateLimit) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "请求频率超限", map[string]interface{}{
				"limit":   apiKey.RateLimit,
				"window":  "1 hour",
				"current": hourlyCount,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestLogMiddleware 请求日志中间件
func RequestLogMiddleware(requestLogRepo repository.RequestLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 只记录使用access_token的请求
		tokenType, exists := c.Get("token_type")
		if !exists || tokenType != "access_token" {
			return
		}

		// 获取用户和API密钥信息
		userID, exists := c.Get("user_id")
		if !exists {
			return
		}

		apiKeyID, exists := c.Get("api_key_id")
		if !exists {
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			return
		}

		apiKeyIDUint, ok := apiKeyID.(uint)
		if !ok {
			return
		}

		// 计算请求耗时
		duration := time.Since(startTime).Milliseconds()

		// 创建请求日志
		log := &models.RequestLog{
			UserID:     userIDUint,
			APIKeyID:   apiKeyIDUint,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			Duration:   duration,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		}

		// 异步保存日志，不阻塞请求
		go func() {
			if err := requestLogRepo.Create(log); err != nil {
				// 记录错误，但不影响响应
				// 在生产环境中应该使用更好的日志系统
			}
		}()
	}
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以根据需要实现管理员权限检查
		// 例如检查用户角色、特定权限等

		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "需要认证", nil)
			c.Abort()
			return
		}

		// 示例：检查是否为管理员用户（可以根据实际需求修改）
		// 这里简单检查用户ID是否为1（假设ID为1的用户是管理员）
		if userIDUint, ok := userID.(uint); !ok || userIDUint != 1 {
			utils.ErrorResponse(c, http.StatusForbidden, "需要管理员权限", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求认证）
func OptionalAuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]
		token, err := authService.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		userID, err := services.ExtractUserIDFromToken(token)
		if err != nil {
			c.Next()
			return
		}

		// 存储认证信息
		c.Set("user_id", userID)
		c.Set("username", claims["username"])
		c.Set("token_type", claims["type"])
		c.Set("authenticated", true)

		if tokenType, ok := claims["type"].(string); ok && tokenType == "access_token" {
			apiKeyID, err := services.ExtractAPIKeyIDFromToken(token)
			if err == nil {
				c.Set("api_key_id", apiKeyID)
			}
		}

		c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户ID的辅助函数
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userIDUint, ok := userID.(uint)
	return userIDUint, ok
}

// GetAPIKeyIDFromContext 从上下文中获取API密钥ID的辅助函数
func GetAPIKeyIDFromContext(c *gin.Context) (uint, bool) {
	apiKeyID, exists := c.Get("api_key_id")
	if !exists {
		return 0, false
	}

	apiKeyIDUint, ok := apiKeyID.(uint)
	return apiKeyIDUint, ok
}

// IsAuthenticated 检查用户是否已认证的辅助函数
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}
