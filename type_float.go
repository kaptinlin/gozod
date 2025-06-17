package gozod

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodFloatConstraint defines supported floating-point types for generic implementation
type ZodFloatConstraint interface {
	~float32 | ~float64
}

// ZodFloatDef defines the configuration for generic floating-point validation
type ZodFloatDef[T ZodFloatConstraint] struct {
	ZodTypeDef
	Type     string     // "float32" or "float64"
	MinValue T          // Type minimum value
	MaxValue T          // Type maximum value
	Checks   []ZodCheck // Float-specific validation checks
}

// ZodFloatInternals contains generic floating-point validator internal state
type ZodFloatInternals[T ZodFloatConstraint] struct {
	ZodTypeInternals
	Def     *ZodFloatDef[T]        // Schema definition
	Checks  []ZodCheck             // Validation checks
	Isst    ZodIssueInvalidType    // Invalid type issue template
	Pattern *regexp.Regexp         // Float pattern (if any)
	Bag     map[string]interface{} // Additional metadata (minimum, maximum, format, coerce flag, etc.)
}

// ZodFloat represents a generic floating-point validation schema
type ZodFloat[T ZodFloatConstraint] struct {
	internals *ZodFloatInternals[T]
}

// GetInternals returns the internal state of the schema
func (z *ZodFloat[T]) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for floating-point type conversion
func (z *ZodFloat[T]) Coerce(input interface{}) (interface{}, bool) {
	coerceFunc := createFloatCoerceFunc[T]()
	return coerceFunc(input)
}

