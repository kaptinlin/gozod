package types

import (
	"errors"
	"fmt"
	"math"
	"regexp"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error definitions for integer transformations
var (
	ErrTransformNilGeneric = errors.New("cannot transform nil generic value")
	ErrExpectedNumeric     = errors.New("expected numeric type")
)

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////   Generic Integer    //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodIntegerConstraint defines supported integer types for generic implementation
type ZodIntegerConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// ZodIntegerDef defines the configuration for generic integer validation
type ZodIntegerDef[T ZodIntegerConstraint] struct {
	core.ZodTypeDef
	Type     core.ZodTypeCode // "int", "int8", "uint32", etc.
	MinValue T                // Type minimum value
	MaxValue T                // Type maximum value
	Checks   []core.ZodCheck  // Integer-specific validation checks
}

// ZodIntegerInternals contains generic integer validator internal state
type ZodIntegerInternals[T ZodIntegerConstraint] struct {
	core.ZodTypeInternals
	Def     *ZodIntegerDef[T]          // Schema definition
	Checks  []core.ZodCheck            // Validation checks
	Isst    issues.ZodIssueInvalidType // Invalid type issue template
	Pattern *regexp.Regexp             // Integer pattern (if any)
	Bag     map[string]any             // Additional metadata (minimum, maximum, format, coerce flag, etc.)
}

// ZodInteger represents a generic integer validation schema
type ZodInteger[T ZodIntegerConstraint] struct {
	internals *ZodIntegerInternals[T]
}

//////////////////////////////////////////
//////////   Core Interface Methods   ///
//////////////////////////////////////////

// GetInternals returns the internal state of the schema
func (z *ZodInteger[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for integer type conversion
func (z *ZodInteger[T]) Coerce(input any) (any, bool) {
	coerceFunc := createIntegerCoerceFunc[T]()
	return coerceFunc(input)
}

// GetZod returns the integer-specific internals for framework usage
func (z *ZodInteger[T]) GetZod() *ZodIntegerInternals[T] {
	return z.internals
}

// Parse implements intelligent type inference and validation
func (z *ZodInteger[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	var zero T
	typeName := getIntegerTypeName(zero)

	return engine.ParsePrimitive[T](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		validateInteger[T],
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodInteger[T]) MustParse(input any, ctx ...*core.ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////////////////////
//////////   Validation Methods   ///////
//////////////////////////////////////////

// Min adds minimum value validation
func (z *ZodInteger[T]) Min(minimum T, params ...any) *ZodInteger[T] {
	check := checks.Gte(minimum, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Max adds maximum value validation
func (z *ZodInteger[T]) Max(maximum T, params ...any) *ZodInteger[T] {
	check := checks.Lte(maximum, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Gt adds greater than validation (exclusive)
func (z *ZodInteger[T]) Gt(value T, params ...any) *ZodInteger[T] {
	check := checks.Gt(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodInteger[T]) Gte(value T, params ...any) *ZodInteger[T] {
	check := checks.Gte(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Lt adds less than validation (exclusive)
func (z *ZodInteger[T]) Lt(value T, params ...any) *ZodInteger[T] {
	check := checks.Lt(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodInteger[T]) Lte(value T, params ...any) *ZodInteger[T] {
	check := checks.Lte(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Positive adds positive number validation (> 0)
func (z *ZodInteger[T]) Positive(params ...any) *ZodInteger[T] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodInteger[T]) Negative(params ...any) *ZodInteger[T] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodInteger[T]) NonNegative(params ...any) *ZodInteger[T] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodInteger[T]) NonPositive(params ...any) *ZodInteger[T] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple of validation
func (z *ZodInteger[T]) MultipleOf(value T, params ...any) *ZodInteger[T] {
	check := checks.MultipleOf(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Refine adds type-safe custom validation logic
func (z *ZodInteger[T]) Refine(fn func(T) bool, params ...any) *ZodInteger[T] {
	result := z.RefineAny(func(v any) bool {
		intVal, isNil, err := extractIntegerValue[T](v)

		if err != nil {
			return false
		}

		if isNil {
			return true // Let Nilable flag handle nil validation
		}

		return fn(intVal)
	}, params...)
	return result.(*ZodInteger[T])
}

// RefineAny adds flexible custom validation logic
func (z *ZodInteger[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Check adds modern validation using direct payload access
func (z *ZodInteger[T]) Check(fn core.CheckFn) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result
}

////////////////////////////
////   DEFAULT METHODS   ////
////////////////////////////

// ZodIntegerDefault is a Default wrapper for integer type
// Provides perfect type safety and chainable method support
type ZodIntegerDefault[T ZodIntegerConstraint] struct {
	*ZodDefault[*ZodInteger[T]] // Embed concrete pointer to enable method promotion
}

// Default adds a default value to the integer schema, returns ZodIntegerDefault support chain call
func (z *ZodInteger[T]) Default(value T) ZodIntegerDefault[T] {
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the integer schema, returns ZodIntegerDefault support chain call
func (z *ZodInteger[T]) DefaultFunc(fn func() T) ZodIntegerDefault[T] {
	genericFn := func() any { return fn() }
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ZodIntegerDefault chainable validation methods

func (s ZodIntegerDefault[T]) Min(minimum T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Min(minimum, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Max(maximum T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Max(maximum, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Gt(value T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Gt(value, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Gte(value T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Gte(value, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Lt(value T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Lt(value, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Lte(value T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Lte(value, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Positive(params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Positive(params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Negative(params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Negative(params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) NonNegative(params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.NonNegative(params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) NonPositive(params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.NonPositive(params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) MultipleOf(value T, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.MultipleOf(value, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Refine(fn func(T) bool, params ...any) ZodIntegerDefault[T] {
	newInner := s.innerType.Refine(fn, params...)
	return ZodIntegerDefault[T]{
		&ZodDefault[*ZodInteger[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodIntegerDefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use extractIntegerValue helper function that already exists
		val, isNil, err := extractIntegerValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, fmt.Errorf("cannot transform nil value")
		}
		return fn(val, ctx)
	})
}

func (s ZodIntegerDefault[T]) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodIntegerDefault[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   PREFAULT METHODS   ////
////////////////////////////

// ZodIntegerPrefault is a Prefault wrapper for integer type
// Provides perfect type safety and chainable method support
type ZodIntegerPrefault[T ZodIntegerConstraint] struct {
	*ZodPrefault[*ZodInteger[T]] // Embed concrete pointer to enable method promotion
}

// Prefault adds a prefault value to the integer schema, returns ZodIntegerPrefault support chain call
func (z *ZodInteger[T]) Prefault(value T) ZodIntegerPrefault[T] {
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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the integer schema, returns ZodIntegerPrefault support chain call
func (z *ZodInteger[T]) PrefaultFunc(fn func() T) ZodIntegerPrefault[T] {
	genericFn := func() any { return fn() }

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// ZodIntegerPrefault chainable validation methods

func (s ZodIntegerPrefault[T]) Min(minimum T, params ...any) ZodIntegerPrefault[T] {
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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Max(maximum T, params ...any) ZodIntegerPrefault[T] {
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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Gt(value T, params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.Gt(value, params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Gte(value T, params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.Gte(value, params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Lt(value T, params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.Lt(value, params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Lte(value T, params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.Lte(value, params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Positive(params ...any) ZodIntegerPrefault[T] {
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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Negative(params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.Negative(params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) NonNegative(params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.NonNegative(params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) NonPositive(params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.NonPositive(params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) MultipleOf(value T, params ...any) ZodIntegerPrefault[T] {
	newInner := s.innerType.MultipleOf(value, params...)

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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Refine(fn func(T) bool, params ...any) ZodIntegerPrefault[T] {
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

	return ZodIntegerPrefault[T]{
		&ZodPrefault[*ZodInteger[T]]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodIntegerPrefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use extractIntegerValue helper function that already exists
		val, isNil, err := extractIntegerValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, fmt.Errorf("cannot transform nil value")
		}
		return fn(val, ctx)
	})
}

func (s ZodIntegerPrefault[T]) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

func (s ZodIntegerPrefault[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
}

////////////////////////////
////   MODIFIER METHODS   ////
////////////////////////////

// Optional adds an optional check to the integer schema, returns ZodType support chain call
func (z *ZodInteger[T]) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the integer schema, returns ZodType support chain call
func (z *ZodInteger[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish adds a nullish check to the integer schema, returns ZodType support chain call
func (z *ZodInteger[T]) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

//////////////////////////////////////////
//////////   Transform Methods   ////////
//////////////////////////////////////////

// Transform creates a type-safe transformation of integer values
func (z *ZodInteger[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use extractIntegerValue helper function that already exists
		val, isNil, err := extractIntegerValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilGeneric
		}
		return fn(val, ctx)
	})
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodInteger[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodInteger[T]) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////////////////////
//////////   Internal Methods   /////////
//////////////////////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodInteger[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodInteger[T]); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy integer-specific definition fields
		if src.internals.Def != nil {
			z.internals.Def.MinValue = src.internals.Def.MinValue
			z.internals.Def.MaxValue = src.internals.Def.MaxValue
		}

		// Copy Pattern state
		if src.internals.Pattern != nil {
			z.internals.Pattern = src.internals.Pattern
		}
	}
}

//////////////////////////////////////////
//////////   Construction   /////////////
//////////////////////////////////////////

// createZodIntegerFromDef creates a ZodInteger from definition
func createZodIntegerFromDef[T ZodIntegerConstraint](def *ZodIntegerDef[T]) *ZodInteger[T] {
	internals := &ZodIntegerInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             issues.ZodIssueInvalidType{Expected: def.Type},
		Pattern:          nil,
		Bag:              make(map[string]any),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		integerDef := &ZodIntegerDef[T]{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			MinValue:   def.MinValue,
			MaxValue:   def.MaxValue,
			Checks:     newDef.Checks,
		}
		return createZodIntegerFromDef(integerDef)
	}

	zodSchema := &ZodInteger[T]{internals: internals}

	// Set up parse function
	internals.Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := zodSchema.Parse(payload.Value, ctx)
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

	// Initialize the schema with proper error handling support
	engine.InitZodType(zodSchema, &def.ZodTypeDef)

	return zodSchema
}

//////////////////////////////////////////
//////////   Utility Functions   ////////
//////////////////////////////////////////

// createIntegerCoerceFunc creates a coercion function for target integer type T
// This replaces multiple previous conversion functions and uses coerce package
func createIntegerCoerceFunc[T ZodIntegerConstraint]() func(any) (T, bool) {
	return func(value any) (T, bool) {
		result, err := coerce.ToInteger[T](value)
		return result, err == nil
	}
}

// convertSafeInteger safely converts int64 to target integer type T with range checking
func convertSafeInteger[T ZodIntegerConstraint](value int64) (T, bool) {
	var zero T
	minVal, maxVal := getIntegerTypeBounds[T]()

	// Check if value is within target type range
	if value >= int64(minVal) && value <= int64(maxVal) {
		return T(value), true
	}

	return zero, false
}

// convertSafeUInteger safely converts uint64 to target integer type T with range checking
func convertSafeUInteger[T ZodIntegerConstraint](value uint64) (T, bool) {
	var zero T
	minVal, maxVal := getIntegerTypeBounds[T]()

	// Check if value is within target type range
	if value <= uint64(maxVal) && (minVal >= 0 || value <= uint64(math.MaxInt64)) {
		return T(value), true
	}

	return zero, false
}

// isIntegerWithinRange checks if integer value is within min and max range

// getIntegerTypeBounds returns min, max values for integer type T
// Simplified version of the previous getIntegerTypeInfo function
func getIntegerTypeBounds[T ZodIntegerConstraint]() (T, T) {
	var zero T

	// Use runtime type assertion to avoid compile-time overflow errors
	switch any(zero).(type) {
	case int:
		return any(math.MinInt).(T), any(math.MaxInt).(T)
	case int8:
		return any(int8(math.MinInt8)).(T), any(int8(math.MaxInt8)).(T)
	case int16:
		return any(int16(math.MinInt16)).(T), any(int16(math.MaxInt16)).(T)
	case int32:
		return any(int32(math.MinInt32)).(T), any(int32(math.MaxInt32)).(T)
	case int64:
		return any(int64(math.MinInt64)).(T), any(int64(math.MaxInt64)).(T)
	case uint:
		return any(uint(0)).(T), any(uint(math.MaxUint)).(T)
	case uint8:
		return any(uint8(0)).(T), any(uint8(math.MaxUint8)).(T)
	case uint16:
		return any(uint16(0)).(T), any(uint16(math.MaxUint16)).(T)
	case uint32:
		return any(uint32(0)).(T), any(uint32(math.MaxUint32)).(T)
	case uint64:
		return any(uint64(0)).(T), any(uint64(math.MaxUint64)).(T)
	default:
		return zero, zero
	}
}

// getIntegerTypeName returns the type name for integer type T
func getIntegerTypeName[T ZodIntegerConstraint](zero T) core.ZodTypeCode {
	switch any(zero).(type) {
	case int:
		return core.ZodTypeInt
	case int8:
		return core.ZodTypeInt8
	case int16:
		return core.ZodTypeInt16
	case int32:
		return core.ZodTypeInt32
	case int64:
		return core.ZodTypeInt64
	case uint:
		return core.ZodTypeUint
	case uint8:
		return core.ZodTypeUint8
	case uint16:
		return core.ZodTypeUint16
	case uint32:
		return core.ZodTypeUint32
	case uint64:
		return core.ZodTypeUint64
	default:
		return "unknown"
	}
}

//////////////////////////////////////////
//////////   Generic Constructors   /////
//////////////////////////////////////////

// Integer creates a new generic integer schema
func Integer[T ZodIntegerConstraint](params ...any) *ZodInteger[T] {
	minValue, maxValue := getIntegerTypeBounds[T]()

	// Determine type name
	var typeName core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case int:
		typeName = core.ZodTypeInt
	case int8:
		typeName = core.ZodTypeInt8
	case int16:
		typeName = core.ZodTypeInt16
	case int32:
		typeName = core.ZodTypeInt32
	case int64:
		typeName = core.ZodTypeInt64
	case uint:
		typeName = core.ZodTypeUint
	case uint8:
		typeName = core.ZodTypeUint8
	case uint16:
		typeName = core.ZodTypeUint16
	case uint32:
		typeName = core.ZodTypeUint32
	case uint64:
		typeName = core.ZodTypeUint64
	default:
		typeName = "unknown"
	}

	def := &ZodIntegerDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeName,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     typeName,
		MinValue: minValue,
		MaxValue: maxValue,
		Checks:   make([]core.ZodCheck, 0),
	}

	schema := createZodIntegerFromDef(def)

	// Apply schema parameters using the same pattern as string.go
	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return p
			})
			def.Error = &errorMap
			schema.internals.Error = &errorMap
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				// Handle string error messages by converting to ZodErrorMap
				if errStr, ok := p.Error.(string); ok {
					errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
						return errStr
					})
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				} else if errorMap, ok := p.Error.(core.ZodErrorMap); ok {
					def.Error = &errorMap
					schema.internals.Error = &errorMap
				}
			}

			if p.Description != "" {
				schema.internals.Bag["description"] = p.Description
			}
			if p.Abort {
				schema.internals.Bag["abort"] = true
			}
			if len(p.Path) > 0 {
				schema.internals.Bag["path"] = p.Path
			}
		}
	}

	return schema
}

//////////////////////////////////////////
//////////////////////////////////////////
//////////                      //////////
//////////   Concrete Types     //////////
//////////                      //////////
//////////////////////////////////////////
//////////////////////////////////////////

// Concrete integer type definitions using the generic base
type ZodInt = ZodInteger[int]
type ZodInt8 = ZodInteger[int8]
type ZodInt16 = ZodInteger[int16]
type ZodInt32 = ZodInteger[int32]
type ZodInt64 = ZodInteger[int64]
type ZodUint = ZodInteger[uint]
type ZodUint8 = ZodInteger[uint8]
type ZodUint16 = ZodInteger[uint16]
type ZodUint32 = ZodInteger[uint32]
type ZodUint64 = ZodInteger[uint64]

//////////////////////////////////////////
//////////   Concrete Constructors   ////
//////////////////////////////////////////

// Int creates a new int schema
func Int(params ...any) *ZodInt {
	return Integer[int](params...)
}

// Int8 creates a new int8 schema
func Int8(params ...any) *ZodInt8 {
	return Integer[int8](params...)
}

// Int16 creates a new int16 schema
func Int16(params ...any) *ZodInt16 {
	return Integer[int16](params...)
}

// Int32 creates a new int32 schema
func Int32(params ...any) *ZodInt32 {
	return Integer[int32](params...)
}

// Int64 creates a new int64 schema
func Int64(params ...any) *ZodInt64 {
	return Integer[int64](params...)
}

// Uint creates a new uint schema
func Uint(params ...any) *ZodUint {
	return Integer[uint](params...)
}

// Uint8 creates a new uint8 schema
func Uint8(params ...any) *ZodUint8 {
	return Integer[uint8](params...)
}

// Uint16 creates a new uint16 schema
func Uint16(params ...any) *ZodUint16 {
	return Integer[uint16](params...)
}

// Uint32 creates a new uint32 schema
func Uint32(params ...any) *ZodUint32 {
	return Integer[uint32](params...)
}

// Uint64 creates a new uint64 schema
func Uint64(params ...any) *ZodUint64 {
	return Integer[uint64](params...)
}

// Byte creates a new byte schema (uint8 alias)
func Byte(params ...any) *ZodUint8 {
	return Uint8(params...)
}

// Rune creates a new rune schema (int32 alias)
func Rune(params ...any) *ZodInt32 {
	return Int32(params...)
}

////////////////////////////
////   UTILITY FUNCTIONS ////
////////////////////////////

// extractIntegerValue smartly extracts integer values using reflectx package
func extractIntegerValue[T ZodIntegerConstraint](input any) (T, bool, error) {
	var zero T

	// Use reflectx.IsNil for better nil checking
	if reflectx.IsNil(input) {
		return zero, true, nil
	}

	// Handle pointer types using reflectx.Deref
	if reflectx.IsPointer(input) {
		if deref, ok := reflectx.Deref(input); ok {
			return extractIntegerValue[T](deref)
		}
		return zero, true, nil // nil pointer
	}

	// Direct type match
	if val, ok := input.(T); ok {
		return val, false, nil
	}

	// Use reflectx for integer type extraction with proper bounds checking
	var zero_sample T
	switch any(zero_sample).(type) {
	case int, int8, int16, int32, int64:
		// Signed integer types
		if val, ok := reflectx.ExtractInt(input); ok {
			if converted, convertOk := convertSafeInteger[T](val); convertOk {
				return converted, false, nil
			}
			return zero, false, fmt.Errorf("integer value %d out of range for type %T", val, zero)
		}
	case uint, uint8, uint16, uint32, uint64:
		// Unsigned integer types - try both signed and unsigned extraction
		if val, ok := reflectx.ExtractUint(input); ok {
			if converted, convertOk := convertSafeUInteger[T](val); convertOk {
				return converted, false, nil
			}
			return zero, false, fmt.Errorf("unsigned integer value %d out of range for type %T", val, zero)
		}
		// Also try signed integers for unsigned types (handle positive values)
		if val, ok := reflectx.ExtractInt(input); ok && val >= 0 {
			if converted, convertOk := convertSafeInteger[T](val); convertOk {
				return converted, false, nil
			}
			return zero, false, fmt.Errorf("integer value %d out of range for type %T", val, zero)
		}
	}

	// Fallback to coerce package for type conversion
	if result, err := coerce.ToInteger[T](input); err == nil {
		return result, false, nil
	}

	return zero, false, fmt.Errorf("%w or *%T, got %T", ErrExpectedNumeric, zero, input)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodInteger[T]) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateInteger validates integer values with checks (generic)
func validateInteger[T ZodIntegerConstraint](value T, checks []core.ZodCheck, ctx *core.ParseContext) error {
	if len(checks) > 0 {
		payload := &core.ParsePayload{
			Value:  value,
			Issues: make([]core.ZodRawIssue, 0),
		}
		engine.RunChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return issues.NewZodError(issues.ConvertRawIssuesToIssues(payload.Issues, ctx))
		}
	}
	return nil
}
