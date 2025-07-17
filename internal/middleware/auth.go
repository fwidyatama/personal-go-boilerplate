package middleware

import (
	"net/http"
	"strings"

	"go-boilerplate/internal/auth"
	"go-boilerplate/internal/formatter"
	"go-boilerplate/pkg/redis"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(authService auth.AuthService, redisClient redis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			formatter.Error(c, http.StatusUnauthorized, "Authorization header required", "Unauthorized")
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			formatter.Error(c, http.StatusUnauthorized, "Invalid authorization header format", "Unauthorized")
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		claims, err := authService.ValidateAccessToken(tokenString)
		if err != nil {
			formatter.Error(c, http.StatusUnauthorized, "Invalid or expired token", "Unauthorized")
			c.Abort()
			return
		}

		// SessionID check against Redis
		redisSessionID, err := redisClient.Get(c.Request.Context(), claims.UserID.String())
		if err != nil || claims.SessionID == "" || claims.SessionID != redisSessionID {
			formatter.Error(c, http.StatusUnauthorized, "Session invalid or expired (logged in elsewhere)", "Unauthorized")
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
