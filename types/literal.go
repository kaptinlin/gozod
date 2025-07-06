package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE CONSTRAINTS
// =============================================================================

// LiteralConstraint allows any comparable type for literal values
type LiteralConstraint[T comparable] interface {
	comparable
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodLiteralDef defines the configuration for literal validation
type ZodLiteralDef[T comparable] struct {
	core.ZodTypeDef
	Values []T // Allowed literal values
}

// ZodLiteralInternals contains literal validator internal state
type ZodLiteralInternals[T comparable] struct {
	core.ZodTypeInternals
	Def     *ZodLiteralDef[T] // Schema definition reference
	Values  map[T]struct{}    // Fast lookup set for allowed values
	Pattern *regexp.Regexp    // Compiled regex pattern for validation
}

// ZodLiteral represents a literal value validation schema with type safety
// T is the base comparable type, R is the constraint type (T | *T)
type ZodLiteral[T comparable, R any] struct {
	internals *ZodLiteralInternals[T]
}

// Type aliases for backward compatibility and clarity
type ZodLiteralValue[T comparable] = ZodLiteral[T, T] // Value type constraint
type ZodLiteralPtr[T comparable] = ZodLiteral[T, *T]  // Pointer type constraint

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodLiteral[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodLiteral[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodLiteral[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Coerce implements Coercible interface - literals don't support coercion
func (z *ZodLiteral[T, R]) Coerce(input any) (any, bool) {
	// Literals must match exactly - no coercion
	return input, false
}

// Parse returns a value that matches the literal values with full type safety
func (z *ZodLiteral[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeLiteral,
		// Custom validator function for literal-specific logic
		func(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
			// Type-specific validation FIRST
			if _, exists := z.internals.Values[value]; !exists {
				var zero T
				return zero, fmt.Errorf("invalid literal value: %v", value)
			}
			// Then run standard checks
			return engine.ApplyChecks[T](value, checks, ctx)
		},
		// Custom converter with literal validation and pointer identity preservation
		func(result any, ctx *core.ParseContext, expectedType core.ZodTypeCode) (R, error) {
			var zero R

			// Handle nil input for optional/nilable schemas
			if result == nil {
				internals := z.GetInternals()
				if internals.Optional || internals.Nilable {
					return zero, nil
				}
				return zero, fmt.Errorf("literal value cannot be nil")
			}

			// First, extract the base value for validation
			var baseValue T
			var found bool

			// Try direct conversion to T
			if value, ok := result.(T); ok {
				baseValue = value
				found = true
			} else if ptr, ok := result.(*T); ok && ptr != nil {
				// Handle pointer to T
				baseValue = *ptr
				found = true
			}

			if !found {
				return zero, fmt.Errorf("type conversion failed: expected %T or *%T but got %T", *new(T), *new(T), result)
			}

			// Validate against literal values
			if _, exists := z.internals.Values[baseValue]; !exists {
				return zero, fmt.Errorf("invalid literal value: %v", baseValue)
			}

			// Use standard converter for constraint type conversion with pointer identity preservation
			return engine.ConvertToConstraintType[T, R](result, ctx, expectedType)
		},
		ctx...,
	)
}

// MustParse is the type-safe variant that panics on error
func (z *ZodLiteral[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodLiteral[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns pointer constraint type because the optional value may be nil
func (z *ZodLiteral[T, R]) Optional() *ZodLiteral[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrConstraint(in)
}

// Nilable returns pointer constraint type because the value may be nil
func (z *ZodLiteral[T, R]) Nilable() *ZodLiteral[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrConstraint(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodLiteral[T, R]) Nullish() *ZodLiteral[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrConstraint(in)
}

// Default preserves the current constraint type R
func (z *ZodLiteral[T, R]) Default(v T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves the current constraint type R
func (z *ZodLiteral[T, R]) DefaultFunc(fn func() T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodLiteral[T, R]) Prefault(v T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodLiteral[T, R]) PrefaultFunc(fn func() T) *ZodLiteral[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Value returns the single literal value. Panics if the schema has multiple values.
func (z *ZodLiteral[T, R]) Value() T {
	if len(z.internals.Def.Values) != 1 {
		panic("literal.Value() is only available on single-value literal schemas")
	}
	return z.internals.Def.Values[0]
}

// Values returns all possible literal values
func (z *ZodLiteral[T, R]) Values() []T {
	values := make([]T, 0, len(z.internals.Def.Values))
	values = append(values, z.internals.Def.Values...)
	return values
}

// Contains checks if a value is in the allowed literal values
func (z *ZodLiteral[T, R]) Contains(value T) bool {
	_, exists := z.internals.Values[value]
	return exists
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodLiteral[T, R]) Transform(fn func(R, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		return fn(input, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodLiteral[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(any(input), ctx)
	}
	return core.NewZodPipe[R, any](z, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies custom validation logic with automatic type conversion
func (z *ZodLiteral[T, R]) Refine(fn func(T) bool, params ...any) *ZodLiteral[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			var zero T
			return fn(zero)
		}
		if literalVal, ok := v.(T); ok {
			return fn(literalVal)
		}
		return false
	}

	check := checks.NewCustom[any](wrapper, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodLiteral[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodLiteral[T, R] {
	check := checks.NewCustom[any](fn, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withInternals creates new instance preserving current constraint type R
func (z *ZodLiteral[T, R]) withInternals(in *core.ZodTypeInternals) *ZodLiteral[T, R] {
	return &ZodLiteral[T, R]{internals: &ZodLiteralInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Values:           z.internals.Values,
		Pattern:          z.internals.Pattern,
	}}
}

// withPtrConstraint creates new instance with pointer constraint type
func (z *ZodLiteral[T, R]) withPtrConstraint(in *core.ZodTypeInternals) *ZodLiteral[T, *T] {
	return &ZodLiteral[T, *T]{internals: &ZodLiteralInternals[T]{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Values:           z.internals.Values,
		Pattern:          z.internals.Pattern,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodLiteral[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodLiteral[T, R]); ok {
		// Preserve original checks to avoid overwriting them
		originalChecks := z.internals.ZodTypeInternals.Checks

		// Copy all state from source
		*z.internals = *src.internals

		// Restore the original checks that were set by the constructor
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// buildLiteralPattern creates regex pattern for literal values
func buildLiteralPattern[T comparable](values []T) *regexp.Regexp {
	if len(values) == 0 {
		return regexp.MustCompile("^$a") // Never matches
	}

	patterns := make([]string, 0, len(values))
	for _, value := range values {
		patterns = append(patterns, regexp.QuoteMeta(fmt.Sprintf("%v", value)))
	}
	patternStr := "^(" + strings.Join(patterns, "|") + ")$"
	return regexp.MustCompile(patternStr)
}

// newZodLiteralFromDef constructs new ZodLiteral from definition
func newZodLiteralFromDef[T comparable](def *ZodLiteralDef[T]) *ZodLiteral[T, T] {
	values := make(map[T]struct{}, len(def.Values))
	for _, value := range def.Values {
		values[value] = struct{}{}
	}

	// Convert typed values to map[any]struct{} for ZodTypeInternals.Values
	anyValues := make(map[any]struct{}, len(values))
	for value := range values {
		anyValues[value] = struct{}{}
	}

	internals := &ZodLiteralInternals[T]{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Values: anyValues, // Set the Values field for discriminator extraction
			Bag:    make(map[string]any),
		},
		Def:     def,
		Values:  values,
		Pattern: buildLiteralPattern(def.Values),
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		literalDef := &ZodLiteralDef[T]{
			ZodTypeDef: *newDef,
			Values:     def.Values,
		}
		return any(newZodLiteralFromDef[T](literalDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodLiteral[T, T]{internals: internals}
}

// =============================================================================
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// Literal creates a single value literal schema with value constraint type
func Literal[T comparable](value T, params ...any) *ZodLiteral[T, T] {
	return LiteralTyped[T]([]T{value}, params...)
}

// LiteralPtr creates a single value literal schema with pointer constraint type
func LiteralPtr[T comparable](value T, params ...any) *ZodLiteral[T, *T] {
	return Literal(value, params...).Nilable()
}

// LiteralOf creates a multi-value literal schema with value constraint type
func LiteralOf[T comparable](values []T, params ...any) *ZodLiteral[T, T] {
	return LiteralTyped[T](values, params...)
}

// LiteralPtrOf creates a multi-value literal schema with pointer constraint type
func LiteralPtrOf[T comparable](values []T, params ...any) *ZodLiteral[T, *T] {
	return LiteralTyped[T](values, params...).Nilable()
}

// LiteralTyped is the generic constructor for literal schemas - returns value constraint type
func LiteralTyped[T comparable](values []T, paramArgs ...any) *ZodLiteral[T, T] {
	param := utils.GetFirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodLiteralDef[T]{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeLiteral,
			Checks: []core.ZodCheck{},
		},
		Values: values,
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodLiteralFromDef[T](def)
}
