package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodNil[T]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse returns a value that matches the generic type T with full type safety.
func (z *ZodNil[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	parseCtx := (*core.ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}
	if parseCtx == nil {
		parseCtx = core.NewParseContext()
	}

	// For nil type, always try core parsing first
	// Optional/Nilable modifiers don't change the core behavior for nil type
	return z.parseNilCore(input, parseCtx)
}

// parseNilCore handles the core nil parsing logic
func (z *ZodNil[T]) parseNilCore(input any, ctx *core.ParseContext) (T, error) {
	// Only accept nil input
	if input != nil {
		return z.tryPrefaultFallback(input, ctx)
	}

	// Validate nil with checks and capture transformed result
	if len(z.internals.ZodTypeInternals.Checks) > 0 {
		transformedInput, err := engine.ApplyChecks[any](input, z.internals.ZodTypeInternals.Checks, ctx)
		if err != nil {
			return z.tryPrefaultFallback(input, ctx)
		}
		input = transformedInput
	}

	// Convert nil to the constraint type T
	return convertNilToConstraintType[T](input)
}

// tryPrefaultFallback attempts to use prefault values when validation fails
func (z *ZodNil[T]) tryPrefaultFallback(originalInput any, ctx *core.ParseContext) (T, error) {
	internals := &z.internals.ZodTypeInternals

	// Try prefault value first
	if internals.PrefaultValue != nil {
		if result, ok := convertAnyToConstraintType[T](internals.PrefaultValue); ok {
			return result, nil
		}
	}

	// Try prefault function
	if internals.PrefaultFunc != nil {
		prefaultValue := internals.PrefaultFunc()
		if result, ok := convertAnyToConstraintType[T](prefaultValue); ok {
			return result, nil
		}
	}

	// No prefault available, return error with schema instance for custom error message
	var zero T
	rawIssue := issues.CreateInvalidTypeIssue(core.ZodTypeNil, originalInput)
	rawIssue.Inst = z // Pass schema instance so FinalizeIssue can extract custom error message
	finalIssue := issues.FinalizeIssue(rawIssue, ctx, nil)
	return zero, issues.NewZodError([]core.ZodIssue{finalIssue})
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

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *any because the optional value may be nil or missing.
func (z *ZodNil[T]) Optional() *ZodNil[*any] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *any because the value may be nil.
func (z *ZodNil[T]) Nilable() *ZodNil[*any] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodNil[T]) Nullish() *ZodNil[*any] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default keeps the current generic type T.
func (z *ZodNil[T]) Default(v any) *ZodNil[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic type T.
func (z *ZodNil[T]) DefaultFunc(fn func() any) *ZodNil[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault keeps the current generic type T.
func (z *ZodNil[T]) Prefault(v any) *ZodNil[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps the current generic type T.
func (z *ZodNil[T]) PrefaultFunc(fn func() any) *ZodNil[T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
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
	return core.NewZodPipe[T, any](z, wrapperFn)
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
			if anyVal, ok := v.(any); ok {
				ptr := &anyVal
				return fn(any(ptr).(T))
			}
			return false
		case any:
			// Schema output is any
			if v == nil {
				return fn(any(nil).(T))
			}
			if anyVal, ok := v.(any); ok {
				return fn(any(anyVal).(T))
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
func (z *ZodNil[T]) RefineAny(fn func(any) bool, params ...any) *ZodNil[T] {
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
		originalChecks := z.internals.ZodTypeInternals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.ZodTypeInternals.Checks = originalChecks
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

// convertNilToConstraintType converts nil to the target constraint type T
func convertNilToConstraintType[T any](value any) (T, error) {
	var zero T
	// For nil type, always return zero value (nil) regardless of constraint type
	// This is the special semantic of nil type: it always represents "no value"
	return zero, nil
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
		if converted, ok := any(value).(T); ok {
			return converted, true
		}
		// For *any constraint, wrap the value in a pointer
		return any(&value).(T), true
	default:
		// For any constraint, try direct conversion
		if converted, ok := any(value).(T); ok {
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

// Null creates a nil schema (alias for Nil)
func Null(args ...any) *ZodNil[any] {
	return Nil(args...)
}

// NullPtr creates a nil schema with pointer constraint (alias for NilPtr)
func NullPtr(args ...any) *ZodNil[*any] {
	return NilPtr(args...)
}
