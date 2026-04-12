package validate

import (
	"math"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Lt reports whether value is less than limit.
func Lt(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) < toFloat64(limit)
}

// Lte reports whether value is less than or equal to limit.
func Lte(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) <= toFloat64(limit)
}

// Gt reports whether value is greater than limit.
func Gt(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) > toFloat64(limit)
}

// Gte reports whether value is greater than or equal to limit.
func Gte(value, limit any) bool {
	if !reflectx.IsNumeric(value) || !reflectx.IsNumeric(limit) {
		return false
	}
	return toFloat64(value) >= toFloat64(limit)
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
