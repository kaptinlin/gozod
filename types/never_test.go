package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestNever_Basic(t *testing.T) {
	t.Run("rejects all values", func(t *testing.T) {
		schema := Never()

		inputs := []any{
			"string", 42, 3.14, true, false, nil,
			[]int{1, 2, 3},
			map[string]int{"key": 42},
		}

		for _, input := range inputs {
			_, err := schema.Parse(input)
			require.Error(t, err, "Never().Parse(%v) should return error", input)
			assert.Contains(t, err.Error(), "expected never, received")
		}
	})

	t.Run("rejects complex structures", func(t *testing.T) {
		schema := Never()

		input := map[string]any{
			"nested": map[string]any{
				"array": []any{1, "two", true},
			},
		}

		_, err := schema.Parse(input)
		assert.Error(t, err)
	})
}

// =============================================================================
// Parse method variants tests
// =============================================================================

func TestNever_ParseVariants(t *testing.T) {
	t.Run("MustParse panics", func(t *testing.T) {
		schema := Never()
		assert.Panics(t, func() { schema.MustParse("test") })
	})

	t.Run("ParseAny rejects input", func(t *testing.T) {
		schema := Never()
		_, err := schema.ParseAny("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("MustParseAny panics", func(t *testing.T) {
		schema := Never()
		assert.Panics(t, func() { schema.MustParseAny("test") })
	})

	t.Run("StrictParse with nil rejects", func(t *testing.T) {
		schema := Never()
		_, err := schema.StrictParse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("MustStrictParse with nil panics", func(t *testing.T) {
		schema := Never()
		assert.Panics(t, func() { schema.MustStrictParse(nil) })
	})
}

// =============================================================================
// Internals and state tests
// =============================================================================

func TestNever_Internals(t *testing.T) {
	t.Run("Internals returns non-nil", func(t *testing.T) {
		schema := Never()
		internals := schema.Internals()
		require.NotNil(t, internals)
		assert.Equal(t, core.ZodTypeNever, internals.Type)
	})

	t.Run("IsOptional defaults to false", func(t *testing.T) {
		schema := Never()
		assert.False(t, schema.IsOptional())
	})

	t.Run("IsNilable defaults to false", func(t *testing.T) {
		schema := Never()
		assert.False(t, schema.IsNilable())
	})

	t.Run("IsOptional after Optional", func(t *testing.T) {
		schema := Never().Optional()
		assert.True(t, schema.IsOptional())
	})

	t.Run("IsNilable after Nilable", func(t *testing.T) {
		schema := Never().Nilable()
		assert.True(t, schema.IsNilable())
	})
}

// =============================================================================
// Modifier tests
// =============================================================================

func TestNever_Modifiers(t *testing.T) {
	t.Run("Optional allows nil", func(t *testing.T) {
		schema := Never().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse("hello")
		assert.Error(t, err)
	})

	t.Run("Nilable allows nil", func(t *testing.T) {
		schema := Never().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("Nullish allows nil", func(t *testing.T) {
		schema := Never().Nullish()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestNever_DefaultAndPrefault(t *testing.T) {
	t.Run("Default short-circuits validation", func(t *testing.T) {
		schema := Never().Default("bypass_never")

		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "bypass_never", result)
	})

	t.Run("Default with Optional", func(t *testing.T) {
		schema := Never().Optional().Default("default_value")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default_value", *result)

		_, err = schema.Parse("input_value")
		assert.Error(t, err)
	})

	t.Run("DefaultFunc with Optional", func(t *testing.T) {
		called := 0
		schema := Never().Optional().DefaultFunc(func() any {
			called++
			return "func_default"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "func_default", *result)
		assert.Equal(t, 1, called)
	})

	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := Never().Default("default_value").Prefault("prefault_value")

		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
	})

	t.Run("DefaultFunc priority over PrefaultFunc", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Never().DefaultFunc(func() any {
			defaultCalled = true
			return "default_func_value"
		}).PrefaultFunc(func() any {
			prefaultCalled = true
			return "prefault_func_value"
		})

		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_func_value", result)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled)
	})

	t.Run("Prefault replaces nil then gets rejected", func(t *testing.T) {
		schema := Never().Prefault("fallback")

		// Non-nil inputs rejected directly.
		nonNilInputs := []any{"string", 42, 3.14, true, false, []int{1, 2, 3}}
		for _, input := range nonNilInputs {
			_, err := schema.Parse(input)
			require.Error(t, err, "Never().Prefault().Parse(%v) should error", input)
			assert.Contains(t, err.Error(), "expected never, received")
		}

		// Nil input replaced by prefault, then rejected.
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("PrefaultFunc only called for nil", func(t *testing.T) {
		called := 0
		schema := Never().PrefaultFunc(func() any {
			called++
			return "dynamic_fallback"
		})

		_, err := schema.Parse("test")
		require.Error(t, err)
		assert.Equal(t, 0, called, "PrefaultFunc should not be called for non-nil input")

		_, err = schema.Parse(nil)
		require.Error(t, err)
		assert.Equal(t, 1, called, "PrefaultFunc should be called for nil input")
	})

	t.Run("Prefault with Optional rejects non-nil", func(t *testing.T) {
		schema := Never().Optional().Prefault("prefault_value")

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		// Nil triggers Prefault which provides invalid value for Never.
		result, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received string")
		assert.Nil(t, result)
	})

	t.Run("Prefault with complex types", func(t *testing.T) {
		fallback := map[string]any{"status": "fallback", "data": []int{1, 2, 3}}
		schema := Never().Prefault(fallback)

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("multiple modifiers combination", func(t *testing.T) {
		schema := Never().
			Optional().
			Default("default").
			Prefault("prefault")

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default", *result)
	})
}

// =============================================================================
// Factory function tests
// =============================================================================

func TestNever_Factories(t *testing.T) {
	t.Run("Never factory", func(t *testing.T) {
		schema := Never()
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverPtr factory", func(t *testing.T) {
		schema := NeverPtr()
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with string", func(t *testing.T) {
		schema := NeverTyped[string, string]()
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with any", func(t *testing.T) {
		schema := NeverTyped[any, any]()
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("NeverTyped with pointer constraint", func(t *testing.T) {
		schema := NeverTyped[any, *any]()
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestNever_Errors(t *testing.T) {
	t.Run("returns ZodError type", func(t *testing.T) {
		schema := Never()
		_, err := schema.Parse("not never")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("custom error message via params", func(t *testing.T) {
		schema := Never(core.SchemaParams{
			Error: "Custom never error",
		})

		_, err := schema.Parse("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Custom never error")
	})

	t.Run("NeverPtr with custom error", func(t *testing.T) {
		schema := NeverPtr(core.SchemaParams{
			Error: "Custom never ptr error",
		})

		_, err := schema.Parse("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Custom never ptr error")
	})

	t.Run("string shorthand error", func(t *testing.T) {
		schema := Never("short error message")

		_, err := schema.Parse("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "short error message")
	})
}

// =============================================================================
// Metadata tests
// =============================================================================

func TestNever_Metadata(t *testing.T) {
	t.Run("Describe sets description", func(t *testing.T) {
		schema := Never().Describe("a never type")

		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "a never type", meta.Description)
	})

	t.Run("Meta stores metadata", func(t *testing.T) {
		schema := Never().Meta(core.GlobalMeta{
			Description: "never schema",
		})

		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "never schema", meta.Description)
	})
}

// =============================================================================
// Transformation and pipeline tests
// =============================================================================

func TestNever_Transform(t *testing.T) {
	t.Run("transform with prefault still rejects", func(t *testing.T) {
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
	t.Run("pipe with prefault still rejects", func(t *testing.T) {
		anySchema := Any()
		schema := Never().Prefault("valid_string").Pipe(anySchema)

		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})
}

// =============================================================================
// Type conversion tests
// =============================================================================

func TestNever_TypeConversion(t *testing.T) {
	t.Run("typed never with string", func(t *testing.T) {
		schema := NeverTyped[string, string]().Prefault("test")
		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("typed never with int", func(t *testing.T) {
		schema := NeverTyped[int, int]().Prefault(42)
		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})

	t.Run("typed never with pointer constraint", func(t *testing.T) {
		schema := NeverTyped[string, *string]().Prefault("test")
		_, err := schema.Parse("anything")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected never, received")
	})
}

// =============================================================================
// Clone and unwrap tests
// =============================================================================

func TestNever_CloneAndUnwrap(t *testing.T) {
	t.Run("CloneFrom copies behavior", func(t *testing.T) {
		original := Never().Prefault("original")
		clone := Never()
		clone.CloneFrom(original)

		got1, err1 := original.Parse("test")
		got2, err2 := clone.Parse("test")

		assert.Equal(t, err1, err2)
		assert.Equal(t, got1, got2)
	})

	t.Run("Unwrap returns self", func(t *testing.T) {
		schema := Never()
		assert.Same(t, schema, schema.Unwrap())
	})
}
