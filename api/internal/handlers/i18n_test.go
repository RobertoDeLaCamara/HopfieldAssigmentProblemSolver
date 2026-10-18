package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func serveCSVIn(t *testing.T, lang, field, content string) *httptest.ResponseRecorder {
	t.Helper()
	router := setupTestRouter()
	router.Use(middleware.Language())
	handler := NewAssignmentHandler(logrus.New())
	handler.httpClient = &MockHTTPClient{} // never reached: every case below is rejected first
	router.POST("/solve/csv", handler.SolveCSV)

	body, ct := csvUpload(t, field, content)
	req := httptest.NewRequest("POST", "/solve/csv", body)
	req.Header.Set("Content-Type", ct)
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCSVErrors_AreTranslated(t *testing.T) {
	cases := []struct {
		name, content, en, es string
	}{
		{"non-numeric cell", "1,2\n3,x\n",
			`Invalid CSV: row 2, column 2 is not a number: "x"`,
			`CSV no válido: la fila 2, columna 2 no es un número: "x"`},
		{"ragged rows", "1,2\n3\n",
			"Invalid CSV: row 2 has 1 values but row 1 has 2",
			"CSV no válido: la fila 2 tiene 1 valores pero la fila 1 tiene 2"},
		{"empty file", "",
			"Invalid CSV: the CSV has no data",
			"CSV no válido: el CSV no contiene datos"},
		{"empty cell", "1,\n3,4\n",
			"Invalid CSV: row 1, column 2 is empty",
			"CSV no válido: la fila 1, columna 2 está vacía"},
		// not a CSV problem but a matrix one: the same Spanish as the JSON endpoint
		{"not square", "1,2,3\n4,5,6\n",
			"The cost matrix must be square. Row 0 has 3 elements, expected 2",
			"La matriz de costes debe ser cuadrada. La fila 0 tiene 3 elementos y se esperaban 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			en := serveCSVIn(t, "", "file", tc.content)
			assert.Equal(t, http.StatusBadRequest, en.Code)
			assert.Equal(t, tc.en, errorOf(t, en))

			es := serveCSVIn(t, "es-ES,es;q=0.9", "file", tc.content)
			assert.Equal(t, http.StatusBadRequest, es.Code)
			assert.Equal(t, tc.es, errorOf(t, es))
		})
	}
}

func TestCSVUploadProblems_AreTranslated(t *testing.T) {
	w := serveCSVIn(t, "es", "matrix", "1,2\n3,4\n")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, `Envía el CSV como multipart/form-data en un campo llamado "file"`, errorOf(t, w))

	big := strings.Repeat("1,2\n", maxCSVUploadBytes/4+10)
	w = serveCSVIn(t, "es", "file", big)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Equal(t, "El CSV supera los 1048576 bytes", errorOf(t, w))
}
