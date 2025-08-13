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

		// Get current count
		current, err := redisClient.Get(c.Request.Context(), key)
		if err != nil {
			// If key doesn't exist, start fresh
			current = "0"
		}

		currentCount, err := strconv.Atoi(current)
		if err != nil {
			currentCount = 0
		}

		// Check if limit exceeded
		if currentCount >= config.RequestsPerMinute {
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

		// Increment counter
		newCount := currentCount + 1
		if currentCount == 0 {
			// Set with expiration for new key
			err = redisClient.Set(c.Request.Context(), key, strconv.Itoa(newCount), 60) // 60 seconds
		} else {
			// Increment existing key
			err = redisClient.Set(c.Request.Context(), key, strconv.Itoa(newCount), -1) // Keep existing TTL
		}

		if err != nil {
			logger.Errorf("error in rate limiter: %v", err.Error())
		}

		// Set rate limit headers
		remaining := config.RequestsPerMinute - newCount
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
		},
		"api": {
			RequestsPerMinute: 60, // General API endpoints
			BurstSize:         10,
			KeyGenerator:      AuthenticatedKeyGenerator,
		},
		"public": {
			RequestsPerMinute: 30, // Public endpoints
			BurstSize:         5,
			KeyGenerator:      DefaultKeyGenerator,
		},
	}
}
