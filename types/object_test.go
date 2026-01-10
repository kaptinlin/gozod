package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestObject_BasicFunctionality(t *testing.T) {
	t.Run("valid object inputs", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String(),
		})

		validInput := map[string]any{
			"name":  "John",
			"age":   25,
			"email": "john@example.com",
		}

		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 25, result["age"])
		assert.Equal(t, "john@example.com", result["email"])
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		invalidInputs := []any{
			"not an object", 123, 3.14, []string{"array"}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		validInput := map[string]any{
			"name": "Alice",
			"age":  30,
		}

		// Test Parse method
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result["name"])
		assert.Equal(t, 30, result["age"])

		// Test MustParse method
		mustResult := schema.MustParse(validInput)
		assert.Equal(t, "Alice", mustResult["name"])
		assert.Equal(t, 30, mustResult["age"])

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid user object"
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}, core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeObject, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("field validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String().Min(2),
			"age":   Int().Min(0),
			"email": String().Email(),
		})

		// Valid input
		validInput := map[string]any{
			"name":  "John",
			"age":   25,
			"email": "john@example.com",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid field validation
		invalidInputs := []map[string]any{
			{"name": "J", "age": 25, "email": "john@example.com"},    // name too short
			{"name": "John", "age": -1, "email": "john@example.com"}, // age negative
			{"name": "John", "age": 25, "email": "invalid-email"},    // invalid email
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestObject_TypeSafety(t *testing.T) {
	t.Run("Object returns map[string]any type", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		require.NotNil(t, schema)

		input := map[string]any{
			"name": "test",
			"age":  123,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, map[string]any{}, result)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, 123, result["age"])
	})

	t.Run("Shape method returns object schema", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()

		schema := Object(core.ObjectSchema{
			"name": stringSchema,
			"age":  intSchema,
		})

		shape := schema.Shape()
		assert.Len(t, shape, 2)
		assert.Contains(t, shape, "name")
		assert.Contains(t, shape, "age")
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		input := map[string]any{
			"name": "test",
			"age":  123,
		}

		result := schema.MustParse(input)
		assert.IsType(t, map[string]any{}, result)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, 123, result["age"])
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestObject_Modifiers(t *testing.T) {
	t.Run("Optional modifier", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		optionalSchema := schema.Optional()

		// Test non-nil value
		input := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := optionalSchema.Parse(input)
		require.NoError(t, err)
		// Optional returns pointer constraint type
		assert.Equal(t, &input, result)

		// Test nil value (should be allowed for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		nilableSchema := schema.Nilable()

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value
		input := map[string]any{
			"name": "test",
			"age":  123,
		}
		result, err = nilableSchema.Parse(input)
		require.NoError(t, err)
		// Nilable returns pointer constraint type
		assert.Equal(t, &input, result)
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		nullishSchema := schema.Nullish()

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value
		input := map[string]any{
			"name": "test",
			"age":  123,
		}
		result, err = nullishSchema.Parse(input)
		require.NoError(t, err)
		// Nullish returns pointer constraint type
		assert.Equal(t, &input, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		defaultValue := map[string]any{
			"name": "default",
			"age":  0,
		}
		defaultSchema := schema.Default(defaultValue)

		// Valid input should override default
		input := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := defaultSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
		assert.IsType(t, map[string]any{}, result)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestObject_Chaining(t *testing.T) {
	t.Run("complex chaining", func(t *testing.T) {
		defaultValue := map[string]any{
			"name": "fallback",
			"age":  0,
		}
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Default(defaultValue).Optional()

		// Test final behavior
		input := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		// Chained with Optional returns pointer constraint type
		assert.Equal(t, &input, result)

		// Test nil handling - Default should short-circuit and return default value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, &defaultValue, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		defaultValue := map[string]any{
			"name": "default",
			"age":  0,
		}
		prefaultValue := map[string]any{
			"name": "prefault",
			"age":  -1,
		}
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Default(defaultValue).Prefault(prefaultValue)

		input := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("modifier immutability", func(t *testing.T) {
		originalSchema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
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

func TestObject_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// When both Default and Prefault are set, Default should take precedence
		schema := Object(core.ObjectSchema{
			"name": String(),
		}).Default(map[string]any{"name": "default_value"}).Prefault(map[string]any{"name": "prefault_value"}).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{"name": "default_value"}, *result)
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Default should bypass validation and return immediately
		schema := Object(core.ObjectSchema{
			"name": String().Min(10), // Strict validation that default won't pass
		}).Default(map[string]any{"name": "short"}).Optional() // "short" < 10 chars

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{"name": "short"}, *result)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value must pass object validation
		validObject := map[string]any{"name": "valid_prefault_name"}
		schema := Object(core.ObjectSchema{
			"name": String().Min(5),
		}).Prefault(validObject).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validObject, *result)
	})

	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		// Non-nil input that fails validation should not trigger Prefault
		schema := Object(core.ObjectSchema{
			"name": String().Min(10),
		}).Prefault(map[string]any{"name": "prefault_fallback"}).Optional()

		// Invalid input should fail validation, not use Prefault
		_, err := schema.Parse(map[string]any{"name": "short"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Object(core.ObjectSchema{
			"name": String(),
		}).DefaultFunc(func() map[string]any {
			defaultCalled = true
			return map[string]any{"name": "default_func"}
		}).PrefaultFunc(func() map[string]any {
			prefaultCalled = true
			return map[string]any{"name": "prefault_func"}
		}).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{"name": "default_func"}, *result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		// Prefault value that fails validation should return error
		invalidPrefault := map[string]any{"name": "x"} // Too short
		schema := Object(core.ObjectSchema{
			"name": String().Min(5),
		}).Prefault(invalidPrefault).Optional()

		_, err := schema.Parse(nil)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "Too small")
		}
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestObject_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		// Only accept objects with non-empty name and positive age
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(obj map[string]any) bool {
			name, hasName := obj["name"].(string)
			age, hasAge := obj["age"].(int)
			return hasName && hasAge && len(name) > 0 && age > 0
		})

		// Valid object
		validInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid object (empty name)
		invalidInput1 := map[string]any{
			"name": "",
			"age":  25,
		}
		_, err = schema.Parse(invalidInput1)
		assert.Error(t, err)

		// Invalid object (negative age)
		invalidInput2 := map[string]any{
			"name": "John",
			"age":  -5,
		}
		_, err = schema.Parse(invalidInput2)
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must have non-empty name and positive age"
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(obj map[string]any) bool {
			name, hasName := obj["name"].(string)
			age, hasAge := obj["age"].(int)
			return hasName && hasAge && len(name) > 0 && age > 0
		}, core.SchemaParams{Error: errorMessage})

		validInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		invalidInput := map[string]any{
			"name": "",
			"age":  25,
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

func TestObject_RefineAny(t *testing.T) {
	t.Run("refineAny validation", func(t *testing.T) {
		// Accept any object that has a "valid" field set to true
		schema := Object(core.ObjectSchema{
			"name":  String(),
			"valid": Bool(),
		}).RefineAny(func(v any) bool {
			if obj, ok := v.(map[string]any); ok {
				if valid, hasValid := obj["valid"].(bool); hasValid {
					return valid
				}
			}
			return false
		})

		// Valid object
		validInput := map[string]any{
			"name":  "John",
			"valid": true,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid object
		invalidInput := map[string]any{
			"name":  "John",
			"valid": false,
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestObject_ValidationMethods(t *testing.T) {
	t.Run("Min size validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Min(2)

		// Valid - has exactly 2 fields
		validInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Valid - has more than 2 fields (in strip mode, extra fields are removed)
		extraInput := map[string]any{
			"name":  "John",
			"age":   25,
			"extra": "field",
		}
		result, err = schema.Parse(extraInput)
		require.NoError(t, err)
		assert.Len(t, result, 2) // extra field should be stripped

		// Invalid - has less than 2 fields
		invalidInput := map[string]any{
			"name": "John",
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("Max size validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Optional(),
			"age":  Int().Optional(),
		}).Max(1)

		// Valid - has exactly 1 field
		validInput := map[string]any{
			"name": "John",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid - has more than 1 field (even after stripping)
		invalidInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("Size validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Size(2)

		// Valid - has exactly 2 fields
		validInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid - has different number of fields
		invalidInput1 := map[string]any{
			"name": "John",
		}
		_, err = schema.Parse(invalidInput1)
		assert.Error(t, err)

		// Extra fields are stripped, so this should pass size check
		extraInput := map[string]any{
			"name":  "John",
			"age":   25,
			"extra": "field",
		}
		result, err = schema.Parse(extraInput)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Property validation", func(t *testing.T) {
		// Create an object schema with property validation
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Property("name", String().Min(3))

		// Valid - name has at least 3 characters
		validInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid - name has less than 3 characters
		invalidInput := map[string]any{
			"name": "Jo",
			"age":  25,
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")

		// Valid - empty object shouldn't trigger property validation if field doesn't exist
		emptyInput := map[string]any{
			"age": 25,
		}
		_, err = schema.Parse(emptyInput)
		assert.Error(t, err) // Should fail due to missing required field, not property validation
	})

	t.Run("Property validation with custom error", func(t *testing.T) {
		customError := "Name must be at least 3 characters long"
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Property("name", String().Min(3), customError)

		invalidInput := map[string]any{
			"name": "Jo",
			"age":  25,
		}
		_, err := schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("Multiple property validations", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String(),
			"email": String(),
			"age":   Int(),
		}).
			Property("name", String().Min(2)).
			Property("email", String().Email()).
			Property("age", Int().Min(0).Max(150))

		// Valid input
		validInput := map[string]any{
			"name":  "John",
			"email": "john@example.com",
			"age":   25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid name
		invalidName := map[string]any{
			"name":  "J",
			"email": "john@example.com",
			"age":   25,
		}
		_, err = schema.Parse(invalidName)
		assert.Error(t, err)

		// Invalid email
		invalidEmail := map[string]any{
			"name":  "John",
			"email": "invalid-email",
			"age":   25,
		}
		_, err = schema.Parse(invalidEmail)
		assert.Error(t, err)

		// Invalid age
		invalidAge := map[string]any{
			"name":  "John",
			"email": "john@example.com",
			"age":   200,
		}
		_, err = schema.Parse(invalidAge)
		assert.Error(t, err)
	})
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestObject_TypeSpecificMethods(t *testing.T) {
	t.Run("Shape method returns field schemas", func(t *testing.T) {
		stringSchema := String()
		intSchema := Int()
		boolSchema := Bool()

		object := Object(core.ObjectSchema{
			"name":   stringSchema,
			"age":    intSchema,
			"active": boolSchema,
		})

		shape := object.Shape()
		assert.Len(t, shape, 3)

		// Verify all fields are present
		assert.Contains(t, shape, "name")
		assert.Contains(t, shape, "age")
		assert.Contains(t, shape, "active")
	})

	t.Run("Pick creates subset object", func(t *testing.T) {
		originalObject := Object(core.ObjectSchema{
			"name":    String(),
			"age":     Int(),
			"email":   String().Email(),
			"address": String().Optional(),
		})

		pickedObject, err := originalObject.Pick([]string{"name", "email"})
		require.NoError(t, err)

		// Should accept objects with only picked fields
		input := map[string]any{
			"name":  "John",
			"email": "john@example.com",
		}
		result, err := pickedObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should reject objects with required fields that weren't picked
		invalidInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		_, err = pickedObject.Parse(invalidInput)
		assert.Error(t, err) // missing email
	})

	t.Run("Omit creates filtered object", func(t *testing.T) {
		originalObject := Object(core.ObjectSchema{
			"name":     String(),
			"age":      Int(),
			"email":    String().Email(),
			"password": String().Min(8),
		})

		publicObject, err := originalObject.Omit([]string{"password"})
		require.NoError(t, err)

		// Should accept objects without omitted field
		input := map[string]any{
			"name":  "John",
			"age":   25,
			"email": "john@example.com",
		}
		result, err := publicObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should still validate remaining fields
		invalidInput := map[string]any{
			"name":  "John",
			"age":   25,
			"email": "invalid-email",
		}
		_, err = publicObject.Parse(invalidInput)
		assert.Error(t, err) // invalid email
	})

	t.Run("Extend adds new fields", func(t *testing.T) {
		baseObject := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		extendedObject, err := baseObject.Extend(core.ObjectSchema{
			"email":   String().Email(),
			"country": String(),
		})
		require.NoError(t, err)

		// Should accept objects with all fields
		input := map[string]any{
			"name":    "John",
			"age":     25,
			"email":   "john@example.com",
			"country": "USA",
		}
		result, err := extendedObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should reject objects missing extended fields
		invalidInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		_, err = extendedObject.Parse(invalidInput)
		assert.Error(t, err) // missing email and country
	})

	t.Run("Merge combines objects", func(t *testing.T) {
		object1 := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		object2 := Object(core.ObjectSchema{
			"email":   String().Email(),
			"country": String(),
		})

		mergedObject := object1.Merge(object2)

		// Should accept objects with all fields from both objects
		input := map[string]any{
			"name":    "John",
			"age":     25,
			"email":   "john@example.com",
			"country": "USA",
		}
		result, err := mergedObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("Partial makes fields optional", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String().Email(),
		})

		partialObject := object.Partial()

		// Should accept objects with only some fields
		input := map[string]any{
			"name": "John",
		}
		result, err := partialObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should accept objects with all fields
		fullInput := map[string]any{
			"name":  "John",
			"age":   25,
			"email": "john@example.com",
		}
		result, err = partialObject.Parse(fullInput)
		require.NoError(t, err)
		assert.Equal(t, fullInput, result)
	})

	t.Run("Partial with specific keys", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String().Email(),
		})

		// Make only name and age optional, email remains required
		partialObject := object.Partial([]string{"name", "age"})

		// Should accept objects with only required field
		input := map[string]any{
			"email": "john@example.com",
		}
		result, err := partialObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should reject objects missing required field
		invalidInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		_, err = partialObject.Parse(invalidInput)
		assert.Error(t, err) // missing required email
	})

	t.Run("Required makes specific fields required", func(t *testing.T) {
		partialObject := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String().Email(),
		}).Partial()

		// Make only name and email required
		requiredObject := partialObject.Required([]string{"name", "email"})

		// Should accept objects with required fields
		input := map[string]any{
			"name":  "John",
			"email": "john@example.com",
		}
		result, err := requiredObject.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should reject objects missing required field
		invalidInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		_, err = requiredObject.Parse(invalidInput)
		assert.Error(t, err) // missing required email
	})

	t.Run("Keyof returns string enum of keys", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String(),
		})

		keyEnum := object.Keyof()
		require.NotNil(t, keyEnum)

		// Should accept valid keys
		result, err := keyEnum.Parse("name")
		require.NoError(t, err)
		assert.Equal(t, "name", result)

		result, err = keyEnum.Parse("age")
		require.NoError(t, err)
		assert.Equal(t, "age", result)

		// Should reject invalid keys
		_, err = keyEnum.Parse("invalid")
		assert.Error(t, err)
	})

	// Pick returns error on invalid key (Go convention: default methods return error)
	t.Run("Pick with invalid key returns error", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		_, err := object.Pick([]string{"name", "invalid"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unrecognized key")
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("Pick with all invalid keys returns error", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		_, err := object.Pick([]string{"unknown"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unrecognized key")
		assert.Contains(t, err.Error(), "unknown")
	})

	t.Run("Omit with invalid key returns error", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		_, err := object.Omit([]string{"name", "invalid"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unrecognized key")
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("Omit with all invalid keys returns error", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		_, err := object.Omit([]string{"unknown"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unrecognized key")
		assert.Contains(t, err.Error(), "unknown")
	})

	t.Run("Pick on refined schema returns error", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(v map[string]any) bool {
			return v["age"].(int) >= 18
		})

		_, err := object.Pick([]string{"name"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "refinements")
	})

	t.Run("Omit on refined schema returns error", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(v map[string]any) bool {
			return v["age"].(int) >= 18
		})

		_, err := object.Omit([]string{"name"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "refinements")
	})

	// MustPick/MustOmit tests - panic on error (Go convention: Must prefix panics)
	t.Run("MustPick with valid keys succeeds", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String(),
		})

		picked := object.MustPick([]string{"name", "email"})
		require.NotNil(t, picked)

		result, err := picked.Parse(map[string]any{"name": "John", "email": "john@example.com"})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("MustPick with invalid key panics", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		assert.Panics(t, func() {
			object.MustPick([]string{"name", "invalid"})
		})
	})

	t.Run("MustPick on refined schema panics", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(v map[string]any) bool {
			return v["age"].(int) >= 18
		})

		assert.Panics(t, func() {
			object.MustPick([]string{"name"})
		})
	})

	t.Run("MustOmit with valid keys succeeds", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name":  String(),
			"age":   Int(),
			"email": String(),
		})

		omitted := object.MustOmit([]string{"email"})
		require.NotNil(t, omitted)

		result, err := omitted.Parse(map[string]any{"name": "John", "age": 25})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("MustOmit with invalid key panics", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		assert.Panics(t, func() {
			object.MustOmit([]string{"unknown"})
		})
	})

	t.Run("MustOmit on refined schema panics", func(t *testing.T) {
		object := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Refine(func(v map[string]any) bool {
			return v["age"].(int) >= 18
		})

		assert.Panics(t, func() {
			object.MustOmit([]string{"name"})
		})
	})
}

// =============================================================================
// Object modes tests
// =============================================================================

func TestObject_Modes(t *testing.T) {
	t.Run("Strip mode (default) removes unknown fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}) // Default is strip mode

		input := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "field",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 25, result["age"])
		assert.NotContains(t, result, "unknown")
	})

	t.Run("Strict mode rejects unknown fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Strict()

		// Valid input without unknown fields
		validInput := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid input with unknown fields
		invalidInput := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "field",
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("Passthrough mode allows unknown fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Passthrough()

		input := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "field",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 25, result["age"])
		assert.Equal(t, "field", result["unknown"])
	})

	t.Run("Catchall validates unknown fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Passthrough().Catchall(String())

		// Valid input with string unknown field
		validInput := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "string value",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid input with non-string unknown field
		invalidInput := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": 123, // Should be string
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestObject_ErrorHandling(t *testing.T) {
	t.Run("custom error messages", func(t *testing.T) {
		customError := "Custom object error"
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}, core.SchemaParams{Error: customError})

		_, err := schema.Parse("not an object")
		assert.Error(t, err)
	})

	t.Run("invalid type error structure", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		_, err := schema.Parse("not an object")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("missing required field error", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		input := map[string]any{
			"name": "John",
			// missing required "age" field
		}
		_, err := schema.Parse(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "age")
	})

	t.Run("field validation error", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name":  String().Min(5),
			"email": String().Email(),
		})

		input := map[string]any{
			"name":  "Jo", // Too short
			"email": "john@example.com",
		}
		_, err := schema.Parse(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("nil handling edge cases", func(t *testing.T) {
		// Regular object should reject nil
		schema := Object(core.ObjectSchema{
			"name": String(),
		})
		_, err := schema.Parse(nil)
		assert.Error(t, err)

		// Nilable object should accept nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Optional object should accept nil
		optionalSchema := schema.Optional()
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty context handling", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		})

		input := map[string]any{
			"name": "John",
		}

		// Should work without context
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Should work with empty context slice
		result, err = schema.Parse(input, []*core.ParseContext{}...)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestObject_EdgeCases(t *testing.T) {
	t.Run("empty object schema", func(t *testing.T) {
		schema := Object(core.ObjectSchema{})

		// Should accept empty object
		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)

		// Should strip unknown fields
		input := map[string]any{
			"unknown": "field",
		}
		result, err = schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("nested objects", func(t *testing.T) {
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
				"name": "John",
				"age":  25,
			},
			"country": "USA",
		}

		result, err := profileSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// Test invalid nested object
		invalidInput := map[string]any{
			"user": map[string]any{
				"name": "John",
				// missing age
			},
			"country": "USA",
		}
		_, err = profileSchema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("complex chaining with all modifiers", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Min(1).Max(10).Strict().Optional()

		// Test with valid input
		input := map[string]any{
			"name": "John",
			"age":  25,
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		// Complex chaining with Optional returns pointer constraint type
		assert.Equal(t, &input, result)

		// Test with nil (should be allowed due to Optional)
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("performance with large objects", func(t *testing.T) {
		// Create schema with many fields
		shape := make(core.ObjectSchema)
		for i := 0; i < 100; i++ {
			shape[fmt.Sprintf("field%d", i)] = String().Optional()
		}
		schema := Object(shape)

		// Create input with all fields
		input := make(map[string]any)
		for i := 0; i < 100; i++ {
			input[fmt.Sprintf("field%d", i)] = fmt.Sprintf("value%d", i)
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("map conversion edge cases", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		})

		// Test with different map types that are NOT currently convertible
		input := map[any]any{
			"name": "John",
		}

		_, err := schema.Parse(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})
}

// =============================================================================
// Constructors tests
// =============================================================================

func TestObject_Constructors(t *testing.T) {
	t.Run("Object constructor with default strip mode", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Should strip unknown fields by default
		input := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "field",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.NotContains(t, result, "unknown")
	})

	t.Run("StrictObject constructor", func(t *testing.T) {
		schema := StrictObject(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Should reject unknown fields
		input := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "field",
		}
		_, err := schema.Parse(input)
		assert.Error(t, err)
	})

	t.Run("LooseObject constructor", func(t *testing.T) {
		schema := LooseObject(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})

		// Should allow unknown fields
		input := map[string]any{
			"name":    "John",
			"age":     25,
			"unknown": "field",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Contains(t, result, "unknown")
	})

	t.Run("Object with custom parameters", func(t *testing.T) {
		customError := "Custom validation error"
		schema := Object(core.ObjectSchema{
			"name": String(),
		}, core.SchemaParams{
			Error: customError,
		})

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestObject_Overwrite(t *testing.T) {
	t.Run("basic object field transformation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		}).Overwrite(func(obj map[string]any) map[string]any {
			// Transform name to uppercase and increment age
			result := make(map[string]any)
			for k, v := range obj {
				switch k {
				case "name":
					if strVal, ok := v.(string); ok {
						result[k] = strings.ToUpper(strVal)
					} else {
						result[k] = v
					}
				case "age":
					if intVal, ok := v.(int); ok {
						result[k] = intVal + 1
					} else {
						result[k] = v
					}
				default:
					result[k] = v
				}
			}
			return result
		})

		input := map[string]any{
			"name": "alice",
			"age":  25,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"name": "ALICE",
			"age":  26,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("object field normalization", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"email": String(),
			"phone": String(),
		}).Overwrite(func(obj map[string]any) map[string]any {
			// Normalize email and phone formats
			result := make(map[string]any)
			for k, v := range obj {
				switch k {
				case "email":
					if strVal, ok := v.(string); ok {
						result[k] = strings.ToLower(strings.TrimSpace(strVal))
					} else {
						result[k] = v
					}
				case "phone":
					if strVal, ok := v.(string); ok {
						// Remove all non-digits for phone
						phone := strings.ReplaceAll(strVal, "-", "")
						phone = strings.ReplaceAll(phone, " ", "")
						phone = strings.ReplaceAll(phone, "(", "")
						phone = strings.ReplaceAll(phone, ")", "")
						result[k] = phone
					} else {
						result[k] = v
					}
				default:
					result[k] = v
				}
			}
			return result
		})

		input := map[string]any{
			"email": "  JOHN@EXAMPLE.COM  ",
			"phone": "(555) 123-4567",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"email": "john@example.com",
			"phone": "5551234567",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("adding computed fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"firstName": String(),
			"lastName":  String(),
		}).Overwrite(func(obj map[string]any) map[string]any {
			// Add full name field
			result := make(map[string]any)
			for k, v := range obj {
				result[k] = v
			}

			// Compute full name
			if firstName, ok := obj["firstName"].(string); ok {
				if lastName, ok := obj["lastName"].(string); ok {
					result["fullName"] = firstName + " " + lastName
				}
			}
			return result
		})

		input := map[string]any{
			"firstName": "Jane",
			"lastName":  "Doe",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"firstName": "Jane",
			"lastName":  "Doe",
			"fullName":  "Jane Doe",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"flag": Bool(),
		}).Overwrite(func(obj map[string]any) map[string]any {
			return obj // Identity transformation
		})

		input := map[string]any{
			"flag": true,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, map[string]any{}, result)
		assert.Equal(t, input, result)
	})

	t.Run("field removal transformation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"public":  String(),
			"private": String().Optional(),
		}).Overwrite(func(obj map[string]any) map[string]any {
			// Remove private fields
			result := make(map[string]any)
			for k, v := range obj {
				if !strings.HasPrefix(k, "private") {
					result[k] = v
				}
			}
			return result
		})

		input := map[string]any{
			"public":  "visible",
			"private": "hidden",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"public": "visible",
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestObject_Check(t *testing.T) {
	simpleShape := core.ObjectSchema{
		"name": String(),
	}

	t.Run("invalid object triggers issues", func(t *testing.T) {
		schema := Object(simpleShape).Check(func(value map[string]any, p *core.ParsePayload) {
			if value["name"] == "" {
				p.AddIssueWithMessage("name is required")
			}
		})

		_, err := schema.Parse(map[string]any{"name": ""})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})

	t.Run("pointer schema adapts", func(t *testing.T) {
		schema := ObjectPtr(simpleShape).Check(func(value *map[string]any, p *core.ParsePayload) {
			if value == nil || (*value)["name"] == "" {
				p.AddIssueWithMessage("empty name")
			}
		})

		_, err := schema.Parse(map[string]any{"name": ""})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})
}

// =============================================================================
// NonOptional tests for Object schema
// =============================================================================

func TestObject_NonOptional(t *testing.T) {
	// base schema
	shape := map[string]core.ZodSchema{
		"name": String(),
	}
	schema := Object(shape).NonOptional()

	// valid
	_, err := schema.Parse(map[string]any{"name": "leo"})
	require.NoError(t, err)

	// nil input should error with expected nonoptional
	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		require.Len(t, zErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zErr.Issues[0].Code)
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}

	// Optional().NonOptional() chain
	chain := Object(shape).Optional().NonOptional()
	_, err = chain.Parse(nil)
	assert.Error(t, err)

	// embedding inside another object
	outer := Object(map[string]core.ZodSchema{
		"inner": Object(shape).Optional().NonOptional(),
	})
	_, err = outer.Parse(map[string]any{"inner": map[string]any{"name": "leo"}})
	require.NoError(t, err)
	_, err = outer.Parse(map[string]any{"inner": nil})
	assert.Error(t, err)
}

// =============================================================================
// ExactOptional tests (TypeScript Zod v4: accepts absent keys, rejects explicit nil)
// =============================================================================

func TestObject_ExactOptional(t *testing.T) {
	t.Run("exactOptional accepts absent keys", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"a": String().ExactOptional(),
		})

		// Absent key should pass
		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("exactOptional accepts valid values", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"a": String().ExactOptional(),
		})

		// Present key with valid value should pass
		result, err := schema.Parse(map[string]any{"a": "hello"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"a": "hello"}, result)
	})

	t.Run("exactOptional rejects explicit nil", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"a": String().ExactOptional(),
		})

		// Explicit nil should fail
		_, err := schema.Parse(map[string]any{"a": nil})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("exactOptional vs optional comparison", func(t *testing.T) {
		optionalSchema := Object(core.ObjectSchema{"a": String().Optional()})
		exactOptionalSchema := Object(core.ObjectSchema{"a": String().ExactOptional()})

		// Both accept absent keys
		_, err1 := optionalSchema.Parse(map[string]any{})
		require.NoError(t, err1)
		_, err2 := exactOptionalSchema.Parse(map[string]any{})
		require.NoError(t, err2)

		// Both accept valid values
		_, err3 := optionalSchema.Parse(map[string]any{"a": "hi"})
		require.NoError(t, err3)
		_, err4 := exactOptionalSchema.Parse(map[string]any{"a": "hi"})
		require.NoError(t, err4)

		// optional() accepts explicit nil
		result, err5 := optionalSchema.Parse(map[string]any{"a": nil})
		require.NoError(t, err5)
		assert.Nil(t, result["a"])

		// exactOptional() rejects explicit nil
		_, err6 := exactOptionalSchema.Parse(map[string]any{"a": nil})
		require.Error(t, err6)
	})

	t.Run("exactOptional type inference", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"a": String().ExactOptional(),
			"b": String().Optional(),
		})

		// Test with both present
		result, err := schema.Parse(map[string]any{
			"a": "hello",
			"b": "world",
		})
		require.NoError(t, err)
		assert.Equal(t, "hello", result["a"])
		assert.Equal(t, "world", result["b"])

		// Test with only optional present as nil
		result2, err := schema.Parse(map[string]any{
			"b": nil, // optional accepts nil
		})
		require.NoError(t, err)
		assert.Nil(t, result2["b"])
	})

	t.Run("exactOptional with validation constraints", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(3).ExactOptional(),
		})

		// Absent key should pass (no validation needed)
		_, err := schema.Parse(map[string]any{})
		require.NoError(t, err)

		// Valid value should pass
		_, err = schema.Parse(map[string]any{"name": "John"})
		require.NoError(t, err)

		// Invalid value should fail validation
		_, err = schema.Parse(map[string]any{"name": "Jo"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 3")

		// Explicit nil should fail (exactOptional rejects nil)
		_, err = schema.Parse(map[string]any{"name": nil})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}

// =============================================================================
// Multiple Error Collection tests (TypeScript Zod v4 object behavior)
// =============================================================================

func TestObject_MultipleErrorCollection(t *testing.T) {
	t.Run("collects multiple field validation errors", func(t *testing.T) {
		// Create object schema with multiple field constraints
		schema := Object(core.ObjectSchema{
			"name":  String().Min(3),
			"age":   Int().Min(18),
			"email": String().Email(),
		})

		// Input with multiple validation failures
		input := map[string]any{
			"name":  "Jo",            // Too short (< 3)
			"age":   16,              // Too young (< 18)
			"email": "invalid-email", // Invalid email format
		}

		result, err := schema.Parse(input)
		require.Error(t, err)
		assert.Nil(t, result)

		// Check that we have all 3 field validation errors
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 3)

		// Verify each error has correct path and preserves original error codes (TypeScript Zod v4 behavior)
		// Note: Go map iteration order is not guaranteed, so we check by field name
		fieldErrorMap := make(map[string]core.ZodIssue)
		for _, issue := range zodErr.Issues {
			if len(issue.Path) == 1 {
				fieldName := issue.Path[0].(string)
				fieldErrorMap[fieldName] = issue
			}
		}

		// Verify name field error
		nameIssue, hasName := fieldErrorMap["name"]
		assert.True(t, hasName, "Should have name field error")
		if hasName {
			assert.Equal(t, core.TooSmall, nameIssue.Code, "Name error should preserve original too_small code")
			assert.Contains(t, nameIssue.Message, "at least 3", "Name error should mention length requirement")
		}

		// Verify age field error
		ageIssue, hasAge := fieldErrorMap["age"]
		assert.True(t, hasAge, "Should have age field error")
		if hasAge {
			assert.Equal(t, core.TooSmall, ageIssue.Code, "Age error should preserve original too_small code")
			assert.Contains(t, ageIssue.Message, "at least 18", "Age error should mention minimum requirement")
		}

		// Verify email field error
		emailIssue, hasEmail := fieldErrorMap["email"]
		assert.True(t, hasEmail, "Should have email field error")
		if hasEmail {
			assert.Equal(t, core.InvalidFormat, emailIssue.Code, "Email error should preserve original invalid_format code")
			assert.Contains(t, emailIssue.Message, "email", "Email error should mention email validation")
		}
	})

	t.Run("handles nested object multiple errors", func(t *testing.T) {
		// Create nested object schema
		schema := Object(core.ObjectSchema{
			"user": Object(core.ObjectSchema{
				"name": String().Min(3),
				"age":  Int().Min(18),
			}),
			"items": Slice[int](Int().Min(10)), // Array of integers >= 10
		})

		// Input with multiple nested validation failures
		input := map[string]any{
			"user": map[string]any{
				"name": "Jo", // Too short (< 3) - path should be ["user", "name"]
				"age":  16,   // Too young (< 18) - path should be ["user", "age"]
			},
			"items": []int{5, 8, 15}, // First two are < 10 - path should be ["items", 0] and ["items", 1]
		}

		result, err := schema.Parse(input)
		require.Error(t, err)
		assert.Nil(t, result)

		// Check that we have all 4 nested errors
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 4)

		// Count errors by type and verify paths
		userErrors := 0
		itemErrors := 0
		for _, issue := range zodErr.Issues {
			if len(issue.Path) >= 2 && issue.Path[0] == "user" {
				userErrors++
				assert.Equal(t, core.TooSmall, issue.Code, "User field error should preserve original too_small code")
				assert.True(t, issue.Path[1] == "name" || issue.Path[1] == "age", "User error should have correct field path")
			} else if len(issue.Path) >= 2 && issue.Path[0] == "items" {
				itemErrors++
				assert.Equal(t, core.TooSmall, issue.Code, "Items error should preserve original too_small code")
				assert.True(t, issue.Path[1] == 0 || issue.Path[1] == 1, "Items error should have correct index path")
			}
		}

		assert.Equal(t, 2, userErrors, "Should have 2 user field errors")
		assert.Equal(t, 2, itemErrors, "Should have 2 items validation errors")
	})

	t.Run("strict mode with unrecognized keys", func(t *testing.T) {
		// Create strict object schema
		schema := Object(core.ObjectSchema{
			"name": String().Min(3),
		}).Strict()

		// Input with field validation error AND unknown field
		input := map[string]any{
			"name":    "Jo",    // Too short (< 3)
			"unknown": "value", // Unknown field in strict mode
		}

		result, err := schema.Parse(input)
		require.Error(t, err)
		assert.Nil(t, result)

		// Check that we have both field error and unrecognized keys error
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 2)

		// Count error types
		fieldErrors := 0
		unrecognizedErrors := 0
		for _, issue := range zodErr.Issues {
			switch issue.Code {
			case core.TooSmall:
				fieldErrors++
				assert.Equal(t, []any{"name"}, issue.Path, "Field error should have correct path")
			case core.UnrecognizedKeys:
				unrecognizedErrors++
				assert.Equal(t, []any{}, issue.Path, "Unrecognized keys error should have empty path")
			case core.InvalidType, core.InvalidValue, core.InvalidFormat, core.InvalidUnion, core.InvalidKey, core.InvalidElement, core.TooBig, core.NotMultipleOf, core.Custom, core.InvalidSchema, core.InvalidDiscriminator, core.IncompatibleTypes, core.MissingRequired, core.TypeConversion, core.NilPointer:
				// These issue codes are not expected in this specific test
			default:
				// Handle unexpected issue codes gracefully
			}
		}

		assert.Equal(t, 1, fieldErrors, "Should have 1 field validation error")
		assert.Equal(t, 1, unrecognizedErrors, "Should have 1 unrecognized keys error")
	})

	t.Run("successful validation with no errors", func(t *testing.T) {
		// Create object schema
		schema := Object(core.ObjectSchema{
			"name": String().Min(3),
			"age":  Int().Min(18),
		})

		// Valid input
		input := map[string]any{
			"name": "John",
			"age":  25,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 25, result["age"])
	})
}

