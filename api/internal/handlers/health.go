package handlers

import (
	"hopfield-assignment-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles service health requests
type HealthHandler struct {
	logger *logrus.Logger
}

// NewHealthHandler creates a new instance of the health handler
func NewHealthHandler(logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// HealthCheck verifies the service status
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:  "healthy",
		Service: "hopfield-assignment-api",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck verifies if the service is ready to receive traffic
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Here you could add additional checks such as:
	// - Database connectivity
	// - External service availability
	// - System resource verification

	response := models.HealthResponse{
		Status:  "ready",
		Service: "hopfield-assignment-api",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// LivenessCheck verifies if the service is alive
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// Basic verification that the service is running
	response := models.HealthResponse{
		Status:  "alive",
		Service: "hopfield-assignment-api",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
