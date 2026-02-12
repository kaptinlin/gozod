package types

import (
	"sort"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Fixed-length Array Tests (without rest parameters)
// =============================================================================

func TestArray_FixedLength(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		schema := Array()

		result, err := schema.Parse([]any{})
		require.NoError(t, err)
		assert.Equal(t, []any{}, result)

		// Non-empty should fail
		_, err = schema.Parse([]any{"something"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected exactly 0")
	})

	t.Run("single element", func(t *testing.T) {
		schema := Array([]any{String()})

		result, err := schema.Parse([]any{"hello"})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello"}, result)

		// Wrong length
		_, err = schema.Parse([]any{})
		assert.Error(t, err)
		_, err = schema.Parse([]any{"hello", "world"})
		assert.Error(t, err)
	})

	t.Run("multiple elements", func(t *testing.T) {
		schema := Array([]any{String(), Int(), Bool()})

		result, err := schema.Parse([]any{"hello", 42, true})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42, true}, result)

		// Wrong length
		_, err = schema.Parse([]any{"hello", 42})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected exactly 3")

		// Wrong types
		_, err = schema.Parse([]any{42, "hello", true})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "element at index 0")
	})

	t.Run("nested arrays", func(t *testing.T) {
		inner := Array([]any{String(), Int()})
		schema := Array([]any{inner, inner})

		validData := []any{
			[]any{"hello", 1},
			[]any{"world", 2},
		}
		result, err := schema.Parse(validData)
		require.NoError(t, err)
		assert.Equal(t, validData, result)
	})

	t.Run("explicit syntax", func(t *testing.T) {
		// Explicit syntax: Array([]any{String(), Int(), Bool()})
		schema := Array([]any{String(), Int(), Bool()})

		result, err := schema.Parse([]any{"hello", 42, true})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42, true}, result)
	})
}

// =============================================================================
// Rest Parameter Tests
// =============================================================================

func TestArray_RestParameters(t *testing.T) {
	t.Run("basic rest validation", func(t *testing.T) {
		// [string, int, ...boolean] - first two fixed, rest are booleans
		schema := Array([]any{String(), Int()}, Bool())

		// Minimum length (exactly fixed items)
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)

		// With rest elements
		result, err = schema.Parse([]any{"hello", 42, true, false, true})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42, true, false, true}, result)

		// Too few elements (less than fixed)
		_, err = schema.Parse([]any{"hello"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected at least 2")

		// Wrong type in fixed position
		_, err = schema.Parse([]any{42, "hello", true})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "element at index 0")

		// Wrong type in rest position
		_, err = schema.Parse([]any{"hello", 42, "not_bool"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rest element at index 2")
	})

	t.Run("empty fixed with rest", func(t *testing.T) {
		// [...string] - all elements are strings
		schema := Array([]any{}, String())

		// Empty array
		result, err := schema.Parse([]any{})
		require.NoError(t, err)
		assert.Equal(t, []any{}, result)

		// All strings
		result, err = schema.Parse([]any{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, []any{"a", "b", "c"}, result)

		// Wrong type
		_, err = schema.Parse([]any{"a", 42, "c"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rest element at index 1")
	})

	t.Run("complex rest validation", func(t *testing.T) {
		// [string, int, ...Array([]any{string, int})] - pairs after first two
		innerSchema := Array([]any{String(), Int()})
		schema := Array([]any{String(), Int()}, innerSchema)

		// Just fixed elements
		result, err := schema.Parse([]any{"name", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"name", 42}, result)

		// With rest elements (nested arrays)
		validData := []any{
			"name", 42,
			[]any{"hello", 1},
			[]any{"world", 2},
		}
		result, err = schema.Parse(validData)
		require.NoError(t, err)
		assert.Equal(t, validData, result)

		// Invalid rest element
		_, err = schema.Parse([]any{"name", 42, []any{"hello", "not_int"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rest element at index 2")
	})

	t.Run("rest with validation constraints", func(t *testing.T) {
		// [string, ...int.min(10)] - integers >= 10 for rest
		schema := Array([]any{String()}, Int().Min(10))

		result, err := schema.Parse([]any{"hello", 15, 20, 100})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 15, 20, 100}, result)

		// Rest element fails validation
		_, err = schema.Parse([]any{"hello", 15, 5, 20})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rest element at index 2")
	})

	t.Run("standard rest parameter syntax", func(t *testing.T) {
		// Standard rest syntax: Array([]any{String(), Int()}, Bool())
		schema := Array([]any{String(), Int()}, Bool())

		result, err := schema.Parse([]any{"hello", 42, true, false})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42, true, false}, result)
	})
}

// =============================================================================
// Type-specific Methods Tests
// =============================================================================

func TestArray_TypeSpecificMethods(t *testing.T) {
	t.Run("Element method", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()
		boolSchema := Bool()
		schema := Array([]any{stringSchema, intSchema, boolSchema})

		assert.Equal(t, stringSchema, schema.Element(0))
		assert.Equal(t, intSchema, schema.Element(1))
		assert.Equal(t, boolSchema, schema.Element(2))
		assert.Nil(t, schema.Element(3))
		assert.Nil(t, schema.Element(-1))
	})

	t.Run("Items method", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()
		schema := Array([]any{stringSchema, intSchema})
		items := schema.ElementSchemas()

		assert.Len(t, items, 2)
		assert.Equal(t, stringSchema, items[0])
		assert.Equal(t, intSchema, items[1])
	})
}

