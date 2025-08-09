package types

import (
	"math/big"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestBigInt_BasicFunctionality(t *testing.T) {
	t.Run("valid big.Int inputs", func(t *testing.T) {
		schema := BigInt()

		// Test positive value
		bigVal := big.NewInt(42)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)

		// Test negative value
		negVal := big.NewInt(-123)
		result, err = schema.Parse(negVal)
		require.NoError(t, err)
		assert.Equal(t, negVal, result)

		// Test zero value
		zeroVal := big.NewInt(0)
		result, err = schema.Parse(zeroVal)
		require.NoError(t, err)
		assert.Equal(t, zeroVal, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := BigInt()

		invalidInputs := []any{
			"not a bigint", 123, 3.14, []byte{1, 2, 3}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := BigInt()
		bigVal := big.NewInt(999)

		// Test Parse method
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)

		// Test MustParse method
		mustResult := schema.MustParse(bigVal)
		assert.Equal(t, bigVal, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestBigInt_TypeSafety(t *testing.T) {
	t.Run("BigInt returns *big.Int type", func(t *testing.T) {
		schema := BigInt()
		require.NotNil(t, schema)

		bigVal := big.NewInt(42)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)
		assert.IsType(t, (*big.Int)(nil), result) // Ensure type is *big.Int
	})

	t.Run("BigIntPtr returns **big.Int type", func(t *testing.T) {
		schema := BigIntPtr()
		require.NotNil(t, schema)

		bigVal := big.NewInt(42)
		result, err := schema.Parse(bigVal)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, bigVal, *result)
		assert.IsType(t, (**big.Int)(nil), result) // Ensure type is **big.Int
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		bigVal := big.NewInt(123)

		// Test *big.Int type
		bigSchema := BigInt()
		result := bigSchema.MustParse(bigVal)
		assert.IsType(t, (*big.Int)(nil), result)
		assert.Equal(t, bigVal, result)

		// Test **big.Int type
		ptrSchema := BigIntPtr().Nilable().Overwrite(func(bi **big.Int) **big.Int {
			if bi == nil || *bi == nil {
				return nil
			}
			abs := new(big.Int).Abs(*bi)
			return &abs
		})
		ptrResult := ptrSchema.MustParse(bigVal)
		assert.IsType(t, (**big.Int)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, bigVal, *ptrResult)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestBigInt_Modifiers(t *testing.T) {
	t.Run("Optional always returns **big.Int", func(t *testing.T) {
		// From *big.Int to **big.Int via Optional
		bigSchema := BigInt()
		optionalSchema := bigSchema.Optional()

		// Type check: ensure it returns *ZodBigInt[**big.Int]
		var _ *ZodBigInt[**big.Int] = optionalSchema

		// Functionality test
		bigVal := big.NewInt(42)
		result, err := optionalSchema.Parse(bigVal)
		require.NoError(t, err)
		assert.IsType(t, (**big.Int)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, bigVal, *result)
	})

	t.Run("Nilable always returns **big.Int", func(t *testing.T) {
		bigSchema := BigInt()
		nilableSchema := bigSchema.Nilable()

		var _ *ZodBigInt[**big.Int] = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultVal := big.NewInt(100)

		// *big.Int maintains *big.Int
		bigSchema := BigInt()
		defaultBigSchema := bigSchema.Default(defaultVal)
		var _ *ZodBigInt[*big.Int] = defaultBigSchema

		// **big.Int maintains **big.Int
		ptrSchema := BigIntPtr()
		defaultPtrSchema := ptrSchema.Default(defaultVal)
		var _ *ZodBigInt[**big.Int] = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultVal := big.NewInt(50)

		// *big.Int maintains *big.Int
		bigSchema := BigInt()
		prefaultBigSchema := bigSchema.Prefault(prefaultVal)
		var _ *ZodBigInt[*big.Int] = prefaultBigSchema

		// **big.Int maintains **big.Int
		ptrSchema := BigIntPtr()
		prefaultPtrSchema := ptrSchema.Prefault(prefaultVal)
		var _ *ZodBigInt[**big.Int] = prefaultPtrSchema
	})

	t.Run("Default priority over Prefault", func(t *testing.T) {
		defaultVal := big.NewInt(100)
		prefaultVal := big.NewInt(200)
		schema := BigIntPtr().Min(big.NewInt(150)).Default(defaultVal).Prefault(prefaultVal)

		// Test with nil input - should use default (100), not prefault (200)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultVal, *result)
	})

	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		defaultVal := big.NewInt(5) // This violates Min(10) but should still work
		schema := BigIntPtr().Min(big.NewInt(10)).Default(defaultVal)

		// Test with nil input - should use default without validation
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultVal, *result)
	})

	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		prefaultVal := big.NewInt(100)
		schema := BigInt().Min(big.NewInt(50)).Prefault(prefaultVal)

		// Test with nil input - should use prefault
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, prefaultVal, result)

		// Test with invalid non-nil input - should return error, not prefault
		invalidVal := big.NewInt(10) // Less than Min(50)
		_, err = schema.Parse(invalidVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 50")
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		prefaultVal := big.NewInt(5) // This violates Min(10)
		schema := BigIntPtr().Min(big.NewInt(10)).Prefault(prefaultVal)

		// Test with nil input - prefault should fail validation
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 10")
	})

	t.Run("Prefault error handling", func(t *testing.T) {
		prefaultVal := big.NewInt(5) // This violates Min(10)
		schema := BigInt().Min(big.NewInt(10)).Prefault(prefaultVal)

		// Test with nil input - should return validation error for prefault value
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 10")
	})

	t.Run("BigIntPtr Prefault only on nil input", func(t *testing.T) {
		prefaultVal := big.NewInt(100)
		schema := BigIntPtr().Min(big.NewInt(50)).Prefault(prefaultVal)

		// Test with nil input - should use prefault
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, prefaultVal, *result)

		// Test with invalid non-nil input - should return error, not prefault
		invalidVal := big.NewInt(10) // Less than Min(50)
		_, err = schema.Parse(invalidVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected bigint to be at least 50")
	})
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestBigInt_Validations(t *testing.T) {
	t.Run("Min validation", func(t *testing.T) {
		schema := BigInt().Min(big.NewInt(10))

		// Valid: value >= minimum
		result, err := schema.Parse(big.NewInt(15))
		require.NoError(t, err)
		expected := big.NewInt(15)
		assert.Equal(t, expected, result)

		// Valid: value == minimum
		result, err = schema.Parse(big.NewInt(10))
		require.NoError(t, err)
		expected = big.NewInt(10)
		assert.Equal(t, expected, result)

		// Invalid: value < minimum
		_, err = schema.Parse(big.NewInt(5))
		assert.Error(t, err)
	})

	t.Run("Max validation", func(t *testing.T) {
		schema := BigInt().Max(big.NewInt(100))

		// Valid: value <= maximum
		result, err := schema.Parse(big.NewInt(50))
		require.NoError(t, err)
		expected := big.NewInt(50)
		assert.Equal(t, expected, result)

		// Valid: value == maximum
		result, err = schema.Parse(big.NewInt(100))
		require.NoError(t, err)
		expected = big.NewInt(100)
		assert.Equal(t, expected, result)

		// Invalid: value > maximum
		_, err = schema.Parse(big.NewInt(150))
		assert.Error(t, err)
	})

	t.Run("Positive validation", func(t *testing.T) {
		schema := BigInt().Positive()

		// Valid: positive number
		result, err := schema.Parse(big.NewInt(42))
		require.NoError(t, err)
		expected := big.NewInt(42)
		assert.Equal(t, expected, result)

		// Invalid: zero
		_, err = schema.Parse(big.NewInt(0))
		assert.Error(t, err)

		// Invalid: negative
		_, err = schema.Parse(big.NewInt(-1))
		assert.Error(t, err)
	})

	t.Run("Negative validation", func(t *testing.T) {
		schema := BigInt().Negative()

		// Valid: negative number
		result, err := schema.Parse(big.NewInt(-42))
		require.NoError(t, err)
		expected := big.NewInt(-42)
		assert.Equal(t, expected, result)

		// Invalid: zero
		_, err = schema.Parse(big.NewInt(0))
		assert.Error(t, err)

		// Invalid: positive
		_, err = schema.Parse(big.NewInt(1))
		assert.Error(t, err)
	})
}

// =============================================================================
// Coercion tests
// =============================================================================

func TestBigInt_Coercion(t *testing.T) {
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
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err, "Failed to parse %v", tt.input)
			assert.Equal(t, tt.expected, result.String())
		}
	})

	t.Run("large number coercion", func(t *testing.T) {
		schema := CoercedBigInt()
		largeNumber := "123456789012345678901234567890"
		result, err := schema.Parse(largeNumber)
		require.NoError(t, err)
		assert.Equal(t, largeNumber, result.String())
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedBigInt().Min(big.NewInt(5))

		// Coercion then validation passes
		result, err := schema.Parse("10")
		require.NoError(t, err)
		assert.Equal(t, "10", result.String())

		// Coercion then validation fails
		_, err = schema.Parse("3")
		assert.Error(t, err)
	})
}

