package engine

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
)

// =============================================================================
// INPUT PREPROCESSING
// =============================================================================

// PreprocessInput performs pointer dereferencing and nil checking in a single reflect operation.
func PreprocessInput(input any) (dereferenced any, isNilPtr bool) {
	if input == nil {
		return nil, true
	}

	// Single reflect.ValueOf call to handle both nil checking and dereferencing
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, true
		}
		return rv.Elem().Interface(), false
	}
	return input, false
}

// =============================================================================
// CORE PARSING FUNCTIONS
// =============================================================================

// Parse is the core parsing function used by all schema types
// Provides unified parsing logic with proper error handling and type safety
func Parse[In, Out any](schema core.ZodType[In, Out], value any, ctx *core.ParseContext) (Out, error) {
	var result Out

	// Validate input parameters
	if schema == nil {
		rawIssue := issues.CreateCustomIssue("Schema cannot be nil", nil, value)
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return result, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	if ctx == nil {
		ctx = core.NewParseContext()
	}

	// Perform the actual parsing via schema's internal parse method
	parsedValue, err := schema.Parse(value, ctx)
	if err != nil {
		return result, err
	}

	// Type assertion with proper error handling
	if typedResult, ok := parsedValue.(Out); ok {
		return typedResult, nil
	}

	// Handle type conversion failure
	rawIssue := issues.CreateCustomIssue("type assertion failed: cannot convert parsed value to output type", nil, parsedValue)
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return result, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// MustParse is a convenience function that panics on parsing failure
// Use with caution - prefer Parse for production code
func MustParse[In, Out any](schema core.ZodType[In, Out], value any, ctx *core.ParseContext) Out {
	result, err := Parse[In, Out](schema, value, ctx)
	if err != nil {
		panic(err)
	}
	return result
}

// ParsePrimitive provides a zero-reflection, zero-allocation parsing path for
// primitive Go types (string, bool, numeric, …). It supports both value and
// pointer inputs while preserving the input shape in the returned value.
//
//	T        – concrete Go type to parse (bool, string, int64 …)
//	input    – raw user input, may be T or *T or anything that can be coerced
//	intr     – pointer to the schema internals (nilable, checks, bag …)
//	expected – constant describing the expected Zod type, purely for error msgs
//	validate – optional extra validator that applies Zod checks after type match
//	ctx      – parsing context for error path tracking (may be nil)
func ParsePrimitive[T any](
	input any,
	intr *core.ZodTypeInternals,
	expected core.ZodTypeCode,
	validate func(T, []core.ZodCheck, *core.ParseContext) error,
	ctx *core.ParseContext,
) (any, error) {
	if ctx == nil {
		ctx = core.NewParseContext()
	}

	// 1. Optimized nil / pointer preprocessing using single reflect operation
	deref, isNilPtr := PreprocessInput(input)
	if isNilPtr {
		if !intr.Nilable {
			raw := issues.CreateInvalidTypeIssue(string(expected), input)
			fin := issues.FinalizeIssue(raw, ctx, nil)
			return nil, issues.NewZodError([]core.ZodIssue{fin})
		}
		// Preserve original nil form - pointer nil stays pointer nil
		switch v := input.(type) {
		case *T:
			return v, nil
		default:
			var nilPtr *T = nil
			return nilPtr, nil
		}
	}

	// 2. Generic coercion when enabled via Bag["coerce"]
	if intr != nil && mapx.GetBoolDefault(intr.Bag, "coerce", false) && utils.IsPrimitiveType(expected) {
		// Try coercing the original input first
		if coerced, err := coerce.To[T](input); err == nil {
			if validate != nil {
				if err := validate(coerced, intr.Checks, ctx); err != nil {
					return nil, err
				}
			}
			return coerced, nil
		}
		// Try coercing the dereferenced value for pointer inputs
		if coerced, err := coerce.To[T](deref); err == nil {
			if validate != nil {
				if err := validate(coerced, intr.Checks, ctx); err != nil {
					return nil, err
				}
			}
			return coerced, nil
		}
	}

	// 3. Optimized type assertion order: check most common types first
	// Direct value match (handles both T and pointer types like *big.Int)
	if val, ok := input.(T); ok {
		if validate != nil {
			if err := validate(val, intr.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return val, nil
	}

	// 4. Direct pointer assertion (handles **T where T is already a pointer type)
	if ptr, ok := input.(*T); ok {
		if validate != nil {
			if err := validate(*ptr, intr.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return ptr, nil
	}

	// 5. Dereferenced value assertion (using preprocessed deref value)
	if val, ok := deref.(T); ok {
		if validate != nil {
			if err := validate(val, intr.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return val, nil
	}

	// 6. All attempts failed - create unified invalid-type error
	raw := issues.CreateInvalidTypeIssue(string(expected), input)
	fin := issues.FinalizeIssue(raw, ctx, nil)
	return nil, issues.NewZodError([]core.ZodIssue{fin})
}

// =============================================================================
// PARSEYPE TEMPLATE SYSTEM
// =============================================================================

// ParseType is the unified parsing engine for all schema types
// Handles type checking, coercion, validation, and error reporting
// T represents the expected value type
func ParseType[T any](
	input any,
	internals *core.ZodTypeInternals,
	expectedType core.ZodTypeCode,
	typeChecker func(any) (T, bool),
	pointerChecker func(any) (*T, bool),
	validator func(T, []core.ZodCheck, *core.ParseContext) error,
	ctx *core.ParseContext,
) (any, error) {
	// 1. Unified nil handling with compile-time type determination
	if input == nil {
		if !internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
			finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		// Compile-time check: if T is already a pointer type, return zero value
		// Otherwise, return a pointer to T's zero value
		var zero T
		if _, isPtr := any(zero).(*T); isPtr {
			return zero, nil // T is already a pointer type
		}
		return (*T)(nil), nil // T is not a pointer, return pointer to zero
	}

	// 2. Optimized pointer type matching
	if ptr, ok := pointerChecker(input); ok {
		if ptr == nil {
			if !internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
				finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
			}
			// Use same compile-time type determination as above
			var zero T
			if _, isPtr := any(zero).(*T); isPtr {
				return zero, nil
			}
			return (*T)(nil), nil
		}

		// Validate dereferenced value but preserve pointer identity
		val := *ptr

		if validator != nil {
			if err := validator(val, internals.Checks, ctx); err != nil {
				return nil, err
			}
		}

		return ptr, nil // Return original pointer to preserve identity
	}

	// 3. Direct type matching
	if value, ok := typeChecker(input); ok {
		if validator != nil {
			if err := validator(value, internals.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return value, nil // T → T (preserve original type)
	}

	// 4. Unified error creation
	rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
}
