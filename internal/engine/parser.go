package engine

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// Static error variables
var (
	ErrUnableToConvert = errors.New("unable to convert value to expected type")
)

// =============================================================================
// PUBLIC PARSING API
// =============================================================================

// tryDirectTypeMatch performs compile-time type matching optimization
// This function reduces runtime type assertions by leveraging Go's type system
func tryDirectTypeMatch[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type matching (fastest path)
	if result, ok := value.(R); ok {
		return result, true
	}

	// Optimized T → *T and *T → T conversions with reduced any() calls
	// Cache the zero type check to avoid repeated any() conversions
	zeroAsAny := any(zero)
	switch zeroAsAny.(type) {
	case *T:
		// R is *T, try to convert T to *T
		if val, ok := value.(T); ok {
			ptr := &val
			if result, ok := any(ptr).(R); ok {
				return result, true
			}
		}
	case T:
		// R is T, try to convert *T to T
		if ptr, ok := value.(*T); ok && ptr != nil {
			if result, ok := any(*ptr).(R); ok {
				return result, true
			}
		}
	}

	// Fast path for common pointer type conversions without repeated any() calls
	switch zeroAsAny.(type) {
	case *string:
		if str, ok := value.(string); ok {
			strPtr := &str
			if result, ok := any(strPtr).(R); ok {
				return result, true
			}
		}
	case *int64:
		if i64, ok := value.(int64); ok {
			i64Ptr := &i64
			if result, ok := any(i64Ptr).(R); ok {
				return result, true
			}
		}
	case *float64:
		if f64, ok := value.(float64); ok {
			f64Ptr := &f64
			if result, ok := any(f64Ptr).(R); ok {
				return result, true
			}
		}
	case *bool:
		if b, ok := value.(bool); ok {
			bPtr := &b
			if result, ok := any(bPtr).(R); ok {
				return result, true
			}
		}
	}

	return zero, false
}

