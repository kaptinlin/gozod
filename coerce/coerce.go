package coerce

import (
	"math/big"
	"time"

	"github.com/kaptinlin/gozod/types"
)

// Package coerce provides constructors that enable automatic type coercion
// before validation. For example, strings like "true" or numbers like 1/0
// will be coerced into boolean values when using Bool().
//
// Example:
//  s := coerce.Bool()
//  v, err := s.Parse("true") // v == true, err == nil
//
// This mirrors the JavaScript Zod API: `z.coerce.boolean()`.
// Currently only boolean coercion is supported; additional helpers will
// be added as the library evolves.

// Bool returns a boolean schema with coercion enabled.
func Bool(params ...any) *types.ZodBool[bool] {
	return types.CoercedBool(params...)
}

// BoolPtr returns a *bool schema with coercion enabled.
func BoolPtr(params ...any) *types.ZodBool[*bool] {
	return types.CoercedBoolPtr(params...)
}

// String returns a string schema with coercion enabled.
func String(params ...any) *types.ZodString[string] {
	return types.CoercedString(params...)
}

// StringPtr returns a *string schema with coercion enabled.
func StringPtr(params ...any) *types.ZodString[*string] {
	return types.CoercedStringPtr(params...)
}

// Number returns a number schema with coercion enabled.
func Number(params ...any) *types.ZodFloatTyped[float64, float64] {
	return types.CoercedNumber(params...)
}

// NumberPtr returns a *number schema with coercion enabled.
func NumberPtr(params ...any) *types.ZodFloatTyped[float64, *float64] {
	return types.CoercedNumberPtr(params...)
}

// Float returns a float schema with coercion enabled.
func Float(params ...any) *types.ZodFloatTyped[float64, float64] {
	return types.CoercedFloat[float64](params...)
}

// FloatPtr returns a *float64 schema with coercion enabled.
func FloatPtr(params ...any) *types.ZodFloatTyped[float64, *float64] {
	return types.CoercedFloatPtr(params...)
}

// Float32 returns a float32 schema with coercion enabled.
func Float32(params ...any) *types.ZodFloatTyped[float32, float32] {
	return types.CoercedFloat32(params...)
}

// Float32Ptr returns a *float32 schema with coercion enabled.
func Float32Ptr(params ...any) *types.ZodFloatTyped[float32, *float32] {
	return types.CoercedFloat32Ptr(params...)
}

// Float64 returns a float64 schema with coercion enabled.
func Float64(params ...any) *types.ZodFloatTyped[float64, float64] {
	return types.CoercedFloat64(params...)
}

// Float64Ptr returns a *float64 schema with coercion enabled.
func Float64Ptr(params ...any) *types.ZodFloatTyped[float64, *float64] {
	return types.CoercedFloat64Ptr(params...)
}

// Integer returns a generic integer schema with coercion enabled.
func Integer(params ...any) *types.ZodIntegerTyped[int64, int64] {
	return types.CoercedInteger(params...)
}

// IntegerPtr returns a *int64 schema with coercion enabled.
func IntegerPtr(params ...any) *types.ZodIntegerTyped[int64, *int64] {
	return types.CoercedIntegerPtr(params...)
}

// Int returns an int schema with coercion enabled.
func Int(params ...any) *types.ZodIntegerTyped[int, int] {
	return types.CoercedInt(params...)
}

// IntPtr returns a *int schema with coercion enabled.
func IntPtr(params ...any) *types.ZodIntegerTyped[int, *int] {
	return types.CoercedIntPtr(params...)
}

// Int8 returns an int8 schema with coercion enabled.
func Int8(params ...any) *types.ZodIntegerTyped[int8, int8] {
	return types.CoercedInt8(params...)
}

// Int8Ptr returns a *int8 schema with coercion enabled.
func Int8Ptr(params ...any) *types.ZodIntegerTyped[int8, *int8] {
	return types.CoercedInt8Ptr(params...)
}

// Int16 returns an int16 schema with coercion enabled.
func Int16(params ...any) *types.ZodIntegerTyped[int16, int16] {
	return types.CoercedInt16(params...)
}

// Int16Ptr returns a *int16 schema with coercion enabled.
func Int16Ptr(params ...any) *types.ZodIntegerTyped[int16, *int16] {
	return types.CoercedInt16Ptr(params...)
}

// Int32 returns an int32 schema with coercion enabled.
func Int32(params ...any) *types.ZodIntegerTyped[int32, int32] {
	return types.CoercedInt32(params...)
}

// Int32Ptr returns a *int32 schema with coercion enabled.
func Int32Ptr(params ...any) *types.ZodIntegerTyped[int32, *int32] {
	return types.CoercedInt32Ptr(params...)
}

