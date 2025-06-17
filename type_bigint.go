package gozod

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodBigIntDef defines the configuration for big.Int validation
type ZodBigIntDef struct {
	ZodTypeDef
	Type string // "bigint"
}

// ZodBigIntInternals contains big.Int validator internal state
type ZodBigIntInternals struct {
	ZodTypeInternals
	Def     *ZodBigIntDef          // Schema definition
	Pattern *regexp.Regexp         // BigInt pattern (if any)
	Bag     map[string]interface{} // Additional metadata (minimum, maximum, coerce flag, etc.)
}

// ZodBigInt represents a big.Int validation schema
type ZodBigInt struct {
	internals *ZodBigIntInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodBigInt) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the bigint-specific internals for framework usage
func (z *ZodBigInt) GetZod() *ZodBigIntInternals {
	return z.internals
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodBigInt) CloneFrom(source any) {
	if src, ok := source.(interface{ GetZod() *ZodBigIntInternals }); ok {
		srcState := src.GetZod()
		tgtState := z.GetZod()

		// Copy Bag state (includes coercion flags, etc.)
		if len(srcState.Bag) > 0 {
			if tgtState.Bag == nil {
				tgtState.Bag = make(map[string]interface{})
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

// Coerce attempts to coerce input to big.Int type
func (z *ZodBigInt) Coerce(input interface{}) (interface{}, bool) {
	coerceFunc := createBigIntCoerceFunc()
	return coerceFunc(input)
}

// Parse validates input with smart type inference
func (z *ZodBigInt) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	typeName := "bigint"
	coerceFunc := createBigIntCoerceFunc()

	return parseType[*big.Int](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		func(v any) (*big.Int, bool) {
			if bi, ok := v.(*big.Int); ok {
				return bi, true
			}
			return nil, false
		},
		func(v any) (**big.Int, bool) {
			if bi, ok := v.(**big.Int); ok {
				return bi, true
			}
			return nil, false
		},
		validateBigInt,
		coerceFunc,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodBigInt) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

//////////////////////////
// VALIDATION METHODS
//////////////////////////

// Min adds minimum value validation
func (z *ZodBigInt) Min(minimum *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckGreaterThan(minimum, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Max adds maximum value validation
func (z *ZodBigInt) Max(maximum *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckLessThan(maximum, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Gt adds greater than validation (exclusive)
func (z *ZodBigInt) Gt(value *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckGreaterThan(value, false, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodBigInt) Gte(value *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckGreaterThan(value, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Lt adds less than validation (exclusive)
func (z *ZodBigInt) Lt(value *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckLessThan(value, false, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodBigInt) Lte(value *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckLessThan(value, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Positive adds positive number validation (> 0)
func (z *ZodBigInt) Positive(params ...SchemaParams) *ZodBigInt {
	return z.Gt(big.NewInt(0), params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodBigInt) Negative(params ...SchemaParams) *ZodBigInt {
	return z.Lt(big.NewInt(0), params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodBigInt) NonNegative(params ...SchemaParams) *ZodBigInt {
	return z.Gte(big.NewInt(0), params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodBigInt) NonPositive(params ...SchemaParams) *ZodBigInt {
	return z.Lte(big.NewInt(0), params...)
}

// MultipleOf adds multiple of validation
func (z *ZodBigInt) MultipleOf(value *big.Int, params ...SchemaParams) *ZodBigInt {
	check := NewZodCheckMultipleOf(value, params...)
	result := AddCheck(z, check)
	return result.(*ZodBigInt)
}

// Refine adds a type-safe refinement check for big.Int types
func (z *ZodBigInt) Refine(fn func(*big.Int) bool, params ...SchemaParams) *ZodBigInt {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractBigIntValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide
		}
		return fn(val)
	}, params...)
	return result.(*ZodBigInt)
}

// RefineAny adds custom validation to the big.Int schema
func (z *ZodBigInt) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[interface{}](fn, params...)
	return AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodBigInt) Check(fn CheckFn) ZodType[any, any] {
	check := NewCustom[interface{}](func(v any) bool {
		payload := &ParsePayload{
			Value:  v,
			Issues: make([]ZodRawIssue, 0),
			Path:   make([]interface{}, 0),
		}
		fn(payload)
		return len(payload.Issues) == 0
	})
	return AddCheck(z, check)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a transformation pipeline for big.Int types
func (z *ZodBigInt) Transform(fn func(*big.Int, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		val, isNil, err := extractBigIntValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBigInt
		}
		return fn(val, ctx)
	})
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodBigInt) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodBigInt) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the big.Int schema optional
func (z *ZodBigInt) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable creates a new big.Int schema that accepts nil values
func (z *ZodBigInt) Nilable() ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodBigInt) setNilable() ZodType[any, any] {
	cloned := Clone(z, func(def *ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodBigInt).internals.Nilable = true
	return cloned
}

// Nullish makes the big.Int schema both optional and nullable
func (z *ZodBigInt) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodBigInt) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodBigIntDefault embeds ZodDefault with concrete pointer for method promotion
type ZodBigIntDefault struct {
	*ZodDefault[*ZodBigInt] // Embed concrete pointer, allows method promotion
}

// ZodBigIntPrefault embeds ZodPrefault with concrete pointer for method promotion
type ZodBigIntPrefault struct {
	*ZodPrefault[*ZodBigInt] // Embed concrete pointer, allows method promotion
}

// Default creates a default wrapper for bigint schema
func (z *ZodBigInt) Default(value *big.Int) ZodBigIntDefault {
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc creates a function-based default wrapper for bigint schema
func (z *ZodBigInt) DefaultFunc(fn func() *big.Int) ZodBigIntDefault {
	genericFn := func() any { return fn() }
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault creates a prefault wrapper for bigint schema
func (z *ZodBigInt) Prefault(value *big.Int) ZodBigIntPrefault {
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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc creates a function-based prefault wrapper for bigint schema
func (z *ZodBigInt) PrefaultFunc(fn func() *big.Int) ZodBigIntPrefault {
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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// Type-safe wrapper methods for ZodBigIntDefault
func (s ZodBigIntDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

func (s ZodBigIntDefault) Min(minimum *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Max(maximum *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Gt(value *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Gt(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Gte(value *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Gte(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Lt(value *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Lt(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Lte(value *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Lte(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Positive(params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Positive(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Negative(params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Negative(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) NonNegative(params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.NonNegative(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) NonPositive(params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.NonPositive(params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) MultipleOf(value *big.Int, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.MultipleOf(value, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Refine(fn func(*big.Int) bool, params ...SchemaParams) ZodBigIntDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodBigIntDefault{
		&ZodDefault[*ZodBigInt]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodBigIntDefault) Transform(fn func(*big.Int, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use the embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of big integer values
		bigIntVal, isNil, err := extractBigIntValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBigInt
		}
		return fn(bigIntVal, ctx)
	})
}

func (s ZodBigIntDefault) Optional() ZodType[any, any] {
	// Wrap the current ZodBigIntDefault instance, maintain Default logic
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodBigIntDefault) Nilable() ZodType[any, any] {
	// Wrap the current ZodBigIntDefault instance, maintain Default logic
	return Nilable(any(s).(ZodType[any, any]))
}

// Type-safe wrapper methods for ZodBigIntPrefault
func (s ZodBigIntPrefault) Min(minimum *big.Int, params ...SchemaParams) ZodBigIntPrefault {
	newInner := s.innerType.Min(minimum, params...)

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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodBigIntPrefault) Max(maximum *big.Int, params ...SchemaParams) ZodBigIntPrefault {
	newInner := s.innerType.Max(maximum, params...)

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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodBigIntPrefault) Positive(params ...SchemaParams) ZodBigIntPrefault {
	newInner := s.innerType.Positive(params...)

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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodBigIntPrefault) MultipleOf(value *big.Int, params ...SchemaParams) ZodBigIntPrefault {
	newInner := s.innerType.MultipleOf(value, params...)

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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodBigIntPrefault) Refine(fn func(*big.Int) bool, params ...SchemaParams) ZodBigIntPrefault {
	newInner := s.innerType.Refine(fn, params...)

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

	return ZodBigIntPrefault{
		&ZodPrefault[*ZodBigInt]{
			internals:     internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

func (s ZodBigIntPrefault) Transform(fn func(*big.Int, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use the embedded ZodPrefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of big integer values
		bigIntVal, isNil, err := extractBigIntValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilBigInt
		}
		return fn(bigIntVal, ctx)
	})
}

func (s ZodBigIntPrefault) Optional() ZodType[any, any] {
	// Wrap the current ZodBigIntPrefault instance, maintain Prefault logic
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodBigIntPrefault) Nilable() ZodType[any, any] {
	// Wrap the current ZodBigIntPrefault instance, maintain Prefault logic
	return Nilable(any(s).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// ZodInt64BigInt wraps ZodBigInt for int64-compatible validation
type ZodInt64BigInt struct {
	*ZodBigInt
}

// ZodUint64BigInt wraps ZodBigInt for uint64-compatible validation
type ZodUint64BigInt struct {
	*ZodBigInt
}

// Parse validates and converts to int64
func (z *ZodInt64BigInt) Parse(input any, ctx ...*ParseContext) (any, error) {
	result, err := z.ZodBigInt.Parse(input, ctx...)
	if err != nil {
		return nil, err
	}
	if bigIntResult, ok := result.(*big.Int); ok {
		if bigIntResult.IsInt64() {
			return bigIntResult.Int64(), nil
		} else {
			// Value too large for int64
			return nil, &ZodError{
				Issues: []ZodIssue{{
					ZodIssueBase: ZodIssueBase{
						Code:    "too_big",
						Message: "BigInt value too large for int64",
						Path:    []interface{}{},
					},
				}},
			}
		}
	}
	return result, nil
}

func (z *ZodInt64BigInt) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Parse validates and converts to uint64
func (z *ZodUint64BigInt) Parse(input any, ctx ...*ParseContext) (any, error) {
	result, err := z.ZodBigInt.Parse(input, ctx...)
	if err != nil {
		return nil, err
	}
	if bigIntResult, ok := result.(*big.Int); ok {
		if bigIntResult.IsUint64() {
			return bigIntResult.Uint64(), nil
		} else {
			// Value too large for uint64 or negative
			return nil, &ZodError{
				Issues: []ZodIssue{{
					ZodIssueBase: ZodIssueBase{
						Code:    "too_big",
						Message: "BigInt value too large for uint64 or negative",
						Path:    []interface{}{},
					},
				}},
			}
		}
	}
	return result, nil
}

func (z *ZodUint64BigInt) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// createZodBigIntFromDef creates a ZodBigInt from definition
func createZodBigIntFromDef(def *ZodBigIntDef) *ZodBigInt {
	internals := &ZodBigIntInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Pattern:          nil,
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		bigintDef := &ZodBigIntDef{
			ZodTypeDef: *newDef,
			Type:       "bigint",
		}
		return createZodBigIntFromDef(bigintDef)
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodBigInt{internals: internals}
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

	schema := &ZodBigInt{internals: internals}

	// Initialize the schema with proper error handling support
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

// NewZodBigInt creates a new big.Int schema with full error handling support
func NewZodBigInt(params ...SchemaParams) *ZodBigInt {
	def := &ZodBigIntDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "bigint",
			Checks: make([]ZodCheck, 0),
		},
		Type: "bigint",
	}

	schema := createZodBigIntFromDef(def)

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

// NewZodInt64BigInt creates a new int64-compatible BigInt schema
func NewZodInt64BigInt(params ...SchemaParams) *ZodInt64BigInt {
	baseBigInt := NewZodBigInt(params...)

	// Add int64 range validation using individual checks
	minInt64 := big.NewInt(-9223372036854775808) // math.MinInt64
	maxInt64 := big.NewInt(9223372036854775807)  // math.MaxInt64

	// Apply range checks directly via AddCheck
	gteCheck := NewZodCheckGreaterThan(minInt64, true)
	lteCheck := NewZodCheckLessThan(maxInt64, true)

	rangedSchema := AddCheck(baseBigInt, gteCheck)
	rangedSchema = AddCheck(rangedSchema, lteCheck)

	if finalBigInt, ok := rangedSchema.(*ZodBigInt); ok {
		return &ZodInt64BigInt{finalBigInt}
	}

	// Fallback - this shouldn't happen but provides safety
	return &ZodInt64BigInt{baseBigInt}
}

// NewZodUint64BigInt creates a new uint64-compatible BigInt schema
func NewZodUint64BigInt(params ...SchemaParams) *ZodUint64BigInt {
	baseBigInt := NewZodBigInt(params...)

	// Add uint64 range validation using individual checks
	minUint64 := big.NewInt(0)
	maxUint64 := new(big.Int).SetUint64(18446744073709551615) // math.MaxUint64

	// Apply range checks directly via AddCheck
	gteCheck := NewZodCheckGreaterThan(minUint64, true)
	lteCheck := NewZodCheckLessThan(maxUint64, true)

	rangedSchema := AddCheck(baseBigInt, gteCheck)
	rangedSchema = AddCheck(rangedSchema, lteCheck)

	if finalBigInt, ok := rangedSchema.(*ZodBigInt); ok {
		return &ZodUint64BigInt{finalBigInt}
	}

	// Fallback - this shouldn't happen but provides safety
	return &ZodUint64BigInt{baseBigInt}
}

// BigInt creates a new big.Int schema (main constructor)
func BigInt(params ...SchemaParams) *ZodBigInt {
	return NewZodBigInt(params...)
}

// CoercedBigInt creates a new big.Int schema with coercion enabled
func CoercedBigInt(params ...SchemaParams) *ZodBigInt {
	// Force coercion to true
	var modifiedParams SchemaParams
	if len(params) > 0 {
		modifiedParams = params[0]
	}
	modifiedParams.Coerce = true

	return NewZodBigInt(modifiedParams)
}

// Int64BigInt creates a new int64-compatible BigInt schema
func Int64BigInt(params ...SchemaParams) *ZodInt64BigInt {
	return NewZodInt64BigInt(params...)
}

// Uint64BigInt creates a new uint64-compatible BigInt schema
func Uint64BigInt(params ...SchemaParams) *ZodUint64BigInt {
	return NewZodUint64BigInt(params...)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// extractBigIntValue extracts big.Int value, handling various input types
func extractBigIntValue(input any) (*big.Int, bool, error) {
	if input == nil {
		return nil, true, nil
	}

	switch v := input.(type) {
	case *big.Int:
		if v == nil {
			return nil, true, nil
		}
		return v, false, nil
	case big.Int:
		return &v, false, nil
	default:
		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedBigInt, input)
	}
}

// createBigIntCoerceFunc creates a coercion function for big.Int
func createBigIntCoerceFunc() func(interface{}) (*big.Int, bool) {
	return func(value interface{}) (*big.Int, bool) {
		// Try direct type assertion first
		if result, ok := value.(*big.Int); ok {
			return new(big.Int).Set(result), true
		}

		// Accept all numeric types directly (including bool for boolean to BigInt coercion)
		switch v := value.(type) {
		case bool:
			if v {
				return big.NewInt(1), true
			}
			return big.NewInt(0), true
		case int:
			return big.NewInt(int64(v)), true
		case int8:
			return big.NewInt(int64(v)), true
		case int16:
			return big.NewInt(int64(v)), true
		case int32:
			return big.NewInt(int64(v)), true
		case int64:
			return big.NewInt(v), true
		case uint:
			return new(big.Int).SetUint64(uint64(v)), true
		case uint8:
			return new(big.Int).SetUint64(uint64(v)), true
		case uint16:
			return new(big.Int).SetUint64(uint64(v)), true
		case uint32:
			return new(big.Int).SetUint64(uint64(v)), true
		case uint64:
			return new(big.Int).SetUint64(v), true
		case string:
			bi := new(big.Int)
			_, ok := bi.SetString(v, 10)
			if ok {
				return bi, true
			}
			// Try base 16
			_, ok = bi.SetString(v, 16)
			if ok {
				return bi, true
			}
			// Try base 8
			_, ok = bi.SetString(v, 8)
			return bi, ok
		case float32:
			if v == float32(int64(v)) {
				return big.NewInt(int64(v)), true
			}
		case float64:
			if v == float64(int64(v)) {
				return big.NewInt(int64(v)), true
			}
		}

		// Use coerceToBigInt from utils.go for complex coercion logic (if available)
		return coerceToBigInt(value)
	}
}
