package types_test

import (
	"testing"

	"github.com/kaptinlin/gozod/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BASIC FUNCTIONALITY TESTS
// =============================================================================

func TestSetBasicParse(t *testing.T) {
	t.Run("valid string set", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := map[string]struct{}{"a": {}, "b": {}, "c": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("valid int set", func(t *testing.T) {
		schema := types.Set[int](types.Int())
		input := map[int]struct{}{1: {}, 2: {}, 3: {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("empty set", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := map[string]struct{}{}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("nil input fails", func(t *testing.T) {
		schema := types.Set[string](types.String())
		_, err := schema.Parse(nil)
		require.Error(t, err)
	})
}

func TestSetStrictParse(t *testing.T) {
	t.Run("valid string set with StrictParse", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := map[string]struct{}{"x": {}, "y": {}}
		result, err := schema.StrictParse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("valid int set with StrictParse", func(t *testing.T) {
		schema := types.Set[int](types.Int())
		input := map[int]struct{}{10: {}, 20: {}}
		result, err := schema.StrictParse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

func TestSetMustParse(t *testing.T) {
	t.Run("MustParse with valid input", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := map[string]struct{}{"a": {}}
		result := schema.MustParse(input)
		assert.Equal(t, input, result)
	})

	t.Run("MustParse panics on invalid input", func(t *testing.T) {
		schema := types.Set[string](types.String())
		assert.Panics(t, func() {
			schema.MustParse(nil)
		})
	})
}

func TestSetMustStrictParse(t *testing.T) {
	t.Run("MustStrictParse with valid input", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := map[string]struct{}{"test": {}}
		result := schema.MustStrictParse(input)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// SIZE VALIDATION TESTS
// =============================================================================

func TestSetMin(t *testing.T) {
	t.Run("valid min size", func(t *testing.T) {
		schema := types.Set[string](types.String()).Min(2)
		input := map[string]struct{}{"a": {}, "b": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("invalid min size", func(t *testing.T) {
		schema := types.Set[string](types.String()).Min(3)
		input := map[string]struct{}{"a": {}, "b": {}}
		_, err := schema.Parse(input)
		require.Error(t, err)
	})
}

func TestSetMax(t *testing.T) {
	t.Run("valid max size", func(t *testing.T) {
		schema := types.Set[string](types.String()).Max(3)
		input := map[string]struct{}{"a": {}, "b": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("invalid max size", func(t *testing.T) {
		schema := types.Set[string](types.String()).Max(1)
		input := map[string]struct{}{"a": {}, "b": {}}
		_, err := schema.Parse(input)
		require.Error(t, err)
	})
}

func TestSetSize(t *testing.T) {
	t.Run("valid exact size", func(t *testing.T) {
		schema := types.Set[string](types.String()).Size(2)
		input := map[string]struct{}{"a": {}, "b": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("invalid exact size - too few", func(t *testing.T) {
		schema := types.Set[string](types.String()).Size(3)
		input := map[string]struct{}{"a": {}, "b": {}}
		_, err := schema.Parse(input)
		require.Error(t, err)
	})

	t.Run("invalid exact size - too many", func(t *testing.T) {
		schema := types.Set[string](types.String()).Size(1)
		input := map[string]struct{}{"a": {}, "b": {}}
		_, err := schema.Parse(input)
		require.Error(t, err)
	})
}

func TestSetNonEmpty(t *testing.T) {
	t.Run("valid non-empty set", func(t *testing.T) {
		schema := types.Set[string](types.String()).NonEmpty()
		input := map[string]struct{}{"a": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("invalid empty set", func(t *testing.T) {
		schema := types.Set[string](types.String()).NonEmpty()
		input := map[string]struct{}{}
		_, err := schema.Parse(input)
		require.Error(t, err)
	})
}

// =============================================================================
// MODIFIER TESTS
// =============================================================================

func TestSetOptional(t *testing.T) {
	t.Run("optional accepts nil", func(t *testing.T) {
		schema := types.Set[string](types.String()).Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("optional accepts valid set", func(t *testing.T) {
		schema := types.Set[string](types.String()).Optional()
		input := map[string]struct{}{"a": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, input, *result)
	})
}

func TestSetNilable(t *testing.T) {
	t.Run("nilable accepts nil", func(t *testing.T) {
		schema := types.Set[string](types.String()).Nilable()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestSetNullish(t *testing.T) {
	t.Run("nullish accepts nil", func(t *testing.T) {
		schema := types.Set[string](types.String()).Nullish()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestSetDefault(t *testing.T) {
	t.Run("default value on nil input", func(t *testing.T) {
		defaultValue := map[string]struct{}{"default": {}}
		schema := types.Set[string](types.String()).Default(defaultValue)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})

	t.Run("default not used with valid input", func(t *testing.T) {
		defaultValue := map[string]struct{}{"default": {}}
		schema := types.Set[string](types.String()).Default(defaultValue)
		input := map[string]struct{}{"actual": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

func TestSetDefaultFunc(t *testing.T) {
	t.Run("defaultFunc on nil input", func(t *testing.T) {
		schema := types.Set[string](types.String()).DefaultFunc(func() map[string]struct{} {
			return map[string]struct{}{"generated": {}}
		})
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Contains(t, result, "generated")
	})
}

// =============================================================================
// ELEMENT VALIDATION TESTS
// =============================================================================

func TestSetElementValidation(t *testing.T) {
	t.Run("validates string elements", func(t *testing.T) {
		schema := types.Set[string](types.String().Min(2))
		input := map[string]struct{}{"ab": {}, "cd": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("fails on invalid string element", func(t *testing.T) {
		schema := types.Set[string](types.String().Min(3))
		input := map[string]struct{}{"ab": {}, "cde": {}}
		_, err := schema.Parse(input)
		require.Error(t, err) // "ab" is too short
	})

	t.Run("validates int elements with range", func(t *testing.T) {
		schema := types.Set[int](types.Int().Min(0).Max(100))
		input := map[int]struct{}{10: {}, 50: {}, 100: {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// SLICE CONVERSION TESTS
// =============================================================================

func TestSetFromSlice(t *testing.T) {
	t.Run("convert slice to set", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := []string{"a", "b", "c"}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Contains(t, result, "a")
		assert.Contains(t, result, "b")
		assert.Contains(t, result, "c")
	})

	t.Run("convert slice to set with duplicates", func(t *testing.T) {
		schema := types.Set[string](types.String())
		input := []string{"a", "b", "a", "b"}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 2) // Duplicates removed
	})

	t.Run("convert int slice to set", func(t *testing.T) {
		schema := types.Set[int](types.Int())
		input := []int{1, 2, 3}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})
}

// =============================================================================
// REFINEMENT TESTS
// =============================================================================

func TestSetRefine(t *testing.T) {
	t.Run("refine with custom validation", func(t *testing.T) {
		schema := types.Set[string](types.String()).Refine(func(s map[string]struct{}) bool {
			// Check that set contains "required"
			_, ok := s["required"]
			return ok
		}, "set must contain 'required'")

		input := map[string]struct{}{"required": {}, "other": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("refine fails validation", func(t *testing.T) {
		schema := types.Set[string](types.String()).Refine(func(s map[string]struct{}) bool {
			_, ok := s["required"]
			return ok
		}, "set must contain 'required'")

		input := map[string]struct{}{"other": {}}
		_, err := schema.Parse(input)
		require.Error(t, err)
	})
}

// =============================================================================
// POINTER VARIANT TESTS
// =============================================================================

func TestSetPtr(t *testing.T) {
	t.Run("SetPtr with valid input", func(t *testing.T) {
		schema := types.SetPtr[string](types.String())
		input := map[string]struct{}{"a": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Contains(t, *result, "a")
	})

	t.Run("SetPtr with nil input fails", func(t *testing.T) {
		schema := types.SetPtr[string](types.String())
		_, err := schema.Parse(nil)
		require.Error(t, err) // Non-optional pointer still requires value
	})
}

// =============================================================================
// METADATA TESTS
// =============================================================================

func TestSetMeta(t *testing.T) {
	t.Run("describe adds description", func(t *testing.T) {
		schema := types.Set[string](types.String()).Describe("A set of strings")
		assert.NotNil(t, schema)
	})
}

// =============================================================================
// COMPOSITION TESTS
// =============================================================================

func TestSetComposition(t *testing.T) {
	t.Run("And composition", func(t *testing.T) {
		schema1 := types.Set[string](types.String()).Min(1)
		schema2 := types.Set[string](types.String()).Max(3)
		combined := schema1.And(schema2)
		assert.NotNil(t, combined)
	})

	t.Run("Or composition", func(t *testing.T) {
		schema1 := types.Set[string](types.String())
		schema2 := types.Set[int](types.Int())
		combined := schema1.Or(schema2)
		assert.NotNil(t, combined)
	})
}

// =============================================================================
// CHAINED METHOD TESTS
// =============================================================================

func TestSetMethodChaining(t *testing.T) {
	t.Run("chain multiple constraints", func(t *testing.T) {
		schema := types.Set[string](types.String()).
			Min(1).
			Max(5).
			NonEmpty()

		input := map[string]struct{}{"a": {}, "b": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// TYPE-SPECIFIC TESTS
// =============================================================================

func TestSetValueType(t *testing.T) {
	t.Run("ValueType returns element schema", func(t *testing.T) {
		elementSchema := types.String()
		schema := types.Set[string](elementSchema)
		assert.Equal(t, elementSchema, schema.ValueType())
	})
}

func TestSetWithNilValueSchema(t *testing.T) {
	t.Run("set with nil schema accepts any comparable values", func(t *testing.T) {
		schema := types.Set[string](nil)
		input := map[string]struct{}{"a": {}, "b": {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

func TestSetEdgeCases(t *testing.T) {
	t.Run("large set", func(t *testing.T) {
		schema := types.Set[int](types.Int())
		input := make(map[int]struct{})
		for i := range 1000 {
			input[i] = struct{}{}
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 1000)
	})

	t.Run("set with complex comparable type", func(t *testing.T) {
		type point struct {
			x, y int
		}
		// Note: This works because struct{x, y int} is comparable
		schema := types.Set[point](nil)
		input := map[point]struct{}{{1, 2}: {}, {3, 4}: {}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

// =============================================================================
// INTERNALS TESTS
// =============================================================================

func TestSetInternals(t *testing.T) {
	t.Run("GetInternals returns valid internals", func(t *testing.T) {
		schema := types.Set[string](types.String())
		internals := schema.GetInternals()
		assert.NotNil(t, internals)
	})

	t.Run("IsOptional returns false by default", func(t *testing.T) {
		schema := types.Set[string](types.String())
		assert.False(t, schema.IsOptional())
	})

	t.Run("IsNilable returns false by default", func(t *testing.T) {
		schema := types.Set[string](types.String())
		assert.False(t, schema.IsNilable())
	})

	t.Run("IsOptional returns true after Optional()", func(t *testing.T) {
		schema := types.Set[string](types.String()).Optional()
		assert.True(t, schema.IsOptional())
	})

	t.Run("IsNilable returns true after Nilable()", func(t *testing.T) {
		schema := types.Set[string](types.String()).Nilable()
		assert.True(t, schema.IsNilable())
	})
}
