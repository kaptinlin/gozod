package types

import (
	"errors"
	"fmt"
	"math/cmplx"
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"

	"github.com/kaptinlin/gozod/internal/issues"
)

// Error definitions for complex transformations
var (
	ErrTransformNilComplex  = errors.New("cannot transform nil complex value")
	ErrExpectedComplex      = errors.New("expected complex type")
	ErrInvalidComplexFormat = errors.New("invalid complex number format")
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodComplexConstraint defines supported complex types for generic implementation
type ZodComplexConstraint interface {
	~complex64 | ~complex128
}

// ZodComplexDef defines the configuration for generic complex number validation
type ZodComplexDef[T ZodComplexConstraint] struct {
	core.ZodTypeDef
	Type core.ZodTypeCode // Type identifier using type-safe constants
}

// ZodComplexInternals contains generic complex number validator internal state
type ZodComplexInternals[T ZodComplexConstraint] struct {
	core.ZodTypeInternals
	Def     *ZodComplexDef[T] // Schema definition
	Pattern *regexp.Regexp    // Complex pattern (if any)
	Bag     map[string]any    // Additional metadata (coerce flag, etc.)
}

// ZodComplex represents a generic complex number validation schema
type ZodComplex[T ZodComplexConstraint] struct {
	internals *ZodComplexInternals[T]
}

// GetInternals returns the internal state of the schema
func (z *ZodComplex[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the complex-specific internals for framework usage
func (z *ZodComplex[T]) GetZod() *ZodComplexInternals[T] {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodComplex[T]) CloneFrom(source any) {
	if src, ok := source.(interface {
		GetZod() *ZodComplexInternals[T]
	}); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		// Copy Bag state (includes coercion flags, etc.)
		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]any)
			}
			for key, value := range srcState.Bag {
				tgtState.Bag[key] = value
			}
		}

		// Copy Pattern state
		if srcState.Pattern != nil {
			tgtState.Pattern = srcState.Pattern
		}
	}
}

// Coerce attempts to coerce input to target complex type using coerce package
func (z *ZodComplex[T]) Coerce(input any) (any, bool) {
	// Use existing coerce functions based on type
	var zero T
	switch any(zero).(type) {
	case complex64:
		if result, err := coerce.ToComplex64(input); err == nil {
			return T(result), true
		}
	case complex128:
		if result, err := coerce.ToComplex128(input); err == nil {
			return T(result), true
		}
	}
	return *new(T), false
}

// Parse validates input with smart type inference using engine.ParseType
func (z *ZodComplex[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// Determine type code from generic parameter
	var typeCode core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case complex64:
		typeCode = core.ZodTypeComplex64
	case complex128:
		typeCode = core.ZodTypeComplex128
	default:
		typeCode = core.ZodTypeComplex128 // Default to complex128
	}

	return engine.ParsePrimitive[T](
		input,
		&z.internals.ZodTypeInternals,
		typeCode,
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) error {
			// Run attached checks, if any
			if len(checks) == 0 {
				return nil
			}
			payload := &core.ParsePayload{
				Value:  value,
				Issues: make([]core.ZodRawIssue, 0),
			}
			engine.RunChecksOnValue(value, checks, payload, ctx)
			if len(payload.Issues) > 0 {
				finalized := make([]core.ZodIssue, len(payload.Issues))
				for i, raw := range payload.Issues {
					finalized[i] = issues.FinalizeIssue(raw, ctx, core.GetConfig())
				}
				return issues.NewZodError(finalized)
			}
			return nil
		},
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodComplex[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Min adds minimum magnitude validation for complex numbers
func (z *ZodComplex[T]) Min(minimum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val, ok := v.(T); ok {
			return cmplx.Abs(complex128(val)) >= minimum
		}
		// support pointer
		if ptr, ok := v.(*T); ok && ptr != nil {
			return cmplx.Abs(complex128(*ptr)) >= minimum
		}
		return false
	}, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Max adds maximum magnitude validation for complex numbers
func (z *ZodComplex[T]) Max(maximum float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val, ok := v.(T); ok {
			return cmplx.Abs(complex128(val)) <= maximum
		}
		if ptr, ok := v.(*T); ok && ptr != nil {
			return cmplx.Abs(complex128(*ptr)) <= maximum
		}
		return false
	}, params...)
	result := engine.AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Gt adds greater than validation for complex magnitude (exclusive)
func (z *ZodComplex[T]) Gt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val, ok := v.(T); ok {
			return cmplx.Abs(complex128(val)) > value
		}
		if ptr, ok := v.(*T); ok && ptr != nil {
			return cmplx.Abs(complex128(*ptr)) > value
		}
		return false
	}, params...)
	return engine.AddCheck(z, check).(*ZodComplex[T])
}

