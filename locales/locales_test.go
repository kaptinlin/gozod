package locales

import (
	"errors"
	"testing"

	"github.com/kaptinlin/gozod"
)

// TestEnglishLocale tests basic English localization
func TestEnglishLocale(t *testing.T) {
	issue := gozod.NewRawIssue("invalid_type", 123,
		gozod.WithExpected("string"),
	)

	message := En()(issue)
	expected := "Invalid input: expected string, received number"
	if message != expected {
		t.Errorf("Expected %q, got %q", expected, message)
	}
}

// TestChineseLocale tests basic Chinese localization
func TestChineseLocale(t *testing.T) {
	issue := gozod.NewRawIssue("invalid_type", 123,
		gozod.WithExpected("string"),
	)

	message := ZhCN()(issue)
	expected := "无效输入：期望 string，实际接收 数字"
	if message != expected {
		t.Errorf("Expected %q, got %q", expected, message)
	}
}

// TestDefaultLocales tests the default locale configuration
func TestDefaultLocales(t *testing.T) {
	// Test that default locales are properly initialized
	if len(DefaultLocales) == 0 {
		t.Error("DefaultLocales should not be empty")
	}

	if _, exists := DefaultLocales["en"]; !exists {
		t.Error("English locale should exist in DefaultLocales")
	}

	if _, exists := DefaultLocales["zh-CN"]; !exists {
		t.Error("Chinese locale should exist in DefaultLocales")
	}

	if _, exists := DefaultLocales["zh"]; !exists {
		t.Error("Chinese alias should exist in DefaultLocales")
	}
}

// TestGetLocalizedError tests the GetLocalizedError function
func TestGetLocalizedError(t *testing.T) {
	issue := gozod.NewRawIssue("invalid_type", 123,
		gozod.WithExpected("string"),
	)

	// Test English
	enMessage := GetLocalizedError(issue, "en")
	if enMessage == "" {
		t.Error("English message should not be empty")
	}

	// Test Chinese
	zhMessage := GetLocalizedError(issue, "zh-CN")
	if zhMessage == "" {
		t.Error("Chinese message should not be empty")
	}

	// Test fallback to English for unknown locale
	fallbackMessage := GetLocalizedError(issue, "unknown")
	if fallbackMessage != enMessage {
		t.Error("Unknown locale should fallback to English")
	}
}

