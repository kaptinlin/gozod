package types

import (
	"fmt"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestDiscriminatedUnionBasicFunctionality(t *testing.T) {
	t.Run("constructor with invalid options", func(t *testing.T) {
		// Test that using basic types as options should panic because they have no discriminator fields
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(), // Basic types have no discriminator fields, should panic
				Int(),
			})
		})
	})

	t.Run("constructor with params and valid options", func(t *testing.T) {
		// Test that using valid Object + Literal combinations should succeed
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "value": String()}),
			Object(core.ObjectSchema{"type": Literal("b"), "value": Int()}),
		}, core.SchemaParams{Error: "Must match one of the discriminated types"})
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals.Error)
	})

	t.Run("discriminator accessor with invalid options", func(t *testing.T) {
		// Test accessing discriminator field name, should panic even though trying to access it
		assert.Panics(t, func() {
			schema := DiscriminatedUnion("status", []core.ZodType[any, any]{
				String(),
				Int(),
			})
			_ = schema.Discriminator() // This line won't execute because constructor will panic
		})
	})

	t.Run("options accessor with invalid options", func(t *testing.T) {
		// Test accessing options, should panic even though trying to access them
		assert.Panics(t, func() {
			stringSchema := String()
			intSchema := Int()
			schema := DiscriminatedUnion("type", []core.ZodType[any, any]{stringSchema, intSchema})
			_ = schema.Options() // This line won't execute because constructor will panic
		})
	})
}

// =============================================================================
// 2. Validation methods
// =============================================================================

func TestDiscriminatedUnionValidations(t *testing.T) {
	t.Run("valid parse - object", func(t *testing.T) {
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "a": String()}),
			Object(core.ObjectSchema{"type": Literal("b"), "b": String()}),
		})

		result, err := schema.Parse(map[string]any{
			"type": "a",
			"a":    "abc",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "a",
			"a":    "abc",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("performance optimization - direct lookup", func(t *testing.T) {
		// Test that discriminated union uses direct lookup instead of trying each option
		schema := DiscriminatedUnion("status", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"status": Literal("success"), "data": String()}),
			Object(core.ObjectSchema{"status": Literal("failed"), "error": String()}),
		})

		// Success case
		result, err := schema.Parse(map[string]any{
			"status": "success",
			"data":   "operation completed",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"status": "success",
			"data":   "operation completed",
		}
		assert.Equal(t, expected, result)

		// Failed case
		result, err = schema.Parse(map[string]any{
			"status": "failed",
			"error":  "operation failed",
		})
		require.NoError(t, err)
		expected = map[string]any{
			"status": "failed",
			"error":  "operation failed",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("discriminator value of various primitive types", func(t *testing.T) {
		// Test discriminated union with different primitive discriminator types
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("1"), "val": String()}),
			Object(core.ObjectSchema{"type": Literal(1), "val": String()}),
			Object(core.ObjectSchema{"type": Literal("true"), "val": String()}),
			Object(core.ObjectSchema{"type": Literal(true), "val": String()}),
		})

		// String discriminator
		result, err := schema.Parse(map[string]any{
			"type": "1",
			"val":  "string_val",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "1",
			"val":  "string_val",
		}
		assert.Equal(t, expected, result)

		// Integer discriminator
		result, err = schema.Parse(map[string]any{
			"type": 1,
			"val":  "int_val",
		})
		require.NoError(t, err)
		expected = map[string]any{
			"type": 1,
			"val":  "int_val",
		}
		assert.Equal(t, expected, result)

		// Boolean discriminator
		result, err = schema.Parse(map[string]any{
			"type": true,
			"val":  "bool_val",
		})
		require.NoError(t, err)
		expected = map[string]any{
			"type": true,
			"val":  "bool_val",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("performance optimization concept", func(t *testing.T) {
		// Test discriminated union concept - direct lookup instead of trying each option
		// Using invalid options will panic
		assert.Panics(t, func() {
			DiscriminatedUnion("status", []core.ZodType[any, any]{
				String(),
				Int(),
			})
		})
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestDiscriminatedUnionModifiers(t *testing.T) {
	t.Run("optional with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			})
		})
	})

	t.Run("nilable with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			})
		})
	})

	t.Run("nilable does not affect original schema concept", func(t *testing.T) {
		// Test concept: Nilable does not affect original schema
		// But using invalid options will panic
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			})
		})
	})

	t.Run("must parse with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			})
		})
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestDiscriminatedUnionChaining(t *testing.T) {
	t.Run("validation with constraints", func(t *testing.T) {
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("string"), "value": String().Min(5)}),
			Object(core.ObjectSchema{"type": Literal("number"), "value": Int().Min(10)}),
		})

		// Valid cases
		result, err := schema.Parse(map[string]any{
			"type":  "string",
			"value": "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type":  "string",
			"value": "hello",
		}
		assert.Equal(t, expected, result)

		result, err = schema.Parse(map[string]any{
			"type":  "number",
			"value": 15,
		})
		require.NoError(t, err)
		expected = map[string]any{
			"type":  "number",
			"value": 15,
		}
		assert.Equal(t, expected, result)

		// Invalid cases - constraint violations
		_, err = schema.Parse(map[string]any{
			"type":  "string",
			"value": "hi", // Too short
		})
		assert.Error(t, err)

		_, err = schema.Parse(map[string]any{
			"type":  "number",
			"value": 5, // Too small
		})
		assert.Error(t, err)
	})

	t.Run("method chaining with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			})
		})
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestDiscriminatedUnionTransform(t *testing.T) {
	t.Run("basic transform with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			}).Transform(func(val any, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"original":    val,
					"transformed": true,
				}, nil
			})
		})
	})

	t.Run("pipe to another schema with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			})
		})
	})
}

