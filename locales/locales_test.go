package locales

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

// =============================================================================
// LOCALE FORMATTER TESTS
// =============================================================================

func TestNewLocaleFormatter(t *testing.T) {
	t.Run("creates formatter with correct locale", func(t *testing.T) {
		formatFunc := func(issue core.ZodRawIssue) string {
			return "test message"
		}

		formatter := NewLocaleFormatter("test-locale", formatFunc)

		if formatter.locale != "test-locale" {
			t.Errorf("Expected locale 'test-locale', got '%s'", formatter.locale)
		}

		if formatter.GetLocale() != "test-locale" {
			t.Errorf("Expected GetLocale() to return 'test-locale', got '%s'", formatter.GetLocale())
		}
	})
}

func TestLocaleFormatter_FormatMessage(t *testing.T) {
	t.Run("formats message using provided function", func(t *testing.T) {
		formatFunc := func(issue core.ZodRawIssue) string {
			return "custom error: " + string(issue.Code)
		}

		formatter := NewLocaleFormatter("test", formatFunc)
		issue := core.ZodRawIssue{Code: core.InvalidType}

		result := formatter.FormatMessage(issue)
		expected := "custom error: invalid_type"

		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

// =============================================================================
// LOCALE REGISTRY TESTS
// =============================================================================

func TestGetLocaleFormatter(t *testing.T) {
	t.Run("returns existing formatter", func(t *testing.T) {
		formatter := GetLocaleFormatter("en")

		if formatter == nil {
			t.Error("Expected formatter, got nil")
		}

		// Test that it actually formats messages
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: map[string]any{"expected": "string"},
			Input:      123,
		}

		message := formatter.FormatMessage(issue)
		if message == "" {
			t.Error("Expected non-empty message")
		}
	})

	t.Run("returns fallback for unknown locale", func(t *testing.T) {
		formatter := GetLocaleFormatter("unknown-locale")

		if formatter == nil {
			t.Error("Expected fallback formatter, got nil")
		}

		// Should fallback to English
		enFormatter := GetLocaleFormatter("en")
		issue := core.ZodRawIssue{Code: core.InvalidType}

		if formatter.FormatMessage(issue) != enFormatter.FormatMessage(issue) {
			t.Error("Expected fallback to English formatter")
		}
	})

	t.Run("handles language-only codes", func(t *testing.T) {
		// Test that "zh-TW" falls back to "zh"
		formatter := GetLocaleFormatter("zh-TW")
		zhFormatter := GetLocaleFormatter("zh")

		issue := core.ZodRawIssue{Code: core.InvalidType}

		if formatter.FormatMessage(issue) != zhFormatter.FormatMessage(issue) {
			t.Error("Expected language-only fallback to work")
		}
	})
}

func TestGetLocalizedError(t *testing.T) {
	t.Run("returns localized error message", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: map[string]any{"expected": "string"},
			Input:      123,
		}

		enMessage := GetLocalizedError(issue, "en")
		zhMessage := GetLocalizedError(issue, "zh-CN")

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

		formatter := GetLocaleFormatter("custom")
		issue := core.ZodRawIssue{Code: core.InvalidType}

		result := formatter.FormatMessage(issue)
		expected := "CUSTOM: invalid_type"

		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Clean up
		delete(DefaultLocales, "custom")
	})
}

func TestGetAvailableLocales(t *testing.T) {
	t.Run("returns list of available locales", func(t *testing.T) {
		locales := GetAvailableLocales()

		if len(locales) == 0 {
			t.Error("Expected at least one locale")
		}

		// Should contain at least "en"
		found := false
		for _, locale := range locales {
			if locale == "en" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected 'en' locale to be available")
		}
	})
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestGetLocalizedErrors(t *testing.T) {
	t.Run("returns localized errors for multiple issues", func(t *testing.T) {
		issues := []core.ZodRawIssue{
			{Code: core.InvalidType, Properties: map[string]any{"expected": "string"}},
			{Code: core.TooSmall, Properties: map[string]any{"minimum": 5}},
		}

		messages, err := GetLocalizedErrors(issues, "en")

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
		messages, err := GetLocalizedErrors([]core.ZodRawIssue{}, "en")

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
		if !contains(joined, "; ") {
			t.Error("Expected joined message to contain separator")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
