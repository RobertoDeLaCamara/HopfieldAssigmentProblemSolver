package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSVMatrix_Valid(t *testing.T) {
	want := [][]float64{{1, 2}, {3, 4.5}}
	cases := map[string]string{
		"plain":                      "1,2\n3,4.5\n",
		"no trailing newline":        "1,2\n3,4.5",
		"CRLF line endings":          "1,2\r\n3,4.5\r\n",
		"spaces around values":       " 1 , 2 \n 3 , 4.5 \n",
		"blank lines are skipped":    "\n1,2\n\n3,4.5\n\n",
		"space-only lines are blank": "1,2\n   \n3,4.5\n",
		"UTF-8 byte order mark":      "\xef\xbb\xbf1,2\n3,4.5\n",
		"semicolon separated":        "1;2\n3;4.5\n",
		"scientific notation":        "1e0,2\n3,4.5\n",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseCSVMatrix(strings.NewReader(in))
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestParseCSVMatrix_Errors(t *testing.T) {
	cases := []struct {
		name, in, contains string
	}{
		{"empty file", "", "no data"},
		{"only blank lines", "\n\n  \n", "no data"},
		{"text in a cell", "1,2\n3,abc\n", `row 2, column 2 is not a number: "abc"`},
		{"empty cell", "1,\n3,4\n", "row 1, column 2 is empty"},
		{"row of only separators", ",,\n", "row 1, column 1 is empty"},
		{"header row is not numeric", "a,b\n1,2\n", `row 1, column 1 is not a number: "a"`},
		{"NaN", "1,NaN\n3,4\n", "row 1, column 2 is not a finite number"},
		{"infinity", "1,2\n+Inf,4\n", "row 2, column 1 is not a finite number"},
		{"ragged rows", "1,2\n3\n", "row 2 has 1 values but row 1 has 2"},
		{"unterminated quote", "1,\"2\n3,4\n", "malformed CSV"},
		{"too many rows", strings.Repeat("1\n", maxCSVRows+1), "more than"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCSVMatrix(strings.NewReader(tc.in))
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tc.contains)
		})
	}
}

// The parser accepts any rectangular numeric CSV; squareness and negatives are Validate's job,
// so JSON and CSV share one set of rules.
func TestParseCSVMatrix_ThenValidate(t *testing.T) {
	m, err := ParseCSVMatrix(strings.NewReader("1,2,3\n4,5,6\n"))
	require.NoError(t, err)
	assert.Error(t, (&AssignmentRequest{CostMatrix: m}).Validate(), "2x3 is not square")

	m, err = ParseCSVMatrix(strings.NewReader("1,-2\n3,4\n"))
	require.NoError(t, err)
	assert.Error(t, (&AssignmentRequest{CostMatrix: m}).Validate(), "negative cost")
}