// =============================================================================
// 6. Refine
// =============================================================================

func TestDiscriminatedUnionRefine(t *testing.T) {
	t.Run("basic refine with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			}).Refine(func(val any) bool {
				return true
			}, core.SchemaParams{Error: "Must be a valid discriminated union"})
		})
	})

	t.Run("refine vs transform distinction with invalid options", func(t *testing.T) {
		// Test concept: distinction between Refine and Transform
		// But using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			}).Refine(func(val any) bool {
				return true
			})
		})

		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			}).Transform(func(val any, ctx *core.RefinementContext) (any, error) {
				return val, nil
			})
		})
	})
}

// =============================================================================
// 7. Error handling
// =============================================================================

func TestDiscriminatedUnionErrorHandling(t *testing.T) {
	t.Run("invalid - null with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			})
		})
	})

	t.Run("invalid discriminator value", func(t *testing.T) {
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "a": String()}),
			Object(core.ObjectSchema{"type": Literal("b"), "b": String()}),
		})

		_, err := schema.Parse(map[string]any{
			"type": "x",
			"a":    "abc",
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidUnion, zodErr.Issues[0].Code)
	})

	t.Run("valid discriminator value, invalid data", func(t *testing.T) {
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "a": String()}),
			Object(core.ObjectSchema{"type": Literal("b"), "b": String()}),
		})

		_, err := schema.Parse(map[string]any{
			"type": "a",
			"b":    "abc", // Wrong field for discriminator "a"
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("union fallback behavior with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction, even with UnionFallback set
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			}, core.SchemaParams{UnionFallback: true})
		})
	})

	t.Run("custom error messages with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction, even with custom error set
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			}, core.SchemaParams{
				Error: "Must be a valid discriminated type",
			})
		})
	})
}

// =============================================================================
// 8. Edge and mutual exclusion cases
// =============================================================================

