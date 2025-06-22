package types

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestBigIntBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := BigInt()
		// Valid bigint
		bigVal := big.NewInt(42)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)
		// Invalid type
		_, err = schema.Parse("not a bigint")
		assert.Error(t, err)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := BigInt()
		// *big.Int input returns *big.Int
		bigVal := big.NewInt(123)
		result1, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.IsType(t, (*big.Int)(nil), result1)
		assert.Equal(t, bigVal, result1)
		// Pointer input returns same pointer
		result2, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result2)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1))
		bigVal := big.NewInt(42)

		result, err := schema.Parse(bigVal)
		require.NoError(t, err)

		// Verify exact pointer identity is preserved
		resultPtr, ok := result.(*big.Int)
		require.True(t, ok, "Result should be *big.Int")
		assert.True(t, resultPtr == bigVal, "Should return the exact same pointer")
		assert.Equal(t, int64(42), resultPtr.Int64())
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := BigInt().Nilable()
		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*big.Int)(nil), result)
		// Valid input keeps type inference
		bigVal := big.NewInt(42)
		result2, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result2)
		assert.IsType(t, (*big.Int)(nil), result2)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := BigInt().Min(big.NewInt(1))
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		bigVal := big.NewInt(5)
		result2, err2 := nilableSchema.Parse(bigVal)
		require.NoError(t, err2)
		assert.Equal(t, bigVal, result2)

		// Test nilable schema rejects invalid values
		smallVal := big.NewInt(0)
		_, err3 := nilableSchema.Parse(smallVal)
		assert.Error(t, err3)

		// 🔥 Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse(bigVal)
		require.NoError(t, err5)
		assert.Equal(t, bigVal, result5)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestBigIntCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := CoercedBigInt()
		tests := []struct {
			input    any
			expected string // Use string representation for comparison
		}{
			{"42", "42"},
			{int(42), "42"},
			{int64(84), "84"},
			{uint(100), "100"},
			{float64(123.0), "123"}, // Should truncate
		}
		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			resultBig, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, resultBig.String())
		}
	})

	t.Run("string coercion with large numbers", func(t *testing.T) {
		schema := CoercedBigInt()
		largeNumber := "123456789012345678901234567890"
		result, err := schema.Parse(largeNumber)
		require.NoError(t, err)
		resultBig, ok := result.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, largeNumber, resultBig.String())
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedBigInt().Min(big.NewInt(5))
		// Coercion then validation passes
		result, err := schema.Parse("50")
		require.NoError(t, err)
		resultBig, ok := result.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, "50", resultBig.String())
		// Coercion then validation fails
		_, err = schema.Parse("3")
		assert.Error(t, err)
	})

	t.Run("failed coercion", func(t *testing.T) {
		schema := CoercedBigInt()
		invalidInputs := []any{
			"not a number",
			[]int{1},                 // slice
			map[string]int{"key": 1}, // map
			float64(3.14),            // non-integer float
		}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Should fail to coerce %v", input)
		}
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestBigIntValidations(t *testing.T) {
	t.Run("range validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  core.ZodType[any, any]
			input   *big.Int
			wantErr bool
		}{
			{"min valid", BigInt().Min(big.NewInt(5)), big.NewInt(10), false},
			{"min invalid", BigInt().Min(big.NewInt(5)), big.NewInt(3), true},
			{"max valid", BigInt().Max(big.NewInt(100)), big.NewInt(50), false},
			{"max invalid", BigInt().Max(big.NewInt(100)), big.NewInt(150), true},
			{"gt valid", BigInt().Gt(big.NewInt(0)), big.NewInt(5), false},
			{"gt invalid", BigInt().Gt(big.NewInt(0)), big.NewInt(0), true},
			{"gte valid", BigInt().Gte(big.NewInt(0)), big.NewInt(0), false},
			{"gte invalid", BigInt().Gte(big.NewInt(0)), big.NewInt(-1), true},
			{"lt valid", BigInt().Lt(big.NewInt(100)), big.NewInt(50), false},
			{"lt invalid", BigInt().Lt(big.NewInt(100)), big.NewInt(100), true},
			{"lte valid", BigInt().Lte(big.NewInt(100)), big.NewInt(100), false},
			{"lte invalid", BigInt().Lte(big.NewInt(100)), big.NewInt(101), true},
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
			input   *big.Int
			wantErr bool
		}{
			{"positive valid", BigInt().Positive(), big.NewInt(5), false},
			{"positive invalid", BigInt().Positive(), big.NewInt(-5), true},
			{"negative valid", BigInt().Negative(), big.NewInt(-5), false},
			{"negative invalid", BigInt().Negative(), big.NewInt(5), true},
			{"non-negative valid", BigInt().NonNegative(), big.NewInt(0), false},
			{"non-negative invalid", BigInt().NonNegative(), big.NewInt(-1), true},
			{"non-positive valid", BigInt().NonPositive(), big.NewInt(0), false},
			{"non-positive invalid", BigInt().NonPositive(), big.NewInt(1), true},
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
		schema := BigInt().MultipleOf(big.NewInt(5))
		// Valid multiple
		result, err := schema.Parse(big.NewInt(25))
		require.NoError(t, err)
		assert.Equal(t, "25", result.(*big.Int).String())
		// Invalid multiple
		_, err = schema.Parse(big.NewInt(23))
		assert.Error(t, err)
	})

	t.Run("large number validations", func(t *testing.T) {
		// Test with very large numbers
		largeMin := new(big.Int)
		largeMin.SetString("123456789012345678901234567890", 10)
		schema := BigInt().Min(largeMin)

		// Valid large number
		largeValid := new(big.Int)
		largeValid.SetString("123456789012345678901234567891", 10)
		result, err := schema.Parse(largeValid)
		require.NoError(t, err)
		assert.Equal(t, largeValid, result)

		// Invalid small number
		_, err = schema.Parse(big.NewInt(42))
		assert.Error(t, err)
	})

	// Additional tests to ensure that consecutive Min/Max calls apply the strictest
	// bounds. GoZod accumulates checks instead of replacing them, but the effective
	// behaviour should match expectations: the tightest constraints decide whether
	// validation passes or fails.

	// -----------------------------------------------------------------------------
	// 3.1 Min/Max override semantics
	// -----------------------------------------------------------------------------
	t.Run("min/max override semantics", func(t *testing.T) {
		// Min override: second Min(10) should make 10 the effective lower bound.
		minSchema := BigInt().Min(big.NewInt(5)).Min(big.NewInt(10))
		_, err := minSchema.Parse(big.NewInt(10)) // boundary should succeed
		require.NoError(t, err)
		_, err = minSchema.Parse(big.NewInt(6)) // below 10 should fail
		assert.Error(t, err)

		// Max override: second Max(1) should make 1 the effective upper bound.
		maxSchema := BigInt().Max(big.NewInt(5)).Max(big.NewInt(1))
		_, err = maxSchema.Parse(big.NewInt(1)) // boundary should succeed
		require.NoError(t, err)
		_, err = maxSchema.Parse(big.NewInt(4)) // above 1 should fail
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestBigIntModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schema := BigInt().Optional()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		bigVal := big.NewInt(42)
		result2, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := BigInt().Nilable()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		bigVal := big.NewInt(42)
		result2, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result2)
	})

	t.Run("nullish modifier", func(t *testing.T) {
		schema := BigInt().Nullish()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		bigVal := big.NewInt(42)
		result2, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result2)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := BigInt()
		// Valid input should not panic
		bigVal := big.NewInt(42)
		result := schema.MustParse(bigVal)
		assert.Equal(t, bigVal, result)
		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestBigIntChaining(t *testing.T) {
	t.Run("multiple validations", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1)).Max(big.NewInt(100)).Positive()
		// Valid input
		bigVal := big.NewInt(50)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)
		// Validation failures
		testCases := []*big.Int{
			big.NewInt(0),   // too small
			big.NewInt(150), // too large
			big.NewInt(-10), // not positive
		}
		for _, input := range testCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("validation with multiple of", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(5)).Max(big.NewInt(100)).MultipleOf(big.NewInt(5))
		// Valid input
		result, err := schema.Parse(big.NewInt(25))
		require.NoError(t, err)
		assert.Equal(t, "25", result.(*big.Int).String())
		// Validation failures
		_, err = schema.Parse(big.NewInt(23)) // not multiple of 5
		assert.Error(t, err)
	})

	t.Run("large number chaining", func(t *testing.T) {
		largeMin := new(big.Int)
		largeMin.SetString("123456789012345678901234567890", 10)
		largeMax := new(big.Int)
		largeMax.SetString("999999999999999999999999999999", 10)

		schema := BigInt().Min(largeMin).Max(largeMax).Positive()

		// Valid large number
		largeValid := new(big.Int)
		largeValid.SetString("500000000000000000000000000000", 10)
		result, err := schema.Parse(largeValid)
		require.NoError(t, err)
		assert.Equal(t, largeValid, result)

		// Invalid large number (too large)
		tooLarge := new(big.Int)
		tooLarge.SetString("9999999999999999999999999999999", 10)
		_, err = schema.Parse(tooLarge)
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestBigIntTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := BigInt().Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			result := new(big.Int)
			result.Mul(val, big.NewInt(2))
			return result, nil
		})
		result, err := schema.Parse(big.NewInt(21))
		require.NoError(t, err)
		resultBig, ok := result.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, "42", resultBig.String())
	})

	t.Run("transform chaining", func(t *testing.T) {
		schema := BigInt().
			Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
				result := new(big.Int)
				result.Mul(val, big.NewInt(2))
				return result, nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if bigVal, ok := val.(*big.Int); ok {
					return fmt.Sprintf("result_%s", bigVal.String()), nil
				}
				return val, nil
			})
		result, err := schema.Parse(big.NewInt(21))
		require.NoError(t, err)
		assert.Equal(t, "result_42", result)
	})

	t.Run("pipe combination", func(t *testing.T) {
		schema := BigInt().
			Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
				return val.String(), nil
			}).
			Pipe(String().Min(2))
		result, err := schema.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})

	t.Run("transform with validation", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1)).Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			if val.Cmp(big.NewInt(0)) < 0 {
				return nil, fmt.Errorf("negative numbers not allowed")
			}
			// Convert to scientific notation
			return fmt.Sprintf("%e", float64(val.Int64())), nil
		})

		result, err := schema.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, "4.200000e+01", result)

		// Validation before transform should fail
		_, err = schema.Parse(big.NewInt(0))
		assert.Error(t, err)
	})

	t.Run("mathematical transforms", func(t *testing.T) {
		schema := BigInt().Positive().Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			squared := new(big.Int)
			squared.Mul(val, val)
			return map[string]any{
				"original": val.String(),
				"squared":  squared.String(),
				"double":   new(big.Int).Mul(val, big.NewInt(2)).String(),
			}, nil
		})

		result, err := schema.Parse(big.NewInt(10))
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "10", resultMap["original"])
		assert.Equal(t, "100", resultMap["squared"])
		assert.Equal(t, "20", resultMap["double"])
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestBigIntRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := BigInt().Refine(func(val *big.Int) bool {
			// Check if number is even
			return val.Bit(0) == 0
		}, core.SchemaParams{
			Error: "Number must be even",
		})
		result, err := schema.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, "42", result.(*big.Int).String())
		_, err = schema.Parse(big.NewInt(43))
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "even")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := big.NewInt(21)

		// Refine: only validates, never modifies
		refineSchema := BigInt().Refine(func(val *big.Int) bool {
			return val.Cmp(big.NewInt(0)) > 0
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := BigInt().Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			result := new(big.Int)
			result.Mul(val, big.NewInt(2))
			return result, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, input, refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		transformResultBig, ok := transformResult.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, "42", transformResultBig.String())

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input.String(), transformResultBig.String(), "Transform should return modified value")
	})

	t.Run("refine preserves pointer identity", func(t *testing.T) {
		schema := BigInt().Refine(func(val *big.Int) bool {
			return val.Cmp(big.NewInt(1)) >= 0
		})

		input := big.NewInt(42)

		result, err := schema.Parse(input)
		require.NoError(t, err)

		// Verify exact pointer identity is preserved
		resultPtr, ok := result.(*big.Int)
		require.True(t, ok)
		assert.True(t, resultPtr == input, "Refine should preserve exact pointer identity")
	})

	t.Run("refine with prime validation", func(t *testing.T) {
		schema := BigInt().Refine(func(val *big.Int) bool {
			// Simple primality test for small numbers
			if val.Cmp(big.NewInt(2)) < 0 {
				return false
			}
			if val.Cmp(big.NewInt(2)) == 0 {
				return true
			}
			if val.Bit(0) == 0 { // even number
				return false
			}
			// Check odd divisors up to sqrt(n)
			sqrt := new(big.Int).Sqrt(val)
			for i := big.NewInt(3); i.Cmp(sqrt) <= 0; i.Add(i, big.NewInt(2)) {
				if new(big.Int).Mod(val, i).Cmp(big.NewInt(0)) == 0 {
					return false
				}
			}
			return true
		}, core.SchemaParams{
			Error: func(issue core.ZodRawIssue) string {
				if input, ok := issue.Input.(*big.Int); ok {
					return fmt.Sprintf("The number %s is not prime", input.String())
				}
				return "Invalid input for prime validation"
			},
		})

		result, err := schema.Parse(big.NewInt(17)) // Prime number
		require.NoError(t, err)
		assert.Equal(t, "17", result.(*big.Int).String())

		_, err = schema.Parse(big.NewInt(15)) // Not prime
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "not prime")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestBigIntErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(10))
		_, err := schema.Parse(big.NewInt(5))
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(10), core.SchemaParams{
			Error: "core.Custom minimum value error",
		})
		_, err := schema.Parse(big.NewInt(5))
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "core.Custom minimum value error")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(10)).Max(big.NewInt(20))
		_, err := schema.Parse(big.NewInt(5))
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})

	t.Run("type mismatch error", func(t *testing.T) {
		schema := BigInt()
		_, err := schema.Parse("not a bigint")
		assert.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("coercion error handling", func(t *testing.T) {
		schema := CoercedBigInt()
		invalidInputs := []any{
			"not a number",
			3.14, // non-integer float
			complex(1, 2),
		}
		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input %v", input)
		}
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestBigIntEdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		schema := BigInt()
		result, err := schema.Parse(big.NewInt(0))
		require.NoError(t, err)
		assert.Equal(t, "0", result.(*big.Int).String())
	})

	t.Run("very large numbers", func(t *testing.T) {
		schema := BigInt()
		// Create a very large number
		largeNumber := new(big.Int)
		largeNumber.SetString("123456789012345678901234567890123456789012345678901234567890", 10)

		result, err := schema.Parse(largeNumber)
		require.NoError(t, err)
		assert.Equal(t, largeNumber, result)
	})

	t.Run("negative numbers", func(t *testing.T) {
		schema := BigInt()
		negativeNumber := big.NewInt(-42)
		result, err := schema.Parse(negativeNumber)
		require.NoError(t, err)
		assert.Equal(t, negativeNumber, result)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := BigInt()
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
		schema := BigInt()
		invalidTypes := []any{
			"42",
			42,
			42.0,
			true,
			[]int{42},
			map[string]int{"key": 42},
		}
		for _, invalidType := range invalidTypes {
			_, err := schema.Parse(invalidType)
			assert.Error(t, err, "Expected error for type %T", invalidType)
		}
	})

	t.Run("modifier combinations", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1)).Positive().Nilable()
		// Valid input
		bigVal := big.NewInt(5)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)
		// nil input
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		// Invalid input
		_, err = schema.Parse(big.NewInt(-5))
		assert.Error(t, err)
	})

	t.Run("precision and boundary tests", func(t *testing.T) {
		schema := BigInt()

		// Test maximum int64 value
		maxInt64 := big.NewInt(9223372036854775807) // math.MaxInt64
		result, err := schema.Parse(maxInt64)
		require.NoError(t, err)
		assert.Equal(t, maxInt64, result)

		// Test beyond int64 range
		beyondInt64 := new(big.Int)
		beyondInt64.SetString("9223372036854775808", 10) // MaxInt64 + 1
		result2, err := schema.Parse(beyondInt64)
		require.NoError(t, err)
		assert.Equal(t, beyondInt64, result2)
	})

	t.Run("comparison edge cases", func(t *testing.T) {
		// Test comparison with equal values
		schema := BigInt().Min(big.NewInt(42)).Max(big.NewInt(42))
		result, err := schema.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, "42", result.(*big.Int).String())

		// Test just outside range
		_, err = schema.Parse(big.NewInt(41))
		assert.Error(t, err)
		_, err = schema.Parse(big.NewInt(43))
		assert.Error(t, err)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestBigIntDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1)).Default(big.NewInt(42))
		// nil input uses default value
		result, err := any(schema).(core.ZodType[any, any]).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "42", result.(*big.Int).String())
		// Valid input normal validation
		bigVal := big.NewInt(5)
		result2, err := any(schema).(core.ZodType[any, any]).Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result2)
		// Invalid input still fails
		_, err = any(schema).(core.ZodType[any, any]).Parse(big.NewInt(0))
		assert.Error(t, err)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		baseSchema := BigInt().DefaultFunc(func() *big.Int {
			counter++
			return big.NewInt(int64(counter * 42))
		})
		schema := any(baseSchema).(core.ZodType[any, any])

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "42", result1.(*big.Int).String())

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "84", result2.(*big.Int).String())

		// Valid input bypasses default generation
		result3, err3 := schema.Parse(big.NewInt(100))
		require.NoError(t, err3)
		assert.Equal(t, "100", result3.(*big.Int).String())

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := BigInt().
			Default(big.NewInt(42)).
			Min(big.NewInt(1)).
			Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
				squared := new(big.Int)
				squared.Mul(val, val)
				return map[string]any{
					"original": val.String(),
					"squared":  squared.String(),
					"hex":      fmt.Sprintf("0x%x", val),
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse(big.NewInt(10))
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, "10", result1Map["original"])
		assert.Equal(t, "100", result1Map["squared"])
		assert.Equal(t, "0xa", result1Map["hex"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, "42", result2Map["original"])
		assert.Equal(t, "1764", result2Map["squared"])
		assert.Equal(t, "0x2a", result2Map["hex"])

		// Invalid input still fails validation
		_, err3 := schema.Parse(big.NewInt(0))
		assert.Error(t, err3, "Zero should fail Min(1) validation")
	})

	t.Run("default value", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1)).Default(big.NewInt(999))
		// nil input uses default value
		result1, err := any(schema).(core.ZodType[any, any]).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "999", result1.(*big.Int).String())
		// Valid input normal validation
		result3, err := any(schema).(core.ZodType[any, any]).Parse(big.NewInt(5))
		require.NoError(t, err)
		assert.Equal(t, "5", result3.(*big.Int).String())
		// Invalid input still fails validation
		_, err = any(schema).(core.ZodType[any, any]).Parse(big.NewInt(0))
		assert.Error(t, err)
	})

	t.Run("default chaining", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(1)).Default(big.NewInt(42))
		// nil input uses default value
		result, err := any(schema).(core.ZodType[any, any]).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "42", result.(*big.Int).String())
		// Valid input normal validation
		result3, err := any(schema).(core.ZodType[any, any]).Parse(big.NewInt(5))
		require.NoError(t, err)
		assert.Equal(t, "5", result3.(*big.Int).String())
		// Invalid input still fails validation
		_, err = any(schema).(core.ZodType[any, any]).Parse(big.NewInt(0))
		assert.Error(t, err)
	})

	t.Run("typed default chaining", func(t *testing.T) {
		// Test that Default returns type-safe wrapper that supports chaining
		schema := BigInt().Default(big.NewInt(42)).Min(big.NewInt(1)).Max(big.NewInt(100)).Positive()

		// Test default functionality
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "42", result1.(*big.Int).String())

		// Test chained validations work
		result2, err2 := schema.Parse(big.NewInt(50))
		require.NoError(t, err2)
		assert.Equal(t, "50", result2.(*big.Int).String())

		// Test validation failures
		_, err3 := schema.Parse(big.NewInt(0)) // below Min(1)
		assert.Error(t, err3)
		_, err4 := schema.Parse(big.NewInt(150)) // above Max(100)
		assert.Error(t, err4)
	})

	t.Run("large number defaults", func(t *testing.T) {
		largeDefault := new(big.Int)
		largeDefault.SetString("123456789012345678901234567890", 10)

		schema := BigInt().Default(largeDefault)

		result, err := any(schema).(core.ZodType[any, any]).Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, largeDefault, result)
	})

	t.Run("prefault value", func(t *testing.T) {
		fallbackValue := big.NewInt(999)
		schema := BigInt().Min(big.NewInt(10)).Prefault(fallbackValue)

		// Valid input should pass through
		validInput := big.NewInt(20)
		result, err := any(schema).(core.ZodType[any, any]).Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)

		// Invalid input should use prefault
		invalidInput := big.NewInt(5) // Below min(10)
		result, err = any(schema).(core.ZodType[any, any]).Parse(invalidInput)
		require.NoError(t, err)
		assert.Equal(t, fallbackValue, result)

		// Invalid type should use prefault
		result, err = any(schema).(core.ZodType[any, any]).Parse("not a bigint")
		require.NoError(t, err)
		assert.Equal(t, fallbackValue, result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		schema := BigInt().Min(big.NewInt(10)).PrefaultFunc(func() *big.Int {
			counter++
			return big.NewInt(int64(1000 + counter))
		})

		// Valid input should not call function
		validInput := big.NewInt(20)
		result, err := any(schema).(core.ZodType[any, any]).Parse(validInput)
		require.NoError(t, err)
		assert.Equal(t, validInput, result)
		assert.Equal(t, 0, counter)

		// Invalid input should call function
		invalidInput := big.NewInt(5) // Below min(10)
		result, err = any(schema).(core.ZodType[any, any]).Parse(invalidInput)
		require.NoError(t, err)
		expectedValue := big.NewInt(1001)
		assert.Equal(t, expectedValue, result)
		assert.Equal(t, 1, counter)

		// Another invalid input should increment counter
		result, err = any(schema).(core.ZodType[any, any]).Parse("invalid")
		require.NoError(t, err)
		expectedValue2 := big.NewInt(1002)
		assert.Equal(t, expectedValue2, result)
		assert.Equal(t, 2, counter)
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultValue := big.NewInt(500)
		fallbackValue := big.NewInt(999)

		defaultSchema := BigInt().Default(defaultValue)
		prefaultSchema := BigInt().Min(big.NewInt(10)).Prefault(fallbackValue)

		// For valid input: different behaviors
		validInput := big.NewInt(20)
		result1, err1 := any(defaultSchema).(core.ZodType[any, any]).Parse(validInput)
		require.NoError(t, err1)
		assert.Equal(t, validInput, result1, "Default: valid input passes through")

		result2, err2 := any(prefaultSchema).(core.ZodType[any, any]).Parse(validInput)
		require.NoError(t, err2)
		assert.Equal(t, validInput, result2, "Prefault: valid input passes through")

		// For nil input: different behaviors
		result3, err3 := any(defaultSchema).(core.ZodType[any, any]).Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, defaultValue, result3, "Default: nil gets default value")

		result4, err4 := any(prefaultSchema).(core.ZodType[any, any]).Parse(nil)
		require.NoError(t, err4)
		assert.Equal(t, fallbackValue, result4, "Prefault: nil fails validation, use fallback")

		// For invalid input: different behaviors
		invalidInput := big.NewInt(5) // Below min(10)
		result5, err5 := any(defaultSchema).(core.ZodType[any, any]).Parse(invalidInput)
		require.NoError(t, err5)
		assert.Equal(t, invalidInput, result5, "Default: valid type passes through")

		result6, err6 := any(prefaultSchema).(core.ZodType[any, any]).Parse(invalidInput)
		require.NoError(t, err6)
		assert.Equal(t, fallbackValue, result6, "Prefault: validation fails, use fallback")
	})
}

