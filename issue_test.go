package gozod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Raw Issue Creation Tests  ///
//////////////////////////////////////////

func TestNewRawIssue(t *testing.T) {
	t.Run("creates raw issue with basic properties", func(t *testing.T) {
		issue := NewRawIssue("invalid_type", "test_input",
			WithExpected("string"),
			WithReceived("number"),
			WithMessage("Custom message"),
		)

		require.Equal(t, "invalid_type", issue.Code)
		require.Equal(t, "test_input", issue.Input)
		require.Equal(t, "Custom message", issue.Message)
		require.NotNil(t, issue.Properties)
		assert.Equal(t, "string", issue.Properties["expected"])
		assert.Equal(t, "number", issue.Properties["received"])
		assert.Empty(t, issue.Path) // Should be initialized as empty slice
	})

	t.Run("creates raw issue with path information", func(t *testing.T) {
		path := []interface{}{"user", "name"}
		issue := NewRawIssue("too_small", "hi",
			WithMinimum(5),
			WithOrigin("string"),
			WithPath(path),
		)

		require.Equal(t, "too_small", issue.Code)
		assert.Equal(t, path, issue.Path)
		assert.Equal(t, 5, issue.Properties["minimum"])
		assert.Equal(t, "string", issue.Properties["origin"])
	})

	t.Run("creates raw issue with all option types", func(t *testing.T) {
		keys := []string{"invalid_key1", "invalid_key2"}
		values := []interface{}{"val1", "val2"}
		params := map[string]interface{}{"custom": "data"}

		issue := NewRawIssue("custom", nil,
			WithKeys(keys),
			WithValues(values),
			WithFormat("email"),
			WithPattern(".*@.*"),
			WithPrefix("test_"),
			WithSuffix("_end"),
			WithIncludes("@"),
			WithDivisor(2),
			WithAlgorithm("HS256"),
			WithParams(params),
			WithInclusive(true),
			WithContinue(false),
		)

		require.Equal(t, "custom", issue.Code)
		assert.Equal(t, keys, issue.Properties["keys"])
		assert.Equal(t, values, issue.Properties["values"])
		assert.Equal(t, "email", issue.Properties["format"])
		assert.Equal(t, ".*@.*", issue.Properties["pattern"])
		assert.Equal(t, "test_", issue.Properties["prefix"])
		assert.Equal(t, "_end", issue.Properties["suffix"])
		assert.Equal(t, "@", issue.Properties["includes"])
		assert.Equal(t, 2, issue.Properties["divisor"])
		assert.Equal(t, "HS256", issue.Properties["algorithm"])
		assert.Equal(t, params, issue.Properties["params"])
		assert.Equal(t, true, issue.Properties["inclusive"])
		assert.Equal(t, false, issue.Continue)
	})

	t.Run("handles nil properties map gracefully", func(t *testing.T) {
		issue := NewRawIssue("custom", "test")
		require.NotNil(t, issue.Properties)
		assert.Empty(t, issue.Properties)
	})
}

func TestNewRawIssueFromMessage(t *testing.T) {
	t.Run("creates raw issue from message with custom code", func(t *testing.T) {
		issue := NewRawIssueFromMessage("Custom validation failed", "test_input", nil)

		require.Equal(t, string(Custom), issue.Code)
		require.Equal(t, "Custom validation failed", issue.Message)
		require.Equal(t, "test_input", issue.Input)
		assert.Empty(t, issue.Path)
		assert.NotNil(t, issue.Properties)
	})

	t.Run("handles nil instance parameter", func(t *testing.T) {
		issue := NewRawIssueFromMessage("Error message", 123, nil)
		assert.Nil(t, issue.Inst)
		assert.Equal(t, 123, issue.Input)
	})
}

//////////////////////////////////////////
//////////   Raw Issue Accessor Tests  ///
//////////////////////////////////////////

