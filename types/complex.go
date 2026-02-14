// Package types provides validation schemas for Go types.
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

// Type constraints for complex number validation.

// ComplexConstraint restricts values to complex64, complex128, or their pointers.
type ComplexConstraint interface {
	complex64 | complex128 | *complex64 | *complex128
}

// Complex64Constraint restricts values to complex64 or *complex64.
type Complex64Constraint interface {
	complex64 | *complex64
}

// Complex128Constraint restricts values to complex128 or *complex128.
type Complex128Constraint interface {
	complex128 | *complex128
}

// Schema definition and internal structures.

// ZodComplexDef holds the configuration for complex number validation.
type ZodComplexDef struct {
	core.ZodTypeDef
}

// ZodComplexInternals contains the internal state of a complex validator.
type ZodComplexInternals struct {
	core.ZodTypeInternals
	Def *ZodComplexDef
}

// ZodComplex is a type-safe complex number validation schema.
type ZodComplex[T ComplexConstraint] struct {
	internals *ZodComplexInternals
}

// Core schema methods.

// Internals returns the internal state of the schema.
func (z *ZodComplex[T]) Internals() *core.ZodTypeInternals {
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

// Parsing methods.

// Coerce converts input to the target complex type.
func (z *ZodComplex[T]) Coerce(input any) (any, bool) {
	var zero T
	switch any(zero).(type) {
	case complex64:
		if r, err := coerce.ToComplex64(input); err == nil {
			return r, true
		}
	case *complex64:
		if r, err := coerce.ToComplex64(input); err == nil {
			return &r, true
		}
	case complex128:
		if r, err := coerce.ToComplex128(input); err == nil {
			return r, true
		}
	case *complex128:
		if r, err := coerce.ToComplex128(input); err == nil {
			return &r, true
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

// MustParse panics on validation failure.
func (z *ZodComplex[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodComplex[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// StrictParse validates input with compile-time type safety.
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

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodComplex[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// Validation check application methods.

// applyChecks64 applies validation checks for complex64 values.
func (z *ZodComplex[T]) applyChecks64(val complex64, chks []core.ZodCheck, ctx *core.ParseContext) (complex64, error) {
	out, err := engine.ApplyChecks[any](val, chks, ctx)
	if err != nil {
		return 0, err
	}
	switch v := out.(type) {
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
		return val, nil
	}
}

// applyChecks128 applies validation checks for complex128 values.
func (z *ZodComplex[T]) applyChecks128(val complex128, chks []core.ZodCheck, ctx *core.ParseContext) (complex128, error) {
	out, err := engine.ApplyChecks[any](val, chks, ctx)
	if err != nil {
		return 0, err
	}
	switch v := out.(type) {
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
		return val, nil
	}
}

// Modifier methods.

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodComplex[T]) Optional() *ZodComplex[*complex128] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodComplex[T]) ExactOptional() *ZodComplex[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodComplex[T]) Nilable() *ZodComplex[*complex128] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodComplex[T]) Nullish() *ZodComplex[*complex128] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default and Prefault methods.

// Default sets a fallback value returned when input is nil (short-circuits).
func (z *ZodComplex[T]) Default(v complex128) *ZodComplex[T] {
	in := z.internals.Clone()
	var zero T
	switch any(zero).(type) {
	case *complex64:
		in.SetDefaultValue(new(complex64(v)))
	case *complex128:
		in.SetDefaultValue(new(v))
	case complex64:
		in.SetDefaultValue(complex64(v))
	default:
		in.SetDefaultValue(v)
	}
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits).
func (z *ZodComplex[T]) DefaultFunc(fn func() complex128) *ZodComplex[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		v := fn()
		var zero T
		switch any(zero).(type) {
		case *complex64:
			return new(complex64(v))
		case *complex128:
			return new(v)
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
		in.SetPrefaultValue(new(complex64(v)))
	case *complex128:
		in.SetPrefaultValue(new(v))
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
			return new(complex64(v))
		case *complex128:
			return new(v)
		case complex64:
			return complex64(v)
		default:
			return v
		}
	})
	return z.withInternals(in)
}

// Metadata methods.

// Meta stores metadata in the global registry.
func (z *ZodComplex[T]) Meta(meta core.GlobalMeta) *ZodComplex[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodComplex[T]) Describe(desc string) *ZodComplex[T] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = desc
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// Validation constraint methods.

// Min adds minimum magnitude validation.
func (z *ZodComplex[T]) Min(minimum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return cmplx.Abs(*c) >= minimum
	}, params...)
	return z.withCheck(check)
}

// Max adds maximum magnitude validation.
func (z *ZodComplex[T]) Max(maximum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return cmplx.Abs(*c) <= maximum
	}, params...)
	return z.withCheck(check)
}

// Gt adds greater-than magnitude validation (exclusive).
func (z *ZodComplex[T]) Gt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return cmplx.Abs(*c) > value
	}, params...)
	return z.withCheck(check)
}

// Gte adds greater-than-or-equal magnitude validation (inclusive).
func (z *ZodComplex[T]) Gte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return cmplx.Abs(*c) >= value
	}, params...)
	return z.withCheck(check)
}

// Lt adds less-than magnitude validation (exclusive).
func (z *ZodComplex[T]) Lt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return cmplx.Abs(*c) < value
	}, params...)
	return z.withCheck(check)
}

// Lte adds less-than-or-equal magnitude validation (inclusive).
func (z *ZodComplex[T]) Lte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return cmplx.Abs(*c) <= value
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

