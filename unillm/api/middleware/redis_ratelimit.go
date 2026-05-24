package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements sliding window rate limiting backed by Redis.
// Supports multi-instance deployment (all instances share the same counters).
type RedisRateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
	prefix string
}

// NewRedisRateLimiter creates a Redis-backed rate limiter.
func NewRedisRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
		prefix: "rl:",
	}
}

// Allow checks if a request is allowed for the given key using Redis sorted sets.
func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-rl.window)
	rkey := rl.prefix + key

	pipe := rl.rdb.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, rkey, "-inf", strconv.FormatFloat(float64(windowStart.UnixMicro()), 'f', 0, 64))

	// Count current entries
	countCmd := pipe.ZCard(ctx, rkey)

	// Add current request
	pipe.ZAdd(ctx, rkey, redis.Z{
		Score:  float64(now.UnixMicro()),
		Member: now.UnixMicro(),
	})

	// Set TTL to auto-cleanup
	pipe.Expire(ctx, rkey, rl.window+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		// On Redis failure, allow the request (fail open)
		return true, err
	}

	count := countCmd.Val()
	if count >= int64(rl.limit) {
		// Over limit — remove the entry we just added
		rl.rdb.ZRemRangeByRank(ctx, rkey, -1, -1)
		return false, nil
	}

	return true, nil
}

// RedisRateLimit middleware uses Redis-backed rate limiting.
func RedisRateLimit(limiter *RedisRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		key := "anon"
		if exists {
			if id, ok := userID.(int64); ok {
				key = strconv.FormatInt(id, 10)
			}
		}

		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			// Log but don't block on Redis errors (fail open)
			c.Next()
			return
		}

		if !allowed {
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
