package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestEnum_BasicFunctionality(t *testing.T) {
	t.Run("valid enum inputs", func(t *testing.T) {
		// String enum
		colorEnum := Enum("red", "green", "blue")

		result, err := colorEnum.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		result, err = colorEnum.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
	})

	t.Run("invalid enum inputs", func(t *testing.T) {
		colorEnum := Enum("red", "green", "blue")

		invalidInputs := []any{
			"yellow", "purple", 123, true, nil, []string{"red"},
		}

		for _, input := range invalidInputs {
			_, err := colorEnum.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("numeric enum", func(t *testing.T) {
		statusEnum := Enum(1, 2, 3, 4)

		result, err := statusEnum.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)

		_, err = statusEnum.Parse(5)
		assert.Error(t, err)
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		enum := Enum("a", "b", "c")

		// Test Parse method
		result, err := enum.Parse("a")
		require.NoError(t, err)
		assert.Equal(t, "a", result)

		// Test MustParse method
		mustResult := enum.MustParse("b")
		assert.Equal(t, "b", mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			enum.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid color"
		enum := EnumMap(map[string]string{
			"red":   "red",
			"green": "green",
			"blue":  "blue",
		}, core.SchemaParams{Error: customError})

		require.NotNil(t, enum)

		_, err := enum.Parse("yellow")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestEnum_TypeSafety(t *testing.T) {
	t.Run("string enum returns string type", func(t *testing.T) {
		enum := Enum("red", "green", "blue")
		require.NotNil(t, enum)

		result, err := enum.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)
		assert.IsType(t, "", result) // Ensure type is string
	})

	t.Run("int enum returns int type", func(t *testing.T) {
		enum := Enum(1, 2, 3)
		require.NotNil(t, enum)

		result, err := enum.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)
		assert.IsType(t, 0, result) // Ensure type is int
	})

	t.Run("custom type enum", func(t *testing.T) {
		type Status int
		const (
			StatusPending Status = iota
			StatusActive
			StatusInactive
		)

		enum := Enum(StatusPending, StatusActive, StatusInactive)

		result, err := enum.Parse(StatusActive)
		require.NoError(t, err)
		assert.Equal(t, StatusActive, result)
		assert.IsType(t, Status(0), result)
	})

	t.Run("EnumPtr returns pointer type", func(t *testing.T) {
		// Create enum schema and then make it optional for pointer behavior
		enum := Enum("a", "b", "c").Optional()
		require.NotNil(t, enum)

		result, err := enum.Parse("a")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "a", *result)
		assert.IsType(t, (*string)(nil), result) // Now returns pointer type
	})
}

// =============================================================================
// ZodEnumPtr tests - Tests for pointer-capable enum types
// =============================================================================

func TestZodEnumPtr_BasicFunctionality(t *testing.T) {
	t.Run("EnumPtr factory function", func(t *testing.T) {
		// Create pointer-capable enum directly
		enum := EnumPtr("red", "green", "blue")
		require.NotNil(t, enum)

		// Test valid parsing returns pointer
		result, err := enum.Parse("red")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "red", *result)

		// Test invalid value
		_, err = enum.Parse("yellow")
		assert.Error(t, err)
	})

	t.Run("EnumSlicePtr factory function", func(t *testing.T) {
		colors := []string{"red", "green", "blue"}
		enum := EnumSlicePtr(colors)
		require.NotNil(t, enum)

		result, err := enum.Parse("green")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "green", *result)
	})

	t.Run("EnumMapPtr factory function", func(t *testing.T) {
		statusMap := map[string]string{"active": "active", "inactive": "inactive", "pending": "pending"}
		enum := EnumMapPtr(statusMap)
		require.NotNil(t, enum)

		result, err := enum.Parse("active")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "active", *result)
	})

	t.Run("Optional/Nilable return ZodEnumPtr", func(t *testing.T) {
		// Test that Optional() returns pointer-capable type
		enum := Enum("a", "b", "c")
		optionalEnum := enum.Optional()
		require.NotNil(t, optionalEnum)

		// Parse valid value - should return pointer
		result, err := optionalEnum.Parse("a")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "a", *result)

		// Parse nil - should return nil pointer for optional behavior
		result, err = optionalEnum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for optional nil

		// Test that Nilable() returns pointer-capable type
		nilableEnum := enum.Nilable()
		require.NotNil(t, nilableEnum)

		result, err = nilableEnum.Parse("b")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "b", *result)
	})
}