// Finite validates that both real and imaginary parts are finite.
func (z *ZodComplex[T]) Finite(params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		c := toComplex128(v)
		if c == nil {
			return false
		}
		return !math.IsInf(real(*c), 0) && !math.IsInf(imag(*c), 0) &&
			!math.IsNaN(real(*c)) && !math.IsNaN(imag(*c))
	}, params...)
	return z.withCheck(check)
}

// Transformation and composition methods.

// Transform applies a transformation function to the parsed value.
func (z *ZodComplex[T]) Transform(fn func(complex128, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapper := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractComplex128(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapper)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodComplex[T]) Overwrite(transform func(T) T, params ...any) *ZodComplex[T] {
	fn := func(input any) any {
		v, ok := toComplexType[T](input)
		if !ok {
			return input
		}
		return transform(v)
	}
	check := checks.NewZodCheckOverwrite(fn, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodComplex[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	fn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractComplex128(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, fn)
}

// Refinement methods.

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodComplex[T]) Refine(fn func(T) bool, params ...any) *ZodComplex[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case complex64:
			if v == nil {
				return false
			}
			c, ok := v.(complex64)
			if !ok {
				return false
			}
			return fn(any(c).(T))
		case *complex64:
			if v == nil {
				return fn(any((*complex64)(nil)).(T))
			}
			c, ok := v.(complex64)
			if !ok {
				return false
			}
			return fn(any(&c).(T))
		case complex128:
			if v == nil {
				return false
			}
			c, ok := v.(complex128)
			if !ok {
				return false
			}
			return fn(any(c).(T))
		case *complex128:
			if v == nil {
				return fn(any((*complex128)(nil)).(T))
			}
			c, ok := v.(complex128)
			if !ok {
				return false
			}
			return fn(any(&c).(T))
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

// Composition methods.

// And creates an intersection with another schema.
func (z *ZodComplex[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodComplex[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// Internal helper methods.

// expectedType returns the schema's type code, defaulting based on T.
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

// withCheck clones internals, adds a check, and returns a new schema.
func (z *ZodComplex[T]) withCheck(check core.ZodCheck) *ZodComplex[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new *complex128 schema from cloned internals.
func (z *ZodComplex[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodComplex[*complex128] {
	return &ZodComplex[*complex128]{
		internals: &ZodComplexInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodComplex[T]) withInternals(in *core.ZodTypeInternals) *ZodComplex[T] {
	return &ZodComplex[T]{
		internals: &ZodComplexInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodComplex[T]) CloneFrom(source any) {
	src, ok := source.(*ZodComplex[T])
	if !ok {
		return
	}
	orig := z.internals.Checks
	*z.internals = *src.internals
	z.internals.Checks = orig
}

// Helper functions for type conversion.

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

// toComplex128 extracts a complex128 pointer from any complex type.
func toComplex128(v any) *complex128 {
	switch val := v.(type) {
	case complex64:
		return new(complex128(val))
	case *complex64:
		if val != nil {
			return new(complex128(*val))
		}
		return nil
	case complex128:
		return new(val)
	case *complex128:
		if val != nil {
			return val
		}
		return nil
	default:
		return nil
	}
}

// toComplexType converts any value to the target complex constraint type T.
func toComplexType[T ComplexConstraint](v any) (T, bool) {
	var zero T
	if v == nil {
		switch any(zero).(type) {
		case *complex64, *complex128:
			return zero, true
		default:
			return zero, false
		}
	}
	c := toComplex128(v)
	if c == nil {
		return zero, false
	}
	switch any(zero).(type) {
	case complex64:
		return any(complex64(*c)).(T), true
	case *complex64:
		return any(new(complex64(*c))).(T), true
	case complex128:
		return any(*c).(T), true
	case *complex128:
		return any(c).(T), true
	default:
		return zero, false
	}
}

// Schema constructor functions.

// newZodComplexFromDef constructs a new ZodComplex from a definition.
func newZodComplexFromDef[T ComplexConstraint](def *ZodComplexDef) *ZodComplex[T] {
	in := &ZodComplexInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		cd := &ZodComplexDef{ZodTypeDef: *d}
		return any(newZodComplexFromDef[T](cd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodComplex[T]{internals: in}
}

// Public constructor functions.

// ComplexTyped creates a generic complex schema with automatic type inference.
func ComplexTyped[T ComplexConstraint](params ...any) *ZodComplex[T] {
	var tc core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		tc = core.ZodTypeComplex64
	default:
		tc = core.ZodTypeComplex128
	}
	return newComplexTyped[T](tc, params...)
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
func newComplexTyped[T ComplexConstraint](tc core.ZodTypeCode, params ...any) *ZodComplex[T] {
	sp := utils.NormalizeParams(params...)
	def := &ZodComplexDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   tc,
			Checks: []core.ZodCheck{},
		},
	}
	utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	return newZodComplexFromDef[T](def)
}

// Coerced constructor functions.

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
	s := Complex64(params...)
	s.internals.SetCoerce(true)
	return s
}

// CoercedComplex64Ptr creates a coerced *complex64 schema.
func CoercedComplex64Ptr(params ...any) *ZodComplex[*complex64] {
	s := Complex64Ptr(params...)
	s.internals.SetCoerce(true)
	return s
}

// CoercedComplex128 creates a coerced complex128 schema.
func CoercedComplex128(params ...any) *ZodComplex[complex128] {
	s := Complex128(params...)
	s.internals.SetCoerce(true)
	return s
}

// CoercedComplex128Ptr creates a coerced *complex128 schema.
func CoercedComplex128Ptr(params ...any) *ZodComplex[*complex128] {
	s := Complex128Ptr(params...)
	s.internals.SetCoerce(true)
	return s
}
