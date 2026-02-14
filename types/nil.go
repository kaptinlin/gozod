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

// ZodNilDef holds the configuration for nil validation.
type ZodNilDef struct {
	core.ZodTypeDef
}

// ZodNilInternals contains the internal state of a nil validator.
type ZodNilInternals struct {
	core.ZodTypeInternals
	Def *ZodNilDef
}

// ZodNil is a type-safe nil validation schema.
// T is the base type and R is the constraint type (may be a pointer for modifiers like Optional/Nilable).
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

// nilValidator validates that the input is nil.
// When internals is provided, custom error messages from schema parameters are used.
func nilValidator[T any](internals *core.ZodTypeInternals) func(T, []core.ZodCheck, *core.ParseContext) (T, error) {
	return func(value T, chks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
		if any(value) == nil || reflectx.IsNil(value) {
			return engine.ApplyChecks(value, chks, ctx)
		}

		var zero T
		if internals != nil {
			return zero, issues.CreateInvalidTypeErrorWithInst(core.ZodTypeNil, value, ctx, internals)
		}
		return zero, issues.CreateInvalidTypeError(core.ZodTypeNil, value, ctx)
	}
}

// convertNilResult converts a parsed result to constraint type R.
// Nil type needs special handling because nil interface values don't match type switch cases in the generic ConvertToConstraintType.
func convertNilResult[T any, R any](
	result any, ctx *core.ParseContext, expectedType core.ZodTypeCode,
) (R, error) {
	if result == nil || reflectx.IsNil(result) {
		var zero R
		return zero, nil
	}
	return engine.ConvertToConstraintType[T, R](result, ctx, expectedType)
}

// Parse validates input and returns a value matching the constraint type R.
func (z *ZodNil[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive(
		input, &z.internals.ZodTypeInternals, z.expectedType(),
		nilValidator[T](&z.internals.ZodTypeInternals), convertNilResult[T, R], ctx...,
	)
}

// MustParse panics on validation failure.
func (z *ZodNil[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// StrictParse validates input with compile-time type safety.
func (z *ZodNil[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict(
		input, &z.internals.ZodTypeInternals, z.expectedType(),
		nilValidator[T](nil), ctx...,
	)
}

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodNil[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodNil[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodNil[T, R]) Optional() *ZodNil[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodNil[T, R]) ExactOptional() *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodNil[T, R]) Nilable() *ZodNil[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodNil[T, R]) Nullish() *ZodNil[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits).
func (z *ZodNil[T, R]) Default(v T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits).
func (z *ZodNil[T, R]) DefaultFunc(fn func() T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodNil[T, R]) Prefault(v T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodNil[T, R]) PrefaultFunc(fn func() T) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata in the global registry.
func (z *ZodNil[T, R]) Meta(meta core.GlobalMeta) *ZodNil[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodNil[T, R]) Describe(desc string) *ZodNil[T, R] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = desc
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed value.
func (z *ZodNil[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractNilValue[T, R](input), ctx)
	}
	return core.NewZodTransform(z, wrapper)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodNil[T, R]) Overwrite(transform func(R) R, params ...any) *ZodNil[T, R] {
	fn := func(input any) any {
		if converted, ok := convertToNilConstraintValue[T, R](input); ok {
			return transform(converted)
		}
		return input
	}
	check := checks.NewZodCheckOverwrite(fn, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodNil[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	fn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractNilValue[T, R](input), ctx)
	}
	return core.NewZodPipe(z, target, fn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies a custom validation function matching the schema's output type R.
func (z *ZodNil[T, R]) Refine(fn func(R) bool, params ...any) *ZodNil[T, R] {
	wrapper := func(v any) bool {
		if converted, ok := convertToNilConstraintValue[T, R](v); ok {
			return fn(converted)
		}
		return false
	}
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}
	check := checks.NewCustom[any](wrapper, msg)
	return z.withCheck(check)
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodNil[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodNil[T, R] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}
	check := checks.NewCustom[any](fn, msg)
	return z.withCheck(check)
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodNil[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodNil[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := convertToNilConstraintValue[T, R](payload.Value()); ok {
			fn(val, payload)
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodNil[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodNil[T, R] {
	return z.Check(fn, params...)
}

// And creates an intersection with another schema.
func (z *ZodNil[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodNil[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// NonOptional removes the optional flag, returning a base constraint.
func (z *ZodNil[T, R]) NonOptional() *ZodNil[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodNil[T, T]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// expectedType returns the schema's type code, defaulting to ZodTypeNil.
func (z *ZodNil[T, R]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeNil
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodNil[T, R]) withCheck(check core.ZodCheck) *ZodNil[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new *T schema from cloned internals.
func (z *ZodNil[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodNil[T, *T] {
	return &ZodNil[T, *T]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new schema preserving generic type R.
func (z *ZodNil[T, R]) withInternals(in *core.ZodTypeInternals) *ZodNil[T, R] {
	return &ZodNil[T, R]{internals: &ZodNilInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodNil[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodNil[T, R]); ok {
		orig := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = orig
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
		return any(new(value)).(R), true
	}

	return zero, false
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// newZodNilFromDef constructs a new ZodNil from a definition.
func newZodNilFromDef[T any, R any](def *ZodNilDef) *ZodNil[T, R] {
	in := &ZodNilInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Nil type accepts nil values by default.
	in.Nilable = true

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		nd := &ZodNilDef{ZodTypeDef: *d}
		return any(newZodNilFromDef[T, R](nd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodNil[T, R]{internals: in}
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

// NilTyped is the generic constructor for nil schemas.
func NilTyped[T any, R any](params ...any) *ZodNil[T, R] {
	sp := utils.NormalizeParams(params...)

	def := &ZodNilDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeNil,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, sp)

	return newZodNilFromDef[T, R](def)
}
