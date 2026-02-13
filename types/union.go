package types

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// ZodUnionDef holds the configuration for union validation.
type ZodUnionDef struct {
	core.ZodTypeDef
	Options []core.ZodSchema
}

// ZodUnionInternals holds the internal state for a union schema.
type ZodUnionInternals struct {
	core.ZodTypeInternals
	Def     *ZodUnionDef
	Options []core.ZodSchema
}

// ZodUnion validates that input matches at least one member schema.
// T is the base type, R is the constraint type (T or *T for optional/nilable).
type ZodUnion[T any, R any] struct {
	internals *ZodUnionInternals
}

// Internals returns the internal state for framework usage.
func (z *ZodUnion[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodUnion[T, R]) IsOptional() bool { return z.internals.IsOptional() }

// IsNilable reports whether this schema accepts nil values.
func (z *ZodUnion[T, R]) IsNilable() bool { return z.internals.IsNilable() }

// Parse validates input against all union member schemas, returning the first match.
func (z *ZodUnion[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	parseCtx := resolveCtx(ctx)

	result, err := engine.ParseComplex[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractType,
		z.extractPtr,
		z.validate,
		parseCtx,
	)
	if err != nil {
		var zero R
		return zero, err
	}
	return convertUnionToConstraint[T, R](result), nil
}

func (z *ZodUnion[T, R]) extractType(input any) (any, bool) {
	return input, true
}

func (z *ZodUnion[T, R]) extractPtr(input any) (*any, bool) {
	if input == nil {
		return nil, true
	}
	return &input, true
}

// validate tries each member schema and returns the first successful match.
// It prefers the schema whose result type matches the original input type.
func (z *ZodUnion[T, R]) validate(
	input any,
	chks []core.ZodCheck,
	parseCtx *core.ParseContext,
) (any, error) {
	var (
		match   any
		matched bool
		errs    []error
	)

	inputType := reflect.TypeOf(input)

	for i, opt := range z.internals.Options {
		if opt == nil {
			continue
		}

		result, err := opt.ParseAny(input, parseCtx)
		if err != nil {
			errs = append(errs, fmt.Errorf("option %d: %w", i, err))
			continue
		}

		if len(chks) > 0 {
			result, err = engine.ApplyChecks[any](result, chks, parseCtx)
			if err != nil {
				errs = append(errs, fmt.Errorf("option %d: %w", i, err))
				continue
			}
		}

		// Prefer the schema whose result type matches the original input type.
		if inputType != nil && reflect.TypeOf(result) == inputType {
			return result, nil
		}

		if !matched {
			match = result
			matched = true
		}
	}

	if matched {
		return match, nil
	}

	if len(errs) == 0 {
		return nil, issues.CreateInvalidSchemaError(
			"no union options provided",
			input,
			parseCtx,
		)
	}
	return nil, issues.CreateInvalidUnionError(errs, input, parseCtx)
}

// MustParse is like Parse but panics on validation failure.
func (z *ZodUnion[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodUnion[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constrained, ok := convertToUnionConstraint[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"union constraint type",
			any(input),
			ctx[0],
		)
	}

	return engine.ParseComplexStrict[any, R](
		constrained,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractType,
		z.extractPtr,
		z.validate,
		ctx...,
	)
}

// MustStrictParse is like StrictParse but panics on validation failure.
func (z *ZodUnion[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodUnion[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a new schema that accepts undefined/missing values.
func (z *ZodUnion[T, R]) Optional() *ZodUnion[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodUnion[T, R]) Nilable() *ZodUnion[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema that accepts both undefined and nil values.
func (z *ZodUnion[T, R]) Nullish() *ZodUnion[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional returns a new schema that enforces non-nil values.
func (z *ZodUnion[T, R]) NonOptional() *ZodUnion[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodUnion[T, T]{internals: &ZodUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Options:          z.internals.Options,
	}}
}

// Default sets a value to use when input is nil, bypassing validation.
func (z *ZodUnion[T, R]) Default(v T) *ZodUnion[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a function to produce the default value when input is nil.
func (z *ZodUnion[T, R]) DefaultFunc(fn func() T) *ZodUnion[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodUnion[T, R]) Prefault(v T) *ZodUnion[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a function to produce the prefault value.
func (z *ZodUnion[T, R]) PrefaultFunc(fn func() T) *ZodUnion[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta attaches metadata to this schema.
func (z *ZodUnion[T, R]) Meta(meta core.GlobalMeta) *ZodUnion[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe sets a human-readable description for this schema.
func (z *ZodUnion[T, R]) Describe(desc string) *ZodUnion[T, R] {
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

// Options returns a copy of all union member schemas.
func (z *ZodUnion[T, R]) Options() []core.ZodSchema {
	out := make([]core.ZodSchema, len(z.internals.Options))
	copy(out, z.internals.Options)
	return out
}

// Transform creates a type-safe transformation pipeline.
func (z *ZodUnion[T, R]) Transform(
	fn func(T, *core.RefinementContext) (any, error),
) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractUnionValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Pipe chains this schema's output into another schema for further validation.
func (z *ZodUnion[T, R]) Pipe(
	target core.ZodType[any],
) *core.ZodPipe[R, any] {
	wrapper := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractUnionValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapper)
}

// Refine adds a custom validation check with type-safe access to the parsed value.
func (z *ZodUnion[T, R]) Refine(
	fn func(R) bool,
	params ...any,
) *ZodUnion[T, R] {
	wrapper := func(v any) bool {
		cv, ok := convertToUnionConstraint[T, R](v)
		if !ok {
			return false
		}
		return fn(cv)
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
	return z.withInternals(in)
}

// RefineAny adds a custom validation check operating on the raw value.
func (z *ZodUnion[T, R]) RefineAny(
	fn func(any) bool,
	params ...any,
) *ZodUnion[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
	return z.withInternals(in)
}

// And creates an intersection of this schema with another.
func (z *ZodUnion[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union of this schema with another.
func (z *ZodUnion[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

func (z *ZodUnion[T, R]) withPtrInternals(
	in *core.ZodTypeInternals,
) *ZodUnion[T, *T] {
	return &ZodUnion[T, *T]{internals: &ZodUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Options:          z.internals.Options,
	}}
}

func (z *ZodUnion[T, R]) withInternals(
	in *core.ZodTypeInternals,
) *ZodUnion[T, R] {
	return &ZodUnion[T, R]{internals: &ZodUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Options:          z.internals.Options,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodUnion[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodUnion[T, R]); ok {
		z.internals = src.internals
	}
}

// resolveCtx returns the first non-nil ParseContext or a new empty one.
func resolveCtx(ctx []*core.ParseContext) *core.ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return &core.ParseContext{}
}

// convertUnionToConstraint converts a value to constraint type R,
// handling pointer wrapping/unwrapping based on whether R is a pointer type.
func convertUnionToConstraint[T any, R any](value any) R {
	rType := reflect.TypeFor[R]()

	if value == nil {
		if rType.Kind() == reflect.Pointer {
			return any((*any)(nil)).(R)
		}
		var zero R
		return zero
	}

	if rType.Kind() == reflect.Pointer {
		if reflect.TypeOf(value).Kind() == reflect.Pointer {
			return any(value).(R) //nolint:unconvert // generic constraint conversion
		}
		v := value
		return any(&v).(R)
	}

	// R is non-pointer; dereference if value is a pointer.
	if reflect.TypeOf(value).Kind() == reflect.Pointer {
		rv := reflect.ValueOf(value)
		if !rv.IsNil() {
			return any(rv.Elem().Interface()).(R) //nolint:unconvert // generic constraint conversion
		}
		var zero R
		return zero
	}

	return any(value).(R) //nolint:unconvert // generic constraint conversion
}

// extractUnionValue extracts the base type T from constraint type R.
func extractUnionValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok && v != nil {
		return any(*v).(T) //nolint:unconvert // generic constraint conversion
	}
	return any(value).(T)
}

// convertToUnionConstraint attempts to convert a value to constraint type R.
func convertToUnionConstraint[T any, R any](value any) (R, bool) {
	var zero R

	if r, ok := any(value).(R); ok { //nolint:unconvert // generic constraint conversion
		return r, true
	}

	// Handle pointer conversion: wrap value as *any when R is *any.
	if _, ok := any(zero).(*any); ok {
		if value != nil {
			v := value
			return any(&v).(R), true
		}
		return any((*any)(nil)).(R), true
	}

	return zero, false
}

func newZodUnionFromDef[T any, R any](def *ZodUnionDef) *ZodUnion[T, R] {
	in := &ZodUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Options:          def.Options,
	}

	in.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodUnionDef{
			ZodTypeDef: *newDef,
			Options:    def.Options,
		}
		return any(newZodUnionFromDef[T, R](d)).(core.ZodType[any])
	}

	schema := &ZodUnion[T, R]{internals: in}

	if def.Error != nil {
		in.Error = def.Error
	}
	for _, chk := range def.Checks {
		in.AddCheck(chk)
	}

	return schema
}

// Union creates a union schema that accepts one of multiple types.
func Union(options []any, args ...any) *ZodUnion[any, any] {
	return UnionTyped[any, any](options, args...)
}

// UnionPtr creates a union schema with pointer constraint type.
func UnionPtr(options []any, args ...any) *ZodUnion[any, *any] {
	return UnionTyped[any, *any](options, args...)
}

// UnionOf creates a union schema from variadic arguments.
func UnionOf(schemas ...any) *ZodUnion[any, any] {
	return Union(schemas)
}

// UnionOfPtr creates a union schema from variadic arguments with pointer constraint.
func UnionOfPtr(schemas ...any) *ZodUnion[any, *any] {
	return UnionPtr(schemas)
}

// UnionTyped creates a typed union schema with generic constraints.
func UnionTyped[T any, R any](options []any, args ...any) *ZodUnion[T, R] {
	params := utils.NormalizeParams(utils.FirstParam(args...))

	schemas := make([]core.ZodSchema, len(options))
	for i, opt := range options {
		if opt == nil {
			continue
		}
		s, err := core.ConvertToZodSchema(opt)
		if err != nil {
			panic(fmt.Sprintf("Union option %d: %v", i, err))
		}
		schemas[i] = s
	}

	def := &ZodUnionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeUnion,
			Checks: []core.ZodCheck{},
		},
		Options: schemas,
	}
	if params != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, params)
	}

	schema := newZodUnionFromDef[T, R](def)

	// Trigger the validation pipeline even when no user-defined checks exist.
	schema.internals.AddCheck(checks.NewCustom[any](func(any) bool { return true }, core.SchemaParams{}))

	return schema
}