func TestZodRawIssueAccessors(t *testing.T) {
	t.Run("accesses string properties correctly", func(t *testing.T) {
		issue := NewRawIssue("invalid_format", "invalid@",
			WithExpected("string"),
			WithReceived("number"),
			WithOrigin("string"),
			WithFormat("email"),
			WithPattern(".*@.*\\..*"),
			WithPrefix("valid_"),
			WithSuffix("_format"),
			WithIncludes("@"),
		)

		assert.Equal(t, "string", issue.GetExpected())
		assert.Equal(t, "number", issue.GetReceived())
		assert.Equal(t, "string", issue.GetOrigin())
		assert.Equal(t, "email", issue.GetFormat())
		assert.Equal(t, ".*@.*\\..*", issue.GetPattern())
		assert.Equal(t, "valid_", issue.GetPrefix())
		assert.Equal(t, "_format", issue.GetSuffix())
		assert.Equal(t, "@", issue.GetIncludes())
	})

	t.Run("accesses numeric and boolean properties correctly", func(t *testing.T) {
		issue := NewRawIssue("too_small", 3,
			WithMinimum(5),
			WithMaximum(100),
			WithInclusive(true),
			WithDivisor(2.5),
		)

		assert.Equal(t, 5, issue.GetMinimum())
		assert.Equal(t, 100, issue.GetMaximum())
		assert.True(t, issue.GetInclusive())
		assert.Equal(t, 2.5, issue.GetDivisor())
	})

	t.Run("accesses slice properties correctly", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		values := []interface{}{"val1", 42, true}

		issue := NewRawIssue("unrecognized_keys", nil,
			WithKeys(keys),
			WithValues(values),
		)

		assert.Equal(t, keys, issue.GetKeys())
		assert.Equal(t, values, issue.GetValues())
	})

	t.Run("returns empty values for missing properties", func(t *testing.T) {
		issue := NewRawIssue("custom", nil)

		assert.Empty(t, issue.GetExpected())
		assert.Empty(t, issue.GetReceived())
		assert.Empty(t, issue.GetOrigin())
		assert.Empty(t, issue.GetFormat())
		assert.Nil(t, issue.GetMinimum())
		assert.Nil(t, issue.GetMaximum())
		assert.False(t, issue.GetInclusive())
		assert.Nil(t, issue.GetDivisor())
		assert.Nil(t, issue.GetKeys())
		assert.Nil(t, issue.GetValues())
	})

	t.Run("handles nil properties map gracefully", func(t *testing.T) {
		issue := ZodRawIssue{
			Code:       "test",
			Properties: nil,
		}

		assert.Empty(t, issue.GetExpected())
		assert.Nil(t, issue.GetMinimum())
		assert.False(t, issue.GetInclusive())
	})
}

//////////////////////////////////////////
//////////   Issue Finalization Tests  ///
//////////////////////////////////////////

