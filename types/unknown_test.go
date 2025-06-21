package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestUnknownBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := Unknown()

		// Unknown accepts all types (similar to Any but with different semantics)
		testCases := []any{
			"string", 42, 3.14, true, false,
			[]int{1, 2, 3}, map[string]int{"a": 1},
		}

		for _, input := range testCases {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}

		// Unknown rejects nil by default
		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Unknown()

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
		resultPtr, ok := result2.(*string)
		require.True(t, ok)
		assert.True(t, resultPtr == &str, "Should return the exact same pointer")
		assert.Equal(t, "world", *resultPtr)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Unknown().Nilable()

		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*any)(nil), result)

		// Valid input keeps type inference
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
		assert.IsType(t, "", result2)
	})

	t.Run("constructors", func(t *testing.T) {
		schema1 := Unknown()
		require.NotNil(t, schema1)
		assert.Equal(t, "unknown", schema1.GetInternals().Type)

		schema2 := Unknown()
		require.NotNil(t, schema2)
		assert.Equal(t, "unknown", schema2.GetInternals().Type)
	})

	t.Run("MustParse success", func(t *testing.T) {
		schema := Unknown()
		result := schema.MustParse("test")
		assert.Equal(t, "test", result)
	})

	t.Run("MustParse panic", func(t *testing.T) {
		schema := Unknown()
		assert.Panics(t, func() {
			schema.MustParse(nil) // Should panic because nil is not allowed by default
		})
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestUnknownCoercion(t *testing.T) {
	t.Run("coerce flag", func(t *testing.T) {
		// Unknown with coerce flag (though Unknown doesn't need coercion)
		schema := Unknown(core.SchemaParams{Coerce: true})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// Verify coerce flag is stored
		assert.True(t, schema.GetZod().Bag["coerce"].(bool))
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestUnknownValidations(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		schema := Unknown().Refine(func(val any) bool {
			if str, ok := val.(string); ok {
				return len(str) >= 3
			}
			return true // Allow non-strings
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)

		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("refineAny validation", func(t *testing.T) {
		schema := Unknown().RefineAny(func(val any) bool {
			return val != nil
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// This should fail because we're testing non-nilable Unknown with nil
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestUnknownModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Unknown().Optional()

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Unknown().Nilable()

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		schema := Unknown().Nullish()

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Unknown()
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse("hello")
		require.NoError(t, err2)
		assert.Equal(t, "hello", result2)

		// Critical: Original schema should remain unchanged
		_, err3 := baseSchema.Parse(nil)
		assert.Error(t, err3, "Original schema should still reject nil")

		result4, err4 := baseSchema.Parse("hello")
		require.NoError(t, err4)
		assert.Equal(t, "hello", result4)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestUnknownChaining(t *testing.T) {
	t.Run("refine chaining", func(t *testing.T) {
		schema := Unknown().
			Refine(func(val any) bool {
				return val != nil
			}).
			RefineAny(func(val any) bool {
				if str, ok := val.(string); ok {
					return len(str) > 0
				}
				return true
			})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("")
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestUnknownTransformPipe(t *testing.T) {
	t.Run("transform values", func(t *testing.T) {
		schema := Unknown().Transform(func(val any, ctx *core.RefinementContext) (any, error) {
			if str, ok := val.(string); ok {
				return "transformed_" + str, nil
			}
			return val, nil
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "transformed_test", result)

		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("transformAny values", func(t *testing.T) {
		schema := Unknown().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			return map[string]any{"value": val, "type": "unknown"}, nil
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		expected := map[string]any{"value": "test", "type": "unknown"}
		assert.Equal(t, expected, result)
	})

	t.Run("pipe to string", func(t *testing.T) {
		schema := Unknown().Pipe(String())

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestUnknownRefine(t *testing.T) {
	t.Run("refine with custom message", func(t *testing.T) {
		schema := Unknown().RefineAny(func(val any) bool {
			return val != "forbidden"
		}, core.SchemaParams{Error: "Value is forbidden"})

		result, err := schema.Parse("allowed")
		require.NoError(t, err)
		assert.Equal(t, "allowed", result)

		_, err = schema.Parse("forbidden")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Value is forbidden")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := "hello"

		// Refine: only validates, never modifies
		refineSchema := Unknown().Refine(func(val any) bool {
			return val != nil
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Unknown().Transform(func(val any, ctx *core.RefinementContext) (any, error) {
			if str, ok := val.(string); ok {
				return map[string]any{"original": str}, nil
			}
			return val, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, "hello", refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		expected := map[string]any{"original": "hello"}
		assert.Equal(t, expected, transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input, transformResult, "Transform should return modified value")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestUnknownErrorHandling(t *testing.T) {
	t.Run("custom error message", func(t *testing.T) {
		schema := Unknown(core.SchemaParams{Error: "core.Custom unknown message"})

		_, err := schema.Parse(nil) // Should fail because nil not allowed by default
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("error function", func(t *testing.T) {
		schema := Unknown(core.SchemaParams{
			Error: func(issue core.ZodRawIssue) string {
				return "Function-based error"
			},
		})

		_, err := schema.Parse(nil)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("parse context error", func(t *testing.T) {
		schema := Unknown()
		ctx := &core.ParseContext{
			Error: func(issue core.ZodRawIssue) string {
				return "Context error"
			},
		}

		_, err := schema.Parse(nil, ctx)
		assert.Error(t, err)
	})
}

// =============================================================================
// 9. Edge cases and internals
// =============================================================================

func TestUnknownEdgeCases(t *testing.T) {
	t.Run("internals access", func(t *testing.T) {
		schema := Unknown()
		internals := schema.GetInternals()

		assert.Equal(t, core.ZodTypeUnknown, internals.Type)
		assert.Equal(t, core.Version, internals.Version)
	})

	t.Run("clone functionality", func(t *testing.T) {
		original := Unknown()
		original.GetZod().Bag["custom"] = "value"

		cloned := &ZodUnknown{internals: &ZodUnknownInternals{
			ZodTypeInternals: core.ZodTypeInternals{},
			Bag:              make(map[string]any),
		}}
		cloned.CloneFrom(original)

		assert.Equal(t, "value", cloned.GetZod().Bag["custom"])
	})

	t.Run("parameters storage", func(t *testing.T) {
		params := core.SchemaParams{
			Description: "Test description",
			Params: map[string]any{
				"custom": "value",
			},
		}

		schema := Unknown(params)
		assert.Equal(t, "Test description", schema.GetZod().Bag["description"])
		assert.Equal(t, "value", schema.GetZod().Bag["custom"])
	})

	t.Run("unwrap returns self", func(t *testing.T) {
		schema := Unknown()
		unwrapped := schema.Unwrap()
		assert.Equal(t, schema, unwrapped)
	})

	t.Run("unknown vs any semantic difference", func(t *testing.T) {
		// Both Unknown and Any accept the same values, but semantically different
		unknownSchema := Unknown()
		anySchema := Any()

		testValue := "test"

		unknownResult, unknownErr := unknownSchema.Parse(testValue)
		anyResult, anyErr := anySchema.Parse(testValue)

		// Both should succeed with same result
		require.NoError(t, unknownErr)
		require.NoError(t, anyErr)
		assert.Equal(t, testValue, unknownResult)
		assert.Equal(t, testValue, anyResult)

		// But they have different type identifiers
		assert.Equal(t, "unknown", unknownSchema.GetInternals().Type)
		assert.Equal(t, "any", anySchema.GetInternals().Type)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestUnknownDefaultAndPrefault(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		defaultVal := "default_value"
		schema := Default(Unknown(), defaultVal)

		result, err := schema.Parse("actual")
		require.NoError(t, err)
		assert.Equal(t, "actual", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})

	t.Run("prefault values", func(t *testing.T) {
		// Unknown types don't have built-in Prefault method
		// This test demonstrates the concept but uses basic validation
		schema := Unknown()

		result, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", result)

		// Test that nil is rejected by default Unknown schema
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := Unknown().DefaultFunc(func() any {
			counter++
			return map[string]any{
				"unknown_generated": true,
				"iteration":         counter,
			}
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		expected1 := map[string]any{"unknown_generated": true, "iteration": 1}
		assert.Equal(t, expected1, result1)
		assert.Equal(t, 1, counter, "Function should be called once for nil input")

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		expected2 := map[string]any{"unknown_generated": true, "iteration": 2}
		assert.Equal(t, expected2, result2)
		assert.Equal(t, 2, counter, "Function should be called twice for second nil input")

		// Valid input should not call function
		result3, err3 := schema.Parse("unknown_value")
		require.NoError(t, err3)
		assert.Equal(t, "unknown_value", result3)
		assert.Equal(t, 2, counter, "Function should not be called for valid input")
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Unknown().PrefaultFunc(func() any {
			counter++
			return map[string]any{
				"unknown_fallback": true,
				"recovery_attempt": counter,
			}
		})

		// Create a refined schema by converting to the correct type first
		refinedSchema := schema.Refine(func(val any) bool {
			// Only allow strings with specific pattern
			if str, ok := val.(string); ok {
				return len(str) >= 4 && str != "fail"
			}
			return true // Allow non-strings
		})

		// Valid input should not call prefault function
		result1, err1 := refinedSchema.Parse("valid") // String with length >= 4
		require.NoError(t, err1)
		assert.Equal(t, "valid", result1)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input should call prefault function
		result2, err2 := refinedSchema.Parse("hi") // String with length < 4
		require.NoError(t, err2)
		expected2 := map[string]any{"unknown_fallback": true, "recovery_attempt": 1}
		assert.Equal(t, expected2, result2)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")

		// Another invalid input should call function again
		result3, err3 := refinedSchema.Parse("fail") // String with forbidden value
		require.NoError(t, err3)
		expected3 := map[string]any{"unknown_fallback": true, "recovery_attempt": 2}
		assert.Equal(t, expected3, result3)
		assert.Equal(t, 2, counter, "Function should increment counter for each invalid input")

		// Valid non-string input should not call function
		result4, err4 := refinedSchema.Parse(123)
		require.NoError(t, err4)
		assert.Equal(t, 123, result4)
		assert.Equal(t, 2, counter, "Counter should remain unchanged for valid non-string input")
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := "unknown_default_value"
		prefaultValue := "unknown_prefault_value"

		// Test default behavior separately
		defaultSchema := Unknown().Default(defaultValue)
		result1, err1 := defaultSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Test prefault behavior separately
		prefaultSchema := Unknown().Prefault(prefaultValue)

		// Create a refined schema by converting to the correct type first
		refinedPrefaultSchema := prefaultSchema.Refine(func(val any) bool {
			// Only allow strings with length > 6
			if str, ok := val.(string); ok {
				return len(str) > 6
			}
			return true // Allow non-strings
		})

		// Valid input succeeds
		result2, err2 := refinedPrefaultSchema.Parse("long_enough_string")
		require.NoError(t, err2)
		assert.Equal(t, "long_enough_string", result2)

		// Invalid input uses prefault
		result3, err3 := refinedPrefaultSchema.Parse("short")
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)

		// Valid non-string input succeeds
		result4, err4 := refinedPrefaultSchema.Parse(456)
		require.NoError(t, err4)
		assert.Equal(t, 456, result4)
	})
}
