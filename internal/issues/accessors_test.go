package issues

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Raw Issue Accessor Tests  ///
//////////////////////////////////////////

func TestZodRawIssueAccessors(t *testing.T) {
	t.Run("GetRawIssueExpected", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidType, "test", WithExpected("string"))
		expected := GetRawIssueExpected(issue)
		assert.Equal(t, "string", expected)
	})

	t.Run("GetRawIssueReceived", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidType, "test", WithReceived("number"))
		received := GetRawIssueReceived(issue)
		assert.Equal(t, "number", received)
	})

	t.Run("GetRawIssueOrigin", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithOrigin("string"))
		origin := GetRawIssueOrigin(issue)
		assert.Equal(t, "string", origin)
	})

	t.Run("GetRawIssueFormat", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithFormat("email"))
		format := GetRawIssueFormat(issue)
		assert.Equal(t, "email", format)
	})

	t.Run("GetRawIssuePattern", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithPattern("^.*$"))
		pattern := GetRawIssuePattern(issue)
		assert.Equal(t, "^.*$", pattern)
	})

	t.Run("GetRawIssuePrefix", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithPrefix("start_"))
		prefix := GetRawIssuePrefix(issue)
		assert.Equal(t, "start_", prefix)
	})

	t.Run("GetRawIssueSuffix", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithSuffix("_end"))
		suffix := GetRawIssueSuffix(issue)
		assert.Equal(t, "_end", suffix)
	})

	t.Run("GetRawIssueIncludes", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithIncludes("substring"))
		includes := GetRawIssueIncludes(issue)
		assert.Equal(t, "substring", includes)
	})

	t.Run("GetRawIssueMinimum", func(t *testing.T) {
		issue := NewRawIssue(core.TooSmall, "test", WithMinimum(5))
		minimum := GetRawIssueMinimum(issue)
		assert.Equal(t, 5, minimum)
	})

	t.Run("GetRawIssueMaximum", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithMaximum(100))
		maximum := GetRawIssueMaximum(issue)
		assert.Equal(t, 100, maximum)
	})

	t.Run("GetRawIssueInclusive", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithInclusive(true))
		inclusive := GetRawIssueInclusive(issue)
		assert.True(t, inclusive)
	})

	t.Run("GetRawIssueDivisor", func(t *testing.T) {
		issue := NewRawIssue(core.NotMultipleOf, "test", WithDivisor(3))
		divisor := GetRawIssueDivisor(issue)
		assert.Equal(t, 3, divisor)
	})

	t.Run("GetRawIssueKeys", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		issue := NewRawIssue(core.UnrecognizedKeys, "test", WithKeys(keys))
		result := GetRawIssueKeys(issue)
		assert.Equal(t, keys, result)
	})

	t.Run("GetRawIssueValues", func(t *testing.T) {
		values := []any{"val1", "val2", 123}
		issue := NewRawIssue(core.InvalidValue, "test", WithValues(values))
		result := GetRawIssueValues(issue)
		assert.Equal(t, values, result)
	})

	t.Run("empty properties return defaults", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: make(map[string]any),
		}

		assert.Empty(t, GetRawIssueExpected(issue))
		assert.Empty(t, GetRawIssueReceived(issue))
		assert.Empty(t, GetRawIssueOrigin(issue))
		assert.Nil(t, GetRawIssueMinimum(issue))
		assert.Nil(t, GetRawIssueMaximum(issue))
		assert.False(t, GetRawIssueInclusive(issue))
		assert.Nil(t, GetRawIssueKeys(issue))
		assert.Nil(t, GetRawIssueValues(issue))
	})

	t.Run("nil properties map", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: nil,
		}

		assert.Empty(t, GetRawIssueExpected(issue))
		assert.Empty(t, GetRawIssueReceived(issue))
		assert.Nil(t, GetRawIssueMinimum(issue))
		assert.False(t, GetRawIssueInclusive(issue))
	})

	t.Run("wrong type values return defaults", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code: core.InvalidType,
			Properties: map[string]any{
				"expected":  123,      // Should be string
				"minimum":   "string", // Should be numeric
				"inclusive": "yes",    // Should be bool
			},
		}

		assert.Empty(t, GetRawIssueExpected(issue))
		assert.Equal(t, "string", GetRawIssueMinimum(issue))
		assert.False(t, GetRawIssueInclusive(issue))
	})
}