func TestFinalizeIssue(t *testing.T) {
	t.Run("finalizes issue with basic properties", func(t *testing.T) {
		rawIssue := NewRawIssue("invalid_type", "test_input",
			WithExpected("string"),
			WithReceived("number"),
			WithMessage("Custom error message"),
		)

		config := &ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		require.Equal(t, "invalid_type", issue.Code)
		require.Equal(t, "Custom error message", issue.Message)
		assert.Empty(t, issue.Path)
		assert.Equal(t, "string", issue.Expected)
		assert.Equal(t, "number", issue.Received)
		assert.Equal(t, "test_input", issue.Input)
	})

	t.Run("includes input when ReportInput is true", func(t *testing.T) {
		rawIssue := NewRawIssue("too_small", "hi")
		ctx := &ParseContext{ReportInput: true}
		config := &ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		assert.Equal(t, "hi", issue.Input)
	})

	t.Run("excludes input when ReportInput is false", func(t *testing.T) {
		rawIssue := NewRawIssue("too_small", "secret_data")
		ctx := &ParseContext{ReportInput: false}
		config := &ZodConfig{}

		issue := FinalizeIssue(rawIssue, ctx, config)

		assert.Nil(t, issue.Input)
	})

	t.Run("uses default message when no message provided", func(t *testing.T) {
		rawIssue := NewRawIssue("invalid_type", nil,
			WithExpected("string"),
			WithReceived("number"),
		)

		config := &ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Contains(t, issue.Message, "Invalid input")
		assert.Contains(t, issue.Message, "string")
		assert.Contains(t, issue.Message, "number")
	})

	t.Run("generates appropriate default messages for different issue types", func(t *testing.T) {
		testCases := []struct {
			name     string
			rawIssue ZodRawIssue
			expected string
		}{
			{
				name: "too_small message",
				rawIssue: NewRawIssue("too_small", 3,
					WithMinimum(5),
					WithOrigin("string"),
				),
				expected: "string must be at least 5",
			},
			{
				name: "too_big message",
				rawIssue: NewRawIssue("too_big", 150,
					WithMaximum(100),
					WithOrigin("number"),
				),
				expected: "number must be at most 100",
			},
			{
				name: "invalid_format message",
				rawIssue: NewRawIssue("invalid_format", "invalid@",
					WithFormat("email"),
				),
				expected: "Invalid email",
			},
			{
				name: "not_multiple_of message",
				rawIssue: NewRawIssue("not_multiple_of", 7,
					WithDivisor(2),
				),
				expected: "Number must be a multiple of 2",
			},
			{
				name:     "unrecognized_keys message",
				rawIssue: NewRawIssue("unrecognized_keys", nil),
				expected: "Unrecognized key(s) in object",
			},
			{
				name:     "invalid_union message",
				rawIssue: NewRawIssue("invalid_union", nil),
				expected: "Invalid input",
			},
			{
				name:     "custom message",
				rawIssue: NewRawIssue("custom", nil),
				expected: "Refinement failed",
			},
		}

		config := &ZodConfig{}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				issue := FinalizeIssue(tc.rawIssue, nil, config)
				assert.Contains(t, issue.Message, tc.expected)
			})
		}
	})

	t.Run("ensures path is never nil", func(t *testing.T) {
		rawIssue := ZodRawIssue{
			Code:  "custom",
			Input: nil,
			Path:  nil, // Explicitly nil
		}

		config := &ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		require.NotNil(t, issue.Path)
		assert.Empty(t, issue.Path)
	})

	t.Run("preserves existing path", func(t *testing.T) {
		path := []interface{}{"user", "profile", "email"}
		rawIssue := NewRawIssue("invalid_format", "invalid@",
			WithPath(path),
		)

		config := &ZodConfig{}
		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Equal(t, path, issue.Path)
	})

	t.Run("TODO: uses context error mapping when provided", func(t *testing.T) {
		t.Skip("TODO: Test context-level error mapping when error resolution chain is fully implemented")

		// This would test:
		// - ctx.Error function override
		// - Priority order: inst.error > ctx.error > config.customError > config.localeError > default
	})

	t.Run("TODO: uses config error mapping when provided", func(t *testing.T) {
		t.Skip("TODO: Test config-level error mapping when ZodConfig methods are implemented")

		// This would test:
		// - config.customError override
		// - config.localeError override
		// - Error resolution priority
	})
}

//////////////////////////////////////////
//////////   Issue Type-Specific Accessor Tests  ///
//////////////////////////////////////////

