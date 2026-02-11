package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.limit)
	assert.Equal(t, time.Minute, rl.window)
}

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow("client1"), "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_Allow_OverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow("client1"))
	}
	assert.False(t, rl.Allow("client1"), "4th request should be denied")
	assert.False(t, rl.Allow("client1"), "5th request should be denied")
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))
	assert.False(t, rl.Allow("client1"), "client1 should be limited")

	// client2 should still be allowed
	assert.True(t, rl.Allow("client2"))
	assert.True(t, rl.Allow("client2"))
	assert.False(t, rl.Allow("client2"), "client2 should be limited")
}

func TestRateLimiter_Allow_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	assert.True(t, rl.Allow("client1"))
	assert.True(t, rl.Allow("client1"))
	assert.False(t, rl.Allow("client1"))

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	assert.True(t, rl.Allow("client1"), "should be allowed after window expires")
	assert.True(t, rl.Allow("client1"), "should be allowed after window expires")
}

func TestRateLimiter_Concurrency(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- rl.Allow("client1")
		}()
	}

	wg.Wait()
	close(allowed)

	allowedCount := 0
	for ok := range allowed {
		if ok {
			allowedCount++
		}
	}

	assert.Equal(t, 100, allowedCount, "exactly 100 requests should be allowed")
}

func TestRateLimitMiddleware_AllowedRequests(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware(3, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}
}

func TestRateLimitMiddleware_RateLimited(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware(2, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Exhaust limit
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// This should be rate limited
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(429), resp["code"])
}
