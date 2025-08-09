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
// Basic functionality tests
// =============================================================================

func TestInt_BasicFunctionality(t *testing.T) {
	t.Run("valid integer inputs", func(t *testing.T) {
		schema := Int64()

		// Test positive value
		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		// Test negative value
		result, err = schema.Parse(int64(-10))
		require.NoError(t, err)
		assert.Equal(t, int64(-10), result)

		// Test zero
		result, err = schema.Parse(int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})

	t.Run("different integer types", func(t *testing.T) {
		// Test int32
		schema32 := Int32()
		result32, err := schema32.Parse(int32(100))
		require.NoError(t, err)
		assert.Equal(t, int32(100), result32)

		// Test uint64
		schemaUint64 := Uint64()
		resultUint64, err := schemaUint64.Parse(uint64(200))
		require.NoError(t, err)
		assert.Equal(t, uint64(200), resultUint64)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Int64()

		invalidInputs := []any{
			"not a number", 3.14, []int{1}, map[string]int{"key": 1}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Int64()

		// Test Parse method
		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		// Test MustParse method
		mustResult := schema.MustParse(int64(42))
		assert.Equal(t, int64(42), mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected an integer value"
		schema := Int64(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeInt64, schema.internals.Def.ZodTypeDef.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestInt_TypeSafety(t *testing.T) {
	t.Run("Int returns int type", func(t *testing.T) {
		schema := Int()
		require.NotNil(t, schema)

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.IsType(t, int(0), result)
	})

	t.Run("Int64 returns int64 type", func(t *testing.T) {
		schema := Int64()
		require.NotNil(t, schema)

		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
		assert.IsType(t, int64(0), result)
	})

	t.Run("IntPtr returns *int type", func(t *testing.T) {
		schema := IntPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)
		assert.IsType(t, (*int)(nil), result)
	})

	t.Run("Int64Ptr returns *int64 type", func(t *testing.T) {
		schema := Int64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(42), *result)
		assert.IsType(t, (*int64)(nil), result)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test int type
		intSchema := Int()
		resultInt := intSchema.MustParse(42)
		assert.IsType(t, int(0), resultInt)
		assert.Equal(t, 42, resultInt)

		// Test int64 type
		int64Schema := Int64()
		resultInt64 := int64Schema.MustParse(int64(42))
		assert.IsType(t, int64(0), resultInt64)
		assert.Equal(t, int64(42), resultInt64)

		// Test *int64 type
		ptrSchema := Int64Ptr()
		ptrResult := ptrSchema.MustParse(int64(42))
		assert.IsType(t, (*int64)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, int64(42), *ptrResult)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestInt_Modifiers(t *testing.T) {
	t.Run("Optional returns type-safe pointer constraint", func(t *testing.T) {
		// Int32().Optional() should return *ZodIntegerTyped[int32, *int32]
		int32Schema := Int32()
		optionalSchema := int32Schema.Optional()

		// Type check: ensure it returns the correct type
		var _ *ZodIntegerTyped[int32, *int32] = optionalSchema

		// Test with value
		result, err := optionalSchema.Parse(int32(42))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int32(42), *result)

		// Test with nil (should work for optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable returns type-safe pointer constraint", func(t *testing.T) {
		// Int64().Nilable() should return *ZodIntegerTyped[int64, *int64]
		int64Schema := Int64()
		nilableSchema := int64Schema.Nilable()

		// Type check: ensure it returns the correct type
		var _ *ZodIntegerTyped[int64, *int64] = nilableSchema

		// Test with value
		result, err := nilableSchema.Parse(int64(123))
		require.NoError(t, err)
		assert.IsType(t, (*int64)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int64(123), *result)

		// Test with nil (should work for nilable)
		result, err = nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nullish returns type-safe pointer constraint", func(t *testing.T) {
		// Uint16().Nullish() should return *ZodIntegerTyped[uint16, *uint16]
		uint16Schema := Uint16()
		nullishSchema := uint16Schema.Nullish()

		// Type check: ensure it returns the correct type
		var _ *ZodIntegerTyped[uint16, *uint16] = nullishSchema

		// Test with value
		result, err := nullishSchema.Parse(uint16(789))
		require.NoError(t, err)
		assert.IsType(t, (*uint16)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, uint16(789), *result)

		// Test with nil (should work for nullish)
		result, err = nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Chaining modifiers preserves type safety", func(t *testing.T) {
		// Test chaining: Int().Optional().Default(42)
		schema := Int().Optional()

		// Type check
		var _ *ZodIntegerTyped[int, *int] = schema

		// Add validation
		validatedSchema := schema.Min(0).Max(100)
		var _ *ZodIntegerTyped[int, *int] = validatedSchema

		// Test functionality
		result, err := validatedSchema.Parse(50)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 50, *result)
	})

	t.Run("Different integer types work correctly", func(t *testing.T) {
		tests := []struct {
			name   string
			schema func() any
		}{
			{"int8", func() any { return Int8().Optional() }},
			{"int16", func() any { return Int16().Optional() }},
			{"int32", func() any { return Int32().Optional() }},
			{"int64", func() any { return Int64().Optional() }},
			{"uint", func() any { return Uint().Optional() }},
			{"uint8", func() any { return Uint8().Optional() }},
			{"uint16", func() any { return Uint16().Optional() }},
			{"uint32", func() any { return Uint32().Optional() }},
			{"uint64", func() any { return Uint64().Optional() }},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				schema := tt.schema()
				assert.NotNil(t, schema)
				// Just ensure they compile and create instances
			})
		}
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestInt_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		schema := Int(). // *ZodIntegerTyped[int, int]
					Default(42). // *ZodIntegerTyped[int, int] (maintains type)
					Optional()   // *ZodIntegerTyped[int, *int] (type conversion)

		var _ *ZodIntegerTyped[int, *int] = schema

		// Test final behavior
		result, err := schema.Parse(int(100))
		require.NoError(t, err)
		assert.IsType(t, (*int)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int(100), *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := Int(). // *ZodIntegerTyped[int, int]
					Default(42). // *ZodIntegerTyped[int, int] (maintains type)
					Optional()   // *ZodIntegerTyped[int, *int] (type conversion)

		var _ *ZodIntegerTyped[int, *int] = schema

		// Test final behavior
		result, err := schema.Parse(int(100))
		require.NoError(t, err)
		assert.IsType(t, (*int)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int(100), *result)
	})

	t.Run("validation chaining", func(t *testing.T) {
		schema := Int64().
			Min(0).
			Max(100).
			Positive()

		result, err := schema.Parse(int64(50))
		require.NoError(t, err)
		assert.Equal(t, int64(50), result)

		// Should fail validation
		_, err = schema.Parse(int64(-5))
		assert.Error(t, err)

		_, err = schema.Parse(int64(150))
		assert.Error(t, err)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Int64().
			Default(42).
			Prefault(0)

		result, err := schema.Parse(int64(100))
		require.NoError(t, err)
		assert.Equal(t, int64(100), result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestInt_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// Int64 type
		schema1 := Int64().Default(100).Prefault(200)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, int64(100), result1) // Should be default, not prefault

		// Int64Ptr type
		schema2 := Int64Ptr().Default(100).Prefault(200)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, int64(100), *result2) // Should be default, not prefault
	})

	// Test 2: Default short-circuits validation
	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Int64 type - default value violates Min(50) but should still work
		schema1 := Int64().Min(50).Default(10)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, int64(10), result1) // Default bypasses validation

		// Int64Ptr type - default value violates Min(50) but should still work
		schema2 := Int64Ptr().Min(50).Default(10)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, int64(10), *result2) // Default bypasses validation
	})

	// Test 3: Prefault goes through full validation
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Int64 type - prefault value passes validation
		schema1 := Int64().Min(50).Prefault(100)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, int64(100), result1) // Prefault passes validation

		// Int64 type - prefault value fails validation
		schema2 := Int64().Min(50).Prefault(10)
		_, err2 := schema2.Parse(nil)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "Too small") // Prefault fails validation

		// Int64Ptr type - prefault value passes validation
		schema3 := Int64Ptr().Min(50).Prefault(100)
		result3, err3 := schema3.Parse(nil)
		require.NoError(t, err3)
		require.NotNil(t, result3)
		assert.Equal(t, int64(100), *result3) // Prefault passes validation

		// Int64Ptr type - prefault value fails validation
		schema4 := Int64Ptr().Min(50).Prefault(10)
		_, err4 := schema4.Parse(nil)
		require.Error(t, err4)
		assert.Contains(t, err4.Error(), "Too small") // Prefault fails validation
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		// Int64 type
		schema1 := Int64().Min(50).Prefault(100)

		// Nil input triggers prefault
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, int64(100), result1)

		// Non-nil input validation failure should not trigger prefault
		_, err2 := schema1.Parse(int64(10)) // Less than Min(50)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "Too small")

		// Int64Ptr type
		schema2 := Int64Ptr().Min(50).Prefault(100)

		// Nil input triggers prefault
		result3, err3 := schema2.Parse(nil)
		require.NoError(t, err3)
		require.NotNil(t, result3)
		assert.Equal(t, int64(100), *result3)

		// Non-nil input validation failure should not trigger prefault
		_, err4 := schema2.Parse(int64(10)) // Less than Min(50)
		require.Error(t, err4)
		assert.Contains(t, err4.Error(), "Too small")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		// DefaultFunc should not be called for non-nil input
		callCount := 0
		schema1 := Int64().DefaultFunc(func() int64 {
			callCount++
			return 100
		})

		// Non-nil input should not call DefaultFunc
		result1, err1 := schema1.Parse(int64(50))
		require.NoError(t, err1)
		assert.Equal(t, int64(50), result1)
		assert.Equal(t, 0, callCount)

		// Nil input should call DefaultFunc
		result2, err2 := schema1.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, int64(100), result2)
		assert.Equal(t, 1, callCount)

		// PrefaultFunc should go through validation
		schema2 := Int64().Min(50).PrefaultFunc(func() int64 {
			return 100 // Valid value
		})
		result3, err3 := schema2.Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, int64(100), result3)

		// PrefaultFunc with invalid value should fail
		schema3 := Int64().Min(50).PrefaultFunc(func() int64 {
			return 10 // Invalid value
		})
		_, err4 := schema3.Parse(nil)
		require.Error(t, err4)
		assert.Contains(t, err4.Error(), "Too small")
	})

	// Test 6: Error handling - prefault validation failure
	t.Run("Error handling - prefault validation failure", func(t *testing.T) {
		// Prefault value fails validation, should return error directly
		schema := Int64().Min(100).Prefault(50)
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
		// Should not try any other fallback
	})
}

