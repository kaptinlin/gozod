package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodUnknownDef defines the configuration for unknown value validation.
type ZodUnknownDef struct {
	core.ZodTypeDef
}

// ZodUnknownInternals contains unknown validator internal state.
type ZodUnknownInternals struct {
	core.ZodTypeInternals
	Def *ZodUnknownDef
}

// ZodUnknown validates unknown values with dual generic parameters.
// T is the base type (any), R is the constraint type (any or *any).
type ZodUnknown[T any, R any] struct {
	internals *ZodUnknownInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodUnknown[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodUnknown[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodUnknown[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// convertUnknownResult handles nil and pointer conversion for the unknown type,
// where engine.ConvertToConstraintType cannot distinguish nil any.
func convertUnknownResult[T any, R any](
	result any,
	_ *core.ParseContext,
	_ core.ZodTypeCode,
) (R, error) {
	var zero R
	if result == nil {
		return zero, nil
	}
	if r, ok := result.(R); ok {
		return r, nil
	}
	// Handle R=*any: wrap value in pointer.
	if _, ok := any(zero).(*any); ok {
		valueCopy := result
		return any(&valueCopy).(R), nil
	}
	return zero, nil
}

// Parse validates input and returns a value matching the generic type R.
func (z *ZodUnknown[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnknown,
		engine.ApplyChecks[T],
		convertUnknownResult[T, R],
		ctx...,
	)
}

// StrictParse provides type-safe parsing with compile-time guarantees.
func (z *ZodUnknown[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnknown,
		engine.ApplyChecks[T],
		ctx...,
	)
}

// MustParse panics on validation error.
func (z *ZodUnknown[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustStrictParse panics on validation error with compile-time type safety.
func (z *ZodUnknown[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodUnknown[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a pointer-typed schema that accepts missing values.
func (z *ZodUnknown[T, R]) Optional() *ZodUnknown[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodUnknown[T, R]) ExactOptional() *ZodUnknown[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a pointer-typed schema that accepts nil values.
func (z *ZodUnknown[T, R]) Nilable() *ZodUnknown[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodUnknown[T, R]) Nullish() *ZodUnknown[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and forces return type to T.
func (z *ZodUnknown[T, R]) NonOptional() *ZodUnknown[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodUnknown[T, T]{
		internals: &ZodUnknownInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default sets a default value used when input is nil.
func (z *ZodUnknown[T, R]) Default(v T) *ZodUnknown[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a function-based default value.
func (z *ZodUnknown[T, R]) DefaultFunc(fn func() T) *ZodUnknown[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault provides a fallback value through the full parsing pipeline.
func (z *ZodUnknown[T, R]) Prefault(v T) *ZodUnknown[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides a dynamic fallback value.
func (z *ZodUnknown[T, R]) PrefaultFunc(fn func() T) *ZodUnknown[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata in the global registry.
func (z *ZodUnknown[T, R]) Meta(meta core.GlobalMeta) *ZodUnknown[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodUnknown[T, R]) Describe(description string) *ZodUnknown[T, R] {
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
// VALIDATION METHODS
// =============================================================================

// Overwrite applies data transformation while preserving type structure.
func (z *ZodUnknown[T, R]) Overwrite(transform func(T) T, params ...any) *ZodUnknown[T, R] {
	transformFunc := checks.NewZodCheckOverwrite(func(value any) any {
		if v, ok := value.(T); ok {
			return transform(v)
		}
		return value
	}, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(transformFunc)
	return z.withInternals(newInternals)
}

// Refine applies a custom validation function matching the schema's output type R.
func (z *ZodUnknown[T, R]) Refine(fn func(R) bool, args ...any) *ZodUnknown[T, R] {
	wrapper := func(v any) bool {
		if r, ok := v.(R); ok {
			return fn(r)
		}
		// Handle *any: wrap value in pointer for pointer constraint types.
		var zero R
		if _, ok := any(zero).(*any); ok {
			if v == nil {
				return fn(any((*any)(nil)).(R))
			}
			valueCopy := v
			return fn(any(&valueCopy).(R))
		}
		return false
	}

	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var errorMessage any
	if normalizedParams.Error != nil {
		errorMessage = normalizedParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion.
func (z *ZodUnknown[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodUnknown[T, R] {
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var errorMessage any
	if normalizedParams.Error != nil {
		errorMessage = normalizedParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodUnknown[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodUnknown[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(R); ok {
			fn(val, payload)
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// With is an alias for Check for Zod v4 API compatibility.
func (z *ZodUnknown[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodUnknown[T, R] {
	return z.Check(fn, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function.
func (z *ZodUnknown[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractUnknownValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a pipeline to another schema.
func (z *ZodUnknown[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractUnknownValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// COMPOSITION METHODS
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodUnknown[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodUnknown[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodUnknown[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodUnknown[T, *T] {
	return &ZodUnknown[T, *T]{internals: &ZodUnknownInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodUnknown[T, R]) withInternals(in *core.ZodTypeInternals) *ZodUnknown[T, R] {
	return &ZodUnknown[T, R]{internals: &ZodUnknownInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodUnknown[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodUnknown[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractUnknownValue extracts base type T from constraint type R.
func extractUnknownValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok && v != nil {
		return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
	}
	return any(value).(T)
}

// =============================================================================
// CONSTRUCTORS
// =============================================================================

// newZodUnknownFromDef constructs a new ZodUnknown from the given definition.
func newZodUnknownFromDef[T any, R any](def *ZodUnknownDef) *ZodUnknown[T, R] {
	internals := &ZodUnknownInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Unknown type defaults to accepting nil values like Zod v4.
	internals.Nilable = true

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		unknownDef := &ZodUnknownDef{ZodTypeDef: *newDef}
		return any(newZodUnknownFromDef[T, R](unknownDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodUnknown[T, R]{internals: internals}
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// Unknown creates a schema that accepts unknown values.
func Unknown(params ...any) *ZodUnknown[any, any] {
	return UnknownTyped[any, any](params...)
}

// UnknownPtr creates a schema that accepts unknown values with pointer constraint.
func UnknownPtr(params ...any) *ZodUnknown[any, *any] {
	return UnknownTyped[any, *any](params...)
}

// UnknownTyped creates a typed unknown schema with generic constraints.
func UnknownTyped[T any, R any](params ...any) *ZodUnknown[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodUnknownDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeUnknown,
			Checks: []core.ZodCheck{},
		},
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodUnknownFromDef[T, R](def)
}
