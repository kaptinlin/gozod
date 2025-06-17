package gozod

import (
	"fmt"
	"math"
	"math/cmplx"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestComplexBasicFunctionality(t *testing.T) {
	t.Run("basic validation complex128", func(t *testing.T) {
		schema := Complex128()
		// Valid complex128
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
		// Invalid type
		_, err = schema.Parse("not a complex")
		assert.Error(t, err)
	})

	t.Run("basic validation complex64", func(t *testing.T) {
		schema := Complex64()
		// Valid complex64
		result, err := schema.Parse(complex64(complex(3.0, 4.0)))
		require.NoError(t, err)
		assert.Equal(t, complex64(complex(3.0, 4.0)), result)
		// Invalid type
		_, err = schema.Parse("not a complex")
		assert.Error(t, err)
	})

	t.Run("smart type inference complex128", func(t *testing.T) {
		schema := Complex128()
		// complex128 input returns complex128
		result1, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.IsType(t, complex128(0), result1)
		assert.Equal(t, complex(3.0, 4.0), result1)
		// Pointer input returns same pointer
		val := complex(2.0, 3.0)
		result2, err := schema.Parse(&val)
		require.NoError(t, err)
		assert.IsType(t, (*complex128)(nil), result2)
		assert.Equal(t, &val, result2)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		schema := Complex128()
		input := complex(3.0, 4.0)
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify exact pointer identity is preserved
		resultPtr, ok := result.(*complex128)
		require.True(t, ok, "Result should be *complex128")
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
		assert.Equal(t, complex(3.0, 4.0), *resultPtr)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Complex128().Nilable()
		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*complex128)(nil), result)
		// Valid input keeps type inference
		result2, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result2)
		assert.IsType(t, complex128(0), result2)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Complex128()
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse(complex(3.0, 4.0))
		require.NoError(t, err2)
		assert.Equal(t, complex(3.0, 4.0), result2)

		// 🔥 Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse(complex(3.0, 4.0))
		require.NoError(t, err5)
		assert.Equal(t, complex(3.0, 4.0), result5)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestComplexCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := Complex128(SchemaParams{Coerce: true})
		tests := []struct {
			input    interface{}
			expected complex128
		}{
			{int(42), complex(42.0, 0.0)},
			{float64(3.14), complex(3.14, 0.0)},
			{complex64(complex(1.0, 2.0)), complex(1.0, 2.0)},
			{"3+4i", complex(3.0, 4.0)},
			{"5", complex(5.0, 0.0)},
		}
		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("failed coercion", func(t *testing.T) {
		schema := Complex128(SchemaParams{Coerce: true})
		invalidInputs := []interface{}{
			"not a number",
			true,                            // boolean
			[]complex128{complex(1.0, 2.0)}, // slice
			map[string]complex128{"key": complex(1.0, 2.0)}, // map
		}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Should fail to coerce %v", input)
		}
	})

	t.Run("cross-type coercion", func(t *testing.T) {
		// Test complex64 to complex128 coercion
		schema := Complex128(SchemaParams{Coerce: true})
		result, err := schema.Parse(complex64(complex(3.0, 4.0)))
		require.NoError(t, err)
		assert.IsType(t, complex128(0), result)
		assert.Equal(t, complex(3.0, 4.0), result)
	})
}

// =============================================================================
// 3. Validation methods (complex-specific)
// =============================================================================

