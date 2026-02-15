package types

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestComplex_BasicFunctionality(t *testing.T) {
	t.Run("valid complex inputs", func(t *testing.T) {
		schema := Complex128()

		// Test positive real and imaginary values
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Test negative values
		result, err = schema.Parse(complex(-2.0, -1.0))
		require.NoError(t, err)
		assert.Equal(t, complex(-2.0, -1.0), result)

		// Test zero
		result, err = schema.Parse(complex(0.0, 0.0))
		require.NoError(t, err)
		assert.Equal(t, complex(0.0, 0.0), result)

		// Test purely real
		result, err = schema.Parse(complex(5.0, 0.0))
		require.NoError(t, err)
		assert.Equal(t, complex(5.0, 0.0), result)

		// Test purely imaginary
		result, err = schema.Parse(complex(0.0, 3.0))
		require.NoError(t, err)
		assert.Equal(t, complex(0.0, 3.0), result)
	})

	t.Run("different complex types", func(t *testing.T) {
		// Test complex64
		schema64 := Complex64()
		result64, err := schema64.Parse(complex64(complex(1.5, 2.5)))
		require.NoError(t, err)
		assert.Equal(t, complex64(complex(1.5, 2.5)), result64)

		// Test complex128
		schema128 := Complex128()
		result128, err := schema128.Parse(complex128(complex(3.5, 4.5)))
		require.NoError(t, err)
		assert.Equal(t, complex128(complex(3.5, 4.5)), result128)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Complex128()

		invalidInputs := []any{
			"not a complex", true, 42, 3.14, []complex128{1 + 2i}, map[string]complex128{"key": 1 + 2i}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Complex128()

		// Test Parse method
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Test MustParse method
		mustResult := schema.MustParse(complex(5.0, 6.0))
		assert.Equal(t, complex(5.0, 6.0), mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a complex value"
		schema := Complex128(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeComplex128, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestComplex_TypeSafety(t *testing.T) {
	t.Run("Complex64 returns complex64 type", func(t *testing.T) {
		schema := Complex64()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex64(complex(1.5, 2.5)))
		require.NoError(t, err)
		assert.Equal(t, complex64(complex(1.5, 2.5)), result)
		assert.IsType(t, complex64(0), result)
	})

	t.Run("Complex128 returns complex128 type", func(t *testing.T) {
		schema := Complex128()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex128(complex(3.5, 4.5)))
		require.NoError(t, err)
		assert.Equal(t, complex128(complex(3.5, 4.5)), result)
		assert.IsType(t, complex128(0), result)
	})

	t.Run("Complex64Ptr returns *complex64 type", func(t *testing.T) {
		schema := Complex64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex64(complex(1.5, 2.5)))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex64(complex(1.5, 2.5)), *result)
		assert.IsType(t, (*complex64)(nil), result)
	})

	t.Run("Complex128Ptr returns *complex128 type", func(t *testing.T) {
		schema := Complex128Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex128(complex(3.5, 4.5)))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex128(complex(3.5, 4.5)), *result)
		assert.IsType(t, (*complex128)(nil), result)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test complex64 type
		complex64Schema := Complex64()
		resultComplex64 := complex64Schema.MustParse(complex64(complex(1.5, 2.5)))
		assert.IsType(t, complex64(0), resultComplex64)
		assert.Equal(t, complex64(complex(1.5, 2.5)), resultComplex64)

		// Test complex128 type
		complex128Schema := Complex128()
		resultComplex128 := complex128Schema.MustParse(complex128(complex(3.5, 4.5)))
		assert.IsType(t, complex128(0), resultComplex128)
		assert.Equal(t, complex128(complex(3.5, 4.5)), resultComplex128)

		// Test *complex128 type
		ptrSchema := Complex128Ptr()
		ptrResult := ptrSchema.MustParse(complex128(complex(5.5, 6.5)))
		assert.IsType(t, (*complex128)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, complex128(complex(5.5, 6.5)), *ptrResult)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestComplex_Modifiers(t *testing.T) {
	t.Run("Optional always returns *complex128", func(t *testing.T) {
		// From complex64 to *complex128 via Optional
		complex64Schema := Complex64()
		optionalSchema := complex64Schema.Optional()

		// Type check: ensure it returns *ZodComplex[*complex128]
		var _ = optionalSchema

		// Functionality test
		result, err := optionalSchema.Parse(complex128(complex(1.5, 2.5)))
		require.NoError(t, err)
		assert.IsType(t, (*complex128)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, complex128(complex(1.5, 2.5)), *result)

		// From *complex64 to *complex128 via Optional (type conversion)
		ptrSchema := Complex64Ptr()
		optionalPtrSchema := ptrSchema.Optional()
		_ = optionalPtrSchema
	})

	t.Run("Nilable always returns *complex128", func(t *testing.T) {
		complex64Schema := Complex64()
		nilableSchema := complex64Schema.Nilable()

		_ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		// complex64 maintains complex64
		complex64Schema := Complex64()
		defaultComplex64Schema := complex64Schema.Default(complex(1.0, 2.0))
		_ = defaultComplex64Schema

		// *complex128 maintains *complex128
		ptrSchema := Complex128Ptr()
		defaultPtrSchema := ptrSchema.Default(complex(3.0, 4.0))
		_ = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		// complex64 maintains complex64
		complex64Schema := Complex64()
		prefaultComplex64Schema := complex64Schema.Prefault(complex(1.0, 2.0))
		_ = prefaultComplex64Schema

		// *complex128 maintains *complex128
		ptrSchema := Complex128Ptr()
		prefaultPtrSchema := ptrSchema.Prefault(complex(3.0, 4.0))
		_ = prefaultPtrSchema
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestComplex_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		schema := Complex64(). // *ZodComplex[complex64]
					Default(complex(1.0, 2.0)). // *ZodComplex[complex64] (maintains type)
					Optional()                  // *ZodComplex[*complex128] (type conversion)

		_ = schema

		// Test final behavior
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.IsType(t, (*complex128)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, complex128(complex(3.0, 4.0)), *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := Complex64Ptr(). // *ZodComplex[*complex64]
						Nilable().                 // *ZodComplex[*complex128] (type conversion)
						Default(complex(1.0, 2.0)) // *ZodComplex[*complex128] (maintains type)

		_ = schema

		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex128(complex(3.0, 4.0)), *result)
	})

	t.Run("validation chaining", func(t *testing.T) {
		schema := Complex128().
			Min(2.0).
			Max(10.0).
			Positive()

		result, err := schema.Parse(complex(3.0, 4.0)) // magnitude = 5.0
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Should fail validation (magnitude too small)
		_, err = schema.Parse(complex(0.5, 0.5)) // magnitude ≈ 0.71
		assert.Error(t, err)

		// Should fail validation (magnitude too large)
		_, err = schema.Parse(complex(8.0, 8.0)) // magnitude ≈ 11.31
		assert.Error(t, err)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Complex128().
			Default(complex(1.0, 2.0)).
			Prefault(complex(3.0, 4.0))

		result, err := schema.Parse(complex(5.0, 6.0))
		require.NoError(t, err)
		assert.Equal(t, complex(5.0, 6.0), result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestComplex_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		// When both Default and Prefault are set, Default should take precedence for nil input
		schema := Complex128().Default(complex(1.0, 2.0)).Prefault(complex(3.0, 4.0))

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, complex(1.0, 2.0), result) // Should be default, not prefault
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Default value should bypass complex validation constraints
		schema := Complex128().Refine(func(c complex128) bool {
			return false // Always fail refinement
		}, "Should never pass").Default(complex(1.0, 2.0))

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, complex(1.0, 2.0), result) // Default bypasses validation
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value must pass all complex validation including refinements
		schema := Complex128().Refine(func(c complex128) bool {
			return cmplx.Abs(c) >= 5.0 // Magnitude must be at least 5
		}, "Magnitude must be at least 5").Prefault(complex(3.0, 4.0)) // |3+4i| = 5

		// Nil input triggers prefault, goes through validation and succeeds
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Non-nil input that fails validation should not trigger prefault
		_, err = schema.Parse(complex(1.0, 1.0)) // |1+1i| < 5
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Magnitude must be at least 5")
	})

	t.Run("Prefault only triggers for nil input", func(t *testing.T) {
		// Non-nil input that fails validation should not trigger Prefault
		schema := Complex128().Prefault(complex(3.0, 4.0))

		// Invalid complex should fail without triggering Prefault
		_, err := schema.Parse("invalid-complex")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected complex, received string")

		// Valid complex should pass normally
		result, err := schema.Parse(complex(5.0, 6.0))
		require.NoError(t, err)
		assert.Equal(t, complex(5.0, 6.0), result)
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		// Test function call behavior and priority
		defaultCalled := false
		prefaultCalled := false

		schema := Complex128().
			DefaultFunc(func() complex128 {
				defaultCalled = true
				return complex(1.0, 2.0)
			}).
			PrefaultFunc(func() complex128 {
				prefaultCalled = true
				return complex(3.0, 4.0)
			})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, complex(1.0, 2.0), result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		// When Prefault value fails validation, should return error
		schema := Complex128().Refine(func(c complex128) bool {
			return false // Always fail
		}, "Custom validation failed").Prefault(complex(1.0, 2.0))

		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Custom validation failed")
	})

	t.Run("Complex128Ptr with Default and Prefault", func(t *testing.T) {
		// Test pointer types with default and prefault
		schema := Complex128Ptr().Default(complex(1.0, 2.0))

		// Nil input should use default
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex(1.0, 2.0), *result)

		// Valid input should override default
		result, err = schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex(3.0, 4.0), *result)
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestComplex_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept complex numbers with positive real part
		schema := Complex128().Refine(func(c complex128) bool {
			return real(c) > 0
		})

		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		_, err = schema.Parse(complex(-1.0, 2.0))
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Real part must be positive"
		schema := Complex128Ptr().Refine(func(c *complex128) bool {
			return c != nil && real(*c) > 0
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex(3.0, 4.0), *result)

		_, err = schema.Parse(complex(-1.0, 2.0))
		assert.Error(t, err)
	})

	t.Run("refine pointer allows nil", func(t *testing.T) {
		schema := Complex128Ptr().Nilable().Refine(func(c *complex128) bool {
			// Accept nil or complex numbers with positive real part
			return c == nil || real(*c) > 0
		})

		// Expect nil to be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid complex should pass
		result, err = schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, complex(3.0, 4.0), *result)

		// Invalid complex should fail
		_, err = schema.Parse(complex(-1.0, 2.0))
		assert.Error(t, err)
	})

	t.Run("refine complex64 type", func(t *testing.T) {
		schema := Complex64().Refine(func(c complex64) bool {
			return real(c) > 0
		})

		result, err := schema.Parse(complex64(complex(3.0, 4.0)))
		require.NoError(t, err)
		assert.Equal(t, complex64(complex(3.0, 4.0)), result)

		_, err = schema.Parse(complex64(complex(-1.0, 2.0)))
		assert.Error(t, err)
	})
}

