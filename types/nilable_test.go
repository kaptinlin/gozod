package types

import (
	"reflect"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestNilableBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := String().Nilable()

		// Valid string input
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// nil input should return nil
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Invalid type should fail
		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := String().Nilable()

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

		// nil input returns typed nil pointer
		result3, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result3)
		// Note: For string nilable, nil should return (*string)(nil)
		assert.IsType(t, (*string)(nil), result3)
	})

	t.Run("package function constructor", func(t *testing.T) {
		schema := Nilable(String())
		require.NotNil(t, schema)

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("with inner type validation", func(t *testing.T) {
		schema := String().Min(5).Nilable()

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = schema.Parse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Invalid string should fail
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("different types nilable", func(t *testing.T) {
		// String nilable
		stringSchema := String().Nilable()
		result, err := stringSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Int nilable
		intSchema := Int().Nilable()
		result, err = intSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Bool nilable
		boolSchema := Bool().Nilable()
		result, err = boolSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestNilableCoercion(t *testing.T) {
	t.Run("coercion with nilable", func(t *testing.T) {
		schema := String(core.SchemaParams{Coerce: true}).Nilable()

		// nil should remain nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Coercible values should be coerced
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestNilableValidations(t *testing.T) {
	t.Run("unwrap method", func(t *testing.T) {
		schema := Nilable(String())
		unwrapped := schema.Unwrap()

		// Unwrapped should be the original string schema
		assert.IsType(t, &ZodString{}, unwrapped)

		// Should validate as string
		result, err := unwrapped.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Should reject nil (since it's no longer nilable)
		_, err = unwrapped.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("complex inner type validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(3),
			"age":  Int().Min(0),
		}).Nilable()

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid object should pass
		obj := map[string]any{"name": "Alice", "age": 25}
		result, err = schema.Parse(obj)
		require.NoError(t, err)
		assert.Equal(t, obj, result)

		// Invalid object should fail
		invalidObj := map[string]any{"name": "Al", "age": -5}
		_, err = schema.Parse(invalidObj)
		assert.Error(t, err)
	})

	t.Run("array nilable validation", func(t *testing.T) {
		schema := Slice(String()).Nilable()

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid array should pass
		arr := []any{"hello", "world"}
		result, err = schema.Parse(arr)
		require.NoError(t, err)
		assert.Equal(t, arr, result)

		// Invalid array element should fail
		invalidArr := []any{"hello", 123}
		_, err = schema.Parse(invalidArr)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestNilableModifiers(t *testing.T) {
	t.Run("nilable.nilable() is noop", func(t *testing.T) {
		schema := String().Nilable()
		doubleNilable := schema.Nilable()

		// Both should behave the same
		result1, err1 := schema.Parse(nil)
		result2, err2 := doubleNilable.Parse(nil)
		assert.Equal(t, err1 == nil, err2 == nil)
		assert.Equal(t, result1, result2)

		result1, err1 = schema.Parse("test")
		result2, err2 = doubleNilable.Parse("test")
		assert.Equal(t, err1 == nil, err2 == nil)
		assert.Equal(t, result1, result2)
	})

	t.Run("MustParse method", func(t *testing.T) {
		schema := String().Nilable()

		// Should not panic for valid values
		result := schema.MustParse("hello")
		assert.Equal(t, "hello", result)

		result = schema.MustParse(nil)
		assert.Nil(t, result)

		// Should panic for invalid values
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestNilableChaining(t *testing.T) {
	t.Run("refine chaining", func(t *testing.T) {
		schema := String().Nilable()
		refinedSchema := schema.RefineAny(func(val any) bool {
			// nil values should pass (nilable behavior)
			if val == nil {
				return true
			}
			if str, ok := val.(string); ok {
				return len(str) > 3
			}
			return false
		}, core.SchemaParams{
			Error: "String must be longer than 3 characters when present",
		})

		// nil should pass
		result, err := refinedSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = refinedSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid string should fail
		_, err = refinedSchema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("multiple validation chaining", func(t *testing.T) {
		schema := String().Min(3).Max(10).Nilable()

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Too short should fail
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		// Too long should fail
		_, err = schema.Parse("this is too long")
		assert.Error(t, err)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		schema := String().Min(2).Nilable()
		input := "hello"
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify not only type and value, but exact pointer identity
		resultPtr, ok := result.(*string)
		require.True(t, ok, "Result should be *string")
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
		assert.Equal(t, "hello", *resultPtr)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestNilableTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := String().Nilable().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			// Nilable passes typed nil pointer to transform function
			if val == nil || (val != nil && reflect.ValueOf(val).Kind() == reflect.Ptr && reflect.ValueOf(val).IsNil()) {
				return "nil_transformed", nil
			}
			if str, ok := val.(string); ok {
				return str + "_transformed", nil
			}
			return val, nil
		})

		// nil should be transformed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "nil_transformed", result)

		// String should be transformed
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello_transformed", result)
	})

	t.Run("transform with complex logic", func(t *testing.T) {
		schema := String().Nilable().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			// Check for nil or typed nil pointer
			if val == nil || (val != nil && reflect.ValueOf(val).Kind() == reflect.Ptr && reflect.ValueOf(val).IsNil()) {
				return map[string]any{"status": "null"}, nil
			}
			if str, ok := val.(string); ok {
				return map[string]any{
					"status": "present",
					"value":  str,
					"length": len(str),
				}, nil
			}
			return val, nil
		})

		// nil should transform to null status
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "null", resultMap["status"])

		// String should transform to present status
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		resultMap, ok = result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "present", resultMap["status"])
		assert.Equal(t, "hello", resultMap["value"])
		assert.Equal(t, 5, resultMap["length"])
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestNilableRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := String().Nilable().RefineAny(func(val any) bool {
			// nil values should pass (nilable behavior)
			if val == nil {
				return true
			}
			if str, ok := val.(string); ok {
				return len(str) > 3
			}
			return false
		})

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid string should fail
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := String().Nilable().RefineAny(func(val any) bool {
			if val == nil {
				return true
			}
			if str, ok := val.(string); ok {
				return str != "forbidden"
			}
			return false
		}, core.SchemaParams{
			Error: "This value is forbidden",
		})

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = schema.Parse("allowed")
		require.NoError(t, err)
		assert.Equal(t, "allowed", result)

		// Forbidden string should fail
		_, err = schema.Parse("forbidden")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "This value is forbidden")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestNilableErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := String().Nilable()

		_, err := schema.Parse(123)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("inner type validation errors", func(t *testing.T) {
		schema := String().Min(5).Nilable()

		// nil should not error
		_, err := schema.Parse(nil)
		assert.NoError(t, err)

		// Invalid inner type should error
		_, err = schema.Parse("hi")
		assert.Error(t, err)
		// Check for the actual error message format
		assert.Contains(t, err.Error(), "Too small: expected string to have >=5 characters")
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestNilableEdgeCases(t *testing.T) {
	t.Run("nested nilable schemas", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"nested": Object(core.ObjectSchema{
				"value": String().Min(10).Nilable(),
			}).Nilable(),
		})

		// Completely nil should pass
		result, err := schema.Parse(map[string]any{
			"nested": nil,
		})
		require.NoError(t, err)
		resultMap := result.(map[string]any)
		assert.Nil(t, resultMap["nested"])

		// Nested object with nil value should pass
		result, err = schema.Parse(map[string]any{
			"nested": map[string]any{
				"value": nil,
			},
		})
		require.NoError(t, err)
		resultMap = result.(map[string]any)
		nestedMap := resultMap["nested"].(map[string]any)
		assert.Nil(t, nestedMap["value"])

		// Valid nested value should pass
		result, err = schema.Parse(map[string]any{
			"nested": map[string]any{
				"value": "hello world",
			},
		})
		require.NoError(t, err)
		resultMap = result.(map[string]any)
		nestedMap = resultMap["nested"].(map[string]any)
		assert.Equal(t, "hello world", nestedMap["value"])

		// Invalid nested value should fail
		_, err = schema.Parse(map[string]any{
			"nested": map[string]any{
				"value": "short",
			},
		})
		assert.Error(t, err)
	})

	t.Run("side effect isolation", func(t *testing.T) {
		baseSchema := String().Min(3)
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse("hello")
		require.NoError(t, err2)
		assert.Equal(t, "hello", result2)

		// Test nilable schema rejects invalid values
		_, err3 := nilableSchema.Parse("hi")
		assert.Error(t, err3)

		// Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse("hello")
		require.NoError(t, err5)
		assert.Equal(t, "hello", result5)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestNilableDefaultAndPrefault(t *testing.T) {
	t.Run("nilable behavior vs optional behavior", func(t *testing.T) {
		nilableSchema := String().Nilable()
		optionalSchema := String().Optional()

		// Both should accept nil
		nilableResult, nilableErr := nilableSchema.Parse(nil)
		optionalResult, optionalErr := optionalSchema.Parse(nil)

		require.NoError(t, nilableErr)
		require.NoError(t, optionalErr)
		assert.Nil(t, nilableResult)
		assert.Nil(t, optionalResult)

		// Both should accept valid strings
		nilableResult, nilableErr = nilableSchema.Parse("hello")
		optionalResult, optionalErr = optionalSchema.Parse("hello")

		require.NoError(t, nilableErr)
		require.NoError(t, optionalErr)
		assert.Equal(t, "hello", nilableResult)
		assert.Equal(t, "hello", optionalResult)

		// Both should reject invalid types
		_, nilableErr = nilableSchema.Parse(123)
		_, optionalErr = optionalSchema.Parse(123)

		assert.Error(t, nilableErr)
		assert.Error(t, optionalErr)
	})
}
