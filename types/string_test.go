package types

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestStringBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := String()
		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		// Invalid type
		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := String()
		// String input returns string
		result1, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.IsType(t, "", result1)
		assert.Equal(t, "hello", result1)
		// Pointer input returns same pointer
		str := "world"
		result2, err := schema.Parse(&str)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result2)
		assert.Equal(t, &str, result2)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		schema := String().Min(2)
		input := "hello"
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify not only type and value, but exact pointer identity
		resultPtr, ok := result.(*string)
		require.True(t, ok, "Result should be *string")
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
		assert.Equal(t, "hello", *resultPtr)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := String().Nilable()
		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*string)(nil), result)
		// Valid input keeps type inference
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
		assert.IsType(t, "", result2)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := String().Min(3)
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse("hello")
		require.NoError(t, err2)
		assert.Equal(t, "hello", result2)

		// Test nilable schema rejects invalid values
		_, err3 := nilableSchema.Parse("hi")
		assert.Error(t, err3)

		// 🔥 Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse("hello")
		require.NoError(t, err5)
		assert.Equal(t, "hello", result5)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestStringCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := Coerce.String()
		tests := []struct {
			input    any
			expected string
		}{
			{123, "123"},
			{true, "true"},
			{false, "false"},
			{12.34, "12.34"},
			{"hello", "hello"},
		}
		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := Coerce.String().Min(3).Max(5)
		// Coercion then validation passes
		result, err := schema.Parse(1234)
		require.NoError(t, err)
		assert.Equal(t, "1234", result)
		// Coercion then validation fails
		_, err = schema.Parse(12)
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestStringValidations(t *testing.T) {
	t.Run("length validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  *ZodString
			input   string
			wantErr bool
		}{
			{"min length valid", String().Min(5), "hello", false},
			{"min length invalid", String().Min(5), "hi", true},
			{"max length valid", String().Max(5), "hello", false},
			{"max length invalid", String().Max(5), "hello world", true},
			{"exact length valid", String().Length(5), "hello", false},
			{"exact length invalid", String().Length(5), "hi", true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("pattern validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  *ZodString
			input   string
			wantErr bool
		}{
			{"starts with valid", String().StartsWith("hello"), "hello world", false},
			{"starts with invalid", String().StartsWith("hello"), "world hello", true},
			{"ends with valid", String().EndsWith("world"), "hello world", false},
			{"ends with invalid", String().EndsWith("world"), "world hello", true},
			{"includes valid", String().Includes("test"), "this is a test string", false},
			{"includes invalid", String().Includes("test"), "this is a string", true},
			{"regex valid", String().Regex(regexp.MustCompile(`^\d+$`)), "12345", false},
			{"regex invalid", String().Regex(regexp.MustCompile(`^\d+$`)), "abc123", true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("format validations", func(t *testing.T) {
		tests := []struct {
			name     string
			schema   *ZodString
			input    string
			expected bool
		}{
			{"valid email", String().Email(), "test@example.com", true},
			{"invalid email", String().Email(), "invalid-email", false},
			{"valid URL", String().URL(), "https://example.com", true},
			{"invalid URL", String().URL(), "not-a-url", false},
			{"valid UUID", String().UUID(), "f47ac10b-58cc-4372-a567-0e02b2c3d479", true},
			{"invalid UUID", String().UUID(), "not-a-uuid", false},
			{"valid datetime", String().DateTime(), "2023-04-15T10:30:00Z", true},
			{"invalid datetime", String().DateTime(), "not-a-datetime", false},
			{"valid date", String().Date(), "2023-04-15", true},
			{"invalid date", String().Date(), "not-a-date", false},
			{"valid time", String().Time(), "10:30:00", true},
			{"invalid time", String().Time(), "not-a-time", false},
			{"valid duration", String().Duration(), "P1Y2M3DT4H5M6S", true},
			{"invalid duration", String().Duration(), "not-a-duration", false},
			{"valid JSON", String().JSON(), `{"key": "value"}`, true},
			{"invalid JSON", String().JSON(), `{key: value}`, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				hasError := err != nil
				if hasError == tt.expected {
					t.Errorf("Expected success=%v, but got error=%v", tt.expected, hasError)
				}
			})
		}
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestStringModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schema := String().Optional()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := String().Nilable()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
	})

	t.Run("nullish modifier", func(t *testing.T) {
		schema := String().Nullish()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := String()
		// Valid input should not panic
		result := schema.MustParse("hello")
		assert.Equal(t, "hello", result)
		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestStringChaining(t *testing.T) {
	t.Run("multiple validations", func(t *testing.T) {
		schema := String().Min(5).Max(50).Email()
		// Valid input
		result, err := schema.Parse("user@example.com")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", result)
		// Validation failures
		testCases := []string{
			"a@b.c",         // too short
			"invalid-email", // not email format
		}
		for _, input := range testCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("validation with format", func(t *testing.T) {
		schema := String().Min(10).Email().StartsWith("user")
		// Valid input
		result, err := schema.Parse("user@example.com")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", result)
		// Validation failure
		_, err = schema.Parse("admin@example.com") // does not start with "user"
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestStringTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := String().Transform(func(val string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(val), nil
		})
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)
	})

	t.Run("transform chaining", func(t *testing.T) {
		schema := String().
			Transform(func(val string, ctx *core.RefinementContext) (any, error) {
				return strings.ToUpper(val), nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if str, ok := val.(string); ok {
					return "PREFIX_" + str, nil
				}
				return val, nil
			})
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "PREFIX_HELLO", result)
	})

	t.Run("pipe combination", func(t *testing.T) {
		schema := String().
			Transform(func(val string, ctx *core.RefinementContext) (any, error) {
				return strings.ToUpper(val), nil
			}).
			Pipe(String().Min(5))
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestStringRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := String().Refine(func(val string) bool {
			return strings.ContainsAny(val, "!@#$%^&*()")
		}, core.SchemaParams{
			Error: "String must contain at least one special character",
		})
		result, err := schema.Parse("password123!")
		require.NoError(t, err)
		assert.Equal(t, "password123!", result)
		_, err = schema.Parse("password123")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "special character")
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := String().Refine(func(val string) bool {
			return !strings.HasPrefix(val, "invalid")
		}, core.SchemaParams{
			Error: func(issue core.ZodRawIssue) string {
				if input, ok := issue.Input.(string); ok {
					return "The string '" + input + "' cannot start with 'invalid'"
				}
				return "Invalid string input"
			},
		})
		result, err := schema.Parse("validstring")
		require.NoError(t, err)
		assert.Equal(t, "validstring", result)
		_, err = schema.Parse("invalidstring")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "cannot start with 'invalid'")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := "hello"

		// Refine: only validates, never modifies
		refineSchema := String().Refine(func(s string) bool {
			return len(s) > 0
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, "hello", refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, "HELLO", transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input, transformResult, "Transform should return modified value")
	})

	t.Run("refine preserves pointer identity", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) >= 2
		})

		input := "hello"
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify exact pointer identity is preserved
		resultPtr, ok := result.(*string)
		require.True(t, ok)
		assert.True(t, resultPtr == inputPtr, "Refine should preserve exact pointer identity")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestStringErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := String().Min(5)
		_, err := schema.Parse("hi")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := String().Min(5, core.SchemaParams{
			Error: "core.Custom minimum length error",
		})
		_, err := schema.Parse("hi")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "core.Custom minimum length error")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := String().Min(5).Email()
		_, err := schema.Parse("hi")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestStringEdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		schema := String()
		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)
		// Empty string with min length validation
		_, err = schema.Min(1).Parse("")
		assert.Error(t, err)
	})

	t.Run("very long string", func(t *testing.T) {
		schema := String().Max(10)
		longString := strings.Repeat("a", 100)
		_, err := schema.Parse(longString)
		assert.Error(t, err)
	})

	t.Run("special characters", func(t *testing.T) {
		schema := String()
		specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
		result, err := schema.Parse(specialChars)
		require.NoError(t, err)
		assert.Equal(t, specialChars, result)
	})

	t.Run("unicode characters", func(t *testing.T) {
		schema := String()
		unicodeStr := "Hello 世界 🌍"
		result, err := schema.Parse(unicodeStr)
		require.NoError(t, err)
		assert.Equal(t, unicodeStr, result)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := String()
		// By default nil is not allowed
		_, err := schema.Parse(nil)
		assert.Error(t, err)
		// Nilable allows nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("type mismatch", func(t *testing.T) {
		schema := String()
		invalidTypes := []any{
			123,
			true,
			false,
			12.34,
			[]string{"hello"},
			map[string]string{"key": "value"},
		}
		for _, invalidType := range invalidTypes {
			_, err := schema.Parse(invalidType)
			assert.Error(t, err, "Expected error for type %T", invalidType)
		}
	})

	t.Run("modifier combinations", func(t *testing.T) {
		schema := String().Min(3).Email().Nilable()
		// Valid input
		result, err := schema.Parse("user@example.com")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", result)
		// nil input
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		// Invalid input
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestStringDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := Default(String().Min(5), "default")
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default", result)
		// Valid input normal validation
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
		// Invalid input still fails
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := String().DefaultFunc(func() string {
			counter++
			return fmt.Sprintf("generated-%d", counter)
		}).Min(10)

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "generated-1", result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "generated-2", result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse("valid_string_long_enough")
		require.NoError(t, err3)
		assert.Equal(t, "valid_string_long_enough", result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := String().
			Default("hello world").
			Min(5).
			Transform(func(s string, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"original": s,
					"upper":    strings.ToUpper(s),
					"length":   len(s),
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse("test input")
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, "test input", result1Map["original"])
		assert.Equal(t, "TEST INPUT", result1Map["upper"])
		assert.Equal(t, 10, result1Map["length"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, "hello world", result2Map["original"])
		assert.Equal(t, "HELLO WORLD", result2Map["upper"])
		assert.Equal(t, 11, result2Map["length"])

		// Invalid input still fails validation
		_, err3 := schema.Parse("hi")
		assert.Error(t, err3, "Short string should fail Min(5) validation")
	})

	t.Run("prefault value", func(t *testing.T) {
		schema := String().Min(5).Prefault("fallback")
		// Any validation failure uses fallback value
		result1, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "fallback", result1)
		result2, err := schema.Parse("hi")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result2)
		// Valid input normal validation
		result3, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result3)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Default(
			String().Min(5).Prefault("fallback"),
			"default",
		)
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default", result)
		// Invalid input uses fallback value
		result2, err := schema.Parse("hi")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result2)
		// Valid input normal validation
		result3, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result3)
	})
}