// =============================================================================
// Additional type-specific tests
// =============================================================================

func TestBigIntTypeSpecific(t *testing.T) {
	t.Run("constructor validation", func(t *testing.T) {
		schema := BigInt()
		bigVal := big.NewInt(42)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)
	})

	t.Run("coerce constructor", func(t *testing.T) {
		schema := CoercedBigInt()
		result, err := schema.Parse("42")
		require.NoError(t, err)
		resultBig, ok := result.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, "42", resultBig.String())
	})

	t.Run("mathematical operations in transforms", func(t *testing.T) {
		schema := BigInt().Positive().Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			// Factorial calculation for small numbers
			if val.Cmp(big.NewInt(10)) > 0 {
				return nil, fmt.Errorf("number too large for factorial")
			}

			factorial := big.NewInt(1)
			for i := big.NewInt(1); i.Cmp(val) <= 0; i.Add(i, big.NewInt(1)) {
				factorial.Mul(factorial, i)
			}

			return factorial, nil
		})

		result, err := schema.Parse(big.NewInt(5))
		require.NoError(t, err)
		resultBig, ok := result.(*big.Int)
		require.True(t, ok)
		assert.Equal(t, "120", resultBig.String()) // 5! = 120
	})

	t.Run("string representation methods", func(t *testing.T) {
		schema := BigInt().Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			return map[string]string{
				"decimal": val.String(),
				"hex":     fmt.Sprintf("0x%x", val),
				"binary":  fmt.Sprintf("0b%b", val),
				"octal":   fmt.Sprintf("0o%o", val),
			}, nil
		})

		result, err := schema.Parse(big.NewInt(42))
		require.NoError(t, err)
		resultMap, ok := result.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "42", resultMap["decimal"])
		assert.Equal(t, "0x2a", resultMap["hex"])
		assert.Equal(t, "0b101010", resultMap["binary"])
		assert.Equal(t, "0o52", resultMap["octal"])
	})

	t.Run("bit manipulation validation", func(t *testing.T) {
		schema := BigInt().Refine(func(val *big.Int) bool {
			// Check if number has exactly 3 bits set
			bitCount := 0
			temp := new(big.Int).Set(val)
			for temp.Cmp(big.NewInt(0)) > 0 {
				if temp.Bit(0) == 1 {
					bitCount++
				}
				temp.Rsh(temp, 1)
			}
			return bitCount == 3
		}, core.SchemaParams{
			Error: "Number must have exactly 3 bits set",
		})

		// 7 = 0b111 (3 bits set)
		result, err := schema.Parse(big.NewInt(7))
		require.NoError(t, err)
		assert.Equal(t, "7", result.(*big.Int).String())

		// 15 = 0b1111 (4 bits set)
		_, err = schema.Parse(big.NewInt(15))
		assert.Error(t, err)
	})
}