// =============================================================================
// Validation tests
// =============================================================================

func TestInt_Validations(t *testing.T) {
	t.Run("min validation", func(t *testing.T) {
		schema := Int64().Min(5)

		// Valid input
		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)

		// Invalid input
		_, err = schema.Parse(int64(3))
		assert.Error(t, err)
	})

	t.Run("max validation", func(t *testing.T) {
		schema := Int64().Max(10)

		// Valid input
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)

		// Invalid input
		_, err = schema.Parse(int64(15))
		assert.Error(t, err)
	})

	t.Run("range validation", func(t *testing.T) {
		schema := Int64().Min(1).Max(10)

		// Valid input
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)

		// Invalid inputs
		_, err = schema.Parse(int64(0))
		assert.Error(t, err)

		_, err = schema.Parse(int64(15))
		assert.Error(t, err)
	})

	t.Run("positive validation", func(t *testing.T) {
		schema := Int64().Positive()

		// Valid input
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)

		// Invalid inputs
		_, err = schema.Parse(int64(0))
		assert.Error(t, err)

		_, err = schema.Parse(int64(-5))
		assert.Error(t, err)
	})

	t.Run("negative validation", func(t *testing.T) {
		schema := Int64().Negative()

		// Valid input
		result, err := schema.Parse(int64(-5))
		require.NoError(t, err)
		assert.Equal(t, int64(-5), result)

		// Invalid inputs
		_, err = schema.Parse(int64(0))
		assert.Error(t, err)

		_, err = schema.Parse(int64(5))
		assert.Error(t, err)
	})

	t.Run("non-negative validation", func(t *testing.T) {
		schema := Int64().NonNegative()

		// Valid inputs
		result, err := schema.Parse(int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)

		result, err = schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)

		// Invalid input
		_, err = schema.Parse(int64(-5))
		assert.Error(t, err)
	})

	t.Run("non-positive validation", func(t *testing.T) {
		schema := Int64().NonPositive()

		// Valid inputs
		result, err := schema.Parse(int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)

		result, err = schema.Parse(int64(-5))
		require.NoError(t, err)
		assert.Equal(t, int64(-5), result)

		// Invalid input
		_, err = schema.Parse(int64(5))
		assert.Error(t, err)
	})

	t.Run("multiple of validation", func(t *testing.T) {
		schema := Int64().MultipleOf(5)

		// Valid inputs
		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)

		result, err = schema.Parse(int64(15))
		require.NoError(t, err)
		assert.Equal(t, int64(15), result)

		// Invalid input
		_, err = schema.Parse(int64(7))
		assert.Error(t, err)
	})

	t.Run("step validation", func(t *testing.T) {
		schema := Int64().Step(3)

		// Valid inputs (multiples of 3)
		result, err := schema.Parse(int64(6))
		require.NoError(t, err)
		assert.Equal(t, int64(6), result)

		result, err = schema.Parse(int64(9))
		require.NoError(t, err)
		assert.Equal(t, int64(9), result)

		// Invalid input
		_, err = schema.Parse(int64(7))
		assert.Error(t, err)
	})

	t.Run("safe validation", func(t *testing.T) {
		schema := Int64().Safe()

		// Valid input (within safe range)
		result, err := schema.Parse(int64(1000000))
		require.NoError(t, err)
		assert.Equal(t, int64(1000000), result)

		// Invalid input (outside safe range)
		_, err = schema.Parse(int64(1 << 54))
		assert.Error(t, err)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestInt_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept even numbers
		schema := Int64().Refine(func(i int64) bool {
			return i%2 == 0
		})

		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)

		_, err = schema.Parse(int64(3))
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be even"
		schema := Int64Ptr().Refine(func(i int64) bool {
			return i%2 == 0
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(10), *result)

		_, err = schema.Parse(int64(3))
		assert.Error(t, err)
	})

	t.Run("refine pointer allows nil", func(t *testing.T) {
		schema := Int64Ptr().Nilable().Refine(func(i int64) bool {
			// Accept even numbers (nil is handled separately by nilable)
			return i%2 == 0
		})

		// Expect nil to be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid value should pass
		result, err = schema.Parse(int64(10))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(10), *result)

		// Invalid value should fail
		_, err = schema.Parse(int64(3))
		assert.Error(t, err)
	})

	t.Run("refine int32 type", func(t *testing.T) {
		schema := Int32().Refine(func(i int32) bool {
			return i%2 == 0
		})

		result, err := schema.Parse(int32(10))
		require.NoError(t, err)
		assert.Equal(t, int32(10), result)

		_, err = schema.Parse(int32(3))
		assert.Error(t, err)
	})
}

func TestInt_RefineAny(t *testing.T) {
	t.Run("refineAny int64 schema", func(t *testing.T) {
		// Only accept even numbers via RefineAny
		schema := Int64().RefineAny(func(v any) bool {
			i, ok := v.(int64)
			return ok && i%2 == 0
		})

		// Valid value passes
		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)

		// Invalid value fails
		_, err = schema.Parse(int64(3))
		assert.Error(t, err)
	})

	t.Run("refineAny pointer schema", func(t *testing.T) {
		// Int64Ptr().RefineAny sees underlying int64 value
		schema := Int64Ptr().RefineAny(func(v any) bool {
			i, ok := v.(int64)
			return ok && i%2 == 0
		})

		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(10), *result)

		_, err = schema.Parse(int64(3))
		assert.Error(t, err)
	})

	t.Run("refineAny nilable schema", func(t *testing.T) {
		// Nil input should bypass checks and be accepted because schema is Nilable()
		schema := Int64Ptr().Nilable().RefineAny(func(v any) bool {
			// Never called for nil input, but keep return true for completeness
			return true
		})

		// nil passes
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid value still passes
		result, err = schema.Parse(int64(10))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(10), *result)
	})
}

// =============================================================================
// Overwrite tests
// =============================================================================

