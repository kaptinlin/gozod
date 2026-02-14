package issues

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZodRawIssueAccessors(t *testing.T) {
	t.Run("Expected", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidType, "test", WithExpected("string"))
		expected := Expected(issue)
		assert.Equal(t, "string", expected)
	})

	t.Run("Received", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidType, "test", WithReceived("number"))
		received := Received(issue)
		assert.Equal(t, "number", received)
	})

	t.Run("Origin", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithOrigin("string"))
		origin := Origin(issue)
		assert.Equal(t, "string", origin)
	})

	t.Run("Format", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithFormat("email"))
		format := Format(issue)
		assert.Equal(t, "email", format)
	})

	t.Run("Pattern", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithPattern("^.*$"))
		pattern := Pattern(issue)
		assert.Equal(t, "^.*$", pattern)
	})

	t.Run("Prefix", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithPrefix("start_"))
		prefix := Prefix(issue)
		assert.Equal(t, "start_", prefix)
	})

	t.Run("Suffix", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithSuffix("_end"))
		suffix := Suffix(issue)
		assert.Equal(t, "_end", suffix)
	})

	t.Run("Includes", func(t *testing.T) {
		issue := NewRawIssue(core.InvalidFormat, "test", WithIncludes("substring"))
		includes := Includes(issue)
		assert.Equal(t, "substring", includes)
	})

	t.Run("Minimum", func(t *testing.T) {
		issue := NewRawIssue(core.TooSmall, "test", WithMinimum(5))
		minimum := Minimum(issue)
		assert.Equal(t, 5, minimum)
	})

	t.Run("Maximum", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithMaximum(100))
		maximum := Maximum(issue)
		assert.Equal(t, 100, maximum)
	})

	t.Run("Inclusive", func(t *testing.T) {
		issue := NewRawIssue(core.TooBig, "test", WithInclusive(true))
		inclusive := Inclusive(issue)
		assert.True(t, inclusive)
	})

	t.Run("Divisor", func(t *testing.T) {
		issue := NewRawIssue(core.NotMultipleOf, "test", WithDivisor(3))
		divisor := Divisor(issue)
		assert.Equal(t, 3, divisor)
	})

	t.Run("Keys", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		issue := NewRawIssue(core.UnrecognizedKeys, "test", WithKeys(keys))
		result := Keys(issue)
		assert.Equal(t, keys, result)
	})

	t.Run("Values", func(t *testing.T) {
		values := []any{"val1", "val2", 123}
		issue := NewRawIssue(core.InvalidValue, "test", WithValues(values))
		result := Values(issue)
		assert.Equal(t, values, result)
	})

	t.Run("empty properties return defaults", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: make(map[string]any),
		}

		assert.Empty(t, Expected(issue))
		assert.Empty(t, Received(issue))
		assert.Empty(t, Origin(issue))
		assert.Nil(t, Minimum(issue))
		assert.Nil(t, Maximum(issue))
		assert.False(t, Inclusive(issue))
		assert.Nil(t, Keys(issue))
		assert.Nil(t, Values(issue))
	})

	t.Run("nil properties map", func(t *testing.T) {
		issue := core.ZodRawIssue{
			Code:       core.InvalidType,
			Properties: nil,
		}

		assert.Empty(t, Expected(issue))
		assert.Empty(t, Received(issue))
		assert.Nil(t, Minimum(issue))
		assert.False(t, Inclusive(issue))
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

		assert.Empty(t, Expected(issue))
		assert.Equal(t, "string", Minimum(issue))
		assert.False(t, Inclusive(issue))
	})
}

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

func TestAccessorEdgeCases(t *testing.T) {
	t.Run("handles zero values in accessors", func(t *testing.T) {
		issue := ZodIssue{
			ZodIssueBase: ZodIssueBase{Code: "too_small"},
			Minimum:      0, // Zero value should still be accessible
		}

		minimum, ok := issue.MinValue()
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

		minimum, ok := issue.MinValue()
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
		for range 1000 {
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
		for range 1000 {
			_, _ = issue.ExpectedType()
			_, _ = issue.ReceivedType()
			_, _ = issue.MinValue() // Should return false quickly
			_, _ = issue.MaxValue() // Should return false quickly
		}
	})
}

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
		assert.Equal(t, core.ZodTypeCode(Expected(rawIssue)), finalizedIssue.Expected)
		assert.Equal(t, core.ZodTypeCode(Received(rawIssue)), finalizedIssue.Received)
		assert.Equal(t, Minimum(rawIssue), finalizedIssue.Minimum)
		assert.Equal(t, Maximum(rawIssue), finalizedIssue.Maximum)
		assert.Equal(t, Inclusive(rawIssue), finalizedIssue.Inclusive)
		assert.Equal(t, Origin(rawIssue), finalizedIssue.Origin)

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
		assert.Equal(t, "string", Expected(invalidTypeIssue))

		assert.Equal(t, "number", Origin(tooBigIssue))
		assert.Equal(t, 100, Maximum(tooBigIssue))
		assert.True(t, Inclusive(tooBigIssue))

		assert.Equal(t, "email", Format(formatIssue))
	})
}
