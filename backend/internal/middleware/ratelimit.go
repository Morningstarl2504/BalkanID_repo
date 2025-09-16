package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"github.com/Morningstarl2504/Balkanid_repo/internal/utils"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check pattern
		if limiter, exists = rl.limiters[key]; !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.limiters[key] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	// Cleanup old limiters periodically
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()
		
		for range ticker.C {
			rl.mu.Lock()
			for key, limiter := range rl.limiters {
				// Remove limiters that haven't been used recently
				if limiter.TokensAt(time.Now()) == float64(rl.burst) {
					delete(rl.limiters, key)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return gin.HandlerFunc(func(c *gin.Context) {
		// Use user ID if authenticated, otherwise use IP
		key := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			key = string(rune(userID.(uint)))
		}

		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	})
}