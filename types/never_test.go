package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestNever_BasicFunctionality(t *testing.T) {
	t.Run("rejects all values", func(t *testing.T) {
		schema := Never()

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
			_, err := schema.Parse(input)
			require.Error(t, err, "Expected error for input: %v", input)

			// Verify error message
			assert.Contains(t, err.Error(), "Never type should not accept any value")
		}
	})

	t.Run("mustParse panic on error", func(t *testing.T) {
		schema := Never()

		assert.Panics(t, func() {
			schema.MustParse("test")
		})
	})

	t.Run("prefault provides fallback", func(t *testing.T) {
		schema := Never().Prefault("fallback")

		// Test various inputs - all should return fallback
		testCases := []any{
			"string", 42, 3.14, true, false, nil, []int{1, 2, 3},
		}

		for _, input := range testCases {
			result, err := schema.Parse(input)
			require.NoError(t, err, "Expected no error with prefault for input: %v", input)
			assert.Equal(t, "fallback", result)
		}
	})

	t.Run("prefaultFunc provides dynamic fallback", func(t *testing.T) {
		callCount := 0
		schema := Never().PrefaultFunc(func() any {
			callCount++
			return "dynamic_fallback"
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "dynamic_fallback", result)
		assert.Equal(t, 1, callCount)
	})
}

// =============================================================================
// Modifier tests
// =============================================================================

