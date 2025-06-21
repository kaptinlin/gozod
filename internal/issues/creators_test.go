package issues

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Message-based Creation Tests ///
//////////////////////////////////////////

func TestNewRawIssueFromMessage(t *testing.T) {
	message := "Custom error message"
	input := "test_input"
	instance := "test_instance"

	issue := NewRawIssueFromMessage(message, input, instance)

	assert.Equal(t, core.Custom, issue.Code)
	assert.Equal(t, message, issue.Message)
	assert.Equal(t, input, issue.Input)
	assert.Equal(t, instance, issue.Inst)
	assert.Empty(t, issue.Path)
	// Properties are initialized lazily, so initially nil
	assert.Nil(t, issue.Properties)
}

//////////////////////////////////////////
//////////   Error Creation Helper Tests  ///
//////////////////////////////////////////

func TestErrorCreationHelpers(t *testing.T) {
	t.Run("CreateInvalidTypeIssue", func(t *testing.T) {
		issue := CreateInvalidTypeIssue("string", "test_input")

		require.Equal(t, core.InvalidType, issue.Code)
		// Check properties using mapx accessors
		expected, _ := issue.Properties["expected"].(string)
		assert.Equal(t, "string", expected)
	})

	t.Run("CreateTooBigIssue", func(t *testing.T) {
		issue := CreateTooBigIssue(100, true, "number", 150)

		require.Equal(t, core.TooBig, issue.Code)
		maximum, _ := issue.Properties["maximum"].(int)
		assert.Equal(t, 100, maximum)
		inclusive, _ := issue.Properties["inclusive"].(bool)
		assert.True(t, inclusive)
		origin, _ := issue.Properties["origin"].(string)
		assert.Equal(t, "number", origin)
	})

	t.Run("CreateTooSmallIssue", func(t *testing.T) {
		issue := CreateTooSmallIssue(5, false, "string", 3)

		require.Equal(t, core.TooSmall, issue.Code)
		minimum, _ := issue.Properties["minimum"].(int)
		assert.Equal(t, 5, minimum)
		inclusive, _ := issue.Properties["inclusive"].(bool)
		assert.False(t, inclusive)
		origin, _ := issue.Properties["origin"].(string)
		assert.Equal(t, "string", origin)
	})

	t.Run("CreateInvalidFormatIssue", func(t *testing.T) {
		issue := CreateInvalidFormatIssue("email", "invalid@", nil)

		require.Equal(t, core.InvalidFormat, issue.Code)
		format, _ := issue.Properties["format"].(string)
		assert.Equal(t, "email", format)
	})

	t.Run("CreateNotMultipleOfIssue", func(t *testing.T) {
		issue := CreateNotMultipleOfIssue(2, "number", 7)

		require.Equal(t, core.NotMultipleOf, issue.Code)
		divisor, _ := issue.Properties["divisor"].(int)
		assert.Equal(t, 2, divisor)
	})

	t.Run("CreateCustomIssue", func(t *testing.T) {
		issue := CreateCustomIssue("Custom validation failed", nil, "test_input")

		require.Equal(t, core.Custom, issue.Code)
		require.Equal(t, "Custom validation failed", issue.Message)
	})

	t.Run("CreateInvalidValueIssue", func(t *testing.T) {
		validValues := []any{"val1", "val2", "val3"}
		issue := CreateInvalidValueIssue(validValues, "invalid_value")

		require.Equal(t, core.InvalidValue, issue.Code)
		values, _ := issue.Properties["values"].([]any)
		assert.Equal(t, validValues, values)
	})

	t.Run("CreateUnrecognizedKeysIssue", func(t *testing.T) {
		keys := []string{"extraKey1", "extraKey2"}
		issue := CreateUnrecognizedKeysIssue(keys, nil)

		require.Equal(t, core.UnrecognizedKeys, issue.Code)
		actualKeys, _ := issue.Properties["keys"].([]string)
		assert.Equal(t, keys, actualKeys)
	})
}

