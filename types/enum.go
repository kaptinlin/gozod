package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodEnumDef defines the schema definition for enum validation
type ZodEnumDef[T comparable] struct {
	core.ZodTypeDef
	Entries map[string]T // Enum key-value mapping
}

// ZodEnumInternals contains the internal state for enum schema
type ZodEnumInternals[T comparable] struct {
	core.ZodTypeInternals
	Def     *ZodEnumDef[T] // Schema definition reference
	Entries map[string]T   // Enum key-value mapping
	Values  map[T]struct{} // Set of valid values for fast lookup
	Pattern *regexp.Regexp // Compiled regex pattern for validation
}

// ZodEnum represents a type-safe enum validation schema with unified constraint
// T is the base comparable type, R is the constraint type (T | *T)
type ZodEnum[T comparable, R any] struct {
	internals *ZodEnumInternals[T]
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodEnum[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodEnum[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodEnum[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using enum-specific parsing logic
func (z *ZodEnum[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeEnum,
		// Validator function with multiple error collection (TypeScript Zod v4 behavior adapted for enum)
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
			return z.validateEnumWithIssues(value, checks, ctx)
		},
		// Use standard converter function
		engine.ConvertToConstraintType[T, R],
		ctx...,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodEnum[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type R.
func (z *ZodEnum[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	// Validator function with multiple error collection (TypeScript Zod v4 behavior adapted for enum)
	validator := func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
		return z.validateEnumWithIssues(value, checks, ctx)
	}

	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeEnum,
		validator,
		ctx...,
	)
}

// MustStrictParse validates input with compile-time type safety and panics on failure.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodEnum[T, R]) MustStrictParse(input R, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodEnum[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *T constraint because the optional value may be nil.
func (z *ZodEnum[T, R]) Optional() *ZodEnum[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *T constraint because the value may be nil.
func (z *ZodEnum[T, R]) Nilable() *ZodEnum[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodEnum[T, R]) Nullish() *ZodEnum[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default keeps the current generic constraint type R.
func (z *ZodEnum[T, R]) Default(v T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic constraint type R.
func (z *ZodEnum[T, R]) DefaultFunc(fn func() T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodEnum[T, R]) Prefault(v T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodEnum[T, R]) PrefaultFunc(fn func() T) *ZodEnum[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this enum schema.
func (z *ZodEnum[T, R]) Meta(meta core.GlobalMeta) *ZodEnum[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// ENUM SPECIFIC METHODS
// =============================================================================

// Enum returns the enum key-value mapping
func (z *ZodEnum[T, R]) Enum() map[string]T {
	result := make(map[string]T, len(z.internals.Entries))
	for k, v := range z.internals.Entries {
		result[k] = v
	}
	return result
}

// Options returns all possible enum values
func (z *ZodEnum[T, R]) Options() []T {
	values := make([]T, 0, len(z.internals.Values))
	for value := range z.internals.Values {
		values = append(values, value)
	}
	return values
}

// Extract creates a sub-enum with specified keys
// Non-existent keys are silently ignored to maintain fluent interface design
func (z *ZodEnum[T, R]) Extract(keys []string, params ...any) *ZodEnum[T, R] {
	newEntries := make(map[string]T)
	for _, key := range keys {
		if value, exists := z.internals.Entries[key]; exists {
			newEntries[key] = value
		}
		// Silently ignore non-existent keys for chainability
	}
	return EnumMapTyped[T, R](newEntries, params...)
}

// Exclude creates a sub-enum excluding specified keys
// Non-existent keys are silently ignored to maintain fluent interface design
func (z *ZodEnum[T, R]) Exclude(keys []string, params ...any) *ZodEnum[T, R] {
	excludeSet := make(map[string]bool, len(keys))
	for _, key := range keys {
		// Silently ignore non-existent keys for chainability
		excludeSet[key] = true
	}

	newEntries := make(map[string]T)
	for key, value := range z.internals.Entries {
		if !excludeSet[key] {
			newEntries[key] = value
		}
	}
	return EnumMapTyped[T, R](newEntries, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
func (z *ZodEnum[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		enumValue := extractEnumValue[T, R](input)
		return fn(enumValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a pipeline using the WrapFn pattern.
func (z *ZodEnum[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		enumValue := extractEnumValue[T, R](input)
		return target.Parse(enumValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation that matches the schema's output type R.
// The callback will be executed even when the value is nil (for *T schemas)
// to align with Zod v4 semantics.
func (z *ZodEnum[T, R]) Refine(fn func(R) bool, params ...any) *ZodEnum[T, R] {
	// Wrapper converts the raw value (always T or nil) into R before calling fn.
	wrapper := func(v any) bool {
		var zero R

		switch any(zero).(type) {
		case *T:
			// Schema output is *T – convert incoming value (T or nil) to *T
			if v == nil {
				return fn(any((*T)(nil)).(R))
			}
			if enumVal, ok := v.(T); ok {
				eCopy := enumVal
				ptr := &eCopy
				return fn(any(ptr).(R))
			}
			return false
		default:
			// Schema output is T
			if v == nil {
				// nil should never reach here for T schema; treat as failure.
				return false
			}
			if enumVal, ok := v.(T); ok {
				return fn(any(enumVal).(R))
			}
			return false
		}
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	// Convert back to the format expected by checks.NewCustom
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)

	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodEnum[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodEnum[T, R] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// validateEnumWithIssues validates enum value and applies checks with multiple error collection (TypeScript Zod v4 behavior adapted for enum)
func (z *ZodEnum[T, R]) validateEnumWithIssues(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	var collectedIssues []core.ZodRawIssue

	// First check if value is in enum and collect any enum validation issues
	if _, exists := z.internals.Values[value]; !exists {
		// Create list of valid values for error message
		validValues := make([]any, 0, len(z.internals.Values))
		for v := range z.internals.Values {
			validValues = append(validValues, v)
		}
		rawIssue := issues.CreateIssue(core.InvalidValue, "Invalid enum value", map[string]any{
			"received": fmt.Sprintf("%v", value),
			"options":  validValues,
		}, value)
		collectedIssues = append(collectedIssues, rawIssue)
	}

	// Apply additional checks and collect their issues (TypeScript Zod v4 behavior)
	if len(checks) > 0 {
		// Use RunChecksOnValue to collect issues instead of ApplyChecks which fails fast
		payload := core.NewParsePayload(value)
		result := engine.RunChecksOnValue(value, checks, payload, ctx)

		if result.HasIssues() {
			checkIssues := result.GetIssues()
			collectedIssues = append(collectedIssues, checkIssues...)
		}

		// Get the potentially transformed value for return
		if result.GetValue() != nil {
			if transformedValue, ok := result.GetValue().(T); ok {
				value = transformedValue
			}
		}
	}

	// If we collected any issues, return them as a combined error
	if len(collectedIssues) > 0 {
		var zero T
		return zero, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return value, nil
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates a new ZodEnum instance of constraint type *T.
// Used by modifiers such as Optional, Nilable, and Nullish that must return a pointer constraint.
func (z *ZodEnum[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodEnum[T, *T] {
	return &ZodEnum[T, *T]{internals: &ZodEnumInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Entries:          z.internals.Entries,
		Values:           z.internals.Values,
		Pattern:          z.internals.Pattern,
	}}
}

// withInternals creates a new ZodEnum instance that keeps the original constraint type R.
// Used by modifiers that retain the original constraint, such as Default, Prefault, and Transform.
func (z *ZodEnum[T, R]) withInternals(in *core.ZodTypeInternals) *ZodEnum[T, R] {
	return &ZodEnum[T, R]{internals: &ZodEnumInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Entries:          z.internals.Entries,
		Values:           z.internals.Values,
		Pattern:          z.internals.Pattern,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodEnum[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodEnum[T, R]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.Checks = originalChecks
	}
}

// extractEnumValue extracts the base enum value T from constraint type R
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

// buildEnumPattern creates regex pattern for enum validation
func buildEnumPattern[T comparable](values map[T]struct{}) *regexp.Regexp {
	if len(values) == 0 {
		return regexp.MustCompile("^$a") // Never matches
	}

	patterns := make([]string, 0, len(values))
	for value := range values {
		patterns = append(patterns, regexp.QuoteMeta(fmt.Sprintf("%v", value)))
	}
	patternStr := "^(" + strings.Join(patterns, "|") + ")$"
	return regexp.MustCompile(patternStr)
}

// newZodEnumFromDef constructs new ZodEnum from definition
func newZodEnumFromDef[T comparable, R any](def *ZodEnumDef[T]) *ZodEnum[T, R] {
	values := make(map[T]struct{}, len(def.Entries))
	for _, value := range def.Entries {
		values[value] = struct{}{}
	}

	// Convert typed values to map[any]struct{} for ZodTypeInternals.Values
	anyValues := make(map[any]struct{}, len(values))
	for value := range values {
		anyValues[value] = struct{}{}
	}

	internals := &ZodEnumInternals[T]{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Values: anyValues, // Set the Values field for discriminator extraction
			Bag:    make(map[string]any),
		},
		Def:     def,
		Entries: def.Entries,
		Values:  values,
		Pattern: buildEnumPattern(values),
	}

	// Provide constructor for AddCheck functionality
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

// Enum creates enum schema from values with type-inference support
func Enum[T comparable](values ...T) *ZodEnum[T, T] {
	return EnumSlice(values)
}

// EnumSlice creates enum schema from slice of values
func EnumSlice[T comparable](values []T) *ZodEnum[T, T] {
	entries := make(map[string]T, len(values))
	for i, value := range values {
		key := fmt.Sprintf("%d", i)
		entries[key] = value
	}
	return EnumMapTyped[T, T](entries)
}

// EnumMap creates enum schema from key-value mapping
func EnumMap[T comparable](entries map[string]T, params ...any) *ZodEnum[T, T] {
	return EnumMapTyped[T, T](entries, params...)
}

// EnumMapTyped is the generic constructor for enum schemas from mapping
func EnumMapTyped[T comparable, R any](entries map[string]T, args ...any) *ZodEnum[T, R] {
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodEnumDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeEnum,
			Checks: []core.ZodCheck{},
		},
		Entries: entries,
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodEnumFromDef[T, R](def)
}

// EnumPtr creates pointer-capable enum schema from values
func EnumPtr[T comparable](values ...T) *ZodEnum[T, *T] {
	return EnumSlicePtr(values)
}

// EnumSlicePtr creates pointer-capable enum schema from slice of values
func EnumSlicePtr[T comparable](values []T) *ZodEnum[T, *T] {
	entries := make(map[string]T, len(values))
	for i, value := range values {
		key := fmt.Sprintf("%d", i)
		entries[key] = value
	}
	return EnumMapTyped[T, *T](entries)
}

// EnumMapPtr creates pointer-capable enum schema from key-value mapping
func EnumMapPtr[T comparable](entries map[string]T, params ...any) *ZodEnum[T, *T] {
	return EnumMapTyped[T, *T](entries, params...)
}