func TestZodEnumPtr_PointerBehavior(t *testing.T) {
	t.Run("pointer identity preservation", func(t *testing.T) {
		enum := EnumPtr("x", "y", "z")

		// Parse same value multiple times
		result1, err1 := enum.Parse("x")
		require.NoError(t, err1)
		require.NotNil(t, result1)

		result2, err2 := enum.Parse("x")
		require.NoError(t, err2)
		require.NotNil(t, result2)

		// Values should be equal but pointers should be different instances
		assert.Equal(t, *result1, *result2)
		assert.Equal(t, "x", *result1)
		assert.Equal(t, "x", *result2)
		// Note: We don't require pointer identity preservation for enum values
		// as they're typically small and copied, not referenced
	})

	t.Run("MustParse returns pointer", func(t *testing.T) {
		enum := EnumPtr("alpha", "beta", "gamma")

		result := enum.MustParse("beta")
		require.NotNil(t, result)
		assert.Equal(t, "beta", *result)

		// Test panic on invalid input
		assert.Panics(t, func() {
			enum.MustParse("invalid")
		})
	})

	t.Run("nil handling in optional enum", func(t *testing.T) {
		enum := Enum(1, 2, 3).Optional()

		// Test nil input - should return nil pointer
		result, err := enum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for optional nil

		// Test valid input - should return pointer
		result, err = enum.Parse(2)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, *result)
	})
}

func TestZodEnumPtr_Refine(t *testing.T) {
	t.Run("refine with pointer enum", func(t *testing.T) {
		// Create pointer enum and refine it
		enum := EnumPtr("apple", "banana", "cherry").Refine(func(s *string) bool {
			if s == nil {
				return false
			}
			return len(*s) >= 5 // Only fruits with 5+ letters
		})

		// "apple", "banana", "cherry" all have 5+ letters, so they should pass
		result, err := enum.Parse("apple")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "apple", *result)

		result, err = enum.Parse("banana")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "banana", *result)

		result, err = enum.Parse("cherry")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "cherry", *result)
	})

	t.Run("refineAny with pointer enum", func(t *testing.T) {
		enum := EnumPtr(10, 20, 30).RefineAny(func(v any) bool {
			n, ok := v.(int)
			return ok && n >= 20 // Only values >= 20
		})

		// 20 and 30 should pass
		result, err := enum.Parse(20)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 20, *result)

		result, err = enum.Parse(30)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 30, *result)

		// 10 should fail
		_, err = enum.Parse(10)
		assert.Error(t, err)
	})

	t.Run("chained refine operations", func(t *testing.T) {
		enum := EnumPtr("red", "green", "blue", "yellow").
			Refine(func(s *string) bool {
				if s == nil {
					return false
				}
				return len(*s) >= 4 // First: length >= 4
			}).
			Refine(func(s *string) bool {
				if s == nil {
					return false
				}
				return *s != "blue" // Second: not blue
			})

		// "green" and "yellow" should pass (length >= 4 and not "blue")
		result, err := enum.Parse("green")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "green", *result)

		result, err = enum.Parse("yellow")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "yellow", *result)

		// "red" should fail (length < 4)
		_, err = enum.Parse("red")
		assert.Error(t, err)

		// "blue" should fail (excluded by second refine)
		_, err = enum.Parse("blue")
		assert.Error(t, err)
	})
}

