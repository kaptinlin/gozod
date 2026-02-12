package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// BoolConstraint restricts values to bool or *bool.
type BoolConstraint interface {
	~bool | ~*bool
}

// ZodBoolDef holds the configuration for boolean validation.
type ZodBoolDef struct {
	core.ZodTypeDef
}

// ZodBoolInternals contains the internal state of a boolean validator.
type ZodBoolInternals struct {
	core.ZodTypeInternals
	Def *ZodBoolDef
}

// ZodBool is a type-safe boolean validation schema.
type ZodBool[T BoolConstraint] struct {
	internals *ZodBoolInternals
}

// Internals returns the internal state of the schema.
func (z *ZodBool[T]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodBool[T]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodBool[T]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Coerce converts input to bool.
func (z *ZodBool[T]) Coerce(input any) (any, bool) {
	r, err := coerce.ToBool(input)
	return r, err == nil
}

// Parse validates input and returns a value matching the generic type T.
func (z *ZodBool[T]) Parse(input any, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitive[bool, T](input, &z.internals.ZodTypeInternals, z.expectedType(), engine.ApplyChecks[bool], engine.ConvertToConstraintType[bool, T], ctx...)
}

// MustParse panics on validation failure.
func (z *ZodBool[T]) MustParse(input any, ctx ...*core.ParseContext) T {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// StrictParse validates input with compile-time type safety.
func (z *ZodBool[T]) StrictParse(input T, ctx ...*core.ParseContext) (T, error) {
	return engine.ParsePrimitiveStrict[bool, T](input, &z.internals.ZodTypeInternals, z.expectedType(), engine.ApplyChecks[bool], ctx...)
}

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodBool[T]) MustStrictParse(input T, ctx ...*core.ParseContext) T {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodBool[T]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodBool[T]) Optional() *ZodBool[*bool] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodBool[T]) ExactOptional() *ZodBool[T] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodBool[T]) Nilable() *ZodBool[*bool] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodBool[T]) Nullish() *ZodBool[*bool] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits).
func (z *ZodBool[T]) Default(v bool) *ZodBool[T] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits).
func (z *ZodBool[T]) DefaultFunc(fn func() bool) *ZodBool[T] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodBool[T]) Prefault(v bool) *ZodBool[T] {
	in := z.internals.Clone()
	var zero T
	switch any(zero).(type) {
	case *bool:
		in.SetPrefaultValue(&v)
	default:
		in.SetPrefaultValue(v)
	}
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodBool[T]) PrefaultFunc(fn func() bool) *ZodBool[T] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		v := fn()
		var zero T
		switch any(zero).(type) {
		case *bool:
			return &v
		default:
			return v
		}
	})
	return z.withInternals(in)
}

// Meta stores metadata in the global registry.
func (z *ZodBool[T]) Meta(meta core.GlobalMeta) *ZodBool[T] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodBool[T]) Describe(desc string) *ZodBool[T] {
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

// Transform applies a transformation function to the parsed value.
func (z *ZodBool[T]) Transform(fn func(bool, *core.RefinementContext) (any, error)) *core.ZodTransform[T, any] {
	wrapper := func(input T, ctx *core.RefinementContext) (any, error) {
		return fn(extractBool(input), ctx)
	}
	return core.NewZodTransform[T, any](z, wrapper)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodBool[T]) Overwrite(transform func(T) T, params ...any) *ZodBool[T] {
	fn := func(input any) any {
		v, ok := convertToBoolType[T](input)
		if !ok {
			return input
		}
		return transform(v)
	}
	check := checks.NewZodCheckOverwrite(fn, params...)
	return z.withCheck(check)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodBool[T]) Pipe(target core.ZodType[any]) *core.ZodPipe[T, any] {
	fn := func(input T, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractBool(input), ctx)
	}
	return core.NewZodPipe[T, any](z, target, fn)
}

// convertToBoolType converts only bool values to the target bool type T.
func convertToBoolType[T BoolConstraint](v any) (T, bool) {
	var zero T

	if v == nil {
		switch any(zero).(type) {
		case *bool:
			return zero, true
		default:
			return zero, false
		}
	}

	var boolValue bool
	var isValid bool

	switch val := v.(type) {
	case bool:
		boolValue, isValid = val, true
	case *bool:
		if val != nil {
			boolValue, isValid = *val, true
		}
	default:
		return zero, false
	}

	if !isValid {
		return zero, false
	}

	switch any(zero).(type) {
	case bool:
		return any(boolValue).(T), true
	case *bool:
		return any(&boolValue).(T), true
	default:
		return zero, false
	}
}

