package handlers

import (
	"hopfield-assignment-api/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestHealthHandler_HealthCheck tests the health endpoint to verify
// that the service returns a successful response with a healthy status.
func TestHealthHandler_HealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	handler := NewHealthHandler(logger)

	// Setup route
	router.GET("/health", handler.HealthCheck)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Parse response
	var response models.HealthResponse
	err := w.Body.UnmarshalJSON(w.Body.Bytes())
	assert.NoError(t, err)

	// Verify response content
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "hopfield-assignment-api", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
}

func TestHealthHandler_ReadinessCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	handler := NewHealthHandler(logger)

	// Setup route
	router.GET("/health/ready", handler.ReadinessCheck)

	// Create request
	req := httptest.NewRequest("GET", "/health/ready", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Parse response
	var response models.HealthResponse
	err := w.Body.UnmarshalJSON(w.Body.Bytes())
	assert.NoError(t, err)

	// Verify response content
	assert.Equal(t, "ready", response.Status)
	assert.Equal(t, "hopfield-assignment-api", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
}

func TestHealthHandler_LivenessCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	handler := NewHealthHandler(logger)

	// Setup route
	router.GET("/health/live", handler.LivenessCheck)

	// Create request
	req := httptest.NewRequest("GET", "/health/live", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Parse response
	var response models.HealthResponse
	err := w.Body.UnmarshalJSON(w.Body.Bytes())
	assert.NoError(t, err)

	// Verify response content
	assert.Equal(t, "alive", response.Status)
	assert.Equal(t, "hopfield-assignment-api", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
}

func TestHealthHandler_AllEndpoints(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	handler := NewHealthHandler(logger)

	// Setup routes
	router.GET("/health", handler.HealthCheck)
	router.GET("/health/ready", handler.ReadinessCheck)
	router.GET("/health/live", handler.LivenessCheck)

	// Test all endpoints
	endpoints := []struct {
		path   string
		status string
	}{
		{"/health", "healthy"},
		{"/health/ready", "ready"},
		{"/health/live", "alive"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", endpoint.path, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			// Parse response
			var response models.HealthResponse
			err := w.Body.UnmarshalJSON(w.Body.Bytes())
			assert.NoError(t, err)

			// Verify response content
			assert.Equal(t, endpoint.status, response.Status)
			assert.Equal(t, "hopfield-assignment-api", response.Service)
			assert.Equal(t, "1.0.0", response.Version)
		})
	}
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	handler := NewHealthHandler(logger)

	// Setup route
	router.GET("/health", handler.HealthCheck)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify JSON structure
	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(w.Body.Bytes())
	assert.NoError(t, err)

	// Check required fields
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "service")
	assert.Contains(t, response, "version")

	// Check field types
	assert.IsType(t, "", response["status"])
	assert.IsType(t, "", response["service"])
	assert.IsType(t, "", response["version"])

	// Check field values
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "hopfield-assignment-api", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
}
