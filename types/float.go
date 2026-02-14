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

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// FloatConstraint restricts values to supported float types or their pointers.
type FloatConstraint interface {
	~float32 | ~float64 | ~*float32 | ~*float64
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodFloatDef defines the configuration for float validation.
type ZodFloatDef struct {
	core.ZodTypeDef
}

// ZodFloatInternals contains float validator internal state.
type ZodFloatInternals struct {
	core.ZodTypeInternals
	Def *ZodFloatDef
}

// ZodFloatTyped represents a floating-point validation schema with dual generic parameters.
// T is the base type (float32, float64), R is the constraint type (T or *T).
type ZodFloatTyped[T FloatConstraint, R any] struct {
	internals *ZodFloatInternals
}

// ZodFloat is a generic type alias for ZodFloatTyped.
type ZodFloat[T FloatConstraint, R any] = ZodFloatTyped[T, R]

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodFloatTyped[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodFloatTyped[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodFloatTyped[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// expectedType returns the ZodTypeCode for the base type T.
func (*ZodFloatTyped[T, R]) expectedType() core.ZodTypeCode {
	return floatTypeCode[T]()
}

// Coerce converts input to the target float type.
func (z *ZodFloatTyped[T, R]) Coerce(input any) (any, bool) {
	var zero T
	switch any(zero).(type) {
	case float32, *float32:
		result, err := coerce.ToFloat[float32](input)
		return result, err == nil
	default: // float64, *float64
		result, err := coerce.ToFloat[float64](input)
		return result, err == nil
	}
}

// Parse validates input and returns a value matching the constraint type R.
func (z *ZodFloatTyped[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		z.expectedType(),
		engine.ApplyChecks[T],
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodFloatTyped[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodFloatTyped[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	expected := z.expectedType()
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

// MustStrictParse validates with compile-time type safety and panics on error.
func (z *ZodFloatTyped[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodFloatTyped[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts the base type T or nil, with constraint type *T.
func (z *ZodFloatTyped[T, R]) Optional() *ZodFloatTyped[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodFloatTyped[T, R]) ExactOptional() *ZodFloatTyped[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts the base type T or nil, with constraint type *T.
func (z *ZodFloatTyped[T, R]) Nilable() *ZodFloatTyped[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodFloatTyped[T, R]) Nullish() *ZodFloatTyped[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes optional flag and returns value constraint T.
func (z *ZodFloatTyped[T, R]) NonOptional() *ZodFloatTyped[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodFloatTyped[T, T]{
		internals: &ZodFloatInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default sets a default value, keeping the current constraint type R.
func (z *ZodFloatTyped[T, R]) Default(v float64) *ZodFloatTyped[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(convertDefaultValue[R](v))
	return z.withInternals(in)
}

// DefaultFunc sets a lazy default value, keeping the current constraint type R.
func (z *ZodFloatTyped[T, R]) DefaultFunc(fn func() float64) *ZodFloatTyped[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return convertDefaultValue[R](fn())
	})
	return z.withInternals(in)
}

// Prefault sets a prefault value, keeping the current constraint type R.
func (z *ZodFloatTyped[T, R]) Prefault(v float64) *ZodFloatTyped[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(convertDefaultValue[R](v))
	return z.withInternals(in)
}

// PrefaultFunc sets a lazy prefault value, keeping the current constraint type R.
func (z *ZodFloatTyped[T, R]) PrefaultFunc(fn func() float64) *ZodFloatTyped[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return convertDefaultValue[R](fn())
	})
	return z.withInternals(in)
}

// Meta stores metadata for this float schema in the global registry.
func (z *ZodFloatTyped[T, R]) Meta(meta core.GlobalMeta) *ZodFloatTyped[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodFloatTyped[T, R]) Describe(description string) *ZodFloatTyped[T, R] {
	in := z.internals.Clone()

	meta, ok := core.GlobalRegistry.Get(z)
	if !ok {
		meta = core.GlobalMeta{}
	}
	meta.Description = description

	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, meta)

	return clone
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min adds minimum value validation (alias for Gte).
func (z *ZodFloatTyped[T, R]) Min(minimum float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.Gte(minimum, params...))
}

// Max adds maximum value validation (alias for Lte).
func (z *ZodFloatTyped[T, R]) Max(maximum float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.Lte(maximum, params...))
}

// Gt adds greater-than validation (exclusive).
func (z *ZodFloatTyped[T, R]) Gt(value float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.Gt(value, params...))
}

// Gte adds greater-than-or-equal validation (inclusive).
func (z *ZodFloatTyped[T, R]) Gte(value float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.Gte(value, params...))
}

// Lt adds less-than validation (exclusive).
func (z *ZodFloatTyped[T, R]) Lt(value float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.Lt(value, params...))
}

// Lte adds less-than-or-equal validation (inclusive).
func (z *ZodFloatTyped[T, R]) Lte(value float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.Lte(value, params...))
}

// Positive adds positive number validation (> 0).
func (z *ZodFloatTyped[T, R]) Positive(params ...any) *ZodFloatTyped[T, R] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0).
func (z *ZodFloatTyped[T, R]) Negative(params ...any) *ZodFloatTyped[T, R] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0).
func (z *ZodFloatTyped[T, R]) NonNegative(params ...any) *ZodFloatTyped[T, R] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0).
func (z *ZodFloatTyped[T, R]) NonPositive(params ...any) *ZodFloatTyped[T, R] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple-of validation.
func (z *ZodFloatTyped[T, R]) MultipleOf(value float64, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.MultipleOf(value, params...))
}

// Step adds step validation (alias for MultipleOf).
func (z *ZodFloatTyped[T, R]) Step(step float64, params ...any) *ZodFloatTyped[T, R] {
	return z.MultipleOf(step, params...)
}

// Int adds integer validation (no decimal part).
func (z *ZodFloatTyped[T, R]) Int(params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.NewCustom[any](func(v any) bool {
		switch val := v.(type) {
		case float32:
			return val == float32(math.Trunc(float64(val)))
		case float64:
			return val == math.Trunc(val)
		case *float32:
			return val == nil || *val == float32(math.Trunc(float64(*val)))
		case *float64:
			return val == nil || *val == math.Trunc(*val)
		default:
			return false
		}
	}, params...))
}

