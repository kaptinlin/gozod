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
			require.NoError(t, err, "Expected no error for input: %v", input)
			assert.Equal(t, input, result, "Expected input to be returned as-is")
		}
	})

	t.Run("mustParse success", func(t *testing.T) {
		schema := Any()
		result := schema.MustParse("test")
		assert.Equal(t, "test", result)
	})

	t.Run("mustParse panic on error", func(t *testing.T) {
		schema := Any().Refine(func(v any) bool {
			return false // Always fail
		})

		assert.Panics(t, func() {
			schema.MustParse("test")
		})
	})

	t.Run("basic validation with refinement", func(t *testing.T) {
		// Only accept strings
		schema := Any().Refine(func(v any) bool {
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

	t.Run("nil handling", func(t *testing.T) {
		schema := Any()

		// Nil should be accepted for Any type
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestAny_TypeSafety(t *testing.T) {
	t.Run("type preservation", func(t *testing.T) {
		schema := Any()

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
		schema := Any()

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

func TestAny_Modifiers(t *testing.T) {
	t.Run("Optional behavior", func(t *testing.T) {
		schema := Any()
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodAny[any, *any]
		_ = optionalSchema

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
		schema := Any()
		nilableSchema := schema.Nilable()

		_ = nilableSchema

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

	t.Run("Default preserves any type", func(t *testing.T) {
		schema := Any()
		defaultSchema := schema.Default("default_value")

		_ = defaultSchema

		// Valid input should override default
		result, err := defaultSchema.Parse("input_value")
		require.NoError(t, err)
		assert.Equal(t, "input_value", result)
	})
}

// =============================================================================
// Refinement tests
// =============================================================================

func TestAny_Refine(t *testing.T) {
	t.Run("basic refinement", func(t *testing.T) {
		// Only accept strings
		schema := Any().Refine(func(v any) bool {
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
		schema := Any().Nilable().Refine(func(v *any) bool {
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
}

// =============================================================================
// Factory function tests
// =============================================================================

func TestAny_Factories(t *testing.T) {
	t.Run("Any factory", func(t *testing.T) {
		schema := Any()
		_ = schema

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("AnyPtr factory", func(t *testing.T) {
		schema := AnyPtr()
		_ = schema

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

func TestAny_NonOptional(t *testing.T) {
	schema := Any().NonOptional()

	_, err := schema.Parse(123)
	require.NoError(t, err)

	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}
}