func TestZodIssueTypeSpecificAccessors(t *testing.T) {
	t.Run("invalid_type issue accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "invalid_type"},
			Expected:     "string",
			Received:     "number",
		}

		expected, ok := issue.GetExpected()
		require.True(t, ok)
		assert.Equal(t, "string", expected)

		received, ok := issue.GetReceived()
		require.True(t, ok)
		assert.Equal(t, "number", received)
	})

	t.Run("too_small issue accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "too_small"},
			Minimum:      10,
		}

		minimum, ok := issue.GetMinimum()
		require.True(t, ok)
		assert.Equal(t, 10, minimum)

		// Wrong type should return false
		_, ok = issue.GetExpected()
		assert.False(t, ok)
	})

	t.Run("too_big issue accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "too_big"},
			Maximum:      100,
		}

		maximum, ok := issue.GetMaximum()
		require.True(t, ok)
		assert.Equal(t, 100, maximum)

		// Wrong type should return false
		_, ok = issue.GetMinimum()
		assert.False(t, ok)
	})

	t.Run("invalid_format issue accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "invalid_format"},
			Format:       "email",
		}

		format, ok := issue.GetFormat()
		require.True(t, ok)
		assert.Equal(t, "email", format)

		// Wrong type should return false
		_, ok = issue.GetMinimum()
		assert.False(t, ok)
	})

	t.Run("not_multiple_of issue accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "not_multiple_of"},
			Divisor:      2.5,
		}

		divisor, ok := issue.GetDivisor()
		require.True(t, ok)
		assert.Equal(t, 2.5, divisor)

		// Wrong type should return false
		_, ok = issue.GetExpected()
		assert.False(t, ok)
	})

	t.Run("returns false for wrong issue types", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "custom"},
			Expected:     "string", // Set but not accessible for custom type
		}

		_, ok := issue.GetExpected()
		assert.False(t, ok, "Should return false for wrong issue type")

		_, ok = issue.GetMinimum()
		assert.False(t, ok, "Should return false for wrong issue type")

		_, ok = issue.GetMaximum()
		assert.False(t, ok, "Should return false for wrong issue type")

		_, ok = issue.GetFormat()
		assert.False(t, ok, "Should return false for wrong issue type")

		_, ok = issue.GetDivisor()
		assert.False(t, ok, "Should return false for wrong issue type")
	})

	t.Run("handles empty values correctly", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "invalid_type"},
			Expected:     "", // Empty but accessible
			Received:     "", // Empty but accessible
		}

		// For empty strings, the accessors should return false because empty is treated as "not set"
		expected, ok := issue.GetExpected()
		require.False(t, ok, "GetExpected should return false for empty strings")
		assert.Empty(t, expected)

		received, ok := issue.GetReceived()
		require.False(t, ok, "GetReceived should return false for empty strings")
		assert.Empty(t, received)

		// Test with actual non-empty values
		issueWithValues := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "invalid_type"},
			Expected:     "string",
			Received:     "number",
		}

		expected, ok = issueWithValues.GetExpected()
		require.True(t, ok, "GetExpected should return true for non-empty strings")
		assert.Equal(t, "string", expected)

		received, ok = issueWithValues.GetReceived()
		require.True(t, ok, "GetReceived should return true for non-empty strings")
		assert.Equal(t, "number", received)
	})
}

//////////////////////////////////////////
//////////   Error Creation Helper Tests  ///
//////////////////////////////////////////