// Finite adds finite number validation (not NaN or Infinity).
func (z *ZodFloatTyped[T, R]) Finite(params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.NewCustom[any](func(v any) bool {
		switch val := v.(type) {
		case float32:
			f := float64(val)
			return !math.IsInf(f, 0) && !math.IsNaN(f)
		case float64:
			return !math.IsInf(val, 0) && !math.IsNaN(val)
		case *float32:
			if val == nil {
				return true
			}
			f := float64(*val)
			return !math.IsInf(f, 0) && !math.IsNaN(f)
		case *float64:
			if val == nil {
				return true
			}
			return !math.IsInf(*val, 0) && !math.IsNaN(*val)
		default:
			return false
		}
	}, params...))
}

// Safe adds safe number validation (within JavaScript safe integer range).
func (z *ZodFloatTyped[T, R]) Safe(params ...any) *ZodFloatTyped[T, R] {
	const safeIntMax = 1<<53 - 1
	const safeIntMin = -(1<<53 - 1)
	return z.Gte(safeIntMin, params...).Lte(safeIntMax, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed float value.
//
// Example:
//
//	schema := Float64().Min(0.0).Transform(func(f float64, ctx *RefinementContext) (string, error) {
//	    return fmt.Sprintf("%.2f", f), nil
//	})
func (z *ZodFloatTyped[T, R]) Transform(fn func(float64, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrap := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractFloatToFloat64[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrap)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this doesn't change the inferred type.
func (z *ZodFloatTyped[T, R]) Overwrite(transform func(T) T, params ...any) *ZodFloatTyped[T, R] {
	wrap := func(input any) any {
		val, ok := convertToFloatType[T](input)
		if !ok {
			return input
		}
		return transform(val)
	}
	return z.withCheck(checks.NewZodCheckOverwrite(wrap, params...))
}

// Pipe creates a pipeline that feeds the parsed value into another schema.
func (z *ZodFloatTyped[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	fn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractFloatToFloat64[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, fn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation using the base type T.
// Nil values bypass validation for nilable schemas per Zod v4 semantics.
//
// Example:
//
//	schema := Float64().Refine(func(f float64) bool {
//	    return f > 0 && f < 100
//	}, "value must be between 0 and 100")
func (z *ZodFloatTyped[T, R]) Refine(fn func(T) bool, params ...any) *ZodFloatTyped[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			return z.IsNilable()
		}

		val, ok := convertToFloatType[T](v)
		if !ok {
			return false
		}

		return fn(val)
	}

	sp := utils.NormalizeParams(params...)

	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}

	return z.withCheck(checks.NewCustom[any](wrapper, msg))
}

// convertToFloatType converts matching float values to the target type T.
func convertToFloatType[T FloatConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *float32, *float64:
			return zero, true
		default:
			return zero, false
		}
	}

	switch any(zero).(type) {
	case float32:
		if val, ok := v.(float32); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*float32); ok && val != nil {
			return any(*val).(T), true
		}
	case float64:
		if val, ok := v.(float64); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*float64); ok && val != nil {
			return any(*val).(T), true
		}
	case *float32:
		if val, ok := v.(float32); ok {
			return any(new(val)).(T), true
		}
		if val, ok := v.(*float32); ok {
			return any(val).(T), true
		}
	case *float64:
		if val, ok := v.(float64); ok {
			return any(new(val)).(T), true
		}
		if val, ok := v.(*float64); ok {
			return any(val).(T), true
		}
	}

	return zero, false
}

