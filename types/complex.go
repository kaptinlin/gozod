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

// ComplexConstraint defines supported complex types for generic implementation
type ComplexConstraint interface {
	complex64 | complex128 | *complex64 | *complex128
}

// Complex64Constraint defines constraint for complex64 types
type Complex64Constraint interface {
	complex64 | *complex64
}

// Complex128Constraint defines constraint for complex128 types
type Complex128Constraint interface {
	complex128 | *complex128
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodComplexDef defines the schema definition for complex number validation
type ZodComplexDef struct {
	core.ZodTypeDef
}

// ZodComplexInternals contains the internal state for complex schema
type ZodComplexInternals struct {
	core.ZodTypeInternals
	Def *ZodComplexDef // Schema definition reference
}

// ZodComplex represents a generic complex number validation schema
type ZodComplex[T ComplexConstraint] struct {
	internals *ZodComplexInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodComplex[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodComplex[T]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodComplex[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce attempts to coerce input to target complex type using coerce package
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
	return *new(T), false
}

// Parse validates input using unified ParsePrimitive API
func (z *ZodComplex[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		return engine.ParsePrimitive[complex64, T](
			input,
			&z.internals.ZodTypeInternals,
			core.ZodTypeComplex64,
			z.applyComplexChecks64,
			engine.ConvertToConstraintType[complex64, T],
			ctx...,
		)
	case complex128, *complex128:
		return engine.ParsePrimitive[complex128, T](
			input,
			&z.internals.ZodTypeInternals,
			core.ZodTypeComplex128,
			z.applyComplexChecks128,
			engine.ConvertToConstraintType[complex128, T],
			ctx...,
		)
	default:
		// Default to complex128
		return engine.ParsePrimitive[complex128, T](
			input,
			&z.internals.ZodTypeInternals,
			core.ZodTypeComplex128,
			z.applyComplexChecks128,
			engine.ConvertToConstraintType[complex128, T],
			ctx...,
		)
	}
}

// MustParse validates the input value and panics on failure
func (z *ZodComplex[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodComplex[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// applyComplexChecks64 handles ApplyChecks for complex64 with pointer type support
func (z *ZodComplex[T]) applyComplexChecks64(value complex64, checks []core.ZodCheck, ctx *core.ParseContext) (complex64, error) {
	transformedValue, err := engine.ApplyChecks[any](value, checks, ctx)
	if err != nil {
		return 0, err
	}

	// Handle potential pointer type from Overwrite transformations
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
		// If transformation returned unexpected type, return original result
		return value, nil
	}
}

// applyComplexChecks128 handles ApplyChecks for complex128 with pointer type support
func (z *ZodComplex[T]) applyComplexChecks128(value complex128, checks []core.ZodCheck, ctx *core.ParseContext) (complex128, error) {
	transformedValue, err := engine.ApplyChecks[any](value, checks, ctx)
	if err != nil {
		return 0, err
	}

	// Handle potential pointer type from Overwrite transformations
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
		// If transformation returned unexpected type, return original result
		return value, nil
	}
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *complex128 for nullable semantics
func (z *ZodComplex[T]) Optional() *ZodComplex[*complex128] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withComplex128PtrInternals(in)
}

// Nilable allows nil values, returns pointer type
func (z *ZodComplex[T]) Nilable() *ZodComplex[*complex128] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withComplex128PtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodComplex[T]) Nullish() *ZodComplex[*complex128] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withComplex128PtrInternals(in)
}

// Default preserves current generic type T
func (z *ZodComplex[T]) Default(v complex128) *ZodComplex[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current generic type T
func (z *ZodComplex[T]) DefaultFunc(fn func() complex128) *ZodComplex[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodComplex[T]) Prefault(v complex128) *ZodComplex[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type T.
func (z *ZodComplex[T]) PrefaultFunc(fn func() complex128) *ZodComplex[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this complex number schema.
func (z *ZodComplex[T]) Meta(meta core.GlobalMeta) *ZodComplex[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// VALIDATION METHODS (USING CHECKS PACKAGE)
// =============================================================================

// Min adds minimum magnitude validation for complex numbers
func (z *ZodComplex[T]) Min(minimum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) >= minimum
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max adds maximum magnitude validation for complex numbers
func (z *ZodComplex[T]) Max(maximum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) <= maximum
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gt adds greater than validation for complex magnitude (exclusive)
func (z *ZodComplex[T]) Gt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) > value
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gte adds greater than or equal validation for complex magnitude (inclusive)
func (z *ZodComplex[T]) Gte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) >= value
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lt adds less than validation for complex magnitude (exclusive)
func (z *ZodComplex[T]) Lt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) < value
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lte adds less than or equal validation for complex magnitude (inclusive)
func (z *ZodComplex[T]) Lte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return cmplx.Abs(*val) <= value
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Positive adds positive magnitude validation (> 0) for complex numbers
func (z *ZodComplex[T]) Positive(params ...any) *ZodComplex[T] {
	return z.Gt(0, params...)
}

// Negative adds negative validation (< 0) - validates real part for complex numbers
func (z *ZodComplex[T]) Negative(params ...any) *ZodComplex[T] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative validation (>= 0) for complex magnitude
func (z *ZodComplex[T]) NonNegative(params ...any) *ZodComplex[T] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive validation (<= 0) for complex magnitude
func (z *ZodComplex[T]) NonPositive(params ...any) *ZodComplex[T] {
	return z.Lte(0, params...)
}

// Finite adds finite validation for complex numbers (no infinite components)
func (z *ZodComplex[T]) Finite(params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val := convertToComplexValue(v); val != nil {
			return !math.IsInf(real(*val), 0) && !math.IsInf(imag(*val), 0) &&
				!math.IsNaN(real(*val)) && !math.IsNaN(imag(*val))
		}
		return false
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using the WrapFn pattern
func (z *ZodComplex[T]) Transform(fn func(complex128, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		complexVal := extractComplex128(input)
		return fn(complexVal, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodComplex[T]) Overwrite(transform func(T) T, params ...any) *ZodComplex[T] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToComplexType[T](input)
		if !ok {
			// If conversion fails, return original value
			return input
		}

		// Apply transformation directly on type T
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using the WrapFn pattern
func (z *ZodComplex[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		complexVal := extractComplex128(input)
		return target.Parse(complexVal, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with automatic type conversion
func (z *ZodComplex[T]) Refine(fn func(T) bool, params ...any) *ZodComplex[T] {
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case complex64:
			// Schema output is complex64
			if v == nil {
				return false
			}
			if complexVal, ok := v.(complex64); ok {
				return fn(any(complexVal).(T))
			}
			return false
		case *complex64:
			// Schema output is *complex64 – convert incoming value to *complex64
			if v == nil {
				return fn(any((*complex64)(nil)).(T))
			}
			if complexVal, ok := v.(complex64); ok {
				cCopy := complexVal
				ptr := &cCopy
				return fn(any(ptr).(T))
			}
			return false
		case complex128:
			// Schema output is complex128
			if v == nil {
				return false
			}
			if complexVal, ok := v.(complex128); ok {
				return fn(any(complexVal).(T))
			}
			return false
		case *complex128:
			// Schema output is *complex128 – convert incoming value to *complex128
			if v == nil {
				return fn(any((*complex128)(nil)).(T))
			}
			if complexVal, ok := v.(complex128); ok {
				cCopy := complexVal
				ptr := &cCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			return false
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodComplex[T]) RefineAny(fn func(any) bool, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withComplex128PtrInternals creates new instance with *complex128 type
func (z *ZodComplex[T]) withComplex128PtrInternals(in *core.ZodTypeInternals) *ZodComplex[*complex128] {
	return &ZodComplex[*complex128]{internals: &ZodComplexInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates new instance preserving generic type T
func (z *ZodComplex[T]) withInternals(in *core.ZodTypeInternals) *ZodComplex[T] {
	return &ZodComplex[T]{internals: &ZodComplexInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodComplex[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodComplex[T]); ok {
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// extractComplex128 extracts complex128 value from generic type T
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

// convertToComplexValue extracts complex128 value from any complex type
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

// convertToComplexType converts any value to the complex constraint type T with strict type checking
func convertToComplexType[T ComplexConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		// Handle nil values for pointer types
		switch any(zero).(type) {
		case *complex64, *complex128:
			return zero, true // zero value for pointer types is nil
		default:
			return zero, false // nil not allowed for value types
		}
	}

	// Extract complex value from input
	var complexValue *complex128 = convertToComplexValue(v)
	if complexValue == nil {
		return zero, false // Reject all non-complex types
	}

	// Convert to target type T
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

// newZodComplexFromDef constructs new ZodComplex from definition
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

	// Provide constructor for AddCheck functionality
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
// It automatically determines the appropriate type code based on the generic type parameter.
// Usage: ComplexTyped[complex64](), ComplexTyped[complex128](), ComplexTyped[*complex64](), etc.
func ComplexTyped[T ComplexConstraint](params ...any) *ZodComplex[T] {
	// Determine type code based on T
	var typeCode core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case complex64, *complex64:
		typeCode = core.ZodTypeComplex64
	case complex128, *complex128:
		typeCode = core.ZodTypeComplex128
	default:
		typeCode = core.ZodTypeComplex128 // default to complex128
	}

	return newComplexTyped[T](typeCode, params...)
}

// Complex64 creates complex64 schema
func Complex64(params ...any) *ZodComplex[complex64] {
	return newComplexTyped[complex64](core.ZodTypeComplex64, params...)
}

// Complex64Ptr creates schema for *complex64
func Complex64Ptr(params ...any) *ZodComplex[*complex64] {
	return newComplexTyped[*complex64](core.ZodTypeComplex64, params...)
}

// Complex128 creates complex128 schema
func Complex128(params ...any) *ZodComplex[complex128] {
	return newComplexTyped[complex128](core.ZodTypeComplex128, params...)
}

// Complex128Ptr creates schema for *complex128
func Complex128Ptr(params ...any) *ZodComplex[*complex128] {
	return newComplexTyped[*complex128](core.ZodTypeComplex128, params...)
}

// Complex creates complex128 schema (default)
func Complex(params ...any) *ZodComplex[complex128] {
	return Complex128(params...)
}

// ComplexPtr creates schema for *complex128 (default)
func ComplexPtr(params ...any) *ZodComplex[*complex128] {
	return Complex128Ptr(params...)
}

// newComplexTyped is the underlying generic function for creating complex schemas,
// allowing for explicit type parameterization. This is an internal function.
func newComplexTyped[T ComplexConstraint](typeCode core.ZodTypeCode, params ...any) *ZodComplex[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodComplexDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeCode,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply the normalized parameters to the schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodComplexFromDef[T](def)
}

// CoercedComplex creates coerced complex128 schema (default)
func CoercedComplex(params ...any) *ZodComplex[complex128] {
	return CoercedComplex128(params...)
}

// CoercedComplexPtr creates coerced *complex128 schema (default)
func CoercedComplexPtr(params ...any) *ZodComplex[*complex128] {
	return CoercedComplex128Ptr(params...)
}

// CoercedComplex64 creates coerced complex64 schema
func CoercedComplex64(params ...any) *ZodComplex[complex64] {
	schema := Complex64(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedComplex64Ptr creates coerced *complex64 schema
func CoercedComplex64Ptr(params ...any) *ZodComplex[*complex64] {
	schema := Complex64Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedComplex128 creates coerced complex128 schema
func CoercedComplex128(params ...any) *ZodComplex[complex128] {
	schema := Complex128(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedComplex128Ptr creates coerced *complex128 schema
func CoercedComplex128Ptr(params ...any) *ZodComplex[*complex128] {
	schema := Complex128Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}
