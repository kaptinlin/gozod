package types

import (
	"errors"
	"fmt"
	"maps"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

var (
	ErrSchemaIsNil                = errors.New("schema is nil")
	ErrSchemaInternalsIsNil       = errors.New("schema internals is nil")
	ErrNoDiscriminatorValues      = errors.New("no discriminator values found for field")
	ErrOptionIsNil                = errors.New("option is nil")
	ErrDuplicateDiscriminator     = errors.New("duplicate discriminator value")
	ErrFailedToBuildDiscriminator = errors.New("failed to build discriminator map")
	ErrNoValidDiscriminators      = errors.New("no valid discriminator values found for field")
)

// ZodDiscriminatedUnionDef holds the configuration for discriminated union validation.
type ZodDiscriminatedUnionDef struct {
	core.ZodTypeDef
	Discriminator string
	Options       []core.ZodSchema
}

// ZodDiscriminatedUnionInternals contains the internal state of a discriminated union validator.
type ZodDiscriminatedUnionInternals struct {
	core.ZodTypeInternals
	Def           *ZodDiscriminatedUnionDef
	Discriminator string
	Options       []core.ZodSchema
	DiscMap       map[any]core.ZodSchema
}

// ZodDiscriminatedUnion is a type-safe discriminated union validation schema.
type ZodDiscriminatedUnion[T any, R any] struct {
	internals *ZodDiscriminatedUnionInternals
}

// Internals returns the internal state of the schema.
func (z *ZodDiscriminatedUnion[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodDiscriminatedUnion[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodDiscriminatedUnion[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input and returns a value matching the constraint type R.
func (z *ZodDiscriminatedUnion[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R
	pctx := resolveCtx(ctx)

	if ce, ok := z.internals.Bag["construction_error"]; ok {
		if msg, ok := ce.(string); ok {
			return zero, issues.CreateInvalidSchemaError(msg, input, pctx)
		}
	}

	if input == nil && (z.internals.Nilable || z.internals.Optional) {
		return zero, nil
	}

	if input == nil {
		if z.internals.DefaultFunc != nil {
			input = z.internals.DefaultFunc()
		} else if z.internals.DefaultValue != nil {
			input = z.internals.DefaultValue
		}
	}

	m, ok := input.(map[string]any)
	if !ok {
		if z.internals.PrefaultFunc != nil {
			return z.Parse(z.internals.PrefaultFunc(), pctx)
		}
		if z.internals.PrefaultValue != nil {
			return z.Parse(z.internals.PrefaultValue, pctx)
		}
		return zero, issues.CreateInvalidTypeError(core.ZodTypeObject, input, pctx)
	}

	dv, exists := m[z.internals.Discriminator]
	if !exists {
		return zero, issues.CreateMissingRequiredError(z.internals.Discriminator, "discriminator field", input, pctx)
	}

	r, err := z.parseVariant(m, dv, pctx)
	if err != nil {
		return zero, err
	}

	if len(z.internals.Checks) > 0 {
		r, err = engine.ApplyChecks[any](r, z.internals.Checks, pctx)
		if err != nil {
			return zero, err
		}
	}

	return convertToDiscriminatedUnionConstraintType[T, R](r), nil
}

// parseVariant dispatches to the matching schema or falls back to trying all options.
func (z *ZodDiscriminatedUnion[T, R]) parseVariant(m map[string]any, dv any, pctx *core.ParseContext) (any, error) {
	if target, ok := z.internals.DiscMap[dv]; ok {
		return target.ParseAny(m, pctx)
	}

	errs := make([]error, 0, len(z.internals.Options))
	for _, opt := range z.internals.Options {
		if opt == nil {
			continue
		}
		r, e := opt.ParseAny(m, pctx)
		if e == nil {
			return r, nil
		}
		errs = append(errs, e)
	}
	return nil, issues.CreateInvalidUnionError(errs, m, pctx)
}

// MustParse panics on validation failure.
func (z *ZodDiscriminatedUnion[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	r, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// StrictParse validates input with compile-time type safety.
func (z *ZodDiscriminatedUnion[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	cv, ok := convertToDiscriminatedUnionConstraintValue[T, R](input)
	if !ok {
		var zero R
		return zero, issues.CreateTypeConversionError(fmt.Sprintf("%T", input), "discriminated union constraint type", any(input), resolveCtx(ctx))
	}
	return z.Parse(cv, ctx...)
}

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodDiscriminatedUnion[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	r, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return r
}

// ParseAny validates input and returns an untyped result.
func (z *ZodDiscriminatedUnion[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodDiscriminatedUnion[T, R]) Optional() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodDiscriminatedUnion[T, R]) Nilable() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodDiscriminatedUnion[T, R]) Nullish() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits).
func (z *ZodDiscriminatedUnion[T, R]) Default(v T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits).
func (z *ZodDiscriminatedUnion[T, R]) DefaultFunc(fn func() T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodDiscriminatedUnion[T, R]) Prefault(v T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodDiscriminatedUnion[T, R]) PrefaultFunc(fn func() T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata in the global registry.
func (z *ZodDiscriminatedUnion[T, R]) Meta(meta core.GlobalMeta) *ZodDiscriminatedUnion[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodDiscriminatedUnion[T, R]) Describe(desc string) *ZodDiscriminatedUnion[T, R] {
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

// Discriminator returns the discriminator field name.
func (z *ZodDiscriminatedUnion[T, R]) Discriminator() string {
	return z.internals.Discriminator
}

// Options returns a copy of all union member schemas.
func (z *ZodDiscriminatedUnion[T, R]) Options() []core.ZodSchema {
	r := make([]core.ZodSchema, len(z.internals.Options))
	copy(r, z.internals.Options)
	return r
}

// DiscriminatorMap returns a copy of the discriminator-to-schema mapping.
func (z *ZodDiscriminatedUnion[T, R]) DiscriminatorMap() map[any]core.ZodSchema {
	r := make(map[any]core.ZodSchema, len(z.internals.DiscMap))
	maps.Copy(r, z.internals.DiscMap)
	return r
}

// Transform applies a transformation function to the parsed value.
func (z *ZodDiscriminatedUnion[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractDiscriminatedUnionValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Pipe creates a validation pipeline with another schema.
func (z *ZodDiscriminatedUnion[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapper := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractDiscriminatedUnionValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapper)
}

// Refine applies a custom validation function matching the schema's output type R.
func (z *ZodDiscriminatedUnion[T, R]) Refine(fn func(R) bool, params ...any) *ZodDiscriminatedUnion[T, R] {
	wrapper := func(v any) bool {
		cv, ok := convertToDiscriminatedUnionConstraintValue[T, R](v)
		if !ok {
			return false
		}
		return fn(cv)
	}
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
}

// RefineAny applies a custom validation function that receives the raw value.
func (z *ZodDiscriminatedUnion[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodDiscriminatedUnion[T, R] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// withCheck clones internals, adds a check, and returns a new schema.
func (z *ZodDiscriminatedUnion[T, R]) withCheck(c core.ZodCheck) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.AddCheck(c)
	return z.withInternals(in)
}

// withPtrInternals creates a new pointer-constraint schema from cloned internals.
func (z *ZodDiscriminatedUnion[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodDiscriminatedUnion[T, *T] {
	return &ZodDiscriminatedUnion[T, *T]{internals: &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Discriminator:    z.internals.Discriminator,
		Options:          z.internals.Options,
		DiscMap:          z.internals.DiscMap,
	}}
}

// withInternals creates a new schema preserving generic type parameters.
func (z *ZodDiscriminatedUnion[T, R]) withInternals(in *core.ZodTypeInternals) *ZodDiscriminatedUnion[T, R] {
	return &ZodDiscriminatedUnion[T, R]{internals: &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Discriminator:    z.internals.Discriminator,
		Options:          z.internals.Options,
		DiscMap:          z.internals.DiscMap,
	}}
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodDiscriminatedUnion[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodDiscriminatedUnion[T, R]); ok {
		z.internals = src.internals
	}
}

// convertToDiscriminatedUnionConstraintType converts a value to constraint type R.
func convertToDiscriminatedUnionConstraintType[T any, R any](v any) R {
	var zero R
	if _, ok := any(zero).(*any); ok {
		if v != nil {
			return any(new(v)).(R)
		}
		return any((*any)(nil)).(R)
	}
	return any(v).(R) //nolint:unconvert // Required for generic type constraint conversion
}

// extractDiscriminatedUnionValue extracts base type T from constraint type R.
func extractDiscriminatedUnionValue[T any, R any](val R) T {
	if v, ok := any(val).(*any); ok {
		if v != nil {
			return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
		}
		var zero T
		return zero
	}
	return any(val).(T)
}

// convertToDiscriminatedUnionConstraintValue converts any value to constraint type R.
func convertToDiscriminatedUnionConstraintValue[T any, R any](v any) (R, bool) {
	var zero R
	if r, ok := any(v).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}
	if _, ok := any(zero).(*any); ok {
		if v != nil {
			return any(new(v)).(R), true
		}
		return any((*any)(nil)).(R), true
	}
	return zero, false
}

// discValues extracts discriminator values from a schema.
func discValues(schema core.ZodSchema, field string) ([]any, error) {
	if schema == nil {
		return nil, ErrSchemaIsNil
	}

	in := schema.Internals()
	if in == nil {
		return nil, ErrSchemaInternalsIsNil
	}

	if vals := extractDiscVals(in); len(vals) > 0 {
		return vals, nil
	}

	return discValuesFromAny(schema, field)
}

// discValuesFromAny extracts discriminator values via type assertion on Shape methods.
func discValuesFromAny(schema any, field string) ([]any, error) {
	if schema == nil {
		return nil, ErrSchemaIsNil
	}

	if s, ok := schema.(interface {
		Internals() *core.ZodTypeInternals
	}); ok {
		if in := s.Internals(); in != nil {
			if vals := extractDiscVals(in); len(vals) > 0 {
				return vals, nil
			}
		}
	}

	if s, ok := schema.(interface {
		Shape() core.ObjectSchema
	}); ok {
		if fs, exists := s.Shape()[field]; exists {
			if fi := fs.Internals(); fi != nil {
				if vals := extractDiscVals(fi); len(vals) > 0 {
					return vals, nil
				}
			}
		}
	}

	if s, ok := schema.(interface {
		Shape() core.StructSchema
	}); ok {
		if fs, exists := s.Shape()[field]; exists {
			if fi := fs.Internals(); fi != nil {
				if vals := extractDiscVals(fi); len(vals) > 0 {
					return vals, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("%w '%s' in schema", ErrNoDiscriminatorValues, field)
}

// extractDiscVals extracts discriminator values from schema internals.
func extractDiscVals(in *core.ZodTypeInternals) []any {
	if in == nil {
		return nil
	}

	if len(in.Values) > 0 {
		vals := make([]any, 0, len(in.Values))
		for v := range in.Values {
			vals = append(vals, v)
		}
		return vals
	}

	if in.Bag != nil {
		if bv, exists := in.Bag["values"]; exists {
			if vm, ok := bv.(map[any]struct{}); ok {
				vals := make([]any, 0, len(vm))
				for v := range vm {
					vals = append(vals, v)
				}
				return vals
			}
		}
	}

	// Only interested in Literal and Enum types here
	//nolint:exhaustive
	switch in.Type {
	case core.ZodTypeLiteral:
		if lv, exists := in.Bag["literal"]; exists {
			return []any{lv}
		}
	case core.ZodTypeEnum:
		if ev, exists := in.Bag["enum"]; exists {
			if em, ok := ev.(map[any]struct{}); ok {
				enumVals := make([]any, 0, len(em))
				for v := range em {
					enumVals = append(enumVals, v)
				}
				return enumVals
			}
		}
	}
	return nil
}

// buildDiscriminatorMap builds the discriminator-to-schema mapping.
func buildDiscriminatorMap(disc string, options []core.ZodSchema) (map[any]core.ZodSchema, error) {
	dm := make(map[any]core.ZodSchema)
	var errs []error

	for i, opt := range options {
		if opt == nil {
			errs = append(errs, fmt.Errorf("%w: option %d", ErrOptionIsNil, i))
			continue
		}
		vals, err := discValues(opt, disc)
		if err != nil {
			errs = append(errs, fmt.Errorf("option %d: %w", i, err))
			continue
		}
		for _, v := range vals {
			if _, exists := dm[v]; exists {
				return nil, fmt.Errorf("%w: %v", ErrDuplicateDiscriminator, v)
			}
			dm[v] = opt
		}
	}

	if len(dm) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("%w: %w", ErrFailedToBuildDiscriminator, errors.Join(errs...))
	}
	if len(dm) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoValidDiscriminators, disc)
	}
	return dm, nil
}

// newZodDiscriminatedUnionFromDef constructs a ZodDiscriminatedUnion from a definition.
// Construction errors are deferred to parse-time for graceful handling.
func newZodDiscriminatedUnionFromDef[T any, R any](def *ZodDiscriminatedUnionDef) *ZodDiscriminatedUnion[T, R] {
	dm, err := buildDiscriminatorMap(def.Discriminator, def.Options)

	in := &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Discriminator:    def.Discriminator,
		Options:          def.Options,
		DiscMap:          dm,
	}

	if err != nil {
		in.Bag["construction_error"] = fmt.Sprintf("DiscriminatedUnion construction error: %v", err)
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		cd := &ZodDiscriminatedUnionDef{
			ZodTypeDef:    *d,
			Discriminator: def.Discriminator,
			Options:       def.Options,
		}
		return any(newZodDiscriminatedUnionFromDef[T, R](cd)).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}
	for _, c := range def.Checks {
		in.AddCheck(c)
	}

	return &ZodDiscriminatedUnion[T, R]{internals: in}
}

// DiscriminatedUnion creates a discriminated union schema with value constraint.
func DiscriminatedUnion(disc string, options []any, args ...any) *ZodDiscriminatedUnion[any, any] {
	return DiscriminatedUnionTyped[any, any](disc, options, args...)
}

// DiscriminatedUnionPtr creates a discriminated union schema with pointer constraint.
func DiscriminatedUnionPtr(disc string, options []any, args ...any) *ZodDiscriminatedUnion[any, *any] {
	return DiscriminatedUnionTyped[any, *any](disc, options, args...)
}

// DiscriminatedUnionTyped creates a typed discriminated union schema.
func DiscriminatedUnionTyped[T any, R any](disc string, options []any, args ...any) *ZodDiscriminatedUnion[T, R] {
	sp := utils.NormalizeParams(utils.FirstParam(args...))

	opts := make([]core.ZodSchema, len(options))
	for i, opt := range options {
		s, err := core.ConvertToZodSchema(opt)
		if err != nil {
			panic(fmt.Sprintf("DiscriminatedUnion option %d: %v", i, err))
		}
		opts[i] = s
	}

	def := &ZodDiscriminatedUnionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeDiscriminated,
			Checks: []core.ZodCheck{},
		},
		Discriminator: disc,
		Options:       opts,
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}
	return newZodDiscriminatedUnionFromDef[T, R](def)
}
