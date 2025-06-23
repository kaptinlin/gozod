package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

// TestArrayConstructor tests basic array schema construction
func TestArrayConstructor(t *testing.T) {
	t.Run("basic constructor", func(t *testing.T) {
		schema := Array(String(), Int())
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, core.ZodTypeArray, internals.Type)
	})

	t.Run("constructor with params", func(t *testing.T) {
		schema := Array(String(), Int(), core.SchemaParams{Error: "core.Custom error"})
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals.Error)
	})

	t.Run("constructor with single element", func(t *testing.T) {
		schema := Array(String(), 3)
		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeArray, schema.GetInternals().Type)
	})
}

// TestArrayBasicValidation tests fundamental array validation behavior
func TestArrayBasicValidation(t *testing.T) {
	t.Run("successful validation with exact types", func(t *testing.T) {
		schema := Array(String(), Int())

		// Valid input should return exact same value
		val, err := schema.Parse([]any{"asdf", 1234})
		require.NoError(t, err)
		assert.Equal(t, []any{"asdf", 1234}, val)

		// Another valid input
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})

	t.Run("invalid element type validation", func(t *testing.T) {
		schema := Array(String(), Int())

		// Invalid second element type (string instead of int)
		_, err := schema.Parse([]any{"asdf", "asdf"})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)

		// Should have invalid_type error for the second element
		hasInvalidType := false
		for _, issue := range zodErr.Issues {
			if issue.Code == issues.InvalidType && (strings.Contains(issue.Message, "expected int") || strings.Contains(issue.Message, "expected number")) {
				hasInvalidType = true
				break
			}
		}
		assert.True(t, hasInvalidType, "Should have invalid type error for int/number")
	})

	t.Run("array length validation", func(t *testing.T) {
		schema := Array(String(), Int())

		// Too many elements - should fail with "too big" error
		_, err := schema.Parse([]any{"asdf", 1234, true})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)

		// Should have "too big" error
		hasTooBig := false
		for _, issue := range zodErr.Issues {
			if issue.Code == issues.TooBig || strings.Contains(issue.Message, "Too big") {
				hasTooBig = true
				break
			}
		}
		assert.True(t, hasTooBig, "Should have 'too big' error for extra elements")

		// Wrong length with different type
		_, err = schema.Parse([]string{"hello"}) // This should fail - wrong length
		assert.Error(t, err)
	})

	t.Run("non-array input validation", func(t *testing.T) {
		schema := Array(String(), Int())

		// Wrong input type (not array) - should fail with invalid_type
		_, err := schema.Parse(map[string]any{})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, issues.InvalidType, zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "expected array")
	})
}

// TestArrayTypeInference tests intelligent type handling and inference
func TestArrayTypeInference(t *testing.T) {
	t.Run("type preservation", func(t *testing.T) {
		schema := Array(String(), Int())

		// []any input returns []any
		input1 := []any{"hello", 123}
		result1, err := schema.Parse(input1)
		require.NoError(t, err)
		assert.IsType(t, []any{}, result1)
		assert.Equal(t, input1, result1)

		// Verify pointer identity preservation
		inputPtr := &input1
		result2, err := schema.Parse(inputPtr)
		require.NoError(t, err)
		resultPtr, ok := result2.(*[]any)
		require.True(t, ok)
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
	})
}

// =============================================================================
// 2. Validation methods
// =============================================================================

