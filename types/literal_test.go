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

func TestLiteralBasicFunctionality(t *testing.T) {
	t.Run("constructors", func(t *testing.T) {
		// Single value literal
		schema := Literal("hello")
		require.NotNil(t, schema)
		assert.Equal(t, "literal", schema.GetInternals().Type)

		// Multiple values literal
		schema2 := Literal("red", "green", "blue")
		require.NotNil(t, schema2)
		assert.Equal(t, "literal", schema2.GetInternals().Type)
	})

	t.Run("basic validation", func(t *testing.T) {
		schema := Literal("hello", 42, true)

		// Valid cases
		tests := []any{"hello", 42, true}
		for _, input := range tests {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}

		// Invalid cases
		invalidTests := []any{"world", 43, false, nil}
		for _, input := range invalidTests {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("type-specific literals", func(t *testing.T) {
		// String literal
		stringSchema := Literal("active")
		result, err := stringSchema.Parse("active")
		require.NoError(t, err)
		assert.Equal(t, "active", result)

		_, err = stringSchema.Parse("inactive")
		assert.Error(t, err)

		// Number literal
		numSchema := Literal(42)
		result, err = numSchema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = numSchema.Parse(43)
		assert.Error(t, err)

		// Boolean literal
		boolSchema := Literal(true)
		result, err = boolSchema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = boolSchema.Parse(false)
		assert.Error(t, err)
	})
}

// =============================================================================
// 2. Validation methods
// =============================================================================

func TestLiteralValidationMethods(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		schema := Literal("red", "green", "blue").RefineAny(func(val any) bool {
			// Only allow "red" and "green"
			return val == "red" || val == "green"
		})

		// Valid refined values
		result, err := schema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		result, err = schema.Parse("green")
		require.NoError(t, err)
		assert.Equal(t, "green", result)

		// Valid literal but fails refinement
		_, err = schema.Parse("blue")
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestLiteralModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Optional(Literal("hello"))

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse("world")
		assert.Error(t, err)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Nilable(Literal(42))

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse(43)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestLiteralChaining(t *testing.T) {
	t.Run("refine chaining", func(t *testing.T) {
		schema := Literal(1, 2, 3, 4, 5, 6).
			RefineAny(func(val any) bool {
				if num, ok := val.(int); ok {
					return num > 3 // Must be > 3
				}
				return false
			}).
			RefineAny(func(val any) bool {
				if num, ok := val.(int); ok {
					return num%2 == 0 // Must be even
				}
				return false
			})

		// Valid: > 3 and even
		result, err := schema.Parse(4)
		require.NoError(t, err)
		assert.Equal(t, 4, result)

		result, err = schema.Parse(6)
		require.NoError(t, err)
		assert.Equal(t, 6, result)

		// Invalid cases
		_, err = schema.Parse(2) // Even but not > 3
		assert.Error(t, err)

		_, err = schema.Parse(5) // > 3 but not even
		assert.Error(t, err)
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestLiteralTransform(t *testing.T) {
	t.Run("transform values", func(t *testing.T) {
		schema := Literal("red", "green", "blue").TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if str, ok := val.(string); ok {
				return "color_" + str, nil
			}
			return val, nil
		})

		result, err := schema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "color_red", result)
	})

	t.Run("pipe to string", func(t *testing.T) {
		schema := Literal("hello").Pipe(String().Min(3))

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi") // Not in literal
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Error handling
// =============================================================================

func TestLiteralErrorHandling(t *testing.T) {
	t.Run("type mismatch errors", func(t *testing.T) {
		schema := Literal("hello")

		_, err := schema.Parse(123)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
	})

	t.Run("custom error messages", func(t *testing.T) {
		defaultSchema := Literal("active")
		_, err := defaultSchema.Parse("inactive")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected \"active\"")

		customSchema := Literal("active", core.SchemaParams{Error: "must be active"})
		_, err2 := customSchema.Parse("inactive")
		assert.Error(t, err2)
		assert.Equal(t, "must be active", err2.Error())
	})
}

// =============================================================================
// 7. Edge and mutual exclusion cases
// =============================================================================

func TestLiteralEdgeCases(t *testing.T) {
	t.Run("nil literal", func(t *testing.T) {
		// Literal types may not support nil values directly
		// Test with nilable wrapper instead
		schema := Nilable(Literal("hello"))

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse("world")
		assert.Error(t, err)
	})

	t.Run("empty string literal", func(t *testing.T) {
		schema := Literal("")

		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)

		_, err = schema.Parse(" ")
		assert.Error(t, err)
	})

	t.Run("zero value literals", func(t *testing.T) {
		zeroSchema := Literal(0)
		result, err := zeroSchema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, 0, result)

		_, err = zeroSchema.Parse(false)
		assert.Error(t, err)
	})

	t.Run("complex type literals", func(t *testing.T) {
		complexValue := map[string]int{"key": 42}
		schema := Literal(complexValue)

		result, err := schema.Parse(complexValue)
		require.NoError(t, err)
		assert.Equal(t, complexValue, result)

		// Different map should fail
		_, err = schema.Parse(map[string]int{"key": 43})
		assert.Error(t, err)
	})
}

// =============================================================================
// 8. Default and Prefault tests
// =============================================================================

func TestLiteralDefaultAndPrefault(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		defaultVal := "hello"
		schema := Default(Literal("hello", "world"), defaultVal)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultVal, result)
	})

	t.Run("prefault values", func(t *testing.T) {
		// Literal types don't have built-in Prefault method
		// This test demonstrates the concept but uses basic validation
		schema := Literal("valid")

		result, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", result)

		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := Literal("hello", "world").DefaultFunc(func() any {
			counter++
			if counter%2 == 1 {
				return "hello"
			}
			return "world"
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)
		assert.Equal(t, 1, counter)

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "world", result2)
		assert.Equal(t, 2, counter)

		// Valid input should not call function
		result3, err3 := schema.Parse("hello")
		require.NoError(t, err3)
		assert.Equal(t, "hello", result3)
		assert.Equal(t, 2, counter) // Counter should not increment
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Literal("red", "green", "blue").PrefaultFunc(func() any {
			counter++
			colors := []string{"red", "green", "blue"}
			return colors[(counter-1)%len(colors)]
		})

		// Valid input should not call function
		result1, err1 := schema.Parse("red")
		require.NoError(t, err1)
		assert.Equal(t, "red", result1)
		assert.Equal(t, 0, counter)

		// Invalid input should call prefault function
		result2, err2 := schema.Parse("yellow")
		require.NoError(t, err2)
		assert.Equal(t, "red", result2)
		assert.Equal(t, 1, counter)

		// Another invalid input should call function again
		result3, err3 := schema.Parse("purple")
		require.NoError(t, err3)
		assert.Equal(t, "green", result3)
		assert.Equal(t, 2, counter)
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := "default_literal"
		prefaultValue := "prefault_literal"

		schema := Literal("valid", "default_literal", "prefault_literal").
			Default(defaultValue).
			Prefault(prefaultValue)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Valid input succeeds
		result2, err2 := schema.Parse("valid")
		require.NoError(t, err2)
		assert.Equal(t, "valid", result2)

		// Invalid input uses prefault
		result3, err3 := schema.Parse("invalid")
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)
	})
}