//////////////////////////////////////////
//////////   Issue Type-Specific Accessor Tests  ///
//////////////////////////////////////////

func TestZodIssueTypeSpecificAccessors(t *testing.T) {
	t.Run("GetIssueExpected for InvalidType", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidType},
			Expected:     core.ZodTypeString,
		}

		expected, ok := GetIssueExpected(issue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeString, expected)
	})

	t.Run("GetIssueExpected for non-InvalidType", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooBig},
			Expected:     core.ZodTypeString,
		}

		expected, ok := GetIssueExpected(issue)
		assert.False(t, ok)
		assert.Empty(t, expected)
	})

	t.Run("GetIssueReceived for InvalidType", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidType},
			Received:     core.ZodTypeNumber,
		}

		received, ok := GetIssueReceived(issue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeNumber, received)
	})

	t.Run("GetIssueMinimum for TooSmall", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooSmall},
			Minimum:      5,
		}

		minimum, ok := GetIssueMinimum(issue)
		require.True(t, ok)
		assert.Equal(t, 5, minimum)
	})

	t.Run("GetIssueMinimum for non-TooSmall", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooBig},
			Minimum:      5,
		}

		minimum, ok := GetIssueMinimum(issue)
		assert.False(t, ok)
		assert.Nil(t, minimum)
	})

	t.Run("GetIssueMaximum for TooBig", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooBig},
			Maximum:      100,
		}

		maximum, ok := GetIssueMaximum(issue)
		require.True(t, ok)
		assert.Equal(t, 100, maximum)
	})

	t.Run("GetIssueFormat for InvalidFormat", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidFormat},
			Format:       "email",
		}

		format, ok := GetIssueFormat(issue)
		require.True(t, ok)
		assert.Equal(t, "email", format)
	})

	t.Run("GetIssueDivisor for NotMultipleOf", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.NotMultipleOf},
			Divisor:      3,
		}

		divisor, ok := GetIssueDivisor(issue)
		require.True(t, ok)
		assert.Equal(t, 3, divisor)
	})

	t.Run("empty field values return false", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidType},
			Expected:     "",
			Received:     "",
		}

		expected, ok := GetIssueExpected(issue)
		assert.False(t, ok)
		assert.Empty(t, expected)

		received, ok := GetIssueReceived(issue)
		assert.False(t, ok)
		assert.Empty(t, received)
	})

	t.Run("nil field values return false", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooSmall},
			Minimum:      nil,
		}

		minimum, ok := GetIssueMinimum(issue)
		assert.False(t, ok)
		assert.Nil(t, minimum)
	})
}

//////////////////////////////////////////
//////////   Accessor Edge Cases Tests ///
//////////////////////////////////////////

func TestAccessorEdgeCases(t *testing.T) {
	t.Run("handles zero values in accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "too_small"},
			Minimum:      0, // Zero value should still be accessible
		}

		minimum, ok := issue.GetMinimum()
		require.True(t, ok)
		assert.Equal(t, 0, minimum)
	})

	t.Run("handles false boolean values in accessors", func(t *testing.T) {
		rawIssue := NewRawIssue("test", nil, WithInclusive(false))
		assert.False(t, rawIssue.GetInclusive()) // False should be preserved
	})

	t.Run("handles nil any values", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "too_small"},
			Minimum:      nil, // Nil any
		}

		minimum, ok := issue.GetMinimum()
		assert.False(t, ok)
		assert.Nil(t, minimum)
	})

	t.Run("handles complex data types in properties", func(t *testing.T) {
		complexData := map[string]any{
			"nested": map[string]any{
				"level": 2,
				"items": []string{"a", "b", "c"},
			},
		}

		issue := NewRawIssue("custom", nil, WithParams(complexData))
		assert.Equal(t, complexData, issue.Properties["params"])
	})

	t.Run("handles very large numbers", func(t *testing.T) {
		largeNumber := int64(9223372036854775807) // max int64

		issue := NewRawIssue("too_big", largeNumber, WithMaximum(largeNumber))
		assert.Equal(t, largeNumber, issue.GetMaximum())
	})

	t.Run("handles unicode strings", func(t *testing.T) {
		unicodeString := "测试字符串 🚀 émojis"

		issue := NewRawIssue("invalid_format", unicodeString,
			WithExpected(unicodeString),
			WithOrigin(unicodeString),
		)

		assert.Equal(t, core.ZodTypeCode(unicodeString), issue.GetExpected())
		assert.Equal(t, unicodeString, issue.GetOrigin())
	})
}

