package types_test

import (
	"testing"

	"github.com/kaptinlin/gozod"
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BASIC TUPLE VALIDATION TESTS
// =============================================================================

func TestTuple_Basic(t *testing.T) {
	t.Run("validates tuple with string and int", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})

	t.Run("validates tuple with three elements", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int(), gozod.Bool())
		result, err := schema.Parse([]any{"test", 123, true})
		require.NoError(t, err)
		assert.Equal(t, []any{"test", 123, true}, result)
	})

	t.Run("validates empty tuple", func(t *testing.T) {
		schema := types.Tuple()
		result, err := schema.Parse([]any{})
		require.NoError(t, err)
		assert.Equal(t, []any{}, result)
	})

	t.Run("validates single element tuple", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		result, err := schema.Parse([]any{"solo"})
		require.NoError(t, err)
		assert.Equal(t, []any{"solo"}, result)
	})
}

// =============================================================================
// TYPE VALIDATION TESTS
// =============================================================================

func TestTuple_TypeValidation(t *testing.T) {
	t.Run("fails on wrong element type", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		_, err := schema.Parse([]any{"hello", "not-int"})
		require.Error(t, err)
		// Error message indicates invalid type for int
		assert.Contains(t, err.Error(), "expected int")
	})

	t.Run("fails on wrong first element type", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		_, err := schema.Parse([]any{123, 42})
		require.Error(t, err)
	})

	t.Run("fails on non-array input", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		_, err := schema.Parse("not-array")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tuple")
	})

	t.Run("fails on nil input (non-optional)", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		_, err := schema.Parse(nil)
		require.Error(t, err)
	})
}

// =============================================================================
// LENGTH VALIDATION TESTS
// =============================================================================

func TestTuple_Length(t *testing.T) {
	t.Run("fails when too few elements", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int(), gozod.Bool())
		_, err := schema.Parse([]any{"hello"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least")
	})

	t.Run("fails when too many elements (no rest)", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		_, err := schema.Parse([]any{"hello", 42, "extra"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at most")
	})

	t.Run("exact length required without rest", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		_, err := schema.Parse([]any{"hello", 42, true, "extra"})
		require.Error(t, err)
	})
}

// =============================================================================
// REST ELEMENT TESTS
// =============================================================================

func TestTuple_Rest(t *testing.T) {
	t.Run("allows additional elements with rest schema", func(t *testing.T) {
		schema := types.TupleWithRest(
			[]core.ZodSchema{gozod.String(), gozod.Int()},
			gozod.Bool(),
		)
		result, err := schema.Parse([]any{"hello", 42, true, false, true})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42, true, false, true}, result)
	})

	t.Run("validates rest elements", func(t *testing.T) {
		schema := types.TupleWithRest(
			[]core.ZodSchema{gozod.String()},
			gozod.Int(),
		)
		_, err := schema.Parse([]any{"hello", "not-int"})
		require.Error(t, err)
	})

	t.Run("works with no rest elements provided", func(t *testing.T) {
		schema := types.TupleWithRest(
			[]core.ZodSchema{gozod.String(), gozod.Int()},
			gozod.Bool(),
		)
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})

	t.Run("rest method creates new tuple with rest", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int()).Rest(gozod.Bool())
		result, err := schema.Parse([]any{"hello", 42, true, false})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42, true, false}, result)
	})
}

// =============================================================================
// OPTIONAL ELEMENT TESTS
// =============================================================================

