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
			assert.Contains(t, err.Error(), "expected never, received")
		}
	})

	t.Run("mustParse panic on error", func(t *testing.T) {
		schema := Never()

		assert.Panics(t, func() {
			schema.MustParse("test")
		})
	})

	t.Run("prefault only replaces nil then gets rejected", func(t *testing.T) {
		schema := Never().Prefault("fallback")

		// Non-nil inputs should be rejected directly by Never
		nonNilInputs := []any{
			"string", 42, 3.14, true, false, []int{1, 2, 3},
		}

		for _, input := range nonNilInputs {
			_, err := schema.Parse(input)
			require.Error(t, err, "Expected error for non-nil input: %v", input)
			assert.Contains(t, err.Error(), "expected never, received")
		}

		// Nil input should be replaced by prefault, then rejected by Never
		_, err := schema.Parse(nil)
		require.Error(t, err, "Expected error even with prefault for nil input")
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("prefaultFunc only replaces nil then gets rejected", func(t *testing.T) {
		callCount := 0
		schema := Never().PrefaultFunc(func() any {
			callCount++
			return "dynamic_fallback"
		})

		// Non-nil input should be rejected directly, prefaultFunc not called
		_, err := schema.Parse("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
		assert.Equal(t, 0, callCount, "PrefaultFunc should not be called for non-nil input")

		// Nil input should trigger prefaultFunc, then get rejected
		_, err = schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
		assert.Equal(t, 1, callCount, "PrefaultFunc should be called for nil input")
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
		var _ = optionalSchema

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

		var _ = nilableSchema

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

		var _ = nullishSchema

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

		// Never should still reject non-nil inputs even with Prefault
		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		// With new priority (Prefault > Optional), nil triggers Prefault which provides invalid value for Never
		result2, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received string")
		assert.Nil(t, result2)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestNever_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := Never().Default("default_value").Prefault("prefault_value")

		// When input is nil, Default should take precedence
		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
	})

	// Test 2: Default short-circuit mechanism
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		// Never type rejects all values, but Default should bypass this
		schema := Never().Default("bypass_never")

		// Default should bypass Never's rejection mechanism
		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "bypass_never", result)
	})

	// Test 3: Prefault requires full validation
	t.Run("Prefault requires full validation", func(t *testing.T) {
		// Never type with Prefault should replace nil then reject
		schema := Never().Prefault("fallback_value")

		// Non-nil input should be rejected directly
		_, err := schema.ParseAny("any_input")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		// Nil input should be replaced by prefault, then rejected
		_, err = schema.ParseAny(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		// For Never type, all inputs should be rejected even with Prefault
		schema := Never().Prefault("fallback_value")

		// Non-nil input should be rejected by Never
		_, err := schema.ParseAny("non_nil_input")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		// Nil input should also be rejected by Never
		_, err2 := schema.ParseAny(nil)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "expected never, received")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Never().DefaultFunc(func() any {
			defaultCalled = true
			return "default_func_value"
		}).PrefaultFunc(func() any {
			prefaultCalled = true
			return "prefault_func_value"
		})

		// DefaultFunc should be called and take precedence
		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_func_value", result)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // PrefaultFunc should not be called
	})
}

// =============================================================================
// Refinement tests
// =============================================================================

// TestNever_Refine test cases have been removed
// Reason: z.never() should reject all inputs, even if prefault is set, refine function will not be executed
// Because prefault values will be rejected by never, these tests are based on incorrect assumptions

// =============================================================================
// Factory function tests
// =============================================================================

func TestNever_Factories(t *testing.T) {
	t.Run("Never factory", func(t *testing.T) {
		schema := Never()
		var _ = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverPtr factory", func(t *testing.T) {
		schema := NeverPtr()
		var _ = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped factory", func(t *testing.T) {
		schema := NeverTyped[string, string]()
		var _ = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with explicit types", func(t *testing.T) {
		schema := NeverTyped[any, any]()
		var _ = schema

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with pointer constraint", func(t *testing.T) {
		schema := NeverTyped[any, *any]()
		var _ = schema

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
		// Transform with Never should still reject all inputs
		schema := Never().Prefault("hello").Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return str + "_transformed", nil
			}
			return input, nil
		})

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})
}

func TestNever_Pipe(t *testing.T) {
	t.Run("pipe with prefault", func(t *testing.T) {
		// Create a simple any schema for piping
		anySchema := Any()

		// Never with prefault should still reject all inputs
		schema := Never().Prefault("valid_string").Pipe(anySchema)

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})
}

// =============================================================================
// Type conversion and helper tests
// =============================================================================

func TestNever_TypeConversion(t *testing.T) {
	t.Run("constraint type conversion", func(t *testing.T) {
		// Never type always rejects all inputs, even if prefault is set
		schema1 := NeverTyped[string, string]().Prefault("test")
		_, err := schema1.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		schema2 := NeverTyped[int, int]().Prefault(42)
		_, err = schema2.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("pointer constraint handling", func(t *testing.T) {
		schema := NeverTyped[string, *string]().Prefault("test")
		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
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
		// Test combination of multiple modifiers, but not including Refine (because Never type should not execute refine functions)
		schema := Never().
			Optional().
			Default("default").
			Prefault("prefault")

		// Non-nil input should be rejected by Never (Never type always rejects non-nil inputs)
		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		// Nil input should use default (Default has higher priority than Prefault)
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

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})
}
