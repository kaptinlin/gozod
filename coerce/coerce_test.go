package coerce

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Boolean type tests
// =============================================================================

func TestBool_Coercion(t *testing.T) {
	t.Run("Bool creates coerced boolean schema", func(t *testing.T) {
		schema := Bool()
		require.NotNil(t, schema)

		// Test string coercion
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.IsType(t, bool(false), result)

		result, err = schema.Parse("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Test empty string coercion (should be false)
		result, err = schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Test numeric coercion
		result, err = schema.Parse(1)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse(0)
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Test direct boolean
		result, err = schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("BoolPtr creates coerced *bool schema", func(t *testing.T) {
		schema := BoolPtr()
		require.NotNil(t, schema)

		// Test string coercion
		result, err := schema.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
		assert.IsType(t, (*bool)(nil), result)

		// Test numeric coercion
		result, err = schema.Parse(1)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		// Test direct boolean
		result, err = schema.Parse(false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("Bool with modifiers", func(t *testing.T) {
		schema := Bool().Optional()

		result, err := schema.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Bool invalid coercion", func(t *testing.T) {
		schema := Bool()

		_, err := schema.Parse("invalid")
		assert.Error(t, err)

		_, err = schema.Parse([]string{"true"})
		assert.Error(t, err)
	})
}

// =============================================================================
// String type tests
// =============================================================================

func TestString_Coercion(t *testing.T) {
	t.Run("String creates coerced string schema", func(t *testing.T) {
		schema := String()
		require.NotNil(t, schema)

		// Test numeric coercion
		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)
		assert.IsType(t, "", result)

		result, err = schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, "3.14", result)

		// Test boolean coercion
		result, err = schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, "true", result)

		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, "false", result)

		// Test direct string
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("StringPtr creates coerced *string schema", func(t *testing.T) {
		schema := StringPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse(123)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "123", *result)
		assert.IsType(t, (*string)(nil), result)
	})

	t.Run("String with modifiers", func(t *testing.T) {
		schema := String().Optional()

		result, err := schema.Parse(123)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "123", *result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Numeric type tests
// =============================================================================

func TestNumber_Coercion(t *testing.T) {
	t.Run("Number creates coerced float64 schema", func(t *testing.T) {
		schema := Number()
		require.NotNil(t, schema)

		// Test string coercion
		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		assert.Equal(t, 123.45, result)
		assert.IsType(t, float64(0), result)

		// Test integer coercion
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123.0, result)

		// Test direct float
		result, err = schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("NumberPtr creates coerced *float64 schema", func(t *testing.T) {
		schema := NumberPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 123.45, *result)
		assert.IsType(t, (*float64)(nil), result)
	})

	t.Run("Number invalid coercion", func(t *testing.T) {
		schema := Number()

		_, err := schema.Parse("not a number")
		assert.Error(t, err)

		_, err = schema.Parse([]int{123})
		assert.Error(t, err)
	})
}

// =============================================================================
// Float type tests
// =============================================================================

func TestFloat_Coercion(t *testing.T) {
	t.Run("Float creates coerced float64 schema", func(t *testing.T) {
		schema := Float()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		assert.Equal(t, 123.45, result)
		assert.IsType(t, float64(0), result)
	})

	t.Run("FloatPtr creates coerced *float64 schema", func(t *testing.T) {
		schema := FloatPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 123.45, *result)
		assert.IsType(t, (*float64)(nil), result)
	})

	t.Run("Float32 creates coerced float32 schema", func(t *testing.T) {
		schema := Float32()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		assert.InDelta(t, float32(123.45), result, 0.001)
		assert.IsType(t, float32(0), result)
	})

	t.Run("Float32Ptr creates coerced *float32 schema", func(t *testing.T) {
		schema := Float32Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.InDelta(t, float32(123.45), *result, 0.001)
		assert.IsType(t, (*float32)(nil), result)
	})

	t.Run("Float64 creates coerced float64 schema", func(t *testing.T) {
		schema := Float64()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		assert.Equal(t, 123.45, result)
		assert.IsType(t, float64(0), result)
	})

	t.Run("Float64Ptr creates coerced *float64 schema", func(t *testing.T) {
		schema := Float64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123.45")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 123.45, *result)
		assert.IsType(t, (*float64)(nil), result)
	})
}

// =============================================================================
// Integer type tests
// =============================================================================

func TestInteger_Coercion(t *testing.T) {
	t.Run("Integer creates coerced int64 schema", func(t *testing.T) {
		schema := Integer()
		require.NotNil(t, schema)

		// Test string coercion
		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, int64(123), result)
		assert.IsType(t, int64(0), result)

		// Test empty string coercion (should be 0)
		result, err = schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)

		// Test float coercion
		result, err = schema.Parse(123.0)
		require.NoError(t, err)
		assert.Equal(t, int64(123), result)

		// Test direct int
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, int64(123), result)
	})

	t.Run("IntegerPtr creates coerced *int64 schema", func(t *testing.T) {
		schema := IntegerPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(123), *result)
		assert.IsType(t, (*int64)(nil), result)
	})

	t.Run("Int creates coerced int schema", func(t *testing.T) {
		schema := Int()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		assert.IsType(t, int(0), result)

		// Test empty string coercion (should be 0)
		result, err = schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("IntPtr creates coerced *int schema", func(t *testing.T) {
		schema := IntPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 123, *result)
		assert.IsType(t, (*int)(nil), result)
	})

	t.Run("Int8 creates coerced int8 schema", func(t *testing.T) {
		schema := Int8()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, int8(123), result)
		assert.IsType(t, int8(0), result)
	})

	t.Run("Int8Ptr creates coerced *int8 schema", func(t *testing.T) {
		schema := Int8Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int8(123), *result)
		assert.IsType(t, (*int8)(nil), result)
	})

	t.Run("Int16 creates coerced int16 schema", func(t *testing.T) {
		schema := Int16()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, int16(123), result)
		assert.IsType(t, int16(0), result)
	})

	t.Run("Int16Ptr creates coerced *int16 schema", func(t *testing.T) {
		schema := Int16Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int16(123), *result)
		assert.IsType(t, (*int16)(nil), result)
	})

	t.Run("Int32 creates coerced int32 schema", func(t *testing.T) {
		schema := Int32()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, int32(123), result)
		assert.IsType(t, int32(0), result)
	})

	t.Run("Int32Ptr creates coerced *int32 schema", func(t *testing.T) {
		schema := Int32Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(123), *result)
		assert.IsType(t, (*int32)(nil), result)
	})

	t.Run("Int64 creates coerced int64 schema", func(t *testing.T) {
		schema := Int64()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, int64(123), result)
		assert.IsType(t, int64(0), result)
	})

	t.Run("Int64Ptr creates coerced *int64 schema", func(t *testing.T) {
		schema := Int64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(123), *result)
		assert.IsType(t, (*int64)(nil), result)
	})

	t.Run("Integer invalid coercion", func(t *testing.T) {
		schema := Integer()

		_, err := schema.Parse("not a number")
		assert.Error(t, err)

		_, err = schema.Parse(123.45) // float with decimal part
		assert.Error(t, err)
	})
}