func TestTuple_OptionalElements(t *testing.T) {
	t.Run("validates tuple with optional element", func(t *testing.T) {
		schema := types.Tuple(
			gozod.String(),
			gozod.Int().Optional(),
		)
		// All elements provided
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		// Optional elements return pointer types, so we check length and first element
		assert.Len(t, result, 2)
		assert.Equal(t, "hello", result[0])
		// Second element is *int due to Optional()
		if ptr, ok := result[1].(*int); ok {
			assert.Equal(t, 42, *ptr)
		}
	})

	t.Run("allows missing optional elements at end", func(t *testing.T) {
		schema := types.Tuple(
			gozod.String(),
			gozod.Int().Optional(),
		)
		// Optional element omitted
		result, err := schema.Parse([]any{"hello"})
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("multiple optional elements", func(t *testing.T) {
		schema := types.Tuple(
			gozod.String(),
			gozod.Int().Optional(),
			gozod.Bool().Optional(),
		)
		// Only required element provided
		result, err := schema.Parse([]any{"test"})
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

// =============================================================================
// MODIFIER TESTS
// =============================================================================

func TestTuple_Optional(t *testing.T) {
	t.Run("optional tuple accepts nil", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("optional tuple accepts valid value", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Optional()
		result, err := schema.Parse([]any{"hello"})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, []any{"hello"}, *result)
	})
}

func TestTuple_Nilable(t *testing.T) {
	t.Run("nilable tuple accepts nil", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Nilable()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable tuple accepts valid value", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Nilable()
		result, err := schema.Parse([]any{"hello"})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestTuple_Nullish(t *testing.T) {
	t.Run("nullish tuple accepts nil", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Nullish()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestTuple_Default(t *testing.T) {
	t.Run("default value used for nil input with optional", func(t *testing.T) {
		// Default must be used with Optional to accept nil
		// When Default is set, it returns the default value for nil input
		schema := types.Tuple(gozod.String()).Optional().Default([]any{"default"})
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		// Default value should be returned
		require.NotNil(t, result)
		assert.Equal(t, []any{"default"}, *result)
	})

	t.Run("non-optional tuple rejects nil", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		_, err := schema.Parse(nil)
		require.Error(t, err)
	})
}

// =============================================================================
// STRICT PARSE TESTS
// =============================================================================

func TestTuple_StrictParse(t *testing.T) {
	t.Run("strict parse with correct type", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		result, err := schema.StrictParse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})

	t.Run("must parse panics on error", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		assert.Panics(t, func() {
			schema.MustParse("not-array")
		})
	})

	t.Run("must strict parse panics on error", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		assert.Panics(t, func() {
			schema.MustStrictParse([]any{123}) // wrong type
		})
	})
}

// =============================================================================
// VALIDATION METHODS TESTS
// =============================================================================

func TestTuple_ValidationMethods(t *testing.T) {
	t.Run("min validation", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int()).Min(2)
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("max validation", func(t *testing.T) {
		schema := types.TupleWithRest(
			[]core.ZodSchema{gozod.String()},
			gozod.Int(),
		).Max(3)
		_, err := schema.Parse([]any{"hello", 1, 2, 3, 4})
		require.Error(t, err)
	})

	t.Run("length validation", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int()).Length(2)
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("non-empty validation", func(t *testing.T) {
		schema := types.Tuple().NonEmpty()
		_, err := schema.Parse([]any{})
		require.Error(t, err)
	})

	t.Run("refine validation", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int()).Refine(func(arr []any) bool {
			if len(arr) >= 2 {
				if num, ok := arr[1].(int); ok {
					return num > 0
				}
			}
			return false
		})
		_, err := schema.Parse([]any{"hello", -5})
		require.Error(t, err)
	})
}

// =============================================================================
// METADATA TESTS
// =============================================================================

func TestTuple_Metadata(t *testing.T) {
	t.Run("describe adds description", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Describe("A simple tuple")
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "A simple tuple", meta.Description)
	})

	t.Run("meta stores metadata", func(t *testing.T) {
		schema := types.Tuple(gozod.String()).Meta(core.GlobalMeta{
			Title:       "Test Tuple",
			Description: "A test tuple schema",
		})
		meta, ok := core.GlobalRegistry.Get(schema)
		require.True(t, ok)
		assert.Equal(t, "Test Tuple", meta.Title)
		assert.Equal(t, "A test tuple schema", meta.Description)
	})
}

// =============================================================================
// INTERNALS TESTS
// =============================================================================

func TestTuple_Internals(t *testing.T) {
	t.Run("get internals returns valid internals", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		internals := schema.GetInternals()
		assert.NotNil(t, internals)
		assert.Equal(t, core.ZodTypeTuple, internals.Type)
	})

	t.Run("get items returns item schemas", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		items := schema.GetItems()
		assert.Len(t, items, 2)
	})

	t.Run("get rest returns nil when no rest", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		rest := schema.GetRest()
		assert.Nil(t, rest)
	})

	t.Run("get rest returns rest schema", func(t *testing.T) {
		schema := types.TupleWithRest(
			[]core.ZodSchema{gozod.String()},
			gozod.Int(),
		)
		rest := schema.GetRest()
		assert.NotNil(t, rest)
	})

	t.Run("is optional returns correct value", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		assert.False(t, schema.IsOptional())

		optSchema := schema.Optional()
		assert.True(t, optSchema.IsOptional())
	})

	t.Run("is nilable returns correct value", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		assert.False(t, schema.IsNilable())

		nilableSchema := schema.Nilable()
		assert.True(t, nilableSchema.IsNilable())
	})
}

