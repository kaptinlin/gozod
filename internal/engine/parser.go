package engine

import (
	"errors"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// INPUT PREPROCESSING
// =============================================================================

// PreprocessInput dereferences one pointer layer (if any) and returns whether
// the original value is a nil pointer. The helper is intentionally minimal so
// that callers can keep full control over pointer semantics.
func PreprocessInput(input any) (dereferenced any, isNilPtr bool) {
	if input == nil {
		return nil, true
	}

	if reflectx.IsNilPointer(input) {
		return nil, true
	}

	// reflectx.Deref returns (v,true) for non-pointers, so it's zero-cost in
	// the common case where the user passes a value.
	dereferenced, _ = reflectx.Deref(input)
	return dereferenced, false
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

	// 1. Cheap nil / pointer preprocessing.
	deref, isNilPtr := PreprocessInput(input)
	if isNilPtr {
		if !intr.Nilable {
			raw := issues.CreateInvalidTypeIssue(string(expected), input)
			fin := issues.FinalizeIssue(raw, ctx, nil)
			return nil, issues.NewZodError([]core.ZodIssue{fin})
		}
		// Keep original nil form – pointer nil stays pointer nil so that
		// callers comparing against (*T)(nil) don't break.
		switch v := input.(type) {
		case *T:
			return v, nil
		default:
			var nilPtr *T = nil
			return nilPtr, nil
		}
	}

	// 2. Generic coercion when the schema opted-in via Bag["coerce"]. This
	// step intentionally precedes pointer/value assertions to mirror the
	// original ParseType semantics (coercion wins over inference).
	if intr != nil && mapx.GetBoolDefault(intr.Bag, "coerce", false) && utils.IsPrimitiveType(expected) {
		if coerced, err := coerce.To[T](input); err == nil {
			if validate != nil {
				if err := validate(coerced, intr.Checks, ctx); err != nil {
					return nil, err
				}
			}
			return coerced, nil
		}
		// If input is a pointer, try dereferencing before coercion.
		if ptr, ok := input.(*T); ok && ptr != nil {
			if coerced, err := coerce.To[T](*ptr); err == nil {
				if validate != nil {
					if err := validate(coerced, intr.Checks, ctx); err != nil {
						return nil, err
					}
				}
				return coerced, nil
			}
		}
	}

	// 3. Direct match: if the input already has type T (covers pointer types
	// like *big.Int as well as base scalars). This must precede the **ptr
	// path to avoid missing single-pointer matches.
	if val, ok := input.(T); ok {
		if validate != nil {
			if err := validate(val, intr.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return val, nil
	}

	// 4. Try direct pointer assertion to preserve identity (handles **T where
	// T itself is a pointer type, e.g., **big.Int).
	if ptr, ok := input.(*T); ok {
		if validate != nil {
			if err := validate(*ptr, intr.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return ptr, nil
	}

	// 5. Try value assertion using the dereferenced value.
	if val, ok := deref.(T); ok {
		if validate != nil {
			if err := validate(val, intr.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return val, nil
	}

	// 6. All attempts failed – create a unified invalid-type error.
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
	// 1. Unified nil handling
	if input == nil {
		if !internals.Nilable {
			rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
			finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
		// For pointer types, return nil; for non-pointer types, return (*T)(nil)
		var zero T
		if reflectx.IsPointer(zero) {
			return zero, nil
		}
		return (*T)(nil), nil
	}

	// 2. Smart type inference: check pointer type matching
	if ptr, ok := pointerChecker(input); ok {
		if ptr == nil {
			if !internals.Nilable {
				rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
				finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
				return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
			}
			// For pointer types, return nil; for non-pointer types, return a typed nil pointer
			var zero T
			if reflectx.IsPointer(zero) {
				return zero, nil
			}
			return (*T)(nil), nil
		}

		// Validate the dereferenced value but preserve pointer identity in result
		val := *ptr

		if validator != nil {
			if err := validator(val, internals.Checks, ctx); err != nil {
				return nil, err
			}
		}

		return ptr, nil // Return original pointer to preserve identity as expected in tests
	}

	// 3. Smart type inference: check direct type matching
	if value, ok := typeChecker(input); ok {
		if validator != nil {
			if err := validator(value, internals.Checks, ctx); err != nil {
				return nil, err
			}
		}
		return value, nil // T → T (keep original type)
	}

	// 4. Unified error creation
	rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
}

// =============================================================================
// UNIFIED VALIDATION ENGINE
// =============================================================================

// ParseInternal is the unified parsing engine used by all schema types
// Centralizes validation logic with proper error handling and type safety
// T represents the concrete value type being validated
func ParseInternal[T any](
	schema core.ZodType[any, any], // Schema used for validation
	input any, // Raw input value
	basePath []any, // Path context for error reporting
	ctx *core.ParseContext, // Global parsing context
	coercer func(any) (T, bool), // Type coercion function
	validator func(T, []core.ZodCheck, *core.ParseContext) error,
	checks []core.ZodCheck, // Validation checks to run
	singlePath bool, // Whether to include path in error reporting
) (any, error) {
	// 1. Type coercion attempt
	value, coerced := coercer(input)
	if !coerced {
		// Create invalid type issue for coercion failure
		expectedType := reflectx.ParsedType((*new(T)))
		rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// 2. Handle nil values after successful coercion using reflectx
	if reflectx.IsNil(value) {
		// Create nil-specific issue
		expectedType := reflectx.ParsedType((*new(T)))
		rawIssue := issues.CreateInvalidTypeIssue(string(expectedType), input)
		finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// 3. Run validation checks - use slicex to safely check for empty
	if validator != nil && !slicex.IsEmpty(checks) {
		if err := validator(value, checks, ctx); err != nil {
			// If validation error is already a ZodError, return it directly
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				return nil, err
			}

			// For other errors, wrap in ZodError
			rawIssue := issues.CreateCustomIssue(err.Error(), nil, input)
			finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
			return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
		}
	}

	return any(value), nil
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// ParseWithDefaults parses with default context
func ParseWithDefaults[In, Out any](schema core.ZodType[In, Out], value any) (Out, error) {
	ctx := core.NewParseContext()
	return Parse[In, Out](schema, value, ctx)
}

// MustParseWithDefaults is a convenience function that uses default context and panics on failure
// For quick validation in initialization or configuration parsing
func MustParseWithDefaults[In, Out any](schema core.ZodType[In, Out], value any) Out {
	ctx := core.NewParseContext()
	result, err := Parse[In, Out](schema, value, ctx)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseFlat parses with flat error formatting (simpler structure)
func ParseFlat[T any](input any, schema core.ZodType[T, T], ctx ...*core.ParseContext) (T, *issues.FlattenedError) {
	var parseCtx *core.ParseContext
	if !slicex.IsEmpty(ctx) {
		parseCtx = ctx[0]
	} else {
		parseCtx = core.NewParseContext()
	}

	result, err := Parse[T, T](schema, input, parseCtx)
	var flatErr *issues.FlattenedError = nil

	if err != nil {
		// Handle ZodError with flattening
		var zodErr *issues.ZodError
		if errors.As(err, &zodErr) {
			flatErr := issues.FlattenError(zodErr)
			return result, flatErr
		}

		// Handle non-ZodError case
		var zero T
		return zero, &issues.FlattenedError{
			FormErrors:  []string{err.Error()},
			FieldErrors: make(map[string][]string),
		}
	}

	return result, flatErr
}
