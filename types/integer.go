package types

import (
	"math"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// IntegerConstraint restricts values to supported integer types or their pointers.
type IntegerConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~*int | ~*int8 | ~*int16 | ~*int32 | ~*int64 | ~*uint | ~*uint8 | ~*uint16 | ~*uint32 | ~*uint64
}

// ZodIntegerDef is the configuration for integer validation.
type ZodIntegerDef struct {
	core.ZodTypeDef
}

// ZodIntegerInternals holds the internal state of an integer validator.
type ZodIntegerInternals struct {
	core.ZodTypeInternals
	Def *ZodIntegerDef
}

// ZodIntegerTyped is an integer validation schema with dual generic parameters.
// T is the base type (int, int32, int64, etc.) and R is the constraint type (T or *T).
type ZodIntegerTyped[T IntegerConstraint, R any] struct {
	internals *ZodIntegerInternals
}

// ZodInteger is a type alias for ZodIntegerTyped providing a unified interface.
type ZodInteger[T IntegerConstraint, R any] = ZodIntegerTyped[T, R]

// Internals returns the internal state of the schema.
func (z *ZodIntegerTyped[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodIntegerTyped[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodIntegerTyped[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// typeCode returns the ZodTypeCode for the base type T.
func (*ZodIntegerTyped[T, R]) typeCode() core.ZodTypeCode {
	return integerTypeCode[T]()
}

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodIntegerTyped[T, R]) withCheck(check core.ZodCheck) *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Coerce converts the input to the target integer type.
func (z *ZodIntegerTyped[T, R]) Coerce(input any) (any, bool) {
	var zero T
	switch any(zero).(type) {
	case int, *int:
		result, err := coerce.ToInteger[int](input)
		return result, err == nil
	case int8, *int8:
		result, err := coerce.ToInteger[int8](input)
		return result, err == nil
	case int16, *int16:
		result, err := coerce.ToInteger[int16](input)
		return result, err == nil
	case int32, *int32:
		result, err := coerce.ToInteger[int32](input)
		return result, err == nil
	case int64, *int64:
		result, err := coerce.ToInteger[int64](input)
		return result, err == nil
	case uint, *uint:
		result, err := coerce.ToInteger[uint](input)
		return result, err == nil
	case uint8, *uint8:
		result, err := coerce.ToInteger[uint8](input)
		return result, err == nil
	case uint16, *uint16:
		result, err := coerce.ToInteger[uint16](input)
		return result, err == nil
	case uint32, *uint32:
		result, err := coerce.ToInteger[uint32](input)
		return result, err == nil
	case uint64, *uint64:
		result, err := coerce.ToInteger[uint64](input)
		return result, err == nil
	default:
		result, err := coerce.ToInteger[int64](input)
		return result, err == nil
	}
}

// Parse validates input and returns a value matching the constraint type R.
func (z *ZodIntegerTyped[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		z.typeCode(),
		engine.ApplyChecks[T],
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// MustParse panics if Parse returns an error.
func (z *ZodIntegerTyped[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety by requiring exact type matching.
func (z *ZodIntegerTyped[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	expected := z.typeCode()
	if z.internals.Type != "" {
		expected = z.internals.Type
	}

	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		expected,
		engine.ApplyChecks[T],
		ctx...,
	)
}

// MustStrictParse panics if StrictParse returns an error.
//
// Example usage:
//
//	schema := gozod.Int().Min(0).Max(100)
//	result := schema.MustStrictParse(42)           // ✅ int → int
//	result := schema.MustStrictParse(&num)         // ❌ compile error
func (z *ZodIntegerTyped[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodIntegerTyped[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// withPtrInternals creates a new ZodIntegerTyped with pointer constraint type *T.
func (z *ZodIntegerTyped[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodIntegerTyped[T, *T] {
	return &ZodIntegerTyped[T, *T]{internals: &ZodIntegerInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// Optional returns a schema that accepts T or nil, with constraint type *T.
func (z *ZodIntegerTyped[T, R]) Optional() *ZodIntegerTyped[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional, ExactOptional only accepts absent keys in object fields.
func (z *ZodIntegerTyped[T, R]) ExactOptional() *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts T or nil, with constraint type *T.
func (z *ZodIntegerTyped[T, R]) Nilable() *ZodIntegerTyped[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodIntegerTyped[T, R]) Nullish() *ZodIntegerTyped[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and returns a value constraint (T).
func (z *ZodIntegerTyped[T, R]) NonOptional() *ZodIntegerTyped[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodIntegerTyped[T, T]{
		internals: &ZodIntegerInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default sets a default value for nil inputs.
func (z *ZodIntegerTyped[T, R]) Default(v int64) *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(convertIntDefaultValue[T](v))
	return z.withInternals(in)
}

// DefaultFunc sets a lazy default value using a function.
func (z *ZodIntegerTyped[T, R]) DefaultFunc(fn func() int64) *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full validation pipeline.
func (z *ZodIntegerTyped[T, R]) Prefault(v int64) *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(convertIntDefaultValue[T](v))
	return z.withInternals(in)
}

// PrefaultFunc sets a lazy prefault value using a function.
func (z *ZodIntegerTyped[T, R]) PrefaultFunc(fn func() R) *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this integer schema in the global registry.
func (z *ZodIntegerTyped[T, R]) Meta(meta core.GlobalMeta) *ZodIntegerTyped[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodIntegerTyped[T, R]) Describe(description string) *ZodIntegerTyped[T, R] {
	in := z.internals.Clone()

	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description

	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// Min adds minimum value validation (alias for Gte).
func (z *ZodIntegerTyped[T, R]) Min(minimum int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.Gte(minimum, params...))
}

// Max adds maximum value validation (alias for Lte).
func (z *ZodIntegerTyped[T, R]) Max(maximum int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.Lte(maximum, params...))
}

// Gt adds greater-than validation (exclusive).
func (z *ZodIntegerTyped[T, R]) Gt(value int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.Gt(value, params...))
}

// Gte adds greater-than-or-equal validation (inclusive).
func (z *ZodIntegerTyped[T, R]) Gte(value int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.Gte(value, params...))
}

// Lt adds less-than validation (exclusive).
func (z *ZodIntegerTyped[T, R]) Lt(value int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.Lt(value, params...))
}

// Lte adds less-than-or-equal validation (inclusive).
func (z *ZodIntegerTyped[T, R]) Lte(value int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.Lte(value, params...))
}

// Positive adds positive number validation (> 0).
func (z *ZodIntegerTyped[T, R]) Positive(params ...any) *ZodIntegerTyped[T, R] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0).
func (z *ZodIntegerTyped[T, R]) Negative(params ...any) *ZodIntegerTyped[T, R] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0).
func (z *ZodIntegerTyped[T, R]) NonNegative(params ...any) *ZodIntegerTyped[T, R] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0).
func (z *ZodIntegerTyped[T, R]) NonPositive(params ...any) *ZodIntegerTyped[T, R] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple-of validation.
func (z *ZodIntegerTyped[T, R]) MultipleOf(value int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.withCheck(checks.MultipleOf(value, params...))
}

// Step adds step validation (alias for MultipleOf).
func (z *ZodIntegerTyped[T, R]) Step(step int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.MultipleOf(step, params...)
}

// Safe adds safe integer validation (within JavaScript safe integer range).
func (z *ZodIntegerTyped[T, R]) Safe(params ...any) *ZodIntegerTyped[T, R] {
	const maxSafeInt = 1<<53 - 1
	const minSafeInt = -(1<<53 - 1)
	return z.Gte(minSafeInt, params...).Lte(maxSafeInt, params...)
}

// Transform applies a transformation function that extracts int64 values.
func (z *ZodIntegerTyped[T, R]) Transform(fn func(int64, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractIntegerToInt64[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodIntegerTyped[T, R]) Overwrite(transform func(T) T, params ...any) *ZodIntegerTyped[T, R] {
	fn := func(input any) any {
		converted, ok := convertToIntegerType[T](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	check := checks.NewZodCheckOverwrite(fn, params...)
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodIntegerTyped[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	fn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractIntegerToInt64[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, fn)
}

// extractIntegerToInt64 converts constraint type R to int64.
func extractIntegerToInt64[T IntegerConstraint, R any](value R) int64 {
	base := extractIntegerValue[T, R](value)
	switch v := any(base).(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		if v > math.MaxInt64 {
			return 0
		}
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		if v > math.MaxInt64 {
			return 0
		}
		return int64(v)
	default:
		return 0
	}
}

// Refine applies type-safe custom validation using the base type T.
func (z *ZodIntegerTyped[T, R]) Refine(fn func(T) bool, params ...any) *ZodIntegerTyped[T, R] {
	wrapper := func(v any) bool {
		if v == nil && z.IsNilable() {
			return true
		}
		converted, ok := convertToIntegerType[T](v)
		if !ok {
			return false
		}
		if v == nil {
			return true
		}
		return fn(converted)
	}

	param := utils.FirstParam(params...)
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(param))
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// And creates an intersection with another schema.
//
// Example:
//
//	schema := gozod.Int().Min(0).And(gozod.Int().Max(100))
//	result, _ := schema.Parse(50)
func (z *ZodIntegerTyped[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
//
// Example:
//
//	schema := gozod.Int().Or(gozod.String())
//	result, _ := schema.Parse(42)
func (z *ZodIntegerTyped[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// convertToIntegerType converts matching integer values to the target type T.
func convertToIntegerType[T IntegerConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
			return zero, true
		default:
			return zero, false
		}
	}

	if val, ok := v.(T); ok {
		return val, true
	}

	// Try coercion using the coerce package.
	switch any(zero).(type) {
	case int:
		if converted, err := coerce.ToInteger[int](v); err == nil {
			return any(converted).(T), true
		}
	case int8:
		if converted, err := coerce.ToInteger[int8](v); err == nil {
			return any(converted).(T), true
		}
	case int16:
		if converted, err := coerce.ToInteger[int16](v); err == nil {
			return any(converted).(T), true
		}
	case int32:
		if converted, err := coerce.ToInteger[int32](v); err == nil {
			return any(converted).(T), true
		}
	case int64:
		if converted, err := coerce.ToInteger[int64](v); err == nil {
			return any(converted).(T), true
		}
	case uint:
		if converted, err := coerce.ToInteger[uint](v); err == nil {
			return any(converted).(T), true
		}
	case uint8:
		if converted, err := coerce.ToInteger[uint8](v); err == nil {
			return any(converted).(T), true
		}
	case uint16:
		if converted, err := coerce.ToInteger[uint16](v); err == nil {
			return any(converted).(T), true
		}
	case uint32:
		if converted, err := coerce.ToInteger[uint32](v); err == nil {
			return any(converted).(T), true
		}
	case uint64:
		if converted, err := coerce.ToInteger[uint64](v); err == nil {
			return any(converted).(T), true
		}
	// Pointer types
	case *int:
		if converted, err := coerce.ToInteger[int](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *int8:
		if converted, err := coerce.ToInteger[int8](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *int16:
		if converted, err := coerce.ToInteger[int16](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *int32:
		if converted, err := coerce.ToInteger[int32](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *int64:
		if converted, err := coerce.ToInteger[int64](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *uint:
		if converted, err := coerce.ToInteger[uint](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *uint8:
		if converted, err := coerce.ToInteger[uint8](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *uint16:
		if converted, err := coerce.ToInteger[uint16](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *uint32:
		if converted, err := coerce.ToInteger[uint32](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	case *uint64:
		if converted, err := coerce.ToInteger[uint64](v); err == nil {
			ptr := &converted
			return any(ptr).(T), true
		}
	}

	return zero, false
}

// RefineAny adds flexible custom validation using an untyped function.
func (z *ZodIntegerTyped[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withInternals creates a new ZodIntegerTyped that keeps the original generic types.
func (z *ZodIntegerTyped[T, R]) withInternals(in *core.ZodTypeInternals) *ZodIntegerTyped[T, R] {
	return &ZodIntegerTyped[T, R]{internals: &ZodIntegerInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodIntegerTyped[T, R]) CloneFrom(source any) {
	src, ok := source.(*ZodIntegerTyped[T, R])
	if !ok {
		return
	}
	orig := z.internals.Checks
	*z.internals = *src.internals
	z.internals.Checks = orig
}

// extractIntegerValue extracts the base integer value T from constraint type R.
func extractIntegerValue[T IntegerConstraint, R any](value R) T {
	if v, ok := any(value).(T); ok {
		return v
	}

	// Handle pointer dereferencing.
	switch v := any(value).(type) {
	case *int:
		if v != nil {
			return any(*v).(T)
		}
	case *int8:
		if v != nil {
			return any(*v).(T)
		}
	case *int16:
		if v != nil {
			return any(*v).(T)
		}
	case *int32:
		if v != nil {
			return any(*v).(T)
		}
	case *int64:
		if v != nil {
			return any(*v).(T)
		}
	case *uint:
		if v != nil {
			return any(*v).(T)
		}
	case *uint8:
		if v != nil {
			return any(*v).(T)
		}
	case *uint16:
		if v != nil {
			return any(*v).(T)
		}
	case *uint32:
		if v != nil {
			return any(*v).(T)
		}
	case *uint64:
		if v != nil {
			return any(*v).(T)
		}
	}

	var zero T
	return zero
}

// newZodIntegerFromDef constructs a new ZodIntegerTyped from the given definition.
func newZodIntegerFromDef[T IntegerConstraint, R any](def *ZodIntegerDef) *ZodIntegerTyped[T, R] {
	in := &ZodIntegerInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}
	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodIntegerDef{ZodTypeDef: *newDef}
		return any(newZodIntegerFromDef[T, R](d)).(core.ZodType[any])
	}
	if def.Error != nil {
		in.Error = def.Error
	}
	return &ZodIntegerTyped[T, R]{internals: in}
}

// IntegerTyped creates a generic integer schema with automatic type inference.
func IntegerTyped[T IntegerConstraint](params ...any) *ZodIntegerTyped[T, T] {
	return newIntegerTyped[T, T](integerTypeCode[T](), params...)
}

// Int creates a standard int schema.
func Int(params ...any) *ZodIntegerTyped[int, int] {
	return newIntegerTyped[int, int](core.ZodTypeInt, params...)
}

// IntPtr creates a schema for *int.
func IntPtr(params ...any) *ZodIntegerTyped[int, *int] {
	return newIntegerTyped[int, *int](core.ZodTypeInt, params...)
}

// Int8 creates an int8 schema.
func Int8(params ...any) *ZodIntegerTyped[int8, int8] {
	return newIntegerTyped[int8, int8](core.ZodTypeInt8, params...)
}

// Int8Ptr creates a schema for *int8.
func Int8Ptr(params ...any) *ZodIntegerTyped[int8, *int8] {
	return newIntegerTyped[int8, *int8](core.ZodTypeInt8, params...)
}

// Int16 creates an int16 schema.
func Int16(params ...any) *ZodIntegerTyped[int16, int16] {
	return newIntegerTyped[int16, int16](core.ZodTypeInt16, params...)
}

// Int16Ptr creates a schema for *int16.
func Int16Ptr(params ...any) *ZodIntegerTyped[int16, *int16] {
	return newIntegerTyped[int16, *int16](core.ZodTypeInt16, params...)
}

// Int32 creates an int32 schema.
func Int32(params ...any) *ZodIntegerTyped[int32, int32] {
	return newIntegerTyped[int32, int32](core.ZodTypeInt32, params...)
}

// Int32Ptr creates a schema for *int32.
func Int32Ptr(params ...any) *ZodIntegerTyped[int32, *int32] {
	return newIntegerTyped[int32, *int32](core.ZodTypeInt32, params...)
}

// Int64 creates an int64 schema.
func Int64(params ...any) *ZodIntegerTyped[int64, int64] {
	return newIntegerTyped[int64, int64](core.ZodTypeInt64, params...)
}

// Int64Ptr creates a schema for *int64.
func Int64Ptr(params ...any) *ZodIntegerTyped[int64, *int64] {
	return newIntegerTyped[int64, *int64](core.ZodTypeInt64, params...)
}

// Uint creates a uint schema.
func Uint(params ...any) *ZodIntegerTyped[uint, uint] {
	return newIntegerTyped[uint, uint](core.ZodTypeUint, params...)
}

// UintPtr creates a schema for *uint.
func UintPtr(params ...any) *ZodIntegerTyped[uint, *uint] {
	return newIntegerTyped[uint, *uint](core.ZodTypeUint, params...)
}

// Uint8 creates a uint8 schema.
func Uint8(params ...any) *ZodIntegerTyped[uint8, uint8] {
	return newIntegerTyped[uint8, uint8](core.ZodTypeUint8, params...)
}

// Uint8Ptr creates a schema for *uint8.
func Uint8Ptr(params ...any) *ZodIntegerTyped[uint8, *uint8] {
	return newIntegerTyped[uint8, *uint8](core.ZodTypeUint8, params...)
}

// Uint16 creates a uint16 schema.
func Uint16(params ...any) *ZodIntegerTyped[uint16, uint16] {
	return newIntegerTyped[uint16, uint16](core.ZodTypeUint16, params...)
}

// Uint16Ptr creates a schema for *uint16.
func Uint16Ptr(params ...any) *ZodIntegerTyped[uint16, *uint16] {
	return newIntegerTyped[uint16, *uint16](core.ZodTypeUint16, params...)
}

// Uint32 creates a uint32 schema.
func Uint32(params ...any) *ZodIntegerTyped[uint32, uint32] {
	return newIntegerTyped[uint32, uint32](core.ZodTypeUint32, params...)
}

// Uint32Ptr creates a schema for *uint32.
func Uint32Ptr(params ...any) *ZodIntegerTyped[uint32, *uint32] {
	return newIntegerTyped[uint32, *uint32](core.ZodTypeUint32, params...)
}

// Uint64 creates a uint64 schema.
func Uint64(params ...any) *ZodIntegerTyped[uint64, uint64] {
	return newIntegerTyped[uint64, uint64](core.ZodTypeUint64, params...)
}

// Uint64Ptr creates a schema for *uint64.
func Uint64Ptr(params ...any) *ZodIntegerTyped[uint64, *uint64] {
	return newIntegerTyped[uint64, *uint64](core.ZodTypeUint64, params...)
}

// Byte creates a uint8 schema (alias for byte).
func Byte(params ...any) *ZodIntegerTyped[uint8, uint8] {
	return Uint8(params...)
}

// BytePtr creates a schema for *uint8 (alias for *byte).
func BytePtr(params ...any) *ZodIntegerTyped[uint8, *uint8] {
	return Uint8Ptr(params...)
}

// Rune creates an int32 schema (alias for rune).
func Rune(params ...any) *ZodIntegerTyped[int32, int32] {
	return Int32(params...)
}

// RunePtr creates a schema for *int32 (alias for *rune).
func RunePtr(params ...any) *ZodIntegerTyped[int32, *int32] {
	return Int32Ptr(params...)
}

// newIntegerTyped creates an integer schema with the given type code.
func newIntegerTyped[T IntegerConstraint, R any](typeCode core.ZodTypeCode, params ...any) *ZodIntegerTyped[T, R] {
	p := utils.NormalizeParams(params...)
	def := &ZodIntegerDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeCode,
			Checks: []core.ZodCheck{},
		},
	}
	utils.ApplySchemaParams(&def.ZodTypeDef, p)
	return newZodIntegerFromDef[T, R](def)
}

// CoercedInteger creates an int64 schema with coercion enabled.
func CoercedInteger(params ...any) *ZodIntegerTyped[int64, int64] {
	schema := Int64(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedIntegerPtr creates a *int64 schema with coercion enabled.
func CoercedIntegerPtr(params ...any) *ZodIntegerTyped[int64, *int64] {
	schema := Int64Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt creates an int schema with coercion enabled.
func CoercedInt(params ...any) *ZodIntegerTyped[int, int] {
	schema := Int(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedIntPtr creates a *int schema with coercion enabled.
func CoercedIntPtr(params ...any) *ZodIntegerTyped[int, *int] {
	schema := IntPtr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt8 creates an int8 schema with coercion enabled.
func CoercedInt8(params ...any) *ZodIntegerTyped[int8, int8] {
	schema := Int8(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt8Ptr creates a *int8 schema with coercion enabled.
func CoercedInt8Ptr(params ...any) *ZodIntegerTyped[int8, *int8] {
	schema := Int8Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt16 creates an int16 schema with coercion enabled.
func CoercedInt16(params ...any) *ZodIntegerTyped[int16, int16] {
	schema := Int16(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt16Ptr creates a *int16 schema with coercion enabled.
func CoercedInt16Ptr(params ...any) *ZodIntegerTyped[int16, *int16] {
	schema := Int16Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt32 creates an int32 schema with coercion enabled.
func CoercedInt32(params ...any) *ZodIntegerTyped[int32, int32] {
	schema := Int32(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt32Ptr creates a *int32 schema with coercion enabled.
func CoercedInt32Ptr(params ...any) *ZodIntegerTyped[int32, *int32] {
	schema := Int32Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt64 creates an int64 schema with coercion enabled.
func CoercedInt64(params ...any) *ZodIntegerTyped[int64, int64] {
	schema := Int64(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedInt64Ptr creates a *int64 schema with coercion enabled.
func CoercedInt64Ptr(params ...any) *ZodIntegerTyped[int64, *int64] {
	schema := Int64Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint creates a uint schema with coercion enabled.
func CoercedUint(params ...any) *ZodIntegerTyped[uint, uint] {
	schema := Uint(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUintPtr creates a *uint schema with coercion enabled.
func CoercedUintPtr(params ...any) *ZodIntegerTyped[uint, *uint] {
	schema := UintPtr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint8 creates a uint8 schema with coercion enabled.
func CoercedUint8(params ...any) *ZodIntegerTyped[uint8, uint8] {
	schema := Uint8(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint8Ptr creates a *uint8 schema with coercion enabled.
func CoercedUint8Ptr(params ...any) *ZodIntegerTyped[uint8, *uint8] {
	schema := Uint8Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint16 creates a uint16 schema with coercion enabled.
func CoercedUint16(params ...any) *ZodIntegerTyped[uint16, uint16] {
	schema := Uint16(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint16Ptr creates a *uint16 schema with coercion enabled.
func CoercedUint16Ptr(params ...any) *ZodIntegerTyped[uint16, *uint16] {
	schema := Uint16Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint32 creates a uint32 schema with coercion enabled.
func CoercedUint32(params ...any) *ZodIntegerTyped[uint32, uint32] {
	schema := Uint32(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint32Ptr creates a *uint32 schema with coercion enabled.
func CoercedUint32Ptr(params ...any) *ZodIntegerTyped[uint32, *uint32] {
	schema := Uint32Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint64 creates a uint64 schema with coercion enabled.
func CoercedUint64(params ...any) *ZodIntegerTyped[uint64, uint64] {
	schema := Uint64(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedUint64Ptr creates a *uint64 schema with coercion enabled.
func CoercedUint64Ptr(params ...any) *ZodIntegerTyped[uint64, *uint64] {
	schema := Uint64Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// Integer creates a flexible integer schema (alias for Int64).
func Integer(params ...any) *ZodInteger[int64, int64] {
	return Int64(params...)
}

// integerTypeCode returns the ZodTypeCode for the given integer type T.
func integerTypeCode[T IntegerConstraint]() core.ZodTypeCode {
	var zero T
	switch any(zero).(type) {
	case int, *int:
		return core.ZodTypeInt
	case int8, *int8:
		return core.ZodTypeInt8
	case int16, *int16:
		return core.ZodTypeInt16
	case int32, *int32:
		return core.ZodTypeInt32
	case int64, *int64:
		return core.ZodTypeInt64
	case uint, *uint:
		return core.ZodTypeUint
	case uint8, *uint8:
		return core.ZodTypeUint8
	case uint16, *uint16:
		return core.ZodTypeUint16
	case uint32, *uint32:
		return core.ZodTypeUint32
	case uint64, *uint64:
		return core.ZodTypeUint64
	default:
		return core.ZodTypeInt64
	}
}

// convertIntDefaultValue converts an int64 to the appropriate type for Default/Prefault.
func convertIntDefaultValue[T IntegerConstraint](v int64) any {
	var zero T
	switch any(zero).(type) {
	case int, *int:
		return int(v)
	case int8, *int8:
		return int8(v) // #nosec G115
	case int16, *int16:
		return int16(v) // #nosec G115
	case int32, *int32:
		return int32(v) // #nosec G115
	case int64, *int64:
		return v
	case uint, *uint:
		return uint(v) // #nosec G115
	case uint8, *uint8:
		return uint8(v) // #nosec G115
	case uint16, *uint16:
		return uint16(v) // #nosec G115
	case uint32, *uint32:
		return uint32(v) // #nosec G115
	case uint64, *uint64:
		return uint64(v) // #nosec G115
	default:
		return v
	}
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodIntegerTyped[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodIntegerTyped[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(R); ok {
			fn(val, payload)
			return
		}
		// Handle pointer/value mismatch: if R is pointer but payload holds value.
		var zero R
		rt := reflect.TypeOf(zero)
		if rt != nil && rt.Kind() == reflect.Pointer {
			et := rt.Elem()
			rv := reflect.ValueOf(payload.Value())
			if rv.IsValid() && rv.Type() == et {
				ptr := reflect.New(et)
				ptr.Elem().Set(rv)
				if casted, ok := ptr.Interface().(R); ok {
					fn(casted, payload)
				}
			}
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodIntegerTyped[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodIntegerTyped[T, R] {
	return z.Check(fn, params...)
}