// =============================================================================
// INPUT TYPE CONVERSION TESTS
// =============================================================================

func TestTuple_InputConversion(t *testing.T) {
	t.Run("converts typed slice to []any", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		// Using a slice literal that gets converted
		input := []interface{}{"hello", 42}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})

	t.Run("works with array input", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		input := [2]any{"hello", 42}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})
}

// =============================================================================
// PARSE ANY TESTS
// =============================================================================

func TestTuple_ParseAny(t *testing.T) {
	t.Run("parse any returns any type", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		result, err := schema.ParseAny([]any{"hello", 42})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// CONSTRUCTOR TESTS
// =============================================================================

func TestTuple_Constructors(t *testing.T) {
	t.Run("TuplePtr creates pointer-returning tuple", func(t *testing.T) {
		schema := types.TuplePtr(gozod.String())
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("TupleTyped with explicit types", func(t *testing.T) {
		schema := types.TupleTyped[[]any, []any](
			[]core.ZodSchema{gozod.String(), gozod.Int()},
			nil,
		)
		result, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Equal(t, []any{"hello", 42}, result)
	})
}

// =============================================================================
// ENHANCED ZOD V4 COMPATIBILITY TESTS
// =============================================================================

func TestTuple_EnhancedCoverage(t *testing.T) {
	t.Run("sparse array input fails validation", func(t *testing.T) {
		// Zod v4: sparse arrays should fail - elements with undefined holes are invalid
		schema := types.Tuple(gozod.String(), gozod.Int())
		// Simulate sparse array behavior with nil elements
		_, err := schema.Parse([]any{nil, nil})
		require.Error(t, err)
	})

	t.Run("optional element followed by required element", func(t *testing.T) {
		// Zod v4 pattern: [string, number?, string]
		// Optional in the middle, required at the end
		schema := types.Tuple(
			gozod.String(),
			gozod.Int().Optional(),
			gozod.String(),
		)

		// Valid: all elements provided (optional has value)
		result1, err := schema.Parse([]any{"first", 42, "last"})
		require.NoError(t, err)
		assert.Len(t, result1, 3)
		assert.Equal(t, "first", result1[0])
		assert.Equal(t, "last", result1[2])

		// Invalid: missing required element at end (only 2 elements)
		_, err = schema.Parse([]any{"first", 42})
		require.Error(t, err)

		// Invalid: only first element
		_, err = schema.Parse([]any{"first"})
		require.Error(t, err)
	})

	t.Run("all optional elements allows empty array", func(t *testing.T) {
		// Zod v4: tuple with all optional elements accepts empty array
		schema := types.Tuple(
			gozod.String().Optional(),
			gozod.Int().Optional(),
			gozod.Bool().Optional(),
		)

		// Empty array should be valid (all items optional)
		result1, err := schema.Parse([]any{})
		require.NoError(t, err)
		assert.Len(t, result1, 0)

		// Partial arrays should be valid
		result2, err := schema.Parse([]any{"hello"})
		require.NoError(t, err)
		assert.Len(t, result2, 1)

		result3, err := schema.Parse([]any{"hello", 42})
		require.NoError(t, err)
		assert.Len(t, result3, 2)

		// Full array should be valid
		result4, err := schema.Parse([]any{"hello", 42, true})
		require.NoError(t, err)
		assert.Len(t, result4, 3)

		// Array that's too long should fail
		_, err = schema.Parse([]any{"hello", 42, true, "extra"})
		require.Error(t, err)
	})

	t.Run("optional elements with rest schema", func(t *testing.T) {
		// Zod v4: [string, number?, string?, ...boolean[]]
		schema := types.TupleWithRest(
			[]core.ZodSchema{
				gozod.String(),
				gozod.Int().Optional(),
				gozod.String().Optional(),
			},
			gozod.Bool(),
		)

		// Valid combinations
		validData := [][]any{
			{"asdf"},
			{"asdf", 1234},
			{"asdf", 1234, "asdf"},
			{"asdf", 1234, "asdf", true, false, true},
		}
		for _, data := range validData {
			result, err := schema.Parse(data)
			require.NoError(t, err, "should parse %v", data)
			assert.Len(t, result, len(data))
		}

		// Invalid combinations - wrong types
		invalidData := [][]any{
			{"asdf", "not-int"},                             // wrong type for optional int
			{"asdf", 1234, "asdf", "not-bool"},              // wrong rest element type
			{"asdf", 1234, "asdf", true, false, "not-bool"}, // wrong rest element type
		}
		for _, data := range invalidData {
			_, err := schema.Parse(data)
			require.Error(t, err, "should fail for %v", data)
		}
	})

	t.Run("tuple immutability - Rest returns new instance", func(t *testing.T) {
		original := types.Tuple(gozod.String(), gozod.Int())
		withRest := original.Rest(gozod.Bool())

		// Original should still reject extra elements
		_, err := original.Parse([]any{"hello", 42, true})
		require.Error(t, err)

		// WithRest should accept rest elements
		result, err := withRest.Parse([]any{"hello", 42, true})
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("tuple immutability - modifiers return new instance", func(t *testing.T) {
		original := types.Tuple(gozod.String())
		optional := original.Optional()
		nilable := original.Nilable()

		// Original should reject nil
		_, err := original.Parse(nil)
		require.Error(t, err)

		// Optional should accept nil
		result1, err := optional.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result1)

		// Nilable should accept nil
		result2, err := nilable.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})

	t.Run("tuple with Refine returns new instance", func(t *testing.T) {
		original := types.Tuple(gozod.String(), gozod.Int())
		refined := original.Refine(func(arr []any) bool {
			if len(arr) >= 2 {
				if num, ok := arr[1].(int); ok {
					return num > 10
				}
			}
			return false
		})

		// Original should accept any int
		_, err := original.Parse([]any{"hello", 5})
		require.NoError(t, err)

		// Refined should reject int <= 10
		_, err = refined.Parse([]any{"hello", 5})
		require.Error(t, err)

		// Refined should accept int > 10
		_, err = refined.Parse([]any{"hello", 15})
		require.NoError(t, err)
	})

	t.Run("error paths for nested tuple validation", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())

		// Error at path [1] for wrong type
		_, err := schema.Parse([]any{"hello", "not-int"})
		require.Error(t, err)
		var zodErr *gozod.ZodError
		require.True(t, gozod.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		// Path should indicate index 1
		assert.Contains(t, zodErr.Issues[0].Path, 1)
	})

	t.Run("error message for too many elements", func(t *testing.T) {
		schema := types.Tuple(gozod.String(), gozod.Int())
		_, err := schema.Parse([]any{"hello", 42, true})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at most")
	})

	t.Run("error message for wrong input type", func(t *testing.T) {
		schema := types.Tuple(gozod.String())
		_, err := schema.Parse(map[string]any{"key": "value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tuple")
	})
}

// =============================================================================
// EDGE CASES
// =============================================================================

func TestTuple_EdgeCases(t *testing.T) {
	t.Run("handles nested tuples", func(t *testing.T) {
		innerTuple := types.Tuple(gozod.String(), gozod.Int())
		outerTuple := types.Tuple(gozod.String(), innerTuple)

		result, err := outerTuple.Parse([]any{"outer", []any{"inner", 42}})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("handles tuple with object schema", func(t *testing.T) {
		objSchema := gozod.Object(core.ObjectSchema{
			"name": gozod.String(),
		})
		tupleSchema := types.Tuple(gozod.String(), objSchema)

		result, err := tupleSchema.Parse([]any{"label", map[string]any{"name": "test"}})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("handles tuple with slice schema", func(t *testing.T) {
		sliceSchema := gozod.Slice[string](gozod.String())
		tupleSchema := types.Tuple(gozod.Int(), sliceSchema)

		result, err := tupleSchema.Parse([]any{42, []any{"a", "b", "c"}})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}