func TestInt_Overwrite(t *testing.T) {
	t.Run("basic overwrite functionality", func(t *testing.T) {
		schema := Int64().Overwrite(func(val int64) int64 {
			return val * 2
		})

		// Test with valid input
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)

		// Test with negative value
		result, err = schema.Parse(int64(-3))
		require.NoError(t, err)
		assert.Equal(t, int64(-6), result)

		// Test with zero
		result, err = schema.Parse(int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})

	t.Run("overwrite preserves type", func(t *testing.T) {
		// Test Int64 preserves type
		schema64 := Int64().Overwrite(func(val int64) int64 {
			return val + 1
		})

		result64, err := schema64.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(6), result64)
		assert.IsType(t, int64(0), result64)

		// Test Int32 preserves type
		schema32 := Int32().Overwrite(func(val int32) int32 {
			return val + 1
		})

		result32, err := schema32.Parse(int32(5))
		require.NoError(t, err)
		assert.Equal(t, int32(6), result32)
		assert.IsType(t, int32(0), result32)

		// Test flexible Integer works with int64 type
		schemaInt := Int64().Overwrite(func(val int64) int64 {
			return val + 1
		})

		resultInt, err := schemaInt.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(6), resultInt)
		assert.IsType(t, int64(0), resultInt)
	})

	t.Run("overwrite with mathematical operations", func(t *testing.T) {
		// Test absolute value transformation
		schema := Int64().Overwrite(func(val int64) int64 {
			if val < 0 {
				return -val
			}
			return val
		})

		// Test positive value remains positive
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)

		// Test negative value becomes positive
		result, err = schema.Parse(int64(-7))
		require.NoError(t, err)
		assert.Equal(t, int64(7), result)
	})

	t.Run("overwrite can be chained with validations", func(t *testing.T) {
		schema := Int64().
			Overwrite(func(val int64) int64 {
				if val < 0 {
					return -val // Convert to absolute value
				}
				return val
			}).
			Max(10)

		// Test positive value within limit
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), result)

		// Test negative value converted to positive and within limit
		result, err = schema.Parse(int64(-8))
		require.NoError(t, err)
		assert.Equal(t, int64(8), result)

		// Test negative value that exceeds limit after conversion
		_, err = schema.Parse(int64(-15))
		assert.Error(t, err)
	})

	t.Run("multiple overwrite calls", func(t *testing.T) {
		schema := Int64().
			Overwrite(func(val int64) int64 {
				return val * 2
			}).
			Overwrite(func(val int64) int64 {
				return val + 1
			})

		// Should apply transformations in order: 5 * 2 = 10, then 10 + 1 = 11
		result, err := schema.Parse(int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(11), result)
	})

	t.Run("overwrite with custom error message", func(t *testing.T) {
		// Test that overwrite still allows custom error messages for other validations
		schema := Int64().
			Overwrite(func(val int64) int64 {
				return val * 2
			}).
			Max(10, "Value must be at most 10 after doubling")

		// Test valid case
		result, err := schema.Parse(int64(4))
		require.NoError(t, err)
		assert.Equal(t, int64(8), result)

		// Test invalid case
		_, err = schema.Parse(int64(6)) // 6 * 2 = 12, which is > 10
		assert.Error(t, err)
	})

	t.Run("overwrite with MustParse", func(t *testing.T) {
		schema := Int64().Overwrite(func(val int64) int64 {
			return val * 3
		})

		result := schema.MustParse(int64(4))
		assert.Equal(t, int64(12), result)

		// Test panic on invalid input type
		assert.Panics(t, func() {
			schema.MustParse("not a number")
		})
	})

	t.Run("overwrite return type matches original schema", func(t *testing.T) {
		// Unlike Transform which returns ZodTransform, Overwrite should return same type
		original := Int64()
		overwritten := original.Overwrite(func(val int64) int64 {
			return val + 1
		})

		// Should be the same underlying type
		var _ *ZodIntegerTyped[int64, int64] = overwritten

		// Can continue chaining int methods
		final := overwritten.Max(100)
		var _ *ZodIntegerTyped[int64, int64] = final
	})

	t.Run("overwrite with different integer types", func(t *testing.T) {
		// Test with uint64
		schemaUint64 := Uint64().Overwrite(func(val uint64) uint64 {
			return val + 1
		})

		resultUint64, err := schemaUint64.Parse(uint64(5))
		require.NoError(t, err)
		assert.Equal(t, uint64(6), resultUint64)
		assert.IsType(t, uint64(0), resultUint64)

		// Test with int8
		schemaInt8 := Int8().Overwrite(func(val int8) int8 {
			return val * 2
		})

		resultInt8, err := schemaInt8.Parse(int8(10))
		require.NoError(t, err)
		assert.Equal(t, int8(20), resultInt8)
		assert.IsType(t, int8(0), resultInt8)
	})

	t.Run("overwrite strict type checking", func(t *testing.T) {
		// Test that overwrite only accepts matching integer types
		schema := Int64().Overwrite(func(val int64) int64 {
			return val * 10
		})

		// Should work with int64 input
		result, err := schema.Parse(int64(3))
		require.NoError(t, err)
		assert.Equal(t, int64(30), result)

		// Should reject int input (different integer type)
		_, err = schema.Parse(int(3))
		assert.Error(t, err, "Should reject int input")

		// Should reject int32 input (different integer type)
		_, err = schema.Parse(int32(5))
		assert.Error(t, err, "Should reject int32 input")

		// Should reject uint input (different integer type)
		_, err = schema.Parse(uint(7))
		assert.Error(t, err, "Should reject uint input")
	})

	t.Run("overwrite int32 strict type checking", func(t *testing.T) {
		// Test that int32 schema only accepts int32 types
		schema := Int32().Overwrite(func(val int32) int32 {
			return val + 100
		})

		// Should work with int32 input
		result, err := schema.Parse(int32(5))
		require.NoError(t, err)
		assert.Equal(t, int32(105), result)

		// Should reject int input (different integer type)
		_, err = schema.Parse(int(5))
		assert.Error(t, err, "Should reject int input")

		// Should reject int64 input (different integer type)
		_, err = schema.Parse(int64(5))
		assert.Error(t, err, "Should reject int64 input")

		// Should reject int8 input (different integer type)
		_, err = schema.Parse(int8(5))
		assert.Error(t, err, "Should reject int8 input")
	})

	t.Run("overwrite transformation showcase", func(t *testing.T) {
		// Demonstrate TypeScript Zod-like behavior
		schema := Int64().Overwrite(func(val int64) int64 {
			// Ensure value is always positive
			if val < 0 {
				return -val
			}
			return val
		})

		// Test positive value (unchanged)
		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		// Test negative value (converted to positive)
		result, err = schema.Parse(int64(-42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		// Test zero (unchanged)
		result, err = schema.Parse(int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})
}

// =============================================================================
// Coercion tests
// =============================================================================

func TestInt_Coercion(t *testing.T) {
	t.Run("string coercion", func(t *testing.T) {
		schema := CoercedInt64()

		// Test string "42" -> int64 42
		result, err := schema.Parse("42")
		require.NoError(t, err, "Should coerce string '42' to int64 42")
		assert.Equal(t, int64(42), result)

		// Test string "-10" -> int64 -10
		result, err = schema.Parse("-10")
		require.NoError(t, err, "Should coerce string '-10' to int64 -10")
		assert.Equal(t, int64(-10), result)
	})

	t.Run("float coercion", func(t *testing.T) {
		schema := CoercedInt64()

		testCases := []struct {
			input    any
			expected int64
			name     string
		}{
			{42.0, 42, "float64 42.0 to 42"},
			{float32(10.0), 10, "float32 10.0 to 10"},
			{-5.0, -5, "float64 -5.0 to -5"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("boolean coercion", func(t *testing.T) {
		schema := CoercedInt64()

		// Boolean values should be coerced to integers
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result)

		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})

	t.Run("integer type coercion", func(t *testing.T) {
		schema := CoercedInt64()

		testCases := []struct {
			input    any
			expected int64
			name     string
		}{
			{int(42), 42, "int 42 to int64 42"},
			{int8(10), 10, "int8 10 to int64 10"},
			{int16(100), 100, "int16 100 to int64 100"},
			{int32(1000), 1000, "int32 1000 to int64 1000"},
			{uint(42), 42, "uint 42 to int64 42"},
			{uint8(10), 10, "uint8 10 to int64 10"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("coerced int32 schema", func(t *testing.T) {
		schema := CoercedInt32()

		// Test string coercion to int32
		result, err := schema.Parse("42")
		require.NoError(t, err)
		assert.IsType(t, int32(0), result)
		assert.Equal(t, int32(42), result)

		// Test float coercion to int32
		result, err = schema.Parse(42.0)
		require.NoError(t, err)
		assert.Equal(t, int32(42), result)
	})

	t.Run("invalid coercion inputs", func(t *testing.T) {
		schema := CoercedInt64()

		// Inputs that cannot be coerced
		invalidInputs := []any{
			"not a number", "invalid", []int{1}, map[string]int{"key": 1}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedInt64().Min(5).Max(100)

		// Coercion then validation passes
		result, err := schema.Parse("50")
		require.NoError(t, err)
		assert.Equal(t, int64(50), result)

		// Coercion then validation fails
		_, err = schema.Parse("3")
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestInt_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := Int64()

		_, err := schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected an integer value"
		schema := Int64Ptr(core.SchemaParams{Error: customError})

		_, err := schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("validation error messages", func(t *testing.T) {
		schema := Int64().Min(10)

		_, err := schema.Parse(int64(5))
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestInt_EdgeCases(t *testing.T) {
	t.Run("nil handling with *int64", func(t *testing.T) {
		schema := Int64Ptr().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid integer
		result, err = schema.Parse(int64(42))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(42), *result)
	})

	t.Run("boundary values", func(t *testing.T) {
		schema := Int64()

		// Test zero
		result, err := schema.Parse(int64(0))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)

		// Test maximum int64
		result, err = schema.Parse(int64(9223372036854775807))
		require.NoError(t, err)
		assert.Equal(t, int64(9223372036854775807), result)

		// Test minimum int64
		result, err = schema.Parse(int64(-9223372036854775808))
		require.NoError(t, err)
		assert.Equal(t, int64(-9223372036854775808), result)
	})

	t.Run("uint64 boundary values", func(t *testing.T) {
		schema := Uint64()

		// Test zero
		result, err := schema.Parse(uint64(0))
		require.NoError(t, err)
		assert.Equal(t, uint64(0), result)

		// Test maximum uint64
		result, err = schema.Parse(uint64(18446744073709551615))
		require.NoError(t, err)
		assert.Equal(t, uint64(18446744073709551615), result)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Int64()

		// Parse with empty context slice
		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	t.Run("different integer sizes", func(t *testing.T) {
		// Test all integer sizes work correctly
		schemas := []struct {
			name   string
			schema any
			input  any
			output any
		}{
			{"int8", Int8(), int8(42), int8(42)},
			{"int16", Int16(), int16(42), int16(42)},
			{"int32", Int32(), int32(42), int32(42)},
			{"uint8", Uint8(), uint8(42), uint8(42)},
			{"uint16", Uint16(), uint16(42), uint16(42)},
			{"uint32", Uint32(), uint32(42), uint32(42)},
		}

		for _, test := range schemas {
			t.Run(test.name, func(t *testing.T) {
				// Use type assertion to get the Parse method
				switch s := test.schema.(type) {
				case *ZodIntegerTyped[int8, int8]:
					result, err := s.Parse(test.input)
					require.NoError(t, err)
					assert.Equal(t, test.output, result)
				case *ZodIntegerTyped[int16, int16]:
					result, err := s.Parse(test.input)
					require.NoError(t, err)
					assert.Equal(t, test.output, result)
				case *ZodIntegerTyped[int32, int32]:
					result, err := s.Parse(test.input)
					require.NoError(t, err)
					assert.Equal(t, test.output, result)
				case *ZodIntegerTyped[uint8, uint8]:
					result, err := s.Parse(test.input)
					require.NoError(t, err)
					assert.Equal(t, test.output, result)
				case *ZodIntegerTyped[uint16, uint16]:
					result, err := s.Parse(test.input)
					require.NoError(t, err)
					assert.Equal(t, test.output, result)
				case *ZodIntegerTyped[uint32, uint32]:
					result, err := s.Parse(test.input)
					require.NoError(t, err)
					assert.Equal(t, test.output, result)
				}
			})
		}
	})
}

// =============================================================================
// Type-specific tests
// =============================================================================

func TestInt_TypeSpecific(t *testing.T) {
	t.Run("Int vs Int64 distinction", func(t *testing.T) {
		intSchema := Int()
		int64Schema := Int64()

		// Both should handle their respective types
		resultInt, err := intSchema.Parse(42)
		require.NoError(t, err)
		assert.IsType(t, int(0), resultInt)

		resultInt64, err := int64Schema.Parse(int64(42))
		require.NoError(t, err)
		assert.IsType(t, int64(0), resultInt64)
	})

	t.Run("Byte and Rune aliases", func(t *testing.T) {
		byteSchema := Byte()
		runeSchema := Rune()

		// Byte should behave like Uint8
		resultByte, err := byteSchema.Parse(uint8(255))
		require.NoError(t, err)
		assert.IsType(t, uint8(0), resultByte)
		assert.Equal(t, uint8(255), resultByte)

		// Rune should behave like Int32
		resultRune, err := runeSchema.Parse(int32(65))
		require.NoError(t, err)
		assert.IsType(t, int32(0), resultRune)
		assert.Equal(t, int32(65), resultRune)
	})

	t.Run("mixed integer types in validation chains", func(t *testing.T) {
		// Test that validation methods work correctly with different integer types
		int32Schema := Int32().Min(1).Max(10).Positive()
		uint64Schema := Uint64().Min(1).Max(10)

		// Both should validate successfully
		result32, err := int32Schema.Parse(int32(5))
		require.NoError(t, err)
		assert.Equal(t, int32(5), result32)

		resultUint64, err := uint64Schema.Parse(uint64(5))
		require.NoError(t, err)
		assert.Equal(t, uint64(5), resultUint64)

		// Both should fail validation
		_, err = int32Schema.Parse(int32(-1))
		assert.Error(t, err)

		_, err = uint64Schema.Parse(uint64(15))
		assert.Error(t, err)
	})

	t.Run("coerced integer type equivalence", func(t *testing.T) {
		CoercedIntegerSchema := CoercedInteger()
		CoercedInt64Schema := CoercedInt64()

		// Both should coerce strings successfully
		resultInt, err1 := CoercedIntegerSchema.Parse("42")
		require.NoError(t, err1)

		resultInt64, err2 := CoercedInt64Schema.Parse("42")
		require.NoError(t, err2)

		// Both CoercedInteger() and CoercedInt64() return int64
		assert.IsType(t, int64(0), resultInt)
		assert.IsType(t, int64(0), resultInt64)
		assert.Equal(t, int64(42), resultInt)
		assert.Equal(t, int64(42), resultInt64)
	})
}

// =============================================================================
// Pointer identity preservation tests
// =============================================================================

func TestInt_PointerIdentityPreservation(t *testing.T) {
	t.Run("Int Optional preserves pointer identity", func(t *testing.T) {
		schema := Int().Optional()

		// Int().Optional() returns *ZodIntegerTyped[int, *int], so we need to use int
		originalInt := int(42)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, int(42), *result)
	})

	t.Run("Int64 Optional preserves pointer identity", func(t *testing.T) {
		schema := Int64().Optional()

		originalInt := int64(123)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, int64(123), *result)
	})

	t.Run("Int64 Nilable preserves pointer identity", func(t *testing.T) {
		schema := Int64().Nilable()

		originalInt := int64(456)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, int64(456), *result)
	})

	t.Run("Int64Ptr Optional preserves pointer identity", func(t *testing.T) {
		schema := Int64Ptr().Optional()

		originalInt := int64(789)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Int64Ptr Optional should preserve pointer identity")
		assert.Equal(t, int64(789), *result)
	})

	t.Run("Int64Ptr Nilable preserves pointer identity", func(t *testing.T) {
		schema := Int64Ptr().Nilable()

		originalInt := int64(-123)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Int64Ptr Nilable should preserve pointer identity")
		assert.Equal(t, int64(-123), *result)
	})

	t.Run("Int64 Nullish preserves pointer identity", func(t *testing.T) {
		schema := Int64().Nullish()

		originalInt := int64(999)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Int64 Nullish should preserve pointer identity")
		assert.Equal(t, int64(999), *result)
	})

	t.Run("Optional handles nil consistently", func(t *testing.T) {
		schema := Int64().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable handles nil consistently", func(t *testing.T) {
		schema := Int64().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default().Optional() chaining preserves pointer identity", func(t *testing.T) {
		schema := Int64().Default(100).Optional()

		originalInt := int64(200)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Default().Optional() should preserve pointer identity")
		assert.Equal(t, int64(200), *result)
	})

	t.Run("Validation with Optional preserves pointer identity", func(t *testing.T) {
		schema := Int64().Min(0).Max(1000).Optional()

		originalInt := int64(500)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Validation().Optional() should preserve pointer identity")
		assert.Equal(t, int64(500), *result)
	})

	t.Run("Refine with Optional preserves pointer identity", func(t *testing.T) {
		schema := Int64().Refine(func(i int64) bool {
			return i > 0 // Only positive numbers
		}).Optional()

		originalInt := int64(42)
		originalPtr := &originalInt

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Refine().Optional() should preserve pointer identity")
		assert.Equal(t, int64(42), *result)
	})

	t.Run("Multiple integer types pointer identity", func(t *testing.T) {
		// Test int32 Optional returns *int32
		int32Schema := Int32().Optional()
		var _ *ZodIntegerTyped[int32, *int32] = int32Schema

		val32 := int32(123)
		ptr32 := &val32
		result32, err := int32Schema.Parse(ptr32)
		require.NoError(t, err)
		assert.True(t, result32 == ptr32, "Pointer identity should be preserved for int32")
		assert.Equal(t, int32(123), *result32)

		// Test int64 Optional returns *int64
		int64Schema := Int64().Optional()
		var _ *ZodIntegerTyped[int64, *int64] = int64Schema

		val64 := int64(456)
		ptr64 := &val64
		result64, err := int64Schema.Parse(ptr64)
		require.NoError(t, err)
		assert.True(t, result64 == ptr64, "Pointer identity should be preserved for int64")
		assert.Equal(t, int64(456), *result64)

		// Test uint32 Optional returns *uint32
		uint32Schema := Uint32().Optional()
		var _ *ZodIntegerTyped[uint32, *uint32] = uint32Schema

		valU32 := uint32(789)
		ptrU32 := &valU32
		resultU32, err := uint32Schema.Parse(ptrU32)
		require.NoError(t, err)
		assert.True(t, resultU32 == ptrU32, "Pointer identity should be preserved for uint32")
		assert.Equal(t, uint32(789), *resultU32)
	})
}

// =============================================================================
// Comprehensive Type Safety Tests for Modifier Methods
// =============================================================================

func TestInt_ComprehensiveTypeSafety(t *testing.T) {
	t.Run("Optional method type safety for all integer types", func(t *testing.T) {
		// Test int8 -> *int8
		int8Schema := Int8()
		int8Optional := int8Schema.Optional()
		var _ *ZodIntegerTyped[int8, *int8] = int8Optional

		result8, err := int8Optional.Parse(int8(42))
		require.NoError(t, err)
		assert.IsType(t, (*int8)(nil), result8)
		require.NotNil(t, result8)
		assert.Equal(t, int8(42), *result8)

		// Test int16 -> *int16
		int16Schema := Int16()
		int16Optional := int16Schema.Optional()
		var _ *ZodIntegerTyped[int16, *int16] = int16Optional

		result16, err := int16Optional.Parse(int16(1000))
		require.NoError(t, err)
		assert.IsType(t, (*int16)(nil), result16)
		require.NotNil(t, result16)
		assert.Equal(t, int16(1000), *result16)

		// Test int32 -> *int32
		int32Schema := Int32()
		int32Optional := int32Schema.Optional()
		var _ *ZodIntegerTyped[int32, *int32] = int32Optional

		result32, err := int32Optional.Parse(int32(100000))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result32)
		require.NotNil(t, result32)
		assert.Equal(t, int32(100000), *result32)

		// Test int64 -> *int64
		int64Schema := Int64()
		int64Optional := int64Schema.Optional()
		var _ *ZodIntegerTyped[int64, *int64] = int64Optional

		result64, err := int64Optional.Parse(int64(1000000))
		require.NoError(t, err)
		assert.IsType(t, (*int64)(nil), result64)
		require.NotNil(t, result64)
		assert.Equal(t, int64(1000000), *result64)

		// Test uint types
		uint8Schema := Uint8()
		uint8Optional := uint8Schema.Optional()
		var _ *ZodIntegerTyped[uint8, *uint8] = uint8Optional

		resultU8, err := uint8Optional.Parse(uint8(255))
		require.NoError(t, err)
		assert.IsType(t, (*uint8)(nil), resultU8)
		require.NotNil(t, resultU8)
		assert.Equal(t, uint8(255), *resultU8)

		uint16Schema := Uint16()
		uint16Optional := uint16Schema.Optional()
		var _ *ZodIntegerTyped[uint16, *uint16] = uint16Optional

		resultU16, err := uint16Optional.Parse(uint16(65535))
		require.NoError(t, err)
		assert.IsType(t, (*uint16)(nil), resultU16)
		require.NotNil(t, resultU16)
		assert.Equal(t, uint16(65535), *resultU16)

		uint32Schema := Uint32()
		uint32Optional := uint32Schema.Optional()
		var _ *ZodIntegerTyped[uint32, *uint32] = uint32Optional

		resultU32, err := uint32Optional.Parse(uint32(4294967295))
		require.NoError(t, err)
		assert.IsType(t, (*uint32)(nil), resultU32)
		require.NotNil(t, resultU32)
		assert.Equal(t, uint32(4294967295), *resultU32)

		uint64Schema := Uint64()
		uint64Optional := uint64Schema.Optional()
		var _ *ZodIntegerTyped[uint64, *uint64] = uint64Optional

		resultU64, err := uint64Optional.Parse(uint64(18446744073709551615))
		require.NoError(t, err)
		assert.IsType(t, (*uint64)(nil), resultU64)
		require.NotNil(t, resultU64)
		assert.Equal(t, uint64(18446744073709551615), *resultU64)

		// Test int and uint
		intSchema := Int()
		intOptional := intSchema.Optional()
		var _ *ZodIntegerTyped[int, *int] = intOptional

		resultInt, err := intOptional.Parse(int(42))
		require.NoError(t, err)
		assert.IsType(t, (*int)(nil), resultInt)
		require.NotNil(t, resultInt)
		assert.Equal(t, int(42), *resultInt)

		uintSchema := Uint()
		uintOptional := uintSchema.Optional()
		var _ *ZodIntegerTyped[uint, *uint] = uintOptional

		resultUint, err := uintOptional.Parse(uint(42))
		require.NoError(t, err)
		assert.IsType(t, (*uint)(nil), resultUint)
		require.NotNil(t, resultUint)
		assert.Equal(t, uint(42), *resultUint)
	})

	t.Run("Nilable method type safety for all integer types", func(t *testing.T) {
		// Test int8 -> *int8
		int8Schema := Int8()
		int8Nilable := int8Schema.Nilable()
		var _ *ZodIntegerTyped[int8, *int8] = int8Nilable

		result8, err := int8Nilable.Parse(int8(42))
		require.NoError(t, err)
		assert.IsType(t, (*int8)(nil), result8)
		require.NotNil(t, result8)
		assert.Equal(t, int8(42), *result8)

		// Test nil handling
		resultNil8, err := int8Nilable.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNil8)

		// Test int32 -> *int32
		int32Schema := Int32()
		int32Nilable := int32Schema.Nilable()
		var _ *ZodIntegerTyped[int32, *int32] = int32Nilable

		result32, err := int32Nilable.Parse(int32(100000))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result32)
		require.NotNil(t, result32)
		assert.Equal(t, int32(100000), *result32)

		// Test nil handling
		resultNil32, err := int32Nilable.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNil32)

		// Test uint64 -> *uint64
		uint64Schema := Uint64()
		uint64Nilable := uint64Schema.Nilable()
		var _ *ZodIntegerTyped[uint64, *uint64] = uint64Nilable

		resultU64, err := uint64Nilable.Parse(uint64(12345))
		require.NoError(t, err)
		assert.IsType(t, (*uint64)(nil), resultU64)
		require.NotNil(t, resultU64)
		assert.Equal(t, uint64(12345), *resultU64)

		// Test nil handling
		resultNilU64, err := uint64Nilable.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNilU64)
	})

	t.Run("Nullish method type safety for all integer types", func(t *testing.T) {
		// Test int16 -> *int16
		int16Schema := Int16()
		int16Nullish := int16Schema.Nullish()
		var _ *ZodIntegerTyped[int16, *int16] = int16Nullish

		result16, err := int16Nullish.Parse(int16(1000))
		require.NoError(t, err)
		assert.IsType(t, (*int16)(nil), result16)
		require.NotNil(t, result16)
		assert.Equal(t, int16(1000), *result16)

		// Test nil handling
		resultNil16, err := int16Nullish.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNil16)

		// Test uint32 -> *uint32
		uint32Schema := Uint32()
		uint32Nullish := uint32Schema.Nullish()
		var _ *ZodIntegerTyped[uint32, *uint32] = uint32Nullish

		resultU32, err := uint32Nullish.Parse(uint32(4000000))
		require.NoError(t, err)
		assert.IsType(t, (*uint32)(nil), resultU32)
		require.NotNil(t, resultU32)
		assert.Equal(t, uint32(4000000), *resultU32)

		// Test nil handling
		resultNilU32, err := uint32Nullish.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNilU32)
	})

	t.Run("Pointer schema modifier methods maintain pointer constraint", func(t *testing.T) {
		// Test Int64Ptr().Optional() maintains *int64
		int64PtrSchema := Int64Ptr()
		int64PtrOptional := int64PtrSchema.Optional()
		var _ *ZodIntegerTyped[int64, *int64] = int64PtrOptional

		// Test with value input (should convert to pointer)
		result, err := int64PtrOptional.Parse(int64(123))
		require.NoError(t, err)
		assert.IsType(t, (*int64)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int64(123), *result)

		// Test Int32Ptr().Nilable() maintains *int32
		int32PtrSchema := Int32Ptr()
		int32PtrNilable := int32PtrSchema.Nilable()
		var _ *ZodIntegerTyped[int32, *int32] = int32PtrNilable

		result32, err := int32PtrNilable.Parse(int32(456))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result32)
		require.NotNil(t, result32)
		assert.Equal(t, int32(456), *result32)

		// Test Uint16Ptr().Nullish() maintains *uint16
		uint16PtrSchema := Uint16Ptr()
		uint16PtrNullish := uint16PtrSchema.Nullish()
		var _ *ZodIntegerTyped[uint16, *uint16] = uint16PtrNullish

		resultU16, err := uint16PtrNullish.Parse(uint16(789))
		require.NoError(t, err)
		assert.IsType(t, (*uint16)(nil), resultU16)
		require.NotNil(t, resultU16)
		assert.Equal(t, uint16(789), *resultU16)
	})
}

func TestInt_TypeEvolutionChaining(t *testing.T) {
	t.Run("type evolution through method chaining", func(t *testing.T) {
		// Test: Int32() -> Default() -> Min() -> Optional()
		// Type evolution: [int32, int32] -> [int32, int32] -> [int32, int32] -> [int32, *int32]
		schema := Int32(). // *ZodIntegerTyped[int32, int32]
					Default(42). // *ZodIntegerTyped[int32, int32] (maintains type)
					Min(0).      // *ZodIntegerTyped[int32, int32] (maintains type)
					Optional()   // *ZodIntegerTyped[int32, *int32] (type conversion)

		var _ *ZodIntegerTyped[int32, *int32] = schema

		// Test with valid value
		result, err := schema.Parse(int32(100))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int32(100), *result)

		// Test with nil - Default should always win (Zod v4 behavior)
		resultNil, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, resultNil) // Default takes precedence over Optional
		assert.Equal(t, int32(42), *resultNil)
	})

	t.Run("complex validation chaining with type conversion", func(t *testing.T) {
		// Test: Uint64() -> Max() -> Positive() -> Nilable()
		// Type evolution: [uint64, uint64] -> [uint64, uint64] -> [uint64, uint64] -> [uint64, *uint64]
		schema := Uint64(). // *ZodIntegerTyped[uint64, uint64]
					Max(1000).  // *ZodIntegerTyped[uint64, uint64] (maintains type)
					Positive(). // *ZodIntegerTyped[uint64, uint64] (maintains type)
					Nilable()   // *ZodIntegerTyped[uint64, *uint64] (type conversion)

		var _ *ZodIntegerTyped[uint64, *uint64] = schema

		// Test with valid value
		result, err := schema.Parse(uint64(500))
		require.NoError(t, err)
		assert.IsType(t, (*uint64)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, uint64(500), *result)

		// Test with nil (should work for nilable)
		resultNil, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNil)

		// Test validation failure
		_, err = schema.Parse(uint64(2000)) // exceeds Max(1000)
		assert.Error(t, err)
	})

	t.Run("reverse type evolution: pointer to value schema", func(t *testing.T) {
		// Start with pointer schema, but modifiers that don't change constraint type
		// should maintain the pointer constraint
		schema := Int16Ptr(). // *ZodIntegerTyped[int16, *int16]
					Min(10).   // *ZodIntegerTyped[int16, *int16] (maintains constraint type)
					Max(100).  // *ZodIntegerTyped[int16, *int16] (maintains constraint type)
					Positive() // *ZodIntegerTyped[int16, *int16] (maintains constraint type)

		var _ *ZodIntegerTyped[int16, *int16] = schema

		// Test with pointer input
		val := int16(50)
		result, err := schema.Parse(&val)
		require.NoError(t, err)
		assert.IsType(t, (*int16)(nil), result)
		assert.True(t, result == &val, "Pointer identity should be preserved")

		// Test with value input (should convert to pointer)
		result2, err := schema.Parse(int16(75))
		require.NoError(t, err)
		assert.IsType(t, (*int16)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, int16(75), *result2)
	})

	t.Run("multiple modifier method applications", func(t *testing.T) {
		// Test applying multiple modifier methods that change constraint types
		originalSchema := Int8() // *ZodIntegerTyped[int8, int8]

		// First conversion: value to pointer
		optionalSchema := originalSchema.Optional() // *ZodIntegerTyped[int8, *int8]
		var _ *ZodIntegerTyped[int8, *int8] = optionalSchema

		// Second conversion: should maintain pointer constraint
		nilableSchema := optionalSchema.Nilable() // *ZodIntegerTyped[int8, *int8]
		var _ *ZodIntegerTyped[int8, *int8] = nilableSchema

		// Third conversion: should maintain pointer constraint
		nullishSchema := nilableSchema.Nullish() // *ZodIntegerTyped[int8, *int8]
		var _ *ZodIntegerTyped[int8, *int8] = nullishSchema

		// Test functionality
		result, err := nullishSchema.Parse(int8(42))
		require.NoError(t, err)
		assert.IsType(t, (*int8)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int8(42), *result)

		// Test nil handling
		resultNil, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, resultNil)
	})
}