func TestZodEnumPtr_Integration(t *testing.T) {
	t.Run("complex type with pointer enum", func(t *testing.T) {
		type Priority int
		const (
			Low Priority = iota
			Medium
			High
		)

		enum := EnumPtr(Low, Medium, High)

		result, err := enum.Parse(High)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, High, *result)
		assert.IsType(t, Priority(0), *result)
	})

	t.Run("optional pointer enum behavior", func(t *testing.T) {
		enum := Enum("small", "medium", "large").Optional()

		// Test with provided value - should return pointer
		result, err := enum.Parse("large")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "large", *result)

		// Test with nil (should return nil pointer)
		result, err = enum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for optional nil
	})

	t.Run("error handling in pointer enum", func(t *testing.T) {
		enum := EnumPtr("valid1", "valid2")

		// Test invalid value
		_, err := enum.Parse("invalid")
		assert.Error(t, err)

		// Test wrong type
		_, err = enum.Parse(123)
		assert.Error(t, err)

		// Test nil in non-optional enum
		_, err = enum.Parse(nil)
		assert.Error(t, err)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestEnum_Modifiers(t *testing.T) {
	t.Run("Optional always returns pointer type", func(t *testing.T) {
		enum := Enum("red", "green", "blue")
		optionalEnum := enum.Optional()

		// Test non-nil value - should return pointer
		result, err := optionalEnum.Parse("red")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "red", *result)
		assert.IsType(t, (*string)(nil), result)

		// Test nil value (should be allowed for optional) - returns nil pointer
		result, err = optionalEnum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for optional nil
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		enum := Enum(1, 2, 3)
		nilableEnum := enum.Nilable()

		// Test nil handling - should return nil pointer
		result, err := nilableEnum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for nilable nil

		// Test valid value - should return pointer
		result, err = nilableEnum.Parse(2)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, *result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		enum := Enum("a", "b", "c")
		defaultEnum := enum.Default("a")

		// Valid input should override default
		result, err := defaultEnum.Parse("b")
		require.NoError(t, err)
		assert.Equal(t, "b", result)
		assert.IsType(t, "", result) // Should still be string, not pointer
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		enum := Enum(1, 2, 3)
		prefaultEnum := enum.Prefault(1)

		// Valid input should override prefault
		result, err := prefaultEnum.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)
		assert.IsType(t, 0, result) // Should still be int, not pointer
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestEnum_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		enum := Enum("red", "green", "blue").
			Default("red"). // Preserves string type
			Optional()      // Now returns pointer type

		// Test final behavior - should return pointer
		result, err := enum.Parse("blue")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.IsType(t, (*string)(nil), result)
		assert.Equal(t, "blue", *result)

		// Test nil handling - should return default value as pointer
		result, err = enum.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)       // default value takes precedence over optional
		assert.Equal(t, "red", *result) // should return default value
	})

	t.Run("complex chaining", func(t *testing.T) {
		enum := Enum(1, 2, 3).Optional()

		result, err := enum.Parse(2)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, *result) // Should dereference pointer

		// Test nil handling - should return nil pointer
		result, err = enum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for optional nil
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		enum := Enum("a", "b", "c").
			Default("a").
			Prefault("b")

		result, err := enum.Parse("c")
		require.NoError(t, err)
		assert.Equal(t, "c", result)
	})
}

// =============================================================================
// Enum specific methods tests
// =============================================================================

