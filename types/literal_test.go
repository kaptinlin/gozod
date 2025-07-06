package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestLiteral_BasicFunctionality(t *testing.T) {
	t.Run("single value literal", func(t *testing.T) {
		schema := Literal("hello")
		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeLiteral, schema.GetInternals().Type)

		// Valid input
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid input
		_, err = schema.Parse("world")
		assert.Error(t, err)
	})

	t.Run("multiple values literal", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green", "blue"})
		require.NotNil(t, schema)

		// Valid inputs
		validValues := []string{"red", "green", "blue"}
		for _, value := range validValues {
			result, err := schema.Parse(value)
			require.NoError(t, err)
			assert.Equal(t, value, result)
		}

		// Invalid inputs
		invalidValues := []string{"yellow", "purple", "orange"}
		for _, value := range invalidValues {
			_, err := schema.Parse(value)
			assert.Error(t, err)
		}
	})

	t.Run("new generic API with type inference", func(t *testing.T) {
		// String literal
		stringSchema := Literal("active")
		result, err := stringSchema.Parse("active")
		require.NoError(t, err)
		assert.Equal(t, "active", result)

		_, err = stringSchema.Parse("pending")
		assert.Error(t, err)

		// Int literal
		intSchema := Literal(42)
		result2, err := intSchema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)

		_, err = intSchema.Parse(43)
		assert.Error(t, err)

		// Bool literal
		boolSchema := Literal(true)
		result3, err := boolSchema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result3)

		_, err = boolSchema.Parse(false)
		assert.Error(t, err)

		// Float literal
		floatSchema := Literal(3.14)
		result4, err := floatSchema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result4)
	})

	t.Run("pointer constraint literals", func(t *testing.T) {
		// Single value pointer constraint
		ptrSchema := LiteralPtr("hello")
		var _ *ZodLiteral[string, *string] = ptrSchema

		str := "hello"
		result, err := ptrSchema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, &str, result)
		assert.True(t, result == &str) // Pointer identity preserved

		// Multiple values pointer constraint
		ptrMultiSchema := LiteralPtrOf([]string{"red", "green", "blue"})
		var _ *ZodLiteral[string, *string] = ptrMultiSchema

		color := "red"
		result2, err := ptrMultiSchema.Parse(&color)
		require.NoError(t, err)
		assert.Equal(t, &color, result2)
		assert.True(t, result2 == &color) // Pointer identity preserved
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Literal(42)

		// Test Parse method
		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)

		// Test MustParse method
		mustResult := schema.MustParse(42)
		assert.Equal(t, 42, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse(43)
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a specific value"
		schema := Literal("valid", core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeLiteral, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestLiteral_TypeSafety(t *testing.T) {
	t.Run("string literal maintains type safety", func(t *testing.T) {
		schema := LiteralOf([]string{"hello", "world"})
		require.NotNil(t, schema)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result) // Ensure type is string

		result, err = schema.Parse("world")
		require.NoError(t, err)
		assert.Equal(t, "world", result)
		assert.IsType(t, "", result)
	})

	t.Run("int literal maintains type safety", func(t *testing.T) {
		schema := LiteralOf([]int{1, 2, 3})
		require.NotNil(t, schema)

		result, err := schema.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)
		assert.IsType(t, 0, result) // Ensure type is int
	})

	t.Run("generic type constraint enforcement", func(t *testing.T) {
		// Custom comparable type
		type Status string
		const (
			StatusActive   Status = "active"
			StatusInactive Status = "inactive"
		)

		schema := LiteralOf([]Status{StatusActive, StatusInactive})
		require.NotNil(t, schema)

		result, err := schema.Parse(StatusActive)
		require.NoError(t, err)
		assert.Equal(t, StatusActive, result)
		assert.IsType(t, Status(""), result)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := LiteralOf([]string{"test"})
		result := schema.MustParse("test")
		assert.IsType(t, "", result)
		assert.Equal(t, "test", result)
	})

	t.Run("type constraint differentiation", func(t *testing.T) {
		// Value constraint type
		valueSchema := LiteralOf([]string{"hello"}) // *ZodLiteral[string, string]
		var _ *ZodLiteral[string, string] = valueSchema

		// Pointer constraint type
		ptrSchema := valueSchema.Optional() // *ZodLiteral[string, *string]
		var _ *ZodLiteral[string, *string] = ptrSchema

		// Verify different return types
		val := valueSchema.MustParse("hello") // string
		var _ string = val

		ptr := ptrSchema.MustParse(&[]string{"hello"}[0]) // *string
		var _ *string = ptr
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestLiteral_Modifiers(t *testing.T) {
	t.Run("Optional returns pointer constraint type", func(t *testing.T) {
		schema := LiteralOf([]string{"hello"})
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodLiteral[string, *string]
		var _ *ZodLiteral[string, *string] = optionalSchema

		// Valid literal value
		str := "hello"
		result, err := optionalSchema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, &str, result)
		assert.True(t, result == &str) // Pointer identity preserved

		// Invalid literal value should still fail
		invalidStr := "world"
		_, err = optionalSchema.Parse(&invalidStr)
		assert.Error(t, err)
	})

	t.Run("Nilable returns pointer constraint type", func(t *testing.T) {
		schema := LiteralOf([]int{42})
		nilableSchema := schema.Nilable()

		var _ *ZodLiteral[int, *int] = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value
		val := 42
		result, err = nilableSchema.Parse(&val)
		require.NoError(t, err)
		assert.Equal(t, &val, result)
		assert.True(t, result == &val) // Pointer identity preserved
	})

	t.Run("Default preserves current constraint type", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green", "blue"})
		defaultSchema := schema.Default("red")
		var _ *ZodLiteral[string, string] = defaultSchema

		// Valid input should override default
		result, err := defaultSchema.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
	})

	t.Run("Prefault preserves current constraint type", func(t *testing.T) {
		schema := LiteralOf([]int{1, 2, 3})
		prefaultSchema := schema.Prefault(1)
		var _ *ZodLiteral[int, int] = prefaultSchema

		// Valid input should override prefault
		result, err := prefaultSchema.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)
	})

	t.Run("Nullish returns pointer constraint type", func(t *testing.T) {
		schema := LiteralOf([]string{"test"})
		nullishSchema := schema.Nullish()
		var _ *ZodLiteral[string, *string] = nullishSchema

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid input with pointer identity
		str := "test"
		result, err = nullishSchema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, &str, result)
		assert.True(t, result == &str) // Pointer identity preserved
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestLiteral_Chaining(t *testing.T) {
	t.Run("modifier chaining with type transitions", func(t *testing.T) {
		// Start with value constraint
		schema := LiteralOf([]string{"active", "inactive"}) // *ZodLiteral[string, string]

		// Chain to Default (preserves constraint) then Optional (changes to pointer constraint)
		chainedSchema := schema.Default("active").Optional() // *ZodLiteral[string, *string]
		var _ *ZodLiteral[string, *string] = chainedSchema

		// Test final behavior
		str := "inactive"
		result, err := chainedSchema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, &str, result)
		assert.True(t, result == &str) // Pointer identity preserved
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := LiteralOf([]bool{true, false}).
			Nilable().
			Default(true)

		var _ *ZodLiteral[bool, *bool] = schema

		// Test with value
		val := false
		result, err := schema.Parse(&val)
		require.NoError(t, err)
		assert.Equal(t, &val, result)
		assert.True(t, result == &val) // Pointer identity preserved
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green", "blue"}).
			Default("red").
			Prefault("green")

		result, err := schema.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestLiteral_DefaultAndPrefault(t *testing.T) {
	t.Run("default value with string literal", func(t *testing.T) {
		schema := LiteralOf([]string{"hello", "world"}).Default("hello")

		// Valid input should override default
		result, err := schema.Parse("world")
		require.NoError(t, err)
		assert.Equal(t, "world", result)
	})

	t.Run("prefault value", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green"}).Prefault("red")

		// Valid input should override prefault
		result, err := schema.Parse("green")
		require.NoError(t, err)
		assert.Equal(t, "green", result)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := LiteralOf([]string{"a", "b"}).DefaultFunc(func() string {
			counter++
			if counter%2 == 1 {
				return "a"
			}
			return "b"
		})

		// Valid input should not call function
		result, err := schema.Parse("a")
		require.NoError(t, err)
		assert.Equal(t, "a", result)
		assert.Equal(t, 0, counter) // Function should not be called
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := LiteralOf([]string{"valid"}).PrefaultFunc(func() string {
			counter++
			return "fallback"
		})

		// Valid input should not call function
		result, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", result)
		assert.Equal(t, 0, counter)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestLiteral_Refine(t *testing.T) {
	t.Run("refine with type-safe function", func(t *testing.T) {
		// Only allow "red" even though "green" and "blue" are valid literals
		schema := LiteralOf([]string{"red", "green", "blue"}).Refine(func(s string) bool {
			return s == "red"
		})

		result, err := schema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		_, err = schema.Parse("green") // Valid literal but fails refinement
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be positive"
		schema := LiteralOf([]int{1, 2, 3, -1}).Refine(func(n int) bool {
			return n > 0
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)

		_, err = schema.Parse(-1)
		assert.Error(t, err)
	})

	t.Run("refine with nil handling", func(t *testing.T) {
		schema := LiteralOf([]string{"hello", "world"}).Nilable().Refine(func(s string) bool {
			return s != "" // Accept non-empty strings
		})

		// Valid string should pass
		str := "hello"
		result, err := schema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, &str, result)
		assert.True(t, result == &str) // Pointer identity preserved

		// nil should pass
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestLiteral_RefineAny(t *testing.T) {
	t.Run("refineAny with flexible validation", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green", "blue"}).RefineAny(func(v any) bool {
			if s, ok := v.(string); ok {
				return len(s) >= 3 // Only allow values with 3+ characters
			}
			return false
		})

		// "red" and "green" and "blue" all have 3+ characters
		result, err := schema.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		result, err = schema.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
	})

	t.Run("refineAny with nil handling", func(t *testing.T) {
		schema := LiteralOf([]int{1, 2, 3}).Nilable().RefineAny(func(v any) bool {
			if v == nil {
				return true // Allow nil
			}
			if i, ok := v.(int); ok {
				return i > 0
			}
			return false
		})

		// Valid int should pass with pointer identity
		val := 2
		result, err := schema.Parse(&val)
		require.NoError(t, err)
		assert.Equal(t, &val, result)
		assert.True(t, result == &val) // Pointer identity preserved

		// nil should pass
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestLiteral_TypeSpecificMethods(t *testing.T) {
	t.Run("Value method returns single value", func(t *testing.T) {
		schema := LiteralOf([]string{"hello"})
		value := schema.Value()
		assert.Equal(t, "hello", value)
	})

	t.Run("Value method panics on multiple values", func(t *testing.T) {
		schema := LiteralOf([]string{"hello", "world"})
		assert.Panics(t, func() {
			schema.Value()
		})
	})

	t.Run("Values method returns all values", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green", "blue"})
		values := schema.Values()
		assert.Len(t, values, 3)
		assert.Contains(t, values, "red")
		assert.Contains(t, values, "green")
		assert.Contains(t, values, "blue")
	})

	t.Run("Contains method checks value membership", func(t *testing.T) {
		schema := LiteralOf([]int{1, 2, 3})

		assert.True(t, schema.Contains(2))
		assert.False(t, schema.Contains(4))
	})
}

// =============================================================================
// Error handling and edge case tests
// =============================================================================

func TestLiteral_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := LiteralOf([]string{"hello"})

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("empty literal schema", func(t *testing.T) {
		schema := LiteralTyped([]string{})
		require.NotNil(t, schema)

		// Any input should fail for empty literal
		_, err := schema.Parse("anything")
		assert.Error(t, err)
	})

	t.Run("nil input without nilable", func(t *testing.T) {
		schema := LiteralOf([]string{"hello"})

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("type mismatch", func(t *testing.T) {
		stringSchema := LiteralOf([]string{"hello"})
		_, err := stringSchema.Parse(42)
		assert.Error(t, err)

		intSchema := LiteralOf([]int{42})
		_, err = intSchema.Parse("hello")
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestLiteral_EdgeCases(t *testing.T) {
	t.Run("zero value literals", func(t *testing.T) {
		// Empty string literal
		emptyStringSchema := LiteralOf([]string{""})
		result, err := emptyStringSchema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)

		_, err = emptyStringSchema.Parse(" ")
		assert.Error(t, err)

		// Zero int literal
		zeroIntSchema := LiteralOf([]int{0})
		result2, err := zeroIntSchema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, 0, result2)

		_, err = zeroIntSchema.Parse(1)
		assert.Error(t, err)

		// False bool literal
		falseBoolSchema := LiteralOf([]bool{false})
		result3, err := falseBoolSchema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result3)

		_, err = falseBoolSchema.Parse(true)
		assert.Error(t, err)
	})

	t.Run("duplicate values in literal", func(t *testing.T) {
		// Should handle duplicate values gracefully
		schema := LiteralOf([]string{"a", "b", "a", "c", "b"})

		result, err := schema.Parse("a")
		require.NoError(t, err)
		assert.Equal(t, "a", result)

		values := schema.Values()
		// Should contain all values including duplicates in original order
		assert.Contains(t, values, "a")
		assert.Contains(t, values, "b")
		assert.Contains(t, values, "c")
	})

	t.Run("single value literal edge cases", func(t *testing.T) {
		schema := LiteralOf([]string{"only"})

		result, err := schema.Parse("only")
		require.NoError(t, err)
		assert.Equal(t, "only", result)

		_, err = schema.Parse("other")
		assert.Error(t, err)

		// Value method should work
		value := schema.Value()
		assert.Equal(t, "only", value)
	})

	t.Run("custom comparable types", func(t *testing.T) {
		type CustomType string
		const (
			CustomA CustomType = "A"
			CustomB CustomType = "B"
		)

		schema := LiteralOf([]CustomType{CustomA, CustomB})

		result, err := schema.Parse(CustomA)
		require.NoError(t, err)
		assert.Equal(t, CustomA, result)
		assert.IsType(t, CustomType(""), result)

		_, err = schema.Parse(CustomType("C"))
		assert.Error(t, err)
	})
}

// =============================================================================
// Pointer identity preservation tests
// =============================================================================

func TestLiteral_PointerIdentityPreservation(t *testing.T) {
	t.Run("String literal Optional correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]string{"red", "green", "blue"}).Optional()

		originalStr := "red"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, "red", *result)
	})

	t.Run("String literal Nilable correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]string{"active", "inactive"}).Nilable()

		originalStr := "active"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// NOW FIXED: Literal.Nilable() correctly returns *string and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, "active", *result)
	})

	t.Run("Int literal Optional correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]int{1, 2, 3}).Optional()

		originalInt := 2
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// NOW FIXED: Literal.Optional() correctly returns *int and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, 2, *result)
	})

	t.Run("Int literal Nilable correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]int{10, 20, 30}).Nilable()

		originalInt := 20
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// NOW FIXED: Literal.Nilable() correctly returns *int and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, 20, *result)
	})

	t.Run("Bool literal Optional correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]bool{true}).Optional()

		originalBool := true
		originalPtr := &originalBool

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// NOW FIXED: Literal.Optional() correctly returns *bool and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, true, *result)
	})

	t.Run("Bool literal Nilable correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]bool{false}).Nilable()

		originalBool := false
		originalPtr := &originalBool

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// NOW FIXED: Literal.Nilable() correctly returns *bool and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, false, *result)
	})

	t.Run("Optional handles nil consistently", func(t *testing.T) {
		schema := LiteralOf([]string{"test"}).Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable handles nil consistently", func(t *testing.T) {
		schema := LiteralOf([]int{123}).Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default().Optional() chaining correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]string{"one", "two", "three"}).Default("one").Optional()

		originalStr := "two"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// NOW FIXED: Chain correctly returns *string and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, "two", *result)
	})

	t.Run("Refine with Optional correctly preserves pointer identity", func(t *testing.T) {
		schema := LiteralOf([]string{"apple", "banana", "cherry"}).
			Refine(func(s string) bool {
				return len(s) > 4 // Only accept longer fruits
			}).Optional()

		originalStr := "banana"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Refined Optional schema correctly returns *string and preserves pointer identity
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, "banana", *result)
	})

	t.Run("Multiple literal types correctly preserve pointer identity", func(t *testing.T) {
		testCases := []struct {
			name string
			test func(t *testing.T)
		}{
			{"string", func(t *testing.T) {
				schema := LiteralOf([]string{"red", "green", "blue"}).Optional()
				originalStr := "red"
				originalPtr := &originalStr

				result, err := schema.Parse(originalPtr)
				require.NoError(t, err)

				assert.True(t, result == originalPtr, "Pointer identity should be preserved")
				assert.Equal(t, "red", *result)
			}},
			{"int", func(t *testing.T) {
				schema := LiteralOf([]int{42, 100, 200}).Optional()
				originalInt := 42
				originalPtr := &originalInt

				result, err := schema.Parse(originalPtr)
				require.NoError(t, err)

				assert.True(t, result == originalPtr, "Pointer identity should be preserved")
				assert.Equal(t, 42, *result)
			}},
			{"bool", func(t *testing.T) {
				schema := LiteralOf([]bool{true, false}).Optional()
				originalBool := true
				originalPtr := &originalBool

				result, err := schema.Parse(originalPtr)
				require.NoError(t, err)

				assert.True(t, result == originalPtr, "Pointer identity should be preserved")
				assert.Equal(t, true, *result)
			}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, tc.test)
		}
	})
}

