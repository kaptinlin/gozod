package gozod

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
	ZodTypeDef
	Type string // "complex64" or "complex128"
}

// ZodComplexInternals contains generic complex number validator internal state
type ZodComplexInternals[T ZodComplexConstraint] struct {
	ZodTypeInternals
	Def     *ZodComplexDef[T]      // Schema definition
	Pattern *regexp.Regexp         // Complex pattern (if any)
	Bag     map[string]interface{} // Additional metadata (coerce flag, etc.)
}

// ZodComplex represents a generic complex number validation schema
type ZodComplex[T ZodComplexConstraint] struct {
	internals *ZodComplexInternals[T]
}

// GetInternals returns the internal state of the schema
func (z *ZodComplex[T]) GetInternals() *ZodTypeInternals {
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

// Coerce attempts to coerce input to target complex type
func (z *ZodComplex[T]) Coerce(input interface{}) (interface{}, bool) {
	coerceFunc := createComplexCoerceFunc[T]()
	return coerceFunc(input)
}

// Parse validates input with smart type inference
func (z *ZodComplex[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	typeName := getComplexTypeName[T]()
	coerceFunc := createComplexCoerceFunc[T]()

	return parseType[T](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		func(v any) (T, bool) {
			if val, ok := v.(T); ok {
				return val, true
			}
			return *new(T), false
		},
		func(v any) (*T, bool) {
			if ptr, ok := v.(*T); ok {
				return ptr, true
			}
			return nil, false
		},
		validateComplex[T],
		coerceFunc,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodComplex[T]) MustParse(input any, ctx ...*ParseContext) any {
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
func (z *ZodComplex[T]) Min(minimum float64, params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckGreaterThan(minimum, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Max adds maximum magnitude validation for complex numbers
func (z *ZodComplex[T]) Max(maximum float64, params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckLessThan(maximum, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Gt adds greater than validation for complex magnitude (exclusive)
func (z *ZodComplex[T]) Gt(value float64, params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckGreaterThan(value, false, params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Gte adds greater than or equal validation for complex magnitude (inclusive)
func (z *ZodComplex[T]) Gte(value float64, params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckGreaterThan(value, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Lt adds less than validation for complex magnitude (exclusive)
func (z *ZodComplex[T]) Lt(value float64, params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckLessThan(value, false, params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Lte adds less than or equal validation for complex magnitude (inclusive)
func (z *ZodComplex[T]) Lte(value float64, params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckLessThan(value, true, params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Positive adds positive magnitude validation (> 0) for complex numbers
func (z *ZodComplex[T]) Positive(params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckPositive(params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Negative adds negative validation (< 0) - validates real part for complex numbers
func (z *ZodComplex[T]) Negative(params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckNegative(params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// NonNegative adds non-negative validation (>= 0) for complex magnitude
func (z *ZodComplex[T]) NonNegative(params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckNonnegative(params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// NonPositive adds non-positive validation (<= 0) for complex magnitude
func (z *ZodComplex[T]) NonPositive(params ...SchemaParams) *ZodComplex[T] {
	check := NewZodCheckNonpositive(params...)
	result := AddCheck(z, check)
	return result.(*ZodComplex[T])
}

// Refine adds a type-safe refinement check for complex types
func (z *ZodComplex[T]) Refine(fn func(T) bool, params ...SchemaParams) *ZodComplex[T] {
	result := z.RefineAny(func(v any) bool {
		val, isNil, err := extractComplexValue[T](v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide
		}
		return fn(val)
	}, params...)
	return result.(*ZodComplex[T])
}

// RefineAny adds custom validation to the complex schema
func (z *ZodComplex[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[interface{}](fn, params...)
	return AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodComplex[T]) Check(fn CheckFn) ZodType[any, any] {
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

// Transform creates a transformation pipeline for complex types
func (z *ZodComplex[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		val, isNil, err := extractComplexValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilComplex
		}
		return fn(val, ctx)
	})
}

// TransformAny creates a transformation pipeline that accepts any input type
func (z *ZodComplex[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a validation pipeline
func (z *ZodComplex[T]) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional makes the complex schema optional
func (z *ZodComplex[T]) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable creates a new complex schema that accepts nil values
func (z *ZodComplex[T]) Nilable() ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodComplex[T]) setNilable() ZodType[any, any] {
	cloned := Clone(z, func(def *ZodTypeDef) {
		// Clone operates on ZodTypeDef level
	})
	cloned.(*ZodComplex[T]).internals.Nilable = true
	return cloned
}

// Nullish makes the complex schema both optional and nullable
func (z *ZodComplex[T]) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodComplex[T]) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
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
func (s ZodComplexDefault[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

func (s ZodComplexDefault[T]) Min(minimum float64, params ...SchemaParams) ZodComplexDefault[T] {
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

func (s ZodComplexDefault[T]) Max(maximum float64, params ...SchemaParams) ZodComplexDefault[T] {
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

func (s ZodComplexDefault[T]) Positive(params ...SchemaParams) ZodComplexDefault[T] {
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

func (s ZodComplexDefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodComplexDefault[T] {
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

func (s ZodComplexDefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use the embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of complex values
		complexVal, isNil, err := extractComplexValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilComplex
		}
		return fn(complexVal, ctx)
	})
}

func (s ZodComplexDefault[T]) Optional() ZodType[any, any] {
	// Wrap the current ZodComplexDefault instance, maintain Default logic
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodComplexDefault[T]) Nilable() ZodType[any, any] {
	// Wrap the current ZodComplexDefault instance, maintain Default logic
	return Nilable(any(s).(ZodType[any, any]))
}

// Type-safe wrapper methods for ZodComplexPrefault
func (s ZodComplexPrefault[T]) Min(minimum float64, params ...SchemaParams) ZodComplexPrefault[T] {
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

func (s ZodComplexPrefault[T]) Max(maximum float64, params ...SchemaParams) ZodComplexPrefault[T] {
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

func (s ZodComplexPrefault[T]) Positive(params ...SchemaParams) ZodComplexPrefault[T] {
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

func (s ZodComplexPrefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodComplexPrefault[T] {
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

func (s ZodComplexPrefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use the embedded ZodPrefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of complex values
		complexVal, isNil, err := extractComplexValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilComplex
		}
		return fn(complexVal, ctx)
	})
}

func (s ZodComplexPrefault[T]) Optional() ZodType[any, any] {
	// Wrap the current ZodComplexPrefault instance, maintain Prefault logic
	return Optional(any(s).(ZodType[any, any]))
}

func (s ZodComplexPrefault[T]) Nilable() ZodType[any, any] {
	// Wrap the current ZodComplexPrefault instance, maintain Prefault logic
	return Nilable(any(s).(ZodType[any, any]))
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
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Pattern:          nil,
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		complexDef := &ZodComplexDef[T]{
			ZodTypeDef: *newDef,
			Type:       def.Type,
		}
		return createZodComplexFromDef(complexDef)
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodComplex[T]{internals: internals}
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

	schema := &ZodComplex[T]{internals: internals}

	// Initialize the schema with proper error handling support
	initZodType(schema, &def.ZodTypeDef)

	return schema
}

// NewZodComplex creates a new generic complex schema with full error handling support
func NewZodComplex[T ZodComplexConstraint](params ...SchemaParams) *ZodComplex[T] {
	typeName := getComplexTypeName[T]()
	def := &ZodComplexDef[T]{
		ZodTypeDef: ZodTypeDef{
			Type:   typeName,
			Checks: make([]ZodCheck, 0),
		},
		Type: typeName,
	}

	schema := createZodComplexFromDef(def)

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

// NewZodComplex64 creates a new complex64 schema
func NewZodComplex64(params ...SchemaParams) *ZodComplex64 {
	return NewZodComplex[complex64](params...)
}

// NewZodComplex128 creates a new complex128 schema
func NewZodComplex128(params ...SchemaParams) *ZodComplex128 {
	return NewZodComplex[complex128](params...)
}

// Complex64 creates a new complex64 schema (main constructor)
func Complex64(params ...SchemaParams) *ZodComplex64 {
	return NewZodComplex64(params...)
}

// Complex128 creates a new complex128 schema (main constructor)
func Complex128(params ...SchemaParams) *ZodComplex128 {
	return NewZodComplex128(params...)
}

// CoercedComplex64 creates a new complex64 schema with coercion enabled
func CoercedComplex64(params ...SchemaParams) *ZodComplex64 {
	// Force coercion to true
	var modifiedParams SchemaParams
	if len(params) > 0 {
		modifiedParams = params[0]
	}
	modifiedParams.Coerce = true

	return NewZodComplex64(modifiedParams)
}

// CoercedComplex128 creates a new complex128 schema with coercion enabled
func CoercedComplex128(params ...SchemaParams) *ZodComplex128 {
	// Force coercion to true
	var modifiedParams SchemaParams
	if len(params) > 0 {
		modifiedParams = params[0]
	}
	modifiedParams.Coerce = true

	return NewZodComplex128(modifiedParams)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// extractComplexValue extracts complex value, handling various input types
func extractComplexValue[T ZodComplexConstraint](input any) (T, bool, error) {
	if input == nil {
		return *new(T), true, nil
	}

	switch v := input.(type) {
	case T:
		return v, false, nil
	case *T:
		if v == nil {
			return *new(T), true, nil
		}
		return *v, false, nil
	default:
		return *new(T), false, fmt.Errorf("%w, got %T", ErrExpectedComplex, input)
	}
}

// validateComplex validates complex number values against checks
func validateComplex[T ZodComplexConstraint](value T, checks []ZodCheck, ctx *ParseContext) error {
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}
	return nil
}

// createComplexCoerceFunc creates a coercion function for complex types
func createComplexCoerceFunc[T ZodComplexConstraint]() func(interface{}) (T, bool) {
	return func(value interface{}) (T, bool) {
		// Try direct type assertion first
		if result, ok := value.(T); ok {
			return result, true
		}

		// Handle cross-complex type coercion
		switch v := value.(type) {
		case complex64:
			return T(complex128(v)), true
		case complex128:
			return T(v), true
		case int:
			return T(complex(float64(v), 0)), true
		case int8:
			return T(complex(float64(v), 0)), true
		case int16:
			return T(complex(float64(v), 0)), true
		case int32:
			return T(complex(float64(v), 0)), true
		case int64:
			return T(complex(float64(v), 0)), true
		case uint:
			return T(complex(float64(v), 0)), true
		case uint8:
			return T(complex(float64(v), 0)), true
		case uint16:
			return T(complex(float64(v), 0)), true
		case uint32:
			return T(complex(float64(v), 0)), true
		case uint64:
			return T(complex(float64(v), 0)), true
		case float32:
			return T(complex(float64(v), 0)), true
		case float64:
			return T(complex(v, 0)), true
		case string:
			// Parse complex strings like "3+4i", "5", etc.
			if parsed, err := parseComplexString(v); err == nil {
				return T(parsed), true
			}
		}

		return *new(T), false
	}
}

// parseComplexString parses complex number from string representation
func parseComplexString(s string) (complex128, error) {
	s = strings.ReplaceAll(s, " ", "")

	// Handle simple real numbers
	if !strings.Contains(s, "i") {
		if real, err := strconv.ParseFloat(s, 64); err == nil {
			return complex(real, 0), nil
		}
	}

	// Handle pure imaginary numbers like "4i"
	if strings.HasSuffix(s, "i") && !strings.Contains(s[:len(s)-1], "+") && !strings.Contains(s[:len(s)-1], "-") {
		imagStr := s[:len(s)-1]
		switch imagStr {
		case "", "+":
			return complex(0, 1), nil
		case "-":
			return complex(0, -1), nil
		default:
			if imag, err := strconv.ParseFloat(imagStr, 64); err == nil {
				return complex(0, imag), nil
			}
		}
	}

	// Handle full complex numbers like "3+4i", "3-4i"
	var real, imag float64
	var err error

	// Find the last + or - that's not at the beginning
	lastOp := -1
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '+' || s[i] == '-' {
			lastOp = i
			break
		}
	}

	if lastOp > 0 && strings.HasSuffix(s, "i") {
		realStr := s[:lastOp]
		imagStr := s[lastOp : len(s)-1]

		if real, err = strconv.ParseFloat(realStr, 64); err != nil {
			return 0, err
		}

		switch imagStr {
		case "+":
			imag = 1
		case "-":
			imag = -1
		default:
			if imag, err = strconv.ParseFloat(imagStr, 64); err != nil {
				return 0, err
			}
		}

		return complex(real, imag), nil
	}

	return 0, ErrInvalidComplexFormat
}

// getComplexTypeName returns the type name for the generic complex type
func getComplexTypeName[T ZodComplexConstraint]() string {
	var zero T
	switch any(zero).(type) {
	case complex64:
		return "complex64"
	case complex128:
		return "complex128"
	default:
		return "complex"
	}
}
