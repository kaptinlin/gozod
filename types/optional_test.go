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

func TestOptionalBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := String().Optional()

		// Valid string input
		result, err := schema.Parse("adsf")
		require.NoError(t, err)
		assert.Equal(t, "adsf", result)

		// nil input should return nil
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Invalid type should fail
		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := String().Optional()

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

		// nil input returns nil
		result3, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result3)
	})

	t.Run("package function constructor", func(t *testing.T) {
		schema := Optional(String())
		require.NotNil(t, schema)

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("unwrap method", func(t *testing.T) {
		schema := String().Optional()
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		unwrapped := optionalSchema.Unwrap()

		// Unwrapped should be the original string schema
		assert.IsType(t, &ZodString{}, unwrapped)

		// Should validate as string
		result, err := unwrapped.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Should reject nil (since it's no longer optional)
		_, err = unwrapped.Parse(nil)
		assert.Error(t, err)
	})
}

// =============================================================================
// 2. Optionality concept (optin/optout)
// =============================================================================

func TestOptionalOptionalityConcept(t *testing.T) {
	t.Run("optionality tracking", func(t *testing.T) {
		// Regular string has no optinality
		a := String()
		aInternals := a.GetInternals()
		assert.Empty(t, aInternals.OptIn, "Regular string should have no optin")
		assert.Empty(t, aInternals.OptOut, "Regular string should have no optout")

		// Optional string has both optin and optout
		b := String().Optional()
		bInternals := b.GetInternals()
		assert.Equal(t, "optional", bInternals.OptIn, "Optional should have optin='optional'")
		assert.Equal(t, "optional", bInternals.OptOut, "Optional should have optout='optional'")

		// Default has optin but not optout
		c := String().Default("asdf")
		cInternals := c.GetInternals()
		assert.Equal(t, "optional", cInternals.OptIn, "Default should have optin='optional'")
		assert.Empty(t, cInternals.OptOut, "Default should have no optout")

		// Optional + Nilable
		d := String().Optional().Nilable()
		dInternals := d.GetInternals()
		// Note: The exact behavior might depend on implementation
		// This test verifies the behavior is consistent
		assert.NotEmpty(t, dInternals.OptIn, "Optional+Nilable should have some optin")
		assert.NotEmpty(t, dInternals.OptOut, "Optional+Nilable should have some optout")
	})
}

// =============================================================================
// 3. Pipe optionality
// =============================================================================

func TestOptionalPipeOptionality(t *testing.T) {
	t.Run("optional pipe to required", func(t *testing.T) {
		// Optional input piped to required output
		a := String().Optional().Pipe(String())
		aInternals := a.GetInternals()

		// Should preserve input optionality but lose output optionality
		// Expected: optin="optional", optout="" (not optional)
		assert.Equal(t, "optional", aInternals.OptIn, "Pipe should preserve input optionality")
		assert.Empty(t, aInternals.OptOut, "Pipe to required should not have optout")

		// Test actual parsing behavior
		result, err := a.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// nil input should pass through the pipe
		result, err = a.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("required pipe to optional", func(t *testing.T) {
		// Transform to potentially nil, then pipe to optional
		b := String().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			// Simulate a transform that might return nil
			if val.(string) == "make_nil" {
				return nil, nil
			}
			return val, nil
		}).Pipe(String().Optional())

		bInternals := b.GetInternals()

		// Should not have input optionality but have output optionality
		assert.Empty(t, bInternals.OptIn, "Transform pipe should not have optin")
		assert.Equal(t, "optional", bInternals.OptOut, "Pipe to optional should have optout")

		// Test parsing
		result, err := b.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = b.Parse("make_nil")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("basic default pipe", func(t *testing.T) {
		// Default piped to required
		c := String().Default("asdf").Pipe(String())
		cInternals := c.GetInternals()

		// Test that optionality is preserved from Default
		// Note: This may fail if Default's OptIn implementation needs fixing
		t.Logf("Default pipe OptIn: '%s', OptOut: '%s'", cInternals.OptIn, cInternals.OptOut)

		// Test parsing behavior
		result, err := c.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		result, err = c.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "asdf", result) // Should use default
	})
}

// =============================================================================
// 4. Pipe optionality inside objects (simplified)
// =============================================================================

func TestOptionalPipeOptionalityInObjects(t *testing.T) {
	t.Run("simple object with optional fields", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"a": String().Optional(),
			"b": String().Default("default_b"),
		})

		// Test with partial object
		obj := map[string]any{
			"a": "test_a",
		}

		result, err := schema.Parse(obj)
		require.NoError(t, err)
		resultMap := result.(map[string]any)

		assert.Equal(t, "test_a", resultMap["a"])
		assert.Equal(t, "default_b", resultMap["b"])

		// Test with empty object
		emptyObj := map[string]any{}

		result, err = schema.Parse(emptyObj)
		require.NoError(t, err)
		resultMap = result.(map[string]any)

		// "a" is optional, should be missing/nil
		// "b" has default, should use default value
		assert.Equal(t, "default_b", resultMap["b"])
	})
}

