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

func TestAny_BasicFunctionality(t *testing.T) {
	t.Run("accepts any value", func(t *testing.T) {
		schema := Any()

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
			require.NoError(t, err, "input: %v", input)
			assert.Equal(t, input, result)
		}
	})

	t.Run("MustParse success", func(t *testing.T) {
		schema := Any()
		result := schema.MustParse("test")
		assert.Equal(t, "test", result)
	})

	t.Run("MustParse panics on error", func(t *testing.T) {
		schema := Any().Refine(func(v any) bool {
			return false
		})
		assert.Panics(t, func() {
			schema.MustParse("test")
		})
	})

	t.Run("basic validation with refinement", func(t *testing.T) {
		schema := Any().Refine(func(v any) bool {
			_, ok := v.(string)
			return ok
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("nil handling", func(t *testing.T) {
		schema := Any()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// StrictParse tests
// =============================================================================

func TestAny_StrictParse(t *testing.T) {
	t.Run("accepts value", func(t *testing.T) {
		schema := Any()
		result, err := schema.StrictParse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("MustStrictParse success", func(t *testing.T) {
		schema := Any()
		result := schema.MustStrictParse(42)
		assert.Equal(t, 42, result)
	})

	t.Run("MustStrictParse panics on error", func(t *testing.T) {
		schema := Any().Refine(func(v any) bool {
			return false
		})
		assert.Panics(t, func() {
			schema.MustStrictParse("test")
		})
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestAny_TypeSafety(t *testing.T) {
	t.Run("type preservation", func(t *testing.T) {
		schema := Any()
		inputs := []any{"string", 42, 3.14, true, []int{1, 2, 3}}

		for _, input := range inputs {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}
	})

	t.Run("complex nested type preservation", func(t *testing.T) {
		schema := Any()
		complexInput := map[string]any{
			"string": "value",
			"number": 42,
			"array":  []any{1, "two", true},
			"nested": map[string]any{"inner": "value"},
		}

		result, err := schema.Parse(complexInput)
		require.NoError(t, err)
		assert.Equal(t, complexInput, result)
	})
}

// =============================================================================
// Modifier tests
// =============================================================================

func TestAny_Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		optionalSchema := Any().Optional()

		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable", func(t *testing.T) {
		nilableSchema := Any().Nilable()

		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = nilableSchema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("Nullish", func(t *testing.T) {
		schema := Any().Nullish()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("value")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "value", *result)
	})

	t.Run("Default", func(t *testing.T) {
		schema := Any().Default("default_value")

		result, err := schema.Parse("input_value")
		require.NoError(t, err)
		assert.Equal(t, "input_value", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
	})

	t.Run("DefaultFunc", func(t *testing.T) {
		called := false
		schema := Any().DefaultFunc(func() any {
			called = true
			return "dynamic_default"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "dynamic_default", result)
		assert.True(t, called)
	})

	t.Run("Prefault", func(t *testing.T) {
		schema := Any().Prefault("fallback")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)
	})

	t.Run("PrefaultFunc", func(t *testing.T) {
		schema := Any().PrefaultFunc(func() any {
			return "dynamic_fallback"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "dynamic_fallback", result)
	})
}

// =============================================================================
// Refinement tests
// =============================================================================

func TestAny_Refine(t *testing.T) {
	t.Run("basic refinement", func(t *testing.T) {
		schema := Any().Refine(func(v any) bool {
			_, ok := v.(string)
			return ok
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("refine with pointer constraints", func(t *testing.T) {
		schema := Any().Nilable().Refine(func(v *any) bool {
			if v == nil {
				return true
			}
			_, ok := (*v).(string)
			return ok
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("RefineAny", func(t *testing.T) {
		schema := Any().RefineAny(func(v any) bool {
			_, ok := v.(int)
			return ok
		})

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = schema.Parse("not int")
		assert.Error(t, err)
	})
}

// =============================================================================
// NonOptional tests
// =============================================================================

func TestAny_NonOptional(t *testing.T) {
	t.Run("rejects nil", func(t *testing.T) {
		schema := Any().NonOptional()

		_, err := schema.Parse(123)
		require.NoError(t, err)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
		var zErr *issues.ZodError
		if issues.IsZodError(err, &zErr) {
			assert.Equal(t, core.ZodTypeNonOptional,
				zErr.Issues[0].Expected)
		}
	})
}

// =============================================================================
// Factory tests
// =============================================================================

func TestAny_Factories(t *testing.T) {
	t.Run("Any", func(t *testing.T) {
		schema := Any()
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("AnyPtr", func(t *testing.T) {
		schema := AnyPtr()
		result, err := schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("Any with params", func(t *testing.T) {
		schema := Any(core.SchemaParams{
			Error: "Custom error",
		})
		require.NotNil(t, schema)
	})
}

// =============================================================================
// Composition tests
// =============================================================================

func TestAny_Composition(t *testing.T) {
	t.Run("And", func(t *testing.T) {
		schema := Any().And(Any())
		require.NotNil(t, schema)
	})

	t.Run("Or", func(t *testing.T) {
		schema := Any().Or(String())
		require.NotNil(t, schema)
	})
}

// =============================================================================
// Copy-on-Write immutability tests
// =============================================================================

func TestAny_CopyOnWrite(t *testing.T) {
	t.Run("modifiers do not mutate original", func(t *testing.T) {
		original := Any()
		_ = original.Optional()
		_ = original.Default("x")

		assert.False(t, original.IsOptional())
	})
}
