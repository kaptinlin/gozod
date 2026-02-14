package locales

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/core"
)

// =============================================================================
// LOCALE FORMATTER TESTS - UPDATED FOR FUNCTIONAL APPROACH
// =============================================================================

func TestLocaleFormatterFunction(t *testing.T) {
	t.Run("formats message using formatter function", func(t *testing.T) {
		formatFunc := func(issue core.ZodRawIssue) string {
			return "custom error: " + string(issue.Code)
		}

		issue := core.ZodRawIssue{Code: core.InvalidType}
		result := formatFunc(issue)
		expected := "custom error: invalid_type"

		assert.Equal(t, expected, result, "Expected '%s', got '%s'", expected, result)
	})
}

// =============================================================================
// LOCALE REGISTRY TESTS
// =============================================================================

func TestLocaleFormatter(t *testing.T) {
	t.Run("returns existing formatter", func(t *testing.T) {
		formatter := LocaleFormatter("en")

		if formatter == nil {
			t.Error("Expected formatter, got nil")
		}

		// Test that it actually formats messages
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: map[string]any{"expected": "string"},
			Input:      123,
		}

		message := formatter(issue)
		if message == "" {
			t.Error("Expected non-empty message")
		}
	})

	t.Run("returns fallback for unknown locale", func(t *testing.T) {
		formatter := LocaleFormatter("unknown-locale")

		if formatter == nil {
			t.Error("Expected fallback formatter, got nil")
		}

		// Should fallback to English
		enFormatter := LocaleFormatter("en")
		issue := core.ZodRawIssue{Code: core.InvalidType}

		if formatter(issue) != enFormatter(issue) {
			t.Error("Expected fallback to English formatter")
		}
	})

	t.Run("handles language-only codes", func(t *testing.T) {
		// Test that an unknown regional variant falls back to the base language
		// e.g., "de-AT" (Austrian German) should fall back to "de" (German)
		formatter := LocaleFormatter("de-AT")
		deFormatter := LocaleFormatter("de")

		issue := core.ZodRawIssue{Code: core.InvalidType}

		if formatter(issue) != deFormatter(issue) {
			t.Error("Expected language-only fallback to work")
		}
	})

	t.Run("zh-TW has its own formatter", func(t *testing.T) {
		// Test that zh-TW now has its own formatter (Traditional Chinese)
		zhTwFormatter := LocaleFormatter("zh-TW")
		zhCnFormatter := LocaleFormatter("zh-CN")

		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: map[string]any{"expected": "string"},
			Input:      123,
		}

		zhTwMessage := zhTwFormatter(issue)
		zhCnMessage := zhCnFormatter(issue)

		// Traditional and Simplified Chinese should have different messages
		if zhTwMessage == zhCnMessage {
			t.Error("Expected zh-TW and zh-CN to have different messages")
		}
	})
}

func TestLocalizedError(t *testing.T) {
	t.Run("returns localized error message", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: map[string]any{"expected": "string"},
			Input:      123,
		}

		enMessage := LocalizedError(issue, "en")
		zhMessage := LocalizedError(issue, "zh-CN")

		if enMessage == "" {
			t.Error("Expected non-empty English message")
		}

		if zhMessage == "" {
			t.Error("Expected non-empty Chinese message")
		}

		// Messages should be different for different locales
		if enMessage == zhMessage {
			t.Error("Expected different messages for different locales")
		}
	})
}

func TestRegisterLocale(t *testing.T) {
	t.Run("registers new locale", func(t *testing.T) {
		customFormat := func(issue core.ZodRawIssue) string {
			return "CUSTOM: " + string(issue.Code)
		}

		RegisterLocale("custom", customFormat)

		formatter := LocaleFormatter("custom")
		issue := core.ZodRawIssue{Code: core.InvalidType}

		result := formatter(issue)
		expected := "CUSTOM: invalid_type"

		assert.Equal(t, expected, result, "Expected '%s', got '%s'", expected, result)

		// Clean up
		delete(DefaultLocales, "custom")
	})
}

func TestAvailableLocales(t *testing.T) {
	t.Run("returns list of available locales", func(t *testing.T) {
		locales := AvailableLocales()

		if len(locales) == 0 {
			t.Error("Expected at least one locale")
		}

		// Should contain at least "en"
		found := slices.Contains(locales, "en")

		if !found {
			t.Error("Expected 'en' locale to be available")
		}
	})
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestLocalizedErrors(t *testing.T) {
	t.Run("returns localized errors for multiple issues", func(t *testing.T) {
		issues := []core.ZodRawIssue{
			{Code: core.InvalidType, Properties: map[string]any{"expected": "string"}},
			{Code: core.TooSmall, Properties: map[string]any{"minimum": 5}},
		}

		messages, err := LocalizedErrors(issues, "en")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(messages))
		}

		for _, message := range messages {
			if message == "" {
				t.Error("Expected non-empty message")
			}
		}
	})

	t.Run("handles empty issues slice", func(t *testing.T) {
		messages, err := LocalizedErrors([]core.ZodRawIssue{}, "en")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if messages != nil {
			t.Error("Expected nil for empty issues")
		}
	})
}

func TestFilterUniqueLocales(t *testing.T) {
	t.Run("removes duplicate locales", func(t *testing.T) {
		locales := []string{"en", "zh-CN", "en", "zh", "zh-CN"}

		unique, err := FilterUniqueLocales(locales)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(unique) >= len(locales) {
			t.Error("Expected fewer unique locales than input")
		}

		// Check that all unique locales are present
		localeMap := make(map[string]bool)
		for _, locale := range unique {
			localeMap[locale] = true
		}

		expectedLocales := []string{"en", "zh-CN", "zh"}
		for _, expected := range expectedLocales {
			if !localeMap[expected] {
				t.Errorf("Expected locale '%s' to be in unique list", expected)
			}
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		unique, err := FilterUniqueLocales([]string{})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if unique != nil {
			t.Error("Expected nil for empty input")
		}
	})
}

func TestValidateLocaleList(t *testing.T) {
	t.Run("separates valid and invalid locales", func(t *testing.T) {
		locales := []string{"en", "zh-CN", "invalid-locale", "zh", "another-invalid"}

		valid, invalid, err := ValidateLocaleList(locales)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedValid := map[string]bool{"en": true, "zh-CN": true, "zh": true}
		for _, locale := range valid {
			if !expectedValid[locale] {
				t.Errorf("Unexpected valid locale: %s", locale)
			}
		}

		expectedInvalid := map[string]bool{"invalid-locale": true, "another-invalid": true}
		for _, locale := range invalid {
			if !expectedInvalid[locale] {
				t.Errorf("Unexpected invalid locale: %s", locale)
			}
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		valid, invalid, err := ValidateLocaleList([]string{})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if valid != nil || invalid != nil {
			t.Error("Expected nil for empty input")
		}
	})
}

func TestJoinLocalizedMessages(t *testing.T) {
	t.Run("joins multiple localized messages", func(t *testing.T) {
		issues := []core.ZodRawIssue{
			{Code: core.InvalidType, Properties: map[string]any{"expected": "string"}},
			{Code: core.TooSmall, Properties: map[string]any{"minimum": 5}},
		}

		joined, err := JoinLocalizedMessages(issues, "en", "; ")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if joined == "" {
			t.Error("Expected non-empty joined message")
		}

		// Should contain the separator
		if !strings.Contains(joined, "; ") {
			t.Error("Expected joined message to contain separator")
		}
	})
}
