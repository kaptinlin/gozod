package types

import (
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// ZodLiteralDef defines the structure of a literal schema
type ZodLiteralDef[T comparable] struct {
	core.ZodTypeDef
	Values []T
}

// ZodLiteralInternals holds the internal state of a literal schema
type ZodLiteralInternals[T comparable] struct {
	core.ZodTypeInternals
	Def *ZodLiteralDef[T]
}

type ZodLiteral[T comparable, R any] struct {
	internals *ZodLiteralInternals[T]
}

func (z *ZodLiteral[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates the input value and returns a typed result.
// It leverages engine.ParsePrimitive so that all modifiers, checks, and transforms
// work consistently across the code-base.
func (z *ZodLiteral[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	// Validator ensures the value is one of the allowed literal values and then
	// delegates to engine.ApplyChecks to run Refine / Check logic.
	validator := func(value T, checks []core.ZodCheck, c *core.ParseContext) (T, error) {
		if !z.Contains(value) {
			// Re-use standardized invalid type error helper so error formatting
			// matches other primitives. We treat a wrong literal as invalid type.
			return value, engine.CreateInvalidTypeError(core.ZodTypeLiteral, value, c)
		}
		// Apply additional checks (Refine, Overwrite, etc.).
		return engine.ApplyChecks[T](value, checks, c)
	}

	// Ensure validator is always executed (even for pointer inputs) by making
	// sure there is at least one check present. We inject a noop custom check
	// when the schema currently has none.
	internalsCopy := z.internals.ZodTypeInternals.Clone()
	if len(internalsCopy.Checks) == 0 {
		internalsCopy.AddCheck(checks.NewCustom[any](func(v any) bool { return true }))
	}

	return engine.ParsePrimitive[T, R](
		input,
		internalsCopy,
		core.ZodTypeLiteral,
		validator,
		engine.ConvertToConstraintType[T, R],
		ctx..., // pass-through context
	)
}

func (z *ZodLiteral[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	val, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return val
}

func (z *ZodLiteral[T, R]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	val, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return val
}

// ParseAny is a zero-overhead wrapper around Parse and is used by reflection-based
// callers that don’t know the concrete type parameters at compile-time.
func (z *ZodLiteral[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	res, err := z.Parse(input, ctx...)
	if err != nil {
		return nil, err
	}
	return any(res), nil
}

func (z *ZodLiteral[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

func (z *ZodLiteral[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Optional returns a new schema that allows the value to be nil.
// It changes the constraint type from R to *T.
func (z *ZodLiteral[T, R]) Optional() *ZodLiteral[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a new schema that allows the value to be nil.
// It changes the constraint type from R to *T.
func (z *ZodLiteral[T, R]) Nilable() *ZodLiteral[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodLiteral[T, R]) Nullish() *ZodLiteral[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value for the schema.
func (z *ZodLiteral[T, R]) Default(v T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// Prefault sets a fallback value if parsing fails.
func (z *ZodLiteral[T, R]) Prefault(v T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

func (z *ZodLiteral[T, R]) DefaultFunc(fn func() T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

func (z *ZodLiteral[T, R]) PrefaultFunc(fn func() T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Refine adds a custom validation check to the schema.
func (z *ZodLiteral[T, R]) Refine(fn func(T) bool, params ...any) *ZodLiteral[T, R] {
	check := checks.NewCustom[any](
		func(data any) bool {
			if typed, ok := data.(T); ok {
				return fn(typed)
			}
			return false
		},
		utils.NormalizeCustomParams(params...),
	)
	in := z.internals.ZodTypeInternals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// RefineAny adds a custom validation check to the schema, accepting `any`.
func (z *ZodLiteral[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodLiteral[T, R] {
	check := checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...))
	in := z.internals.ZodTypeInternals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// Value returns the first literal value. Useful for single-value literals.
func (z *ZodLiteral[T, R]) Value() T {
	if len(z.internals.Def.Values) == 0 {
		var zero T
		return zero
	}
	if len(z.internals.Def.Values) > 1 {
		panic("Value() cannot be used on multi-value literal schema; use Values() instead")
	}
	return z.internals.Def.Values[0]
}

// Values returns all literal values.
func (z *ZodLiteral[T, R]) Values() []T {
	return z.internals.Def.Values
}

// Contains checks if a given value is one of the allowed literal values.
func (z *ZodLiteral[T, R]) Contains(v T) bool {
	for _, val := range z.internals.Def.Values {
		if val == v {
			return true
		}
	}
	return false
}

// withInternals creates a new instance with the same constraint type R.
func (z *ZodLiteral[T, R]) withInternals(in *core.ZodTypeInternals) *ZodLiteral[T, R] {
	return &ZodLiteral[T, R]{internals: &ZodLiteralInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withPtrInternals creates a new instance with a pointer constraint type *T.
func (z *ZodLiteral[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodLiteral[T, *T] {
	return &ZodLiteral[T, *T]{internals: &ZodLiteralInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func newZodLiteralFromDef[T comparable, R any](def *ZodLiteralDef[T]) *ZodLiteral[T, R] {
	internals := &ZodLiteralInternals[T]{
		core.ZodTypeInternals{
			Type: core.ZodTypeLiteral,
		},
		def,
	}

	values := make(map[any]struct{})
	if len(def.Values) > 0 {
		// Disallow pointer literals – they introduce identity semantics and are
		// not supported by design (see tests).
		firstValRV := reflect.ValueOf(def.Values[0])
		if firstValRV.Kind() == reflect.Ptr {
			panic("pointer literals are not supported in Literal() constructor")
		}

		// Use reflection to check if the type is comparable before adding to the map.
		if firstValRV.Type().Comparable() {
			for _, val := range def.Values {
				values[val] = struct{}{}
			}
		}
	}
	internals.Values = values

	schema := &ZodLiteral[T, R]{
		internals: internals,
	}
	return schema
}

func LiteralTyped[T comparable, R any](value T, params ...any) *ZodLiteral[T, R] {
	def := &ZodLiteralDef[T]{
		Values: []T{value},
	}

	if len(params) > 0 {
		normalizedParams := utils.NormalizeParams(params...)
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodLiteralFromDef[T, R](def)
}

func Literal[T comparable](value T, params ...any) *ZodLiteral[T, T] {
	return LiteralTyped[T, T](value, params...)
}

// LiteralPtr creates a literal schema for a single value with a pointer return type
func LiteralPtr[T comparable](value T, params ...any) *ZodLiteral[T, *T] {
	return Literal(value, params...).Nilable()
}

func LiteralOf[T comparable](values []T, params ...any) *ZodLiteral[T, T] {
	def := &ZodLiteralDef[T]{
		Values: values,
	}

	if len(params) > 0 {
		normalizedParams := utils.NormalizeParams(params...)
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodLiteralFromDef[T, T](def)
}

// LiteralPtrOf creates a multi-value literal schema with a pointer return type
func LiteralPtrOf[T comparable](values []T, params ...any) *ZodLiteral[T, *T] {
	return LiteralOf[T](values, params...).Nilable()
}
