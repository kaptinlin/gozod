package gozod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestNeverBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := Never()

		// Never always fails for any input
		testCases := []interface{}{
			"string", 42, true, nil, []int{1, 2, 3}, map[string]interface{}{"key": "value"},
		}

		for _, input := range testCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)

			var zodErr *ZodError
			require.True(t, IsZodError(err, &zodErr))
			assert.Equal(t, "invalid_type", zodErr.Issues[0].Code)
			assert.Equal(t, "never", zodErr.Issues[0].Expected)
		}
	})

	t.Run("smart type inference with nilable", func(t *testing.T) {
		schema := Never().Nilable()

		// Nilable Never accepts nil and returns typed nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*interface{})(nil), result)

		// Still fails for actual values
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("constructors", func(t *testing.T) {
		schema1 := Never()
		require.NotNil(t, schema1)
		assert.Equal(t, "never", schema1.GetInternals().Type)

		schema2 := NewZodNever()
		require.NotNil(t, schema2)
		assert.Equal(t, "never", schema2.GetInternals().Type)
	})

	t.Run("MustParse panics", func(t *testing.T) {
		schema := Never()
		assert.Panics(t, func() {
			schema.MustParse("test")
		})
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Never()
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema rejects non-nil values
		_, err2 := nilableSchema.Parse("hello")
		assert.Error(t, err2)

		// Critical: Original schema should remain unchanged
		_, err3 := baseSchema.Parse(nil)
		assert.Error(t, err3, "Original schema should still reject nil")

		_, err4 := baseSchema.Parse("hello")
		assert.Error(t, err4, "Original schema should reject all values")
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestNeverCoercion(t *testing.T) {
	t.Run("no coercion available", func(t *testing.T) {
		// Never type doesn't support coercion - it always fails
		schema := Never(SchemaParams{Coerce: true})
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestNeverValidations(t *testing.T) {
	t.Run("refine never called", func(t *testing.T) {
		// Refine is never called because Parse always fails
		called := false
		schema := Never().RefineAny(func(val any) bool {
			called = true
			return true
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)
		assert.False(t, called, "Refine should not be called for Never type")
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestNeverModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Never().Optional()

		// Optional never passes for nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Still fails for actual values
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Never().Nilable()

		// Nilable never passes for nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Still fails for actual values
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		schema := Never().Nullish()

		// Nullish never passes for nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Still fails for actual values
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestNeverChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		schema := Never().Nilable()

		// Verify chaining works
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("refine chaining", func(t *testing.T) {
		schema := Never().
			Refine(func(val any) bool {
				return false // Never reached
			}).
			RefineAny(func(val any) bool {
				return false // Never reached
			})

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestNeverTransformPipe(t *testing.T) {
	t.Run("transform never called", func(t *testing.T) {
		called := false
		schema := Never().Transform(func(val any, ctx *RefinementContext) (any, error) {
			called = true
			return val, nil
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)
		assert.False(t, called, "Transform should not be called for Never type")
	})

	t.Run("transformAny never called", func(t *testing.T) {
		called := false
		schema := Never().TransformAny(func(val any, ctx *RefinementContext) (any, error) {
			called = true
			return val, nil
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)
		assert.False(t, called, "TransformAny should not be called for Never type")
	})

	t.Run("pipe composition", func(t *testing.T) {
		schema := Never().Pipe(String())

		// Never fails before reaching the piped schema
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestNeverRefine(t *testing.T) {
	t.Run("refine with custom message", func(t *testing.T) {
		schema := Never().RefineAny(func(val any) bool {
			return false
		}, SchemaParams{Error: "Custom never error"})

		_, err := schema.Parse("test")
		assert.Error(t, err)
		// Never fails at Parse level, not at Refine level
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := "hello"

		// Refine: never called because Parse always fails
		refineSchema := Never().Refine(func(val any) bool {
			return true // Never reached
		})
		_, refineErr := refineSchema.Parse(input)

		// Transform: never called because Parse always fails
		transformSchema := Never().Transform(func(val any, ctx *RefinementContext) (any, error) {
			return val, nil // Never reached
		})
		_, transformErr := transformSchema.Parse(input)

		// Both should fail at Parse level
		assert.Error(t, refineErr)
		assert.Error(t, transformErr)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestNeverErrorHandling(t *testing.T) {
	t.Run("custom error message", func(t *testing.T) {
		schema := Never(SchemaParams{Error: "Custom never message"})

		_, err := schema.Parse("test")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("error function", func(t *testing.T) {
		schema := Never(SchemaParams{
			Error: func(issue ZodRawIssue) string {
				return "Function-based error"
			},
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues[0].Message)
	})

	t.Run("parse context error", func(t *testing.T) {
		schema := Never()
		ctx := &ParseContext{
			Error: func(issue ZodRawIssue) string {
				return "Context error"
			},
		}

		_, err := schema.Parse("test", ctx)
		assert.Error(t, err)
	})

	t.Run("error structure", func(t *testing.T) {
		schema := Never()
		_, err := schema.Parse("test")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, "invalid_type", zodErr.Issues[0].Code)
		assert.Equal(t, "never", zodErr.Issues[0].Expected)
		assert.Contains(t, zodErr.Issues[0].Message, "never type should never receive any value")
	})
}

// =============================================================================
// 9. Edge cases and internals
// =============================================================================

func TestNeverEdgeCases(t *testing.T) {
	t.Run("internals access", func(t *testing.T) {
		schema := Never()
		internals := schema.GetInternals()

		assert.Equal(t, ZodTypeNever, internals.Type)
		assert.Equal(t, Version, internals.Version)
	})

	t.Run("constructor variants", func(t *testing.T) {
		// Test different constructors
		schema1 := Never()
		schema2 := NewZodNever()

		assert.NotNil(t, schema1)
		assert.NotNil(t, schema2)
		assert.Equal(t, schema1.GetInternals().Type, schema2.GetInternals().Type)
	})

	t.Run("parameters storage", func(t *testing.T) {
		params := SchemaParams{
			Params: map[string]interface{}{
				"custom": "value",
			},
		}

		schema := Never(params)
		assert.Equal(t, "value", schema.GetZod().Bag["custom"])
	})

	t.Run("clone functionality", func(t *testing.T) {
		original := Never()
		original.GetZod().Bag["custom"] = "value"

		cloned := &ZodNever{internals: &ZodNeverInternals{
			ZodTypeInternals: ZodTypeInternals{},
			Bag:              make(map[string]interface{}),
		}}
		cloned.CloneFrom(original)

		assert.Equal(t, "value", cloned.GetZod().Bag["custom"])
	})

	t.Run("unwrap returns self", func(t *testing.T) {
		schema := Never()
		unwrapped := schema.Unwrap()
		assert.Equal(t, schema, unwrapped)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestNeverDefaultAndPrefault(t *testing.T) {
	t.Run("default with never", func(t *testing.T) {
		// Never with default - the default will be used for nil, but Never still fails for non-nil
		schema := Default(Never(), "default_value")

		// For nil input, Default returns the default value without calling Never
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Equal(t, "default_value", result)

		// For non-nil input, Never still fails
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("prefault with never", func(t *testing.T) {
		// Never types don't have built-in Prefault method
		// This test demonstrates that Never always fails regardless of prefault
		schema := Never()

		_, err := schema.Parse("any_value")
		assert.Error(t, err)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}