//////////////////////////////////////////
//////////   Helper Functions with Options Tests ///
//////////////////////////////////////////

func TestCreationHelpersWithOptions(t *testing.T) {
	t.Run("helper functions accept additional options", func(t *testing.T) {
		path := []any{"user", "email"}
		issue := NewRawIssue(core.InvalidFormat, "invalid@",
			WithPath(path),
			WithMessage("Custom email error"),
			WithFormat("email"),
		)

		assert.Equal(t, path, issue.Path)
		assert.Equal(t, "Custom email error", issue.Message)
		format, _ := issue.Properties["format"].(string)
		assert.Equal(t, "email", format)
	})

	t.Run("CreateInvalidTypeIssue with additional properties", func(t *testing.T) {
		issue := CreateInvalidTypeIssue("string", "input")

		// Add additional properties using mapx
		issue.Properties["custom"] = "value"
		issue.Path = []any{"data", "field"}
		issue.Message = "Custom type error"

		assert.Equal(t, []any{"data", "field"}, issue.Path)
		assert.Equal(t, "Custom type error", issue.Message)
		custom, _ := issue.Properties["custom"].(string)
		assert.Equal(t, "value", custom)
	})

	t.Run("CreateTooBigIssue with continue flag", func(t *testing.T) {
		issue := CreateTooBigIssue(100, false, "number", 200)
		issue.Message = "Value is way too big"
		issue.Continue = true

		assert.Equal(t, "Value is way too big", issue.Message)
		assert.True(t, issue.Continue)
	})

	t.Run("CreateTooSmallIssue with pattern", func(t *testing.T) {
		issue := CreateTooSmallIssue(5, true, "string", 2)
		issue.Message = "String too short"
		issue.Properties["pattern"] = "^.{5,}$"

		assert.Equal(t, "String too short", issue.Message)
		pattern, _ := issue.Properties["pattern"].(string)
		assert.Equal(t, "^.{5,}$", pattern)
	})

	t.Run("CreateCustomIssue with params", func(t *testing.T) {
		params := map[string]any{"rule": "custom_rule", "value": 42}
		props := map[string]any{"params": params}
		issue := CreateCustomIssue("Custom error", props, "input")
		issue.Path = []any{"custom", "field"}

		actualParams, _ := issue.Properties["params"].(map[string]any)
		assert.Equal(t, params, actualParams)
		assert.Equal(t, []any{"custom", "field"}, issue.Path)
	})
}

//////////////////////////////////////////
//////////   Complex Creation Scenarios Tests ///
//////////////////////////////////////////

