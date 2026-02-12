package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodNeverDef defines the configuration for never validation.
type ZodNeverDef struct {
	core.ZodTypeDef
}

// ZodNeverInternals contains never validator internal state.
type ZodNeverInternals struct {
	core.ZodTypeInternals
	Def *ZodNeverDef
}

// ZodNever represents a never validation schema that always rejects input.
type ZodNever[T any, R any] struct {
	internals *ZodNeverInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodNever[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts missing values.
func (z *ZodNever[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodNever[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// validateNeverValue rejects all values with an invalid type error.
// Checks are applied first to support custom refinement error messages.
func validateNeverValue[T any](value T, chks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	if len(chks) > 0 {
		validated, err := engine.ApplyChecks[T](value, chks, ctx)
		if err != nil {
			var zero T
			return zero, err
		}
		value = validated
	}
	var zero T
	return zero, issues.CreateInvalidTypeError(core.ZodTypeNever, value, ctx)
}

// Parse validates the input and always rejects non-nil values.
func (z *ZodNever[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeNever,
		validateNeverValue[T],
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// MustParse validates the input and panics on failure.
func (z *ZodNever[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input and returns any type.
func (z *ZodNever[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// MustParseAny validates the input and panics on failure.
func (z *ZodNever[T, R]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse requires exact type matching for compile-time type safety.
func (z *ZodNever[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeNever,
		validateNeverValue[T],
		ctx...,
	)
}

// MustStrictParse requires exact type matching and panics on failure.
func (z *ZodNever[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a pointer constraint that accepts missing values.
func (z *ZodNever[T, R]) Optional() *ZodNever[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a pointer constraint that accepts nil values.
func (z *ZodNever[T, R]) Nilable() *ZodNever[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a pointer constraint that accepts both nil and missing values.
func (z *ZodNever[T, R]) Nullish() *ZodNever[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value returned when input is nil.
func (z *ZodNever[T, R]) Default(value T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(value)
	return z.withInternals(in)
}

// DefaultFunc sets a function that provides the default value.
func (z *ZodNever[T, R]) DefaultFunc(fn func() T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full parsing pipeline.
func (z *ZodNever[T, R]) Prefault(value T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(value)
	return z.withInternals(in)
}

// PrefaultFunc sets a function that provides the fallback value.
func (z *ZodNever[T, R]) PrefaultFunc(fn func() T) *ZodNever[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this schema.
func (z *ZodNever[T, R]) Meta(meta core.GlobalMeta) *ZodNever[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodNever[T, R]) Describe(description string) *ZodNever[T, R] {
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

// Transform applies a transformation function to the parsed value.
func (z *ZodNever[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractNeverValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a validation pipeline with the given target schema.
func (z *ZodNever[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractNeverValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation function with constraint type R.
func (z *ZodNever[T, R]) Refine(fn func(R) bool, params ...any) *ZodNever[T, R] {
	wrapper := func(v any) bool {
		if converted, ok := convertToNeverConstraintValue[T, R](v); ok {
			return fn(converted)
		}
		return false
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds a custom validation function with any type.
func (z *ZodNever[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodNever[T, R] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodNever with pointer constraint type.
func (z *ZodNever[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodNever[T, *T] {
	return &ZodNever[T, *T]{internals: &ZodNeverInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new ZodNever preserving the current constraint type.
func (z *ZodNever[T, R]) withInternals(in *core.ZodTypeInternals) *ZodNever[T, R] {
	return &ZodNever[T, R]{internals: &ZodNeverInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodNever[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodNever[T, R]); ok {
		z.internals = src.internals
	}
}

// Unwrap returns the schema itself.
func (z *ZodNever[T, R]) Unwrap() *ZodNever[T, R] {
	return z
}

// extractNeverValue extracts base type T from constraint type R.
func extractNeverValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *any:
		if v != nil {
			return any(*v).(T) //nolint:unconvert
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToNeverConstraintValue converts any value to constraint type R.
func convertToNeverConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok { //nolint:unconvert
		return r, true
	}

	// Handle pointer constraint: wrap value in *any
	if _, ok := any(zero).(*any); ok {
		if value == nil {
			return any((*any)(nil)).(R), true
		}
		return any(&value).(R), true
	}

	return zero, false
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// newZodNeverFromDef constructs a new ZodNever from the given definition.
func newZodNeverFromDef[T any, R any](def *ZodNeverDef) *ZodNever[T, R] {
	internals := &ZodNeverInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		neverDef := &ZodNeverDef{ZodTypeDef: *newDef}
		return any(newZodNeverFromDef[T, R](neverDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodNever[T, R]{internals: internals}
}

// Never creates a never schema that always rejects input.
func Never(params ...any) *ZodNever[any, any] {
	return NeverTyped[any, any](params...)
}

// NeverPtr creates a never schema with pointer constraint.
func NeverPtr(params ...any) *ZodNever[any, *any] {
	return NeverTyped[any, *any](params...)
}

// NeverTyped creates a typed never schema with explicit generic constraints.
func NeverTyped[T any, R any](params ...any) *ZodNever[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodNeverDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNever,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodNeverFromDef[T, R](def)
}
