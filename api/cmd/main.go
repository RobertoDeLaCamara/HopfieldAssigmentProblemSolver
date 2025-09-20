package main

import (
	"hopfield-assignment-api/internal/handlers"
	"hopfield-assignment-api/pkg/middleware"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	// Create router
	router := gin.New()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(gin.Recovery())

	// Create handlers
	healthHandler := handlers.NewHealthHandler(logger)
	assignmentHandler := handlers.NewAssignmentHandler(logger)

	// Health routes
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/ready", healthHandler.ReadinessCheck)
	router.GET("/health/live", healthHandler.LivenessCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/solve", assignmentHandler.SolveAssignment)
		api.POST("/solve/batch", assignmentHandler.SolveBatch)
	}

	// Root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Hopfield Assignment Problem Solver",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Get port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.WithField("port", port).Info("Starting server")
	if err := router.Run(":" + port); err != nil {
		logger.WithError(err).Fatal("Error starting server")
	}
}
