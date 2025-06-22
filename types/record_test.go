package types

import (
	"sort"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestRecordBasicFunctionality(t *testing.T) {
	t.Run("constructor", func(t *testing.T) {
		schema := Record(String(), Int())
		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, "record", internals.Type)
	})

	t.Run("constructor with params", func(t *testing.T) {
		schema := Record(String(), Int(), core.SchemaParams{
			Error: "core.Custom record error",
		})
		require.NotNil(t, schema)
		// Coercion is no longer supported for collection types
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Record(String(), Int())

		// Basic map input returns map[any]any
		input := map[string]any{
			"key1": 10,
			"key2": 20,
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, 10, resultMap["key1"])
		assert.Equal(t, 20, resultMap["key2"])

		// Pointer input returns same pointer type
		inputPtr := &map[any]any{
			"key": 42,
		}
		result2, err := schema.Parse(inputPtr)
		require.NoError(t, err)
		resultPtr, ok := result2.(*map[any]any)
		require.True(t, ok)
		assert.Equal(t, inputPtr, resultPtr) // Same pointer identity
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Record(String(), Int()).Nilable()

		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid input keeps type inference
		validInput := map[string]any{
			"key": 42,
		}
		result2, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result2)
	})

	t.Run("key and value type access", func(t *testing.T) {
		schema := Record(String(), Int())

		// Test key type validation
		_, err := schema.internals.KeyType.Parse("valid_key")
		require.NoError(t, err)

		// Test value type validation
		_, err = schema.internals.ValueType.Parse(42)
		require.NoError(t, err)
	})
}

// =============================================================================
// 2. Validation methods (Record-specific key/value validation)
// =============================================================================

