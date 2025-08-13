package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-boilerplate/internal/formatter"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/redis"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	KeyGenerator      func(*gin.Context) string
	FailOpen          bool // If true, allow requests when Redis is unavailable; if false, return 503
}

// DefaultKeyGenerator generates rate limit key based on IP
func DefaultKeyGenerator(c *gin.Context) string {
	return "rate_limit:" + c.ClientIP()
}

// AuthenticatedKeyGenerator generates rate limit key based on user ID if authenticated, otherwise IP
func AuthenticatedKeyGenerator(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return "rate_limit:user:" + fmt.Sprintf("%v", userID)
	}
	return "rate_limit:ip:" + c.ClientIP()
}

// RateLimitMiddleware creates a rate limiting middleware using Redis
func RateLimitMiddleware(redisClient redis.RedisClient, config RateLimitConfig) gin.HandlerFunc {
	if config.KeyGenerator == nil {
		config.KeyGenerator = DefaultKeyGenerator
	}

	return func(c *gin.Context) {
		key := config.KeyGenerator(c)

		// Atomically increment the counter
		newCount, err := redisClient.Incr(c.Request.Context(), key)
		if err != nil {
			// Check if it's a key not found error (shouldn't happen with INCR, but handle it)
			if redis.IsKeyNotFoundError(err) {
				// This shouldn't happen with INCR, but if it does, treat as first request
				newCount = 1
				err = redisClient.Expire(c.Request.Context(), key, 60) // 60 seconds
				if err != nil {
					logger.Errorf("error setting rate limit expiration after key not found: %v", err.Error())
				}
			} else {
				// This is a connection or server error
				logger.Errorf("Redis error in rate limiter: %v", err.Error())

				if config.FailOpen {
					// Fail open: allow the request but log the error
					logger.Warnf("Rate limiter failing open due to Redis error: %v", err.Error())
					c.Next()
					return
				} else {
					// Fail closed: return service unavailable
					formatter.Error(c, http.StatusServiceUnavailable, "Rate limit service unavailable", "Service Unavailable")
					c.Abort()
					return
				}
			}
		}

		// If this is the first request (count == 1), set the expiration
		if newCount == 1 {
			err = redisClient.Expire(c.Request.Context(), key, 60) // 60 seconds
			if err != nil {
				logger.Errorf("error setting rate limit expiration: %v", err.Error())
			}
		}

		// Check if limit exceeded
		if newCount > int64(config.RequestsPerMinute) {
			// Get TTL to inform client when to retry
			ttl, _ := redisClient.TTL(c.Request.Context(), key)

			c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(ttl.Seconds()), 10))

			formatter.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "Too Many Requests")
			c.Abort()
			return
		}

		// Set rate limit headers
		remaining := config.RequestsPerMinute - int(newCount)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

// CreateRateLimitConfigs returns common rate limit configurations
func CreateRateLimitConfigs() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		"auth": {
			RequestsPerMinute: 5, // Stricter for auth endpoints
			BurstSize:         2,
			KeyGenerator:      DefaultKeyGenerator,
			FailOpen:          false, // Fail closed for auth endpoints for security
		},
		"api": {
			RequestsPerMinute: 60, // General API endpoints
			BurstSize:         10,
			KeyGenerator:      AuthenticatedKeyGenerator,
			FailOpen:          true, // Fail open for API endpoints for availability
		},
		"public": {
			RequestsPerMinute: 30, // Public endpoints
			BurstSize:         5,
			KeyGenerator:      DefaultKeyGenerator,
			FailOpen:          true, // Fail open for public endpoints for availability
		},
	}
}
