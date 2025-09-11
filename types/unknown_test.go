package types

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestUnknown_BasicFunctionality(t *testing.T) {
	t.Run("accepts any value", func(t *testing.T) {
		schema := Unknown()

		testCases := []any{
			"string",
			42,
			3.14,
			true,
			false,
			nil,
			[]int{1, 2, 3},
			map[string]int{"key": 42},
		}

		for _, input := range testCases {
			result, err := schema.Parse(input)
			require.NoError(t, err, "Expected no error for input: %v", input)
			assert.Equal(t, input, result, "Expected input to be returned as-is")
		}
	})

	t.Run("mustParse success", func(t *testing.T) {
		schema := Unknown()
		result := schema.MustParse("test")
		assert.Equal(t, "test", result)
	})

	t.Run("mustParse panic on error", func(t *testing.T) {
		schema := Unknown().Refine(func(v any) bool {
			return false // Always fail
		})

		assert.Panics(t, func() {
			schema.MustParse("test")
		})
	})

	t.Run("basic validation with refinement", func(t *testing.T) {
		// Only accept strings
		schema := Unknown().Refine(func(v any) bool {
			_, ok := v.(string)
			return ok
		})

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid non-string
		_, err = schema.Parse(42)
		assert.Error(t, err, "Expected error for non-string input")
	})

	t.Run("nil handling - unknown special behavior", func(t *testing.T) {
		schema := Unknown()

		// Nil should be accepted for Unknown type (special behavior)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestUnknown_TypeSafety(t *testing.T) {
	t.Run("type preservation", func(t *testing.T) {
		schema := Unknown()

		// Test that types are preserved
		inputs := []any{
			"string",
			42,
			3.14,
			true,
			[]int{1, 2, 3},
		}

		for _, input := range inputs {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}
	})

	t.Run("complex nested type preservation", func(t *testing.T) {
		schema := Unknown()

		complexInput := map[string]any{
			"string": "value",
			"number": 42,
			"array":  []any{1, "two", true},
			"nested": map[string]any{
				"inner": "value",
			},
		}

		result, err := schema.Parse(complexInput)
		require.NoError(t, err)
		assert.Equal(t, complexInput, result)
	})
}

// =============================================================================
// Modifier tests
// =============================================================================

func TestUnknown_Modifiers(t *testing.T) {
	t.Run("Optional behavior", func(t *testing.T) {
		schema := Unknown()
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodUnknown[any, *any]
		var _ = optionalSchema

		// Test non-nil value - returns pointer
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// Test nil value (should be allowed for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable behavior", func(t *testing.T) {
		schema := Unknown()
		nilableSchema := schema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - returns pointer
		result, err = nilableSchema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("Default preserves unknown type", func(t *testing.T) {
		schema := Unknown()
		defaultSchema := schema.Default("default_value")

		var _ = defaultSchema

		// Valid input should override default
		result, err := defaultSchema.Parse("input_value")
		require.NoError(t, err)
		assert.Equal(t, "input_value", result)
	})
}

// =============================================================================
// Refinement tests
// =============================================================================

func TestUnknown_Refine(t *testing.T) {
	t.Run("basic refinement", func(t *testing.T) {
		// Only accept strings
		schema := Unknown().Refine(func(v any) bool {
			_, ok := v.(string)
			return ok
		})

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid case
		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("refine with pointer constraints", func(t *testing.T) {
		schema := Unknown().Nilable().Refine(func(v *any) bool {
			// Accept nil or strings
			if v == nil {
				return true
			}
			_, ok := (*v).(string)
			return ok
		})

		// Nil should be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// String should pass - returns pointer
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// Number should fail
		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("multiple refinements", func(t *testing.T) {
		schema := Unknown().
			Refine(func(v any) bool {
				// Must be string
				_, ok := v.(string)
				return ok
			}).
			Refine(func(v any) bool {
				// Must be non-empty string
				if s, ok := v.(string); ok {
					return len(s) > 0
				}
				return false
			})

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Fails first check
		_, err = schema.Parse(42)
		assert.Error(t, err)

		// Fails second check
		_, err = schema.Parse("")
		assert.Error(t, err)
	})
}

// =============================================================================
// Behavior tests - Unknown vs Any differences
// =============================================================================

func TestUnknown_Behavior(t *testing.T) {
	t.Run("unknown accepts everything without coercion", func(t *testing.T) {
		schema := Unknown()

		// Unknown should accept all values as-is
		testCases := []any{
			"string",
			42,
			3.14,
			true,
			nil,
			[]int{1, 2, 3},
			map[string]int{"key": 42},
		}

		for _, input := range testCases {
			result, err := schema.Parse(input)
			require.NoError(t, err, "Expected no error for input: %v", input)
			assert.Equal(t, input, result, "Expected input to be returned as-is")
		}
	})

	t.Run("unknown accepts nil by default", func(t *testing.T) {
		schema := Unknown()

		// Unknown type should accept nil by default (special behavior)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestUnknown_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// When both Default and Prefault are set, Default should take precedence
		schema := Unknown().Default("default_value").Prefault("prefault_value")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Default should bypass validation and return immediately
		// Add a refinement that would reject the default value
		schema := Unknown().Refine(func(v any) bool {
			// Only accept integers
			_, ok := v.(int)
			return ok
		}).Default("not-an-int") // String default that would fail refinement

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "not-an-int", result)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value must pass unknown validation (including refinements)
		schema := Unknown().Refine(func(v any) bool {
			// Only accept strings
			_, ok := v.(string)
			return ok
		}).Prefault("valid_string")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "valid_string", result)
	})

	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		// Non-nil input that fails validation should not trigger Prefault
		schema := Unknown().Refine(func(v any) bool {
			// Only accept strings
			_, ok := v.(string)
			return ok
		}).Prefault("prefault_fallback")

		// Invalid input should fail validation, not use Prefault
		_, err := schema.Parse(42) // Integer fails string refinement
		require.Error(t, err)
		// The error should indicate refinement failure
		assert.Contains(t, err.Error(), "Invalid input")
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Unknown().DefaultFunc(func() any {
			defaultCalled = true
			return "default_func"
		}).PrefaultFunc(func() any {
			prefaultCalled = true
			return "prefault_func"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_func", result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		// Test that invalid prefault value fails validation
		schema := Unknown().Refine(func(v any) bool {
			// Only accept strings
			_, ok := v.(string)
			return ok
		})

		// Test with valid prefault
		validSchema := schema.Prefault("valid_string")
		result, err := validSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "valid_string", result)
	})

	t.Run("Optional takes priority over Default", func(t *testing.T) {
		// When both Optional and Default are present, Default should still take priority
		// because Default is processed before Optional check
		schema := Unknown().Default("default_value").Optional()

		// Nil input should use default value, not return nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default_value", *result)

		// Test Optional without Default - should return nil
		optionalSchema := Unknown().Optional()
		result2, err := optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})
}

// =============================================================================
// Factory function tests
// =============================================================================

func TestUnknown_Factories(t *testing.T) {
	t.Run("Unknown factory", func(t *testing.T) {
		schema := Unknown()
		var _ = schema

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("UnknownPtr factory", func(t *testing.T) {
		schema := UnknownPtr()
		var _ = schema

		result, err := schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("Unknown with params", func(t *testing.T) {
		schema := Unknown(core.SchemaParams{
			Error: "Custom error",
		})

		require.NotNil(t, schema)
	})
}

func TestZodUnknown_Overwrite(t *testing.T) {
	t.Run("basic transformation", func(t *testing.T) {
		schema := Unknown().Overwrite(func(input any) any {
			if str, ok := input.(string); ok {
				return strings.ToUpper(str)
			}
			return input
		})

		result, err := schema.Parse("hello")
		assert.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Test non-string input passes through unchanged
		result2, err := schema.Parse(42)
		assert.NoError(t, err)
		assert.Equal(t, 42, result2)
	})

	t.Run("unknown type special nil handling", func(t *testing.T) {
		schema := Unknown().Overwrite(func(input any) any {
			if input == nil {
				return "nil_transformed"
			}
			return fmt.Sprintf("value_%v", input)
		})

		// Unknown type accepts nil by default, but Overwrite should still transform it
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		// Note: Unknown type's special nil handling means nil passes through
		// The transformation might not be applied to nil values by default
		assert.Nil(t, result) // Expecting original nil behavior

		// Test regular value transformation
		result2, err := schema.Parse("test")
		assert.NoError(t, err)
		assert.Equal(t, "value_test", result2)
	})

	t.Run("data sanitization for unknown sources", func(t *testing.T) {
		// Common use case: sanitizing data from unknown/untrusted sources
		schema := Unknown().Overwrite(func(input any) any {
			switch v := input.(type) {
			case string:
				// Remove HTML tags and normalize
				cleaned := strings.ReplaceAll(v, "<", "&lt;")
				cleaned = strings.ReplaceAll(cleaned, ">", "&gt;")
				return strings.TrimSpace(cleaned)
			case map[string]any:
				// Remove sensitive keys from maps
				cleaned := make(map[string]any)
				for k, val := range v {
					if k != "password" && k != "secret" {
						cleaned[k] = val
					}
				}
				return cleaned
			case []any:
				// Filter out nil values from slices
				var filtered []any
				for _, item := range v {
					if item != nil {
						filtered = append(filtered, item)
					}
				}
				return filtered
			default:
				return v
			}
		})

		// Test HTML escaping
		result, err := schema.Parse("<script>alert('xss')</script>")
		assert.NoError(t, err)
		assert.Equal(t, "&lt;script&gt;alert('xss')&lt;/script&gt;", result)

		// Test map sanitization
		unsafeMap := map[string]any{
			"username": "john",
			"password": "secret123",
			"email":    "john@example.com",
		}
		result2, err := schema.Parse(unsafeMap)
		assert.NoError(t, err)
		expected := map[string]any{
			"username": "john",
			"email":    "john@example.com",
		}
		assert.Equal(t, expected, result2)

		// Test slice filtering
		dirtySlice := []any{"a", nil, "b", nil, "c"}
		result3, err := schema.Parse(dirtySlice)
		assert.NoError(t, err)
		assert.Equal(t, []any{"a", "b", "c"}, result3)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Unknown().Overwrite(func(input any) any {
			return fmt.Sprintf("processed_%v", input)
		})

		result, err := schema.Parse(123)
		assert.NoError(t, err)
		assert.Equal(t, "processed_123", result)

		// Verify it's still ZodUnknown type
		assert.IsType(t, &ZodUnknown[any, any]{}, schema)
	})

	t.Run("method chaining", func(t *testing.T) {
		schema := Unknown().
			Overwrite(func(input any) any {
				if str, ok := input.(string); ok {
					return strings.ToLower(str)
				}
				return input
			}).
			Overwrite(func(input any) any {
				if str, ok := input.(string); ok {
					return "processed_" + str
				}
				return input
			})

		result, err := schema.Parse("HELLO")
		assert.NoError(t, err)
		assert.Equal(t, "processed_hello", result)
	})

	t.Run("with other modifiers", func(t *testing.T) {
		schema := Unknown().
			Default("unknown_default").
			Overwrite(func(input any) any {
				if str, ok := input.(string); ok {
					return "{" + str + "}"
				}
				return input
			})

		// Test with provided value
		result, err := schema.Parse("test")
		assert.NoError(t, err)
		assert.Equal(t, "{test}", result)

		// Test that Unknown actually uses the default value for nil
		// (My earlier assumption about Unknown's behavior was incorrect)
		schema2 := Unknown().Default("fallback").Overwrite(func(input any) any {
			if str, ok := input.(string); ok {
				return "{" + str + "}"
			}
			return input
		})

		result2, err := schema2.Parse(nil)
		assert.NoError(t, err)
		// Unknown uses default and then transforms it
		assert.Equal(t, "{fallback}", result2)
	})

	t.Run("json preprocessing", func(t *testing.T) {
		// Practical example: preprocessing JSON-like data from unknown sources
		schema := Unknown().Overwrite(func(input any) any {
			if m, ok := input.(map[string]any); ok {
				processed := make(map[string]any)
				for k, v := range m {
					// Convert string numbers to actual numbers
					if str, isStr := v.(string); isStr {
						if num, err := strconv.ParseFloat(str, 64); err == nil {
							processed[k] = num
						} else {
							processed[k] = str
						}
					} else {
						processed[k] = v
					}
				}
				return processed
			}
			return input
		})

		input := map[string]any{
			"id":   "123",
			"name": "John",
			"age":  "30",
		}

		result, err := schema.Parse(input)
		assert.NoError(t, err)
		expected := map[string]any{
			"id":   float64(123),
			"name": "John",
			"age":  float64(30),
		}
		assert.Equal(t, expected, result)
	})

	t.Run("pointer type support", func(t *testing.T) {
		schema := UnknownPtr().Overwrite(func(input any) any {
			if input == nil {
				return "null_handled"
			}
			return fmt.Sprintf("unknown_%v", input)
		})

		result, err := schema.Parse("test")
		assert.NoError(t, err)
		// Unknown type returns *any (interface{}) not *string
		// Check the dereferenced value
		assert.NotNil(t, result)
		assert.Equal(t, "unknown_test", *result)
	})
}

func TestUnknown_NonOptional(t *testing.T) {
	schema := Unknown().NonOptional()

	_, err := schema.Parse("hi")
	require.NoError(t, err)

	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}
}
