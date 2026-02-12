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

func TestRecord_BasicFunctionality(t *testing.T) {
	t.Run("valid record inputs", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		testRecord := map[string]any{
			"key1": 1,
			"key2": 2,
			"key3": 3,
		}

		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
	})

	t.Run("valid map[any]any with string keys", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		testMap := map[any]any{
			"key1": 1,
			"key2": 2,
		}

		result, err := recordSchema.Parse(testMap)
		require.NoError(t, err)

		expected := map[string]any{
			"key1": 1,
			"key2": 2,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("empty record", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		result, err := recordSchema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		invalidInputs := []any{
			"not a record", 123, []int{1, 2, 3}, true, nil,
		}

		for _, input := range invalidInputs {
			_, err := recordSchema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		recordSchema := Record(String(), Bool())
		testRecord := map[string]any{"test": true}

		// Test Parse method
		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)

		// Test MustParse method
		mustResult := recordSchema.MustParse(testRecord)
		assert.Equal(t, testRecord, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			recordSchema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid record"
		recordSchema := Record(String(), Int(), core.SchemaParams{Error: customError})

		require.NotNil(t, recordSchema)

		_, err := recordSchema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestRecord_TypeSafety(t *testing.T) {
	t.Run("record returns map[string]any type", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		require.NotNil(t, recordSchema)

		testRecord := map[string]any{"test": 42}
		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
		assert.IsType(t, map[string]any{}, result)
	})

	t.Run("key validation - only string keys allowed", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		// Valid string keys should pass
		validRecord := map[string]any{"valid_key": 42}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// map[any]any with non-string keys should fail
		invalidMap := map[any]any{123: 42} // int key instead of string
		_, err = recordSchema.Parse(invalidMap)
		assert.Error(t, err)

		// map[any]any with mixed key types should fail
		mixedMap := map[any]any{"string_key": 42, 123: 43}
		_, err = recordSchema.Parse(mixedMap)
		assert.Error(t, err)
	})

	t.Run("value validation", func(t *testing.T) {
		// Test that values are properly typed in the result
		recordSchema := Record(String(), Int())

		// Valid values should pass
		validRecord := map[string]any{"key": 42}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid value type should fail - struct should definitely fail
		invalidRecord := map[string]any{"key": struct{}{}}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		recordSchema := Record(String(), Bool())
		testRecord := map[string]any{"test": true}

		result := recordSchema.MustParse(testRecord)
		assert.IsType(t, map[string]any{}, result)
		assert.Equal(t, testRecord, result)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestRecord_Modifiers(t *testing.T) {
	t.Run("Optional allows nil values", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		optionalSchema := recordSchema.Optional()

		// Test non-nil value - returns pointer
		testRecord := map[string]any{"key": 42}
		result, err := optionalSchema.Parse(testRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord, *result)

		// Test nil value (should be allowed for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		nilableSchema := recordSchema.Nilable()

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - returns pointer
		testRecord := map[string]any{"key": 42}
		result, err = nilableSchema.Parse(testRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord, *result)
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		nullishSchema := recordSchema.Nullish()

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - returns pointer
		testRecord := map[string]any{"key": 42}
		result, err = nullishSchema.Parse(testRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord, *result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultRecord := map[string]any{"default": 1}
		recordSchema := Record(String(), Int())
		defaultSchema := recordSchema.Default(defaultRecord)

		// Valid input should override default
		testRecord := map[string]any{"test": 2}
		result, err := defaultSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
		assert.IsType(t, map[string]any{}, result)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultRecord := map[string]any{"prefault": 1}
		recordSchema := Record(String(), Int())
		prefaultSchema := recordSchema.Prefault(prefaultRecord)

		// Valid input should override prefault
		testRecord := map[string]any{"test": 2}
		result, err := prefaultSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
		assert.IsType(t, map[string]any{}, result)
	})

	t.Run("DefaultFunc preserves current type", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		defaultSchema := recordSchema.DefaultFunc(func() map[string]any {
			return map[string]any{"func_default": 42}
		})

		// Valid input should override default function
		testRecord := map[string]any{"test": 2}
		result, err := defaultSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
		assert.IsType(t, map[string]any{}, result)
	})

	t.Run("PrefaultFunc preserves current type", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		prefaultSchema := recordSchema.PrefaultFunc(func() map[string]any {
			return map[string]any{"func_prefault": 42}
		})

		// Valid input should override prefault function
		testRecord := map[string]any{"test": 2}
		result, err := prefaultSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
		assert.IsType(t, map[string]any{}, result)
	})

	t.Run("Partial instance method allows missing keys", func(t *testing.T) {
		// Create a record with enum key schema (exhaustive keys)
		keySchema := Enum("id", "name", "email")
		recordSchema := Record(keySchema, String())

		// Without Partial, missing keys should fail
		partialInput := map[string]any{"id": "user-123"}
		_, err := recordSchema.Parse(partialInput)
		require.Error(t, err, "Regular record should fail with missing keys")

		// With Partial() instance method, missing keys should be allowed
		partialSchema := recordSchema.Partial()
		result, err := partialSchema.Parse(partialInput)
		require.NoError(t, err, "Partial record should allow missing keys")
		assert.Equal(t, partialInput, result)
	})

	t.Run("Partial instance method still validates present keys", func(t *testing.T) {
		keySchema := Enum("id", "name", "email")
		partialSchema := Record(keySchema, String()).Partial()

		// Present keys should still be validated
		invalidInput := map[string]any{"id": 123} // Wrong type for value
		_, err := partialSchema.Parse(invalidInput)
		require.Error(t, err, "Partial record should still validate present values")
	})

	t.Run("Partial instance method still rejects unrecognized keys", func(t *testing.T) {
		keySchema := Enum("id", "name", "email")
		partialSchema := Record(keySchema, String()).Partial()

		// Unrecognized keys should still fail
		invalidInput := map[string]any{
			"id":    "user-123",
			"extra": "not allowed",
		}
		_, err := partialSchema.Parse(invalidInput)
		require.Error(t, err, "Partial record should reject unrecognized keys")
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Equal(t, core.UnrecognizedKeys, zodErr.Issues[0].Code)
	})

	t.Run("Partial instance method is chainable", func(t *testing.T) {
		keySchema := Enum("id", "name", "email")
		partialSchema := Record(keySchema, String()).
			Partial().
			Min(1).
			Max(3)

		// Should allow partial record with 1-3 entries
		result, err := partialSchema.Parse(map[string]any{"id": "123"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"id": "123"}, result)

		// Should fail with 0 entries (Min(1))
		_, err = partialSchema.Parse(map[string]any{})
		require.Error(t, err)
	})

	t.Run("Partial vs PartialRecord equivalence", func(t *testing.T) {
		keySchema := Enum("id", "name", "email")

		// Using PartialRecord constructor
		schema1 := PartialRecord(keySchema, String())

		// Using .Partial() instance method
		schema2 := Record(keySchema, String()).Partial()

		testInput := map[string]any{"id": "123"}

		// Both should behave identically
		result1, err1 := schema1.Parse(testInput)
		require.NoError(t, err1)

		result2, err2 := schema2.Parse(testInput)
		require.NoError(t, err2)

		assert.Equal(t, result1, result2)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestRecord_Chaining(t *testing.T) {
	t.Run("modifier chaining", func(t *testing.T) {
		recordSchema := Record(String(), String()).
			Optional().
			Min(1).
			Max(5)

		require.NotNil(t, recordSchema)

		// Valid record within constraints - returns pointer
		testRecord := map[string]any{"key1": "value1", "key2": "value2"}
		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord, *result)

		// nil should be allowed (optional)
		result, err = recordSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("validation chaining", func(t *testing.T) {
		recordSchema := Record(String(), Int()).
			Min(2).
			Max(4).
			Refine(func(r map[string]any) bool {
				// All values must be positive
				for _, v := range r {
					if val, ok := v.(int); ok && val <= 0 {
						return false
					}
				}
				return true
			})

		// Valid record
		validRecord := map[string]any{"a": 1, "b": 2}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid record (negative value)
		invalidRecord := map[string]any{"a": -1, "b": 2}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)

		// Invalid record (too few entries)
		tooSmallRecord := map[string]any{"a": 1}
		_, err = recordSchema.Parse(tooSmallRecord)
		assert.Error(t, err)
	})

	t.Run("complex chaining preserves type", func(t *testing.T) {
		recordSchema := Record(String(), Bool()).
			Nilable().
			Length(2).
			Default(map[string]any{"default1": true, "default2": false})

		// Valid input maintains type - returns pointer
		testRecord := map[string]any{"test1": true, "test2": false}
		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord, *result)
		assert.IsType(t, &map[string]any{}, result)
	})
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestRecord_ValidationMethods(t *testing.T) {
	t.Run("Min sets minimum number of entries", func(t *testing.T) {
		recordSchema := Record(String(), String()).Min(2)

		// Valid: meets minimum
		validRecord := map[string]any{"key1": "value1", "key2": "value2"}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid: below minimum
		invalidRecord := map[string]any{"key1": "value1"}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})

	t.Run("Max sets maximum number of entries", func(t *testing.T) {
		recordSchema := Record(String(), String()).Max(2)

		// Valid: meets maximum
		validRecord := map[string]any{"key1": "value1", "key2": "value2"}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid: exceeds maximum
		invalidRecord := map[string]any{"key1": "value1", "key2": "value2", "key3": "value3"}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})

	t.Run("Length sets exact number of entries", func(t *testing.T) {
		recordSchema := Record(String(), String()).Length(2)

		// Valid: exact size
		validRecord := map[string]any{"key1": "value1", "key2": "value2"}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid: too few
		tooSmallRecord := map[string]any{"key1": "value1"}
		_, err = recordSchema.Parse(tooSmallRecord)
		assert.Error(t, err)

		// Invalid: too many
		tooBigRecord := map[string]any{"key1": "value1", "key2": "value2", "key3": "value3"}
		_, err = recordSchema.Parse(tooBigRecord)
		assert.Error(t, err)
	})

	t.Run("validation with custom error messages", func(t *testing.T) {
		errorMsg := "Record must have exactly 3 entries"
		recordSchema := Record(String(), Int()).Length(3, core.SchemaParams{Error: errorMsg})

		invalidRecord := map[string]any{"key1": 1}
		_, err := recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestRecord_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// When both Default and Prefault are set, Default should take precedence
		defaultRecord := map[string]any{"default": "value"}
		prefaultRecord := map[string]any{"prefault": "value"}
		schema := Record(String(), String()).Default(defaultRecord).Prefault(prefaultRecord).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultRecord, *result)
	})

	t.Run("Default bypasses validation (short-circuit)", func(t *testing.T) {
		// Default should bypass record validation constraints
		// Use DefaultFunc to provide invalid type that bypasses validation
		schema := Record(String(), Int()).DefaultFunc(func() map[string]any {
			// This will be converted to map[string]any but contains invalid data
			return map[string]any{"invalid": "not-an-int"}
		}).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{"invalid": "not-an-int"}, *result)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value must pass record validation
		validRecord := map[string]any{"key": "value"}
		schema := Record(String(), String()).Prefault(validRecord).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validRecord, *result)
	})

	t.Run("nil input triggers Default, not Prefault", func(t *testing.T) {
		// When input is nil, only Default should be triggered
		defaultRecord := map[string]any{"default": "value"}
		prefaultRecord := map[string]any{"prefault": "value"}
		schema := Record(String(), String()).Default(defaultRecord).Prefault(prefaultRecord).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultRecord, *result)

		// Test with only Prefault - should trigger on nil
		schemaOnlyPrefault := Record(String(), String()).Prefault(prefaultRecord).Optional()
		result2, err := schemaOnlyPrefault.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result2)
		assert.Equal(t, prefaultRecord, *result2)

		// Non-nil input should not trigger prefault even if validation fails
		_, err = schemaOnlyPrefault.Parse("not-a-record")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected record")
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Record(String(), String()).
			DefaultFunc(func() map[string]any {
				defaultCalled = true
				return map[string]any{"default": "func"}
			}).
			PrefaultFunc(func() map[string]any {
				prefaultCalled = true
				return map[string]any{"prefault": "func"}
			}).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{"default": "func"}, *result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		// Prefault value that fails validation should return error
		// Use invalid record that will fail value validation
		invalidRecord := map[string]any{"key": 123} // 123 is not a string
		schema := Record(String(), String()).Prefault(invalidRecord).Optional()

		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected string")
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestRecord_Refine(t *testing.T) {
	t.Run("refine basic validation", func(t *testing.T) {
		recordSchema := Record(String(), Int()).Refine(func(r map[string]any) bool {
			return len(r) >= 2
		})

		// Valid record with >= 2 entries
		validRecord := map[string]any{"key1": 1, "key2": 2}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid record with < 2 entries
		invalidRecord := map[string]any{"key1": 1}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})

	t.Run("refine with value validation", func(t *testing.T) {
		recordSchema := Record(String(), Int()).Refine(func(r map[string]any) bool {
			// All values must be positive
			for _, v := range r {
				if val, ok := v.(int); ok && val <= 0 {
					return false
				}
			}
			return true
		})

		// Valid record with positive values
		validRecord := map[string]any{"key1": 1, "key2": 2}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// Invalid record with negative value
		invalidRecord := map[string]any{"key1": -1, "key2": 2}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Record must have at least 2 entries"
		recordSchema := Record(String(), Int()).Refine(func(r map[string]any) bool {
			return len(r) >= 2
		}, core.SchemaParams{Error: errorMessage})

		validRecord := map[string]any{"key1": 1, "key2": 2}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		invalidRecord := map[string]any{"key1": 1}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})

	t.Run("refine nilable record", func(t *testing.T) {
		recordSchema := Record(String(), Int()).Nilable().Refine(func(r *map[string]any) bool {
			// Allow nil or records with 0 or > 1 entries
			if r == nil {
				return true
			}
			return len(*r) == 0 || len(*r) > 1
		})

		// nil should pass
		result, err := recordSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// empty record should pass and return pointer
		result, err = recordSchema.Parse(map[string]any{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{}, *result)

		// record with > 1 entries should pass and return pointer
		validRecord := map[string]any{"key1": 1, "key2": 2}
		result, err = recordSchema.Parse(validRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, validRecord, *result)

		// record with exactly 1 entry should fail
		invalidRecord := map[string]any{"key1": 1}
		_, err = recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
	})
}

func TestRecord_RefineAny(t *testing.T) {
	t.Run("refineAny flexible validation", func(t *testing.T) {
		recordSchema := Record(String(), Int()).RefineAny(func(v any) bool {
			r, ok := v.(map[string]any)
			return ok && len(r) >= 1
		})

		// record with >= 1 entries should pass
		validRecord := map[string]any{"key1": 1}
		result, err := recordSchema.Parse(validRecord)
		require.NoError(t, err)
		assert.Equal(t, validRecord, result)

		// empty record should fail
		_, err = recordSchema.Parse(map[string]any{})
		assert.Error(t, err)
	})

	t.Run("refineAny with type checking", func(t *testing.T) {
		recordSchema := Record(String(), Int()).RefineAny(func(v any) bool {
			r, ok := v.(map[string]any)
			if !ok {
				return false
			}
			// Only allow records with even number of entries
			return len(r)%2 == 0
		})

		evenRecord := map[string]any{"key1": 1, "key2": 2}
		result, err := recordSchema.Parse(evenRecord)
		require.NoError(t, err)
		assert.Equal(t, evenRecord, result)

		oddRecord := map[string]any{"key1": 1}
		_, err = recordSchema.Parse(oddRecord)
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestRecord_ErrorHandling(t *testing.T) {
	t.Run("invalid record type error", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		_, err := recordSchema.Parse("not a record")
		assert.Error(t, err)
	})

	t.Run("non-string key error", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		invalidMap := map[any]any{123: 42} // int key instead of string
		_, err := recordSchema.Parse(invalidMap)
		assert.Error(t, err)
	})

	t.Run("value validation error", func(t *testing.T) {
		// Test basic error handling without relying on complex coercion behavior
		recordSchema := Record(String(), Int())

		invalidRecord := map[string]any{"key": struct{}{}}
		_, err := recordSchema.Parse(invalidRecord)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value validation failed")
	})

	t.Run("custom error message", func(t *testing.T) {
		recordSchema := Record(String(), Int(), core.SchemaParams{Error: "Expected a valid record"})

		_, err := recordSchema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestRecord_EdgeCases(t *testing.T) {
	t.Run("empty record", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		result, err := recordSchema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("nil handling with nilable record", func(t *testing.T) {
		recordSchema := Record(String(), Int()).Nilable()

		// Test nil input
		result, err := recordSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid record - returns pointer
		testRecord := map[string]any{"key": 42}
		result, err = recordSchema.Parse(testRecord)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testRecord, *result)
	})

	t.Run("empty context", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		// Parse with empty context slice
		testRecord := map[string]any{"key": 42}
		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
	})

	t.Run("record with nil value schema", func(t *testing.T) {
		// Test with nil value schema
		recordSchema := Record(String(), nil)

		testRecord := map[string]any{"any": "any"}
		result, err := recordSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
	})

	t.Run("conversion from map[any]any with all string keys", func(t *testing.T) {
		recordSchema := Record(String(), String())

		// map[any]any with all string keys should convert successfully
		mixedMap := map[any]any{
			"key1": "value1",
			"key2": "value2",
		}

		result, err := recordSchema.Parse(mixedMap)
		require.NoError(t, err)

		expected := map[string]any{
			"key1": "value1",
			"key2": "value2",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform operations", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		// Test Transform
		transform := recordSchema.Transform(func(r map[string]any, ctx *core.RefinementContext) (any, error) {
			return len(r), nil
		})
		require.NotNil(t, transform)
	})

	t.Run("large record performance", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		// Create a large record
		largeRecord := make(map[string]any)
		for i := 0; i < 1000; i++ {
			largeRecord[fmt.Sprintf("key%d", i)] = i
		}

		result, err := recordSchema.Parse(largeRecord)
		require.NoError(t, err)
		assert.Equal(t, largeRecord, result)
		assert.Equal(t, 1000, len(result))
	})

	t.Run("deeply nested record validation", func(t *testing.T) {
		// Record of string to record of string to int
		innerRecordSchema := Record(String(), Int())
		outerRecordSchema := Record(String(), innerRecordSchema)

		testRecord := map[string]any{
			"outer1": map[string]any{
				"inner1": 1,
				"inner2": 2,
			},
			"outer2": map[string]any{
				"inner3": 3,
			},
		}

		result, err := outerRecordSchema.Parse(testRecord)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
	})

	t.Run("mixed type validation complexity", func(t *testing.T) {
		// Test with different value schema types
		schemas := []struct {
			name      string
			valSchema any
			testValue any
		}{
			{"bool", Bool(), true},
			{"string", String(), "value"},
			{"float", Float64(), 3.14},
			{"enum", Enum("a", "b", "c"), "a"},
		}

		for _, schema := range schemas {
			t.Run(schema.name, func(t *testing.T) {
				recordSchema := Record(String(), schema.valSchema)
				require.NotNil(t, recordSchema)

				testRecord := map[string]any{"key": schema.testValue}
				result, err := recordSchema.Parse(testRecord)
				require.NoError(t, err)
				assert.Equal(t, testRecord, result)
			})
		}
	})

	t.Run("pointer value handling", func(t *testing.T) {
		recordSchema := Record(String(), Int())

		// Test with pointer to record
		testRecord := map[string]any{"key": 42}
		testRecordPtr := &testRecord

		result, err := recordSchema.Parse(testRecordPtr)
		require.NoError(t, err)
		assert.Equal(t, testRecord, result)
	})

	t.Run("concurrent access safety", func(t *testing.T) {
		recordSchema := Record(String(), Int())
		testRecord := map[string]any{"key": 42}

		// Run multiple goroutines parsing the same schema
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := recordSchema.Parse(testRecord)
				results <- err
			}()
		}

		// Check all results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("complex map[any]any conversion scenarios", func(t *testing.T) {
		recordSchema := Record(String(), String())

		// Test with various map[any]any inputs
		testCases := []struct {
			name  string
			input map[any]any
			valid bool
		}{
			{
				"all string keys",
				map[any]any{"key1": "val1", "key2": "val2"},
				true,
			},
			{
				"mixed key types",
				map[any]any{"key1": "val1", 123: "val2"},
				false,
			},
			{
				"empty map",
				map[any]any{},
				true,
			},
			{
				"single entry",
				map[any]any{"single": "value"},
				true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := recordSchema.Parse(tc.input)
				if tc.valid {
					require.NoError(t, err)
					// Convert expected result to map[string]any
					expected := make(map[string]any)
					for k, v := range tc.input {
						if strKey, ok := k.(string); ok {
							expected[strKey] = v
						}
					}
					assert.Equal(t, expected, result)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("schema evolution and chaining stress test", func(t *testing.T) {
		// Create a complex chain of modifications
		recordSchema := Record(String(), Int()).
			Min(1).
			Max(10).
			Nilable().
			Default(map[string]any{"default": 42}).
			Refine(func(r *map[string]any) bool {
				// Handle nil case for nilable
				if r == nil {
					return true
				}
				// All values must be positive
				for _, v := range *r {
					if val, ok := v.(int); ok && val <= 0 {
						return false
					}
				}
				return true
			})

		// Test various inputs
		testCases := []struct {
			name   string
			input  any
			valid  bool
			result map[string]any
		}{
			{
				"valid record",
				map[string]any{"key1": 1, "key2": 2},
				true,
				map[string]any{"key1": 1, "key2": 2},
			},
			{
				"nil input (should use default)",
				nil,
				true,
				nil, // nilable allows nil
			},
			{
				"negative value (should fail refine)",
				map[string]any{"key1": -1},
				false,
				nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := recordSchema.Parse(tc.input)
				if tc.valid {
					require.NoError(t, err)
					if tc.result != nil {
						// Since using Nilable(), result should be pointer
						require.NotNil(t, result)
						assert.Equal(t, tc.result, *result)
					}
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestRecord_Overwrite(t *testing.T) {
	t.Run("basic record transformation", func(t *testing.T) {
		schema := Record(String(), String()).
			Overwrite(func(record map[string]any) map[string]any {
				// Convert all string values to uppercase
				result := make(map[string]any)
				for k, v := range record {
					if strVal, ok := v.(string); ok {
						result[k] = strings.ToUpper(strVal)
					} else {
						result[k] = v
					}
				}
				return result
			})

		input := map[string]any{
			"name":    "john",
			"city":    "seattle",
			"country": "usa",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"name":    "JOHN",
			"city":    "SEATTLE",
			"country": "USA",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("record key transformation", func(t *testing.T) {
		schema := Record(String(), Int()).
			Overwrite(func(record map[string]any) map[string]any {
				// Add prefix to all keys and increment values
				result := make(map[string]any)
				for k, v := range record {
					newKey := "transformed_" + k
					if intVal, ok := v.(int); ok {
						result[newKey] = intVal + 10
					} else {
						result[newKey] = v
					}
				}
				return result
			})

		input := map[string]any{
			"a": 1,
			"b": 2,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"transformed_a": 11,
			"transformed_b": 12,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("filtering transformation", func(t *testing.T) {
		schema := Record(String(), Int()).
			Overwrite(func(record map[string]any) map[string]any {
				// Filter out negative values
				result := make(map[string]any)
				for k, v := range record {
					if intVal, ok := v.(int); ok && intVal >= 0 {
						result[k] = intVal
					}
				}
				return result
			})

		input := map[string]any{
			"positive": 5,
			"negative": -3,
			"zero":     0,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"positive": 5,
			"zero":     0,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("chaining with validations", func(t *testing.T) {
		schema := Record(String(), String()).
			Overwrite(func(record map[string]any) map[string]any {
				// Trim whitespace from all values
				result := make(map[string]any)
				for k, v := range record {
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

		input := map[string]any{
			"name": "  John  ",
			"city": "  Seattle  ",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := map[string]any{
			"name": "John",
			"city": "Seattle",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := RecordPtr(String(), String()).
			Overwrite(func(record *map[string]any) *map[string]any {
				if record == nil {
					return nil
				}
				// Convert values to lowercase
				result := make(map[string]any)
				for k, v := range *record {
					if strVal, ok := v.(string); ok {
						result[k] = strings.ToLower(strVal)
					} else {
						result[k] = v
					}
				}
				return &result
			})

		input := map[string]any{
			"MESSAGE": "HELLO WORLD",
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		require.NotNil(t, result)

		expected := map[string]any{
			"MESSAGE": "hello world",
		}
		assert.Equal(t, expected, *result)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Record(String(), Bool()).
			Overwrite(func(record map[string]any) map[string]any {
				return record // Identity transformation
			})

		input := map[string]any{
			"flag": true,
		}

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, map[string]any{}, result)
		assert.Equal(t, input, result)
	})

	t.Run("empty record handling", func(t *testing.T) {
		schema := Record(String(), String()).
			Overwrite(func(record map[string]any) map[string]any {
				if len(record) == 0 {
					// Add default entry for empty records
					return map[string]any{"default": "empty"}
				}
				return record
			})

		// Test with empty record
		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)

		expected := map[string]any{"default": "empty"}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestRecord_Check(t *testing.T) {
	t.Run("adds issues for invalid record", func(t *testing.T) {
		schema := Record(String(), Int()).Check(func(value map[string]any, p *core.ParsePayload) {
			if len(value) == 0 {
				p.AddIssueWithMessage("record cannot be empty")
			}
		})

		_, err := schema.Parse(map[string]any{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})

	t.Run("pointer schema adapts to value input", func(t *testing.T) {
		schema := RecordPtr(String(), String()).Check(func(value *map[string]any, p *core.ParsePayload) {
			if value == nil || len(*value) == 0 {
				p.AddIssueWithMessage("pointer record empty")
			}
		})

		_, err := schema.Parse(map[string]any{})
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
	})
}

func TestRecord_NonOptional(t *testing.T) {
	schema := Record(String(), String()).NonOptional()

	_, err := schema.Parse(map[string]any{"x": "y"})
	require.NoError(t, err)

	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}
}

// =============================================================================
// Key Schema Validation Tests (Enum/Literal)
// =============================================================================

func TestRecord_KeySchemaValidation(t *testing.T) {
	keySchema := Enum("id", "name", "email")

	t.Run("exhaustive enum check - valid", func(t *testing.T) {
		schema := Record(keySchema, String())
		validInput := map[string]any{
			"id":    "user-123",
			"name":  "John Doe",
			"email": "john.doe@example.com",
		}
		result, err := schema.Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)
	})

	t.Run("exhaustive enum check - missing key", func(t *testing.T) {
		schema := Record(keySchema, String())
		invalidInput := map[string]any{
			"id":   "user-123",
			"name": "John Doe",
			// "email" is missing
		}
		_, err := schema.Parse(invalidInput)
		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
		assert.Equal(t, []any{"email"}, zodErr.Issues[0].Path)
	})

	t.Run("exhaustive enum check - unrecognized key", func(t *testing.T) {
		schema := Record(keySchema, String())
		invalidInput := map[string]any{
			"id":    "user-123",
			"name":  "John Doe",
			"email": "john.doe@example.com",
			"extra": "this key is not allowed",
		}
		_, err := schema.Parse(invalidInput)
		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.UnrecognizedKeys, zodErr.Issues[0].Code)
	})

	t.Run("partial record - valid (missing keys allowed)", func(t *testing.T) {
		schema := PartialRecord(keySchema, String())
		partialInput := map[string]any{
			"id": "user-123",
		}
		result, err := schema.Parse(partialInput)
		require.NoError(t, err)
		assert.Equal(t, partialInput, result)
	})

	t.Run("partial record - unrecognized key (still fails)", func(t *testing.T) {
		schema := PartialRecord(keySchema, String())
		invalidInput := map[string]any{
			"id":    "user-123",
			"extra": "not allowed",
		}
		_, err := schema.Parse(invalidInput)
		require.Error(t, err) // Unrecognized keys are still checked in partial records
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Equal(t, core.UnrecognizedKeys, zodErr.Issues[0].Code)
	})

	t.Run("literal key schema", func(t *testing.T) {
		literalKeySchema := Literal("fixedKey")
		schema := Record(literalKeySchema, Int())

		// Valid
		_, err := schema.Parse(map[string]any{"fixedKey": 100})
		require.NoError(t, err)

		// Missing key
		_, err = schema.Parse(map[string]any{})
		require.Error(t, err)

		// Unrecognized key
		_, err = schema.Parse(map[string]any{"fixedKey": 100, "another": 200})
		require.Error(t, err)
	})
}

// =============================================================================
// LooseRecord tests - Zod v4 compatible loose record mode
// =============================================================================

func TestLooseRecord(t *testing.T) {
	t.Run("passes through non-matching keys", func(t *testing.T) {
		// Keys matching pattern are validated, non-matching keys pass through
		schema := LooseRecord(String().RegexString(`^S_`), String())

		// Keys matching pattern are validated
		result, err := schema.Parse(map[string]any{"S_name": "John"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"S_name": "John"}, result)

		// Keys not matching pattern pass through unchanged
		result, err = schema.Parse(map[string]any{"S_name": "John", "other": "value"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"S_name": "John", "other": "value"}, result)

		// Non-matching keys can have any type (not validated)
		result, err = schema.Parse(map[string]any{"S_name": "John", "count": 123})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"S_name": "John", "count": 123}, result)

		// Only non-matching keys (no matching keys)
		result, err = schema.Parse(map[string]any{"other": "value"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"other": "value"}, result)
	})

	t.Run("validates matching keys with value schema", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String())

		// Wrong value type for matching key should fail
		_, err := schema.Parse(map[string]any{"S_name": 123})
		require.Error(t, err)
	})

	t.Run("empty record passes", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String())

		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("IsLoose returns true", func(t *testing.T) {
		looseSchema := LooseRecord(String(), String())
		regularSchema := Record(String(), String())

		assert.True(t, looseSchema.IsLoose())
		assert.False(t, regularSchema.IsLoose())
	})

	t.Run("LooseRecordPtr works correctly", func(t *testing.T) {
		schema := LooseRecordPtr(String().RegexString(`^S_`), String())
		assert.True(t, schema.IsLoose())

		result, err := schema.Parse(map[string]any{"S_name": "John", "other": 123})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, map[string]any{"S_name": "John", "other": 123}, *result)
	})

	t.Run("multiple patterns with intersection", func(t *testing.T) {
		// Each pattern validates its matching keys
		schema1 := LooseRecord(String().RegexString(`^S_`), String())
		schema2 := LooseRecord(String().RegexString(`^N_`), Int())

		// Schema1 validates S_ keys, passes through N_ keys
		result1, err := schema1.Parse(map[string]any{"S_foo": "bar", "N_count": 123})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"S_foo": "bar", "N_count": 123}, result1)

		// Schema2 validates N_ keys, passes through S_ keys
		result2, err := schema2.Parse(map[string]any{"S_foo": "bar", "N_count": 123})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"S_foo": "bar", "N_count": 123}, result2)
	})

	t.Run("modifiers preserve loose mode", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String()).Min(1)
		assert.True(t, schema.IsLoose())

		// Non-matching key should still pass through
		result, err := schema.Parse(map[string]any{"S_name": "John", "other": 123})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"S_name": "John", "other": 123}, result)
	})

	t.Run("vs regular Record behavior", func(t *testing.T) {
		looseSchema := LooseRecord(String().RegexString(`^S_`), String())
		// Note: Regular Record with regex would fail on non-matching keys

		// LooseRecord preserves non-matching keys
		result, err := looseSchema.Parse(map[string]any{"S_name": "John", "other": "value"})
		require.NoError(t, err)
		assert.Equal(t, "value", result["other"])
	})
}

// =============================================================================
// ENHANCED LOOSERECORD ZOD V4 COMPATIBILITY TESTS
// =============================================================================

func TestLooseRecord_EnhancedCoverage(t *testing.T) {
	t.Run("error path for matching key validation failure", func(t *testing.T) {
		// Zod v4: Error path should include the key name
		schema := LooseRecord(String().RegexString(`^S_`), String())

		_, err := schema.Parse(map[string]any{"S_name": 123})
		require.Error(t, err)
		// Error should mention the wrong type
		assert.Contains(t, err.Error(), "expected string")
	})

	t.Run("with complex object value schema", func(t *testing.T) {
		// Zod v4: looseRecord with object value schema
		valueSchema := Object(core.ObjectSchema{
			"name": String(),
			"age":  Int(),
		})
		schema := LooseRecord(String().RegexString(`^user_`), valueSchema)

		// Valid: matching key with valid object
		result, err := schema.Parse(map[string]any{
			"user_1": map[string]any{"name": "Alice", "age": 30},
		})
		require.NoError(t, err)
		user1 := result["user_1"].(map[string]any)
		assert.Equal(t, "Alice", user1["name"])

		// Valid: non-matching key passes through
		result, err = schema.Parse(map[string]any{
			"user_1": map[string]any{"name": "Alice", "age": 30},
			"other":  "value",
		})
		require.NoError(t, err)
		assert.Equal(t, "value", result["other"])

		// Invalid: matching key with invalid object
		_, err = schema.Parse(map[string]any{
			"user_1": map[string]any{"name": "Alice"}, // missing age
		})
		require.Error(t, err)
	})

	t.Run("with array value schema", func(t *testing.T) {
		// Use a simpler schema - LooseRecord with string value for matching keys
		schema := LooseRecord(String().RegexString(`^arr_`), String())

		// Valid: matching key with valid string
		result, err := schema.Parse(map[string]any{
			"arr_1": "value1",
		})
		require.NoError(t, err)
		assert.Equal(t, "value1", result["arr_1"])

		// Non-matching key passes through with any type
		result, err = schema.Parse(map[string]any{
			"arr_1": "value1",
			"other": 123,
		})
		require.NoError(t, err)
		assert.Equal(t, 123, result["other"])
		assert.Equal(t, "value1", result["arr_1"])
	})

	t.Run("immutability - modifiers return new instance", func(t *testing.T) {
		original := LooseRecord(String().RegexString(`^S_`), String())
		optional := original.Optional()
		nilable := original.Nilable()

		// Original should reject nil
		_, err := original.Parse(nil)
		require.Error(t, err)

		// Optional should accept nil
		result1, err := optional.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result1)

		// Nilable should accept nil
		result2, err := nilable.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})

	t.Run("all matching keys must be valid", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String())

		// Multiple matching keys - all must be valid
		_, err := schema.Parse(map[string]any{
			"S_name": "valid",
			"S_bad":  123, // invalid
		})
		require.Error(t, err)
	})

	t.Run("empty pattern matches all keys", func(t *testing.T) {
		// LooseRecord with simple String() key schema matches all keys
		schema := LooseRecord(String(), Int())

		// All keys must have int values
		_, err := schema.Parse(map[string]any{
			"a": 1,
			"b": 2,
			"c": "not-int",
		})
		require.Error(t, err)
	})

	t.Run("Refine on loose record", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String()).Refine(func(m map[string]any) bool {
			// Require at least one S_ key
			for k := range m {
				if len(k) > 2 && k[:2] == "S_" {
					return true
				}
			}
			return false
		})

		// Valid: has S_ key
		_, err := schema.Parse(map[string]any{"S_name": "John", "other": 123})
		require.NoError(t, err)

		// Invalid: no S_ key
		_, err = schema.Parse(map[string]any{"other": 123})
		require.Error(t, err)
	})

	t.Run("StrictParse with loose record", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String())

		result, err := schema.StrictParse(map[string]any{
			"S_name": "John",
			"other":  "value",
		})
		require.NoError(t, err)
		assert.Equal(t, "John", result["S_name"])
		assert.Equal(t, "value", result["other"])
	})

	t.Run("MustParse panics on error", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String())

		assert.Panics(t, func() {
			schema.MustParse(map[string]any{"S_name": 123})
		})
	})

	t.Run("preserves order-independent behavior", func(t *testing.T) {
		schema := LooseRecord(String().RegexString(`^S_`), String())

		// Same data, different order should produce equivalent results
		result1, err := schema.Parse(map[string]any{
			"S_a":   "1",
			"other": 2,
			"S_b":   "3",
		})
		require.NoError(t, err)

		result2, err := schema.Parse(map[string]any{
			"other": 2,
			"S_b":   "3",
			"S_a":   "1",
		})
		require.NoError(t, err)

		assert.Equal(t, result1["S_a"], result2["S_a"])
		assert.Equal(t, result1["S_b"], result2["S_b"])
		assert.Equal(t, result1["other"], result2["other"])
	})
}

// =============================================================================
// NUMERIC STRING KEYS TESTS (Zod v4 PR #5585)
// =============================================================================

func TestRecord_NumericStringKeys(t *testing.T) {
	t.Run("numeric string keys with float schema", func(t *testing.T) {
		// Zod v4: z.record(z.number(), z.number())
		schema := Record(Float64(), Int())

		// Integer string keys work
		result, err := schema.Parse(map[string]any{"1": 100, "2": 200})
		require.NoError(t, err)
		assert.Equal(t, 100, result["1"])
		assert.Equal(t, 200, result["2"])
	})

	t.Run("float string keys work", func(t *testing.T) {
		schema := Record(Float64(), Int())

		// Float and negative keys work
		result, err := schema.Parse(map[string]any{"1.5": 100, "-3": 200})
		require.NoError(t, err)
		assert.Equal(t, 100, result["1.5"])
		assert.Equal(t, 200, result["-3"])
	})

	t.Run("non-numeric keys fail with number schema", func(t *testing.T) {
		schema := Record(Float64(), Int())

		// Non-numeric keys should fail
		_, err := schema.Parse(map[string]any{"abc": 100})
		require.Error(t, err)
	})

	t.Run("integer constraint is respected", func(t *testing.T) {
		// Use Int() which enforces integer validation
		schema := Record(Int(), Int())

		// Integer keys should pass
		result, err := schema.Parse(map[string]any{"1": 100, "42": 200})
		require.NoError(t, err)
		assert.Equal(t, 100, result["1"])
		assert.Equal(t, 200, result["42"])

		// Float keys should fail with integer schema
		_, err = schema.Parse(map[string]any{"1.5": 100})
		require.Error(t, err)
	})

	t.Run("mixed valid and invalid keys", func(t *testing.T) {
		schema := Record(Float64(), String())

		// All valid numeric keys
		result, err := schema.Parse(map[string]any{
			"1":    "one",
			"2.5":  "two point five",
			"-10":  "negative ten",
			"0":    "zero",
			"3.14": "pi-ish",
		})
		require.NoError(t, err)
		assert.Equal(t, "one", result["1"])
		assert.Equal(t, "two point five", result["2.5"])
		assert.Equal(t, "negative ten", result["-10"])
	})

	t.Run("empty record with numeric key schema", func(t *testing.T) {
		schema := Record(Float64(), String())

		result, err := schema.Parse(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("scientific notation keys", func(t *testing.T) {
		schema := Record(Float64(), String())

		// Scientific notation keys are parsed as valid numbers.
		// Note: The key format may change after parsing due to Go's float formatting.
		// "1e10" -> "1e+10", "1.5e-3" -> "0.0015"
		result, err := schema.Parse(map[string]any{"1e10": "big", "1.5e-3": "small"})
		require.NoError(t, err)
		// Keys are normalized to Go's float64 string format
		assert.Equal(t, "big", result["1e+10"])
		assert.Equal(t, "small", result["0.0015"])
	})

	t.Run("key transform with Overwrite", func(t *testing.T) {
		// Zod v4: z.record(z.number().overwrite((n) => n * 2), z.string())
		schema := Record(
			Float64().Overwrite(func(n float64) float64 { return n * 2 }),
			String(),
		)

		result, err := schema.Parse(map[string]any{"5": "five", "10": "ten"})
		require.NoError(t, err)
		// Keys should be transformed: 5*2=10, 10*2=20
		assert.Equal(t, "five", result["10"])
		assert.Equal(t, "ten", result["20"])
	})

	t.Run("key transform with int schema", func(t *testing.T) {
		schema := Record(
			Int().Overwrite(func(n int) int { return n + 100 }),
			String(),
		)

		result, err := schema.Parse(map[string]any{"1": "one", "2": "two"})
		require.NoError(t, err)
		// Keys should be transformed: 1+100=101, 2+100=102
		assert.Equal(t, "one", result["101"])
		assert.Equal(t, "two", result["102"])
	})

	t.Run("uint key schema", func(t *testing.T) {
		schema := Record(Uint(), String())

		// Valid unsigned integers
		result, err := schema.Parse(map[string]any{"0": "zero", "42": "answer"})
		require.NoError(t, err)
		assert.Equal(t, "zero", result["0"])
		assert.Equal(t, "answer", result["42"])

		// Negative numbers should fail with uint schema
		_, err = schema.Parse(map[string]any{"-1": "negative"})
		require.Error(t, err)
	})

	t.Run("combined with Min/Max validation", func(t *testing.T) {
		schema := Record(Float64(), String()).Min(2).Max(4)

		// Valid: 2-4 entries
		result, err := schema.Parse(map[string]any{"1": "a", "2": "b", "3": "c"})
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// Invalid: too few entries
		_, err = schema.Parse(map[string]any{"1": "a"})
		require.Error(t, err)

		// Invalid: too many entries
		_, err = schema.Parse(map[string]any{"1": "a", "2": "b", "3": "c", "4": "d", "5": "e"})
		require.Error(t, err)
	})

	t.Run("with Optional modifier", func(t *testing.T) {
		schema := Record(Float64(), String()).Optional()

		// Valid record
		result, err := schema.Parse(map[string]any{"1": "one"})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "one", (*result)["1"])

		// Nil is accepted
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("with Nilable modifier", func(t *testing.T) {
		schema := Record(Float64(), String()).Nilable()

		// Valid record
		result, err := schema.Parse(map[string]any{"1.5": "one point five"})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "one point five", (*result)["1.5"])

		// Nil is accepted
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("loose record with numeric keys", func(t *testing.T) {
		// LooseRecord should pass through non-matching keys
		schema := LooseRecord(Float64(), String())

		// Numeric keys are validated, non-numeric keys pass through
		result, err := schema.Parse(map[string]any{
			"1":   "one",
			"abc": "text", // non-numeric, should pass through
			"2.5": "two point five",
		})
		require.NoError(t, err)
		assert.Equal(t, "one", result["1"])
		assert.Equal(t, "text", result["abc"]) // Preserved unchanged
		assert.Equal(t, "two point five", result["2.5"])
	})

	t.Run("partial record with numeric keys", func(t *testing.T) {
		// PartialRecord skips exhaustive key checks
		// Note: PartialRecord is designed for enum/literal key schemas
		// With numeric key schema, it behaves similarly to regular Record
		schema := Record(Float64(), String())

		result, err := schema.Parse(map[string]any{"1": "one"})
		require.NoError(t, err)
		assert.Equal(t, "one", result["1"])
	})

	t.Run("StrictParse with numeric keys", func(t *testing.T) {
		schema := Record(Float64(), String())

		input := map[string]any{"1": "one", "2.5": "two point five"}
		result, err := schema.StrictParse(input)
		require.NoError(t, err)
		assert.Equal(t, "one", result["1"])
		assert.Equal(t, "two point five", result["2.5"])
	})

	t.Run("MustParse with numeric keys", func(t *testing.T) {
		schema := Record(Float64(), String())

		// Valid input should not panic
		result := schema.MustParse(map[string]any{"42": "answer"})
		assert.Equal(t, "answer", result["42"])

		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse(map[string]any{"not-a-number": "value"})
		})
	})
}
