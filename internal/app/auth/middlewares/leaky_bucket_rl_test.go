package middlewares

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test basic rate limiting
func TestBasicLeakyBucketRateLimiting(t *testing.T) {
	rl := NewRateLimiterLb(1, 1)

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// First request should pass
	w := performRequestLb(router, "GET", "/test")
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should be rate limited
	w = performRequestLb(router, "GET", "/test")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// Test token refill
func TestTokenRefill(t *testing.T) {
	rl := NewRateLimiterLb(1, 1)

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// First request should pass
	w := performRequestLb(router, "GET", "/test")
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should be rate limited
	w = performRequestLb(router, "GET", "/test")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Wait for tokens to refill
	time.Sleep(time.Second)

	// Next request should pass
	w = performRequestLb(router, "GET", "/test")
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test concurrent requests
func TestConcurrentRequests(t *testing.T) {
	rl := NewRateLimiterLb(5, 5)

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	var wg sync.WaitGroup
	successfulRequests := 0
	failedRequests := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := performRequestLb(router, "GET", "/test")
			if w.Code == http.StatusOK {
				successfulRequests++
			} else {
				failedRequests++
			}
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, successfulRequests, 5)
	assert.GreaterOrEqual(t, failedRequests, 5)
}

// Test bucket cleanup
func TestBucketCleanup(t *testing.T) {
	rl := NewRateLimiterLb(1, 1)

	go rl.CleanupOldBuckets(2 * time.Second)

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// First request should pass
	w := performRequestLb(router, "GET", "/test")
	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(3 * time.Second)

	rl.mu.Lock()
	_, exists := rl.buckets["127.0.0.1"]
	rl.mu.Unlock()
	assert.False(t, exists, "The bucket should be cleaned up after 2 seconds")
}

// Helper function to perform requests
func performRequestLb(router *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.RemoteAddr = "127.0.0.1:6000"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
