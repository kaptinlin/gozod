package coerce

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToBool(t *testing.T) {
	t.Run("direct bool values", func(t *testing.T) {
		result, err := ToBool(true)
		require.NoError(t, err)
		assert.True(t, result)

		result, err = ToBool(false)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("string true values", func(t *testing.T) {
		trueStrings := []string{"true", "1", "yes", "on"}
		for _, s := range trueStrings {
			result, err := ToBool(s)
			require.NoError(t, err, "input: %s", s)
			assert.True(t, result, "input: %s", s)
		}
	})

	t.Run("string false values", func(t *testing.T) {
		falseStrings := []string{"false", "0", "no", "off", ""}
		for _, s := range falseStrings {
			result, err := ToBool(s)
			require.NoError(t, err, "input: %s", s)
			assert.False(t, result, "input: %s", s)
		}
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		_, err := ToBool("invalid")
		assert.Error(t, err)
	})

	t.Run("integer values", func(t *testing.T) {
		result, err := ToBool(1)
		require.NoError(t, err)
		assert.True(t, result)

		result, err = ToBool(0)
		require.NoError(t, err)
		assert.False(t, result)

		result, err = ToBool(-1)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("float values", func(t *testing.T) {
		result, err := ToBool(1.0)
		require.NoError(t, err)
		assert.True(t, result)

		result, err = ToBool(0.0)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("nil pointer returns error", func(t *testing.T) {
		var p *bool
		_, err := ToBool(p)
		assert.Error(t, err)
	})

	t.Run("valid pointer dereferences", func(t *testing.T) {
		val := true
		result, err := ToBool(&val)
		require.NoError(t, err)
		assert.True(t, result)
	})
}

func TestToString(t *testing.T) {
	t.Run("string passthrough", func(t *testing.T) {
		result, err := ToString("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("integer types", func(t *testing.T) {
		result, err := ToString(42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)

		result, err = ToString(int64(64))
		require.NoError(t, err)
		assert.Equal(t, "64", result)
	})

	t.Run("float types", func(t *testing.T) {
		result, err := ToString(3.14)
		require.NoError(t, err)
		assert.Equal(t, "3.14", result)
	})

	t.Run("bool types", func(t *testing.T) {
		result, err := ToString(true)
		require.NoError(t, err)
		assert.Equal(t, "true", result)

		result, err = ToString(false)
		require.NoError(t, err)
		assert.Equal(t, "false", result)
	})

	t.Run("[]byte", func(t *testing.T) {
		result, err := ToString([]byte("hello"))
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("big.Int", func(t *testing.T) {
		val := big.NewInt(123456789)
		result, err := ToString(val)
		require.NoError(t, err)
		assert.Equal(t, "123456789", result)
	})

	t.Run("nil big.Int pointer", func(t *testing.T) {
		var val *big.Int
		_, err := ToString(val)
		assert.Error(t, err)
	})

	t.Run("time.Time", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result, err := ToString(tm)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15T10:30:00Z", result)
	})
}

func TestToTime(t *testing.T) {
	t.Run("time.Time passthrough", func(t *testing.T) {
		input := time.Now()
		result, err := ToTime(input)
		require.NoError(t, err)
		assert.True(t, result.Equal(input))
	})

	t.Run("RFC3339 string", func(t *testing.T) {
		result, err := ToTime("2024-01-15T10:30:00Z")
		require.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
	})

	t.Run("date only string", func(t *testing.T) {
		result, err := ToTime("2024-01-15")
		require.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
	})

	t.Run("unix timestamp int64", func(t *testing.T) {
		result, err := ToTime(int64(1705312200))
		require.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		_, err := ToTime("not a time")
		assert.Error(t, err)
	})
}

func TestToInt64(t *testing.T) {
	t.Run("direct int64", func(t *testing.T) {
		result, err := ToInt64(int64(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	t.Run("other integer types", func(t *testing.T) {
		result, err := ToInt64(42)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		result, err = ToInt64(int8(8))
		require.NoError(t, err)
		assert.Equal(t, int64(8), result)
	})

	t.Run("unsigned integers", func(t *testing.T) {
		result, err := ToInt64(uint(42))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		// uint64 overflow
		_, err = ToInt64(uint64(math.MaxInt64 + 1))
		assert.Error(t, err)
	})

	t.Run("whole number floats", func(t *testing.T) {
		result, err := ToInt64(42.0)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		// Non-whole number should error
		_, err = ToInt64(42.5)
		assert.Error(t, err)
	})

	t.Run("string numbers", func(t *testing.T) {
		result, err := ToInt64("42")
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)

		result, err = ToInt64("")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)

		_, err = ToInt64("abc")
		assert.Error(t, err)
	})

	t.Run("bool", func(t *testing.T) {
		result, err := ToInt64(true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result)

		result, err = ToInt64(false)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})
}

func TestToFloat64(t *testing.T) {
	t.Run("direct float64", func(t *testing.T) {
		result, err := ToFloat64(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("integer types", func(t *testing.T) {
		result, err := ToFloat64(42)
		require.NoError(t, err)
		assert.Equal(t, 42.0, result)
	})

	t.Run("string numbers", func(t *testing.T) {
		result, err := ToFloat64("3.14")
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)

		result, err = ToFloat64("")
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)

		_, err = ToFloat64("abc")
		assert.Error(t, err)
	})

	t.Run("NaN returns error", func(t *testing.T) {
		_, err := ToFloat64(math.NaN())
		assert.Error(t, err)
	})
}

func TestToBigInt(t *testing.T) {
	t.Run("from int", func(t *testing.T) {
		result, err := ToBigInt(42)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Int64())
	})

	t.Run("from string", func(t *testing.T) {
		result, err := ToBigInt("12345678901234567890")
		require.NoError(t, err)
		assert.Equal(t, "12345678901234567890", result.String())
	})

	t.Run("from hex string", func(t *testing.T) {
		result, err := ToBigInt("0xFF")
		require.NoError(t, err)
		assert.Equal(t, int64(255), result.Int64())
	})

	t.Run("from bool", func(t *testing.T) {
		result, err := ToBigInt(true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Int64())
	})

	t.Run("from non-whole float returns error", func(t *testing.T) {
		_, err := ToBigInt(3.14)
		assert.Error(t, err)
	})
}

func TestToInteger(t *testing.T) {
	t.Run("to int", func(t *testing.T) {
		result, err := ToInteger[int](42)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("to int8", func(t *testing.T) {
		result, err := ToInteger[int8](42)
		require.NoError(t, err)
		assert.Equal(t, int8(42), result)
	})

	t.Run("int8 overflow", func(t *testing.T) {
		_, err := ToInteger[int8](200)
		assert.Error(t, err)
	})

	t.Run("uint8 negative", func(t *testing.T) {
		_, err := ToInteger[uint8](-1)
		assert.Error(t, err)
	})
}

func TestToFloat(t *testing.T) {
	t.Run("to float32", func(t *testing.T) {
		result, err := ToFloat[float32](3.14)
		require.NoError(t, err)
		assert.InDelta(t, float32(3.14), result, 0.01)
	})

	t.Run("to float64", func(t *testing.T) {
		result, err := ToFloat[float64](3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})
}

func TestToComplex64(t *testing.T) {
	t.Run("from complex64", func(t *testing.T) {
		input := complex64(3 + 4i)
		result, err := ToComplex64(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("from complex128", func(t *testing.T) {
		input := complex128(3 + 4i)
		result, err := ToComplex64(input)
		require.NoError(t, err)
		assert.Equal(t, complex64(3+4i), result)
	})

	t.Run("from int", func(t *testing.T) {
		result, err := ToComplex64(5)
		require.NoError(t, err)
		assert.Equal(t, complex64(5+0i), result)
	})
}

func TestToComplex128(t *testing.T) {
	t.Run("from complex128", func(t *testing.T) {
		input := complex128(3 + 4i)
		result, err := ToComplex128(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("from float64", func(t *testing.T) {
		result, err := ToComplex128(3.14)
		require.NoError(t, err)
		assert.Equal(t, complex(3.14, 0), result)
	})
}

func TestToComplexFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    complex128
		wantErr bool
	}{
		{"real only", "3.14", complex(3.14, 0), false},
		{"imaginary only i", "2i", complex(0, 2), false},
		{"imaginary only j", "2j", complex(0, 2), false},
		{"complex positive", "3+4i", complex(3, 4), false},
		{"complex negative", "3-4i", complex(3, -4), false},
		{"just i", "i", complex(0, 1), false},
		{"just +i", "+i", complex(0, 1), false},
		{"just -i", "-i", complex(0, -1), false},
		{"empty", "", complex(0, 0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToComplexFromString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestTo(t *testing.T) {
	t.Run("to string", func(t *testing.T) {
		result, err := To[string](42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})

	t.Run("to bool", func(t *testing.T) {
		result, err := To[bool]("true")
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("to int", func(t *testing.T) {
		result, err := To[int]("42")
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("to float64", func(t *testing.T) {
		result, err := To[float64]("3.14")
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})
}

func TestToLiteral(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		result, err := ToLiteral(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("bool string converts", func(t *testing.T) {
		result, err := ToLiteral("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = ToLiteral("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("empty string stays empty", func(t *testing.T) {
		result, err := ToLiteral("")
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}