func TestComplex_RefineAny(t *testing.T) {
	t.Run("refineAny complex128 schema", func(t *testing.T) {
		// Only accept complex numbers with positive real part via RefineAny
		schema := Complex128().RefineAny(func(v any) bool {
			c, ok := v.(complex128)
			return ok && real(c) > 0
		})

		// Valid value passes
		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Invalid value fails
		_, err = schema.Parse(complex(-1.0, 2.0))
		assert.Error(t, err)
	})

	t.Run("refineAny pointer schema", func(t *testing.T) {
		// Complex128Ptr().RefineAny sees underlying complex128 value
		schema := Complex128Ptr().RefineAny(func(v any) bool {
			c, ok := v.(complex128)
			return ok && real(c) > 0
		})

		result, err := schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, complex(3.0, 4.0), *result)

		_, err = schema.Parse(complex(-1.0, 2.0))
		assert.Error(t, err)
	})

	t.Run("refineAny nilable schema", func(t *testing.T) {
		// Nil input should bypass checks and be accepted because schema is Nilable()
		schema := Complex128Ptr().Nilable().RefineAny(func(v any) bool {
			// Never called for nil input, but return true for completeness
			return true
		})

		// nil passes
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid value still passes
		result, err = schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, complex(3.0, 4.0), *result)
	})
}