func TestEnum_EnumSpecificMethods(t *testing.T) {
	t.Run("Enum method returns mapping", func(t *testing.T) {
		enum := EnumMap(map[string]string{
			"red":   "RED",
			"green": "GREEN",
			"blue":  "BLUE",
		})

		mapping := enum.Enum()
		assert.Equal(t, map[string]string{
			"red":   "RED",
			"green": "GREEN",
			"blue":  "BLUE",
		}, mapping)
	})

	t.Run("Options method returns all values", func(t *testing.T) {
		enum := Enum("red", "green", "blue")

		options := enum.Options()
		assert.Len(t, options, 3)
		assert.Contains(t, options, "red")
		assert.Contains(t, options, "green")
		assert.Contains(t, options, "blue")
	})

	t.Run("Extract creates sub-enum", func(t *testing.T) {
		originalEnum := EnumMap(map[string]string{
			"red":    "RED",
			"green":  "GREEN",
			"blue":   "BLUE",
			"yellow": "YELLOW",
		})

		subEnum := originalEnum.Extract([]string{"red", "blue"})

		// Should accept extracted values
		result, err := subEnum.Parse("RED")
		require.NoError(t, err)
		assert.Equal(t, "RED", result)

		// Should reject non-extracted values
		_, err = subEnum.Parse("GREEN")
		assert.Error(t, err)
	})

	t.Run("Exclude creates filtered enum", func(t *testing.T) {
		originalEnum := EnumMap(map[string]string{
			"red":   "RED",
			"green": "GREEN",
			"blue":  "BLUE",
		})

		filteredEnum := originalEnum.Exclude([]string{"green"})

		// Should accept non-excluded values
		result, err := filteredEnum.Parse("RED")
		require.NoError(t, err)
		assert.Equal(t, "RED", result)

		// Should reject excluded values
		_, err = filteredEnum.Parse("GREEN")
		assert.Error(t, err)
	})

	t.Run("Extract with invalid key silently ignores", func(t *testing.T) {
		enum := EnumMap(map[string]string{"red": "RED", "blue": "BLUE"})

		// Extract with one valid key and one invalid key
		subEnum := enum.Extract([]string{"red", "invalid"})

		// Should only contain the valid key
		assert.Len(t, subEnum.Enum(), 1)
		assert.Equal(t, "RED", subEnum.Enum()["red"])
		_, exists := subEnum.Enum()["invalid"]
		assert.False(t, exists)
	})

	t.Run("Exclude with invalid key silently ignores", func(t *testing.T) {
		enum := EnumMap(map[string]string{"red": "RED", "blue": "BLUE"})

		// Exclude with one valid key and one invalid key
		filteredEnum := enum.Exclude([]string{"red", "invalid"})

		// Should only exclude the valid key
		assert.Len(t, filteredEnum.Enum(), 1)
		assert.Equal(t, "BLUE", filteredEnum.Enum()["blue"])
		_, exists := filteredEnum.Enum()["red"]
		assert.False(t, exists)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestEnum_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		enum := Enum("a", "b", "c").Default("a").Prefault("b")

		// When input is nil, Default should take precedence
		result, err := enum.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "a", result)
	})

	// Test 2: Default short-circuit mechanism
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		// Create an enum where default value is not in the allowed values
		enum := Enum("valid1", "valid2").Default("invalid_default")

		// Default should bypass validation even if it's not in the enum
		result, err := enum.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "invalid_default", result)
	})

	// Test 3: Prefault requires full validation
	t.Run("Prefault requires full validation", func(t *testing.T) {
		// Create an enum where prefault value is not in the allowed values
		enum := Enum("valid1", "valid2").Prefault("invalid_prefault")

		// Prefault should fail validation if it's not in the enum
		_, err := enum.ParseAny(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid enum value")
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		enum := Enum("valid1", "valid2").Prefault("valid1")

		// Non-nil input that fails validation should not trigger Prefault
		_, err := enum.ParseAny("invalid_input")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid enum value")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		enum := Enum("a", "b", "c").DefaultFunc(func() string {
			defaultCalled = true
			return "a"
		}).PrefaultFunc(func() string {
			prefaultCalled = true
			return "b"
		})

		// DefaultFunc should be called and take precedence
		result, err := enum.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "a", result)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // PrefaultFunc should not be called
	})

	// Test 6: Error handling for Prefault validation failure
	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		enum := Enum(1, 2, 3).Prefault(999) // Prefault value not in enum

		// Should return validation error, not attempt fallback
		_, err := enum.ParseAny(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid enum value")
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestEnum_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		// Only accept values starting with "r"
		enum := Enum("red", "green", "blue").Refine(func(s string) bool {
			return len(s) > 0 && s[0] == 'r'
		})

		result, err := enum.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		_, err = enum.Parse("green")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be a primary color"
		enum := Enum("red", "green", "blue", "yellow").Refine(func(s string) bool {
			return s == "red" || s == "green" || s == "blue"
		}, core.SchemaParams{Error: errorMessage})

		result, err := enum.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		_, err = enum.Parse("yellow")
		assert.Error(t, err)
	})

	t.Run("refine nilable enum", func(t *testing.T) {
		enum := Enum(1, 2, 3).Nilable().Refine(func(n *int) bool {
			if n == nil {
				return true // Allow nil values in nilable enum
			}
			return *n == 0 || *n > 1 // 0 is zero value for nil
		})

		// nil should pass (returns nil pointer)
		result, err := enum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for nilable nil

		// value > 1 should pass (returns pointer)
		result, err = enum.Parse(2)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 2, *result)

		// value <= 1 should fail (except 0)
		_, err = enum.Parse(1)
		assert.Error(t, err)
	})
}