func TestInt_CompilationTimeTypeSafety(t *testing.T) {
	t.Run("compile-time type assertions for all integer types", func(t *testing.T) {
		// These should all compile without issues - testing compile-time type safety

		// Base value schemas
		var _ *ZodIntegerTyped[int8, int8] = Int8()
		var _ *ZodIntegerTyped[int16, int16] = Int16()
		var _ *ZodIntegerTyped[int32, int32] = Int32()
		var _ *ZodIntegerTyped[int64, int64] = Int64()
		var _ *ZodIntegerTyped[int, int] = Int()
		var _ *ZodIntegerTyped[uint8, uint8] = Uint8()
		var _ *ZodIntegerTyped[uint16, uint16] = Uint16()
		var _ *ZodIntegerTyped[uint32, uint32] = Uint32()
		var _ *ZodIntegerTyped[uint64, uint64] = Uint64()
		var _ *ZodIntegerTyped[uint, uint] = Uint()

		// Base pointer schemas
		var _ *ZodIntegerTyped[int8, *int8] = Int8Ptr()
		var _ *ZodIntegerTyped[int16, *int16] = Int16Ptr()
		var _ *ZodIntegerTyped[int32, *int32] = Int32Ptr()
		var _ *ZodIntegerTyped[int64, *int64] = Int64Ptr()
		var _ *ZodIntegerTyped[int, *int] = IntPtr()
		var _ *ZodIntegerTyped[uint8, *uint8] = Uint8Ptr()
		var _ *ZodIntegerTyped[uint16, *uint16] = Uint16Ptr()
		var _ *ZodIntegerTyped[uint32, *uint32] = Uint32Ptr()
		var _ *ZodIntegerTyped[uint64, *uint64] = Uint64Ptr()
		var _ *ZodIntegerTyped[uint, *uint] = UintPtr()

		// Optional conversions (value -> pointer)
		var _ *ZodIntegerTyped[int8, *int8] = Int8().Optional()
		var _ *ZodIntegerTyped[int16, *int16] = Int16().Optional()
		var _ *ZodIntegerTyped[int32, *int32] = Int32().Optional()
		var _ *ZodIntegerTyped[int64, *int64] = Int64().Optional()
		var _ *ZodIntegerTyped[int, *int] = Int().Optional()
		var _ *ZodIntegerTyped[uint8, *uint8] = Uint8().Optional()
		var _ *ZodIntegerTyped[uint16, *uint16] = Uint16().Optional()
		var _ *ZodIntegerTyped[uint32, *uint32] = Uint32().Optional()
		var _ *ZodIntegerTyped[uint64, *uint64] = Uint64().Optional()
		var _ *ZodIntegerTyped[uint, *uint] = Uint().Optional()

		// Nilable conversions (value -> pointer)
		var _ *ZodIntegerTyped[int8, *int8] = Int8().Nilable()
		var _ *ZodIntegerTyped[int16, *int16] = Int16().Nilable()
		var _ *ZodIntegerTyped[int32, *int32] = Int32().Nilable()
		var _ *ZodIntegerTyped[int64, *int64] = Int64().Nilable()
		var _ *ZodIntegerTyped[int, *int] = Int().Nilable()
		var _ *ZodIntegerTyped[uint8, *uint8] = Uint8().Nilable()
		var _ *ZodIntegerTyped[uint16, *uint16] = Uint16().Nilable()
		var _ *ZodIntegerTyped[uint32, *uint32] = Uint32().Nilable()
		var _ *ZodIntegerTyped[uint64, *uint64] = Uint64().Nilable()
		var _ *ZodIntegerTyped[uint, *uint] = Uint().Nilable()

		// Nullish conversions (value -> pointer)
		var _ *ZodIntegerTyped[int8, *int8] = Int8().Nullish()
		var _ *ZodIntegerTyped[int16, *int16] = Int16().Nullish()
		var _ *ZodIntegerTyped[int32, *int32] = Int32().Nullish()
		var _ *ZodIntegerTyped[int64, *int64] = Int64().Nullish()
		var _ *ZodIntegerTyped[int, *int] = Int().Nullish()
		var _ *ZodIntegerTyped[uint8, *uint8] = Uint8().Nullish()
		var _ *ZodIntegerTyped[uint16, *uint16] = Uint16().Nullish()
		var _ *ZodIntegerTyped[uint32, *uint32] = Uint32().Nullish()
		var _ *ZodIntegerTyped[uint64, *uint64] = Uint64().Nullish()
		var _ *ZodIntegerTyped[uint, *uint] = Uint().Nullish()

		// Pointer schema modifier methods (pointer -> pointer)
		var _ *ZodIntegerTyped[int8, *int8] = Int8Ptr().Optional()
		var _ *ZodIntegerTyped[int16, *int16] = Int16Ptr().Nilable()
		var _ *ZodIntegerTyped[int32, *int32] = Int32Ptr().Nullish()
		var _ *ZodIntegerTyped[int64, *int64] = Int64Ptr().Optional()
		var _ *ZodIntegerTyped[uint8, *uint8] = Uint8Ptr().Nilable()
		var _ *ZodIntegerTyped[uint16, *uint16] = Uint16Ptr().Nullish()
		var _ *ZodIntegerTyped[uint32, *uint32] = Uint32Ptr().Optional()
		var _ *ZodIntegerTyped[uint64, *uint64] = Uint64Ptr().Nilable()

		// Complex chaining type assertions
		var _ *ZodIntegerTyped[int32, *int32] = Int32().Default(42).Optional()
		var _ *ZodIntegerTyped[uint64, *uint64] = Uint64().Min(0).Max(1000).Nilable()
		var _ *ZodIntegerTyped[int16, *int16] = Int16().Positive().Nullish()
		var _ *ZodIntegerTyped[int8, *int8] = Int8().NonNegative().Step(2).Optional()

		// Validation method chaining maintains constraint type
		var _ *ZodIntegerTyped[int32, int32] = Int32().Min(0).Max(100).Positive()
		var _ *ZodIntegerTyped[uint64, uint64] = Uint64().Max(1000).NonNegative()
		var _ *ZodIntegerTyped[int16, *int16] = Int16Ptr().Min(-100).Max(100)
		var _ *ZodIntegerTyped[uint8, *uint8] = Uint8Ptr().Max(255).MultipleOf(5)
	})

	t.Run("method chaining preserves type safety", func(t *testing.T) {
		// Test that complex method chaining maintains proper types

		// Value schema with validation chain then conversion
		schema1 := Int32().Min(0).Max(1000).Positive().Optional()
		var _ *ZodIntegerTyped[int32, *int32] = schema1

		// Pointer schema with validation chain
		schema2 := Int64Ptr().Min(-1000).Max(1000).NonNegative()
		var _ *ZodIntegerTyped[int64, *int64] = schema2

		// Multiple conversions
		schema3 := Int16().Optional().Nilable().Nullish()
		var _ *ZodIntegerTyped[int16, *int16] = schema3

		// Test that they all work functionally
		result1, err1 := schema1.Parse(int32(500))
		require.NoError(t, err1)
		assert.IsType(t, (*int32)(nil), result1)
		require.NotNil(t, result1)
		assert.Equal(t, int32(500), *result1)

		result2, err2 := schema2.Parse(int64(500))
		require.NoError(t, err2)
		assert.IsType(t, (*int64)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, int64(500), *result2)

		result3, err3 := schema3.Parse(int16(100))
		require.NoError(t, err3)
		assert.IsType(t, (*int16)(nil), result3)
		require.NotNil(t, result3)
		assert.Equal(t, int16(100), *result3)
	})
}

