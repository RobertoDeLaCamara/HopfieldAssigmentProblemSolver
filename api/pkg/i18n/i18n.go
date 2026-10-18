// Package i18n holds the gateway's user-facing error messages in English and Spanish and picks the
// language from the Accept-Language header. English is the default and the fallback: a code or a
// language without a translation is answered in English, never with an empty message.
package i18n

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Lang is a supported response language.
type Lang string

const (
	English Lang = "en"
	Spanish Lang = "es"
)

// Message codes. models.ValidationError carries one of these plus its arguments, so the same
// error can be rendered in any language.
const (
	MatrixEmpty          = "matrix_empty"
	MatrixTooSmall       = "matrix_too_small"
	MatrixTooLarge       = "matrix_too_large"
	MatrixNotSquare      = "matrix_not_square"
	CostNegative         = "cost_negative"
	CostTooLarge         = "cost_too_large"
	BatchTooLarge        = "batch_too_large"
	BatchEmpty           = "batch_empty"
	InvalidRequestFormat = "invalid_request_format"
)

// catalog[lang][code] is a fmt format string; every language uses the same verbs in the same order.
var catalog = map[Lang]map[string]string{
	English: {
		MatrixEmpty:          "The cost matrix cannot be empty",
		MatrixTooSmall:       "Matrix size %dx%d is too small. Minimum size is %dx%d",
		MatrixTooLarge:       "Matrix size %dx%d exceeds the maximum allowed size of %dx%d",
		MatrixNotSquare:      "The cost matrix must be square. Row %d has %d elements, expected %d",
		CostNegative:         "Cost at position [%d][%d] cannot be negative: %f",
		CostTooLarge:         "Cost at position [%d][%d] is %g, which exceeds the maximum allowed value of %g",
		BatchTooLarge:        "Batch size of %d exceeds the maximum of %d problems. Please split it into smaller batches",
		BatchEmpty:           "At least one problem is required in the batch",
		InvalidRequestFormat: "Invalid request format: %s",
	},
	Spanish: {
		MatrixEmpty:          "La matriz de costes no puede estar vacía",
		MatrixTooSmall:       "El tamaño de matriz %dx%d es demasiado pequeño. El tamaño mínimo es %dx%d",
		MatrixTooLarge:       "El tamaño de matriz %dx%d supera el máximo permitido de %dx%d",
		MatrixNotSquare:      "La matriz de costes debe ser cuadrada. La fila %d tiene %d elementos y se esperaban %d",
		CostNegative:         "El coste en la posición [%d][%d] no puede ser negativo: %f",
		CostTooLarge:         "El coste en la posición [%d][%d] es %g, que supera el valor máximo permitido de %g",
		BatchTooLarge:        "Un lote de %d problemas supera el máximo de %d. Divídelo en lotes más pequeños",
		BatchEmpty:           "Se requiere al menos un problema en el lote",
		InvalidRequestFormat: "Formato de solicitud no válido: %s",
	},
}

// T renders a message in the given language, falling back to English.
func T(lang Lang, code string, args ...interface{}) string {
	format, ok := catalog[lang][code]
	if !ok {
		format, ok = catalog[English][code]
		if !ok {
			return code
		}
	}
	return fmt.Sprintf(format, args...)
}

// Codes lists every message code (those of the English catalog), sorted.
func Codes() []string {
	codes := make([]string, 0, len(catalog[English]))
	for code := range catalog[English] {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// ParseAcceptLanguage returns the best supported language for an Accept-Language header:
// the highest q-value among English and Spanish (regional variants count: "es-ES" is Spanish),
// the first listed on a tie. English when the header is empty, malformed, "*", or asks only for
// unsupported languages or for q=0.
func ParseAcceptLanguage(header string) Lang {
	best, bestQ := English, 0.0
	for _, part := range strings.Split(header, ",") {
		tag, q := parseRange(part)
		if q <= 0 {
			continue
		}
		var lang Lang
		switch primary(tag) {
		case "es":
			lang = Spanish
		case "en":
			lang = English
		default:
			continue
		}
		if q > bestQ {
			best, bestQ = lang, q
		}
	}
	return best
}

func parseRange(part string) (tag string, q float64) {
	fields := strings.Split(strings.TrimSpace(part), ";")
	tag, q = strings.TrimSpace(fields[0]), 1.0
	for _, p := range fields[1:] {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "q=") {
			v, err := strconv.ParseFloat(strings.TrimPrefix(p, "q="), 64)
			if err != nil {
				return tag, 0
			}
			q = v
		}
	}
	return tag, q
}

func primary(tag string) string {
	tag = strings.ToLower(tag)
	if i := strings.IndexAny(tag, "-_"); i >= 0 {
		tag = tag[:i]
	}
	return tag
}
