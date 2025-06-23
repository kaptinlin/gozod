package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestObjectBasicFunctionality(t *testing.T) {
	t.Run("constructor", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, core.ZodTypeObject, internals.Type)
	})

	t.Run("constructor with params", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"email": String().Email(),
		}, core.SchemaParams{
			Error: "core.Custom object error",
		})

		require.NotNil(t, schema)
		// Coercion is no longer supported for collection types
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Map input returns map
		input := map[string]any{
			"name": "Alice",
			"age":  30,
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, map[string]any{}, result)

		// Pointer input returns same pointer
		result2, err := schema.Parse(&input)
		require.NoError(t, err)
		assert.IsType(t, (*map[string]any)(nil), result2)
		assert.Equal(t, &input, result2) // Same pointer identity
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		}).Nilable()

		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*map[string]any)(nil), result)

		// Valid input keeps type inference
		validInput := map[string]any{"name": "Alice"}
		result2, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result2)
		assert.IsType(t, map[string]any{}, result2)
	})

	t.Run("complex object type inference", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"f1": Number(),
			"f2": String().Optional(),
			"f3": String().Nilable(),
			"f4": Slice(Object(core.ObjectSchema{
				"t": Union([]core.ZodType[any, any]{String(), Bool()}),
			})),
		})

		input := map[string]any{
			"f1": 42.0,
			"f2": "optional string",
			"f3": "nullable string",
			"f4": []any{
				map[string]any{"t": "string"},
				map[string]any{"t": true},
			},
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test with nil f3
		input2 := map[string]any{
			"f1": 42.0,
			"f3": nil,
			"f4": []any{
				map[string]any{"t": false},
			},
		}

		result2, err := schema.Parse(input2)
		require.NoError(t, err)
		assert.NotNil(t, result2)
	})
}

// =============================================================================
// 2. Validation methods (coercion no longer supported for collection types)
// =============================================================================

