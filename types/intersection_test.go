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

func TestIntersection_BasicFunctionality(t *testing.T) {
	t.Run("valid intersection inputs", func(t *testing.T) {
		// Create intersection of two compatible schemas
		intersection := Intersection(String(), String())

		// Test with compatible input
		result, err := intersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("invalid intersection inputs", func(t *testing.T) {
		// Create intersection with conflicting schemas
		intersection := Intersection(String().Min(5), String().Max(3))

		// This should fail as no string can be both >5 and <3 chars
		_, err := intersection.Parse("test")
		assert.Error(t, err)
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		intersection := Intersection(Bool(), Bool())

		// Test Parse method
		result, err := intersection.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Test MustParse method
		mustResult := intersection.MustParse(false)
		assert.Equal(t, false, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			intersection.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected intersection match"
		intersection := Intersection(String(), String(), core.SchemaParams{Error: customError})

		require.NotNil(t, intersection)

		_, err := intersection.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestIntersection_TypeSafety(t *testing.T) {
	t.Run("intersection returns merged type", func(t *testing.T) {
		intersection := Intersection(String(), String())
		require.NotNil(t, intersection)

		result, err := intersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
		assert.IsType(t, "", result) // Ensure type is string
	})

	t.Run("intersection with different types", func(t *testing.T) {
		intersection := Intersection(String(), Int())

		// Should fail as string and int are incompatible
		_, err := intersection.Parse("test")
		assert.Error(t, err)

		_, err = intersection.Parse(123)
		assert.Error(t, err)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		intersection := Intersection(Bool(), Bool())

		result := intersection.MustParse(true)
		assert.IsType(t, true, result)
		assert.Equal(t, true, result)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestIntersection_Modifiers(t *testing.T) {
	t.Run("Optional makes intersection optional", func(t *testing.T) {
		intersection := Intersection(String(), String())
		optionalIntersection := intersection.Optional()

		// Test non-nil value - returns pointer
		result, err := optionalIntersection.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)

		// Test nil value (should be handled by Optional)
		result, err = optionalIntersection.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		intersection := Intersection(String(), String())
		nilableIntersection := intersection.Nilable()

		// Test nil handling
		result, err := nilableIntersection.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - returns pointer
		result, err = nilableIntersection.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("Default preserves intersection type", func(t *testing.T) {
		intersection := Intersection(String(), String())
		defaultIntersection := intersection.Default("default")

		// Valid input should override default
		result, err := defaultIntersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Prefault preserves intersection type", func(t *testing.T) {
		intersection := Intersection(String(), String())
		prefaultIntersection := intersection.Prefault("prefault")

		// Valid input should override prefault
		result, err := prefaultIntersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestIntersection_Chaining(t *testing.T) {
	t.Run("intersection chaining", func(t *testing.T) {
		schema := Intersection(String(), String()).
			Default("default").
			Optional()

		// Test final behavior - returns pointer due to Optional
		result, err := schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := Intersection(Bool(), Bool()).
			Nilable().
			Default(true)

		result, err := schema.Parse(false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Intersection(String(), String()).
			Default("default").
			Prefault("prefault")

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestIntersection_DefaultAndPrefault(t *testing.T) {
	t.Run("default value behavior", func(t *testing.T) {
		intersection := Intersection(String(), String()).Default("default")

		// Valid input should override default
		result, err := intersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// Test default function
		intersectionFunc := Intersection(String(), String()).DefaultFunc(func() any {
			return "func-default"
		})
		result2, err := intersectionFunc.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result2)
	})

	t.Run("prefault value behavior", func(t *testing.T) {
		intersection := Intersection(String(), String()).Prefault("prefault")

		// Valid input should work normally
		result, err := intersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// Test prefault function
		intersectionFunc := Intersection(String(), String()).PrefaultFunc(func() any {
			return "func-prefault"
		})
		result2, err := intersectionFunc.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result2)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestIntersection_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		// Only accept strings with length > 3
		intersection := Intersection(String().Min(3), String().Max(10)).Refine(func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) > 3 && s != "test"
			}
			return false
		})

		result, err := intersection.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = intersection.Parse("test")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be a valid intersection"
		intersection := Intersection(String(), String()).Refine(func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) > 3
			}
			return false
		}, core.SchemaParams{Error: errorMessage})

		result, err := intersection.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = intersection.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refine with complex intersection", func(t *testing.T) {
		intersection := Intersection(String().Min(3), String().Max(10)).Refine(func(v any) bool {
			if s, ok := v.(string); ok {
				return s != "forbidden"
			}
			return false
		})

		// Should pass all validations
		result, err := intersection.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Should fail refine validation
		_, err = intersection.Parse("forbidden")
		assert.Error(t, err)

		// Should fail schema validation (too short)
		_, err = intersection.Parse("hi")
		assert.Error(t, err)
	})
}

func TestIntersection_RefineAny(t *testing.T) {
	t.Run("refineAny flexible validation", func(t *testing.T) {
		intersection := Intersection(String(), String()).RefineAny(func(v any) bool {
			s, ok := v.(string)
			return ok && len(s) >= 4
		})

		// String with >= 4 chars should pass
		result, err := intersection.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = intersection.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// String with < 4 chars should fail
		_, err = intersection.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refineAny with type checking", func(t *testing.T) {
		intersection := Intersection(Int(), Int()).RefineAny(func(v any) bool {
			n, ok := v.(int)
			return ok && n%2 == 0 // Only even numbers
		})

		result, err := intersection.Parse(4)
		require.NoError(t, err)
		assert.Equal(t, 4, result)

		_, err = intersection.Parse(3)
		assert.Error(t, err)
	})
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestIntersection_TypeSpecificMethods(t *testing.T) {
	t.Run("Left returns left schema", func(t *testing.T) {
		leftSchema := String()
		rightSchema := Int()
		intersection := Intersection(leftSchema, rightSchema)

		left := intersection.Left()
		// Note: Left() returns the wrapped schema, not the original
		assert.NotNil(t, left)
	})

	t.Run("Right returns right schema", func(t *testing.T) {
		leftSchema := String()
		rightSchema := Int()
		intersection := Intersection(leftSchema, rightSchema)

		right := intersection.Right()
		// Note: Right() returns the wrapped schema, not the original
		assert.NotNil(t, right)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestIntersection_ErrorHandling(t *testing.T) {
	t.Run("left schema validation error", func(t *testing.T) {
		intersection := Intersection(String().Min(10), String())

		_, err := intersection.Parse("short")
		assert.Error(t, err)
	})

	t.Run("right schema validation error", func(t *testing.T) {
		intersection := Intersection(String(), String().Min(10))

		_, err := intersection.Parse("short")
		assert.Error(t, err)
	})

	t.Run("both schemas validation error", func(t *testing.T) {
		intersection := Intersection(String().Min(10), String().Max(3))

		_, err := intersection.Parse("medium")
		assert.Error(t, err)
	})

	t.Run("merge conflict error", func(t *testing.T) {
		// Create schemas that return different types
		intersection := Intersection(String(), Int())

		// This should fail during merge
		_, err := intersection.Parse("test")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		intersection := Intersection(String(), String(), core.SchemaParams{Error: "Expected intersection match"})

		_, err := intersection.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestIntersection_EdgeCases(t *testing.T) {
	t.Run("intersection of two structs", func(t *testing.T) {
		type structA struct {
			A string
		}
		type structB struct {
			B int
		}
		schemaA := Struct[structA]()
		schemaB := Struct[structB]()
		intersection := Intersection(schemaA, schemaB)

		// Valid input (map) should be merged
		input := map[string]any{"A": "hello", "B": 123}
		result, err := intersection.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{"A": "hello", "B": 123}
		assert.Equal(t, expected, result)

		// Invalid input (missing field)
		_, err = intersection.Parse(map[string]any{"A": "hello"})
		assert.Error(t, err)
	})

	t.Run("intersection with overlapping fields", func(t *testing.T) {
		type structC struct {
			Overlap string
		}
		type structD struct {
			Overlap string
		}
		schemaC := Struct[structC]()
		schemaD := Struct[structD]()
		intersection := Intersection(schemaC, schemaD)

		// Should merge successfully
		input := map[string]any{"Overlap": "value"}
		result, err := intersection.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("intersection with nil value", func(t *testing.T) {
		intersection := Intersection(String(), String())
		_, err := intersection.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("map merging with intersection", func(t *testing.T) {
		intersection := Intersection(String(), String())
		_, err := intersection.Parse(map[string]any{"name": "test", "age": "30"})
		assert.Error(t, err)
	})
}

func TestIntersection_NonOptional(t *testing.T) {
	t.Run("basic non-optional", func(t *testing.T) {
		schema := Intersection(String().Min(2), String().Max(5)).Optional().NonOptional()

		// valid string should pass
		result, err := schema.Parse("abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", result)
		assert.IsType(t, "", result)

		// nil should now fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("chained optional and non-optional", func(t *testing.T) {
		schema := Intersection(String(), String()).Optional().NonOptional().Optional().NonOptional()

		// valid string should pass
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional on already non-optional schema", func(t *testing.T) {
		schema := Intersection(Int(), Int()).NonOptional()

		// valid int should pass
		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		assert.IsType(t, 0, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional with intersection of structs", func(t *testing.T) {
		type Name struct {
			Name string `json:"name"`
		}
		type Age struct {
			Age int `json:"age"`
		}

		schema := Intersection(Struct[Name](), Struct[Age]()).Optional().NonOptional()

		// valid map should pass and be merged
		input := map[string]any{"name": "test", "age": 30}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{"name": "test", "age": 30}
		assert.Equal(t, expected, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}
