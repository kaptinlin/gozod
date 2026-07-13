package types

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"

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

// DiscriminatorError describes why a discriminated union option is invalid.
type DiscriminatorError struct {
	Option int
	Field  string
	Value  any
	Err    error
}

func (e *DiscriminatorError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Option >= 0 {
		if e.Value != nil {
			return fmt.Sprintf("discriminated union option %d field %q value %v: %v", e.Option, e.Field, e.Value, e.Err)
		}
		return fmt.Sprintf("discriminated union option %d field %q: %v", e.Option, e.Field, e.Err)
	}
	return fmt.Sprintf("discriminated union field %q: %v", e.Field, e.Err)
}

// Unwrap returns the construction error category.
func (e *DiscriminatorError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

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
	r, err := engine.ParseComplex[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeObject,
		z.extractType,
		z.extractPtr,
		z.validate,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}
	return convertToDiscriminatedUnionConstraintType[T, R](r), nil
}

func (z *ZodDiscriminatedUnion[T, R]) extractType(input any) (any, bool) {
	return input, true
}

func (z *ZodDiscriminatedUnion[T, R]) extractPtr(input any) (*any, bool) {
	ptr, ok := input.(*any)
	return ptr, ok
}

func (z *ZodDiscriminatedUnion[T, R]) validate(input any, chks []core.ZodCheck, pctx *core.ParseContext) (any, error) {
	m, ok := input.(map[string]any)
	if !ok {
		return nil, issues.CreateInvalidTypeError(core.ZodTypeObject, input, pctx)
	}

	dv, exists := m[z.internals.Discriminator]
	if !exists {
		return nil, issues.CreateMissingRequiredError(z.internals.Discriminator, "discriminator field", input, pctx)
	}

	r, err := z.parseVariant(m, dv, pctx)
	if err != nil {
		return nil, err
	}
	if len(chks) > 0 {
		return engine.ApplyChecks[any](r, chks, pctx)
	}
	return r, nil
}

// parseVariant dispatches to the schema indexed by the discriminator value.
func (z *ZodDiscriminatedUnion[T, R]) parseVariant(m map[string]any, dv any, pctx *core.ParseContext) (any, error) {
	if target, ok := z.internals.DiscMap[dv]; ok {
		return target.ParseAny(m, pctx)
	}
	values := slices.Collect(maps.Keys(z.internals.DiscMap))
	slices.SortFunc(values, compareDiscriminatorValues)
	return nil, issues.CreateInvalidDiscriminatorError(z.internals.Discriminator, values, m, pctx)
}

func compareDiscriminatorValues(a, b any) int {
	if result := cmp.Compare(fmt.Sprintf("%T", a), fmt.Sprintf("%T", b)); result != 0 {
		return result
	}
	return cmp.Compare(fmt.Sprint(a), fmt.Sprint(b))
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
	if !ok && any(input) == nil {
		var zero R
		cv, ok = zero, true
	}
	if !ok {
		var zero R
		return zero, issues.CreateTypeConversionError(fmt.Sprintf("%T", input), "discriminated union constraint type", any(input), resolveCtx(ctx))
	}
	return engine.ParseComplexStrict[any, R](
		cv,
		&z.internals.ZodTypeInternals,
		core.ZodTypeObject,
		z.extractType,
		z.extractPtr,
		z.validate,
		ctx...,
	)
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

// Meta returns a schema with merged metadata.
func (z *ZodDiscriminatedUnion[T, R]) Meta(meta core.GlobalMeta) *ZodDiscriminatedUnion[T, R] {
	clone := z.withInternals(z.internals.Clone())
	core.ApplySchemaMeta(z, clone, meta)
	return clone
}

// Describe returns a schema with the description.
func (z *ZodDiscriminatedUnion[T, R]) Describe(desc string) *ZodDiscriminatedUnion[T, R] {
	return z.Meta(core.GlobalMeta{Description: desc})
}

// Discriminator returns the discriminator field name.
func (z *ZodDiscriminatedUnion[T, R]) Discriminator() string {
	return z.internals.Discriminator
}

// Options returns a copy of all union member schemas.
func (z *ZodDiscriminatedUnion[T, R]) Options() []core.ZodSchema {
	if len(z.internals.Options) == 0 {
		return []core.ZodSchema{}
	}
	return slices.Clone(z.internals.Options)
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
	clone := &ZodDiscriminatedUnion[T, *T]{internals: &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Discriminator:    z.internals.Discriminator,
		Options:          z.internals.Options,
		DiscMap:          z.internals.DiscMap,
	}}
	finalizeClone(clone)
	return clone
}

// withInternals creates a new schema preserving generic type parameters.
func (z *ZodDiscriminatedUnion[T, R]) withInternals(in *core.ZodTypeInternals) *ZodDiscriminatedUnion[T, R] {
	clone := &ZodDiscriminatedUnion[T, R]{internals: &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Discriminator:    z.internals.Discriminator,
		Options:          z.internals.Options,
		DiscMap:          z.internals.DiscMap,
	}}
	finalizeClone(clone)
	return clone
}

