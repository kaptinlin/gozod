package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestUnion_BasicFunctionality(t *testing.T) {
	t.Run("valid union inputs", func(t *testing.T) {
		// String or Int union
		stringOrInt := Union([]any{
			String(),
			Int(),
		})

		// Test string input
		result, err := stringOrInt.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test int input
		result, err = stringOrInt.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// Test bool input - should fail
		_, err = stringOrInt.Parse(true)
		assert.Error(t, err)
	})

	t.Run("union of different types", func(t *testing.T) {
		// String, Int, Bool union
		multiUnion := Union([]any{
			String(),
			Int(),
			Bool(),
		})

		testCases := []struct {
			input      any
			expected   any
			shouldPass bool
		}{
			{"test", "test", true},
			{123, 123, true},
			{true, true, true},
			{3.14, 3.14, false},             // float not in union
			{[]string{}, []string{}, false}, // slice not in union
		}

		for _, tc := range testCases {
			result, err := multiUnion.Parse(tc.input)
			if tc.shouldPass {
				require.NoError(t, err, "Expected success for input: %v", tc.input)
				assert.Equal(t, tc.expected, result)
			} else {
				assert.Error(t, err, "Expected error for input: %v", tc.input)
			}
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Union([]any{String(), Bool()})

		// Test Parse method
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test MustParse method
		mustResult := schema.MustParse(true)
		assert.Equal(t, true, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected string or boolean"
		schema := Union([]any{
			String(),
			Bool(),
		}, core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeUnion, schema.internals.Def.Type)

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("empty union", func(t *testing.T) {
		emptyUnion := Union([]any{})

		_, err := emptyUnion.Parse("anything")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no union options provided")
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestUnion_TypeSafety(t *testing.T) {
	t.Run("union returns any type", func(t *testing.T) {
		schema := Union([]any{String(), Int()})
		require.NotNil(t, schema)

		// String input
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
		assert.IsType(t, "", result)

		// Int input
		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.IsType(t, 0, result)
	})

	t.Run("UnionOf variadic constructor", func(t *testing.T) {
		schema := UnionOf(String(), Int(), Bool())
		require.NotNil(t, schema)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)

		result, err = schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("Options method returns member schemas", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()

		schema := Union([]any{stringSchema, intSchema})

		options := schema.Options()
		assert.Len(t, options, 2)

		// With the new ZodSchema interface, these may be the same instances
		// or wrapped instances depending on implementation
		// The important thing is that they function correctly
		assert.Equal(t, len(options), 2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Union([]any{String(), Bool()})

		result := schema.MustParse("test")
		assert.IsType(t, "", result)
		assert.Equal(t, "test", result)

		result = schema.MustParse(false)
		assert.IsType(t, false, result)
		assert.Equal(t, false, result)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestUnion_Modifiers(t *testing.T) {
	t.Run("Optional modifier", func(t *testing.T) {
		schema := Union([]any{String(), Int()})
		optionalSchema := schema.Optional()

		// Test non-nil value - returns pointer
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// Test nil value (should be allowed for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		schema := Union([]any{String(), Int()})
		nilableSchema := schema.Nilable()

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - returns pointer
		result, err = nilableSchema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		schema := Union([]any{String(), Int()})
		nullishSchema := schema.Nullish()

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - returns pointer
		result, err = nullishSchema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		schema := Union([]any{String(), Int()})
		defaultSchema := schema.Default("default")

		// Valid input should override default
		result, err := defaultSchema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		assert.IsType(t, 0, result)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestUnion_Chaining(t *testing.T) {
	t.Run("complex chaining", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).
			Default("fallback").
			Optional()

		// Test final behavior - returns pointer due to Optional
		result, err := schema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)

		// Test nil handling
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).
			Default("default").
			Prefault("prefault")

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("modifier immutability", func(t *testing.T) {
		originalSchema := Union([]any{String(), Int()})
		modifiedSchema := originalSchema.Optional()

		// Original should not be affected by modifier
		_, err1 := originalSchema.Parse(nil)
		assert.Error(t, err1, "Original schema should reject nil")

		// Modified schema should have new behavior
		result2, err2 := modifiedSchema.Parse(nil)
		require.NoError(t, err2)
		assert.Nil(t, result2)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestUnion_DefaultAndPrefault(t *testing.T) {
	t.Run("Default with function", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).DefaultFunc(func() any {
			return "function default"
		})

		result, err := schema.Parse("input")
		require.NoError(t, err)
		assert.Equal(t, "input", result)
	})

	t.Run("Prefault with function", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).PrefaultFunc(func() any {
			return "function prefault"
		})

		result, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
	})

	t.Run("Default vs Prefault behavior", func(t *testing.T) {
		// Default should be used when input is nil/undefined
		defaultSchema := Union([]any{String()}).Default("default")

		// Prefault should be used when validation fails
		prefaultSchema := Union([]any{String()}).Prefault("prefault")

		// Test valid input (both should return the input)
		result1, err1 := defaultSchema.Parse("hello")
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)

		result2, err2 := prefaultSchema.Parse("hello")
		require.NoError(t, err2)
		assert.Equal(t, "hello", result2)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestUnion_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		// Only accept strings longer than 3 chars or positive numbers
		schema := Union([]any{String(), Int()}).Refine(func(v any) bool {
			switch val := v.(type) {
			case string:
				return len(val) > 3
			case int:
				return val > 0
			default:
				return false
			}
		})

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Valid int
		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// Invalid string (too short)
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		// Invalid int (negative)
		_, err = schema.Parse(-5)
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be non-empty string or positive number"
		schema := Union([]any{String(), Int()}).Refine(func(v any) bool {
			switch val := v.(type) {
			case string:
				return len(val) > 0
			case int:
				return val > 0
			default:
				return false
			}
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("")
		assert.Error(t, err)

		_, err = schema.Parse(-1)
		assert.Error(t, err)
	})
}

func TestUnion_RefineAny(t *testing.T) {
	t.Run("refineAny validation", func(t *testing.T) {
		// Accept any value that converts to string "valid"
		schema := Union([]any{String(), Bool()}).RefineAny(func(v any) bool {
			return v == "valid" || v == true
		})

		// Valid string
		result, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", result)

		// Valid bool
		result, err = schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid string
		_, err = schema.Parse("invalid")
		assert.Error(t, err)

		// Invalid bool
		_, err = schema.Parse(false)
		assert.Error(t, err)
	})
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestUnion_TypeSpecificMethods(t *testing.T) {
	t.Run("Options method returns all member schemas", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()
		boolSchema := Bool()

		union := Union([]any{stringSchema, intSchema, boolSchema})

		options := union.Options()
		assert.Len(t, options, 3)

		// Verify all options are present (order should be preserved)
		// We can't directly compare schema instances, so we test functionality
		_, err1 := options[0].ParseAny("test")
		assert.NoError(t, err1, "First option should accept strings")

		_, err2 := options[1].ParseAny(123)
		assert.NoError(t, err2, "Second option should accept ints")

		_, err3 := options[2].ParseAny(true)
		assert.NoError(t, err3, "Third option should accept bools")
	})

	t.Run("UnionOf constructor", func(t *testing.T) {
		// Test variadic constructor
		union := UnionOf(String(), Int(), Bool())

		options := union.Options()
		assert.Len(t, options, 3)

		// Test functionality
		result, err := union.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = union.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		result, err = union.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})
}

// =============================================================================
// Error handling and edge case tests
// =============================================================================

func TestUnion_ErrorHandling(t *testing.T) {
	t.Run("union with all failing members", func(t *testing.T) {
		schema := Union([]any{String(), Bool()})
		_, err := schema.Parse(123.45)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no union member matched")
	})

	t.Run("invalid union member", func(t *testing.T) {
		// This should panic during construction, not during parse
		assert.Panics(t, func() {
			Union([]any{String(), 123}) // 123 is not a schema
		})
	})
}

func TestUnion_EdgeCases(t *testing.T) {
	t.Run("union with discriminated union", func(t *testing.T) {
		// Create object schemas with literal constraints for discriminator field
		dogSchema := Object(core.ObjectSchema{
			"type": Literal("dog"),
			"bark": String(),
		})
		catSchema := Object(core.ObjectSchema{
			"type": Literal("cat"),
			"meow": String(),
		})

		// Discriminated union
		animalSchema := DiscriminatedUnion("type", []any{
			dogSchema,
			catSchema,
		})

		// Regular union containing the discriminated union
		schema := Union([]any{
			animalSchema,
			Int(),
		})

		// Test valid discriminated union member
		result, err := schema.Parse(map[string]any{
			"type": "dog",
			"bark": "woof",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "dog",
			"bark": "woof",
		}
		assert.Equal(t, expected, result)

		// Test other valid union member
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)
	})

	t.Run("union of unions", func(t *testing.T) {
		stringOrInt := UnionOf(String(), Int())
		boolOrFloat := UnionOf(Bool(), Float())

		schema := UnionOf(stringOrInt, boolOrFloat)

		// Test nested union members
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("coerce number to string in union", func(t *testing.T) {
		schema := Union([]any{CoercedString(), Int()})

		// Coercion should work, but int takes precedence for int input
		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.IsType(t, 0, result) // Should be int, not string
	})

	t.Run("union with nil schema", func(t *testing.T) {
		schema := Union([]any{String(), nil, Int()})
		_, err := schema.Parse(true) // Should still fail, but not panic
		assert.Error(t, err)
	})
}

func TestUnion_NonOptional(t *testing.T) {
	t.Run("basic non-optional", func(t *testing.T) {
		schema := UnionOf(String(), Int()).Optional().NonOptional()

		// valid string should pass
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result)

		// valid int should pass
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		assert.IsType(t, 0, result)

		// nil should now fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("chained optional and non-optional", func(t *testing.T) {
		schema := UnionOf(String(), Bool()).Optional().NonOptional().Optional().NonOptional()

		// valid bool should pass
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.IsType(t, false, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional on already non-optional schema", func(t *testing.T) {
		schema := UnionOf(Float(), String()).NonOptional()

		// valid float should pass
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
		assert.IsType(t, 0.0, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}