// =============================================================================
// Chaining and Transform tests
// =============================================================================

func TestBigInt_Chaining(t *testing.T) {
	t.Run("chain multiple validations", func(t *testing.T) {
		schema := BigInt().
			Min(big.NewInt(10)).
			Max(big.NewInt(100)).
			Positive()

		// Valid value in range
		result, err := schema.Parse(big.NewInt(50))
		require.NoError(t, err)
		expected := big.NewInt(50)
		assert.Equal(t, expected, result)

		// Invalid: below minimum
		_, err = schema.Parse(big.NewInt(5))
		assert.Error(t, err)

		// Invalid: above maximum
		_, err = schema.Parse(big.NewInt(150))
		assert.Error(t, err)
	})
}

func TestBigInt_Transform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := BigInt()
		transform := schema.Transform(func(val *big.Int, ctx *core.RefinementContext) (any, error) {
			// Convert to string representation
			return val.String(), nil
		})

		result, err := transform.Parse(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})
}

// =============================================================================
// Overwrite functionality tests
// =============================================================================

func TestBigInt_Overwrite(t *testing.T) {
	t.Run("basic transformations", func(t *testing.T) {
		// Test absolute value transformation
		absSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Abs(bi)
		})

		// Test with positive number
		positiveInput := big.NewInt(42)
		result, err := absSchema.Parse(positiveInput)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Int64())

		// Test with negative number
		negativeInput := big.NewInt(-42)
		result, err = absSchema.Parse(negativeInput)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Int64())
	})

	t.Run("arithmetic transformations", func(t *testing.T) {
		// Test multiplication by 2
		doubleSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mul(bi, big.NewInt(2))
		})

		input := big.NewInt(21)
		result, err := doubleSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Int64())

		// Test addition
		addTenSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Add(bi, big.NewInt(10))
		})

		input = big.NewInt(5)
		result, err = addTenSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, int64(15), result.Int64())
	})

	t.Run("modular arithmetic", func(t *testing.T) {
		// Test modulo operation
		modSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mod(bi, big.NewInt(10))
		})

		input := big.NewInt(123)
		result, err := modSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Int64())

		// Test with negative number
		input = big.NewInt(-17)
		result, err = modSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Int64()) // Go's mod behavior: -17 % 10 = 3
	})

	t.Run("power and root operations", func(t *testing.T) {
		// Test square operation
		squareSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Mul(bi, bi)
		})

		input := big.NewInt(7)
		result, err := squareSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, int64(49), result.Int64())

		// Test exponentiation
		powerSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Exp(bi, big.NewInt(3), nil)
		})

		input = big.NewInt(4)
		result, err = powerSchema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, int64(64), result.Int64())
	})

	t.Run("conditional transformations", func(t *testing.T) {
		// Test conditional transformation based on value
		clampSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			// Clamp between 0 and 100
			if bi.Cmp(big.NewInt(0)) < 0 {
				return big.NewInt(0)
			}
			if bi.Cmp(big.NewInt(100)) > 0 {
				return big.NewInt(100)
			}
			return new(big.Int).Set(bi) // Return a copy
		})

		// Test negative value (should be clamped to 0)
		result, err := clampSchema.Parse(big.NewInt(-10))
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Int64())

		// Test value within range (should remain unchanged)
		result, err = clampSchema.Parse(big.NewInt(50))
		require.NoError(t, err)
		assert.Equal(t, int64(50), result.Int64())

		// Test value above range (should be clamped to 100)
		result, err = clampSchema.Parse(big.NewInt(150))
		require.NoError(t, err)
		assert.Equal(t, int64(100), result.Int64())
	})

	t.Run("chaining with validations", func(t *testing.T) {
		// Test chaining Overwrite with Min validation
		positiveDoubleSchema := BigInt().
			Overwrite(func(bi *big.Int) *big.Int {
				return new(big.Int).Mul(bi, big.NewInt(2))
			}).
			Min(big.NewInt(10))

		// Test with value that becomes valid after transformation
		result, err := positiveDoubleSchema.Parse(big.NewInt(7))
		require.NoError(t, err)
		assert.Equal(t, int64(14), result.Int64())

		// Test with value that fails validation even after transformation
		_, err = positiveDoubleSchema.Parse(big.NewInt(2))
		assert.Error(t, err)
	})

	t.Run("multiple overwrite calls", func(t *testing.T) {
		// Test chaining multiple Overwrite calls
		multiTransformSchema := BigInt().
			Overwrite(func(bi *big.Int) *big.Int {
				return new(big.Int).Abs(bi) // First: get absolute value
			}).
			Overwrite(func(bi *big.Int) *big.Int {
				return new(big.Int).Mul(bi, big.NewInt(3)) // Second: multiply by 3
			}).
			Overwrite(func(bi *big.Int) *big.Int {
				return new(big.Int).Add(bi, big.NewInt(1)) // Third: add 1
			})

		// Test with negative input: -5 -> 5 -> 15 -> 16
		result, err := multiTransformSchema.Parse(big.NewInt(-5))
		require.NoError(t, err)
		assert.Equal(t, int64(16), result.Int64())

		// Test with positive input: 3 -> 3 -> 9 -> 10
		result, err = multiTransformSchema.Parse(big.NewInt(3))
		require.NoError(t, err)
		assert.Equal(t, int64(10), result.Int64())
	})

	t.Run("large number handling", func(t *testing.T) {
		// Test with very large numbers
		largeNumSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			// Add a large number
			large := new(big.Int)
			large.SetString("999999999999999999999999999999", 10)
			return new(big.Int).Add(bi, large)
		})

		input := new(big.Int)
		input.SetString("123456789012345678901234567890", 10)

		result, err := largeNumSchema.Parse(input)
		require.NoError(t, err)

		expected := new(big.Int)
		expected.SetString("1123456789012345678901234567889", 10)
		assert.Equal(t, 0, result.Cmp(expected))
	})

	t.Run("type preservation", func(t *testing.T) {
		// Test that Overwrite preserves the original type
		bigIntSchema := BigInt()
		overwriteSchema := bigIntSchema.Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Abs(bi)
		})

		// Both should have the same type
		testValue := big.NewInt(-42)

		result1, err1 := bigIntSchema.Parse(testValue)
		require.NoError(t, err1)

		result2, err2 := overwriteSchema.Parse(testValue)
		require.NoError(t, err2)

		// Both results should be of type *big.Int
		assert.IsType(t, (*big.Int)(nil), result1)
		assert.IsType(t, (*big.Int)(nil), result2)
	})

	t.Run("pointer type handling", func(t *testing.T) {
		// Pointer Overwrite should now work and preserve pointer identity
		ptrSchema := BigIntPtr().Nilable().Overwrite(func(bi **big.Int) **big.Int {
			if bi == nil || *bi == nil {
				return nil
			}
			abs := new(big.Int).Abs(*bi)
			return &abs
		})

		// Test with non-nil value
		testValue := big.NewInt(-42)
		result, err := ptrSchema.Parse(testValue)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, *result)
		assert.Equal(t, int64(42), (*result).Int64())

		// Test with nil value
		result, err = ptrSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("coerced bigint overwrite", func(t *testing.T) {
		// Test with coerced BigInt schema
		coercedSchema := CoercedBigInt().Overwrite(func(bi *big.Int) *big.Int {
			// Always return the square of the input
			return new(big.Int).Mul(bi, bi)
		})

		// Test with string input that can be coerced
		result, err := coercedSchema.Parse("7")
		require.NoError(t, err)
		assert.Equal(t, int64(49), result.Int64())

		// Test with int input that can be coerced
		result, err = coercedSchema.Parse(6)
		require.NoError(t, err)
		assert.Equal(t, int64(36), result.Int64())

		// Test with float input that can be coerced
		result, err = coercedSchema.Parse(5.0)
		require.NoError(t, err)
		assert.Equal(t, int64(25), result.Int64())
	})

	t.Run("error handling", func(t *testing.T) {
		// Test that invalid inputs still produce errors
		schema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Abs(bi)
		})

		// Invalid input should still cause an error
		_, err := schema.Parse("not a number")
		assert.Error(t, err)

		_, err = schema.Parse(3.14)
		assert.Error(t, err)

		_, err = schema.Parse(nil)
		assert.Error(t, err)

		_, err = schema.Parse([]int{1, 2, 3})
		assert.Error(t, err)
	})

	t.Run("immutability", func(t *testing.T) {
		// Test that original values are not modified
		originalValue := big.NewInt(42)
		originalCopy := new(big.Int).Set(originalValue)

		negateSchema := BigInt().Overwrite(func(bi *big.Int) *big.Int {
			return new(big.Int).Neg(bi)
		})

		result, err := negateSchema.Parse(originalValue)
		require.NoError(t, err)

		// Result should be negated
		assert.Equal(t, int64(-42), result.Int64())

		// Original value should remain unchanged
		assert.Equal(t, 0, originalValue.Cmp(originalCopy))
		assert.Equal(t, int64(42), originalValue.Int64())
	})
}