// =============================================================================
// Unsigned integer type tests
// =============================================================================

func TestUnsignedInteger_Coercion(t *testing.T) {
	t.Run("Uint creates coerced uint schema", func(t *testing.T) {
		schema := Uint()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, uint(123), result)
		assert.IsType(t, uint(0), result)
	})

	t.Run("UintPtr creates coerced *uint schema", func(t *testing.T) {
		schema := UintPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint(123), *result)
		assert.IsType(t, (*uint)(nil), result)
	})

	t.Run("Uint8 creates coerced uint8 schema", func(t *testing.T) {
		schema := Uint8()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, uint8(123), result)
		assert.IsType(t, uint8(0), result)
	})

	t.Run("Uint8Ptr creates coerced *uint8 schema", func(t *testing.T) {
		schema := Uint8Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint8(123), *result)
		assert.IsType(t, (*uint8)(nil), result)
	})

	t.Run("Uint16 creates coerced uint16 schema", func(t *testing.T) {
		schema := Uint16()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, uint16(123), result)
		assert.IsType(t, uint16(0), result)
	})

	t.Run("Uint16Ptr creates coerced *uint16 schema", func(t *testing.T) {
		schema := Uint16Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint16(123), *result)
		assert.IsType(t, (*uint16)(nil), result)
	})

	t.Run("Uint32 creates coerced uint32 schema", func(t *testing.T) {
		schema := Uint32()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, uint32(123), result)
		assert.IsType(t, uint32(0), result)
	})

	t.Run("Uint32Ptr creates coerced *uint32 schema", func(t *testing.T) {
		schema := Uint32Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint32(123), *result)
		assert.IsType(t, (*uint32)(nil), result)
	})

	t.Run("Uint64 creates coerced uint64 schema", func(t *testing.T) {
		schema := Uint64()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		assert.Equal(t, uint64(123), result)
		assert.IsType(t, uint64(0), result)
	})

	t.Run("Uint64Ptr creates coerced *uint64 schema", func(t *testing.T) {
		schema := Uint64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(123), *result)
		assert.IsType(t, (*uint64)(nil), result)
	})

	t.Run("Uint negative value error", func(t *testing.T) {
		schema := Uint()

		_, err := schema.Parse("-123")
		assert.Error(t, err)

		_, err = schema.Parse(-123)
		assert.Error(t, err)
	})
}

