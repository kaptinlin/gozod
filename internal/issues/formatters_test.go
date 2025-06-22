package issues

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//////////////////////////////////////////
//////////   Default Message Generation Tests ///
//////////////////////////////////////////

func TestGenerateDefaultMessage(t *testing.T) {
	t.Run("generates appropriate default messages for different issue types", func(t *testing.T) {
		testCases := []struct {
			name     string
			rawIssue ZodRawIssue
			expected string
		}{
			{
				name: "too_small message",
				rawIssue: NewRawIssue(core.TooSmall, 3,
					WithMinimum(5),
					WithOrigin("string"),
				),
				expected: "Too small: expected string to have >=5 characters",
			},
			{
				name: "too_big message",
				rawIssue: NewRawIssue("too_big", 150,
					WithMaximum(100),
					WithOrigin("number"),
				),
				expected: "Too big: expected number to be <=100",
			},
			{
				name: "invalid_format message",
				rawIssue: NewRawIssue(core.InvalidFormat, "invalid@",
					WithFormat("email"),
				),
				expected: "Invalid email address",
			},
			{
				name: "not_multiple_of message",
				rawIssue: NewRawIssue("not_multiple_of", 7,
					WithDivisor(2),
				),
				expected: "Invalid number: must be a multiple of 2",
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
				rawIssue: NewRawIssue(core.Custom, nil),
				expected: "Invalid input",
			},
			{
				name: "invalid_type message",
				rawIssue: NewRawIssue(core.InvalidType, 123,
					WithExpected("string"),
				),
				expected: "Invalid input: expected string, received number",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				message := GenerateDefaultMessage(tc.rawIssue)
				assert.Contains(t, message, tc.expected)
			})
		}
	})

	t.Run("handles missing properties gracefully", func(t *testing.T) {
		// Test invalid_type without expected/received
		rawIssue := NewRawIssue(core.InvalidType, nil)
		message := GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, "Invalid input")

		// Test too_small without minimum
		rawIssue = NewRawIssue(core.TooSmall, nil)
		message = GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, "Too small")

		// Test invalid_format without format
		rawIssue = NewRawIssue(core.InvalidFormat, nil)
		message = GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, "Invalid")
	})

	t.Run("handles unknown issue codes", func(t *testing.T) {
		rawIssue := NewRawIssue("unknown_code", nil)
		message := GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, "Invalid input")
	})

	t.Run("uses origin information when available", func(t *testing.T) {
		rawIssue := NewRawIssue(core.TooSmall, "hi",
			WithMinimum(5),
			WithOrigin("string"),
		)
		message := GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, "string")
		assert.Contains(t, message, ">=")
	})

	t.Run("handles inclusive/exclusive bounds", func(t *testing.T) {
		// Inclusive bounds
		rawIssue := NewRawIssue(core.TooSmall, 5,
			WithMinimum(5),
			WithInclusive(true),
			WithOrigin("number"),
		)
		message := GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, ">=")

		// Exclusive bounds
		rawIssue = NewRawIssue(core.TooSmall, 4,
			WithMinimum(5),
			WithInclusive(false),
			WithOrigin("number"),
		)
		message = GenerateDefaultMessage(rawIssue)
		assert.Contains(t, message, ">")
	})
}

//////////////////////////////////////////
//////////   Type Detection Compatibility Tests ///
//////////////////////////////////////////

func TestParsedTypeToString(t *testing.T) {
	t.Run("matches reference type detection behaviour", func(t *testing.T) {
		// Create a big.Int for testing
		bigIntValue := big.NewInt(123)

		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{"string type", "hello", "string"},
			{"number type", 42, "number"},
			{"boolean type", true, "boolean"},
			{"null type", nil, "null"},
			{"array type", []int{1, 2, 3}, "array"},
			{"slice type", []string{"a", "b"}, "array"}, // Go slices map to array
			{"object type", map[string]any{"key": "value"}, "object"},
			{"map type", map[int]string{1: "one"}, "object"}, // Go maps map to object
			{"float64 type", 3.14, "number"},
			{"float32 type", float32(2.71), "number"},
			{"NaN value", math.NaN(), "NaN"},
			{"positive infinity", math.Inf(1), "Infinity"},
			{"negative infinity", math.Inf(-1), "Infinity"},
			{"bigint type", bigIntValue, "bigint"}, // Use actual *big.Int
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := ParsedTypeToString(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("handles Go-specific types", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{"int8", int8(8), "number"},
			{"int16", int16(16), "number"},
			{"int32", int32(32), "number"},
			{"int64", int64(64), "number"},
			{"uint8", uint8(8), "number"},
			{"uint16", uint16(16), "number"},
			{"uint32", uint32(32), "number"},
			{"uint64", uint64(64), "number"},
			{"complex64", complex64(1 + 2i), "complex"},
			{"complex128", complex128(3 + 4i), "complex"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := ParsedTypeToString(tc.input)
				// For Go numeric types, they should all map to "number" just like the original reference implementation
				if strings.Contains(tc.expected, "number") {
					assert.Equal(t, "number", result)
				} else {
					assert.Equal(t, tc.expected, result)
				}
			})
		}
	})
}