func TestComplexValidations(t *testing.T) {
	t.Run("magnitude validations", func(t *testing.T) {
		schema := Complex128().Min(1.0)
		// Valid magnitude
		result, err := schema.Parse(complex(3.0, 4.0)) // |3+4i| = 5
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
		// Invalid magnitude (too small)
		_, err = schema.Parse(complex(0.1, 0.1)) // |0.1+0.1i| ≈ 0.14
		assert.Error(t, err)
	})

	t.Run("positive magnitude validation", func(t *testing.T) {
		schema := Complex128().Positive()
		// Valid positive magnitude
		result, err := schema.Parse(complex(3.0, 4.0)) // |3+4i| = 5 > 0
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
		// Zero magnitude should fail
		_, err = schema.Parse(complex(0.0, 0.0))
		assert.Error(t, err)
	})

	t.Run("non-negative magnitude validation", func(t *testing.T) {
		schema := Complex128().NonNegative()
		// Valid non-negative magnitude
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
		// Zero magnitude should pass
		result2, err := schema.Parse(complex(0.0, 0.0))
		require.NoError(t, err)
		assert.Equal(t, complex(0.0, 0.0), result2)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestComplexModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schema := Complex128().Optional()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result2)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Complex128()
		// Valid input should not panic
		result := schema.MustParse(complex(3.0, 4.0))
		assert.Equal(t, complex(3.0, 4.0), result)
		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestComplexChaining(t *testing.T) {
	t.Run("multiple validations", func(t *testing.T) {
		schema := Complex128().Min(1.0).Max(10.0).Positive()
		// Valid input
		result, err := schema.Parse(complex(5.0, 3.0))
		require.NoError(t, err)
		assert.Equal(t, complex(5.0, 3.0), result)
		// Validation failures
		testCases := []complex128{
			complex(0.1, 0.1),  // magnitude too small
			complex(15.0, 3.0), // magnitude too large
		}
		for _, input := range testCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("magnitude constraints", func(t *testing.T) {
		schema := Complex128().Min(1.0).Max(5.0)
		// Valid input (proper magnitude)
		result, err := schema.Parse(complex(3.0, 4.0)) // |3+4i| = 5
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
		// Invalid magnitude
		_, err = schema.Parse(complex(0.1, 0.1)) // magnitude too small
		assert.Error(t, err)
		// Invalid magnitude
		_, err = schema.Parse(complex(10.0, 10.0)) // magnitude too large
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestComplexTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Complex128().Transform(func(val complex128, ctx *RefinementContext) (any, error) {
			// Complex conjugate
			return cmplx.Conj(val), nil
		})
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, -4.0), result)
	})

	t.Run("mathematical transforms", func(t *testing.T) {
		schema := Complex128().Transform(func(val complex128, ctx *RefinementContext) (any, error) {
			return map[string]any{
				"original":  val,
				"conjugate": cmplx.Conj(val),
				"magnitude": cmplx.Abs(val),
				"phase":     cmplx.Phase(val),
				"real":      real(val),
				"imaginary": imag(val),
			}, nil
		})

		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, complex(3.0, 4.0), resultMap["original"])
		assert.Equal(t, complex(3.0, -4.0), resultMap["conjugate"])
		assert.Equal(t, 5.0, resultMap["magnitude"])
		assert.InDelta(t, math.Atan2(4.0, 3.0), resultMap["phase"], 0.001)
		assert.Equal(t, 3.0, resultMap["real"])
		assert.Equal(t, 4.0, resultMap["imaginary"])
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestComplexRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Complex128().Refine(func(val complex128) bool {
			// Only allow pure real numbers
			return imag(val) == 0
		}, SchemaParams{
			Error: "Number must be real",
		})
		result, err := schema.Parse(complex(42.0, 0.0))
		require.NoError(t, err)
		assert.Equal(t, complex(42.0, 0.0), result)
		_, err = schema.Parse(complex(3.0, 4.0))
		assert.Error(t, err)
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := complex(3.0, 4.0)

		// Refine: only validates, never modifies
		refineSchema := Complex128().Refine(func(val complex128) bool {
			return cmplx.Abs(val) > 1.0
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Complex128().Transform(func(val complex128, ctx *RefinementContext) (any, error) {
			return cmplx.Conj(val), nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, complex(3.0, 4.0), refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, complex(3.0, -4.0), transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input, transformResult, "Transform should return modified value")
	})

	t.Run("unit circle validation", func(t *testing.T) {
		schema := Complex128().Refine(func(val complex128) bool {
			// Check if complex number is on or near unit circle
			magnitude := cmplx.Abs(val)
			return math.Abs(magnitude-1.0) < 0.01
		}, SchemaParams{
			Error: func(issue ZodRawIssue) string {
				if input, ok := issue.Input.(complex128); ok {
					magnitude := cmplx.Abs(input)
					return fmt.Sprintf("Complex number %.3f+%.3fi has magnitude %.3f, not on unit circle", real(input), imag(input), magnitude)
				}
				return "Invalid input for unit circle validation"
			},
		})

		// On unit circle
		result, err := schema.Parse(complex(math.Cos(math.Pi/4), math.Sin(math.Pi/4)))
		require.NoError(t, err)
		assert.InDelta(t, math.Sqrt(2)/2, real(result.(complex128)), 0.001)

		// Not on unit circle
		_, err = schema.Parse(complex(2.0, 0.0))
		assert.Error(t, err)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestComplexErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Complex128().Min(5.1) // Use 5.1 so that |3+4i| = 5.0 fails
		_, err := schema.Parse(complex(3.0, 4.0))
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
	})

	t.Run("type mismatch error", func(t *testing.T) {
		schema := Complex128()
		_, err := schema.Parse("not a complex")
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Equal(t, string(InvalidType), zodErr.Issues[0].Code)
	})

	t.Run("special values error handling", func(t *testing.T) {
		schema := Complex128().Positive()
		specialValues := []complex128{
			complex(math.Inf(1), 0.0),  // +Inf real
			complex(0.0, math.Inf(-1)), // -Inf imag
			complex(math.NaN(), 0.0),   // NaN real
			complex(0.0, math.NaN()),   // NaN imag
		}
		for _, val := range specialValues {
			_, err := schema.Parse(val)
			// Note: These special values may or may not error depending on implementation
			// Just check that parsing doesn't panic
			_ = err
		}
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestComplexEdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		schema := Complex128()
		result, err := schema.Parse(complex(0.0, 0.0))
		require.NoError(t, err)
		assert.Equal(t, complex(0.0, 0.0), result)
	})

	t.Run("pure real and imaginary numbers", func(t *testing.T) {
		schema := Complex128()
		// Pure real
		result1, err := schema.Parse(complex(5.0, 0.0))
		require.NoError(t, err)
		assert.Equal(t, complex(5.0, 0.0), result1)
		// Pure imaginary
		result2, err := schema.Parse(complex(0.0, 5.0))
		require.NoError(t, err)
		assert.Equal(t, complex(0.0, 5.0), result2)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := Complex128()
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
		schema := Complex128()
		invalidTypes := []interface{}{
			"3+4i",
			3.14,
			42,
			true,
			[]complex128{complex(1.0, 2.0)},
			map[string]complex128{"key": complex(1.0, 2.0)},
		}
		for _, invalidType := range invalidTypes {
			_, err := schema.Parse(invalidType)
			assert.Error(t, err, "Expected error for type %T", invalidType)
		}
	})

	t.Run("cross-precision compatibility", func(t *testing.T) {
		// Test that different complex types don't accidentally validate
		schemaComplex128 := Complex128()
		_, err := schemaComplex128.Parse(complex64(complex(3.0, 4.0)))
		assert.Error(t, err, "complex128 schema should reject complex64 without coercion")

		schemaComplex64 := Complex64()
		_, err = schemaComplex64.Parse(complex(3.0, 4.0)) // complex128
		assert.Error(t, err, "complex64 schema should reject complex128 without coercion")
	})

	t.Run("precision edge cases", func(t *testing.T) {
		schema := Complex128()
		// Test very small differences
		val1 := complex(1.0, 1.0)
		val2 := complex(1.0+1e-15, 1.0+1e-15)

		result1, err := schema.Parse(val1)
		require.NoError(t, err)
		assert.Equal(t, val1, result1)

		result2, err := schema.Parse(val2)
		require.NoError(t, err)
		assert.Equal(t, val2, result2)
		assert.NotEqual(t, result1, result2)
	})

	t.Run("boundary magnitude calculations", func(t *testing.T) {
		schema := Complex128().Min(5.0).Max(5.0)
		// Exactly magnitude 5.0
		result, err := schema.Parse(complex(3.0, 4.0)) // |3+4i| = 5
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Also magnitude 5.0 but different components
		result2, err := schema.Parse(complex(0.0, 5.0)) // |0+5i| = 5
		require.NoError(t, err)
		assert.Equal(t, complex(0.0, 5.0), result2)

		// Slightly off
		_, err = schema.Parse(complex(3.1, 4.0)) // |3.1+4i| ≈ 5.015
		assert.Error(t, err)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestComplexDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := Complex128().Default(complex(1.0, 1.0))
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, complex(1.0, 1.0), result)
		// Valid input normal validation
		result2, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result2)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := Complex128().DefaultFunc(func() complex128 {
			counter++
			return complex(float64(counter), float64(counter))
		})

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, complex(1.0, 1.0), result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, complex(2.0, 2.0), result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse(complex(10.0, 10.0))
		require.NoError(t, err3)
		assert.Equal(t, complex(10.0, 10.0), result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("mathematical constants as defaults", func(t *testing.T) {
		tests := []struct {
			name     string
			schema   ZodComplexDefault[complex128]
			expected complex128
		}{
			{"unit imaginary", Complex128().Default(complex(0.0, 1.0)), complex(0.0, 1.0)},
			{"golden ratio", Complex128().Default(complex((1.0+math.Sqrt(5.0))/2.0, 0.0)), complex((1.0+math.Sqrt(5.0))/2.0, 0.0)},
			{"euler's identity", Complex128().Default(cmplx.Exp(complex(0.0, math.Pi))), cmplx.Exp(complex(0.0, math.Pi))},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(nil)
				require.NoError(t, err)
				if tt.name == "euler's identity" {
					assert.InDelta(t, real(tt.expected), real(result.(complex128)), 0.001)
					assert.InDelta(t, imag(tt.expected), imag(result.(complex128)), 0.001)
				} else {
					assert.Equal(t, tt.expected, result)
				}
			})
		}
	})

	t.Run("prefault value", func(t *testing.T) {
		fallbackValue := complex(10.0, 0.0)
		schema := Complex128().Refine(func(val complex128) bool {
			return real(val) > 2.0 // Only allow real part > 2
		}).Prefault(fallbackValue)

		// Valid input succeeds
		result1, err1 := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err1)
		assert.Equal(t, complex(3.0, 4.0), result1)

		// Invalid input uses prefault (real part too small)
		result2, err2 := schema.Parse(complex(1.0, 1.0))
		require.NoError(t, err2)
		assert.Equal(t, fallbackValue, result2)
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Complex128().Refine(func(val complex128) bool {
			return real(val) > 2.0 // Only allow real part > 2
		}).PrefaultFunc(func() complex128 {
			counter++
			return complex(float64(counter*10), 0.0) // Generate different fallback values
		})

		// Valid input succeeds and doesn't call function
		result1, err1 := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err1)
		assert.Equal(t, complex(3.0, 4.0), result1)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input calls prefault function (real part too small)
		result2, err2 := schema.Parse(complex(1.0, 1.0))
		require.NoError(t, err2)
		assert.Equal(t, complex(10.0, 0.0), result2)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")

		// Another invalid input calls function again
		result3, err3 := schema.Parse(complex(0.5, 0.5))
		require.NoError(t, err3)
		assert.Equal(t, complex(20.0, 0.0), result3)
		assert.Equal(t, 2, counter, "Function should increment counter for each invalid input")

		// Valid input still doesn't call function
		result4, err4 := schema.Parse(complex(4.0, 3.0))
		require.NoError(t, err4)
		assert.Equal(t, complex(4.0, 3.0), result4)
		assert.Equal(t, 2, counter, "Counter should remain unchanged for valid input")
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := complex(0.0, 0.0)
		prefaultValue := complex(100.0, 0.0)

		schema := Complex128().
			Refine(func(val complex128) bool {
				return real(val) > 2.0 // Only allow real part > 2
			}).
			Default(defaultValue).
			Prefault(prefaultValue)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, defaultValue, result1)

		// Valid input succeeds
		result2, err2 := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err2)
		assert.Equal(t, complex(3.0, 4.0), result2)

		// Invalid input uses prefault
		result3, err3 := schema.Parse(complex(1.0, 1.0))
		require.NoError(t, err3)
		assert.Equal(t, prefaultValue, result3)
	})
}