// =============================================================================
// Complex type tests
// =============================================================================

func TestComplex_Coercion(t *testing.T) {
	t.Run("Complex creates coerced complex128 schema", func(t *testing.T) {
		schema := Complex()
		require.NotNil(t, schema)

		// Test direct complex
		result, err := schema.Parse(complex(3, 4))
		require.NoError(t, err)
		assert.Equal(t, complex128(3+4i), result)
		assert.IsType(t, complex128(0), result)

		// Test numeric coercion
		result, err = schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, complex128(3.14+0i), result)
	})

	t.Run("ComplexPtr creates coerced *complex128 schema", func(t *testing.T) {
		schema := ComplexPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex(3, 4))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex128(3+4i), *result)
		assert.IsType(t, (*complex128)(nil), result)
	})

	t.Run("Complex64 creates coerced complex64 schema", func(t *testing.T) {
		schema := Complex64()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex64(3 + 4i))
		require.NoError(t, err)
		assert.Equal(t, complex64(3+4i), result)
		assert.IsType(t, complex64(0), result)
	})

	t.Run("Complex64Ptr creates coerced *complex64 schema", func(t *testing.T) {
		schema := Complex64Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex64(3 + 4i))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex64(3+4i), *result)
		assert.IsType(t, (*complex64)(nil), result)
	})

	t.Run("Complex128 creates coerced complex128 schema", func(t *testing.T) {
		schema := Complex128()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex128(3 + 4i))
		require.NoError(t, err)
		assert.Equal(t, complex128(3+4i), result)
		assert.IsType(t, complex128(0), result)
	})

	t.Run("Complex128Ptr creates coerced *complex128 schema", func(t *testing.T) {
		schema := Complex128Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse(complex128(3 + 4i))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, complex128(3+4i), *result)
		assert.IsType(t, (*complex128)(nil), result)
	})
}

// =============================================================================
// BigInt type tests
// =============================================================================

func TestBigInt_Coercion(t *testing.T) {
	t.Run("BigInt creates coerced *big.Int schema", func(t *testing.T) {
		schema := BigInt()
		require.NotNil(t, schema)

		// Test string coercion
		result, err := schema.Parse("123456789012345678901234567890")
		require.NoError(t, err)
		require.NotNil(t, result)
		expected := big.NewInt(0)
		expected.SetString("123456789012345678901234567890", 10)
		assert.Equal(t, expected, result)
		assert.IsType(t, (*big.Int)(nil), result)

		// Test int coercion
		result, err = schema.Parse(123)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, big.NewInt(123), result)
	})

	t.Run("BigIntPtr creates coerced **big.Int schema", func(t *testing.T) {
		schema := BigIntPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, *result)
		assert.Equal(t, big.NewInt(123), *result)
		assert.IsType(t, (**big.Int)(nil), result)
	})

	t.Run("BigInt invalid coercion", func(t *testing.T) {
		schema := BigInt()

		_, err := schema.Parse("not a number")
		assert.Error(t, err)

		_, err = schema.Parse([]int{123})
		assert.Error(t, err)
	})
}

// =============================================================================
// Time type tests
// =============================================================================

