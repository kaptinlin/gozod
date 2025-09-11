package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// NilConstraint restricts values to any or *any
type NilConstraint interface {
	any
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodNilDef defines the configuration for nil validation
type ZodNilDef struct {
	core.ZodTypeDef
}

// ZodNilInternals contains nil validator internal state
type ZodNilInternals struct {
	core.ZodTypeInternals
	Def *ZodNilDef // Schema definition
}

// ZodNil represents a nil validation schema with type safety
type ZodNil[T any] struct {
	internals *ZodNilInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodNil[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodNil[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodNil[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// validateNilValue is the validator function for Nil type
// Nil type should only accept nil values
func validateNilValue[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// For Nil type, accept any nil value (including the original nil input)
	if any(value) == nil || reflectx.IsNil(value) {
		// Apply checks to the nil value
		return engine.ApplyChecks[T](value, checks, ctx)
	}

	// Reject non-nil values for normal parsing
	var zero T
	return zero, issues.CreateInvalidTypeError(core.ZodTypeNil, value, ctx)
}

// convertNilTypeToConstraint handles type conversion for Nil type, preserving nil values
// Special handling for Prefault values that may be non-nil
func convertNilTypeToConstraint[T any](
	result any,
	ctx *core.ParseContext,
	expectedType core.ZodTypeCode,
) (T, error) {
	// For nil values, convert directly
	if result == nil || reflectx.IsNil(result) {
		var zero T
		return zero, nil
	}

	// For non-nil values (e.g., from Prefault), allow conversion
	// This enables Prefault values to reach the Refine stage
	converted, ok := convertAnyToConstraintType[T](result)
	if !ok {
		// If conversion fails, return an error
		var zero T
		return zero, issues.CreateInvalidTypeError(expectedType, result, ctx)
	}
	return converted, nil
}

// validateNilValueWithCustomError is the validator function for Nil type with custom error support
// Nil type should only accept nil values
func validateNilValueWithCustomError[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext, internals *core.ZodTypeInternals) (T, error) {
	// For Nil type, accept any nil value (including the original nil input)
	if any(value) == nil || reflectx.IsNil(value) {
		// Apply checks to the nil value
		return engine.ApplyChecks[T](value, checks, ctx)
	}

	// Reject non-nil values for normal parsing
	// Use the new function that supports custom error messages
	var zero T
	return zero, issues.CreateInvalidTypeErrorWithInst(core.ZodTypeNil, value, ctx, internals)
}

// Parse returns a value that matches the generic type T with full type safety
// Nil type uses the standard ParsePrimitive with custom validator and converter
func (z *ZodNil[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	// Use the internally recorded type code by default, fall back to nil if not set
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeNil
	}

	internals := &z.internals.ZodTypeInternals

	// Create validator function that captures internals
	validator := func(value any, checks []core.ZodCheck, ctx *core.ParseContext) (any, error) {
		return validateNilValueWithCustomError[any](value, checks, ctx, internals)
	}

	// Use the standard ParsePrimitive with nil-specific validator and converter
	return engine.ParsePrimitive[any, T](
		input,
		internals,
		expectedType,
		validator,
		convertNilTypeToConstraint[T],
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodNil[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodNil[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// MustParseAny is the any-type variant that panics on error.
func (z *ZodNil[T]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodNil[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Use the internally recorded type code by default, fall back to nil if not set
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeNil
	}

	return engine.ParsePrimitiveStrict[any, T](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		validateNilValue[any],
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on validation failure.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodNil[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *any because the optional value may be nil or missing.
func (z *ZodNil[T]) Optional() *ZodNil[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *any because the value may be nil.
func (z *ZodNil[T]) Nilable() *ZodNil[*any] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodNil[T]) Nullish() *ZodNil[*any] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default keeps the current generic type T.
func (z *ZodNil[T]) Default(v any) *ZodNil[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic type T.
func (z *ZodNil[T]) DefaultFunc(fn func() any) *ZodNil[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current generic type T.
func (z *ZodNil[T]) Prefault(v any) *ZodNil[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type T.
func (z *ZodNil[T]) PrefaultFunc(fn func() any) *ZodNil[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(fn)
	return z.withInternals(in)
}

// Meta stores metadata for this nil schema.
func (z *ZodNil[T]) Meta(meta core.GlobalMeta) *ZodNil[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodNil[T]) Transform(fn func(any, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		nilValue := extractNil(input)
		return fn(nilValue, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodNil[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		nilValue := extractNil(input)
		return target.Parse(nilValue, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// TYPE CONVERSION - NO LONGER NEEDED (USING WRAPFN PATTERN)
// =============================================================================

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function that matches the schema's output
// type T. The callback will be executed even when the value is nil (for *any
// schemas) to align with Zod v4 semantics.
func (z *ZodNil[T]) Refine(fn func(T) bool, params ...any) *ZodNil[T] {
	// Wrapper converts the raw value (always nil) into T before calling fn.
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case *any:
			// Schema output is *any – convert incoming value (nil) to *any
			if v == nil {
				nilPtr := (*any)(nil)
				return fn(any(nilPtr).(T))
			}
			ptr := &v
			return fn(any(ptr).(T))
		case any:
			// Schema output is any
			if v == nil {
				return fn(any(nil).(T))
			}
			return fn(any(v).(T)) //nolint:unconvert // Required for generic type constraint conversion
		default:
			// Unsupported type – should never happen
			return false
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodNil[T]) RefineAny(fn func(any) bool, params ...any) *ZodNil[T] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodNil instance of type *any.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer type.
func (z *ZodNil[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodNil[*any] {
	return &ZodNil[*any]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new ZodNil instance that keeps the original generic type T.
// Used by modifiers that retain the original type, such as Default, Prefault, and Transform.
func (z *ZodNil[T]) withInternals(in *core.ZodTypeInternals) *ZodNil[T] {
	return &ZodNil[T]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodNil[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodNil[T]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.Checks = originalChecks
	}
}

// extractNil extracts nil value from generic type T
func extractNil[T any](value T) any {
	if ptr, ok := any(value).(*any); ok {
		if ptr != nil {
			return *ptr
		}
		return nil
	}
	return any(value)
}

// convertAnyToConstraintType converts any value to constraint type T if possible
func convertAnyToConstraintType[T any](value any) (T, bool) {
	var zero T

	// Handle nil input first
	if value == nil {
		return zero, true // zero value for any constraint type
	}

	// Type conversion logic based on constraint type
	switch any(zero).(type) {
	case *any:
		// For *any constraint, try direct conversion first (in case value is already *any)
		if converted, ok := any(value).(T); ok { //nolint:unconvert // Required for generic type constraint conversion
			return converted, true
		}
		// For *any constraint, wrap the value in a pointer
		return any(&value).(T), true
	default:
		// For any constraint, try direct conversion
		if converted, ok := any(value).(T); ok { //nolint:unconvert // Required for generic type constraint conversion
			return converted, true
		}
		return zero, false
	}
}

// newZodNilFromDef constructs a new ZodNil from the given definition.
// Internal helper used by the constructor chain.
func newZodNilFromDef[T any](def *ZodNilDef) *ZodNil[T] {
	internals := &ZodNilInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Nil type should accept nil values by default
	internals.Nilable = true

	// Provide a constructor so that AddCheck can create new schema instances.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		nilDef := &ZodNilDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodNilFromDef[T](nilDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodNil[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Nil creates a nil schema following Zod TypeScript v4 pattern
// Usage:
//
//	Nil()                    // no parameters
//	Nil("custom error")      // string shorthand
//	Nil(SchemaParams{...})   // full parameters
func Nil(params ...any) *ZodNil[any] {
	return NilTyped[any](params...)
}

// NilPtr creates a schema for *any
func NilPtr(params ...any) *ZodNil[*any] {
	return NilTyped[*any](params...)
}

// NilTyped is the underlying generic function for creating nil schemas
// allowing for explicit type parameterization
func NilTyped[T any](params ...any) *ZodNil[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodNilDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNil,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply the normalized parameters to the schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodNilFromDef[T](def)
}
