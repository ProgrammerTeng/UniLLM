package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleRateLimiter implements a per-key sliding window rate limiter in memory.
// For production, replace with Redis-based implementation.
type SimpleRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *SimpleRateLimiter {
	rl := &SimpleRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *SimpleRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove expired entries
	reqs := rl.requests[key]
	valid := reqs[:0]
	for _, t := range reqs {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

func (rl *SimpleRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for key, reqs := range rl.requests {
			valid := reqs[:0]
			for _, t := range reqs {
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
		rl.mu.Unlock()
	}
}

// RateLimit middleware limits requests per user.
func RateLimit(limiter *SimpleRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		key := "anon"
		if exists {
			key = toString(userID)
		}

		if !limiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": map[string]string{
					"message": "rate limit exceeded",
					"type":    "rate_limit_error",
				},
			})
			return
		}
		c.Next()
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	if i, ok := v.(int64); ok {
		return strconv.FormatInt(i, 10)
	}
	return "unknown"
}