func TestEnum_RefineAny(t *testing.T) {
	t.Run("refineAny flexible validation", func(t *testing.T) {
		enum := Enum("red", "green", "blue").RefineAny(func(v any) bool {
			s, ok := v.(string)
			return ok && len(s) >= 4
		})

		// "green" and "blue" should pass (>= 4 chars)
		result, err := enum.Parse("green")
		require.NoError(t, err)
		assert.Equal(t, "green", result)

		result, err = enum.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)

		// "red" should fail (< 4 chars)
		_, err = enum.Parse("red")
		assert.Error(t, err)
	})

	t.Run("refineAny with type checking", func(t *testing.T) {
		enum := Enum(1, 2, 3).RefineAny(func(v any) bool {
			n, ok := v.(int)
			return ok && n%2 == 0 // Only even numbers
		})

		result, err := enum.Parse(2)
		require.NoError(t, err)
		assert.Equal(t, 2, result)

		_, err = enum.Parse(1)
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestEnum_ErrorHandling(t *testing.T) {
	t.Run("invalid value error", func(t *testing.T) {
		enum := Enum("a", "b", "c")

		_, err := enum.Parse("d")
		assert.Error(t, err)
	})

	t.Run("invalid type error", func(t *testing.T) {
		enum := Enum("a", "b", "c")

		_, err := enum.Parse(123)
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		enum := EnumMap(map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
		}, core.SchemaParams{Error: "Expected a valid option"})

		_, err := enum.Parse("d")
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestEnum_EdgeCases(t *testing.T) {
	t.Run("empty enum", func(t *testing.T) {
		enum := EnumSlice([]string{})
		require.NotNil(t, enum)

		// Any input should fail for empty enum
		_, err := enum.Parse("anything")
		assert.Error(t, err)
	})

	t.Run("single value enum", func(t *testing.T) {
		enum := Enum("only")

		result, err := enum.Parse("only")
		require.NoError(t, err)
		assert.Equal(t, "only", result)

		_, err = enum.Parse("other")
		assert.Error(t, err)
	})

	t.Run("nil handling with pointer enum", func(t *testing.T) {
		enum := Enum("a", "b", "c").Nilable()

		// Test nil input - should return nil pointer
		result, err := enum.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result) // nil pointer for nilable nil

		// Test valid value - should return pointer
		result, err = enum.Parse("a")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "a", *result)
	})

	t.Run("duplicate values in enum", func(t *testing.T) {
		// Should handle duplicate values gracefully
		enum := Enum("a", "b", "a", "c", "b")

		result, err := enum.Parse("a")
		require.NoError(t, err)
		assert.Equal(t, "a", result)

		options := enum.Options()
		// Should contain unique values only
		assert.Contains(t, options, "a")
		assert.Contains(t, options, "b")
		assert.Contains(t, options, "c")
	})

	t.Run("empty context", func(t *testing.T) {
		enum := Enum("a", "b", "c")

		// Parse with empty context slice
		result, err := enum.Parse("a")
		require.NoError(t, err)
		assert.Equal(t, "a", result)
	})
}

// =============================================================================
// Multi-error collection tests (TypeScript Zod v4 behavior)
// =============================================================================

func TestEnum_MultiErrorCollection(t *testing.T) {
	t.Run("collect enum value and refinement errors", func(t *testing.T) {
		// Create an enum with refinement that will fail
		colorEnum := Enum("red", "green", "blue").Refine(func(value string) bool {
			return len(value) > 5 // All enum values are too short, so this will always fail
		}, "Color name must be longer than 5 characters")

		// Test with invalid enum value
		_, err := colorEnum.Parse("yellow") // Invalid enum value
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 1 error for invalid enum value (refinement won't run for invalid values)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.IssueCode("invalid_value"), zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "Invalid enum value")

		// Test with valid enum value that fails refinement
		_, err = colorEnum.Parse("red") // Valid enum value but fails refinement
		require.Error(t, err)

		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 1 error for failed refinement (enum validation passes)
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.IssueCode("custom"), zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "Color name must be longer than 5 characters")
	})

	t.Run("multiple refinements on enum", func(t *testing.T) {
		// Create enum with multiple refinements that will fail
		numberEnum := Enum(1, 2, 3, 4, 5).
			Refine(func(value int) bool {
				return value%2 == 0 // Must be even
			}, "Number must be even").
			Refine(func(value int) bool {
				return value > 10 // Must be greater than 10
			}, "Number must be greater than 10")

		// Test with valid enum value that fails both refinements
		_, err := numberEnum.Parse(3) // Valid enum, but odd and not > 10
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should collect both refinement errors
		assert.Len(t, zodErr.Issues, 2)

		// Verify both refinement error messages are present
		errorMessages := make([]string, len(zodErr.Issues))
		for i, issue := range zodErr.Issues {
			errorMessages[i] = issue.Message
			assert.Equal(t, core.IssueCode("custom"), issue.Code)
		}

		assert.Contains(t, errorMessages, "Number must be even")
		assert.Contains(t, errorMessages, "Number must be greater than 10")
	})

	t.Run("enum value validation only", func(t *testing.T) {
		// Simple enum without additional checks
		statusEnum := Enum("active", "inactive", "pending")

		// Test with invalid enum value
		_, err := statusEnum.Parse("unknown")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 1 error for invalid enum value
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.IssueCode("invalid_value"), zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "Invalid enum value")
	})

	t.Run("enum with passing refinements", func(t *testing.T) {
		// Create enum with refinements that will pass
		colorEnum := Enum("red", "green", "blue").
			Refine(func(value string) bool {
				return len(value) >= 3 // All colors have at least 3 characters
			}, "Color name must be at least 3 characters").
			Refine(func(value string) bool {
				return value != "yellow" // None of our colors are yellow
			}, "Color cannot be yellow")

		// Test with valid enum value that passes all refinements
		result, err := colorEnum.Parse("red")
		require.NoError(t, err)
		assert.Equal(t, "red", result)

		// Test with valid enum value that passes all refinements
		result, err = colorEnum.Parse("blue")
		require.NoError(t, err)
		assert.Equal(t, "blue", result)
	})

	t.Run("numeric enum with refinements", func(t *testing.T) {
		// Create numeric enum with validation
		scoreEnum := Enum(0, 1, 2, 3, 4, 5).
			Refine(func(value int) bool {
				return value >= 2 // Minimum score of 2
			}, "Score must be at least 2")

		// Test with valid enum but failing refinement
		_, err := scoreEnum.Parse(1) // Valid enum value but score < 2
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))

		// Should have 1 refinement error
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.IssueCode("custom"), zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "Score must be at least 2")

		// Test with valid enum and passing refinement
		result, err := scoreEnum.Parse(3)
		require.NoError(t, err)
		assert.Equal(t, 3, result)
	})
}