func TestTime_Coercion(t *testing.T) {
	t.Run("Time creates coerced time.Time schema", func(t *testing.T) {
		schema := Time()
		require.NotNil(t, schema)

		// Test string coercion
		result, err := schema.Parse("2023-01-15T10:30:00Z")
		require.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, "2023-01-15T10:30:00Z")
		assert.Equal(t, expected, result)
		assert.IsType(t, time.Time{}, result)

		// Test direct time.Time
		now := time.Now()
		result, err = schema.Parse(now)
		require.NoError(t, err)
		assert.Equal(t, now, result)
	})

	t.Run("TimePtr creates coerced *time.Time schema", func(t *testing.T) {
		schema := TimePtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("2023-01-15T10:30:00Z")
		require.NoError(t, err)
		require.NotNil(t, result)
		expected, _ := time.Parse(time.RFC3339, "2023-01-15T10:30:00Z")
		assert.Equal(t, expected, *result)
		assert.IsType(t, (*time.Time)(nil), result)
	})

	t.Run("Time invalid coercion", func(t *testing.T) {
		schema := Time()

		_, err := schema.Parse("not a time")
		assert.Error(t, err)

		// Note: Integer to time coercion might be supported (unix timestamp)
		// so we test with a clearly invalid type instead
		_, err = schema.Parse([]int{123})
		assert.Error(t, err)
	})
}

// =============================================================================
// StringBool type tests
// =============================================================================