//////////////////////////////////////////
//////////   String Formatting Compatibility Tests ///
//////////////////////////////////////////

func TestStringifyPrimitive(t *testing.T) {
	t.Run("formats primitives like reference implementation", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			expected string
		}{
			{"string with quotes", "hello", `"hello"`},
			{"null value", nil, "null"},
			{"boolean true", true, "true"},
			{"boolean false", false, "false"},
			{"integer", 42, "42"},
			{"float without decimals", 5.0, "5"},
			{"float with decimals", 3.14, "3.14"},
			{"NaN value", math.NaN(), "NaN"},
			{"positive infinity", math.Inf(1), "Infinity"},
			{"negative infinity", math.Inf(-1), "-Infinity"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := StringifyPrimitive(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})
}

//////////////////////////////////////////
//////////   Size Constraint Formatting Tests ///
//////////////////////////////////////////

func TestSizeConstraintFormatting(t *testing.T) {
	t.Run("formats string length constraints", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}

		// Too small string
		rawIssue := NewRawIssue(core.TooSmall, "hi",
			WithMinimum(5),
			WithOrigin("string"),
			WithInclusive(true),
		)
		message := formatter.formatSizeConstraint(rawIssue, true)
		assert.Equal(t, "Too small: expected string to have >=5 characters", message)

		// Too big string
		rawIssue = NewRawIssue("too_big", "hello world",
			WithMaximum(5),
			WithOrigin("string"),
			WithInclusive(true),
		)
		message = formatter.formatSizeConstraint(rawIssue, false)
		assert.Equal(t, "Too big: expected string to have <=5 characters", message)
	})

	t.Run("formats array length constraints", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}

		// Too small array
		rawIssue := NewRawIssue(core.TooSmall, []string{"a", "b"},
			WithMinimum(3),
			WithOrigin("array"),
			WithInclusive(true),
		)
		message := formatter.formatSizeConstraint(rawIssue, true)
		assert.Equal(t, "Too small: expected array to have >=3 items", message)
	})

	t.Run("handles exclusive bounds", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}

		// Exclusive minimum
		rawIssue := NewRawIssue(core.TooSmall, 5,
			WithMinimum(5),
			WithOrigin("number"),
			WithInclusive(false),
		)
		message := formatter.formatSizeConstraint(rawIssue, true)
		assert.Equal(t, "Too small: expected number to be >5", message)
	})
}

//////////////////////////////////////////
//////////   Format Noun Tests ///
//////////////////////////////////////////