//////////////////////////////////////////
//////////   Accessor Performance Tests ///
//////////////////////////////////////////

func TestAccessorPerformance(t *testing.T) {
	t.Run("accessor methods are efficient", func(t *testing.T) {
		// Create an issue with many properties
		issue := NewRawIssue("complex", nil,
			WithExpected("string"),
			WithReceived("number"),
			WithMinimum(0),
			WithMaximum(100),
			WithInclusive(true),
			WithOrigin("test"),
			WithFormat("format"),
			WithPattern(".*"),
			WithPrefix("pre_"),
			WithSuffix("_suf"),
			WithIncludes("inc"),
			WithDivisor(2),
			WithKeys([]string{"k1", "k2"}),
			WithValues([]any{"v1", "v2"}),
			WithAlgorithm("ALG"),
		)

		// Test that repeated access is efficient
		for i := 0; i < 1000; i++ {
			_ = issue.GetExpected()
			_ = issue.GetReceived()
			_ = issue.GetMinimum()
			_ = issue.GetMaximum()
			_ = issue.GetInclusive()
			_ = issue.GetOrigin()
			_ = issue.GetFormat()
			_ = issue.GetPattern()
			_ = issue.GetPrefix()
			_ = issue.GetSuffix()
			_ = issue.GetIncludes()
			_ = issue.GetDivisor()
			_ = issue.GetKeys()
			_ = issue.GetValues()
		}
	})

	t.Run("type-specific accessors are efficient", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "invalid_type"},
			Expected:     "string",
			Received:     "number",
		}

		// Test that type checking and access is efficient
		for i := 0; i < 1000; i++ {
			_, _ = issue.GetExpected()
			_, _ = issue.GetReceived()
			_, _ = issue.GetMinimum() // Should return false quickly
			_, _ = issue.GetMaximum() // Should return false quickly
		}
	})
}

//////////////////////////////////////////
//////////   Accessor Integration Tests ///
//////////////////////////////////////////

func TestAccessorIntegration(t *testing.T) {
	t.Run("raw issue to finalized issue accessor consistency", func(t *testing.T) {
		// Create raw issue with properties
		rawIssue := NewRawIssue(core.InvalidType, "test_input",
			WithExpected("string"),
			WithReceived("number"),
			WithMinimum(5),
			WithMaximum(100),
			WithInclusive(true),
			WithOrigin("validation"),
		)

		// Finalize it
		finalizedIssue := FinalizeIssue(rawIssue, nil, nil)

		// Check that properties are preserved correctly
		assert.Equal(t, core.ZodTypeCode(GetRawIssueExpected(rawIssue)), finalizedIssue.Expected)
		assert.Equal(t, core.ZodTypeCode(GetRawIssueReceived(rawIssue)), finalizedIssue.Received)
		assert.Equal(t, GetRawIssueMinimum(rawIssue), finalizedIssue.Minimum)
		assert.Equal(t, GetRawIssueMaximum(rawIssue), finalizedIssue.Maximum)
		assert.Equal(t, GetRawIssueInclusive(rawIssue), finalizedIssue.Inclusive)
		assert.Equal(t, GetRawIssueOrigin(rawIssue), finalizedIssue.Origin)

		// Check type-specific accessors work on finalized issue
		expected, ok := GetIssueExpected(finalizedIssue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeString, expected)

		received, ok := GetIssueReceived(finalizedIssue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeNumber, received)
	})

	t.Run("accessor methods work with raw issues", func(t *testing.T) {
		// Test that raw issues work with accessors
		invalidTypeIssue := NewRawIssue(core.InvalidType, "test", WithExpected("string"), WithReceived("number"))
		tooBigIssue := NewRawIssue(core.TooBig, "150", WithOrigin("number"), WithMaximum(100), WithInclusive(true))
		formatIssue := NewRawIssue(core.InvalidFormat, "invalid@", WithFormat("email"))

		// Verify accessors work correctly
		assert.Equal(t, "string", GetRawIssueExpected(invalidTypeIssue))

		assert.Equal(t, "number", GetRawIssueOrigin(tooBigIssue))
		assert.Equal(t, 100, GetRawIssueMaximum(tooBigIssue))
		assert.True(t, GetRawIssueInclusive(tooBigIssue))

		assert.Equal(t, "email", GetRawIssueFormat(formatIssue))
	})
}