func TestComplexCreationScenarios(t *testing.T) {
	t.Run("create issue with all possible properties", func(t *testing.T) {
		keys := []string{"key1", "key2"}
		values := []any{"val1", "val2"}
		params := map[string]any{"custom": "data", "level": 2}
		path := []any{"nested", "deep", "field"}
		instance := "complex_schema"

		issue := NewRawIssue(core.Custom, "complex_input",
			WithExpected("complex_type"),
			WithReceived("simple_type"),
			WithMinimum(10),
			WithMaximum(100),
			WithInclusive(true),
			WithOrigin("validation"),
			WithFormat("complex_format"),
			WithPattern("^complex.*$"),
			WithPrefix("complex_"),
			WithSuffix("_complex"),
			WithIncludes("complex"),
			WithDivisor(5),
			WithKeys(keys),
			WithValues(values),
			WithAlgorithm("COMPLEX256"),
			WithParams(params),
			WithPath(path),
			WithInst(instance),
			WithContinue(false),
			WithMessage("Complex validation error"),
		)

		// Verify all properties are set correctly
		assert.Equal(t, core.Custom, issue.Code)
		assert.Equal(t, "complex_input", issue.Input)
		assert.Equal(t, "Complex validation error", issue.Message)
		assert.Equal(t, "complex_type", issue.Properties["expected"])
		assert.Equal(t, "simple_type", issue.Properties["received"])
		assert.Equal(t, 10, issue.Properties["minimum"])
		assert.Equal(t, 100, issue.Properties["maximum"])
		assert.True(t, issue.Properties["inclusive"].(bool))
		assert.Equal(t, "validation", issue.Properties["origin"])
		assert.Equal(t, "complex_format", issue.Properties["format"])
		assert.Equal(t, "^complex.*$", issue.Properties["pattern"])
		assert.Equal(t, "complex_", issue.Properties["prefix"])
		assert.Equal(t, "_complex", issue.Properties["suffix"])
		assert.Equal(t, "complex", issue.Properties["includes"])
		assert.Equal(t, 5, issue.Properties["divisor"])
		assert.Equal(t, keys, issue.Properties["keys"])
		assert.Equal(t, values, issue.Properties["values"])
		assert.Equal(t, "COMPLEX256", issue.Properties["algorithm"])
		assert.Equal(t, params, issue.Properties["params"])
		assert.Equal(t, path, issue.Path)
		assert.Equal(t, instance, issue.Inst)
		assert.False(t, issue.Continue)
	})

	t.Run("create multiple issues with different types", func(t *testing.T) {
		issues := []core.ZodRawIssue{
			CreateInvalidTypeIssue("string", "123"),
			CreateTooBigIssue(100, true, "number", 150),
			CreateTooSmallIssue(5, false, "string", "hi"),
			CreateInvalidFormatIssue("email", "invalid@", nil),
			CreateNotMultipleOfIssue(2, "number", 7),
			CreateCustomIssue("Custom error", nil, "data"),
		}

		require.Len(t, issues, 6)

		// Verify each issue type
		assert.Equal(t, core.InvalidType, issues[0].Code)
		assert.Equal(t, core.TooBig, issues[1].Code)
		assert.Equal(t, core.TooSmall, issues[2].Code)
		assert.Equal(t, core.InvalidFormat, issues[3].Code)
		assert.Equal(t, core.NotMultipleOf, issues[4].Code)
		assert.Equal(t, core.Custom, issues[5].Code)

		// Verify properties are correctly set for each
		assert.Equal(t, "string", issues[0].Properties["expected"])
		assert.Equal(t, 100, issues[1].Properties["maximum"])
		assert.Equal(t, 5, issues[2].Properties["minimum"])
		assert.Equal(t, "email", issues[3].Properties["format"])
		assert.Equal(t, 2, issues[4].Properties["divisor"])
		assert.Equal(t, "Custom error", issues[5].Message)
	})
}

//////////////////////////////////////////
//////////   Edge Cases Tests          ///
//////////////////////////////////////////

func TestCreationEdgeCases(t *testing.T) {
	t.Run("nil properties handling", func(t *testing.T) {
		issue := CreateCustomIssue("Error", nil, "input")
		assert.NotNil(t, issue.Properties)
	})

	t.Run("empty slices", func(t *testing.T) {
		issue := CreateInvalidValueIssue([]any{}, "input")
		values, _ := issue.Properties["values"].([]any)
		assert.Empty(t, values)
	})

	t.Run("duplicate keys in slice", func(t *testing.T) {
		keys := []string{"key1", "key2", "key1", "key3", "key2"}
		issue := CreateUnrecognizedKeysIssue(keys, nil)

		// The slicex.Unique should remove duplicates
		actualKeys, _ := issue.Properties["keys"].([]string)
		assert.Len(t, actualKeys, 3) // Should be unique
	})
}

//////////////////////////////////////////
//////////   Performance Tests         ///
//////////////////////////////////////////

func TestCreationPerformance(t *testing.T) {
	t.Run("large value arrays", func(t *testing.T) {
		largeValues := make([]any, 1000)
		for i := range largeValues {
			largeValues[i] = i
		}

		issue := CreateInvalidValueIssue(largeValues, "input")
		values, _ := issue.Properties["values"].([]any)
		assert.Len(t, values, 1000)
	})
}