func TestGetFormatNoun(t *testing.T) {
	t.Run("returns standard format nouns", func(t *testing.T) {
		testCases := []struct {
			format   string
			expected string
		}{
			{"email", "email address"},
			{"url", "URL"},
			{"uuid", "UUID"},
			{"datetime", "ISO datetime"},
			{"ipv4", "IPv4 address"},
			{"base64", "base64-encoded string"},
			{"jwt", "JWT"},
		}

		for _, tc := range testCases {
			t.Run(tc.format, func(t *testing.T) {
				result := GetFormatNoun(tc.format)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("returns Go-specific format nouns", func(t *testing.T) {
		testCases := []struct {
			format   string
			expected string
		}{
			{"int8", "8-bit integer"},
			{"int64", "64-bit integer"},
			{"uint32", "32-bit unsigned integer"},
			{"float32", "32-bit float"},
			{"complex128", "128-bit complex number"},
		}

		for _, tc := range testCases {
			t.Run(tc.format, func(t *testing.T) {
				result := GetFormatNoun(tc.format)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("returns format name for unknown formats", func(t *testing.T) {
		result := GetFormatNoun("unknown_format")
		assert.Equal(t, "unknown_format", result)
	})
}

//////////////////////////////////////////
//////////   String Validation Formatting Tests ///
//////////////////////////////////////////

func TestStringValidationFormatting(t *testing.T) {
	t.Run("formats starts_with validation", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}
		rawIssue := NewRawIssue(core.InvalidFormat, "hello",
			WithFormat("starts_with"),
			WithPrefix("test"),
		)
		message := formatter.formatStringValidation(rawIssue, "starts_with")
		assert.Equal(t, `Invalid string: must start with "test"`, message)
	})

	t.Run("formats ends_with validation", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}
		rawIssue := NewRawIssue(core.InvalidFormat, "hello",
			WithFormat("ends_with"),
			WithSuffix("world"),
		)
		message := formatter.formatStringValidation(rawIssue, "ends_with")
		assert.Equal(t, `Invalid string: must end with "world"`, message)
	})

	t.Run("formats includes validation", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}
		rawIssue := NewRawIssue(core.InvalidFormat, "hello",
			WithFormat("includes"),
			WithIncludes("ell"),
		)
		message := formatter.formatStringValidation(rawIssue, "includes")
		assert.Equal(t, `Invalid string: must include "ell"`, message)
	})

	t.Run("formats regex validation", func(t *testing.T) {
		formatter := &DefaultMessageFormatter{}
		rawIssue := NewRawIssue(core.InvalidFormat, "hello",
			WithFormat("regex"),
			WithPattern("[0-9]+"),
		)
		message := formatter.formatStringValidation(rawIssue, "regex")
		assert.Equal(t, "Invalid string: must match pattern [0-9]+", message)
	})
}

//////////////////////////////////////////
//////////   ZodError Creation Tests   ///
//////////////////////////////////////////

func TestNewZodError(t *testing.T) {
	t.Run("creates error with basic properties", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Invalid input: expected string, received number",
					Path:    []any{"name"},
				},
				Expected: "string",
				Received: "number",
			},
		}

		err := NewZodError(issues)

		require.NotNil(t, err)
		require.Equal(t, "ZodError", err.Name)
		require.Len(t, err.Issues, 1)
		require.Len(t, err.Zod.Def, 1)
		assert.Equal(t, issues, err.Issues)
		assert.Equal(t, issues, err.Zod.Def)
		assert.Nil(t, err.Type)
		assert.Nil(t, err.Zod.Output)
	})

	t.Run("creates error with multiple issues", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Invalid type",
					Path:    []any{"name"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.TooSmall,
					Message: "Too small",
					Path:    []any{"age"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidFormat,
					Message: "Invalid email",
					Path:    []any{"email"},
				},
			},
		}

		err := NewZodError(issues)

		require.Len(t, err.Issues, 3)
		assert.Equal(t, core.InvalidType, err.Issues[0].Code)
		assert.Equal(t, core.TooSmall, err.Issues[1].Code)
		assert.Equal(t, core.InvalidFormat, err.Issues[2].Code)
	})

	t.Run("creates error with empty issues slice", func(t *testing.T) {
		issues := []ZodIssue{}
		err := NewZodError(issues)

		require.NotNil(t, err)
		assert.Empty(t, err.Issues)
		assert.Empty(t, err.Zod.Def)
	})

	t.Run("creates error with nil issues slice", func(t *testing.T) {
		err := NewZodError(nil)

		require.NotNil(t, err)
		assert.Nil(t, err.Issues)
		assert.Nil(t, err.Zod.Def)
	})
}

//////////////////////////////////////////
//////////   ZodError Interface Tests  ///
//////////////////////////////////////////

func TestZodErrorInterface(t *testing.T) {
	t.Run("implements error interface correctly", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Invalid input: expected string, received number",
					Path:    []any{"user", "name"},
				},
			},
		}

		zodErr := NewZodError(issues)
		var err error = zodErr

		require.NotNil(t, err)
		errorStr := err.Error()
		assert.Contains(t, errorStr, "Invalid input: expected string, received number")
		assert.Contains(t, errorStr, "user.name: Invalid input: expected string, received number")
	})

	t.Run("error string contains all issues", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "First error",
					Path:    []any{"field1"},
				},
			},
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.TooSmall,
					Message: "Second error",
					Path:    []any{"field2"},
				},
			},
		}

		err := NewZodError(issues)
		errorStr := err.Error()

		assert.Contains(t, errorStr, "First error")
		assert.Contains(t, errorStr, "Second error")
		assert.Contains(t, errorStr, "field1")
		assert.Contains(t, errorStr, "field2")
	})
}

//////////////////////////////////////////
//////////   IsZodError Function Tests ///
//////////////////////////////////////////

