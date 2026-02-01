package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIKeyAuth middleware for API key authentication
func APIKeyAuth(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from environment or use default for development
		expectedAPIKey := os.Getenv("API_KEY")
		
		// If no API key is configured, skip authentication (development mode)
		if expectedAPIKey == "" {
			logger.Warn("API_KEY not configured - authentication disabled (development mode)")
			c.Next()
			return
		}

		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header with Bearer scheme
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			logger.WithFields(logrus.Fields{
				"ip":   c.ClientIP(),
				"path": c.Request.URL.Path,
			}).Warn("Request without API key")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API key required. Provide X-API-Key header or Authorization: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedAPIKey)) != 1 {
			logger.WithFields(logrus.Fields{
				"ip":   c.ClientIP(),
				"path": c.Request.URL.Path,
			}).Warn("Invalid API key")
			
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid API key",
			})
			c.Abort()
			return
		}

		// API key is valid
		c.Set("authenticated", true)
		c.Next()
	}
}

// RateLimitInfo adds rate limit information to response headers
func RateLimitInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", "99")
		c.Header("X-RateLimit-Reset", "3600")
		c.Next()
	}
}
