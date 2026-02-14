// Package locales provides pre-configured error message formatters for different languages.
// To use a locale, pass its factory function to gozod.SetConfig().
//
// Example:
//
//	gozod.SetConfig(locales.ZhCN()) // Switch to Chinese messages globally.
//	gozod.SetConfig(locales.EN())   // Switch back to English.
package locales

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// LOCALE SYSTEM - FUNCTIONAL APPROACH MATCHING TYPESCRIPT ZOD V4
// =============================================================================

// LocaleErrorMap maps locale names to their corresponding formatter functions
// Following TypeScript Zod v4's functional approach instead of struct-based patterns
type LocaleErrorMap map[string]func(core.ZodRawIssue) string

// DefaultLocales contains the default supported locales using functional approach
// Supports both full locale codes (zh-CN) and short codes (zh, en)
// Following TypeScript Zod v4's pattern of mapping locales to formatter functions
var DefaultLocales = LocaleErrorMap{
	// English
	"en": formatEn,

	// Chinese
	"zh-CN": formatZhCN, // Simplified Chinese (China)
	"zh-TW": formatZhTw, // Traditional Chinese (Taiwan)
	"zh":    formatZhCN, // Chinese fallback to Simplified

	// European Languages
	"de": formatDe, // German
	"fr": formatFr, // French
	"es": formatEs, // Spanish
	"it": formatIt, // Italian
	"pt": formatPt, // Portuguese
	"nl": formatNl, // Dutch
	"pl": formatPl, // Polish
	"ru": formatRu, // Russian
	"uk": formatUk, // Ukrainian
	"cs": formatCs, // Czech
	"da": formatDa, // Danish
	"sv": formatSv, // Swedish
	"tr": formatTr, // Turkish
	"hu": formatHu, // Hungarian
	"fi": formatFi, // Finnish
	"no": formatNo, // Norwegian
	"bg": formatBg, // Bulgarian

	// Asian Languages
	"ja": formatJa, // Japanese
	"ko": formatKo, // Korean
	"vi": formatVi, // Vietnamese
	"th": formatTh, // Thai
	"id": formatId, // Indonesian
	"ms": formatMs, // Malay
	"ta": formatTa, // Tamil

	// Middle Eastern Languages
	"ar": formatAr, // Arabic
	"fa": formatFa, // Persian (Farsi)
	"he": formatHe, // Hebrew
	"ur": formatUr, // Urdu
}

// LocaleFormatter returns a formatter function for the given locale.
// Falls back to English if the locale is not found, ensuring robust operation.
func LocaleFormatter(locale string) func(core.ZodRawIssue) string {
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

// LocalizedError returns a localized error message for the given issue and locale.
// This is a convenience function for backward compatibility and simple usage.
func LocalizedError(issue core.ZodRawIssue, locale string) string {
	formatter := LocaleFormatter(locale)
	return formatter(issue)
}

// RegisterLocale adds a new locale to the default locales map
// Allows runtime registration of additional locales and custom formatters
func RegisterLocale(locale string, formatFunc func(core.ZodRawIssue) string) {
	DefaultLocales[locale] = formatFunc
}

// AvailableLocales returns a list of all registered locale identifiers.
// Useful for UI components that need to display available localization options.
// Uses slicex for better slice handling.
func AvailableLocales() []string {
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
// SLICEX-ENHANCED UTILITY FUNCTIONS
// =============================================================================

// LocalizedErrors returns localized error messages for multiple issues.
// Uses slicex for efficient slice processing.
func LocalizedErrors(issues []core.ZodRawIssue, locale string) ([]string, error) {
	if slicex.IsEmpty(issues) {
		return nil, nil
	}

	formatter := LocaleFormatter(locale)
	messages, err := slicex.Map(issues, func(issue any) any {
		if rawIssue, ok := issue.(core.ZodRawIssue); ok {
			return formatter(rawIssue)
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

// JoinLocalizedMessages joins multiple localized messages with a separator.
// Utilizes slicex.Join for consistent formatting.
func JoinLocalizedMessages(issues []core.ZodRawIssue, locale string, separator string) (string, error) {
	messages, err := LocalizedErrors(issues, locale)
	if err != nil {
		return "", err
	}

	if slicex.IsEmpty(messages) {
		return "", nil
	}

	return slicex.Join(messages, separator), nil
}
