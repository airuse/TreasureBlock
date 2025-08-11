package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	Count     int
	LastTry   time.Time
	BlockedAt *time.Time
}

// BruteForceProtector 暴力破解防护器
type BruteForceProtector struct {
	attempts    map[string]*LoginAttempt // 记录IP的登录尝试次数
	mutex       sync.RWMutex             // 互斥锁
	maxAttempts int                      // 最大尝试次数
	blockTime   time.Duration            // 阻止时间
	windowTime  time.Duration            // 时间窗口
}

// NewBruteForceProtector 创建暴力破解防护器
func NewBruteForceProtector() *BruteForceProtector {
	protector := &BruteForceProtector{
		attempts:    make(map[string]*LoginAttempt),
		maxAttempts: 5,
		blockTime:   5 * time.Minute,
		windowTime:  1 * time.Second,
	}

	// 启动清理goroutine
	go protector.cleanupRoutine()

	return protector
}

// LoginAttemptMiddleware 登录尝试中间件
func (bf *BruteForceProtector) LoginAttemptMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api/auth/login" && c.Request.Method == "POST" {
			clientIP := c.ClientIP()

			if bf.isBlocked(clientIP) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"error":   fmt.Sprintf("登录尝试过多，请在 %d 分钟后再试", int(bf.blockTime.Minutes())),
					"data": gin.H{
						"blocked_until": bf.getBlockedUntil(clientIP),
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()

		// 检查登录结果
		if c.Request.URL.Path == "/api/auth/login" && c.Request.Method == "POST" {
			if c.Writer.Status() == http.StatusUnauthorized {
				bf.recordFailedAttempt(c.ClientIP())
			} else if c.Writer.Status() == http.StatusOK {
				bf.clearAttempts(c.ClientIP())
			}
		}
	}
}

// isBlocked 检查IP是否被阻止
func (bf *BruteForceProtector) isBlocked(ip string) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	attempt, exists := bf.attempts[ip]
	if !exists {
		return false
	}

	if attempt.BlockedAt != nil && time.Since(*attempt.BlockedAt) < bf.blockTime {
		return true
	}

	return false
}

// getBlockedUntil 获取阻止截止时间
func (bf *BruteForceProtector) getBlockedUntil(ip string) *time.Time {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	attempt, exists := bf.attempts[ip]
	if !exists || attempt.BlockedAt == nil {
		return nil
	}

	until := attempt.BlockedAt.Add(bf.blockTime)
	return &until
}

// recordFailedAttempt 记录失败尝试
func (bf *BruteForceProtector) recordFailedAttempt(ip string) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	now := time.Now()
	attempt, exists := bf.attempts[ip]

	if !exists {
		bf.attempts[ip] = &LoginAttempt{
			Count:   1,
			LastTry: now,
		}
		return
	}

	// 如果超过时间窗口，重置计数
	if now.Sub(attempt.LastTry) > bf.windowTime {
		attempt.Count = 1
		attempt.BlockedAt = nil
	} else {
		attempt.Count++
	}

	attempt.LastTry = now

	// 如果超过最大尝试次数，标记为阻止
	if attempt.Count >= bf.maxAttempts {
		blockTime := now
		attempt.BlockedAt = &blockTime
	}
}

// clearAttempts 清除尝试记录
func (bf *BruteForceProtector) clearAttempts(ip string) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()
	delete(bf.attempts, ip)
}

// cleanupRoutine 清理过期记录
func (bf *BruteForceProtector) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bf.cleanup()
		}
	}
}

// cleanup 清理过期记录
func (bf *BruteForceProtector) cleanup() {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	now := time.Now()
	for ip, attempt := range bf.attempts {
		// 清理超过窗口时间且未被阻止的记录
		if attempt.BlockedAt == nil && now.Sub(attempt.LastTry) > bf.windowTime {
			delete(bf.attempts, ip)
			continue
		}

		// 清理阻止时间已过的记录
		if attempt.BlockedAt != nil && now.Sub(*attempt.BlockedAt) > bf.blockTime {
			delete(bf.attempts, ip)
		}
	}
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止XSS攻击
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// HTTPS相关（仅在HTTPS时设置）
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:")

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Next()
	}
}

// HTTPSRedirectMiddleware HTTPS重定向中间件
func HTTPSRedirectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为HTTPS
		if c.Request.TLS == nil && c.GetHeader("X-Forwarded-Proto") != "https" {
			// 只对GET请求进行重定向
			if c.Request.Method != "GET" {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "HTTPS required for this operation",
				})
				c.Abort()
				return
			}

			// 构建HTTPS URL
			host := c.Request.Host
			if strings.Contains(host, ":8080") {
				host = strings.Replace(host, ":8080", ":8443", 1)
			}

			target := "https://" + host + c.Request.URL.RequestURI()
			c.Redirect(http.StatusMovedPermanently, target)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SanitizeMiddleware 敏感信息清理中间件
func SanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求（不包含敏感信息）
		if gin.Mode() == gin.DebugMode {
			c.Set("request_sanitized", sanitizeForLog(c.Request))
		}

		c.Next()
	}
}

// sanitizeForLog 清理敏感信息用于日志
func sanitizeForLog(data interface{}) interface{} {
	// 这里可以实现敏感信息清理逻辑
	// 例如移除密码、令牌等敏感字段
	return data
}

// isSensitiveField 判断是否为敏感字段
func isSensitiveField(field string) bool {
	sensitiveFields := []string{
		"password", "secret", "token", "key", "auth",
		"credential", "private", "confidential", "passwd",
	}

	fieldLower := strings.ToLower(field)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, sensitive) {
			return true
		}
	}
	return false
}

// IPWhitelistMiddleware IP白名单中间件
func IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	whitelistMap := make(map[string]bool)
	for _, ip := range whitelist {
		whitelistMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 如果白名单为空，则允许所有IP
		if len(whitelistMap) == 0 {
			c.Next()
			return
		}

		// 检查IP是否在白名单中
		if !whitelistMap[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied from this IP address",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestSizeLimitMiddleware 请求大小限制中间件
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Request body too large, maximum %d bytes allowed", maxSize),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
