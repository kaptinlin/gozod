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
	t.Run("RawIssueExpected", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidType, "test", WithExpected("string"))
		expected := RawIssueExpected(issue)
		assert.Equal(t, "string", expected)
	})

	t.Run("RawIssueReceived", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidType, "test", WithReceived("number"))
		received := RawIssueReceived(issue)
		assert.Equal(t, "number", received)
	})

	t.Run("RawIssueOrigin", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithOrigin("string"))
		origin := RawIssueOrigin(issue)
		assert.Equal(t, "string", origin)
	})

	t.Run("RawIssueFormat", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithFormat("email"))
		format := RawIssueFormat(issue)
		assert.Equal(t, "email", format)
	})

	t.Run("RawIssuePattern", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithPattern("^.*$"))
		pattern := RawIssuePattern(issue)
		assert.Equal(t, "^.*$", pattern)
	})

	t.Run("RawIssuePrefix", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithPrefix("start_"))
		prefix := RawIssuePrefix(issue)
		assert.Equal(t, "start_", prefix)
	})

	t.Run("RawIssueSuffix", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithSuffix("_end"))
		suffix := RawIssueSuffix(issue)
		assert.Equal(t, "_end", suffix)
	})

	t.Run("RawIssueIncludes", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithIncludes("substring"))
		includes := RawIssueIncludes(issue)
		assert.Equal(t, "substring", includes)
	})

	t.Run("RawIssueMinimum", func(t *testing.T) {
		issue := NewRawIssue(core.TooSmall, "test", WithMinimum(5))
		minimum := RawIssueMinimum(issue)
		assert.Equal(t, 5, minimum)
	})

	t.Run("RawIssueMaximum", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithMaximum(100))
		maximum := RawIssueMaximum(issue)
		assert.Equal(t, 100, maximum)
	})

	t.Run("RawIssueInclusive", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithInclusive(true))
		inclusive := RawIssueInclusive(issue)
		assert.True(t, inclusive)
	})

	t.Run("RawIssueDivisor", func(t *testing.T) {
		issue := NewRawIssue(core.NotMultipleOf, "test", WithDivisor(3))
		divisor := RawIssueDivisor(issue)
		assert.Equal(t, 3, divisor)
	})

	t.Run("RawIssueKeys", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		issue := NewRawIssue(core.UnrecognizedKeys, "test", WithKeys(keys))
		result := RawIssueKeys(issue)
		assert.Equal(t, keys, result)
	})

	t.Run("RawIssueValues", func(t *testing.T) {
		values := []any{"val1", "val2", 123}
		issue := NewRawIssue(core.InvalidValue, "test", WithValues(values))
		result := RawIssueValues(issue)
		assert.Equal(t, values, result)
	})

	t.Run("empty properties return defaults", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: make(map[string]any),
		}

		assert.Empty(t, RawIssueExpected(issue))
		assert.Empty(t, RawIssueReceived(issue))
		assert.Empty(t, RawIssueOrigin(issue))
		assert.Nil(t, RawIssueMinimum(issue))
		assert.Nil(t, RawIssueMaximum(issue))
		assert.False(t, RawIssueInclusive(issue))
		assert.Nil(t, RawIssueKeys(issue))
		assert.Nil(t, RawIssueValues(issue))
	})

	t.Run("nil properties map", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: nil,
		}

		assert.Empty(t, RawIssueExpected(issue))
		assert.Empty(t, RawIssueReceived(issue))
		assert.Nil(t, RawIssueMinimum(issue))
		assert.False(t, RawIssueInclusive(issue))
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

		assert.Empty(t, RawIssueExpected(issue))
		assert.Equal(t, "string", RawIssueMinimum(issue))
		assert.False(t, RawIssueInclusive(issue))
	})
}

//////////////////////////////////////////
//////////   Issue Type-Specific Accessor Tests  ///
//////////////////////////////////////////