func TestErrorCreationHelpers(t *testing.T) {
	t.Run("CreateInvalidTypeIssue", func(t *testing.T) {
		issue := CreateInvalidTypeIssue("test_input", "string", "number")

		require.Equal(t, string(InvalidType), issue.Code)
		require.Equal(t, "test_input", issue.Input)
		assert.Equal(t, "string", issue.GetExpected())
		assert.Equal(t, "number", issue.GetReceived())
	})

	t.Run("CreateTooBigIssue", func(t *testing.T) {
		issue := CreateTooBigIssue(150, "number", 100, true)

		require.Equal(t, string(TooBig), issue.Code)
		require.Equal(t, 150, issue.Input)
		assert.Equal(t, "number", issue.GetOrigin())
		assert.Equal(t, 100, issue.GetMaximum())
		assert.True(t, issue.GetInclusive())
	})

	t.Run("CreateTooSmallIssue", func(t *testing.T) {
		issue := CreateTooSmallIssue(3, "string", 5, false)

		require.Equal(t, string(TooSmall), issue.Code)
		require.Equal(t, 3, issue.Input)
		assert.Equal(t, "string", issue.GetOrigin())
		assert.Equal(t, 5, issue.GetMinimum())
		assert.False(t, issue.GetInclusive())
	})

	t.Run("CreateInvalidFormatIssue", func(t *testing.T) {
		issue := CreateInvalidFormatIssue("invalid@", "email")

		require.Equal(t, string(InvalidFormat), issue.Code)
		require.Equal(t, "invalid@", issue.Input)
		assert.Equal(t, "string", issue.GetOrigin())
		assert.Equal(t, "email", issue.GetFormat())
	})

	t.Run("CreateNotMultipleOfIssue", func(t *testing.T) {
		issue := CreateNotMultipleOfIssue(7, 2)

		require.Equal(t, string(NotMultipleOf), issue.Code)
		require.Equal(t, 7, issue.Input)
		assert.Equal(t, 2, issue.GetDivisor())
	})

	t.Run("CreateCustomIssue", func(t *testing.T) {
		issue := CreateCustomIssue("test_input", "Custom validation failed")

		require.Equal(t, string(Custom), issue.Code)
		require.Equal(t, "test_input", issue.Input)
		require.Equal(t, "Custom validation failed", issue.Message)
	})

	t.Run("helper functions accept additional options", func(t *testing.T) {
		path := []interface{}{"user", "email"}
		issue := CreateInvalidFormatIssue("invalid@", "email",
			WithPath(path),
			WithMessage("Custom email error"),
		)

		assert.Equal(t, path, issue.Path)
		assert.Equal(t, "Custom email error", issue.Message)
	})
}

//////////////////////////////////////////
//////////   Issue Constants Tests     ///
//////////////////////////////////////////

func TestIssueConstants(t *testing.T) {
	t.Run("all issue codes are defined correctly", func(t *testing.T) {
		expectedCodes := map[IssueCode]string{
			InvalidType:      "invalid_type",
			TooBig:           "too_big",
			TooSmall:         "too_small",
			InvalidFormat:    "invalid_format",
			NotMultipleOf:    "not_multiple_of",
			UnrecognizedKeys: "unrecognized_keys",
			InvalidUnion:     "invalid_union",
			InvalidKey:       "invalid_key",
			InvalidElement:   "invalid_element",
			InvalidValue:     "invalid_value",
			Custom:           "custom",
		}

		for code, expected := range expectedCodes {
			assert.Equal(t, expected, string(code))
		}
	})

	t.Run("issue codes can be used for comparison", func(t *testing.T) {
		issue := NewRawIssue(string(InvalidType), nil)
		assert.Equal(t, string(InvalidType), issue.Code)

		finalizedIssue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: string(TooBig)},
		}
		assert.Equal(t, string(TooBig), finalizedIssue.Code)
	})
}

//////////////////////////////////////////
//////////   ZodIssue Interface Tests  ///
//////////////////////////////////////////

func TestZodIssueInterface(t *testing.T) {
	t.Run("ZodIssue implements error interface", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "custom",
				Message: "Test error message",
			},
		}

		var err error = issue
		assert.Equal(t, "Test error message", err.Error())
	})

	t.Run("ZodIssue string representation", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "invalid_type",
				Message: "Expected string, received number",
				Path:    []interface{}{"user", "name"},
			},
		}

		str := issue.String()
		assert.Contains(t, str, "ZodIssue")
		assert.Contains(t, str, "invalid_type")
		assert.Contains(t, str, "Expected string, received number")
		assert.Contains(t, str, "[user name]")
	})

	t.Run("ZodIssue handles empty path", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "custom",
				Message: "Error",
				Path:    []interface{}{},
			},
		}

		str := issue.String()
		assert.Contains(t, str, "Path: []")
	})

	t.Run("ZodIssue handles nil path", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{
				Code:    "custom",
				Message: "Error",
				Path:    nil,
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			_ = issue.String()
		})
	})
}

//////////////////////////////////////////
//////////   Edge Cases and Error Handling Tests  ///
//////////////////////////////////////////

