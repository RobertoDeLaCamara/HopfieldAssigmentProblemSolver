package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hopfield-assignment-api/internal/models"
	"hopfield-assignment-api/pkg/i18n"
	"hopfield-assignment-api/pkg/metrics"
	"hopfield-assignment-api/pkg/middleware"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// localize renders an error for the client in the request's language. Only validation errors have
// translations; anything else (solver or network errors) is returned as is.
func localize(c *gin.Context, err error) string {
	var ve *models.ValidationError
	if errors.As(err, &ve) {
		return ve.Localized(middleware.LanguageFrom(c))
	}
	return err.Error()
}

// HTTPDoer is an interface for making HTTP requests (allows mocking in tests)
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// AssignmentHandler handles requests related to the assignment problem
type AssignmentHandler struct {
	logger      *logrus.Logger
	hopfieldURL string
	httpClient  HTTPDoer
}

// NewAssignmentHandler creates a new instance of the handler
func NewAssignmentHandler(logger *logrus.Logger) *AssignmentHandler {
	hopfieldURL := os.Getenv("HOPFIELD_SERVICE_URL")
	if hopfieldURL == "" {
		hopfieldURL = "http://hopfield-service:5000"
	}

	return &AssignmentHandler{
		logger:      logger,
		hopfieldURL: hopfieldURL,
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
			Error:   i18n.T(middleware.LanguageFrom(c), i18n.InvalidRequestFormat, err.Error()),
		})
		return
	}

	h.validateAndSolve(c, req)
}

// maxCSVUploadBytes caps the size of an uploaded CSV (a 50x50 matrix is a few KB).
const maxCSVUploadBytes = 1 << 20

// SolveCSV solves an assignment problem sent as a CSV file (multipart/form-data, field "file").
// The response is the same as for POST /solve.
func (h *AssignmentHandler) SolveCSV(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCSVUploadBytes)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, models.APIResponse{
				Success: false,
				Error:   i18n.T(middleware.LanguageFrom(c), i18n.CSVTooLarge, maxCSVUploadBytes),
			})
			return
		}
		h.logger.WithError(err).Warn("Missing CSV file")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   i18n.T(middleware.LanguageFrom(c), i18n.CSVMissingFile),
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.WithError(err).Warn("Could not open the uploaded CSV")
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: i18n.T(middleware.LanguageFrom(c), i18n.CSVUnreadable)})
		return
	}
	defer func() { _ = file.Close() }()

	matrix, err := models.ParseCSVMatrix(file)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid CSV")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   i18n.T(middleware.LanguageFrom(c), i18n.CSVInvalid, localize(c, err)),
		})
		return
	}

	h.validateAndSolve(c, models.AssignmentRequest{CostMatrix: matrix})
}

// validateAndSolve is the part of POST /solve and POST /solve/csv after the input is parsed.
func (h *AssignmentHandler) validateAndSolve(c *gin.Context, req models.AssignmentRequest) {
	// Validate the cost matrix
	if err := req.Validate(); err != nil {
		h.logger.WithError(err).Warn("Validation error")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   localize(c, err),
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
			Error:   i18n.T(middleware.LanguageFrom(c), i18n.InvalidRequestFormat, err.Error()),
		})
		return
	}

	if len(req.Problems) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   i18n.T(middleware.LanguageFrom(c), i18n.BatchEmpty),
		})
		return
	}

	if err := models.ValidateBatchSize(len(req.Problems)); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: localize(c, err)})
		return
	}

	// Process each problem with a per-iteration timeout context.
	// The closure scopes `defer cancel()` to a single iteration, preventing
	// timer/goroutine accumulation seen with `defer cancel()` inside a for loop.
	processProblem := func(problem models.BatchProblem) models.BatchResult {
		assignmentReq := models.AssignmentRequest{
			CostMatrix: problem.CostMatrix,
		}

		if err := assignmentReq.Validate(); err != nil {
			return models.BatchResult{ID: problem.ID, Success: false, Error: localize(c, err)}
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		result, err := h.callHopfieldServiceWithContext(ctx, assignmentReq)
		if err != nil {
			return models.BatchResult{ID: problem.ID, Success: false, Error: err.Error()}
		}
		return models.BatchResult{ID: problem.ID, Success: true, Result: result}
	}

	results := make([]models.BatchResult, 0, len(req.Problems))
	for _, problem := range req.Problems {
		results = append(results, processProblem(problem))
	}

	c.JSON(http.StatusOK, models.BatchResponse{
		Success: true,
		Results: results,
	})
}

// solverResponse is the envelope the Python solver answers with; Result is left raw.
type solverResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// callHopfieldServiceWithContext calls the Python service to solve the problem with context
func (h *AssignmentHandler) callHopfieldServiceWithContext(ctx context.Context, req models.AssignmentRequest) (result *models.AssignmentResponse, err error) {
	start := time.Now()
	defer func() {
		iterations := 0
		if result != nil {
			iterations = result.Iterations
		}
		metrics.ObserveSolverCall(len(req.CostMatrix), time.Since(start), iterations, err)
	}()

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
	defer func() { _ = resp.Body.Close() }()

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

	// Parse successful response. The result stays raw so it is decoded once, straight
	// into AssignmentResponse, instead of generic value -> Marshal -> Unmarshal.
	var apiResp solverResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("service error: %s", apiResp.Error)
	}

	var assignmentResp models.AssignmentResponse
	// An absent result decodes to the zero value, as it did when it went through "null".
	if len(apiResp.Result) > 0 {
		if err := json.Unmarshal(apiResp.Result, &assignmentResp); err != nil {
			return nil, fmt.Errorf("error parsing result: %w", err)
		}
	}

	return &assignmentResp, nil
}