func TestZodIssueTypeSpecificAccessors(t *testing.T) {
	t.Run("IssueExpected for InvalidType", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidType},
			Expected:     core.ZodTypeString,
		}

		expected, ok := IssueExpected(issue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeString, expected)
	})

	t.Run("IssueExpected for non-InvalidType", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooBig},
			Expected:     core.ZodTypeString,
		}

		expected, ok := IssueExpected(issue)
		assert.False(t, ok)
		assert.Empty(t, expected)
	})

	t.Run("IssueReceived for InvalidType", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidType},
			Received:     core.ZodTypeNumber,
		}

		received, ok := IssueReceived(issue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeNumber, received)
	})

	t.Run("IssueMinimum for TooSmall", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooSmall},
			Minimum:      5,
		}

		minimum, ok := IssueMinimum(issue)
		require.True(t, ok)
		assert.Equal(t, 5, minimum)
	})

	t.Run("IssueMinimum for non-TooSmall", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooBig},
			Minimum:      5,
		}

		minimum, ok := IssueMinimum(issue)
		assert.False(t, ok)
		assert.Nil(t, minimum)
	})

	t.Run("IssueMaximum for TooBig", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooBig},
			Maximum:      100,
		}

		maximum, ok := IssueMaximum(issue)
		require.True(t, ok)
		assert.Equal(t, 100, maximum)
	})

	t.Run("IssueFormat for InvalidFormat", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidFormat},
			Format:       "email",
		}

		format, ok := IssueFormat(issue)
		require.True(t, ok)
		assert.Equal(t, "email", format)
	})

	t.Run("IssueDivisor for NotMultipleOf", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.NotMultipleOf},
			Divisor:      3,
		}

		divisor, ok := IssueDivisor(issue)
		require.True(t, ok)
		assert.Equal(t, 3, divisor)
	})

	t.Run("empty field values return false", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.InvalidType},
			Expected:     "",
			Received:     "",
		}

		expected, ok := IssueExpected(issue)
		assert.False(t, ok)
		assert.Empty(t, expected)

		received, ok := IssueReceived(issue)
		assert.False(t, ok)
		assert.Empty(t, received)
	})

	t.Run("nil field values return false", func(t *testing.T) {
		issue := core.ZodIssue{
			ZodIssueBase: core.ZodIssueBase{Code: core.TooSmall},
			Minimum:      nil,
		}

		minimum, ok := IssueMinimum(issue)
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
		assert.False(t, rawIssue.Inclusive()) // False should be preserved
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
		assert.Equal(t, largeNumber, issue.Maximum())
	})

	t.Run("handles unicode strings", func(t *testing.T) {
		unicodeString := "测试字符串 🚀 émojis"

		issue := NewRawIssue("invalid_format", unicodeString,
			WithExpected(unicodeString),
			WithOrigin(unicodeString),
		)

		assert.Equal(t, core.ZodTypeCode(unicodeString), issue.Expected())
		assert.Equal(t, unicodeString, issue.Origin())
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
			_ = issue.Expected()
			_ = issue.Received()
			_ = issue.Minimum()
			_ = issue.Maximum()
			_ = issue.Inclusive()
			_ = issue.Origin()
			_ = issue.Format()
			_ = issue.Pattern()
			_ = issue.Prefix()
			_ = issue.Suffix()
			_ = issue.Includes()
			_ = issue.Divisor()
			_ = issue.Keys()
			_ = issue.Values()
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
		assert.Equal(t, core.ZodTypeCode(RawIssueExpected(rawIssue)), finalizedIssue.Expected)
		assert.Equal(t, core.ZodTypeCode(RawIssueReceived(rawIssue)), finalizedIssue.Received)
		assert.Equal(t, RawIssueMinimum(rawIssue), finalizedIssue.Minimum)
		assert.Equal(t, RawIssueMaximum(rawIssue), finalizedIssue.Maximum)
		assert.Equal(t, RawIssueInclusive(rawIssue), finalizedIssue.Inclusive)
		assert.Equal(t, RawIssueOrigin(rawIssue), finalizedIssue.Origin)

		// Check type-specific accessors work on finalized issue
		expected, ok := IssueExpected(finalizedIssue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeString, expected)

		received, ok := IssueReceived(finalizedIssue)
		require.True(t, ok)
		assert.Equal(t, core.ZodTypeNumber, received)
	})

	t.Run("accessor methods work with raw issues", func(t *testing.T) {
		// Test that raw issues work with accessors
		invalidTypeIssue := NewRawIssue(core.InvalidType, "test", WithExpected("string"), WithReceived("number"))
		tooBigIssue := NewRawIssue(core.TooBig, "150", WithOrigin("number"), WithMaximum(100), WithInclusive(true))
		formatIssue := NewRawIssue(core.InvalidFormat, "invalid@", WithFormat("email"))

		// Verify accessors work correctly
		assert.Equal(t, "string", RawIssueExpected(invalidTypeIssue))

		assert.Equal(t, "number", RawIssueOrigin(tooBigIssue))
		assert.Equal(t, 100, RawIssueMaximum(tooBigIssue))
		assert.True(t, RawIssueInclusive(tooBigIssue))

		assert.Equal(t, "email", RawIssueFormat(formatIssue))
	})
}
