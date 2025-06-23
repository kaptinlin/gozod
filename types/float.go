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

//////////////////////////
// CORE TYPE DEFINITIONS
//////////////////////////

// ZodFloatConstraint defines supported floating-point types for generic implementation
type ZodFloatConstraint interface {
	~float32 | ~float64
}

// ZodFloatDef defines the configuration for generic floating-point validation
type ZodFloatDef[T ZodFloatConstraint] struct {
	core.ZodTypeDef
	Type     core.ZodTypeCode // Type identifier using type-safe constants
	MinValue T                // Type minimum value
	MaxValue T                // Type maximum value
	Checks   []core.ZodCheck  // Float-specific validation checks
}

// ZodFloatInternals contains generic floating-point validator internal state
type ZodFloatInternals[T ZodFloatConstraint] struct {
	core.ZodTypeInternals
	Def     *ZodFloatDef[T]            // Schema definition
	Checks  []core.ZodCheck            // Validation checks
	Isst    issues.ZodIssueInvalidType // Invalid type issue template
	Pattern *regexp.Regexp             // Float pattern (if any)
	Bag     map[string]any             // Additional metadata (minimum, maximum, format, coerce flag, etc.)
}

// ZodFloat represents a generic floating-point validation schema
type ZodFloat[T ZodFloatConstraint] struct {
	internals *ZodFloatInternals[T]
}

