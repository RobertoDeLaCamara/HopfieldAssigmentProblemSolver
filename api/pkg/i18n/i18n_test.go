package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAcceptLanguage(t *testing.T) {
	cases := []struct {
		header string
		want   Lang
	}{
		{"", English},
		{"es", Spanish},
		{"ES", Spanish},
		{"es-ES", Spanish},
		{"es_MX", Spanish},
		{"en", English},
		{"en-US,en;q=0.9", English},
		{"es-ES,es;q=0.9,en;q=0.8", Spanish},
		{"en-US,en;q=0.9,es;q=0.8", English},
		{"fr-FR,fr;q=0.9,es;q=0.7", Spanish}, // best supported one
		{"fr,de;q=0.8", English},             // nothing supported -> default
		{"*", English},
		{"es;q=0", English}, // q=0 means "not acceptable"
		{"es;q=0, en", English},
		{"en;q=0.5, es;q=0.5", English}, // tie -> first listed
		{"es;q=0.5, en;q=0.5", Spanish},
		{"es;q=notanumber", English}, // malformed q is ignored, not trusted
		{";;;,,,", English},
	}
	for _, tc := range cases {
		t.Run(tc.header, func(t *testing.T) {
			assert.Equal(t, tc.want, ParseAcceptLanguage(tc.header))
		})
	}
}

func TestT_FallsBackToEnglish(t *testing.T) {
	assert.Equal(t, "The cost matrix cannot be empty", T(English, MatrixEmpty))
	assert.Equal(t, "La matriz de costes no puede estar vacía", T(Spanish, MatrixEmpty))
	assert.Equal(t, "The cost matrix cannot be empty", T(Lang("fr"), MatrixEmpty), "unsupported language")
	assert.Equal(t, "no_such_code", T(Spanish, "no_such_code"), "unknown code is returned as is, never empty")
}

func TestT_FormatsArguments(t *testing.T) {
	assert.Equal(t, "Cost at position [1][2] cannot be negative: -3.000000", T(English, CostNegative, 1, 2, -3.0))
	assert.Equal(t, "El coste en la posición [1][2] no puede ser negativo: -3.000000", T(Spanish, CostNegative, 1, 2, -3.0))
}

// A Spanish message that takes different arguments than its English twin would print %!d(MISSING) or
// extra garbage; this keeps the two catalogs in step.
func TestCatalogsHaveTheSameCodesAndVerbs(t *testing.T) {
	verbs := func(s string) string {
		var out []string
		for i := 0; i < len(s)-1; i++ {
			if s[i] == '%' {
				out = append(out, string(s[i+1]))
				i++
			}
		}
		return strings.Join(out, "")
	}
	assert.Equal(t, len(catalog[English]), len(catalog[Spanish]), "same number of codes")
	for _, code := range Codes() {
		es, ok := catalog[Spanish][code]
		if assert.True(t, ok, "Spanish translation for %q", code) {
			assert.Equal(t, verbs(catalog[English][code]), verbs(es), "format verbs of %q", code)
			assert.NotEqual(t, catalog[English][code], es, "%q is translated", code)
		}
	}
}
