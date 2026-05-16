package coerce_test

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/kaptinlin/gozod/pkg/coerce"
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

		result, err = ToInt64(uint64(math.MaxInt64))
		require.NoError(t, err)
		assert.Equal(t, int64(math.MaxInt64), result)

		_, err = ToInt64(uint64(math.MaxInt64 + 1))
		assert.ErrorIs(t, err, ErrOverflow)
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

	t.Run("uint64 overflows signed conversion", func(t *testing.T) {
		_, err := ToInteger[int64](uint64(math.MaxInt64 + 1))
		assert.ErrorIs(t, err, ErrOverflow)
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
		{"implicit positive imaginary j", "+j", complex(0, 1), false},
		{"implicit negative imaginary j", "-j", complex(0, -1), false},
		{"complex positive", "3+4i", complex(3, 4), false},
		{"complex positive j", "3+4j", complex(3, 4), false},
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

func TestConversionErrorsWrapSentinels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "unsupported", err: NewUnsupportedError("struct {}", "bool"), want: ErrUnsupported},
		{name: "format", err: NewFormatError("abc", "int64"), want: ErrInvalidFmt},
		{name: "overflow", err: NewOverflowError(256, "uint8"), want: ErrOverflow},
		{name: "empty input", err: NewEmptyInputError("complex"), want: ErrEmptyInput},
		{name: "negative", err: NewNegativeError(-1, "uint"), want: ErrNegative},
		{name: "not whole", err: NewNotWholeError(1.5), want: ErrNotWhole},
		{name: "nil pointer", err: NewNilPointerError("string"), want: ErrNilPointer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorIs(t, tt.err, tt.want)
		})
	}
}

func TestToBool_UnsupportedType(t *testing.T) {
	t.Parallel()

	_, err := ToBool(struct{}{})
	require.ErrorIs(t, err, ErrUnsupported)
}

func TestToString_UnsupportedType(t *testing.T) {
	t.Parallel()

	_, err := ToString(struct{}{})
	require.ErrorIs(t, err, ErrUnsupported)
}

func TestToTime_AdditionalInputs(t *testing.T) {
	t.Parallel()

	t.Run("unix timestamp float32", func(t *testing.T) {
		t.Parallel()

		result, err := ToTime(float32(1705312200))
		require.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ToTime(struct{}{})
		require.ErrorIs(t, err, ErrUnsupported)
	})
}

func TestToFloat64_AdditionalInputs(t *testing.T) {
	t.Parallel()

	t.Run("float32 NaN", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat64(float32(math.NaN()))
		require.ErrorIs(t, err, ErrInvalidFmt)
	})

	t.Run("big int", func(t *testing.T) {
		t.Parallel()

		result, err := ToFloat64(big.NewInt(42))
		require.NoError(t, err)
		assert.Equal(t, 42.0, result)
	})

	t.Run("complex magnitude", func(t *testing.T) {
		t.Parallel()

		result, err := ToFloat64(complex(3, 4))
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat64(struct{}{})
		require.ErrorIs(t, err, ErrUnsupported)
	})
}

func TestToBigInt_AdditionalInputs(t *testing.T) {
	t.Parallel()

	t.Run("copies input big int", func(t *testing.T) {
		t.Parallel()

		input := big.NewInt(42)
		result, err := ToBigInt(input)
		require.NoError(t, err)
		input.SetInt64(7)
		assert.Equal(t, int64(42), result.Int64())
	})

	t.Run("empty string becomes zero", func(t *testing.T) {
		t.Parallel()

		result, err := ToBigInt("   ")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Int64())
	})

	t.Run("invalid string returns format error", func(t *testing.T) {
		t.Parallel()

		_, err := ToBigInt("not-a-number")
		require.ErrorIs(t, err, ErrInvalidFmt)
	})

	t.Run("nil big int pointer", func(t *testing.T) {
		t.Parallel()

		var input *big.Int
		_, err := ToBigInt(input)
		require.ErrorIs(t, err, ErrNilPointer)
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ToBigInt(struct{}{})
		require.ErrorIs(t, err, ErrUnsupported)
	})
}

func TestToInteger_ErrorContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T) error
		want error
	}{
		{name: "non whole float32 truncates", run: func(t *testing.T) error {
			got, err := ToInteger[int](float32(3.9))
			assert.Equal(t, 3, got)
			return err
		}, want: nil},
		{name: "non whole float64", run: func(t *testing.T) error {
			_, err := ToInteger[int](3.9)
			return err
		}, want: ErrNotWhole},
		{name: "uint16 overflow", run: func(t *testing.T) error {
			_, err := ToInteger[uint16](70000)
			return err
		}, want: ErrOverflow},
		{name: "unsupported type", run: func(t *testing.T) error {
			_, err := ToInteger[int](struct{}{})
			return err
		}, want: ErrUnsupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.run(t)
			if tt.want == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestToFloat_ErrorContracts(t *testing.T) {
	t.Parallel()

	t.Run("float32 overflow", func(t *testing.T) {
		t.Parallel()

		_, err := ToFloat[float32](math.MaxFloat64)
		require.ErrorIs(t, err, ErrOverflow)
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var input *float64
		_, err := ToFloat[float64](input)
		require.ErrorIs(t, err, ErrNilPointer)
	})
}

func TestToComplex_ErrorContracts(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var input *complex128
		_, err := ToComplex128(input)
		require.ErrorIs(t, err, ErrNilPointer)
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ToComplex128(struct{}{})
		require.ErrorIs(t, err, ErrUnsupported)
	})
}

