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

// ZodAnyDef defines the configuration for any value validation.
type ZodAnyDef struct {
	core.ZodTypeDef
}

// ZodAnyInternals contains any validator internal state.
type ZodAnyInternals struct {
	core.ZodTypeInternals
	Def *ZodAnyDef
}

// ZodAny validates any value with dual generic parameters.
// T is the base type (any), R is the constraint type (any or *any).
type ZodAny[T any, R any] struct {
	internals *ZodAnyInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodAny[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodAny[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodAny[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// convertAnyResult handles nil and pointer conversion for the any type,
// where engine.ConvertToConstraintType cannot distinguish nil any.
func convertAnyResult[T any, R any](
	result any,
	_ *core.ParseContext,
	_ core.ZodTypeCode,
) (R, error) {
	var zero R
	if result == nil {
		return zero, nil
	}
	// Direct type match (covers R=any case)
	if r, ok := result.(R); ok {
		return r, nil
	}
	// Handle R=*any: wrap value in pointer
	if _, ok := any(zero).(*any); ok {
		valueCopy := result
		return any(&valueCopy).(R), nil
	}
	return zero, nil
}

// Parse validates input and returns a value matching the generic type R.
func (z *ZodAny[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeAny
	}
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		engine.ApplyChecks[T],
		convertAnyResult[T, R],
		ctx...,
	)
}

// StrictParse provides type-safe parsing with compile-time guarantees.
func (z *ZodAny[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	expectedType := z.internals.Type
	if expectedType == "" {
		expectedType = core.ZodTypeAny
	}
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		expectedType,
		engine.ApplyChecks[T],
		ctx...,
	)
}

// MustParse panics on validation error.
func (z *ZodAny[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustStrictParse panics on validation error with compile-time type safety.
func (z *ZodAny[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodAny[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a pointer-typed schema that accepts missing values.
func (z *ZodAny[T, R]) Optional() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodAny[T, R]) ExactOptional() *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a pointer-typed schema that accepts nil values.
func (z *ZodAny[T, R]) Nilable() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodAny[T, R]) Nullish() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and forces return type to T.
// Sets internals.Type to ZodTypeNonOptional for error reporting.
func (z *ZodAny[T, R]) NonOptional() *ZodAny[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodAny[T, T]{
		internals: &ZodAnyInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default sets a default value used when input is nil.
func (z *ZodAny[T, R]) Default(v T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a function-based default value.
func (z *ZodAny[T, R]) DefaultFunc(fn func() T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault provides a fallback value through the full parsing pipeline.
func (z *ZodAny[T, R]) Prefault(v T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides a dynamic fallback value.
func (z *ZodAny[T, R]) PrefaultFunc(fn func() T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata in the global registry.
func (z *ZodAny[T, R]) Meta(meta core.GlobalMeta) *ZodAny[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodAny[T, R]) Describe(description string) *ZodAny[T, R] {
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
func (z *ZodAny[T, R]) Overwrite(transform func(T) T, params ...any) *ZodAny[T, R] {
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
func (z *ZodAny[T, R]) Refine(fn func(R) bool, args ...any) *ZodAny[T, R] {
	wrapper := func(v any) bool {
		// Direct type match
		if r, ok := v.(R); ok {
			return fn(r)
		}
		// Handle *any: wrap value in pointer for pointer constraint types
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
func (z *ZodAny[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodAny[T, R] {
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
func (z *ZodAny[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodAny[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// With is an alias for Check for Zod v4 API compatibility.
func (z *ZodAny[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodAny[T, R] {
	return z.Check(fn, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function.
func (z *ZodAny[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractAnyValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a pipeline to another schema.
func (z *ZodAny[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractAnyValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// COMPOSITION METHODS
// =============================================================================

// And creates an intersection with another schema.
func (z *ZodAny[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodAny[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodAny[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodAny[T, *T] {
	return &ZodAny[T, *T]{internals: &ZodAnyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodAny[T, R]) withInternals(in *core.ZodTypeInternals) *ZodAny[T, R] {
	return &ZodAny[T, R]{internals: &ZodAnyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodAny[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodAny[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractAnyValue extracts base type T from constraint type R.
func extractAnyValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok && v != nil {
		return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
	}
	return any(value).(T)
}

// =============================================================================
// CONSTRUCTORS
// =============================================================================

// newZodAnyFromDef constructs a new ZodAny from the given definition.
func newZodAnyFromDef[T any, R any](def *ZodAnyDef) *ZodAny[T, R] {
	internals := &ZodAnyInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Any type defaults to accepting nil values like Zod v4.
	internals.Nilable = true

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		anyDef := &ZodAnyDef{ZodTypeDef: *newDef}
		return any(newZodAnyFromDef[T, R](anyDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodAny[T, R]{internals: internals}
}

// Any creates a schema that accepts any value.
func Any(params ...any) *ZodAny[any, any] {
	return AnyTyped[any, any](params...)
}

// AnyPtr creates a schema that accepts any value with pointer constraint.
func AnyPtr(params ...any) *ZodAny[any, *any] {
	return AnyTyped[any, *any](params...)
}

// AnyTyped creates a typed any schema with generic constraints.
func AnyTyped[T any, R any](params ...any) *ZodAny[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodAnyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeAny,
			Checks: []core.ZodCheck{},
		},
	}

	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodAnyFromDef[T, R](def)
}