// TestComprehensiveErrorTypes tests all error types in both languages
func TestComprehensiveErrorTypes(t *testing.T) {
	testCases := []struct {
		name       string
		issue      gozod.ZodRawIssue
		expectedEn string
		expectedZh string
	}{
		{
			name: "invalid_type",
			issue: gozod.NewRawIssue("invalid_type", 42,
				gozod.WithExpected("string"),
				gozod.WithReceived("number"),
			),
			expectedEn: "Invalid input: expected string, received number",
			expectedZh: "无效输入：期望 string，实际接收 数字",
		},
		{
			name: "too_small_string",
			issue: gozod.NewRawIssue("too_small", "hi",
				gozod.WithOrigin("string"),
				gozod.WithMinimum(5),
				gozod.WithInclusive(true),
			),
			expectedEn: "Too small: expected string to have >=5 characters",
			expectedZh: "数值过小：期望 string >=5 字符",
		},
		{
			name: "too_big_string",
			issue: gozod.NewRawIssue("too_big", "hello world",
				gozod.WithOrigin("string"),
				gozod.WithMaximum(3),
				gozod.WithInclusive(true),
			),
			expectedEn: "Too big: expected string to have <=3 characters",
			expectedZh: "数值过大：期望 string <=3 字符",
		},
		{
			name: "invalid_format_email",
			issue: gozod.NewRawIssue("invalid_format", "not-an-email",
				gozod.WithFormat("email"),
			),
			expectedEn: "Invalid email address",
			expectedZh: "无效电子邮件",
		},
		{
			name: "invalid_format_url",
			issue: gozod.NewRawIssue("invalid_format", "not-a-url",
				gozod.WithFormat("url"),
			),
			expectedEn: "Invalid URL",
			expectedZh: "无效URL",
		},
		{
			name: "invalid_format_starts_with",
			issue: gozod.NewRawIssue("invalid_format", "hello",
				gozod.WithFormat("starts_with"),
				gozod.WithPrefix("prefix"),
			),
			expectedEn: "Invalid string: must start with \"prefix\"",
			expectedZh: "无效字符串：必须以 \"prefix\" 开头",
		},
		{
			name: "invalid_format_ends_with",
			issue: gozod.NewRawIssue("invalid_format", "hello",
				gozod.WithFormat("ends_with"),
				gozod.WithSuffix("suffix"),
			),
			expectedEn: "Invalid string: must end with \"suffix\"",
			expectedZh: "无效字符串：必须以 \"suffix\" 结尾",
		},
		{
			name: "invalid_format_includes",
			issue: gozod.NewRawIssue("invalid_format", "hello",
				gozod.WithFormat("includes"),
				gozod.WithIncludes("world"),
			),
			expectedEn: "Invalid string: must include \"world\"",
			expectedZh: "无效字符串：必须包含 \"world\"",
		},
		{
			name: "invalid_format_regex",
			issue: gozod.NewRawIssue("invalid_format", "hello",
				gozod.WithFormat("regex"),
				gozod.WithPattern("[0-9]+"),
			),
			expectedEn: "Invalid string: must match pattern [0-9]+",
			expectedZh: "无效字符串：必须满足正则表达式 [0-9]+",
		},
		{
			name: "not_multiple_of",
			issue: gozod.NewRawIssue("not_multiple_of", 7,
				gozod.WithDivisor(5),
			),
			expectedEn: "Invalid number: must be a multiple of 5",
			expectedZh: "无效数字：必须是 5 的倍数",
		},
		{
			name: "unrecognized_keys_single",
			issue: gozod.NewRawIssue("unrecognized_keys", map[string]interface{}{},
				gozod.WithKeys([]string{"extra"}),
			),
			expectedEn: "Unrecognized key: \"extra\"",
			expectedZh: "出现未知的键(key): \"extra\"",
		},
		{
			name: "unrecognized_keys_multiple",
			issue: gozod.NewRawIssue("unrecognized_keys", map[string]interface{}{},
				gozod.WithKeys([]string{"extra1", "extra2"}),
			),
			expectedEn: "Unrecognized keys: \"extra1\", \"extra2\"",
			expectedZh: "出现未知的键(key): \"extra1\", \"extra2\"",
		},
		{
			name: "invalid_key",
			issue: gozod.NewRawIssue("invalid_key", "invalid",
				gozod.WithOrigin("object"),
			),
			expectedEn: "Invalid key in object",
			expectedZh: "object 中的键(key)无效",
		},
		{
			name:       "invalid_union",
			issue:      gozod.NewRawIssue("invalid_union", "invalid"),
			expectedEn: "Invalid input",
			expectedZh: "无效输入",
		},
		{
			name: "invalid_element",
			issue: gozod.NewRawIssue("invalid_element", "invalid",
				gozod.WithOrigin("array"),
			),
			expectedEn: "Invalid value in array",
			expectedZh: "array 中包含无效值(value)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test English
			enResult := En()(tc.issue)
			if enResult != tc.expectedEn {
				t.Errorf("English: expected %q, got %q", tc.expectedEn, enResult)
			}

			// Test Chinese
			zhResult := ZhCN()(tc.issue)
			if zhResult != tc.expectedZh {
				t.Errorf("Chinese: expected %q, got %q", tc.expectedZh, zhResult)
			}
		})
	}
}

