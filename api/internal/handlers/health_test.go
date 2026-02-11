package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"hopfield-assignment-api/internal/models"
)

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

	// Check status code
	assert.Equal(t, 200, w.Code)

	// Check response body
	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "assignment-api", response.Service)
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

	// Check status code
	assert.Equal(t, 200, w.Code)

	// Check response body
	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response.Status)
	assert.Equal(t, "assignment-api", response.Service)
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

	// Check status code
	assert.Equal(t, 200, w.Code)

	// Check response body
	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response.Status)
	assert.Equal(t, "assignment-api", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
}

func TestHealthHandler_CurrentTime(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logrus.New()
	handler := NewHealthHandler(logger)

	// Setup route
	router.GET("/time", handler.CurrentTime)

	// Create request
	req := httptest.NewRequest("GET", "/time", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, 200, w.Code)

	// Check response body structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "time")
	
	// Verify time format is ISO 8601
	timeStr, ok := response["time"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, timeStr)
	
	// Verify it's a valid ISO 8601 timestamp
	_, err = time.Parse(time.RFC3339, timeStr)
	assert.NoError(t, err)
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
	router.GET("/time", handler.CurrentTime)

	endpoints := []struct {
		path string
	}{
		{"/health"},
		{"/health/ready"},
		{"/health/live"},
		{"/time"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", endpoint.path, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, 200, w.Code)

			// Check response body structure
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			if endpoint.path != "/time" {
				assert.Contains(t, response, "status")
				assert.Contains(t, response, "service")
				assert.Contains(t, response, "version")
			} else {
				assert.Contains(t, response, "time")
				// Verify time format
				timeStr, ok := response["time"].(string)
				assert.True(t, ok)
				_, err := time.Parse(time.RFC3339, timeStr)
				assert.NoError(t, err)
			}
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

	// Check response structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "service")
	assert.Contains(t, response, "version")
}