// Refine applies a custom validation function matching the schema's output type T.
func (z *ZodBool[T]) Refine(fn func(T) bool, params ...any) *ZodBool[T] {
	wrapper := func(v any) bool {
		var zero T
		switch any(zero).(type) {
		case bool:
			if v == nil {
				return false
			}
			if b, ok := v.(bool); ok {
				return fn(any(b).(T))
			}
			return false
		case *bool:
			if v == nil {
				return fn(any((*bool)(nil)).(T))
			}
			if b, ok := v.(bool); ok {
				cp := b
				return fn(any(&cp).(T))
			}
			return false
		default:
			return false
		}
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
func (z *ZodBool[T]) RefineAny(fn func(any) bool, params ...any) *ZodBool[T] {
	sp := utils.NormalizeParams(params...)
	var msg any
	if sp.Error != nil {
		msg = sp.Error
	}
	check := checks.NewCustom[any](fn, msg)
	return z.withCheck(check)
}

// And creates an intersection with another schema.
func (z *ZodBool[T]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodBool[T]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// expectedType returns the schema's type code, defaulting to ZodTypeBool.
func (z *ZodBool[T]) expectedType() core.ZodTypeCode {
	if z.internals.Type != "" {
		return z.internals.Type
	}
	return core.ZodTypeBool
}

// withCheck clones internals, adds a check, and returns a new schema (Copy-on-Write).
func (z *ZodBool[T]) withCheck(check core.ZodCheck) *ZodBool[T] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// withPtrInternals creates a new *bool schema from cloned internals.
func (z *ZodBool[T]) withPtrInternals(in *core.ZodTypeInternals) *ZodBool[*bool] {
	return &ZodBool[*bool]{internals: &ZodBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withInternals creates a new schema preserving generic type T.
func (z *ZodBool[T]) withInternals(in *core.ZodTypeInternals) *ZodBool[T] {
	return &ZodBool[T]{internals: &ZodBoolInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodBool[T]) CloneFrom(source any) {
	if src, ok := source.(*ZodBool[T]); ok {
		orig := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = orig
	}
}

// extractBool extracts the underlying bool from a generic constraint type.
func extractBool[T BoolConstraint](value T) bool {
	if ptr, ok := any(value).(*bool); ok {
		return ptr != nil && *ptr
	}
	return any(value).(bool)
}

// newZodBoolFromDef constructs a new ZodBool from a definition.
func newZodBoolFromDef[T BoolConstraint](def *ZodBoolDef) *ZodBool[T] {
	in := &ZodBoolInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def: def,
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		bd := &ZodBoolDef{ZodTypeDef: *d}
		return any(newZodBoolFromDef[T](bd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}

	return &ZodBool[T]{internals: in}
}

// Bool creates a bool validation schema.
func Bool(params ...any) *ZodBool[bool] {
	return BoolTyped[bool](params...)
}

// BoolPtr creates a *bool validation schema.
func BoolPtr(params ...any) *ZodBool[*bool] {
	return BoolTyped[*bool](params...)
}

// BoolTyped is the generic constructor for boolean schemas.
func BoolTyped[T BoolConstraint](params ...any) *ZodBool[T] {
	sp := utils.NormalizeParams(params...)

	def := &ZodBoolDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeBool,
			Checks: []core.ZodCheck{},
		},
	}

	utils.ApplySchemaParams(&def.ZodTypeDef, sp)

	return newZodBoolFromDef[T](def)
}

// CoercedBool creates a coerced bool schema that converts input.
func CoercedBool(args ...any) *ZodBool[bool] {
	s := Bool(args...)
	s.internals.SetCoerce(true)
	return s
}

// CoercedBoolPtr creates a coerced *bool schema that converts input.
func CoercedBoolPtr(args ...any) *ZodBool[*bool] {
	s := BoolPtr(args...)
	s.internals.SetCoerce(true)
	return s
}

// Check adds a custom validation function that can push multiple issues.
func (z *ZodBool[T]) Check(fn func(value T, payload *core.ParsePayload), params ...any) *ZodBool[T] {
	wrapper := func(payload *core.ParsePayload) {
		if val, ok := payload.Value().(T); ok {
			fn(val, payload)
			return
		}

		var zero T
		if _, ok := any(zero).(*bool); ok {
			if b, ok := payload.Value().(bool); ok {
				bCopy := b
				fn(any(&bCopy).(T), payload)
			}
		}
	}
	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	return z.withCheck(check)
}

// With is an alias for Check (Zod v4 API compatibility).
func (z *ZodBool[T]) With(fn func(value T, payload *core.ParsePayload), params ...any) *ZodBool[T] {
	return z.Check(fn, params...)
}

// NonOptional removes the optional flag, returning a bool constraint.
func (z *ZodBool[T]) NonOptional() *ZodBool[bool] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodBool[bool]{
		internals: &ZodBoolInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}
