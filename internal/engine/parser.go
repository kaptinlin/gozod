package engine

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// PUBLIC PARSING API
// =============================================================================

// ParsePrimitive provides unified, type-safe parsing for primitive types.
// Automatically handles optional/nilable/default/prefault/transform modifiers with optimal performance.
//
//	T            – primitive Go type (bool, string, int64, float64, etc.)
//	R            – constraint type (T | *T) for type-safe output
//	input        – raw user input, supports T, *T, and coercible types
//	internals    – schema internals containing validation rules and modifiers
//	expectedType – type code for error messages
//	validator    – validation function for applying checks to T values
//	converter    – function to convert parsed result to constraint type R
//	ctx          – parsing context (optional variadic parameter)
func ParsePrimitive[T any, R any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	converter func(any, *core.ParseContext, core.ZodTypeCode) (R, error),
	ctx ...*core.ParseContext,
) (R, error) {
	parseCtx := getOrCreateContext(ctx...)

	// Fast path: handle modifiers first to avoid unnecessary processing
	if result, handled, err := processModifiers[T](input, internals, expectedType, func(value any) (any, error) {
		return parsePrimitiveValue[T](value, internals, expectedType, validator, parseCtx)
	}, parseCtx); handled {
		if err != nil {
			var zero R
			return zero, err
		}
		transformed, transformErr := applyTransformIfPresent(result, internals, parseCtx)
		if transformErr != nil {
			var zero R
			return zero, transformErr
		}
		return converter(transformed, parseCtx, expectedType)
	}

	// Regular parsing path
	result, err := parsePrimitiveValue[T](input, internals, expectedType, validator, parseCtx)
	if err == nil {
		transformed, transformErr := applyTransformIfPresent(result, internals, parseCtx)
		if transformErr != nil {
			var zero R
			return zero, transformErr
		}
		return converter(transformed, parseCtx, expectedType)
	}

	// Fallback to prefault if available
	fallbackResult, fallbackErr := tryPrefaultFallback[T](internals, func(value any) (any, error) {
		return parsePrimitiveValue[T](value, internals, expectedType, validator, parseCtx)
	}, err)
	if !errors.Is(fallbackErr, err) {
		// Prefault was attempted
		if fallbackErr != nil {
			var zero R
			return zero, fallbackErr
		}
		transformed, transformErr := applyTransformIfPresent(fallbackResult, internals, parseCtx)
		if transformErr != nil {
			var zero R
			return zero, transformErr
		}
		return converter(transformed, parseCtx, expectedType)
	}

	// No prefault, return original error
	var zero R
	return zero, err
}

// ConvertToConstraintType provides universal type conversion for constraint types (T | *T).
// This function handles the common pattern where a schema can output either T or *T.
//
//	T      – base type (bool, string, int64, float64, etc.)
//	R      – constraint type (must be T | *T)
//	result – parsed result from engine (T or nil)
//	ctx    – parse context for error reporting
//	expectedType – type code for error messages
func ConvertToConstraintType[T any, R any](
	result any,
	ctx *core.ParseContext,
	expectedType core.ZodTypeCode,
) (R, error) {
	var zero R
	switch any(zero).(type) {
	case *T:
		// R is pointer type *T
		if result == nil {
			return any((*T)(nil)).(R), nil
		}
		// CRITICAL FIX: Preserve pointer identity when input is already the correct pointer type
		if ptr, ok := result.(*T); ok {
			return any(ptr).(R), nil
		}
		if val, ok := result.(T); ok {
			ptr := &val
			return any(ptr).(R), nil
		}
		return zero, CreateInvalidTypeError(expectedType, result, ctx)
	case T:
		// R is value type T. The engine may return either T or *T (when pointer
		// identity was preserved). Handle both transparently.
		if result == nil {
			var t T
			return any(t).(R), nil
		}
		if val, ok := result.(T); ok {
			return any(val).(R), nil
		}
		if ptr, ok := result.(*T); ok {
			if ptr == nil {
				var t T
				return any(t).(R), nil
			}
			return any(*ptr).(R), nil
		}
		return zero, CreateInvalidTypeError(expectedType, result, ctx)
	default:
		return zero, CreateInvalidTypeError(expectedType, result, ctx)
	}
}

