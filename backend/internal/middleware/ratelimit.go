package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[key]

	validRequests := make([]time.Time, 0, len(requests))
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// GetRemainingRequests returns remaining requests for a key
func (rl *RateLimiter) GetRemainingRequests(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[key]
	count := 0
	for _, t := range requests {
		if t.After(windowStart) {
			count++
		}
	}

	remaining := rl.limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// RateLimitMiddleware returns a Gin middleware for rate limiting
func RateLimitMiddleware(rl *RateLimiter, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			c.Next()
			return
		}

		if !rl.Allow(tenantID) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(time.Minute).Format(time.RFC3339))
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		remaining := rl.GetRemainingRequests(tenantID)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", time.Now().Add(time.Minute).Format(time.RFC3339))

		c.Next()
	}
}
