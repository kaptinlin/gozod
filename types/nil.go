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
// TYPE DEFINITIONS
// =============================================================================

// ZodNilDef defines the configuration for nil validation.
type ZodNilDef struct {
	core.ZodTypeDef
}

// ZodNilInternals contains nil validator internal state.
type ZodNilInternals struct {
	core.ZodTypeInternals
	Def *ZodNilDef
}

// ZodNil represents a nil validation schema that only accepts nil values.
// T is the base type and R is the constraint type (may be a pointer for
// modifiers like Optional/Nilable).
type ZodNil[T any, R any] struct {
	internals *ZodNilInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodNil[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts missing values.
func (z *ZodNil[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodNil[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// validateNilValue validates that the input is nil.
// Non-nil values are rejected with an invalid type error.
func validateNilValue[T any](value T, chks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	if any(value) == nil || reflectx.IsNil(value) {
		return engine.ApplyChecks(value, chks, ctx)
	}
	var zero T
	return zero, issues.CreateInvalidTypeError(core.ZodTypeNil, value, ctx)
}

// newNilValidator creates a validator that supports custom error messages
// from schema parameters via the internals' Error field.
func newNilValidator[T any](internals *core.ZodTypeInternals) func(T, []core.ZodCheck, *core.ParseContext) (T, error) {
	return func(value T, chks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
		if any(value) == nil || reflectx.IsNil(value) {
			return engine.ApplyChecks(value, chks, ctx)
		}
		var zero T
		return zero, issues.CreateInvalidTypeErrorWithInst(core.ZodTypeNil, value, ctx, internals)
	}
}

// convertNilResult converts a parsed result to constraint type R.
// Nil type needs special handling because nil interface values don't
// match type switch cases in the generic ConvertToConstraintType.
func convertNilResult[T any, R any](
	result any, ctx *core.ParseContext, expectedType core.ZodTypeCode,
) (R, error) {
	if result == nil || reflectx.IsNil(result) {
		var zero R
		return zero, nil
	}
	return engine.ConvertToConstraintType[T, R](result, ctx, expectedType)
}

// Parse validates the input and returns the constraint type R.
func (z *ZodNil[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input, &z.internals.ZodTypeInternals, core.ZodTypeNil,
		newNilValidator[T](&z.internals.ZodTypeInternals), convertNilResult[T, R], ctx...,
	)
}

// MustParse validates the input and panics on failure.
func (z *ZodNil[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input and returns any type.
func (z *ZodNil[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// MustParseAny validates the input and panics on failure.
func (z *ZodNil[T, R]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse requires exact type matching for compile-time type safety.
func (z *ZodNil[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input, &z.internals.ZodTypeInternals, core.ZodTypeNil,
		validateNilValue[T], ctx...,
	)
}

// MustStrictParse requires exact type matching and panics on failure.
func (z *ZodNil[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
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
func (z *ZodNil[T, R]) Optional() *ZodNil[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a pointer constraint that accepts nil values.
func (z *ZodNil[T, R]) Nilable() *ZodNil[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a pointer constraint accepting both nil and missing values.
func (z *ZodNil[T, R]) Nullish() *ZodNil[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value returned when input is nil.
func (z *ZodNil[T, R]) Default(value T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(value)
	return z.withInternals(in)
}

// DefaultFunc sets a function that provides the default value.
func (z *ZodNil[T, R]) DefaultFunc(fn func() T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full parsing pipeline.
func (z *ZodNil[T, R]) Prefault(value T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(value)
	return z.withInternals(in)
}

// PrefaultFunc sets a function that provides the fallback value.
func (z *ZodNil[T, R]) PrefaultFunc(fn func() T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this schema.
func (z *ZodNil[T, R]) Meta(meta core.GlobalMeta) *ZodNil[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodNil[T, R]) Describe(description string) *ZodNil[T, R] {
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
func (z *ZodNil[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractNilValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a validation pipeline with the given target schema.
func (z *ZodNil[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractNilValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation function with constraint type R.
func (z *ZodNil[T, R]) Refine(fn func(R) bool, params ...any) *ZodNil[T, R] {
	wrapper := func(v any) bool {
		if converted, ok := convertToNilConstraintValue[T, R](v); ok {
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
func (z *ZodNil[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodNil[T, R] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodNil with pointer constraint type.
func (z *ZodNil[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodNil[T, *T] {
	return &ZodNil[T, *T]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new ZodNil preserving the current constraint type.
func (z *ZodNil[T, R]) withInternals(in *core.ZodTypeInternals) *ZodNil[T, R] {
	return &ZodNil[T, R]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodNil[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodNil[T, R]); ok {
		z.internals = src.internals
	}
}

// extractNilValue extracts base type T from constraint type R.
// Handles nil values safely since nil is the primary value for this type.
func extractNilValue[T any, R any](value R) T {
	if any(value) == nil {
		var zero T
		return zero
	}
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

// convertToNilConstraintValue converts any value to constraint type R.
func convertToNilConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match.
	if r, ok := any(value).(R); ok { //nolint:unconvert
		return r, true
	}

	// Handle pointer constraint: wrap value in *any.
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

// newZodNilFromDef constructs a new ZodNil from the given definition.
func newZodNilFromDef[T any, R any](def *ZodNilDef) *ZodNil[T, R] {
	internals := &ZodNilInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Nil type accepts nil values by default.
	internals.Nilable = true

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		nilDef := &ZodNilDef{ZodTypeDef: *newDef}
		return any(newZodNilFromDef[T, R](nilDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodNil[T, R]{internals: internals}
}

// Nil creates a nil schema that only accepts nil values.
//
// Usage:
//
//	Nil()                    // no parameters
//	Nil("custom error")      // string shorthand
//	Nil(SchemaParams{...})   // full parameters
func Nil(params ...any) *ZodNil[any, any] {
	return NilTyped[any, any](params...)
}

// NilPtr creates a nil schema with pointer constraint.
func NilPtr(params ...any) *ZodNil[any, *any] {
	return NilTyped[any, *any](params...)
}

// NilTyped creates a typed nil schema with explicit generic constraints.
func NilTyped[T any, R any](params ...any) *ZodNil[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodNilDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNil,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)

	return newZodNilFromDef[T, R](def)
}