func TestToComplexFromString_ErrorContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "invalid real part", input: "x+1i"},
		{name: "missing imaginary suffix", input: "1+2"},
		{name: "invalid imaginary part", input: "1+xi"},
		{name: "malformed value", input: "not-complex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ToComplexFromString(tt.input)
			require.ErrorIs(t, err, ErrInvalidFmt)
		})
	}
}

func TestTo_AdditionalTargets(t *testing.T) {
	t.Parallel()

	t.Run("to time", func(t *testing.T) {
		t.Parallel()

		result, err := To[time.Time]("2024-01-15")
		require.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
	})

	t.Run("to big int", func(t *testing.T) {
		t.Parallel()

		result, err := To[*big.Int]("42")
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Int64())
	})

	t.Run("to complex128", func(t *testing.T) {
		t.Parallel()

		result, err := To[complex128]("3+4i")
		require.NoError(t, err)
		assert.Equal(t, complex(3, 4), result)
	})

	t.Run("fallback conversion", func(t *testing.T) {
		t.Parallel()

		result, err := To[int32](int16(7))
		require.NoError(t, err)
		assert.Equal(t, int32(7), result)
	})
}

func TestToLiteral_NumericContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  any
	}{
		{name: "string one becomes true", input: "1", want: true},
		{name: "string zero becomes false", input: "0", want: false},
		{name: "string integer becomes int", input: "42", want: 42},
		{name: "large integer within native range becomes int", input: "9223372036854775807", want: int(9223372036854775807)},
		{name: "decimal becomes float64", input: "3.14", want: 3.14},
		{name: "whole float string becomes int", input: "2.0", want: 2},
		{name: "non numeric string stays string", input: "hello", want: "hello"},
		{name: "float32 zero becomes false", input: float32(0), want: false},
		{name: "float64 one becomes true", input: float64(1), want: true},
		{name: "int64 two stays int64", input: int64(2), want: int64(2)},
		{name: "uint64 two stays uint64", input: uint64(2), want: uint64(2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToLiteral(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestToString_AdditionalInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{name: "int8", input: int8(8), want: "8"},
		{name: "uint16", input: uint16(16), want: "16"},
		{name: "float32", input: float32(1.5), want: "1.5"},
		{name: "complex", input: complex(3, 4), want: "(3+4i)"},
		{name: "big int value", input: *big.NewInt(99), want: "99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToString(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToInt64_ErrorContracts(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var input *int64
		_, err := ToInt64(input)
		require.ErrorIs(t, err, ErrNilPointer)
	})

	t.Run("float32 not whole", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt64(float32(1.25))
		require.ErrorIs(t, err, ErrNotWhole)
	})

	t.Run("float64 overflow", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt64(math.MaxFloat64)
		require.ErrorIs(t, err, ErrOverflow)
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ToInt64(struct{}{})
		require.ErrorIs(t, err, ErrUnsupported)
	})
}

func TestToInteger_BoundsByTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T) error
		want error
	}{
		{name: "int16 overflow", run: func(t *testing.T) error {
			_, err := ToInteger[int16](40000)
			return err
		}, want: ErrOverflow},
		{name: "int32 overflow", run: func(t *testing.T) error {
			_, err := ToInteger[int32](float64(math.MaxInt64))
			return err
		}, want: ErrOverflow},
		{name: "uint negative", run: func(t *testing.T) error {
			_, err := ToInteger[uint](-1)
			return err
		}, want: ErrNegative},
		{name: "uint32 overflow", run: func(t *testing.T) error {
			_, err := ToInteger[uint32](uint64(math.MaxUint32) + 1)
			return err
		}, want: ErrOverflow},
		{name: "uint64 negative", run: func(t *testing.T) error {
			_, err := ToInteger[uint64](-1)
			return err
		}, want: ErrNegative},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorIs(t, tt.run(t), tt.want)
		})
	}
}

func TestToFloat_MoreTargets(t *testing.T) {
	t.Parallel()

	t.Run("already target float32", func(t *testing.T) {
		t.Parallel()

		got, err := ToFloat[float32](float32(1.25))
		require.NoError(t, err)
		assert.Equal(t, float32(1.25), got)
	})

	t.Run("bool true", func(t *testing.T) {
		t.Parallel()

		got, err := ToFloat[float64](true)
		require.NoError(t, err)
		assert.Equal(t, 1.0, got)
	})
}