// CloneFrom copies configuration from another schema of the same type.
func (z *ZodDiscriminatedUnion[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodDiscriminatedUnion[T, R]); ok && src != nil {
		cloneWithPreservedChecks(src, z, func() {
			z.internals = &ZodDiscriminatedUnionInternals{
				ZodTypeInternals: *src.internals.Clone(),
				Def:              src.internals.Def,
				Discriminator:    src.internals.Discriminator,
				Options:          src.internals.Options,
				DiscMap:          src.internals.DiscMap,
			}
		})
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
		return slices.Collect(maps.Keys(in.Values))
	}
	return nil
}

// buildDiscriminatorMap builds the discriminator-to-schema mapping.
func buildDiscriminatorMap(disc string, options []core.ZodSchema) (map[any]core.ZodSchema, error) {
	dm := make(map[any]core.ZodSchema)
	var errs []error

	for i, opt := range options {
		if opt == nil {
			errs = append(errs, &DiscriminatorError{Option: i, Field: disc, Err: ErrOptionIsNil})
			continue
		}
		vals, err := discValues(opt, disc)
		if err != nil {
			errs = append(errs, &DiscriminatorError{Option: i, Field: disc, Err: err})
			continue
		}
		for _, v := range vals {
			if _, exists := dm[v]; exists {
				return nil, &DiscriminatorError{
					Option: i,
					Field:  disc,
					Value:  v,
					Err:    ErrDuplicateDiscriminator,
				}
			}
			dm[v] = opt
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("%w: %w", ErrFailedToBuildDiscriminator, errors.Join(errs...))
	}
	if len(dm) == 0 {
		return nil, &DiscriminatorError{Option: -1, Field: disc, Err: ErrNoValidDiscriminators}
	}
	return dm, nil
}

// newZodDiscriminatedUnionFromDef constructs a ZodDiscriminatedUnion from a definition.
func newZodDiscriminatedUnionFromDef[T any, R any](def *ZodDiscriminatedUnionDef) (*ZodDiscriminatedUnion[T, R], error) {
	dm, err := buildDiscriminatorMap(def.Discriminator, def.Options)
	if err != nil {
		return nil, err
	}

	in := &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Discriminator:    def.Discriminator,
		Options:          def.Options,
		DiscMap:          dm,
	}

	in.Constructor = func(d *core.ZodTypeDef) core.ZodType[any] {
		cd := &ZodDiscriminatedUnionDef{
			ZodTypeDef:    *d,
			Discriminator: def.Discriminator,
			Options:       def.Options,
		}
		clone, cloneErr := newZodDiscriminatedUnionFromDef[T, R](cd)
		if cloneErr != nil {
			panic(fmt.Sprintf("rebuild discriminated union: %v", cloneErr))
		}
		return any(clone).(core.ZodType[any])
	}

	if def.Error != nil {
		in.Error = def.Error
	}
	for _, c := range def.Checks {
		in.AddCheck(c)
	}

	return &ZodDiscriminatedUnion[T, R]{internals: in}, nil
}

// DiscriminatedUnion creates a discriminated union schema with value constraint.
// It returns an error when an option cannot provide unique discriminator values.
func DiscriminatedUnion(disc string, options []core.ZodSchema, args ...any) (*ZodDiscriminatedUnion[any, any], error) {
	return DiscriminatedUnionTyped[any, any](disc, options, args...)
}

// DiscriminatedUnionPtr creates a discriminated union schema with pointer constraint.
// It returns an error when an option cannot provide unique discriminator values.
func DiscriminatedUnionPtr(disc string, options []core.ZodSchema, args ...any) (*ZodDiscriminatedUnion[any, *any], error) {
	return DiscriminatedUnionTyped[any, *any](disc, options, args...)
}

// DiscriminatedUnionTyped creates a typed discriminated union schema.
// It returns an error when an option cannot provide unique discriminator values.
func DiscriminatedUnionTyped[T any, R any](disc string, options []core.ZodSchema, args ...any) (*ZodDiscriminatedUnion[T, R], error) {
	sp := utils.NormalizeParams(utils.FirstParam(args...))

	def := &ZodDiscriminatedUnionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeDiscriminated,
			Checks: []core.ZodCheck{},
		},
		Discriminator: disc,
		Options:       slices.Clone(options),
	}
	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}
	return newZodDiscriminatedUnionFromDef[T, R](def)
}

// MustDiscriminatedUnion creates a discriminated union or panics if its definition is invalid.
func MustDiscriminatedUnion(disc string, options []core.ZodSchema, args ...any) *ZodDiscriminatedUnion[any, any] {
	return MustDiscriminatedUnionTyped[any, any](disc, options, args...)
}

// MustDiscriminatedUnionPtr creates a pointer-constrained discriminated union or panics if invalid.
func MustDiscriminatedUnionPtr(disc string, options []core.ZodSchema, args ...any) *ZodDiscriminatedUnion[any, *any] {
	return MustDiscriminatedUnionTyped[any, *any](disc, options, args...)
}

// MustDiscriminatedUnionTyped creates a typed discriminated union or panics if invalid.
func MustDiscriminatedUnionTyped[T any, R any](disc string, options []core.ZodSchema, args ...any) *ZodDiscriminatedUnion[T, R] {
	schema, err := DiscriminatedUnionTyped[T, R](disc, options, args...)
	if err != nil {
		panic(err)
	}
	return schema
}