// =============================================================================
// NonOptional tests
// =============================================================================

func TestBigInt_NonOptional(t *testing.T) {
	// basic
	schema := BigInt().NonOptional()

	v := big.NewInt(123)
	r, err := schema.Parse(v)
	require.NoError(t, err)
	assert.Equal(t, v, r)
	assert.IsType(t, (*big.Int)(nil), r)

	_, err = schema.Parse(nil)
	assert.Error(t, err)
	var zErr *issues.ZodError
	if issues.IsZodError(err, &zErr) {
		assert.Equal(t, core.InvalidType, zErr.Issues[0].Code)
		assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
	}

	// Optional().NonOptional()
	chain := BigInt().Optional().NonOptional()
	var _ *ZodBigInt[*big.Int] = chain
	_, err = chain.Parse(nil)
	assert.Error(t, err)

	// object embedding
	obj := Object(map[string]core.ZodSchema{
		"id": BigInt().Optional().NonOptional(),
	})
	_, err = obj.Parse(map[string]any{"id": big.NewInt(999)})
	require.NoError(t, err)
	_, err = obj.Parse(map[string]any{"id": nil})
	assert.Error(t, err)

	// BigIntPtr().NonOptional()
	ptrSchema := BigIntPtr().NonOptional()
	var _ *ZodBigInt[*big.Int] = ptrSchema
	vv := big.NewInt(456)
	res2, err := ptrSchema.Parse(&vv)
	require.NoError(t, err)
	assert.Equal(t, vv, res2)
	_, err = ptrSchema.Parse(nil)
	assert.Error(t, err)
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestBigInt_StrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := BigInt()

		// Test StrictParse with exact type match
		bigVal := big.NewInt(12345)
		result, err := schema.StrictParse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)
		assert.IsType(t, (*big.Int)(nil), result)

		// Test StrictParse with negative value
		negVal := big.NewInt(-9876)
		negResult, err := schema.StrictParse(negVal)
		require.NoError(t, err)
		assert.Equal(t, negVal, negResult)

		// Test StrictParse with zero
		zeroVal := big.NewInt(0)
		zeroResult, err := schema.StrictParse(zeroVal)
		require.NoError(t, err)
		assert.Equal(t, zeroVal, zeroResult)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := BigInt().Refine(func(b *big.Int) bool {
			return b.Cmp(big.NewInt(100)) >= 0 // Must be >= 100
		}, "Must be at least 100")

		// Valid case
		bigVal := big.NewInt(150)
		result, err := schema.StrictParse(bigVal)
		require.NoError(t, err)
		assert.Equal(t, bigVal, result)

		// Invalid case - too small
		smallVal := big.NewInt(50)
		_, err = schema.StrictParse(smallVal)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Must be at least 100")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := BigIntPtr()
		bigVal := big.NewInt(999)

		// Test with valid pointer input
		result, err := schema.StrictParse(&bigVal)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, bigVal, *result)
		assert.IsType(t, (**big.Int)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		defaultVal := big.NewInt(42)
		schema := BigIntPtr().Default(defaultVal)
		var nilPtr **big.Int = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultVal, *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		prefaultVal := big.NewInt(1000)
		schema := BigIntPtr().Refine(func(b **big.Int) bool {
			return b != nil && *b != nil && (*b).Cmp(big.NewInt(500)) >= 0
		}, "Must be at least 500").Prefault(prefaultVal)
		smallVal := big.NewInt(100) // Too small

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema.StrictParse(&smallVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Must be at least 500")

		// Test with nil input (should use prefault)
		result, err := schema.StrictParse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, prefaultVal, *result)
	})

	t.Run("large numbers", func(t *testing.T) {
		schema := BigInt()

		// Test with very large number
		largeVal := new(big.Int)
		largeVal.SetString("123456789012345678901234567890", 10)
		result, err := schema.StrictParse(largeVal)
		require.NoError(t, err)
		assert.Equal(t, largeVal, result)

		// Test with very large negative number
		negLargeVal := new(big.Int)
		negLargeVal.SetString("-987654321098765432109876543210", 10)
		negResult, err := schema.StrictParse(negLargeVal)
		require.NoError(t, err)
		assert.Equal(t, negLargeVal, negResult)
	})
}

