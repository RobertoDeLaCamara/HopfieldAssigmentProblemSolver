package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssignmentRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     AssignmentRequest
		expectError bool
		errorField  string
	}{
		{
			name: "valid 2x2 matrix",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{1, 2}, {3, 4}},
			},
			expectError: false,
		},
		{
			name: "valid 3x3 matrix",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			},
			expectError: false,
		},
		{
			name: "valid 1x1 matrix",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{5}},
			},
			expectError: false,
		},
		{
			name: "empty matrix",
			request: AssignmentRequest{
				CostMatrix: [][]float64{},
			},
			expectError: true,
			errorField:  "cost_matrix",
		},
		{
			name: "non-square matrix 2x3",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}},
			},
			expectError: true,
			errorField:  "cost_matrix",
		},
		{
			name: "non-square matrix 3x2",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{1, 2}, {3, 4}, {5, 6}},
			},
			expectError: true,
			errorField:  "cost_matrix",
		},
		{
			name: "matrix with negative costs",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{1, -2}, {3, 4}},
			},
			expectError: true,
			errorField:  "cost_matrix",
		},
		{
			name: "matrix with zero costs",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{0, 1}, {2, 0}},
			},
			expectError: false,
		},
		{
			name: "matrix with large costs",
			request: AssignmentRequest{
				CostMatrix: [][]float64{{1e6, 2e6}, {3e6, 4e6}},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if validationErr, ok := err.(*ValidationError); ok {
					assert.Equal(t, tt.errorField, validationErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssignmentRequest_ToJSON(t *testing.T) {
	request := AssignmentRequest{
		CostMatrix: [][]float64{{1, 2}, {3, 4}},
	}

	jsonData, err := request.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it's valid JSON by unmarshaling
	var unmarshaled AssignmentRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, request.CostMatrix, unmarshaled.CostMatrix)
}

func TestAssignmentRequest_FromJSON(t *testing.T) {
	jsonData := []byte(`{"cost_matrix": [[1, 2], [3, 4]]}`)

	var request AssignmentRequest
	err := request.FromJSON(jsonData)
	assert.NoError(t, err)
	assert.Equal(t, [][]float64{{1, 2}, {3, 4}}, request.CostMatrix)
}

func TestAssignmentRequest_FromJSON_Invalid(t *testing.T) {
	jsonData := []byte(`{"invalid": "data"}`)

	var request AssignmentRequest
	err := request.FromJSON(jsonData)
	assert.Error(t, err)
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "cost_matrix",
		Message: "La matriz debe ser cuadrada",
	}

	assert.Equal(t, "La matriz debe ser cuadrada", err.Error())
}

func TestAPIResponse_Structure(t *testing.T) {
	// Test success response
	successResp := APIResponse{
		Success: true,
		Result:  "test result",
	}

	assert.True(t, successResp.Success)
	assert.Equal(t, "test result", successResp.Result)
	assert.Empty(t, successResp.Error)

	// Test error response
	errorResp := APIResponse{
		Success: false,
		Error:   "test error",
	}

	assert.False(t, errorResp.Success)
	assert.Empty(t, errorResp.Result)
	assert.Equal(t, "test error", errorResp.Error)
}

func TestAssignmentResponse_Structure(t *testing.T) {
	response := AssignmentResponse{
		Assignments: []int{1, 0, 2},
		TotalCost:   15.5,
		Iterations:  100,
		CostMatrix:  [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
	}

	assert.Equal(t, []int{1, 0, 2}, response.Assignments)
	assert.Equal(t, 15.5, response.TotalCost)
	assert.Equal(t, 100, response.Iterations)
	assert.Equal(t, [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, response.CostMatrix)
}

func TestBatchProblem_Structure(t *testing.T) {
	problem := BatchProblem{
		ID:         "test_problem",
		CostMatrix: [][]float64{{1, 2}, {3, 4}},
	}

	assert.Equal(t, "test_problem", problem.ID)
	assert.Equal(t, [][]float64{{1, 2}, {3, 4}}, problem.CostMatrix)
}

func TestBatchRequest_Structure(t *testing.T) {
	request := BatchRequest{
		Problems: []BatchProblem{
			{ID: "problem_1", CostMatrix: [][]float64{{1, 2}, {3, 4}}},
			{ID: "problem_2", CostMatrix: [][]float64{{5, 6}, {7, 8}}},
		},
	}

	assert.Len(t, request.Problems, 2)
	assert.Equal(t, "problem_1", request.Problems[0].ID)
	assert.Equal(t, "problem_2", request.Problems[1].ID)
}

func TestBatchResult_Structure(t *testing.T) {
	// Test success result
	successResult := BatchResult{
		ID:      "test_problem",
		Success: true,
		Result:  "test result",
	}

	assert.Equal(t, "test_problem", successResult.ID)
	assert.True(t, successResult.Success)
	assert.Equal(t, "test result", successResult.Result)
	assert.Empty(t, successResult.Error)

	// Test error result
	errorResult := BatchResult{
		ID:      "test_problem",
		Success: false,
		Error:   "test error",
	}

	assert.Equal(t, "test_problem", errorResult.ID)
	assert.False(t, errorResult.Success)
	assert.Empty(t, errorResult.Result)
	assert.Equal(t, "test error", errorResult.Error)
}

func TestBatchResponse_Structure(t *testing.T) {
	response := BatchResponse{
		Success: true,
		Results: []BatchResult{
			{ID: "problem_1", Success: true, Result: "result_1"},
			{ID: "problem_2", Success: false, Error: "error_2"},
		},
	}

	assert.True(t, response.Success)
	assert.Len(t, response.Results, 2)
	assert.Equal(t, "problem_1", response.Results[0].ID)
	assert.True(t, response.Results[0].Success)
	assert.Equal(t, "problem_2", response.Results[1].ID)
	assert.False(t, response.Results[1].Success)
}

func TestHealthResponse_Structure(t *testing.T) {
	response := HealthResponse{
		Status:  "healthy",
		Service: "test-service",
		Version: "1.0.0",
	}

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "test-service", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
}
