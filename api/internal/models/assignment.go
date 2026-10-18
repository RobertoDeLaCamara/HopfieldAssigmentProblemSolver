package models

import (
	"encoding/json"

	"hopfield-assignment-api/pkg/i18n"
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
		return newValidationError("problems", i18n.BatchTooLarge, n, MaxBatchSize)
	}
	return nil
}

// Validate validates the cost matrix
func (r *AssignmentRequest) Validate() error {
	if len(r.CostMatrix) == 0 {
		return newValidationError("cost_matrix", i18n.MatrixEmpty)
	}

	n := len(r.CostMatrix)
	if n < MinMatrixSize {
		return newValidationError("cost_matrix", i18n.MatrixTooSmall, n, n, MinMatrixSize, MinMatrixSize)
	}
	if n > MaxMatrixSize {
		return newValidationError("cost_matrix", i18n.MatrixTooLarge, n, n, MaxMatrixSize, MaxMatrixSize)
	}
	for i, row := range r.CostMatrix {
		if len(row) != n {
			return newValidationError("cost_matrix", i18n.MatrixNotSquare, i, len(row), n)
		}
		for j, cost := range row {
			if cost < 0 {
				return newValidationError("cost_matrix", i18n.CostNegative, i, j, cost)
			}
			if cost > MaxCostValue {
				return newValidationError("cost_matrix", i18n.CostTooLarge, i, j, cost, MaxCostValue)
			}
		}
	}

	return nil
}

// ValidationError represents a validation error. Message is the English text (what Error returns);
// Code and Args let Localized render the same error in another language (see pkg/i18n).
type ValidationError struct {
	Field   string
	Message string
	Code    string
	Args    []interface{}
}

func newValidationError(field, code string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Code:    code,
		Args:    args,
		Message: i18n.T(i18n.English, code, args...),
	}
}

// Localized renders the error in the given language (English when there is no translation).
func (e *ValidationError) Localized(lang i18n.Lang) string {
	if e.Code == "" {
		return e.Message
	}
	return i18n.T(lang, e.Code, e.Args...)
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
