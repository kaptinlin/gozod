package types

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodAnyDef defines the configuration for any value validation
type ZodAnyDef struct {
	core.ZodTypeDef
}

// ZodAnyInternals contains any validator internal state
type ZodAnyInternals struct {
	core.ZodTypeInternals
	Def *ZodAnyDef // Schema definition
}

// ZodAny represents an any-value validation schema with dual generic parameters
// T = base type (any), R = constraint type (any or *any)
type ZodAny[T any, R any] struct {
	internals *ZodAnyInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodAny[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodAny[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodAny[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// validateAnyValue is the Any type validator - accepts any value and applies checks
func validateAnyValue[T any](value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// Any type accepts any value, only need to apply checks
	return engine.ApplyChecks[T](value, checks, ctx)
}

// Parse returns the input value as-is with modifier support using engine.ParsePrimitive
func (z *ZodAny[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	// Create converter adapter for engine.ParsePrimitive
	converter := func(value any, ctx *core.ParseContext, expectedType core.ZodTypeCode) (R, error) {
		return convertToAnyConstraintType[T, R](value), nil
	}

	return engine.ParsePrimitive[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeAny,
		validateAnyValue[T],
		converter,
		ctx...,
	)
}

// StrictParse provides type-safe parsing with compile-time guarantees using engine.ParsePrimitiveStrict
func (z *ZodAny[T, R]) StrictParse(input R, ctx ...*core.ParseContext) (R, error) {
	return engine.ParsePrimitiveStrict[T, R](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeAny,
		validateAnyValue[T],
		ctx...,
	)
}

// MustParse is the variant that panics on error
func (z *ZodAny[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodAny[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional makes the any type optional and returns pointer constraint
func (z *ZodAny[T, R]) Optional() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the any type nilable and returns pointer constraint
func (z *ZodAny[T, R]) Nilable() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodAny[T, R]) Nullish() *ZodAny[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and returns base constraint type T.
func (z *ZodAny[T, R]) NonOptional() *ZodAny[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodAny[T, T]{
		internals: &ZodAnyInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default preserves current constraint type R
func (z *ZodAny[T, R]) Default(v T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodAny[T, R]) DefaultFunc(fn func() T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodAny[T, R]) Prefault(v T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodAny[T, R]) PrefaultFunc(fn func() T) *ZodAny[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this any schema.
func (z *ZodAny[T, R]) Meta(meta core.GlobalMeta) *ZodAny[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodAny[T, R]) Describe(description string) *ZodAny[T, R] {
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
// VALIDATION METHODS
// =============================================================================

// Overwrite applies data transformation while preserving type structure
// This allows for data cleaning, normalization, and preprocessing operations
func (z *ZodAny[T, R]) Overwrite(transform func(T) T, params ...any) *ZodAny[T, R] {
	// Create transformation function that wraps the user's transform function
	transformFunc := checks.NewZodCheckOverwrite(func(value any) any {
		// Convert any to T, apply transformation, and return as any
		return convertToAnyType[T, R](transform(extractAnyValue[T, R](convertToAnyConstraintType[T, R](value))))
	}, params...)

	// Clone current internals and add the overwrite check
	newInternals := z.internals.Clone()
	newInternals.AddCheck(transformFunc)

	return z.withInternals(newInternals)
}

// Refine applies type-safe validation with constraint type R
func (z *ZodAny[T, R]) Refine(fn func(R) bool, args ...any) *ZodAny[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToAnyConstraintValue[T, R](v); ok {
			return fn(constraintValue)
		}
		return false
	}

	// Use unified parameter handling
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var errorMessage any
	if normalizedParams.Error != nil {
		errorMessage = normalizedParams
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodAny[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodAny[T, R] {
	// Use unified parameter handling
	param := utils.FirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var errorMessage any
	if normalizedParams.Error != nil {
		errorMessage = normalizedParams
	}

	check := checks.NewCustom[any](fn, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
func (z *ZodAny[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		anyValue := extractAnyValue[T, R](input)
		return fn(anyValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a pipeline using the WrapFn pattern.
func (z *ZodAny[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		anyValue := extractAnyValue[T, R](input)
		return target.Parse(anyValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodAny[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodAny[T, *T] {
	return &ZodAny[T, *T]{internals: &ZodAnyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodAny[T, R]) withInternals(in *core.ZodTypeInternals) *ZodAny[T, R] {
	return &ZodAny[T, R]{internals: &ZodAnyInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodAny[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodAny[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToAnyConstraintType converts a base type T to constraint type R
func convertToAnyConstraintType[T any, R any](value any) R {
	var zero R
	switch any(zero).(type) {
	case *any:
		// Need to return *any from any
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R)
		}
		return any((*any)(nil)).(R)
	default:
		// Return value directly as R, handle nil case safely
		if value == nil {
			return zero // Return zero value for nil
		}
		return any(value).(R) //nolint:unconvert // Required for generic type constraint conversion
	}
}

// extractAnyValue extracts base type T from constraint type R
func extractAnyValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok {
		if v != nil {
			return any(*v).(T) //nolint:unconvert // Required for generic type constraint conversion
		}
		var zero T
		return zero
	}
	return any(value).(T)
}

// convertToAnyConstraintValue converts any value to constraint type R if possible
func convertToAnyConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	// Handle pointer conversion for any types
	if _, ok := any(zero).(*any); ok {
		// Need to convert any to *any
		if value != nil {
			valueCopy := value
			return any(&valueCopy).(R), true
		}
		return any((*any)(nil)).(R), true
	}

	return zero, false
}

// convertToAnyType helper function for Overwrite transformations
// Converts transformed result back to any type for further processing
func convertToAnyType[T any, R any](value T) any {
	return any(value)
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodAnyFromDef constructs new ZodAny from definition
func newZodAnyFromDef[T any, R any](def *ZodAnyDef) *ZodAny[T, R] {
	internals := &ZodAnyInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
	}

	// Any type should default to accepting nil values like in Zod v4
	internals.Nilable = true

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		anyDef := &ZodAnyDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodAnyFromDef[T, R](anyDef)).(core.ZodType[any])
	}

	schema := &ZodAny[T, R]{internals: internals}

	// Set error if provided
	if def.Error != nil {
		internals.Error = def.Error
	}

	// Set checks if provided
	if len(def.Checks) > 0 {
		for _, check := range def.Checks {
			internals.AddCheck(check)
		}
	}

	return schema
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// Any creates any schema that accepts any value - returns value constraint
func Any(params ...any) *ZodAny[any, any] {
	return AnyTyped[any, any](params...)
}

// AnyPtr creates any schema that accepts any value - returns pointer constraint
func AnyPtr(params ...any) *ZodAny[any, *any] {
	return AnyTyped[any, *any](params...)
}

// AnyTyped creates typed any schema with generic constraints
func AnyTyped[T any, R any](params ...any) *ZodAny[T, R] {
	param := utils.FirstParam(params...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodAnyDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeAny,
			Checks: []core.ZodCheck{},
		},
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodAnyFromDef[T, R](def)
}
