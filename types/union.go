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

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodUnionDef defines the schema definition for union validation.
type ZodUnionDef struct {
	core.ZodTypeDef
	Options []core.ZodSchema
}

// ZodUnionInternals contains the internal state for union schema.
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

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state for framework usage.
func (z *ZodUnion[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodUnion[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodUnion[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input against all union member schemas, returning the first match.
func (z *ZodUnion[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	result, err := engine.ParseComplex[any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractUnionType,
		z.extractUnionPtr,
		z.validateUnionValue,
		parseCtx,
	)
	if err != nil {
		var zero R
		return zero, err
	}
	return convertToUnionConstraintType[T, R](result), nil
}

func (z *ZodUnion[T, R]) extractUnionType(input any) (any, bool) {
	return input, true
}

func (z *ZodUnion[T, R]) extractUnionPtr(input any) (*any, bool) {
	if input == nil {
		return nil, true
	}
	return &input, true
}

// validateUnionValue tries each member schema and returns the first successful match.
// It prefers the schema whose result type matches the original input type.
func (z *ZodUnion[T, R]) validateUnionValue(input any, checks []core.ZodCheck, parseCtx *core.ParseContext) (any, error) {
	var (
		firstSuccess any
		successFound bool
		allErrors    []error
	)

	inputType := reflect.TypeOf(input)

	for i, option := range z.internals.Options {
		if option == nil {
			continue
		}

		result, err := option.ParseAny(input, parseCtx)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
			continue
		}

		// Apply custom checks on the union itself.
		if len(checks) > 0 {
			result, err = engine.ApplyChecks[any](result, checks, parseCtx)
			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
				continue
			}
		}

		// Prefer the schema whose result type matches the original input type.
		if inputType != nil && reflect.TypeOf(result) == inputType {
			return result, nil
		}

		if !successFound {
			firstSuccess = result
			successFound = true
		}
	}

	if successFound {
		return firstSuccess, nil
	}

	if len(allErrors) == 0 {
		return nil, issues.CreateInvalidSchemaError("no union options provided", input, parseCtx)
	}
	return nil, issues.CreateInvalidUnionError(allErrors, input, parseCtx)
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
// The input must exactly match the schema's base type T.
func (z *ZodUnion[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput, ok := convertToUnionConstraintValue[T, R](input)
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
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractUnionType,
		z.extractUnionPtr,
		z.validateUnionValue,
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

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodUnion[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a new schema that accepts undefined/missing values.
// The result type changes to *T to represent the optional nature.
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
	return &ZodUnion[T, T]{
		internals: &ZodUnionInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Options:          z.internals.Options,
		},
	}
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
	in.SetDefaultFunc(func() any {
		return fn()
	})
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
func (z *ZodUnion[T, R]) Describe(description string) *ZodUnion[T, R] {
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
// TYPE-SPECIFIC METHODS
// =============================================================================

// Options returns a copy of all union member schemas.
func (z *ZodUnion[T, R]) Options() []core.ZodSchema {
	result := make([]core.ZodSchema, len(z.internals.Options))
	copy(result, z.internals.Options)
	return result
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation pipeline.
func (z *ZodUnion[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractUnionValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe chains this schema's output into another schema for further validation.
func (z *ZodUnion[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractUnionValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation check with type-safe access to the parsed value.
func (z *ZodUnion[T, R]) Refine(fn func(R) bool, params ...any) *ZodUnion[T, R] {
	wrapper := func(v any) bool {
		if cv, ok := convertToUnionConstraintValue[T, R](v); ok {
			return fn(cv)
		}
		return false
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny adds a custom validation check operating on the raw value.
func (z *ZodUnion[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodUnion[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection of this schema with another.
func (z *ZodUnion[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union of this schema with another, enabling chaining.
func (z *ZodUnion[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodUnion[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodUnion[T, *T] {
	return &ZodUnion[T, *T]{internals: &ZodUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Options:          z.internals.Options,
	}}
}

func (z *ZodUnion[T, R]) withInternals(in *core.ZodTypeInternals) *ZodUnion[T, R] {
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

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToUnionConstraintType converts a value to constraint type R,
// handling pointer wrapping/unwrapping based on whether R is a pointer type.
func convertToUnionConstraintType[T any, R any](value any) R {
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
			return any(value).(R) //nolint:unconvert // Required for generic type constraint conversion
		}
		valueCopy := value
		return any(&valueCopy).(R)
	}

	// R is a non-pointer type; dereference if value is a pointer.
	if reflect.TypeOf(value).Kind() == reflect.Pointer {
		rv := reflect.ValueOf(value)
		if !rv.IsNil() {
			return any(rv.Elem().Interface()).(R) //nolint:unconvert // Required for generic type constraint conversion
		}
		var zero R
		return zero
	}

	return any(value).(R) //nolint:unconvert // Required for generic type constraint conversion
}

// extractUnionValue unwraps a constraint type R back to base type T.
func extractUnionValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok && v != nil {
		return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
	}
	return any(value).(T)
}

// convertToUnionConstraintValue attempts to convert a value to constraint type R.
func convertToUnionConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	// Handle pointer conversion: wrap value as *any when R is *any.
	if _, ok := any(zero).(*any); ok {
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R), true
		}
		return any((*any)(nil)).(R), true
	}

	return zero, false
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

func newZodUnionFromDef[T any, R any](def *ZodUnionDef) *ZodUnion[T, R] {
	internals := &ZodUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Options:          def.Options,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		unionDef := &ZodUnionDef{
			ZodTypeDef: *newDef,
			Options:    def.Options,
		}
		return any(newZodUnionFromDef[T, R](unionDef)).(core.ZodType[any])
	}

	schema := &ZodUnion[T, R]{internals: internals}

	if def.Error != nil {
		internals.Error = def.Error
	}
	for _, check := range def.Checks {
		internals.AddCheck(check)
	}

	return schema
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

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
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	wrappedOptions := make([]core.ZodSchema, len(options))
	for i, option := range options {
		if option == nil {
			continue
		}
		zodSchema, err := core.ConvertToZodSchema(option)
		if err != nil {
			panic(fmt.Sprintf("Union option %d: %v", i, err))
		}
		wrappedOptions[i] = zodSchema
	}

	def := &ZodUnionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeUnion,
			Checks: []core.ZodCheck{},
		},
		Options: wrappedOptions,
	}
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	schema := newZodUnionFromDef[T, R](def)

	// Trigger the validation pipeline even when no user-defined checks exist.
	schema.internals.AddCheck(checks.NewCustom[any](func(any) bool { return true }, core.SchemaParams{}))

	return schema
}
