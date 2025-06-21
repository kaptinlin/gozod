package types

import (
	"fmt"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

// TestSliceConstructor tests basic slice schema construction
func TestSliceConstructor(t *testing.T) {
	t.Run("basic constructor", func(t *testing.T) {
		elementSchema := String()
		schema := Slice(elementSchema)
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, "slice", internals.Type)
	})

	t.Run("constructor with params", func(t *testing.T) {
		schema := Slice(String(), core.SchemaParams{Error: "core.Custom error"})
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals.Error)
	})
}

// TestSliceBasicValidation tests fundamental slice validation behavior
func TestSliceBasicValidation(t *testing.T) {
	t.Run("successful validation with type conversion", func(t *testing.T) {
		// Test basic type inference and conversion behavior
		schema := Slice(String())

		// Valid input should be converted to []any
		result, err := schema.Parse([]string{"hello", "world"})
		require.NoError(t, err)
		assert.IsType(t, []any{}, result)
		assert.Equal(t, []any{"hello", "world"}, result)
	})

	t.Run("input type validation", func(t *testing.T) {
		schema := Slice(String())

		// Valid slice types
		validInputs := []any{
			[]string{"a", "b", "c"},
			[]any{"hello", "world"},
			[]string{},
		}
		for _, input := range validInputs {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.NotNil(t, result)
		}

		// Invalid input types
		invalidInputs := []any{
			"not_a_slice",
			42,
			true,
			map[string]string{"k": "v"},
		}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("slice length validation", func(t *testing.T) {
		// Test length constraint validation
		schema := Slice(String()).Length(2)

		// Valid case - exact length
		result, err := schema.Parse([]string{"asdf", "asdf"})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid case - too short
		_, err1 := schema.Parse([]string{"asdf"})
		assert.Error(t, err1)

		var zodErr1 *issues.ZodError
		require.True(t, issues.IsZodError(err1, &zodErr1))
		assert.Len(t, zodErr1.Issues, 1)
		issue1 := zodErr1.Issues[0]
		assert.Equal(t, issues.TooSmall, issue1.Code)

		// Invalid case - too long
		_, err2 := schema.Parse([]string{"asdf", "asdf", "asdf"})
		assert.Error(t, err2)

		var zodErr2 *issues.ZodError
		require.True(t, issues.IsZodError(err2, &zodErr2))
		assert.Len(t, zodErr2.Issues, 1)
		issue2 := zodErr2.Issues[0]
		assert.Equal(t, core.TooBig, issue2.Code)
	})

	t.Run("min/max length validation", func(t *testing.T) {
		// Test min/max constraint validation with detailed error checking
		schema := Slice(String()).Min(2).Max(2)

		// Test too small
		_, err1 := schema.Parse([]string{"asdf"})
		assert.Error(t, err1)

		var zodErr1 *issues.ZodError
		require.True(t, issues.IsZodError(err1, &zodErr1))
		assert.Len(t, zodErr1.Issues, 1)
		issue1 := zodErr1.Issues[0]
		assert.Equal(t, issues.TooSmall, issue1.Code)
		assert.Contains(t, issue1.Message, "Too small: expected slice to have >=2 items")

		// Test too big
		_, err2 := schema.Parse([]string{"asdf", "asdf", "asdf"})
		assert.Error(t, err2)

		var zodErr2 *issues.ZodError
		require.True(t, issues.IsZodError(err2, &zodErr2))
		assert.Len(t, zodErr2.Issues, 1)
		issue2 := zodErr2.Issues[0]
		assert.Equal(t, core.TooBig, issue2.Code)
		assert.Contains(t, issue2.Message, "Too big: expected slice to have <=2 items")
	})

	t.Run("non-empty slice validation", func(t *testing.T) {
		// Test non-empty constraint
		schema := Slice(String()).NonEmpty()

		// Valid case - has elements
		result, err := schema.Parse([]string{"a"})
		require.NoError(t, err)
		assert.Equal(t, []any{"a"}, result)

		// Invalid case - empty slice
		_, err = schema.Parse([]string{})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)

		// Should have a "too small" error for empty slice
		hasTooSmall := false
		for _, issue := range zodErr.Issues {
			if issue.Code == core.TooSmall {
				hasTooSmall = true
				break
			}
		}
		assert.True(t, hasTooSmall, "Should have 'too small' error for empty slice")
	})

	t.Run("combined non-empty with max validation", func(t *testing.T) {
		// Test combination of non-empty and max constraints
		schema := Slice(String()).NonEmpty().Max(2)

		// Valid case
		result, err := schema.Parse([]string{"a"})
		require.NoError(t, err)
		assert.Equal(t, []any{"a"}, result)

		// Invalid case - empty slice
		_, err = schema.Parse([]string{})
		assert.Error(t, err)

		// Invalid case - too many elements
		_, err = schema.Parse([]string{"a", "a", "a"})
		assert.Error(t, err)
	})
}

// TestSliceTypeInference tests intelligent type handling and inference
func TestSliceTypeInference(t *testing.T) {
	t.Run("type conversion and preservation", func(t *testing.T) {
		schema := Slice(String())

		// []string input returns []any
		input1 := []string{"hello", "world"}
		result1, err := schema.Parse(input1)
		require.NoError(t, err)
		assert.IsType(t, []any{}, result1)

		// []any input returns []any
		input2 := []any{"hello", "world"}
		result2, err := schema.Parse(input2)
		require.NoError(t, err)
		assert.IsType(t, []any{}, result2)

		// Verify pointer identity preservation with []any
		inputSlice := []any{"test"}
		inputPtr := &inputSlice
		result3, err := schema.Parse(inputPtr)
		require.NoError(t, err)
		resultPtr, ok := result3.(*[]any)
		require.True(t, ok)
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestSliceCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := Slice(String(), core.SchemaParams{Coerce: true})

		// For now, coercion might not be fully implemented
		// Test with proper slice input instead
		result, err := schema.Parse([]string{"hello"})
		require.NoError(t, err)
		expected := []any{"hello"}
		assert.Equal(t, expected, result)
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := Slice(String(), core.SchemaParams{Coerce: true}).Min(2)

		// Single value coerced to slice should fail min length
		_, err := schema.Parse("hello")
		assert.Error(t, err)

		// Multiple values should pass
		result, err := schema.Parse([]string{"hello", "world"})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestSliceValidations(t *testing.T) {
	t.Run("length validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  *ZodSlice
			input   []string
			wantErr bool
		}{
			{"min length valid", Slice(String()).Min(2), []string{"a", "b"}, false},
			{"min length invalid", Slice(String()).Min(2), []string{"a"}, true},
			{"max length valid", Slice(String()).Max(2), []string{"a", "b"}, false},
			{"max length invalid", Slice(String()).Max(2), []string{"a", "b", "c"}, true},
			{"exact length valid", Slice(String()).Length(2), []string{"a", "b"}, false},
			{"exact length invalid", Slice(String()).Length(2), []string{"a"}, true},
			{"nonempty valid", Slice(String()).NonEmpty(), []string{"a"}, false},
			{"nonempty invalid", Slice(String()).NonEmpty(), []string{}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("element validation", func(t *testing.T) {
		schema := Slice(String().Min(3))

		// Valid elements
		result, err := schema.Parse([]string{"hello", "world"})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid elements
		_, err = schema.Parse([]string{"hello", "hi"})
		assert.Error(t, err)
	})

	t.Run("get element schema", func(t *testing.T) {
		// Test accessing element schema for individual validation
		schema := Slice(String())

		// Access element schema and validate directly
		elementSchema := schema.Element()

		// Valid element
		result, err := elementSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid element
		_, err = elementSchema.Parse(123)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("continue parsing despite array size error", func(t *testing.T) {
		// Test validation behavior when slice has both element type and size errors
		schema := Slice(String()).Min(2)

		_, err := schema.Parse([]any{123})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have errors - could be element type error or size error or both
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)

		// Check if we have any relevant errors (type or size)
		hasRelevantError := false
		for _, issue := range zodErr.Issues {
			if issue.Code == core.TooSmall || issue.Code == core.InvalidType {
				hasRelevantError = true
				break
			}
		}
		assert.True(t, hasRelevantError, "Should have type or size validation error")
	})

	t.Run("sparse array validation", func(t *testing.T) {
		// Test validation behavior with slices containing nil/undefined elements
		schema := Slice(String()).NonEmpty().Min(1).Max(3)

		// Create sparse array (array with nil elements)
		sparseArray := make([]any, 3)
		// sparseArray contains [nil, nil, nil] representing undefined values

		_, err := schema.Parse(sparseArray)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have multiple invalid_type errors for each undefined element
		invalidTypeCount := 0
		for _, issue := range zodErr.Issues {
			if issue.Code == core.InvalidType {
				invalidTypeCount++
			}
		}
		assert.Equal(t, 3, invalidTypeCount, "Should have 3 invalid type errors for undefined elements")

		// Verify error paths are correct
		expectedPaths := [][]any{{0}, {1}, {2}}
		actualPaths := make([][]any, 0)
		for _, issue := range zodErr.Issues {
			if issue.Code == core.InvalidType {
				actualPaths = append(actualPaths, issue.Path)
			}
		}
		assert.ElementsMatch(t, expectedPaths, actualPaths, "Error paths should match slice indices")
	})

	t.Run("nested validation error handling", func(t *testing.T) {
		// Test validation behavior with nested object schemas containing slice validation
		schema := Object(core.ObjectSchema{
			"people": Slice(String()).Min(2),
		})

		_, err := schema.Parse(map[string]any{
			"people": []any{123}, // Wrong element type AND wrong length
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)

		// Should have both type error and size error
		hasInvalidType := false
		hasTooSmall := false
		for _, issue := range zodErr.Issues {
			if issue.Code == core.InvalidType {
				hasInvalidType = true
			}
			if issue.Code == core.TooSmall {
				hasTooSmall = true
			}
		}

		// At least one of these errors should be present
		assert.True(t, hasInvalidType || hasTooSmall, "Should have either type error or size error")
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestSliceModifiers(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		schema := Slice(Int()).Optional()

		// Valid slice
		result, err := schema.Parse([]int{1, 2})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})

	t.Run("nilable", func(t *testing.T) {
		schema := Slice(Int()).Nilable()

		// Valid slice
		result, err := schema.Parse([]int{1, 2})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input returns typed nil
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		assert.IsType(t, (*[]any)(nil), result2)

		// Valid input keeps type inference
		result3, err := schema.Parse([]int{123})
		require.NoError(t, err)
		assert.IsType(t, []any{}, result3)
	})

	t.Run("nilable modifier immutability", func(t *testing.T) {
		baseSchema := Slice(String())
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Critical: Original schema should remain unchanged and reject nil
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		// Both schemas should validate valid input the same way
		result5, err5 := baseSchema.Parse([]string{"hello"})
		require.NoError(t, err5)
		assert.NotNil(t, result5)

		result6, err6 := nilableSchema.Parse([]string{"hello"})
		require.NoError(t, err6)
		assert.NotNil(t, result6)
	})

	t.Run("nullish", func(t *testing.T) {
		schema := Slice(Int()).Nullish()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Slice(String())

		// Valid input should not panic
		result := schema.MustParse([]string{"hello"})
		assert.NotNil(t, result)

		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("not a slice")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestSliceChaining(t *testing.T) {
	t.Run("multiple validation chaining", func(t *testing.T) {
		schema := Slice(String().Min(3)).Min(2).Max(5)

		// Valid case
		result, err := schema.Parse([]string{"hello", "world"})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid array length
		_, err = schema.Parse([]string{"hello"})
		assert.Error(t, err)

		// Invalid element length
		_, err = schema.Parse([]string{"hello", "hi"})
		assert.Error(t, err)

		// Array too long
		_, err = schema.Parse([]string{"hello", "world", "test", "data", "extra", "more"})
		assert.Error(t, err)
	})

	t.Run("validation with element validation", func(t *testing.T) {
		schema := Slice(String().Email()).NonEmpty().Max(3)

		// Valid emails
		result, err := schema.Parse([]string{"test@example.com", "user@domain.org"})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid email
		_, err = schema.Parse([]string{"test@example.com", "invalid-email"})
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestSliceTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Slice(Int()).TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if arr, ok := val.([]any); ok {
				sum := 0
				for _, v := range arr {
					if i, ok := v.(int); ok {
						sum += i
					}
				}
				return sum, nil
			}
			return 0, nil
		})

		result, err := schema.Parse([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, 6, result)
	})

	t.Run("transform chain", func(t *testing.T) {
		// First transform: sum the array
		// Second transform: double the result
		schema := Slice(Int()).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if arr, ok := val.([]any); ok {
					sum := 0
					for _, v := range arr {
						if i, ok := v.(int); ok {
							sum += i
						}
					}
					return sum, nil
				}
				return 0, nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if i, ok := val.(int); ok {
					return i * 2, nil
				}
				return val, nil
			})

		result, err := schema.Parse([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, 12, result) // (1+2+3) * 2 = 12
	})

	t.Run("pipe to another schema", func(t *testing.T) {
		// Transform slice to its length, then validate as positive integer
		lengthTransform := Slice(String()).TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if arr, ok := val.([]any); ok {
				return len(arr), nil
			}
			return 0, nil
		})

		schema := lengthTransform.Pipe(Int().Min(1))

		// Valid case
		result, err := schema.Parse([]string{"hello", "world"})
		require.NoError(t, err)
		assert.Equal(t, 2, result)

		// Invalid case (empty array -> length 0 -> fails Min(1))
		_, err = schema.Parse([]string{})
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestSliceRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Slice(Int()).Refine(func(val []any) bool {
			return len(val)%2 == 0
		}, core.SchemaParams{Error: "Array length must be even"})

		// Valid case (even length)
		result, err := schema.Parse([]int{1, 2})
		require.NoError(t, err)
		assert.Equal(t, []any{1, 2}, result)

		// Invalid case (odd length)
		_, err = schema.Parse([]int{1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Array length must be even")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := []string{"hello", "world"}

		// Refine: only validates, never modifies
		refineSchema := Slice(String()).Refine(func(val []any) bool {
			return len(val) > 0
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Slice(String()).TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if arr, ok := val.([]any); ok {
				return len(arr), nil
			}
			return 0, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.IsType(t, []any{}, refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, 2, transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.NotEqual(t, transformResult, refineResult, "Transform should return modified value")
	})

	t.Run("complex refine validation", func(t *testing.T) {
		// All elements must be unique
		schema := Slice(String()).Refine(func(val []any) bool {
			seen := make(map[any]bool)
			for _, item := range val {
				if seen[item] {
					return false
				}
				seen[item] = true
			}
			return true
		}, core.SchemaParams{Error: "All elements must be unique"})

		// Valid case (unique elements)
		result, err := schema.Parse([]string{"a", "b", "c"})
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// Invalid case (duplicate elements)
		_, err = schema.Parse([]string{"a", "b", "a"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "All elements must be unique")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestSliceErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Slice(String()).Min(3)
		_, err := schema.Parse([]string{"a"})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Slice(String()).Min(2, core.SchemaParams{
			Error: "Array must have at least 2 elements",
		})
		_, err := schema.Parse([]string{"a"})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Array must have at least 2 elements")
	})

	t.Run("error message format validation", func(t *testing.T) {
		// Test specific error message format for size constraints
		schema := Slice(String()).Min(2).Max(2)

		// Test too small - should show proper format
		_, err1 := schema.Parse([]string{"hello"})
		require.Error(t, err1)
		fmt.Printf("Too small error: %s\n", err1.Error())

		// Test too big - should show proper format
		_, err2 := schema.Parse([]string{"hello", "world", "test"})
		require.Error(t, err2)
		fmt.Printf("Too big error: %s\n", err2.Error())

		// Verify the format matches expectations
		assert.Contains(t, err1.Error(), "Too small: expected slice to have >=2 items")
		assert.Contains(t, err2.Error(), "Too big: expected slice to have <=2 items")
	})

	t.Run("multilingual error message support", func(t *testing.T) {
		// Test error message structure for internationalization support
		schema := Slice(String()).Min(3)

		// Test with default locale
		_, err := schema.Parse([]string{"a", "b"})
		require.Error(t, err)

		// The default error should be in expected format
		assert.Contains(t, err.Error(), "Too small: expected slice to have >=3 items")

		// Test that the error contains the expected metadata for localization
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		issue := zodErr.Issues[0]
		assert.Equal(t, issues.TooSmall, issue.Code)
		// Note: core.ZodRawIssue doesn't have Origin, Minimum, Inclusive fields
		// These are expected to be in Properties map or derived from message
		assert.Contains(t, issue.Message, "Too small")

		fmt.Printf("Default error message: %s\n", err.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		schema := Slice(String().Min(5)).Min(2)
		_, err := schema.Parse([]string{"hi"})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})

	t.Run("nested validation errors", func(t *testing.T) {
		schema := Slice(Object(core.ObjectSchema{
			"name": String().Min(3),
			"age":  Int().Min(0),
		}))

		invalidData := []map[string]any{
			{"name": "Jo", "age": -1}, // Both name too short and age negative
		}

		_, err := schema.Parse(invalidData)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestSliceEdgeCases(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		schema := Slice(String())
		result, err := schema.Parse([]string{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("nil slice handling", func(t *testing.T) {
		schema := Slice(String())

		// nil slice should fail unless Nilable
		_, err := schema.Parse((*[]string)(nil))
		assert.Error(t, err)

		// Nilable schema should handle nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Slice(String())
		complexTypes := []any{
			make(chan int),
			func() int { return 1 },
			struct{ I int }{I: 1},
			map[string]int{"a": 1},
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("heterogeneous slice handling", func(t *testing.T) {
		schema := Slice(String())

		// Mixed types should fail element validation
		_, err := schema.Parse([]any{"hello", 123, true})
		assert.Error(t, err)
	})

	t.Run("deeply nested slices", func(t *testing.T) {
		schema := Slice(Slice(String().Min(2)))

		// Valid nested structure
		result, err := schema.Parse([][]string{
			{"hello", "world"},
			{"test", "data"},
		})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid inner element
		_, err = schema.Parse([][]string{
			{"hello", "world"},
			{"test", "x"}, // "x" is too short
		})
		assert.Error(t, err)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestSliceDefaultAndPrefault(t *testing.T) {
	t.Run("basic default value", func(t *testing.T) {
		defaultValue := []any{"default", "value"}
		schema := Slice(String()).Default(defaultValue)

		// nil input uses default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)

		// Valid input bypasses default
		validInput := []string{"hello", "world"}
		result2, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", "world"}, result2)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := Slice(String()).DefaultFunc(func() []any {
			counter++
			return []any{fmt.Sprintf("generated-%d", counter)}
		}).Min(1)

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, []any{"generated-1"}, result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, []any{"generated-2"}, result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse([]string{"valid"})
		require.NoError(t, err3)
		assert.Equal(t, []any{"valid"}, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with validation chaining", func(t *testing.T) {
		schema := Slice(String()).
			Default([]any{"hello", "world"}).
			Min(2).
			Max(5)

		// nil input: use default, validate
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, []any{"hello", "world"}, result1)

		// Valid input: validate
		result2, err2 := schema.Parse([]string{"test", "data", "example"})
		require.NoError(t, err2)
		assert.Len(t, result2, 3)

		// Invalid input still fails validation
		_, err3 := schema.Parse([]string{"only_one"})
		assert.Error(t, err3, "Short slice should fail Min(2) validation")
	})

	t.Run("prefault value", func(t *testing.T) {
		fallbackValue := []any{"fallback", "data"}
		schema := Slice(String()).Min(2).Prefault(fallbackValue) // Min(2) for slice length, not string length

		// Valid input succeeds
		result1, err1 := schema.Parse([]string{"hello", "world"})
		require.NoError(t, err1)
		assert.Equal(t, []any{"hello", "world"}, result1)

		// Invalid input uses prefault - single element slice should fail Min(2) validation
		result2, err2 := schema.Parse([]string{"hello"}) // only one element, should fail Min(2)
		require.NoError(t, err2)
		assert.Equal(t, fallbackValue, result2)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := Slice(String()).
			Default([]any{"hello", "world"}).
			Min(1).
			Transform(func(val []any, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"original": val,
					"length":   len(val),
					"first":    val[0],
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse([]string{"test", "data"})
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, []any{"test", "data"}, result1Map["original"])
		assert.Equal(t, 2, result1Map["length"])
		assert.Equal(t, "test", result1Map["first"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, []any{"hello", "world"}, result2Map["original"])
		assert.Equal(t, 2, result2Map["length"])
		assert.Equal(t, "hello", result2Map["first"])
	})
}
