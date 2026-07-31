package validate_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

type namedInt64 int64

func TestNumericComparisons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "less than", got: validate.Lt(2, 3), want: true},
		{name: "less than rejects equal", got: validate.Lt(3, 3), want: false},
		{name: "less than or equal accepts equal", got: validate.Lte(3, 3), want: true},
		{name: "greater than", got: validate.Gt(4, 3), want: true},
		{name: "greater than rejects equal", got: validate.Gt(3, 3), want: false},
		{name: "greater than or equal accepts equal", got: validate.Gte(3, 3), want: true},
		{name: "non numeric value", got: validate.Gt("4", 3), want: false},
		{name: "non numeric limit", got: validate.Lt(2, "3"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestNumericComparisons_PreserveNamedIntegerPrecision(t *testing.T) {
	value := namedInt64(math.MaxInt64 - 1)
	assert.True(t, validate.Lt(value, int64(math.MaxInt64)))
	assert.False(t, validate.Gte(value, int64(math.MaxInt64)))
}

func TestNumericComparisons_PreserveEveryIntegerKind(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	maxUint := ^uint(0)
	maxUintptr := ^uintptr(0)
	tests := []struct {
		name  string
		lower any
		upper any
	}{
		{name: "int", lower: maxInt - 1, upper: maxInt},
		{name: "int8", lower: int8(math.MaxInt8 - 1), upper: int8(math.MaxInt8)},
		{name: "int16", lower: int16(math.MaxInt16 - 1), upper: int16(math.MaxInt16)},
		{name: "int32", lower: int32(math.MaxInt32 - 1), upper: int32(math.MaxInt32)},
		{name: "int64", lower: int64(math.MaxInt64 - 1), upper: int64(math.MaxInt64)},
		{name: "uint", lower: maxUint - 1, upper: maxUint},
		{name: "uint8", lower: uint8(math.MaxUint8 - 1), upper: uint8(math.MaxUint8)},
		{name: "uint16", lower: uint16(math.MaxUint16 - 1), upper: uint16(math.MaxUint16)},
		{name: "uint32", lower: uint32(math.MaxUint32 - 1), upper: uint32(math.MaxUint32)},
		{name: "uint64", lower: uint64(math.MaxUint64 - 1), upper: uint64(math.MaxUint64)},
		{name: "uintptr", lower: maxUintptr - 1, upper: maxUintptr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.True(t, validate.Lt(test.lower, test.upper))
			assert.True(t, validate.Lte(test.upper, test.upper))
			assert.True(t, validate.Gt(test.upper, test.lower))
			assert.True(t, validate.Gte(test.lower, test.lower))
		})
	}
}

func TestNumericComparisons_PreserveMixedIntegerDomains(t *testing.T) {
	assert.True(t, validate.Lt(int64(math.MinInt64), int64(math.MinInt64+1)))
	assert.True(t, validate.Lt(int64(-1), uint64(0)))
	assert.True(t, validate.Gt(uint64(0), int64(-1)))
	assert.True(t, validate.Lt(int64(math.MaxInt64), uint64(math.MaxInt64)+1))
	assert.True(t, validate.Gt(uint64(math.MaxUint64), int64(math.MaxInt64)))
	assert.True(t, validate.Lte(int64(42), uint64(42)))
	assert.True(t, validate.Gte(uint64(42), int64(42)))
}

func TestNumericComparisons_PreserveFloatBehavior(t *testing.T) {
	assert.True(t, validate.Lt(0.1, 0.2))
	assert.True(t, validate.Gte(float32(1.5), float64(1.5)))
}

func TestNumericSignPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "positive", got: validate.Positive(1), want: true},
		{name: "positive rejects zero", got: validate.Positive(0), want: false},
		{name: "negative", got: validate.Negative(-1.5), want: true},
		{name: "negative rejects zero", got: validate.Negative(0), want: false},
		{name: "non positive accepts zero", got: validate.NonPositive(0), want: true},
		{name: "non positive rejects positive", got: validate.NonPositive(1), want: false},
		{name: "non negative accepts zero", got: validate.NonNegative(0), want: true},
		{name: "non negative rejects negative", got: validate.NonNegative(-1), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestMultipleOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		divisor any
		want    bool
	}{
		{name: "integer multiple", value: 10, divisor: 5, want: true},
		{name: "integer non multiple", value: 7, divisor: 3, want: false},
		{name: "negative multiple", value: -12, divisor: 4, want: true},
		{name: "zero divisor", value: 10, divisor: 0, want: false},
		{name: "non numeric value", value: "10", divisor: 5, want: false},
		{name: "non numeric divisor", value: 10, divisor: "5", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.MultipleOf(tt.value, tt.divisor))
		})
	}
}

func TestMultipleOf_PreservesLargeIntegerPrecision(t *testing.T) {
	assert.False(t, validate.MultipleOf(int64(math.MaxInt64-1), int64(math.MaxInt64)))
	assert.True(t, validate.MultipleOf(uint64(math.MaxUint64), uint8(3)))
	assert.False(t, validate.MultipleOf(uint64(math.MaxUint64-1), uint8(3)))
	assert.True(t, validate.MultipleOf(namedInt64(-12), uint8(4)))
	assert.False(t, validate.MultipleOf(uint64(math.MaxUint64), int64(0)))
}

func TestMultipleOf_PreservesFloatTolerance(t *testing.T) {
	assert.True(t, validate.MultipleOf(0.3, 0.1))
}
