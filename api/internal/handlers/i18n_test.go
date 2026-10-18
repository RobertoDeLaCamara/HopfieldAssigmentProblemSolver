package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hopfield-assignment-api/internal/models"
	"hopfield-assignment-api/pkg/middleware"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func postJSON(t *testing.T, path, acceptLanguage string, body interface{}, solverCalled bool) (*httptest.ResponseRecorder, *MockHTTPClient) {
	t.Helper()
	router := setupTestRouter()
	router.Use(middleware.Language())
	handler := NewAssignmentHandler(logrus.New())
	mockClient := &MockHTTPClient{}
	handler.httpClient = mockClient
	if solverCalled {
		mockClient.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       createMockResponseBody(t, true, createMockAssignmentResponse()),
		}, nil)
	}
	router.POST("/solve", handler.SolveAssignment)
	router.POST("/solve/batch", handler.SolveBatch)

	raw, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", path, bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, mockClient
}

func errorOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp models.APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp.Error
}

func TestValidationErrors_AreTranslatedByAcceptLanguage(t *testing.T) {
	notSquare := models.AssignmentRequest{CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}}}
	cases := []struct {
		name, lang, want string
	}{
		{"no header -> English", "", "The cost matrix must be square. Row 0 has 3 elements, expected 2"},
		{"es", "es", "La matriz de costes debe ser cuadrada. La fila 0 tiene 3 elementos y se esperaban 2"},
		{"es-ES with q-values", "es-ES,es;q=0.9,en;q=0.8", "La matriz de costes debe ser cuadrada. La fila 0 tiene 3 elementos y se esperaban 2"},
		{"English preferred over Spanish", "en-US,en;q=0.9,es;q=0.8", "The cost matrix must be square. Row 0 has 3 elements, expected 2"},
		{"unsupported language -> English", "fr", "The cost matrix must be square. Row 0 has 3 elements, expected 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, mockClient := postJSON(t, "/solve", tc.lang, notSquare, false)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, tc.want, errorOf(t, w))
			mockClient.AssertNotCalled(t, "Do", mock.Anything)
		})
	}
}

func TestNegativeCostAndSizeLimits_InSpanish(t *testing.T) {
	w, _ := postJSON(t, "/solve", "es", models.AssignmentRequest{CostMatrix: [][]float64{{1, -1}, {3, 4}}}, false)
	assert.Equal(t, "El coste en la posición [0][1] no puede ser negativo: -1.000000", errorOf(t, w))

	w, _ = postJSON(t, "/solve", "es", models.AssignmentRequest{CostMatrix: [][]float64{{5}}}, false)
	assert.Equal(t, "El tamaño de matriz 1x1 es demasiado pequeño. El tamaño mínimo es 2x2", errorOf(t, w))

	w, _ = postJSON(t, "/solve", "es", models.AssignmentRequest{CostMatrix: [][]float64{{1e10, 1}, {1, 1}}}, false)
	assert.Contains(t, errorOf(t, w), "supera el valor máximo permitido")
}

func TestMalformedRequestAndEmptyBatch_InSpanish(t *testing.T) {
	w, _ := postJSON(t, "/solve", "es", map[string]interface{}{"invalid_field": "x"}, false)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, errorOf(t, w), "Formato de solicitud no válido: ")

	w, _ = postJSON(t, "/solve/batch", "es", models.BatchRequest{Problems: []models.BatchProblem{}}, false)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Se requiere al menos un problema en el lote", errorOf(t, w))

	w, _ = postJSON(t, "/solve/batch", "en", models.BatchRequest{Problems: []models.BatchProblem{}}, false)
	assert.Equal(t, "At least one problem is required in the batch", errorOf(t, w))
}

func TestBatchItemValidationErrors_AreTranslatedAndSolverErrorsAreNot(t *testing.T) {
	batch := models.BatchRequest{Problems: []models.BatchProblem{
		{ID: "bad", CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}}},
		{ID: "good", CostMatrix: [][]float64{{1, 2}, {3, 4}}},
	}}
	w, _ := postJSON(t, "/solve/batch", "es", batch, true)
	require.Equal(t, http.StatusOK, w.Code)

	var resp models.BatchResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Results, 2)
	assert.False(t, resp.Results[0].Success)
	assert.Contains(t, resp.Results[0].Error, "La matriz de costes debe ser cuadrada")
	assert.True(t, resp.Results[1].Success)
}

func TestContentLanguageHeader(t *testing.T) {
	w, _ := postJSON(t, "/solve", "es", models.AssignmentRequest{CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}}}, false)
	assert.Equal(t, "es", w.Header().Get("Content-Language"))
	w, _ = postJSON(t, "/solve", "", models.AssignmentRequest{CostMatrix: [][]float64{{1, 2, 3}, {4, 5, 6}}}, false)
	assert.Equal(t, "en", w.Header().Get("Content-Language"))
}
