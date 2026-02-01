package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hopfield-assignment-api/internal/models"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AssignmentHandler handles requests related to the assignment problem
type AssignmentHandler struct {
	logger     *logrus.Logger
	hopfieldURL string
	httpClient *http.Client
}

// NewAssignmentHandler creates a new instance of the handler
func NewAssignmentHandler(logger *logrus.Logger) *AssignmentHandler {
	hopfieldURL := os.Getenv("HOPFIELD_SERVICE_URL")
	if hopfieldURL == "" {
		hopfieldURL = "http://hopfield-service:5000"
	}

	return &AssignmentHandler{
		logger:       logger,
		hopfieldURL:  hopfieldURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SolveAssignment solves an assignment problem
func (h *AssignmentHandler) SolveAssignment(c *gin.Context) {
	var req models.AssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid request format")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate the cost matrix
	if err := req.Validate(); err != nil {
		h.logger.WithError(err).Warn("Validation error")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Call the Hopfield service with context for timeout control
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := h.callHopfieldServiceWithContext(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Error calling Hopfield service")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Internal server error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Result:  result,
	})
}

// SolveBatch solves multiple assignment problems in batch
func (h *AssignmentHandler) SolveBatch(c *gin.Context) {
	var req models.BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Error parsing batch request")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	if len(req.Problems) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "At least one problem is required in the batch",
		})
		return
	}

	// Process each problem with timeout
	results := make([]models.BatchResult, 0, len(req.Problems))
	for _, problem := range req.Problems {
		assignmentReq := models.AssignmentRequest{
			CostMatrix: problem.CostMatrix,
		}

		// Validate the individual problem
		if err := assignmentReq.Validate(); err != nil {
			results = append(results, models.BatchResult{
				ID:      problem.ID,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		// Solve the problem with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		result, err := h.callHopfieldServiceWithContext(ctx, assignmentReq)
		if err != nil {
			results = append(results, models.BatchResult{
				ID:      problem.ID,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		results = append(results, models.BatchResult{
			ID:      problem.ID,
			Success: true,
			Result:  result,
		})
	}

	c.JSON(http.StatusOK, models.BatchResponse{
		Success: true,
		Results: results,
	})
}

// callHopfieldServiceWithContext calls the Python service to solve the problem with context
func (h *AssignmentHandler) callHopfieldServiceWithContext(ctx context.Context, req models.AssignmentRequest) (*models.AssignmentResponse, error) {
	// Prepare the request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error serializing request: %w", err)
	}

	// Create the HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, "POST", h.hopfieldURL+"/solve", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		var errorResp models.APIResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf("service error (code %d): %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("service error: %s", errorResp.Error)
	}

	// Parse successful response
	var apiResp models.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("service error: %s", apiResp.Error)
	}

	// Convert result to AssignmentResponse
	resultBytes, err := json.Marshal(apiResp.Result)
	if err != nil {
		return nil, fmt.Errorf("error serializing result: %w", err)
	}

	var assignmentResp models.AssignmentResponse
	if err := json.Unmarshal(resultBytes, &assignmentResp); err != nil {
		return nil, fmt.Errorf("error parsing result: %w", err)
	}

	return &assignmentResp, nil
}