func TestInt_ErrorHandlingWithTypeSafety(t *testing.T) {
	t.Run("type mismatch errors with modifier methods", func(t *testing.T) {
		// Int32 Optional should reject other integer types
		schema := Int32().Optional()

		// Should reject int64
		_, err := schema.Parse(int64(42))
		assert.Error(t, err)

		// Should reject int8
		_, err = schema.Parse(int8(42))
		assert.Error(t, err)

		// Should reject uint32
		_, err = schema.Parse(uint32(42))
		assert.Error(t, err)

		// Should accept int32
		result, err := schema.Parse(int32(42))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int32(42), *result)
	})

	t.Run("validation errors maintain type safety", func(t *testing.T) {
		// Create schema with validation that will fail
		schema := Int64().Min(100).Optional()

		// Validation failure should still maintain type information
		_, err := schema.Parse(int64(50)) // below minimum
		assert.Error(t, err)

		// Successful validation should return correct type
		result, err := schema.Parse(int64(150))
		require.NoError(t, err)
		assert.IsType(t, (*int64)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int64(150), *result)
	})

	t.Run("refine errors with type safety", func(t *testing.T) {
		// Create refine schema with type conversion
		schema := Int32().Refine(func(i int32) bool {
			return i%2 == 0 // only even numbers
		}).Optional()

		// Refine failure
		_, err := schema.Parse(int32(43)) // odd number
		assert.Error(t, err)

		// Refine success with correct type
		result, err := schema.Parse(int32(44))
		require.NoError(t, err)
		assert.IsType(t, (*int32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int32(44), *result)
	})
}