// =============================================================================
// SafeExtend tests - Zod v4 Compatible
// =============================================================================

func TestObject_SafeExtend(t *testing.T) {
	t.Run("SafeExtend allows overwriting existing keys", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"email": String(),
		})

		// SafeExtend with more specific validation
		extended := schema.SafeExtend(core.ObjectSchema{
			"email": String().Email(), // Override with more specific validation
		})

		// Invalid email should now fail
		_, err := extended.Parse(map[string]any{"email": "not-an-email"})
		require.Error(t, err)

		// Valid email should pass
		result, err := extended.Parse(map[string]any{"email": "test@example.com"})
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result["email"])
	})

	t.Run("SafeExtend chaining preserves and overrides properties", func(t *testing.T) {
		// From Zod v4 test: safeExtend chaining preserves and overrides properties
		schema1 := Object(core.ObjectSchema{
			"email": String(),
		})

		schema2 := schema1.SafeExtend(core.ObjectSchema{
			"email": String().Email(),
		})

		schema3 := schema2.SafeExtend(core.ObjectSchema{
			"name": String(),
		})

		result, err := schema3.Parse(map[string]any{
			"email": "test@example.com",
			"name":  "John",
		})
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result["email"])
		assert.Equal(t, "John", result["name"])
	})

	t.Run("SafeExtend works with refined schemas", func(t *testing.T) {
		// Unlike Extend, SafeExtend should work with refined schemas
		schema := Object(core.ObjectSchema{
			"name": String(),
		}).Refine(func(m map[string]any) bool {
			return m["name"] != ""
		})

		// SafeExtend should not return error for refined schemas
		extended := schema.SafeExtend(core.ObjectSchema{
			"age": Int(),
		})

		result, err := extended.Parse(map[string]any{
			"name": "John",
			"age":  30,
		})
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 30, result["age"])
	})

	t.Run("SafeExtend adds new fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String(),
		})

		extended := schema.SafeExtend(core.ObjectSchema{
			"age":   Int(),
			"email": String().Email(),
		})

		result, err := extended.Parse(map[string]any{
			"name":  "John",
			"age":   30,
			"email": "john@example.com",
		})
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 30, result["age"])
		assert.Equal(t, "john@example.com", result["email"])
	})

	t.Run("SafeExtend preserves immutability", func(t *testing.T) {
		original := Object(core.ObjectSchema{
			"name": String(),
		})

		extended := original.SafeExtend(core.ObjectSchema{
			"age": Int(),
		})

		// Original should not be affected
		originalShape := original.Shape()
		assert.Len(t, originalShape, 1)
		assert.Contains(t, originalShape, "name")
		assert.NotContains(t, originalShape, "age")

		// Extended should have both fields
		extendedShape := extended.Shape()
		assert.Len(t, extendedShape, 2)
		assert.Contains(t, extendedShape, "name")
		assert.Contains(t, extendedShape, "age")
	})

	t.Run("SafeExtend with constructor field in shape", func(t *testing.T) {
		// From Zod v4 test: extend with constructor field in shape
		baseSchema := Object(core.ObjectSchema{
			"name": String(),
		})

		extendedSchema := baseSchema.SafeExtend(core.ObjectSchema{
			"constructor": String(),
			"age":         Int(),
		})

		result, err := extendedSchema.Parse(map[string]any{
			"name":        "John",
			"constructor": "Person",
			"age":         30,
		})
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, "Person", result["constructor"])
		assert.Equal(t, 30, result["age"])
	})
}
