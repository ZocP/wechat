package middleware

import (
	"net/http"
	"sync"
	"time"

	"pickup/internal/model"

	"github.com/gin-gonic/gin"
)

// RateLimiter 简单的内存限流器
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// 定期清理过期的空键，防止内存泄漏
	go rl.cleanup()

	return rl
}

// cleanup 定期清理过期的空键
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for key, requests := range rl.requests {
			var valid []time.Time
			for _, t := range requests {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mutex.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 清理过期请求
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// 检查是否超过限制
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// 记录当前请求
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// 使用客户端IP作为限流键
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, model.NewErrorResponse(
				429, "rate limit exceeded",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
