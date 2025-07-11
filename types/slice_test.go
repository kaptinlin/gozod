package types

import (
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlice_BasicGeneric(t *testing.T) {
	t.Run("string slice basic functionality", func(t *testing.T) {
		schema := Slice[string](String())

		result, err := schema.Parse([]string{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("int slice basic functionality", func(t *testing.T) {
		schema := Slice[int](Int())

		result, err := schema.Parse([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("optional modifier returns pointer constraint", func(t *testing.T) {
		schema := Slice[string](String()).Optional()

		// Test with value
		testSlice := []string{"test"}
		result, err := schema.Parse(testSlice)
		require.NoError(t, err)
		assert.Equal(t, &testSlice, result)

		// Test with nil
		nilResult, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, nilResult)
	})

	t.Run("validation methods work", func(t *testing.T) {
		schema := Slice[string](String()).Min(2).Max(5)

		// Valid length
		result, err := schema.Parse([]string{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)

		// Too short
		_, err = schema.Parse([]string{"a"})
		assert.Error(t, err)

		// Too long
		_, err = schema.Parse([]string{"a", "b", "c", "d", "e", "f"})
		assert.Error(t, err)
	})

	t.Run("refine with type-safe functions", func(t *testing.T) {
		schema := Slice[string](String()).Refine(func(s []string) bool {
			return len(s) > 1
		})

		// Valid
		result, err := schema.Parse([]string{"a", "b"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, result)

		// Invalid
		_, err = schema.Parse([]string{"a"})
		assert.Error(t, err)
	})

	t.Run("default preserves type", func(t *testing.T) {
		defaultSlice := []string{"default"}
		schema := Slice[string](String()).Default(defaultSlice)

		// Valid input overrides default
		result, err := schema.Parse([]string{"test"})
		require.NoError(t, err)
		assert.Equal(t, []string{"test"}, result)
	})

	t.Run("element validation", func(t *testing.T) {
		schema := Slice[string](String())

		// Valid elements
		result, err := schema.Parse([]string{"valid", "elements"})
		require.NoError(t, err)
		assert.Equal(t, []string{"valid", "elements"}, result)

		// Invalid element type - test with []any containing non-string
		_, err = schema.Parse([]any{"valid", 123})
		assert.Error(t, err)
	})

	t.Run("empty slice", func(t *testing.T) {
		schema := Slice[string](String())

		result, err := schema.Parse([]string{})
		require.NoError(t, err)
		assert.Equal(t, []string{}, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Slice[string](String())

		invalidInputs := []any{
			"not a slice", 123, true, map[string]any{"key": "value"}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})
}

func TestSlice_TypeConstraints(t *testing.T) {
	t.Run("slice with value constraint", func(t *testing.T) {
		schema := Slice[string](String())

		result, err := schema.Parse([]string{"test"})
		require.NoError(t, err)
		assert.IsType(t, []string{}, result)
		assert.Equal(t, []string{"test"}, result)
	})

	t.Run("slice with pointer constraint from Optional", func(t *testing.T) {
		schema := Slice[string](String()).Optional()

		testSlice := []string{"test"}
		result, err := schema.Parse(testSlice)
		require.NoError(t, err)
		assert.IsType(t, &[]string{}, result)
		assert.Equal(t, &testSlice, result)
	})

	t.Run("slice with pointer constraint from Nilable", func(t *testing.T) {
		schema := Slice[string](String()).Nilable()

		testSlice := []string{"test"}
		result, err := schema.Parse(testSlice)
		require.NoError(t, err)
		assert.IsType(t, &[]string{}, result)
		assert.Equal(t, &testSlice, result)

		// Test nil
		nilResult, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, nilResult)
	})
}

func TestSlice_ErrorHandling(t *testing.T) {
	t.Run("custom error message", func(t *testing.T) {
		schema := Slice[string](String(), core.SchemaParams{Error: "Expected a valid slice"})

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Overwrite functionality tests
// =============================================================================

func TestSlice_Overwrite(t *testing.T) {
	t.Run("basic slice transformations", func(t *testing.T) {
		// Test sorting slice
		sortSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			sorted := make([]int, len(s))
			copy(sorted, s)
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i] > sorted[j] {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			return sorted
		})

		input := []int{3, 1, 4, 1, 5, 9, 2, 6}
		result, err := sortSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 1, 2, 3, 4, 5, 6, 9}, result)
	})

	t.Run("string slice transformations", func(t *testing.T) {
		// Test trimming all strings
		trimSchema := Slice[string](String()).Overwrite(func(s []string) []string {
			trimmed := make([]string, len(s))
			for i, str := range s {
				trimmed[i] = strings.TrimSpace(str)
			}
			return trimmed
		})

		input := []string{"  hello  ", "\tworld\n", " go ", ""}
		result, err := trimSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []string{"hello", "world", "go", ""}, result)

		// Test converting to lowercase
		lowerSchema := Slice[string](String()).Overwrite(func(s []string) []string {
			lower := make([]string, len(s))
			for i, str := range s {
				lower[i] = strings.ToLower(str)
			}
			return lower
		})

		input = []string{"HELLO", "World", "GO"}
		result, err = lowerSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []string{"hello", "world", "go"}, result)
	})

	t.Run("filtering and deduplication", func(t *testing.T) {
		// Test removing duplicates
		dedupeSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			seen := make(map[int]bool)
			result := []int{}
			for _, v := range s {
				if !seen[v] {
					seen[v] = true
					result = append(result, v)
				}
			}
			return result
		})

		input := []int{1, 2, 2, 3, 1, 4, 3, 5}
		result, err := dedupeSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result)

		// Test filtering even numbers
		evenSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			evens := []int{}
			for _, v := range s {
				if v%2 == 0 {
					evens = append(evens, v)
				}
			}
			return evens
		})

		input = []int{1, 2, 3, 4, 5, 6, 7, 8}
		result, err = evenSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []int{2, 4, 6, 8}, result)
	})

	t.Run("mathematical transformations", func(t *testing.T) {
		// Test doubling all values
		doubleSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			doubled := make([]int, len(s))
			for i, v := range s {
				doubled[i] = v * 2
			}
			return doubled
		})

		input := []int{1, 2, 3, 4, 5}
		result, err := doubleSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []int{2, 4, 6, 8, 10}, result)

		// Test calculating cumulative sum
		cumulativeSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			cumulative := make([]int, len(s))
			sum := 0
			for i, v := range s {
				sum += v
				cumulative[i] = sum
			}
			return cumulative
		})

		input = []int{1, 2, 3, 4, 5}
		result, err = cumulativeSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 3, 6, 10, 15}, result)
	})

	t.Run("structural transformations", func(t *testing.T) {
		// Test reversing slice
		reverseSchema := Slice[string](String()).Overwrite(func(s []string) []string {
			reversed := make([]string, len(s))
			for i, v := range s {
				reversed[len(s)-1-i] = v
			}
			return reversed
		})

		input := []string{"a", "b", "c", "d", "e"}
		result, err := reverseSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []string{"e", "d", "c", "b", "a"}, result)

		// Test padding slice to minimum length
		padSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			minLen := 5
			if len(s) >= minLen {
				return s
			}
			padded := make([]int, minLen)
			copy(padded, s)
			// Fill remaining with zeros (default)
			return padded
		})

		intInput := []int{1, 2, 3}
		intResult, err := padSchema.Parse(intInput)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 0, 0}, intResult)
	})

	t.Run("chaining with validations", func(t *testing.T) {
		// Test chaining Overwrite with Min validation
		sortedMinSchema := Slice[int](Int()).
			Overwrite(func(s []int) []int {
				// Sort first
				sorted := make([]int, len(s))
				copy(sorted, s)
				for i := 0; i < len(sorted)-1; i++ {
					for j := i + 1; j < len(sorted); j++ {
						if sorted[i] > sorted[j] {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					}
				}
				return sorted
			}).
			Min(3) // Then require at least 3 elements

		// Test with valid input
		result, err := sortedMinSchema.Parse([]int{3, 1, 4, 1, 5})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 1, 3, 4, 5}, result)

		// Test with invalid input (too short)
		_, err = sortedMinSchema.Parse([]int{3, 1})
		assert.Error(t, err)
	})

	t.Run("multiple overwrite calls", func(t *testing.T) {
		// Test chaining multiple Overwrite calls
		multiTransformSchema := Slice[int](Int()).
			Overwrite(func(s []int) []int {
				// First: remove zeros
				filtered := []int{}
				for _, v := range s {
					if v != 0 {
						filtered = append(filtered, v)
					}
				}
				return filtered
			}).
			Overwrite(func(s []int) []int {
				// Second: double all values
				doubled := make([]int, len(s))
				for i, v := range s {
					doubled[i] = v * 2
				}
				return doubled
			}).
			Overwrite(func(s []int) []int {
				// Third: sort
				sorted := make([]int, len(s))
				copy(sorted, s)
				for i := 0; i < len(sorted)-1; i++ {
					for j := i + 1; j < len(sorted); j++ {
						if sorted[i] > sorted[j] {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					}
				}
				return sorted
			})

		// Test: [3, 0, 1, 0, 2] -> [3, 1, 2] -> [6, 2, 4] -> [2, 4, 6]
		result, err := multiTransformSchema.Parse([]int{3, 0, 1, 0, 2})
		require.NoError(t, err)
		assert.Equal(t, []int{2, 4, 6}, result)
	})

	t.Run("empty slice handling", func(t *testing.T) {
		// Test transformation that handles empty slices
		safeSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			if len(s) == 0 {
				return []int{0} // Provide default
			}
			return s
		})

		// Test empty slice
		result, err := safeSchema.Parse([]int{})
		require.NoError(t, err)
		assert.Equal(t, []int{0}, result)

		// Test non-empty slice
		result, err = safeSchema.Parse([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("type preservation", func(t *testing.T) {
		// Test that Overwrite preserves the original type
		sliceSchema := Slice[int](Int())
		overwriteSchema := sliceSchema.Overwrite(func(s []int) []int {
			// Sort the slice
			sorted := make([]int, len(s))
			copy(sorted, s)
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i] > sorted[j] {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			return sorted
		})

		// Both should have the same type
		testValue := []int{3, 1, 4}

		result1, err1 := sliceSchema.Parse(testValue)
		require.NoError(t, err1)

		result2, err2 := overwriteSchema.Parse(testValue)
		require.NoError(t, err2)

		// Both results should be of type []int
		assert.IsType(t, []int{}, result1)
		assert.IsType(t, []int{}, result2)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		// Pointer Overwrite should now work and preserve pointer identity
		ptrSchema := SlicePtr[int](Int()).Nilable().Overwrite(func(s *[]int) *[]int {
			if s == nil {
				return nil
			}
			// Sort the slice
			sorted := make([]int, len(*s))
			copy(sorted, *s)
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i] > sorted[j] {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			return &sorted
		})

		// Test with non-nil value
		testValue := []int{3, 1, 4}
		result, err := ptrSchema.Parse(&testValue)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []int{1, 3, 4}, *result)

		// Test with nil value
		result, err = ptrSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("error handling", func(t *testing.T) {
		// Test that invalid inputs still produce errors
		schema := Slice[int](Int()).Overwrite(func(s []int) []int {
			return s // Identity transformation
		})

		// Invalid input should still cause an error
		_, err := schema.Parse("not a slice")
		assert.Error(t, err)

		_, err = schema.Parse(12345)
		assert.Error(t, err)

		_, err = schema.Parse(map[string]int{"a": 1})
		assert.Error(t, err)
	})

	t.Run("immutability", func(t *testing.T) {
		// Test that original values are not modified
		originalSlice := []int{3, 1, 4, 1, 5}
		originalCopy := make([]int, len(originalSlice))
		copy(originalCopy, originalSlice)

		sortSchema := Slice[int](Int()).Overwrite(func(s []int) []int {
			sorted := make([]int, len(s))
			copy(sorted, s)
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i] > sorted[j] {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			return sorted
		})

		result, err := sortSchema.Parse(originalSlice)
		require.NoError(t, err)

		// Result should be sorted
		assert.Equal(t, []int{1, 1, 3, 4, 5}, result)

		// Original slice should remain unchanged
		assert.Equal(t, originalCopy, originalSlice)
	})

	t.Run("complex element transformations", func(t *testing.T) {
		// Test slice of custom struct types (using map as example)
		type Person struct {
			Name string
			Age  int
		}

		// Create a slice transformation that normalizes names
		normalizeSchema := Slice[Person](Any()).Overwrite(func(people []Person) []Person {
			normalized := make([]Person, len(people))
			for i, person := range people {
				normalized[i] = Person{
					Name: strings.TrimSpace(strings.ToLower(person.Name)),
					Age:  person.Age,
				}
			}
			return normalized
		})

		input := []Person{
			{Name: "  ALICE  ", Age: 30},
			{Name: "Bob\n", Age: 25},
			{Name: "\tCharlie ", Age: 35},
		}

		result, err := normalizeSchema.Parse(input)
		require.NoError(t, err)

		expected := []Person{
			{Name: "alice", Age: 30},
			{Name: "bob", Age: 25},
			{Name: "charlie", Age: 35},
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestSlice_Check(t *testing.T) {
	t.Run("adds multiple issues for invalid input", func(t *testing.T) {
		schema := Slice[int](nil).Check(func(value []int, p *core.ParsePayload) {
			if len(value) == 0 {
				p.AddIssueWithMessage("slice cannot be empty")
			}
			if len(value) > 3 {
				p.AddIssueWithCode(core.TooBig, "too many elements")
			}
		})

		_, err := schema.Parse([]int{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)

		_, err = schema.Parse([]int{1, 2, 3, 4, 5})
		require.Error(t, err)
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Equal(t, core.TooBig, zErr.Issues[0].Code)
	})

	t.Run("succeeds for valid input", func(t *testing.T) {
		schema := Slice[int](nil).Check(func(value []int, p *core.ParsePayload) {
			if len(value)%2 != 0 {
				p.AddIssueWithMessage("length must be even")
			}
		})
		res, err := schema.Parse([]int{1, 2})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, res)
	})

	t.Run("works with pointer types", func(t *testing.T) {
		schema := SlicePtr[int](nil).Check(func(value *[]int, p *core.ParsePayload) {
			if value != nil && len(*value) < 2 {
				p.AddIssueWithMessage("need at least two elements")
			}
		})
		small := []int{1}
		_, err := schema.Parse(&small)
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)

		good := []int{1, 2, 3}
		res, err := schema.Parse(&good)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, 3, len(*res))
	})
}