// ParseComplex provides unified parsing for complex types (struct, slice, map, interface{}, etc.).
// Automatically handles optional/nilable/default/prefault/transform modifiers with optimal performance.
//
//	T              – complex Go type (struct, slice, map, custom types, etc.)
//	input          – raw user input, supports T, *T, and convertible types
//	internals      – schema internals containing validation rules and modifiers
//	expectedType   – type code for error messages
//	typeExtractor  – function to extract value of type T from input
//	ptrExtractor   – function to extract pointer to T from input
//	validator      – optional validation function for applying checks
//	ctx            – parsing context (optional variadic parameter)
func ParseComplex[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	typeExtractor func(any) (T, bool),
	ptrExtractor func(any) (*T, bool),
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx ...*core.ParseContext,
) (any, error) {
	parseCtx := getOrCreateContext(ctx...)

	// Fast path: handle modifiers first to avoid unnecessary processing
	if result, handled, err := processModifiers[T](input, internals, expectedType, func(value any) (any, error) {
		return parseComplexValue[T](value, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
	}, parseCtx); handled {
		if err != nil {
			return nil, err
		}
		return applyTransformIfPresent(result, internals, parseCtx)
	}

	// Regular parsing path
	result, err := parseComplexValue[T](input, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
	if err == nil {
		return applyTransformIfPresent(result, internals, parseCtx)
	}

	// Fallback to prefault if available
	return tryPrefaultFallback[T](internals, func(value any) (any, error) {
		return parseComplexValue[T](value, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
	}, err)
}

// =============================================================================
// MODIFIER PROCESSING LAYER
// =============================================================================

// processModifiers handles optional/nilable/default/prefault logic with optimal performance.
// Returns (result, handled, error) where handled=true means processing is complete.
// DESIGN DECISION: Optional/Nilable takes priority over Default for nil inputs
// This differs from TypeScript Zod but matches our test expectations
func processModifiers[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	parseCore func(any) (any, error),
	ctx *core.ParseContext,
) (result any, handled bool, err error) {
	// Fast path: non-nil input doesn't need modifier processing
	if input != nil {
		return nil, false, nil
	}

	// NonOptional flag – nil not allowed; produce dedicated error immediately
	if internals.NonOptional {
		return nil, true, CreateNonOptionalError(ctx)
	}

	// DESIGN DECISION: For nil input, Optional/Nilable takes priority over Default
	// This ensures that .Optional().Default() and .Default().Optional() both return nil for nil input
	// This is different from TypeScript Zod where Default takes priority

	// Check Optional/Nilable first (highest priority for nil input)
	if internals.Optional || internals.Nilable {
		return nil, true, nil
	}

	// Try default values only if not optional/nilable
	if internals.DefaultValue != nil {
		result, err := parseCore(internals.DefaultValue)
		if err == nil {
			return result, true, nil
		}
		// Default failed, try prefault
		return tryPrefaultWithError[T](internals, parseCore, err)
	}

	// Try default function
	if internals.DefaultFunc != nil {
		defaultValue := internals.DefaultFunc()
		result, err := parseCore(defaultValue)
		if err == nil {
			return result, true, nil
		}
		// Default failed, try prefault
		return tryPrefaultWithError[T](internals, parseCore, err)
	}

	// No default or optional/nilable available, return type error
	rawIssue := issues.CreateInvalidTypeIssue(expectedType, input)
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return nil, true, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// applyTransformIfPresent applies transformation if configured, with performance optimization.
func applyTransformIfPresent(result any, internals *core.ZodTypeInternals, ctx *core.ParseContext) (any, error) {
	if internals.Transform == nil {
		return result, nil // Fast path: no transform
	}

	refCtx := &core.RefinementContext{ParseContext: ctx}
	return internals.Transform(result, refCtx)
}

// tryPrefaultFallback attempts prefault fallback when parsing fails.
func tryPrefaultFallback[T any](
	internals *core.ZodTypeInternals,
	parseCore func(any) (any, error),
	originalErr error,
) (any, error) {
	// Fast path: no prefault available
	if internals.PrefaultValue == nil && internals.PrefaultFunc == nil {
		return nil, originalErr
	}

	// Try prefault value first (more common)
	if internals.PrefaultValue != nil {
		return parseCore(internals.PrefaultValue)
	}

	// Try prefault function
	if internals.PrefaultFunc != nil {
		prefaultValue := internals.PrefaultFunc()
		return parseCore(prefaultValue)
	}

	return nil, originalErr
}

// tryPrefaultWithError is a helper for handling prefault after default failure.
func tryPrefaultWithError[T any](
	internals *core.ZodTypeInternals,
	parseCore func(any) (any, error),
	defaultErr error,
) (any, bool, error) {
	result, err := tryPrefaultFallback[T](internals, parseCore, defaultErr)
	if errors.Is(err, defaultErr) {
		// No prefault was attempted
		return nil, true, defaultErr
	}
	// Prefault was attempted
	return result, true, err
}

// =============================================================================
// LEGACY CORE VALUE PARSING LAYER (for backward compatibility during transition)
// =============================================================================

// parsePrimitiveValue handles primitive type parsing with optimized coercion support.
func parsePrimitiveValue[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	// Fast path: direct type match (most common case)
	if val, ok := input.(T); ok {
		return validateAndReturnWithPrefault(val, internals.Checks, validator, internals, expectedType, ctx)
	}

	// Handle pointer types efficiently with identity preservation
	if ptr, ok := input.(*T); ok {
		if ptr == nil {
			return handleNilPointer[T](internals, expectedType, ctx)
		}
		// Always validate through the original pointer so we can keep identity and
		// support Overwrite(*T) *T transformations consistently across all types.
		return validateAndReturnPtr(*ptr, ptr, internals.Checks, validator, ctx)
	}

	// Handle nil input
	if input == nil {
		return handleNilPointer[T](internals, expectedType, ctx)
	}

	// Try pointer dereferencing with reflection (slower path)
	if deref, isNilPtr := dereferencePointer(input); isNilPtr {
		return handleNilPointer[T](internals, expectedType, ctx)
	} else if val, ok := deref.(T); ok {
		return validateAndReturnWithPrefault(val, internals.Checks, validator, internals, expectedType, ctx)
	}

	// Try coercion if enabled (slowest path)
	if shouldTryCoercion(internals) {
		if coerced, err := coerce.To[T](input); err == nil {
			return validateAndReturnWithPrefault(coerced, internals.Checks, validator, internals, expectedType, ctx)
		}
		// Try coercing dereferenced value
		if deref, _ := dereferencePointer(input); deref != nil {
			if coerced, err := coerce.To[T](deref); err == nil {
				return validateAndReturnWithPrefault(coerced, internals.Checks, validator, internals, expectedType, ctx)
			}
		}
	}

	// All attempts failed
	return nil, CreateInvalidTypeError(expectedType, input, ctx)
}

// parseComplexValue handles complex type parsing with type extractors.
func parseComplexValue[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	typeExtractor func(any) (T, bool),
	ptrExtractor func(any) (*T, bool),
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	// Handle nil input first
	if input == nil {
		return handleNilComplex[T](internals, expectedType, ctx)
	}

	// Try pointer extraction first (preserves original pointer identity)
	if ptr, ok := ptrExtractor(input); ok {
		if ptr == nil {
			return handleNilComplex[T](internals, expectedType, ctx)
		}
		// Validate dereferenced value but return original pointer
		return validateAndReturnPtr(*ptr, ptr, internals.Checks, validator, ctx)
	}

	// Try direct type extraction
	if value, ok := typeExtractor(input); ok {
		return validateAndReturn(value, internals.Checks, validator, ctx)
	}

	// All attempts failed
	return nil, CreateInvalidTypeError(expectedType, input, ctx)
}

// =============================================================================
// UTILITY AND HELPER FUNCTIONS
// =============================================================================

// getOrCreateContext efficiently gets or creates a parse context.
func getOrCreateContext(ctx ...*core.ParseContext) *core.ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return core.NewParseContext()
}

// dereferencePointer efficiently handles pointer dereferencing with reflection.
func dereferencePointer(input any) (dereferenced any, isNilPtr bool) {
	if input == nil {
		return nil, true
	}

	rv := reflect.ValueOf(input)
	if rv.Kind() != reflect.Ptr {
		return input, false
	}

	if rv.IsNil() {
		return nil, true
	}

	return rv.Elem().Interface(), false
}

// shouldTryCoercion determines if coercion should be attempted for performance.
func shouldTryCoercion(internals *core.ZodTypeInternals) bool {
	// Coercion is opt-in at the type level (e.g., String, Int, etc.)
	return internals.Coerce
}

// validateAndReturn validates a value and returns it if successful.
func validateAndReturn[T any](
	value T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	if validator != nil && len(checks) > 0 {
		validatedValue, err := validator(value, checks, ctx)
		if err != nil {
			return nil, err
		}
		return validatedValue, nil
	}
	return value, nil
}

// validateAndReturnWithPrefault validates a value and returns it if successful,
// or attempts prefault fallback if validation fails.
func validateAndReturnWithPrefault[T any](
	value T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (any, error) {
	validatedValue, err := validator(value, checks, ctx)
	if err == nil {
		return validatedValue, nil
	}

	// Validation failed, try prefault
	if internals.PrefaultValue != nil {
		return validateAndReturn(internals.PrefaultValue.(T), checks, validator, ctx)
	}
	if internals.PrefaultFunc != nil {
		return validateAndReturn(internals.PrefaultFunc().(T), checks, validator, ctx)
	}
	return nil, err
}

// validateAndReturnPtr validates a dereferenced value but returns the original pointer.
func validateAndReturnPtr[T any](
	value T,
	originalPtr *T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	if validator != nil && len(checks) > 0 {
		// CONSERVATIVE FIX: For specific case of pointer-type Overwrite transformations,
		// try pointer validation first, but only use it if it produces a different result

		// Check if we have Overwrite checks and try pointer validation
		hasOverwrite := false
		for _, check := range checks {
			if checkInternals := check.GetZod(); checkInternals != nil && checkInternals.Def != nil {
				if checkInternals.Def.Check == "overwrite" {
					hasOverwrite = true
					break
				}
			}
		}

		if hasOverwrite {
			// Try pointer validation first - this handles func(*T) *T transformations
			if validatedPtr, err := ApplyChecks[*T](originalPtr, checks, ctx); err == nil {
				// Check if the pointer validation created a new pointer
				if validatedPtr != originalPtr {
					// Different pointer, transformation definitely occurred
					return validatedPtr, nil
				}
			}
			// Pointer validation failed or produced no change, continue with value validation
		}

		// Standard value validation for func(T) T transformations
		validatedValue, err := validator(value, checks, ctx)
		if err != nil {
			return nil, err
		}
		// Update the original pointer with the validated (potentially modified) value
		*originalPtr = validatedValue
	}
	return originalPtr, nil
}

// handleNilPointer handles nil pointer cases for primitive types.
func handleNilPointer[T any](internals *core.ZodTypeInternals, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	if !internals.Optional && !internals.Nilable {
		if internals.NonOptional {
			return nil, CreateNonOptionalError(ctx)
		}
		return nil, CreateInvalidTypeError(expectedType, nil, ctx)
	}
	// Return typed nil pointer
	var nilPtr *T = nil
	return nilPtr, nil
}

// handleNilComplex handles nil cases for complex types with compile-time type determination.
func handleNilComplex[T any](internals *core.ZodTypeInternals, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	if !internals.Optional && !internals.Nilable {
		if internals.NonOptional {
			return nil, CreateNonOptionalError(ctx)
		}
		return nil, CreateInvalidTypeError(expectedType, nil, ctx)
	}

	var zero T
	if _, isPtr := any(zero).(*T); isPtr {
		return zero, nil
	}
	return (*T)(nil), nil
}

// CreateInvalidTypeError creates a standardized invalid type error.
func CreateInvalidTypeError(expectedType core.ZodTypeCode, input any, ctx *core.ParseContext) error {
	rawIssue := issues.CreateInvalidTypeIssue(expectedType, input)
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return issues.NewZodError([]core.ZodIssue{finalIssue})
}

// ApplyChecks provides universal validation and transformation for any type.
// This function handles validation logic that is common across all types.
// It returns the potentially modified value after running checks (e.g., overwrite transforms).
//
// Renamed from ApplyChecks to better reflect its dual purpose:
// - Validates input according to schema rules
// - Applies transformations (like Overwrite) that modify the value
//
//	T         – any type (primitive, complex, composite, etc.)
//	value     – the value to validate and potentially transform
//	checks    – validation checks to run (including transformations)
//	ctx       – parse context for error reporting
func ApplyChecks[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	payload := core.NewParsePayload(value)
	result := RunChecksOnValue(value, checks, payload, ctx)
	if result.HasIssues() {
		return value, issues.NewZodError(issues.ConvertRawIssuesToIssues(result.GetIssues(), ctx))
	}

	if result.GetValue() == nil {
		var zero T
		return zero, nil
	}

	// Attempt direct assertion first.
	if castVal, ok := result.GetValue().(T); ok {
		return castVal, nil
	}

	// Handle pointer/value mismatches generically using reflection.
	var zero T
	zeroTyp := reflect.TypeOf(zero)

	if zeroTyp != nil {
		valRV := reflect.ValueOf(result.GetValue())

		// Case 1: T is a pointer type, but we got a value – wrap it.
		if zeroTyp.Kind() == reflect.Ptr {
			elemTyp := zeroTyp.Elem()
			if valRV.IsValid() && valRV.Type() == elemTyp {
				ptr := reflect.New(elemTyp)
				ptr.Elem().Set(valRV)
				if converted, ok := ptr.Interface().(T); ok {
					return converted, nil
				}
			}
		}

		// Case 2: T is a value type, but we got *T – dereference it.
		if valRV.IsValid() && valRV.Kind() == reflect.Ptr {
			elemTyp := valRV.Type().Elem()
			if elemTyp == zeroTyp {
				// Dereference pointer to get value
				if deref := valRV.Elem(); deref.IsValid() {
					if converted, ok := deref.Interface().(T); ok {
						return converted, nil
					}
				}
			}
		}
	}

	// Fallback: panic with clearer diagnostic instead of unsafe conversion panic.
	panic(fmt.Sprintf("ApplyChecks: unable to convert value of type %T to expected generic type", result.GetValue()))
}

// CreateNonOptionalError returns a standardized error for nonoptional nil input.
func CreateNonOptionalError(ctx *core.ParseContext) error {
	raw := issues.CreateNonOptionalIssue(nil)
	final := issues.FinalizeIssue(raw, ctx, nil)
	return issues.NewZodError([]core.ZodIssue{final})
}
