package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"hopfield-assignment-api/internal/models"
)

// MockHTTPClient es un mock del cliente HTTP
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAssignmentHandler_SolveAssignment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid request",
			requestBody: models.AssignmentRequest{
				CostMatrix: [][]float64{{1, 2}, {3, 4}},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "invalid request - missing cost matrix",
			requestBody: map[string]interface{}{
				"invalid_field": "test",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid request - non-square matrix",
			requestBody: models.AssignmentRequest{
				CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid request - empty matrix",
			requestBody: models.AssignmentRequest{
				CostMatrix: [][]float64{},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := setupTestRouter()
			logger := logrus.New()
			handler := NewAssignmentHandler(logger)

			// Mock HTTP client
			mockClient := &MockHTTPClient{}
			handler.httpClient = mockClient

			// Setup mock response for valid requests
			if !tt.expectError {
				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       createMockResponseBody(t, true, createMockAssignmentResponse()),
				}
				mockClient.On("Do", mock.Anything).Return(mockResponse, nil)
			}

			// Setup route
			router.POST("/solve", handler.SolveAssignment)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/solve", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			} else {
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}

			// Verify mock expectations
			if !tt.expectError {
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestAssignmentHandler_SolveBatch(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
		expectMockCall bool
	}{
		{
			name: "valid batch request",
			requestBody: models.BatchRequest{
				Problems: []models.BatchProblem{
					{
						ID:         "problem_1",
						CostMatrix: [][]float64{{1, 2}, {3, 4}},
					},
					{
						ID:         "problem_2",
						CostMatrix: [][]float64{{5, 6}, {7, 8}},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			expectMockCall: true,
		},
		{
			name: "empty problems list",
			requestBody: models.BatchRequest{
				Problems: []models.BatchProblem{},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing problems field",
			requestBody: map[string]interface{}{
				"invalid_field": "test",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "batch with invalid problem",
			requestBody: models.BatchRequest{
				Problems: []models.BatchProblem{
					{
						ID:         "problem_1",
						CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}}, // Non-square matrix
					},
				},
			},
			expectedStatus: http.StatusOK, // Batch should still return success with error in result
			expectError:    false,
			expectMockCall: false, // Validation short-circuits before HTTP call
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := setupTestRouter()
			logger := logrus.New()
			handler := NewAssignmentHandler(logger)

			// Mock HTTP client
			mockClient := &MockHTTPClient{}
			handler.httpClient = mockClient

			// Setup mock response only when handler will actually call the solver
			if tt.expectMockCall {
				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       createMockResponseBody(t, true, createMockAssignmentResponse()),
				}
				mockClient.On("Do", mock.Anything).Return(mockResponse, nil)
			}

			// Setup route
			router.POST("/solve/batch", handler.SolveBatch)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/solve/batch", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			} else {
				var response models.BatchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.Results)

				// Check that we have the expected number of results
				if tt.name == "batch with invalid problem" {
					assert.Len(t, response.Results, 1)
					assert.False(t, response.Results[0].Success)
				}
			}

			// Verify mock expectations only when a call was expected
			if tt.expectMockCall {
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestAssignmentHandler_CallHopfieldServiceWithContext(t *testing.T) {
	tests := []struct {
		name           string
		request        models.AssignmentRequest
		mockResponse   *http.Response
		mockError      error
		expectError    bool
		expectedResult *models.AssignmentResponse
	}{
		{
			name: "successful call",
			request: models.AssignmentRequest{
				CostMatrix: [][]float64{{1, 2}, {3, 4}},
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       createMockResponseBody(t, true, createMockAssignmentResponse()),
			},
			mockError:   nil,
			expectError: false,
			expectedResult: &models.AssignmentResponse{
				Assignments: []int{1, 0},
				TotalCost:   5.0,
				Iterations:  10,
				CostMatrix:  [][]float64{{1, 2}, {3, 4}},
			},
		},
		{
			name: "service error",
			request: models.AssignmentRequest{
				CostMatrix: [][]float64{{1, 2}, {3, 4}},
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       createMockErrorResponseBody(t, "Internal server error"),
			},
			mockError:   nil,
			expectError: true,
		},
		{
			name: "network error",
			request: models.AssignmentRequest{
				CostMatrix: [][]float64{{1, 2}, {3, 4}},
			},
			mockResponse: nil,
			mockError:    assert.AnError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			handler := NewAssignmentHandler(logger)

			// Mock HTTP client
			mockClient := &MockHTTPClient{}
			handler.httpClient = mockClient

			// Setup mock expectations
			if tt.mockResponse != nil {
				mockClient.On("Do", mock.Anything).Return(tt.mockResponse, tt.mockError)
			} else {
				mockClient.On("Do", mock.Anything).Return(nil, tt.mockError)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call method
			result, err := handler.callHopfieldServiceWithContext(ctx, tt.request)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.Assignments, result.Assignments)
					assert.Equal(t, tt.expectedResult.TotalCost, result.TotalCost)
					assert.Equal(t, tt.expectedResult.Iterations, result.Iterations)
					assert.Equal(t, tt.expectedResult.CostMatrix, result.CostMatrix)
				}
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestAssignmentHandler_ValidateMatrix(t *testing.T) {
	tests := []struct {
		name        string
		costMatrix  [][]float64
		expectError bool
	}{
		{
			name:        "valid square matrix",
			costMatrix:  [][]float64{{1, 2}, {3, 4}},
			expectError: false,
		},
		{
			name:        "empty matrix",
			costMatrix:  [][]float64{},
			expectError: true,
		},
		{
			name:        "non-square matrix",
			costMatrix:  [][]float64{{1, 2, 3}, {4, 5, 6}},
			expectError: true,
		},
		{
			name:        "negative cost",
			costMatrix:  [][]float64{{1, -2}, {3, 4}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.AssignmentRequest{
				CostMatrix: tt.costMatrix,
			}

			err := req.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions

func createMockAssignmentResponse() models.AssignmentResponse {
	return models.AssignmentResponse{
		Assignments: []int{1, 0},
		TotalCost:   5.0,
		Iterations:  10,
		CostMatrix:  [][]float64{{1, 2}, {3, 4}},
	}
}

func createMockResponseBody(t *testing.T, success bool, result interface{}) io.ReadCloser {
	response := models.APIResponse{
		Success: success,
		Result:  result,
	}
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	return io.NopCloser(bytes.NewBuffer(jsonData))
}

func createMockErrorResponseBody(t *testing.T, errorMsg string) io.ReadCloser {
	response := models.APIResponse{
		Success: false,
		Error:   errorMsg,
	}
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	return io.NopCloser(bytes.NewBuffer(jsonData))
}

// The solver result is decoded once from raw JSON; these pin the edge cases so the
// response stays identical to the old Marshal/Unmarshal round trip (cnds-style golden cases).
func TestAssignmentHandler_CallHopfieldService_ResultDecoding(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantErr   bool
		wantCost  float64
		wantAssig []int
	}{
		{"extra unknown fields are ignored",
			`{"success":true,"request_id":"x","result":{"assignments":[1,0],"total_cost":5,"iterations":3,"cost_matrix":[[1,2],[3,4]],"extra":{"a":1}}}`,
			false, 5, []int{1, 0}},
		{"absent result is an empty response, not an error",
			`{"success":true}`, false, 0, nil},
		{"null result is an empty response, not an error",
			`{"success":true,"result":null}`, false, 0, nil},
		{"result of the wrong type is an error",
			`{"success":true,"result":"nope"}`, true, 0, nil},
		{"success=false carries the solver message",
			`{"success":false,"error":"bad matrix"}`, true, 0, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewAssignmentHandler(logrus.New())
			mockClient := &MockHTTPClient{}
			handler.httpClient = mockClient
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(tc.body)),
			}, nil)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			got, err := handler.callHopfieldServiceWithContext(ctx, models.AssignmentRequest{CostMatrix: [][]float64{{1, 2}, {3, 4}}})

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCost, got.TotalCost)
			assert.Equal(t, tc.wantAssig, got.Assignments)
		})
	}
}
