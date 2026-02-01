package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"hopfield-assignment-api/internal/models"
)

type HealthHandler struct {
	logger *logrus.Logger
}

func NewHealthHandler(logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:  "healthy",
		Service: "hopfield-assignment-api",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Here you could add additional checks such as:
	// - Database connectivity
	// - External service availability
	// - System resource verification

	response := models.HealthResponse{
		Status:  "ready",
		Service: "assignment-api",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// Basic verification that the service is running
	response := models.HealthResponse{
		Status:  "alive",
		Service: "assignment-api",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) CurrentTime(c *gin.Context) {
	// Return current time in ISO 8601 format
	currentTime := time.Now().Format(time.RFC3339)
	
	response := map[string]interface{}{
		"time": currentTime,
	}
	
	c.JSON(http.StatusOK, response)
}