// Parse validates and parses input with smart type inference
func (z *ZodFloat[T]) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	typeName := getFloatTypeName[T]()
	coerceFunc := createFloatCoerceFunc[T]()

	return parseType[T](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		func(v any) (T, bool) { val, ok := v.(T); return val, ok },
		func(v any) (*T, bool) { ptr, ok := v.(*T); return ptr, ok },
		validateFloat[T],
		coerceFunc,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodFloat[T]) MustParse(input any, ctx ...*ParseContext) any {
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
func (z *ZodFloat[T]) Min(minimum T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckGreaterThan(float64(minimum), true, params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Max adds maximum value validation
func (z *ZodFloat[T]) Max(maximum T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckLessThan(float64(maximum), true, params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Gt adds greater than validation (exclusive)
func (z *ZodFloat[T]) Gt(value T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckGreaterThan(float64(value), false, params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodFloat[T]) Gte(value T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckGreaterThan(float64(value), true, params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Lt adds less than validation (exclusive)
func (z *ZodFloat[T]) Lt(value T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckLessThan(float64(value), false, params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodFloat[T]) Lte(value T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckLessThan(float64(value), true, params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Positive adds positive number validation (> 0)
func (z *ZodFloat[T]) Positive(params ...SchemaParams) *ZodFloat[T] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodFloat[T]) Negative(params ...SchemaParams) *ZodFloat[T] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodFloat[T]) NonNegative(params ...SchemaParams) *ZodFloat[T] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodFloat[T]) NonPositive(params ...SchemaParams) *ZodFloat[T] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple of validation
func (z *ZodFloat[T]) MultipleOf(value T, params ...SchemaParams) *ZodFloat[T] {
	check := NewZodCheckMultipleOf(float64(value), params...)
	result := AddCheck(z, check)
	return result.(*ZodFloat[T])
}

// Int adds integer validation (no decimal part)
func (z *ZodFloat[T]) Int(params ...SchemaParams) ZodType[any, any] {
	check := NewZodCheckNumberFormat(NumberFormatSafeint, params...)
	return AddCheck(z, check)
}

// Finite adds finite number validation (not NaN or Infinity)
func (z *ZodFloat[T]) Finite(params ...SchemaParams) ZodType[any, any] {
	var format ZodNumberFormats
	var zero T

	switch any(zero).(type) {
	case float32:
		format = NumberFormatFloat32
	case float64:
		format = NumberFormatFloat64
	}

	check := NewZodCheckNumberFormat(format, params...)
	return AddCheck(z, check)
}

// Safe adds safe number validation (within JavaScript safe integer range)
func (z *ZodFloat[T]) Safe(params ...SchemaParams) ZodType[any, any] {
	const maxSafeInt = 1<<53 - 1
	const minSafeInt = -(1<<53 - 1)
	return z.Gte(T(minSafeInt), params...).Lte(T(maxSafeInt), params...)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a type-safe transformation of floating-point values
func (z *ZodFloat[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	wrappedFn := func(input any, ctx *RefinementContext) (any, error) {
		floatVal, isNil, err := extractFloatValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return (*T)(nil), nil
		}
		return fn(floatVal, ctx)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodFloat[T]) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  z,
		out: transform,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodFloat[T]) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  z,
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional adds an optional check to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable adds a nilable check to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Nilable() ZodType[any, any] {
	return Nilable(any(z).(ZodType[any, any]))
}

// Nullish combines optional and nilable
func (z *ZodFloat[T]) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds a flexible validation function to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Refine(fn func(T) bool, params ...SchemaParams) *ZodFloat[T] {
	result := z.RefineAny(func(v any) bool {
		floatVal, isNil, err := extractFloatValue[T](v)

		if err != nil {
			return false
		}

		if isNil {
			return true // Let Nilable flag handle nil validation
		}

		return fn(floatVal)
	}, params...)
	return result.(*ZodFloat[T])
}

// RefineAny adds a flexible validation function to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(z, check)
}

// Check adds a modern validation function to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Check(fn CheckFn) ZodType[any, any] {
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

// Unwrap returns the inner type (for basic types, returns self), returns ZodType support chain call
func (z *ZodFloat[T]) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

//////////////////////////
// WRAPPER TYPES
//////////////////////////

// ZodFloatDefault is a default value wrapper for floating-point type
type ZodFloatDefault[T ZodFloatConstraint] struct {
	*ZodDefault[*ZodFloat[T]]
}

// Default adds a default value to the floating-point, returns ZodFloatDefault support chain call
func (z *ZodFloat[T]) Default(value T) ZodFloatDefault[T] {
	return ZodFloatDefault[T]{
		&ZodDefault[*ZodFloat[T]]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the floating-point, returns ZodFloatDefault support chain call
func (z *ZodFloat[T]) DefaultFunc(fn func() T) ZodFloatDefault[T] {
	genericFn := func() any { return fn() }
	return ZodFloatDefault[T]{
		&ZodDefault[*ZodFloat[T]]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// ZodFloatDefault chain call methods

// Min adds a minimum value validation to the floating-point, returns ZodFloatDefault support chain call
func (s ZodFloatDefault[T]) Min(minimum T, params ...SchemaParams) ZodFloatDefault[T] {
	newInner := s.innerType.Min(minimum, params...)
	return ZodFloatDefault[T]{
		&ZodDefault[*ZodFloat[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Max adds a maximum value validation to the floating-point, returns ZodFloatDefault support chain call
func (s ZodFloatDefault[T]) Max(maximum T, params ...SchemaParams) ZodFloatDefault[T] {
	newInner := s.innerType.Max(maximum, params...)
	return ZodFloatDefault[T]{
		&ZodDefault[*ZodFloat[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Positive adds a positive number validation to the floating-point, returns ZodFloatDefault support chain call
func (s ZodFloatDefault[T]) Positive(params ...SchemaParams) ZodFloatDefault[T] {
	newInner := s.innerType.Positive(params...)
	return ZodFloatDefault[T]{
		&ZodDefault[*ZodFloat[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the floating-point, returns ZodFloatDefault support chain call
func (s ZodFloatDefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodFloatDefault[T] {
	newInner := s.innerType.Refine(fn, params...)
	return ZodFloatDefault[T]{
		&ZodDefault[*ZodFloat[T]]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

func (s ZodFloatDefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		floatVal, isNil, err := extractFloatValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilGeneric
		}
		return fn(floatVal, ctx)
	})
}

// Optional adds an optional check to the floating-point, returns ZodType support chain call
func (s ZodFloatDefault[T]) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the floating-point, returns ZodType support chain call
func (s ZodFloatDefault[T]) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ZodFloatPrefault is a prefault value wrapper for floating-point type, returns ZodFloatPrefault support chain call
type ZodFloatPrefault[T ZodFloatConstraint] struct {
	*ZodPrefault[*ZodFloat[T]]
}

// Prefault adds a prefault value to the floating-point, returns ZodFloatPrefault support chain call
func (z *ZodFloat[T]) Prefault(value T) ZodFloatPrefault[T] {
	return ZodFloatPrefault[T]{
		&ZodPrefault[*ZodFloat[T]]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the floating-point, returns ZodFloatPrefault support chain call
func (z *ZodFloat[T]) PrefaultFunc(fn func() T) ZodFloatPrefault[T] {
	genericFn := func() any { return fn() }
	return ZodFloatPrefault[T]{
		&ZodPrefault[*ZodFloat[T]]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

// ZodFloatPrefault chain call methods

// Positive adds a positive number validation to the floating-point, returns ZodFloatPrefault support chain call
func (f ZodFloatPrefault[T]) Positive(params ...SchemaParams) ZodFloatPrefault[T] {
	newInner := f.innerType.Positive(params...)
	return ZodFloatPrefault[T]{
		&ZodPrefault[*ZodFloat[T]]{
			innerType:     newInner,
			prefaultValue: f.prefaultValue,
			prefaultFunc:  f.prefaultFunc,
			isFunction:    f.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the floating-point, returns ZodFloatPrefault support chain call
func (f ZodFloatPrefault[T]) Refine(fn func(T) bool, params ...SchemaParams) ZodFloatPrefault[T] {
	newInner := f.innerType.Refine(fn, params...)
	return ZodFloatPrefault[T]{
		&ZodPrefault[*ZodFloat[T]]{
			innerType:     newInner,
			prefaultValue: f.prefaultValue,
			prefaultFunc:  f.prefaultFunc,
			isFunction:    f.isFunction,
		},
	}
}

// Transform adds a data transformation function to the floating-point, returns ZodType support transform pipeline
func (f ZodFloatPrefault[T]) Transform(fn func(T, *RefinementContext) (any, error)) ZodType[any, any] {
	return f.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		floatVal, isNil, err := extractFloatValue[T](input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilGeneric
		}
		return fn(floatVal, ctx)
	})
}

// Optional adds an optional check to the floating-point, returns ZodType support chain call
func (f ZodFloatPrefault[T]) Optional() ZodType[any, any] {
	return Optional(any(f).(ZodType[any, any]))
}

// Nilable adds a nilable check to the floating-point, returns ZodType support chain call
func (f ZodFloatPrefault[T]) Nilable() ZodType[any, any] {
	return Nilable(any(f).(ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodFloatFromDef creates a ZodFloat from definition
func createZodFloatFromDef[T ZodFloatConstraint](def *ZodFloatDef[T]) *ZodFloat[T] {
	internals := &ZodFloatInternals[T]{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             ZodIssueInvalidType{Expected: def.Type},
		Pattern:          nil,
		Bag:              make(map[string]interface{}),
	}

	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		floatDef := &ZodFloatDef[T]{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			MinValue:   def.MinValue,
			MaxValue:   def.MaxValue,
			Checks:     newDef.Checks,
		}
		return createZodFloatFromDef(floatDef)
	}

	schema := &ZodFloat[T]{internals: internals}

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

	initZodType(schema, &def.ZodTypeDef)
	return schema
}

// NewZodFloat creates a new generic floating-point schema
func NewZodFloat[T ZodFloatConstraint](params ...SchemaParams) *ZodFloat[T] {
	minValue, maxValue := getFloatTypeBounds[T]()
	typeName := getFloatTypeName[T]()

	def := &ZodFloatDef[T]{
		ZodTypeDef: ZodTypeDef{
			Type: typeName,
		},
		Type:     typeName,
		MinValue: minValue,
		MaxValue: maxValue,
		Checks:   make([]ZodCheck, 0),
	}

	schema := createZodFloatFromDef(def)

	if len(params) > 0 {
		param := params[0]

		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			schema.internals.ZodTypeInternals.Bag["coerce"] = true
		}

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

// Concrete floating-point type definitions
type ZodFloat32 = ZodFloat[float32]
type ZodFloat64 = ZodFloat[float64]
type ZodNumber = ZodFloat64

// NewZodFloat32 creates a new float32 schema
func NewZodFloat32(params ...SchemaParams) *ZodFloat32 {
	return NewZodFloat[float32](params...)
}

// NewZodFloat64 creates a new float64 schema
func NewZodFloat64(params ...SchemaParams) *ZodFloat64 {
	return NewZodFloat[float64](params...)
}

// Float32 creates a new float32 schema (main constructor)
func Float32(params ...SchemaParams) *ZodFloat32 {
	return NewZodFloat32(params...)
}

// Float64 creates a new float64 schema (main constructor)
func Float64(params ...SchemaParams) *ZodFloat64 {
	return NewZodFloat64(params...)
}

// Number creates a new number schema (alias for Float64)
func Number(params ...SchemaParams) *ZodFloat64 {
	return NewZodFloat64(params...)
}

// CoercedFloat32 creates a new float32 schema with coercion enabled
func CoercedFloat32(params ...SchemaParams) *ZodFloat32 {
	var modifiedParams SchemaParams
	if len(params) > 0 {
		modifiedParams = params[0]
	}
	modifiedParams.Coerce = true
	return NewZodFloat32(modifiedParams)
}

// CoercedFloat64 creates a new float64 schema with coercion enabled
func CoercedFloat64(params ...SchemaParams) *ZodFloat64 {
	var modifiedParams SchemaParams
	if len(params) > 0 {
		modifiedParams = params[0]
	}
	modifiedParams.Coerce = true
	return NewZodFloat64(modifiedParams)
}

// CoercedNumber creates a new number schema with coercion enabled (alias)
func CoercedNumber(params ...SchemaParams) *ZodFloat64 {
	return CoercedFloat64(params...)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodFloat[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodFloat[T]); ok {
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		if src.internals.Def != nil {
			z.internals.Def.MinValue = src.internals.Def.MinValue
			z.internals.Def.MaxValue = src.internals.Def.MaxValue
		}
	}
}

// createFloatCoerceFunc creates a coercion function for target floating-point type T
func createFloatCoerceFunc[T ZodFloatConstraint]() func(interface{}) (T, bool) {
	return func(value interface{}) (T, bool) {
		var zero T

		if result, ok := value.(T); ok {
			return result, true
		}

		if b, ok := value.(bool); ok {
			if b {
				return T(1), true
			}
			return T(0), true
		}

		if str, ok := value.(string); ok {
			str = strings.TrimSpace(str)
			if str == "" {
				return zero, false
			}
			if f, err := strconv.ParseFloat(str, 64); err == nil {
				switch any(zero).(type) {
				case float32:
					if f >= -math.MaxFloat32 && f <= math.MaxFloat32 {
						return T(f), true
					}
					return zero, false
				case float64:
					return T(f), true
				}
			}
			return zero, false
		}

		switch v := value.(type) {
		case int:
			return T(v), true
		case int8:
			return T(v), true
		case int16:
			return T(v), true
		case int32:
			return T(v), true
		case int64:
			return T(v), true
		case uint:
			return T(v), true
		case uint8:
			return T(v), true
		case uint16:
			return T(v), true
		case uint32:
			return T(v), true
		case uint64:
			return T(v), true
		case float32:
			switch any(zero).(type) {
			case float32:
				return T(v), true
			case float64:
				return T(v), true
			}
		case float64:
			switch any(zero).(type) {
			case float32:
				if v >= -math.MaxFloat32 && v <= math.MaxFloat32 {
					return T(v), true
				}
				return zero, false
			case float64:
				return T(v), true
			}
		}

		return zero, false
	}
}

// getFloatTypeName returns the type name for floating-point type T
func getFloatTypeName[T ZodFloatConstraint]() string {
	var zero T
	switch any(zero).(type) {
	case float32:
		return "float32"
	case float64:
		return "float64"
	default:
		return "unknown"
	}
}

// getFloatTypeBounds returns min, max values for floating-point type T
func getFloatTypeBounds[T ZodFloatConstraint]() (T, T) {
	var zero T
	switch any(zero).(type) {
	case float32:
		return any(float32(-math.MaxFloat32)).(T), any(float32(math.MaxFloat32)).(T)
	case float64:
		return any(float64(-math.MaxFloat64)).(T), any(float64(math.MaxFloat64)).(T)
	default:
		return zero, zero
	}
}

// extractFloatValue extracts floating-point value from input with smart handling
func extractFloatValue[T ZodFloatConstraint](input any) (T, bool, error) {
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
