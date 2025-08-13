package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-boilerplate/internal/auth"
	"go-boilerplate/internal/config"
	"go-boilerplate/internal/middleware"
	"go-boilerplate/internal/registry"
	"go-boilerplate/internal/router"
	"go-boilerplate/pkg/database"
	"go-boilerplate/pkg/logger"
	redis "go-boilerplate/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger.New(cfg.Log.Level)

	// Initialize database
	dbWrap, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := dbWrap.Close(); err != nil {
			logger.Errorf("failed to close db: %v", err)
		}
	}()
	db := dbWrap.DB

	// Initialize Redis client
	redisAddr := cfg.Redis.GetRedisAddr()
	redisClient := redis.NewRedis(redisAddr, cfg.Redis.Password, cfg.Redis.DB)

	// Initialize auth service
	authService := auth.NewAuthService(
		cfg.Auth.JWTSecret,
		cfg.Auth.RefreshSecret,
		cfg.Auth.AccessTTL,
		cfg.Auth.RefreshTTL,
		redisClient,
	)

	// Initialize registry
	appRegistry := registry.NewAppRegistry(db, authService, redisClient)

	// Setup Gin
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware(logger.Global))
	r.Use(gin.Logger()) // Add Gin logger for HTTP request logs
	r.Use(gin.Recovery())

	// Register routes
	router.RegisterRoutes(r, appRegistry)

	// Graceful shutdown setup
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Infof("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	logger.Infof("Server exiting")
}