func TestObjectValidations(t *testing.T) {
	t.Run("required field validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(2),
			"age":  Int().Min(0).Max(120),
		})

		// Valid object
		input := map[string]any{
			"name": "Alice",
			"age":  30,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "Alice", resultMap["name"])
		assert.Equal(t, 30, resultMap["age"])

		// Missing required field
		invalidInput := map[string]any{
			"name": "Alice",
			// missing age
		}

		_, err = schema.Parse(invalidInput)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		hasRequiredError := false
		for _, issue := range zodErr.Issues {
			if issue.Code == core.InvalidType && len(issue.Path) > 0 && issue.Path[0] == "age" {
				hasRequiredError = true
				break
			}
		}
		assert.True(t, hasRequiredError)
	})

	t.Run("field validation errors", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(5),
			"age":  Int().Min(18),
		})

		input := map[string]any{
			"name": "A", // Too short
			"age":  10,  // Too young
		}

		_, err := schema.Parse(input)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Greater(t, len(zodErr.Issues), 0)
	})

	t.Run("optional fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":    String(),
			"age":     Int(),
			"address": String().Optional(),
		})

		// With optional field
		input1 := map[string]any{
			"name":    "Alice",
			"age":     30,
			"address": "123 Main St",
		}

		result1, err := schema.Parse(input1)
		require.NoError(t, err)
		resultMap1 := result1.(map[string]any)
		assert.Equal(t, "123 Main St", resultMap1["address"])

		// Without optional field
		input2 := map[string]any{
			"name": "Bob",
			"age":  25,
		}

		result2, err := schema.Parse(input2)
		require.NoError(t, err)
		resultMap2 := result2.(map[string]any)
		assert.Equal(t, "Bob", resultMap2["name"])
		_, hasAddress := resultMap2["address"]
		assert.False(t, hasAddress)
	})

	t.Run("nested object validation", func(t *testing.T) {
		userSchema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		profileSchema := Object(core.ObjectSchema{
			"user":    userSchema,
			"country": String(),
		})

		input := map[string]any{
			"user": map[string]any{
				"name": "Alice",
				"age":  30,
			},
			"country": "US",
		}

		result, err := profileSchema.Parse(input)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		userMap := resultMap["user"].(map[string]any)
		assert.Equal(t, "Alice", userMap["name"])
		assert.Equal(t, "US", resultMap["country"])
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestObjectModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Optional()

		// Valid object
		validInput := map[string]any{
			"name": "Alice",
			"age":  30,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input should succeed
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Nilable()

		// Valid object
		validInput := map[string]any{
			"name": "Alice",
			"age":  30,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*map[string]any)(nil), result)
	})

	t.Run("object modes", func(t *testing.T) {
		baseSchema := core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}

		// Strip mode (default)
		stripSchema := Object(baseSchema)
		input := map[string]any{
			"name":    "Alice",
			"age":     30,
			"unknown": "should be stripped",
		}

		result, err := stripSchema.Parse(input)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "Alice", resultMap["name"])
		assert.Equal(t, 30, resultMap["age"])
		_, hasUnknown := resultMap["unknown"]
		assert.False(t, hasUnknown, "Unknown key should be stripped")

		// Strict mode
		strictSchema := StrictObject(baseSchema)
		_, err = strictSchema.Parse(input)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		hasUnrecognizedKeys := false
		for _, issue := range zodErr.Issues {
			if issue.Code == core.UnrecognizedKeys {
				hasUnrecognizedKeys = true
				break
			}
		}
		assert.True(t, hasUnrecognizedKeys)

		// Loose mode
		looseSchema := LooseObject(baseSchema)
		result, err = looseSchema.Parse(input)
		require.NoError(t, err)

		resultMap = result.(map[string]any)
		assert.Equal(t, "Alice", resultMap["name"])
		assert.Equal(t, 30, resultMap["age"])
		assert.Equal(t, "should be stripped", resultMap["unknown"])
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestObjectChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(2),
			"age":  Int().Min(0),
		}).Strict().Passthrough().Strip()

		// Test that final method (Strip) takes effect
		input := map[string]any{
			"name":    "Alice",
			"age":     30,
			"unknown": "should be stripped",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "Alice", resultMap["name"])
		assert.Equal(t, 30, resultMap["age"])
		_, hasUnknown := resultMap["unknown"]
		assert.False(t, hasUnknown, "Unknown key should be stripped")
	})

	t.Run("catchall with validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		}).Catchall(String())

		input := map[string]any{
			"name":    "Alice",
			"extra":   "valid string",
			"another": "also string",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap := result.(map[string]any)
		assert.Equal(t, "Alice", resultMap["name"])
		assert.Equal(t, "valid string", resultMap["extra"])
		assert.Equal(t, "also string", resultMap["another"])

		// Test catchall validation failure
		invalidInput := map[string]any{
			"name":  "Alice",
			"extra": 123, // Should fail string validation
		}

		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestObjectTransform(t *testing.T) {
	t.Run("transform", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Transform(func(obj map[string]any, ctx *core.RefinementContext) (any, error) {
			// Transform: add a computed field
			obj["computed"] = "transformed"
			return obj, nil
		})

		result, err := schema.Parse(map[string]any{
			"name": "Alice",
			"age":  30,
		})
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, "transformed", resultMap["computed"])
	})

	t.Run("transform with type change", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Transform(func(obj map[string]any, ctx *core.RefinementContext) (any, error) {
			// Transform to string representation
			return obj["name"].(string) + " is " + string(rune(obj["age"].(int)+'0')) + " years old", nil
		})

		result, err := schema.Parse(map[string]any{
			"name": "Alice",
			"age":  5, // Single digit for simple conversion
		})
		require.NoError(t, err)
		assert.Equal(t, "Alice is 5 years old", result)
	})

	t.Run("pipe", func(t *testing.T) {
		objectSchema := Object(core.ObjectSchema{
			"value": String(),
		})

		stringSchema := String().Min(1)

		pipeSchema := objectSchema.Pipe(stringSchema)

		// This should fail because object doesn't match string
		_, err := pipeSchema.Parse(map[string]any{
			"value": "test",
		})
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Refine
// =============================================================================

func TestObjectRefine(t *testing.T) {
	t.Run("refine", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(obj map[string]any) bool {
			// Only allow objects where age is greater than name length
			name, hasName := obj["name"].(string)
			age, hasAge := obj["age"].(int)
			return hasName && hasAge && age > len(name)
		})

		// Valid case (age > name length)
		_, err := schema.Parse(map[string]any{
			"name": "Alice",
			"age":  30,
		})
		require.NoError(t, err)

		// Invalid case (age <= name length)
		_, err = schema.Parse(map[string]any{
			"name": "Alice",
			"age":  3,
		})
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"password":        String().Min(8),
			"confirmPassword": String().Min(8),
		}).Refine(func(obj map[string]any) bool {
			password, hasPassword := obj["password"].(string)
			confirmPassword, hasConfirm := obj["confirmPassword"].(string)
			return hasPassword && hasConfirm && password == confirmPassword
		}, core.SchemaParams{
			Error: "Passwords must match",
		})

		// Valid case
		_, err := schema.Parse(map[string]any{
			"password":        "password123",
			"confirmPassword": "password123",
		})
		require.NoError(t, err)

		// Invalid case
		_, err = schema.Parse(map[string]any{
			"password":        "password123",
			"confirmPassword": "different123",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Passwords must match")
	})
}

// =============================================================================
// 7. Error handling
// =============================================================================