// Gte adds greater than or equal validation for complex magnitude (inclusive)
func (z *ZodComplex[T]) Gte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val, ok := v.(T); ok {
			return cmplx.Abs(complex128(val)) >= value
		}
		if ptr, ok := v.(*T); ok && ptr != nil {
			return cmplx.Abs(complex128(*ptr)) >= value
		}
		return false
	}, params...)
	return engine.AddCheck(z, check).(*ZodComplex[T])
}

// Lt adds less than validation for complex magnitude (exclusive)
func (z *ZodComplex[T]) Lt(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val, ok := v.(T); ok {
			return cmplx.Abs(complex128(val)) < value
		}
		if ptr, ok := v.(*T); ok && ptr != nil {
			return cmplx.Abs(complex128(*ptr)) < value
		}
		return false
	}, params...)
	return engine.AddCheck(z, check).(*ZodComplex[T])
}

// Lte adds less than or equal validation for complex magnitude (inclusive)
func (z *ZodComplex[T]) Lte(value float64, params ...any) *ZodComplex[T] {
	check := checks.NewCustom[any](func(v any) bool {
		if val, ok := v.(T); ok {
			return cmplx.Abs(complex128(val)) <= value
		}
		if ptr, ok := v.(*T); ok && ptr != nil {
			return cmplx.Abs(complex128(*ptr)) <= value
		}
		return false
	}, params...)
	return engine.AddCheck(z, check).(*ZodComplex[T])
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

// Refine adds a type-safe refinement check for complex types using reflectx
func (z *ZodComplex[T]) Refine(fn func(T) bool, params ...any) *ZodComplex[T] {
	result := z.RefineAny(func(v any) bool {
		// Use reflectx.Deref for pointer handling
		derefed, ok := reflectx.Deref(v)
		if !ok {
			return false // nil pointer
		}

		// Check if it's nil pointer using reflectx
		if reflectx.IsNilPointer(v) {
			return true // Let upper logic decide
		}

		// Try to convert to target type
		if val, ok := derefed.(T); ok {
			return fn(val)
		}
		return false
	}, params...)
	return result.(*ZodComplex[T])
}

// RefineAny adds custom validation to the complex schema
func (z *ZodComplex[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodComplex[T]) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](func(v any) bool {
		payload := &core.ParsePayload{
			Value:  v,
			Issues: make([]core.ZodRawIssue, 0),
			Path:   make([]any, 0),
		}
		fn(payload)
		return len(payload.Issues) == 0
	})
	return engine.AddCheck(z, check)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a transformation pipeline for complex types
func (z *ZodComplex[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use the ZodTransform to create the transformation
	transform := Transform[any, any](func(input any, ctx *core.RefinementContext) (any, error) {
		// Smart handling of complex values using reflectx
		derefed, ok := reflectx.Deref(input)
		if !ok {
			return nil, ErrTransformNilComplex
		}

		if reflectx.IsNilPointer(input) {
			return nil, ErrTransformNilComplex
		}

		if complexVal, ok := derefed.(T); ok {
			return fn(complexVal, ctx)
		}

		return nil, fmt.Errorf("%w, got %T", ErrExpectedComplex, input)
	})
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodComplex[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodComplex[T]) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: out,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the complex schema optional
func (z *ZodComplex[T]) Optional() core.ZodType[any, any] {
	return Optional(any(z).(interface{ GetInternals() *core.ZodTypeInternals }))
}

// Nilable creates a new complex schema that accepts nil values
func (z *ZodComplex[T]) Nilable() core.ZodType[any, any] {
	cloned := engine.Clone(z, func(def *core.ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodComplex[T]).internals.SetNilable()
	return cloned
}

// Nullish makes the complex schema both optional and nullable
func (z *ZodComplex[T]) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(interface{ GetInternals() *core.ZodTypeInternals }))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodComplex[T]) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodComplexDefault embeds ZodDefault with concrete pointer for method promotion
type ZodComplexDefault[T ZodComplexConstraint] struct {
	*ZodDefault[*ZodComplex[T]] // Embed concrete pointer, allows method promotion
}

// ZodComplexPrefault embeds ZodPrefault with concrete pointer for method promotion
type ZodComplexPrefault[T ZodComplexConstraint] struct {
	*ZodPrefault[*ZodComplex[T]] // Embed concrete pointer, allows method promotion
}

// Default creates a default wrapper for complex schema
func (z *ZodComplex[T]) Default(value T) ZodComplexDefault[T] {
	return ZodComplexDefault[T]{
		&ZodDefault[*ZodComplex[T]]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a function-based default wrapper for complex schema
func (z *ZodComplex[T]) DefaultFunc(fn func() T) ZodComplexDefault[T] {
	genericFn := func() any { return fn() }
	return ZodComplexDefault[T]{
		&ZodDefault[*ZodComplex[T]]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault creates a prefault wrapper for complex schema
func (z *ZodComplex[T]) Prefault(value T) ZodComplexPrefault[T] {
	// Construct Prefault internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodComplexPrefault[T]{
		&ZodPrefault[*ZodComplex[T]]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a function-based prefault wrapper for complex schema
func (z *ZodComplex[T]) PrefaultFunc(fn func() T) ZodComplexPrefault[T] {
	genericFn := func() any { return fn() }

	// Construct Prefault internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodComplexPrefault[T]{
		&ZodPrefault[*ZodComplex[T]]{
			internals:     internals,
			innerType:     z,
			prefaultValue: *new(T),
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Type-safe wrapper methods for ZodComplexDefault
func (s ZodComplexDefault[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

func (s ZodComplexDefault[T]) Min(minimum float64, params ...any) ZodComplexDefault[T] {
	newInner := s.innerType.Min(minimum, params...)
	return ZodComplexDefault[T]{
		&ZodDefault[*ZodComplex[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodComplexDefault[T]) Max(maximum float64, params ...any) ZodComplexDefault[T] {
	newInner := s.innerType.Max(maximum, params...)
	return ZodComplexDefault[T]{
		&ZodDefault[*ZodComplex[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodComplexDefault[T]) Positive(params ...any) ZodComplexDefault[T] {
	newInner := s.innerType.Positive(params...)
	return ZodComplexDefault[T]{
		&ZodDefault[*ZodComplex[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodComplexDefault[T]) Refine(fn func(T) bool, params ...any) ZodComplexDefault[T] {
	newInner := s.innerType.Refine(fn, params...)
	return ZodComplexDefault[T]{
		&ZodDefault[*ZodComplex[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodComplexDefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use the embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Smart handling of complex values using reflectx
		derefed, ok := reflectx.Deref(input)
		if !ok {
			return nil, ErrTransformNilComplex
		}

		if reflectx.IsNilPointer(input) {
			return nil, ErrTransformNilComplex
		}

		if complexVal, ok := derefed.(T); ok {
			return fn(complexVal, ctx)
		}

		return nil, fmt.Errorf("%w, got %T", ErrExpectedComplex, input)
	})
}

func (s ZodComplexDefault[T]) Optional() core.ZodType[any, any] {
	// Wrap the current ZodComplexDefault instance, maintain Default logic
	return Optional(any(s).(interface{ GetInternals() *core.ZodTypeInternals }))
}

func (s ZodComplexDefault[T]) Nilable() core.ZodType[any, any] {
	// Wrap the current ZodComplexDefault instance, maintain Default logic
	return any(s).(core.ZodType[any, any]).Nilable()
}

// Type-safe wrapper methods for ZodComplexPrefault
func (s ZodComplexPrefault[T]) Min(minimum float64, params ...any) ZodComplexPrefault[T] {
	newInner := s.innerType.Min(minimum, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodComplexPrefault[T]{
		&ZodPrefault[*ZodComplex[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodComplexPrefault[T]) Max(maximum float64, params ...any) ZodComplexPrefault[T] {
	newInner := s.innerType.Max(maximum, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodComplexPrefault[T]{
		&ZodPrefault[*ZodComplex[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodComplexPrefault[T]) Positive(params ...any) ZodComplexPrefault[T] {
	newInner := s.innerType.Positive(params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodComplexPrefault[T]{
		&ZodPrefault[*ZodComplex[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodComplexPrefault[T]) Refine(fn func(T) bool, params ...any) ZodComplexPrefault[T] {
	newInner := s.innerType.Refine(fn, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &core.ZodTypeInternals{
		Version:     core.Version,
		Type:        core.ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodComplexPrefault[T]{
		&ZodPrefault[*ZodComplex[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodComplexPrefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	// Use the embedded ZodPrefault's TransformAny method
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Smart handling of complex values using reflectx
		derefed, ok := reflectx.Deref(input)
		if !ok {
			return nil, ErrTransformNilComplex
		}

		if reflectx.IsNilPointer(input) {
			return nil, ErrTransformNilComplex
		}

		if complexVal, ok := derefed.(T); ok {
			return fn(complexVal, ctx)
		}

		return nil, fmt.Errorf("%w, got %T", ErrExpectedComplex, input)
	})
}

func (s ZodComplexPrefault[T]) Optional() core.ZodType[any, any] {
	// Wrap the current ZodComplexPrefault instance, maintain Prefault logic
	return Optional(any(s).(interface{ GetInternals() *core.ZodTypeInternals }))
}

func (s ZodComplexPrefault[T]) Nilable() core.ZodType[any, any] {
	// Wrap the current ZodComplexPrefault instance, maintain Prefault logic
	return any(s).(core.ZodType[any, any]).Nilable()
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// Type aliases for convenience
type ZodComplex64 = ZodComplex[complex64]
type ZodComplex128 = ZodComplex[complex128]

// createZodComplexFromDef creates a ZodComplex from definition
func createZodComplexFromDef[T ZodComplexConstraint](def *ZodComplexDef[T]) *ZodComplex[T] {
	internals := &ZodComplexInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Pattern:          nil,
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		complexDef := &ZodComplexDef[T]{
			ZodTypeDef: *newDef,
			Type:       def.Type,
		}
		return createZodComplexFromDef(complexDef)
	}

	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		schema := &ZodComplex[T]{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := core.ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}
		payload.Value = result
		return payload
	}

	zodSchema := &ZodComplex[T]{internals: internals}

	// Initialize the schema with proper error handling support
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

// Complex creates a new generic complex schema
func Complex[T ZodComplexConstraint](params ...any) *ZodComplex[T] {
	// Determine type name from generic parameter
	var typeName core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case complex64:
		typeName = core.ZodTypeComplex64
	case complex128:
		typeName = core.ZodTypeComplex128
	default:
		typeName = core.ZodTypeComplex64 // Default fallback
	}

	def := &ZodComplexDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeName,
			Checks: make([]core.ZodCheck, 0),
		},
		Type: typeName,
	}

	schema := createZodComplexFromDef(def)

	// Apply schema parameters using engine.ApplySchemaParams
	if len(params) > 0 {
		// Convert any params to core.SchemaParams format
		if param, ok := params[0].(core.SchemaParams); ok {
			engine.ApplySchemaParams(&def.ZodTypeDef, param)
			// Note: Coerce handling is now done via dedicated constructors
		}
	}

	return schema
}

// Complex64 creates a new complex64 schema
func Complex64(params ...any) *ZodComplex64 {
	return Complex[complex64](params...)
}

// Complex128 creates a new complex128 schema
func Complex128(params ...any) *ZodComplex128 {
	return Complex[complex128](params...)
}