func TestTo_IntegralAndFloatTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{name: "int8", run: func(t *testing.T) {
			got, err := To[int8]("8")
			require.NoError(t, err)
			assert.Equal(t, int8(8), got)
		}},
		{name: "int16", run: func(t *testing.T) {
			got, err := To[int16]("16")
			require.NoError(t, err)
			assert.Equal(t, int16(16), got)
		}},
		{name: "int32", run: func(t *testing.T) {
			got, err := To[int32]("32")
			require.NoError(t, err)
			assert.Equal(t, int32(32), got)
		}},
		{name: "int64", run: func(t *testing.T) {
			got, err := To[int64]("64")
			require.NoError(t, err)
			assert.Equal(t, int64(64), got)
		}},
		{name: "uint", run: func(t *testing.T) {
			got, err := To[uint]("7")
			require.NoError(t, err)
			assert.Equal(t, uint(7), got)
		}},
		{name: "uint8", run: func(t *testing.T) {
			got, err := To[uint8]("8")
			require.NoError(t, err)
			assert.Equal(t, uint8(8), got)
		}},
		{name: "uint16", run: func(t *testing.T) {
			got, err := To[uint16]("16")
			require.NoError(t, err)
			assert.Equal(t, uint16(16), got)
		}},
		{name: "uint32", run: func(t *testing.T) {
			got, err := To[uint32]("32")
			require.NoError(t, err)
			assert.Equal(t, uint32(32), got)
		}},
		{name: "uint64", run: func(t *testing.T) {
			got, err := To[uint64]("64")
			require.NoError(t, err)
			assert.Equal(t, uint64(64), got)
		}},
		{name: "float32", run: func(t *testing.T) {
			got, err := To[float32]("1.5")
			require.NoError(t, err)
			assert.Equal(t, float32(1.5), got)
		}},
		{name: "complex64", run: func(t *testing.T) {
			got, err := To[complex64]("3+4i")
			require.NoError(t, err)
			assert.Equal(t, complex64(3+4i), got)
		}},
		{name: "unsupported nil target", run: func(t *testing.T) {
			_, err := To[any](42)
			require.ErrorIs(t, err, ErrUnsupported)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.run(t)
		})
	}
}

func TestToString_ErrorContracts(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var input *string
		_, err := ToString(input)
		require.ErrorIs(t, err, ErrNilPointer)
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ToString(struct{}{})
		require.ErrorIs(t, err, ErrUnsupported)
	})
}

func TestToTime_NumericInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
	}{
		{name: "int", input: 1705312200},
		{name: "float64", input: float64(1705312200)},
		{name: "float32", input: float32(1705312200)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToTime(tt.input)
			require.NoError(t, err)
			assert.Equal(t, 2024, got.Year())
		})
	}
}

func TestToFloat64_PrimitiveInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  float64
	}{
		{name: "uint", input: uint(42), want: 42},
		{name: "bool true", input: true, want: 1},
		{name: "bool false", input: false, want: 0},
		{name: "big int value", input: *big.NewInt(9), want: 9},
		{name: "complex64 magnitude", input: complex64(3 + 4i), want: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToFloat64(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToBigInt_PrimitiveInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  int64
	}{
		{name: "uint", input: uint(42), want: 42},
		{name: "float32 whole", input: float32(12), want: 12},
		{name: "float64 whole", input: float64(13), want: 13},
		{name: "bool false", input: false, want: 0},
		{name: "big int value", input: *big.NewInt(14), want: 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToBigInt(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Int64())
		})
	}
}

func TestToInteger_MoreInputs(t *testing.T) {
	t.Parallel()

	t.Run("uint64 within range", func(t *testing.T) {
		t.Parallel()

		got, err := ToInteger[int](uint64(42))
		require.NoError(t, err)
		assert.Equal(t, 42, got)
	})

	t.Run("invalid string", func(t *testing.T) {
		t.Parallel()

		_, err := ToInteger[int]("not-int")
		require.ErrorIs(t, err, ErrInvalidFmt)
	})

	t.Run("bool false", func(t *testing.T) {
		t.Parallel()

		got, err := ToInteger[int](false)
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var input *int
		_, err := ToInteger[int](input)
		require.ErrorIs(t, err, ErrNilPointer)
	})
}

type coerceAlias string

func TestTo_FallbackAndErrorPropagation(t *testing.T) {
	t.Parallel()

	t.Run("fallback target", func(t *testing.T) {
		t.Parallel()

		got, err := To[coerceAlias]("value")
		require.NoError(t, err)
		assert.Equal(t, coerceAlias("value"), got)
	})

	t.Run("bool target propagates format error", func(t *testing.T) {
		t.Parallel()

		_, err := To[bool]("maybe")
		require.ErrorIs(t, err, ErrInvalidFmt)
	})

	t.Run("int target propagates overflow error", func(t *testing.T) {
		t.Parallel()

		_, err := To[int8](200)
		require.ErrorIs(t, err, ErrOverflow)
	})

	t.Run("complex target propagates format error", func(t *testing.T) {
		t.Parallel()

		_, err := To[complex64]("not-complex")
		require.ErrorIs(t, err, ErrInvalidFmt)
	})
}
