package router

import (
	"go-boilerplate/internal/handler"
	"go-boilerplate/internal/middleware"
	"go-boilerplate/internal/registry"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, reg *registry.AppRegistry) {
	// Add security headers middleware
	router.Use(middleware.SecurityHeadersMiddleware())

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware([]string{"http://localhost:3000", "http://localhost:8080"}, true))

	// Add request size limit (10MB)
	router.Use(middleware.RequestSizeMiddleware(10 << 20))

	// Rate limit configurations
	rateLimitConfigs := middleware.CreateRateLimitConfigs()

	router.Static("/swagger-ui/", "./swagger-ui")
	router.StaticFile("/swagger/openapi.yaml", "./docs/openapi.yaml")

	api := router.Group("/api/v1")
	{
		// Public endpoints with rate limiting
		api.GET("/health",
			middleware.RateLimitMiddleware(reg.Redis, rateLimitConfigs["public"]),
			handler.HealthCheck(reg.DB))
		api.GET("/metrics",
			middleware.RateLimitMiddleware(reg.Redis, rateLimitConfigs["public"]),
			handler.MetricsHandler())

		// Auth endpoints with stricter rate limiting
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(reg.Redis, rateLimitConfigs["auth"]))
		{
			auth.POST("/register", reg.Handlers.AuthHandler.Register)
			auth.POST("/login", reg.Handlers.AuthHandler.Login)
			auth.POST("/refresh", reg.Handlers.AuthHandler.Refresh)
			auth.POST("/forgot-password", reg.Handlers.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", reg.Handlers.AuthHandler.ResetPassword)

			// Authenticated auth endpoints
			authProtected := auth.Group("/")
			authProtected.Use(middleware.AuthMiddleware(reg.Auth, reg.Redis))
			{
				authProtected.POST("/logout", reg.Handlers.AuthHandler.Logout)
				authProtected.PUT("/change-password", reg.Handlers.AuthHandler.ChangePassword)
			}
		}

		// Protected user endpoints
		users := api.Group("/users")
		users.Use(middleware.RateLimitMiddleware(reg.Redis, rateLimitConfigs["api"]))
		users.Use(middleware.AuthMiddleware(reg.Auth, reg.Redis))
		{
			users.POST("/", reg.Handlers.UserHandler.Create)
			users.GET("/", reg.Handlers.UserHandler.List)
			users.GET("/:id", reg.Handlers.UserHandler.GetByID)
			users.PUT("/:id", reg.Handlers.UserHandler.Update)
			users.DELETE("/:id", reg.Handlers.UserHandler.Delete)
			users.GET("/me", reg.Handlers.UserHandler.GetMe)
		}
	}
}
