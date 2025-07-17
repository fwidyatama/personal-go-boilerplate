package router

import (
	"go-boilerplate/internal/handler"
	"go-boilerplate/internal/middleware"
	"go-boilerplate/internal/registry"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, reg *registry.AppRegistry) {
	router.Static("/swagger-ui/", "./swagger-ui")
	router.StaticFile("/swagger/openapi.yaml", "./docs/openapi.yaml")

	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(reg.Auth, reg.Redis))
		{
			users.POST("/", reg.Handlers.UserHandler.Create)
			users.GET("/", reg.Handlers.UserHandler.List)
			users.GET("/:id", reg.Handlers.UserHandler.GetByID)
			users.PUT("/:id", reg.Handlers.UserHandler.Update)
			users.DELETE("/:id", reg.Handlers.UserHandler.Delete)
			users.GET("/me", reg.Handlers.UserHandler.GetMe)
		}
		auth := api.Group("/auth")
		{
			auth.POST("/register", reg.Handlers.AuthHandler.Register)
			auth.POST("/login", reg.Handlers.AuthHandler.Login)
			auth.POST("/refresh", reg.Handlers.AuthHandler.Refresh)
			auth.POST("/logout", reg.Handlers.AuthHandler.Logout)
			auth.POST("/forgot-password", reg.Handlers.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", reg.Handlers.AuthHandler.ResetPassword)
			auth.PUT("/change-password", reg.Handlers.AuthHandler.ChangePassword)
		}

		api.GET("/health", handler.HealthCheck(reg.DB))
		api.GET("/metrics", handler.MetricsHandler())
	}
}