// GetInternals returns the internal state of the schema
func (z *ZodFloat[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for floating-point type conversion
func (z *ZodFloat[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToFloat[T](input)
	return result, err == nil
}

// Parse validates and parses input with smart type inference
func (z *ZodFloat[T]) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	var zero T
	typeName := getFloatTypeName(zero)

	return engine.ParsePrimitive[T](
		input,
		&z.internals.ZodTypeInternals,
		typeName,
		validateFloat[T],
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodFloat[T]) MustParse(input any, ctx ...*core.ParseContext) any {
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
func (z *ZodFloat[T]) Min(minimum T, params ...any) *ZodFloat[T] {
	check := checks.Gte(minimum, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Max adds maximum value validation
func (z *ZodFloat[T]) Max(maximum T, params ...any) *ZodFloat[T] {
	check := checks.Lte(maximum, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Gt adds greater than validation (exclusive)
func (z *ZodFloat[T]) Gt(value T, params ...any) *ZodFloat[T] {
	check := checks.Gt(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodFloat[T]) Gte(value T, params ...any) *ZodFloat[T] {
	check := checks.Gte(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Lt adds less than validation (exclusive)
func (z *ZodFloat[T]) Lt(value T, params ...any) *ZodFloat[T] {
	check := checks.Lt(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodFloat[T]) Lte(value T, params ...any) *ZodFloat[T] {
	check := checks.Lte(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Positive adds positive number validation (> 0)
func (z *ZodFloat[T]) Positive(params ...any) *ZodFloat[T] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodFloat[T]) Negative(params ...any) *ZodFloat[T] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0)
func (z *ZodFloat[T]) NonNegative(params ...any) *ZodFloat[T] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0)
func (z *ZodFloat[T]) NonPositive(params ...any) *ZodFloat[T] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple of validation
func (z *ZodFloat[T]) MultipleOf(value T, params ...any) *ZodFloat[T] {
	check := checks.MultipleOf(value, params...)
	result := engine.AddCheck(any(z).(core.ZodType[any, any]), check)
	return result.(*ZodFloat[T])
}

// Int adds integer validation (no decimal part)
func (z *ZodFloat[T]) Int(params ...any) core.ZodType[any, any] {
	// Create a custom check for integer validation using reflectx
	check := checks.NewCustom[any](func(v any) bool {
		if !reflectx.IsFloat(v) {
			return false
		}
		if floatVal, ok := reflectx.ExtractFloat(v); ok {
			return floatVal == math.Trunc(floatVal)
		}
		return false
	}, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Finite adds finite number validation (not NaN or Infinity)
func (z *ZodFloat[T]) Finite(params ...any) core.ZodType[any, any] {
	// Create a custom check for finite validation using reflectx
	check := checks.NewCustom[any](func(v any) bool {
		if !reflectx.IsFloat(v) {
			return false
		}
		if floatVal, ok := reflectx.ExtractFloat(v); ok {
			return !math.IsInf(floatVal, 0) && !math.IsNaN(floatVal)
		}
		return false
	}, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Safe adds safe number validation (within JavaScript safe integer range)
func (z *ZodFloat[T]) Safe(params ...any) core.ZodType[any, any] {
	const maxSafeInt = 1<<53 - 1
	const minSafeInt = -(1<<53 - 1)
	return z.Gte(T(minSafeInt), params...).Lte(T(maxSafeInt), params...)
}

//////////////////////////
// TRANSFORM METHODS
//////////////////////////

// Transform creates a type-safe transformation of floating-point values
func (z *ZodFloat[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	wrappedFn := func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart value extraction
		if reflectx.IsNil(input) {
			return (*T)(nil), nil
		}

		// Handle pointer dereferencing
		actualInput := input
		if reflectx.IsPointer(input) {
			if deref, ok := reflectx.Deref(input); ok {
				actualInput = deref
			} else {
				return (*T)(nil), nil // nil pointer
			}
		}

		// Try direct type match first
		if floatVal, ok := actualInput.(T); ok {
			return fn(floatVal, ctx)
		}

		// Use reflectx.ExtractFloat for unified extraction
		if floatVal, ok := reflectx.ExtractFloat(actualInput); ok {
			return fn(T(floatVal), ctx)
		}

		return nil, fmt.Errorf("expected float type, got %T", input)
	}
	return z.TransformAny(wrappedFn)
}

// TransformAny provides flexible transformation with any input/output types
func (z *ZodFloat[T]) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	transform := Transform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: any(transform).(core.ZodType[any, any]),
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

// Pipe operation for pipeline chaining
func (z *ZodFloat[T]) Pipe(schema core.ZodType[any, any]) core.ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(core.ZodType[any, any]),
		out: schema,
		def: core.ZodTypeDef{Type: "pipe"},
	}
}

//////////////////////////
// MODIFIER METHODS
//////////////////////////

// Optional adds an optional check to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Optional() core.ZodType[any, any] {
	return Optional(any(z).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(z).(core.ZodType[any, any]))
}

// Nullish combines optional and nilable
func (z *ZodFloat[T]) Nullish() core.ZodType[any, any] {
	return Nullish(any(z).(core.ZodType[any, any]))
}

// Refine adds a flexible validation function to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Refine(fn func(T) bool, params ...any) *ZodFloat[T] {
	result := z.RefineAny(func(v any) bool {
		// Use reflectx for smart value extraction
		if reflectx.IsNil(v) {
			return true // Let Nilable flag handle nil validation
		}

		// Handle pointer dereferencing
		actualValue := v
		if reflectx.IsPointer(v) {
			if deref, ok := reflectx.Deref(v); ok {
				actualValue = deref
			} else {
				return true // nil pointer
			}
		}

		// Try direct type match first
		if floatVal, ok := actualValue.(T); ok {
			return fn(floatVal)
		}

		// Use reflectx.ExtractFloat for unified extraction
		if floatVal, ok := reflectx.ExtractFloat(actualValue); ok {
			return fn(T(floatVal))
		}

		return false // Invalid type
	}, params...)
	return result.(*ZodFloat[T])
}

// RefineAny adds a flexible validation function to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	check := checks.NewCustom[any](fn, params...)
	return engine.AddCheck(any(z).(core.ZodType[any, any]), check)
}

// Check adds a modern validation function to the floating-point, returns ZodType support chain call
func (z *ZodFloat[T]) Check(fn core.CheckFn) core.ZodType[any, any] {
	custom := Custom(fn, core.SchemaParams{})
	custom.GetInternals().Parse = func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
		result, err := z.Parse(payload.Value, ctx)
		if err != nil {
			payload.Issues = append(payload.Issues, core.ZodRawIssue{
				Code:    core.InvalidType,
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
func (z *ZodFloat[T]) Unwrap() core.ZodType[any, any] {
	return any(z).(core.ZodType[any, any])
}

//////////////////////////
// VALIDATION FUNCTIONS
//////////////////////////

// validateFloat validates float values with checks (generic)
func validateFloat[T ZodFloatConstraint](value T, checks []core.ZodCheck, ctx *core.ParseContext) error {
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
func (s ZodFloatDefault[T]) Min(minimum T, params ...any) ZodFloatDefault[T] {
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
func (s ZodFloatDefault[T]) Max(maximum T, params ...any) ZodFloatDefault[T] {
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
func (s ZodFloatDefault[T]) Positive(params ...any) ZodFloatDefault[T] {
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
func (s ZodFloatDefault[T]) Refine(fn func(T) bool, params ...any) ZodFloatDefault[T] {
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

func (s ZodFloatDefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart value extraction
		if reflectx.IsNil(input) {
			return nil, fmt.Errorf("cannot transform nil float value")
		}

		// Handle pointer dereferencing
		actualInput := input
		if reflectx.IsPointer(input) {
			if deref, ok := reflectx.Deref(input); ok {
				actualInput = deref
			} else {
				return nil, fmt.Errorf("cannot transform nil float value")
			}
		}

		// Try direct type match first
		if floatVal, ok := actualInput.(T); ok {
			return fn(floatVal, ctx)
		}

		// Use reflectx.ExtractFloat for unified extraction
		if floatVal, ok := reflectx.ExtractFloat(actualInput); ok {
			return fn(T(floatVal), ctx)
		}

		return nil, fmt.Errorf("expected float type, got %T", input)
	})
}

// Optional adds an optional check to the floating-point, returns ZodType support chain call
func (s ZodFloatDefault[T]) Optional() core.ZodType[any, any] {
	return Optional(any(s).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the floating-point, returns ZodType support chain call
func (s ZodFloatDefault[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(s).(core.ZodType[any, any]))
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
func (f ZodFloatPrefault[T]) Positive(params ...any) ZodFloatPrefault[T] {
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
func (f ZodFloatPrefault[T]) Refine(fn func(T) bool, params ...any) ZodFloatPrefault[T] {
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
func (f ZodFloatPrefault[T]) Transform(fn func(T, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return f.TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
		// Use reflectx for smart value extraction
		if reflectx.IsNil(input) {
			return nil, fmt.Errorf("cannot transform nil float value")
		}

		// Handle pointer dereferencing
		actualInput := input
		if reflectx.IsPointer(input) {
			if deref, ok := reflectx.Deref(input); ok {
				actualInput = deref
			} else {
				return nil, fmt.Errorf("cannot transform nil float value")
			}
		}

		// Try direct type match first
		if floatVal, ok := actualInput.(T); ok {
			return fn(floatVal, ctx)
		}

		// Use reflectx.ExtractFloat for unified extraction
		if floatVal, ok := reflectx.ExtractFloat(actualInput); ok {
			return fn(T(floatVal), ctx)
		}

		return nil, fmt.Errorf("expected float type, got %T", input)
	})
}

// Optional adds an optional check to the floating-point, returns ZodType support chain call
func (f ZodFloatPrefault[T]) Optional() core.ZodType[any, any] {
	return Optional(any(f).(core.ZodType[any, any]))
}

// Nilable adds a nilable check to the floating-point, returns ZodType support chain call
func (f ZodFloatPrefault[T]) Nilable() core.ZodType[any, any] {
	return Nilable(any(f).(core.ZodType[any, any]))
}

//////////////////////////
// CONSTRUCTOR FUNCTIONS
//////////////////////////

// createZodFloatFromDef creates a ZodFloat from definition
func createZodFloatFromDef[T ZodFloatConstraint](def *ZodFloatDef[T]) *ZodFloat[T] {
	internals := &ZodFloatInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Checks:           def.Checks,
		Isst:             issues.ZodIssueInvalidType{Expected: def.Type},
		Pattern:          nil,
		Bag:              make(map[string]any),
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any, any] {
		floatDef := &ZodFloatDef[T]{
			ZodTypeDef: *newDef,
			Type:       def.Type,
			MinValue:   def.MinValue,
			MaxValue:   def.MaxValue,
			Checks:     newDef.Checks,
		}
		return createZodFloatFromDef(floatDef)
	}

	zodSchema := &ZodFloat[T]{internals: internals}

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

	engine.InitZodType(zodSchema, &def.ZodTypeDef)
	return zodSchema
}

// Float creates a new generic floating-point schema
func Float[T ZodFloatConstraint](params ...any) *ZodFloat[T] {
	var zero T
	typeName := getFloatTypeName(zero)
	minValue, maxValue := getFloatTypeBounds[T]()

	def := &ZodFloatDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeName,
			Checks: make([]core.ZodCheck, 0),
		},
		Type:     typeName,
		MinValue: minValue,
		MaxValue: maxValue,
		Checks:   make([]core.ZodCheck, 0),
	}

	schema := createZodFloatFromDef(def)

	if len(params) > 0 {
		param := params[0]

		// Handle different parameter types
		switch p := param.(type) {
		case string:
			// String parameter becomes Error field
			errorMap := issues.CreateErrorMap(p)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		case core.SchemaParams:
			// Handle core.SchemaParams
			if p.Error != nil {
				errorMap := issues.CreateErrorMap(p.Error)
				if errorMap != nil {
					def.Error = errorMap
					schema.internals.Error = errorMap
				}
			}
		}
	}

	return schema
}

// Concrete floating-point type definitions
type ZodFloat32 = ZodFloat[float32]
type ZodFloat64 = ZodFloat[float64]
type ZodNumber = ZodFloat64

// Float32 creates a new float32 schema
func Float32(params ...any) *ZodFloat32 {
	return Float[float32](params...)
}

// Float64 creates a new float64 schema
func Float64(params ...any) *ZodFloat64 {
	return Float[float64](params...)
}

// Number creates a new number schema (alias for Float64)
func Number(params ...any) *ZodFloat64 {
	return Float64(params...)
}

//////////////////////////
// UTILITY FUNCTIONS
//////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodFloat[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodFloat[T]); ok {
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]any)
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

// getFloatTypeName returns the type name for floating-point type T
func getFloatTypeName[T ZodFloatConstraint](zero T) core.ZodTypeCode {
	switch any(zero).(type) {
	case float32:
		return core.ZodTypeFloat32
	case float64:
		return core.ZodTypeFloat64
	default:
		return core.ZodTypeFloat64 // Default fallback
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
