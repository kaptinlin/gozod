package locales

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// LOCALE SYSTEM - UNIFIED MESSAGE FORMATTING WITH TYPESCRIPT ZOD V4 COMPATIBILITY
// =============================================================================

// LocaleFormatter implements MessageFormatter for specific locales
// Provides localized error messages following TypeScript Zod v4 patterns
type LocaleFormatter struct {
	locale     string                        // Locale identifier (e.g., "en", "zh-CN")
	formatFunc func(core.ZodRawIssue) string // Locale-specific formatting function
}

// NewLocaleFormatter creates a new locale-specific formatter
// This allows for easy registration of new locales with custom formatting logic
func NewLocaleFormatter(locale string, formatFunc func(core.ZodRawIssue) string) *LocaleFormatter {
	return &LocaleFormatter{
		locale:     locale,
		formatFunc: formatFunc,
	}
}

// FormatMessage implements the MessageFormatter interface
// Delegates to the locale-specific formatting function
func (lf *LocaleFormatter) FormatMessage(raw core.ZodRawIssue) string {
	return lf.formatFunc(raw)
}

// GetLocale returns the locale identifier for this formatter
func (lf *LocaleFormatter) GetLocale() string {
	return lf.locale
}

// =============================================================================
// LOCALE REGISTRY AND MANAGEMENT
// =============================================================================

// LocaleMap maps locale names to their corresponding formatters
type LocaleMap map[string]issues.MessageFormatter

// DefaultLocales contains the default supported locales using the new formatter system
// Supports both full locale codes (zh-CN) and short codes (zh, en)
var DefaultLocales = LocaleMap{
	"en":    NewLocaleFormatter("en", formatEn),      // English
	"zh-CN": NewLocaleFormatter("zh-CN", formatZhCN), // Simplified Chinese (China)
	"zh":    NewLocaleFormatter("zh", formatZhCN),    // Chinese fallback
}

// GetLocaleFormatter returns a formatter for the given locale
// Falls back to English if the locale is not found, ensuring robust operation
func GetLocaleFormatter(locale string) issues.MessageFormatter {
	if formatter, exists := DefaultLocales[locale]; exists {
		return formatter
	}

	// Try fallback for language-only codes (e.g., "zh" for "zh-CN")
	if len(locale) > 2 && locale[2] == '-' {
		langCode := locale[:2]
		if formatter, exists := DefaultLocales[langCode]; exists {
			return formatter
		}
	}

	// Final fallback to English
	return DefaultLocales["en"]
}

// GetLocalizedError returns a localized error message for the given issue and locale
// This is a convenience function for backward compatibility and simple usage
func GetLocalizedError(issue core.ZodRawIssue, locale string) string {
	formatter := GetLocaleFormatter(locale)
	return formatter.FormatMessage(issue)
}

// RegisterLocale adds a new locale to the default locales map
// Allows runtime registration of additional locales and custom formatters
func RegisterLocale(locale string, formatFunc func(core.ZodRawIssue) string) {
	DefaultLocales[locale] = NewLocaleFormatter(locale, formatFunc)
}

// GetAvailableLocales returns a list of all registered locale identifiers
// Useful for UI components that need to display available localization options
// Now uses slicex for better slice handling
func GetAvailableLocales() []string {
	locales := make([]string, 0, len(DefaultLocales))
	for locale := range DefaultLocales {
		locales = append(locales, locale)
	}

	// Use slicex to sort and deduplicate locales
	if sortedLocales, err := slicex.Unique(locales); err == nil {
		if typedLocales, err := slicex.ToTyped[string](sortedLocales); err == nil {
			return typedLocales
		}
	}

	return locales
}

// =============================================================================
// BACKWARD COMPATIBILITY FUNCTIONS
// =============================================================================

// En returns the English locale formatter function for backward compatibility
// This function matches the old interface: func(ZodRawIssue) string
// Preserves existing API while integrating with the new formatter system
func En() func(core.ZodRawIssue) string {
	return formatEn
}

// ZhCN returns the Chinese locale formatter function for backward compatibility
// This function matches the old interface: func(ZodRawIssue) string
// Preserves existing API while integrating with the new formatter system
func ZhCN() func(core.ZodRawIssue) string {
	return formatZhCN
}

// =============================================================================
// SLICEX-ENHANCED UTILITY FUNCTIONS
// =============================================================================

// GetLocalizedErrors returns localized error messages for multiple issues
// Uses slicex for efficient slice processing
func GetLocalizedErrors(issues []core.ZodRawIssue, locale string) ([]string, error) {
	if slicex.IsEmpty(issues) {
		return nil, nil
	}

	formatter := GetLocaleFormatter(locale)
	messages, err := slicex.Map(issues, func(issue any) any {
		if rawIssue, ok := issue.(core.ZodRawIssue); ok {
			return formatter.FormatMessage(rawIssue)
		}
		return "Invalid input"
	})

	if err != nil {
		return nil, err
	}

	return slicex.ToTyped[string](messages)
}

// FilterUniqueLocales removes duplicate locales from a slice
// Demonstrates slicex.Unique functionality
func FilterUniqueLocales(locales []string) ([]string, error) {
	if slicex.IsEmpty(locales) {
		return nil, nil
	}

	uniqueLocales, err := slicex.Unique(locales)
	if err != nil {
		return nil, err
	}

	return slicex.ToTyped[string](uniqueLocales)
}

// ValidateLocaleList checks if all locales in a slice are supported
// Returns valid locales and invalid ones separately using slicex.Filter
func ValidateLocaleList(locales []string) (valid []string, invalid []string, err error) {
	if slicex.IsEmpty(locales) {
		return nil, nil, nil
	}

	// Filter valid locales
	validAny, err := slicex.Filter(locales, func(locale any) bool {
		if localeStr, ok := locale.(string); ok {
			_, exists := DefaultLocales[localeStr]
			return exists
		}
		return false
	})
	if err != nil {
		return nil, nil, err
	}

	// Filter invalid locales
	invalidAny, err := slicex.Filter(locales, func(locale any) bool {
		if localeStr, ok := locale.(string); ok {
			_, exists := DefaultLocales[localeStr]
			return !exists
		}
		return false
	})
	if err != nil {
		return nil, nil, err
	}

	valid, err = slicex.ToTyped[string](validAny)
	if err != nil {
		return nil, nil, err
	}

	invalid, err = slicex.ToTyped[string](invalidAny)
	if err != nil {
		return nil, nil, err
	}

	return valid, invalid, nil
}

// JoinLocalizedMessages joins multiple localized messages with a separator
// Utilizes slicex.Join for consistent formatting
func JoinLocalizedMessages(issues []core.ZodRawIssue, locale string, separator string) (string, error) {
	messages, err := GetLocalizedErrors(issues, locale)
	if err != nil {
		return "", err
	}

	if slicex.IsEmpty(messages) {
		return "", nil
	}

	return slicex.Join(messages, separator), nil
}
