package tests

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
// 	"github.com/stretchr/testify/assert"
// 	"golang.org/x/time/rate"
// )

// // Test basic rate limiting
// func TestBasicRateLimiting(t *testing.T) {
// 	rl := middlewares.NewRateLimiter()
// 	rl.SetRateLimitConfig("/test", middlewares.RateLimitConfig{Rate: rate.Every(time.Second), Burst: 1})

// 	router := gin.New()
// 	router.Use(middlewares.RateLimit(rl))
// 	router.GET("/test", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})

// 	// First request should pass
// 	w := performRequest(router, "GET", "/test")
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// Second request should be rate limited
// 	w = performRequest(router, "GET", "/test")
// 	assert.Equal(t, http.StatusTooManyRequests, w.Code)
// }

// // Test different limits for different endpoints
// func TestDifferentLimitsForEndpoints(t *testing.T) {
// 	rl := middlewares.NewRateLimiter()
// 	rl.SetRateLimitConfig("/test1", middlewares.RateLimitConfig{Rate: rate.Every(time.Second), Burst: 1})
// 	rl.SetRateLimitConfig("/test2", middlewares.RateLimitConfig{Rate: rate.Every(2 * time.Second), Burst: 1})

// 	router := gin.New()
// 	router.Use(middlewares.RateLimit(rl))
// 	router.GET("/test1", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})
// 	router.GET("/test2", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})

// 	// Test /test1 endpoint
// 	w := performRequest(router, "GET", "/test1")
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	w = performRequest(router, "GET", "/test1")
// 	assert.Equal(t, http.StatusTooManyRequests, w.Code)

// 	// Test /test2 endpoint
// 	w = performRequest(router, "GET", "/test2")
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	w = performRequest(router, "GET", "/test2")
// 	assert.Equal(t, http.StatusTooManyRequests, w.Code)
// }

// // Test custom default rate limit
// func TestCustomDefaultRateLimit(t *testing.T) {
// 	rl := middlewares.NewRateLimiter()
// 	rl.SetRateLimitConfig("/test", middlewares.RateLimitConfig{Rate: rate.Every(time.Second), Burst: 1})

// 	router := gin.New()
// 	router.Use(middlewares.RateLimit(rl))
// 	router.GET("/default", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})

// 	// First request should pass
// 	w := performRequest(router, "GET", "/default")
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// Second request should be rate limited
// 	w = performRequest(router, "GET", "/default")
// 	assert.Equal(t, http.StatusTooManyRequests, w.Code)
// }

// // Test concurrency handling
// func TestConcurrencyHandling(t *testing.T) {
// 	rl := middlewares.NewRateLimiter()
// 	rl.SetRateLimitConfig("/test", middlewares.RateLimitConfig{Rate: rate.Every(time.Second), Burst: 5})

// 	router := gin.New()
// 	router.Use(middlewares.RateLimit(rl))
// 	router.GET("/test", func(c *gin.Context) {
// 		c.String(http.StatusOK, "OK")
// 	})

// 	var wg sync.WaitGroup
// 	successfulRequests := 0
// 	failedRequests := 0

// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			w := performRequest(router, "GET", "/test")
// 			if w.Code == http.StatusOK {
// 				successfulRequests++
// 			} else {
// 				failedRequests++
// 			}
// 		}()
// 	}

// 	wg.Wait()
// 	assert.LessOrEqual(t, successfulRequests, 5)
// 	assert.GreaterOrEqual(t, failedRequests, 5)
// }

// // Helper function to perform requests
// func performRequest(router *gin.Engine, method, path string) *httptest.ResponseRecorder {
// 	req, _ := http.NewRequest(method, path, nil)
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	return w
// }
