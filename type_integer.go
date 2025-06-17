package gozod

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
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
	ZodTypeDef
	Type     string     // "int", "int8", "uint32", etc.
	MinValue T          // Type minimum value
	MaxValue T          // Type maximum value
	Checks   []ZodCheck // Integer-specific validation checks
}

// ZodIntegerInternals contains generic integer validator internal state
type ZodIntegerInternals[T ZodIntegerConstraint] struct {
	ZodTypeInternals
	Def     *ZodIntegerDef[T]      // Schema definition
	Checks  []ZodCheck             // Validation checks
	Isst    ZodIssueInvalidType    // Invalid type issue template
	Pattern *regexp.Regexp         // Integer pattern (if any)
	Bag     map[string]interface{} // Additional metadata (minimum, maximum, format, coerce flag, etc.)
}

// ZodInteger represents a generic integer validation schema
type ZodInteger[T ZodIntegerConstraint] struct {
	internals *ZodIntegerInternals[T]
}

//////////////////////////////////////////
//////////   Core Interface Methods   ///
//////////////////////////////////////////

// GetInternals returns the internal state of the schema
func (z *ZodInteger[T]) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for integer type conversion
func (z *ZodInteger[T]) Coerce(input interface{}) (interface{}, bool) {
	coerceFunc := createIntegerCoerceFunc[T]()
	return coerceFunc(input)
}

// GetZod returns the integer-specific internals for framework usage
func (z *ZodInteger[T]) GetZod() *ZodIntegerInternals[T] {
	return z.internals
}