func TestInt_MustParseTypeSafety(t *testing.T) {
	t.Run("MustParse maintains type safety with modifier methods", func(t *testing.T) {
		// Test Optional MustParse
		optionalSchema := Int32().Optional()
		result := optionalSchema.MustParse(int32(42))
		assert.IsType(t, (*int32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, int32(42), *result)

		// Test nil handling
		resultNil := optionalSchema.MustParse(nil)
		assert.Nil(t, resultNil)

		// Test Nilable MustParse
		nilableSchema := Int64().Nilable()
		result64 := nilableSchema.MustParse(int64(123))
		assert.IsType(t, (*int64)(nil), result64)
		require.NotNil(t, result64)
		assert.Equal(t, int64(123), *result64)

		// Test Nullish MustParse
		nullishSchema := Uint16().Nullish()
		resultU16 := nullishSchema.MustParse(uint16(456))
		assert.IsType(t, (*uint16)(nil), resultU16)
		require.NotNil(t, resultU16)
		assert.Equal(t, uint16(456), *resultU16)
	})

	t.Run("MustParse panics maintain type safety context", func(t *testing.T) {
		// Even when panicking, the schema type information should be clear
		optionalSchema := Int32().Optional()

		// This should panic, but the schema is correctly typed
		assert.Panics(t, func() {
			optionalSchema.MustParse("not an integer")
		})

		// Verify the schema type is still correct after panic recovery
		var _ *ZodIntegerTyped[int32, *int32] = optionalSchema
	})
}

func TestInt_Check(t *testing.T) {
	t.Run("adds multiple issues for invalid input", func(t *testing.T) {
		schema := Int64().Check(func(value int64, p *core.ParsePayload) {
			if value < 0 {
				p.AddIssueWithMessage("value must be non-negative")
			}
			if value%2 != 0 {
				p.AddIssueWithCode(core.NotMultipleOf, "value must be even")
			}
		})

		_, err := schema.Parse(int64(-3))
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		require.Len(t, zodErr.Issues, 2)

		assert.Equal(t, "value must be non-negative", zodErr.Issues[0].Message)
		assert.Equal(t, core.Custom, zodErr.Issues[0].Code)
		assert.Equal(t, core.NotMultipleOf, zodErr.Issues[1].Code)
	})

	t.Run("succeeds for valid input", func(t *testing.T) {
		schema := Int64().Check(func(value int64, p *core.ParsePayload) {
			if value < 0 {
				p.AddIssueWithMessage("negative")
			}
		})

		result, err := schema.Parse(int64(10))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)
	})

	t.Run("works with pointer types", func(t *testing.T) {
		schema := Int64Ptr().Check(func(value *int64, p *core.ParsePayload) {
			if value != nil && *value < 100 {
				p.AddIssueWithMessage("too small")
			}
		})

		small := int64(10)
		_, err := schema.Parse(&small)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)

		big := int64(200)
		res, err := schema.Parse(&big)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, int64(200), *res)
	})
}

