package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// StructuredLogging middleware provides structured logging with request context
func StructuredLogging(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID
		requestID, _ := c.Get("request_id")

		// Log with structured fields
		fields := logrus.Fields{
			"request_id":  requestID,
			"method":      c.Request.Method,
			"path":        path,
			"query":       raw,
			"status":      c.Writer.Status(),
			"latency_ms":  latency.Milliseconds(),
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"error":       c.Errors.String(),
		}

		// Add authentication info if available
		if authenticated, exists := c.Get("authenticated"); exists {
			fields["authenticated"] = authenticated
		}

		// Log based on status code
		statusCode := c.Writer.Status()
		if statusCode >= 500 {
			logger.WithFields(fields).Error("Request completed with server error")
		} else if statusCode >= 400 {
			logger.WithFields(fields).Warn("Request completed with client error")
		} else {
			logger.WithFields(fields).Info("Request completed successfully")
		}
	}
}

// RequestMetrics middleware tracks request metrics
type RequestMetrics struct {
	TotalRequests   int64
	ErrorCount      int64
	AvgLatency      time.Duration
	MaxLatency      time.Duration
}

var metrics = &RequestMetrics{}

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		metrics.TotalRequests++

		if c.Writer.Status() >= 400 {
			metrics.ErrorCount++
		}

		if latency > metrics.MaxLatency {
			metrics.MaxLatency = latency
		}

		// Simple moving average
		metrics.AvgLatency = (metrics.AvgLatency + latency) / 2
	}
}

// GetMetrics returns current metrics
func GetMetrics() *RequestMetrics {
	return metrics
}
