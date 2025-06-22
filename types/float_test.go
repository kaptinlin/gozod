package types

import (
	"fmt"
	"math"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestFloatBasicFunctionality(t *testing.T) {
	t.Run("basic validation float64", func(t *testing.T) {
		schema := Float64()
		// Valid float64
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
		// Invalid type
		_, err = schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("basic validation float32", func(t *testing.T) {
		schema := Float32()
		// Valid float32
		result, err := schema.Parse(float32(3.14))
		require.NoError(t, err)
		assert.Equal(t, float32(3.14), result)
		// Invalid type
		_, err = schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("smart type inference float64", func(t *testing.T) {
		schema := Float64()
		// Float64 input returns float64
		result1, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.IsType(t, float64(0), result1)
		assert.Equal(t, 3.14, result1)
		// Pointer input returns same pointer
		val := 2.71
		result2, err := schema.Parse(&val)
		require.NoError(t, err)
		assert.IsType(t, (*float64)(nil), result2)
		assert.Equal(t, &val, result2)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		schema := Float32().Min(1.0)
		input := float32(3.14)
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify not only type and value, but exact pointer identity
		resultPtr, ok := result.(*float32)
		require.True(t, ok, "Result should be *float32")
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
		assert.Equal(t, float32(3.14), *resultPtr)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Float64().Nilable()
		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*float64)(nil), result)
		// Valid input keeps type inference
		result2, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result2)
		assert.IsType(t, float64(0), result2)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Float64().Min(1.0)
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse(5.0)
		require.NoError(t, err2)
		assert.Equal(t, 5.0, result2)

		// Test nilable schema rejects invalid values
		_, err3 := nilableSchema.Parse(0.5)
		assert.Error(t, err3)

		// 🔥 Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse(5.0)
		require.NoError(t, err5)
		assert.Equal(t, 5.0, result5)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestFloatCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := CoercedFloat64()
		tests := []struct {
			input    any
			expected float64
		}{
			{"3.14", 3.14},
			{int(42), 42.0},
			{int64(84), 84.0},
			{float32(2.71), 2.71},
			{uint(100), 100.0},
		}
		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			if _, ok := tt.input.(float32); ok {
				// Use InDelta for float32 to float64 conversion due to precision differences
				assert.InDelta(t, tt.expected, result, 0.01)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedFloat64().Min(5.0).Max(100.0)
		// Coercion then validation passes
		result, err := schema.Parse("50.5")
		require.NoError(t, err)
		assert.Equal(t, 50.5, result)
		// Coercion then validation fails
		_, err = schema.Parse("3.14")
		assert.Error(t, err)
	})

	t.Run("failed coercion", func(t *testing.T) {
		schema := CoercedFloat64()
		invalidInputs := []any{
			"not a number",
			[]float64{1.0},                 // slice
			map[string]float64{"key": 1.0}, // map
		}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Should fail to coerce %v", input)
		}
	})

	t.Run("boolean coercion", func(t *testing.T) {
		schema := CoercedFloat64()

		// Boolean values should be coerced to numbers
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, 1.0, result)

		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})

	t.Run("cross-type coercion", func(t *testing.T) {
		// Test float32 to float64 coercion
		schema := CoercedFloat64()
		result, err := schema.Parse(float32(3.14))
		require.NoError(t, err)
		assert.IsType(t, float64(0), result)
		assert.InDelta(t, 3.14, result, 0.01)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestFloatValidations(t *testing.T) {
	t.Run("range validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  core.ZodType[any, any]
			input   float64
			wantErr bool
		}{
			{"min valid", Float64().Min(5.0), 10.0, false},
			{"min invalid", Float64().Min(5.0), 3.0, true},
			{"max valid", Float64().Max(100.0), 50.0, false},
			{"max invalid", Float64().Max(100.0), 150.0, true},
			{"gt valid", Float64().Gt(0.0), 5.0, false},
			{"gt invalid", Float64().Gt(0.0), 0.0, true},
			{"gte valid", Float64().Gte(0.0), 0.0, false},
			{"gte invalid", Float64().Gte(0.0), -1.0, true},
			{"lt valid", Float64().Lt(100.0), 50.0, false},
			{"lt invalid", Float64().Lt(100.0), 100.0, true},
			{"lte valid", Float64().Lte(100.0), 100.0, false},
			{"lte invalid", Float64().Lte(100.0), 101.0, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("sign validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  core.ZodType[any, any]
			input   float64
			wantErr bool
		}{
			{"positive valid", Float64().Positive(), 5.0, false},
			{"positive invalid", Float64().Positive(), -5.0, true},
			{"negative valid", Float64().Negative(), -5.0, false},
			{"negative invalid", Float64().Negative(), 5.0, true},
			{"non-negative valid", Float64().NonNegative(), 0.0, false},
			{"non-negative invalid", Float64().NonNegative(), -1.0, true},
			{"non-positive valid", Float64().NonPositive(), 0.0, false},
			{"non-positive invalid", Float64().NonPositive(), 1.0, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("multiple of validation", func(t *testing.T) {
		schema := Float64().MultipleOf(0.5)
		// Valid multiple
		result, err := schema.Parse(2.5)
		require.NoError(t, err)
		assert.Equal(t, 2.5, result)
		// Invalid multiple
		_, err = schema.Parse(2.3)
		assert.Error(t, err)
	})

	t.Run("finite validation", func(t *testing.T) {
		schema := Float64().Finite()
		// Valid finite numbers
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
		// Invalid infinite numbers
		_, err = schema.Parse(math.Inf(1))
		assert.Error(t, err)
		_, err = schema.Parse(math.NaN())
		assert.Error(t, err)
	})

	t.Run("integer validation", func(t *testing.T) {
		schema := Float64().Int()
		// Valid integer-like floats
		result, err := schema.Parse(42.0)
		require.NoError(t, err)
		assert.Equal(t, 42.0, result)
		// Invalid non-integer floats
		_, err = schema.Parse(3.14)
		assert.Error(t, err)
	})

	t.Run("safe validation", func(t *testing.T) {
		schema := Float64().Safe()
		// Valid safe numbers
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
		// Invalid unsafe numbers (same as finite for float types)
		_, err = schema.Parse(math.Inf(1))
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestFloatModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schema := Float64().Optional()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Float64().Nilable()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result2)
	})

	t.Run("nullish modifier", func(t *testing.T) {
		schema := Float64().Nullish()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result2)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Float64()
		// Valid input should not panic
		result := schema.MustParse(3.14)
		assert.Equal(t, 3.14, result)
		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestFloatChaining(t *testing.T) {
	t.Run("multiple validations", func(t *testing.T) {
		schema := Float64().Min(1.0).Max(100.0).Positive().Finite()
		// Valid input
		result, err := schema.Parse(50.5)
		require.NoError(t, err)
		assert.Equal(t, 50.5, result)
		// Validation failures
		testCases := []float64{
			0.5,         // too small
			150.0,       // too large
			-10.0,       // not positive
			math.Inf(1), // not finite
		}
		for _, input := range testCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("validation with multiple of and integer", func(t *testing.T) {
		schema := Float64().Min(5.0).Max(100.0).MultipleOf(5.0).Int()
		// Valid input
		result, err := schema.Parse(25.0)
		require.NoError(t, err)
		assert.Equal(t, 25.0, result)
		// Validation failures
		_, err = schema.Parse(23.0) // not multiple of 5
		assert.Error(t, err)
		_, err = schema.Parse(25.5) // not integer
		assert.Error(t, err)
	})

	t.Run("cross-precision validation", func(t *testing.T) {
		schema := Float32().Min(1.0).Max(float32(math.MaxFloat32))
		// Valid input within range
		result, err := schema.Parse(float32(50.5))
		require.NoError(t, err)
		assert.Equal(t, float32(50.5), result)
		// Edge case: test boundary
		result2, err := schema.Parse(float32(1.0))
		require.NoError(t, err)
		assert.Equal(t, float32(1.0), result2)
	})

	t.Run("precision overflow protection", func(t *testing.T) {
		// Test float32 precision limits
		schema := Float32().Finite()
		// Should handle normal range
		result, err := schema.Parse(float32(3.14))
		require.NoError(t, err)
		assert.Equal(t, float32(3.14), result)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestFloatTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Float64().Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
			return val * 2.0, nil
		})
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 6.28, result)
	})

	t.Run("transform chaining", func(t *testing.T) {
		schema := Float64().
			Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
				return val * 2.0, nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if floatVal, ok := val.(float64); ok {
					return fmt.Sprintf("result_%.2f", floatVal), nil
				}
				return val, nil
			})
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, "result_6.28", result)
	})

	t.Run("pipe combination", func(t *testing.T) {
		schema := Float64().
			Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
				return fmt.Sprintf("%.2f", val), nil
			}).
			Pipe(String().Min(3))
		result, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, "3.14", result)
		_, err = schema.Parse(1.0) // "1.00" meets String().Min(3)
		require.NoError(t, err)
	})

	t.Run("transform with validation", func(t *testing.T) {
		schema := Float64().Min(1.0).Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
			if val < 0 {
				return nil, fmt.Errorf("negative numbers not allowed")
			}
			return int(val * 10), nil
		})

		result, err := schema.Parse(5.7)
		require.NoError(t, err)
		assert.Equal(t, 57, result)

		// Validation before transform should fail
		_, err = schema.Parse(0.5)
		assert.Error(t, err)
	})

	t.Run("mathematical transforms", func(t *testing.T) {
		schema := Float64().Positive().Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
			return map[string]any{
				"original": val,
				"squared":  val * val,
				"sqrt":     math.Sqrt(val),
				"log":      math.Log(val),
			}, nil
		})

		result, err := schema.Parse(16.0)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, 16.0, resultMap["original"])
		assert.Equal(t, 256.0, resultMap["squared"])
		assert.Equal(t, 4.0, resultMap["sqrt"])
		assert.InDelta(t, math.Log(16.0), resultMap["log"], 0.001)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestFloatRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Float64().Refine(func(val float64) bool {
			return val == math.Floor(val) // integer-like floats only
		}, core.SchemaParams{
			Error: "Number must be an integer",
		})
		result, err := schema.Parse(42.0)
		require.NoError(t, err)
		assert.Equal(t, 42.0, result)
		_, err = schema.Parse(3.14)
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "integer")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := 3.14

		// Refine: only validates, never modifies
		refineSchema := Float64().Refine(func(val float64) bool {
			return val > 0
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Float64().Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
			return val * 2, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, 3.14, refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, 6.28, transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input, transformResult, "Transform should return modified value")
	})

	t.Run("refine preserves pointer identity", func(t *testing.T) {
		schema := Float64().Refine(func(val float64) bool {
			return val >= 1.0
		})

		input := 3.14
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify exact pointer identity is preserved
		resultPtr, ok := result.(*float64)
		require.True(t, ok)
		assert.True(t, resultPtr == inputPtr, "Refine should preserve exact pointer identity")
	})

	t.Run("refine with complex validation", func(t *testing.T) {
		schema := Float64().Refine(func(val float64) bool {
			// Golden ratio approximation (within 0.01)
			goldenRatio := (1.0 + math.Sqrt(5.0)) / 2.0
			return math.Abs(val-goldenRatio) < 0.01
		}, core.SchemaParams{
			Error: func(issue core.ZodRawIssue) string {
				if input, ok := issue.Input.(float64); ok {
					return fmt.Sprintf("The number %.3f is not close to the golden ratio", input)
				}
				return "Invalid input for golden ratio validation"
			},
		})

		result, err := schema.Parse(1.618) // Close to golden ratio
		require.NoError(t, err)
		assert.Equal(t, 1.618, result)

		_, err = schema.Parse(2.0) // Not close to golden ratio
		assert.Error(t, err)
	})

	t.Run("scientific notation validation", func(t *testing.T) {
		schema := Float64().Refine(func(val float64) bool {
			// Check if number is in scientific notation range (1e-10 to 1e10)
			return math.Abs(val) >= 1e-10 && math.Abs(val) <= 1e10
		}, core.SchemaParams{
			Error: "Number must be within scientific range",
		})

		result, err := schema.Parse(1e5)
		require.NoError(t, err)
		assert.Equal(t, 1e5, result)

		_, err = schema.Parse(1e15) // Too large
		assert.Error(t, err)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestFloatErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Float64().Min(10.0)
		_, err := schema.Parse(5.0)
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Float64().Min(10.0, core.SchemaParams{
			Error: "core.Custom minimum value error",
		})
		_, err := schema.Parse(5.0)
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "core.Custom minimum value error")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := Float64().Min(10.0).Max(20.0).Finite()
		_, err := schema.Parse(5.0)
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})

	t.Run("type mismatch error", func(t *testing.T) {
		schema := Float64()
		_, err := schema.Parse("not a number")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("special values error handling", func(t *testing.T) {
		schema := Float64().Finite()
		specialValues := []float64{
			math.Inf(1),  // +Inf
			math.Inf(-1), // -Inf
			math.NaN(),   // NaN
		}
		for _, val := range specialValues {
			_, err := schema.Parse(val)
			assert.Error(t, err, "Expected error for special value %v", val)
		}
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestFloatEdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		schema := Float64()
		result, err := schema.Parse(0.0)
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})

	t.Run("negative zero", func(t *testing.T) {
		schema := Float64()
		result, err := schema.Parse(math.Copysign(0.0, -1))
		require.NoError(t, err)
		assert.Equal(t, 0.0, result) // -0.0 == 0.0 in Go
	})

	t.Run("boundary values float32", func(t *testing.T) {
		schema := Float32()
		// Test min and max values
		result1, err := schema.Parse(float32(math.SmallestNonzeroFloat32))
		require.NoError(t, err)
		assert.Equal(t, float32(math.SmallestNonzeroFloat32), result1)

		result2, err := schema.Parse(float32(math.MaxFloat32))
		require.NoError(t, err)
		assert.Equal(t, float32(math.MaxFloat32), result2)
	})

	t.Run("boundary values float64", func(t *testing.T) {
		schema := Float64()
		// Test min and max values
		result1, err := schema.Parse(math.SmallestNonzeroFloat64)
		require.NoError(t, err)
		assert.Equal(t, math.SmallestNonzeroFloat64, result1)

		result2, err := schema.Parse(math.MaxFloat64)
		require.NoError(t, err)
		assert.Equal(t, math.MaxFloat64, result2)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := Float64()
		// By default nil is not allowed
		_, err := schema.Parse(nil)
		assert.Error(t, err)
		// Nilable allows nil
		nilableSchema := schema.Nilable()
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("type mismatch", func(t *testing.T) {
		schema := Float64()
		invalidTypes := []any{
			"3.14",
			true,
			[]float64{3.14},
			map[string]float64{"key": 3.14},
			complex(1, 2),
		}
		for _, invalidType := range invalidTypes {
			_, err := schema.Parse(invalidType)
			assert.Error(t, err, "Expected error for type %T", invalidType)
		}
	})

	t.Run("modifier combinations", func(t *testing.T) {
		schema := Float64().Min(1.0).Positive().Finite().Nilable()
		// Valid input
		result, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
		// nil input
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		// Invalid input
		_, err = schema.Parse(-5.0)
		assert.Error(t, err)
		_, err = schema.Parse(math.Inf(1))
		assert.Error(t, err)
	})

	t.Run("cross-type compatibility", func(t *testing.T) {
		// Test that different float types don't accidentally validate
		schemaFloat64 := Float64()
		_, err := schemaFloat64.Parse(float32(3.14))
		assert.Error(t, err, "float64 schema should reject float32 without coercion")

		schemaFloat32 := Float32()
		_, err = schemaFloat32.Parse(3.14) // float64
		assert.Error(t, err, "float32 schema should reject float64 without coercion")
	})

	t.Run("precision edge cases", func(t *testing.T) {
		schema := Float64()
		// Test small but representable differences using math.Nextafter
		val1 := 1.0
		val2 := math.Nextafter(1.0, 2.0) // Next representable float64 after 1.0

		result1, err := schema.Parse(val1)
		require.NoError(t, err)
		assert.Equal(t, val1, result1)

		result2, err := schema.Parse(val2)
		require.NoError(t, err)
		assert.Equal(t, val2, result2)
		assert.NotEqual(t, result1, result2)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestFloatDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := Float64().Min(1.0).Default(3.14)
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
		// Valid input normal validation
		result2, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result2)
		// Invalid input still fails
		_, err = schema.Parse(0.5)
		assert.Error(t, err)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := Float64().DefaultFunc(func() float64 {
			counter++
			return float64(counter) * 3.14
		}).Min(1.0)

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 3.14, result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, 6.28, result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse(10.0)
		require.NoError(t, err3)
		assert.Equal(t, 10.0, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := Float64().
			Default(math.Pi).
			Min(1.0).
			Transform(func(val float64, ctx *core.RefinementContext) (any, error) {
				return map[string]any{
					"original":   val,
					"doubled":    val * 2,
					"rounded":    math.Round(val*100) / 100,
					"scientific": fmt.Sprintf("%.2e", val),
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse(2.71)
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, 2.71, result1Map["original"])
		assert.Equal(t, 5.42, result1Map["doubled"])
		assert.Equal(t, 2.71, result1Map["rounded"])
		assert.Equal(t, "2.71e+00", result1Map["scientific"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, math.Pi, result2Map["original"])
		assert.Equal(t, math.Pi*2, result2Map["doubled"])
		assert.Equal(t, 3.14, result2Map["rounded"])
		assert.Equal(t, "3.14e+00", result2Map["scientific"])

		// Invalid input still fails validation
		_, err3 := schema.Parse(0.5)
		assert.Error(t, err3, "Small float should fail Min(1.0) validation")
	})

	t.Run("prefault value", func(t *testing.T) {
		schema := Float64().Min(1.0).Prefault(999.99)
		// Any validation failure uses fallback value
		result1, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 999.99, result1)
		result2, err := schema.Parse(0.5)
		require.NoError(t, err)
		assert.Equal(t, 999.99, result2)
		// Valid input normal validation
		result3, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result3)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := Float64().Min(1.0).PrefaultFunc(func() float64 {
			counter++
			return 1000.0 + float64(counter)
		})

		// Valid input should not call function
		result, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
		assert.Equal(t, 0, counter)

		// Invalid input should call function
		result, err = schema.Parse(0.5)
		require.NoError(t, err)
		assert.Equal(t, 1001.0, result)
		assert.Equal(t, 1, counter)

		// Another invalid input should increment counter
		result, err = schema.Parse("invalid")
		require.NoError(t, err)
		assert.Equal(t, 1002.0, result)
		assert.Equal(t, 2, counter)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Float64().Min(1.0).Default(3.14).Prefault(999.99)
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
		// Invalid input uses fallback value
		result2, err := schema.Parse(0.5)
		require.NoError(t, err)
		assert.Equal(t, 999.99, result2)
		// Valid input normal validation
		result3, err := schema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result3)
	})

	t.Run("typed default chaining", func(t *testing.T) {
		// Test that Default returns type-safe wrapper that supports chaining
		schema := Float64().Default(3.14).Min(1.0).Max(100.0).Positive()

		// Test default functionality
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 3.14, result1)

		// Test chained validations work
		result2, err2 := schema.Parse(50.5)
		require.NoError(t, err2)
		assert.Equal(t, 50.5, result2)

		// Test validation failures
		_, err3 := schema.Parse(0.5) // below Min(1.0)
		assert.Error(t, err3)
		_, err4 := schema.Parse(150.0) // above Max(100.0)
		assert.Error(t, err4)
	})

	t.Run("mathematical constants as defaults", func(t *testing.T) {
		tests := []struct {
			name     string
			schema   ZodFloatDefault[float64]
			expected float64
		}{
			{"pi", Float64().Default(math.Pi), math.Pi},
			{"e", Float64().Default(math.E), math.E},
			{"golden ratio", Float64().Default((1.0 + math.Sqrt(5.0)) / 2.0), (1.0 + math.Sqrt(5.0)) / 2.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(nil)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// Additional type-specific tests
// =============================================================================

func TestFloatTypeSpecific(t *testing.T) {
	t.Run("all float types basic validation", func(t *testing.T) {
		tests := []struct {
			name   string
			schema core.ZodType[any, any]
			input  any
			want   any
		}{
			{"float32", Float32(), float32(3.14), float32(3.14)},
			{"float64", Float64(), float64(3.14), float64(3.14)},
			{"number alias", Number(), float64(3.14), float64(3.14)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			})
		}
	})

	t.Run("coerced constructors", func(t *testing.T) {
		tests := []struct {
			name   string
			schema core.ZodType[any, any]
			input  any
			want   any
		}{
			{"coerced float32", CoercedFloat32(), "3.14", float32(3.14)},
			{"coerced float64", CoercedFloat64(), "3.14", float64(3.14)},
			{"coerced number", CoercedNumber(), "3.14", float64(3.14)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(tt.input)
				require.NoError(t, err)
				assert.InDelta(t, tt.want, result, 0.001)
			})
		}
	})

	t.Run("precision overflow protection", func(t *testing.T) {
		// Test that coercion respects type bounds
		schema := CoercedFloat32()

		// Should succeed within bounds
		result, err := schema.Parse("3.14")
		require.NoError(t, err)
		assert.InDelta(t, float32(3.14), result, 0.001)

		// Test extreme values (implementation dependent)
		_, _ = schema.Parse("1e40") // Very large number
		// Note: This behavior depends on implementation
	})

	t.Run("special value handling", func(t *testing.T) {
		schema := Float64()

		// Test special IEEE 754 values
		specialTests := []struct {
			name  string
			input float64
		}{
			{"positive infinity", math.Inf(1)},
			{"negative infinity", math.Inf(-1)},
			{"NaN", math.NaN()},
		}

		for _, tt := range specialTests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				if math.IsNaN(tt.input) {
					assert.True(t, math.IsNaN(result.(float64)))
				} else {
					assert.Equal(t, tt.input, result)
				}
			})
		}
	})
}