func TestEdgeCasesAndErrorHandling(t *testing.T) {
	t.Run("handles missing properties gracefully", func(t *testing.T) {
		issue := ZodRawIssue{
			Code:       "test",
			Properties: nil,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			_ = issue.GetExpected()
			_ = issue.GetMinimum()
			_ = issue.GetInclusive()
		})
	})

	t.Run("handles wrong property types gracefully", func(t *testing.T) {
		issue := ZodRawIssue{
			Code: "test",
			Properties: map[string]interface{}{
				"expected":  123,     // Wrong type (should be string)
				"minimum":   "hello", // Wrong type (should be number)
				"inclusive": "yes",   // Wrong type (should be bool)
			},
		}

		// Should return empty/default values for wrong types
		assert.Empty(t, issue.GetExpected())
		assert.Equal(t, "hello", issue.GetMinimum()) // Returns interface{}, so preserves wrong type
		assert.False(t, issue.GetInclusive())        // Wrong type returns false
	})

	t.Run("finalization with nil config", func(t *testing.T) {
		rawIssue := NewRawIssue("custom", nil)

		// Should not panic with nil config
		assert.NotPanics(t, func() {
			issue := FinalizeIssue(rawIssue, nil, nil)
			assert.Equal(t, "custom", issue.Code)
		})
	})

	t.Run("finalization with nil context", func(t *testing.T) {
		rawIssue := NewRawIssue("custom", "test_input")
		config := &ZodConfig{}

		issue := FinalizeIssue(rawIssue, nil, config)

		assert.Equal(t, "custom", issue.Code)
		// When ctx is nil, input is included by default (ReportInput defaults to true)
		assert.Equal(t, "test_input", issue.Input)
	})

	t.Run("property mapping handles all field types", func(t *testing.T) {
		props := map[string]interface{}{
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
			"divisor":   2.5,
			"keys":      []string{"key1", "key2"},
			"values":    []interface{}{"val1", "val2"},
			"origin":    "string",
			"key":       "field_name",
			"params":    map[string]interface{}{"custom": "data"},
		}

		issue := ZodIssue{}

		// Should not panic
		assert.NotPanics(t, func() {
			mapPropertiesToIssue(&issue, props)
		})

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
		assert.Equal(t, 2.5, issue.Divisor)
		assert.Equal(t, []string{"key1", "key2"}, issue.Keys)
		assert.Equal(t, []interface{}{"val1", "val2"}, issue.Values)
		assert.Equal(t, "string", issue.Origin)
		assert.Equal(t, "field_name", issue.Key)
		assert.Equal(t, map[string]interface{}{"custom": "data"}, issue.Params)
	})

	t.Run("property mapping with nil properties", func(t *testing.T) {
		issue := ZodIssue{}

		// Should not panic
		assert.NotPanics(t, func() {
			mapPropertiesToIssue(&issue, nil)
		})

		// Should remain unchanged
		assert.Empty(t, issue.Expected)
		assert.Nil(t, issue.Minimum)
	})

	t.Run("property mapping with wrong types", func(t *testing.T) {
		props := map[string]interface{}{
			"expected":  123,                      // Wrong type
			"keys":      "not_a_slice",            // Wrong type
			"inclusive": "not_a_bool",             // Wrong type
			"params":    []string{"not", "a_map"}, // Wrong type
		}

		issue := ZodIssue{}

		// Should not panic with wrong types
		assert.NotPanics(t, func() {
			mapPropertiesToIssue(&issue, props)
		})

		// Wrong types should not be assigned
		assert.Empty(t, issue.Expected)  // Wrong type, should remain empty
		assert.Nil(t, issue.Keys)        // Wrong type, should remain nil
		assert.False(t, issue.Inclusive) // Wrong type, should remain false
		assert.Nil(t, issue.Params)      // Wrong type, should remain nil
	})
}

//////////////////////////////////////////
//////////   Performance Tests         ///
//////////////////////////////////////////

