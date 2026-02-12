package engine

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// ErrUnableToConvert indicates a type conversion failure.
var ErrUnableToConvert = errors.New("unable to convert value to expected type")

// tryDirectTypeMatch attempts direct and pointer type conversions between T and R.
func tryDirectTypeMatch[T any, R any](value any) (R, bool) {
	var zero R

	if r, ok := value.(R); ok {
		return r, true
	}

	// T → *T and *T → T conversions.
	switch any(zero).(type) {
	case *T:
		if v, ok := value.(T); ok {
			p := &v
			if r, ok := any(p).(R); ok {
				return r, true
			}
		}
	case T:
		if p, ok := value.(*T); ok && p != nil {
			if r, ok := any(*p).(R); ok {
				return r, true
			}
		}
	}

	// Fast path for common pointer type conversions.
	switch any(zero).(type) {
	case *string:
		if v, ok := value.(string); ok {
			if r, ok := any(&v).(R); ok {
				return r, true
			}
		}
	case *int64:
		if v, ok := value.(int64); ok {
			if r, ok := any(&v).(R); ok {
				return r, true
			}
		}
	case *float64:
		if v, ok := value.(float64); ok {
			if r, ok := any(&v).(R); ok {
				return r, true
			}
		}
	case *bool:
		if v, ok := value.(bool); ok {
			if r, ok := any(&v).(R); ok {
				return r, true
			}
		}
	}

	return zero, false
}