// Parse validates and parses input with smart type inference
func (z *ZodInteger[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	var zero T
	typeName := getIntegerTypeName(zero)
	coerceFunc := createIntegerCoerceFunc[T]()

	return parseType[T](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		func(v any) (T, bool) { val, ok := v.(T); return val, ok },
		func(v any) (*T, bool) { ptr, ok := v.(*T); return ptr, ok },
		validateInteger[T],
		coerceFunc,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodInteger[T]) MustParse(input any, ctx ...*ParseContext) any {
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
func (z *ZodInteger[T]) Min(minimum T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckGreaterThan(float64(minimum), true, params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Max adds maximum value validation
func (z *ZodInteger[T]) Max(maximum T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckLessThan(float64(maximum), true, params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Gt adds greater than validation (exclusive)
func (z *ZodInteger[T]) Gt(value T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckGreaterThan(float64(value), false, params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodInteger[T]) Gte(value T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckGreaterThan(float64(value), true, params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Lt adds less than validation (exclusive)
func (z *ZodInteger[T]) Lt(value T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckLessThan(float64(value), false, params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodInteger[T]) Lte(value T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckLessThan(float64(value), true, params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Positive adds positive number validation (> 0)
func (z *ZodInteger[T]) Positive(params ...SchemaParams) *ZodInteger[T] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodInteger[T]) Negative(params ...SchemaParams) *ZodInteger[T] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodInteger[T]) NonNegative(params ...SchemaParams) *ZodInteger[T] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodInteger[T]) NonPositive(params ...SchemaParams) *ZodInteger[T] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple of validation
func (z *ZodInteger[T]) MultipleOf(value T, params ...SchemaParams) *ZodInteger[T] {
	check := NewZodCheckMultipleOf(float64(value), params...)
	result := AddCheck(any(z).(ZodType[any, any]), check)
	return result.(*ZodInteger[T])
}

// Refine adds type-safe custom validation logic
func (z *ZodInteger[T]) Refine(fn func(T) bool, params ...SchemaParams) *ZodInteger[T] {
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
func (z *ZodInteger[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Check adds modern validation using direct payload access
func (z *ZodInteger[T]) Check(fn CheckFn) ZodType[any, any] {
	custom := NewZodCustom(fn, SchemaParams{})
	custom.GetInternals().Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, ZodRawIssue{
				Code:    "invalid_type",
				Message: err.Error(),
				Input:   payload.Value,
			})
			return payload
		}
		payload.Value = result

		fn(payload)
		return payload
	}
	return custom
}

////////////////////////////
////   INTEGER DEFAULT WRAPPER ////
////////////////////////////

// ZodIntegerDefault is a Default wrapper for integer types
// Provides perfect type safety and chainable method support
type ZodIntegerDefault[T ZodIntegerConstraint] struct {
	*ZodDefault[*ZodInteger[T]] // Embed concrete pointer to enable method promotion
}

////////////////////////////
////   DEFAULT METHODS   ////
////////////////////////////

// Default adds a default value to the integer schema, returns ZodIntegerDefault support chain call
// Compile-time type safety: Int().Default("string") will fail to compile
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

////////////////////////////
////   INTEGERDEFAULT CHAINING METHODS ////
////////////////////////////

// Min adds minimum value validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Min(minimum T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Max adds maximum value validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Max(maximum T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Gt adds greater than validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Gt(value T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Gte adds greater than or equal validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Gte(value T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Lt adds less than validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Lt(value T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Lte adds less than or equal validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Lte(value T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Positive adds positive number validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Positive(params ...SchemaParams) ZodIntegerDefault[T] {
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

// Negative adds negative number validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) Negative(params ...SchemaParams) ZodIntegerDefault[T] {
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

// NonNegative adds non-negative number validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) NonNegative(params ...SchemaParams) ZodIntegerDefault[T] {
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

// NonPositive adds non-positive number validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) NonPositive(params ...SchemaParams) ZodIntegerDefault[T] {
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

// MultipleOf adds multiple validation, returns ZodIntegerDefault for method chaining
func (s ZodIntegerDefault[T]) MultipleOf(value T, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Refine adds a flexible validation function to the integer schema, returns ZodIntegerDefault support chain call
func (s ZodIntegerDefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodIntegerDefault[T] {
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

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (s ZodIntegerDefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of integer values
		intVal, isNil, err := extractIntegerValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilGeneric
		}
		return fn(intVal, ctx)
	})
}

// Optional adds an optional check to the integer schema, returns ZodType support chain call
func (s ZodIntegerDefault[T]) Optional() ZodType[any, any] {
	// Wrap current ZodIntegerDefault instance, maintain Default logic
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the integer schema, returns ZodType support chain call
func (s ZodIntegerDefault[T]) Nilable() ZodType[any, any] {
	// Wrap current ZodIntegerDefault instance, maintain Default logic
	return Nilable(any(s).(ZodType[any, any]))
}

//////////////////////////////////////////
//////////   Wrapper Methods   //////////
//////////////////////////////////////////

// Optional adds an optional check to the integer schema, returns ZodType support chain call
func (z *ZodInteger[T]) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable adds a nilable check to the integer schema, returns ZodType support chain call
func (z *ZodInteger[T]) Nilable() ZodType[any, any] {
	return Nilable(any(z).(ZodType[any, any]))
}

// Nullish adds a nullish check to the integer schema, returns ZodType support chain call
func (z *ZodInteger[T]) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

////////////////////////////
////   INTEGER PREFAULT WRAPPER ////
////////////////////////////

// ZodIntegerPrefault is a Prefault wrapper for integer types
// Provides perfect type safety and chainable method support
type ZodIntegerPrefault[T ZodIntegerConstraint] struct {
	*ZodPrefault[*ZodInteger[T]] // Embed concrete pointer to enable method promotion
}

////////////////////////////
////   PREFAULT METHODS   ////
////////////////////////////

// Prefault adds a prefault value to the integer schema, returns ZodIntegerPrefault support chain call
// Compile-time type safety: Int().Prefault("invalid") will fail to compile
func (z *ZodInteger[T]) Prefault(value T) ZodIntegerPrefault[T] {
	// Construct Prefault internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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

	// Construct Prefault internals, Type = "prefault", copy inner type's checks/coerce/optional/nilable
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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

////////////////////////////
////   INTEGERPREFAULT CHAINING METHODS ////
////////////////////////////

// Min adds minimum value validation, returns ZodIntegerPrefault for method chaining
func (i ZodIntegerPrefault[T]) Min(minimum T, params ...SchemaParams) ZodIntegerPrefault[T] {
	newInner := i.innerType.Min(minimum, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
			prefaultValue: i.prefaultValue,
			prefaultFunc:  i.prefaultFunc,
			isFunction:    i.isFunction,
		},
	}
}

// Max adds maximum value validation, returns ZodIntegerPrefault for method chaining
func (i ZodIntegerPrefault[T]) Max(maximum T, params ...SchemaParams) ZodIntegerPrefault[T] {
	newInner := i.innerType.Max(maximum, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
			prefaultValue: i.prefaultValue,
			prefaultFunc:  i.prefaultFunc,
			isFunction:    i.isFunction,
		},
	}
}

// Positive adds positive number validation, returns ZodIntegerPrefault for method chaining
func (i ZodIntegerPrefault[T]) Positive(params ...SchemaParams) ZodIntegerPrefault[T] {
	newInner := i.innerType.Positive(params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
			prefaultValue: i.prefaultValue,
			prefaultFunc:  i.prefaultFunc,
			isFunction:    i.isFunction,
		},
	}
}

// MultipleOf adds multiple validation, returns ZodIntegerPrefault for method chaining
func (i ZodIntegerPrefault[T]) MultipleOf(value T, params ...SchemaParams) ZodIntegerPrefault[T] {
	newInner := i.innerType.MultipleOf(value, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
			prefaultValue: i.prefaultValue,
			prefaultFunc:  i.prefaultFunc,
			isFunction:    i.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the integer schema, returns ZodIntegerPrefault support chain call
func (i ZodIntegerPrefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodIntegerPrefault[T] {
	newInner := i.innerType.Refine(fn, params...)

	// Construct new internals
	baseInternals := newInner.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
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
			prefaultValue: i.prefaultValue,
			prefaultFunc:  i.prefaultFunc,
			isFunction:    i.isFunction,
		},
	}
}

// Transform adds data transformation, returns a generic ZodType support transform pipeline
func (i ZodIntegerPrefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return i.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of integer values
		intVal, isNil, err := extractIntegerValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilGeneric
		}
		return fn(intVal, ctx)
	})
}

// Optional makes the integer optional
func (i ZodIntegerPrefault[T]) Optional() ZodType[any, any] {
	// Wrap current ZodIntegerPrefault instance, maintain Prefault logic
	return Optional(any(i).(ZodType[any, any]))
}

// Nilable makes the integer nilable
func (i ZodIntegerPrefault[T]) Nilable() ZodType[any, any] {
	// Wrap current ZodIntegerPrefault instance, maintain Prefault logic
	return Nilable(any(i).(ZodType[any, any]))
}

//////////////////////////////////////////
//////////   Transform Methods   ////////
//////////////////////////////////////////

// Transform creates a type-safe transformation of integer values
func (z *ZodInteger[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	wrappedFn := func(input any, ctx *RefinementContext) (any, error) {
		intVal, isNil, err := extractIntegerValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return (*T)(nil), nil
		}
		return fn(intVal, ctx)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodInteger[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodInteger[T]) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
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
				z.internals.Bag = make(map[string]interface{})
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
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             ZodIssueInvalidType{Expected: def.Type},
		Pattern:          nil,
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		integerDef := &ZodIntegerDef[T]{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			MinValue:   def.MinValue,
			MaxValue:   def.MaxValue,
			Checks:     newDef.Checks,
		}
		return createZodIntegerFromDef(integerDef)
	}

	schema := &ZodInteger[T]{internals: internals}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
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
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

//////////////////////////////////////////
//////////   Utility Functions   ////////
//////////////////////////////////////////

// createIntegerCoerceFunc creates a coercion function for target integer type T
// This replaces multiple previous conversion functions and uses utils.go functions
func createIntegerCoerceFunc[T ZodIntegerConstraint]() func(interface{}) (T, bool) {
	return func(value interface{}) (T, bool) {
		var zero T

		// Try direct type assertion first
		if result, ok := value.(T); ok {
			return result, true
		}

		// Handle string values explicitly
		if str, ok := value.(string); ok {
			// Trim whitespace
			str = strings.TrimSpace(str)
			if str == "" {
				return zero, false
			}
			// Try to parse as integer
			if i, err := strconv.ParseInt(str, 10, 64); err == nil {
				return convertSafeInteger[T](i)
			}
			return zero, false
		}

		// Accept all numeric types directly
		switch v := value.(type) {
		case int:
			return convertSafeInteger[T](int64(v))
		case int8:
			return convertSafeInteger[T](int64(v))
		case int16:
			return convertSafeInteger[T](int64(v))
		case int32:
			return convertSafeInteger[T](int64(v))
		case int64:
			return convertSafeInteger[T](v)
		case uint:
			return convertSafeUInteger[T](uint64(v))
		case uint8:
			return convertSafeUInteger[T](uint64(v))
		case uint16:
			return convertSafeUInteger[T](uint64(v))
		case uint32:
			return convertSafeUInteger[T](uint64(v))
		case uint64:
			return convertSafeUInteger[T](v)
		case float32:
			// Only allow exact integers in strict mode
			if v == float32(int32(v)) {
				return convertSafeInteger[T](int64(v))
			}
		case float64:
			// Only allow exact integers in strict mode
			if v == float64(int64(v)) {
				return convertSafeInteger[T](int64(v))
			}
		}

		return zero, false
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
func getIntegerTypeName[T ZodIntegerConstraint](zero T) string {
	switch any(zero).(type) {
	case int:
		return "int"
	case int8:
		return "int8"
	case int16:
		return "int16"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case uint:
		return "uint"
	case uint8:
		return "uint8"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case uint64:
		return "uint64"
	default:
		return "unknown"
	}
}

//////////////////////////////////////////
//////////   Generic Constructors   /////
//////////////////////////////////////////

// NewZodInteger creates a new generic integer schema
func NewZodInteger[T ZodIntegerConstraint](params ...SchemaParams) *ZodInteger[T] {
	minValue, maxValue := getIntegerTypeBounds[T]()

	// Determine type name
	var typeName string
	var zero T
	switch any(zero).(type) {
	case int:
		typeName = "int"
	case int8:
		typeName = "int8"
	case int16:
		typeName = "int16"
	case int32:
		typeName = "int32"
	case int64:
		typeName = "int64"
	case uint:
		typeName = "uint"
	case uint8:
		typeName = "uint8"
	case uint16:
		typeName = "uint16"
	case uint32:
		typeName = "uint32"
	case uint64:
		typeName = "uint64"
	default:
		typeName = "unknown"
	}

	def := &ZodIntegerDef[T]{
		ZodTypeDef: ZodTypeDef{
			Type:   typeName,
			Checks: make([]ZodCheck, 0),
		},
		Type:     typeName,
		MinValue: minValue,
		MaxValue: maxValue,
		Checks:   make([]ZodCheck, 0),
	}

	schema := createZodIntegerFromDef(def)

	// Apply schema parameters using utility functions from utils.go
	if len(params) > 0 {
		param := params[0]

		// Store coerce flag in bag for parseType to access
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			schema.internals.ZodTypeInternals.Bag["coerce"] = true // Also set in ZodTypeInternals for parseType
		}

		// Handle schema-level error mapping using utility function from utils.go
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
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

// NewZodInt creates a new int schema
func NewZodInt(params ...SchemaParams) *ZodInt {
	return NewZodInteger[int](params...)
}

// NewZodInt8 creates a new int8 schema
func NewZodInt8(params ...SchemaParams) *ZodInt8 {
	return NewZodInteger[int8](params...)
}

// NewZodInt16 creates a new int16 schema
func NewZodInt16(params ...SchemaParams) *ZodInt16 {
	return NewZodInteger[int16](params...)
}

// NewZodInt32 creates a new int32 schema
func NewZodInt32(params ...SchemaParams) *ZodInt32 {
	return NewZodInteger[int32](params...)
}

// NewZodInt64 creates a new int64 schema
func NewZodInt64(params ...SchemaParams) *ZodInt64 {
	return NewZodInteger[int64](params...)
}

// NewZodUint creates a new uint schema
func NewZodUint(params ...SchemaParams) *ZodUint {
	return NewZodInteger[uint](params...)
}

// NewZodUint8 creates a new uint8 schema
func NewZodUint8(params ...SchemaParams) *ZodUint8 {
	return NewZodInteger[uint8](params...)
}

// NewZodUint16 creates a new uint16 schema
func NewZodUint16(params ...SchemaParams) *ZodUint16 {
	return NewZodInteger[uint16](params...)
}

// NewZodUint32 creates a new uint32 schema
func NewZodUint32(params ...SchemaParams) *ZodUint32 {
	return NewZodInteger[uint32](params...)
}

// NewZodUint64 creates a new uint64 schema
func NewZodUint64(params ...SchemaParams) *ZodUint64 {
	return NewZodInteger[uint64](params...)
}

//////////////////////////////////////////
//////////   Public Constructors   //////
//////////////////////////////////////////

// Int creates a new int schema
func Int(params ...SchemaParams) *ZodInt {
	return NewZodInt(params...)
}

// Int8 creates a new int8 schema
func Int8(params ...SchemaParams) *ZodInt8 {
	return NewZodInt8(params...)
}

// Int16 creates a new int16 schema
func Int16(params ...SchemaParams) *ZodInt16 {
	return NewZodInt16(params...)
}

// Int32 creates a new int32 schema
func Int32(params ...SchemaParams) *ZodInt32 {
	return NewZodInt32(params...)
}

// Int64 creates a new int64 schema
func Int64(params ...SchemaParams) *ZodInt64 {
	return NewZodInt64(params...)
}

// Uint creates a new uint schema
func Uint(params ...SchemaParams) *ZodUint {
	return NewZodUint(params...)
}

// Uint8 creates a new uint8 schema
func Uint8(params ...SchemaParams) *ZodUint8 {
	return NewZodUint8(params...)
}

// Uint16 creates a new uint16 schema
func Uint16(params ...SchemaParams) *ZodUint16 {
	return NewZodUint16(params...)
}

// Uint32 creates a new uint32 schema
func Uint32(params ...SchemaParams) *ZodUint32 {
	return NewZodUint32(params...)
}

// Uint64 creates a new uint64 schema
func Uint64(params ...SchemaParams) *ZodUint64 {
	return NewZodUint64(params...)
}

// Byte creates a new byte schema (uint8 alias)
func Byte(params ...SchemaParams) *ZodUint8 {
	return NewZodUint8(params...)
}

// Rune creates a new rune schema (int32 alias)
func Rune(params ...SchemaParams) *ZodInt32 {
	return NewZodInt32(params...)
}

////////////////////////////
////   UTILITY FUNCTIONS ////
////////////////////////////

// extractIntegerValue smartly extracts integer values
func extractIntegerValue[T ZodIntegerConstraint](input any) (T, bool, error) {
	var zero T
	switch v := input.(type) {
	case T:
		return v, false, nil
	case *T:
		if v == nil {
			return zero, true, nil
		}
		return *v, false, nil
	default:
		return zero, false, fmt.Errorf("%w or *%T, got %T", ErrExpectedNumeric, zero, input)
	}
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodInteger[T]) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}
