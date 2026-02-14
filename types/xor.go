package types

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodXorDef defines the schema definition for exclusive union validation.
type ZodXorDef struct {
	core.ZodTypeDef
	Options []core.ZodSchema
}

// ZodXorInternals contains the internal state for exclusive union schema.
type ZodXorInternals struct {
	core.ZodTypeInternals
	Def     *ZodXorDef
	Options []core.ZodSchema
}

// ZodXor validates that input matches exactly one member schema.
// T is the base type, R is the constraint type (T or *T for optional/nilable).
type ZodXor[T any, R any] struct {
	internals *ZodXorInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state for framework usage.
func (z *ZodXor[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodXor[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodXor[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input ensuring exactly one option matches.
func (z *ZodXor[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
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

func (z *ZodXor[T, R]) extractType(input any) (any, bool) {
	return input, true
}

func (z *ZodXor[T, R]) extractPtr(input any) (*any, bool) {
	if input == nil {
		return nil, true
	}
	return &input, true
}

// validate checks that exactly one option matches (Zod v4 exclusive union semantics).
// This is an unexported helper method for internal validation logic.
func (z *ZodXor[T, R]) validate(input any, chks []core.ZodCheck, parseCtx *core.ParseContext) (any, error) {
	successes := make([]any, 0, len(z.internals.Options))
	var allErrors []error

	for i, option := range z.internals.Options {
		if option == nil {
			continue
		}

		result, err := option.ParseAny(input, parseCtx)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
			continue
		}

		// Apply custom checks on the xor schema itself.
		if len(chks) > 0 {
			result, err = engine.ApplyChecks[any](result, chks, parseCtx)
			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
				continue
			}
		}

		successes = append(successes, result)
	}

	switch len(successes) {
	case 1:
		return successes[0], nil
	case 0:
		if len(allErrors) == 0 {
			return nil, issues.CreateInvalidSchemaError("no xor options provided", input, parseCtx)
		}
		return nil, issues.CreateInvalidUnionError(allErrors, input, parseCtx)
	default:
		return nil, issues.CreateInvalidXorError(len(successes), input, parseCtx)
	}
}

// MustParse is like Parse but panics on validation failure.
func (z *ZodXor[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
// The input must exactly match the schema's base type T.
func (z *ZodXor[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput, ok := convertToUnionConstraint[T, R](input)
	if !ok {
		var zero R
		parseCtx := resolveCtx(ctx)
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"xor constraint type",
			any(input),
			parseCtx,
		)
	}

	return engine.ParseComplexStrict[any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractType,
		z.extractPtr,
		z.validate,
		ctx...,
	)
}

// MustStrictParse is like StrictParse but panics on validation failure.
func (z *ZodXor[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result for runtime scenarios.
func (z *ZodXor[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a new schema that accepts undefined/missing values.
func (z *ZodXor[T, R]) Optional() *ZodXor[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a new schema that accepts nil values.
func (z *ZodXor[T, R]) Nilable() *ZodXor[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema that accepts both undefined and nil values.
func (z *ZodXor[T, R]) Nullish() *ZodXor[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional returns a new schema that enforces non-nil values.
func (z *ZodXor[T, R]) NonOptional() *ZodXor[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)
	return &ZodXor[T, T]{
		internals: &ZodXorInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Options:          z.internals.Options,
		},
	}
}

// Default sets a value to use when input is nil, bypassing validation.
func (z *ZodXor[T, R]) Default(v T) *ZodXor[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a function to produce the default value when input is nil.
func (z *ZodXor[T, R]) DefaultFunc(fn func() T) *ZodXor[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodXor[T, R]) Prefault(v T) *ZodXor[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a function to produce the prefault value.
func (z *ZodXor[T, R]) PrefaultFunc(fn func() T) *ZodXor[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta attaches metadata to this schema.
func (z *ZodXor[T, R]) Meta(meta core.GlobalMeta) *ZodXor[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe sets a human-readable description for this schema.
func (z *ZodXor[T, R]) Describe(description string) *ZodXor[T, R] {
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

// Options returns a copy of all exclusive union member schemas.
func (z *ZodXor[T, R]) Options() []core.ZodSchema {
	result := make([]core.ZodSchema, len(z.internals.Options))
	copy(result, z.internals.Options)
	return result
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation pipeline.
func (z *ZodXor[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractUnionValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe chains this schema's output into another schema for further validation.
func (z *ZodXor[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractUnionValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation check with type-safe access to the parsed value.
func (z *ZodXor[T, R]) Refine(fn func(R) bool, params ...any) *ZodXor[T, R] {
	wrapper := func(v any) bool {
		cv, ok := convertToUnionConstraint[T, R](v)
		if !ok {
			return false
		}
		return fn(cv)
	}
	return z.withCheck(checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...)))
}

// RefineAny adds a custom validation check operating on the raw value.
func (z *ZodXor[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodXor[T, R] {
	return z.withCheck(checks.NewCustom[any](fn, utils.NormalizeCustomParams(params...)))
}

// =============================================================================
// COMPOSITION METHODS (Zod v4 Compatibility)
// =============================================================================

// And creates an intersection of this schema with another.
func (z *ZodXor[T, R]) And(other any) *ZodIntersection[any, any] {
	return Intersection(z, other)
}

// Or creates a union of this schema with another, enabling chaining.
func (z *ZodXor[T, R]) Or(other any) *ZodUnion[any, any] {
	return Union([]any{z, other})
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodXor[T, R]) withCheck(c core.ZodCheck) *ZodXor[T, R] {
	in := z.internals.Clone()
	in.AddCheck(c)
	return z.withInternals(in)
}

func (z *ZodXor[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodXor[T, *T] {
	return &ZodXor[T, *T]{
		internals: &ZodXorInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Options:          z.internals.Options,
		},
	}
}

func (z *ZodXor[T, R]) withInternals(in *core.ZodTypeInternals) *ZodXor[T, R] {
	return &ZodXor[T, R]{
		internals: &ZodXorInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Options:          z.internals.Options,
		},
	}
}

// CloneFrom copies configuration from another schema.
func (z *ZodXor[T, R]) CloneFrom(source any) {
	src, ok := source.(*ZodXor[T, R])
	if !ok {
		return
	}
	z.internals = src.internals
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

func newZodXorFromDef[T any, R any](def *ZodXorDef) *ZodXor[T, R] {
	internals := &ZodXorInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Options:          def.Options,
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		xorDef := &ZodXorDef{
			ZodTypeDef: *newDef,
			Options:    def.Options,
		}
		return any(newZodXorFromDef[T, R](xorDef)).(core.ZodType[any])
	}

	schema := &ZodXor[T, R]{internals: internals}

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

// Xor creates an exclusive union schema that requires exactly one option to match.
// Unlike Union which succeeds when any option matches, Xor fails if zero or multiple match.
func Xor(options []any, args ...any) *ZodXor[any, any] {
	return XorTyped[any, any](options, args...)
}

// XorPtr creates an exclusive union schema with pointer constraint type.
func XorPtr(options []any, args ...any) *ZodXor[any, *any] {
	return XorTyped[any, *any](options, args...)
}

// XorOf creates an exclusive union schema from variadic arguments.
func XorOf(schemas ...any) *ZodXor[any, any] {
	return Xor(schemas)
}

// XorOfPtr creates an exclusive union schema from variadic arguments with pointer constraint.
func XorOfPtr(schemas ...any) *ZodXor[any, *any] {
	return XorPtr(schemas)
}

// XorTyped creates a typed exclusive union schema with generic constraints.
func XorTyped[T any, R any](options []any, args ...any) *ZodXor[T, R] {
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	wrappedOptions := make([]core.ZodSchema, len(options))
	for i, option := range options {
		if option == nil {
			continue
		}
		zodSchema, err := core.ConvertToZodSchema(option)
		if err != nil {
			panic(fmt.Sprintf("Xor option %d: %v", i, err))
		}
		wrappedOptions[i] = zodSchema
	}

	def := &ZodXorDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeXor,
			Checks: []core.ZodCheck{},
		},
		Options: wrappedOptions,
	}
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	schema := newZodXorFromDef[T, R](def)

	// Trigger the validation pipeline even when no user-defined checks exist.
	schema.internals.AddCheck(checks.NewCustom[any](func(any) bool { return true }, core.SchemaParams{}))

	return schema
}