func TestStringBool_Coercion(t *testing.T) {
	t.Run("StringBool creates coerced bool schema", func(t *testing.T) {
		schema := StringBool()
		require.NotNil(t, schema)

		// Test string to bool coercion
		testCases := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"1", true},
			{"0", false},
			{"yes", true},
			{"no", false},
			{"on", true},
			{"off", false},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result, err := schema.Parse(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
				assert.IsType(t, bool(false), result)
			})
		}

		// Test direct boolean
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("StringBoolPtr creates coerced *bool schema", func(t *testing.T) {
		schema := StringBoolPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
		assert.IsType(t, (*bool)(nil), result)
	})

	t.Run("StringBool invalid coercion", func(t *testing.T) {
		schema := StringBool()

		_, err := schema.Parse("maybe")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// Parameter passing tests
// =============================================================================

func TestCoercion_ParameterPassing(t *testing.T) {
	t.Run("functions accept custom parameters", func(t *testing.T) {
		// Test with custom error message
		schema := Bool("custom error message")
		require.NotNil(t, schema)

		// Test with multiple parameters
		schema2 := String("param1", "param2")
		require.NotNil(t, schema2)

		// Test with no parameters
		schema3 := Int()
		require.NotNil(t, schema3)
	})
}

// =============================================================================
// Modifier chaining tests
// =============================================================================

func TestCoercion_ModifierChaining(t *testing.T) {
	t.Run("coerced schemas support modifier chaining", func(t *testing.T) {
		// Test Bool with Optional and Default - Default should short-circuit
		schema1 := Bool().Optional().Default(true)
		result, err := schema1.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result) // Default short-circuits, returns default value

		// Test order independence - Default should short-circuit regardless of order
		schema2 := Bool().Default(true).Optional()
		result, err = schema2.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result) // Default short-circuits regardless of order

		// Test with actual value - both should work the same
		result, err = schema1.Parse("true")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		// Test String with modifiers
		stringSchema := String().Nilable()
		result2, err := stringSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)

		// Test Integer with modifiers
		// For Optional, nil input should return nil, not default
		intSchema := Integer().Optional()
		result3, err := intSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result3)

		// Test with actual value
		result3, err = intSchema.Parse("123")
		require.NoError(t, err)
		require.NotNil(t, result3)
		assert.Equal(t, int64(123), *result3)
	})

	t.Run("pointer schemas support modifier chaining", func(t *testing.T) {
		// Test BoolPtr with modifiers
		schema := BoolPtr().Nilable().Default(true)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result) // Default takes precedence and short-circuits

		// Test StringPtr with modifiers
		stringSchema := StringPtr().Optional()
		result2, err := stringSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestCoercion_TypeSafety(t *testing.T) {
	t.Run("schemas return correct types", func(t *testing.T) {
		// Test that each schema returns the expected type
		boolResult, _ := Bool().Parse(true)
		assert.IsType(t, bool(false), boolResult)

		boolPtrResult, _ := BoolPtr().Parse(true)
		assert.IsType(t, (*bool)(nil), boolPtrResult)

		stringResult, _ := String().Parse("test")
		assert.IsType(t, "", stringResult)

		stringPtrResult, _ := StringPtr().Parse("test")
		assert.IsType(t, (*string)(nil), stringPtrResult)

		intResult, _ := Int().Parse(123)
		assert.IsType(t, int(0), intResult)

		int8Result, _ := Int8().Parse(123)
		assert.IsType(t, int8(0), int8Result)

		uint64Result, _ := Uint64().Parse(123)
		assert.IsType(t, uint64(0), uint64Result)

		float32Result, _ := Float32().Parse(3.14)
		assert.IsType(t, float32(0), float32Result)

		complexResult, _ := Complex().Parse(3 + 4i)
		assert.IsType(t, complex128(0), complexResult)

		bigIntResult, _ := BigInt().Parse(123)
		assert.IsType(t, (*big.Int)(nil), bigIntResult)

		timeResult, _ := Time().Parse("2023-01-15T10:30:00Z")
		assert.IsType(t, time.Time{}, timeResult)

		stringBoolResult, _ := StringBool().Parse("true")
		assert.IsType(t, bool(false), stringBoolResult)
	})
}

// =============================================================================
// Edge cases and error handling
// =============================================================================

func TestCoercion_EdgeCases(t *testing.T) {
	t.Run("nil input handling", func(t *testing.T) {
		// Non-nilable schemas should reject nil
		schemas := []any{
			Bool(), String(), Int(), Float(), Complex(), BigInt(), Time(), StringBool(),
		}

		for _, schema := range schemas {
			if s, ok := schema.(interface{ Parse(any) (any, error) }); ok {
				_, err := s.Parse(nil)
				assert.Error(t, err, "Schema should reject nil input")
			}
		}
	})

	t.Run("empty string handling", func(t *testing.T) {
		// Test how different schemas handle empty strings
		// Empty strings should have sensible default coercion behavior

		// Bool should convert empty string to false
		boolSchema := Bool()
		result, err := boolSchema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Integer should convert empty string to 0
		intSchema := Integer()
		result2, err := intSchema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result2)

		// String should preserve empty string
		stringSchema := String()
		result3, err := stringSchema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result3)

		// Test clearly invalid inputs still return errors
		_, err = boolSchema.Parse("invalid_bool")
		assert.Error(t, err)

		_, err = intSchema.Parse("not_a_number")
		assert.Error(t, err)
	})

	t.Run("overflow handling", func(t *testing.T) {
		// Test integer overflow
		int8Schema := Int8()
		_, err := int8Schema.Parse(128) // Overflow for int8
		assert.Error(t, err)

		uint8Schema := Uint8()
		_, err = uint8Schema.Parse(256) // Overflow for uint8
		assert.Error(t, err)
	})

	t.Run("precision handling", func(t *testing.T) {
		// Test float precision
		float32Schema := Float32()
		result, err := float32Schema.Parse(3.14159265359)
		require.NoError(t, err)
		// float32 has limited precision
		assert.InDelta(t, float32(3.14159265359), result, 0.00001)
	})
}

// =============================================================================
// MustParse tests
// =============================================================================

func TestCoercion_MustParse(t *testing.T) {
	t.Run("MustParse successful cases", func(t *testing.T) {
		// Test Bool MustParse
		result := Bool().MustParse(true)
		assert.Equal(t, true, result)

		// Test String MustParse with coercion
		stringResult := String().MustParse(123)
		assert.Equal(t, "123", stringResult)

		// Test Integer MustParse with coercion
		intResult := Integer().MustParse("456")
		assert.Equal(t, int64(456), intResult)
	})

	t.Run("MustParse panic cases", func(t *testing.T) {
		// Test Bool MustParse panic
		assert.Panics(t, func() {
			Bool().MustParse("invalid")
		})

		// Test Integer MustParse panic
		assert.Panics(t, func() {
			Integer().MustParse("not a number")
		})

		// Test Uint MustParse panic with negative
		assert.Panics(t, func() {
			Uint().MustParse(-123)
		})
	})
}
