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

// FloatConstraint restricts values to supported float types or their pointers.
type FloatConstraint interface {
	~float32 | ~float64 | ~*float32 | ~*float64
}

// Float32Constraint restricts values to float32 or *float32.
type Float32Constraint interface {
	float32 | *float32
}

// Float64Constraint restricts values to float64 or *float64.
type Float64Constraint interface {
	float64 | *float64
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodFloatDef defines the configuration for float validation
type ZodFloatDef struct {
	core.ZodTypeDef
}

// ZodFloatInternals contains float validator internal state
type ZodFloatInternals struct {
	core.ZodTypeInternals
	Def *ZodFloatDef // Schema definition
}

// ZodFloatTyped represents a floating-point validation schema with dual generic parameters
// T = base type (float32, float64, etc.), R = constraint type (T or *T)
type ZodFloatTyped[T FloatConstraint, R any] struct {
	internals *ZodFloatInternals
}

// ZodFloat is now a generic type alias for ZodFloatTyped
// This provides a unified interface for all float types
type ZodFloat[T FloatConstraint, R any] = ZodFloatTyped[T, R]

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodFloatTyped[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodFloatTyped[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodFloatTyped[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce implements Coercible interface for float type conversion
func (z *ZodFloatTyped[T, R]) Coerce(input any) (any, bool) {
	var zero T
	switch any(zero).(type) {
	case float32, *float32:
		result, err := coerce.ToFloat[float32](input)
		return result, err == nil
	case float64, *float64:
		result, err := coerce.ToFloat[float64](input)
		return result, err == nil
	default:
		// Fallback to float64
		result, err := coerce.ToFloat[float64](input)
		return result, err == nil
	}
}

// Parse returns a value that matches the constraint type R with full type safety.
func (z *ZodFloatTyped[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero T

	// Determine the type code based on T
	var typeCode core.ZodTypeCode
	switch any(zero).(type) {
	case float32, *float32:
		typeCode = core.ZodTypeFloat32
	case float64, *float64:
		typeCode = core.ZodTypeFloat64
	default:
		// Fallback to float64
		typeCode = core.ZodTypeFloat64
	}

	// Determine if we have a pointer constraint type R to enable pointer identity preservation
	var zeroR R
	isPointerConstraint := false
	switch any(zeroR).(type) {
	case *float32, *float64:
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
			// Apply all float-specific checks
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
				return zeroR, fmt.Errorf("float value cannot be nil")
			}

			// First, try to convert result directly to constraint type R (preserves pointer identity)
			if directValue, ok := result.(R); ok {
				return directValue, nil
			}

			// Special handling for pointer constraint types to preserve pointer identity
			var zeroPtr R
			switch any(zeroPtr).(type) {
			case *float32:
				if ptr, ok := result.(*float32); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			case *float64:
				if ptr, ok := result.(*float64); ok {
					if convertedPtr, ok2 := any(ptr).(R); ok2 {
						return convertedPtr, nil
					}
				}
			}

			// Fallback: try to convert to base type T and then to constraint type R
			if value, ok := result.(T); ok {
				// Convert base type T to constraint type R (this will create new pointers for pointer types)
				return convertToFloatConstraintType[T, R](value), nil
			}

			// Handle pointer dereferencing if needed (for cases where input pointer doesn't match expected type)
			switch v := result.(type) {
			case *float32:
				if v != nil {
					if floatVal, ok := any(*v).(T); ok {
						return convertToFloatConstraintType[T, R](floatVal), nil
					}
				}
			case *float64:
				if v != nil {
					if floatVal, ok := any(*v).(T); ok {
						return convertToFloatConstraintType[T, R](floatVal), nil
					}
				}
			}

			return zeroR, fmt.Errorf("type conversion failed: expected %T but got %T", *new(T), result)
		},
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodFloatTyped[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodFloatTyped[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// withPtrInternals creates a new ZodFloatTyped instance with pointer constraint type *T.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer constraint.
func (z *ZodFloatTyped[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodFloatTyped[T, *T] {
	return &ZodFloatTyped[T, *T]{internals: &ZodFloatInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// Optional returns a schema that accepts the base type T or nil, with constraint type *T.
func (z *ZodFloatTyped[T, R]) Optional() *ZodFloatTyped[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts the base type T or nil, with constraint type *T.
func (z *ZodFloatTyped[T, R]) Nilable() *ZodFloatTyped[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility.
func (z *ZodFloatTyped[T, R]) Nullish() *ZodFloatTyped[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes optional flag and returns value constraint (T).
// Mirrors Optional() -> NonOptional() pattern, enabling strict non-nil validation.
func (z *ZodFloatTyped[T, R]) NonOptional() *ZodFloatTyped[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodFloatTyped[T, T]{
		internals: &ZodFloatInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default keeps the current constraint type R.
func (z *ZodFloatTyped[T, R]) Default(v float64) *ZodFloatTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current constraint type R.
func (z *ZodFloatTyped[T, R]) DefaultFunc(fn func() float64) *ZodFloatTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current constraint type R.
func (z *ZodFloatTyped[T, R]) Prefault(v float64) *ZodFloatTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current constraint type R.
func (z *ZodFloatTyped[T, R]) PrefaultFunc(fn func() float64) *ZodFloatTyped[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min adds minimum value validation (alias for Gte)
func (z *ZodFloatTyped[T, R]) Min(minimum float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.Gte(minimum, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max adds maximum value validation (alias for Lte)
func (z *ZodFloatTyped[T, R]) Max(maximum float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.Lte(maximum, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gt adds greater than validation (exclusive)
func (z *ZodFloatTyped[T, R]) Gt(value float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.Gt(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Gte adds greater than or equal validation (inclusive)
func (z *ZodFloatTyped[T, R]) Gte(value float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.Gte(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lt adds less than validation (exclusive)
func (z *ZodFloatTyped[T, R]) Lt(value float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.Lt(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Lte adds less than or equal validation (inclusive)
func (z *ZodFloatTyped[T, R]) Lte(value float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.Lte(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Positive adds positive number validation (> 0)
func (z *ZodFloatTyped[T, R]) Positive(params ...any) *ZodFloatTyped[T, R] {
	return z.Gt(0, params...)
}

// Negative adds negative number validation (< 0)
func (z *ZodFloatTyped[T, R]) Negative(params ...any) *ZodFloatTyped[T, R] {
	return z.Lt(0, params...)
}

// NonNegative adds non-negative number validation (>= 0, alias for nonnegative)
func (z *ZodFloatTyped[T, R]) NonNegative(params ...any) *ZodFloatTyped[T, R] {
	return z.Gte(0, params...)
}

// NonPositive adds non-positive number validation (<= 0, alias for nonpositive)
func (z *ZodFloatTyped[T, R]) NonPositive(params ...any) *ZodFloatTyped[T, R] {
	return z.Lte(0, params...)
}

// MultipleOf adds multiple of validation
func (z *ZodFloatTyped[T, R]) MultipleOf(value float64, params ...any) *ZodFloatTyped[T, R] {
	check := checks.MultipleOf(value, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Step adds step validation (alias for MultipleOf to match Zod)
func (z *ZodFloatTyped[T, R]) Step(step float64, params ...any) *ZodFloatTyped[T, R] {
	return z.MultipleOf(step, params...)
}

// Int adds integer validation (no decimal part)
func (z *ZodFloatTyped[T, R]) Int(params ...any) *ZodFloatTyped[T, R] {
	check := checks.NewCustom[any](func(v any) bool {
		switch val := v.(type) {
		case float32:
			return val == float32(math.Trunc(float64(val)))
		case float64:
			return val == math.Trunc(val)
		case *float32:
			if val == nil {
				return true
			}
			return *val == float32(math.Trunc(float64(*val)))
		case *float64:
			if val == nil {
				return true
			}
			return *val == math.Trunc(*val)
		default:
			return false
		}
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Finite adds finite number validation (not NaN or Infinity)
func (z *ZodFloatTyped[T, R]) Finite(params ...any) *ZodFloatTyped[T, R] {
	check := checks.NewCustom[any](func(v any) bool {
		switch val := v.(type) {
		case float32:
			return !math.IsInf(float64(val), 0) && !math.IsNaN(float64(val))
		case float64:
			return !math.IsInf(val, 0) && !math.IsNaN(val)
		case *float32:
			if val == nil {
				return true
			}
			return !math.IsInf(float64(*val), 0) && !math.IsNaN(float64(*val))
		case *float64:
			if val == nil {
				return true
			}
			return !math.IsInf(*val, 0) && !math.IsNaN(*val)
		default:
			return false
		}
	}, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Safe adds safe number validation (within JavaScript safe integer range)
func (z *ZodFloatTyped[T, R]) Safe(params ...any) *ZodFloatTyped[T, R] {
	const maxSafeInt = 1<<53 - 1
	const minSafeInt = -(1<<53 - 1)
	return z.Gte(minSafeInt, params...).Lte(maxSafeInt, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
// Float types implement direct extraction of float64 values for transformation.
//
// WrapFn Implementation:
//  1. Create a wrapper function that extracts float64 from constraint type R
//  2. Apply the user's transformation function to the extracted float64
//  3. Return a ZodTransform with the wrapper function
//
// Zero Redundancy:
//   - No floatTypeConverter structure needed
//   - Direct function composition for maximum performance
//   - Type-safe extraction from constraint type R to float64
//
// Example:
//
//	schema := Float64().Min(0.0).Transform(func(f float64, ctx *RefinementContext) (string, error) {
//	    return fmt.Sprintf("%.2f", f), nil
//	})
func (z *ZodFloatTyped[T, R]) Transform(fn func(float64, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	// WrapFn Pattern: Create wrapper function for type-safe extraction
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		float64Value := extractFloatToFloat64[T, R](input) // Use existing extraction logic
		return fn(float64Value, ctx)
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodFloatTyped[T, R]) Overwrite(transform func(T) T, params ...any) *ZodFloatTyped[T, R] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToFloatType[T](input)
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
func (z *ZodFloatTyped[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	// WrapFn Pattern: Create target function for type conversion and validation
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		// Extract float64 value from constraint type R
		float64Value := extractFloatToFloat64[T, R](input)
		// Apply target schema to the extracted float64
		return target.Parse(float64Value, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[R, any](z, targetFn)
}

// =============================================================================
// TYPE CONVERSION
// =============================================================================

// extractFloatToFloat64 converts constraint type R to float64 for WrapFn pattern transformations
func extractFloatToFloat64[T FloatConstraint, R any](value R) float64 {
	// First extract the base type T from constraint type R
	baseValue := extractFloatValue[T, R](value)

	// Then convert T to float64
	switch v := any(baseValue).(type) {
	case float32:
		return float64(v)
	case float64:
		return v
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
func (z *ZodFloatTyped[T, R]) Refine(fn func(T) bool, params ...any) *ZodFloatTyped[T, R] {
	wrapper := func(v any) bool {
		converted, ok := convertToFloatType[T](v)
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

// convertToFloatType converts only matching float values to the target float type T with strict type checking
func convertToFloatType[T FloatConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		// Handle nil values for pointer types
		switch any(zero).(type) {
		case *float32, *float64:
			return zero, true // zero value for pointer types is nil
		default:
			return zero, false // nil not allowed for value types
		}
	}

	// Only accept matching float types - no cross-type conversion
	switch any(zero).(type) {
	case float32:
		if val, ok := v.(float32); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*float32); ok && val != nil {
			return any(*val).(T), true
		}
	case float64:
		if val, ok := v.(float64); ok {
			return any(val).(T), true
		}
		if val, ok := v.(*float64); ok && val != nil {
			return any(*val).(T), true
		}
	// Pointer types
	case *float32:
		if val, ok := v.(float32); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*float32); ok {
			return any(val).(T), true
		}
	case *float64:
		if val, ok := v.(float64); ok {
			ptr := &val
			return any(ptr).(T), true
		}
		if val, ok := v.(*float64); ok {
			return any(val).(T), true
		}
	}

	return zero, false // Reject all non-matching float types
}

// RefineAny adds flexible custom validation logic
func (z *ZodFloatTyped[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodFloatTyped[T, R] {
	check := checks.NewCustom[any](fn, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withInternals creates a new ZodFloat instance that keeps the original constraint type R.
// Used by modifiers that retain the original type, such as Default, Prefault, and validation methods.
func (z *ZodFloatTyped[T, R]) withInternals(in *core.ZodTypeInternals) *ZodFloatTyped[T, R] {
	return &ZodFloatTyped[T, R]{internals: &ZodFloatInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodFloatTyped[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodFloatTyped[T, R]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.ZodTypeInternals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// extractFloatValue extracts the base float value T from constraint type R
func extractFloatValue[T FloatConstraint, R any](value R) T {
	// Handle direct assignment (when T == R)
	if directValue, ok := any(value).(T); ok {
		return directValue
	}

	// Handle pointer dereferencing
	switch v := any(value).(type) {
	case *float32:
		if v != nil {
			return any(*v).(T)
		}
	case *float64:
		if v != nil {
			return any(*v).(T)
		}
	}

	// Fallback to zero value
	var zero T
	return zero
}

// newZodFloatFromDef constructs a new ZodFloat from the given definition.
// Internal helper used by the constructor chain.
func newZodFloatFromDef[T FloatConstraint, R any](def *ZodFloatDef) *ZodFloatTyped[T, R] {
	internals := &ZodFloatInternals{
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
		floatDef := &ZodFloatDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodFloatFromDef[T, R](floatDef)).(core.ZodType[any])
	}

	if def.ZodTypeDef.Error != nil {
		internals.ZodTypeInternals.Error = def.ZodTypeDef.Error
	}

	return &ZodFloatTyped[T, R]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// FloatTyped creates a generic float schema with automatic type inference.
// It automatically determines the appropriate type code based on the generic type parameter.
// Usage: FloatTyped[float32](), FloatTyped[float64](), etc.
func FloatTyped[T FloatConstraint](params ...any) *ZodFloatTyped[T, T] {
	// Determine type code based on T
	var typeCode core.ZodTypeCode
	var zero T
	switch any(zero).(type) {
	case float32, *float32:
		typeCode = core.ZodTypeFloat32
	case float64, *float64:
		typeCode = core.ZodTypeFloat64
	default:
		typeCode = core.ZodTypeFloat64 // fallback
	}

	return newFloatTyped[T, T](typeCode, params...)
}

// Float32 creates a float32 schema
func Float32(params ...any) *ZodFloatTyped[float32, float32] {
	return newFloatTyped[float32, float32](core.ZodTypeFloat32, params...)
}

// Float32Ptr creates a schema for *float32.
func Float32Ptr(params ...any) *ZodFloatTyped[float32, *float32] {
	return newFloatTyped[float32, *float32](core.ZodTypeFloat32, params...)
}

// Float64 creates a float64 schema
func Float64(params ...any) *ZodFloatTyped[float64, float64] {
	return newFloatTyped[float64, float64](core.ZodTypeFloat64, params...)
}

// Float64Ptr creates a schema for *float64.
func Float64Ptr(params ...any) *ZodFloatTyped[float64, *float64] {
	return newFloatTyped[float64, *float64](core.ZodTypeFloat64, params...)
}

// Float creates a flexible float64 schema (alias for Float64).
func Float(params ...any) *ZodFloatTyped[float64, float64] {
	return Float64(params...)
}

// FloatPtr creates a schema for *float64 (alias for Float64Ptr).
func FloatPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return Float64Ptr(params...)
}

// Number creates a number schema (alias for Float64).
func Number(params ...any) *ZodFloatTyped[float64, float64] {
	return Float64(params...)
}

// NumberPtr creates a schema for *float64 (alias for Float64Ptr).
func NumberPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return Float64Ptr(params...)
}

// newFloatTyped is the underlying generic function for creating float schemas,
// allowing for explicit type parameterization. This is an internal function.
func newFloatTyped[T FloatConstraint, R any](typeCode core.ZodTypeCode, params ...any) *ZodFloatTyped[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodFloatDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   typeCode,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply the normalized parameters to the schema definition.
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodFloatFromDef[T, R](def)
}

// CoercedFloat32 creates a float32 schema with coercion enabled
func CoercedFloat32(params ...any) *ZodFloatTyped[float32, float32] {
	schema := Float32(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedFloat32Ptr creates a *float32 schema with coercion enabled
func CoercedFloat32Ptr(params ...any) *ZodFloatTyped[float32, *float32] {
	schema := Float32Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedFloat64 creates a float64 schema with coercion enabled
func CoercedFloat64(params ...any) *ZodFloatTyped[float64, float64] {
	schema := Float64(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedFloat64Ptr creates a *float64 schema with coercion enabled
func CoercedFloat64Ptr(params ...any) *ZodFloatTyped[float64, *float64] {
	schema := Float64Ptr(params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedFloat creates a flexible float schema with coercion enabled
func CoercedFloat[T FloatConstraint](params ...any) *ZodFloatTyped[T, T] {
	schema := FloatTyped[T](params...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedFloatPtr creates a *float64 schema with coercion enabled
func CoercedFloatPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return CoercedFloat64Ptr(params...)
}

// CoercedNumber creates a number schema with coercion enabled (alias for CoercedFloat64)
func CoercedNumber(params ...any) *ZodFloatTyped[float64, float64] {
	return CoercedFloat64(params...)
}

// CoercedNumberPtr creates a *number schema with coercion enabled (alias for CoercedFloat64Ptr)
func CoercedNumberPtr(params ...any) *ZodFloatTyped[float64, *float64] {
	return CoercedFloat64Ptr(params...)
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToFloatConstraintType converts a base type T to constraint type R
func convertToFloatConstraintType[T FloatConstraint, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *float32:
		if floatVal, ok := any(value).(float32); ok {
			floatCopy := floatVal
			return any(&floatCopy).(R)
		}
	case *float64:
		if floatVal, ok := any(value).(float64); ok {
			floatCopy := floatVal
			return any(&floatCopy).(R)
		}
	default:
		// For value types, return T directly as R
		return any(value).(R)
	}
	return zero
}

// Check adds a custom validation function for ZodFloatTyped that can report multiple issues.
func (z *ZodFloatTyped[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodFloatTyped[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// direct assertion
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// pointer/value mismatch handling
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
	check := checks.NewCustom[R](wrapper, utils.GetFirstParam(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}
