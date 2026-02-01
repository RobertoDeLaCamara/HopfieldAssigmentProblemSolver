package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"hopfield-assignment-api/api/internal/handlers"
	"hopfield-assignment-api/api/pkg/middleware"
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

	// Assignment endpoints
	assignmentHandler := handlers.NewAssignmentHandler(logger)
	router.POST("/solve", assignmentHandler.SolveAssignment)
	router.POST("/solve/batch", assignmentHandler.SolveBatch)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	logger.Infof("Starting server on port %s", port)
	
	// Create server with timeout
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}
	
	logger.Info("Server exited properly")
}