func TestIsZodError(t *testing.T) {
	t.Run("identifies ZodError correctly", func(t *testing.T) {
		zodErr := NewZodError([]ZodIssue{
			{ZodIssueBase: ZodIssueBase{Code: "test", Message: "test"}},
		})

		var target *ZodError
		result := IsZodError(zodErr, &target)

		require.True(t, result)
		require.NotNil(t, target)
		assert.Equal(t, zodErr, target)
	})

	t.Run("returns false for non-ZodError", func(t *testing.T) {
		//nolint:err113 // Intentional regular error for testing IsZodError function
		regularErr := fmt.Errorf("regular error")

		var target *ZodError
		result := IsZodError(regularErr, &target)

		assert.False(t, result)
		assert.Nil(t, target)
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		var target *ZodError
		result := IsZodError(nil, &target)

		assert.False(t, result)
		assert.Nil(t, target)
	})

	t.Run("works without target parameter", func(t *testing.T) {
		zodErr := NewZodError([]ZodIssue{
			{ZodIssueBase: ZodIssueBase{Code: "test", Message: "test"}},
		})

		result := IsZodError(zodErr, nil)
		assert.True(t, result)

		//nolint:err113 // Intentional regular error for testing IsZodError function
		regularErr := fmt.Errorf("regular error")
		result = IsZodError(regularErr, nil)
		assert.False(t, result)
	})

	t.Run("handles wrapped errors", func(t *testing.T) {
		zodErr := NewZodError([]ZodIssue{
			{ZodIssueBase: ZodIssueBase{Code: "test", Message: "test"}},
		})
		wrappedErr := fmt.Errorf("wrapped: %w", zodErr)

		var target *ZodError
		result := IsZodError(wrappedErr, &target)

		require.True(t, result)
		require.NotNil(t, target)
		assert.Equal(t, zodErr, target)
	})
}

//////////////////////////////////////////
//////////   Comprehensive Compatibility Tests ///
//////////////////////////////////////////

func TestReferenceCompatibility(t *testing.T) {
	t.Run("invalid_type messages match exactly", func(t *testing.T) {
		testCases := []struct {
			name         string
			input        any
			expectedType string
			expectedMsg  string
		}{
			{"string to number", 42, "string", "Invalid input: expected string, received number"},
			{"number to string", "hello", "number", "Invalid input: expected number, received string"},
			{"boolean to string", true, "string", "Invalid input: expected string, received boolean"},
			{"null to string", nil, "string", "Invalid input: expected string, received null"},
			{"array to string", []int{1, 2}, "string", "Invalid input: expected string, received array"},
			{"object to string", map[string]int{"key": 1}, "string", "Invalid input: expected string, received object"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				rawIssue := NewRawIssue(core.InvalidType, tc.input,
					WithExpected(tc.expectedType),
				)
				message := GenerateDefaultMessage(rawIssue)
				assert.Equal(t, tc.expectedMsg, message)
			})
		}
	})

	t.Run("invalid_value message format", func(t *testing.T) {
		// Single value case
		rawIssue := NewRawIssue("invalid_value", "invalid",
			WithValues([]any{"valid"}),
		)
		message := GenerateDefaultMessage(rawIssue)
		assert.Equal(t, `Invalid input: expected "valid"`, message)

		// Multiple values case
		rawIssue = NewRawIssue("invalid_value", "invalid",
			WithValues([]any{"option1", "option2", "option3"}),
		)
		message = GenerateDefaultMessage(rawIssue)
		assert.Equal(t, `Invalid option: expected one of "option1"|"option2"|"option3"`, message)
	})

	t.Run("array size constraint messages", func(t *testing.T) {
		// Array too small
		rawIssue := NewRawIssue(core.TooSmall, []string{"a", "b"},
			WithMinimum(3),
			WithOrigin("array"),
			WithInclusive(true),
		)
		message := GenerateDefaultMessage(rawIssue)
		assert.Equal(t, "Too small: expected array to have >=3 items", message)

		// Array too big
		rawIssue = NewRawIssue("too_big", []string{"a", "b", "c", "d"},
			WithMaximum(2),
			WithOrigin("array"),
			WithInclusive(true),
		)
		message = GenerateDefaultMessage(rawIssue)
		assert.Equal(t, "Too big: expected array to have <=2 items", message)
	})

	t.Run("unrecognized_keys message variations", func(t *testing.T) {
		// Single key
		rawIssue := NewRawIssue("unrecognized_keys", map[string]any{},
			WithKeys([]string{"extra"}),
		)
		message := GenerateDefaultMessage(rawIssue)
		assert.Equal(t, `Unrecognized key: "extra"`, message)

		// Multiple keys
		rawIssue = NewRawIssue("unrecognized_keys", map[string]any{},
			WithKeys([]string{"extra1", "extra2"}),
		)
		message = GenerateDefaultMessage(rawIssue)
		assert.Equal(t, `Unrecognized keys: "extra1", "extra2"`, message)
	})

	t.Run("special numeric values handled correctly", func(t *testing.T) {
		// NaN
		rawIssue := NewRawIssue(core.InvalidType, math.NaN(),
			WithExpected("number"),
		)
		message := GenerateDefaultMessage(rawIssue)
		assert.Equal(t, "Invalid input: expected number, received NaN", message)

		// Infinity
		rawIssue = NewRawIssue(core.InvalidType, math.Inf(1),
			WithExpected("number"),
		)
		message = GenerateDefaultMessage(rawIssue)
		assert.Equal(t, "Invalid input: expected number, received Infinity", message)
	})
}

