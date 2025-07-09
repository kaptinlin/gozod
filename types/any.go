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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodAny[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse returns the input value as-is with modifier support using direct validation
func (z *ZodAny[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil input
	if input == nil {
		// NonOptional -> dedicated error
		if z.internals.NonOptional {
			return *new(R), engine.CreateNonOptionalError(parseCtx)
		}
		// Check if nil is allowed (optional/nilable) - this takes precedence
		if z.internals.ZodTypeInternals.Optional || z.internals.ZodTypeInternals.Nilable {
			var zero R
			return zero, nil
		}

		// Try default value
		if z.internals.ZodTypeInternals.DefaultValue != nil {
			return z.Parse(z.internals.ZodTypeInternals.DefaultValue, parseCtx)
		}

		// Try default function
		if z.internals.ZodTypeInternals.DefaultFunc != nil {
			defaultValue := z.internals.ZodTypeInternals.DefaultFunc()
			return z.Parse(defaultValue, parseCtx)
		}

		// For Any type, nil is accepted by default unless explicitly restricted by checks
		if len(z.internals.ZodTypeInternals.Checks) > 0 {
			transformedValue, err := engine.ApplyChecks(input, z.internals.ZodTypeInternals.Checks, parseCtx)
			if err != nil {
				// Try prefault on validation failure
				if z.internals.ZodTypeInternals.PrefaultValue != nil {
					return z.Parse(z.internals.ZodTypeInternals.PrefaultValue, parseCtx)
				}
				if z.internals.ZodTypeInternals.PrefaultFunc != nil {
					prefaultValue := z.internals.ZodTypeInternals.PrefaultFunc()
					return z.Parse(prefaultValue, parseCtx)
				}

				return *new(R), err
			}
			// Use the transformed value from ApplyChecks
			return convertToAnyConstraintType[T, R](transformedValue), nil
		}
		var zero R
		return zero, nil
	}

	// Any type accepts any value - run validation checks if any exist
	if len(z.internals.ZodTypeInternals.Checks) > 0 {
		transformedValue, err := engine.ApplyChecks(input, z.internals.ZodTypeInternals.Checks, parseCtx)
		if err != nil {
			// Try prefault on validation failure
			if z.internals.ZodTypeInternals.PrefaultValue != nil {
				return z.Parse(z.internals.ZodTypeInternals.PrefaultValue, parseCtx)
			}
			if z.internals.ZodTypeInternals.PrefaultFunc != nil {
				prefaultValue := z.internals.ZodTypeInternals.PrefaultFunc()
				return z.Parse(prefaultValue, parseCtx)
			}

			return *new(R), err
		}
		// Use the transformed value from ApplyChecks
		input = transformedValue
	}

	// Apply transform if present
	if z.internals.ZodTypeInternals.Transform != nil {
		refCtx := &core.RefinementContext{ParseContext: parseCtx}
		result, err := z.internals.ZodTypeInternals.Transform(input, refCtx)
		if err != nil {
			return *new(R), err
		}
		return convertToAnyConstraintType[T, R](result), nil
	}

	// Convert result to constraint type R and return
	return convertToAnyConstraintType[T, R](input), nil
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the any type nilable and returns pointer constraint
func (z *ZodAny[T, R]) Nilable() *ZodAny[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodAny[T, R]) Nullish() *ZodAny[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and returns base constraint type T.
func (z *ZodAny[T, R]) NonOptional() *ZodAny[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodAny[T, R]) DefaultFunc(fn func() T) *ZodAny[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodAny[T, R]) Prefault(v T) *ZodAny[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodAny[T, R]) PrefaultFunc(fn func() T) *ZodAny[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var checkParams any
	if normalizedParams.Error != nil {
		checkParams = normalizedParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodAny[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodAny[T, R] {
	// Use unified parameter handling
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	var checkParams any
	if normalizedParams.Error != nil {
		checkParams = normalizedParams
	}

	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
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
		// Return value directly as R
		return any(value).(R)
	}
}

// extractAnyValue extracts base type T from constraint type R
func extractAnyValue[T any, R any](value R) T {
	if v, ok := any(value).(*any); ok {
		if v != nil {
			return any(*v).(T)
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
	if r, ok := any(value).(R); ok {
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
	param := utils.GetFirstParam(params...)
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
