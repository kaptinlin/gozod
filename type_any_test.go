package gozod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestAnyBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := Any()

		// Any accepts all types
		testCases := []interface{}{
			"string", 42, 3.14, true, false,
			[]int{1, 2, 3}, map[string]int{"a": 1},
		}

		for _, input := range testCases {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}

		// Any rejects nil by default
		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Any()

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
		schema := Any().Nilable()

		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*interface{})(nil), result)

		// Valid input keeps type inference
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
		assert.IsType(t, "", result2)
	})

	t.Run("constructors", func(t *testing.T) {
		schema1 := Any()
		require.NotNil(t, schema1)
		assert.Equal(t, "any", schema1.GetInternals().Type)

		schema2 := NewZodAny()
		require.NotNil(t, schema2)
		assert.Equal(t, "any", schema2.GetInternals().Type)
	})

	t.Run("MustParse success", func(t *testing.T) {
		schema := Any()
		result := schema.MustParse("test")
		assert.Equal(t, "test", result)
	})

	t.Run("MustParse panic", func(t *testing.T) {
		schema := Any()
		assert.Panics(t, func() {
			schema.MustParse(nil) // Should panic because nil is not allowed by default
		})
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestAnyCoercion(t *testing.T) {
	t.Run("coerce flag", func(t *testing.T) {
		// Any with coerce flag (though Any doesn't need coercion)
		schema := Any(SchemaParams{Coerce: true})

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

func TestAnyValidations(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		schema := Any().Refine(func(val any) bool {
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
		schema := Any().RefineAny(func(val any) bool {
			return val != nil
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// This should fail because we're testing non-nilable Any with nil
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestAnyModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Any().Optional()

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Any().Nilable()

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		schema := Any().Nullish()

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Any()
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

func TestAnyChaining(t *testing.T) {
	t.Run("refine chaining", func(t *testing.T) {
		schema := Any().
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

func TestAnyTransformPipe(t *testing.T) {
	t.Run("transform values", func(t *testing.T) {
		schema := Any().Transform(func(val any, ctx *RefinementContext) (any, error) {
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
		schema := Any().TransformAny(func(val any, ctx *RefinementContext) (any, error) {
			return map[string]any{"value": val, "type": "any"}, nil
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		expected := map[string]any{"value": "test", "type": "any"}
		assert.Equal(t, expected, result)
	})

	t.Run("pipe to string", func(t *testing.T) {
		schema := Any().Pipe(String())

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

func TestAnyRefine(t *testing.T) {
	t.Run("refine with custom message", func(t *testing.T) {
		schema := Any().RefineAny(func(val any) bool {
			return val != "forbidden"
		}, SchemaParams{Error: "Value is forbidden"})

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
		refineSchema := Any().Refine(func(val any) bool {
			return val != nil
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Any().Transform(func(val any, ctx *RefinementContext) (any, error) {
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

func TestAnyErrorHandling(t *testing.T) {
	t.Run("custom error message", func(t *testing.T) {
		schema := Any(SchemaParams{Error: "Custom any message"})

		_, err := schema.Parse(nil) // Should fail because nil not allowed by default
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("error function", func(t *testing.T) {
		schema := Any(SchemaParams{
			Error: func(issue ZodRawIssue) string {
				return "Function-based error"
			},
		})

		_, err := schema.Parse(nil)
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("parse context error", func(t *testing.T) {
		schema := Any()
		ctx := &ParseContext{
			Error: func(issue ZodRawIssue) string {
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

func TestAnyEdgeCases(t *testing.T) {
	t.Run("internals access", func(t *testing.T) {
		schema := Any()
		internals := schema.GetInternals()

		assert.Equal(t, ZodTypeAny, internals.Type)
		assert.Equal(t, Version, internals.Version)
	})

	t.Run("clone functionality", func(t *testing.T) {
		original := Any()
		original.GetZod().Bag["custom"] = "value"

		cloned := &ZodAny{internals: &ZodAnyInternals{
			ZodTypeInternals: ZodTypeInternals{},
			Bag:              make(map[string]interface{}),
		}}
		cloned.CloneFrom(original)

		assert.Equal(t, "value", cloned.GetZod().Bag["custom"])
	})

	t.Run("parameters storage", func(t *testing.T) {
		params := SchemaParams{
			Description: "Test description",
			Params: map[string]interface{}{
				"custom": "value",
			},
		}

		schema := Any(params)
		assert.Equal(t, "Test description", schema.GetZod().Bag["description"])
		assert.Equal(t, "value", schema.GetZod().Bag["custom"])
	})

	t.Run("unwrap returns self", func(t *testing.T) {
		schema := Any()
		unwrapped := schema.Unwrap()
		assert.Equal(t, schema, unwrapped)
	})

	t.Run("any vs unknown semantic difference", func(t *testing.T) {
		// Both Any and Unknown accept the same values, but semantically different
		anySchema := Any()
		unknownSchema := Unknown()

		testValue := "test"

		anyResult, anyErr := anySchema.Parse(testValue)
		unknownResult, unknownErr := unknownSchema.Parse(testValue)

		// Both should succeed with same result
		require.NoError(t, anyErr)
		require.NoError(t, unknownErr)
		assert.Equal(t, testValue, anyResult)
		assert.Equal(t, testValue, unknownResult)

		// But they have different type identifiers
		assert.Equal(t, "any", anySchema.GetInternals().Type)
		assert.Equal(t, "unknown", unknownSchema.GetInternals().Type)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestAnyDefaultAndPrefault(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		defaultVal := "default_value"
		schema := Default(Any(), defaultVal)

		result, err := schema.Parse("actual")
		require.NoError(t, err)
		assert.Equal(t, "actual", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})

	t.Run("prefault values", func(t *testing.T) {
		// Any types don't have built-in Prefault method
		// This test demonstrates the concept but uses basic validation
		schema := Any()

		result, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", result)

		// Test that nil is rejected by default Any schema
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := Any().DefaultFunc(func() any {
			counter++
			return map[string]any{
				"generated": true,
				"count":     counter,
			}
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		expected1 := map[string]any{"generated": true, "count": 1}
		assert.Equal(t, expected1, result1)
		assert.Equal(t, 1, counter, "Function should be called once for nil input")

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		expected2 := map[string]any{"generated": true, "count": 2}
		assert.Equal(t, expected2, result2)
		assert.Equal(t, 2, counter, "Function should be called twice for second nil input")

		// Valid input should not call function
		result3, err3 := schema.Parse("actual_value")
		require.NoError(t, err3)
		assert.Equal(t, "actual_value", result3)
		assert.Equal(t, 2, counter, "Function should not be called for valid input")
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Any().PrefaultFunc(func() any {
			counter++
			return map[string]any{
				"fallback": true,
				"attempt":  counter,
			}
		}).Refine(func(val any) bool {
			// Only allow strings with length > 3
			if str, ok := val.(string); ok {
				return len(str) > 3
			}
			return true // Allow non-strings
		})

		// Valid input should not call prefault function
		result1, err1 := schema.Parse("hello") // String with length > 3
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input should call prefault function
		result2, err2 := schema.Parse("hi") // String with length <= 3
		require.NoError(t, err2)
		expected2 := map[string]any{"fallback": true, "attempt": 1}
		assert.Equal(t, expected2, result2)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")

		// Another invalid input should call function again
		result3, err3 := schema.Parse("bye") // Another short string
		require.NoError(t, err3)
		expected3 := map[string]any{"fallback": true, "attempt": 2}
		assert.Equal(t, expected3, result3)
		assert.Equal(t, 2, counter, "Function should increment counter for each invalid input")

		// Valid non-string input should not call function
		result4, err4 := schema.Parse(42)
		require.NoError(t, err4)
		assert.Equal(t, 42, result4)
		assert.Equal(t, 2, counter, "Counter should remain unchanged for valid non-string input")
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := "default_any_value"
		prefaultValue := "prefault_any_value"

		// Test default behavior separately
		defaultSchema := Any().Default(defaultValue)
		result1, err1 := defaultSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Test prefault behavior separately
		prefaultSchema := Any().Prefault(prefaultValue).Refine(func(val any) bool {
			// Only allow strings with length > 5
			if str, ok := val.(string); ok {
				return len(str) > 5
			}
			return true // Allow non-strings
		})

		// Valid input succeeds
		result2, err2 := prefaultSchema.Parse("long_enough")
		require.NoError(t, err2)
		assert.Equal(t, "long_enough", result2)

		// Invalid input uses prefault
		result3, err3 := prefaultSchema.Parse("short")
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)

		// Valid non-string input succeeds
		result4, err4 := prefaultSchema.Parse(123)
		require.NoError(t, err4)
		assert.Equal(t, 123, result4)
	})
}
