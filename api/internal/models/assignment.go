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

// Validate validates the cost matrix
func (r *AssignmentRequest) Validate() error {
	if len(r.CostMatrix) == 0 {
		return &ValidationError{Field: "cost_matrix", Message: "The cost matrix cannot be empty"}
	}

	n := len(r.CostMatrix)
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
			_ = j // Avoid unused variable warning
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

// FromJSON creates an AssignmentRequest from JSON
func (r *AssignmentRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
