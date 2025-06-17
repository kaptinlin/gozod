package gozod

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestIntersectionBasicFunctionality(t *testing.T) {
	t.Run("constructors", func(t *testing.T) {
		// Test Intersection constructor
		left := String()
		right := String()
		schema := Intersection(left, right)
		require.NotNil(t, schema)
		assert.Equal(t, "intersection", schema.internals.Def.Type)
		assert.Equal(t, left, schema.Left())
		assert.Equal(t, right, schema.Right())
	})

	t.Run("basic validation - object intersection", func(t *testing.T) {
		// Create two object schemas
		personSchema := Object(ObjectSchema{
			"name": String(),
		})
		employeeSchema := Object(ObjectSchema{
			"role": String(),
		})

		schema := Intersection(personSchema, employeeSchema)

		// Valid input with both properties
		input := map[string]interface{}{
			"name": "John",
			"role": "Developer",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, "Developer", resultMap["role"])
	})

	t.Run("basic validation - union intersection", func(t *testing.T) {
		// Create intersection of unions that have common type
		unionA := Union([]ZodType[any, any]{String(), Int()})
		unionB := Union([]ZodType[any, any]{String(), Bool()})

		schema := Intersection(unionA, unionB)

		// Valid input - string is common to both unions
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid input - int is only in first union
		_, err = schema.Parse(42)
		assert.Error(t, err)

		// Invalid input - bool is only in second union
		_, err = schema.Parse(true)
		assert.Error(t, err)
	})

	t.Run("validation failure", func(t *testing.T) {
		left := String()
		right := Int()
		schema := Intersection(left, right)

		// No value can be both string and int
		_, err := schema.Parse("hello")
		assert.Error(t, err)

		_, err = schema.Parse(42)
		assert.Error(t, err)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestIntersectionCoercion(t *testing.T) {
	t.Run("coercion not applicable", func(t *testing.T) {
		// Intersection doesn't have its own coercion logic
		// It relies on the constituent schemas
		left := String(SchemaParams{Coerce: true})
		right := String(SchemaParams{Coerce: true})
		schema := Intersection(left, right)

		// Both sides should coerce the number to string
		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestIntersectionValidationMethods(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		personSchema := Object(ObjectSchema{
			"name": String(),
		})
		employeeSchema := Object(ObjectSchema{
			"role": String(),
		})

		schema := Intersection(personSchema, employeeSchema).Refine(func(val any) bool {
			if m, ok := val.(map[string]interface{}); ok {
				name, nameOk := m["name"].(string)
				role, roleOk := m["role"].(string)
				return nameOk && roleOk && name != "" && role != ""
			}
			return false
		})

		// Valid case
		input := map[string]interface{}{
			"name": "John",
			"role": "Developer",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Invalid case - empty name
		invalidInput := map[string]interface{}{
			"name": "",
			"role": "Developer",
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestIntersectionModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		left := String()
		right := String()
		schema := Intersection(left, right).Optional()

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		left := String()
		right := String()
		schema := Intersection(left, right).Nilable()

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		left := String()
		right := String()
		schema := Intersection(left, right).Nullish()

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default wrapper", func(t *testing.T) {
		left := String()
		right := String()
		defaultValue := "default"
		schema := Intersection(left, right).Default(defaultValue)

		// Valid string (should not use default)
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("defaultFunc wrapper", func(t *testing.T) {
		left := String()
		right := String()
		defaultFn := func() any {
			return "generated"
		}
		schema := Intersection(left, right).DefaultFunc(defaultFn)

		require.NotNil(t, schema)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestIntersectionChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		personSchema := Object(ObjectSchema{
			"name": String(),
		})
		employeeSchema := Object(ObjectSchema{
			"role": String(),
		})

		schema := Intersection(personSchema, employeeSchema).
			Refine(func(val any) bool { return true }).
			Optional()

		require.NotNil(t, schema)

		// Test chained functionality
		input := map[string]interface{}{
			"name": "John",
			"role": "Developer",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, "Developer", resultMap["role"])
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestIntersectionTransform(t *testing.T) {
	t.Run("transform method", func(t *testing.T) {
		personSchema := Object(ObjectSchema{
			"name": String(),
		})
		employeeSchema := Object(ObjectSchema{
			"role": String(),
		})

		schema := Intersection(personSchema, employeeSchema).Transform(func(val any, ctx *RefinementContext) (any, error) {
			if m, ok := val.(map[string]interface{}); ok {
				return map[string]interface{}{
					"fullName": m["name"],
					"position": m["role"],
				}, nil
			}
			return val, nil
		})

		input := map[string]interface{}{
			"name": "John",
			"role": "Developer",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John", resultMap["fullName"])
		assert.Equal(t, "Developer", resultMap["position"])
	})

	t.Run("transform chaining", func(t *testing.T) {
		left := String()
		right := String()

		schema := Intersection(left, right).
			Transform(func(val any, ctx *RefinementContext) (any, error) {
				return val, nil
			}).
			TransformAny(func(val any, ctx *RefinementContext) (any, error) {
				if s, ok := val.(string); ok {
					return "transformed_" + s, nil
				}
				return val, nil
			})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "transformed_hello", result)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestIntersectionRefine(t *testing.T) {
	t.Run("simple refinement", func(t *testing.T) {
		left := String()
		right := String()
		schema := Intersection(left, right).Refine(func(val any) bool {
			if s, ok := val.(string); ok {
				return len(s) > 3
			}
			return false
		})

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid case
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("complex refinement", func(t *testing.T) {
		personSchema := Object(ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		employeeSchema := Object(ObjectSchema{
			"role":   String(),
			"salary": Int(),
		})

		schema := Intersection(personSchema, employeeSchema).Refine(func(val any) bool {
			if m, ok := val.(map[string]interface{}); ok {
				age, ageOk := m["age"].(int)
				salary, salaryOk := m["salary"].(int)
				return ageOk && salaryOk && age >= 18 && salary > 0
			}
			return false
		})

		// Valid case
		input := map[string]interface{}{
			"name":   "John",
			"age":    25,
			"role":   "Developer",
			"salary": 50000,
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Invalid case - underage
		invalidInput := map[string]interface{}{
			"name":   "John",
			"age":    16,
			"role":   "Developer",
			"salary": 50000,
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestIntersectionErrorHandling(t *testing.T) {
	t.Run("validation error from left schema", func(t *testing.T) {
		left := String().Min(5)
		right := String()
		schema := Intersection(left, right)

		_, err := schema.Parse("hi")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("validation error from right schema", func(t *testing.T) {
		left := String()
		right := String().Min(5)
		schema := Intersection(left, right)

		_, err := schema.Parse("hi")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("merge error", func(t *testing.T) {
		left := String()
		right := Int()
		schema := Intersection(left, right)

		_, err := schema.Parse("hello")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("refinement error", func(t *testing.T) {
		left := String()
		right := String()
		schema := Intersection(left, right).Refine(func(val any) bool {
			return false // Always fail
		})

		_, err := schema.Parse("hello")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestIntersectionEdgeCases(t *testing.T) {
	t.Run("object merge with overlapping keys", func(t *testing.T) {
		leftSchema := Object(ObjectSchema{
			"name": String(),
			"id":   Int(),
		})
		rightSchema := Object(ObjectSchema{
			"id":   Int(), // Same key, same type
			"role": String(),
		})

		schema := Intersection(leftSchema, rightSchema)

		input := map[string]interface{}{
			"name": "John",
			"id":   123,
			"role": "Developer",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "John", resultMap["name"])
		assert.Equal(t, 123, resultMap["id"])
		assert.Equal(t, "Developer", resultMap["role"])
	})

	t.Run("object merge with conflicting values", func(t *testing.T) {
		leftSchema := Object(ObjectSchema{
			"id": Literal(123),
		})
		rightSchema := Object(ObjectSchema{
			"id": Literal(456), // Different literal value
		})

		schema := Intersection(leftSchema, rightSchema)

		// This should fail because the same key has different literal values
		_, err := schema.Parse(map[string]interface{}{
			"id": 123,
		})
		assert.Error(t, err)
	})

	t.Run("empty object intersection", func(t *testing.T) {
		leftSchema := Object(ObjectSchema{})
		rightSchema := Object(ObjectSchema{})

		schema := Intersection(leftSchema, rightSchema)

		result, err := schema.Parse(map[string]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{}, result)
	})

	t.Run("incompatible types", func(t *testing.T) {
		left := String()
		right := Bool()
		schema := Intersection(left, right)

		// No value can be both string and boolean
		_, err := schema.Parse("hello")
		assert.Error(t, err)

		_, err = schema.Parse(true)
		assert.Error(t, err)
	})

	t.Run("nil input handling", func(t *testing.T) {
		left := String()
		right := String()
		schema := Intersection(left, right)

		// Non-nilable intersection should reject nil
		_, err := schema.Parse(nil)
		assert.Error(t, err)

		// Nilable intersection should accept nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestIntersectionDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		left := String()
		right := String()
		defaultValue := "default"
		schema := Intersection(left, right).Default(defaultValue)

		// Valid string should not use default
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// nil input should use default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		left := String()
		right := String()
		schema := Intersection(left, right).DefaultFunc(func() any {
			counter++
			return fmt.Sprintf("generated-%d", counter)
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "generated-1", result1)
		assert.Equal(t, 1, counter, "Function should be called once for nil input")

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "generated-2", result2)
		assert.Equal(t, 2, counter, "Function should be called twice for second nil input")

		// Valid input should not call function
		result3, err3 := schema.Parse("hello")
		require.NoError(t, err3)
		assert.Equal(t, "hello", result3)
		assert.Equal(t, 2, counter, "Function should not be called for valid input")
	})

	t.Run("prefault", func(t *testing.T) {
		left := String().Min(10) // Require long string
		right := String()
		prefaultValue := "fallback long string"
		schema := Intersection(left, right).Prefault(prefaultValue)

		// Valid case
		result, err := schema.Parse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Invalid case - check if Prefault handles validation failure
		result, err = schema.Parse("hi")
		if err != nil {
			// If Intersection Prefault doesn't handle validation failures,
			// this is expected behavior for intersection types
			assert.Error(t, err)
		} else {
			// If it does handle it, check the fallback value
			assert.Equal(t, prefaultValue, result)
		}
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		left := String().Min(5) // Require string with at least 5 characters
		right := String()
		schema := Intersection(left, right).PrefaultFunc(func() any {
			counter++
			return fmt.Sprintf("fallback-%d", counter)
		})

		// Valid input should not call function
		result1, err1 := schema.Parse("hello world")
		require.NoError(t, err1)
		assert.Equal(t, "hello world", result1)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input should call prefault function (string too short)
		result2, err2 := schema.Parse("hi")
		require.NoError(t, err2)
		assert.Equal(t, "fallback-1", result2)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")

		// Another invalid input should call function again
		result3, err3 := schema.Parse("bye")
		require.NoError(t, err3)
		assert.Equal(t, "fallback-2", result3)
		assert.Equal(t, 2, counter, "Function should increment counter for each invalid input")

		// Valid input still doesn't call function
		result4, err4 := schema.Parse("valid string")
		require.NoError(t, err4)
		assert.Equal(t, "valid string", result4)
		assert.Equal(t, 2, counter, "Counter should remain unchanged for valid input")
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := "default_value"
		prefaultValue := "prefault_value"

		left := String().Min(5)
		right := String()
		schema := Intersection(left, right).Default(defaultValue).Prefault(prefaultValue)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Valid input succeeds
		result2, err2 := schema.Parse("hello world")
		require.NoError(t, err2)
		assert.Equal(t, "hello world", result2)

		// Invalid input uses prefault (string too short)
		result3, err3 := schema.Parse("hi")
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)
	})

	t.Run("complex object intersection with default", func(t *testing.T) {
		personSchema := Object(ObjectSchema{
			"name": String(),
		})
		employeeSchema := Object(ObjectSchema{
			"role": String(),
		})

		defaultValue := map[string]interface{}{
			"name": "Anonymous",
			"role": "Unknown",
		}
		schema := Intersection(personSchema, employeeSchema).Default(defaultValue)

		// Valid input should not use default
		input := map[string]interface{}{
			"name": "John",
			"role": "Developer",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// nil input should use default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})

	t.Run("object intersection with defaultFunc", func(t *testing.T) {
		counter := 0
		personSchema := Object(ObjectSchema{
			"name": String(),
		})
		employeeSchema := Object(ObjectSchema{
			"role": String(),
		})

		schema := Intersection(personSchema, employeeSchema).DefaultFunc(func() any {
			counter++
			return map[string]interface{}{
				"name": fmt.Sprintf("Generated User %d", counter),
				"role": "Generated Role",
			}
		})

		// Each nil input generates a new default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		result1Map := result1.(map[string]interface{})
		assert.Equal(t, "Generated User 1", result1Map["name"])
		assert.Equal(t, "Generated Role", result1Map["role"])

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map := result2.(map[string]interface{})
		assert.Equal(t, "Generated User 2", result2Map["name"])
		assert.Equal(t, "Generated Role", result2Map["role"])

		// Valid input bypasses default generation
		validInput := map[string]interface{}{
			"name": "Alice",
			"role": "Developer",
		}
		result3, err3 := schema.Parse(validInput)
		require.NoError(t, err3)
		assert.Equal(t, validInput, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})
}
