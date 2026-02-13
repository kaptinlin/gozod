package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

var errUnexpectedSliceType = errors.New("unexpected result type from slice parser")

// ZodSliceDef defines the configuration for a slice schema.
type ZodSliceDef struct {
	core.ZodTypeDef
	Element any
}

// ZodSliceInternals contains the internal state for a slice schema.
type ZodSliceInternals[T any] struct {
	core.ZodTypeInternals
	Def     *ZodSliceDef
	Element any
}

// ZodSlice is a type-safe slice validation schema.
// T is the element type, R is the constraint type (value or pointer).
type ZodSlice[T any, R any] struct {
	internals *ZodSliceInternals[T]
}

// Internals returns the internal state for framework usage.
func (z *ZodSlice[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodSlice[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodSlice[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input and returns the parsed slice value.
func (z *ZodSlice[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	result, err := engine.ParseComplex[[]T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSlice,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
	if err != nil {
		return zero, err
	}

	switch v := result.(type) {
	case []T:
		return toSliceConstraint[T, R](v), nil
	case *[]T:
		return toSliceConstraint[T, R](v), nil
	case **[]T:
		if v != nil {
			return toSliceConstraint[T, R](*v), nil
		}
		return toSliceConstraint[T, R](nil), nil
	case nil:
		return toSliceConstraint[T, R](nil), nil
	default:
		if typed, ok := result.(R); ok {
			return typed, nil
		}
		return zero, fmt.Errorf("%w: %T", errUnexpectedSliceType, result)
	}
}

// MustParse validates input and panics on error.
func (z *ZodSlice[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodSlice[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParseComplexStrict[[]T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeSlice,
		z.extractForEngine,
		z.extractPtrForEngine,
		z.validateForEngine,
		ctx...,
	)
}

// MustStrictParse validates input with type safety and panics on error.
func (z *ZodSlice[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodSlice[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodSlice[T, R]) Optional() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
func (z *ZodSlice[T, R]) ExactOptional() *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodSlice[T, R]) Nilable() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodSlice[T, R]) Nullish() *ZodSlice[T, *[]T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional enforces non-nil value constraint.
func (z *ZodSlice[T, R]) NonOptional() *ZodSlice[T, []T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodSlice[T, []T]{
		internals: &ZodSliceInternals[T]{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Element:          z.internals.Element,
		},
	}
}

// Default sets a default value used when input is nil.
// Short-circuits validation and returns the default immediately.
func (z *ZodSlice[T, R]) Default(v []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a function that provides the default value.
func (z *ZodSlice[T, R]) DefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value used through the full parsing pipeline.
func (z *ZodSlice[T, R]) Prefault(v []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a function that provides the fallback value.
func (z *ZodSlice[T, R]) PrefaultFunc(fn func() []T) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this slice schema in the global registry.
func (z *ZodSlice[T, R]) Meta(meta core.GlobalMeta) *ZodSlice[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe sets a description for this schema in the global registry.
func (z *ZodSlice[T, R]) Describe(description string) *ZodSlice[T, R] {
	in := z.internals.Clone()

	meta, ok := core.GlobalRegistry.Get(z)
	if !ok {
		meta = core.GlobalMeta{}
	}
	meta.Description = description

	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, meta)

	return clone
}

// Min adds a minimum element count constraint.
func (z *ZodSlice[T, R]) Min(n int, params ...any) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.MinSize(n, params...))
	return z.withInternals(in)
}

// Max adds a maximum element count constraint.
func (z *ZodSlice[T, R]) Max(n int, params ...any) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.MaxSize(n, params...))
	return z.withInternals(in)
}

// Length adds an exact element count constraint.
func (z *ZodSlice[T, R]) Length(n int, params ...any) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.Size(n, params...))
	return z.withInternals(in)
}

// NonEmpty requires at least one element.
func (z *ZodSlice[T, R]) NonEmpty(params ...any) *ZodSlice[T, R] {
	return z.Min(1, params...)
}

// Element returns the element schema.
func (z *ZodSlice[T, R]) Element() core.ZodSchema {
	if schema, ok := z.internals.Element.(core.ZodSchema); ok {
		return schema
	}
	return nil
}

// Transform applies a transformation function to the parsed slice value.
func (z *ZodSlice[T, R]) Transform(
	fn func(R, *core.RefinementContext) (any, error),
) *core.ZodTransform[R, any] {
	wrapper := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapper)
}

// Overwrite transforms the slice value while preserving the original type.
func (z *ZodSlice[T, R]) Overwrite(fn func(R) R, params ...any) *ZodSlice[T, R] {
	wrap := func(input any) any {
		if converted, ok := toSliceType[T, R](input); ok {
			return fn(converted)
		}
		return input
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewZodCheckOverwrite(wrap, params...))
	return z.withInternals(in)
}

// Pipe passes the parsed value to a target schema.
func (z *ZodSlice[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapper := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapper)
}

// Refine adds type-safe custom validation.
func (z *ZodSlice[T, R]) Refine(fn func(R) bool, params ...any) *ZodSlice[T, R] {
	wrapper := func(v any) bool {
		return fn(toSliceConstraint[T, R](v))
	}
	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
	return z.withInternals(in)
}

// RefineAny adds custom validation without type conversion.
func (z *ZodSlice[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodSlice[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
	return z.withInternals(in)
}

// And creates an intersection with another schema.
func (z *ZodSlice[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union with another schema.
func (z *ZodSlice[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// Check adds a custom validation function that can report multiple issues.
func (z *ZodSlice[T, R]) Check(
	fn func(R, *core.ParsePayload),
	params ...any,
) *ZodSlice[T, R] {
	wrapper := func(p *core.ParsePayload) {
		if val, ok := p.Value().(R); ok {
			fn(val, p)
			return
		}

		// Pointer type adaptation: wrap value type in pointer for R = *T.
		rt := reflect.TypeFor[R]()
		if rt.Kind() == reflect.Pointer {
			elem := rt.Elem()
			rv := reflect.ValueOf(p.Value())
			if rv.IsValid() && rv.Type() == elem {
				ptr := reflect.New(elem)
				ptr.Elem().Set(rv)
				if v, ok := ptr.Interface().(R); ok {
					fn(v, p)
				}
			}
		}
	}

	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
	return z.withInternals(in)
}

// With is an alias for Check.
func (z *ZodSlice[T, R]) With(
	fn func(R, *core.ParsePayload),
	params ...any,
) *ZodSlice[T, R] {
	return z.Check(fn, params...)
}

// withInternals creates a new ZodSlice keeping the constraint type R.
func (z *ZodSlice[T, R]) withInternals(in *core.ZodTypeInternals) *ZodSlice[T, R] {
	return &ZodSlice[T, R]{
		internals: &ZodSliceInternals[T]{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Element:          z.internals.Element,
		},
	}
}

// withPtrInternals creates a new ZodSlice with pointer constraint *[]T.
func (z *ZodSlice[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodSlice[T, *[]T] {
	return &ZodSlice[T, *[]T]{
		internals: &ZodSliceInternals[T]{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Element:          z.internals.Element,
		},
	}
}

// CloneFrom copies configuration from another schema.
func (z *ZodSlice[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodSlice[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractForEngine extracts []T from input for engine.ParseComplex.
func (z *ZodSlice[T, R]) extractForEngine(value any) ([]T, bool) {
	if s, ok := value.([]T); ok {
		return s, true
	}

	if items, ok := value.([]any); ok {
		result := make([]T, len(items))
		for i, elem := range items {
			typed, ok := elem.(T)
			if !ok {
				return nil, false
			}
			result[i] = typed
		}
		return result, true
	}

	if value == nil {
		return nil, false
	}

	// Fall back to slicex for non-standard slice types.
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Slice {
		if converted, err := slicex.ToAny(value); err == nil && converted != nil {
			result := make([]T, len(converted))
			for i, elem := range converted {
				typed, ok := elem.(T)
				if !ok {
					return nil, false
				}
				result[i] = typed
			}
			return result, true
		}
	}

	return nil, false
}

// extractPtrForEngine extracts *[]T from input for engine.ParseComplex.
func (z *ZodSlice[T, R]) extractPtrForEngine(value any) (*[]T, bool) {
	if ptr, ok := value.(*[]T); ok {
		return ptr, true
	}
	if s, ok := value.([]T); ok {
		return &s, true
	}
	return nil, false
}

// validateForEngine validates slice elements using the element schema.
func (z *ZodSlice[T, R]) validateForEngine(
	value []T,
	chks []core.ZodCheck,
	ctx *core.ParseContext,
) ([]T, error) {
	payload := core.NewParsePayload(value)
	result := engine.RunChecksOnValue(value, chks, payload, ctx)

	var errs []core.ZodRawIssue
	if result.HasIssues() {
		errs = append(errs, result.Issues()...)
	}

	// Use the potentially transformed value for element validation.
	validated := value
	if result.Value() != nil {
		if converted, ok := result.Value().([]T); ok {
			validated = converted
		}
	}

	if schema, ok := z.internals.Element.(core.ZodSchema); ok && schema != nil {
		for i, elem := range validated {
			if err := validateElement(elem, schema); err != nil {
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					for _, issue := range zodErr.Issues {
						errs = append(errs, issues.ConvertZodIssueToRawWithProperties(issue, []any{i}))
					}
				} else {
					raw := issues.CreateIssue(core.Custom, err.Error(), nil, elem)
					raw.Path = []any{i}
					errs = append(errs, raw)
				}
			}
		}
	}

	if len(errs) > 0 {
		return nil, issues.CreateArrayValidationIssues(errs)
	}

	return validated, nil
}

// toSliceType converts any value to the slice constraint type R.
func toSliceType[T any, R any](v any) (R, bool) {
	var zero R

	if s, ok := v.([]T); ok {
		return toSliceConstraint[T, R](s), true
	}
	if ptr, ok := v.(*[]T); ok {
		return toSliceConstraint[T, R](ptr), true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		elemType := reflect.TypeFor[T]()
		s := make([]T, rv.Len())
		for i := range rv.Len() {
			elem := rv.Index(i).Interface()
			if reflect.TypeOf(elem).ConvertibleTo(elemType) {
				s[i] = reflect.ValueOf(elem).Convert(elemType).Interface().(T)
			} else {
				return zero, false
			}
		}
		return toSliceConstraint[T, R](s), true
	}

	return zero, false
}

// toSliceConstraint converts a value to the constraint type R.
func toSliceConstraint[T any, R any](value any) R {
	var zero R

	if value == nil {
		return zero
	}

	rt := reflect.TypeFor[R]()
	if rt != nil && rt.Kind() == reflect.Pointer {
		switch v := value.(type) {
		case []T:
			return any(&v).(R)
		case *[]T:
			return any(v).(R)
		}
	} else {
		switch v := value.(type) {
		case []T:
			return any(v).(R)
		case *[]T:
			if v != nil {
				return any(*v).(R)
			}
		}
	}

	if converted, ok := value.(R); ok {
		return converted
	}

	return zero
}

// newZodSliceFromDef constructs a new ZodSlice from a definition.
func newZodSliceFromDef[T any, R any](def *ZodSliceDef) *ZodSlice[T, R] {
	internals := &ZodSliceInternals[T]{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Element:          def.Element,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodSliceDef{
			ZodTypeDef: *newDef,
			Element:    def.Element,
		}
		return any(newZodSliceFromDef[T, R](d)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	for _, check := range def.Checks {
		internals.AddCheck(check)
	}

	return &ZodSlice[T, R]{internals: internals}
}

// Slice creates a slice schema with element validation.
func Slice[T any](elementSchema any, args ...any) *ZodSlice[T, []T] {
	return SliceTyped[T, []T](elementSchema, args...)
}

// SlicePtr creates a slice schema with pointer constraint.
func SlicePtr[T any](elementSchema any, args ...any) *ZodSlice[T, *[]T] {
	return SliceTyped[T, *[]T](elementSchema, args...)
}

// SliceTyped creates a slice schema with an explicit constraint type.
func SliceTyped[T any, R any](elementSchema any, args ...any) *ZodSlice[T, R] {
	param := utils.FirstParam(args...)
	sp := utils.NormalizeParams(param)

	def := &ZodSliceDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeSlice,
			Checks: []core.ZodCheck{},
		},
		Element: elementSchema,
	}

	if sp != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, sp)
	}

	schema := newZodSliceFromDef[T, R](def)

	// Add a minimal check to trigger element validation.
	if elementSchema != nil {
		alwaysTrue := func(any) bool { return true }
		schema.internals.AddCheck(checks.NewCustom[any](alwaysTrue, core.SchemaParams{}))
	}

	return schema
}