// RefineAny adds flexible custom validation logic that accepts any type.
func (z *ZodFloatTyped[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodFloatTyped[T, R] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// Check adds a custom validation function that can report multiple issues.
//
// Example:
//
//	schema := Float64().Check(func(value float64, payload *core.ParsePayload) {
//	    if value < 0 {
//	        payload.AddIssue(core.NewIssue("value must be positive"))
//	    }
//	})
func (z *ZodFloatTyped[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodFloatTyped[T, R] {
	wrapper := func(p *core.ParsePayload) {
		if val, ok := p.Value().(R); ok {
			fn(val, p)
			return
		}

		var zero R
		rt := reflect.TypeOf(zero)
		if rt != nil && rt.Kind() == reflect.Pointer {
			elem := rt.Elem()
			rv := reflect.ValueOf(p.Value())
			if rv.IsValid() && rv.Type() == elem {
				ptr := reflect.New(elem)
				ptr.Elem().Set(rv)
				if v, ok := ptr.Interface().(R); ok {
					fn(v, p)
				}
			}
		}
	}
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodFloatTyped[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodFloatTyped[T, R] {
	return z.Check(fn, params...)
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodFloatTyped[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodFloatTyped[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodFloatTyped[T, R]) withCheck(check core.ZodCheck) *ZodFloatTyped[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withInternals creates a new instance preserving the constraint type R.
func (z *ZodFloatTyped[T, R]) withInternals(in *core.ZodTypeInternals) *ZodFloatTyped[T, R] {
	return &ZodFloatTyped[T, R]{internals: &ZodFloatInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withPtrInternals creates a new instance with pointer constraint type *T.
func (z *ZodFloatTyped[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodFloatTyped[T, *T] {
	return &ZodFloatTyped[T, *T]{internals: &ZodFloatInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodFloatTyped[T, R]) CloneFrom(source any) {
	src, ok := source.(*ZodFloatTyped[T, R])
	if !ok {
		return
	}
	orig := z.internals.Checks
	*z.internals = *src.internals
	z.internals.Checks = orig
}

// floatTypeCode returns the ZodTypeCode for the given float type T.
func floatTypeCode[T FloatConstraint]() core.ZodTypeCode {
	var zero T
	switch any(zero).(type) {
	case float32, *float32:
		return core.ZodTypeFloat32
	default: // float64, *float64
		return core.ZodTypeFloat64
	}
}

// convertDefaultValue converts a float64 value to the appropriate constraint type R.
func convertDefaultValue[R any](v float64) any {
	var zero R
	switch any(zero).(type) {
	case *float32:
		return new(float32(v))
	case *float64:
		return new(v)
	case float32:
		return float32(v)
	default: // float64
		return v
	}
}

// extractFloatValue extracts the base float value T from constraint type R.
func extractFloatValue[T FloatConstraint, R any](value R) T {
	if v, ok := any(value).(T); ok {
		return v
	}

	switch v := any(value).(type) {
	case *float32:
		if v != nil {
			return any(*v).(T)
		}
	case *float64:
		if v != nil {
			return any(*v).(T)
		}
	}

	var zero T
	return zero
}

// extractFloatToFloat64 converts constraint type R to float64.
func extractFloatToFloat64[T FloatConstraint, R any](value R) float64 {
	base := extractFloatValue[T, R](value)

	switch v := any(base).(type) {
	case float32:
		return float64(v)
	default: // float64
		return v.(float64)
	}
}

// newZodFloatFromDef constructs a ZodFloat from the given definition.
func newZodFloatFromDef[T FloatConstraint, R any](def *ZodFloatDef) *ZodFloatTyped[T, R] {
	in := &ZodFloatInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		return any(newZodFloatFromDef[T, R](&ZodFloatDef{ZodTypeDef: *d})).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodFloatTyped[T, R]{internals: in}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// FloatTyped creates a generic float schema with automatic type inference.
// Usage: FloatTyped[float32](), FloatTyped[float64](), etc.
func FloatTyped[T FloatConstraint](params ...any) *ZodFloatTyped[T, T] {
	return newFloatTyped[T, T](floatTypeCode[T](), params...)
}

// Float32 creates a float32 schema.
func Float32(params ...any) *ZodFloatTyped[float32, float32] {
	return newFloatTyped[float32, float32](core.ZodTypeFloat32, params...)
}

// Float32Ptr creates a schema for *float32.
func Float32Ptr(params ...any) *ZodFloatTyped[float32, *float32] {
	return newFloatTyped[float32, *float32](core.ZodTypeFloat32, params...)
}

// Float64 creates a float64 schema.
func Float64(params ...any) *ZodFloatTyped[float64, float64] {
	return newFloatTyped[float64, float64](core.ZodTypeFloat64, params...)
}

// Float64Ptr creates a schema for *float64.
func Float64Ptr(params ...any) *ZodFloatTyped[float64, *float64] {
	return newFloatTyped[float64, *float64](core.ZodTypeFloat64, params...)
}

// Float creates a flexible float64 schema (alias for Float64).
func Float(params ...any) *ZodFloatTyped[float64, float64] {
	return Float64(params...)
}

// FloatPtr creates a schema for *float64 (alias for Float64Ptr).
func FloatPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return Float64Ptr(params...)
}

// Number creates a number schema (alias for Float64).
func Number(params ...any) *ZodFloatTyped[float64, float64] {
	return Float64(params...)
}

// NumberPtr creates a schema for *float64 (alias for Float64Ptr).
func NumberPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return Float64Ptr(params...)
}

// newFloatTyped creates a float schema with explicit type parameterization.
func newFloatTyped[T FloatConstraint, R any](typeCode core.ZodTypeCode, params ...any) *ZodFloatTyped[T, R] {
	sp := utils.NormalizeParams(params...)

	def := &ZodFloatDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeCode,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, sp)

	return newZodFloatFromDef[T, R](def)
}

// CoercedFloat32 creates a float32 schema with coercion enabled.
func CoercedFloat32(params ...any) *ZodFloatTyped[float32, float32] {
	schema := Float32(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedFloat32Ptr creates a *float32 schema with coercion enabled.
func CoercedFloat32Ptr(params ...any) *ZodFloatTyped[float32, *float32] {
	schema := Float32Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedFloat64 creates a float64 schema with coercion enabled.
func CoercedFloat64(params ...any) *ZodFloatTyped[float64, float64] {
	schema := Float64(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedFloat64Ptr creates a *float64 schema with coercion enabled.
func CoercedFloat64Ptr(params ...any) *ZodFloatTyped[float64, *float64] {
	schema := Float64Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedFloat creates a flexible float schema with coercion enabled.
func CoercedFloat[T FloatConstraint](params ...any) *ZodFloatTyped[T, T] {
	schema := FloatTyped[T](params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedFloatPtr creates a *float64 schema with coercion enabled.
func CoercedFloatPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return CoercedFloat64Ptr(params...)
}

// CoercedNumber creates a number schema with coercion enabled (alias for CoercedFloat64).
func CoercedNumber(params ...any) *ZodFloatTyped[float64, float64] {
	return CoercedFloat64(params...)
}

// CoercedNumberPtr creates a *number schema with coercion enabled (alias for CoercedFloat64Ptr).
func CoercedNumberPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return CoercedFloat64Ptr(params...)
}