//////////////////////////////////////////
//////////   Edge Cases Tests          ///
//////////////////////////////////////////

func TestFormatterEdgeCases(t *testing.T) {
	t.Run("handles nil issues gracefully", func(t *testing.T) {
		err := NewZodError(nil)

		// All functions should handle nil issues without panicking
		assert.NotPanics(t, func() {
			_ = FormatError(err)
		})
		assert.NotPanics(t, func() {
			_ = TreeifyError(err)
		})
		assert.NotPanics(t, func() {
			_ = FlattenError(err)
		})
		assert.NotPanics(t, func() {
			_ = PrettifyError(err)
		})
	})

	t.Run("handles issues with nil paths", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.Custom,
					Message: "Error with nil path",
					Path:    nil,
				},
			},
		}

		err := NewZodError(issues)

		assert.NotPanics(t, func() {
			formatted := FormatError(err)
			assert.NotNil(t, formatted)
		})

		assert.NotPanics(t, func() {
			tree := TreeifyError(err)
			assert.NotNil(t, tree)
		})

		assert.NotPanics(t, func() {
			flattened := FlattenError(err)
			assert.NotNil(t, flattened)
		})
	})

	t.Run("handles issues with empty messages", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.Custom,
					Message: "",
					Path:    []any{"field"},
				},
			},
		}

		err := NewZodError(issues)
		prettified := PrettifyError(err)

		// Should still format properly with empty message
		assert.Contains(t, prettified, "field: Invalid input")
	})

	t.Run("handles very deep nesting", func(t *testing.T) {
		deepPath := []any{}
		for i := 0; i < 20; i++ {
			deepPath = append(deepPath, fmt.Sprintf("level%d", i))
		}

		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: "Deep nested error",
					Path:    deepPath,
				},
			},
		}

		err := NewZodError(issues)

		// Should handle deep nesting without issues
		assert.NotPanics(t, func() {
			_ = FormatError(err)
			_ = TreeifyError(err)
			_ = FlattenError(err)
			_ = PrettifyError(err)
		})
	})

	t.Run("handles large numbers of issues", func(t *testing.T) {
		issues := []ZodIssue{}
		for i := 0; i < 100; i++ {
			issues = append(issues, ZodIssue{
				ZodIssueBase: ZodIssueBase{
					Code:    core.InvalidType,
					Message: fmt.Sprintf("Error %d", i),
					Path:    []any{fmt.Sprintf("field%d", i)},
				},
			})
		}

		err := NewZodError(issues)

		// Should handle large numbers of issues efficiently
		assert.NotPanics(t, func() {
			formatted := FormatError(err)
			assert.NotNil(t, formatted)
		})

		prettified := PrettifyError(err)
		assert.Contains(t, prettified, "Error 0")
		assert.Contains(t, prettified, "Error 99")
	})

	t.Run("handles special characters in paths and messages", func(t *testing.T) {
		issues := []ZodIssue{
			{
				ZodIssueBase: ZodIssueBase{
					Code:    core.Custom,
					Message: "Error with unicode: 中文测试 🚀",
					Path:    []any{"field@special", "nested-field", "数据"},
				},
			},
		}

		err := NewZodError(issues)

		assert.NotPanics(t, func() {
			formatted := FormatError(err)
			tree := TreeifyError(err)
			flattened := FlattenError(err)
			prettified := PrettifyError(err)

			assert.NotNil(t, formatted)
			assert.NotNil(t, tree)
			assert.NotNil(t, flattened)
			assert.NotEmpty(t, prettified)
		})
	})
}
