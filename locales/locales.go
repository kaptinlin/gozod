package locales

import "github.com/kaptinlin/gozod"

// LocaleMap maps locale names to error message functions
type LocaleMap map[string]func(gozod.ZodRawIssue) string

// DefaultLocales contains the default supported locales
var DefaultLocales = LocaleMap{
	"en":    En(),
	"zh-CN": ZhCN(),
	"zh":    ZhCN(),
}

// GetLocalizedError returns a localized error message for the given issue and locale
// Falls back to English if the locale is not found
func GetLocalizedError(issue gozod.ZodRawIssue, locale string) string {
	if fn, exists := DefaultLocales[locale]; exists {
		return fn(issue)
	}
	// Fallback to English
	return En()(issue)
}
