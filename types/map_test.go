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

func TestMap_BasicFunctionality(t *testing.T) {
	t.Run("valid map inputs", func(t *testing.T) {
		// String key to int value map
		keySchema := String()
		valueSchema := Int()
		mapSchema := Map(keySchema, valueSchema)

		testMap := map[any]any{
			"key1": 1,
			"key2": 2,
			"key3": 3,
		}

		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})

	t.Run("empty map", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		result, err := mapSchema.Parse(map[any]any{})
		require.NoError(t, err)
		assert.Equal(t, map[any]any{}, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		invalidInputs := []any{
			"not a map", 123, []int{1, 2, 3}, true, nil,
		}

		for _, input := range invalidInputs {
			_, err := mapSchema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		mapSchema := Map(String(), Bool())
		testMap := map[any]any{"test": true}

		// Test Parse method
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)

		// Test MustParse method
		mustResult := mapSchema.MustParse(testMap)
		assert.Equal(t, testMap, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			mapSchema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid map"
		mapSchema := Map(String(), Int(), core.SchemaParams{Error: customError})

		require.NotNil(t, mapSchema)

		_, err := mapSchema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestMap_TypeSafety(t *testing.T) {
	t.Run("map returns map[any]any type", func(t *testing.T) {
		mapSchema := Map(String(), Int())
		require.NotNil(t, mapSchema)

		testMap := map[any]any{"test": 42}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
		assert.IsType(t, map[any]any{}, result)
	})

	t.Run("key validation", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		// Valid keys should pass
		validMap := map[any]any{"valid_key": 42}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// Invalid key type should fail
		invalidMap := map[any]any{123: 42} // int key instead of string
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("value validation", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		// Valid values should pass
		validMap := map[any]any{"key": 42}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// Invalid value type should fail
		invalidMap := map[any]any{"key": "not_an_int"}
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		mapSchema := Map(String(), Bool())
		testMap := map[any]any{"test": true}

		result := mapSchema.MustParse(testMap)
		assert.IsType(t, map[any]any{}, result)
		assert.Equal(t, testMap, result)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestMap_Modifiers(t *testing.T) {
	t.Run("Optional allows nil values", func(t *testing.T) {
		mapSchema := Map(String(), Int())
		optionalSchema := mapSchema.Optional()

		// Test non-nil value - should return pointer
		testMap := map[any]any{"key": 42}
		result, err := optionalSchema.Parse(testMap)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testMap, *result)

		// Test nil value (should be allowed for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		mapSchema := Map(String(), Int())
		nilableSchema := mapSchema.Nilable()

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - should return pointer
		testMap := map[any]any{"key": 42}
		result, err = nilableSchema.Parse(testMap)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testMap, *result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultMap := map[any]any{"default": 1}
		mapSchema := Map(String(), Int())
		defaultSchema := mapSchema.Default(defaultMap)

		// Valid input should override default
		testMap := map[any]any{"test": 2}
		result, err := defaultSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
		assert.IsType(t, map[any]any{}, result)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultMap := map[any]any{"prefault": 1}
		mapSchema := Map(String(), Int())
		prefaultSchema := mapSchema.Prefault(prefaultMap)

		// Valid input should override prefault
		testMap := map[any]any{"test": 2}
		result, err := prefaultSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
		assert.IsType(t, map[any]any{}, result)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestMap_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		defaultMap := map[any]any{"default": 1}
		mapSchema := Map(String(), Int()).
			Default(defaultMap). // Preserves map type
			Optional()           // Now returns pointer type

		// Test final behavior - should return pointer
		testMap := map[any]any{"test": 2}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testMap, *result)

		// Test nil handling - Default should short-circuit and return default value
		result, err = mapSchema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, &defaultMap, result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		mapSchema := Map(String(), Int()).
			Nilable().
			Min(1)

		testMap := map[any]any{"key": 42}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testMap, *result)

		// Test nil handling
		result, err = mapSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		defaultMap := map[any]any{"default": 1}
		prefaultMap := map[any]any{"prefault": 2}
		mapSchema := Map(String(), Int()).
			Default(defaultMap).
			Prefault(prefaultMap)

		testMap := map[any]any{"test": 3}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestMap_ValidationMethods(t *testing.T) {
	t.Run("Min validation", func(t *testing.T) {
		mapSchema := Map(String(), Int()).Min(2)

		// Should pass with 2+ entries
		validMap := map[any]any{"key1": 1, "key2": 2}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// Should fail with < 2 entries
		invalidMap := map[any]any{"key1": 1}
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("Max validation", func(t *testing.T) {
		mapSchema := Map(String(), Int()).Max(2)

		// Should pass with <= 2 entries
		validMap := map[any]any{"key1": 1, "key2": 2}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// Should fail with > 2 entries
		invalidMap := map[any]any{"key1": 1, "key2": 2, "key3": 3}
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("Size validation", func(t *testing.T) {
		mapSchema := Map(String(), Int()).Length(2)

		// Should pass with exactly 2 entries
		validMap := map[any]any{"key1": 1, "key2": 2}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// Should fail with != 2 entries
		invalidMap1 := map[any]any{"key1": 1}
		_, err = mapSchema.Parse(invalidMap1)
		assert.Error(t, err)

		invalidMap2 := map[any]any{"key1": 1, "key2": 2, "key3": 3}
		_, err = mapSchema.Parse(invalidMap2)
		assert.Error(t, err)
	})

	t.Run("NonEmpty validation", func(t *testing.T) {
		mapSchema := Map(String(), Int()).NonEmpty()

		// Should pass with 1+ entries
		validMap := map[any]any{"key1": 1}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// Should pass with multiple entries
		multiMap := map[any]any{"key1": 1, "key2": 2, "key3": 3}
		result, err = mapSchema.Parse(multiMap)
		require.NoError(t, err)
		assert.Equal(t, multiMap, result)

		// Should fail with empty map
		emptyMap := map[any]any{}
		_, err = mapSchema.Parse(emptyMap)
		assert.Error(t, err)
	})

	t.Run("NonEmpty with custom error message", func(t *testing.T) {
		customError := "Map cannot be empty"
		mapSchema := Map(String(), Int()).NonEmpty(customError)

		// Should fail with custom error message
		emptyMap := map[any]any{}
		_, err := mapSchema.Parse(emptyMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), customError)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestMap_DefaultAndPrefault(t *testing.T) {
	t.Run("default value behavior", func(t *testing.T) {
		defaultMap := map[any]any{"default": 1}
		mapSchema := Map(String(), Int()).Default(defaultMap)

		// Valid input should override default
		testMap := map[any]any{"test": 2}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)

		// Test default function
		mapFunc := Map(String(), Int()).DefaultFunc(func() map[any]any {
			return map[any]any{"func": 1}
		})
		result2, err := mapFunc.Parse(map[any]any{"test": 2})
		require.NoError(t, err)
		assert.Equal(t, map[any]any{"test": 2}, result2)
	})

	t.Run("prefault value behavior", func(t *testing.T) {
		prefaultMap := map[any]any{"prefault": 1}
		mapSchema := Map(String(), Int()).Prefault(prefaultMap)

		// Valid input should work normally
		testMap := map[any]any{"test": 2}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)

		// Test prefault function
		mapFunc := Map(String(), Int()).PrefaultFunc(func() map[any]any {
			return map[any]any{"func": 1}
		})
		result2, err := mapFunc.Parse(map[any]any{"test": 2})
		require.NoError(t, err)
		assert.Equal(t, map[any]any{"test": 2}, result2)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestMap_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		// Only accept maps with more than 1 entry
		mapSchema := Map(String(), Int()).Refine(func(m map[any]any) bool {
			return len(m) > 1
		})

		validMap := map[any]any{"key1": 1, "key2": 2}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		invalidMap := map[any]any{"key1": 1}
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Map must have at least 2 entries"
		mapSchema := Map(String(), Int()).Refine(func(m map[any]any) bool {
			return len(m) >= 2
		}, core.SchemaParams{Error: errorMessage})

		validMap := map[any]any{"key1": 1, "key2": 2}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		invalidMap := map[any]any{"key1": 1}
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("refine nilable map", func(t *testing.T) {
		mapSchema := Map(String(), Int()).Nilable().Refine(func(m *map[any]any) bool {
			// Allow nil or maps with 0 or > 1 entries
			if m == nil {
				return true
			}
			return len(*m) == 0 || len(*m) > 1
		})

		// nil should pass
		result, err := mapSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// empty map should pass and return pointer
		result, err = mapSchema.Parse(map[any]any{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[any]any{}, *result)

		// map with > 1 entries should pass and return pointer
		validMap := map[any]any{"key1": 1, "key2": 2}
		result, err = mapSchema.Parse(validMap)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validMap, *result)

		// map with exactly 1 entry should fail
		invalidMap := map[any]any{"key1": 1}
		_, err = mapSchema.Parse(invalidMap)
		assert.Error(t, err)
	})
}

func TestMap_RefineAny(t *testing.T) {
	t.Run("refineAny flexible validation", func(t *testing.T) {
		mapSchema := Map(String(), Int()).RefineAny(func(v any) bool {
			m, ok := v.(map[any]any)
			return ok && len(m) >= 1
		})

		// map with >= 1 entries should pass
		validMap := map[any]any{"key1": 1}
		result, err := mapSchema.Parse(validMap)
		require.NoError(t, err)
		assert.Equal(t, validMap, result)

		// empty map should fail
		_, err = mapSchema.Parse(map[any]any{})
		assert.Error(t, err)
	})

	t.Run("refineAny with type checking", func(t *testing.T) {
		mapSchema := Map(String(), Int()).RefineAny(func(v any) bool {
			m, ok := v.(map[any]any)
			if !ok {
				return false
			}
			// Only allow maps with even number of entries
			return len(m)%2 == 0
		})

		evenMap := map[any]any{"key1": 1, "key2": 2}
		result, err := mapSchema.Parse(evenMap)
		require.NoError(t, err)
		assert.Equal(t, evenMap, result)

		oddMap := map[any]any{"key1": 1}
		_, err = mapSchema.Parse(oddMap)
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestMap_ErrorHandling(t *testing.T) {
	t.Run("invalid map type error", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		_, err := mapSchema.Parse("not a map")
		assert.Error(t, err)
	})

	t.Run("key validation error", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		invalidMap := map[any]any{123: 42} // int key instead of string
		_, err := mapSchema.Parse(invalidMap)
		assert.Error(t, err)
		// Now we expect the original error message without wrapping
		assert.Contains(t, err.Error(), "Invalid input: expected string, received number")
	})

	t.Run("value validation error", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		invalidMap := map[any]any{"key": "not_an_int"}
		_, err := mapSchema.Parse(invalidMap)
		assert.Error(t, err)
		// Now we expect the original error message without wrapping
		assert.Contains(t, err.Error(), "Invalid input: expected int, received string")
	})

	t.Run("custom error message", func(t *testing.T) {
		mapSchema := Map(String(), Int(), core.SchemaParams{Error: "Expected a valid map"})

		_, err := mapSchema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestMap_EdgeCases(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		result, err := mapSchema.Parse(map[any]any{})
		require.NoError(t, err)
		assert.Equal(t, map[any]any{}, result)
	})

	t.Run("nil handling with nilable map", func(t *testing.T) {
		mapSchema := Map(String(), Int()).Nilable()

		// Test nil input
		result, err := mapSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid map - should return pointer
		testMap := map[any]any{"key": 42}
		result, err = mapSchema.Parse(testMap)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testMap, *result)
	})

	t.Run("empty context", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		// Parse with empty context slice
		testMap := map[any]any{"key": 42}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})

	t.Run("map with nil schemas", func(t *testing.T) {
		// Test with nil key and value schemas
		mapSchema := Map(nil, nil)

		testMap := map[any]any{"any": "any"}
		result, err := mapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})

	t.Run("large map performance", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		// Create a large map
		largeMap := make(map[any]any)
		for i := range 1000 {
			largeMap[fmt.Sprintf("key%d", i)] = i
		}

		result, err := mapSchema.Parse(largeMap)
		require.NoError(t, err)
		assert.Equal(t, largeMap, result)
		assert.Equal(t, 1000, len(result))
	})

	t.Run("deeply nested map validation", func(t *testing.T) {
		// Map of string to map of string to int
		innerMapSchema := Map(String(), Int())
		outerMapSchema := Map(String(), innerMapSchema)

		testMap := map[any]any{
			"outer1": map[any]any{
				"inner1": 1,
				"inner2": 2,
			},
			"outer2": map[any]any{
				"inner3": 3,
			},
		}

		result, err := outerMapSchema.Parse(testMap)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})

	t.Run("mixed type validation complexity", func(t *testing.T) {
		// Test with different combinations of schemas
		schemas := []struct {
			name      string
			keySchema any
			valSchema any
		}{
			{"string-bool", String(), Bool()},
			{"int-string", Int(), String()},
			{"bool-float", Bool(), Float64()},
			{"enum-int", Enum("a", "b", "c"), Int()},
		}

		for _, schema := range schemas {
			t.Run(schema.name, func(t *testing.T) {
				mapSchema := Map(schema.keySchema, schema.valSchema)
				require.NotNil(t, mapSchema)

				// Test with appropriate values based on key type
				var testMap map[any]any
				switch schema.name {
				case "string-bool":
					testMap = map[any]any{"key": true}
				case "int-string":
					testMap = map[any]any{42: "value"}
				case "bool-float":
					testMap = map[any]any{true: 3.14}
				case "enum-int":
					testMap = map[any]any{"a": 1}
				}

				result, err := mapSchema.Parse(testMap)
				require.NoError(t, err)
				assert.Equal(t, testMap, result)
			})
		}
	})

	t.Run("pointer value handling", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		// Test with pointer to map
		testMap := map[any]any{"key": 42}
		testMapPtr := &testMap

		result, err := mapSchema.Parse(testMapPtr)
		require.NoError(t, err)
		assert.Equal(t, testMap, result)
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		mapSchema := Map(String(), Int())
		testMap := map[any]any{"key": 42}

		// Run multiple goroutines parsing the same schema
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for range numGoroutines {
			go func() {
				_, err := mapSchema.Parse(testMap)
				results <- err
			}()
		}

		// Check all results
		for range numGoroutines {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("transform operations", func(t *testing.T) {
		mapSchema := Map(String(), Int())

		// Test Transform
		transform := mapSchema.Transform(func(m map[any]any, ctx *core.RefinementContext) (any, error) {
			return len(m), nil
		})
		require.NotNil(t, transform)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestMap_Overwrite(t *testing.T) {
	t.Run("basic map transformation", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), String()).
			Overwrite(func(m map[any]any) map[any]any {
				// Convert all string values to uppercase
				result := make(map[any]any)
				for k, v := range m {
					if strVal, ok := v.(string); ok {
						result[k] = strings.ToUpper(strVal)
					} else {
						result[k] = v
					}
				}
				return result
			})

		input := map[any]any{
			"name":    "john",
			"city":    "seattle",
			"country": "usa",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[any]any{
			"name":    "JOHN",
			"city":    "SEATTLE",
			"country": "USA",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("map key transformation", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), Int()).
			Overwrite(func(m map[any]any) map[any]any {
				// Convert string keys to uppercase and increment values
				result := make(map[any]any)
				for k, v := range m {
					newKey := k
					if strKey, ok := k.(string); ok {
						newKey = strings.ToUpper(strKey)
					}

					if intVal, ok := v.(int); ok {
						result[newKey] = intVal + 10
					} else {
						result[newKey] = v
					}
				}
				return result
			})

		input := map[any]any{
			"a": 1,
			"b": 2,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[any]any{
			"A": 11,
			"B": 12,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("filtering transformation", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), Int()).
			Overwrite(func(m map[any]any) map[any]any {
				// Filter out negative values
				result := make(map[any]any)
				for k, v := range m {
					if intVal, ok := v.(int); ok && intVal >= 0 {
						result[k] = intVal
					}
				}
				return result
			})

		input := map[any]any{
			"positive": 5,
			"negative": -3,
			"zero":     0,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[any]any{
			"positive": 5,
			"zero":     0,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("complex key types", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), String()).
			Overwrite(func(m map[any]any) map[any]any {
				// Transform based on key type
				result := make(map[any]any)
				for k, v := range m {
					switch key := k.(type) {
					case string:
						// String keys: uppercase value
						if strVal, ok := v.(string); ok {
							result[key] = strings.ToUpper(strVal)
						} else {
							result[key] = v
						}
					case int:
						// Integer keys: prepend "num_" to value
						if strVal, ok := v.(string); ok {
							result[key] = "num_" + strVal
						} else {
							result[key] = v
						}
					default:
						result[key] = v
					}
				}
				return result
			})

		input := map[any]any{
			"name": "alice",
			123:    "value",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[any]any{
			"name": "ALICE",
			123:    "num_value",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("chaining with validations", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), String()).
			Overwrite(func(m map[any]any) map[any]any {
				// Trim whitespace from all string values
				result := make(map[any]any)
				for k, v := range m {
					if strVal, ok := v.(string); ok {
						result[k] = strings.TrimSpace(strVal)
					} else {
						result[k] = v
					}
				}
				return result
			}).
			Min(1).
			Max(5)

		input := map[any]any{
			"name": "  John  ",
			"city": "  Seattle  ",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[any]any{
			"name": "John",
			"city": "Seattle",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := MapTyped[map[any]any, *map[any]any](Any(), String()).
			Overwrite(func(m *map[any]any) *map[any]any {
				if m == nil {
					return nil
				}
				// Convert values to lowercase
				result := make(map[any]any)
				for k, v := range *m {
					if strVal, ok := v.(string); ok {
						result[k] = strings.ToLower(strVal)
					} else {
						result[k] = v
					}
				}
				return &result
			})

		input := map[any]any{
			"MESSAGE": "HELLO WORLD",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		require.NotNil(t, result)

		expected := map[any]any{
			"MESSAGE": "hello world",
		}
		assert.Equal(t, expected, *result)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), Bool()).
			Overwrite(func(m map[any]any) map[any]any {
				return m // Identity transformation
			})

		input := map[any]any{
			"flag": true,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, map[any]any{}, result)
		assert.Equal(t, input, result)
	})

	t.Run("empty map handling", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), String()).
			Overwrite(func(m map[any]any) map[any]any {
				if len(m) == 0 {
					// Add default entry for empty maps
					return map[any]any{"default": "empty"}
				}
				return m
			})

		// Test with empty map
		result, err := schema.Parse(map[any]any{})
		require.NoError(t, err)

		expected := map[any]any{"default": "empty"}
		assert.Equal(t, expected, result)
	})

	t.Run("mixed value types", func(t *testing.T) {
		schema := MapTyped[map[any]any, map[any]any](Any(), Any()).
			Overwrite(func(m map[any]any) map[any]any {
				// Transform different value types
				result := make(map[any]any)
				for k, v := range m {
					switch val := v.(type) {
					case string:
						result[k] = strings.ToUpper(val)
					case int:
						result[k] = val * 2
					case bool:
						result[k] = !val
					default:
						result[k] = val
					}
				}
				return result
			})

		input := map[any]any{
			"str":  "hello",
			"num":  5,
			"bool": false,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[any]any{
			"str":  "HELLO",
			"num":  10,
			"bool": true,
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestMap_Check(t *testing.T) {
	t.Run("adds multiple issues for invalid input", func(t *testing.T) {
		schema := Map(String(), Int()).Check(func(value map[any]any, p *core.ParsePayload) {
			if len(value) == 0 {
				p.AddIssueWithMessage("map cannot be empty")
			}
			if len(value) > 2 {
				p.AddIssueWithCode(core.TooBig, "too many pairs")
			}
		})

		_, err := schema.Parse(map[any]any{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)

		_, err = schema.Parse(map[any]any{"a": 1, "b": 2, "c": 3})
		require.Error(t, err)
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})

	t.Run("works with pointer schema and value map", func(t *testing.T) {
		schema := MapPtr(String(), Int()).Check(func(value *map[any]any, p *core.ParsePayload) {
			if value == nil || len(*value) == 0 {
				p.AddIssueWithMessage("pointer map is empty")
			}
		})

		_, err := schema.Parse(map[any]any{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})
}

func TestMap_NonOptional(t *testing.T) {
	schema := Map(String(), Int64()).NonOptional()

	m := map[any]any{"a": int64(1)}
	_, err := schema.Parse(m)
	require.NoError(t, err)

	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}
}

// =============================================================================
// Multi-error collection tests (TypeScript Zod v4 behavior)
// =============================================================================

func TestMap_MultiErrorCollection(t *testing.T) {
	t.Run("collect all key-value validation errors", func(t *testing.T) {
		// Create a map schema with validations that will fail for multiple entries
		keySchema := String().Min(5, "Key must be at least 5 characters")
		valueSchema := Int().Min(10, "Value must be at least 10")
		mapSchema := Map(keySchema, valueSchema)

		// Test with invalid data that will trigger multiple errors
		invalidMap := map[any]any{
			"a":     5,  // Key too short (< 5 chars), Value too small (< 10)
			"bb":    8,  // Key too short (< 5 chars), Value too small (< 10)
			"valid": 15, // Valid key and value
			"c":     3,  // Key too short (< 5 chars), Value too small (< 10)
		}

		_, err := mapSchema.Parse(invalidMap)
		require.Error(t, err)

		// Check that it's a ZodError with multiple issues
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 6 validation errors (3 key errors + 3 value errors)
		assert.Len(t, zodErr.Issues, 6)

		// Check that all errors have correct paths and codes
		keyErrorsCount := 0
		valueErrorsCount := 0

		for _, issue := range zodErr.Issues {
			assert.Len(t, issue.Path, 1, "Each error should have map key in path")
			key := issue.Path[0]

			// Verify path corresponds to invalid keys
			assert.Contains(t, []any{"a", "bb", "c"}, key, "Error should be for invalid keys only")

			// Count error types
			if issue.Code == core.IssueCode("too_small") {
				if key == "a" || key == "bb" || key == "c" {
					// Could be either key or value error for these keys
					if strings.Contains(issue.Message, "Key must be at least 5 characters") {
						keyErrorsCount++
					} else if strings.Contains(issue.Message, "Value must be at least 10") {
						valueErrorsCount++
					}
				}
			}
		}

		// Should have errors for both keys and values
		assert.Equal(t, 3, keyErrorsCount, "Should have 3 key validation errors")
		assert.Equal(t, 3, valueErrorsCount, "Should have 3 value validation errors")
	})

	t.Run("collect key-only validation errors", func(t *testing.T) {
		// Create schema that only validates keys (no value validation)
		keySchema := String().Email("Invalid email format")
		mapSchema := Map(keySchema, nil) // No value schema

		// Test with invalid keys
		invalidMap := map[any]any{
			"not-email1":        "value1",
			"not-email2":        "value2",
			"valid@example.com": "value3", // Valid key
			"not-email3":        "value4",
		}

		_, err := mapSchema.Parse(invalidMap)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 3 key validation errors (only for invalid email keys)
		assert.Len(t, zodErr.Issues, 3)

		for _, issue := range zodErr.Issues {
			assert.Equal(t, core.IssueCode("invalid_format"), issue.Code)
			assert.Contains(t, issue.Message, "Invalid email format")
			assert.Len(t, issue.Path, 1, "Each error should have map key in path")
			key := issue.Path[0]
			assert.Contains(t, []any{"not-email1", "not-email2", "not-email3"}, key)
		}
	})

	t.Run("collect value-only validation errors", func(t *testing.T) {
		// Create schema that only validates values (no key validation)
		valueSchema := Int().Max(100, "Value must be at most 100")
		mapSchema := Map(nil, valueSchema) // No key schema

		// Test with invalid values
		invalidMap := map[any]any{
			"key1": 150, // Too large
			"key2": 50,  // Valid value
			"key3": 200, // Too large
			"key4": 300, // Too large
		}

		_, err := mapSchema.Parse(invalidMap)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 3 value validation errors
		assert.Len(t, zodErr.Issues, 3)

		for _, issue := range zodErr.Issues {
			assert.Equal(t, core.IssueCode("too_big"), issue.Code)
			assert.Contains(t, issue.Message, "Value must be at most 100")
			assert.Len(t, issue.Path, 1, "Each error should have map key in path")
			key := issue.Path[0]
			assert.Contains(t, []any{"key1", "key3", "key4"}, key)
		}
	})

	t.Run("mixed valid and invalid entries", func(t *testing.T) {
		// Test with a mix of valid and invalid entries
		keySchema := String().Min(3)
		valueSchema := Int().Min(5)
		mapSchema := Map(keySchema, valueSchema)

		mixedMap := map[any]any{
			"ab":   1,  // Key too short, Value too small
			"abc":  10, // Valid key and value
			"abcd": 3,  // Valid key, Value too small
			"xy":   15, // Key too short, Valid value
		}

		_, err := mapSchema.Parse(mixedMap)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should collect all validation errors, no early termination
		assert.Len(t, zodErr.Issues, 4) // 2 key errors + 2 value errors

		// Verify all errors have correct path structure
		for _, issue := range zodErr.Issues {
			assert.Len(t, issue.Path, 1, "Each error should have map key in path")
			key := issue.Path[0]
			assert.Contains(t, []any{"ab", "abcd", "xy"}, key, "Errors should only be for invalid entries")
		}
	})
}