func TestObjectErrorHandling(t *testing.T) {
	t.Run("type error", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		})
		_, err := schema.Parse("not an object")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("field error", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		_, err := schema.Parse(map[string]any{
			"name": "Alice",
			"age":  "not a number",
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("multiple field errors", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String().Min(5),
			"age":   Int().Min(18),
			"email": String().Email(),
		})

		_, err := schema.Parse(map[string]any{
			"name":  "A",            // Too short
			"age":   10,             // Too young
			"email": "not-an-email", // Invalid format
		})

		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 2) // At least 2 errors
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(2, core.SchemaParams{
				Error: "Name must be at least 2 characters",
			}),
		})

		_, err := schema.Parse(map[string]any{
			"name": "A",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Name must be at least 2 characters")
	})
}

// =============================================================================
// 8. Edge and mutual exclusion cases
// =============================================================================

func TestObjectEdgeCases(t *testing.T) {
	t.Run("empty object", func(t *testing.T) {
		schema := Object(core.ObjectSchema{})

		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Empty object with extra keys should be stripped
		result, err = schema.Parse(map[string]any{
			"extra": "value",
		})
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, 0, len(resultMap))
	})

	t.Run("deeply nested objects", func(t *testing.T) {
		level3Schema := Object(core.ObjectSchema{
			"value": String(),
		})

		level2Schema := Object(core.ObjectSchema{
			"level3": level3Schema,
			"name":   String(),
		})

		level1Schema := Object(core.ObjectSchema{
			"level2": level2Schema,
			"id":     Int(),
		})

		input := map[string]any{
			"id": 1,
			"level2": map[string]any{
				"name": "Level 2",
				"level3": map[string]any{
					"value": "Deep value",
				},
			},
		}

		result, err := level1Schema.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify deep structure
		resultMap := result.(map[string]any)
		level2Map := resultMap["level2"].(map[string]any)
		level3Map := level2Map["level3"].(map[string]any)
		assert.Equal(t, "Deep value", level3Map["value"])
	})

	t.Run("object with all optional fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"field1": String().Optional(),
			"field2": Int().Optional(),
			"field3": Bool().Optional(),
		})

		// Empty object should be valid
		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)

		// Partial object should be valid
		input := map[string]any{
			"field1": "value1",
		}

		result, err = schema.Parse(input)
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, "value1", resultMap["field1"])
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		})

		complexTypes := []any{
			make(chan int),
			func() int { return 1 },
			[]any{1, 2, 3},
			"string",
			123,
			true,
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Should reject type %T", input)
		}
	})

	t.Run("shape access", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String().Email(),
		})

		shape := schema.Shape()

		assert.Equal(t, 3, len(shape))
		assert.Contains(t, shape, "name")
		assert.Contains(t, shape, "age")
		assert.Contains(t, shape, "email")

		// Verify we can access individual field schemas
		nameSchema := shape["name"]
		assert.NotNil(t, nameSchema)

		// Test that shape returns the actual schemas
		_, err := nameSchema.Parse("test")
		require.NoError(t, err)
	})

	t.Run("keyof method", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String().Email(),
		})

		keySchema := schema.Keyof()

		// Test valid keys
		result1, err := keySchema.Parse("name")
		require.NoError(t, err)
		assert.Equal(t, "name", result1)

		result2, err := keySchema.Parse("age")
		require.NoError(t, err)
		assert.Equal(t, "age", result2)

		result3, err := keySchema.Parse("email")
		require.NoError(t, err)
		assert.Equal(t, "email", result3)

		// Test invalid key
		_, err = keySchema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("empty object keyof", func(t *testing.T) {
		emptySchema := Object(core.ObjectSchema{})
		keySchema := emptySchema.Keyof()

		// Should be Never type for empty objects
		_, err := keySchema.Parse("anything")
		assert.Error(t, err)
	})

	t.Run("object operations", func(t *testing.T) {
		baseSchema := Object(core.ObjectSchema{
			"name":    String(),
			"age":     Int(),
			"email":   String().Email(),
			"address": String().Optional(),
		})

		// Pick method
		pickedSchema := baseSchema.Pick([]string{"name", "email"})
		assert.Equal(t, 2, len(pickedSchema.internals.Shape))
		assert.Contains(t, pickedSchema.internals.Shape, "name")
		assert.Contains(t, pickedSchema.internals.Shape, "email")

		// Omit method
		omittedSchema := baseSchema.Omit([]string{"age", "address"})
		assert.Equal(t, 2, len(omittedSchema.internals.Shape))
		assert.Contains(t, omittedSchema.internals.Shape, "name")
		assert.Contains(t, omittedSchema.internals.Shape, "email")

		// Extend method
		extension := core.ObjectSchema{
			"phone":   String(),
			"country": String().Default("US"),
		}
		extendedSchema := baseSchema.Extend(extension)
		assert.Equal(t, 6, len(extendedSchema.internals.Shape))
		assert.Contains(t, extendedSchema.internals.Shape, "phone")

		// Partial method
		partialSchema := baseSchema.Partial()
		assert.Equal(t, 4, len(partialSchema.internals.Shape))

		// Test partial validation - all fields should be optional
		result, err := partialSchema.Parse(map[string]any{
			"name": "Alice",
			// other fields missing but should be okay
		})
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, "Alice", resultMap["name"])
	})

	t.Run("merge objects", func(t *testing.T) {
		schema1 := Object(core.ObjectSchema{
			"a": String(),
			"b": String().Optional(),
		})

		schema2 := Object(core.ObjectSchema{
			"a": String().Optional(), // Override with optional
			"b": String(),            // Override with required
		})

		merged := schema1.Merge(schema2)

		// Test that schema2 fields override schema1
		input := map[string]any{
			"b": "required field",
			// "a" is optional in merged schema
		}

		result, err := merged.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// 9. Default and Prefault tests
// =============================================================================

func TestObjectDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		defaultValue := map[string]any{
			"name": "Default Name",
			"age":  25,
		}

		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Default(defaultValue)

		// nil input should use default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)

		// Valid input should override default
		validInput := map[string]any{
			"name": "Alice",
			"age":  30,
		}

		result, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)
	})

	t.Run("default function", func(t *testing.T) {
		counter := 0
		schema := Object(core.ObjectSchema{
			"name": String(),
			"id":   Int(),
		}).DefaultFunc(func() map[string]any {
			counter++
			return map[string]any{
				"name": "Generated",
				"id":   counter,
			}
		})

		// Each nil input generates a new default value
		result1, err := schema.Parse(nil)
		require.NoError(t, err)
		result1Map := result1.(map[string]any)
		assert.Equal(t, "Generated", result1Map["name"])
		assert.Equal(t, 1, result1Map["id"])

		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		result2Map := result2.(map[string]any)
		assert.Equal(t, "Generated", result2Map["name"])
		assert.Equal(t, 2, result2Map["id"])

		// Valid input bypasses default generation
		validInput := map[string]any{
			"name": "Alice",
			"id":   100,
		}

		result3, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("prefault value", func(t *testing.T) {
		fallbackValue := map[string]any{
			"name": "Fallback Name",
			"age":  25, // Valid age that passes validation
		}

		schema := Object(core.ObjectSchema{
			"name": String().Min(5),
			"age":  Int().Min(18),
		}).Prefault(fallbackValue)

		// Valid input should pass through
		validInput := map[string]any{
			"name": "Alice",
			"age":  25,
		}

		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid input should use fallback
		invalidInput := map[string]any{
			"name": "A", // Too short
			"age":  10,  // Too young
		}

		result, err = schema.Parse(invalidInput)
		require.NoError(t, err)
		assert.Equal(t, fallbackValue, result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := Object(core.ObjectSchema{
			"name": String().Min(5),
			"id":   Int().Min(1),
		}).PrefaultFunc(func() map[string]any {
			counter++
			return map[string]any{
				"name": "Fallback",
				"id":   counter,
			}
		})

		// Valid input should pass through
		validInput := map[string]any{
			"name": "Alice",
			"id":   1,
		}

		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid input should use fallback function
		invalidInput := map[string]any{
			"name": "A", // Too short
			"id":   0,   // Too small
		}

		result, err = schema.Parse(invalidInput)
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, "Fallback", resultMap["name"])
		assert.Equal(t, 1, resultMap["id"])

		// Another invalid input should increment counter
		result2, err := schema.Parse(invalidInput)
		require.NoError(t, err)
		result2Map := result2.(map[string]any)
		assert.Equal(t, "Fallback", result2Map["name"])
		assert.Equal(t, 2, result2Map["id"])

		// Counter should only increment for invalid inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with chaining", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Default(map[string]any{
			"name": "Default",
			"age":  25,
		}).Pick([]string{"name"})

		// nil input should use default and then pick
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, "Default", resultMap["name"])
		_, hasAge := resultMap["age"]
		assert.False(t, hasAge, "Age should be omitted by Pick")
	})

	t.Run("prefault with chaining", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(5),
			"age":  Int().Min(18),
		}).Prefault(map[string]any{
			"name": "Fallback Name",
			"age":  20,
		}).Partial()

		// Invalid input should use fallback and then make partial
		invalidInput := map[string]any{
			"name": "A", // Too short
		}

		result, err := schema.Parse(invalidInput)
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Equal(t, "Fallback Name", resultMap["name"])
		assert.Equal(t, 20, resultMap["age"])
	})
}
