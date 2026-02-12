package types

import (
	"math"
	"math/cmplx"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// ComplexConstraint defines the supported complex types for generic schemas.
type ComplexConstraint interface {
	complex64 | complex128 | *complex64 | *complex128
}

// Complex64Constraint defines the constraint for complex64 types.
type Complex64Constraint interface {
	complex64 | *complex64
}

// Complex128Constraint defines the constraint for complex128 types.
type Complex128Constraint interface {
	complex128 | *complex128
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodComplexDef holds the schema definition for complex number validation.
type ZodComplexDef struct {
	core.ZodTypeDef
}

// ZodComplexInternals contains the internal state for a complex schema.
type ZodComplexInternals struct {
	core.ZodTypeInternals
	Def *ZodComplexDef
}

// ZodComplex is a generic complex number validation schema.
type ZodComplex[T ComplexConstraint] struct {
	internals *ZodComplexInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodComplex[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodComplex[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodComplex[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce converts input to the target complex type, implementing the Coercible interface.
func (z *ZodComplex[T]) Coerce(input any) (any, bool) {
	var zero T
	switch any(zero).(type) {
	case complex64:
		if result, err := coerce.ToComplex64(input); err == nil {
			return result, true
		}
	case *complex64:
		if result, err := coerce.ToComplex64(input); err == nil {
			return &result, true
		}
	case complex128:
		if result, err := coerce.ToComplex128(input); err == nil {
			return result, true
		}
	case *complex128:
		if result, err := coerce.ToComplex128(input); err == nil {
			return &result, true
		}
	}
	return zero, false
}

// Parse validates input and returns a value matching the generic type T.
func (z *ZodComplex[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		return engine.ParsePrimitive[complex64, T](
			input,
			&z.internals.ZodTypeInternals,
			z.expectedType(),
			z.applyChecks64,
			engine.ConvertToConstraintType[complex64, T],
			ctx...,
		)
	default:
		return engine.ParsePrimitive[complex128, T](
			input,
			&z.internals.ZodTypeInternals,
			z.expectedType(),
			z.applyChecks128,
			engine.ConvertToConstraintType[complex128, T],
			ctx...,
		)
	}
}

// MustParse validates input and panics on failure.
func (z *ZodComplex[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodComplex[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse provides compile-time type safety by requiring exact type T.
func (z *ZodComplex[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		return engine.ParsePrimitiveStrict[complex64, T](
			input,
			&z.internals.ZodTypeInternals,
			z.expectedType(),
			z.applyChecks64,
			ctx...,
		)
	default:
		return engine.ParsePrimitiveStrict[complex128, T](
			input,
			&z.internals.ZodTypeInternals,
			z.expectedType(),
			z.applyChecks128,
			ctx...,
		)
	}
}

// MustStrictParse provides compile-time type safety and panics on failure.
func (z *ZodComplex[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// applyChecks64 applies validation checks for complex64 values.
func (z *ZodComplex[T]) applyChecks64(value complex64, chks []core.ZodCheck, ctx *core.ParseContext) (complex64, error) {
	transformedValue, err := engine.ApplyChecks[any](value, chks, ctx)
	if err != nil {
		return 0, err
	}

	switch v := transformedValue.(type) {
	case complex64:
		return v, nil
	case *complex64:
		if v != nil {
			return *v, nil
		}
		return 0, nil
	case complex128:
		return complex64(v), nil
	case *complex128:
		if v != nil {
			return complex64(*v), nil
		}
		return 0, nil
	default:
		return value, nil
	}
}

// applyChecks128 applies validation checks for complex128 values.
func (z *ZodComplex[T]) applyChecks128(value complex128, chks []core.ZodCheck, ctx *core.ParseContext) (complex128, error) {
	transformedValue, err := engine.ApplyChecks[any](value, chks, ctx)
	if err != nil {
		return 0, err
	}

	switch v := transformedValue.(type) {
	case complex128:
		return v, nil
	case *complex128:
		if v != nil {
			return *v, nil
		}
		return 0, nil
	case complex64:
		return complex128(v), nil
	case *complex64:
		if v != nil {
			return complex128(*v), nil
		}
		return 0, nil
	default:
		return value, nil
	}
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a new schema that accepts nil, with *complex128 constraint.
func (z *ZodComplex[T]) Optional() *ZodComplex[*complex128] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withComplex128PtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional, ExactOptional only accepts absent keys in object fields.
func (z *ZodComplex[T]) ExactOptional() *ZodComplex[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a new schema that accepts nil values, with *complex128 constraint.
func (z *ZodComplex[T]) Nilable() *ZodComplex[*complex128] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withComplex128PtrInternals(in)
}

// Nullish returns a new schema combining optional and nilable modifiers.
func (z *ZodComplex[T]) Nullish() *ZodComplex[*complex128] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withComplex128PtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodComplex[T]) Default(v complex128) *ZodComplex[T] {
	in := z.internals.Clone()
	var zero T
	switch any(zero).(type) {
	case *complex64:
		val := complex64(v)
		in.SetDefaultValue(&val)
	case *complex128:
		in.SetDefaultValue(&v)
	case complex64:
		in.SetDefaultValue(complex64(v))
	default:
		in.SetDefaultValue(v)
	}
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodComplex[T]) DefaultFunc(fn func() complex128) *ZodComplex[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		v := fn()
		var zero T
		switch any(zero).(type) {
		case *complex64:
			val := complex64(v)
			return &val
		case *complex128:
			return &v
		case complex64:
			return complex64(v)
		default:
			return v
		}
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodComplex[T]) Prefault(v complex128) *ZodComplex[T] {
	in := z.internals.Clone()
	var zero T
	switch any(zero).(type) {
	case *complex64:
		val := complex64(v)
		in.SetPrefaultValue(&val)
	case *complex128:
		in.SetPrefaultValue(&v)
	case complex64:
		in.SetPrefaultValue(complex64(v))
	default:
		in.SetPrefaultValue(v)
	}
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodComplex[T]) PrefaultFunc(fn func() complex128) *ZodComplex[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		v := fn()
		var zero T
		switch any(zero).(type) {
		case *complex64:
			val := complex64(v)
			return &val
		case *complex128:
			return &v
		case complex64:
			return complex64(v)
		default:
			return v
		}
	})
	return z.withInternals(in)
}

// Meta stores metadata for this complex number schema.
func (z *ZodComplex[T]) Meta(meta core.GlobalMeta) *ZodComplex[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodComplex[T]) Describe(description string) *ZodComplex[T] {
	newInternals := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(newInternals)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// VALIDATION METHODS (USING CHECKS PACKAGE)
// =============================================================================

// Min adds minimum magnitude validation for complex numbers.
func (z *ZodComplex[T]) Min(minimum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) >= minimum
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// Max adds maximum magnitude validation for complex numbers.
func (z *ZodComplex[T]) Max(maximum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) <= maximum
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// Gt adds greater-than magnitude validation (exclusive).
func (z *ZodComplex[T]) Gt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) > value
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// Gte adds greater-than-or-equal magnitude validation (inclusive).
func (z *ZodComplex[T]) Gte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) >= value
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// Lt adds less-than magnitude validation (exclusive).
func (z *ZodComplex[T]) Lt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) < value
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// Lte adds less-than-or-equal magnitude validation (inclusive).
func (z *ZodComplex[T]) Lte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) <= value
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// Positive adds positive magnitude validation (> 0).
func (z *ZodComplex[T]) Positive(params ...any) *ZodComplex[T] {
	return z.Gt(0, params...)
}

// Negative adds negative validation (< 0) for complex magnitude.
func (z *ZodComplex[T]) Negative(params ...any) *ZodComplex[T] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative validation (>= 0) for complex magnitude.
func (z *ZodComplex[T]) NonNegative(params ...any) *ZodComplex[T] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive validation (<= 0) for complex magnitude.
func (z *ZodComplex[T]) NonPositive(params ...any) *ZodComplex[T] {
	return z.Lte(0, params...)
}

// Finite validates that both real and imaginary parts are finite (no Inf or NaN).
func (z *ZodComplex[T]) Finite(params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return !math.IsInf(real(*val), 0) && !math.IsInf(imag(*val), 0) &&
				!math.IsNaN(real(*val)) && !math.IsNaN(imag(*val))
		}
		return false
	}, params...)
	return z.withCheck(check)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation pipeline.
func (z *ZodComplex[T]) Transform(fn func(complex128, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractComplex128(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodComplex[T]) Overwrite(transform func(T) T, params ...any) *ZodComplex[T] {
	transformAny := func(input any) any {
		converted, ok := convertToComplexType[T](input)
		if !ok {
			return input
		}
		return transform(converted)
	}
	check := checks.NewZodCheckOverwrite(transformAny, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodComplex[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	targetFn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractComplex128(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodComplex[T]) Refine(fn func(T) bool, params ...any) *ZodComplex[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case complex64:
			if v == nil {
				return false
			}
			if complexVal, ok := v.(complex64); ok {
				return fn(any(complexVal).(T))
			}
			return false
		case *complex64:
			if v == nil {
				return fn(any((*complex64)(nil)).(T))
			}
			if complexVal, ok := v.(complex64); ok {
				cCopy := complexVal
				return fn(any(&cCopy).(T))
			}
			return false
		case complex128:
			if v == nil {
				return false
			}
			if complexVal, ok := v.(complex128); ok {
				return fn(any(complexVal).(T))
			}
			return false
		case *complex128:
			if v == nil {
				return fn(any((*complex128)(nil)).(T))
			}
			if complexVal, ok := v.(complex128); ok {
				cCopy := complexVal
				return fn(any(&cCopy).(T))
			}
			return false
		default:
			return false
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodComplex[T]) RefineAny(fn func(any) bool, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodComplex[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodComplex[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// expectedType returns the schema's type code, with fallback based on T.
func (z *ZodComplex[T]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		return core.ZodTypeComplex64
	default:
		return core.ZodTypeComplex128
	}
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodComplex[T]) withCheck(check core.ZodCheck) *ZodComplex[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withComplex128PtrInternals creates a new *complex128 schema from cloned internals.
func (z *ZodComplex[T]) withComplex128PtrInternals(in *core.ZodTypeInternals) *ZodComplex[*complex128] {
	return &ZodComplex[*complex128]{internals: &ZodComplexInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodComplex[T]) withInternals(in *core.ZodTypeInternals) *ZodComplex[T] {
	return &ZodComplex[T]{internals: &ZodComplexInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodComplex[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodComplex[T]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractComplex128 extracts the underlying complex128 from a generic constraint type.
func extractComplex128[T ComplexConstraint](value T) complex128 {
	switch v := any(value).(type) {
	case complex64:
		return complex128(v)
	case *complex64:
		if v != nil {
			return complex128(*v)
		}
		return 0
	case complex128:
		return v
	case *complex128:
		if v != nil {
			return *v
		}
		return 0
	default:
		return 0
	}
}

// convertToComplexValue extracts a complex128 pointer from any complex type.
func convertToComplexValue(v any) *complex128 {
	switch val := v.(type) {
	case complex64:
		result := complex128(val)
		return &result
	case *complex64:
		if val != nil {
			result := complex128(*val)
			return &result
		}
		return nil
	case complex128:
		return &val
	case *complex128:
		if val != nil {
			return val
		}
		return nil
	default:
		return nil
	}
}

// convertToComplexType converts any value to the target complex constraint type T.
func convertToComplexType[T ComplexConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *complex64, *complex128:
			return zero, true
		default:
			return zero, false
		}
	}

	complexValue := convertToComplexValue(v)
	if complexValue == nil {
		return zero, false
	}

	switch any(zero).(type) {
	case complex64:
		return any(complex64(*complexValue)).(T), true
	case *complex64:
		c64 := complex64(*complexValue)
		return any(&c64).(T), true
	case complex128:
		return any(*complexValue).(T), true
	case *complex128:
		return any(complexValue).(T), true
	default:
		return zero, false
	}
}

// newZodComplexFromDef constructs a new ZodComplex from a definition.
func newZodComplexFromDef[T ComplexConstraint](def *ZodComplexDef) *ZodComplex[T] {
	internals := &ZodComplexInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		complexDef := &ZodComplexDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodComplexFromDef[T](complexDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodComplex[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// ComplexTyped creates a generic complex schema with automatic type inference.
func ComplexTyped[T ComplexConstraint](params ...any) *ZodComplex[T] {
	var typeCode core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		typeCode = core.ZodTypeComplex64
	default:
		typeCode = core.ZodTypeComplex128
	}
	return newComplexTyped[T](typeCode, params...)
}

// Complex64 creates a complex64 schema.
func Complex64(params ...any) *ZodComplex[complex64] {
	return newComplexTyped[complex64](core.ZodTypeComplex64, params...)
}

// Complex64Ptr creates a *complex64 schema.
func Complex64Ptr(params ...any) *ZodComplex[*complex64] {
	return newComplexTyped[*complex64](core.ZodTypeComplex64, params...)
}

// Complex128 creates a complex128 schema.
func Complex128(params ...any) *ZodComplex[complex128] {
	return newComplexTyped[complex128](core.ZodTypeComplex128, params...)
}

// Complex128Ptr creates a *complex128 schema.
func Complex128Ptr(params ...any) *ZodComplex[*complex128] {
	return newComplexTyped[*complex128](core.ZodTypeComplex128, params...)
}

// Complex creates a complex128 schema (default alias).
func Complex(params ...any) *ZodComplex[complex128] {
	return Complex128(params...)
}

// ComplexPtr creates a *complex128 schema (default alias).
func ComplexPtr(params ...any) *ZodComplex[*complex128] {
	return Complex128Ptr(params...)
}

// newComplexTyped creates a complex schema with explicit type parameterization.
func newComplexTyped[T ComplexConstraint](typeCode core.ZodTypeCode, params ...any) *ZodComplex[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodComplexDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeCode,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodComplexFromDef[T](def)
}

// CoercedComplex creates a coerced complex128 schema (default alias).
func CoercedComplex(params ...any) *ZodComplex[complex128] {
	return CoercedComplex128(params...)
}

// CoercedComplexPtr creates a coerced *complex128 schema (default alias).
func CoercedComplexPtr(params ...any) *ZodComplex[*complex128] {
	return CoercedComplex128Ptr(params...)
}

// CoercedComplex64 creates a coerced complex64 schema.
func CoercedComplex64(params ...any) *ZodComplex[complex64] {
	schema := Complex64(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedComplex64Ptr creates a coerced *complex64 schema.
func CoercedComplex64Ptr(params ...any) *ZodComplex[*complex64] {
	schema := Complex64Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedComplex128 creates a coerced complex128 schema.
func CoercedComplex128(params ...any) *ZodComplex[complex128] {
	schema := Complex128(params...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedComplex128Ptr creates a coerced *complex128 schema.
func CoercedComplex128Ptr(params ...any) *ZodComplex[*complex128] {
	schema := Complex128Ptr(params...)
	schema.internals.SetCoerce(true)
	return schema
}
