package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PublicRateLimiter 公开API限流器
type PublicRateLimiter struct {
	requests map[string]*RequestRecord
	mutex    sync.RWMutex
	limit    int           // 每小时最大请求数
	window   time.Duration // 时间窗口
}

// RequestRecord 请求记录
type RequestRecord struct {
	Count    int       // 请求次数
	FirstReq time.Time // 第一次请求时间
	LastReq  time.Time // 最后一次请求时间
}

// NewPublicRateLimiter 创建公开API限流器
func NewPublicRateLimiter(limit int, window time.Duration) *PublicRateLimiter {
	limiter := &PublicRateLimiter{
		requests: make(map[string]*RequestRecord),
		limit:    limit,
		window:   window,
	}

	// 启动清理协程
	go limiter.cleanupRoutine()

	return limiter
}

// PublicRateLimitMiddleware 公开API限流中间件
func (prl *PublicRateLimiter) PublicRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 检查是否超过限制
		if prl.isRateLimited(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       "请求过于频繁，请稍后再试",
				"message":     "公开API限流：每小时最多100次请求",
				"retry_after": "1小时",
			})
			c.Abort()
			return
		}

		// 记录请求
		prl.recordRequest(clientIP)

		c.Next()
	}
}

// isRateLimited 检查是否超过限流
func (prl *PublicRateLimiter) isRateLimited(clientIP string) bool {
	prl.mutex.RLock()
	defer prl.mutex.RUnlock()

	record, exists := prl.requests[clientIP]
	if !exists {
		return false
	}

	// 检查是否在时间窗口内
	if time.Since(record.FirstReq) > prl.window {
		return false
	}

	// 检查请求次数
	return record.Count >= prl.limit
}

// recordRequest 记录请求
func (prl *PublicRateLimiter) recordRequest(clientIP string) {
	prl.mutex.Lock()
	defer prl.mutex.Unlock()

	now := time.Now()
	record, exists := prl.requests[clientIP]

	if !exists {
		// 新客户端
		prl.requests[clientIP] = &RequestRecord{
			Count:    1,
			FirstReq: now,
			LastReq:  now,
		}
		return
	}

	// 检查是否超过时间窗口
	if time.Since(record.FirstReq) > prl.window {
		// 重置计数
		record.Count = 1
		record.FirstReq = now
		record.LastReq = now
	} else {
		// 增加计数
		record.Count++
		record.LastReq = now
	}
}

// cleanupRoutine 清理过期记录
func (prl *PublicRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			prl.cleanup()
		}
	}
}

// cleanup 清理过期的请求记录
func (prl *PublicRateLimiter) cleanup() {
	prl.mutex.Lock()
	defer prl.mutex.Unlock()

	now := time.Now()
	for ip, record := range prl.requests {
		if now.Sub(record.FirstReq) > prl.window {
			delete(prl.requests, ip)
		}
	}
}

// GetStats 获取限流统计信息
func (prl *PublicRateLimiter) GetStats() map[string]interface{} {
	prl.mutex.RLock()
	defer prl.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_clients"] = len(prl.requests)
	stats["limit_per_hour"] = prl.limit
	stats["window_duration"] = prl.window.String()

	// 统计活跃客户端
	activeClients := 0
	now := time.Now()
	for _, record := range prl.requests {
		if now.Sub(record.FirstReq) <= prl.window {
			activeClients++
		}
	}
	stats["active_clients"] = activeClients

	return stats
}
