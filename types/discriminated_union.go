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

// Sentinel errors for discriminated union construction and validation.
var (
	ErrSchemaIsNil                   = errors.New("schema is nil")
	ErrSchemaInternalsIsNil          = errors.New("schema internals is nil")
	ErrNoDiscriminatorValuesFound    = errors.New("no discriminator values found for field")
	ErrOptionIsNil                   = errors.New("option is nil")
	ErrDuplicateDiscriminatorValue   = errors.New("duplicate discriminator value")
	ErrFailedToBuildDiscriminatorMap = errors.New("failed to build discriminator map")
	ErrNoValidDiscriminatorValues    = errors.New("no valid discriminator values found for field")
)

// ZodDiscriminatedUnionDef is the configuration for discriminated union validation.
type ZodDiscriminatedUnionDef struct {
	core.ZodTypeDef
	Discriminator string
	Options       []core.ZodSchema
}

// ZodDiscriminatedUnionInternals contains discriminated union validator internal state.
type ZodDiscriminatedUnionInternals struct {
	core.ZodTypeInternals
	Def           *ZodDiscriminatedUnionDef
	Discriminator string
	Options       []core.ZodSchema
	DiscMap       map[any]core.ZodSchema
}

// ZodDiscriminatedUnion is a discriminated union validation schema.
type ZodDiscriminatedUnion[T any, R any] struct {
	internals *ZodDiscriminatedUnionInternals
}

// GetInternals returns the internal state for framework usage.
func (z *ZodDiscriminatedUnion[T, R]) GetInternals() *core.ZodTypeInternals {
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
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	if constructionError, exists := z.internals.Bag["construction_error"]; exists {
		if errMsg, ok := constructionError.(string); ok {
			var zero R
			return zero, issues.CreateInvalidSchemaError(errMsg, input, parseCtx)
		}
	}

	if input == nil {
		if z.internals.Nilable || z.internals.Optional {
			var zero R
			return zero, nil
		}
	}

	if input == nil && (z.internals.DefaultValue != nil || z.internals.DefaultFunc != nil) {
		if z.internals.DefaultFunc != nil {
			input = z.internals.DefaultFunc()
		} else {
			input = z.internals.DefaultValue
		}
	}

	inputMap, ok := input.(map[string]any)
	if !ok {
		if z.internals.PrefaultValue != nil || z.internals.PrefaultFunc != nil {
			var pv any
			if z.internals.PrefaultFunc != nil {
				pv = z.internals.PrefaultFunc()
			} else {
				pv = z.internals.PrefaultValue
			}
			return z.Parse(pv, parseCtx)
		}
		var zero R
		return zero, issues.CreateInvalidTypeError(core.ZodTypeObject, input, parseCtx)
	}

	discVal, exists := inputMap[z.internals.Discriminator]
	if !exists {
		var zero R
		return zero, issues.CreateMissingRequiredError(z.internals.Discriminator, "discriminator field", input, parseCtx)
	}

	targetSchema, found := z.internals.DiscMap[discVal]
	var result any
	var err error
	if found {
		result, err = targetSchema.ParseAny(inputMap, parseCtx)
	} else {
		var errs []error
		for _, option := range z.internals.Options {
			if option == nil {
				continue
			}
			if res, e := option.ParseAny(inputMap, parseCtx); e == nil {
				result = res
				err = nil
				break
			} else {
				errs = append(errs, e)
			}
		}
		if result == nil {
			var zero R
			return zero, issues.CreateInvalidUnionError(errs, input, parseCtx)
		}
	}

	if err != nil {
		var zero R
		return zero, err
	}

	if len(z.internals.Checks) > 0 {
		checked, validationErr := engine.ApplyChecks[any](result, z.internals.Checks, parseCtx)
		if validationErr != nil {
			var zero R
			return zero, validationErr
		}
		result = checked
	}

	return convertToDiscriminatedUnionConstraintType[T, R](result), nil
}

