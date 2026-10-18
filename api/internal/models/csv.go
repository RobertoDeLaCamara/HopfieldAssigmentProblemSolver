package models

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// maxCSVRows bounds the work done on an upload before Validate() applies the size limits.
const maxCSVRows = 1000

// ParseCSVMatrix reads a cost matrix from CSV: one matrix row per line, numeric cells.
//
// Accepted: comma or semicolon separated (Spanish-locale spreadsheets export ';'), LF or CRLF
// line endings, a UTF-8 byte order mark, blank lines (skipped) and spaces around values.
// Rejected, with the 1-based row and column in the message: non-numeric, empty, NaN or infinite
// cells, rows of different lengths, an empty file. Squareness and the size/cost limits are
// checked afterwards by AssignmentRequest.Validate, so JSON and CSV share one set of rules.
func ParseCSVMatrix(r io.Reader) ([][]float64, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read the CSV: %w", err)
	}
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))

	reader := csv.NewReader(bufio.NewReader(bytes.NewReader(raw)))
	reader.Comma = detectDelimiter(raw)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // ragged rows are reported with their position below

	var matrix [][]float64
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("malformed CSV: %w", err)
		}
		if len(record) == 1 && strings.TrimSpace(record[0]) == "" {
			continue // a line with only spaces is blank, like an empty line
		}
		if len(matrix) >= maxCSVRows {
			return nil, fmt.Errorf("the CSV has more than %d rows", maxCSVRows)
		}

		row := make([]float64, len(record))
		for j, cell := range record {
			cell = strings.TrimSpace(cell)
			if cell == "" {
				return nil, fmt.Errorf("row %d, column %d is empty", len(matrix)+1, j+1)
			}
			v, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				return nil, fmt.Errorf("row %d, column %d is not a number: %q", len(matrix)+1, j+1, cell)
			}
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, fmt.Errorf("row %d, column %d is not a finite number: %q", len(matrix)+1, j+1, cell)
			}
			row[j] = v
		}
		if len(matrix) > 0 && len(row) != len(matrix[0]) {
			return nil, fmt.Errorf("row %d has %d values but row 1 has %d", len(matrix)+1, len(row), len(matrix[0]))
		}
		matrix = append(matrix, row)
	}

	if len(matrix) == 0 {
		return nil, errors.New("the CSV has no data")
	}
	return matrix, nil
}

// detectDelimiter picks ';' when the first non-blank line uses it and has no comma.
func detectDelimiter(raw []byte) rune {
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, ";") && !strings.Contains(line, ",") {
			return ';'
		}
		break
	}
	return ','
}
