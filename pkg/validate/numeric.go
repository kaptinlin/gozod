package validate

import (
	"math"
	"reflect"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Lt reports whether value is less than limit.
func Lt(value, limit any) bool {
	if order, ok := compareIntegers(value, limit); ok {
		return order < 0
	}
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) < toFloat64(limit)
}

// Lte reports whether value is less than or equal to limit.
func Lte(value, limit any) bool {
	if order, ok := compareIntegers(value, limit); ok {
		return order <= 0
	}
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) <= toFloat64(limit)
}

// Gt reports whether value is greater than limit.
func Gt(value, limit any) bool {
	if order, ok := compareIntegers(value, limit); ok {
		return order > 0
	}
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) > toFloat64(limit)
}

// Gte reports whether value is greater than or equal to limit.
func Gte(value, limit any) bool {
	if order, ok := compareIntegers(value, limit); ok {
		return order >= 0
	}
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) >= toFloat64(limit)
}

type integerMagnitude struct {
	negative  bool
	magnitude uint64
}

func compareIntegers(a, b any) (int, bool) {
	left, ok := normalizeInteger(a)
	if !ok {
		return 0, false
	}
	right, ok := normalizeInteger(b)
	if !ok {
		return 0, false
	}

	if left.negative != right.negative {
		if left.negative {
			return -1, true
		}
		return 1, true
	}
	if left.magnitude == right.magnitude {
		return 0, true
	}
	if left.magnitude < right.magnitude {
		if left.negative {
			return 1, true
		}
		return -1, true
	}
	if left.negative {
		return -1, true
	}
	return 1, true
}

func normalizeInteger(value any) (integerMagnitude, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return integerMagnitude{}, false
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		signed := rv.Int()
		if signed < 0 {
			magnitude := uint64(-(signed + 1)) //nolint:gosec // Negation makes the checked value non-negative; +1 below handles MinInt64.
			return integerMagnitude{negative: true, magnitude: magnitude + 1}, true
		}
		return integerMagnitude{magnitude: uint64(signed)}, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return integerMagnitude{magnitude: rv.Uint()}, true
	default:
		return integerMagnitude{}, false
	}
}

// Positive reports whether the numeric value is positive (> 0).
func Positive(value any) bool { return Gt(value, 0) }

// Negative reports whether the numeric value is negative (< 0).
func Negative(value any) bool { return Lt(value, 0) }

// NonPositive reports whether the numeric value is non-positive (<= 0).
func NonPositive(value any) bool { return Lte(value, 0) }

// NonNegative reports whether the numeric value is non-negative (>= 0).
func NonNegative(value any) bool { return Gte(value, 0) }

// MultipleOf reports whether value is a multiple of the given divisor.
func MultipleOf(value, divisor any) bool {
	integerValue, valueIsInteger := normalizeInteger(value)
	integerDivisor, divisorIsInteger := normalizeInteger(divisor)
	if valueIsInteger && divisorIsInteger {
		return integerDivisor.magnitude != 0 && integerValue.magnitude%integerDivisor.magnitude == 0
	}
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(divisor) {
		return false
	}
	val := toFloat64(value)
	div := toFloat64(divisor)
	if div == 0 {
		return false
	}
	epsilon := max(1e-10, math.Abs(div)*1e-6)
	remainder := math.Abs(math.Mod(val, div))
	return remainder < epsilon || math.Abs(remainder-math.Abs(div)) < epsilon
}
