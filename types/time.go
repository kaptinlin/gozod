package types

import (
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// TimeConstraint restricts values to time.Time or *time.Time.
type TimeConstraint interface {
	time.Time | *time.Time
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodTimeDef defines the configuration for time validation
type ZodTimeDef struct {
	core.ZodTypeDef
}

// ZodTimeInternals contains time validator internal state
type ZodTimeInternals struct {
	core.ZodTypeInternals
	Def *ZodTimeDef // Schema definition
}

// ZodTime represents a time validation schema with type safety
type ZodTime[T TimeConstraint] struct {
	internals *ZodTimeInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodTime[T]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodTime[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodTime[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce implements Coercible interface for time type conversion
func (z *ZodTime[T]) Coerce(input any) (any, bool) {
	result, err := coerce.ToTime(input)
	return result, err == nil
}

// Parse returns a value that matches the generic type T with full type safety.
func (z *ZodTime[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[time.Time, T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeTime,
		engine.ApplyChecks[time.Time],
		engine.ConvertToConstraintType[time.Time, T],
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error.
func (z *ZodTime[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodTime[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	// Use the internally recorded type code by default, fall back to time if not set
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeTime
	}

	return engine.ParsePrimitiveStrict[time.Time, T](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		engine.ApplyChecks[time.Time],
		ctx...,
	)
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodTime[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodTime[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *time.Time because the optional value may be nil.
func (z *ZodTime[T]) Optional() *ZodTime[*time.Time] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *time.Time because the value may be nil.
func (z *ZodTime[T]) Nilable() *ZodTime[*time.Time] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodTime[T]) Nullish() *ZodTime[*time.Time] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default keeps the current generic type T.
func (z *ZodTime[T]) Default(v time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic type T.
func (z *ZodTime[T]) DefaultFunc(fn func() time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current generic type T.
func (z *ZodTime[T]) Prefault(v time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type T.
func (z *ZodTime[T]) PrefaultFunc(fn func() time.Time) *ZodTime[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this time schema.
func (z *ZodTime[T]) Meta(meta core.GlobalMeta) *ZodTime[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodTime[T]) Describe(description string) *ZodTime[T] {
	newInternals := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(newInternals)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodTime[T]) Transform(fn func(time.Time, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapperFn := func(input T, ctx *core.RefinementContext) (any, error) {
		timeValue := extractTime(input)
		return fn(timeValue, ctx)
	}
	return core.NewZodTransform[T, any](z, wrapperFn)
}

// Overwrite transforms time value while keeping the same type
func (z *ZodTime[T]) Overwrite(transform func(T) T, params ...any) *ZodTime[T] {
	check := checks.NewZodCheckOverwrite(func(input any) any {
		// Simple type assertion - let the validation handle incorrect types
		if val, ok := input.(T); ok {
			return transform(val)
		}
		return input
	}, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodTime[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	wrapperFn := func(input T, ctx *core.ParseContext) (any, error) {
		timeValue := extractTime(input)
		return target.Parse(timeValue, ctx)
	}
	return core.NewZodPipe[T, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function that matches the schema's output
// type T. The callback will be executed even when the value is nil (for *time.Time
// schemas) to align with Zod v4 semantics.
func (z *ZodTime[T]) Refine(fn func(T) bool, params ...any) *ZodTime[T] {
	// Wrapper converts the raw value (always time.Time or nil) into T before calling fn.
	wrapper := func(v any) bool {
		var zero T

		switch any(zero).(type) {
		case time.Time:
			// Schema output is time.Time
			if v == nil {
				// nil should never reach here for time.Time schema; treat as failure.
				return false
			}
			if timeVal, ok := v.(time.Time); ok {
				return fn(any(timeVal).(T))
			}
			return false
		case *time.Time:
			// Schema output is *time.Time – convert incoming value (time.Time or nil) to *time.Time
			if v == nil {
				return fn(any((*time.Time)(nil)).(T))
			}
			if timeVal, ok := v.(time.Time); ok {
				tCopy := timeVal
				ptr := &tCopy
				return fn(any(ptr).(T))
			}
			return false
		default:
			// Unsupported type – should never happen
			return false
		}
	}

	// Use unified parameter handling with CustomParams
	param := utils.FirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)

	check := checks.NewCustom[any](wrapper, customParams)

	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds flexible custom validation logic
func (z *ZodTime[T]) RefineAny(fn func(any) bool, params ...any) *ZodTime[T] {
	// Use unified parameter handling with CustomParams
	param := utils.FirstParam(params...)
	customParams := utils.NormalizeCustomParams(param)

	check := checks.NewCustom[any](fn, customParams)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodTime instance of type *time.Time.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer type.
func (z *ZodTime[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodTime[*time.Time] {
	return &ZodTime[*time.Time]{internals: &ZodTimeInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new ZodTime instance that keeps the original generic type T.
// Used by modifiers that retain the original type, such as Default, Prefault, and Transform.
func (z *ZodTime[T]) withInternals(in *core.ZodTypeInternals) *ZodTime[T] {
	return &ZodTime[T]{internals: &ZodTimeInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodTime[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodTime[T]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.Checks = originalChecks
	}
}

// extractTime extracts time.Time value from generic type T
func extractTime[T TimeConstraint](value T) time.Time {
	if ptr, ok := any(value).(*time.Time); ok {
		if ptr != nil {
			return *ptr
		}
		return time.Time{}
	}
	return any(value).(time.Time)
}

// newZodTimeFromDef constructs a new ZodTime from the given definition.
// Internal helper used by the constructor chain.
func newZodTimeFromDef[T TimeConstraint](def *ZodTimeDef) *ZodTime[T] {
	internals := &ZodTimeInternals{
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
		timeDef := &ZodTimeDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodTimeFromDef[T](timeDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodTime[T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Time creates a time.Time schema following Zod TypeScript v4 pattern
// Usage:
//
//	Time()                    // no parameters
//	Time("custom error")      // string shorthand
//	Time(SchemaParams{...})   // full parameters
func Time(params ...any) *ZodTime[time.Time] {
	return TimeTyped[time.Time](params...)
}

// TimePtr creates a schema for *time.Time
func TimePtr(params ...any) *ZodTime[*time.Time] {
	return TimeTyped[*time.Time](params...)
}

// TimeTyped is the underlying generic function for creating time schemas
// allowing for explicit type parameterization
func TimeTyped[T TimeConstraint](params ...any) *ZodTime[T] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodTimeDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeTime,
			Checks: []core.ZodCheck{},
		},
	}

	// Apply the normalized parameters to the schema definition
	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodTimeFromDef[T](def)
}

// CoercedTime creates a time.Time schema with coercion enabled
func CoercedTime(args ...any) *ZodTime[time.Time] {
	schema := Time(args...)
	schema.internals.SetCoerce(true)
	return schema
}

// CoercedTimePtr creates a *time.Time schema with coercion enabled
func CoercedTimePtr(args ...any) *ZodTime[*time.Time] {
	schema := TimePtr(args...)
	schema.internals.SetCoerce(true)
	return schema
}
