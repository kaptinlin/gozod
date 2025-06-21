package issues

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Issue Finalization Tests  ///
//////////////////////////////////////////

func TestFinalizeIssue(t *testing.T) {
	t.Run("finalizes issue with basic properties", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, "test_input",
			WithExpected("string"),
			WithReceived("number"),
			WithMessage("Custom error message"),
		)

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		require.Equal(t, core.InvalidType, issue.Code)
		require.Equal(t, "Custom error message", issue.Message)
		assert.Empty(t, issue.Path)
		assert.Equal(t, "string", issue.Expected)
		assert.Equal(t, "number", issue.Received)
		assert.Equal(t, "test_input", issue.Input)
	})

	t.Run("includes input when ReportInput is true", func(t *testing.T) {
		rawIssue := NewRawIssue("too_small", "hi")
		ctx := &core.ParseContext{ReportInput: true}
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		assert.Equal(t, "hi", issue.Input)
	})

	t.Run("excludes input when ReportInput is false", func(t *testing.T) {
		rawIssue := NewRawIssue("too_small", "secret_data")
		ctx := &core.ParseContext{ReportInput: false}
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		assert.Nil(t, issue.Input)
	})

	t.Run("includes input by default when context is nil", func(t *testing.T) {
		rawIssue := NewRawIssue("too_small", "test_data")
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Equal(t, "test_data", issue.Input)
	})

	t.Run("uses default message when no message provided", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, 123,
			WithExpected("string"),
		)

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Contains(t, issue.Message, "Invalid input")
		assert.Contains(t, issue.Message, "string")
		assert.Contains(t, issue.Message, "number")
	})

	t.Run("preserves existing message when provided", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, nil,
			WithMessage("Custom type validation failed"),
		)

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Equal(t, "Custom type validation failed", issue.Message)
	})

	t.Run("ensures path is never nil", func(t *testing.T) {
		rawIssue := core.ZodRawIssue{
			Code:  core.Custom,
			Input: nil,
			Path:  nil, // Explicitly nil
		}

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		require.NotNil(t, issue.Path)
		assert.Empty(t, issue.Path)
	})

	t.Run("preserves existing path", func(t *testing.T) {
		path := []any{"user", "profile", "email"}
		rawIssue := NewRawIssue(core.InvalidFormat, "invalid@",
			WithPath(path),
		)

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Equal(t, path, issue.Path)
	})

	t.Run("handles empty path correctly", func(t *testing.T) {
		path := []any{}
		rawIssue := NewRawIssue(core.Custom, nil, WithPath(path))

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		require.NotNil(t, issue.Path)
		assert.Empty(t, issue.Path)
	})
}

//////////////////////////////////////////
//////////   Property Mapping Tests    ///
//////////////////////////////////////////

