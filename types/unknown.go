package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// ZodUnknownDef holds the configuration for unknown value validation.
type ZodUnknownDef struct {
	core.ZodTypeDef
}

// ZodUnknownInternals contains the internal state of an unknown validator.
type ZodUnknownInternals struct {
	core.ZodTypeInternals
	Def *ZodUnknownDef
}

// ZodUnknown validates unknown values with dual generic parameters.
// T is the base type, R is the constraint type (T or *T).
type ZodUnknown[T any, R any] struct {
	internals *ZodUnknownInternals
}

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
func convertUnknownResult[T any, R any](result any, _ *core.ParseContext, _ core.ZodTypeCode) (R, error) {
	var zero R
	if result == nil {
		return zero, nil
	}
	if r, ok := result.(R); ok {
		return r, nil
	}
	// R=*any: wrap value in pointer.
	if _, ok := any(zero).(*any); ok {
		v := result
		return any(&v).(R), nil
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
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// MustStrictParse panics on validation error with compile-time type safety.
func (z *ZodUnknown[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodUnknown[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

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
func (z *ZodUnknown[T, R]) Describe(desc string) *ZodUnknown[T, R] {
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

// Overwrite applies data transformation while preserving type structure.
func (z *ZodUnknown[T, R]) Overwrite(transform func(T) T, params ...any) *ZodUnknown[T, R] {
	check := checks.NewZodCheckOverwrite(func(value any) any {
		if v, ok := value.(T); ok {
			return transform(v)
		}
		return value
	}, params...)
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Refine applies a custom validation function matching the schema's output type R.
func (z *ZodUnknown[T, R]) Refine(fn func(R) bool, args ...any) *ZodUnknown[T, R] {
	wrapper := func(v any) bool {
		if r, ok := v.(R); ok {
			return fn(r)
		}
		// R=*any: wrap value in pointer for pointer constraint types.
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
func (z *ZodUnknown[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodUnknown[T, R] {
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
func (z *ZodUnknown[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodUnknown[T, R] {
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

// With is an alias for Check for Zod v4 API compatibility.
func (z *ZodUnknown[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodUnknown[T, R] {
	return z.Check(fn, params...)
}

// Transform applies a transformation function.
func (z *ZodUnknown[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractUnknownValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Pipe creates a pipeline to another schema.
func (z *ZodUnknown[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	fn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractUnknownValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, fn)
}

// And creates an intersection with another schema.
func (z *ZodUnknown[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodUnknown[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

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
		orig := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = orig
	}
}

// extractUnknownValue extracts base type T from constraint type R.
func extractUnknownValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok && v != nil {
		return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
	}
	return any(value).(T)
}

// newZodUnknownFromDef constructs a new ZodUnknown from the given definition.
func newZodUnknownFromDef[T any, R any](def *ZodUnknownDef) *ZodUnknown[T, R] {
	in := &ZodUnknownInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	// Unknown type defaults to accepting nil values like Zod v4.
	in.Nilable = true

	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodUnknownDef{ZodTypeDef: *newDef}
		return any(newZodUnknownFromDef[T, R](d)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodUnknown[T, R]{internals: in}
}

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
	p := utils.NormalizeParams(params...)

	def := &ZodUnknownDef{
		ZodTypeDef: core.ZodTypeDef{
			Type: core.ZodTypeUnknown,
		},
	}

	if p != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, p)
	}

	return newZodUnknownFromDef[T, R](def)
}
