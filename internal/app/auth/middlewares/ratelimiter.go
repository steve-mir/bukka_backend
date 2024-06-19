package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	Rate  rate.Limit
	Burst int
}

type RateLimiter struct {
	limiters map[string]map[string]*rate.Limiter
	configs  map[string]RateLimitConfig
	mu       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]map[string]*rate.Limiter),
		configs:  make(map[string]RateLimitConfig),
	}
}

func (rl *RateLimiter) SetRateLimitConfig(endpoint string, rateLimit RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.configs[endpoint] = rateLimit
}

func (rl *RateLimiter) getLimiter(ip, endpoint string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if _, exists := rl.limiters[endpoint]; !exists {
		rl.limiters[endpoint] = make(map[string]*rate.Limiter)
	}

	if limiter, exists := rl.limiters[endpoint][ip]; exists {
		return limiter
	}

	config, exists := rl.configs[endpoint]
	if !exists {
		config = RateLimitConfig{Rate: rate.Every(time.Second), Burst: 5}
	}
	limiter := rate.NewLimiter(config.Rate, config.Burst)

	rl.limiters[endpoint][ip] = limiter
	return limiter
}

func RateLimit(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath()
		limiter := rl.getLimiter(ip, endpoint)

		if limiter.Allow() {
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
		}
	}
}