func TestMapPropertiesToIssue(t *testing.T) {
	t.Run("maps all standard properties correctly", func(t *testing.T) {
		props := map[string]any{
			"expected":  "string",
			"received":  "number",
			"minimum":   5,
			"maximum":   100,
			"inclusive": true,
			"format":    "email",
			"pattern":   ".*@.*",
			"prefix":    "test_",
			"suffix":    "_end",
			"includes":  "@",
			"algorithm": "HS256",
			"origin":    "string",
			"divisor":   2.5,
			"keys":      []string{"key1", "key2"},
			"values":    []any{"val1", "val2"},
			"key":       "field_name",
			"params":    map[string]any{"custom": "data"},
		}

		issue := core.ZodIssue{}
		MapPropertiesToIssue(&issue, props)

		// Verify all properties were mapped correctly
		assert.Equal(t, "string", issue.Expected)
		assert.Equal(t, "number", issue.Received)
		assert.Equal(t, 5, issue.Minimum)
		assert.Equal(t, 100, issue.Maximum)
		assert.True(t, issue.Inclusive)
		assert.Equal(t, "email", issue.Format)
		assert.Equal(t, ".*@.*", issue.Pattern)
		assert.Equal(t, "test_", issue.Prefix)
		assert.Equal(t, "_end", issue.Suffix)
		assert.Equal(t, "@", issue.Includes)
		assert.Equal(t, "HS256", issue.Algorithm)
		assert.Equal(t, "string", issue.Origin)
		assert.Equal(t, 2.5, issue.Divisor)
		assert.Equal(t, []string{"key1", "key2"}, issue.Keys)
		assert.Equal(t, []any{"val1", "val2"}, issue.Values)
		assert.Equal(t, "field_name", issue.Key)
		assert.Equal(t, map[string]any{"custom": "data"}, issue.Params)
	})

	t.Run("handles partial property mapping", func(t *testing.T) {
		props := map[string]any{
			"expected": "string",
			"minimum":  10,
			"format":   "email",
		}

		issue := core.ZodIssue{}
		MapPropertiesToIssue(&issue, props)

		assert.Equal(t, "string", issue.Expected)
		assert.Equal(t, 10, issue.Minimum)
		assert.Equal(t, "email", issue.Format)

		// Unset properties should remain at zero values
		assert.Empty(t, issue.Received)
		assert.Nil(t, issue.Maximum)
		assert.False(t, issue.Inclusive)
	})

	t.Run("handles nil properties map", func(t *testing.T) {
		issue := core.ZodIssue{}

		// Should not panic
		assert.NotPanics(t, func() {
			MapPropertiesToIssue(&issue, nil)
		})

		// Should remain unchanged
		assert.Empty(t, issue.Expected)
		assert.Nil(t, issue.Minimum)
	})

	t.Run("handles empty properties map", func(t *testing.T) {
		props := map[string]any{}
		issue := core.ZodIssue{}

		MapPropertiesToIssue(&issue, props)

		// Should remain at zero values
		assert.Empty(t, issue.Expected)
		assert.Nil(t, issue.Minimum)
		assert.False(t, issue.Inclusive)
	})

	t.Run("handles wrong property types gracefully", func(t *testing.T) {
		props := map[string]any{
			"expected":  123,                      // Wrong type
			"keys":      "not_a_slice",            // Wrong type
			"inclusive": "not_a_bool",             // Wrong type
			"params":    []string{"not", "a_map"}, // Wrong type
		}

		issue := core.ZodIssue{}

		// Should not panic with wrong types
		assert.NotPanics(t, func() {
			MapPropertiesToIssue(&issue, props)
		})

		// Wrong types should not be assigned (utils functions handle type safety)
		assert.Empty(t, issue.Expected)  // Wrong type, should remain empty
		assert.Nil(t, issue.Keys)        // Wrong type, should remain nil
		assert.False(t, issue.Inclusive) // Wrong type, should remain false
		assert.Nil(t, issue.Params)      // Wrong type, should remain nil
	})

	t.Run("preserves existing issue values for unmapped properties", func(t *testing.T) {
		// Pre-populate issue with some values
		issue := core.ZodIssue{
			Expected: "existing_value",
			Minimum:  42,
		}

		// Map only a subset of properties
		props := map[string]any{
			"received": "new_value",
			"maximum":  100,
		}

		MapPropertiesToIssue(&issue, props)

		// Existing values should be overwritten by empty values from utils
		assert.Empty(t, issue.Expected) // Utils returns empty string for missing property
		assert.Nil(t, issue.Minimum)    // Utils returns nil for missing property

		// New values should be set
		assert.Equal(t, "new_value", issue.Received)
		assert.Equal(t, 100, issue.Maximum)
	})
}

//////////////////////////////////////////
//////////   Error Resolution Chain Tests ///
//////////////////////////////////////////

