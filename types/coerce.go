package types

import "github.com/kaptinlin/gozod/core"

// =============================================================================
// COERCE NAMESPACE
// =============================================================================

// CoerceNamespace exposes factory helpers for coercive schemas via a namespace,
// e.g.  Coerce.String().Max(10).
//
// Design goal: mirror the ergonomics of the JavaScript zod API.
//
// All concrete CoercedXXX constructors delegate to enableCoercion which writes
// the "coerce" flag into ZodTypeInternals.Bag and then makes the type-specific
// Bag field point to the same map so that both the engine and tests see the
// flag.
// =============================================================================

type CoerceNamespace struct{}

// Global singleton, matching the JS implementation style.
var Coerce = &CoerceNamespace{}

// =============================================================================
// PRIMITIVE TYPE COERCERS
// =============================================================================

func (c *CoerceNamespace) String(params ...any) *ZodString         { return CoercedString(params...) }
func (c *CoerceNamespace) Number(params ...any) *ZodFloat[float64] { return CoercedNumber(params...) }
func (c *CoerceNamespace) Bool(params ...any) *ZodBool             { return CoercedBool(params...) }
func (c *CoerceNamespace) BigInt(params ...any) *ZodBigInt         { return CoercedBigInt(params...) }
func (c *CoerceNamespace) Complex64(params ...any) *ZodComplex[complex64] {
	return CoercedComplex64(params...)
}
func (c *CoerceNamespace) Complex128(params ...any) *ZodComplex[complex128] {
	return CoercedComplex128(params...)
}

// CoercedString creates a string schema with coercion enabled
func CoercedString(params ...any) *ZodString {
	schema := String(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}

// CoercedNumber creates a number schema with coercion enabled (float64)
func CoercedNumber(params ...any) *ZodFloat[float64] {
	schema := Float64(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}

// CoercedBool creates a boolean schema with coercion enabled
func CoercedBool(params ...any) *ZodBool {
	schema := Bool(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}

// CoercedBigInt creates a bigint schema with coercion enabled
func CoercedBigInt(params ...any) *ZodBigInt {
	schema := BigInt(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}

// CoercedComplex64 creates a complex64 schema with coercion enabled
func CoercedComplex64(params ...any) *ZodComplex[complex64] {
	schema := Complex64(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}

// CoercedComplex128 creates a complex128 schema with coercion enabled
func CoercedComplex128(params ...any) *ZodComplex[complex128] {
	schema := Complex128(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}

// -------------------------- Integer series --------------------------

func CoercedInt(params ...any) *ZodInteger[int] {
	schema := Int(params...)
	enableCoercion(&schema.internals.ZodTypeInternals)
	return schema
}
func CoercedInt8(params ...any) *ZodInteger[int8] {
	s := Int8(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedInt16(params ...any) *ZodInteger[int16] {
	s := Int16(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedInt32(params ...any) *ZodInteger[int32] {
	s := Int32(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedInt64(params ...any) *ZodInteger[int64] {
	s := Int64(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}

func CoercedUint(params ...any) *ZodInteger[uint] {
	s := Uint(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedUint8(params ...any) *ZodInteger[uint8] {
	s := Uint8(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedUint16(params ...any) *ZodInteger[uint16] {
	s := Uint16(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedUint32(params ...any) *ZodInteger[uint32] {
	s := Uint32(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedUint64(params ...any) *ZodInteger[uint64] {
	s := Uint64(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}

// -------------------------- Float series --------------------------

func CoercedFloat32(params ...any) *ZodFloat[float32] {
	s := Float32(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}
func CoercedFloat64(params ...any) *ZodFloat[float64] {
	s := Float64(params...)
	enableCoercion(&s.internals.ZodTypeInternals)
	return s
}

// =============================================================================
// INTERNAL UTILITIES
// =============================================================================

// enableCoercion writes "coerce=true" into ZodTypeInternals.Bag to let the
// parser know that generic coercion should be attempted.
func enableCoercion(internals *core.ZodTypeInternals) {
	if internals.Bag == nil {
		internals.Bag = make(map[string]any)
	}
	internals.Bag["coerce"] = true
}