// Int64 returns an int64 schema with coercion enabled.
func Int64(params ...any) *types.ZodIntegerTyped[int64, int64] {
	return types.CoercedInt64(params...)
}

// Int64Ptr returns a *int64 schema with coercion enabled.
func Int64Ptr(params ...any) *types.ZodIntegerTyped[int64, *int64] {
	return types.CoercedInt64Ptr(params...)
}

// Uint returns a uint schema with coercion enabled.
func Uint(params ...any) *types.ZodIntegerTyped[uint, uint] {
	return types.CoercedUint(params...)
}

// UintPtr returns a *uint schema with coercion enabled.
func UintPtr(params ...any) *types.ZodIntegerTyped[uint, *uint] {
	return types.CoercedUintPtr(params...)
}

// Uint8 returns a uint8 schema with coercion enabled.
func Uint8(params ...any) *types.ZodIntegerTyped[uint8, uint8] {
	return types.CoercedUint8(params...)
}

// Uint8Ptr returns a *uint8 schema with coercion enabled.
func Uint8Ptr(params ...any) *types.ZodIntegerTyped[uint8, *uint8] {
	return types.CoercedUint8Ptr(params...)
}

// Uint16 returns a uint16 schema with coercion enabled.
func Uint16(params ...any) *types.ZodIntegerTyped[uint16, uint16] {
	return types.CoercedUint16(params...)
}

// Uint16Ptr returns a *uint16 schema with coercion enabled.
func Uint16Ptr(params ...any) *types.ZodIntegerTyped[uint16, *uint16] {
	return types.CoercedUint16Ptr(params...)
}

// Uint32 returns a uint32 schema with coercion enabled.
func Uint32(params ...any) *types.ZodIntegerTyped[uint32, uint32] {
	return types.CoercedUint32(params...)
}

// Uint32Ptr returns a *uint32 schema with coercion enabled.
func Uint32Ptr(params ...any) *types.ZodIntegerTyped[uint32, *uint32] {
	return types.CoercedUint32Ptr(params...)
}

// Uint64 returns a uint64 schema with coercion enabled.
func Uint64(params ...any) *types.ZodIntegerTyped[uint64, uint64] {
	return types.CoercedUint64(params...)
}

// Uint64Ptr returns a *uint64 schema with coercion enabled.
func Uint64Ptr(params ...any) *types.ZodIntegerTyped[uint64, *uint64] {
	return types.CoercedUint64Ptr(params...)
}

// Complex returns a complex128 schema with coercion enabled.
func Complex(params ...any) *types.ZodComplex[complex128] {
	return types.CoercedComplex(params...)
}

// ComplexPtr returns a *complex128 schema with coercion enabled.
func ComplexPtr(params ...any) *types.ZodComplex[*complex128] {
	return types.CoercedComplexPtr(params...)
}

// Complex64 returns a complex64 schema with coercion enabled.
func Complex64(params ...any) *types.ZodComplex[complex64] {
	return types.CoercedComplex64(params...)
}

// Complex64Ptr returns a *complex64 schema with coercion enabled.
func Complex64Ptr(params ...any) *types.ZodComplex[*complex64] {
	return types.CoercedComplex64Ptr(params...)
}

// Complex128 returns a complex128 schema with coercion enabled.
func Complex128(params ...any) *types.ZodComplex[complex128] {
	return types.CoercedComplex128(params...)
}

// Complex128Ptr returns a *complex128 schema with coercion enabled.
func Complex128Ptr(params ...any) *types.ZodComplex[*complex128] {
	return types.CoercedComplex128Ptr(params...)
}

// BigInt returns a big.Int schema with coercion enabled.
func BigInt(params ...any) *types.ZodBigInt[*big.Int] {
	return types.CoercedBigInt(params...)
}

// BigIntPtr returns a **big.Int schema with coercion enabled.
func BigIntPtr(params ...any) *types.ZodBigInt[**big.Int] {
	return types.CoercedBigIntPtr(params...)
}

// Time returns a time.Time schema with coercion enabled.
func Time(params ...any) *types.ZodTime[time.Time] {
	return types.CoercedTime(params...)
}

// TimePtr returns a *time.Time schema with coercion enabled.
func TimePtr(params ...any) *types.ZodTime[*time.Time] {
	return types.CoercedTimePtr(params...)
}

// StringBool returns a bool schema with string-to-boolean coercion enabled.
func StringBool(params ...any) *types.ZodStringBool[bool] {
	return types.CoercedStringBool(params...)
}

// StringBoolPtr returns a *bool schema with string-to-boolean coercion enabled.
func StringBoolPtr(params ...any) *types.ZodStringBool[*bool] {
	return types.CoercedStringBoolPtr(params...)
}