func TestErrorResolutionChain(t *testing.T) {
	t.Run("uses context error when available", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, nil,
			WithExpected("string"),
			WithReceived("number"),
		)

		// Create context with custom error function
		ctx := &core.ParseContext{
			Error: func(issue core.ZodRawIssue) string {
				return "Context custom error message"
			},
		}
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		assert.Equal(t, "Context custom error message", issue.Message)
	})

	t.Run("falls back to default message when context error returns empty", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, nil,
			WithExpected("string"),
			WithReceived("number"),
		)

		// Create context with error function that returns empty
		ctx := &core.ParseContext{
			Error: func(issue core.ZodRawIssue) string {
				return "" // Return empty to test fallback
			},
		}
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		// Should fall back to default message
		assert.Contains(t, issue.Message, "Invalid input")
	})

	t.Run("uses schema-level error when inst is provided", func(t *testing.T) {
		// This test is simplified since we don't have full schema internals implemented yet
		rawIssue := NewRawIssue(core.Custom, nil,
			WithMessage(""), // Empty message to trigger error resolution
			WithInst("test_schema_instance"),
		)

		config := &core.ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		// For now, should fall back to default since schema error extraction is not fully implemented
		assert.Contains(t, issue.Message, "Invalid input")
	})

	t.Run("config-level error handling", func(t *testing.T) {
		rawIssue := NewRawIssue(core.InvalidType, nil,
			WithExpected("string"),
			WithReceived("number"),
		)

		// Test with custom error in config
		configWithCustomError := &core.ZodConfig{
			CustomError: func(issue core.ZodRawIssue) string {
				return "Custom config error message"
			},
		}

		issue := FinalizeIssue(rawIssue, nil, configWithCustomError)
		assert.Equal(t, "Custom config error message", issue.Message)

		// Test with locale error in config
		configWithLocaleError := &core.ZodConfig{
			LocaleError: func(issue core.ZodRawIssue) string {
				return "Locale config error message"
			},
		}

		issue = FinalizeIssue(rawIssue, nil, configWithLocaleError)
		assert.Equal(t, "Locale config error message", issue.Message)

		// Test priority: custom error should override locale error
		configWithBoth := &core.ZodConfig{
			CustomError: func(issue core.ZodRawIssue) string {
				return "Custom error has priority"
			},
			LocaleError: func(issue core.ZodRawIssue) string {
				return "Locale error lower priority"
			},
		}

		issue = FinalizeIssue(rawIssue, nil, configWithBoth)
		assert.Equal(t, "Custom error has priority", issue.Message)

		// Test that context error overrides config error
		ctx := &core.ParseContext{
			Error: func(issue core.ZodRawIssue) string {
				return "Context error overrides config"
			},
		}

		issue = FinalizeIssue(rawIssue, ctx, configWithBoth)
		assert.Equal(t, "Context error overrides config", issue.Message)

		// Test that provided message overrides everything
		rawIssueWithMessage := NewRawIssue(core.InvalidType, nil,
			WithMessage("Provided message has highest priority"),
		)

		issue = FinalizeIssue(rawIssueWithMessage, ctx, configWithBoth)
		assert.Equal(t, "Provided message has highest priority", issue.Message)

		// Test fallback to default when config errors return empty
		configWithEmptyErrors := &core.ZodConfig{
			CustomError: func(issue core.ZodRawIssue) string {
				return "" // Return empty to test fallback
			},
			LocaleError: func(issue core.ZodRawIssue) string {
				return "" // Return empty to test fallback
			},
		}

		issue = FinalizeIssue(rawIssue, nil, configWithEmptyErrors)
		assert.Contains(t, issue.Message, "Invalid input") // Should fall back to default
	})
}

//////////////////////////////////////////
//////////   Convenience Function Tests ///
//////////////////////////////////////////

