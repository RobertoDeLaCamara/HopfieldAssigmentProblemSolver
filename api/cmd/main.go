package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"your-module/api/internal/handlers"
	"your-module/api/pkg/middleware"
)

func main() {
	// Configure logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Configure Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup routes
	router := gin.Default()
	
	// Add logging middleware
	router.Use(middleware.LoggingMiddleware(logger))

	// Add CORS middleware
	router.Use(middleware.CORS())

	// Health endpoints
	healthHandler := handlers.NewHealthHandler(logger)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/ready", healthHandler.ReadinessCheck)
	router.GET("/health/live", healthHandler.LivenessCheck)
	router.GET("/time", healthHandler.CurrentTime)

	// Assignment endpoints (without hopfield dependency)
	assignmentHandler := handlers.NewAssignmentHandler(logger)
	router.POST("/solve", assignmentHandler.SolveAssignment)
	router.POST("/solve/batch", assignmentHandler.SolveBatch)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	logger.Infof("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