// parseTypedValue is the main parsing function with optimized type handling
// This replaces the old parseCore function with clearer naming and better performance
func parseTypedValue[T any, R any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (R, error) {
	var zero R

	// Ultra-fast path: nil input handling
	if input == nil {
		// For nil input, we should not process modifiers here as it leads to recursion
		// Instead, return error and let the caller handle modifiers
		return zero, issues.CreateInvalidTypeError(expectedType, input, ctx)
	}

	// Fast path: direct type matching
	if result, ok := tryDirectTypeMatch[T, R](input); ok {
		// Apply validation and transformation, with Prefault fallback on failure
		validated, err := validateAndReturn[R](result, internals, expectedType, ctx)
		if err != nil {
			return zero, err
		}
		return validated, nil
	}

	// Standard path: type conversion + validation
	converted, err := convertToType[T](input, expectedType, ctx)
	if err != nil {
		return zero, err
	}

	result, ok := converted.(R)
	if !ok {
		return zero, issues.CreateInvalidTypeError(expectedType, input, ctx)
	}

	// Apply validation and transformation, with Prefault fallback on failure
	validated, err := validateAndReturn[R](result, internals, expectedType, ctx)
	if err != nil {
		return zero, err
	}
	return validated, nil
}

// validateAndReturn performs validation and applies transformation if needed
func validateAndReturn[R any](
	value R,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	ctx *core.ParseContext,
) (R, error) {
	// Apply validation checks if present
	if len(internals.Checks) > 0 {
		validatedValue, err := ApplyChecks[R](value, internals.Checks, ctx)
		if err != nil {
			return value, err
		}
		value = validatedValue
	}

	// Apply transformation if present
	if internals.Transform != nil {
		refCtx := &core.RefinementContext{ParseContext: ctx}
		transformed, err := internals.Transform(any(value), refCtx)
		if err != nil {
			return value, err
		}
		if finalResult, ok := transformed.(R); ok {
			return finalResult, nil
		}
		// Type mismatch after transformation
		return value, issues.CreateInvalidTypeError(expectedType, transformed, ctx)
	}

	return value, nil
}

// convertToType performs type conversion with coercion support
func convertToType[T any](input any, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	// Try direct type match first
	if result, ok := input.(T); ok {
		return result, nil
	}

	// Try pointer dereferencing
	if ptr, ok := input.(*T); ok && ptr != nil {
		return *ptr, nil
	}

	// Try type coercion using existing coerce package
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

// isNilInput performs compile-time nil checking to avoid reflection
func isNilInput[R any](input R) bool {
	// Fast path: direct nil comparison for interfaces and pointers
	// This handles the most common cases without reflection
	if any(input) == nil {
		return true
	}

	// Use reflection only when necessary for complex types
	v := reflect.ValueOf(input)
	if !v.IsValid() {
		return true
	}

	// Check for nil pointers, interfaces, slices, maps, channels, and functions
	switch v.Kind() { //nolint:exhaustive // only nillable kinds need explicit handling
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

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

	// Handle modifiers first
	if result, handled, err := processModifiers[T](input, internals, expectedType, func(value any) (any, error) {
		return parsePrimitiveValue[T](value, internals, expectedType, validator, parseCtx)
	}, parseCtx); handled {
		if err != nil {
			var zero R
			return zero, err
		}
		// Check if this is a Default/DefaultFunc short-circuit case
		// These should skip Transform completely
		isDefaultShortCircuit := (internals.DefaultValue != nil && isNilInput(input)) ||
			(internals.DefaultFunc != nil && isNilInput(input))

		if isDefaultShortCircuit {
			// Default/DefaultFunc values skip Transform - direct conversion only
			return converter(result, parseCtx, expectedType)
		}

		// For all other handled cases (Optional/Nilable nil, Unknown type, etc.)
		// apply Transform if present
		transformed, transformErr := applyTransformIfPresent(result, internals, parseCtx)
		if transformErr != nil {
			var zero R
			return zero, transformErr
		}
		return converter(transformed, parseCtx, expectedType)
	} else if result != nil {
		// Prefault case: use the returned value as new input and continue parsing
		input = result
	}

	// Regular parsing path (including prefault values)
	result, err := parsePrimitiveValue[T](input, internals, expectedType, validator, parseCtx)
	if err == nil {
		transformed, transformErr := applyTransformIfPresent(result, internals, parseCtx)
		if transformErr != nil {
			var zero R
			return zero, transformErr
		}
		return converter(transformed, parseCtx, expectedType)
	}

	// Return parsing error
	var zero R
	return zero, err
}

// ParsePrimitiveStrict provides optimized, type-safe parsing for primitive types with compile-time guarantees.
// Uses Go 1.24 generics to eliminate unnecessary type conversions and function call overhead.
// Implements zero-overhead fast paths for common validation scenarios.
//
//	T            – primitive Go type (bool, string, int64, float64, etc.)
//	R            – constraint type (T | *T) for type-safe output
//	input        – input already typed as R (compile-time guarantee)
//	internals    – schema internals containing validation rules and modifiers
//	expectedType – type code for error messages
//	validator    – validation function for applying checks to T values
//	ctx          – parsing context (optional variadic parameter)
func ParsePrimitiveStrict[T any, R any](
	input R,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx ...*core.ParseContext,
) (R, error) {
	var zero R
	parseCtx := getOrCreateContext(ctx...)

	// Ultra-fast path: no validation, no transformation, no modifiers
	// This is the most common case, return input directly with zero overhead
	// Reorder conditions to check most common cases first
	if !isNilInput(input) && len(internals.Checks) == 0 &&
		internals.Transform == nil && internals.DefaultValue == nil &&
		internals.PrefaultValue == nil && !internals.Optional &&
		!internals.Nilable && !internals.NonOptional &&
		internals.DefaultFunc == nil {
		return input, nil
	}

	// Optimized nil handling using processModifiersStrict
	if isNilInput(input) {
		result, handled, err := processModifiersStrict[T](input, internals, expectedType, func(newInput any) (any, error) {
			// For non-nil input from modifiers, use parsePrimitiveValue directly
			return parsePrimitiveValue[T](newInput, internals, expectedType, validator, parseCtx)
		}, parseCtx)
		if handled {
			if err != nil {
				return zero, err
			}
			// Use ConvertToConstraintType for proper type conversion from modifiers
			convertedResult, err := ConvertToConstraintType[T, R](result, parseCtx, expectedType)
			if err != nil {
				return zero, err
			}
			return convertedResult, nil
		}
		// If not handled but result is not nil, it means we have a prefault value
		// Process the prefault value through the full validation pipeline
		if result != nil {
			// Use parsePrimitiveValue to process the prefault value with full validation
			validatedResult, err := parsePrimitiveValue[T](result, internals, expectedType, validator, parseCtx)
			if err != nil {
				return zero, err
			}
			// Convert the validated result to R type
			convertedResult, err := ConvertToConstraintType[T, R](validatedResult, parseCtx, expectedType)
			if err != nil {
				return zero, err
			}
			return convertedResult, nil
		}
		return zero, issues.CreateInvalidTypeError(expectedType, input, parseCtx)
	}

	// Fast path: direct validation without type conversion
	if len(internals.Checks) > 0 {
		// Try to extract T value from input R
		var tValue T
		var hasValue bool

		// Check if input R can be converted to T for validation
		// Case 1: R is *T and we need to extract T value
		inputType := reflect.TypeOf(input)
		if inputType != nil && inputType.Kind() == reflect.Ptr {
			// input is a pointer, try to dereference it
			inputValue := reflect.ValueOf(input)
			if !inputValue.IsNil() {
				derefValue := inputValue.Elem().Interface()
				if val, ok := derefValue.(T); ok {
					tValue = val
					hasValue = true
				}
			} else {
				// Handle nil pointer case
				// Handle nil pointer using processModifiers
				result, handled, err := processModifiers[T](nil, internals, expectedType, func(newInput any) (any, error) {
					return parsePrimitiveValue[T](newInput, internals, expectedType, validator, parseCtx)
				}, parseCtx)
				if handled {
					if err != nil {
						return zero, err
					}
					// Use ConvertToConstraintType for proper type conversion from modifiers
					convertedResult, err := ConvertToConstraintType[T, R](result, parseCtx, expectedType)
					if err != nil {
						return zero, err
					}
					return convertedResult, nil
				}
				return zero, issues.CreateInvalidTypeError(expectedType, nil, parseCtx)
			}
			// If pointer dereference failed, try direct cast as fallback
			if !hasValue {
				if val, ok := any(input).(T); ok {
					tValue = val
					hasValue = true
				}
			}
		} else if val, ok := any(input).(T); ok {
			// R is T and input is T
			tValue = val
			hasValue = true
		} else {
			// Type mismatch, use parsePrimitiveValue for conversion
			result, err := parsePrimitiveValue[T](any(input), internals, expectedType, validator, parseCtx)
			if err != nil {
				return zero, err
			}
			// Convert result back to constraint type R
			convertedResult, err := ConvertToConstraintType[T, R](result, parseCtx, expectedType)
			if err != nil {
				return zero, err
			}
			return convertedResult, nil
		}

		if !hasValue {
			return zero, issues.CreateInvalidTypeError(expectedType, input, parseCtx)
		}

		// Execute validation
		validatedValue, err := validator(tValue, internals.Checks, parseCtx)
		if err != nil {
			return zero, err
		}

		// Convert back to R type
		var result R
		var zeroR R
		switch any(zeroR).(type) {
		case *T:
			ptr := &validatedValue
			result = any(ptr).(R)
		case T:
			result = any(validatedValue).(R)
		}

		// Apply transformation if present (validation already done above)
		if internals.Transform != nil {
			refCtx := &core.RefinementContext{ParseContext: parseCtx}
			transformed, err := internals.Transform(any(result), refCtx)
			if err != nil {
				return result, err
			}
			if finalResult, ok := transformed.(R); ok {
				return finalResult, nil
			}
			// Type mismatch after transformation
			return result, issues.CreateInvalidTypeError(expectedType, transformed, parseCtx)
		}
		return result, nil
	}

	// No validation needed, just apply transformation if present
	return validateAndReturn[R](input, internals, expectedType, parseCtx)
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

	// Handle modifiers first
	if result, handled, err := processModifiers[T](input, internals, expectedType, func(value any) (any, error) {
		return parseComplexValue[T](value, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
	}, parseCtx); handled {
		if err != nil {
			return nil, err
		}
		// Check if this is a Default/DefaultFunc short-circuit case
		// These should skip Transform completely
		isDefaultShortCircuit := (internals.DefaultValue != nil && isNilInput(input)) ||
			(internals.DefaultFunc != nil && isNilInput(input))

		if isDefaultShortCircuit {
			// Default/DefaultFunc values skip Transform - return directly
			return result, nil
		}

		// For all other handled cases (Optional/Nilable nil, Unknown type, etc.)
		// apply Transform if present
		return applyTransformIfPresent(result, internals, parseCtx)
	} else if result != nil {
		// Prefault case: use the returned value as new input and continue parsing
		input = result
	}

	// Regular parsing path (with potentially replaced input from prefault)
	result, err := parseComplexValue[T](input, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
	if err == nil {
		return applyTransformIfPresent(result, internals, parseCtx)
	}

	return nil, err
}

// ParseComplexStrict provides optimized, type-safe parsing for complex types with compile-time guarantees.
// Uses Go 1.24 generics to eliminate unnecessary type conversions and function call overhead.
// Implements zero-overhead fast paths for common validation scenarios.
//
//	T            – complex Go type (struct, slice, map, custom types, etc.)
//	R            – constraint type (T | *T) for type-safe output
//	input        – input already typed as R (compile-time guarantee)
//	internals    – schema internals containing validation rules and modifiers
//	expectedType – type code for error messages
//	typeExtractor – function to extract value of type T from input
//	ptrExtractor  – function to extract pointer to T from input
//	validator    – validation function for applying checks to T values
//	ctx          – parsing context (optional variadic parameter)
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
	parseCtx := getOrCreateContext(ctx...)

	// Ultra-fast path: no validation, no transformation, no modifiers
	// This is the most common case, return input directly with zero overhead
	// Skip ultra-fast path for struct types as they need field validation
	if !isNilInput(input) && len(internals.Checks) == 0 &&
		internals.Transform == nil && internals.DefaultValue == nil &&
		internals.PrefaultValue == nil && !internals.Optional &&
		!internals.Nilable && !internals.NonOptional &&
		internals.DefaultFunc == nil && expectedType != core.ZodTypeStruct {
		return input, nil
	}

	// Handle nil input with modifiers
	if isNilInput(input) {
		if internals.Optional || internals.Nilable {
			return input, nil // Return nil as-is for optional/nilable
		}

		// Use processModifiersStrict for consistent handling
		result, handled, err := processModifiersStrict[T](input, internals, expectedType, func(newInput any) (any, error) {
			// For non-nil input from modifiers, use parseComplexValue directly
			return parseComplexValue[T](newInput, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
		}, parseCtx)
		if handled {
			if err != nil {
				return zero, err
			}
			if convertedResult, ok := result.(R); ok {
				return convertedResult, nil
			}
			return zero, issues.CreateInvalidTypeError(expectedType, result, parseCtx)
		}

		// Try prefault values (replace input and continue parsing)
		if internals.PrefaultValue != nil {
			// Prefault requires full parsing and validation
			result, err := ParseComplex[T](internals.PrefaultValue, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
			if err != nil {
				return zero, err
			}
			if convertedResult, ok := result.(R); ok {
				return convertedResult, nil
			}
		}

		if internals.PrefaultFunc != nil {
			prefaultValue := internals.PrefaultFunc()
			// Prefault requires full parsing and validation
			result, err := ParseComplex[T](prefaultValue, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
			if err != nil {
				return zero, err
			}
			if convertedResult, ok := result.(R); ok {
				return convertedResult, nil
			}
		}

		return zero, issues.CreateNonOptionalError(parseCtx)
	}

	// Fast path for validation-only scenarios
	// Skip fast path for struct types as they need field validation
	if len(internals.Checks) > 0 && internals.Transform == nil &&
		internals.DefaultValue == nil && internals.PrefaultValue == nil &&
		internals.DefaultFunc == nil && expectedType != core.ZodTypeStruct {
		// Extract value for validation
		var valueToValidate T
		var extracted bool

		// Try pointer extraction first
		if ptr, ok := ptrExtractor(input); ok && ptr != nil {
			valueToValidate = *ptr
			extracted = true
		} else if value, ok := typeExtractor(input); ok {
			valueToValidate = value
			extracted = true
		}

		if extracted {
			_, err := validator(valueToValidate, internals.Checks, parseCtx)
			if err != nil {
				return zero, err
			}
			return input, nil // Return original input after successful validation
		}
	}

	// Fallback to regular complex parsing for complex scenarios
	result, err := ParseComplex[T](input, internals, expectedType, typeExtractor, ptrExtractor, validator, parseCtx)
	if err != nil {
		return zero, err
	}

	// Convert result back to constraint type R
	if convertedResult, ok := result.(R); ok {
		return convertedResult, nil
	}

	// This should not happen in well-formed schemas, but provide safety
	return zero, issues.CreateInvalidTypeError(expectedType, result, parseCtx)
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

	// Fast path: direct type match
	if directResult, ok := result.(R); ok {
		return directResult, nil
	}

	// Handle nil result
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

	// Type conversion logic
	switch any(zero).(type) {
	case **T:
		// R is double pointer type **T
		if dblPtr, ok := result.(**T); ok {
			return any(dblPtr).(R), nil
		}
		if ptr, ok := result.(*T); ok {
			dblPtr := &ptr
			return any(dblPtr).(R), nil
		}
		if val, ok := result.(T); ok {
			ptr := &val
			dblPtr := &ptr
			return any(dblPtr).(R), nil
		}
		// Try type conversion for T
		if converted, err := convertToType[T](result, expectedType, ctx); err == nil {
			if val, ok := converted.(T); ok {
				ptr := &val
				dblPtr := &ptr
				return any(dblPtr).(R), nil
			}
		}
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	case *T:
		// R is pointer type *T
		if ptr, ok := result.(*T); ok {
			return any(ptr).(R), nil
		}
		if val, ok := result.(T); ok {
			ptr := &val
			return any(ptr).(R), nil
		}
		// Try type conversion for T
		if converted, err := convertToType[T](result, expectedType, ctx); err == nil {
			if val, ok := converted.(T); ok {
				ptr := &val
				return any(ptr).(R), nil
			}
		}
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	case T:
		// R is value type T. The engine may return either T or *T (when pointer
		// identity was preserved). Handle both transparently.
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
		// Try type conversion for T
		if converted, err := convertToType[T](result, expectedType, ctx); err == nil {
			if val, ok := converted.(T); ok {
				return any(val).(R), nil
			}
		}
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	default:
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	}
}

// =============================================================================

// =============================================================================
// CORE PARSING LAYER
// =============================================================================

// parsePrimitiveValue performs core primitive type parsing with optimized type checking
// Uses fast path for direct type matches and falls back to coercion when needed
func parsePrimitiveValue[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	// Fast path: direct type match (most common case)
	if val, ok := performFastTypeCheck[T](input); ok {
		return validateWithPrefault(val, internals.Checks, validator, ctx)
	}

	// Handle pointer types efficiently with identity preservation
	if ptr, ok := input.(*T); ok {
		if ptr == nil {
			return handleNilPointer[T](internals, expectedType, ctx)
		}
		// Always validate through the original pointer so we can keep identity and
		// support Overwrite(*T) *T transformations consistently across all types.
		return validatePointer(*ptr, ptr, internals.Checks, validator, ctx)
	}

	// Handle nil input - special case for ZodTypeNil
	if input == nil {
		// For nil types, allow nil input to go through normal validation
		if expectedType == core.ZodTypeNil {
			// For nil type, treat nil as a valid value and validate it
			var nilValue T
			return validateWithPrefault(nilValue, internals.Checks, validator, ctx)
		}
		return handleNilPointer[T](internals, expectedType, ctx)
	}

	// Try pointer dereferencing with reflection (slower path)
	if deref, isNilPtr := dereferencePointer(input); isNilPtr {
		return handleNilPointer[T](internals, expectedType, ctx)
	} else if val, ok := performFastTypeCheck[T](deref); ok {
		return validateWithPrefault(val, internals.Checks, validator, ctx)
	}

	// Try coercion if enabled (slowest path)
	if shouldAttemptTypeCoercion(internals) {
		if coerced, err := coerce.To[T](input); err == nil {
			return validateWithPrefault(coerced, internals.Checks, validator, ctx)
		}
	}

	// All attempts failed - create error with internals for custom message support
	raw := issues.CreateInvalidTypeIssue(expectedType, input)
	raw.Inst = internals // Pass internals so custom error messages can be used
	final := issues.FinalizeIssue(raw, ctx, nil)
	return nil, issues.NewZodError([]core.ZodIssue{final})
}

// parseComplexValue handles complex type parsing with type extractors
// Prioritizes pointer extraction to preserve original pointer identity
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
		return validatePointer(*ptr, ptr, internals.Checks, validator, ctx)
	}

	// Try direct type extraction
	if value, ok := typeExtractor(input); ok {
		return validateValue(value, internals.Checks, validator, ctx, expectedType)
	}

	// All attempts failed
	return nil, issues.CreateInvalidTypeError(expectedType, input, ctx)
}

// =============================================================================
// UTILITY AND HELPER FUNCTIONS
// =============================================================================

// getOrCreateContext efficiently gets or creates a parse context for parsing operations.
// Returns existing context if provided, otherwise creates a new one.
func getOrCreateContext(ctx ...*core.ParseContext) *core.ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return core.NewParseContext()
}

// dereferencePointer efficiently handles pointer dereferencing with reflection.
// Returns the dereferenced value and whether the original pointer was nil.
func dereferencePointer(input any) (dereferenced any, isNilPtr bool) {
	// Fast path: direct nil check
	if input == nil {
		return nil, true
	}

	// Fast path: try common pointer types without reflection
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

	// Fallback to reflection for other pointer types
	rv := reflect.ValueOf(input)
	if rv.Kind() != reflect.Ptr {
		return input, false
	}

	if rv.IsNil() {
		return nil, true
	}

	return rv.Elem().Interface(), false
}

// shouldAttemptTypeCoercion determines if type coercion should be attempted based on schema settings.
// Coercion is opt-in at the type level (e.g., String, Int, etc.) for performance optimization.
func shouldAttemptTypeCoercion(internals *core.ZodTypeInternals) bool {
	// Coercion is opt-in at the type level (e.g., String, Int, etc.)
	return internals.Coerce
}

// validateValue validates a value using the provided validator function.
// Returns early if no checks are present to optimize performance.
func validateValue[T any](
	value T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
	expectedType core.ZodTypeCode,
) (any, error) {
	// Special handling for lazy types
	if expectedType == core.ZodTypeLazy {
		// For lazy types, skip validation and return value directly
		return value, nil
	}

	if validator != nil && len(checks) > 0 {
		validatedValue, err := validator(value, checks, ctx)
		if err != nil {
			return nil, err
		}
		return validatedValue, nil
	}
	return value, nil
}

// validateWithPrefault validates a value and returns it if successful.
// Provides validation for non-nil values with proper error handling.
func validateWithPrefault[T any](
	value T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	validatedValue, err := validator(value, checks, ctx)
	if err == nil {
		return validatedValue, nil
	}

	// For validation failures, return the error directly
	// Each type should handle its own prefault logic in its validator
	return nil, err
}

// hasOverwriteCheck determines if any overwrite checks are present in the check list.
// Used to optimize pointer validation by avoiding unnecessary operations.
func hasOverwriteCheck(checks []core.ZodCheck) bool {
	for _, check := range checks {
		if checkInternals := check.GetZod(); checkInternals != nil && checkInternals.Def != nil {
			if checkInternals.Def.Check == "overwrite" {
				return true
			}
		}
	}
	return false
}

// validatePointerWithOverwrite handles pointer validation for overwrite transformations
func validatePointerWithOverwrite[T any](
	originalPtr *T,
	checks []core.ZodCheck,
	ctx *core.ParseContext,
) (*T, bool) {
	if validatedPtr, err := ApplyChecks[*T](originalPtr, checks, ctx); err == nil {
		if validatedPtr != originalPtr {
			return validatedPtr, true
		}
	}
	return originalPtr, false
}

func validatePointer[T any](
	value T,
	originalPtr *T,
	checks []core.ZodCheck,
	validator func(T, []core.ZodCheck, *core.ParseContext) (T, error),
	ctx *core.ParseContext,
) (any, error) {
	if validator == nil {
		return originalPtr, nil
	}

	// Handle Overwrite transformations
	if hasOverwriteCheck(checks) {
		if newPtr, transformed := validatePointerWithOverwrite(originalPtr, checks, ctx); transformed {
			return newPtr, nil
		}
	}

	// Standard value validation
	validatedValue, err := validator(value, checks, ctx)
	if err != nil {
		return nil, err
	}
	*originalPtr = validatedValue
	return originalPtr, nil
}

// handleNilPointer handles nil pointer cases for primitive types.
func handleNilPointer[T any](internals *core.ZodTypeInternals, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	// Use processModifiers for nil input handling
	result, handled, err := processModifiers[T](nil, internals, expectedType, func(any) (any, error) {
		return parseTypedValue[T, *T](nil, internals, expectedType, ctx)
	}, ctx)
	if handled {
		return result, err
	}
	return nil, issues.CreateInvalidTypeError(expectedType, nil, ctx)
}

// handleNilComplex handles nil cases for complex types with compile-time type determination.
func handleNilComplex[T any](internals *core.ZodTypeInternals, expectedType core.ZodTypeCode, ctx *core.ParseContext) (any, error) {
	// Use processModifiers for nil input handling
	result, handled, err := processModifiers[T](nil, internals, expectedType, func(any) (any, error) {
		return parseTypedValue[T, any](nil, internals, expectedType, ctx)
	}, ctx)
	if handled {
		return result, err
	}
	return nil, issues.CreateInvalidTypeError(expectedType, nil, ctx)
}

// =============================================================================
// ERROR HANDLING LAYER
// =============================================================================

// convertResultToType performs optimized type conversion with minimal reflection usage.
func convertResultToType[T any](result any) (T, error) {
	var zero T

	// Fast path: direct type assertion (most common case)
	if castVal, ok := result.(T); ok {
		return castVal, nil
	}

	// Slower path: handle pointer/value mismatches using reflection
	zeroTyp := reflect.TypeOf(zero)
	if zeroTyp != nil {
		valRV := reflect.ValueOf(result)

		// Case 1: T is a pointer type, but we got a value – wrap it
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

		// Case 2: T is a value type, but we got *T – dereference it
		if valRV.IsValid() && valRV.Kind() == reflect.Ptr {
			elemTyp := valRV.Type().Elem()
			if elemTyp == zeroTyp {
				if deref := valRV.Elem(); deref.IsValid() {
					if converted, ok := deref.Interface().(T); ok {
						return converted, nil
					}
				}
			}
		}
	}

	return zero, fmt.Errorf("%w: value of type %T", ErrUnableToConvert, result)
}

// performFastTypeCheck performs optimized type checking with early return for common cases.
// Provides fast path for direct type matches to minimize reflection overhead.
func performFastTypeCheck[T any](input any) (T, bool) {
	if val, ok := input.(T); ok {
		return val, true
	}
	var zero T
	return zero, false
}

// ApplyChecks provides universal validation and transformation for any type
// Handles validation logic common across all types and applies transformations
// Returns the potentially modified value after running checks (e.g., overwrite transforms)
func ApplyChecks[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	if len(checks) == 0 {
		return value, nil
	}

	payload := core.NewParsePayload(value)
	result := RunChecksOnValue(value, checks, payload, ctx)

	if result.HasIssues() {
		return value, issues.NewZodError(issues.ConvertRawIssuesToIssues(result.GetIssues(), ctx))
	}

	if result.GetValue() == nil {
		var zero T
		return zero, nil
	}

	return convertResultToType[T](result.GetValue())
}