// TestNumericConstraints tests numeric constraint error messages
func TestNumericConstraints(t *testing.T) {
	t.Run("too_small_number", func(t *testing.T) {
		issue := gozod.NewRawIssue("too_small", 16,
			gozod.WithOrigin("number"),
			gozod.WithMinimum(18),
			gozod.WithInclusive(true),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Too small: expected number to be >=18"
		expectedZh := "数值过小：期望 number >=18"

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})

	t.Run("too_big_number", func(t *testing.T) {
		issue := gozod.NewRawIssue("too_big", 70,
			gozod.WithOrigin("number"),
			gozod.WithMaximum(65),
			gozod.WithInclusive(true),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Too big: expected number to be <=65"
		expectedZh := "数值过大：期望 number <=65"

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})
}

// TestArrayConstraints tests array constraint error messages
func TestArrayConstraints(t *testing.T) {
	t.Run("too_small_array", func(t *testing.T) {
		issue := gozod.NewRawIssue("too_small", []string{"a", "b"},
			gozod.WithOrigin("array"),
			gozod.WithMinimum(3),
			gozod.WithInclusive(true),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Too small: expected array to have >=3 items"
		expectedZh := "数值过小：期望 array >=3 项"

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})
}

// TestInvalidValueMessages tests invalid_value error messages
func TestInvalidValueMessages(t *testing.T) {
	t.Run("single_value", func(t *testing.T) {
		issue := gozod.NewRawIssue("invalid_value", "invalid",
			gozod.WithValues([]interface{}{"valid"}),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Invalid input: expected \"valid\""
		expectedZh := "无效输入：期望 \"valid\""

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})

	t.Run("multiple_values", func(t *testing.T) {
		issue := gozod.NewRawIssue("invalid_value", "invalid",
			gozod.WithValues([]interface{}{"option1", "option2", "option3"}),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Invalid option: expected one of \"option1\"|\"option2\"|\"option3\""
		expectedZh := "无效选项：期望以下之一 \"option1\"|\"option2\"|\"option3\""

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("unknown_error_code", func(t *testing.T) {
		issue := gozod.NewRawIssue("unknown_code", "test")

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Invalid input"
		expectedZh := "无效输入"

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})

	t.Run("empty_origin", func(t *testing.T) {
		issue := gozod.NewRawIssue("too_small", "test",
			gozod.WithOrigin(""),
			gozod.WithMinimum(5),
			gozod.WithInclusive(true),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		// Should handle empty origin gracefully
		if enResult == "" {
			t.Error("English result should not be empty")
		}
		if zhResult == "" {
			t.Error("Chinese result should not be empty")
		}
	})

	t.Run("exclusive_boundaries", func(t *testing.T) {
		issue := gozod.NewRawIssue("too_small", 5,
			gozod.WithOrigin("number"),
			gozod.WithMinimum(5),
			gozod.WithInclusive(false),
		)

		enResult := En()(issue)
		zhResult := ZhCN()(issue)

		expectedEn := "Too small: expected number to be >5"
		expectedZh := "数值过小：期望 number >5"

		if enResult != expectedEn {
			t.Errorf("English: expected %q, got %q", expectedEn, enResult)
		}
		if zhResult != expectedZh {
			t.Errorf("Chinese: expected %q, got %q", expectedZh, zhResult)
		}
	})
}

// TestIntegrationWithGozod tests integration with actual GoZod schemas
func TestIntegrationWithGozod(t *testing.T) {
	t.Run("string_validation_with_chinese_locale", func(t *testing.T) {
		schema := gozod.String().Min(5, gozod.SchemaParams{
			Error: ZhCN(),
		})

		_, err := schema.Parse("hi")
		if err == nil {
			t.Error("Expected validation error")
			return
		}

		var zodErr *gozod.ZodError
		if !errors.As(err, &zodErr) {
			t.Error("Expected ZodError")
			return
		}

		if len(zodErr.Issues) == 0 {
			t.Error("Expected at least one issue")
			return
		}

		issue := zodErr.Issues[0]
		if issue.Message == "" {
			t.Error("Expected non-empty Chinese error message")
		}

		// Should contain Chinese characters
		if !containsChinese(issue.Message) {
			t.Errorf("Expected Chinese message, got: %s", issue.Message)
		}
	})

	t.Run("global_locale_configuration", func(t *testing.T) {
		// Save original config
		originalConfig := gozod.GetConfig()

		// Set Chinese locale globally
		gozod.Config(&gozod.ZodConfig{
			LocaleError: ZhCN(),
		})

		schema := gozod.String().Min(5)
		_, err := schema.Parse("hi")

		// Restore original config
		gozod.Config(originalConfig)

		if err == nil {
			t.Error("Expected validation error")
			return
		}

		var zodErr *gozod.ZodError
		if !errors.As(err, &zodErr) {
			t.Error("Expected ZodError")
			return
		}

		if len(zodErr.Issues) == 0 {
			t.Error("Expected at least one issue")
			return
		}

		issue := zodErr.Issues[0]
		if !containsChinese(issue.Message) {
			t.Errorf("Expected Chinese message from global config, got: %s", issue.Message)
		}
	})
}

// TestTypeScriptZodCompatibility tests compatibility with TypeScript Zod v4
func TestTypeScriptZodCompatibility(t *testing.T) {
	compatibilityTests := []struct {
		name       string
		issue      gozod.ZodRawIssue
		expectedEn string
		expectedZh string
	}{
		{
			name: "typescript_invalid_type_format",
			issue: gozod.NewRawIssue("invalid_type", 42,
				gozod.WithExpected("string"),
				gozod.WithReceived("number"),
			),
			expectedEn: "Invalid input: expected string, received number",
			expectedZh: "无效输入：期望 string，实际接收 数字",
		},
		{
			name: "typescript_too_small_format",
			issue: gozod.NewRawIssue("too_small", "hi",
				gozod.WithOrigin("string"),
				gozod.WithMinimum(5),
				gozod.WithInclusive(true),
			),
			expectedEn: "Too small: expected string to have >=5 characters",
			expectedZh: "数值过小：期望 string >=5 字符",
		},
		{
			name: "typescript_email_format",
			issue: gozod.NewRawIssue("invalid_format", "not-an-email",
				gozod.WithFormat("email"),
			),
			expectedEn: "Invalid email address",
			expectedZh: "无效电子邮件",
		},
	}

	for _, tc := range compatibilityTests {
		t.Run(tc.name, func(t *testing.T) {
			enResult := En()(tc.issue)
			zhResult := ZhCN()(tc.issue)

			if enResult != tc.expectedEn {
				t.Errorf("TypeScript compatibility failed (EN): expected %q, got %q", tc.expectedEn, enResult)
			}
			if zhResult != tc.expectedZh {
				t.Errorf("TypeScript compatibility failed (ZH): expected %q, got %q", tc.expectedZh, zhResult)
			}
		})
	}
}

// Helper function to check if a string contains Chinese characters
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}