// =============================================================================
// Validation Methods Tests
// =============================================================================

func TestArray_ValidationMethods(t *testing.T) {
	t.Run("Min validation", func(t *testing.T) {
		schema := Array([]any{String()}).Min(3)

		// Fixed-length takes precedence over Min
		_, err := schema.Parse([]any{"hello"})
		assert.Error(t, err) // Expects exactly 1, not >= 3
	})

	t.Run("Max validation", func(t *testing.T) {
		schema := Array([]any{String(), Int()}).Max(1)

		// Max validation conflicts with fixed-length requirement
		_, err := schema.Parse([]any{"hello", 42})
		assert.Error(t, err) // Expects exactly 2, but max is 1
	})

	t.Run("Length validation", func(t *testing.T) {
		schema := Array([]any{String(), Int()}).Length(2)

		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})

	t.Run("NonEmpty validation", func(t *testing.T) {
		schema := Array([]any{String()}).NonEmpty()

		result, err := schema.Parse([]any{"hello"})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello"}, result)
	})
}

// =============================================================================
// Modifier Methods Tests
// =============================================================================

func TestArray_Modifiers(t *testing.T) {
	t.Run("Optional modifier", func(t *testing.T) {
		schema := Array([]any{String(), Int()}).Optional()

		// Valid value
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []any{"hello", 42}, *result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable modifier", func(t *testing.T) {
		schema := Array([]any{String(), Int()}).Nilable()

		// Valid value
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []any{"hello", 42}, *result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default modifier", func(t *testing.T) {
		defaultValue := []any{"default", 1}
		schema := Array([]any{String(), Int()}).Default(defaultValue)

		result, err := schema.Parse([]any{"test", 2})
		require.NoError(t, err)
		assert.Equal(t, []any{"test", 2}, result)
	})
}

// =============================================================================
// Refine Tests
// =============================================================================

