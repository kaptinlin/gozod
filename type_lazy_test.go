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

func TestLazyBasicFunctionality(t *testing.T) {
	t.Run("basic lazy evaluation", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(3)
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		})

		// String input returns string
		result1, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.IsType(t, "", result1)
		assert.Equal(t, "hello", result1)

		// Pointer input returns same pointer
		str := "world"
		result2, err := schema.Parse(&str)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result2)
		assert.Equal(t, &str, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(3)
		}).Nilable()

		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid input keeps type inference
		result2, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)
		assert.IsType(t, "", result2)
	})

	t.Run("deferred evaluation caching", func(t *testing.T) {
		evaluationCount := 0
		schema := Lazy(func() ZodType[any, any] {
			evaluationCount++
			return String().Min(3)
		})

		// Schema created but not evaluated yet
		assert.Equal(t, 0, evaluationCount)

		// First parse triggers evaluation
		_, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, 1, evaluationCount)

		// Second parse uses cached schema
		_, err = schema.Parse("world")
		require.NoError(t, err)
		assert.Equal(t, 1, evaluationCount, "Schema should be cached")
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestLazyCoercion(t *testing.T) {
	t.Run("delegates coercion to inner schema", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return Int()
		})

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("coercion with complex schema", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return Object(ObjectSchema{
				"count": Int(),
			})
		})

		data := map[string]interface{}{
			"count": 42,
		}

		result, err := schema.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, data, result)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestLazyValidations(t *testing.T) {
	t.Run("validation delegation", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(5)
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("complex validation", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return Object(ObjectSchema{
				"name":  String().Min(2),
				"email": String().Email(),
				"age":   Int().Min(0).Max(120),
			})
		})

		validData := map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
			"age":   25,
		}

		result, err := schema.Parse(validData)
		require.NoError(t, err)
		assert.Equal(t, validData, result)

		invalidData := map[string]interface{}{
			"name":  "A", // Too short
			"email": "invalid-email",
			"age":   150, // Too old
		}

		_, err = schema.Parse(invalidData)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestLazyModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		lazySchema := Lazy(func() ZodType[any, any] {
			return String().Email()
		})
		schema := Optional(lazySchema)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(3)
		}).Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		lazySchema := Lazy(func() ZodType[any, any] {
			return String()
		})
		schema := Nullish(lazySchema)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("MustParse method", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		})

		// Should not panic for valid values
		result := schema.MustParse("hello")
		assert.Equal(t, "hello", result)

		// Should panic for invalid values
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestLazyChaining(t *testing.T) {
	t.Run("refine method", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		}).RefineAny(func(s interface{}) bool {
			str, ok := s.(string)
			if !ok {
				return false
			}
			return len(str) > 3
		}, SchemaParams{Error: "String must be longer than 3 characters"})

		_, err := schema.Parse("hi")
		assert.Error(t, err)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("chained methods", func(t *testing.T) {
		lazySchema := Lazy(func() ZodType[any, any] {
			return String()
		})
		refinedSchema := lazySchema.RefineAny(func(s interface{}) bool {
			return len(s.(string)) > 2
		})
		schema := refinedSchema.TransformAny(func(s interface{}, ctx *RefinementContext) (interface{}, error) {
			return "Processed: " + s.(string), nil
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "Processed: hello", result)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestLazyTransform(t *testing.T) {
	t.Run("transform method", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		}).TransformAny(func(s interface{}, ctx *RefinementContext) (interface{}, error) {
			str := s.(string)
			return "Lazy: " + str, nil
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "Lazy: hello", result)
	})

	t.Run("pipe to string", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		}).Pipe(String().Min(10))

		result, err := schema.Parse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		_, err = schema.Parse("short")
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestLazyRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		}).RefineAny(func(val any) bool {
			str, ok := val.(string)
			return ok && len(str) > 3
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		}).RefineAny(func(val any) bool {
			str, ok := val.(string)
			return ok && len(str) >= 5
		}, SchemaParams{
			Error: "String must be at least 5 characters",
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "String must be at least 5 characters")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestLazyErrorHandling(t *testing.T) {
	t.Run("propagates inner schema errors", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(5)
		})

		_, err := schema.Parse("hi")
		assert.Error(t, err)

		var zodErr *ZodError
		if IsZodError(err, &zodErr) {
			assert.Greater(t, len(zodErr.Issues), 0)
		}
	})

	t.Run("handles getter function errors", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return nil // This should cause an error
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("error structure", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(5)
		})

		_, err := schema.Parse("hi")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestLazyEdgeCases(t *testing.T) {
	t.Run("recursive schema evaluation", func(t *testing.T) {
		var TreeNodeSchema ZodType[any, any]
		TreeNodeSchema = Lazy(func() ZodType[any, any] {
			return Object(ObjectSchema{
				"value":    String(),
				"children": Slice(TreeNodeSchema),
			})
		})

		treeData := map[string]interface{}{
			"value": "root",
			"children": []interface{}{
				map[string]interface{}{
					"value":    "child1",
					"children": []interface{}{},
				},
				map[string]interface{}{
					"value": "child2",
					"children": []interface{}{
						map[string]interface{}{
							"value":    "grandchild",
							"children": []interface{}{},
						},
					},
				},
			},
		}

		result, err := TreeNodeSchema.Parse(treeData)
		require.NoError(t, err)
		assert.Equal(t, treeData, result)
	})

	t.Run("mutual recursion", func(t *testing.T) {
		var personSchema, companySchema ZodType[any, any]

		personSchema = Lazy(func() ZodType[any, any] {
			return Object(ObjectSchema{
				"name":    String().Min(1),
				"company": Optional(companySchema),
			})
		})

		companySchema = Lazy(func() ZodType[any, any] {
			return Object(ObjectSchema{
				"name":      String().Min(1),
				"employees": Slice(personSchema),
			})
		})

		companyData := map[string]interface{}{
			"name": "Tech Corp",
			"employees": []interface{}{
				map[string]interface{}{
					"name": "Alice",
					"company": map[string]interface{}{
						"name":      "Tech Corp",
						"employees": []interface{}{},
					},
				},
			},
		}

		result, err := companySchema.Parse(companyData)
		require.NoError(t, err)
		assert.Equal(t, companyData, result)
	})

	t.Run("nil getter function", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic caught: %v", r)
			}
		}()

		schema := Lazy(nil)
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("multiple lazy schemas independence", func(t *testing.T) {
		schema1 := Lazy(func() ZodType[any, any] {
			return String().Min(3)
		})

		schema2 := Lazy(func() ZodType[any, any] {
			return Int().Min(10)
		})

		// Both schemas should work independently
		result1, err1 := schema1.Parse("hello")
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)

		result2, err2 := schema2.Parse(15)
		require.NoError(t, err2)
		assert.Equal(t, 15, result2)

		// Cross-validation should fail appropriately
		_, err := schema1.Parse(42)
		assert.Error(t, err)

		_, err = schema2.Parse("hello")
		assert.Error(t, err)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestLazyDefaultAndPrefault(t *testing.T) {
	t.Run("lazy with default values", func(t *testing.T) {
		// Lazy schemas don't have direct default/prefault support
		// but can be wrapped with Optional/Nilable
		lazySchema := Lazy(func() ZodType[any, any] {
			return String().Min(3)
		})

		optionalSchema := Optional(lazySchema)

		result, err := optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("lazy with nilable behavior", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return String()
		}).Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("unwrap method", func(t *testing.T) {
		innerSchema := String().Min(3)
		lazySchema := Lazy(func() ZodType[any, any] {
			return innerSchema
		})

		// Force evaluation by parsing once
		_, _ = lazySchema.Parse("hello")

		// Now unwrap should return the inner schema
		unwrapped := lazySchema.Unwrap()
		assert.NotNil(t, unwrapped)

		// Unwrapped schema should work the same as inner schema
		result, err := unwrapped.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(3)
		}).(*ZodLazy).DefaultFunc(func() any {
			counter++
			return fmt.Sprintf("lazy-default-%d", counter)
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "lazy-default-1", result1)
		assert.Equal(t, 1, counter)

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "lazy-default-2", result2)
		assert.Equal(t, 2, counter)

		// Valid input should not call function
		result3, err3 := schema.Parse("hello")
		require.NoError(t, err3)
		assert.Equal(t, "hello", result3)
		assert.Equal(t, 2, counter) // Counter should not increment
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Lazy(func() ZodType[any, any] {
			return String().Min(5) // Require at least 5 characters
		}).(*ZodLazy).PrefaultFunc(func() any {
			counter++
			return fmt.Sprintf("lazy-prefault-%d", counter)
		})

		// Valid input should not call function
		result1, err1 := schema.Parse("hello")
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)
		assert.Equal(t, 0, counter)

		// Invalid input should call prefault function
		result2, err2 := schema.Parse("hi") // Too short
		require.NoError(t, err2)
		assert.Equal(t, "lazy-prefault-1", result2)
		assert.Equal(t, 1, counter)

		// Another invalid input should call function again
		result3, err3 := schema.Parse("bye") // Also too short
		require.NoError(t, err3)
		assert.Equal(t, "lazy-prefault-2", result3)
		assert.Equal(t, 2, counter)
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		// Test Default and PrefaultFunc separately since they return ZodType[any, any]

		// Test Default behavior
		defaultSchema := Lazy(func() ZodType[any, any] {
			return String().Min(5)
		}).(*ZodLazy).Default("lazy-default")

		result1, err1 := defaultSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "lazy-default", result1)

		result2, err2 := defaultSchema.Parse("hello")
		require.NoError(t, err2)
		assert.Equal(t, "hello", result2)

		// Test PrefaultFunc behavior separately
		counter := 0
		prefaultSchema := Lazy(func() ZodType[any, any] {
			return String().Min(5)
		}).(*ZodLazy).PrefaultFunc(func() any {
			counter++
			return fmt.Sprintf("lazy-prefault-%d", counter)
		})

		result3, err3 := prefaultSchema.Parse("hello")
		require.NoError(t, err3)
		assert.Equal(t, "hello", result3)
		assert.Equal(t, 0, counter)

		result4, err4 := prefaultSchema.Parse("hi") // Too short
		require.NoError(t, err4)
		assert.Equal(t, "lazy-prefault-1", result4)
		assert.Equal(t, 1, counter)
	})

	t.Run("lazy with nested schemas and defaults", func(t *testing.T) {
		schema := Lazy(func() ZodType[any, any] {
			return Object(ObjectSchema{
				"name": String().Default("Lazy User"),
				"age":  Int().Default(25),
			})
		}).(*ZodLazy).DefaultFunc(func() any {
			return map[string]interface{}{
				"name": "Lazy Default User",
				"age":  30,
			}
		})

		// nil input uses lazy schema's default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, "Lazy Default User", resultMap["name"])
		assert.Equal(t, 30, resultMap["age"])

		// Valid input succeeds
		validInput := map[string]interface{}{
			"name": "Alice",
			"age":  28,
		}
		result, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)
	})
}
