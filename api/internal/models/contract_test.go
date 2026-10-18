package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gateway side of the validation contract shared with the Python solver (../../../contract).
// The same cases run in hopfield/tests/test_contract.py: if the two services stop agreeing on
// what is valid, one of the suites fails.

type contractFile struct {
	Limits struct {
		MatrixSize struct{ Min, Max int }     `json:"matrix_size"`
		CostValue  struct{ Min, Max float64 } `json:"cost_value"`
		BatchSize  struct{ Max int }          `json:"batch_size"`
	} `json:"limits"`
	MatrixCases []struct {
		Name     string      `json:"name"`
		Matrix   [][]float64 `json:"matrix"`
		Generate *struct {
			Size int     `json:"size"`
			Fill float64 `json:"fill"`
		} `json:"generate"`
		Valid bool `json:"valid"`
	} `json:"matrix_cases"`
	BatchSizeCases []struct {
		Name  string `json:"name"`
		Size  int    `json:"size"`
		Valid bool   `json:"valid"`
	} `json:"batch_size_cases"`
}

func loadContract(t *testing.T) contractFile {
	t.Helper()
	raw, err := os.ReadFile("../../../contract/validation_cases.json")
	require.NoError(t, err)
	var c contractFile
	require.NoError(t, json.Unmarshal(raw, &c))
	return c
}

func TestContract_MatrixCases(t *testing.T) {
	for _, tc := range loadContract(t).MatrixCases {
		t.Run(tc.Name, func(t *testing.T) {
			matrix := tc.Matrix
			if tc.Generate != nil {
				matrix = make([][]float64, tc.Generate.Size)
				for i := range matrix {
					matrix[i] = make([]float64, tc.Generate.Size)
					for j := range matrix[i] {
						matrix[i][j] = tc.Generate.Fill
					}
				}
			}
			err := (&AssignmentRequest{CostMatrix: matrix}).Validate()
			if tc.Valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestContract_BatchSizeCases(t *testing.T) {
	for _, tc := range loadContract(t).BatchSizeCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateBatchSize(tc.Size)
			if tc.Valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestContract_LimitsEqualTheConstants(t *testing.T) {
	l := loadContract(t).Limits
	assert.Equal(t, l.MatrixSize.Min, MinMatrixSize)
	assert.Equal(t, l.MatrixSize.Max, MaxMatrixSize)
	assert.Equal(t, l.CostValue.Min, MinCostValue)
	assert.Equal(t, l.CostValue.Max, MaxCostValue)
	assert.Equal(t, l.BatchSize.Max, MaxBatchSize)
}
