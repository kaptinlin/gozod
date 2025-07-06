package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// BoolConstraint restricts values to bool or *bool.
type BoolConstraint interface {
	bool | *bool
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodBoolDef defines the configuration for boolean validation
type ZodBoolDef struct {
	core.ZodTypeDef
}

// ZodBoolInternals contains boolean validator internal state
type ZodBoolInternals struct {
	core.ZodTypeInternals
	Def *ZodBoolDef // Schema definition
}

// ZodBool represents a boolean validation schema with type safety
type ZodBool[T BoolConstraint] struct {
	internals *ZodBoolInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodBool[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodBool[T]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodBool[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce implements Coercible interface for boolean type conversion
func (z *ZodBool[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToBool(input)
	return result, err == nil
}

// Parse returns a value that matches the generic type T with full type safety.
func (z *ZodBool[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeBool
	}
	return engine.ParsePrimitive[bool, T](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		engine.ApplyChecks[bool],
		engine.ConvertToConstraintType[bool, T],
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodBool[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodBool[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *bool because the optional value may be nil.
func (z *ZodBool[T]) Optional() *ZodBool[*bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *bool because the value may be nil.
func (z *ZodBool[T]) Nilable() *ZodBool[*bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodBool[T]) Nullish() *ZodBool[*bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default keeps the current generic type T.
func (z *ZodBool[T]) Default(v bool) *ZodBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic type T.
func (z *ZodBool[T]) DefaultFunc(fn func() bool) *ZodBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current generic type T.
func (z *ZodBool[T]) Prefault(v bool) *ZodBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type T.
func (z *ZodBool[T]) PrefaultFunc(fn func() bool) *ZodBool[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
func (z *ZodBool[T]) Transform(fn func(bool, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	// WrapFn Pattern: Create wrapper function for type-safe extraction
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		boolValue := extractBool(input) // Use existing extraction logic
		return fn(boolValue, ctx)
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodBool[T]) Overwrite(transform func(T) T, params ...any) *ZodBool[T] {
	// Create a transformation function that works with the exact type T
	transformAny := func(input any) any {
		// Try to convert input to type T
		converted, ok := convertToBoolType[T](input)
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
func (z *ZodBool[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	// WrapFn Pattern: Create target function for type conversion and validation
	targetFn := func(input T, ctx *core.ParseContext) (any, error) {
		// Extract bool value from constraint type T
		boolValue := extractBool(input)
		// Apply target schema to the extracted bool
		return target.Parse(boolValue, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[T, any](z, targetFn)
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToBoolType converts only bool values to the target bool type T with strict type checking
func convertToBoolType[T BoolConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		// Handle nil values for pointer types
		switch any(zero).(type) {
		case *bool:
			return zero, true // zero value for pointer types is nil
		default:
			return zero, false // nil not allowed for value types
		}
	}

	// Extract bool value from input
	var boolValue bool
	var isValid bool

	switch val := v.(type) {
	case bool:
		boolValue, isValid = val, true
	case *bool:
		if val != nil {
			boolValue, isValid = *val, true
		}
	default:
		return zero, false // Reject all non-bool types
	}

	if !isValid {
		return zero, false
	}

	// Convert to target type T
	switch any(zero).(type) {
	case bool:
		return any(boolValue).(T), true
	case *bool:
		return any(&boolValue).(T), true
	default:
		return zero, false
	}
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function that matches the schema's output
// type T.
func (z *ZodBool[T]) Refine(fn func(T) bool, params ...any) *ZodBool[T] {
	// Wrapper converts the raw value (always bool or nil) into T before calling fn.
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case bool:
			// Schema output is bool
			if v == nil {
				// nil should never reach here for bool schema; treat as failure.
				return false
			}
			if boolVal, ok := v.(bool); ok {
				return fn(any(boolVal).(T))
			}
			return false
		case *bool:
			// Schema output is *bool – convert incoming value (bool or nil) to *bool
			if v == nil {
				return fn(any((*bool)(nil)).(T))
			}
			if boolVal, ok := v.(bool); ok {
				bCopy := boolVal
				ptr := &bCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			// Unsupported type – should never happen
			return false
		}
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	// Convert back to the format expected by checks.NewCustom
	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)

	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodBool[T]) RefineAny(fn func(any) bool, params ...any) *ZodBool[T] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodBool instance of type *bool.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer type.
func (z *ZodBool[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodBool[*bool] {
	return &ZodBool[*bool]{internals: &ZodBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new ZodBool instance that keeps the original generic type T.
// Used by modifiers that retain the original type, such as Default, Prefault, and Transform.
func (z *ZodBool[T]) withInternals(in *core.ZodTypeInternals) *ZodBool[T] {
	return &ZodBool[T]{internals: &ZodBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodBool[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodBool[T]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.ZodTypeInternals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// extractBool extracts boolean value from generic type T
func extractBool[T BoolConstraint](value T) bool {
	if ptr, ok := any(value).(*bool); ok {
		return ptr != nil && *ptr
	}
	return any(value).(bool)
}

// newZodBoolFromDef constructs a new ZodBool from the given definition.
// Internal helper used by the constructor chain.
func newZodBoolFromDef[T BoolConstraint](def *ZodBoolDef) *ZodBool[T] {
	internals := &ZodBoolInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Provide a constructor so that AddCheck can create new schema instances.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		boolDef := &ZodBoolDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodBoolFromDef[T](boolDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodBool[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Bool creates a bool schema following Zod TypeScript v4 pattern
// Usage:
//
//	Bool()                    // no parameters
//	Bool("custom error")      // string shorthand
//	Bool(SchemaParams{...})   // full parameters
func Bool(params ...any) *ZodBool[bool] {
	return BoolTyped[bool](params...)
}

// BoolPtr creates a schema for *bool
func BoolPtr(params ...any) *ZodBool[*bool] {
	return BoolTyped[*bool](params...)
}

// BoolTyped is the underlying generic function for creating boolean schemas
// allowing for explicit type parameterization
func BoolTyped[T BoolConstraint](params ...any) *ZodBool[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodBoolDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeBool,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply the normalized parameters to the schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodBoolFromDef[T](def)
}

// CoercedBool creates a bool schema with coercion enabled
func CoercedBool(args ...any) *ZodBool[bool] {
	schema := Bool(args...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// CoercedBoolPtr creates a *bool schema with coercion enabled
func CoercedBoolPtr(args ...any) *ZodBool[*bool] {
	schema := BoolPtr(args...)
	schema.internals.ZodTypeInternals.SetCoerce(true)
	return schema
}

// Check adds a custom validation function that can push multiple issues via ParsePayload.
func (z *ZodBool[T]) Check(fn func(value T, payload *core.ParsePayload), params ...any) *ZodBool[T] {
	wrapper := func(payload *core.ParsePayload) {
		// 1. Try direct type assertion to T
		if val, ok := payload.GetValue().(T); ok {
			fn(val, payload)
			return
		}

		// 2. If T is *bool but payload value is bool, take address and call
		var zero T
		if _, ok := any(zero).(*bool); ok {
			if b, ok := payload.GetValue().(bool); ok {
				copy := b
				ptr := &copy
				fn(any(ptr).(T), payload)
			}
			// No special handling for other types
		}
	}
	check := checks.NewCustom[T](wrapper, utils.GetFirstParam(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonOptional removes the optional flag and forces return type to bool.
// It also sets internals.Type to ZodTypeNonOptional so that error reporting
// uses "nonoptional" rather than "bool" when nil is encountered.
func (z *ZodBool[T]) NonOptional() *ZodBool[bool] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodBool[bool]{
		internals: &ZodBoolInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}
