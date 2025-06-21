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

func TestStringBoolBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := StringBool()

		// Test truthy values
		truthyValues := []string{"true", "1", "yes", "on", "y", "enabled", "TRUE"}
		for _, value := range truthyValues {
			result, err := schema.Parse(value)
			require.NoError(t, err)
			assert.Equal(t, true, result)
		}

		// Test falsy values
		falsyValues := []string{"false", "0", "no", "off", "n", "disabled", "FALSE"}
		for _, value := range falsyValues {
			result, err := schema.Parse(value)
			require.NoError(t, err)
			assert.Equal(t, false, result)
		}

		// Test invalid values
		invalidValues := []any{"other", "", "maybe", 123, true, false, nil, struct{}{}}
		for _, value := range invalidValues {
			_, err := schema.Parse(value)
			assert.Error(t, err)
		}
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := StringBool()

		// String input returns bool
		result1, err := schema.Parse("true")
		require.NoError(t, err)
		assert.IsType(t, true, result1)
		assert.Equal(t, true, result1)

		// Pointer input returns bool (not pointer)
		str := "false"
		result2, err := schema.Parse(&str)
		require.NoError(t, err)
		assert.IsType(t, false, result2)
		assert.Equal(t, false, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := StringBool().Nilable()

		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*bool)(nil), result)

		// Valid input keeps type inference
		result2, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result2)
		assert.IsType(t, true, result2)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestStringBoolCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := StringBool(nil, core.SchemaParams{Coerce: true})

		// Numbers to strings
		result1, err := schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		// Boolean to string (should fail - booleans are not coerced to strings)
		_, err = schema.Parse(true)
		assert.Error(t, err)
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := StringBool(nil, core.SchemaParams{Coerce: true})

		// Valid coerced values
		result, err := schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid coerced values
		_, err = schema.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestStringBoolValidations(t *testing.T) {
	t.Run("custom truthy and falsy values", func(t *testing.T) {
		schema := StringBool(&StringBoolOptions{
			Truthy: []string{"y"},
			Falsy:  []string{"N"},
		})

		// Test custom truthy (case insensitive by default)
		result1, err := schema.Parse("y")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse("Y")
		require.NoError(t, err)
		assert.Equal(t, true, result2)

		// Test custom falsy (case insensitive by default)
		result3, err := schema.Parse("n")
		require.NoError(t, err)
		assert.Equal(t, false, result3)

		result4, err := schema.Parse("N")
		require.NoError(t, err)
		assert.Equal(t, false, result4)

		// Test default values should fail
		_, err = schema.Parse("true")
		assert.Error(t, err)

		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("case sensitive validation", func(t *testing.T) {
		schema := StringBool(&StringBoolOptions{
			Truthy: []string{"y"},
			Falsy:  []string{"N"},
			Case:   "sensitive",
		})

		// Test exact case match
		result1, err := schema.Parse("y")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse("N")
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		// Test case mismatch should fail
		_, err = schema.Parse("Y")
		assert.Error(t, err)

		_, err = schema.Parse("n")
		assert.Error(t, err)

		// Test default values should fail
		_, err = schema.Parse("TRUE")
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestStringBoolModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schema := StringBool().Optional()

		// Valid input
		result1, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		// nil input should succeed
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)

		// Invalid input should fail
		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := StringBool().Nilable()

		// Valid input
		result1, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result1)

		// nil input should succeed, return typed nil pointer
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		assert.IsType(t, (*bool)(nil), result2)

		// Invalid input should fail
		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("nullish modifier", func(t *testing.T) {
		schema := StringBool().Nullish()

		// Valid input
		result1, err := schema.Parse("yes")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		// nil input should succeed
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := StringBool()

		// Valid input should not panic
		result := schema.MustParse("on")
		assert.Equal(t, true, result)

		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := StringBool()
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse("true")
		require.NoError(t, err2)
		assert.Equal(t, true, result2)

		// Test nilable schema rejects invalid values
		_, err3 := nilableSchema.Parse("invalid")
		assert.Error(t, err3)

		// Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse("true")
		require.NoError(t, err5)
		assert.Equal(t, true, result5)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestStringBoolChaining(t *testing.T) {
	t.Run("chaining with modifiers", func(t *testing.T) {
		schema := StringBool().Nilable()

		// Test chained validation
		result1, err := schema.Parse("enabled")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)

		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("custom options with modifiers", func(t *testing.T) {
		schema := StringBool(&StringBoolOptions{
			Truthy: []string{"yes", "ok"},
			Falsy:  []string{"no", "nope"},
		}).Optional()

		result1, err := schema.Parse("yes")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse("nope")
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		result3, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result3)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestStringBoolTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := StringBool().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if b, ok := val.(bool); ok {
				if b {
					return "YES", nil
				}
				return "NO", nil
			}
			return val, nil
		})

		result1, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, "YES", result1)

		result2, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, "NO", result2)
	})

	t.Run("transform chain", func(t *testing.T) {
		schema := StringBool().
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if b, ok := val.(bool); ok {
					if b {
						return 1, nil
					}
					return 0, nil
				}
				return val, nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if i, ok := val.(int); ok {
					return i * 10, nil
				}
				return val, nil
			})

		result1, err := schema.Parse("yes")
		require.NoError(t, err)
		assert.Equal(t, 10, result1)

		result2, err := schema.Parse("no")
		require.NoError(t, err)
		assert.Equal(t, 0, result2)
	})

	t.Run("pipe to another schema", func(t *testing.T) {
		stringSchema := String().Min(2)
		schema := StringBool().Pipe(stringSchema)

		// This should fail because bool -> string pipe doesn't work directly
		// The pipe would need proper type conversion
		_, err := schema.Parse("true")
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestStringBoolRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := StringBool().RefineAny(func(val any) bool {
			if b, ok := val.(bool); ok {
				return b == true // Only allow true values
			}
			return false
		})

		// Valid input (true)
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid input (false fails refine)
		_, err = schema.Parse("false")
		assert.Error(t, err)

		// Invalid input (not a valid stringbool)
		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := StringBool().RefineAny(func(val any) bool {
			if b, ok := val.(bool); ok {
				return b == true
			}
			return false
		}, core.SchemaParams{
			Error: "Only true values are allowed",
		})

		_, err := schema.Parse("false")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Only true values are allowed")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := "true"

		// Refine: only validates, never modifies
		refineSchema := StringBool().RefineAny(func(val any) bool {
			if b, ok := val.(bool); ok {
				return b == true
			}
			return false
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := StringBool().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if b, ok := val.(bool); ok {
				if b {
					return "ENABLED", nil
				}
				return "DISABLED", nil
			}
			return val, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original converted value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, true, refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, "ENABLED", transformResult)

		// Key distinction: Refine preserves converted value, Transform modifies
		assert.IsType(t, true, refineResult, "Refine should return bool from stringbool conversion")
		assert.IsType(t, "", transformResult, "Transform should return modified string")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestStringBoolErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := StringBool()
		_, err := schema.Parse("invalid")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidValue, zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := StringBool(nil, core.SchemaParams{Error: "Custom stringbool error"})
		_, err := schema.Parse("invalid")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Custom stringbool error")
	})

	t.Run("type error vs value error", func(t *testing.T) {
		schema := StringBool()

		// Type error (not a string)
		_, err1 := schema.Parse(123)
		assert.Error(t, err1)
		var zodErr1 *issues.ZodError
		require.True(t, issues.IsZodError(err1, &zodErr1))
		assert.Equal(t, core.InvalidType, zodErr1.Issues[0].Code)

		// Value error (string but invalid value)
		_, err2 := schema.Parse("invalid")
		assert.Error(t, err2)
		var zodErr2 *issues.ZodError
		require.True(t, issues.IsZodError(err2, &zodErr2))
		assert.Equal(t, core.InvalidValue, zodErr2.Issues[0].Code)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestStringBoolEdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		schema := StringBool()
		_, err := schema.Parse("")
		assert.Error(t, err)
	})

	t.Run("whitespace strings", func(t *testing.T) {
		schema := StringBool()
		_, err := schema.Parse(" true ")
		assert.Error(t, err) // Whitespace should not be trimmed
	})

	t.Run("nil pointer handling", func(t *testing.T) {
		schema := StringBool()
		var nilPtr *string = nil

		// Non-nilable should reject nil pointer
		_, err := schema.Parse(nilPtr)
		assert.Error(t, err)

		// Nilable should accept nil pointer
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nilPtr)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("case insensitive by default", func(t *testing.T) {
		schema := StringBool()

		// Mixed case should work
		result1, err := schema.Parse("True")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse("FALSE")
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		result3, err := schema.Parse("YES")
		require.NoError(t, err)
		assert.Equal(t, true, result3)
	})

	t.Run("modifier combinations", func(t *testing.T) {
		// Optional + custom values
		schema := StringBool(&StringBoolOptions{
			Truthy: []string{"si"},
			Falsy:  []string{"no"},
		}).Optional()

		result1, err := schema.Parse("si")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)

		_, err = schema.Parse("yes")
		assert.Error(t, err) // Default values should not work
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestStringBoolDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := StringBool().Default(true)

		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Valid input normal validation
		result2, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		// Invalid input still fails
		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := StringBool().DefaultFunc(func() bool {
			counter++
			return counter%2 == 1
		})

		// Each nil input generates a new default value
		result1, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		// Valid input bypasses default generation
		result3, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("prefault value", func(t *testing.T) {
		schema := StringBool().Prefault(false)

		// Valid input normal validation
		result1, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		// Invalid input uses prefault value
		result2, err := schema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, false, result2)

		// nil input still fails (prefault doesn't handle nil)
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := StringBool().PrefaultFunc(func() bool {
			counter++
			return true
		})

		// Valid input normal validation
		result1, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result1)
		assert.Equal(t, 0, counter) // No prefault called

		// Invalid input uses prefault function
		result2, err := schema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, true, result2)
		assert.Equal(t, 1, counter) // Prefault called once
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultSchema := StringBool().Default(true)
		prefaultSchema := StringBool().Prefault(false)

		// Default handles nil, prefault doesn't
		result1, err := defaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result1)

		_, err = prefaultSchema.Parse(nil)
		assert.Error(t, err)

		// Prefault handles invalid values, default doesn't
		_, err = defaultSchema.Parse("invalid")
		assert.Error(t, err)

		result2, err := prefaultSchema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, false, result2)
	})

	t.Run("custom options with default", func(t *testing.T) {
		schema := StringBool(&StringBoolOptions{
			Truthy: []string{"si", "oui"},
			Falsy:  []string{"no", "non"},
		}).Default(true)

		// Custom truthy value
		result, err := schema.Parse("si")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Custom falsy value
		result, err = schema.Parse("non")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// nil input uses default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid input still fails
		_, err = schema.Parse("maybe")
		assert.Error(t, err)
	})
}