// =============================================================================
// Additional type-specific tests
// =============================================================================

func TestComplexTypeSpecific(t *testing.T) {
	t.Run("all complex types basic validation", func(t *testing.T) {
		tests := []struct {
			name   string
			schema ZodType[any, any]
			input  any
			want   any
		}{
			{"complex64", Complex64(), complex64(complex(3.0, 4.0)), complex64(complex(3.0, 4.0))},
			{"complex128", Complex128(), complex(3.0, 4.0), complex(3.0, 4.0)},
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
			schema ZodType[any, any]
			input  any
			want   any
		}{
			{"coerced complex64", CoercedComplex64(), "3+4i", complex64(complex(3.0, 4.0))},
			{"coerced complex128", CoercedComplex128(), "3+4i", complex(3.0, 4.0)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(tt.input)
				require.NoError(t, err)
				if c64, ok := tt.want.(complex64); ok {
					resultC64, ok := result.(complex64)
					require.True(t, ok)
					assert.InDelta(t, real(c64), real(resultC64), 0.001)
					assert.InDelta(t, imag(c64), imag(resultC64), 0.001)
				} else {
					assert.Equal(t, tt.want, result)
				}
			})
		}
	})

	t.Run("polar coordinate conversions", func(t *testing.T) {
		schema := Complex128().Transform(func(val complex128, ctx *RefinementContext) (any, error) {
			magnitude := cmplx.Abs(val)
			phase := cmplx.Phase(val)
			return map[string]float64{
				"magnitude": magnitude,
				"phase":     phase,
				"degrees":   phase * 180.0 / math.Pi,
			}, nil
		})

		result, err := schema.Parse(complex(1.0, 1.0))
		require.NoError(t, err)
		resultMap, ok := result.(map[string]float64)
		require.True(t, ok)
		assert.InDelta(t, math.Sqrt(2), resultMap["magnitude"], 0.001)
		assert.InDelta(t, math.Pi/4, resultMap["phase"], 0.001)
		assert.InDelta(t, 45.0, resultMap["degrees"], 0.001)
	})

	t.Run("quadrant validation", func(t *testing.T) {
		schema := Complex128().Refine(func(val complex128) bool {
			// Only allow first quadrant (both real and imaginary positive)
			return real(val) >= 0 && imag(val) >= 0
		}, SchemaParams{
			Error: "Complex number must be in first quadrant",
		})

		// First quadrant
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Other quadrants
		invalidInputs := []complex128{
			complex(-3.0, 4.0),  // second quadrant
			complex(-3.0, -4.0), // third quadrant
			complex(3.0, -4.0),  // fourth quadrant
		}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("complex arithmetic operations", func(t *testing.T) {
		schema := Complex128().Transform(func(val complex128, ctx *RefinementContext) (any, error) {
			return map[string]complex128{
				"original": val,
				"squared":  val * val,
				"sqrt":     cmplx.Sqrt(val),
				"exp":      cmplx.Exp(val),
				"log":      cmplx.Log(val),
			}, nil
		})

		result, err := schema.Parse(complex(1.0, 0.0)) // Real number 1
		require.NoError(t, err)
		resultMap, ok := result.(map[string]complex128)
		require.True(t, ok)
		assert.Equal(t, complex(1.0, 0.0), resultMap["original"])
		assert.InDelta(t, 1.0, real(resultMap["squared"]), 0.001)
		assert.InDelta(t, 1.0, real(resultMap["sqrt"]), 0.001)
		assert.InDelta(t, math.E, real(resultMap["exp"]), 0.001)
		assert.InDelta(t, 0.0, real(resultMap["log"]), 0.001)
	})
}
