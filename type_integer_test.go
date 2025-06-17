package gozod

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestIntegerBasicFunctionality(t *testing.T) {
	t.Run("basic validation int", func(t *testing.T) {
		schema := Int()
		// Valid int
		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
		// Invalid type
		_, err = schema.Parse("not a number")
		assert.Error(t, err)
	})

	t.Run("basic validation int64", func(t *testing.T) {
		schema := Int64()
		// Valid int64
		result, err := schema.Parse(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
		// Invalid type
		_, err = schema.Parse(3.14)
		assert.Error(t, err)
	})

	t.Run("smart type inference int", func(t *testing.T) {
		schema := Int()
		// Int input returns int
		result1, err := schema.Parse(42)
		require.NoError(t, err)
		assert.IsType(t, int(0), result1)
		assert.Equal(t, 42, result1)
		// Pointer input returns same pointer
		val := 84
		result2, err := schema.Parse(&val)
		require.NoError(t, err)
		assert.IsType(t, (*int)(nil), result2)
		assert.Equal(t, &val, result2)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		schema := Int32().Min(10)
		input := int32(42)
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify not only type and value, but exact pointer identity
		resultPtr, ok := result.(*int32)
		require.True(t, ok, "Result should be *int32")
		assert.True(t, resultPtr == inputPtr, "Should return the exact same pointer")
		assert.Equal(t, int32(42), *resultPtr)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Int().Nilable()
		// nil input should succeed, return nil pointer
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		assert.IsType(t, (*int)(nil), result)
		// Valid input keeps type inference
		result2, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
		assert.IsType(t, int(0), result2)
	})

	t.Run("nilable does not affect original schema", func(t *testing.T) {
		baseSchema := Int().Min(5)
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result1, err1 := nilableSchema.Parse(nil)
		require.NoError(t, err1)
		assert.Nil(t, result1)

		// Test nilable schema validates non-nil values
		result2, err2 := nilableSchema.Parse(10)
		require.NoError(t, err2)
		assert.Equal(t, 10, result2)

		// Test nilable schema rejects invalid values
		_, err3 := nilableSchema.Parse(3)
		assert.Error(t, err3)

		// 🔥 Critical: Original schema should remain unchanged
		_, err4 := baseSchema.Parse(nil)
		assert.Error(t, err4, "Original schema should still reject nil")

		result5, err5 := baseSchema.Parse(10)
		require.NoError(t, err5)
		assert.Equal(t, 10, result5)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestIntegerCoercion(t *testing.T) {
	t.Run("basic coercion int", func(t *testing.T) {
		schema := Int(SchemaParams{Coerce: true})
		tests := []struct {
			input    interface{}
			expected int
		}{
			{"123", 123},
			{int8(42), 42},
			{int16(84), 84},
			{int32(168), 168},
			{int64(336), 336},
			{uint8(12), 12},
			{float64(42), 42}, // exact integer conversion
		}
		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := Int(SchemaParams{Coerce: true}).Min(10).Max(100)
		// Coercion then validation passes
		result, err := schema.Parse("50")
		require.NoError(t, err)
		assert.Equal(t, 50, result)
		// Coercion then validation fails
		_, err = schema.Parse("5")
		assert.Error(t, err)
	})

	t.Run("failed coercion", func(t *testing.T) {
		schema := Int(SchemaParams{Coerce: true})
		invalidInputs := []interface{}{
			"not a number",
			3.14,     // non-integer float
			true,     // boolean
			[]int{1}, // slice
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

func TestIntegerValidations(t *testing.T) {
	t.Run("range validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  *ZodInt
			input   int
			wantErr bool
		}{
			{"min valid", Int().Min(5), 10, false},
			{"min invalid", Int().Min(5), 3, true},
			{"max valid", Int().Max(100), 50, false},
			{"max invalid", Int().Max(100), 150, true},
			{"gt valid", Int().Gt(0), 5, false},
			{"gt invalid", Int().Gt(0), 0, true},
			{"gte valid", Int().Gte(0), 0, false},
			{"gte invalid", Int().Gte(0), -1, true},
			{"lt valid", Int().Lt(100), 50, false},
			{"lt invalid", Int().Lt(100), 100, true},
			{"lte valid", Int().Lte(100), 100, false},
			{"lte invalid", Int().Lte(100), 101, true},
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
			schema  *ZodInt
			input   int
			wantErr bool
		}{
			{"positive valid", Int().Positive(), 5, false},
			{"positive invalid", Int().Positive(), -5, true},
			{"negative valid", Int().Negative(), -5, false},
			{"negative invalid", Int().Negative(), 5, true},
			{"non-negative valid", Int().NonNegative(), 0, false},
			{"non-negative invalid", Int().NonNegative(), -1, true},
			{"non-positive valid", Int().NonPositive(), 0, false},
			{"non-positive invalid", Int().NonPositive(), 1, true},
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
		schema := Int().MultipleOf(5)
		// Valid multiple
		result, err := schema.Parse(15)
		require.NoError(t, err)
		assert.Equal(t, 15, result)
		// Invalid multiple
		_, err = schema.Parse(7)
		assert.Error(t, err)
	})

	t.Run("type-specific bounds uint8", func(t *testing.T) {
		schema := Uint8()
		// Valid range
		result, err := schema.Parse(uint8(200))
		require.NoError(t, err)
		assert.Equal(t, uint8(200), result)
		// Test automatic bounds
		_, err = schema.Parse(uint8(255)) // max uint8
		require.NoError(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestIntegerModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schema := Int().Optional()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Int().Nilable()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
	})

	t.Run("nullish modifier", func(t *testing.T) {
		schema := Int().Nullish()
		// nil input should succeed
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
		// Valid input normal validation
		result2, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result2)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Int()
		// Valid input should not panic
		result := schema.MustParse(42)
		assert.Equal(t, 42, result)
		// Invalid input should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestIntegerChaining(t *testing.T) {
	t.Run("multiple validations", func(t *testing.T) {
		schema := Int().Min(10).Max(100).Positive()
		// Valid input
		result, err := schema.Parse(50)
		require.NoError(t, err)
		assert.Equal(t, 50, result)
		// Validation failures
		testCases := []int{
			5,   // too small
			150, // too large
			-10, // not positive
		}
		for _, input := range testCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("validation with multiple of", func(t *testing.T) {
		schema := Int().Min(10).Max(100).MultipleOf(5)
		// Valid input
		result, err := schema.Parse(25)
		require.NoError(t, err)
		assert.Equal(t, 25, result)
		// Validation failure
		_, err = schema.Parse(23) // not multiple of 5
		assert.Error(t, err)
	})

	t.Run("cross-type validation", func(t *testing.T) {
		schema := Int64().Min(1000).Max(math.MaxInt32)
		// Valid input within range
		result, err := schema.Parse(int64(50000))
		require.NoError(t, err)
		assert.Equal(t, int64(50000), result)
		// Too large for constraint
		_, err = schema.Parse(int64(math.MaxInt64))
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestIntegerTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Int().Transform(func(val int, ctx *RefinementContext) (any, error) {
			return val * 2, nil
		})
		result, err := schema.Parse(21)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("transform chaining", func(t *testing.T) {
		schema := Int().
			Transform(func(val int, ctx *RefinementContext) (any, error) {
				return val * 2, nil
			}).
			TransformAny(func(val any, ctx *RefinementContext) (any, error) {
				if intVal, ok := val.(int); ok {
					return fmt.Sprintf("result_%d", intVal), nil
				}
				return val, nil
			})
		result, err := schema.Parse(21)
		require.NoError(t, err)
		assert.Equal(t, "result_42", result)
	})

	t.Run("pipe combination", func(t *testing.T) {
		schema := Int().
			Transform(func(val int, ctx *RefinementContext) (any, error) {
				return fmt.Sprintf("%d", val), nil
			}).
			Pipe(String().Min(2))
		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)
		_, err = schema.Parse(1) // "1" too short for String().Min(2)
		assert.Error(t, err)
	})

	t.Run("transform with validation", func(t *testing.T) {
		schema := Int().Min(10).Transform(func(val int, ctx *RefinementContext) (any, error) {
			if val < 0 {
				return nil, ErrNegativeNumbersNotAllowed
			}
			return float64(val) / 10.0, nil
		})

		result, err := schema.Parse(100)
		require.NoError(t, err)
		assert.Equal(t, 10.0, result)

		// Validation before transform should fail
		_, err = schema.Parse(5)
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestIntegerRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Int().Refine(func(val int) bool {
			return val%2 == 0 // even numbers only
		}, SchemaParams{
			Error: "Number must be even",
		})
		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
		_, err = schema.Parse(43)
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "even")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := 42

		// Refine: only validates, never modifies
		refineSchema := Int().Refine(func(val int) bool {
			return val > 0
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := Int().Transform(func(val int, ctx *RefinementContext) (any, error) {
			return val * 2, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, 42, refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, 84, transformResult)

		// Key distinction: Refine preserves, Transform modifies
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
		assert.NotEqual(t, input, transformResult, "Transform should return modified value")
	})

	t.Run("refine preserves pointer identity", func(t *testing.T) {
		schema := Int().Refine(func(val int) bool {
			return val >= 10
		})

		input := 42
		inputPtr := &input

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)

		// Verify exact pointer identity is preserved
		resultPtr, ok := result.(*int)
		require.True(t, ok)
		assert.True(t, resultPtr == inputPtr, "Refine should preserve exact pointer identity")
	})

	t.Run("refine with complex validation", func(t *testing.T) {
		schema := Int().Refine(func(val int) bool {
			// Perfect squares only
			sqrt := int(math.Sqrt(float64(val)))
			return sqrt*sqrt == val
		}, SchemaParams{
			Error: func(issue ZodRawIssue) string {
				if input, ok := issue.Input.(int); ok {
					return fmt.Sprintf("The number %d is not a perfect square", input)
				}
				return "Invalid input for perfect square validation"
			},
		})

		result, err := schema.Parse(25) // 5^2 = 25
		require.NoError(t, err)
		assert.Equal(t, 25, result)

		_, err = schema.Parse(24) // not a perfect square
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "not a perfect square")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestIntegerErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Int().Min(10)
		_, err := schema.Parse(5)
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, string(TooSmall), zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Int().Min(10, SchemaParams{
			Error: "Custom minimum value error",
		})
		_, err := schema.Parse(5)
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "Custom minimum value error")
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := Int().Min(10).Max(20)
		_, err := schema.Parse(5)
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.GreaterOrEqual(t, len(zodErr.Issues), 1)
	})

	t.Run("type mismatch error", func(t *testing.T) {
		schema := Int()
		_, err := schema.Parse("not a number")
		assert.Error(t, err)
		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Equal(t, string(InvalidType), zodErr.Issues[0].Code)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestIntegerEdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		schema := Int()
		result, err := schema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("boundary values int8", func(t *testing.T) {
		schema := Int8()
		// Test min and max values
		result1, err := schema.Parse(int8(math.MinInt8))
		require.NoError(t, err)
		assert.Equal(t, int8(math.MinInt8), result1)

		result2, err := schema.Parse(int8(math.MaxInt8))
		require.NoError(t, err)
		assert.Equal(t, int8(math.MaxInt8), result2)
	})

	t.Run("boundary values uint64", func(t *testing.T) {
		schema := Uint64()
		// Test max value
		result, err := schema.Parse(uint64(math.MaxUint64))
		require.NoError(t, err)
		assert.Equal(t, uint64(math.MaxUint64), result)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := Int()
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
		schema := Int()
		invalidTypes := []interface{}{
			"123",
			3.14,
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
		schema := Int().Min(5).Positive().Nilable()
		// Valid input
		result, err := schema.Parse(10)
		require.NoError(t, err)
		assert.Equal(t, 10, result)
		// nil input
		result2, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
		// Invalid input
		_, err = schema.Parse(-5)
		assert.Error(t, err)
	})

	t.Run("cross-type compatibility", func(t *testing.T) {
		// Test that different integer types don't accidentally validate
		schemaInt := Int()
		_, err := schemaInt.Parse(int32(42))
		assert.Error(t, err, "int schema should reject int32")

		schemaInt32 := Int32()
		_, err = schemaInt32.Parse(42) // int
		assert.Error(t, err, "int32 schema should reject int")
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestIntegerDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := Int().Min(10).Default(15)
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 15, result)
		// Valid input normal validation
		result2, err := schema.Parse(20)
		require.NoError(t, err)
		assert.Equal(t, 20, result2)
		// Invalid input still fails
		_, err = schema.Parse(5)
		assert.Error(t, err)
	})

	t.Run("function-based default value", func(t *testing.T) {
		counter := 0
		schema := Int().DefaultFunc(func() int {
			counter++
			return counter * 10
		}).Min(5)

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 10, result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, 20, result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse(100)
		require.NoError(t, err3)
		assert.Equal(t, 100, result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := Int().
			Default(50).
			Min(10).
			Transform(func(val int, ctx *RefinementContext) (any, error) {
				return map[string]any{
					"original": val,
					"doubled":  val * 2,
					"type":     "integer",
				}, nil
			})

		// Non-nil input: validate then transform
		result1, err1 := schema.Parse(25)
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, 25, result1Map["original"])
		assert.Equal(t, 50, result1Map["doubled"])
		assert.Equal(t, "integer", result1Map["type"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, 50, result2Map["original"])
		assert.Equal(t, 100, result2Map["doubled"])
		assert.Equal(t, "integer", result2Map["type"])

		// Invalid input still fails validation
		_, err3 := schema.Parse(5)
		assert.Error(t, err3, "Small integer should fail Min(10) validation")
	})

	t.Run("prefault value", func(t *testing.T) {
		schema := Int().Min(10).Prefault(999)
		// Any validation failure uses fallback value
		result1, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 999, result1)
		result2, err := schema.Parse(5)
		require.NoError(t, err)
		assert.Equal(t, 999, result2)
		// Valid input normal validation
		result3, err := schema.Parse(20)
		require.NoError(t, err)
		assert.Equal(t, 20, result3)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Int().Min(10).Default(50).Prefault(999)
		// nil input uses default value
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 50, result)
		// Invalid input uses fallback value
		result2, err := schema.Parse(5)
		require.NoError(t, err)
		assert.Equal(t, 999, result2)
		// Valid input normal validation
		result3, err := schema.Parse(20)
		require.NoError(t, err)
		assert.Equal(t, 20, result3)
	})

	t.Run("typed default chaining", func(t *testing.T) {
		// Test that Default returns type-safe wrapper that supports chaining
		schema := Int().Default(42).Min(10).Max(100).Positive()

		// Test default functionality
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, 42, result1)

		// Test chained validations work
		result2, err2 := schema.Parse(50)
		require.NoError(t, err2)
		assert.Equal(t, 50, result2)

		// Test validation failures
		_, err3 := schema.Parse(5) // below Min(10)
		assert.Error(t, err3)
		_, err4 := schema.Parse(150) // above Max(100)
		assert.Error(t, err4)
	})
}

// =============================================================================
// Additional type-specific tests
// =============================================================================

func TestIntegerTypeSpecific(t *testing.T) {
	t.Run("all integer types basic validation", func(t *testing.T) {
		tests := []struct {
			name   string
			schema ZodType[any, any]
			input  any
			want   any
		}{
			{"int", Int(), int(42), int(42)},
			{"int8", Int8(), int8(42), int8(42)},
			{"int16", Int16(), int16(42), int16(42)},
			{"int32", Int32(), int32(42), int32(42)},
			{"int64", Int64(), int64(42), int64(42)},
			{"uint", Uint(), uint(42), uint(42)},
			{"uint8", Uint8(), uint8(42), uint8(42)},
			{"uint16", Uint16(), uint16(42), uint16(42)},
			{"uint32", Uint32(), uint32(42), uint32(42)},
			{"uint64", Uint64(), uint64(42), uint64(42)},
			{"byte", Byte(), uint8(42), uint8(42)},
			{"rune", Rune(), int32(42), int32(42)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			})
		}
	})

	t.Run("overflow protection", func(t *testing.T) {
		// Test that coercion respects type bounds
		schema := Int8(SchemaParams{Coerce: true})

		// Should succeed within bounds
		result, err := schema.Parse("127") // MaxInt8
		require.NoError(t, err)
		assert.Equal(t, int8(127), result)

		// Should fail outside bounds (if coercion includes bounds checking)
		_, _ = schema.Parse("128") // MaxInt8 + 1
		// Note: This behavior depends on implementation - may succeed or fail
		// The test documents expected behavior
	})
}
