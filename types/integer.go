package types

import (
	"fmt"
	"math"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// IntegerConstraint restricts values to supported integer types or their pointers.
type IntegerConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~*int | ~*int8 | ~*int16 | ~*int32 | ~*int64 | ~*uint | ~*uint8 | ~*uint16 | ~*uint32 | ~*uint64
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodIntegerDef defines the configuration for integer validation
type ZodIntegerDef struct {
	core.ZodTypeDef
}

// ZodIntegerInternals contains integer validator internal state
type ZodIntegerInternals struct {
	core.ZodTypeInternals
	Def *ZodIntegerDef // Schema definition
}

// ZodIntegerTyped represents an integer validation schema with dual generic parameters
// T = base type (int, int32, int64, etc.), R = constraint type (T or *T)
type ZodIntegerTyped[T IntegerConstraint, R any] struct {
	internals *ZodIntegerInternals
}

// ZodInteger represents a flexible integer validation schema that accepts any integer type
// This is a type alias for ZodIntegerTyped[int64, int64] to provide a unified interface
type ZodInteger[T IntegerConstraint, R any] = ZodIntegerTyped[T, R]

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodIntegerTyped[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodIntegerTyped[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodIntegerTyped[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce implements Coercible interface for integer type conversion
func (z *ZodIntegerTyped[T, R]) Coerce(input any) (any, bool) {
	var zero T
	switch any(zero).(type) {
	case int, *int:
		result, err := coerce.ToInteger[int](input)
		return result, err == nil
	case int8, *int8:
		result, err := coerce.ToInteger[int8](input)
		return result, err == nil
	case int16, *int16:
		result, err := coerce.ToInteger[int16](input)
		return result, err == nil
	case int32, *int32:
		result, err := coerce.ToInteger[int32](input)
		return result, err == nil
	case int64, *int64:
		result, err := coerce.ToInteger[int64](input)
		return result, err == nil
	case uint, *uint:
		result, err := coerce.ToInteger[uint](input)
		return result, err == nil
	case uint8, *uint8:
		result, err := coerce.ToInteger[uint8](input)
		return result, err == nil
	case uint16, *uint16:
		result, err := coerce.ToInteger[uint16](input)
		return result, err == nil
	case uint32, *uint32:
		result, err := coerce.ToInteger[uint32](input)
		return result, err == nil
	case uint64, *uint64:
		result, err := coerce.ToInteger[uint64](input)
		return result, err == nil
	default:
		// Fallback to int64
		result, err := coerce.ToInteger[int64](input)
		return result, err == nil
	}
}

// Parse returns a value that matches the constraint type R with full type safety.
func (z *ZodIntegerTyped[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero T

	// Determine the type code based on T
	var typeCode core.ZodTypeCode
	switch any(zero).(type) {
	case int, *int:
		typeCode = core.ZodTypeInt
	case int8, *int8:
		typeCode = core.ZodTypeInt8
	case int16, *int16:
		typeCode = core.ZodTypeInt16
	case int32, *int32:
		typeCode = core.ZodTypeInt32
	case int64, *int64:
		typeCode = core.ZodTypeInt64
	case uint, *uint:
		typeCode = core.ZodTypeUint
	case uint8, *uint8:
		typeCode = core.ZodTypeUint8
	case uint16, *uint16:
		typeCode = core.ZodTypeUint16
	case uint32, *uint32:
		typeCode = core.ZodTypeUint32
	case uint64, *uint64:
		typeCode = core.ZodTypeUint64
	default:
		// Fallback to int64
		typeCode = core.ZodTypeInt64
	}

	// Determine if we have a pointer constraint type R to enable pointer identity preservation
	var zeroR R
	isPointerConstraint := false
	switch any(zeroR).(type) {
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		isPointerConstraint = true
	}

	// Temporarily enable Optional flag for pointer constraint types to ensure pointer identity preservation in ParsePrimitive
	originalInternals := z.internals
	if isPointerConstraint && !originalInternals.ZodTypeInternals.Optional && !originalInternals.ZodTypeInternals.Nilable {
		// Create a copy of internals with Optional flag temporarily enabled
		tempInternals := *originalInternals
		tempInternals.ZodTypeInternals.SetOptional(true)
		z.internals = &tempInternals

		// Restore original internals after parsing
		defer func() {
			z.internals = originalInternals
		}()
	}

	// Use ParsePrimitive with custom validator and converter functions
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		typeCode,
		// Validator function - validates the base type T
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
			// Apply all integer-specific checks
			return engine.ApplyChecks[T](value, checks, ctx)
		},
		// Converter function - converts validated T to constraint type R
		func(result any, ctx *core.ParseContext, expectedType core.ZodTypeCode) (R, error) {
			var zeroR R

			// Handle nil input for optional/nilable schemas
			if result == nil {
				internals := originalInternals // Use original internals for nil checking
				if internals.ZodTypeInternals.Optional || internals.ZodTypeInternals.Nilable {
					return zeroR, nil
				}
				return zeroR, fmt.Errorf("integer value cannot be nil")
			}

			// First, try to convert result directly to constraint type R (preserves pointer identity)
			if directValue, ok := result.(R); ok {
				return directValue, nil
			}

			// Special handling for pointer constraint types to preserve pointer identity
			var zeroPtr R
			switch any(zeroPtr).(type) {
			case *int:
				if ptr, ok := result.(*int); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *int8:
				if ptr, ok := result.(*int8); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *int16:
				if ptr, ok := result.(*int16); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *int32:
				if ptr, ok := result.(*int32); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *int64:
				if ptr, ok := result.(*int64); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *uint:
				if ptr, ok := result.(*uint); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *uint8:
				if ptr, ok := result.(*uint8); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *uint16:
				if ptr, ok := result.(*uint16); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *uint32:
				if ptr, ok := result.(*uint32); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *uint64:
				if ptr, ok := result.(*uint64); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			}

			// Fallback: try to convert to base type T and then to constraint type R
			if value, ok := result.(T); ok {
				// Convert base type T to constraint type R (this will create new pointers for pointer types)
				return convertToIntegerConstraintType[T, R](value), nil
			}

			// Handle pointer dereferencing if needed (for cases where input pointer doesn't match expected type)
			switch v := result.(type) {
			case *int:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *int8:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *int16:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *int32:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *int64:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *uint:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *uint8:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *uint16:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *uint32:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			case *uint64:
				if v != nil {
					if intVal, ok := any(*v).(T); ok {
						return convertToIntegerConstraintType[T, R](intVal), nil
					}
				}
			}

			return zeroR, fmt.Errorf("type conversion failed: expected %T but got %T", *new(T), result)
		},
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodIntegerTyped[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodIntegerTyped[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// ZODGENERICINTEGER MODIFIER METHODS
// =============================================================================

// withPtrInternals creates a new ZodIntegerTyped instance with pointer constraint type *T.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer constraint.
func (z *ZodIntegerTyped[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodIntegerTyped[T, *T] {
	return &ZodIntegerTyped[T, *T]{internals: &ZodIntegerInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// Optional returns a schema that accepts the base type T or nil, with constraint type *T.
func (z *ZodIntegerTyped[T, R]) Optional() *ZodIntegerTyped[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts the base type T or nil, with constraint type *T.
func (z *ZodIntegerTyped[T, R]) Nilable() *ZodIntegerTyped[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility.
func (z *ZodIntegerTyped[T, R]) Nullish() *ZodIntegerTyped[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes optional flag and returns value constraint (T).
// It is the counterpart of Optional() when you need to revert a previously optional schema
// back to a required field while keeping strong type safety.
func (z *ZodIntegerTyped[T, R]) NonOptional() *ZodIntegerTyped[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
	// ensure field is required and mark NonOptional for custom error reporting
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodIntegerTyped[T, T]{
		internals: &ZodIntegerInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default keeps the current generic type T.
func (z *ZodIntegerTyped[T, R]) Default(v int64) *ZodIntegerTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic type T.
func (z *ZodIntegerTyped[T, R]) DefaultFunc(fn func() int64) *ZodIntegerTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current generic type T.
func (z *ZodIntegerTyped[T, R]) Prefault(v int64) *ZodIntegerTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type R.
func (z *ZodIntegerTyped[T, R]) PrefaultFunc(fn func() R) *ZodIntegerTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this integer schema in the global registry.
func (z *ZodIntegerTyped[T, R]) Meta(meta core.GlobalMeta) *ZodIntegerTyped[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// VALIDATION METHODS (ONLY ZODS SUPPORTED METHODS)
// =============================================================================

// Min adds minimum value validation (alias for Gte)
func (z *ZodIntegerTyped[T, R]) Min(minimum int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.Gte(minimum, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max adds maximum value validation (alias for Lte)
func (z *ZodIntegerTyped[T, R]) Max(maximum int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.Lte(maximum, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gt adds greater than validation (exclusive)
func (z *ZodIntegerTyped[T, R]) Gt(value int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.Gt(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodIntegerTyped[T, R]) Gte(value int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.Gte(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lt adds less than validation (exclusive)
func (z *ZodIntegerTyped[T, R]) Lt(value int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.Lt(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodIntegerTyped[T, R]) Lte(value int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.Lte(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Positive adds positive number validation (> 0)
func (z *ZodIntegerTyped[T, R]) Positive(params ...any) *ZodIntegerTyped[T, R] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodIntegerTyped[T, R]) Negative(params ...any) *ZodIntegerTyped[T, R] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0, alias for nonnegative)
func (z *ZodIntegerTyped[T, R]) NonNegative(params ...any) *ZodIntegerTyped[T, R] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0, alias for nonpositive)
func (z *ZodIntegerTyped[T, R]) NonPositive(params ...any) *ZodIntegerTyped[T, R] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple of validation
func (z *ZodIntegerTyped[T, R]) MultipleOf(value int64, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.MultipleOf(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Step adds step validation (alias for MultipleOf to match Zod)
func (z *ZodIntegerTyped[T, R]) Step(step int64, params ...any) *ZodIntegerTyped[T, R] {
	return z.MultipleOf(step, params...)
}

// Safe adds safe integer validation (within JavaScript safe integer range)
func (z *ZodIntegerTyped[T, R]) Safe(params ...any) *ZodIntegerTyped[T, R] {
	const maxSafeInt = 1<<53 - 1
	const minSafeInt = -(1<<53 - 1)
	return z.Gte(minSafeInt, params...).Lte(maxSafeInt, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
// Integer types implement direct extraction of int64 values for transformation.
func (z *ZodIntegerTyped[T, R]) Transform(fn func(int64, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	// WrapFn Pattern: Create wrapper function for type-safe extraction
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		int64Value := extractIntegerToInt64[T, R](input) // Use existing extraction logic
		return fn(int64Value, ctx)
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodIntegerTyped[T, R]) Overwrite(transform func(T) T, params ...any) *ZodIntegerTyped[T, R] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToIntegerType[T](input)
		if !ok {
			// If conversion fails, return original value
			return input
		}

		// Apply transformation directly on type T
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using the WrapFn pattern.
// Instead of using adapter structures, this creates a target function that handles type conversion.
func (z *ZodIntegerTyped[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	// WrapFn Pattern: Create target function for type conversion and validation
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		// Extract int64 value from constraint type R
		int64Value := extractIntegerToInt64[T, R](input)
		// Apply target schema to the extracted int64
		return target.Parse(int64Value, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// TYPE CONVERSION
// =============================================================================

// extractIntegerToInt64 converts constraint type R to int64 for WrapFn pattern transformations
func extractIntegerToInt64[T IntegerConstraint, R any](value R) int64 {
	// First extract the base type T from constraint type R
	baseValue := extractIntegerValue[T, R](value)

	// Then convert T to int64
	switch v := any(baseValue).(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		if v > math.MaxInt64 {
			return 0
		}
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		if v > math.MaxInt64 {
			return 0
		}
		return int64(v)
	default:
		return 0
	}
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation using the base type T instead of constraint
// type T. The callback will be executed even when the value is nil (for pointer
// schemas) to align with Zod v4 semantics.
func (z *ZodIntegerTyped[T, R]) Refine(fn func(T) bool, params ...any) *ZodIntegerTyped[T, R] {
	wrapper := func(v any) bool {
		converted, ok := convertToIntegerType[T](v)
		if !ok {
			return false
		}
		return fn(converted)
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// convertToIntegerType converts only matching integer values to the target integer type T with strict type checking
func convertToIntegerType[T IntegerConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		// Handle nil values for pointer types
		switch any(zero).(type) {
		case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
			return zero, true // zero value for pointer types is nil
		default:
			return zero, false // nil not allowed for value types
		}
	}

	// Only accept matching integer types - no cross-type conversion
	switch any(zero).(type) {
	case int:
		if val, ok := v.(int); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*int); ok && val != nil {
			return any(*val).(T), true
		}
	case int8:
		if val, ok := v.(int8); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*int8); ok && val != nil {
			return any(*val).(T), true
		}
	case int16:
		if val, ok := v.(int16); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*int16); ok && val != nil {
			return any(*val).(T), true
		}
	case int32:
		if val, ok := v.(int32); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*int32); ok && val != nil {
			return any(*val).(T), true
		}
	case int64:
		if val, ok := v.(int64); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*int64); ok && val != nil {
			return any(*val).(T), true
		}
	case uint:
		if val, ok := v.(uint); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*uint); ok && val != nil {
			return any(*val).(T), true
		}
	case uint8:
		if val, ok := v.(uint8); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*uint8); ok && val != nil {
			return any(*val).(T), true
		}
	case uint16:
		if val, ok := v.(uint16); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*uint16); ok && val != nil {
			return any(*val).(T), true
		}
	case uint32:
		if val, ok := v.(uint32); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*uint32); ok && val != nil {
			return any(*val).(T), true
		}
	case uint64:
		if val, ok := v.(uint64); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*uint64); ok && val != nil {
			return any(*val).(T), true
		}
	// Pointer types
	case *int:
		if val, ok := v.(int); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*int); ok {
			return any(val).(T), true
		}
	case *int8:
		if val, ok := v.(int8); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*int8); ok {
			return any(val).(T), true
		}
	case *int16:
		if val, ok := v.(int16); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*int16); ok {
			return any(val).(T), true
		}
	case *int32:
		if val, ok := v.(int32); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*int32); ok {
			return any(val).(T), true
		}
	case *int64:
		if val, ok := v.(int64); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*int64); ok {
			return any(val).(T), true
		}
	case *uint:
		if val, ok := v.(uint); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*uint); ok {
			return any(val).(T), true
		}
	case *uint8:
		if val, ok := v.(uint8); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*uint8); ok {
			return any(val).(T), true
		}
	case *uint16:
		if val, ok := v.(uint16); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*uint16); ok {
			return any(val).(T), true
		}
	case *uint32:
		if val, ok := v.(uint32); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*uint32); ok {
			return any(val).(T), true
		}
	case *uint64:
		if val, ok := v.(uint64); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*uint64); ok {
			return any(val).(T), true
		}
	}

	return zero, false // Reject all non-matching integer types
}

// RefineAny adds flexible custom validation logic
func (z *ZodIntegerTyped[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodIntegerTyped[T, R] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withInternals creates a new ZodInt instance that keeps the original generic type T.
// Used by modifiers that retain the original type, such as Default, Prefault, and validation methods.
func (z *ZodIntegerTyped[T, R]) withInternals(in *core.ZodTypeInternals) *ZodIntegerTyped[T, R] {
	return &ZodIntegerTyped[T, R]{internals: &ZodIntegerInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodIntegerTyped[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodIntegerTyped[T, R]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.ZodTypeInternals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// extractIntegerValue extracts the base integer value T from constraint type R
func extractIntegerValue[T IntegerConstraint, R any](value R) T {
	// Handle direct assignment (when T == R)
	if directValue, ok := any(value).(T); ok {
		return directValue
	}

	// Handle pointer dereferencing
	switch v := any(value).(type) {
	case *int:
		if v != nil {
			return any(*v).(T)
		}
	case *int8:
		if v != nil {
			return any(*v).(T)
		}
	case *int16:
		if v != nil {
			return any(*v).(T)
		}
	case *int32:
		if v != nil {
			return any(*v).(T)
		}
	case *int64:
		if v != nil {
			return any(*v).(T)
		}
	case *uint:
		if v != nil {
			return any(*v).(T)
		}
	case *uint8:
		if v != nil {
			return any(*v).(T)
		}
	case *uint16:
		if v != nil {
			return any(*v).(T)
		}
	case *uint32:
		if v != nil {
			return any(*v).(T)
		}
	case *uint64:
		if v != nil {
			return any(*v).(T)
		}
	}

	// Fallback to zero value
	var zero T
	return zero
}

// newZodIntegerFromDef constructs a new ZodInteger from the given definition.
// Internal helper used by the constructor chain.
func newZodIntegerFromDef[T IntegerConstraint, R any](def *ZodIntegerDef) *ZodIntegerTyped[T, R] {
	internals := &ZodIntegerInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.ZodTypeDef.Type,
			Checks: def.ZodTypeDef.Checks,
			Coerce: def.ZodTypeDef.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		intDef := &ZodIntegerDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodIntegerFromDef[T, R](intDef)).(core.ZodType[any])
	}

	if def.ZodTypeDef.Error != nil {
		internals.ZodTypeInternals.Error = def.ZodTypeDef.Error
	}

	return &ZodIntegerTyped[T, R]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// IntegerTyped creates a generic integer schema with automatic type inference.
// It automatically determines the appropriate type code based on the generic type parameter.
// Usage: IntegerTyped[int](), IntegerTyped[uint32](), IntegerTyped[int64](), etc.
func IntegerTyped[T IntegerConstraint](params ...any) *ZodIntegerTyped[T, T] {
	// Determine type code based on T
	var typeCode core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case int, *int:
		typeCode = core.ZodTypeInt
	case int8, *int8:
		typeCode = core.ZodTypeInt8
	case int16, *int16:
		typeCode = core.ZodTypeInt16
	case int32, *int32:
		typeCode = core.ZodTypeInt32
	case int64, *int64:
		typeCode = core.ZodTypeInt64
	case uint, *uint:
		typeCode = core.ZodTypeUint
	case uint8, *uint8:
		typeCode = core.ZodTypeUint8
	case uint16, *uint16:
		typeCode = core.ZodTypeUint16
	case uint32, *uint32:
		typeCode = core.ZodTypeUint32
	case uint64, *uint64:
		typeCode = core.ZodTypeUint64
	default:
		typeCode = core.ZodTypeInt64 // fallback
	}

	return newIntegerTyped[T, T](typeCode, params...)
}

// Int creates a standard int schema
func Int(params ...any) *ZodIntegerTyped[int, int] {
	return newIntegerTyped[int, int](core.ZodTypeInt, params...)
}

// IntPtr creates a schema for *int.
func IntPtr(params ...any) *ZodIntegerTyped[int, *int] {
	return newIntegerTyped[int, *int](core.ZodTypeInt, params...)
}

// Int8 creates an int8 schema.
func Int8(params ...any) *ZodIntegerTyped[int8, int8] {
	return newIntegerTyped[int8, int8](core.ZodTypeInt8, params...)
}

// Int8Ptr creates a schema for *int8.
func Int8Ptr(params ...any) *ZodIntegerTyped[int8, *int8] {
	return newIntegerTyped[int8, *int8](core.ZodTypeInt8, params...)
}

// Int16 creates an int16 schema.
func Int16(params ...any) *ZodIntegerTyped[int16, int16] {
	return newIntegerTyped[int16, int16](core.ZodTypeInt16, params...)
}

// Int16Ptr creates a schema for *int16.
func Int16Ptr(params ...any) *ZodIntegerTyped[int16, *int16] {
	return newIntegerTyped[int16, *int16](core.ZodTypeInt16, params...)
}

// Int32 creates an int32 schema.
func Int32(params ...any) *ZodIntegerTyped[int32, int32] {
	return newIntegerTyped[int32, int32](core.ZodTypeInt32, params...)
}

// Int32Ptr creates a schema for *int32.
func Int32Ptr(params ...any) *ZodIntegerTyped[int32, *int32] {
	return newIntegerTyped[int32, *int32](core.ZodTypeInt32, params...)
}

// Int64 creates an int64 schema.
func Int64(params ...any) *ZodIntegerTyped[int64, int64] {
	return newIntegerTyped[int64, int64](core.ZodTypeInt64, params...)
}

// Int64Ptr creates a schema for *int64.
func Int64Ptr(params ...any) *ZodIntegerTyped[int64, *int64] {
	return newIntegerTyped[int64, *int64](core.ZodTypeInt64, params...)
}

// Uint creates a uint schema.
func Uint(params ...any) *ZodIntegerTyped[uint, uint] {
	return newIntegerTyped[uint, uint](core.ZodTypeUint, params...)
}

// UintPtr creates a schema for *uint.
func UintPtr(params ...any) *ZodIntegerTyped[uint, *uint] {
	return newIntegerTyped[uint, *uint](core.ZodTypeUint, params...)
}

// Uint8 creates a uint8 schema.
func Uint8(params ...any) *ZodIntegerTyped[uint8, uint8] {
	return newIntegerTyped[uint8, uint8](core.ZodTypeUint8, params...)
}

// Uint8Ptr creates a schema for *uint8.
func Uint8Ptr(params ...any) *ZodIntegerTyped[uint8, *uint8] {
	return newIntegerTyped[uint8, *uint8](core.ZodTypeUint8, params...)
}

// Uint16 creates a uint16 schema.
func Uint16(params ...any) *ZodIntegerTyped[uint16, uint16] {
	return newIntegerTyped[uint16, uint16](core.ZodTypeUint16, params...)
}

// Uint16Ptr creates a schema for *uint16.
func Uint16Ptr(params ...any) *ZodIntegerTyped[uint16, *uint16] {
	return newIntegerTyped[uint16, *uint16](core.ZodTypeUint16, params...)
}

// Uint32 creates a uint32 schema.
func Uint32(params ...any) *ZodIntegerTyped[uint32, uint32] {
	return newIntegerTyped[uint32, uint32](core.ZodTypeUint32, params...)
}

// Uint32Ptr creates a schema for *uint32.
func Uint32Ptr(params ...any) *ZodIntegerTyped[uint32, *uint32] {
	return newIntegerTyped[uint32, *uint32](core.ZodTypeUint32, params...)
}

// Uint64 creates a uint64 schema.
func Uint64(params ...any) *ZodIntegerTyped[uint64, uint64] {
	return newIntegerTyped[uint64, uint64](core.ZodTypeUint64, params...)
}

// Uint64Ptr creates a schema for *uint64.
func Uint64Ptr(params ...any) *ZodIntegerTyped[uint64, *uint64] {
	return newIntegerTyped[uint64, *uint64](core.ZodTypeUint64, params...)
}

// Byte creates a uint8 schema (alias for byte).
func Byte(params ...any) *ZodIntegerTyped[uint8, uint8] {
	return Uint8(params...)
}

// BytePtr creates a schema for *uint8 (alias for *byte).
func BytePtr(params ...any) *ZodIntegerTyped[uint8, *uint8] {
	return Uint8Ptr(params...)
}

// Rune creates an int32 schema (alias for rune).
func Rune(params ...any) *ZodIntegerTyped[int32, int32] {
	return Int32(params...)
}

// RunePtr creates a schema for *int32 (alias for *rune).
func RunePtr(params ...any) *ZodIntegerTyped[int32, *int32] {
	return Int32Ptr(params...)
}

// newIntegerTyped is the underlying generic function for creating integer schemas,
// allowing for explicit type parameterization. This is an internal function.
func newIntegerTyped[T IntegerConstraint, R any](typeCode core.ZodTypeCode, params ...any) *ZodIntegerTyped[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodIntegerDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeCode,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply the normalized parameters to the schema definition.
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodIntegerFromDef[T, R](def)
}

// CoercedInteger creates a int64 schema with coercion enabled
func CoercedInteger(params ...any) *ZodIntegerTyped[int64, int64] {
	schema := Int64(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedIntegerPtr creates a *int64 schema with coercion enabled
func CoercedIntegerPtr(params ...any) *ZodIntegerTyped[int64, *int64] {
	schema := Int64Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt creates an int schema with coercion enabled
func CoercedInt(params ...any) *ZodIntegerTyped[int, int] {
	schema := Int(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedIntPtr creates a *int schema with coercion enabled
func CoercedIntPtr(params ...any) *ZodIntegerTyped[int, *int] {
	schema := IntPtr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt8 creates an int8 schema with coercion enabled
func CoercedInt8(params ...any) *ZodIntegerTyped[int8, int8] {
	schema := Int8(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt8Ptr creates a *int8 schema with coercion enabled
func CoercedInt8Ptr(params ...any) *ZodIntegerTyped[int8, *int8] {
	schema := Int8Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt16 creates an int16 schema with coercion enabled
func CoercedInt16(params ...any) *ZodIntegerTyped[int16, int16] {
	schema := Int16(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt16Ptr creates a *int16 schema with coercion enabled
func CoercedInt16Ptr(params ...any) *ZodIntegerTyped[int16, *int16] {
	schema := Int16Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt32 creates an int32 schema with coercion enabled
func CoercedInt32(params ...any) *ZodIntegerTyped[int32, int32] {
	schema := Int32(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt32Ptr creates a *int32 schema with coercion enabled
func CoercedInt32Ptr(params ...any) *ZodIntegerTyped[int32, *int32] {
	schema := Int32Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt64 creates an int64 schema with coercion enabled
func CoercedInt64(params ...any) *ZodIntegerTyped[int64, int64] {
	schema := Int64(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedInt64Ptr creates a *int64 schema with coercion enabled
func CoercedInt64Ptr(params ...any) *ZodIntegerTyped[int64, *int64] {
	schema := Int64Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint creates a uint schema with coercion enabled
func CoercedUint(params ...any) *ZodIntegerTyped[uint, uint] {
	schema := Uint(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUintPtr creates a *uint schema with coercion enabled
func CoercedUintPtr(params ...any) *ZodIntegerTyped[uint, *uint] {
	schema := UintPtr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint8 creates a uint8 schema with coercion enabled
func CoercedUint8(params ...any) *ZodIntegerTyped[uint8, uint8] {
	schema := Uint8(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint8Ptr creates a *uint8 schema with coercion enabled
func CoercedUint8Ptr(params ...any) *ZodIntegerTyped[uint8, *uint8] {
	schema := Uint8Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint16 creates a uint16 schema with coercion enabled
func CoercedUint16(params ...any) *ZodIntegerTyped[uint16, uint16] {
	schema := Uint16(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint16Ptr creates a *uint16 schema with coercion enabled
func CoercedUint16Ptr(params ...any) *ZodIntegerTyped[uint16, *uint16] {
	schema := Uint16Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint32 creates a uint32 schema with coercion enabled
func CoercedUint32(params ...any) *ZodIntegerTyped[uint32, uint32] {
	schema := Uint32(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint32Ptr creates a *uint32 schema with coercion enabled
func CoercedUint32Ptr(params ...any) *ZodIntegerTyped[uint32, *uint32] {
	schema := Uint32Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint64 creates a uint64 schema with coercion enabled
func CoercedUint64(params ...any) *ZodIntegerTyped[uint64, uint64] {
	schema := Uint64(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedUint64Ptr creates a *uint64 schema with coercion enabled
func CoercedUint64Ptr(params ...any) *ZodIntegerTyped[uint64, *uint64] {
	schema := Uint64Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// =============================================================================
// NON-GENERIC INTEGER FUNCTIONS
// =============================================================================

// Integer creates a flexible integer schema that accepts any integer type
// Now returns ZodInteger[int64, int64] which is equivalent to ZodIntegerTyped[int64, int64]
func Integer(params ...any) *ZodInteger[int64, int64] {
	return Int64(params...)
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToIntegerConstraintType converts a base type T to constraint type R
func convertToIntegerConstraintType[T IntegerConstraint, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *int:
		if intVal, ok := any(value).(int); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *int8:
		if intVal, ok := any(value).(int8); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *int16:
		if intVal, ok := any(value).(int16); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *int32:
		if intVal, ok := any(value).(int32); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *int64:
		if intVal, ok := any(value).(int64); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *uint:
		if intVal, ok := any(value).(uint); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *uint8:
		if intVal, ok := any(value).(uint8); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *uint16:
		if intVal, ok := any(value).(uint16); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *uint32:
		if intVal, ok := any(value).(uint32); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	case *uint64:
		if intVal, ok := any(value).(uint64); ok {
			intCopy := intVal
			return any(&intCopy).(R)
		}
	default:
		// For value types, return T directly as R
		return any(value).(R)
	}
	return zero
}

// Check adds a custom validation function for integer schemas that can push multiple issues.
func (z *ZodIntegerTyped[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodIntegerTyped[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// Attempt direct assertion first
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch: if R is pointer but payload holds value
		var zero R
		zeroTyp := reflect.TypeOf(zero)
		if zeroTyp != nil && zeroTyp.Kind() == reflect.Ptr {
			elemTyp := zeroTyp.Elem()
			valRV := reflect.ValueOf(payload.GetValue())
			if valRV.IsValid() && valRV.Type() == elemTyp {
				ptr := reflect.New(elemTyp)
				ptr.Elem().Set(valRV)
				if casted, ok := ptr.Interface().(R); ok {
					fn(casted, payload)
				}
			}
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	in := z.internals.ZodTypeInternals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}
