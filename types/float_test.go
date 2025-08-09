package types

import (
	"math"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestFloat_BasicFunctionality(t *testing.T) {
	t.Run("valid float inputs", func(t *testing.T) {
		schema := Float64()

		// Test positive value
		result, err := schema.Parse(float64(42.5))
		require.NoError(t, err)
		assert.Equal(t, float64(42.5), result)

		// Test negative value
		result, err = schema.Parse(float64(-10.5))
		require.NoError(t, err)
		assert.Equal(t, float64(-10.5), result)

		// Test zero
		result, err = schema.Parse(float64(0))
		require.NoError(t, err)
		assert.Equal(t, float64(0), result)
	})

	t.Run("different float types", func(t *testing.T) {
		// Test float32
		schema32 := Float32()
		result32, err := schema32.Parse(float32(100.5))
		require.NoError(t, err)
		assert.Equal(t, float32(100.5), result32)

		// Test float64
		schema64 := Float64()
		result64, err := schema64.Parse(float64(200.5))
		require.NoError(t, err)
		assert.Equal(t, float64(200.5), result64)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Float64()

		invalidInputs := []any{
			"not a number", true, []float64{1.0}, map[string]float64{"key": 1.0}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Float64()

		// Test Parse method
		result, err := schema.Parse(float64(42.5))
		require.NoError(t, err)
		assert.Equal(t, float64(42.5), result)

		// Test MustParse method
		mustResult := schema.MustParse(float64(42.5))
		assert.Equal(t, float64(42.5), mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a float value"
		schema := Float64(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeFloat64, schema.internals.Def.ZodTypeDef.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestFloat_TypeSafety(t *testing.T) {
	t.Run("Float32 returns float32 type", func(t *testing.T) {
		schema := Float32()
		require.NotNil(t, schema)

		result, err := schema.Parse(float32(100.5))
		require.NoError(t, err)
		assert.Equal(t, float32(100.5), result)
		assert.IsType(t, float32(0), result)
	})

	t.Run("Float64 returns float64 type", func(t *testing.T) {
		schema := Float64()
		require.NotNil(t, schema)

		result, err := schema.Parse(float64(200.5))
		require.NoError(t, err)
		assert.Equal(t, float64(200.5), result)
		assert.IsType(t, float64(0), result)
	})

	t.Run("Float32Ptr returns *float32 type", func(t *testing.T) {
		schema := Float32Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(float32(100.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float32(100.5), *result)
		assert.IsType(t, (*float32)(nil), result)
	})

	t.Run("Float64Ptr returns *float64 type", func(t *testing.T) {
		schema := Float64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(float64(200.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(200.5), *result)
		assert.IsType(t, (*float64)(nil), result)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test float32 type
		float32Schema := Float32()
		resultFloat32 := float32Schema.MustParse(float32(100.5))
		assert.IsType(t, float32(0), resultFloat32)
		assert.Equal(t, float32(100.5), resultFloat32)

		// Test float64 type
		float64Schema := Float64()
		resultFloat64 := float64Schema.MustParse(float64(200.5))
		assert.IsType(t, float64(0), resultFloat64)
		assert.Equal(t, float64(200.5), resultFloat64)

		// Test *float64 type
		ptrSchema := Float64Ptr()
		ptrResult := ptrSchema.MustParse(float64(200.5))
		assert.IsType(t, (*float64)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, float64(200.5), *ptrResult)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestFloat_Modifiers(t *testing.T) {
	t.Run("Float32 Optional returns *float32 (type safe)", func(t *testing.T) {
		// Float32().Optional() now returns *ZodFloatTyped[float32, *float32]
		float32Schema := Float32()
		optionalSchema := float32Schema.Optional()

		// Type check: ensure it returns *ZodFloatTyped[float32, *float32]
		var _ *ZodFloatTyped[float32, *float32] = optionalSchema

		// Functionality test with float32
		result, err := optionalSchema.Parse(float32(100.5))
		require.NoError(t, err)
		assert.IsType(t, (*float32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, float32(100.5), *result)

		// From *float32 to *float32 via Optional (preserves type)
		ptrSchema := Float32Ptr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ *ZodFloatTyped[float32, *float32] = optionalPtrSchema
	})

	t.Run("Float32 Nilable returns *float32 (type safe)", func(t *testing.T) {
		float32Schema := Float32()
		nilableSchema := float32Schema.Nilable()

		var _ *ZodFloatTyped[float32, *float32] = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Float64 Optional returns *float64 (type safe)", func(t *testing.T) {
		float64Schema := Float64()
		optionalSchema := float64Schema.Optional()

		var _ *ZodFloatTyped[float64, *float64] = optionalSchema

		// Test functionality
		result, err := optionalSchema.Parse(float64(200.5))
		require.NoError(t, err)
		assert.IsType(t, (*float64)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, float64(200.5), *result)
	})

	t.Run("Default preserves current constraint type", func(t *testing.T) {
		// float32 maintains float32, float32
		float32Schema := Float32()
		defaultFloat32Schema := float32Schema.Default(100.5)
		var _ *ZodFloatTyped[float32, float32] = defaultFloat32Schema

		// *float64 maintains float64, *float64
		ptrSchema := Float64Ptr()
		defaultPtrSchema := ptrSchema.Default(200.5)
		var _ *ZodFloatTyped[float64, *float64] = defaultPtrSchema
	})

	t.Run("Prefault preserves current constraint type", func(t *testing.T) {
		// float32 maintains float32, float32
		float32Schema := Float32()
		prefaultFloat32Schema := float32Schema.Prefault(100.5)
		var _ *ZodFloatTyped[float32, float32] = prefaultFloat32Schema

		// *float64 maintains float64, *float64
		ptrSchema := Float64Ptr()
		prefaultPtrSchema := ptrSchema.Prefault(200.5)
		var _ *ZodFloatTyped[float64, *float64] = prefaultPtrSchema
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestFloat_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type preservation
		schema := Float32(). // *ZodFloatTyped[float32, float32]
					Default(100.5). // *ZodFloatTyped[float32, float32] (maintains constraint type)
					Optional()      // *ZodFloatTyped[float32, *float32] (changes to pointer constraint)

		var _ *ZodFloatTyped[float32, *float32] = schema

		// Test final behavior
		result, err := schema.Parse(float32(200.5))
		require.NoError(t, err)
		assert.IsType(t, (*float32)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, float32(200.5), *result)
	})

	t.Run("complex chaining preserves base type", func(t *testing.T) {
		schema := Float32Ptr(). // *ZodFloatTyped[float32, *float32]
					Nilable().     // *ZodFloatTyped[float32, *float32] (maintains both types)
					Default(100.5) // *ZodFloatTyped[float32, *float32] (maintains constraint type)

		var _ *ZodFloatTyped[float32, *float32] = schema

		result, err := schema.Parse(float32(200.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float32(200.5), *result)
		assert.IsType(t, (*float32)(nil), result)
	})

	t.Run("validation chaining", func(t *testing.T) {
		schema := Float64().
			Min(0).
			Max(100).
			Positive()

		result, err := schema.Parse(float64(50.5))
		require.NoError(t, err)
		assert.Equal(t, float64(50.5), result)

		// Should fail validation
		_, err = schema.Parse(float64(-5.5))
		assert.Error(t, err)

		_, err = schema.Parse(float64(150.5))
		assert.Error(t, err)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Float64().
			Default(200.5).
			Prefault(100.5)

		result, err := schema.Parse(float64(100.5))
		require.NoError(t, err)
		assert.Equal(t, float64(100.5), result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestFloat_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// Float64 type
		schema1 := Float64().Default(100.5).Prefault(200.5)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 100.5, result1) // Should be default, not prefault

		// Float64Ptr type
		schema2 := Float64Ptr().Default(100.5).Prefault(200.5)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, 100.5, *result2) // Should be default, not prefault
	})

	// Test 2: Default short-circuits validation
	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Float64 type - default value violates Min(50.0) but should still work
		schema1 := Float64().Min(50.0).Default(10.0)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 10.0, result1) // Default bypasses validation

		// Float64Ptr type - default value violates Min(50.0) but should still work
		schema2 := Float64Ptr().Min(50.0).Default(10.0)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, 10.0, *result2) // Default bypasses validation
	})

	// Test 3: Prefault goes through full validation
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Float64 type - prefault value passes validation
		schema1 := Float64().Min(50.0).Prefault(100.0)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 100.0, result1) // Prefault passes validation

		// Float64 type - prefault value fails validation
		schema2 := Float64().Min(50.0).Prefault(10.0)
		_, err2 := schema2.Parse(nil)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "Too small") // Prefault fails validation

		// Float64Ptr type - prefault value passes validation
		schema3 := Float64Ptr().Min(50.0).Prefault(100.0)
		result3, err3 := schema3.Parse(nil)
		require.NoError(t, err3)
		require.NotNil(t, result3)
		assert.Equal(t, 100.0, *result3) // Prefault passes validation

		// Float64Ptr type - prefault value fails validation
		schema4 := Float64Ptr().Min(50.0).Prefault(10.0)
		_, err4 := schema4.Parse(nil)
		require.Error(t, err4)
		assert.Contains(t, err4.Error(), "Too small") // Prefault fails validation
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		// Float64 type
		schema1 := Float64().Min(50.0).Prefault(100.0)

		// Nil input triggers prefault
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 100.0, result1)

		// Non-nil input validation failure should not trigger prefault
		_, err2 := schema1.Parse(10.0) // Less than Min(50.0)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "Too small")

		// Float64Ptr type
		schema2 := Float64Ptr().Min(50.0).Prefault(100.0)

		// Nil input triggers prefault
		result3, err3 := schema2.Parse(nil)
		require.NoError(t, err3)
		require.NotNil(t, result3)
		assert.Equal(t, 100.0, *result3)

		// Non-nil input validation failure should not trigger prefault
		_, err4 := schema2.Parse(10.0) // Less than Min(50.0)
		require.Error(t, err4)
		assert.Contains(t, err4.Error(), "Too small")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		// DefaultFunc should not be called for non-nil input
		callCount := 0
		schema1 := Float64().DefaultFunc(func() float64 {
			callCount++
			return 100.0
		})

		// Non-nil input should not call DefaultFunc
		result1, err1 := schema1.Parse(50.0)
		require.NoError(t, err1)
		assert.Equal(t, 50.0, result1)
		assert.Equal(t, 0, callCount)

		// Nil input should call DefaultFunc
		result2, err2 := schema1.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, 100.0, result2)
		assert.Equal(t, 1, callCount)

		// PrefaultFunc should go through validation
		schema2 := Float64().Min(50.0).PrefaultFunc(func() float64 {
			return 100.0 // Valid value
		})
		result3, err3 := schema2.Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, 100.0, result3)

		// PrefaultFunc with invalid value should fail
		schema3 := Float64().Min(50.0).PrefaultFunc(func() float64 {
			return 10.0 // Invalid value
		})
		_, err4 := schema3.Parse(nil)
		require.Error(t, err4)
		assert.Contains(t, err4.Error(), "Too small")
	})

	// Test 6: Error handling - prefault validation failure
	t.Run("Error handling - prefault validation failure", func(t *testing.T) {
		// Prefault value fails validation, should return error directly
		schema := Float64().Min(100.0).Prefault(50.0)
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
		// Should not try any other fallback
	})
}

// =============================================================================
// Validation tests
// =============================================================================

func TestFloat_Validations(t *testing.T) {
	t.Run("min validation", func(t *testing.T) {
		schema := Float64().Min(5.0)

		// Valid input
		result, err := schema.Parse(float64(10.5))
		require.NoError(t, err)
		assert.Equal(t, float64(10.5), result)

		// Invalid input
		_, err = schema.Parse(float64(3.5))
		assert.Error(t, err)
	})

	t.Run("max validation", func(t *testing.T) {
		schema := Float64().Max(10.0)

		// Valid input
		result, err := schema.Parse(float64(5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(5.5), result)

		// Invalid input
		_, err = schema.Parse(float64(15.5))
		assert.Error(t, err)
	})

	t.Run("range validation", func(t *testing.T) {
		schema := Float64().Min(1.0).Max(10.0)

		// Valid input
		result, err := schema.Parse(float64(5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(5.5), result)

		// Invalid inputs
		_, err = schema.Parse(float64(0.5))
		assert.Error(t, err)

		_, err = schema.Parse(float64(15.5))
		assert.Error(t, err)
	})

	t.Run("positive validation", func(t *testing.T) {
		schema := Float64().Positive()

		// Valid input
		result, err := schema.Parse(float64(5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(5.5), result)

		// Invalid inputs
		_, err = schema.Parse(float64(0.0))
		assert.Error(t, err)

		_, err = schema.Parse(float64(-5.5))
		assert.Error(t, err)
	})

	t.Run("negative validation", func(t *testing.T) {
		schema := Float64().Negative()

		// Valid input
		result, err := schema.Parse(float64(-5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(-5.5), result)

		// Invalid inputs
		_, err = schema.Parse(float64(0.0))
		assert.Error(t, err)

		_, err = schema.Parse(float64(5.5))
		assert.Error(t, err)
	})

	t.Run("non-negative validation", func(t *testing.T) {
		schema := Float64().NonNegative()

		// Valid inputs
		result, err := schema.Parse(float64(0.0))
		require.NoError(t, err)
		assert.Equal(t, float64(0.0), result)

		result, err = schema.Parse(float64(5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(5.5), result)

		// Invalid input
		_, err = schema.Parse(float64(-5.5))
		assert.Error(t, err)
	})

	t.Run("non-positive validation", func(t *testing.T) {
		schema := Float64().NonPositive()

		// Valid inputs
		result, err := schema.Parse(float64(0.0))
		require.NoError(t, err)
		assert.Equal(t, float64(0.0), result)

		result, err = schema.Parse(float64(-5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(-5.5), result)

		// Invalid input
		_, err = schema.Parse(float64(5.5))
		assert.Error(t, err)
	})

	t.Run("multiple of validation", func(t *testing.T) {
		schema := Float64().MultipleOf(2.5)

		// Valid inputs
		result, err := schema.Parse(float64(5.0))
		require.NoError(t, err)
		assert.Equal(t, float64(5.0), result)

		result, err = schema.Parse(float64(7.5))
		require.NoError(t, err)
		assert.Equal(t, float64(7.5), result)

		// Invalid input
		_, err = schema.Parse(float64(6.0))
		assert.Error(t, err)
	})

	t.Run("step validation", func(t *testing.T) {
		schema := Float64().Step(2.5)

		// Valid inputs (multiples of 2.5)
		result, err := schema.Parse(float64(5.0))
		require.NoError(t, err)
		assert.Equal(t, float64(5.0), result)

		result, err = schema.Parse(float64(7.5))
		require.NoError(t, err)
		assert.Equal(t, float64(7.5), result)

		// Invalid input
		_, err = schema.Parse(float64(6.0))
		assert.Error(t, err)
	})

	t.Run("int validation", func(t *testing.T) {
		schema := Float64().Int()

		// Valid inputs (integers)
		result, err := schema.Parse(float64(5.0))
		require.NoError(t, err)
		assert.Equal(t, float64(5.0), result)

		result, err = schema.Parse(float64(-10.0))
		require.NoError(t, err)
		assert.Equal(t, float64(-10.0), result)

		// Invalid input (has decimal part)
		_, err = schema.Parse(float64(5.5))
		assert.Error(t, err)
	})

	t.Run("finite validation", func(t *testing.T) {
		schema := Float64().Finite()

		// Valid input (finite number)
		result, err := schema.Parse(float64(5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(5.5), result)

		// Invalid inputs (infinite or NaN)
		_, err = schema.Parse(math.Inf(1))
		assert.Error(t, err)

		_, err = schema.Parse(math.Inf(-1))
		assert.Error(t, err)

		_, err = schema.Parse(math.NaN())
		assert.Error(t, err)
	})

	t.Run("safe validation", func(t *testing.T) {
		schema := Float64().Safe()

		// Valid input (within safe range)
		result, err := schema.Parse(float64(1000000.5))
		require.NoError(t, err)
		assert.Equal(t, float64(1000000.5), result)

		// Invalid input (outside safe range)
		_, err = schema.Parse(float64(1 << 54))
		assert.Error(t, err)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestFloat_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept values greater than 5.0
		schema := Float64().Refine(func(f float64) bool {
			return f > 5.0
		})

		result, err := schema.Parse(float64(10.5))
		require.NoError(t, err)
		assert.Equal(t, float64(10.5), result)

		_, err = schema.Parse(float64(3.5))
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be greater than 5.0"
		schema := Float64Ptr().Refine(func(f float64) bool {
			return f > 5.0
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse(float64(10.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(10.5), *result)

		_, err = schema.Parse(float64(3.5))
		assert.Error(t, err)
	})

	t.Run("refine pointer allows nil", func(t *testing.T) {
		schema := Float64Ptr().Nilable().Refine(func(f float64) bool {
			// Only check values greater than 5.0 (nil is handled by Nilable)
			return f > 5.0
		})

		// Expect nil to be accepted (handled by Nilable)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid value should pass
		result, err = schema.Parse(float64(10.5))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(10.5), *result)

		// Invalid value should fail
		_, err = schema.Parse(float64(3.5))
		assert.Error(t, err)
	})

	t.Run("refine float32 type", func(t *testing.T) {
		schema := Float32().Refine(func(f float32) bool {
			return f > 5.0
		})

		result, err := schema.Parse(float32(10.5))
		require.NoError(t, err)
		assert.Equal(t, float32(10.5), result)

		_, err = schema.Parse(float32(3.5))
		assert.Error(t, err)
	})
}

func TestFloat_RefineAny(t *testing.T) {
	t.Run("refineAny float64 schema", func(t *testing.T) {
		// Only accept values greater than 5.0 via RefineAny
		schema := Float64().RefineAny(func(v any) bool {
			f, ok := v.(float64)
			return ok && f > 5.0
		})

		// Valid value passes
		result, err := schema.Parse(float64(10.5))
		require.NoError(t, err)
		assert.Equal(t, float64(10.5), result)

		// Invalid value fails
		_, err = schema.Parse(float64(3.5))
		assert.Error(t, err)
	})

	t.Run("refineAny pointer schema", func(t *testing.T) {
		// Float64Ptr().RefineAny sees underlying float64 value
		schema := Float64Ptr().RefineAny(func(v any) bool {
			f, ok := v.(float64)
			return ok && f > 5.0
		})

		result, err := schema.Parse(float64(10.5))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(10.5), *result)

		_, err = schema.Parse(float64(3.5))
		assert.Error(t, err)
	})

	t.Run("refineAny nilable schema", func(t *testing.T) {
		// Nil input should bypass checks and be accepted because schema is Nilable()
		schema := Float64Ptr().Nilable().RefineAny(func(v any) bool {
			// Never called for nil input, but keep return true for completeness
			return true
		})

		// nil passes
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid value still passes
		result, err = schema.Parse(float64(10.5))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(10.5), *result)
	})
}

// =============================================================================
// Overwrite tests
// =============================================================================

func TestFloat_Overwrite(t *testing.T) {
	t.Run("basic overwrite functionality", func(t *testing.T) {
		schema := Float64().Overwrite(func(val float64) float64 {
			return val * 2
		})

		// Test with valid input
		result, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 10.0, result)

		// Test with negative value
		result, err = schema.Parse(-3.5)
		require.NoError(t, err)
		assert.Equal(t, -7.0, result)

		// Test with zero
		result, err = schema.Parse(0.0)
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})

	t.Run("overwrite preserves type", func(t *testing.T) {
		// Test Float64 preserves type
		schema64 := Float64().Overwrite(func(val float64) float64 {
			return val + 1
		})

		result64, err := schema64.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 6.0, result64)
		assert.IsType(t, float64(0), result64)

		// Test Float32 preserves type
		schema32 := Float32().Overwrite(func(val float32) float32 {
			return val + 1
		})

		result32, err := schema32.Parse(float32(5.0))
		require.NoError(t, err)
		assert.Equal(t, float32(6.0), result32)
		assert.IsType(t, float32(0), result32)
	})

	t.Run("overwrite with type conversion", func(t *testing.T) {
		schema := Float64().Overwrite(math.Floor)

		// Test with float that has decimal
		result, err := schema.Parse(5.7)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)

		// Test with negative float
		result, err = schema.Parse(-3.9)
		require.NoError(t, err)
		assert.Equal(t, -4.0, result)
	})

	t.Run("overwrite can be chained with validations", func(t *testing.T) {
		schema := Float64().
			Overwrite(math.Abs).
			Max(10.0)

		// Test positive value within limit
		result, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)

		// Test negative value converted to positive and within limit
		result, err = schema.Parse(-8.0)
		require.NoError(t, err)
		assert.Equal(t, 8.0, result)

		// Test negative value that exceeds limit after conversion
		_, err = schema.Parse(-15.0)
		assert.Error(t, err)
	})

	t.Run("multiple overwrite calls", func(t *testing.T) {
		schema := Float64().
			Overwrite(func(val float64) float64 {
				return val * 2
			}).
			Overwrite(func(val float64) float64 {
				return val + 1
			})

		// Should apply transformations in order: 5 * 2 = 10, then 10 + 1 = 11
		result, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 11.0, result)
	})

	t.Run("overwrite with custom error message", func(t *testing.T) {
		// Test that overwrite still allows custom error messages for other validations
		schema := Float64().
			Overwrite(func(val float64) float64 {
				return val * 2
			}).
			Max(10.0, "Value must be at most 10 after doubling")

		// Test valid case
		result, err := schema.Parse(4.0)
		require.NoError(t, err)
		assert.Equal(t, 8.0, result)

		// Test invalid case
		_, err = schema.Parse(6.0) // 6 * 2 = 12, which is > 10
		assert.Error(t, err)
	})

	t.Run("overwrite with MustParse", func(t *testing.T) {
		schema := Float64().Overwrite(func(val float64) float64 {
			return val * 3
		})

		result := schema.MustParse(4.0)
		assert.Equal(t, 12.0, result)

		// Test panic on invalid input type
		assert.Panics(t, func() {
			schema.MustParse("not a number")
		})
	})

	t.Run("overwrite return type matches original schema", func(t *testing.T) {
		// Unlike Transform which returns ZodTransform, Overwrite should return same type
		original := Float64()
		overwritten := original.Overwrite(func(val float64) float64 {
			return val + 1
		})

		// Should be the same underlying type
		var _ *ZodFloatTyped[float64, float64] = overwritten

		// Can continue chaining float methods
		final := overwritten.Max(100.0)
		var _ *ZodFloatTyped[float64, float64] = final
	})

	t.Run("overwrite strict type checking", func(t *testing.T) {
		// Test that overwrite only accepts matching float types
		schema := Float64().Overwrite(func(val float64) float64 {
			return val * 2
		})

		// Should work with float64 input
		result, err := schema.Parse(float64(5.0))
		require.NoError(t, err)
		assert.Equal(t, 10.0, result)

		// Should reject int input
		_, err = schema.Parse(int(5))
		assert.Error(t, err, "Should reject int input")

		// Should reject int32 input
		_, err = schema.Parse(int32(3))
		assert.Error(t, err, "Should reject int32 input")
	})

	t.Run("overwrite float32 strict type checking", func(t *testing.T) {
		// Test that float32 schema only accepts float32 types
		schema := Float32().Overwrite(func(val float32) float32 {
			return val + 1
		})

		// Should work with float32 input
		result, err := schema.Parse(float32(5.0))
		require.NoError(t, err)
		assert.Equal(t, float32(6.0), result)

		// Should reject int input
		_, err = schema.Parse(int(5))
		assert.Error(t, err, "Should reject int input")

		// Should reject int64 input
		_, err = schema.Parse(int64(5))
		assert.Error(t, err, "Should reject int64 input")

		// Should reject float64 input (different float type)
		_, err = schema.Parse(float64(5.0))
		assert.Error(t, err, "Should reject float64 input")
	})
}

// =============================================================================
// Coercion tests
// =============================================================================

func TestFloat_Coercion(t *testing.T) {
	t.Run("string coercion", func(t *testing.T) {
		schema := CoercedFloat64()

		// Test string "100.5" -> float64 100.5
		result, err := schema.Parse("100.5")
		require.NoError(t, err, "Should coerce string '100.5' to float64 100.5")
		assert.Equal(t, float64(100.5), result)

		// Test string "-10.5" -> float64 -10.5
		result, err = schema.Parse("-10.5")
		require.NoError(t, err, "Should coerce string '-10.5' to float64 -10.5")
		assert.Equal(t, float64(-10.5), result)
	})

	t.Run("integer coercion", func(t *testing.T) {
		schema := CoercedFloat64()

		testCases := []struct {
			input    any
			expected float64
			name     string
		}{
			{42, 42.0, "int 42 to 42.0"},
			{int32(10), 10.0, "int32 10 to 10.0"},
			{int64(-5), -5.0, "int64 -5 to -5.0"},
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
		schema := CoercedFloat64()

		// Boolean values should be coerced to floats
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, float64(1.0), result)

		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, float64(0.0), result)
	})

	t.Run("float type coercion", func(t *testing.T) {
		schema := CoercedFloat64()

		testCases := []struct {
			input    any
			expected float64
			name     string
		}{
			{float32(100.5), 100.5, "float32 100.5 to float64 100.5"},
			{float64(200.5), 200.5, "float64 200.5 to float64 200.5"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.InDelta(t, tc.expected, result, 0.0001)
			})
		}
	})

	t.Run("coerced float32 schema", func(t *testing.T) {
		schema := CoercedFloat32()

		// Test string coercion to float32
		result, err := schema.Parse("100.5")
		require.NoError(t, err)
		assert.IsType(t, float32(0), result)
		assert.InDelta(t, float32(100.5), result, 0.0001)

		// Test int coercion to float32
		result, err = schema.Parse(100)
		require.NoError(t, err)
		assert.Equal(t, float32(100.0), result)
	})

	t.Run("invalid coercion inputs", func(t *testing.T) {
		schema := CoercedFloat64()

		// Inputs that cannot be coerced
		invalidInputs := []any{
			"not a number", "invalid", []float64{1.0}, map[string]float64{"key": 1.0}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedFloat64().Min(5.0).Max(100.0)

		// Coercion then validation passes
		result, err := schema.Parse("50.5")
		require.NoError(t, err)
		assert.Equal(t, float64(50.5), result)

		// Coercion then validation fails
		_, err = schema.Parse("3.5")
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestFloat_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := Float64()

		_, err := schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a float value"
		schema := Float64Ptr(core.SchemaParams{Error: customError})

		_, err := schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("validation error messages", func(t *testing.T) {
		schema := Float64().Min(10.0)

		_, err := schema.Parse(float64(5.0))
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestFloat_EdgeCases(t *testing.T) {
	t.Run("nil handling with *float64", func(t *testing.T) {
		schema := Float64Ptr().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid float
		result, err = schema.Parse(float64(200.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(200.5), *result)
	})

	t.Run("special float values", func(t *testing.T) {
		schema := Float64()

		// Test positive infinity
		result, err := schema.Parse(math.Inf(1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(result, 1))

		// Test negative infinity
		result, err = schema.Parse(math.Inf(-1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(result, -1))

		// Test NaN
		result, err = schema.Parse(math.NaN())
		require.NoError(t, err)
		assert.True(t, math.IsNaN(result))
	})

	t.Run("very large and small values", func(t *testing.T) {
		schema := Float64()

		// Test very large value
		result, err := schema.Parse(float64(1e308))
		require.NoError(t, err)
		assert.Equal(t, float64(1e308), result)

		// Test very small value
		result, err = schema.Parse(float64(1e-308))
		require.NoError(t, err)
		assert.Equal(t, float64(1e-308), result)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Float64()

		// Parse with empty context slice
		result, err := schema.Parse(float64(200.5))
		require.NoError(t, err)
		assert.Equal(t, float64(200.5), result)
	})

	t.Run("precision handling", func(t *testing.T) {
		schema := Float64()

		// Test precision preservation
		value := float64(0.1) + float64(0.2)
		result, err := schema.Parse(value)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})
}

// =============================================================================
// Type-specific tests
// =============================================================================

func TestFloat_TypeSpecific(t *testing.T) {
	t.Run("Float32 vs Float64 distinction", func(t *testing.T) {
		float32Schema := Float32()
		float64Schema := Float64()

		// Both should handle their respective types
		resultFloat32, err := float32Schema.Parse(float32(100.5))
		require.NoError(t, err)
		assert.IsType(t, float32(0), resultFloat32)

		resultFloat64, err := float64Schema.Parse(float64(200.5))
		require.NoError(t, err)
		assert.IsType(t, float64(0), resultFloat64)
	})

	t.Run("Number alias", func(t *testing.T) {
		numberSchema := Number()

		// Number should behave like Float64
		result, err := numberSchema.Parse(float64(200.5))
		require.NoError(t, err)
		assert.IsType(t, float64(0), result)
		assert.Equal(t, float64(200.5), result)
	})

	t.Run("mixed float types in validation chains", func(t *testing.T) {
		// Test that validation methods work correctly with different float types
		float32Schema := Float32().Min(1.0).Max(10.0).Positive()
		float64Schema := Float64().Min(1.0).Max(10.0).Positive()

		// Both should validate successfully
		result32, err := float32Schema.Parse(float32(5.5))
		require.NoError(t, err)
		assert.Equal(t, float32(5.5), result32)

		result64, err := float64Schema.Parse(float64(5.5))
		require.NoError(t, err)
		assert.Equal(t, float64(5.5), result64)

		// Both should fail validation
		_, err = float32Schema.Parse(float32(-1.5))
		assert.Error(t, err)

		_, err = float64Schema.Parse(-1.5)
		assert.Error(t, err)
	})

	t.Run("coerced float type equivalence", func(t *testing.T) {
		coercedFloat32Schema := CoercedFloat32()
		coercedFloat64Schema := CoercedFloat64()

		// Both should coerce strings successfully
		resultFloat32, err1 := coercedFloat32Schema.Parse("100.5")
		require.NoError(t, err1)

		resultFloat64, err2 := coercedFloat64Schema.Parse("100.5")
		require.NoError(t, err2)

		// Types should be different but values equivalent
		assert.IsType(t, float32(0), resultFloat32)
		assert.IsType(t, float64(0), resultFloat64)
		assert.InDelta(t, float32(100.5), resultFloat32, 0.0001)
		assert.Equal(t, float64(100.5), resultFloat64)
	})

	t.Run("finite vs infinite validation", func(t *testing.T) {
		finiteSchema := Float64().Finite()
		regularSchema := Float64()

		// Regular schema should accept infinity
		result, err := regularSchema.Parse(math.Inf(1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(result, 1))

		// Finite schema should reject infinity
		_, err = finiteSchema.Parse(math.Inf(1))
		assert.Error(t, err)

		// Both should accept finite values
		result, err = finiteSchema.Parse(float64(100.5))
		require.NoError(t, err)
		assert.Equal(t, float64(100.5), result)

		result, err = regularSchema.Parse(float64(100.5))
		require.NoError(t, err)
		assert.Equal(t, float64(100.5), result)
	})
}

// =============================================================================
// Pointer identity preservation tests
// =============================================================================

func TestFloat_PointerIdentityPreservation(t *testing.T) {
	t.Run("Float32 Optional preserves pointer identity", func(t *testing.T) {
		schema := Float32().Optional()

		originalFloat := float32(3.14)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer (now type-safe with *float32)
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, float32(3.14), *result)
	})

	t.Run("Float32 Nilable preserves pointer identity", func(t *testing.T) {
		schema := Float32().Nilable()

		originalFloat := float32(2.718)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float32 Nilable should preserve pointer identity")
		assert.Equal(t, float32(2.718), *result)
	})

	t.Run("Float32 Nullish preserves pointer identity", func(t *testing.T) {
		schema := Float32().Nullish()

		originalFloat := float32(1.414)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float32 Nullish should preserve pointer identity")
		assert.Equal(t, float32(1.414), *result)
	})

	t.Run("Float32 Default().Optional() chaining preserves pointer identity", func(t *testing.T) {
		schema := Float32().Default(1.0).Optional()

		originalFloat := float32(5.5)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float32 Default().Optional() should preserve pointer identity")
		assert.Equal(t, float32(5.5), *result)
	})

	t.Run("Float32 Min/Max with Optional preserves pointer identity", func(t *testing.T) {
		schema := Float32().Min(0.0).Max(10.0).Optional()

		originalFloat := float32(5.0)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float32 validation with Optional should preserve pointer identity")
		assert.Equal(t, float32(5.0), *result)
	})

	t.Run("Float32 Refine with Optional preserves pointer identity", func(t *testing.T) {
		schema := Float32().Refine(func(f float32) bool {
			return f > 0 // Only positive numbers
		}).Optional()

		originalFloat := float32(42.42)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float32 Refine().Optional() should preserve pointer identity")
		assert.Equal(t, float32(42.42), *result)
	})

	t.Run("Float64 Optional preserves pointer identity", func(t *testing.T) {
		schema := Float64().Optional()

		originalFloat := float64(2.718)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, float64(2.718), *result)
	})

	t.Run("Float64 Nilable preserves pointer identity", func(t *testing.T) {
		schema := Float64().Nilable()

		originalFloat := float64(1.414)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, float64(1.414), *result)
	})

	t.Run("Float64Ptr Optional preserves pointer identity", func(t *testing.T) {
		schema := Float64Ptr().Optional()

		originalFloat := float64(0.577)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float64Ptr Optional should preserve pointer identity")
		assert.Equal(t, float64(0.577), *result)
	})

	t.Run("Float64Ptr Nilable preserves pointer identity", func(t *testing.T) {
		schema := Float64Ptr().Nilable()

		originalFloat := float64(-1.618)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float64Ptr Nilable should preserve pointer identity")
		assert.Equal(t, float64(-1.618), *result)
	})

	t.Run("Float64 Nullish preserves pointer identity", func(t *testing.T) {
		schema := Float64().Nullish()

		originalFloat := float64(6.283)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Float64 Nullish should preserve pointer identity")
		assert.Equal(t, float64(6.283), *result)
	})

	t.Run("Optional handles nil consistently", func(t *testing.T) {
		schema := Float64().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable handles nil consistently", func(t *testing.T) {
		schema := Float64().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default().Optional() chaining preserves pointer identity", func(t *testing.T) {
		schema := Float64().Default(1.0).Optional()

		originalFloat := float64(2.5)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Default().Optional() should preserve pointer identity")
		assert.Equal(t, float64(2.5), *result)
	})

	t.Run("Validation with Optional preserves pointer identity", func(t *testing.T) {
		schema := Float64().Min(0.0).Max(100.0).Optional()

		originalFloat := float64(50.5)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Validation().Optional() should preserve pointer identity")
		assert.Equal(t, float64(50.5), *result)
	})

	t.Run("Refine with Optional preserves pointer identity", func(t *testing.T) {
		schema := Float64().Refine(func(f float64) bool {
			return f > 0 // Only positive numbers
		}).Optional()

		originalFloat := float64(42.42)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Refine().Optional() should preserve pointer identity")
		assert.Equal(t, float64(42.42), *result)
	})

	t.Run("Multiple float types pointer identity", func(t *testing.T) {
		testCases := []struct {
			name  string
			value float64
		}{
			{"positive", 123.456},
			{"negative", -789.012},
			{"zero", 0.0},
			{"small", 0.001},
			{"large", 999999.999},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema := Float64().Optional()
				originalPtr := &tc.value

				result, err := schema.Parse(originalPtr)
				require.NoError(t, err)

				assert.True(t, result == originalPtr, "Pointer identity should be preserved for %s", tc.name)
				assert.Equal(t, tc.value, *result)
			})
		}
	})

	t.Run("Number alias pointer identity", func(t *testing.T) {
		schema := Number().Optional() // Number is alias for Float64

		originalFloat := float64(3.14159)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Number alias should preserve pointer identity")
		assert.Equal(t, float64(3.14159), *result)
	})
}

// =============================================================================
// Flexible Float Tests (new Float() function)
// =============================================================================

func TestFloat_FlexibleTypes(t *testing.T) {
	t.Run("Float() is now an alias for Float64()", func(t *testing.T) {
		schema := Float()

		// Test float64 (should work)
		result1, err1 := schema.Parse(float64(3.14))
		require.NoError(t, err1)
		assert.Equal(t, float64(3.14), result1)
		assert.IsType(t, float64(0), result1)

		// Test float32 (should fail - no longer flexible)
		_, err2 := schema.Parse(float32(2.718))
		assert.Error(t, err2)
	})

	t.Run("rejects non-float types", func(t *testing.T) {
		schema := Float()

		// Test string rejection
		_, err1 := schema.Parse("3.14")
		assert.Error(t, err1)

		// Test integer rejection
		_, err2 := schema.Parse(42)
		assert.Error(t, err2)

		// Test other types
		invalidInputs := []any{
			"not a number",
			42,
			true,
			false,
			[]float64{1.0},
			map[string]float64{"key": 1.0},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v (%T)", input, input)
		}
	})

	t.Run("pointer handling works correctly", func(t *testing.T) {
		schema := Float()

		// Test *float64 (should work and convert to value)
		val1 := float64(3.14)
		result1, err1 := schema.Parse(&val1)
		require.NoError(t, err1)
		assert.Equal(t, float64(3.14), result1)
		assert.IsType(t, float64(0), result1)

		// Test *float32 (should fail - no longer flexible)
		val2 := float32(2.718)
		_, err2 := schema.Parse(&val2)
		assert.Error(t, err2)
	})

	t.Run("handles nil pointers correctly", func(t *testing.T) {
		// Default schema should reject nil
		schema := Float()
		_, err := schema.Parse(nil)
		assert.Error(t, err)

		// Nilable schema should accept nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Optional schema should accept nil
		optionalSchema := schema.Optional()
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("works with validations and modifiers", func(t *testing.T) {
		schema := Float().Min(0.0).Max(10.0)

		// Valid values (float64 only)
		result1, err1 := schema.Parse(float64(5.5))
		require.NoError(t, err1)
		assert.Equal(t, float64(5.5), result1)

		result2, err2 := schema.Parse(float64(8.8))
		require.NoError(t, err2)
		assert.Equal(t, float64(8.8), result2)

		// Invalid values (outside range)
		_, err3 := schema.Parse(float64(-1.0))
		assert.Error(t, err3)

		_, err4 := schema.Parse(float64(15.0))
		assert.Error(t, err4)

		// Invalid type (float32 no longer accepted)
		_, err5 := schema.Parse(float32(5.5))
		assert.Error(t, err5)
	})

	t.Run("pointer validation converts to value", func(t *testing.T) {
		schema := Float().Min(0.0)

		val := float64(42.42)
		originalPtr := &val

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Should convert pointer to value
		assert.Equal(t, float64(42.42), result)
		assert.IsType(t, float64(0), result)
	})

	t.Run("MustParse works with float64 types", func(t *testing.T) {
		schema := Float()

		// Should work with float64
		result1 := schema.MustParse(float64(3.14))
		assert.Equal(t, float64(3.14), result1)

		// Should panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})

		// Should panic on float32 (no longer flexible)
		assert.Panics(t, func() {
			schema.MustParse(float32(2.718))
		})
	})

	t.Run("edge cases and special values", func(t *testing.T) {
		schema := Float()

		// Test zero values
		result1, err1 := schema.Parse(float64(0.0))
		require.NoError(t, err1)
		assert.Equal(t, float64(0.0), result1)

		// Test negative zero
		result2, err2 := schema.Parse(float64(-0.0))
		require.NoError(t, err2)
		assert.Equal(t, float64(-0.0), result2)

		// Test very large values
		result3, err3 := schema.Parse(float64(1e308))
		require.NoError(t, err3)
		assert.Equal(t, float64(1e308), result3)
	})

	t.Run("custom error messages work", func(t *testing.T) {
		customError := "Expected float64 type"
		schema := Float(core.SchemaParams{Error: customError})

		_, err := schema.Parse("not a float")
		assert.Error(t, err)
		// Note: The exact error message checking might depend on the error formatting implementation
	})

	t.Run("comparisons with existing functions", func(t *testing.T) {
		// Test that new Float() is now identical to Float64()
		floatSchema := Float()
		float32Schema := Float32()
		float64Schema := Float64()

		// Float() now only accepts float64 (no longer flexible)
		_, err1 := floatSchema.Parse(float32(3.14))
		assert.Error(t, err1, "Float() should reject float32")

		result2, err2 := floatSchema.Parse(float64(3.14))
		require.NoError(t, err2)
		assert.IsType(t, float64(0), result2)

		// Type-specific schemas work as expected
		result3, err3 := float32Schema.Parse(float32(3.14))
		require.NoError(t, err3)
		assert.IsType(t, float32(0), result3)

		result4, err4 := float64Schema.Parse(float64(3.14))
		require.NoError(t, err4)
		assert.IsType(t, float64(0), result4)
	})
}

// =============================================================================
// COMPREHENSIVE TYPE SAFETY TESTS
// =============================================================================

func TestFloat_ComprehensiveTypeSafety(t *testing.T) {
	t.Run("Float32 type safety chain", func(t *testing.T) {
		// Test complete type evolution
		valueSchema := Float32() // *ZodFloatTyped[float32, float32]
		var _ *ZodFloatTyped[float32, float32] = valueSchema

		optionalSchema := valueSchema.Optional() // *ZodFloatTyped[float32, *float32]
		var _ *ZodFloatTyped[float32, *float32] = optionalSchema

		withDefaultSchema := optionalSchema.Default(5.5) // *ZodFloatTyped[float32, *float32]
		var _ *ZodFloatTyped[float32, *float32] = withDefaultSchema

		withValidationSchema := withDefaultSchema.Min(0.0).Max(100.0) // *ZodFloatTyped[float32, *float32]
		var _ *ZodFloatTyped[float32, *float32] = withValidationSchema

		// Test runtime behavior
		result, err := withValidationSchema.Parse(float32(42.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float32(42.5), *result)
	})

	t.Run("Float64 type safety chain", func(t *testing.T) {
		// Test complete type evolution
		valueSchema := Float64() // *ZodFloatTyped[float64, float64]
		var _ *ZodFloatTyped[float64, float64] = valueSchema

		nilableSchema := valueSchema.Nilable() // *ZodFloatTyped[float64, *float64]
		var _ *ZodFloatTyped[float64, *float64] = nilableSchema

		withDefaultSchema := nilableSchema.Default(10.5) // *ZodFloatTyped[float64, *float64]
		var _ *ZodFloatTyped[float64, *float64] = withDefaultSchema

		// Test runtime behavior
		result, err := withDefaultSchema.Parse(float64(99.9))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(99.9), *result)

		// Test nil handling - Default should always win (Zod v4 behavior)
		nilResult, err := withDefaultSchema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, nilResult) // Default takes precedence over Nilable
		assert.Equal(t, 10.5, *nilResult)
	})

	t.Run("All float types Optional modifier", func(t *testing.T) {
		// Float32
		float32Optional := Float32().Optional()
		var _ *ZodFloatTyped[float32, *float32] = float32Optional

		// Float64
		float64Optional := Float64().Optional()
		var _ *ZodFloatTyped[float64, *float64] = float64Optional

		// Number (alias for Float64)
		numberOptional := Number().Optional()
		var _ *ZodFloatTyped[float64, *float64] = numberOptional

		// Test all work at runtime
		tests := []struct {
			name   string
			schema any
			input  any
			expect any
		}{
			{"Float32", float32Optional, float32(1.5), float32(1.5)},
			{"Float64", float64Optional, float64(2.5), float64(2.5)},
			{"Number", numberOptional, float64(3.5), float64(3.5)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				switch schema := tt.schema.(type) {
				case *ZodFloatTyped[float32, *float32]:
					result, err := schema.Parse(tt.input)
					require.NoError(t, err)
					require.NotNil(t, result)
					assert.Equal(t, tt.expect, *result)
				case *ZodFloatTyped[float64, *float64]:
					result, err := schema.Parse(tt.input)
					require.NoError(t, err)
					require.NotNil(t, result)
					assert.Equal(t, tt.expect, *result)
				}
			})
		}
	})

	t.Run("Pointer types preserve identity", func(t *testing.T) {
		// Float32Ptr
		float32PtrSchema := Float32Ptr()
		var _ *ZodFloatTyped[float32, *float32] = float32PtrSchema

		originalFloat32 := float32(42.42)
		originalPtr32 := &originalFloat32

		result32, err := float32PtrSchema.Parse(originalPtr32)
		require.NoError(t, err)
		assert.True(t, result32 == originalPtr32, "Float32Ptr should preserve pointer identity")

		// Float64Ptr
		float64PtrSchema := Float64Ptr()
		var _ *ZodFloatTyped[float64, *float64] = float64PtrSchema

		originalFloat64 := float64(99.99)
		originalPtr64 := &originalFloat64

		result64, err := float64PtrSchema.Parse(originalPtr64)
		require.NoError(t, err)
		assert.True(t, result64 == originalPtr64, "Float64Ptr should preserve pointer identity")
	})
}

func TestFloat_TypeEvolutionChaining(t *testing.T) {
	t.Run("basic type evolution: value to pointer", func(t *testing.T) {
		// Float32: value -> pointer
		schema := Float32(). // *ZodFloatTyped[float32, float32]
					Default(1.0). // *ZodFloatTyped[float32, float32]
					Optional()    // *ZodFloatTyped[float32, *float32]

		var _ *ZodFloatTyped[float32, *float32] = schema

		result, err := schema.Parse(float32(42.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float32(42.5), *result)
		assert.IsType(t, (*float32)(nil), result)
	})

	t.Run("complex chaining with validation", func(t *testing.T) {
		// Complex chaining preserving types throughout
		schema := Float64(). // *ZodFloatTyped[float64, float64]
					Min(10.0).     // *ZodFloatTyped[float64, float64]
					Max(100.0).    // *ZodFloatTyped[float64, float64]
					Positive().    // *ZodFloatTyped[float64, float64]
					Default(50.0). // *ZodFloatTyped[float64, float64]
					Nilable()      // *ZodFloatTyped[float64, *float64]

		var _ *ZodFloatTyped[float64, *float64] = schema

		// Test valid value
		result, err := schema.Parse(float64(75.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(75.5), *result)

		// Test nil handling - Default should always win (Zod v4 behavior)
		nilResult, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, nilResult) // Default takes precedence over Nilable
		assert.Equal(t, float64(50.0), *nilResult)
	})

	t.Run("reverse type evolution: pointer to value schema", func(t *testing.T) {
		// This tests whether pointer identity is preserved throughout chain
		schema := Float32Ptr(). // *ZodFloatTyped[float32, *float32]
					Min(5.0).  // *ZodFloatTyped[float32, *float32]
					Max(50.0). // *ZodFloatTyped[float32, *float32]
					Positive() // *ZodFloatTyped[float32, *float32]

		var _ *ZodFloatTyped[float32, *float32] = schema

		// Test pointer identity preservation
		originalFloat := float32(25.5)
		originalPtr := &originalFloat

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result == originalPtr, "Pointer identity should be preserved throughout validation chain")
		assert.Equal(t, float32(25.5), *result)
	})
}

func TestFloat_CompilationTimeTypeSafety(t *testing.T) {
	t.Run("compile time type assertions", func(t *testing.T) {
		// These should all compile without issues

		// Float32 chains
		var _ *ZodFloatTyped[float32, float32] = Float32()
		var _ *ZodFloatTyped[float32, float32] = Float32().Default(1.0)
		var _ *ZodFloatTyped[float32, float32] = Float32().Min(0.0).Max(100.0)
		var _ *ZodFloatTyped[float32, *float32] = Float32().Optional()
		var _ *ZodFloatTyped[float32, *float32] = Float32().Nilable()
		var _ *ZodFloatTyped[float32, *float32] = Float32().Nullish()
		var _ *ZodFloatTyped[float32, *float32] = Float32().Default(1.0).Optional()

		// Float64 chains
		var _ *ZodFloatTyped[float64, float64] = Float64()
		var _ *ZodFloatTyped[float64, float64] = Float64().Default(1.0)
		var _ *ZodFloatTyped[float64, float64] = Float64().Min(0.0).Max(100.0)
		var _ *ZodFloatTyped[float64, *float64] = Float64().Optional()
		var _ *ZodFloatTyped[float64, *float64] = Float64().Nilable()
		var _ *ZodFloatTyped[float64, *float64] = Float64().Nullish()
		var _ *ZodFloatTyped[float64, *float64] = Float64().Default(1.0).Optional()

		// Pointer types
		var _ *ZodFloatTyped[float32, *float32] = Float32Ptr()
		var _ *ZodFloatTyped[float64, *float64] = Float64Ptr()
		var _ *ZodFloatTyped[float64, *float64] = NumberPtr()

		// Complex chains
		var _ *ZodFloatTyped[float32, *float32] = Float32().Min(0).Max(100).Positive().Default(50).Optional()
		var _ *ZodFloatTyped[float64, *float64] = Float64().Finite().Positive().Default(10.5).Nilable()
	})
}

func TestFloat_ErrorHandlingWithTypeSafety(t *testing.T) {
	t.Run("type-safe error handling", func(t *testing.T) {
		schema := Float32().Min(10.0).Max(100.0).Optional()
		var _ *ZodFloatTyped[float32, *float32] = schema

		// Test validation error
		_, err := schema.Parse(float32(5.0))
		assert.Error(t, err)

		// Test successful parsing
		result, err := schema.Parse(float32(50.0))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float32(50.0), *result)
		assert.IsType(t, (*float32)(nil), result)
	})

	t.Run("nil handling with type safety", func(t *testing.T) {
		nilableSchema := Float64().Nilable()
		var _ *ZodFloatTyped[float64, *float64] = nilableSchema

		// Test nil input
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test non-nil input
		result, err = nilableSchema.Parse(float64(42.5))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(42.5), *result)
	})
}

func TestFloat_MustParseTypeSafety(t *testing.T) {
	t.Run("MustParse preserves type safety", func(t *testing.T) {
		// Float32 Optional
		schema32 := Float32().Optional()
		var _ *ZodFloatTyped[float32, *float32] = schema32

		result32 := schema32.MustParse(float32(42.5))
		assert.IsType(t, (*float32)(nil), result32)
		require.NotNil(t, result32)
		assert.Equal(t, float32(42.5), *result32)

		// Float64 Nilable
		schema64 := Float64().Nilable()
		var _ *ZodFloatTyped[float64, *float64] = schema64

		result64 := schema64.MustParse(float64(99.9))
		assert.IsType(t, (*float64)(nil), result64)
		require.NotNil(t, result64)
		assert.Equal(t, float64(99.9), *result64)

		// Test nil handling
		nilResult := schema64.MustParse(nil)
		assert.Nil(t, nilResult)
	})
}

// =============================================================================
// Check Method Tests
// =============================================================================

func TestFloat_Check(t *testing.T) {
	t.Run("adds multiple issues for invalid input", func(t *testing.T) {
		schema := Float64().Check(func(value float64, p *core.ParsePayload) {
			if value < 0 {
				p.AddIssueWithMessage("must be non-negative")
			}
			if value > 100 {
				p.AddIssueWithCode(core.TooBig, "too big")
			}
		})

		_, err := schema.Parse(-5.5)
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		require.Len(t, zErr.Issues, 1)

		_, err = schema.Parse(150.0)
		require.Error(t, err)
		require.True(t, issues.IsZodError(err, &zErr))
		require.Len(t, zErr.Issues, 1)
	})

	t.Run("succeeds for valid input", func(t *testing.T) {
		schema := Float64().Check(func(value float64, p *core.ParsePayload) {
			if value < 0 {
				p.AddIssueWithMessage("neg")
			}
		})
		res, err := schema.Parse(12.34)
		require.NoError(t, err)
		assert.Equal(t, 12.34, res)
	})

	t.Run("works with pointer types", func(t *testing.T) {
		schema := Float64Ptr().Check(func(value *float64, p *core.ParsePayload) {
			if value != nil && *value < 10 {
				p.AddIssueWithMessage("too small")
			}
		})
		small := 3.14
		_, err := schema.Parse(&small)
		require.Error(t, err)
		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)

		big := 42.0
		result, err := schema.Parse(&big)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 42.0, *result)
	})
}

// =============================================================================
// NonOptional tests
// =============================================================================

func TestFloat_NonOptional(t *testing.T) {
	schema := Float64().NonOptional()

	// valid input
	r, err := schema.Parse(float64(1.23))
	require.NoError(t, err)
	assert.Equal(t, float64(1.23), r)

	// nil should error
	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.InvalidType, zErr.Issues[0].Code)
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}

	// chain
	chain := Float64().Optional().NonOptional()
	var _ *ZodFloatTyped[float64, float64] = chain
	_, err = chain.Parse(nil)
	assert.Error(t, err)

	// object embedding
	obj := Object(map[string]core.ZodSchema{
		"price": Float64().Optional().NonOptional(),
	})
	_, err = obj.Parse(map[string]any{"price": float64(99.9)})
	require.NoError(t, err)
	_, err = obj.Parse(map[string]any{"price": nil})
	assert.Error(t, err)

	// Float64Ptr().NonOptional()
	val := float64(7.89)
	ptrSchema := Float64Ptr().NonOptional()
	res2, err := ptrSchema.Parse(&val)
	require.NoError(t, err)
	assert.Equal(t, float64(7.89), res2)

	_, err = ptrSchema.Parse(nil)
	assert.Error(t, err)
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestFloat_StrictParse(t *testing.T) {
	t.Run("basic functionality float32", func(t *testing.T) {
		schema := Float32()

		// Test StrictParse with exact type match
		result, err := schema.StrictParse(float32(42.5))
		require.NoError(t, err)
		assert.Equal(t, float32(42.5), result)
		assert.IsType(t, float32(0), result)

		// Test StrictParse with negative value
		negResult, err := schema.StrictParse(float32(-10.5))
		require.NoError(t, err)
		assert.Equal(t, float32(-10.5), negResult)
	})

	t.Run("basic functionality float64", func(t *testing.T) {
		schema := Float64()

		// Test StrictParse with exact type match
		result, err := schema.StrictParse(float64(123.456))
		require.NoError(t, err)
		assert.Equal(t, float64(123.456), result)
		assert.IsType(t, float64(0), result)

		// Test StrictParse with zero
		zeroResult, err := schema.StrictParse(float64(0))
		require.NoError(t, err)
		assert.Equal(t, float64(0), zeroResult)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := Float64().Min(10.0)

		// Valid case
		result, err := schema.StrictParse(float64(15.5))
		require.NoError(t, err)
		assert.Equal(t, float64(15.5), result)

		// Invalid case - too small
		_, err = schema.StrictParse(float64(5.0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 10")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := Float64Ptr()
		floatVal := float64(99.99)

		// Test with valid pointer input
		result, err := schema.StrictParse(&floatVal)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(99.99), *result)
		assert.IsType(t, (*float64)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := Float32Ptr().Default(42.0)
		var nilPtr *float32 = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float32(42.0), *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		schema := Float64Ptr().Min(100.0).Prefault(float64(200.0))
		smallVal := float64(50.0) // Too small for Min(100.0)

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema.StrictParse(&smallVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 100")

		// Test with nil input (should use prefault)
		result, err := schema.StrictParse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(200.0), *result)
	})

	t.Run("special float values", func(t *testing.T) {
		schema := Float64()

		// Test infinity
		infResult, err := schema.StrictParse(math.Inf(1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(infResult, 1))

		// Test negative infinity
		negInfResult, err := schema.StrictParse(math.Inf(-1))
		require.NoError(t, err)
		assert.True(t, math.IsInf(negInfResult, -1))

		// Test NaN
		nanResult, err := schema.StrictParse(math.NaN())
		require.NoError(t, err)
		assert.True(t, math.IsNaN(nanResult))
	})
}

func TestFloat_MustStrictParse(t *testing.T) {
	t.Run("basic functionality float32", func(t *testing.T) {
		schema := Float32()

		// Test MustStrictParse with valid input
		result := schema.MustStrictParse(float32(123.45))
		assert.Equal(t, float32(123.45), result)
		assert.IsType(t, float32(0), result)

		// Test MustStrictParse with zero
		zeroResult := schema.MustStrictParse(float32(0))
		assert.Equal(t, float32(0), zeroResult)
	})

	t.Run("basic functionality float64", func(t *testing.T) {
		schema := Float64()

		// Test MustStrictParse with valid input
		result := schema.MustStrictParse(float64(456.789))
		assert.Equal(t, float64(456.789), result)
		assert.IsType(t, float64(0), result)

		// Test MustStrictParse with negative value
		negResult := schema.MustStrictParse(float64(-99.99))
		assert.Equal(t, float64(-99.99), negResult)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := Float64().Min(100.0)

		// Test panic with validation failure
		assert.Panics(t, func() {
			schema.MustStrictParse(float64(50.0)) // Too small, should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := Float32Ptr()
		floatVal := float32(77.77)

		// Test with valid pointer input
		result := schema.MustStrictParse(&floatVal)
		require.NotNil(t, result)
		assert.Equal(t, float32(77.77), *result)
		assert.IsType(t, (*float32)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := Float64Ptr().Default(float64(88.88))
		var nilPtr *float64 = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, float64(88.88), *result)
	})

	t.Run("special float values", func(t *testing.T) {
		schema := Float64()

		// Test infinity
		infResult := schema.MustStrictParse(math.Inf(1))
		assert.True(t, math.IsInf(infResult, 1))

		// Test NaN
		nanResult := schema.MustStrictParse(math.NaN())
		assert.True(t, math.IsNaN(nanResult))
	})
}