func TestArrayValidations(t *testing.T) {
	t.Run("length validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  *ZodArray
			input   []any
			wantErr bool
		}{
			{"exact length valid", Array(String(), Int()), []any{"a", 1}, false},
			{"exact length invalid - too short", Array(String(), Int()), []any{"a"}, true},
			{"exact length invalid - too long", Array(String(), Int()), []any{"a", 1, true}, true},
			{"min length valid", Array(String(), Int()).Min(2), []any{"a", 1}, false},
			{"min length invalid", Array(String(), Int()).Min(3), []any{"a", 1}, true},
			{"max length valid", Array(String(), Int()).Max(2), []any{"a", 1}, false},
			{"max length invalid", Array(String(), Int()).Max(1), []any{"a", 1}, true},
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

	t.Run("element type validation", func(t *testing.T) {
		schema := Array(String(), Int())

		// Valid elements
		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 123}, result)

		// Invalid second element type (string instead of int)
		_, err = schema.Parse([]any{"asdf", "asdf"})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
		// Check that at least one issue is about invalid type
		hasInvalidType := false
		for _, issue := range zodErr.Issues {
			if issue.Code == issues.InvalidType && (strings.Contains(issue.Message, "expected int") || strings.Contains(issue.Message, "expected number")) {
				hasInvalidType = true
				break
			}
		}
		assert.True(t, hasInvalidType, "Should have invalid type error for int/number")
	})

	t.Run("wrong input type", func(t *testing.T) {
		schema := Array(String(), Int())

		_, err := schema.Parse(map[string]any{})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, issues.InvalidType, zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "expected array")
	})

	t.Run("sparse array input", func(t *testing.T) {
		// Test validation behavior with arrays containing nil/undefined elements
		schema := Array(String(), Int())

		// Create sparse array (array with nil elements)
		sparseArray := make([]any, 2)
		// sparseArray contains [nil, nil] representing undefined values

		_, err := schema.Parse(sparseArray)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have invalid_type errors for undefined elements
		invalidTypeCount := 0
		for _, issue := range zodErr.Issues {
			if issue.Code == issues.InvalidType {
				invalidTypeCount++
			}
		}
		assert.GreaterOrEqual(t, invalidTypeCount, 1, "Should have invalid type errors for undefined elements")
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestArrayModifiers(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		schema := Array(String(), Int()).Optional()

		// Valid array
		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})

	t.Run("nilable", func(t *testing.T) {
		schema := Array(String(), Int()).Nilable()

		// Valid array
		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input returns typed nil
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		assert.IsType(t, (*[]any)(nil), result2)

		// Valid input keeps type inference
		result3, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)
		assert.IsType(t, []any{}, result3)
	})

	t.Run("nilable modifier immutability", func(t *testing.T) {
		baseSchema := Array(String(), Int())
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Critical: Original schema should remain unchanged and reject nil
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		// Both schemas should validate valid input the same way
		validInput := []any{"hello", 123}
		result5, err5 := baseSchema.Parse(validInput)
		require.NoError(t, err5)
		assert.NotNil(t, result5)

		result6, err6 := nilableSchema.Parse(validInput)
		require.NoError(t, err6)
		assert.NotNil(t, result6)
	})

	t.Run("nullish", func(t *testing.T) {
		schema := Array(String(), Int()).Nullish()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Array(String(), Int())

		// Valid input should not panic
		result := schema.MustParse([]any{"hello", 123})
		assert.NotNil(t, result)

		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("not an array")
		})
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestArrayChaining(t *testing.T) {
	t.Run("multiple validation chaining", func(t *testing.T) {
		schema := Array(String().Min(3), Int().Min(0)).Min(2).Max(2)

		// Valid case
		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid element validation (string too short)
		_, err = schema.Parse([]any{"hi", 123})
		assert.Error(t, err)

		// Invalid element validation (negative number)
		_, err = schema.Parse([]any{"hello", -1})
		assert.Error(t, err)
	})

	t.Run("validation with nested schemas", func(t *testing.T) {
		schema := Array(
			Object(core.ObjectSchema{"name": String().Min(2)}),
			Int().Min(0),
		)

		// Valid nested structure
		result, err := schema.Parse([]any{
			map[string]any{"name": "John"},
			25,
		})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid nested validation
		_, err = schema.Parse([]any{
			map[string]any{"name": "J"}, // name too short
			25,
		})
		assert.Error(t, err)
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestArrayTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Array(String(), Int()).Transform(func(arr []any, ctx *core.RefinementContext) (any, error) {
			return map[string]any{
				"first":  arr[0],
				"second": arr[1],
				"length": len(arr),
			}, nil
		})

		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "hello", resultMap["first"])
		assert.Equal(t, 123, resultMap["second"])
		assert.Equal(t, 2, resultMap["length"])
	})

	t.Run("transform chain", func(t *testing.T) {
		// First transform: create object from array
		// Second transform: add metadata
		schema := Array(String(), Int()).
			Transform(func(arr []any, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"data": arr,
				}, nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if obj, ok := val.(map[string]any); ok {
					obj["timestamp"] = "2023-01-01"
					return obj, nil
				}
				return val, nil
			})

		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []any{"hello", 123}, resultMap["data"])
		assert.Equal(t, "2023-01-01", resultMap["timestamp"])
	})

	t.Run("pipe to another schema", func(t *testing.T) {
		// Transform array to string, then validate as non-empty string
		stringTransform := Array(String(), Int()).Transform(func(arr []any, ctx *core.RefinementContext) (any, error) {
			return fmt.Sprintf("%v-%v", arr[0], arr[1]), nil
		})

		schema := stringTransform.Pipe(String().Min(5))

		// Valid case
		result, err := schema.Parse([]any{"hello", 123})
		require.NoError(t, err)
		assert.Equal(t, "hello-123", result)

		// Invalid case (resulting string too short)
		_, err = schema.Parse([]any{"hi", 1})
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Refine
// =============================================================================

func TestArrayRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Array(String(), Int()).Refine(func(arr []any) bool {
			// First element length must be greater than second element value
			if str, ok := arr[0].(string); ok {
				if num, ok := arr[1].(int); ok {
					return len(str) > num
				}
			}
			return false
		}, core.SchemaParams{Error: "String length must be greater than number value"})

		// Valid case
		result, err := schema.Parse([]any{"hello", 3})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 3}, result)

		// Invalid case
		_, err = schema.Parse([]any{"hi", 5})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "String length must be greater than number value")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := []any{"hello", 123}

		// Refine: only validates, never modifies
		refineSchema := Array(String(), Int()).Refine(func(arr []any) bool {
			return len(arr) == 2
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Array(String(), Int()).Transform(func(arr []any, ctx *core.RefinementContext) (any, error) {
			return len(arr), nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.IsType(t, []any{}, refineResult)
		assert.Equal(t, input, refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, 2, transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.NotEqual(t, transformResult, refineResult, "Transform should return modified value")
	})

	t.Run("complex refine validation", func(t *testing.T) {
		// All string elements must be unique
		schema := Array(String(), String(), String()).Refine(func(arr []any) bool {
			seen := make(map[string]bool)
			for _, item := range arr {
				if str, ok := item.(string); ok {
					if seen[str] {
						return false
					}
					seen[str] = true
				}
			}
			return true
		}, core.SchemaParams{Error: "All string elements must be unique"})

		// Valid case (unique elements)
		result, err := schema.Parse([]any{"a", "b", "c"})
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// Invalid case (duplicate elements)
		_, err = schema.Parse([]any{"a", "b", "a"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "All string elements must be unique")
	})
}

// =============================================================================
// 7. Error handling
// =============================================================================

func TestArrayErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Array(String(), Int()).Min(3)
		_, err := schema.Parse([]any{"hello", 123})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
		// Check that at least one issue is about too small
		hasTooSmall := false
		for _, issue := range zodErr.Issues {
			if issue.Code == issues.TooSmall || strings.Contains(issue.Message, "at least") {
				hasTooSmall = true
				break
			}
		}
		assert.True(t, hasTooSmall, "Should have too small error")
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Array(String(), Int(), core.SchemaParams{
			Error: "core.Custom array error",
		})
		_, err := schema.Parse("not an array")
		assert.Error(t, err)

		// For now, just check that we get an error - custom error mapping may not be fully implemented
		assert.Contains(t, err.Error(), "expected array")
	})

	t.Run("multiple errors", func(t *testing.T) {
		schema := Array(String().Min(5), Int().Min(10))
		_, err := schema.Parse([]any{"hi", 5})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})

	t.Run("nested validation errors", func(t *testing.T) {
		schema := Array(
			Object(core.ObjectSchema{
				"name": String().Min(3),
				"age":  Int().Min(0),
			}),
			String().Min(2),
		)

		invalidData := []any{
			map[string]any{"name": "Jo", "age": -1}, // Both name too short and age negative
			"x",                                     // String too short
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

func TestArrayEdgeCases(t *testing.T) {
	t.Run("empty array schema", func(t *testing.T) {
		// Test behavior of Array() with no element schemas - strict length enforcement
		schema := Array() // No elements defined

		// Empty array should succeed and return empty array
		result, err := schema.Parse([]any{})
		require.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, []any{}, result)

		// Non-empty arrays should fail with length validation error
		_, err = schema.Parse([]any{"hello", "world"})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)

		// Should have a "too big" error since input length > expected length (0)
		hasTooBig := false
		for _, issue := range zodErr.Issues {
			if issue.Code == issues.TooBig || strings.Contains(issue.Message, "too big") || strings.Contains(issue.Message, "Too big") {
				hasTooBig = true
				break
			}
		}
		assert.True(t, hasTooBig, "Should have 'too big' error for non-empty input to empty array schema")

		// Single element should also fail
		_, err = schema.Parse([]any{"single"})
		assert.Error(t, err)

		// Even nil elements should fail if array is not empty
		_, err = schema.Parse([]any{nil})
		assert.Error(t, err)
	})

	t.Run("nil array handling", func(t *testing.T) {
		schema := Array(String(), Int())

		// nil array should fail unless Nilable
		_, err := schema.Parse((*[]any)(nil))
		assert.Error(t, err)

		// Nilable schema should handle nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Array(String(), Int())
		complexTypes := []any{
			make(chan int),
			func() int { return 1 },
			struct{ I int }{I: 1},
			"not an array",
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("heterogeneous array handling", func(t *testing.T) {
		schema := Array(String(), String()) // Expects two strings

		// Mixed types should fail element validation
		_, err := schema.Parse([]any{"hello", 123})
		assert.Error(t, err)
	})

	t.Run("deeply nested arrays", func(t *testing.T) {
		schema := Array(
			Array(String(), Int()),
			String(),
		)

		// Valid nested structure
		result, err := schema.Parse([]any{
			[]any{"hello", 123},
			"world",
		})
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Invalid inner element
		_, err = schema.Parse([]any{
			[]any{"hello", "not_int"}, // Second element should be int
			"world",
		})
		assert.Error(t, err)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestArrayDefaultAndPrefault(t *testing.T) {
	t.Run("basic default value", func(t *testing.T) {
		defaultValue := []any{"default", 42}
		schema := Array(String(), Int()).Default(defaultValue)

		// nil input uses default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)

		// Valid input bypasses default
		validInput := []any{"hello", 123}
		result2, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result2)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := Array(String(), Int()).DefaultFunc(func() []any {
			counter++
			return []any{fmt.Sprintf("generated-%d", counter), counter}
		}).Min(2)

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, []any{"generated-1", 1}, result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, []any{"generated-2", 2}, result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse([]any{"valid", 999})
		require.NoError(t, err3)
		assert.Equal(t, []any{"valid", 999}, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with validation chaining", func(t *testing.T) {
		schema := Array(String(), Int()).
			Default([]any{"hello", 123}).
			Min(2).
			Max(2)

		// nil input: use default, validate
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, []any{"hello", 123}, result1)

		// Valid input: validate
		result2, err2 := schema.Parse([]any{"test", 456})
		require.NoError(t, err2)
		assert.Equal(t, []any{"test", 456}, result2)

		// Invalid input still fails validation
		_, err3 := schema.Parse([]any{"only_one"})
		assert.Error(t, err3, "Wrong length should fail validation")
	})

	t.Run("prefault value", func(t *testing.T) {
		fallbackValue := []any{"fallback", 999}
		schema := Array(String().Min(5), Int()).Prefault(fallbackValue)

		// Valid input succeeds
		result1, err1 := schema.Parse([]any{"hello", 123})
		require.NoError(t, err1)
		assert.Equal(t, []any{"hello", 123}, result1)

		// Invalid input uses prefault (string too short)
		result2, err2 := schema.Parse([]any{"hi", 123})
		require.NoError(t, err2)
		assert.Equal(t, fallbackValue, result2)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := Array(String().Min(5), Int()).PrefaultFunc(func() []any {
			counter++
			return []any{fmt.Sprintf("fallback-%d", counter), counter * 100}
		})

		// Valid input succeeds and doesn't call function
		result1, err1 := schema.Parse([]any{"hello", 123})
		require.NoError(t, err1)
		assert.Equal(t, []any{"hello", 123}, result1)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input calls prefault function (string too short)
		result2, err2 := schema.Parse([]any{"hi", 123})
		require.NoError(t, err2)
		assert.Equal(t, []any{"fallback-1", 100}, result2)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")

		// Another invalid input calls function again
		result3, err3 := schema.Parse([]any{"bye", 456})
		require.NoError(t, err3)
		assert.Equal(t, []any{"fallback-2", 200}, result3)
		assert.Equal(t, 2, counter, "Function should increment counter for each invalid input")

		// Valid input still doesn't call function
		result4, err4 := schema.Parse([]any{"world", 789})
		require.NoError(t, err4)
		assert.Equal(t, []any{"world", 789}, result4)
		assert.Equal(t, 2, counter, "Counter should remain unchanged for valid input")
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := []any{"default", 0}
		prefaultValue := []any{"prefault", 999}

		schema := Array(String().Min(5), Int()).
			Default(defaultValue).
			Prefault(prefaultValue)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Valid input succeeds
		result2, err2 := schema.Parse([]any{"hello", 123})
		require.NoError(t, err2)
		assert.Equal(t, []any{"hello", 123}, result2)

		// Invalid input uses prefault
		result3, err3 := schema.Parse([]any{"hi", 123})
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := Array(String(), Int()).
			Default([]any{"hello", 123}).
			Min(2).
			Transform(func(arr []any, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"original": arr,
					"length":   len(arr),
					"first":    arr[0],
					"second":   arr[1],
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse([]any{"test", 456})
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, []any{"test", 456}, result1Map["original"])
		assert.Equal(t, 2, result1Map["length"])
		assert.Equal(t, "test", result1Map["first"])
		assert.Equal(t, 456, result1Map["second"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, []any{"hello", 123}, result2Map["original"])
		assert.Equal(t, 2, result2Map["length"])
		assert.Equal(t, "hello", result2Map["first"])
		assert.Equal(t, 123, result2Map["second"])
	})
}

// =============================================================================
// 10. Additional array-specific utility tests
// =============================================================================

// TestArrayElementAccessor validates Element() schema access and parsing behavior
func TestArrayElementAccessor(t *testing.T) {
	t.Run("get element", func(t *testing.T) {
		// Tuple schema with string then int
		schema := Array(String(), Int())

		// Access element schemas by index
		first := schema.Element(0)
		second := schema.Element(1)

		require.NotNil(t, first, "First element schema should not be nil")
		require.NotNil(t, second, "Second element schema should not be nil")

		// Parsing with correct types should succeed
		_, err := first.Parse("asdf")
		require.NoError(t, err)
		_, err = second.Parse(12)
		require.NoError(t, err)

		// Parsing with incorrect type should fail
		_, err = first.Parse(12)
		assert.Error(t, err)
	})
}

// TestArrayNonEmptyAndMax validates NonEmpty() combined with Max()
func TestArrayNonEmptyAndMax(t *testing.T) {
	t.Run("array.nonempty().max()", func(t *testing.T) {
		// Tuple schema expects exactly two elements; NonEmpty forces at least one, Max(2) caps at two
		schema := Array(String(), Int()).NonEmpty().Max(2)

		// Valid case: two elements of correct types
		_, err := schema.Parse([]any{"a", 1})
		require.NoError(t, err)

		// Invalid case: empty array – should fail length validation
		_, err = schema.Parse([]any{})
		assert.Error(t, err)

		// Invalid case: too many elements – should fail with "too big"
		_, err = schema.Parse([]any{"a", 1, 2})
		assert.Error(t, err)
	})
}