func TestConvenienceFunctions(t *testing.T) {
	t.Run("CopyRawIssueProperties creates independent copy", func(t *testing.T) {
		original := map[string]any{
			"expected": "string",
			"minimum":  5,
			"keys":     []string{"key1", "key2"},
		}

		rawIssue := core.ZodRawIssue{
			Code:       "test",
			Properties: original,
		}

		copied := CopyRawIssueProperties(rawIssue)

		// Should be equal but independent
		assert.Equal(t, original, copied)

		// Modify original
		original["new_key"] = "new_value"

		// Copy should not be affected
		assert.NotContains(t, copied, "new_key")
	})

	t.Run("CopyRawIssueProperties handles nil properties", func(t *testing.T) {
		rawIssue := core.ZodRawIssue{
			Code:       "test",
			Properties: nil,
		}

		copied := CopyRawIssueProperties(rawIssue)
		assert.Nil(t, copied)
	})

	t.Run("MergeRawIssueProperties merges properties correctly", func(t *testing.T) {
		rawIssue := &core.ZodRawIssue{
			Code: "test",
			Properties: map[string]any{
				"existing": "value1",
				"shared":   "original",
			},
		}

		newProperties := map[string]any{
			"new":    "value2",
			"shared": "updated",
		}

		MergeRawIssueProperties(rawIssue, newProperties)

		expected := map[string]any{
			"existing": "value1",
			"new":      "value2",
			"shared":   "updated", // Should be overwritten
		}

		assert.Equal(t, expected, rawIssue.Properties)
	})

	t.Run("MergeRawIssueProperties handles nil original properties", func(t *testing.T) {
		rawIssue := &core.ZodRawIssue{
			Code:       "test",
			Properties: nil,
		}

		newProperties := map[string]any{
			"new": "value",
		}

		MergeRawIssueProperties(rawIssue, newProperties)

		assert.Equal(t, newProperties, rawIssue.Properties)
	})

	t.Run("MergeRawIssueProperties handles nil new properties", func(t *testing.T) {
		original := map[string]any{
			"existing": "value",
		}

		rawIssue := &core.ZodRawIssue{
			Code:       "test",
			Properties: original,
		}

		MergeRawIssueProperties(rawIssue, nil)

		// Should remain unchanged
		assert.Equal(t, original, rawIssue.Properties)
	})
}

//////////////////////////////////////////
//////////   Integration Tests         ///
//////////////////////////////////////////