// =============================================================================
// Coercion tests
// =============================================================================

func TestComplex_Coercion(t *testing.T) {
	t.Run("string coercion", func(t *testing.T) {
		schema := CoercedComplex128()

		// Test string "3+4i" -> complex128(3+4i)
		result, err := schema.Parse("3+4i")
		require.NoError(t, err, "Should coerce string '3+4i' to complex128")
		assert.Equal(t, complex(3.0, 4.0), result)

		// Test string "-1-2i" -> complex128(-1-2i)
		result, err = schema.Parse("-1-2i")
		require.NoError(t, err, "Should coerce string '-1-2i' to complex128")
		assert.Equal(t, complex(-1.0, -2.0), result)
	})

	t.Run("numeric coercion", func(t *testing.T) {
		schema := CoercedComplex128()

		testCases := []struct {
			input    any
			expected complex128
			name     string
		}{
			{42, complex(42.0, 0.0), "int 42 to complex(42, 0)"},
			{3.14, complex(3.14, 0.0), "float64 3.14 to complex(3.14, 0)"},
			{float32(2.5), complex(2.5, 0.0), "float32 2.5 to complex(2.5, 0)"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("complex type coercion", func(t *testing.T) {
		schema := CoercedComplex128()

		testCases := []struct {
			input    any
			expected complex128
			name     string
		}{
			{complex64(complex(1.5, 2.5)), complex(1.5, 2.5), "complex64 to complex128"},
			{complex128(complex(3.5, 4.5)), complex(3.5, 4.5), "complex128 to complex128"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("coerced complex64 schema", func(t *testing.T) {
		schema := CoercedComplex64()

		// Test string coercion to complex64
		result, err := schema.Parse("3+4i")
		require.NoError(t, err)
		assert.IsType(t, complex64(0), result)
		assert.Equal(t, complex64(complex(3.0, 4.0)), result)

		// Test int coercion to complex64
		result, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, complex64(complex(42.0, 0.0)), result)
	})

	t.Run("invalid coercion inputs", func(t *testing.T) {
		schema := CoercedComplex128()

		// Inputs that cannot be coerced
		invalidInputs := []any{
			"not a complex", "invalid", []complex128{1 + 2i}, map[string]complex128{"key": 1 + 2i}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedComplex128().Min(5.0).Max(10.0)

		// Coercion then validation passes
		result, err := schema.Parse("3+4i") // magnitude = 5.0
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		// Coercion then validation fails
		_, err = schema.Parse("1+1i") // magnitude ≈ 1.41
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling and edge case tests
// =============================================================================

func TestComplex_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := Complex128()

		_, err := schema.Parse("not a complex")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a complex value"
		schema := Complex128Ptr(core.SchemaParams{Error: customError})

		_, err := schema.Parse("not a complex")
		assert.Error(t, err)
	})

	t.Run("validation error messages", func(t *testing.T) {
		schema := Complex128().Min(10.0)

		_, err := schema.Parse(complex(3.0, 4.0)) // magnitude = 5.0 < 10.0
		assert.Error(t, err)
	})

	t.Run("nil handling with *complex128", func(t *testing.T) {
		schema := Complex128Ptr().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid complex
		result, err = schema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex(3.0, 4.0), *result)
	})

	t.Run("special complex values", func(t *testing.T) {
		schema := Complex128()

		// Test complex with infinite real part
		result, err := schema.Parse(complex(math.Inf(1), 4.0))
		require.NoError(t, err)
		assert.True(t, math.IsInf(real(result), 1))
		assert.Equal(t, 4.0, imag(result))

		// Test complex with infinite imaginary part
		result, err = schema.Parse(complex(3.0, math.Inf(-1)))
		require.NoError(t, err)
		assert.Equal(t, 3.0, real(result))
		assert.True(t, math.IsInf(imag(result), -1))

		// Test complex with NaN real part
		result, err = schema.Parse(complex(math.NaN(), 4.0))
		require.NoError(t, err)
		assert.True(t, math.IsNaN(real(result)))
		assert.Equal(t, 4.0, imag(result))

		// Test complex with NaN imaginary part
		result, err = schema.Parse(complex(3.0, math.NaN()))
		require.NoError(t, err)
		assert.Equal(t, 3.0, real(result))
		assert.True(t, math.IsNaN(imag(result)))
	})

	t.Run("very large and small values", func(t *testing.T) {
		schema := Complex128()

		// Test very large magnitude
		result, err := schema.Parse(complex(1e100, 1e100))
		require.NoError(t, err)
		assert.Equal(t, complex(1e100, 1e100), result)

		// Test very small magnitude
		result, err = schema.Parse(complex(1e-100, 1e-100))
		require.NoError(t, err)
		assert.Equal(t, complex(1e-100, 1e-100), result)
	})

	t.Run("precision handling", func(t *testing.T) {
		schema := Complex128()

		// Test precision preservation
		value := complex(0.1+0.2, 0.3+0.4)
		result, err := schema.Parse(value)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("magnitude calculations", func(t *testing.T) {
		testCases := []struct {
			input     complex128
			magnitude float64
			name      string
		}{
			{complex(3.0, 4.0), 5.0, "3+4i has magnitude 5"},
			{complex(0.0, 1.0), 1.0, "pure imaginary i has magnitude 1"},
			{complex(1.0, 0.0), 1.0, "pure real 1 has magnitude 1"},
			{complex(0.0, 0.0), 0.0, "zero has magnitude 0"},
			{complex(-3.0, -4.0), 5.0, "-3-4i has magnitude 5"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				calculated := cmplx.Abs(tc.input)
				assert.InDelta(t, tc.magnitude, calculated, 1e-10)
			})
		}
	})

	t.Run("Complex64 vs Complex128 distinction", func(t *testing.T) {
		complex64Schema := Complex64()
		complex128Schema := Complex128()

		// Both should handle their respective types
		resultComplex64, err := complex64Schema.Parse(complex64(complex(1.5, 2.5)))
		require.NoError(t, err)
		assert.IsType(t, complex64(0), resultComplex64)

		resultComplex128, err := complex128Schema.Parse(complex128(complex(3.5, 4.5)))
		require.NoError(t, err)
		assert.IsType(t, complex128(0), resultComplex128)
	})

	t.Run("Complex alias", func(t *testing.T) {
		complexSchema := Complex()

		// Complex should behave like Complex128
		result, err := complexSchema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.IsType(t, complex128(0), result)
		assert.Equal(t, complex(3.0, 4.0), result)
	})

	t.Run("finite vs infinite validation", func(t *testing.T) {
		finiteSchema := Complex128().Finite()
		regularSchema := Complex128()

		// Regular schema should accept infinity
		result, err := regularSchema.Parse(complex(math.Inf(1), 4.0))
		require.NoError(t, err)
		assert.True(t, math.IsInf(real(result), 1))

		// Finite schema should reject infinity
		_, err = finiteSchema.Parse(complex(math.Inf(1), 4.0))
		assert.Error(t, err)

		// Both should accept finite values
		result, err = finiteSchema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)

		result, err = regularSchema.Parse(complex(3.0, 4.0))
		require.NoError(t, err)
		assert.Equal(t, complex(3.0, 4.0), result)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestComplex_Overwrite(t *testing.T) {
	t.Run("basic complex transformation", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			// Transform complex number by adding 1+1i
			return c + (1 + 1i)
		})

		input := 2 + 3i
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := complex128(3 + 4i)
		assert.Equal(t, expected, result)
	})

	t.Run("complex conjugate transformation", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			// Return complex conjugate
			return complex(real(c), -imag(c))
		})

		input := 3 + 4i
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := complex128(3 - 4i)
		assert.Equal(t, expected, result)
	})

	t.Run("magnitude normalization", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			// Normalize to unit magnitude
			magnitude := real(c)*real(c) + imag(c)*imag(c)
			if magnitude == 0 {
				return c
			}
			sqrtMag := math.Sqrt(magnitude)
			return complex(real(c)/sqrtMag, imag(c)/sqrtMag)
		})

		input := 3 + 4i // magnitude = 5
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := complex128(0.6 + 0.8i) // 3/5 + 4i/5
		assert.InDelta(t, real(expected), real(result), 1e-10)
		assert.InDelta(t, imag(expected), imag(result), 1e-10)
	})

	t.Run("complex64 type handling", func(t *testing.T) {
		schema := Complex64().Overwrite(func(c complex64) complex64 {
			// Double both real and imaginary parts
			return complex64(2 * c)
		})

		input := complex64(1 + 2i)
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := complex64(2 + 4i)
		assert.Equal(t, expected, result)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := ComplexPtr().Overwrite(func(c *complex128) *complex128 {
			if c == nil {
				// Return default complex number
				return new(complex128(1 + 0i))
			}
			// Add π to the real part
			return new(complex(real(*c)+math.Pi, imag(*c)))
		})

		input := complex128(1 + 2i)
		result, err := schema.Parse(input)
		require.NoError(t, err)
		require.NotNil(t, result)

		expected := complex128(1+math.Pi) + 2i
		assert.InDelta(t, real(expected), real(*result), 1e-10)
		assert.InDelta(t, imag(expected), imag(*result), 1e-10)
	})

	t.Run("chaining with validations", func(t *testing.T) {
		schema := Complex().
			Overwrite(func(c complex128) complex128 {
				// Ensure imaginary part is positive
				if imag(c) < 0 {
					return complex(real(c), -imag(c))
				}
				return c
			}).
			Refine(func(c complex128) bool {
				// Validate that imaginary part is indeed positive after transformation
				return imag(c) >= 0
			}, "Imaginary part must be non-negative")

		// Test with negative imaginary part
		input := 3 - 4i
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := complex128(3 + 4i) // Imaginary part should be flipped to positive
		assert.Equal(t, expected, result)
	})

	t.Run("polar to rectangular conversion", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			// Treat input as polar coordinates (r, θ) and convert to rectangular
			r := real(c)
			theta := imag(c)

			// Convert polar to rectangular
			return complex(r*math.Cos(theta), r*math.Sin(theta))
		})

		// Input: r=2, θ=π/4 (45 degrees)
		input := complex(2, math.Pi/4)
		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Expected: (2*cos(π/4), 2*sin(π/4)) = (√2, √2)
		expectedReal := 2 * math.Cos(math.Pi/4)
		expectedImag := 2 * math.Sin(math.Pi/4)
		assert.InDelta(t, expectedReal, real(result), 1e-10)
		assert.InDelta(t, expectedImag, imag(result), 1e-10)
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			return c // Identity transformation
		})

		input := complex128(5 + 6i)
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.IsType(t, complex128(0), result)
		assert.Equal(t, input, result)
	})

	t.Run("multiple transformations", func(t *testing.T) {
		schema := Complex().
			Overwrite(func(c complex128) complex128 {
				// First transformation: add 1 to real part
				return complex(real(c)+1, imag(c))
			}).
			Overwrite(func(c complex128) complex128 {
				// Second transformation: multiply imaginary part by 2
				return complex(real(c), imag(c)*2)
			})

		input := complex128(2 + 3i)
		result, err := schema.Parse(input)
		require.NoError(t, err)

		expected := complex128(3 + 6i) // real: 2+1=3, imag: 3*2=6
		assert.Equal(t, expected, result)
	})

	t.Run("error handling preservation", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			return c
		})

		// Invalid input should still fail validation
		_, err := schema.Parse("not a complex number")
		assert.Error(t, err)

		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("phase shift transformation", func(t *testing.T) {
		schema := Complex().Overwrite(func(c complex128) complex128 {
			// Apply 90-degree phase shift (multiply by i)
			return c * 1i
		})

		input := complex128(3 + 4i)
		result, err := schema.Parse(input)
		require.NoError(t, err)

		// (3 + 4i) * i = 3i + 4i² = 3i - 4 = -4 + 3i
		expected := complex128(-4 + 3i)
		assert.Equal(t, expected, result)
	})

	t.Run("default value interaction", func(t *testing.T) {
		schema := Complex().
			Default(complex128(1 + 1i)).
			Overwrite(func(c complex128) complex128 {
				// Square the complex number
				return c * c
			})

		// Test with actual input
		result1, err := schema.Parse(complex128(2 + 0i))
		require.NoError(t, err)
		expected1 := complex128(4 + 0i) // (2+0i)² = 4
		assert.Equal(t, expected1, result1)

		// Test nil input uses default -> transformed
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		expected2 := complex128(0 + 2i) // (1+1i)² = 1 + 2i - 1 = 2i
		assert.Equal(t, expected2, result2)
	})
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestComplex_StrictParse(t *testing.T) {
	t.Run("basic functionality complex64", func(t *testing.T) {
		schema := Complex64()

		// Test StrictParse with exact type match
		result, err := schema.StrictParse(complex64(3 + 4i))
		require.NoError(t, err)
		assert.Equal(t, complex64(3+4i), result)
		assert.IsType(t, complex64(0), result)

		// Test StrictParse with negative values
		negResult, err := schema.StrictParse(complex64(-2 - 1i))
		require.NoError(t, err)
		assert.Equal(t, complex64(-2-1i), negResult)
	})

	t.Run("basic functionality complex128", func(t *testing.T) {
		schema := Complex128()

		// Test StrictParse with exact type match
		result, err := schema.StrictParse(complex128(5 + 6i))
		require.NoError(t, err)
		assert.Equal(t, complex128(5+6i), result)
		assert.IsType(t, complex128(0), result)

		// Test StrictParse with zero
		zeroResult, err := schema.StrictParse(complex128(0 + 0i))
		require.NoError(t, err)
		assert.Equal(t, complex128(0+0i), zeroResult)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := Complex128().Refine(func(c complex128) bool {
			return cmplx.Abs(c) >= 5.0 // Magnitude must be at least 5
		}, "Magnitude must be at least 5")

		// Valid case
		result, err := schema.StrictParse(complex128(3 + 4i)) // |3+4i| = 5
		require.NoError(t, err)
		assert.Equal(t, complex128(3+4i), result)

		// Invalid case - magnitude too small
		_, err = schema.StrictParse(complex128(1 + 1i)) // |1+1i| ≈ 1.414
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Magnitude must be at least 5")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := Complex128Ptr()
		complexVal := complex128(7 + 8i)

		// Test with valid pointer input
		result, err := schema.StrictParse(&complexVal)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex128(7+8i), *result)
		assert.IsType(t, (*complex128)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := Complex64Ptr().Default(complex(1, 2))
		var nilPtr *complex64 = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex64(1+2i), *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		schema := Complex128Ptr().Refine(func(c *complex128) bool {
			return c != nil && cmplx.Abs(*c) >= 10.0
		}, "Magnitude must be at least 10").Prefault(complex128(10 + 0i))
		smallVal := complex128(2 + 2i) // Magnitude too small

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema.StrictParse(&smallVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Magnitude must be at least 10")

		// Test with nil input (should use prefault and pass validation)
		result, err := schema.StrictParse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex128(10+0i), *result)
	})

	t.Run("special complex values", func(t *testing.T) {
		schema := Complex128()

		// Test with infinity
		infResult, err := schema.StrictParse(complex(math.Inf(1), 0))
		require.NoError(t, err)
		assert.True(t, math.IsInf(real(infResult), 1))

		// Test with NaN
		nanResult, err := schema.StrictParse(complex(math.NaN(), 0))
		require.NoError(t, err)
		assert.True(t, math.IsNaN(real(nanResult)))

		// Test purely real number
		realResult, err := schema.StrictParse(complex128(42 + 0i))
		require.NoError(t, err)
		assert.Equal(t, complex128(42+0i), realResult)

		// Test purely imaginary number
		imagResult, err := schema.StrictParse(complex128(0 + 42i))
		require.NoError(t, err)
		assert.Equal(t, complex128(0+42i), imagResult)
	})
}