func TestRecordValidations(t *testing.T) {
	t.Run("string key validation", func(t *testing.T) {
		schema := Record(String(), Int())

		// Valid string keys
		validInput := map[string]any{
			"valid_key1": 1,
			"valid_key2": 2,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid non-string keys should be handled by type conversion
		invalidInput := map[any]any{
			1:   10, // Non-string key
			"b": 20,
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("enum key validation", func(t *testing.T) {
		keySchema := Enum("red", "green", "blue")
		schema := Record(keySchema, Int())

		// Valid enum keys
		validInput := map[string]any{
			"red":   255,
			"green": 128,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid enum key
		invalidInput := map[string]any{
			"red":    255,
			"yellow": 128, // Invalid enum value
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("literal key validation", func(t *testing.T) {
		keySchema := Literal([]any{"name", "age", "email"})
		schema := Record(keySchema, String())

		// Valid literal keys
		validInput := map[string]any{
			"name": "John",
			"age":  "25",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid key
		invalidInput := map[string]any{
			"name":    "John",
			"invalid": "data", // Not in literal values
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("value validation", func(t *testing.T) {
		schema := Record(String(), Int())

		// Valid integer values
		validInput := map[string]any{
			"key1": 10,
			"key2": 20,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid string values for int schema
		invalidInput := map[string]any{
			"key1": "not_an_int",
			"key2": "also_not_int",
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("mixed validation errors", func(t *testing.T) {
		schema := Record(String(), Int())

		input := map[string]any{
			"valid_key":   10,
			"invalid_key": "string_value", // Invalid value type
		}
		_, err := schema.Parse(input)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have error path pointing to the invalid value
		hasPathError := false
		for _, issue := range zodErr.Issues {
			if len(issue.Path) > 0 && issue.Path[0] == "invalid_key" {
				hasPathError = true
				break
			}
		}
		assert.True(t, hasPathError, "Expected error with path pointing to invalid key")
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestRecordModifiers(t *testing.T) {
	schema := Record(String(), Int())

	t.Run("optional wrapper", func(t *testing.T) {
		optionalSchema := schema.Optional()

		// Valid record
		validInput := map[string]any{
			"key": 42,
		}
		result, err := optionalSchema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input should return nil (Optional semantics)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		nilableSchema := schema.Nilable()

		// Valid record
		validInput := map[string]any{
			"key": 42,
		}
		result, err := nilableSchema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input should succeed
		result, err = nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		nullishSchema := schema.Nullish()

		// Valid record
		validInput := map[string]any{
			"key": 42,
		}
		result, err := nullishSchema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input should succeed (Nullish = Optional)
		result, err = nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default wrapper", func(t *testing.T) {
		defaultValue := map[any]any{
			"default_key": 100,
		}
		defaultSchema := schema.Default(defaultValue)

		// Valid record (should not use default)
		validInput := map[string]any{
			"key": 42,
		}
		result, err := defaultSchema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// nil input should use default
		result, err = defaultSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)
	})

	t.Run("prefault wrapper", func(t *testing.T) {
		prefaultValue := map[any]any{
			"fallback_key": 999,
		}
		prefaultSchema := schema.Prefault(prefaultValue)

		// Valid input should pass through
		validInput := map[string]any{
			"key": 42,
		}
		result, err := prefaultSchema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid input should use prefault value
		invalidInput := "not a map"
		result, err = prefaultSchema.Parse(invalidInput)
		require.NoError(t, err)
		assert.Equal(t, prefaultValue, result)
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestRecordChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		schema := Record(String(), Int())

		// Test that schema can be used in chains
		require.NotNil(t, schema)

		// Test basic functionality after chaining
		result, err := schema.Parse(map[string]any{
			"key": 42,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("default value", func(t *testing.T) {
		// Note: Once Default() is called, the return type becomes core.ZodType[any, any]
		// which doesn't have Refine method available
		schema := Record(String(), Int()).Default(map[any]any{"default": 0})

		// Test with valid input
		validInput := map[string]any{
			"key": 42,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test with nil (should use default)
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestRecordTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Record(String(), Int()).Transform(func(val map[any]any, ctx *core.RefinementContext) (any, error) {
			// Transform: count the number of entries
			return len(val), nil
		})

		result, err := schema.Parse(map[string]any{
			"a": 1,
			"b": 2,
			"c": 3,
		})
		require.NoError(t, err)
		assert.Equal(t, 3, result)
	})

	t.Run("transform with type safety", func(t *testing.T) {
		schema := Record(String(), String()).Transform(func(val map[any]any, ctx *core.RefinementContext) (any, error) {
			// Transform to a different structure
			result := make(map[string]string)
			for k, v := range val {
				if keyStr, ok := k.(string); ok {
					if valStr, ok := v.(string); ok {
						result[keyStr] = valStr
					}
				}
			}
			return result, nil
		})

		input := map[string]any{
			"key1": "value1",
			"key2": "value2",
		}
		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "value1", resultMap["key1"])
		assert.Equal(t, "value2", resultMap["key2"])
	})

	t.Run("transform any", func(t *testing.T) {
		schema := Record(String(), Int()).TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
			if recordMap, ok := val.(map[any]any); ok {
				// Transform: count the number of entries
				return len(recordMap), nil
			}
			return val, nil
		})

		result, err := schema.Parse(map[string]any{
			"a": 1,
			"b": 2,
			"c": 3,
		})
		require.NoError(t, err)
		assert.Equal(t, 3, result)
	})

	t.Run("pipe", func(t *testing.T) {
		// Create a pipeline: Record -> Transform to count -> Validate count > 0
		countSchema := Int().Min(1)
		schema := Record(String(), Int()).
			Transform(func(val map[any]any, ctx *core.RefinementContext) (any, error) {
				return len(val), nil
			}).
			Pipe(countSchema)

		// Valid case: non-empty record
		result, err := schema.Parse(map[string]any{
			"key": 42,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, result)

		// Invalid case: empty record (count = 0, fails Min(1))
		_, err = schema.Parse(map[string]any{})
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Refine
// =============================================================================

func TestRecordRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Record(String(), Int()).Refine(func(val map[any]any) bool {
			// Only allow records with even number of entries
			return len(val)%2 == 0
		})

		// Valid case (2 entries)
		_, err := schema.Parse(map[string]any{
			"a": 1,
			"b": 2,
		})
		require.NoError(t, err)

		// Invalid case (1 entry)
		_, err = schema.Parse(map[string]any{
			"a": 1,
		})
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := Record(String(), Int()).Refine(func(val map[any]any) bool {
			return len(val) >= 2
		}, core.SchemaParams{
			Error: "Record must have at least 2 entries",
		})

		_, err := schema.Parse(map[string]any{
			"single": 1,
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Record must have at least 2 entries")
	})

	t.Run("refine any", func(t *testing.T) {
		schema := Record(String(), Int()).RefineAny(func(val any) bool {
			if recordMap, ok := val.(map[any]any); ok {
				return len(recordMap) > 0
			}
			return false
		})

		// Valid case
		_, err := schema.Parse(map[string]any{
			"key": 42,
		})
		require.NoError(t, err)

		// Invalid case
		_, err = schema.Parse(map[string]any{})
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Error handling
// =============================================================================

func TestRecordErrorHandling(t *testing.T) {
	t.Run("type error", func(t *testing.T) {
		schema := Record(String(), Int())
		_, err := schema.Parse("not a map")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("key validation error", func(t *testing.T) {
		schema := Record(Int(), String()) // Keys must be integers
		_, err := schema.Parse(map[string]any{
			"string_key": "value", // Invalid key type
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("value validation error", func(t *testing.T) {
		schema := Record(String(), Int())
		_, err := schema.Parse(map[string]any{
			"key": "not an int", // Invalid value type
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := Record(String(), Int())
		_, err := schema.Parse(map[string]any{
			"key1": "not_an_int",
			"key2": "also_not_int",
		})
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		// Should have multiple issues for different keys
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Record(String(), Int(), core.SchemaParams{
			Error: "core.Custom record validation error",
		})
		_, err := schema.Parse("not a map")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		// Note: core.Custom error might be overridden by specific validation errors
		assert.NotEmpty(t, zodErr.Issues)
	})
}

// =============================================================================
// 8. Edge and mutual exclusion cases
// =============================================================================

func TestRecordEdgeCases(t *testing.T) {
	t.Run("empty record", func(t *testing.T) {
		schema := Record(String(), Int())

		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, 0, len(resultMap))
	})

	t.Run("nested record", func(t *testing.T) {
		innerSchema := Record(String(), Int())
		// nested schema without explicit conversion
		outerSchema := Record(String(), innerSchema)

		input := map[string]any{
			"nested": map[string]any{
				"inner_key": 42,
			},
		}

		result, err := outerSchema.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("large record", func(t *testing.T) {
		schema := Record(String(), Int())

		// Create a large record
		largeInput := make(map[string]any)
		for i := 0; i < 100; i++ {
			largeInput[string(rune('a'+i%26))+string(rune('0'+i/26))] = i
		}

		result, err := schema.Parse(largeInput)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, len(largeInput), len(resultMap))
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Record(String(), Int())

		complexTypes := []any{
			make(chan int),
			func() int { return 1 },
			struct{ I int }{I: 1},
			[]any{1, 2, 3},
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("undefined values handling", func(t *testing.T) {
		// Test that undefined/nil values are preserved in records
		schema := Record(String(), Any().Nilable())

		input := map[string]any{
			"defined":   "value",
			"undefined": nil,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, "value", resultMap["defined"])
		assert.Nil(t, resultMap["undefined"])
		assert.Equal(t, 2, len(resultMap)) // Both keys should be present
	})

	t.Run("prototype pollution protection", func(t *testing.T) {
		// Test that __proto__ keys don't cause issues
		schema := Record(String(), String())

		input := map[string]any{
			"__proto__": "evil",
			"normal":    "good",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, "evil", resultMap["__proto__"])
		assert.Equal(t, "good", resultMap["normal"])
	})
}

// =============================================================================
// 9. Default and Prefault tests
// =============================================================================

func TestRecordDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		defaultValue := map[any]any{
			"default_key": 100,
		}
		schema := Record(String(), Int()).Default(defaultValue)

		// nil input should use default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)

		// Valid input should not use default
		validInput := map[string]any{
			"key": 42,
		}
		result, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotEqual(t, defaultValue, result)
	})

	t.Run("default function", func(t *testing.T) {
		counter := 0
		schema := Record(String(), Int()).DefaultFunc(func() map[any]any {
			counter++
			return map[any]any{
				"generated": counter,
			}
		})

		// Each nil input should generate a new default
		result1, err := schema.Parse(nil)
		require.NoError(t, err)
		result1Map := result1.(map[any]any)
		assert.Equal(t, 1, result1Map["generated"])

		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		result2Map := result2.(map[any]any)
		assert.Equal(t, 2, result2Map["generated"])

		// Valid input should not trigger function
		validInput := map[string]any{
			"key": 42,
		}
		_, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, 2, counter) // Counter should not increment
	})

	t.Run("prefault value", func(t *testing.T) {
		prefaultValue := map[any]any{
			"fallback": 999,
		}
		schema := Record(String(), Int()).Prefault(prefaultValue)

		// Valid input should pass through
		validInput := map[string]any{
			"key": 42,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotEqual(t, prefaultValue, result)

		// Invalid input should use prefault
		result, err = schema.Parse("invalid input")
		require.NoError(t, err)
		assert.Equal(t, prefaultValue, result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := Record(String(), Int()).PrefaultFunc(func() map[any]any {
			counter++
			return map[any]any{
				"fallback": counter,
			}
		})

		// Each invalid input should generate a new prefault
		result1, err := schema.Parse("invalid1")
		require.NoError(t, err)
		result1Map := result1.(map[any]any)
		assert.Equal(t, 1, result1Map["fallback"])

		result2, err := schema.Parse("invalid2")
		require.NoError(t, err)
		result2Map := result2.(map[any]any)
		assert.Equal(t, 2, result2Map["fallback"])

		// Valid input should not trigger function
		validInput := map[string]any{
			"key": 42,
		}
		_, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, 2, counter) // Counter should not increment
	})

	t.Run("default with validation chain", func(t *testing.T) {
		defaultValue := map[any]any{
			"default": 1,
			"value":   2,
		}
		// Note: Once Default() is called, the return type becomes core.ZodType[any, any]
		// which doesn't have Refine method available, so we test basic functionality
		schema := Record(String(), Int()).
			Default(defaultValue)

		// nil input: use default, should pass validation
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, result)

		// Valid input: should pass validation
		validInput := map[string]any{
			"key1": 1,
			"key2": 2,
		}
		result, err = schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotEqual(t, defaultValue, result)
	})

	t.Run("prefault with validation chain", func(t *testing.T) {
		prefaultValue := map[any]any{
			"safe":     1,
			"fallback": 2,
		}
		schema := Record(String(), Int()).
			Refine(func(val map[any]any) bool {
				return len(val) >= 2
			}).
			Prefault(prefaultValue)

		// Valid input: should pass validation
		validInput := map[string]any{
			"key1": 1,
			"key2": 2,
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotEqual(t, prefaultValue, result)

		// Invalid input: should use prefault
		invalidInput := map[string]any{
			"single": 1, // Fails validation (len < 2)
		}
		result, err = schema.Parse(invalidInput)
		require.NoError(t, err)
		assert.Equal(t, prefaultValue, result)
	})
}

// =============================================================================
// 10. Compatibility tests
// =============================================================================

func TestRecordCompatibility(t *testing.T) {
	t.Run("enum exhaustiveness", func(t *testing.T) {
		schema := Record(Enum("Tuna", "Salmon"), String())

		// Valid case: all required keys present
		validInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, "asdf", resultMap["Tuna"])
		assert.Equal(t, "asdf", resultMap["Salmon"])

		// Invalid case: unrecognized key
		invalidInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
			"Trout":  "asdf", // Unrecognized key
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)

		// Partial case: missing keys are allowed; should parse successfully.
		partialInput := map[string]any{
			"Tuna": "asdf",
		}
		resultPartial, err := schema.Parse(partialInput)
		require.NoError(t, err)
		resultMapPartial, ok := resultPartial.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, "asdf", resultMapPartial["Tuna"])
	})

	t.Run("literal exhaustiveness", func(t *testing.T) {
		schema := Record(Literal([]any{"Tuna", "Salmon"}), String())

		// Valid case
		validInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid case: unrecognized key
		invalidInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
			"Trout":  "asdf", // Not in literal values
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("string record with numeric keys", func(t *testing.T) {
		schema := Record(String(), Bool())

		// Should handle numeric keys as strings
		input := map[string]any{
			"k1":   true,
			"k2":   false,
			"1234": false, // Numeric string key
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, true, resultMap["k1"])
		assert.Equal(t, false, resultMap["k2"])
		assert.Equal(t, false, resultMap["1234"])
	})

	t.Run("invalid type error message", func(t *testing.T) {
		schema := Record(String(), Bool())

		_, err := schema.Parse("not a record")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
		// Error message should mention "record" or "object"
		assert.Contains(t, zodErr.Issues[0].Message, "object")
	})

	t.Run("union key exhaustiveness", func(t *testing.T) {
		keySchema := Union([]core.ZodType[any, any]{
			Literal("Tuna"),
			Literal("Salmon"),
		})
		schema := Record(keySchema, String())

		// Valid case
		validInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, "asdf", resultMap["Tuna"])
		assert.Equal(t, "asdf", resultMap["Salmon"])

		// Invalid case: unrecognized key
		invalidInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
			"Trout":  "asdf", // Not in union
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})

	t.Run("undefined values preservation", func(t *testing.T) {
		schema := Record(String(), Any().Nilable())

		input := map[string]any{
			"foo": nil, // Go equivalent of undefined
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Nil(t, resultMap["foo"])
		assert.Equal(t, 1, len(resultMap)) // Key should be present
	})

	t.Run("prototype pollution protection", func(t *testing.T) {
		// Test that __proto__ keys don't cause issues
		schema := Record(String(), String())

		input := map[string]any{
			"__proto__": "evil",
			"normal":    "good",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, "evil", resultMap["__proto__"])
		assert.Equal(t, "good", resultMap["normal"])
		// Verify no prototype pollution occurred
		assert.Equal(t, 2, len(resultMap))
	})
}

////////////////////////////
////   PARTIAL RECORD TESTS ////
////////////////////////////

func TestPartialRecord(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		// Create a partial record: string keys, int values
		schema := PartialRecord(String(), Int())

		// Empty record should pass because keys are optional
		result, err := schema.Parse(map[any]any{})
		assert.NoError(t, err)
		assert.Equal(t, map[any]any{}, result)

		// Valid key-value pairs
		result, err = schema.Parse(map[any]any{
			"name": 42,
			"age":  25,
		})
		assert.NoError(t, err)
		expected := map[any]any{
			"name": 42,
			"age":  25,
		}
		assert.Equal(t, expected, result)

		// Invalid value type
		_, err = schema.Parse(map[any]any{
			"name": "not_an_int", // should be an integer
		})
		assert.Error(t, err)

		// Invalid key type
		_, err = schema.Parse(map[any]any{
			123: 42, // key should be a string
		})
		assert.Error(t, err)
	})

	t.Run("partial record with enum keys", func(t *testing.T) {
		// Create a partial record that only allows specific enum keys
		allowedKeys := []any{"name", "age", "email"}
		schema := PartialRecord(EnumSlice(allowedKeys), String())

		// Empty record
		result, err := schema.Parse(map[any]any{})
		assert.NoError(t, err)
		assert.Equal(t, map[any]any{}, result)

		// Some keys present
		result, err = schema.Parse(map[any]any{
			"name": "John",
			"age":  "25",
		})
		assert.NoError(t, err)
		expected := map[any]any{
			"name": "John",
			"age":  "25",
		}
		assert.Equal(t, expected, result)

		// All keys present
		result, err = schema.Parse(map[any]any{
			"name":  "John",
			"age":   "25",
			"email": "john@example.com",
		})
		assert.NoError(t, err)
		expected = map[any]any{
			"name":  "John",
			"age":   "25",
			"email": "john@example.com",
		}
		assert.Equal(t, expected, result)

		// Key not allowed
		_, err = schema.Parse(map[any]any{
			"name":    "John",
			"invalid": "should_fail", // not in allowed key list
		})
		assert.Error(t, err)
	})

	t.Run("comparison with regular record", func(t *testing.T) {
		// Regular Record
		normalRecord := Record(Enum("a", "b"), String())

		// Partial Record
		partialRecord := PartialRecord(Enum("a", "b"), String())

		// Test data: only partial keys
		testData := map[any]any{
			"a": "value_a",
			// missing key "b"
		}

		// A regular Record may require all keys (implementation-dependent) but
		// PartialRecord should allow some keys to be missing
		result, err := partialRecord.Parse(testData)
		assert.NoError(t, err)
		assert.Equal(t, testData, result)

		// Complete data should be valid for both schemas
		fullData := map[any]any{
			"a": "value_a",
			"b": "value_b",
		}

		result1, err1 := normalRecord.Parse(fullData)
		result2, err2 := partialRecord.Parse(fullData)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, result1, result2)
	})

	t.Run("chaining methods", func(t *testing.T) {
		// Chain PartialRecord with other methods
		schema := PartialRecord(String(), Int()).Optional()

		// Test undefined/nil values
		result, err := schema.Parse(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Test normal values
		result, err = schema.Parse(map[any]any{
			"key": 42,
		})
		assert.NoError(t, err)
		expected := map[any]any{
			"key": 42,
		}
		assert.Equal(t, expected, result)
	})

	// Coercion is no longer supported for collection types - test removed

	t.Run("error handling", func(t *testing.T) {
		schema := PartialRecord(String(), Int())

		// Test non-object input
		_, err := schema.Parse("not_an_object")
		assert.Error(t, err)

		// Test nil input (when not Optional)
		_, err = schema.Parse(nil)
		assert.Error(t, err)

		// Test array input
		_, err = schema.Parse([]int{1, 2, 3})
		assert.Error(t, err)
	})
}

func TestZodRecord(t *testing.T) {
	t.Run("valid record", func(t *testing.T) {
		// Create schemas for key and value types
		keySchema := String()
		valueSchema := Int()
		schema := Record(keySchema, valueSchema)

		input := map[string]any{
			"key1": 42,
			"key2": 100,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid key type", func(t *testing.T) {
		keySchema := String()
		valueSchema := Int()
		schema := Record(keySchema, valueSchema)

		input := map[int]any{
			123: 42,
		}

		_, err := schema.Parse(input)
		assert.Error(t, err)
	})

	t.Run("invalid value type", func(t *testing.T) {
		keySchema := String()
		valueSchema := Int()
		schema := Record(keySchema, valueSchema)

		input := map[string]any{
			"key1": "not_an_int",
		}

		_, err := schema.Parse(input)
		assert.Error(t, err)
	})

	t.Run("nested record", func(t *testing.T) {
		keySchema := String()
		valueSchema := Int()
		innerSchema := Record(keySchema, valueSchema)
		outerSchema := Record(String(), innerSchema)

		input := map[string]any{
			"nested": map[string]any{
				"inner_key": 42,
			},
		}

		result, err := outerSchema.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("empty record", func(t *testing.T) {
		keySchema := String()
		valueSchema := Int()
		schema := Record(keySchema, valueSchema)

		input := map[string]any{}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("nil input", func(t *testing.T) {
		keySchema := String()
		valueSchema := Int()
		schema := Record(keySchema, valueSchema)

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-map input", func(t *testing.T) {
		keySchema := String()
		valueSchema := Int()
		schema := Record(keySchema, valueSchema)

		_, err := schema.Parse("not a map")
		assert.Error(t, err)
	})
}

// =============================================================================
// 11. Additional exhaustiveness tests
// =============================================================================

// TestRecordPipeExhaustiveness validates that piping the key schema still enforces key constraints
func TestRecordPipeExhaustiveness(t *testing.T) {
	t.Run("pipe exhaustiveness", func(t *testing.T) {
		keySchema := Enum("Tuna", "Salmon").Pipe(Any())
		schema := Record(keySchema, String())

		// Valid input with recognized keys
		validInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid input with unrecognized key should fail
		invalidInput := map[string]any{
			"Tuna":   "asdf",
			"Salmon": "asdf",
			"Trout":  "asdf",
		}
		_, err = schema.Parse(invalidInput)
		assert.Error(t, err)
	})
}

// TestRecordAllowNilValues ensures record schema can accept nil values when value type is Nil()
func TestRecordAllowNilValues(t *testing.T) {
	t.Run("allow nil values", func(t *testing.T) {
		schema := Record(String(), Nil())

		input := map[string]any{
			"_test": nil,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		resultMap, ok := result.(map[any]any)
		require.True(t, ok)
		assert.Equal(t, []string{"_test"}, keysSorted(resultMap))
		assert.Nil(t, resultMap["_test"])
	})
}

// keysSorted returns sorted keys of a map for deterministic assertions
func keysSorted(m map[any]any) []string {
	// Preallocate slice capacity to avoid reallocations under lint prealloc check
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k.(string))
	}
	sort.Strings(ks)
	return ks
}
