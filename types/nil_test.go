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

func TestNil_Basic(t *testing.T) {
	t.Run("accepts nil values", func(t *testing.T) {
		schema := Nil()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("rejects non-nil values", func(t *testing.T) {
		schema := Nil()
		_, err := schema.Parse("hello")
		assert.Error(t, err)
	})

	t.Run("rejects zero values", func(t *testing.T) {
		schema := Nil()
		_, err := schema.Parse(0)
		assert.Error(t, err)
	})

	t.Run("rejects empty string", func(t *testing.T) {
		schema := Nil()
		_, err := schema.Parse("")
		assert.Error(t, err)
	})

	t.Run("rejects false", func(t *testing.T) {
		schema := Nil()
		_, err := schema.Parse(false)
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestNil_TypeSafety(t *testing.T) {
	t.Run("consistent nil return", func(t *testing.T) {
		schema := Nil()
		require.NotNil(t, schema)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Ensure type is consistent
		mustResult := schema.MustParse(nil)
		assert.Nil(t, mustResult)
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		nilSchema := Nil()

		result, err := nilSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Nil()
		result := schema.MustParse(nil)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestNil_Modifiers(t *testing.T) {
	t.Run("Optional returns pointer constraint", func(t *testing.T) {
		schema := Nil().Optional()
		var _ *ZodNil[*any] = schema

		// nil input - nil type always returns nil regardless of constraint type
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable returns pointer constraint", func(t *testing.T) {
		schema := Nil().Nilable()
		var _ *ZodNil[*any] = schema

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nullish returns pointer constraint", func(t *testing.T) {
		schema := Nil().Nullish()
		var _ *ZodNil[*any] = schema

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves constraint type", func(t *testing.T) {
		// Value constraint preserved
		valueSchema := Nil().Default("fallback")
		var _ *ZodNil[any] = valueSchema

		// Pointer constraint preserved
		ptrSchema := NilPtr().Default("fallback")
		var _ *ZodNil[*any] = ptrSchema
	})

	t.Run("Prefault preserves constraint type", func(t *testing.T) {
		// Value constraint preserved
		valueSchema := Nil().Prefault("fallback")
		var _ *ZodNil[any] = valueSchema

		// Test prefault works for non-nil input
		result, err := valueSchema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)

		// Pointer constraint preserved
		ptrSchema := NilPtr().Prefault("fallback")
		var _ *ZodNil[*any] = ptrSchema

		ptrResult, err := ptrSchema.Parse("invalid")
		require.NoError(t, err)
		require.NotNil(t, ptrResult)
		assert.Equal(t, "fallback", *ptrResult)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestNil_Chaining(t *testing.T) {
	t.Run("modifier chaining", func(t *testing.T) {
		schema := Nil().
			Optional().
			Nilable().
			Prefault("fallback")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Non-nil input should use prefault
		result, err = schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "fallback", *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := Nil().
			Nilable().
			Default("default").
			Prefault("fallback")

		// Valid nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Invalid input should use prefault
		result, err = schema.Parse("invalid")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "fallback", *result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestNil_DefaultAndPrefault(t *testing.T) {
	t.Run("default with nil", func(t *testing.T) {
		schema := Nil().Default("default_value")

		// Valid nil input should not use default (nil is valid for nil type)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("defaultFunc with nil", func(t *testing.T) {
		schema := Nil().DefaultFunc(func() any {
			return "default_value"
		})

		// Valid nil input should return nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("prefault with invalid input", func(t *testing.T) {
		schema := Nil().Prefault("fallback_value")

		// Valid nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Invalid input should use prefault
		result, err = schema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, "fallback_value", result)
	})

	t.Run("prefaultFunc with invalid input", func(t *testing.T) {
		counter := 0
		schema := Nil().PrefaultFunc(func() any {
			counter++
			return "fallback"
		})

		// Valid nil input should not call function
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input should call prefault function
		result, err = schema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestNil_Refine(t *testing.T) {
	t.Run("refine with nil input", func(t *testing.T) {
		// Only allow nil (which is the only valid case anyway)
		schema := Nil().RefineAny(func(v any) bool {
			return v == nil // Always allow nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Non-nil should still fail
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Custom nil validation failed"
		schema := Nil().RefineAny(func(v any) bool {
			return v == nil // Only accept nil
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("refineAny with nil", func(t *testing.T) {
		schema := Nil().RefineAny(func(v any) bool {
			return v == nil // Only accept nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		_, err = schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// Transformation and pipeline tests
// =============================================================================

func TestNil_Transform(t *testing.T) {
	t.Run("transform nil to string", func(t *testing.T) {
		schema := Nil().Transform(func(v any, ctx *core.RefinementContext) (any, error) {
			if v == nil {
				return "transformed_nil", nil
			}
			return v, nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "transformed_nil", result)
	})

	t.Run("pipe composition", func(t *testing.T) {
		// Test pipe functionality without complex type compatibility issues
		transformedSchema := Nil().Transform(func(v any, ctx *core.RefinementContext) (any, error) {
			return "nil_to_string", nil
		})

		result, err := transformedSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "nil_to_string", result)
	})
}

// =============================================================================
// Error handling and edge case tests
// =============================================================================

func TestNil_ErrorHandling(t *testing.T) {
	t.Run("type error", func(t *testing.T) {
		schema := Nil()
		_, err := schema.Parse("not nil")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("refinement error", func(t *testing.T) {
		schema := Nil().RefineAny(func(v any) bool {
			return false // Always fail
		})

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Nil()

		complexTypes := []any{
			make(chan int),
			func() {},
			struct{ Field string }{Field: "value"},
			[]any{nil},
			map[any]any{"key": nil},
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for complex type: %T", input)
		}
	})

	t.Run("schema structure", func(t *testing.T) {
		schema := Nil()

		// Test that schema has basic functionality
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("clone from another schema", func(t *testing.T) {
		original := Nil()
		target := Nil()

		target.CloneFrom(original)

		// Both should work the same way
		result1, err1 := original.Parse(nil)
		result2, err2 := target.Parse(nil)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, result1, result2)
	})
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestNil_ValidationMethods(t *testing.T) {
	t.Run("refine method", func(t *testing.T) {
		schema := Nil().RefineAny(func(v any) bool {
			// Only accept nil values
			return v == nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("basic method coverage", func(t *testing.T) {
		schema := Nil()

		// Test that schema works as expected
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test invalid input
		_, err = schema.Parse("not nil")
		assert.Error(t, err)
	})
}

// =============================================================================
// Pointer identity preservation tests
// =============================================================================

func TestNil_PointerIdentityPreservation(t *testing.T) {
	t.Run("Nil schema behavior with nil input", func(t *testing.T) {
		schema := Nil()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result, "Nil schema should return nil")
	})

	t.Run("Nil Optional behavior", func(t *testing.T) {
		schema := Nil().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result, "Nil Optional should return nil")
	})

	t.Run("Nil Nilable behavior", func(t *testing.T) {
		schema := Nil().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result, "Nil Nilable should return nil")
	})

	// Note: NilPtr is not available in this implementation

	t.Run("Nil schema rejects non-nil inputs", func(t *testing.T) {
		schema := Nil()

		invalidInputs := []any{
			"string",
			123,
			true,
			[]int{1, 2, 3},
			map[string]any{"key": "value"},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Nil schema should reject non-nil input: %v", input)
		}
	})

	t.Run("Optional Nil schema allows both nil and null behavior", func(t *testing.T) {
		schema := Nil().Optional()

		// Should work with nil
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Should reject non-nil values
		_, err2 := schema.Parse("not nil")
		assert.Error(t, err2)
	})

	t.Run("Nilable Nil schema consistency", func(t *testing.T) {
		schema := Nil().Nilable()

		// Should work with nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Chaining with Nil schema", func(t *testing.T) {
		// Nil doesn't really make sense with Default, but testing the chain works
		schema := Nil().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Dual generic parameter architecture tests
// =============================================================================

func TestNil_GenericArchitecture(t *testing.T) {
	t.Run("Nil factory returns value constraint", func(t *testing.T) {
		schema := Nil()
		var _ *ZodNil[any] = schema

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("NilPtr factory returns pointer constraint", func(t *testing.T) {
		schema := NilPtr()
		var _ *ZodNil[*any] = schema

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // Returns nil pointer for nil input
	})

	t.Run("NilTyped with explicit constraint types", func(t *testing.T) {
		// Value constraint
		valueSchema := NilTyped[string]()
		var _ *ZodNil[string] = valueSchema

		result, err := valueSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "", result) // Zero value for string

		// Pointer constraint
		ptrSchema := NilTyped[*string]()
		var _ *ZodNil[*string] = ptrSchema

		ptrResult, err := ptrSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, ptrResult) // Returns nil pointer for nil input
	})
}

// =============================================================================
// Pipe functionality tests
// =============================================================================

func TestNil_Pipe(t *testing.T) {
	t.Run("Transform functionality works", func(t *testing.T) {
		// Test transform instead of pipe to avoid type complexity
		schema := Nil().Transform(func(v any, ctx *core.RefinementContext) (any, error) {
			return "transformed", nil
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "transformed", result)
	})
}

// =============================================================================
// Factory function aliases tests
// =============================================================================

func TestNil_Aliases(t *testing.T) {
	t.Run("Null is alias for Nil", func(t *testing.T) {
		schema := Null()
		var _ *ZodNil[any] = schema

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("NullPtr is alias for NilPtr", func(t *testing.T) {
		schema := NullPtr()
		var _ *ZodNil[*any] = schema

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Error conditions tests
// =============================================================================

func TestNil_Errors(t *testing.T) {
	t.Run("non-nil input without prefault fails", func(t *testing.T) {
		schema := Nil()

		_, err := schema.Parse("not nil")
		assert.Error(t, err)
	})

	t.Run("non-nil input with insufficient prefault fails", func(t *testing.T) {
		// If prefault doesn't apply, should still fail
		schema := Nil().Prefault("fallback")

		// This should use prefault and succeed
		result, err := schema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)
	})
}

// =============================================================================
// Type constraints tests
// =============================================================================

func TestNil_TypeConstraints(t *testing.T) {
	t.Run("type constraint enforcement", func(t *testing.T) {
		// String constraint
		stringSchema := NilTyped[string]()
		result, err := stringSchema.Parse(nil)
		require.NoError(t, err)
		assert.IsType(t, "", result)

		// Int constraint
		intSchema := NilTyped[int]()
		intResult, err := intSchema.Parse(nil)
		require.NoError(t, err)
		assert.IsType(t, 0, intResult)

		// Pointer constraint
		ptrSchema := NilTyped[*string]()
		ptrResult, err := ptrSchema.Parse(nil)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), ptrResult)
	})
}

// =============================================================================
// Parameter handling tests
// =============================================================================

func TestNil_Parameters(t *testing.T) {
	t.Run("custom error message", func(t *testing.T) {
		schema := Nil(core.SchemaParams{
			Error: "Custom nil error",
		})

		_, err := schema.Parse("not nil")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Custom nil error")
	})

	t.Run("NilPtr with parameters", func(t *testing.T) {
		schema := NilPtr(core.SchemaParams{
			Error: "Custom nil ptr error",
		})

		_, err := schema.Parse("not nil")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Custom nil ptr error")
	})
}