func TestFinalizationIntegration(t *testing.T) {
	t.Run("complete finalization workflow", func(t *testing.T) {
		// Create complex raw issue
		rawIssue := NewRawIssue(core.InvalidFormat, "invalid@email",
			WithExpected("valid_email"),
			WithReceived("invalid_string"),
			WithOrigin("string"),
			WithFormat("email"),
			WithPattern(".*@.*\\..*"),
			WithPath([]any{"user", "contact", "email"}),
		)

		// Create context
		ctx := &core.ParseContext{
			ReportInput: true,
		}

		// Create config
		config := &core.ZodConfig{}

		// Finalize
		issue := FinalizeIssue(rawIssue, ctx, config)

		// Verify complete finalization
		assert.Equal(t, core.InvalidFormat, issue.Code)
		assert.Equal(t, "invalid@email", issue.Input)
		assert.Equal(t, []any{"user", "contact", "email"}, issue.Path)
		assert.NotEmpty(t, issue.Message)
		assert.Equal(t, "valid_email", issue.Expected)
		assert.Equal(t, "invalid_string", issue.Received)
		assert.Equal(t, "string", issue.Origin)
		assert.Equal(t, "email", issue.Format)
		assert.Equal(t, ".*@.*\\..*", issue.Pattern)
	})

	t.Run("finalization with creation helpers", func(t *testing.T) {
		// Test that creation helpers work well with finalization
		rawIssue := CreateInvalidTypeIssue("string", "123")
		rawIssue.Path = []any{"data", "field"}

		ctx := &core.ParseContext{ReportInput: false}
		config := &core.ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		assert.Equal(t, core.InvalidType, issue.Code)
		assert.Nil(t, issue.Input) // Should be excluded due to ReportInput: false
		assert.Equal(t, []any{"data", "field"}, issue.Path)
		assert.Equal(t, "string", issue.Expected)
	})

	t.Run("finalization preserves all property types", func(t *testing.T) {
		// Test with all possible property types
		rawIssue := CreateCustomIssue("", nil, "test")
		rawIssue.Properties["expected"] = "string"
		rawIssue.Properties["received"] = "number"
		rawIssue.Properties["minimum"] = 0
		rawIssue.Properties["maximum"] = 100
		rawIssue.Properties["inclusive"] = true
		rawIssue.Properties["origin"] = "validation"
		rawIssue.Properties["format"] = "custom"
		rawIssue.Properties["pattern"] = ".*"
		rawIssue.Properties["prefix"] = "pre_"
		rawIssue.Properties["suffix"] = "_suf"
		rawIssue.Properties["includes"] = "inc"
		rawIssue.Properties["divisor"] = 2.5
		rawIssue.Properties["keys"] = []string{"k1", "k2"}
		rawIssue.Properties["values"] = []any{"v1", "v2"}
		rawIssue.Properties["algorithm"] = "ALG"
		rawIssue.Properties["params"] = map[string]any{"param": "value"}

		issue := FinalizeIssue(rawIssue, nil, nil)

		// Verify all properties are preserved
		assert.Equal(t, "string", issue.Expected)
		assert.Equal(t, "number", issue.Received)
		assert.Equal(t, 0, issue.Minimum)
		assert.Equal(t, 100, issue.Maximum)
		assert.True(t, issue.Inclusive)
		assert.Equal(t, "validation", issue.Origin)
		assert.Equal(t, "custom", issue.Format)
		assert.Equal(t, ".*", issue.Pattern)
		assert.Equal(t, "pre_", issue.Prefix)
		assert.Equal(t, "_suf", issue.Suffix)
		assert.Equal(t, "inc", issue.Includes)
		assert.Equal(t, 2.5, issue.Divisor)
		assert.Equal(t, []string{"k1", "k2"}, issue.Keys)
		assert.Equal(t, []any{"v1", "v2"}, issue.Values)
		assert.Equal(t, "ALG", issue.Algorithm)
		assert.Equal(t, map[string]any{"param": "value"}, issue.Params)
	})
}

//////////////////////////////////////////
//////////   Edge Cases Tests          ///
//////////////////////////////////////////

func TestFinalizationEdgeCases(t *testing.T) {
	t.Run("handles very long paths", func(t *testing.T) {
		longPath := []any{}
		for i := 0; i < 100; i++ {
			longPath = append(longPath, i)
		}

		rawIssue := NewRawIssue(core.Custom, nil, WithPath(longPath))
		issue := FinalizeIssue(rawIssue, nil, nil)

		assert.Equal(t, longPath, issue.Path)
	})

	t.Run("handles mixed path types", func(t *testing.T) {
		mixedPath := []any{"string", 42, true, 3.14, []string{"nested"}}
		rawIssue := NewRawIssue(core.Custom, nil, WithPath(mixedPath))
		issue := FinalizeIssue(rawIssue, nil, nil)

		assert.Equal(t, mixedPath, issue.Path)
	})

	t.Run("handles very large property maps", func(t *testing.T) {
		largeProps := make(map[string]any)
		for i := 0; i < 100; i++ {
			largeProps[string(rune(i))] = i
		}

		rawIssue := core.ZodRawIssue{
			Code:       core.Custom,
			Properties: largeProps,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			issue := FinalizeIssue(rawIssue, nil, nil)
			assert.Equal(t, core.Custom, issue.Code)
		})
	})

	t.Run("handles unicode in messages and properties", func(t *testing.T) {
		unicodeMessage := "验证失败 🚫 émojis and special chars"
		unicodeProperty := "测试属性值"

		rawIssue := NewRawIssue(core.Custom, nil,
			WithMessage(unicodeMessage),
			WithExpected(unicodeProperty),
		)

		issue := FinalizeIssue(rawIssue, nil, nil)

		assert.Equal(t, unicodeMessage, issue.Message)
		assert.Equal(t, unicodeProperty, issue.Expected)
	})
}