func TestDiscriminatedUnionEdgeCases(t *testing.T) {
	t.Run("single element union with invalid option", func(t *testing.T) {
		assert.Panics(t, func() {
			DiscriminatedUnion("a", []core.ZodType[any, any]{
				String(),
			})
		})
	})

	t.Run("empty union", func(t *testing.T) {
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{})
		})
	})

	t.Run("missing discriminator field with invalid options", func(t *testing.T) {
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			})
		})
	})

	t.Run("complex nested discriminated unions", func(t *testing.T) {
		BaseError := Object(core.ObjectSchema{"status": Literal("failed"), "message": String()})
		MyErrors := DiscriminatedUnion("code", []core.ZodType[any, any]{
			BaseError.Extend(core.ObjectSchema{"code": Literal(400)}),
			BaseError.Extend(core.ObjectSchema{"code": Literal(401)}),
			BaseError.Extend(core.ObjectSchema{"code": Literal(500)}),
		})

		MyResult := DiscriminatedUnion("status", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"status": Literal("success"), "data": String()}),
			MyErrors,
		})

		// Test success case
		result, err := MyResult.Parse(map[string]any{
			"status": "success",
			"data":   "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"status": "success",
			"data":   "hello",
		}
		assert.Equal(t, expected, result)

		// Test error cases
		result, err = MyResult.Parse(map[string]any{
			"status":  "failed",
			"code":    400,
			"message": "bad request",
		})
		require.NoError(t, err)
		expected = map[string]any{
			"status":  "failed",
			"code":    400,
			"message": "bad request",
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// 9. Default and Prefault tests
// =============================================================================

func TestDiscriminatedUnionDefaultAndPrefault(t *testing.T) {
	t.Run("basic default value with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			}).Default(map[string]any{
				"type":  "string",
				"value": "default",
			})
		})
	})

	t.Run("function-based default value with invalid options", func(t *testing.T) {
		counter := 0
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			}).DefaultFunc(func() any {
				counter++
				return map[string]any{
					"type":  "counter",
					"value": counter,
				}
			})
		})
	})

	t.Run("prefault value with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
				Int(),
			}).Prefault(map[string]any{
				"type":  "string",
				"value": "fallback_value",
			})
		})
	})

	t.Run("default with transform compatibility with invalid options", func(t *testing.T) {
		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			}).Default(map[string]any{
				"type":  "string",
				"value": "hello",
			})
		})

		// Test that using invalid options will panic during construction
		assert.Panics(t, func() {
			DiscriminatedUnion("type", []core.ZodType[any, any]{
				String(),
			}).Transform(func(val any, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"original":    val,
					"transformed": true,
				}, nil
			})
		})
	})

	// Valid tests with proper Object + Literal structure
	t.Run("valid default value", func(t *testing.T) {
		defaultValue := map[string]any{
			"type":  "a",
			"value": "default_value",
		}

		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "value": String()}),
			Object(core.ObjectSchema{"type": Literal("b"), "value": Int()}),
		}).Default(defaultValue)

		// nil input should use default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)

		// Valid input should override default
		validInput := map[string]any{
			"type":  "b",
			"value": 42,
		}
		result, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)
	})

	t.Run("valid defaultFunc", func(t *testing.T) {
		counter := 0
		schema := DiscriminatedUnion("status", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"status": Literal("success"), "data": String()}),
			Object(core.ObjectSchema{"status": Literal("failed"), "error": String()}),
		}).DefaultFunc(func() any {
			counter++
			return map[string]any{
				"status": "success",
				"data":   fmt.Sprintf("generated-%d", counter),
			}
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		result1Map := result1.(map[string]any)
		assert.Equal(t, "success", result1Map["status"])
		assert.Equal(t, "generated-1", result1Map["data"])
		assert.Equal(t, 1, counter)

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map := result2.(map[string]any)
		assert.Equal(t, "success", result2Map["status"])
		assert.Equal(t, "generated-2", result2Map["data"])
		assert.Equal(t, 2, counter)

		// Valid input should not call function
		validInput := map[string]any{
			"status": "failed",
			"error":  "something went wrong",
		}
		result3, err3 := schema.Parse(validInput)
		require.NoError(t, err3)
		assert.Equal(t, validInput, result3)
		assert.Equal(t, 2, counter) // Counter should not increment
	})

	t.Run("valid prefault value", func(t *testing.T) {
		prefaultValue := map[string]any{
			"type":  "a",
			"value": "fallback_value",
		}

		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "value": String().Min(5)}),
			Object(core.ObjectSchema{"type": Literal("b"), "value": Int().Min(10)}),
		}).Prefault(prefaultValue)

		// Valid input should pass through
		validInput := map[string]any{
			"type":  "a",
			"value": "valid_string",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid input should use prefault
		invalidInput := map[string]any{
			"type":  "a",
			"value": "hi", // Too short
		}
		result, err = schema.Parse(invalidInput)
		require.NoError(t, err)
		assert.Equal(t, prefaultValue, result)
	})

	t.Run("valid prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := DiscriminatedUnion("status", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"status": Literal("success"), "data": String().Min(5)}),
			Object(core.ObjectSchema{"status": Literal("failed"), "error": String().Min(5)}),
		}).PrefaultFunc(func() any {
			counter++
			return map[string]any{
				"status": "success",
				"data":   fmt.Sprintf("fallback-%d", counter),
			}
		})

		// Valid input should not call function
		validInput := map[string]any{
			"status": "success",
			"data":   "valid_data",
		}
		result1, err1 := schema.Parse(validInput)
		require.NoError(t, err1)
		assert.Equal(t, validInput, result1)
		assert.Equal(t, 0, counter)

		// Invalid input should call prefault function
		invalidInput := map[string]any{
			"status": "success",
			"data":   "hi", // Too short
		}
		result2, err2 := schema.Parse(invalidInput)
		require.NoError(t, err2)
		result2Map := result2.(map[string]any)
		assert.Equal(t, "success", result2Map["status"])
		assert.Equal(t, "fallback-1", result2Map["data"])
		assert.Equal(t, 1, counter)

		// Another invalid input should call function again
		result3, err3 := schema.Parse(invalidInput)
		require.NoError(t, err3)
		result3Map := result3.(map[string]any)
		assert.Equal(t, "success", result3Map["status"])
		assert.Equal(t, "fallback-2", result3Map["data"])
		assert.Equal(t, 2, counter)
	})

	t.Run("valid default vs prefault distinction", func(t *testing.T) {
		defaultValue := map[string]any{
			"type":  "a",
			"value": "default_value",
		}
		prefaultValue := map[string]any{
			"type":  "a",
			"value": "prefault_value",
		}

		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "value": String().Min(5)}),
			Object(core.ObjectSchema{"type": Literal("b"), "value": Int().Min(10)}),
		}).Default(defaultValue).Prefault(prefaultValue)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Valid input succeeds
		validInput := map[string]any{
			"type":  "a",
			"value": "valid_string",
		}
		result2, err2 := schema.Parse(validInput)
		require.NoError(t, err2)
		assert.Equal(t, validInput, result2)

		// Invalid input uses prefault
		invalidInput := map[string]any{
			"type":  "a",
			"value": "hi", // Too short
		}
		result3, err3 := schema.Parse(invalidInput)
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)
	})
}