func TestNever_Modifiers(t *testing.T) {
	t.Run("Optional behavior", func(t *testing.T) {
		schema := Never()
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodNever[any, *any]
		var _ *ZodNever[any, *any] = optionalSchema

		// Test nil value (should be allowed for optional)
		result, err := optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test non-nil value should still fail (unless prefault is set)
		_, err = optionalSchema.Parse("hello")
		assert.Error(t, err)
	})

	t.Run("Nilable behavior", func(t *testing.T) {
		schema := Never()
		nilableSchema := schema.Nilable()

		var _ *ZodNever[any, *any] = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test non-nil value should still fail
		_, err = nilableSchema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("Nullish behavior", func(t *testing.T) {
		schema := Never()
		nullishSchema := schema.Nullish()

		var _ *ZodNever[any, *any] = nullishSchema

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test non-nil value should still fail
		_, err = nullishSchema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("Default with Optional", func(t *testing.T) {
		schema := Never().Optional().Default("default_value")

		// Test nil input should use default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default_value", *result)

		// Test non-nil value should still fail (unless prefault)
		_, err = schema.Parse("input_value")
		assert.Error(t, err)
	})

	t.Run("DefaultFunc with Optional", func(t *testing.T) {
		callCount := 0
		schema := Never().Optional().DefaultFunc(func() any {
			callCount++
			return "func_default"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "func_default", *result)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Prefault with Optional", func(t *testing.T) {
		schema := Never().Optional().Prefault("prefault_value")

		// Prefault should override Never's rejection for non-nil
		result, err := schema.Parse("anything")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "prefault_value", *result)

		// Nil should still work for Optional
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})
}

// =============================================================================
// Refinement tests
// =============================================================================

func TestNever_Refine(t *testing.T) {
	t.Run("basic refinement with prefault", func(t *testing.T) {
		// Never with prefault + refinement
		schema := Never().Prefault("fallback").Refine(func(v any) bool {
			// Only accept specific fallback value
			return v == "fallback"
		})

		// Should pass because prefault provides "fallback" which passes refinement
		result, err := schema.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)
	})

	t.Run("refinement fails on prefault", func(t *testing.T) {
		schema := Never().Prefault("fallback").Refine(func(v any) bool {
			// Reject the fallback value
			return v != "fallback"
		})

		// Should fail because prefault provides "fallback" but refinement rejects it
		_, err := schema.Parse("anything")
		assert.Error(t, err)
	})

	t.Run("refine with pointer constraints", func(t *testing.T) {
		schema := Never().Nilable().Prefault("test").Refine(func(v *any) bool {
			// Accept nil or specific values
			if v == nil {
				return true
			}
			return *v == "test"
		})

		// Nil should be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Non-nil should get prefault value and pass refinement
		result2, err := schema.Parse("anything")
		require.NoError(t, err)
		require.NotNil(t, result2)
		assert.Equal(t, "test", *result2)
	})

	t.Run("RefineAny with flexible validation", func(t *testing.T) {
		schema := Never().Prefault(42).RefineAny(func(v any) bool {
			// Only accept numbers
			_, ok := v.(int)
			return ok
		})

		result, err := schema.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}

// =============================================================================
// Factory function tests
// =============================================================================

func TestNever_Factories(t *testing.T) {
	t.Run("Never factory", func(t *testing.T) {
		schema := Never()
		var _ *ZodNever[any, any] = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverPtr factory", func(t *testing.T) {
		schema := NeverPtr()
		var _ *ZodNever[any, *any] = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped factory", func(t *testing.T) {
		schema := NeverTyped[string, string]()
		var _ *ZodNever[string, string] = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with explicit types", func(t *testing.T) {
		schema := NeverTyped[any, any]()
		var _ *ZodNever[any, any] = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with pointer constraint", func(t *testing.T) {
		schema := NeverTyped[any, *any]()
		var _ *ZodNever[any, *any] = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("Never with params", func(t *testing.T) {
		schema := Never(core.SchemaParams{
			Error: "Custom never error",
		})

		require.NotNil(t, schema)
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// Transformation and Pipeline tests
// =============================================================================

func TestNever_Transform(t *testing.T) {
	t.Run("transform with prefault", func(t *testing.T) {
		// Transform is rarely useful with Never, but can work with prefault
		schema := Never().Prefault("hello").Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return str + "_transformed", nil
			}
			return input, nil
		})

		result, err := schema.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, "hello_transformed", result)
	})
}

func TestNever_Pipe(t *testing.T) {
	t.Run("pipe with prefault", func(t *testing.T) {
		// Create a simple any schema for piping
		anySchema := Any()

		// Never with prefault piped to any validation
		schema := Never().Prefault("valid_string").Pipe(anySchema)

		result, err := schema.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, "valid_string", result)
	})
}

// =============================================================================
// Type conversion and helper tests
// =============================================================================

func TestNever_TypeConversion(t *testing.T) {
	t.Run("constraint type conversion", func(t *testing.T) {
		// Test with different constraint types
		schema1 := NeverTyped[string, string]().Prefault("test")
		result1, err := schema1.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, "test", result1)

		schema2 := NeverTyped[int, int]().Prefault(42)
		result2, err := schema2.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
	})

	t.Run("pointer constraint handling", func(t *testing.T) {
		schema := NeverTyped[string, *string]().Prefault("test")
		result, err := schema.Parse("anything")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})
}

// =============================================================================
// Clone and Unwrap tests
// =============================================================================

func TestNever_CloneAndUnwrap(t *testing.T) {
	t.Run("clone functionality", func(t *testing.T) {
		original := Never().Prefault("original")
		clone := Never()
		clone.CloneFrom(original)

		// Both should behave the same way
		result1, err1 := original.Parse("test")
		result2, err2 := clone.Parse("test")

		assert.Equal(t, err1, err2)
		assert.Equal(t, result1, result2)
	})

	t.Run("unwrap functionality", func(t *testing.T) {
		schema := Never()
		unwrapped := schema.Unwrap()
		assert.Same(t, schema, unwrapped)
	})
}

// =============================================================================
// Edge cases and error handling
// =============================================================================

func TestNever_EdgeCases(t *testing.T) {
	t.Run("multiple modifiers combination", func(t *testing.T) {
		schema := Never().
			Optional().
			Default("default").
			Prefault("prefault").
			Refine(func(v *any) bool {
				return v != nil && *v == "prefault"
			})

		// Non-nil input should use prefault
		result, err := schema.Parse("anything")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "prefault", *result)

		// Nil input should use default
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result2)
		assert.Equal(t, "default", *result2)
	})

	t.Run("complex nested structures", func(t *testing.T) {
		// Never should reject even complex structures
		schema := Never()

		complexInput := map[string]any{
			"nested": map[string]any{
				"array": []any{1, "two", true},
			},
		}

		_, err := schema.Parse(complexInput)
		assert.Error(t, err)
	})

	t.Run("prefault with complex types", func(t *testing.T) {
		complexFallback := map[string]any{
			"status": "fallback",
			"data":   []int{1, 2, 3},
		}

		schema := Never().Prefault(complexFallback)

		result, err := schema.Parse("anything")
		require.NoError(t, err)
		assert.Equal(t, complexFallback, result)
	})
}