func TestBigInt_MustStrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := BigInt()

		// Test MustStrictParse with valid input
		bigVal := big.NewInt(54321)
		result := schema.MustStrictParse(bigVal)
		assert.Equal(t, bigVal, result)
		assert.IsType(t, (*big.Int)(nil), result)

		// Test MustStrictParse with zero
		zeroVal := big.NewInt(0)
		zeroResult := schema.MustStrictParse(zeroVal)
		assert.Equal(t, zeroVal, zeroResult)

		// Test MustStrictParse with negative value
		negVal := big.NewInt(-11111)
		negResult := schema.MustStrictParse(negVal)
		assert.Equal(t, negVal, negResult)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := BigInt().Refine(func(b *big.Int) bool {
			return b.Cmp(big.NewInt(1000)) >= 0 // Must be >= 1000
		}, "Must be at least 1000")

		// Test panic with validation failure
		smallVal := big.NewInt(500)
		assert.Panics(t, func() {
			schema.MustStrictParse(smallVal) // Too small, should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := BigIntPtr()
		bigVal := big.NewInt(77777)

		// Test with valid pointer input
		result := schema.MustStrictParse(&bigVal)
		require.NotNil(t, result)
		assert.Equal(t, bigVal, *result)
		assert.IsType(t, (**big.Int)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		defaultVal := big.NewInt(88888)
		schema := BigIntPtr().Default(defaultVal)
		var nilPtr **big.Int = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, defaultVal, *result)
	})

	t.Run("large numbers", func(t *testing.T) {
		schema := BigInt()

		// Test with very large number
		largeVal := new(big.Int)
		largeVal.SetString("999888777666555444333222111000", 10)
		result := schema.MustStrictParse(largeVal)
		assert.Equal(t, largeVal, result)

		// Test with very large negative number
		negLargeVal := new(big.Int)
		negLargeVal.SetString("-111222333444555666777888999000", 10)
		negResult := schema.MustStrictParse(negLargeVal)
		assert.Equal(t, negLargeVal, negResult)
	})
}