// =============================================================================
// NonOptional tests
// =============================================================================

func TestInt_NonOptional(t *testing.T) {
	// --- Basic behaviour ---
	schema := Int64().NonOptional()

	// Valid input
	res, err := schema.Parse(int64(123))
	require.NoError(t, err)
	assert.Equal(t, int64(123), res)
	assert.IsType(t, int64(0), res)

	// Nil should error with nonoptional expected tag
	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		require.Len(t, zErr.Issues, 1)
		assert.Equal(t, core.InvalidType, zErr.Issues[0].Code)
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}

	// --- Optional().NonOptional() chain ---
	chain := Int64().Optional().NonOptional()
	var _ *ZodIntegerTyped[int64, int64] = chain

	res2, err := chain.Parse(int64(55))
	require.NoError(t, err)
	assert.Equal(t, int64(55), res2)

	_, err = chain.Parse(nil)
	assert.Error(t, err)

	// --- Object embedding ---
	objSchema := Object(map[string]core.ZodSchema{
		"num": Int64().Optional().NonOptional(),
	})

	// valid
	_, err = objSchema.Parse(map[string]any{"num": int64(99)})
	require.NoError(t, err)

	// nil
	_, err = objSchema.Parse(map[string]any{"num": nil})
	assert.Error(t, err)

	// missing
	_, err = objSchema.Parse(map[string]any{})
	assert.Error(t, err)

	// --- Int64Ptr().NonOptional() ---
	val := int64(444)
	ptrSchema := Int64Ptr().NonOptional()
	var _ *ZodIntegerTyped[int64, int64] = ptrSchema

	r1, err := ptrSchema.Parse(&val)
	require.NoError(t, err)
	assert.Equal(t, int64(444), r1)

	r2, err := ptrSchema.Parse(int64(333))
	require.NoError(t, err)
	assert.Equal(t, int64(333), r2)

	_, err = ptrSchema.Parse(nil)
	assert.Error(t, err)
}

// =============================================================================
// IsOptional and IsNilable tests
// =============================================================================

