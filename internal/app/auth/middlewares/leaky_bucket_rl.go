package middlewares

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ab -n 20 -c 5 -r -s 1  http://localhost:6000/v1/auth/test

// for i in {1..30}; do                                                                                                                                                ─╯
//     curl http://localhost:6000/v1/auth/test
//     sleep 0.5  # 1 second delay
// done

type LeakyBucket struct {
	capacity   int
	rate       int
	tokens     int
	lastUpdate time.Time
}

type RateLimiterLb struct {
	buckets  map[string]*LeakyBucket
	capacity int
	rate     int
	mu       sync.Mutex
}

func NewRateLimiterLb(capacity, rate int) *RateLimiterLb {
	return &RateLimiterLb{
		buckets:  make(map[string]*LeakyBucket),
		capacity: capacity,
		rate:     rate,
	}
}

func (rl *RateLimiterLb) getBucket(ip string) *LeakyBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, exists := rl.buckets[ip]; exists {
		return bucket
	}

	bucket := &LeakyBucket{
		capacity:   rl.capacity,
		rate:       rl.rate,
		tokens:     rl.capacity,
		lastUpdate: time.Now(),
	}
	rl.buckets[ip] = bucket
	return bucket
}

func (lb *LeakyBucket) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(lb.lastUpdate).Seconds()
	lb.lastUpdate = now

	// Add tokens based on elapsed time and rate
	lb.tokens += int(elapsed * float64(lb.rate))
	if lb.tokens > lb.capacity {
		lb.tokens = lb.capacity
	}

	log.Printf("Elapsed: %f, Tokens: %d", elapsed, lb.tokens)

	if lb.tokens > 0 {
		lb.tokens--
		log.Printf("Allowing request. Tokens left: %d", lb.tokens)
		return true
	}

	log.Println("Rejecting request. No tokens left.")
	return false
}

func RateLimitMiddleware(rl *RateLimiterLb) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		bucket := rl.getBucket(ip)
		if bucket.Allow() {
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
		}
	}
}

func (rl *RateLimiterLb) CleanupOldBuckets(interval time.Duration) {
	log.Println("Cleaning")
	for {
		time.Sleep(interval)
		rl.mu.Lock()
		for ip, bucket := range rl.buckets {
			log.Println("Wiping", ip, bucket)
			if time.Since(bucket.lastUpdate) > interval {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}
