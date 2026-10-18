package middleware

import (
	"hopfield-assignment-api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// LangKey is the Gin context key holding the response language chosen from Accept-Language.
const LangKey = "lang"

// Language picks the response language (English or Spanish) from the Accept-Language header.
func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := i18n.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set(LangKey, lang)
		c.Header("Content-Language", string(lang))
		c.Writer.Header().Add("Vary", "Accept-Language")
		c.Next()
	}
}

// LanguageFrom returns the language Language() chose for this request, English if it did not run.
func LanguageFrom(c *gin.Context) i18n.Lang {
	if v, ok := c.Get(LangKey); ok {
		if lang, ok := v.(i18n.Lang); ok {
			return lang
		}
	}
	return i18n.English
}
