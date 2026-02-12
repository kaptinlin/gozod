package types

import (
	"fmt"
	"maps"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodEnumDef defines the configuration for enum validation.
type ZodEnumDef[T comparable] struct {
	core.ZodTypeDef
	Entries map[string]T
}

// ZodEnumInternals contains enum validator internal state.
type ZodEnumInternals[T comparable] struct {
	core.ZodTypeInternals
	Def     *ZodEnumDef[T]
	Entries map[string]T
	Values  map[T]struct{}
}

// ZodEnum represents a type-safe enum validation schema.
// T is the base comparable type, R is the constraint type (T or *T).
type ZodEnum[T comparable, R any] struct {
	internals *ZodEnumInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodEnum[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodEnum[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodEnum[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input and returns a value matching the generic type R.
func (z *ZodEnum[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeEnum,
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
			return z.validateEnum(value, checks, ctx)
		},
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// MustParse validates input and panics on failure.
func (z *ZodEnum[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type R.
func (z *ZodEnum[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeEnum,
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
			return z.validateEnum(value, checks, ctx)
		},
		ctx...,
	)
}

// MustStrictParse provides compile-time type safety and panics on failure.
func (z *ZodEnum[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns any type for runtime interface usage.
func (z *ZodEnum[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a new schema that accepts nil, with *T constraint.
func (z *ZodEnum[T, R]) Optional() *ZodEnum[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// ExactOptional accepts absent keys but rejects explicit nil values.
// Unlike Optional, ExactOptional only accepts absent keys in object fields.
func (z *ZodEnum[T, R]) ExactOptional() *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetExactOptional(true)
	return z.withInternals(in)
}

// Nilable returns a new schema that accepts nil values, with *T constraint.
func (z *ZodEnum[T, R]) Nilable() *ZodEnum[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish returns a new schema combining optional and nilable modifiers.
func (z *ZodEnum[T, R]) Nullish() *ZodEnum[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default sets a fallback value returned when input is nil (short-circuits validation).
func (z *ZodEnum[T, R]) Default(v T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets a fallback function called when input is nil (short-circuits validation).
func (z *ZodEnum[T, R]) DefaultFunc(fn func() T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets a fallback value that goes through the full validation pipeline.
func (z *ZodEnum[T, R]) Prefault(v T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	var zero R
	switch any(zero).(type) {
	case *T:
		in.SetPrefaultValue(&v)
	default:
		in.SetPrefaultValue(v)
	}
	return z.withInternals(in)
}

// PrefaultFunc sets a fallback function that goes through the full validation pipeline.
func (z *ZodEnum[T, R]) PrefaultFunc(fn func() T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		v := fn()
		var zero R
		switch any(zero).(type) {
		case *T:
			return &v
		default:
			return v
		}
	})
	return z.withInternals(in)
}

// Meta stores metadata for this enum schema in the global registry.
func (z *ZodEnum[T, R]) Meta(meta core.GlobalMeta) *ZodEnum[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodEnum[T, R]) Describe(description string) *ZodEnum[T, R] {
	in := z.internals.Clone()
	existing, ok := core.GlobalRegistry.Get(z)
	if !ok {
		existing = core.GlobalMeta{}
	}
	existing.Description = description
	clone := z.withInternals(in)
	core.GlobalRegistry.Add(clone, existing)
	return clone
}

// =============================================================================
// ENUM SPECIFIC METHODS
// =============================================================================

// Enum returns a copy of the enum key-value mapping.
func (z *ZodEnum[T, R]) Enum() map[string]T {
	result := make(map[string]T, len(z.internals.Entries))
	maps.Copy(result, z.internals.Entries)
	return result
}

// Options returns all possible enum values.
func (z *ZodEnum[T, R]) Options() []T {
	values := make([]T, 0, len(z.internals.Values))
	for value := range z.internals.Values {
		values = append(values, value)
	}
	return values
}

// Extract creates a sub-enum containing only the specified keys.
// Non-existent keys are silently ignored.
func (z *ZodEnum[T, R]) Extract(keys []string, params ...any) *ZodEnum[T, R] {
	newEntries := make(map[string]T)
	for _, key := range keys {
		if value, exists := z.internals.Entries[key]; exists {
			newEntries[key] = value
		}
	}
	return EnumMapTyped[T, R](newEntries, params...)
}

// Exclude creates a sub-enum without the specified keys.
// Non-existent keys are silently ignored.
func (z *ZodEnum[T, R]) Exclude(keys []string, params ...any) *ZodEnum[T, R] {
	excludeSet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		excludeSet[key] = struct{}{}
	}

	newEntries := make(map[string]T)
	for key, value := range z.internals.Entries {
		if _, excluded := excludeSet[key]; !excluded {
			newEntries[key] = value
		}
	}
	return EnumMapTyped[T, R](newEntries, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function to the parsed enum value.
func (z *ZodEnum[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(extractEnumValue[T, R](input), ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a validation pipeline to a target schema.
func (z *ZodEnum[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(extractEnumValue[T, R](input), ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation matching the schema's output type R.
// The callback receives nil for *T schemas when the value is nil (Zod v4 semantics).
func (z *ZodEnum[T, R]) Refine(fn func(R) bool, params ...any) *ZodEnum[T, R] {
	wrapper := func(v any) bool {
		var zero R
		switch any(zero).(type) {
		case *T:
			if v == nil {
				return fn(any((*T)(nil)).(R))
			}
			if enumVal, ok := v.(T); ok {
				return fn(any(&enumVal).(R))
			}
			return false
		default:
			if v == nil {
				return false
			}
			if enumVal, ok := v.(T); ok {
				return fn(any(enumVal).(R))
			}
			return false
		}
	}

	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// RefineAny applies validation without type conversion.
func (z *ZodEnum[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodEnum[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error
	}

	check := checks.NewCustom[any](fn, errorMessage)
	in := z.internals.Clone()
	in.AddCheck(check)
	return z.withInternals(in)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// validateEnum validates the enum value and applies checks, collecting all issues.
func (z *ZodEnum[T, R]) validateEnum(value T, chks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	var collectedIssues []core.ZodRawIssue

	if _, exists := z.internals.Values[value]; !exists {
		validValues := make([]any, 0, len(z.internals.Values))
		for v := range z.internals.Values {
			validValues = append(validValues, v)
		}
		collectedIssues = append(collectedIssues, issues.CreateIssue(
			core.InvalidValue, "Invalid enum value", map[string]any{
				"received": fmt.Sprintf("%v", value),
				"options":  validValues,
			}, value))
	}

	if len(chks) > 0 {
		payload := core.NewParsePayload(value)
		result := engine.RunChecksOnValue(value, chks, payload, ctx)
		if result.HasIssues() {
			collectedIssues = append(collectedIssues, result.Issues()...)
		}
		if result.Value() != nil {
			if transformed, ok := result.Value().(T); ok {
				value = transformed
			}
		}
	}

	if len(collectedIssues) > 0 {
		var zero T
		return zero, issues.CreateArrayValidationIssues(collectedIssues)
	}
	return value, nil
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodEnum with *T constraint for Optional/Nilable/Nullish.
func (z *ZodEnum[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodEnum[T, *T] {
	return &ZodEnum[T, *T]{internals: &ZodEnumInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Entries:          z.internals.Entries,
		Values:           z.internals.Values,
	}}
}

// withInternals creates a new ZodEnum keeping the original constraint type R.
func (z *ZodEnum[T, R]) withInternals(in *core.ZodTypeInternals) *ZodEnum[T, R] {
	return &ZodEnum[T, R]{internals: &ZodEnumInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Entries:          z.internals.Entries,
		Values:           z.internals.Values,
	}}
}

// CloneFrom copies configuration from another schema, preserving original checks.
func (z *ZodEnum[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodEnum[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// extractEnumValue extracts the base value T from constraint type R.
func extractEnumValue[T comparable, R any](value R) T {
	if ptr, ok := any(value).(*T); ok {
		if ptr != nil {
			return *ptr
		}
		var zero T
		return zero
	}
	return any(value).(T)
}

// newZodEnumFromDef constructs a new ZodEnum from a definition.
func newZodEnumFromDef[T comparable, R any](def *ZodEnumDef[T]) *ZodEnum[T, R] {
	values := make(map[T]struct{}, len(def.Entries))
	for _, value := range def.Entries {
		values[value] = struct{}{}
	}

	anyValues := make(map[any]struct{}, len(values))
	for value := range values {
		anyValues[value] = struct{}{}
	}

	internals := &ZodEnumInternals[T]{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Values: anyValues,
			Bag:    make(map[string]any),
		},
		Def:     def,
		Entries: def.Entries,
		Values:  values,
	}

	// Provide constructor for AddCheck functionality.
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		enumDef := &ZodEnumDef[T]{
			ZodTypeDef: *newDef,
			Entries:    def.Entries,
		}
		return any(newZodEnumFromDef[T, R](enumDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodEnum[T, R]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Enum creates an enum schema from values with type inference.
func Enum[T comparable](values ...T) *ZodEnum[T, T] {
	return EnumSlice(values)
}

// EnumSlice creates an enum schema from a slice of values.
func EnumSlice[T comparable](values []T) *ZodEnum[T, T] {
	entries := make(map[string]T, len(values))
	for i, value := range values {
		entries[fmt.Sprintf("%d", i)] = value
	}
	return EnumMapTyped[T, T](entries)
}

// EnumMap creates an enum schema from a key-value mapping.
func EnumMap[T comparable](entries map[string]T, params ...any) *ZodEnum[T, T] {
	return EnumMapTyped[T, T](entries, params...)
}

// EnumMapTyped is the generic constructor for enum schemas.
func EnumMapTyped[T comparable, R any](entries map[string]T, args ...any) *ZodEnum[T, R] {
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodEnumDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeEnum,
			Checks: []core.ZodCheck{},
		},
		Entries: entries,
	}

	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodEnumFromDef[T, R](def)
}

// EnumPtr creates a pointer-capable enum schema from values.
func EnumPtr[T comparable](values ...T) *ZodEnum[T, *T] {
	return EnumSlicePtr(values)
}

// EnumSlicePtr creates a pointer-capable enum schema from a slice of values.
func EnumSlicePtr[T comparable](values []T) *ZodEnum[T, *T] {
	entries := make(map[string]T, len(values))
	for i, value := range values {
		entries[fmt.Sprintf("%d", i)] = value
	}
	return EnumMapTyped[T, *T](entries)
}

// EnumMapPtr creates a pointer-capable enum schema from a key-value mapping.
func EnumMapPtr[T comparable](entries map[string]T, params ...any) *ZodEnum[T, *T] {
	return EnumMapTyped[T, *T](entries, params...)
}