func TestComplex_MustStrictParse(t *testing.T) {
	t.Run("basic functionality complex64", func(t *testing.T) {
		schema := Complex64()

		// Test MustStrictParse with valid input
		result := schema.MustStrictParse(complex64(9 + 10i))
		assert.Equal(t, complex64(9+10i), result)
		assert.IsType(t, complex64(0), result)

		// Test MustStrictParse with zero
		zeroResult := schema.MustStrictParse(complex64(0 + 0i))
		assert.Equal(t, complex64(0+0i), zeroResult)
	})

	t.Run("basic functionality complex128", func(t *testing.T) {
		schema := Complex128()

		// Test MustStrictParse with valid input
		result := schema.MustStrictParse(complex128(11 + 12i))
		assert.Equal(t, complex128(11+12i), result)
		assert.IsType(t, complex128(0), result)

		// Test MustStrictParse with negative values
		negResult := schema.MustStrictParse(complex128(-5 - 6i))
		assert.Equal(t, complex128(-5-6i), negResult)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := Complex128().Refine(func(c complex128) bool {
			return cmplx.Abs(c) >= 10.0 // Magnitude must be at least 10
		}, "Magnitude must be at least 10")

		// Test panic with validation failure
		assert.Panics(t, func() {
			schema.MustStrictParse(complex128(1 + 1i)) // Too small, should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := Complex64Ptr()
		complexVal := complex64(13 + 14i)

		// Test with valid pointer input
		result := schema.MustStrictParse(&complexVal)
		require.NotNil(t, result)
		assert.Equal(t, complex64(13+14i), *result)
		assert.IsType(t, (*complex64)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := Complex128Ptr().Default(complex128(15 + 16i))
		var nilPtr *complex128 = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, complex128(15+16i), *result)
	})

	t.Run("special complex values", func(t *testing.T) {
		schema := Complex128()

		// Test with infinity
		infResult := schema.MustStrictParse(complex(math.Inf(1), 0))
		assert.True(t, math.IsInf(real(infResult), 1))

		// Test with NaN
		nanResult := schema.MustStrictParse(complex(math.NaN(), 0))
		assert.True(t, math.IsNaN(real(nanResult)))

		// Test purely real number
		realResult := schema.MustStrictParse(complex128(100 + 0i))
		assert.Equal(t, complex128(100+0i), realResult)

		// Test purely imaginary number
		imagResult := schema.MustStrictParse(complex128(0 + 100i))
		assert.Equal(t, complex128(0+100i), imagResult)
	})
}