func TestInt_IsOptionalAndIsNilable(t *testing.T) {
	t.Run("basic schema - not optional, not nilable", func(t *testing.T) {
		schema := Int()

		assert.False(t, schema.IsOptional(), "Basic int schema should not be optional")
		assert.False(t, schema.IsNilable(), "Basic int schema should not be nilable")
	})

	t.Run("optional schema - is optional, not nilable", func(t *testing.T) {
		schema := Int().Optional()

		assert.True(t, schema.IsOptional(), "Optional int schema should be optional")
		assert.False(t, schema.IsNilable(), "Optional int schema should not be nilable")
	})

	t.Run("nilable schema - not optional, is nilable", func(t *testing.T) {
		schema := Int().Nilable()

		assert.False(t, schema.IsOptional(), "Nilable int schema should not be optional")
		assert.True(t, schema.IsNilable(), "Nilable int schema should be nilable")
	})

	t.Run("nullish schema - is optional and nilable", func(t *testing.T) {
		schema := Int().Nullish()

		assert.True(t, schema.IsOptional(), "Nullish int schema should be optional")
		assert.True(t, schema.IsNilable(), "Nullish int schema should be nilable")
	})

	t.Run("chained modifiers", func(t *testing.T) {
		// Optional then Nilable
		schema1 := Int().Optional().Nilable()
		assert.True(t, schema1.IsOptional(), "Optional().Nilable() should be optional")
		assert.True(t, schema1.IsNilable(), "Optional().Nilable() should be nilable")

		// Nilable then Optional
		schema2 := Int().Nilable().Optional()
		assert.True(t, schema2.IsOptional(), "Nilable().Optional() should be optional")
		assert.True(t, schema2.IsNilable(), "Nilable().Optional() should be nilable")
	})

	t.Run("nonoptional modifier resets optional flag", func(t *testing.T) {
		schema := Int().Optional().NonOptional()

		assert.False(t, schema.IsOptional(), "Optional().NonOptional() should not be optional")
		assert.False(t, schema.IsNilable(), "Optional().NonOptional() should not be nilable")
	})

	t.Run("different integer types", func(t *testing.T) {
		// Test various integer types
		testCases := []struct {
			name   string
			schema any
		}{
			{"Int8", Int8()},
			{"Int16", Int16()},
			{"Int32", Int32()},
			{"Int64", Int64()},
			{"Uint8", Uint8()},
			{"Uint16", Uint16()},
			{"Uint32", Uint32()},
			{"Uint64", Uint64()},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Use reflection to call IsOptional and IsNilable methods
				schemaValue := reflect.ValueOf(tc.schema)

				isOptionalMethod := schemaValue.MethodByName("IsOptional")
				isNilableMethod := schemaValue.MethodByName("IsNilable")

				require.True(t, isOptionalMethod.IsValid(), "IsOptional method should exist")
				require.True(t, isNilableMethod.IsValid(), "IsNilable method should exist")

				isOptionalResult := isOptionalMethod.Call(nil)[0].Bool()
				isNilableResult := isNilableMethod.Call(nil)[0].Bool()

				assert.False(t, isOptionalResult, "%s schema should not be optional", tc.name)
				assert.False(t, isNilableResult, "%s schema should not be nilable", tc.name)
			})
		}
	})

	t.Run("pointer types", func(t *testing.T) {
		// IntPtr basic
		ptrSchema := IntPtr()
		assert.False(t, ptrSchema.IsOptional(), "IntPtr schema should not be optional")
		assert.False(t, ptrSchema.IsNilable(), "IntPtr schema should not be nilable")

		// IntPtr with modifiers
		optionalPtrSchema := IntPtr().Optional()
		assert.True(t, optionalPtrSchema.IsOptional(), "IntPtr().Optional() should be optional")
		assert.False(t, optionalPtrSchema.IsNilable(), "IntPtr().Optional() should not be nilable")

		nilablePtrSchema := IntPtr().Nilable()
		assert.False(t, nilablePtrSchema.IsOptional(), "IntPtr().Nilable() should not be optional")
		assert.True(t, nilablePtrSchema.IsNilable(), "IntPtr().Nilable() should be nilable")
	})

	t.Run("consistency with GetInternals", func(t *testing.T) {
		// Test basic int schema
		basicSchema := Int()
		assert.Equal(t, basicSchema.GetInternals().IsOptional(), basicSchema.IsOptional(),
			"Basic schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, basicSchema.GetInternals().IsNilable(), basicSchema.IsNilable(),
			"Basic schema: IsNilable() should match GetInternals().IsNilable()")

		// Test optional int schema
		optionalSchema := Int().Optional()
		assert.Equal(t, optionalSchema.GetInternals().IsOptional(), optionalSchema.IsOptional(),
			"Optional schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, optionalSchema.GetInternals().IsNilable(), optionalSchema.IsNilable(),
			"Optional schema: IsNilable() should match GetInternals().IsNilable()")

		// Test nilable int schema
		nilableSchema := Int().Nilable()
		assert.Equal(t, nilableSchema.GetInternals().IsOptional(), nilableSchema.IsOptional(),
			"Nilable schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, nilableSchema.GetInternals().IsNilable(), nilableSchema.IsNilable(),
			"Nilable schema: IsNilable() should match GetInternals().IsNilable()")

		// Test nullish int schema
		nullishSchema := Int().Nullish()
		assert.Equal(t, nullishSchema.GetInternals().IsOptional(), nullishSchema.IsOptional(),
			"Nullish schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, nullishSchema.GetInternals().IsNilable(), nullishSchema.IsNilable(),
			"Nullish schema: IsNilable() should match GetInternals().IsNilable()")

		// Test nonoptional int schema
		nonoptionalSchema := Int().Optional().NonOptional()
		assert.Equal(t, nonoptionalSchema.GetInternals().IsOptional(), nonoptionalSchema.IsOptional(),
			"NonOptional schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, nonoptionalSchema.GetInternals().IsNilable(), nonoptionalSchema.IsNilable(),
			"NonOptional schema: IsNilable() should match GetInternals().IsNilable()")
	})
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestInt_StrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		// Test different integer types
		schemaInt := Int()
		result, err := schemaInt.StrictParse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.IsType(t, 0, result)

		schemaInt64 := Int64()
		result64, err := schemaInt64.StrictParse(int64(123))
		require.NoError(t, err)
		assert.Equal(t, int64(123), result64)
		assert.IsType(t, int64(0), result64)

		schemaUint32 := Uint32()
		resultUint32, err := schemaUint32.StrictParse(uint32(456))
		require.NoError(t, err)
		assert.Equal(t, uint32(456), resultUint32)
		assert.IsType(t, uint32(0), resultUint32)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := Int().Min(10).Max(100)

		// Valid case
		result, err := schema.StrictParse(50)
		require.NoError(t, err)
		assert.Equal(t, 50, result)

		// Invalid case - below minimum
		_, err = schema.StrictParse(5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected integer to be at least 10")

		// Invalid case - above maximum
		_, err = schema.StrictParse(150)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too big: expected integer to be at most 100")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := IntPtr()
		value := 789

		// Test with valid pointer input
		result, err := schema.StrictParse(&value)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 789, *result)
		assert.IsType(t, (*int)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := IntPtr().Default(999)
		var nilPtr *int = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 999, *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		// Test case 1: Valid prefault value passes validation
		schema1 := IntPtr().Refine(func(i int) bool {
			return i > 0 // Only allow positive values
		}, "Must be positive").Prefault(10)
		negativeValue := -5

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema1.StrictParse(&negativeValue)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Must be positive")

		// Test with nil input (should use valid prefault)
		result1, err := schema1.StrictParse(nil)
		require.NoError(t, err)
		require.NotNil(t, result1)
		assert.Equal(t, 10, *result1) // Valid prefault value

		// Test case 2: Invalid prefault value fails validation
		schema2 := IntPtr().Refine(func(i int) bool {
			return i > 0 // Only allow positive values
		}, "Must be positive").Prefault(-1)

		// Test with nil input (should fail prefault validation)
		_, err2 := schema2.StrictParse(nil)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "Must be positive") // Prefault fails validation
	})

	t.Run("all integer types", func(t *testing.T) {
		// Test int8
		schemaInt8 := Int8()
		resultInt8, err := schemaInt8.StrictParse(int8(127))
		require.NoError(t, err)
		assert.Equal(t, int8(127), resultInt8)

		// Test int16
		schemaInt16 := Int16()
		resultInt16, err := schemaInt16.StrictParse(int16(32767))
		require.NoError(t, err)
		assert.Equal(t, int16(32767), resultInt16)

		// Test uint8
		schemaUint8 := Uint8()
		resultUint8, err := schemaUint8.StrictParse(uint8(255))
		require.NoError(t, err)
		assert.Equal(t, uint8(255), resultUint8)

		// Test uint16
		schemaUint16 := Uint16()
		resultUint16, err := schemaUint16.StrictParse(uint16(65535))
		require.NoError(t, err)
		assert.Equal(t, uint16(65535), resultUint16)

		// Test uint64
		schemaUint64 := Uint64()
		resultUint64, err := schemaUint64.StrictParse(uint64(18446744073709551615))
		require.NoError(t, err)
		assert.Equal(t, uint64(18446744073709551615), resultUint64)
	})
}

func TestInt_MustStrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		// Test different integer types
		schemaInt := Int()
		result := schemaInt.MustStrictParse(42)
		assert.Equal(t, 42, result)
		assert.IsType(t, 0, result)

		schemaInt64 := Int64()
		result64 := schemaInt64.MustStrictParse(int64(123))
		assert.Equal(t, int64(123), result64)
		assert.IsType(t, int64(0), result64)

		schemaUint32 := Uint32()
		resultUint32 := schemaUint32.MustStrictParse(uint32(456))
		assert.Equal(t, uint32(456), resultUint32)
		assert.IsType(t, uint32(0), resultUint32)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := Int().Min(10).Max(100)

		// Test panic with validation failure
		assert.Panics(t, func() {
			schema.MustStrictParse(5) // Below minimum, should panic
		})

		assert.Panics(t, func() {
			schema.MustStrictParse(150) // Above maximum, should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := IntPtr()
		value := 789

		// Test with valid pointer input
		result := schema.MustStrictParse(&value)
		require.NotNil(t, result)
		assert.Equal(t, 789, *result)
		assert.IsType(t, (*int)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := IntPtr().Default(999)
		var nilPtr *int = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, 999, *result)
	})

	t.Run("all integer types", func(t *testing.T) {
		// Test various integer types
		assert.Equal(t, int8(-128), Int8().MustStrictParse(int8(-128)))
		assert.Equal(t, int16(-32768), Int16().MustStrictParse(int16(-32768)))
		assert.Equal(t, int32(-2147483648), Int32().MustStrictParse(int32(-2147483648)))
		assert.Equal(t, int64(-9223372036854775808), Int64().MustStrictParse(int64(-9223372036854775808)))
		assert.Equal(t, uint(0), Uint().MustStrictParse(uint(0)))
		assert.Equal(t, uint8(0), Uint8().MustStrictParse(uint8(0)))
		assert.Equal(t, uint16(0), Uint16().MustStrictParse(uint16(0)))
		assert.Equal(t, uint32(0), Uint32().MustStrictParse(uint32(0)))
		assert.Equal(t, uint64(0), Uint64().MustStrictParse(uint64(0)))
	})

	t.Run("edge values", func(t *testing.T) {
		// Test boundary values for different types
		assert.Equal(t, int8(127), Int8().MustStrictParse(int8(127)))
		assert.Equal(t, int8(-128), Int8().MustStrictParse(int8(-128)))
		assert.Equal(t, uint8(255), Uint8().MustStrictParse(uint8(255)))
		assert.Equal(t, uint8(0), Uint8().MustStrictParse(uint8(0)))

		assert.Equal(t, int16(32767), Int16().MustStrictParse(int16(32767)))
		assert.Equal(t, int16(-32768), Int16().MustStrictParse(int16(-32768)))
		assert.Equal(t, uint16(65535), Uint16().MustStrictParse(uint16(65535)))
		assert.Equal(t, uint16(0), Uint16().MustStrictParse(uint16(0)))
	})
}
