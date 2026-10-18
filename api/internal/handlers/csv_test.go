package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hopfield-assignment-api/internal/models"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func csvUpload(t *testing.T, field, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(field, "costs.csv")
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return body, w.FormDataContentType()
}

func serveCSV(t *testing.T, body *bytes.Buffer, contentType string, solverCalled bool) (*httptest.ResponseRecorder, *MockHTTPClient) {
	t.Helper()
	router := setupTestRouter()
	handler := NewAssignmentHandler(logrus.New())
	mockClient := &MockHTTPClient{}
	handler.httpClient = mockClient
	if solverCalled {
		mockClient.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       createMockResponseBody(t, true, createMockAssignmentResponse()),
		}, nil)
	}
	router.POST("/solve/csv", handler.SolveCSV)

	req := httptest.NewRequest("POST", "/solve/csv", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w, mockClient
}

func TestSolveCSV_ValidUploadReturnsTheSameShapeAsJSON(t *testing.T) {
	body, ct := csvUpload(t, "file", "1,2\n3,4\n")
	w, mockClient := serveCSV(t, body, ct, true)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var resp models.APIResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	result, ok := resp.Result.(map[string]interface{})
	require.True(t, ok)
	for _, key := range []string{"assignments", "total_cost", "iterations", "cost_matrix"} {
		assert.Contains(t, result, key)
	}
	mockClient.AssertExpectations(t)
}

func TestSolveCSV_SemicolonFileFromASpanishSpreadsheet(t *testing.T) {
	body, ct := csvUpload(t, "file", "1;2\r\n3;4\r\n")
	w, _ := serveCSV(t, body, ct, true)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func TestSolveCSV_RejectsBadInputWithAClearMessageAndNeverCallsTheSolver(t *testing.T) {
	cases := []struct {
		name, content, contains string
	}{
		{"non-numeric cell", "1,2\n3,x\n", "row 2, column 2 is not a number"},
		{"ragged rows", "1,2\n3\n", "row 2 has 1 values"},
		{"not square", "1,2,3\n4,5,6\n", "square"},
		{"negative cost", "1,-2\n3,4\n", "negative"},
		{"empty file", "", "no data"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, ct := csvUpload(t, "file", tc.content)
			w, mockClient := serveCSV(t, body, ct, false)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			var resp models.APIResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.False(t, resp.Success)
			assert.Contains(t, resp.Error, tc.contains)
			mockClient.AssertNotCalled(t, "Do", mock.Anything)
		})
	}
}

func TestSolveCSV_WrongFieldNameOrNotMultipart(t *testing.T) {
	body, ct := csvUpload(t, "matrix", "1,2\n3,4\n")
	w, _ := serveCSV(t, body, ct, false)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `field named \"file\"`)

	w, _ = serveCSV(t, bytes.NewBufferString(`{"cost_matrix":[[1,2],[3,4]]}`), "application/json", false)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSolveCSV_UploadAboveTheLimitIs413(t *testing.T) {
	big := strings.Repeat("1,2\n", maxCSVUploadBytes/4+10)
	body, ct := csvUpload(t, "file", big)
	w, _ := serveCSV(t, body, ct, false)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code, w.Body.String())
}
