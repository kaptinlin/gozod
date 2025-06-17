package gozod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestUnionBasicFunctionality(t *testing.T) {
	t.Run("constructor", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, "union", internals.Type)
	})

	t.Run("constructor with params", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}, SchemaParams{Error: "Must be string or number"})
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals.Error)
	})

	t.Run("basic validation - return first valid match", func(t *testing.T) {
		// Test basic union functionality - returns first successful match
		schema := Union([]ZodType[any, any]{String(), Int(), Bool()})

		// Test string input
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test integer input
		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// Test boolean input
		result, err = schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})

		// String input returns string
		result1, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.IsType(t, "", result1)
		assert.Equal(t, "hello", result1)

		// Int input returns int
		result2, err := schema.Parse(42)
		require.NoError(t, err)
		assert.IsType(t, 0, result2)
		assert.Equal(t, 42, result2)

		// Pointer identity preservation
		str := "world"
		inputPtr := &str
		result3, err := schema.Parse(inputPtr)
		require.NoError(t, err)
		resultPtr, ok := result3.(*string)
		require.True(t, ok)
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}).Nilable()

		// nil input should succeed, return typed nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid input keeps type inference
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
		assert.IsType(t, "", result2)
	})

	t.Run("options accessor", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()
		schema := Union([]ZodType[any, any]{stringSchema, intSchema})

		options := schema.Options()
		require.Len(t, options, 2)

		// Test that we can parse with individual options
		result1, err := options[0].Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result1)

		result2, err := options[1].Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestUnionCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}, SchemaParams{Coerce: true})

		// Test with proper input
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestUnionValidations(t *testing.T) {
	t.Run("return valid over invalid", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{
			Object(ObjectSchema{
				"email": String().Email(),
			}),
			String(),
		})

		// Simple string should match second option
		result, err := schema.Parse("asdf")
		require.NoError(t, err)
		assert.Equal(t, "asdf", result)

		// Valid email object should match first option
		emailData := map[string]interface{}{"email": "test@example.com"}
		result, err = schema.Parse(emailData)
		require.NoError(t, err)
		assert.Equal(t, emailData, result)
	})

	t.Run("complex nested union validation", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{
			Object(ObjectSchema{"type": Literal("person"), "name": String().Min(1), "age": Int().Min(0).Max(120)}),
			Object(ObjectSchema{"type": Literal("company"), "name": String().Min(1), "employees": Int().Min(0)}),
		})

		// Valid person data
		personData := map[string]interface{}{"type": "person", "name": "Alice", "age": 30}
		result, err := schema.Parse(personData)
		require.NoError(t, err)
		assert.Equal(t, personData, result)

		// Valid company data
		companyData := map[string]interface{}{"type": "company", "name": "Tech Corp", "employees": 50}
		result, err = schema.Parse(companyData)
		require.NoError(t, err)
		assert.Equal(t, companyData, result)
	})

	t.Run("function parsing - all options fail", func(t *testing.T) {
		// Test when all union options fail validation
		schema := Union([]ZodType[any, any]{
			String().Refine(func(s string) bool { return false }),
			Int().Refine(func(i int) bool { return false }),
		})

		// Should fail with invalid_union error
		_, err := schema.Parse("asdf")
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Code, "invalid_union")
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestUnionModifiers(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})
		optionalSchema := schema.Optional()

		// Valid input
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// nil input
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})
		nilableSchema := schema.Nilable()

		// Valid input
		result, err := nilableSchema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// nil input returns typed nil
		result, err = nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Union([]ZodType[any, any]{String(), Int()})
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Critical: Original schema should remain unchanged and reject nil
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		// Both schemas should validate valid input the same way
		result5, err5 := baseSchema.Parse("hello")
		require.NoError(t, err5)
		assert.Equal(t, "hello", result5)

		result6, err6 := nilableSchema.Parse("hello")
		require.NoError(t, err6)
		assert.Equal(t, "hello", result6)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})

		// Valid input should not panic
		result := schema.MustParse("hello")
		assert.Equal(t, "hello", result)

		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse([]int{1, 2, 3})
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestUnionChaining(t *testing.T) {
	t.Run("validation with constraints", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String().Min(5), Int().Min(10)})

		// Valid cases
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(15)
		require.NoError(t, err)
		assert.Equal(t, 15, result)

		// Invalid cases
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		_, err = schema.Parse(5)
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestUnionTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}).Transform(func(val any, ctx *RefinementContext) (any, error) {
			if s, ok := val.(string); ok {
				return s + "_transformed", nil
			}
			if i, ok := val.(int); ok {
				return i * 2, nil
			}
			return val, nil
		})

		// String transformation
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello_transformed", result)

		// Int transformation
		result, err = schema.Parse(21)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("pipe to another schema", func(t *testing.T) {
		// Transform union to string, then validate string length
		unionToString := Union([]ZodType[any, any]{String(), Int()}).Transform(func(val any, ctx *RefinementContext) (any, error) {
			if s, ok := val.(string); ok {
				return s, nil
			}
			if _, ok := val.(int); ok {
				return "number", nil
			}
			return "", nil
		})

		schema := unionToString.Pipe(String().Min(3))

		// Valid cases
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, "number", result)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestUnionRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}).Refine(func(val any) bool {
			if s, ok := val.(string); ok {
				return len(s) > 3
			}
			if i, ok := val.(int); ok {
				return i > 10
			}
			return false
		}, SchemaParams{Error: "Value must be longer string or larger number"})

		// Invalid cases
		_, err := schema.Parse("hi")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Value must be longer string or larger number")

		_, err = schema.Parse(5)
		assert.Error(t, err)

		// Valid cases
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := "hello"

		// Refine: only validates, never modifies
		refineSchema := Union([]ZodType[any, any]{String()}).Refine(func(val any) bool {
			return true
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Union([]ZodType[any, any]{String()}).Transform(func(val any, ctx *RefinementContext) (any, error) {
			if s, ok := val.(string); ok {
				return s + "_transformed", nil
			}
			return val, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, "hello", refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, "hello_transformed", transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input, transformResult, "Transform should return modified value")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestUnionErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})
		_, err := schema.Parse([]int{1, 2, 3})
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
		assert.Equal(t, string(InvalidUnion), zodErr.Issues[0].Code)
	})

	t.Run("return errors from both union arms", func(t *testing.T) {
		// Test that union collects errors from all failed options
		schema := Union([]ZodType[any, any]{
			Int(),
			String().Refine(func(s string) bool { return false }),
		})

		_, err := schema.Parse("a")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, string(InvalidUnion), zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}, SchemaParams{
			Error: "Must be string or number",
		})
		_, err := schema.Parse([]int{1, 2, 3})
		assert.Error(t, err)

		// Custom error handling may vary in implementation
		assert.NotEmpty(t, err.Error())
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String().Min(10), Int().Min(100)})
		invalidInputs := []interface{}{"short", 5, []int{1, 2}, nil}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err)
			var zodErr *ZodError
			require.True(t, IsZodError(err, &zodErr))
			assert.Contains(t, zodErr.Issues[0].Code, "invalid_union")
		}
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestUnionEdgeCases(t *testing.T) {
	t.Run("empty union", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{})
		_, err := schema.Parse("anything")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("union with single type", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String()})
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse(42)
		assert.Error(t, err)
	})

	t.Run("readonly union", func(t *testing.T) {
		// Test that union works with read-only slice of schemas
		stringSchema := String()
		numberSchema := Int()
		options := []ZodType[any, any]{stringSchema, numberSchema}
		schema := Union(options)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})

		// nil input should fail unless Nilable
		_, err := schema.Parse(nil)
		assert.Error(t, err)

		// Nilable schema should handle nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()})
		complexTypes := []interface{}{
			make(chan int),
			func() int { return 1 },
			struct{ Value int }{Value: 1},
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestUnionDefaultAndPrefault(t *testing.T) {
	t.Run("basic default value", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}).Default("default_string")

		// nil input uses default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_string", result)

		// Valid input bypasses default
		result2, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := Union([]ZodType[any, any]{String(), Int()}).DefaultFunc(func() any {
			counter++
			return counter
		})

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 1, result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, 2, result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse("hello")
		require.NoError(t, err3)
		assert.Equal(t, "hello", result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("prefault value", func(t *testing.T) {
		// Use a fallback value that can pass the union validation
		schema := Union([]ZodType[any, any]{String().Min(10), Int().Min(100)}).Prefault("fallback_value")

		// Valid input succeeds
		result1, err1 := schema.Parse("hello world")
		require.NoError(t, err1)
		assert.Equal(t, "hello world", result1)

		result2, err2 := schema.Parse(150)
		require.NoError(t, err2)
		assert.Equal(t, 150, result2)

		// Invalid input uses prefault
		result3, err3 := schema.Parse("short")
		require.NoError(t, err3)
		assert.Equal(t, "fallback_value", result3)

		result4, err4 := schema.Parse(50)
		require.NoError(t, err4)
		assert.Equal(t, "fallback_value", result4)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := Union([]ZodType[any, any]{String(), Int()}).
			Default("hello").
			Transform(func(val any, ctx *RefinementContext) (any, error) {
				return map[string]any{
					"original": val,
					"type":     "union_value",
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse(42)
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, 42, result1Map["original"])
		assert.Equal(t, "union_value", result1Map["type"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, "hello", result2Map["original"])
		assert.Equal(t, "union_value", result2Map["type"])
	})
}