// =============================================================================
// Additional type support tests
// =============================================================================

func TestLiteral_AdditionalPrimitiveTypes(t *testing.T) {
	t.Run("All primitive type literals with generic API", func(t *testing.T) {
		// Test generic API with different primitive types
		testCases := []struct {
			name string
			test func(t *testing.T)
		}{
			{"float32", func(t *testing.T) {
				schema := LiteralOf([]float32{float32(1.1), float32(2.2)})
				result, err := schema.Parse(float32(1.1))
				require.NoError(t, err)
				assert.Equal(t, float32(1.1), result)
				assert.IsType(t, float32(0), result)
			}},
			{"float64", func(t *testing.T) {
				schema := LiteralOf([]float64{3.14})
				result, err := schema.Parse(3.14)
				require.NoError(t, err)
				assert.Equal(t, 3.14, result)
				assert.IsType(t, float64(0), result)
			}},
			{"int8", func(t *testing.T) {
				schema := LiteralOf([]int8{int8(1), int8(2)})
				result, err := schema.Parse(int8(1))
				require.NoError(t, err)
				assert.Equal(t, int8(1), result)
				assert.IsType(t, int8(0), result)
			}},
			{"int16", func(t *testing.T) {
				schema := LiteralOf([]int16{int16(100)})
				result, err := schema.Parse(int16(100))
				require.NoError(t, err)
				assert.Equal(t, int16(100), result)
				assert.IsType(t, int16(0), result)
			}},
			{"int32", func(t *testing.T) {
				schema := LiteralOf([]int32{int32(1000), int32(2000)})
				result, err := schema.Parse(int32(1000))
				require.NoError(t, err)
				assert.Equal(t, int32(1000), result)
				assert.IsType(t, int32(0), result)
			}},
			{"int64", func(t *testing.T) {
				schema := LiteralOf([]int64{int64(10000)})
				result, err := schema.Parse(int64(10000))
				require.NoError(t, err)
				assert.Equal(t, int64(10000), result)
				assert.IsType(t, int64(0), result)
			}},
			{"uint", func(t *testing.T) {
				schema := LiteralOf([]uint{uint(10), uint(20)})
				result, err := schema.Parse(uint(10))
				require.NoError(t, err)
				assert.Equal(t, uint(10), result)
				assert.IsType(t, uint(0), result)
			}},
			{"uint8", func(t *testing.T) {
				schema := LiteralOf([]uint8{uint8(255)})
				result, err := schema.Parse(uint8(255))
				require.NoError(t, err)
				assert.Equal(t, uint8(255), result)
				assert.IsType(t, uint8(0), result)
			}},
			{"uint16", func(t *testing.T) {
				schema := LiteralOf([]uint16{uint16(65535), uint16(32768)})
				result, err := schema.Parse(uint16(65535))
				require.NoError(t, err)
				assert.Equal(t, uint16(65535), result)
				assert.IsType(t, uint16(0), result)
			}},
			{"uint32", func(t *testing.T) {
				schema := LiteralOf([]uint32{uint32(4294967295)})
				result, err := schema.Parse(uint32(4294967295))
				require.NoError(t, err)
				assert.Equal(t, uint32(4294967295), result)
				assert.IsType(t, uint32(0), result)
			}},
			{"uint64", func(t *testing.T) {
				schema := LiteralOf([]uint64{uint64(18446744073709551615), uint64(9223372036854775808)})
				result, err := schema.Parse(uint64(18446744073709551615))
				require.NoError(t, err)
				assert.Equal(t, uint64(18446744073709551615), result)
				assert.IsType(t, uint64(0), result)
			}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, tc.test)
		}
	})

	t.Run("Pointer constraint constructors work correctly", func(t *testing.T) {
		// Test LiteralPtr (single value)
		ptrSchema := LiteralPtrOf([]string{"hello"})
		var _ *ZodLiteral[string, *string] = ptrSchema

		str := "hello"
		result, err := ptrSchema.Parse(&str)
		require.NoError(t, err)
		assert.Equal(t, &str, result)
		assert.True(t, result == &str) // Pointer identity preserved

		// Test LiteralPtrOf (multiple values)
		ptrMultiSchema := LiteralPtrOf([]string{"hello", "world"})
		var _ *ZodLiteral[string, *string] = ptrMultiSchema

		str2 := "world"
		result2, err := ptrMultiSchema.Parse(&str2)
		require.NoError(t, err)
		assert.Equal(t, &str2, result2)
		assert.True(t, result2 == &str2) // Pointer identity preserved
	})
}