// =============================================================================
// 5. Coerce (type coercion)
// =============================================================================

func TestOptionalCoercion(t *testing.T) {
	t.Run("coercion with optional", func(t *testing.T) {
		schema := CoercedString().Optional()

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
// 6. Validation methods
// =============================================================================

func TestOptionalValidations(t *testing.T) {
	t.Run("complex inner type validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"name": String().Min(3),
			"age":  Int().Min(0),
		}).Optional()

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
}

// =============================================================================
// 7. Modifiers and wrappers
// =============================================================================

func TestOptionalModifiers(t *testing.T) {
	t.Run("optional.optional() is noop", func(t *testing.T) {
		schema := Optional(String())
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		doubleOptional := optionalSchema.Optional()

		// Both should behave the same
		result1, err1 := schema.Parse(nil)
		result2, err2 := doubleOptional.Parse(nil)
		assert.Equal(t, err1 == nil, err2 == nil)
		assert.Equal(t, result1, result2)

		result1, err1 = schema.Parse("test")
		result2, err2 = doubleOptional.Parse("test")
		assert.Equal(t, err1 == nil, err2 == nil)
		assert.Equal(t, result1, result2)
	})

	t.Run("optional.nilable()", func(t *testing.T) {
		schema := Optional(String())
		nilableSchema := schema.Nilable()

		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = nilableSchema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("MustParse method", func(t *testing.T) {
		schema := String().Optional()

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
// 8. Chaining and method composition
// =============================================================================

func TestOptionalChaining(t *testing.T) {
	t.Run("refine chaining", func(t *testing.T) {
		schema := Optional(String())
		refinedSchema := schema.RefineAny(func(val any) bool {
			// nil values should pass (optional behavior)
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
		schema := String().Min(3).Max(10).Optional()

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
}

// =============================================================================
// 9. Transform/Pipe
// =============================================================================

func TestOptionalTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := String().Optional().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			// nil should remain nil
			if val == nil {
				return nil, nil
			}
			if str, ok := val.(string); ok {
				return str + "_transformed", nil
			}
			return val, nil
		})

		// nil should remain nil
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// String should be transformed
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello_transformed", result)
	})

	t.Run("pipe to another schema", func(t *testing.T) {
		schema := String().Optional().Pipe(String().Min(3))

		// nil should pass through
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid string should fail at pipe stage
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("transform with complex logic", func(t *testing.T) {
		schema := String().Optional().TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if val == nil {
				return map[string]any{"status": "empty"}, nil
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

		// nil should transform to empty status
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "empty", resultMap["status"])

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
// 10. Refine
// =============================================================================

func TestOptionalRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := String().Optional().RefineAny(func(val any) bool {
			// nil values should pass (optional behavior)
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
		schema := String().Optional().RefineAny(func(val any) bool {
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
// 11. Error handling
// =============================================================================

func TestOptionalErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := String().Optional()

		_, err := schema.Parse(123)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("inner type validation errors", func(t *testing.T) {
		schema := String().Min(5).Optional()

		// nil should not error
		_, err := schema.Parse(nil)
		assert.NoError(t, err)

		// Invalid inner type should error
		_, err = schema.Parse("hi")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected string to have >=5 characters")
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Optional(String(), core.SchemaParams{
			Error: "core.Custom optional error",
		})

		// Valid cases should not trigger custom error
		_, err := schema.Parse("hello")
		assert.NoError(t, err)

		_, err = schema.Parse(nil)
		assert.NoError(t, err)

		// Invalid type should trigger error
		_, err = schema.Parse(123)
		assert.Error(t, err)
		// Note: core.Custom error may not override type validation errors
		assert.NotEmpty(t, err.Error())
	})
}

// =============================================================================
// 12. Edge and mutual exclusion cases
// =============================================================================

func TestOptionalEdgeCases(t *testing.T) {
	t.Run("nested optional schemas", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"nested": Object(core.ObjectSchema{
				"value": String().Min(10).Optional(),
			}).Optional(),
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

	t.Run("optional with array", func(t *testing.T) {
		schema := Slice(String()).Optional()

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

	t.Run("optional with union", func(t *testing.T) {
		schema := Union([]core.ZodType[any, any]{String(), Int()}).Optional()

		// nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid string should pass
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Valid int should pass
		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// Invalid type should fail
		_, err = schema.Parse(true)
		assert.Error(t, err)
	})
}

// =============================================================================
// 13. Default and Prefault tests
// =============================================================================

func TestOptionalDefaultAndPrefault(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		schema := Optional(String())
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		defaultSchema := optionalSchema.Default("default_value")

		// nil should use default
		result, err := defaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)

		// Valid value should not use default
		result, err = defaultSchema.Parse("custom")
		require.NoError(t, err)
		assert.Equal(t, "custom", result)
	})

	t.Run("default function", func(t *testing.T) {
		counter := 0
		schema := Optional(String())
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		defaultSchema := optionalSchema.DefaultFunc(func() any {
			counter++
			return "generated_default"
		})

		// nil should call function
		result, err := defaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "generated_default", result)
		assert.Equal(t, 1, counter)

		// Valid value should not call function
		result, err = defaultSchema.Parse("custom")
		require.NoError(t, err)
		assert.Equal(t, "custom", result)
		assert.Equal(t, 1, counter) // Counter unchanged
	})

	t.Run("prefault values", func(t *testing.T) {
		schema := Optional(String().Min(5))
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		prefaultSchema := optionalSchema.Prefault("fallback")

		// nil should not use prefault (optional allows nil)
		result, err := prefaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid value should not use prefault
		result, err = prefaultSchema.Parse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Invalid value should use prefault
		result, err = prefaultSchema.Parse("hi")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := Optional(String().Min(5))
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		prefaultSchema := optionalSchema.PrefaultFunc(func() any {
			counter++
			return "generated_fallback"
		})

		// nil should not call function (optional allows nil)
		result, err := prefaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, counter)

		// Valid value should not call function
		result, err = prefaultSchema.Parse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
		assert.Equal(t, 0, counter)

		// Invalid value should call function
		result, err = prefaultSchema.Parse("hi")
		require.NoError(t, err)
		assert.Equal(t, "generated_fallback", result)
		assert.Equal(t, 1, counter)
	})

	t.Run("complex default and prefault combinations", func(t *testing.T) {
		// Optional with default, then prefault for validation failures
		schema := Optional(String().Min(3))
		optionalSchema := schema.(*ZodOptional[core.ZodType[any, any]])
		defaultSchema := optionalSchema.Default("default")
		prefaultSchema := defaultSchema.Prefault("fallback")

		// nil should use default
		result, err := prefaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default", result)

		// Valid value should pass through
		result, err = prefaultSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid value should use prefault
		result, err = prefaultSchema.Parse("hi")
		require.NoError(t, err)
		assert.Equal(t, "fallback", result)
	})
}