// MustParse validates input and panics on failure.
func (z *ZodDiscriminatedUnion[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type T.
func (z *ZodDiscriminatedUnion[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	constraintInput, ok := convertToDiscriminatedUnionConstraintValue[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"discriminated union constraint type",
			any(input),
			ctx[0],
		)
	}

	return z.Parse(constraintInput, ctx...)
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
func (z *ZodDiscriminatedUnion[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodDiscriminatedUnion[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// Optional creates an optional schema that returns pointer constraint.
func (z *ZodDiscriminatedUnion[T, R]) Optional() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint.
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

// Default sets a default value used when input is nil.
func (z *ZodDiscriminatedUnion[T, R]) Default(v T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a default value function used when input is nil.
func (z *ZodDiscriminatedUnion[T, R]) DefaultFunc(fn func() T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides a fallback value on validation failure.
func (z *ZodDiscriminatedUnion[T, R]) Prefault(v T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback value function on validation failure.
func (z *ZodDiscriminatedUnion[T, R]) PrefaultFunc(fn func() T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this schema.
func (z *ZodDiscriminatedUnion[T, R]) Meta(meta core.GlobalMeta) *ZodDiscriminatedUnion[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodDiscriminatedUnion[T, R]) Describe(description string) *ZodDiscriminatedUnion[T, R] {
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

// Discriminator returns the discriminator field name.
func (z *ZodDiscriminatedUnion[T, R]) Discriminator() string {
	return z.internals.Discriminator
}

// Options returns a copy of all union member schemas.
func (z *ZodDiscriminatedUnion[T, R]) Options() []core.ZodSchema {
	result := make([]core.ZodSchema, len(z.internals.Options))
	copy(result, z.internals.Options)
	return result
}

// DiscriminatorMap returns a copy of the discriminator value to schema mapping.
func (z *ZodDiscriminatedUnion[T, R]) DiscriminatorMap() map[any]core.ZodSchema {
	result := make(map[any]core.ZodSchema, len(z.internals.DiscMap))
	maps.Copy(result, z.internals.DiscMap)
	return result
}

// Transform creates a type-safe transformation pipeline.
func (z *ZodDiscriminatedUnion[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractDiscriminatedUnionValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a validation pipeline to another schema.
func (z *ZodDiscriminatedUnion[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractDiscriminatedUnionValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// Refine applies type-safe validation with constraint type R.
func (z *ZodDiscriminatedUnion[T, R]) Refine(fn func(R) bool, params ...any) *ZodDiscriminatedUnion[T, R] {
	wrapper := func(v any) bool {
		if cv, ok := convertToDiscriminatedUnionConstraintValue[T, R](v); ok {
			return fn(cv)
		}
		return false
	}

	p := utils.NormalizeParams(params...)

	var errMsg any
	if p.Error != nil {
		errMsg = p.Error
	}

	return z.withCheck(checks.NewCustom[any](wrapper, errMsg))
}

// RefineAny provides flexible validation without type conversion.
func (z *ZodDiscriminatedUnion[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodDiscriminatedUnion[T, R] {
	p := utils.NormalizeParams(params...)

	var errMsg any
	if p.Error != nil {
		errMsg = p.Error
	}

	return z.withCheck(checks.NewCustom[any](fn, errMsg))
}

func (z *ZodDiscriminatedUnion[T, R]) withCheck(check core.ZodCheck) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

func (z *ZodDiscriminatedUnion[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodDiscriminatedUnion[T, *T] {
	return &ZodDiscriminatedUnion[T, *T]{internals: &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Discriminator:    z.internals.Discriminator,
		Options:          z.internals.Options,
		DiscMap:          z.internals.DiscMap,
	}}
}

func (z *ZodDiscriminatedUnion[T, R]) withInternals(in *core.ZodTypeInternals) *ZodDiscriminatedUnion[T, R] {
	return &ZodDiscriminatedUnion[T, R]{internals: &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Discriminator:    z.internals.Discriminator,
		Options:          z.internals.Options,
		DiscMap:          z.internals.DiscMap,
	}}
}

// CloneFrom copies configuration from another schema.
func (z *ZodDiscriminatedUnion[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodDiscriminatedUnion[T, R]); ok {
		z.internals = src.internals
	}
}

// convertToDiscriminatedUnionConstraintType converts a base type T to constraint type R.
func convertToDiscriminatedUnionConstraintType[T any, R any](value any) R {
	var zero R
	switch any(zero).(type) {
	case *any:
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R)
		}
		return any((*any)(nil)).(R)
	default:
		return any(value).(R) //nolint:unconvert // Required for generic type constraint conversion
	}
}

// extractDiscriminatedUnionValue extracts base type T from constraint type R.
func extractDiscriminatedUnionValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *any:
		if v != nil {
			return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToDiscriminatedUnionConstraintValue converts any value to constraint type R.
func convertToDiscriminatedUnionConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	if _, ok := any(zero).(*any); ok {
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R), true
		}
		return any((*any)(nil)).(R), true
	}

	return zero, false
}

// getDiscriminatorValues extracts discriminator values from a schema.
func getDiscriminatorValues(schema core.ZodSchema, discriminatorField string) ([]any, error) {
	if schema == nil {
		return nil, ErrSchemaIsNil
	}

	internals := schema.GetInternals()
	if internals == nil {
		return nil, ErrSchemaInternalsIsNil
	}

	values := extractDiscriminatorFromInternals(internals, discriminatorField)
	if len(values) > 0 {
		return values, nil
	}

	return getDiscriminatorValuesFromAnySchema(schema, discriminatorField)
}

// getDiscriminatorValuesFromAnySchema extracts discriminator values via type assertion on Shape methods.
func getDiscriminatorValuesFromAnySchema(schema any, discriminatorField string) ([]any, error) {
	if schema == nil {
		return nil, ErrSchemaIsNil
	}

	if s, ok := schema.(interface {
		GetInternals() *core.ZodTypeInternals
	}); ok {
		if internals := s.GetInternals(); internals != nil {
			values := extractDiscriminatorFromInternals(internals, discriminatorField)
			if len(values) > 0 {
				return values, nil
			}
		}
	}

	if s, ok := schema.(interface {
		Shape() core.ObjectSchema
	}); ok {
		shape := s.Shape()
		if fieldSchema, exists := shape[discriminatorField]; exists {
			if fi := fieldSchema.GetInternals(); fi != nil {
				values := extractDiscriminatorFromInternals(fi, discriminatorField)
				if len(values) > 0 {
					return values, nil
				}
			}
		}
	}

	if s, ok := schema.(interface {
		Shape() core.StructSchema
	}); ok {
		shape := s.Shape()
		if fieldSchema, exists := shape[discriminatorField]; exists {
			if fi := fieldSchema.GetInternals(); fi != nil {
				values := extractDiscriminatorFromInternals(fi, discriminatorField)
				if len(values) > 0 {
					return values, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("%w '%s' in schema", ErrNoDiscriminatorValuesFound, discriminatorField)
}

// extractDiscriminatorFromInternals extracts discriminator values from schema internals.
func extractDiscriminatorFromInternals(internals *core.ZodTypeInternals, _ string) []any {
	if internals == nil {
		return nil
	}

	var values []any

	if len(internals.Values) > 0 {
		for value := range internals.Values {
			values = append(values, value)
		}
		return values
	}

	if internals.Bag != nil {
		if bagValues, exists := internals.Bag["values"]; exists {
			if valueMap, ok := bagValues.(map[any]struct{}); ok {
				for value := range valueMap {
					values = append(values, value)
				}
				return values
			}
		}
	}

	//nolint:exhaustive // Only interested in Literal and Enum types here
	switch internals.Type {
	case core.ZodTypeLiteral:
		if literalValue, exists := internals.Bag["literal"]; exists {
			values = append(values, literalValue)
		}
	case core.ZodTypeEnum:
		if enumValues, exists := internals.Bag["enum"]; exists {
			if enumMap, ok := enumValues.(map[any]struct{}); ok {
				for value := range enumMap {
					values = append(values, value)
				}
			}
		}
	}

	return values
}

// buildDiscriminatorMap builds the discriminator value to schema mapping.
func buildDiscriminatorMap(discriminator string, options []core.ZodSchema) (map[any]core.ZodSchema, error) {
	discMap := make(map[any]core.ZodSchema)
	var errs []error

	for i, option := range options {
		if option == nil {
			errs = append(errs, fmt.Errorf("%w: option %d", ErrOptionIsNil, i))
			continue
		}

		values, err := getDiscriminatorValues(option, discriminator)
		if err != nil {
			errs = append(errs, fmt.Errorf("option %d: %w", i, err))
			continue
		}

		for _, value := range values {
			if _, exists := discMap[value]; exists {
				return nil, fmt.Errorf("%w: %v", ErrDuplicateDiscriminatorValue, value)
			}
			discMap[value] = option
		}
	}

	if len(discMap) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrFailedToBuildDiscriminatorMap, errs)
	}

	if len(discMap) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoValidDiscriminatorValues, discriminator)
	}

	return discMap, nil
}

// newZodDiscriminatedUnionFromDef constructs a ZodDiscriminatedUnion from a definition.
// Construction errors are deferred to parse-time for graceful handling.
func newZodDiscriminatedUnionFromDef[T any, R any](def *ZodDiscriminatedUnionDef) *ZodDiscriminatedUnion[T, R] {
	discMap, err := buildDiscriminatorMap(def.Discriminator, def.Options)

	internals := &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Discriminator:    def.Discriminator,
		Options:          def.Options,
		DiscMap:          discMap,
	}

	if err != nil {
		internals.Bag["construction_error"] = fmt.Sprintf("DiscriminatedUnion construction error: %v", err)
	}

	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		d := &ZodDiscriminatedUnionDef{
			ZodTypeDef:    *newDef,
			Discriminator: def.Discriminator,
			Options:       def.Options,
		}
		return any(newZodDiscriminatedUnionFromDef[T, R](d)).(core.ZodType[any])
	}

	schema := &ZodDiscriminatedUnion[T, R]{internals: internals}

	if def.Error != nil {
		internals.Error = def.Error
	}

	if len(def.Checks) > 0 {
		for _, check := range def.Checks {
			internals.AddCheck(check)
		}
	}

	return schema
}

// DiscriminatedUnion creates a discriminated union schema with value constraint.
func DiscriminatedUnion(discriminator string, options []any, args ...any) *ZodDiscriminatedUnion[any, any] {
	return DiscriminatedUnionTyped[any, any](discriminator, options, args...)
}

// DiscriminatedUnionPtr creates a discriminated union schema with pointer constraint.
func DiscriminatedUnionPtr(discriminator string, options []any, args ...any) *ZodDiscriminatedUnion[any, *any] {
	return DiscriminatedUnionTyped[any, *any](discriminator, options, args...)
}

// DiscriminatedUnionTyped creates a typed discriminated union schema with generic constraints.
func DiscriminatedUnionTyped[T any, R any](discriminator string, options []any, args ...any) *ZodDiscriminatedUnion[T, R] {
	param := utils.FirstParam(args...)
	params := utils.NormalizeParams(param)

	wrappedOptions := make([]core.ZodSchema, len(options))
	for i, option := range options {
		zodSchema, err := core.ConvertToZodSchema(option)
		if err != nil {
			panic(fmt.Sprintf("DiscriminatedUnion option %d: %v", i, err))
		}
		wrappedOptions[i] = zodSchema
	}

	def := &ZodDiscriminatedUnionDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeDiscriminated,
			Checks: []core.ZodCheck{},
		},
		Discriminator: discriminator,
		Options:       wrappedOptions,
	}

	if params != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, params)
	}

	return newZodDiscriminatedUnionFromDef[T, R](def)
}
