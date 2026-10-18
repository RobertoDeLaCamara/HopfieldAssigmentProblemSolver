package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hopfield-assignment-api/pkg/i18n"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLanguage_SetsContextAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Language())
	var got i18n.Lang
	r.GET("/x", func(c *gin.Context) { got = LanguageFrom(c); c.String(http.StatusOK, "ok") })

	for header, want := range map[string]i18n.Lang{"": i18n.English, "es-ES,es;q=0.9": i18n.Spanish, "fr": i18n.English} {
		req := httptest.NewRequest("GET", "/x", nil)
		if header != "" {
			req.Header.Set("Accept-Language", header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, want, got, header)
		assert.Equal(t, string(want), w.Header().Get("Content-Language"))
		assert.Contains(t, w.Header().Values("Vary"), "Accept-Language")
	}
}

func TestLanguageFrom_DefaultsToEnglishWithoutTheMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, i18n.English, LanguageFrom(c))
}
