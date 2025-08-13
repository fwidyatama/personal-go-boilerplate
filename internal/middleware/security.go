package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	ContentSecurityPolicy     string
	ReferrerPolicy            string
	PermissionsPolicy         string
	CrossOriginEmbedderPolicy string
	CrossOriginOpenerPolicy   string
	CrossOriginResourcePolicy string
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ContentSecurityPolicy:     "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'none'; worker-src 'none'; frame-ancestors 'none'; form-action 'self'; base-uri 'self'; manifest-src 'self'",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "camera=(), microphone=(), geolocation=(), interest-cohort=()",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(config ...SecurityConfig) gin.HandlerFunc {
	var cfg SecurityConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultSecurityConfig()
	}

	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", cfg.ReferrerPolicy)
		c.Header("Permissions-Policy", cfg.PermissionsPolicy)

		// Content Security Policy
		if cfg.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}

		// Cross-Origin headers
		if cfg.CrossOriginEmbedderPolicy != "" {
			c.Header("Cross-Origin-Embedder-Policy", cfg.CrossOriginEmbedderPolicy)
		}
		if cfg.CrossOriginOpenerPolicy != "" {
			c.Header("Cross-Origin-Opener-Policy", cfg.CrossOriginOpenerPolicy)
		}
		if cfg.CrossOriginResourcePolicy != "" {
			c.Header("Cross-Origin-Resource-Policy", cfg.CrossOriginResourcePolicy)
		}

		// Remove server information
		c.Header("Server", "")

		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}

		// Only set CORS headers for allowed origins
		if allowed {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-Id")
			c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-Id, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours

			if allowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			if allowed {
				c.AbortWithStatus(204)
			} else {
				c.AbortWithStatus(403) // Forbidden for non-allowed origins
			}
			return
		}

		c.Next()
	}
}

// RequestSizeMiddleware limits request body size
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	})
}
