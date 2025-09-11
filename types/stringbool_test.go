package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality tests
// =============================================================================

func TestStringBool_BasicFunctionality(t *testing.T) {
	t.Run("valid string inputs with default options", func(t *testing.T) {
		schema := StringBool()

		// Test truthy values (case insensitive by default)
		truthyValues := []string{"true", "1", "yes", "on", "y", "enabled", "TRUE", "Yes", "ON"}
		for _, value := range truthyValues {
			result, err := schema.Parse(value)
			require.NoError(t, err, "Should parse truthy value: %s", value)
			assert.Equal(t, true, result)
		}

		// Test falsy values (case insensitive by default)
		falsyValues := []string{"false", "0", "no", "off", "n", "disabled", "FALSE", "No", "OFF"}
		for _, value := range falsyValues {
			result, err := schema.Parse(value)
			require.NoError(t, err, "Should parse falsy value: %s", value)
			assert.Equal(t, false, result)
		}
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := StringBool()

		invalidInputs := []any{
			123, 3.14, true, false, []string{"true"}, nil, struct{}{},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("invalid string values", func(t *testing.T) {
		schema := StringBool()

		invalidStrings := []string{"maybe", "unknown", "", "2", "invalid"}
		for _, value := range invalidStrings {
			_, err := schema.Parse(value)
			assert.Error(t, err, "Expected error for invalid string: %s", value)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := StringBool()

		// Test Parse method
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Test MustParse method
		mustResult := schema.MustParse("false")
		assert.Equal(t, false, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid string boolean"
		schema := StringBool(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeStringBool, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// 2. Type safety tests
// =============================================================================

func TestStringBool_TypeSafety(t *testing.T) {
	t.Run("StringBool returns bool type", func(t *testing.T) {
		schema := StringBool()
		require.NotNil(t, schema)

		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.IsType(t, bool(false), result) // Ensure type is bool

		result, err = schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)
		assert.IsType(t, bool(false), result)
	})

	t.Run("StringBoolPtr returns *bool type", func(t *testing.T) {
		schema := StringBoolPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
		assert.IsType(t, (*bool)(nil), result) // Ensure type is *bool

		result, err = schema.Parse("false")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		boolSchema := StringBool()   // bool type
		ptrSchema := StringBoolPtr() // *bool type

		// Test bool type
		result1, err1 := boolSchema.Parse("true")
		require.NoError(t, err1)
		assert.IsType(t, bool(false), result1)
		assert.Equal(t, true, result1)

		// Test *bool type
		result2, err2 := ptrSchema.Parse("true")
		require.NoError(t, err2)
		assert.IsType(t, (*bool)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, true, *result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test bool type
		boolSchema := StringBool()
		result := boolSchema.MustParse("true")
		assert.IsType(t, bool(false), result)
		assert.Equal(t, true, result)

		// Test *bool type
		ptrSchema := StringBoolPtr()
		ptrResult := ptrSchema.MustParse("true")
		assert.IsType(t, (*bool)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, true, *ptrResult)
	})
}

// =============================================================================
// 3. Modifier methods tests
// =============================================================================

func TestStringBool_Modifiers(t *testing.T) {
	t.Run("Optional always returns *bool", func(t *testing.T) {
		// From bool to *bool via Optional
		boolSchema := StringBool()
		optionalSchema := boolSchema.Optional()

		// Type check: ensure it returns *ZodStringBool[*bool]
		var _ = optionalSchema

		// Functionality test
		result, err := optionalSchema.Parse("true")
		require.NoError(t, err)
		assert.IsType(t, (*bool)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		// From *bool to *bool via Optional (maintains type)
		ptrSchema := StringBoolPtr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ = optionalPtrSchema
	})

	t.Run("Nilable always returns *bool", func(t *testing.T) {
		boolSchema := StringBool()
		nilableSchema := boolSchema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		// bool maintains bool
		boolSchema := StringBool()
		defaultBoolSchema := boolSchema.Default(true)
		var _ = defaultBoolSchema

		// *bool maintains *bool
		ptrSchema := StringBoolPtr()
		defaultPtrSchema := ptrSchema.Default(false)
		var _ = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		// bool maintains bool
		boolSchema := StringBool()
		prefaultBoolSchema := boolSchema.Prefault("true")
		var _ = prefaultBoolSchema

		// *bool maintains *bool
		ptrSchema := StringBoolPtr()
		prefaultPtrSchema := ptrSchema.Prefault("false")
		var _ = prefaultPtrSchema
	})
}

// =============================================================================
// 4. Chaining tests
// =============================================================================

func TestStringBool_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		schema := StringBool(). // *ZodStringBool[bool]
					Default(false). // *ZodStringBool[bool] (maintains type)
					Optional()      // *ZodStringBool[*bool] (type conversion)

		var _ = schema

		// Test final behavior
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.IsType(t, (*bool)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := StringBoolPtr(). // *ZodStringBool[*bool]
						Nilable().    // *ZodStringBool[*bool] (maintains type)
						Default(true) // *ZodStringBool[*bool] (maintains type)

		var _ = schema

		result, err := schema.Parse("false")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := StringBool().
			Default(true).
			Prefault("false")

		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})
}

// =============================================================================
// 5. Default and prefault tests
// =============================================================================

func TestStringBool_DefaultAndPrefault(t *testing.T) {
	t.Run("default value behavior", func(t *testing.T) {
		schema := StringBool().Default(true)

		// Valid input should override default
		result, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Nil input should use default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("default function behavior", func(t *testing.T) {
		called := false
		schema := StringBool().DefaultFunc(func() bool {
			called = true
			return true
		})

		// Valid input should not call function
		result, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)
		assert.False(t, called)

		// Nil input should call function
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.True(t, called)
	})

	// Test Default priority over Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := StringBool().Default(true).Prefault("false")

		// Nil input should use Default (higher priority), not Prefault
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result) // Default value, not Prefault value
	})

	// Test Default short-circuit bypasses validation
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		// Create a schema with custom validation that would reject the default value
		options := &StringBoolOptions{
			Truthy: []string{"yes"},
			Falsy:  []string{"no"},
			Case:   "sensitive",
		}
		schema := StringBool(options).Default(true) // Default bypasses validation

		// Nil input should use default and bypass validation
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Valid input should still go through validation
		result, err = schema.Parse("yes")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid input should fail validation
		_, err = schema.Parse("true")
		require.Error(t, err)
	})

	t.Run("prefault value behavior", func(t *testing.T) {
		schema := StringBool().Prefault("false")

		// Valid input should work normally
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid input should return error (NOT use prefault)
		_, err = schema.Parse("invalid")
		require.Error(t, err)

		// Nil input should use prefault
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	// Test Prefault goes through full validation
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Create a schema with custom validation that would reject the prefault value
		options := &StringBoolOptions{
			Truthy: []string{"yes"},
			Falsy:  []string{"no"},
			Case:   "sensitive",
		}
		schema := StringBool(options).Prefault("false") // This should fail validation

		// Nil input should trigger prefault, but prefault should fail validation
		_, err := schema.Parse(nil)
		require.Error(t, err) // Prefault value fails validation
	})

	t.Run("prefault function behavior", func(t *testing.T) {
		called := false
		schema := StringBool().PrefaultFunc(func() string {
			called = true
			return "false"
		})

		// Valid input should not call function
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.False(t, called)

		// Invalid input should return error (NOT call function)
		_, err = schema.Parse("invalid")
		require.Error(t, err)
		assert.False(t, called)

		// Nil input should call function
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
		assert.True(t, called)
	})

	// Test Prefault error handling
	t.Run("Prefault error handling", func(t *testing.T) {
		schema := StringBool().PrefaultFunc(func() string {
			// This will create a string value that should pass StringBool validation
			return "true"
		})

		// Nil input should use prefault function
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	// Test StringBoolPtr behavior
	t.Run("StringBoolPtr Default and Prefault", func(t *testing.T) {
		schema := StringBoolPtr().Default(true)

		// Nil input should use default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		// Test Prefault with pointer type
		prefaultSchema := StringBoolPtr().Prefault("false")
		result2, err2 := prefaultSchema.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, false, *result2)
	})
}

// =============================================================================
// 6. Refine tests
// =============================================================================

func TestStringBool_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept true values
		schema := StringBool().Refine(func(b bool) bool {
			return b == true
		})

		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be true"
		schema := StringBoolPtr().Refine(func(b *bool) bool {
			return b != nil && *b == true
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("refine pointer allows nil", func(t *testing.T) {
		schema := StringBoolPtr().Nilable().Refine(func(b *bool) bool {
			// Accept nil or true
			return b == nil || (b != nil && *b)
		})

		// Expect nil to be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// true should pass
		result, err = schema.Parse("true")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, true, *result)

		// false should fail (refine returns false)
		_, err = schema.Parse("false")
		assert.Error(t, err)
	})
}

func TestStringBool_RefineAny(t *testing.T) {
	t.Run("refineAny bool schema", func(t *testing.T) {
		// Only accept true values via RefineAny on StringBool() schema
		schema := StringBool().RefineAny(func(v any) bool {
			b, ok := v.(bool)
			return ok && b
		})

		// true passes
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// false fails
		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("refineAny pointer schema", func(t *testing.T) {
		// StringBoolPtr().RefineAny sees underlying bool value
		schema := StringBoolPtr().RefineAny(func(v any) bool {
			b, ok := v.(bool)
			return ok && b // accept only true
		})

		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, true, *result)

		_, err = schema.Parse("false")
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Coercion tests
// =============================================================================

func TestStringBool_Coercion(t *testing.T) {
	t.Run("string coercion with coerced schema", func(t *testing.T) {
		schema := CoercedStringBool()

		// Test string "true" -> bool true
		result, err := schema.Parse("true")
		require.NoError(t, err, "Should coerce string 'true' to bool true")
		assert.Equal(t, true, result)

		// Test string "false" -> bool false
		result, err = schema.Parse("false")
		require.NoError(t, err, "Should coerce string 'false' to bool false")
		assert.Equal(t, false, result)
	})

	t.Run("numeric coercion to string then bool", func(t *testing.T) {
		schema := CoercedStringBool()

		// Numbers should be coerced to strings first, then to bool
		// Since default truthy/falsy lists include "1" and "0", this should work
		result, err := schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// But numbers that don't map to truthy/falsy should fail
		_, err = schema.Parse(2)
		assert.Error(t, err, "Numeric input 2 should fail as it doesn't map to valid bool string")
	})

	t.Run("custom options with numeric string support", func(t *testing.T) {
		options := &StringBoolOptions{
			Truthy: []string{"true", "1", "yes"},
			Falsy:  []string{"false", "0", "no"},
			Case:   "insensitive",
		}
		schema := CoercedStringBool(options)

		// Now numeric inputs that coerce to "1" or "0" should work
		result, err := schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("invalid coercion inputs", func(t *testing.T) {
		schema := CoercedStringBool()

		// Inputs that cannot be coerced to valid string bools
		invalidInputs := []any{
			[]string{"true"}, struct{}{}, map[string]any{"key": "value"},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})
}

// =============================================================================
// 8. Type-specific methods tests (StringBool Configuration)
// =============================================================================

func TestStringBool_TypeSpecificMethods(t *testing.T) {
	t.Run("custom truthy and falsy values", func(t *testing.T) {
		options := &StringBoolOptions{
			Truthy: []string{"yes", "ok", "positive"},
			Falsy:  []string{"no", "cancel", "negative"},
			Case:   "sensitive",
		}
		schema := StringBool(options)

		// Test custom truthy values
		result, err := schema.Parse("yes")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse("positive")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Test custom falsy values
		result, err = schema.Parse("no")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		result, err = schema.Parse("negative")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Test that default values no longer work
		_, err = schema.Parse("true")
		assert.Error(t, err)

		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("case sensitive mode", func(t *testing.T) {
		options := &StringBoolOptions{
			Truthy: []string{"True", "YES"},
			Falsy:  []string{"False", "NO"},
			Case:   "sensitive",
		}
		schema := StringBool(options)

		// Exact case should work
		result, err := schema.Parse("True")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse("False")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Different case should fail
		_, err = schema.Parse("true")
		assert.Error(t, err)

		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("case insensitive mode", func(t *testing.T) {
		options := &StringBoolOptions{
			Truthy: []string{"True", "YES"},
			Falsy:  []string{"False", "NO"},
			Case:   "insensitive",
		}
		schema := StringBool(options)

		// Different cases should all work
		testCases := []struct {
			input    string
			expected bool
		}{
			{"True", true}, {"true", true}, {"TRUE", true},
			{"YES", true}, {"yes", true}, {"Yes", true},
			{"False", false}, {"false", false}, {"FALSE", false},
			{"NO", false}, {"no", false}, {"No", false},
		}

		for _, tc := range testCases {
			result, err := schema.Parse(tc.input)
			require.NoError(t, err, "Should parse input: %s", tc.input)
			assert.Equal(t, tc.expected, result)
		}
	})
}

// =============================================================================
// 9. Error handling and edge case tests
// =============================================================================

func TestStringBool_ErrorHandling(t *testing.T) {
	t.Run("custom error messages", func(t *testing.T) {
		customError := "Expected a valid boolean string"
		schema := StringBool(core.SchemaParams{Error: customError})

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
		// Note: Specific error message checking would depend on error implementation
	})

	t.Run("nil handling with different modifiers", func(t *testing.T) {
		// Regular schema rejects nil
		schema := StringBool()
		_, err := schema.Parse(nil)
		assert.Error(t, err)

		// Optional schema accepts nil
		optionalSchema := StringBool().Optional()
		optionalResult, err := optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, optionalResult)

		// Nilable schema accepts nil
		nilableSchema := StringBool().Nilable()
		nilableResult, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, nilableResult)
	})

	t.Run("empty string handling", func(t *testing.T) {
		// Empty string should be rejected by default
		schema := StringBool()
		_, err := schema.Parse("")
		assert.Error(t, err)

		// But can be included in falsy values
		options := &StringBoolOptions{
			Truthy: []string{"true", "1"},
			Falsy:  []string{"false", "0", ""},
			Case:   "insensitive",
		}
		schemaWithEmpty := StringBool(options)
		result, err := schemaWithEmpty.Parse("")
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("pointer string input", func(t *testing.T) {
		schema := StringBool()

		// Test *string input
		str := "true"
		result, err := schema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Test nil *string input
		var nilStr *string = nil
		_, err = schema.Parse(nilStr)
		assert.Error(t, err)

		// Test nil *string with nilable schema
		nilableSchema := StringBool().Nilable()
		nilableResult, err := nilableSchema.Parse(nilStr)
		require.NoError(t, err)
		assert.Nil(t, nilableResult)
	})
}

func TestStringBool_EdgeCases(t *testing.T) {
	t.Run("empty truthy and falsy lists", func(t *testing.T) {
		options := &StringBoolOptions{
			Truthy: []string{},
			Falsy:  []string{},
			Case:   "insensitive",
		}
		schema := StringBool(options)

		// Any string input should fail
		_, err := schema.Parse("anything")
		assert.Error(t, err)
	})

	t.Run("overlapping truthy and falsy values", func(t *testing.T) {
		// If a value appears in both lists, truthy takes precedence
		options := &StringBoolOptions{
			Truthy: []string{"maybe", "true"},
			Falsy:  []string{"maybe", "false"},
			Case:   "insensitive",
		}
		schema := StringBool(options)

		// "maybe" should be true (truthy checked first)
		result, err := schema.Parse("maybe")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("whitespace handling", func(t *testing.T) {
		schema := StringBool()

		// Whitespace should not be automatically trimmed
		_, err := schema.Parse(" true ")
		assert.Error(t, err)

		// Unless explicitly included in the lists
		options := &StringBoolOptions{
			Truthy: []string{"true", " true ", "  yes  "},
			Falsy:  []string{"false", " false ", "  no  "},
			Case:   "insensitive",
		}
		schemaWithWhitespace := StringBool(options)
		result, err := schemaWithWhitespace.Parse(" true ")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})
}

// =============================================================================
// Pointer identity preservation tests
// =============================================================================

func TestStringBool_PointerIdentityPreservation(t *testing.T) {
	t.Run("StringBool Optional preserves pointer identity with string input", func(t *testing.T) {
		schema := StringBool().Optional()

		originalString := "true"
		originalPtr := &originalString

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be a pointer to bool (converted from string)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("StringBool Nilable preserves pointer identity with string input", func(t *testing.T) {
		schema := StringBool().Nilable()

		originalString := "false"
		originalPtr := &originalString

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be a pointer to bool (converted from string)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("StringBoolPtr Optional preserves pointer identity with string input", func(t *testing.T) {
		schema := StringBoolPtr().Optional()

		originalString := "true"
		originalPtr := &originalString

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be a pointer to bool (converted from string)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("StringBoolPtr Nilable preserves pointer identity with string input", func(t *testing.T) {
		schema := StringBoolPtr().Nilable()

		originalString := "false"
		originalPtr := &originalString

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be a pointer to bool (converted from string)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("Optional handles nil consistently", func(t *testing.T) {
		schema := StringBool().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable handles nil consistently", func(t *testing.T) {
		schema := StringBool().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default().Optional() chaining preserves pointer identity with string input", func(t *testing.T) {
		schema := StringBool().Default(false).Optional()

		originalString := "true"
		originalPtr := &originalString

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be a pointer to bool (converted from string)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("Pointer identity with string inputs", func(t *testing.T) {
		testCases := []struct {
			name        string
			stringInput string
			expected    bool
		}{
			{"true", "true", true},
			{"false", "false", false},
			{"1", "1", true},
			{"0", "0", false},
			{"yes", "yes", true},
			{"no", "no", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema := StringBool().Optional()

				// Test with string input - this will convert from string to bool
				originalPtr := &tc.stringInput

				result, err := schema.Parse(originalPtr)
				require.NoError(t, err)

				// This is a conversion, so we just check the value
				assert.Equal(t, tc.expected, *result)
			})
		}
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestStringBool_Overwrite(t *testing.T) {
	t.Run("basic boolean transformation", func(t *testing.T) {
		schema := StringBool().Overwrite(func(b bool) bool {
			return !b // Invert boolean value
		})

		// Test true -> false
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Test false -> true
		result, err = schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Test with custom truthy value
		result, err = schema.Parse("yes")
		require.NoError(t, err)
		assert.Equal(t, false, result) // "yes" -> true -> inverted to false
	})

	t.Run("conditional transformation", func(t *testing.T) {
		schema := StringBool().Overwrite(func(b bool) bool {
			// Convert false to true, keep true as true (always true)
			return true
		})

		result, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := StringBoolPtr().Overwrite(func(b *bool) *bool {
			if b == nil {
				falseVal := false
				return &falseVal
			}
			inverted := !(*b)
			return &inverted
		})

		// Test normal case
		result, err := schema.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)

		// Test nil case through optional
		schema = schema.Optional()
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("chaining with other validations", func(t *testing.T) {
		schema := StringBool().
			Overwrite(func(b bool) bool {
				return !b // Invert first
			}).
			Refine(func(b bool) bool {
				return b == false // Only allow false values (which were originally true)
			}, "Must be originally true")

		// Should pass: "true" -> inverted to false -> passes refine
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Should fail: "false" -> inverted to true -> fails refine
		_, err = schema.Parse("false")
		assert.Error(t, err)
	})

	t.Run("multiple transformations", func(t *testing.T) {
		schema := StringBool().
			Overwrite(func(b bool) bool {
				return !b // First inversion
			}).
			Overwrite(func(b bool) bool {
				return !b // Second inversion (back to original)
			})

		// Should be back to original value
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := StringBool().Overwrite(func(b bool) bool {
			return b // Identity transformation
		})

		result, err := schema.Parse("yes")
		require.NoError(t, err)
		assert.IsType(t, true, result)
		assert.Equal(t, true, result)
	})

	t.Run("custom truthy/falsy handling", func(t *testing.T) {
		// Create StringBool with custom truthy/falsy values
		schema := StringBool(&StringBoolOptions{
			Truthy: []string{"on", "enabled", "active"},
			Falsy:  []string{"off", "disabled", "inactive"},
			Case:   "insensitive",
		}).Overwrite(func(b bool) bool {
			return !b // Invert the parsed boolean
		})

		// Test custom truthy -> false
		result, err := schema.Parse("enabled")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Test custom falsy -> true
		result, err = schema.Parse("inactive")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("error handling preservation", func(t *testing.T) {
		schema := StringBool().Overwrite(func(b bool) bool {
			return !b
		})

		// Invalid input should still fail validation
		_, err := schema.Parse("invalid")
		assert.Error(t, err)

		// Non-string input should still fail
		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("default value interaction", func(t *testing.T) {
		schema := StringBool().
			Default(true).
			Overwrite(func(b bool) bool {
				return !b // Invert boolean
			})

		// Test with actual input
		result, err := schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, true, result) // false -> inverted to true

		// Test nil input uses default -> transformed
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, false, result) // default true -> inverted to false
	})
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestStringBool_StrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := StringBool()

		// Test StrictParse with truthy bool values
		truthyValues := []bool{true}
		for _, value := range truthyValues {
			result, err := schema.StrictParse(value)
			require.NoError(t, err, "Should parse truthy value: %v", value)
			assert.Equal(t, true, result)
			assert.IsType(t, true, result)
		}

		// Test StrictParse with falsy bool values
		falsyValues := []bool{false}
		for _, value := range falsyValues {
			result, err := schema.StrictParse(value)
			require.NoError(t, err, "Should parse falsy value: %v", value)
			assert.Equal(t, false, result)
			assert.IsType(t, false, result)
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		schema := StringBool()

		// Test bool values (no case sensitivity for bools)
		boolValues := []bool{true, false}
		for _, value := range boolValues {
			_, err := schema.StrictParse(value)
			require.NoError(t, err, "Should parse bool value: %v", value)
		}
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := StringBool().Refine(func(b bool) bool {
			return b == true // Only allow true values
		}, "Must be true")

		// Valid case
		result, err := schema.StrictParse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid case - false not allowed
		_, err = schema.StrictParse(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Must be true")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := StringBoolPtr()
		boolVal := true

		// Test with valid pointer input
		result, err := schema.StrictParse(&boolVal)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
		assert.IsType(t, (*bool)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := StringBoolPtr().Default(true)
		var nilPtr *bool = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		schema := StringBoolPtr().Refine(func(b *bool) bool {
			return b != nil && *b == true // Only allow true values
		}, "Must be true").Prefault("true") // Use true as prefault to pass validation
		falseBoolVal := false

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema.StrictParse(&falseBoolVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Must be true")

		// Test with nil input - StrictParse expects boolean input type, not string prefault
		// So this should fail with type error since nil is not a valid boolean
		_, err2 := schema.StrictParse(nil)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "expected boolean")
	})

	t.Run("invalid type values", func(t *testing.T) {
		schema := StringBool()

		// Test with invalid type values - these should cause type errors
		// Note: In strict parsing, we expect bool type, so other types should error
		// This test demonstrates type safety of StrictParse
		_, err := schema.Parse("invalid") // Use Parse instead for string input
		assert.Error(t, err, "Should error for invalid string value")
	})
}

func TestStringBool_MustStrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := StringBool()

		// Test MustStrictParse with truthy values
		result := schema.MustStrictParse(true)
		assert.Equal(t, true, result)
		assert.IsType(t, true, result)

		// Test MustStrictParse with falsy values
		falseResult := schema.MustStrictParse(false)
		assert.Equal(t, false, falseResult)
		assert.IsType(t, false, falseResult)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := StringBool().Refine(func(b bool) bool {
			return b == true // Only allow true values
		}, "Must be true")

		// Test panic with validation failure
		assert.Panics(t, func() {
			schema.MustStrictParse(false) // Should panic
		})

		// Test panic with invalid type through Parse method
		schemaInvalid := StringBool()
		assert.Panics(t, func() {
			schemaInvalid.MustParse("invalid") // Should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := StringBoolPtr()
		boolVal := true

		// Test with valid pointer input
		result := schema.MustStrictParse(&boolVal)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
		assert.IsType(t, (*bool)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := StringBoolPtr().Default(false)
		var nilPtr *bool = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("bool value parsing", func(t *testing.T) {
		schema := StringBool()

		// Test bool values
		assert.Equal(t, true, schema.MustStrictParse(true))
		assert.Equal(t, false, schema.MustStrictParse(false))
	})
}