// parseTypedValue parses input into type R with validation and transformation.
// Returns an error for nil input, letting the caller handle modifiers.
func parseTypedValue[T any, R any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (R, error) {
	var zero R

	if input == nil {
		return zero, issues.CreateInvalidTypeError(expectedType, input, ctx)
	}

	if r, ok := tryDirectTypeMatch[T, R](input); ok {
		return validateAndReturn(r, internals, expectedType, ctx)
	}

	converted, err := convertToType[T](input, expectedType, ctx)
	if err != nil {
		return zero, err
	}

	r, ok := converted.(R)
	if !ok {
		return zero, issues.CreateInvalidTypeError(expectedType, input, ctx)
	}

	return validateAndReturn(r, internals, expectedType, ctx)
}

// validateAndReturn applies validation checks and transformation to a value.
func validateAndReturn[R any](
	value R,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (R, error) {
	if len(internals.Checks) > 0 {
		v, err := ApplyChecks(value, internals.Checks, ctx)
		if err != nil {
			return value, err
		}
		value = v
	}

	if internals.Transform != nil {
		transformed, err := internals.Transform(any(value), &core.RefinementContext{ParseContext: ctx})
		if err != nil {
			return value, err
		}
		if r, ok := transformed.(R); ok {
			return r, nil
		}
		return value, issues.CreateInvalidTypeError(expectedType, transformed, ctx)
	}

	return value, nil
}

// convertToType converts input to type T, trying direct match, pointer deref, then coercion.
func convertToType[T any](input any, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	if result, ok := input.(T); ok {
		return result, nil
	}

	if ptr, ok := input.(*T); ok && ptr != nil {
		return *ptr, nil
	}

	var zero T
	switch any(zero).(type) {
	case string:
		if str, err := coerce.ToString(input); err == nil {
			return any(str).(T), nil
		}
	case int:
		if i, err := coerce.ToInteger[int](input); err == nil {
			return any(i).(T), nil
		}
	case int8:
		if i8, err := coerce.ToInteger[int8](input); err == nil {
			return any(i8).(T), nil
		}
	case int16:
		if i16, err := coerce.ToInteger[int16](input); err == nil {
			return any(i16).(T), nil
		}
	case int32:
		if i32, err := coerce.ToInteger[int32](input); err == nil {
			return any(i32).(T), nil
		}
	case int64:
		if i64, err := coerce.ToInt64(input); err == nil {
			return any(i64).(T), nil
		}
	case uint:
		if u, err := coerce.ToInteger[uint](input); err == nil {
			return any(u).(T), nil
		}
	case uint8:
		if u8, err := coerce.ToInteger[uint8](input); err == nil {
			return any(u8).(T), nil
		}
	case uint16:
		if u16, err := coerce.ToInteger[uint16](input); err == nil {
			return any(u16).(T), nil
		}
	case uint32:
		if u32, err := coerce.ToInteger[uint32](input); err == nil {
			return any(u32).(T), nil
		}
	case uint64:
		if u64, err := coerce.ToInteger[uint64](input); err == nil {
			return any(u64).(T), nil
		}
	case float32:
		if f32, err := coerce.ToFloat[float32](input); err == nil {
			return any(f32).(T), nil
		}
	case float64:
		if f64, err := coerce.ToFloat64(input); err == nil {
			return any(f64).(T), nil
		}
	case bool:
		if b, err := coerce.ToBool(input); err == nil {
			return any(b).(T), nil
		}
	}

	return nil, issues.CreateInvalidTypeError(expectedType, input, ctx)
}

// isNilInput reports whether input is nil, using reflection only for nillable kinds.
func isNilInput[R any](input R) bool {
	if any(input) == nil {
		return true
	}
	rv := reflect.ValueOf(input)
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() { //nolint:exhaustive // only nillable kinds need explicit handling
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

// ParsePrimitive provides unified, type-safe parsing for primitive types.
// Handles optional/nilable/default/prefault/transform modifiers automatically.
func ParsePrimitive[T any, R any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	converter func(any, *core.ParseContext, core.ZodTypeCode) (R, error),
	ctx ...*core.ParseContext,
) (R, error) {
	pc := getOrCreateContext(ctx...)

	if r, handled, err := processModifiers[T](input, internals, expectedType, func(value any) (any, error) {
		return parsePrimitiveValue(value, internals, expectedType, validator, pc)
	}, pc); handled {
		if err != nil {
			var zero R
			return zero, err
		}
		// Default/DefaultFunc short-circuit: skip Transform.
		if isNilInput(input) && (internals.DefaultValue != nil || internals.DefaultFunc != nil) {
			return converter(r, pc, expectedType)
		}
		// Other handled cases: apply Transform if present.
		transformed, err := applyTransformIfPresent(r, internals, pc)
		if err != nil {
			var zero R
			return zero, err
		}
		return converter(transformed, pc, expectedType)
	} else if r != nil {
		// Prefault: use returned value as new input.
		input = r
	}

	r, err := parsePrimitiveValue[T](input, internals, expectedType, validator, pc)
	if err != nil {
		var zero R
		return zero, err
	}
	transformed, err := applyTransformIfPresent(r, internals, pc)
	if err != nil {
		var zero R
		return zero, err
	}
	return converter(transformed, pc, expectedType)
}

// ParsePrimitiveStrict provides type-safe parsing for primitive types with compile-time guarantees.
func ParsePrimitiveStrict[T any, R any](
	input R,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx ...*core.ParseContext,
) (R, error) {
	var zero R
	pc := getOrCreateContext(ctx...)

	// Fast path: no modifiers, return input directly.
	if !isNilInput(input) && len(internals.Checks) == 0 &&
		internals.Transform == nil && internals.DefaultValue == nil &&
		internals.PrefaultValue == nil && !internals.Optional &&
		!internals.Nilable && !internals.NonOptional &&
		internals.DefaultFunc == nil {
		return input, nil
	}

	// Nil handling via modifiers.
	if isNilInput(input) {
		r, handled, err := processModifiersStrict[T](input, internals, expectedType, func(v any) (any, error) {
			return parsePrimitiveValue[T](v, internals, expectedType, validator, pc)
		}, pc)
		if handled {
			if err != nil {
				return zero, err
			}
			return ConvertToConstraintType[T, R](r, pc, expectedType)
		}
		// Not handled but non-nil result means prefault value.
		if r != nil {
			validated, err := parsePrimitiveValue[T](r, internals, expectedType, validator, pc)
			if err != nil {
				return zero, err
			}
			return ConvertToConstraintType[T, R](validated, pc, expectedType)
		}
		return zero, issues.CreateInvalidTypeError(expectedType, input, pc)
	}

	// Direct validation path.
	if len(internals.Checks) > 0 {
		var tVal T
		var ok bool

		rt := reflect.TypeOf(input)
		if rt != nil && rt.Kind() == reflect.Ptr {
			rv := reflect.ValueOf(input)
			if !rv.IsNil() {
				if v, match := rv.Elem().Interface().(T); match {
					tVal = v
					ok = true
				}
			} else {
				// Nil pointer: delegate to modifiers.
				r, handled, err := processModifiers[T](nil, internals, expectedType, func(v any) (any, error) {
					return parsePrimitiveValue[T](v, internals, expectedType, validator, pc)
				}, pc)
				if handled {
					if err != nil {
						return zero, err
					}
					return ConvertToConstraintType[T, R](r, pc, expectedType)
				}
				return zero, issues.CreateInvalidTypeError(expectedType, nil, pc)
			}
			if !ok {
				if v, match := any(input).(T); match {
					tVal = v
					ok = true
				}
			}
		} else if v, match := any(input).(T); match {
			tVal = v
			ok = true
		} else {
			// Type mismatch: fall back to parsePrimitiveValue.
			r, err := parsePrimitiveValue[T](any(input), internals, expectedType, validator, pc)
			if err != nil {
				return zero, err
			}
			return ConvertToConstraintType[T, R](r, pc, expectedType)
		}

		if !ok {
			return zero, issues.CreateInvalidTypeError(expectedType, input, pc)
		}

		validated, err := validator(tVal, internals.Checks, pc)
		if err != nil {
			return zero, err
		}

		// Convert validated T back to R.
		var result R
		switch any(zero).(type) {
		case *T:
			p := &validated
			result = any(p).(R)
		case T:
			result = any(validated).(R)
		}

		// Apply transformation if present.
		if internals.Transform != nil {
			transformed, err := internals.Transform(any(result), &core.RefinementContext{ParseContext: pc})
			if err != nil {
				return result, err
			}
			if r, ok := transformed.(R); ok {
				return r, nil
			}
			return result, issues.CreateInvalidTypeError(expectedType, transformed, pc)
		}
		return result, nil
	}

	// No checks: just apply transformation if present.
	return validateAndReturn[R](input, internals, expectedType, pc)
}

// ParseComplex provides unified parsing for complex types (struct, slice, map, etc.).
// Handles optional/nilable/default/prefault/transform modifiers automatically.
func ParseComplex[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	typeExtractor func(any) (T, bool),
	ptrExtractor func(any) (*T, bool),
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx ...*core.ParseContext,
) (any, error) {
	pc := getOrCreateContext(ctx...)

	if r, handled, err := processModifiers[T](input, internals, expectedType, func(v any) (any, error) {
		return parseComplexValue[T](v, internals, expectedType, typeExtractor, ptrExtractor, validator, pc)
	}, pc); handled {
		if err != nil {
			return nil, err
		}
		// Default/DefaultFunc short-circuit: skip Transform.
		if isNilInput(input) && (internals.DefaultValue != nil || internals.DefaultFunc != nil) {
			return r, nil
		}
		return applyTransformIfPresent(r, internals, pc)
	} else if r != nil {
		input = r
	}

	r, err := parseComplexValue[T](input, internals, expectedType, typeExtractor, ptrExtractor, validator, pc)
	if err != nil {
		return nil, err
	}
	return applyTransformIfPresent(r, internals, pc)
}

// ParseComplexStrict provides type-safe parsing for complex types with compile-time guarantees.
func ParseComplexStrict[T any, R any](
	input R,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	typeExtractor func(any) (T, bool),
	ptrExtractor func(any) (*T, bool),
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx ...*core.ParseContext,
) (R, error) {
	var zero R
	pc := getOrCreateContext(ctx...)

	// Fast path: no modifiers, return input directly.
	// Struct types always need field validation.
	if !isNilInput(input) && len(internals.Checks) == 0 &&
		internals.Transform == nil && internals.DefaultValue == nil &&
		internals.PrefaultValue == nil && !internals.Optional &&
		!internals.Nilable && !internals.NonOptional &&
		internals.DefaultFunc == nil && expectedType != core.ZodTypeStruct {
		return input, nil
	}

	// Nil handling.
	if isNilInput(input) {
		if internals.Optional || internals.Nilable {
			return input, nil
		}

		r, handled, err := processModifiersStrict[T](input, internals, expectedType, func(v any) (any, error) {
			return parseComplexValue[T](v, internals, expectedType, typeExtractor, ptrExtractor, validator, pc)
		}, pc)
		if handled {
			if err != nil {
				return zero, err
			}
			if cr, ok := r.(R); ok {
				return cr, nil
			}
			return zero, issues.CreateInvalidTypeError(expectedType, r, pc)
		}

		// Prefault values: full parsing and validation.
		if internals.PrefaultValue != nil {
			r, err := ParseComplex[T](internals.PrefaultValue, internals, expectedType, typeExtractor, ptrExtractor, validator, pc)
			if err != nil {
				return zero, err
			}
			if cr, ok := r.(R); ok {
				return cr, nil
			}
		}

		if internals.PrefaultFunc != nil {
			r, err := ParseComplex[T](internals.PrefaultFunc(), internals, expectedType, typeExtractor, ptrExtractor, validator, pc)
			if err != nil {
				return zero, err
			}
			if cr, ok := r.(R); ok {
				return cr, nil
			}
		}

		return zero, issues.CreateNonOptionalError(pc)
	}

	// Validation-only fast path (skip for struct types).
	if len(internals.Checks) > 0 && internals.Transform == nil &&
		internals.DefaultValue == nil && internals.PrefaultValue == nil &&
		internals.DefaultFunc == nil && expectedType != core.ZodTypeStruct {
		var val T
		var extracted bool

		if p, ok := ptrExtractor(input); ok && p != nil {
			val = *p
			extracted = true
		} else if v, ok := typeExtractor(input); ok {
			val = v
			extracted = true
		}

		if extracted {
			if _, err := validator(val, internals.Checks, pc); err != nil {
				return zero, err
			}
			return input, nil
		}
	}

	// Fallback to regular complex parsing.
	r, err := ParseComplex[T](input, internals, expectedType, typeExtractor, ptrExtractor, validator, pc)
	if err != nil {
		return zero, err
	}
	if cr, ok := r.(R); ok {
		return cr, nil
	}
	return zero, issues.CreateInvalidTypeError(expectedType, r, pc)
}

// ConvertToConstraintType converts a parsed result to constraint type R (T, *T, or **T).
func ConvertToConstraintType[T any, R any](
	result any,
	ctx *core.ParseContext,
	expectedType core.ZodTypeCode,
) (R, error) {
	var zero R

	if r, ok := result.(R); ok {
		return r, nil
	}

	if result == nil {
		switch any(zero).(type) {
		case **T:
			return any((**T)(nil)).(R), nil
		case *T:
			return any((*T)(nil)).(R), nil
		case T:
			var t T
			return any(t).(R), nil
		}
	}

	switch any(zero).(type) {
	case **T:
		if pp, ok := result.(**T); ok {
			return any(pp).(R), nil
		}
		if p, ok := result.(*T); ok {
			return any(&p).(R), nil
		}
		if v, ok := result.(T); ok {
			p := &v
			return any(&p).(R), nil
		}
		if c, err := convertToType[T](result, expectedType, ctx); err == nil {
			if v, ok := c.(T); ok {
				p := &v
				return any(&p).(R), nil
			}
		}
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	case *T:
		if p, ok := result.(*T); ok {
			return any(p).(R), nil
		}
		if v, ok := result.(T); ok {
			return any(&v).(R), nil
		}
		if c, err := convertToType[T](result, expectedType, ctx); err == nil {
			if v, ok := c.(T); ok {
				return any(&v).(R), nil
			}
		}
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	case T:
		// R is value type T. The engine may return T or *T.
		if v, ok := result.(T); ok {
			return any(v).(R), nil
		}
		if p, ok := result.(*T); ok {
			if p == nil {
				var t T
				return any(t).(R), nil
			}
			return any(*p).(R), nil
		}
		if c, err := convertToType[T](result, expectedType, ctx); err == nil {
			if v, ok := c.(T); ok {
				return any(v).(R), nil
			}
		}
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	default:
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	}
}

// parsePrimitiveValue parses a primitive value with type checking and optional coercion.
func parsePrimitiveValue[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	if v, ok := input.(T); ok {
		return validateWithChecks(v, internals.Checks, validator, ctx)
	}

	if p, ok := input.(*T); ok {
		if p == nil {
			return handleNilPointer[T](internals, expectedType, ctx)
		}
		return validatePointer(*p, p, internals.Checks, validator, ctx)
	}

	if input == nil {
		if expectedType == core.ZodTypeNil {
			var zero T
			return validateWithChecks(zero, internals.Checks, validator, ctx)
		}
		return handleNilPointer[T](internals, expectedType, ctx)
	}

	// Pointer dereferencing via reflection (slower path).
	if deref, nilPtr := dereferencePointer(input); nilPtr {
		return handleNilPointer[T](internals, expectedType, ctx)
	} else if v, ok := deref.(T); ok {
		return validateWithChecks(v, internals.Checks, validator, ctx)
	}

	if internals.Coerce {
		if v, err := coerce.To[T](input); err == nil {
			return validateWithChecks(v, internals.Checks, validator, ctx)
		}
	}

	raw := issues.CreateInvalidTypeIssue(expectedType, input)
	raw.Inst = internals
	return nil, issues.NewZodError([]core.ZodIssue{issues.FinalizeIssue(raw, ctx, nil)})
}

// parseComplexValue parses a complex type using type and pointer extractors.
func parseComplexValue[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	typeExtractor func(any) (T, bool),
	ptrExtractor func(any) (*T, bool),
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	if input == nil {
		return handleNilComplex[T](internals, expectedType, ctx)
	}

	if p, ok := ptrExtractor(input); ok {
		if p == nil {
			return handleNilComplex[T](internals, expectedType, ctx)
		}
		return validatePointer(*p, p, internals.Checks, validator, ctx)
	}

	if v, ok := typeExtractor(input); ok {
		return validateValue(v, internals.Checks, validator, ctx, expectedType)
	}

	return nil, issues.CreateInvalidTypeError(expectedType, input, ctx)
}

// getOrCreateContext returns the first provided context, or creates a new one.
func getOrCreateContext(ctx ...*core.ParseContext) *core.ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return core.NewParseContext()
}

// dereferencePointer dereferences a pointer, returning the value and whether it was nil.
func dereferencePointer(input any) (any, bool) {
	if input == nil {
		return nil, true
	}

	switch v := input.(type) {
	case *string:
		if v == nil {
			return nil, true
		}
		return *v, false
	case *int64:
		if v == nil {
			return nil, true
		}
		return *v, false
	case *float64:
		if v == nil {
			return nil, true
		}
		return *v, false
	case *bool:
		if v == nil {
			return nil, true
		}
		return *v, false
	case *int:
		if v == nil {
			return nil, true
		}
		return *v, false
	}

	// Reflection fallback for other pointer types.
	rv := reflect.ValueOf(input)
	if rv.Kind() != reflect.Ptr {
		return input, false
	}
	if rv.IsNil() {
		return nil, true
	}
	return rv.Elem().Interface(), false
}

// validateValue validates a value, skipping validation for lazy types.
func validateValue[T any](
	value T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
	expectedType core.ZodTypeCode,
) (any, error) {
	if expectedType == core.ZodTypeLazy {
		return value, nil
	}
	if validator != nil && len(checks) > 0 {
		v, err := validator(value, checks, ctx)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return value, nil
}

// validateWithChecks validates a value and returns the result or error.
func validateWithChecks[T any](
	value T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	v, err := validator(value, checks, ctx)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// hasOverwriteCheck reports whether any check is an overwrite check.
func hasOverwriteCheck(checks []core.ZodCheck) bool {
	for _, c := range checks {
		if ci := c.Zod(); ci != nil && ci.Def != nil && ci.Def.Check == "overwrite" {
			return true
		}
	}
	return false
}

// validatePointerWithOverwrite applies overwrite checks to a pointer.
func validatePointerWithOverwrite[T any](
	ptr *T,
	checks []core.ZodCheck,
	ctx *core.ParseContext,
) (*T, bool) {
	if vp, err := ApplyChecks(ptr, checks, ctx); err == nil && vp != ptr {
		return vp, true
	}
	return ptr, false
}

// validatePointer validates a value through its pointer, handling overwrite transformations.
func validatePointer[T any](
	value T,
	ptr *T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	if validator == nil {
		return ptr, nil
	}

	if hasOverwriteCheck(checks) {
		if np, changed := validatePointerWithOverwrite(ptr, checks, ctx); changed {
			return np, nil
		}
	}

	v, err := validator(value, checks, ctx)
	if err != nil {
		return nil, err
	}
	*ptr = v
	return ptr, nil
}

// handleNilPointer handles nil pointer cases for primitive types.
func handleNilPointer[T any](internals *core.ZodTypeInternals, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	r, handled, err := processModifiers[T](nil, internals, expectedType, func(any) (any, error) {
		return parseTypedValue[T, *T](nil, internals, expectedType, ctx)
	}, ctx)
	if handled {
		return r, err
	}
	return nil, issues.CreateInvalidTypeError(expectedType, nil, ctx)
}

// handleNilComplex handles nil cases for complex types.
func handleNilComplex[T any](internals *core.ZodTypeInternals, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	r, handled, err := processModifiers[T](nil, internals, expectedType, func(any) (any, error) {
		return parseTypedValue[T, any](nil, internals, expectedType, ctx)
	}, ctx)
	if handled {
		return r, err
	}
	return nil, issues.CreateInvalidTypeError(expectedType, nil, ctx)
}

// convertResultToType converts a result to type T, handling pointer/value mismatches.
func convertResultToType[T any](result any) (T, error) {
	var zero T

	if v, ok := result.(T); ok {
		return v, nil
	}

	// Handle pointer/value mismatches using reflection.
	zt := reflect.TypeOf(zero)
	if zt != nil {
		rv := reflect.ValueOf(result)

		// T is pointer, got value: wrap it.
		if zt.Kind() == reflect.Ptr {
			et := zt.Elem()
			if rv.IsValid() && rv.Type() == et {
				p := reflect.New(et)
				p.Elem().Set(rv)
				if v, ok := p.Interface().(T); ok {
					return v, nil
				}
			}
		}

		// T is value, got pointer: dereference it.
		if rv.IsValid() && rv.Kind() == reflect.Ptr && rv.Type().Elem() == zt {
			if d := rv.Elem(); d.IsValid() {
				if v, ok := d.Interface().(T); ok {
					return v, nil
				}
			}
		}
	}

	return zero, fmt.Errorf("%w: value of type %T", ErrUnableToConvert, result)
}

// ApplyChecks validates a value against checks and applies transformations.
func ApplyChecks[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	if len(checks) == 0 {
		return value, nil
	}

	payload := core.NewParsePayload(value)
	r := RunChecksOnValue(value, checks, payload, ctx)

	if r.HasIssues() {
		return value, issues.NewZodError(issues.ConvertRawIssuesToIssues(r.Issues(), ctx))
	}

	if r.Value() == nil {
		var zero T
		return zero, nil
	}

	return convertResultToType[T](r.Value())
}
