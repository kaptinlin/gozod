package types

import (
	"reflect"
	"slices"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodLiteralDef defines the configuration for a literal schema.
type ZodLiteralDef[T comparable] struct {
	core.ZodTypeDef
	Values []T
}

// ZodLiteralInternals contains literal validator internal state.
type ZodLiteralInternals[T comparable] struct {
	core.ZodTypeInternals
	Def *ZodLiteralDef[T]
}

// ZodLiteral represents a literal validation schema for exact value matching.
type ZodLiteral[T comparable, R any] struct {
	internals *ZodLiteralInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema.
func (z *ZodLiteral[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodLiteral[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodLiteral[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// withCheck clones internals, adds a check, and returns a new instance.
func (z *ZodLiteral[T, R]) withCheck(check core.ZodCheck) *ZodLiteral[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// validateLiteral ensures the value is one of the allowed literal values,
// then delegates to engine.ApplyChecks for Refine/Check logic.
func (z *ZodLiteral[T, R]) validateLiteral(value T, chks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	if !z.Contains(value) {
		return value, issues.CreateInvalidTypeError(core.ZodTypeLiteral, value, ctx)
	}
	return engine.ApplyChecks[T](value, chks, ctx)
}

// Parse validates the input value and returns a typed result.
func (z *ZodLiteral[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeLiteral,
		z.validateLiteral,
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// StrictParse provides compile-time type safety by requiring exact type matching.
func (z *ZodLiteral[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeLiteral,
		z.validateLiteral,
		ctx...,
	)
}

// MustParse validates input and panics on error.
func (z *ZodLiteral[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustStrictParse validates input with strict type matching and panics on error.
func (z *ZodLiteral[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// MustParseAny validates input and panics on error, returning an untyped result.
func (z *ZodLiteral[T, R]) MustParseAny(input any, ctx ...*core.ParseContext) any {
	result, err := z.ParseAny(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodLiteral[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts nil, with constraint type *T.
func (z *ZodLiteral[T, R]) Optional() *ZodLiteral[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil, with constraint type *T.
func (z *ZodLiteral[T, R]) Nilable() *ZodLiteral[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility.
func (z *ZodLiteral[T, R]) Nullish() *ZodLiteral[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a default value returned when input is nil, bypassing validation.
func (z *ZodLiteral[T, R]) Default(v T) *ZodLiteral[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// Prefault sets a prefault value that goes through the full parsing pipeline when input is nil.
func (z *ZodLiteral[T, R]) Prefault(v T) *ZodLiteral[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a factory function that provides the default value when input is nil.
func (z *ZodLiteral[T, R]) DefaultFunc(fn func() T) *ZodLiteral[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// PrefaultFunc sets a factory function that provides the prefault value through the full parsing pipeline.
func (z *ZodLiteral[T, R]) PrefaultFunc(fn func() T) *ZodLiteral[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Describe registers a description in the global registry.
func (z *ZodLiteral[T, R]) Describe(description string) *ZodLiteral[T, R] {
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

// Meta stores metadata for this literal schema.
func (z *ZodLiteral[T, R]) Meta(meta core.GlobalMeta) *ZodLiteral[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a typed custom validation function.
func (z *ZodLiteral[T, R]) Refine(fn func(T) bool, params ...any) *ZodLiteral[T, R] {
	wrapper := func(data any) bool {
		var zero R
		switch any(zero).(type) {
		case *T:
			if data == nil {
				return true
			}
			if typed, ok := data.(T); ok {
				return fn(typed)
			}
			return false
		default:
			if typed, ok := data.(T); ok {
				return fn(typed)
			}
			return false
		}
	}
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
}

// RefineAny adds a flexible custom validation function accepting any input.
func (z *ZodLiteral[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodLiteral[T, R] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Value returns the single literal value. Panics if the schema has multiple values.
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

// Contains reports whether v is one of the allowed literal values.
func (z *ZodLiteral[T, R]) Contains(v T) bool {
	return slices.Contains(z.internals.Def.Values, v)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withInternals creates a new instance preserving the generic constraint type R.
func (z *ZodLiteral[T, R]) withInternals(in *core.ZodTypeInternals) *ZodLiteral[T, R] {
	return &ZodLiteral[T, R]{internals: &ZodLiteralInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// withPtrInternals creates a new instance with pointer constraint type *T.
func (z *ZodLiteral[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodLiteral[T, *T] {
	return &ZodLiteral[T, *T]{internals: &ZodLiteralInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// newZodLiteralFromDef constructs a new ZodLiteral from the given definition.
func newZodLiteralFromDef[T comparable, R any](def *ZodLiteralDef[T]) *ZodLiteral[T, R] {
	internals := &ZodLiteralInternals[T]{
		core.ZodTypeInternals{
			Type: core.ZodTypeLiteral,
		},
		def,
	}

	values := make(map[any]struct{})
	if len(def.Values) > 0 {
		firstValRV := reflect.ValueOf(def.Values[0])
		if firstValRV.Kind() == reflect.Pointer {
			panic("pointer literals are not supported in Literal() constructor")
		}
		if firstValRV.Type().Comparable() {
			for _, val := range def.Values {
				values[val] = struct{}{}
			}
		}
	}
	internals.Values = values

	return &ZodLiteral[T, R]{internals: internals}
}

// LiteralTyped creates a literal schema with explicit type parameters.
func LiteralTyped[T comparable, R any](value T, params ...any) *ZodLiteral[T, R] {
	def := &ZodLiteralDef[T]{
		Values: []T{value},
	}
	if len(params) > 0 {
		utils.ApplySchemaParams(&def.ZodTypeDef, utils.NormalizeParams(params...))
	}
	return newZodLiteralFromDef[T, R](def)
}

// Literal creates a literal schema for a single value with type inference.
func Literal[T comparable](value T, params ...any) *ZodLiteral[T, T] {
	return LiteralTyped[T, T](value, params...)
}

// LiteralPtr creates a literal schema for a single value with a pointer return type.
func LiteralPtr[T comparable](value T, params ...any) *ZodLiteral[T, *T] {
	return Literal(value, params...).Nilable()
}

// LiteralOf creates a multi-value literal schema with type inference.
func LiteralOf[T comparable](values []T, params ...any) *ZodLiteral[T, T] {
	def := &ZodLiteralDef[T]{
		Values: values,
	}
	if len(params) > 0 {
		utils.ApplySchemaParams(&def.ZodTypeDef, utils.NormalizeParams(params...))
	}
	return newZodLiteralFromDef[T, T](def)
}

// LiteralPtrOf creates a multi-value literal schema with a pointer return type.
func LiteralPtrOf[T comparable](values []T, params ...any) *ZodLiteral[T, *T] {
	return LiteralOf(values, params...).Nilable()
}
