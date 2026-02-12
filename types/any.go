package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// ZodAnyDef holds the configuration for any-value validation.
type ZodAnyDef struct {
	core.ZodTypeDef
}

// ZodAnyInternals contains the internal state of an any validator.
type ZodAnyInternals struct {
	core.ZodTypeInternals
	Def *ZodAnyDef
}

// ZodAny validates any value with dual generic parameters.
// T is the base type, and R is the constraint type (T or *T).
type ZodAny[T any, R any] struct {
	internals *ZodAnyInternals
}

// Internals returns the internal state of the schema.
func (z *ZodAny[T, R]) Internals() *core.ZodTypeInternals {
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

// convertAnyResult handles nil and pointer conversion for the any type.
// This is necessary because engine.ConvertToConstraintType cannot distinguish nil any values.
func convertAnyResult[T any, R any](result any, _ *core.ParseContext, _ core.ZodTypeCode) (R, error) {
	var zero R
	if result == nil {
		return zero, nil
	}

	if r, ok := result.(R); ok {
		return r, nil
	}

	// For R=*any, wrap the value in a pointer.
	if _, ok := any(zero).(*any); ok {
		v := result
		return any(&v).(R), nil
	}

	return zero, nil
}

// Parse validates the input and returns a value matching the generic type R.
func (z *ZodAny[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	typ := z.internals.Type
	if typ == "" {
		typ = core.ZodTypeAny
	}
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		typ,
		engine.ApplyChecks[T],
		convertAnyResult[T, R],
		ctx...,
	)
}

// StrictParse provides type-safe parsing with compile-time guarantees.
func (z *ZodAny[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	typ := z.internals.Type
	if typ == "" {
		typ = core.ZodTypeAny
	}
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		typ,
		engine.ApplyChecks[T],
		ctx...,
	)
}

// MustParse validates the input and panics if validation fails.
func (z *ZodAny[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// MustStrictParse validates the input with compile-time type safety and panics if validation fails.
func (z *ZodAny[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates the input and returns an untyped result.
func (z *ZodAny[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a pointer-typed schema that accepts missing values.
func (z *ZodAny[T, R]) Optional() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional returns a schema that accepts absent keys but rejects explicit nil values.
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

// Nullish returns a schema that combines optional and nilable modifiers.
func (z *ZodAny[T, R]) Nullish() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and forces the return type to T.
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

// Default sets a default value to use when the input is nil.
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

// Prefault provides a fallback value that goes through the full parsing pipeline.
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

// Meta stores the provided metadata in the global registry.
func (z *ZodAny[T, R]) Meta(meta core.GlobalMeta) *ZodAny[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodAny[T, R]) Describe(desc string) *ZodAny[T, R] {
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

// Overwrite applies data transformation while preserving the type structure.
func (z *ZodAny[T, R]) Overwrite(transform func(T) T, params ...any) *ZodAny[T, R] {
	check := checks.NewZodCheckOverwrite(
		func(value any) any {
			if v, ok := value.(T); ok {
				return transform(v)
			}
			return value
		},
		params...,
	)

	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Refine applies a custom validation function that matches the schema's output type R.
func (z *ZodAny[T, R]) Refine(fn func(R) bool, args ...any) *ZodAny[T, R] {
	wrapper := func(v any) bool {
		if r, ok := v.(R); ok {
			return fn(r)
		}

		// For R=*any, wrap the value in a pointer for pointer constraint types.
		var zero R
		if _, ok := any(zero).(*any); ok {
			if v == nil {
				return fn(any((*any)(nil)).(R))
			}
			val := v
			return fn(any(&val).(R))
		}

		return false
	}

	param := utils.FirstParam(args...)
	norm := utils.NormalizeParams(param)

	var msg any
	if norm.Error != nil {
		msg = norm
	}

	check := checks.NewCustom[any](wrapper, msg)
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// RefineAny provides flexible validation without type conversion.
func (z *ZodAny[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodAny[T, R] {
	param := utils.FirstParam(args...)
	norm := utils.NormalizeParams(param)

	var msg any
	if norm.Error != nil {
		msg = norm
	}

	check := checks.NewCustom[any](fn, msg)
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodAny[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodAny[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(R); ok {
			fn(val, payload)
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// With is an alias for Check that provides Zod v4 API compatibility.
func (z *ZodAny[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodAny[T, R] {
	return z.Check(fn, params...)
}

// Transform applies a transformation function.
func (z *ZodAny[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractAnyValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Pipe creates a pipeline to another schema.
func (z *ZodAny[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	fn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractAnyValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, fn)
}

// And creates an intersection with another schema.
func (z *ZodAny[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodAny[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

func (z *ZodAny[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodAny[T, *T] {
	return &ZodAny[T, *T]{
		internals: &ZodAnyInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

func (z *ZodAny[T, R]) withInternals(in *core.ZodTypeInternals) *ZodAny[T, R] {
	return &ZodAny[T, R]{
		internals: &ZodAnyInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// CloneFrom copies configuration from another schema.
func (z *ZodAny[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodAny[T, R]); ok {
		orig := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = orig
	}
}

// extractAnyValue extracts the base type T from the constraint type R.
func extractAnyValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok && v != nil {
		return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion.
	}
	return any(value).(T)
}

// newZodAnyFromDef constructs a new ZodAny instance from the given definition.
func newZodAnyFromDef[T any, R any](def *ZodAnyDef) *ZodAny[T, R] {
	in := &ZodAnyInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// The any type defaults to accepting nil values, matching Zod v4 behavior.
	in.Nilable = true

	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodAnyDef{ZodTypeDef: *newDef}
		return any(newZodAnyFromDef[T, R](d)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodAny[T, R]{internals: in}
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
	p := utils.NormalizeParams(params...)

	def := &ZodAnyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type: core.ZodTypeAny,
		},
	}

	if p != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, p)
	}

	return newZodAnyFromDef[T, R](def)
}