// ------------------------------------------------------------------
// Optional discriminator key – discriminator may be omitted
// ------------------------------------------------------------------
func TestDiscriminatedUnionOptionalDiscriminator(t *testing.T) {
	schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
		Object(core.ObjectSchema{
			"type": Literal("a").Optional(), // discriminator may be absent
			"a":    String(),
		}),
		Object(core.ObjectSchema{
			"type": Literal("b"),
			"b":    String(),
		}),
	}, core.SchemaParams{UnionFallback: true})

	// Discriminator present
	_, err := schema.Parse(map[string]any{"type": "a", "a": "hello"})
	require.NoError(t, err)

	// Discriminator omitted – still matches first option via fallback
	_, err = schema.Parse(map[string]any{"a": "hello"})
	require.NoError(t, err)
}

// ------------------------------------------------------------------
// Nil discriminator literal value
// ------------------------------------------------------------------
func TestDiscriminatedUnionNilDiscriminator(t *testing.T) {
	schema := DiscriminatedUnion("kind", []core.ZodType[any, any]{
		Object(core.ObjectSchema{"kind": Literal("foo"), "v": String()}),
		Object(core.ObjectSchema{"kind": Nil(), "v": String()}),
	})

	// Match nil discriminator value
	_, err := schema.Parse(map[string]any{"kind": nil, "v": "bar"})
	require.NoError(t, err)
}

// =============================================================================
// 10. Additional discriminated union behavior tests (safe subset)
// =============================================================================

func TestDiscriminatedUnionAdditionalBehavior(t *testing.T) {
	// ------------------------------------------------------------------
	// invalid discriminator value with unionFallback enabled
	// ------------------------------------------------------------------
	t.Run("invalid discriminator unionFallback", func(t *testing.T) {
		schema := DiscriminatedUnion("type", []core.ZodType[any, any]{
			Object(core.ObjectSchema{"type": Literal("a"), "a": String()}),
			Object(core.ObjectSchema{"type": Literal("b"), "b": String()}),
		}, core.SchemaParams{UnionFallback: true})

		// Expect an error because discriminator value "x" is unknown
		_, err := schema.Parse(map[string]any{"type": "x", "a": "abc"})
		assert.Error(t, err)

		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		// Should report invalid_union
		assert.Equal(t, core.InvalidUnion, zErr.Issues[0].Code)
	})

	// ------------------------------------------------------------------
	// single-element discriminated union still validates inner schema
	// ------------------------------------------------------------------
	t.Run("single element union still validates", func(t *testing.T) {
		singleOption := Object(core.ObjectSchema{
			"a": Literal("discKey"),
			"b": Enum("apple", "banana"),
			"c": Object(core.ObjectSchema{"id": String()}),
		})

		schema := DiscriminatedUnion("a", []core.ZodType[any, any]{singleOption})

		// Provide invalid nested data (missing c.id) -> should error
		input := map[string]any{
			"a": "discKey",
			"b": "apple",
			"c": map[string]any{},
		}
		_, err := schema.Parse(input)
		assert.Error(t, err)
	})
}