func TestPerformance(t *testing.T) {
	t.Run("raw issue creation is efficient", func(t *testing.T) {
		// This test ensures issue creation doesn't have performance regressions
		for i := 0; i < 1000; i++ {
			issue := NewRawIssue("invalid_type", i,
				WithExpected("string"),
				WithReceived("number"),
			)
			assert.Equal(t, "invalid_type", issue.Code)
		}
	})

	t.Run("issue finalization is efficient", func(t *testing.T) {
		config := &ZodConfig{}

		// Test finalization performance
		for i := 0; i < 1000; i++ {
			rawIssue := NewRawIssue("too_small", i,
				WithMinimum(100),
				WithOrigin("number"),
			)

			issue := FinalizeIssue(rawIssue, nil, config)
			assert.Equal(t, "too_small", issue.Code)
		}
	})

	t.Run("property mapping is efficient", func(t *testing.T) {
		props := map[string]interface{}{
			"expected": "string",
			"received": "number",
			"minimum":  5,
			"maximum":  100,
		}

		// Test property mapping performance
		for i := 0; i < 1000; i++ {
			issue := ZodIssue{}
			mapPropertiesToIssue(&issue, props)
			assert.Equal(t, "string", issue.Expected)
		}
	})
}

//////////////////////////////////////////
//////////   Integration Tests          ///
//////////////////////////////////////////

func TestIntegration(t *testing.T) {
	t.Run("full issue lifecycle from creation to finalization", func(t *testing.T) {
		// Create raw issue
		rawIssue := NewRawIssue("invalid_type", "123",
			WithExpected("string"),
			WithReceived("number"),
			WithPath([]interface{}{"user", "name"}),
		)

		// Verify raw issue properties
		require.Equal(t, "invalid_type", rawIssue.Code)
		require.Equal(t, "string", rawIssue.GetExpected())
		require.Equal(t, "number", rawIssue.GetReceived())

		// Finalize issue
		ctx := &ParseContext{ReportInput: true}
		config := &ZodConfig{}
		finalIssue := FinalizeIssue(rawIssue, ctx, config)

		// Verify finalized issue
		require.Equal(t, "invalid_type", finalIssue.Code)
		require.Equal(t, "123", finalIssue.Input)
		require.Equal(t, []interface{}{"user", "name"}, finalIssue.Path)
		assert.Equal(t, "string", finalIssue.Expected)
		assert.Equal(t, "number", finalIssue.Received)

		// Verify type-specific accessors work
		expected, ok := finalIssue.GetExpected()
		require.True(t, ok)
		assert.Equal(t, "string", expected)
	})

	t.Run("issue creation with all helper functions", func(t *testing.T) {
		testCases := []struct {
			name   string
			create func() ZodRawIssue
		}{
			{
				name:   "invalid type helper",
				create: func() ZodRawIssue { return CreateInvalidTypeIssue("test", "string", "number") },
			},
			{
				name:   "too big helper",
				create: func() ZodRawIssue { return CreateTooBigIssue(150, "number", 100, true) },
			},
			{
				name:   "too small helper",
				create: func() ZodRawIssue { return CreateTooSmallIssue(3, "string", 5, false) },
			},
			{
				name:   "invalid format helper",
				create: func() ZodRawIssue { return CreateInvalidFormatIssue("invalid@", "email") },
			},
			{
				name:   "not multiple of helper",
				create: func() ZodRawIssue { return CreateNotMultipleOfIssue(7, 2) },
			},
			{
				name:   "custom helper",
				create: func() ZodRawIssue { return CreateCustomIssue("test", "Custom error") },
			},
		}

		config := &ZodConfig{}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				rawIssue := tc.create()

				// Should be able to finalize without errors
				finalIssue := FinalizeIssue(rawIssue, nil, config)
				assert.NotEmpty(t, finalIssue.Code)
				assert.NotEmpty(t, finalIssue.Message)
			})
		}
	})
}