func TestArray_Refine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		// Only accept arrays where first string has length > 3
		schema := Array([]any{String(), Int()}).Refine(func(arr []any) bool {
			if len(arr) > 0 {
				if str, ok := arr[0].(string); ok {
					return len(str) > 3
				}
			}
			return false
		})

		// Valid
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)

		// Invalid (string too short)
		_, err = schema.Parse([]any{"hi", 42})
		assert.Error(t, err)
	})

	t.Run("refine with rest parameters", func(t *testing.T) {
		// Only accept if all rest elements are even numbers
		schema := Array([]any{String()}, Int()).Refine(func(arr []any) bool {
			for i := 1; i < len(arr); i++ { // Skip first element (string)
				if num, ok := arr[i].(int); ok {
					if num%2 != 0 {
						return false
					}
				}
			}
			return true
		})

		// Valid (all rest elements are even)
		result, err := schema.Parse([]any{"hello", 2, 4, 6})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 2, 4, 6}, result)

		// Invalid (contains odd number)
		_, err = schema.Parse([]any{"hello", 2, 3, 4})
		assert.Error(t, err)
	})
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestArray_ErrorHandling(t *testing.T) {
	t.Run("invalid input type", func(t *testing.T) {
		schema := Array([]any{String(), Int()})

		invalidInputs := []any{"string", 123, true, map[string]any{}}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Array([]any{String(), Int()}, core.SchemaParams{Error: "Custom array error"})

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("MustParse panic", func(t *testing.T) {
		schema := Array([]any{String(), Int()})

		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func TestArray_EdgeCases(t *testing.T) {
	t.Run("nil schemas in array", func(t *testing.T) {
		// nil schema should allow any value at that position
		schema := Array([]any{String(), nil, Bool()})

		result, err := schema.Parse([]any{"hello", "anything", true})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", "anything", true}, result)
	})

	t.Run("pointer input handling", func(t *testing.T) {
		schema := Array([]any{String(), Int()})
		testArray := []any{"hello", 42}

		result, err := schema.Parse(&testArray)
		require.NoError(t, err)
		assert.Equal(t, testArray, result)
	})

	t.Run("concurrent access", func(t *testing.T) {
		schema := Array([]any{String(), Int()})
		testArray := []any{"hello", 42}

		results := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func() {
				_, err := schema.Parse(testArray)
				results <- err
			}()
		}

		for i := 0; i < 5; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("transform operations", func(t *testing.T) {
		schema := Array([]any{String(), Int()})

		transform := schema.Transform(func(arr []any, ctx *core.RefinementContext) (any, error) {
			return len(arr), nil
		})
		require.NotNil(t, transform)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestArray_Overwrite(t *testing.T) {
	t.Run("basic array transformation", func(t *testing.T) {
		// Create array schema with element validation
		schema := Array([]any{Int(), String(), Bool()}).
			Overwrite(func(arr []any) []any {
				// Double numeric values, uppercase strings, negate booleans
				result := make([]any, len(arr))
				for i, val := range arr {
					switch v := val.(type) {
					case int:
						result[i] = v * 2
					case string:
						result[i] = strings.ToUpper(v)
					case bool:
						result[i] = !v
					default:
						result[i] = val
					}
				}
				return result
			})

		input := []any{5, "hello", true}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := []any{10, "HELLO", false}
		assert.Equal(t, expected, result)
	})

	t.Run("array sorting transformation", func(t *testing.T) {
		// Create array of integers and sort them
		schema := Array([]any{Int(), Int(), Int()}).
			Overwrite(func(arr []any) []any {
				// Convert to int slice, sort, and convert back
				intSlice := make([]int, len(arr))
				for i, val := range arr {
					if intVal, ok := val.(int); ok {
						intSlice[i] = intVal
					}
				}
				sort.Ints(intSlice)

				result := make([]any, len(intSlice))
				for i, val := range intSlice {
					result[i] = val
				}
				return result
			})

		input := []any{3, 1, 4}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := []any{1, 3, 4}
		assert.Equal(t, expected, result)
	})

	t.Run("chaining with other validations", func(t *testing.T) {
		schema := Array([]any{String(), String()}).
			Overwrite(func(arr []any) []any {
				// Trim whitespace from all strings
				result := make([]any, len(arr))
				for i, val := range arr {
					if strVal, ok := val.(string); ok {
						result[i] = strings.TrimSpace(strVal)
					} else {
						result[i] = val
					}
				}
				return result
			}).
			Length(2)

		input := []any{"  hello  ", "  world  "}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := []any{"hello", "world"}
		assert.Equal(t, expected, result)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Array([]any{String()}).
			Overwrite(func(arr []any) []any {
				return arr // Identity transformation
			})

		// Should maintain array type, not change to different type
		input := []any{"test"}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, []any{}, result)
		assert.Equal(t, input, result)
	})

	t.Run("error handling in transformation", func(t *testing.T) {
		// Use Rest() to allow variable length arrays
		schema := Array([]any{}, Int()).
			Overwrite(func(arr []any) []any {
				// Transform should not cause validation to fail
				// even if transformation logic has edge cases
				if len(arr) == 0 {
					return []any{} // Return empty array for empty input
				}
				return arr
			})

		// Test with empty array
		result, err := schema.Parse([]any{})
		require.NoError(t, err)
		assert.Equal(t, []any{}, result)

		// Test with normal array
		input := []any{42}
		result, err = schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestArray_Check(t *testing.T) {
	t.Run("adds multiple issues for invalid input", func(t *testing.T) {
		schema := Array().Check(func(value []any, p *core.ParsePayload) {
			if len(value) == 0 {
				p.AddIssueWithMessage("array cannot be empty")
			}
			if len(value) > 3 {
				p.AddIssueWithCode(core.TooBig, "too many elements")
			}
		})

		// Case 1: empty array – expect one issue
		_, err := schema.Parse([]any{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)

		// Case 2: too many elements – expect one issue
		_, err = schema.Parse([]any{1, 2, 3, 4})
		require.Error(t, err)
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})

	t.Run("works with pointer schema and value slice", func(t *testing.T) {
		schema := ArrayPtr().Check(func(value *[]any, p *core.ParsePayload) {
			if value == nil || len(*value) == 0 {
				p.AddIssueWithMessage("pointer array is empty")
			}
		})

		_, err := schema.Parse([]any{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})
}

func TestArray_NonOptional(t *testing.T) {
	schema := Array([]any{String()}).NonOptional()

	// valid
	_, err := schema.Parse([]any{"hi"})
	require.NoError(t, err)

	// nil error
	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}

	chain := Array([]any{String()}).Optional().NonOptional()
	_, err = chain.Parse(nil)
	assert.Error(t, err)
}

// Test multiple error collection (added for testing raw issue refactoring)
func TestArray_MultipleErrorCollection(t *testing.T) {
	t.Run("collects multiple element validation errors", func(t *testing.T) {
		// Create array with 3 string elements, each requiring min length 5
		schema := Array([]any{String().Min(5), String().Min(5), String().Min(5)})

		// Input with multiple validation failures
		input := []any{"hi", "ok", "bye"} // All strings are too short (< 5 chars)

		result, err := schema.Parse(input)
		require.Error(t, err)
		assert.Nil(t, result)

		// Check that we have multiple issues in the error
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 3 element validation errors (one for each element)
		assert.Len(t, zodErr.Issues, 3)

		// Check that each error has the correct path (array index)
		for i, issue := range zodErr.Issues {
			assert.Equal(t, []any{i}, issue.Path, "Issue %d should have path [%d]", i, i)
			assert.Contains(t, issue.Message, "Too small", "Issue %d should be a 'too small' error", i)
		}
	})

	t.Run("fails fast on length errors (TypeScript Zod v4 behavior)", func(t *testing.T) {
		// Create array expecting exactly 2 integers, each > 10
		schema := Array([]any{Int().Min(10), Int().Min(10)})

		// Input with wrong length - should fail fast without validating elements
		input := []any{5, 8, 15} // 3 elements (should be 2), first two are < 10

		result, err := schema.Parse(input)
		require.Error(t, err)
		assert.Nil(t, result)

		// Check that we have only length error (fail fast behavior)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have only 1 length error (fail fast on length, no element validation)
		assert.Len(t, zodErr.Issues, 1)

		// Verify it's a length error
		issue := zodErr.Issues[0]
		assert.Empty(t, issue.Path, "Length error should have empty path")
		assert.Equal(t, core.TooBig, issue.Code, "Should be a too_big error")
		assert.Contains(t, issue.Message, "expected exactly 2", "Length error message")
	})

	t.Run("collects multiple element errors when length is correct", func(t *testing.T) {
		// Create array expecting exactly 2 integers, each > 10
		schema := Array([]any{Int().Min(10), Int().Min(10)})

		// Input with correct length but element validation failures
		input := []any{5, 8} // 2 elements (correct length), but both are < 10

		result, err := schema.Parse(input)
		require.Error(t, err)
		assert.Nil(t, result)

		// Check that we have multiple element issues
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 2 element validation errors
		assert.Len(t, zodErr.Issues, 2)

		// All should be element errors with proper paths
		for i, issue := range zodErr.Issues {
			assert.Equal(t, []any{i}, issue.Path, "Issue %d should have path [%d]", i, i)
			assert.Equal(t, core.InvalidElement, issue.Code, "Issue %d should be invalid_element", i)
			assert.Contains(t, issue.Message, "Too small", "Issue %d should be a 'too small' error", i)
		}
	})
}
