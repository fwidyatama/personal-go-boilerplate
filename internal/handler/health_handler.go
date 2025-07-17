package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp string            `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version" example:"1.0.0"`
}

// HealthCheck returns a health check handler
func HealthCheck(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "healthy"
		services := make(map[string]string)

		// Check database
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			status = "unhealthy"
			services["database"] = "unhealthy"
		} else {
			services["database"] = "healthy"
		}

		response := HealthResponse{
			Status:    status,
			Timestamp: time.Now().Format("2006-01-02T15:04:05Z"),
			Services:  services,
			Version:   "1.0.0",
		}

		if status == "healthy" {
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusServiceUnavailable, response)
		}
	}
}

// MetricsHandler returns a metrics handler for Prometheus
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real implementation, you would expose Prometheus metrics
		// For now, return a simple metrics endpoint
		c.JSON(http.StatusOK, gin.H{
			"message": "Metrics endpoint - implement Prometheus metrics here",
		})
	}
}
