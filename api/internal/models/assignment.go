package models

import (
	"encoding/json"
	"fmt"
)

// AssignmentRequest represents the request to solve an assignment problem
type AssignmentRequest struct {
	CostMatrix [][]float64 `json:"cost_matrix" binding:"required"`
}

// AssignmentResponse represents the response from the Hopfield algorithm
type AssignmentResponse struct {
	Assignments []int       `json:"assignments"`
	TotalCost   float64     `json:"total_cost"`
	Iterations  int         `json:"iterations"`
	CostMatrix  [][]float64 `json:"cost_matrix"`
}

// APIResponse represents the standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BatchProblem represents an individual problem in a batch
type BatchProblem struct {
	ID         string      `json:"id" binding:"required"`
	CostMatrix [][]float64 `json:"cost_matrix" binding:"required"`
}

// BatchRequest represents a batch processing request
type BatchRequest struct {
	Problems []BatchProblem `json:"problems" binding:"required"`
}

// BatchResult represents the result of an individual problem in a batch
type BatchResult struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BatchResponse represents the batch processing response
type BatchResponse struct {
	Success bool          `json:"success"`
	Results []BatchResult `json:"results"`
	Error   string        `json:"error,omitempty"`
}

// HealthResponse represents the health endpoint response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// Limits shared with the Python solver; both are checked against contract/validation_cases.json.
// A request the gateway accepts but the solver rejects would come back as an opaque 500.
const (
	MinMatrixSize = 2
	MaxMatrixSize = 50
	MinCostValue  = 0.0
	MaxCostValue  = 1e9
	MaxBatchSize  = 100
)

// ValidateBatchSize rejects a batch with more problems than the solver accepts.
func ValidateBatchSize(n int) error {
	if n > MaxBatchSize {
		return &ValidationError{
			Field:   "problems",
			Message: fmt.Sprintf("Batch size of %d exceeds the maximum of %d problems. Please split it into smaller batches", n, MaxBatchSize),
		}
	}
	return nil
}

// Validate validates the cost matrix
func (r *AssignmentRequest) Validate() error {
	if len(r.CostMatrix) == 0 {
		return &ValidationError{Field: "cost_matrix", Message: "The cost matrix cannot be empty"}
	}

	n := len(r.CostMatrix)
	if n < MinMatrixSize {
		return &ValidationError{
			Field:   "cost_matrix",
			Message: fmt.Sprintf("Matrix size %dx%d is too small. Minimum size is %dx%d", n, n, MinMatrixSize, MinMatrixSize),
		}
	}
	if n > MaxMatrixSize {
		return &ValidationError{
			Field:   "cost_matrix",
			Message: fmt.Sprintf("Matrix size %dx%d exceeds the maximum allowed size of %dx%d", n, n, MaxMatrixSize, MaxMatrixSize),
		}
	}
	for i, row := range r.CostMatrix {
		if len(row) != n {
			return &ValidationError{
				Field:   "cost_matrix",
				Message: fmt.Sprintf("The cost matrix must be square. Row %d has %d elements, expected %d", i, len(row), n),
			}
		}
		for j, cost := range row {
			if cost < 0 {
				return &ValidationError{
					Field:   "cost_matrix",
					Message: fmt.Sprintf("Cost at position [%d][%d] cannot be negative: %f", i, j, cost),
				}
			}
			if cost > MaxCostValue {
				return &ValidationError{
					Field:   "cost_matrix",
					Message: fmt.Sprintf("Cost at position [%d][%d] is %g, which exceeds the maximum allowed value of %g", i, j, cost, MaxCostValue),
				}
			}
		}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ToJSON converts the structure to JSON
func (r *AssignmentRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON creates an AssignmentRequest from JSON and validates it.
// Returns an error if the JSON is malformed or the resulting request fails validation.
func (r *AssignmentRequest) FromJSON(data []byte) error {
	if err := json.Unmarshal(data, r); err != nil {
		return err
	}
	return r.Validate()
}
