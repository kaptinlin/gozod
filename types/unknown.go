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

// ZodUnknownDef defines the configuration for unknown value validation
type ZodUnknownDef struct {
	core.ZodTypeDef
}

// ZodUnknownInternals contains unknown validator internal state
type ZodUnknownInternals struct {
	core.ZodTypeInternals
	Def *ZodUnknownDef // Schema definition
}

// ZodUnknown represents an unknown-value validation schema with dual generic parameters
// T = base type (any), R = constraint type (any or *any)
// Unknown accepts any value but provides validation and modifier support
type ZodUnknown[T any, R any] struct {
	internals *ZodUnknownInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodUnknown[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodUnknown[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodUnknown[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse returns the input value as-is with full modifier and validation support using direct validation
// Unknown type has special behavior: it accepts nil by default (unlike other types)
func (z *ZodUnknown[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil input
	if input == nil {
		if z.internals.NonOptional {
			return *new(R), engine.CreateNonOptionalError(parseCtx)
		}
		// Check if nil is allowed (optional/nilable)
		if z.internals.Optional || z.internals.Nilable {
			var zero R
			return zero, nil
		}

		// Try default values
		if z.internals.ZodTypeInternals.DefaultValue != nil {
			return z.Parse(z.internals.ZodTypeInternals.DefaultValue, parseCtx)
		}
		if z.internals.ZodTypeInternals.DefaultFunc != nil {
			defaultValue := z.internals.ZodTypeInternals.DefaultFunc()
			return z.Parse(defaultValue, parseCtx)
		}

		// SPECIAL: Unknown type accepts nil by default (this is what makes it different)
		var zero R
		return zero, nil
	}

	// For non-nil values, use direct validation
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
		return convertToUnknownConstraintType[T, R](result), nil
	}

	// Convert result to constraint type R and return
	return convertToUnknownConstraintType[T, R](input), nil
}

// MustParse is the variant that panics on error
func (z *ZodUnknown[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodUnknown[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional makes the unknown type optional and returns pointer constraint
func (z *ZodUnknown[T, R]) Optional() *ZodUnknown[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable makes the unknown type nilable and returns pointer constraint
func (z *ZodUnknown[T, R]) Nilable() *ZodUnknown[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodUnknown[T, R]) Nullish() *ZodUnknown[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional makes the unknown type non-optional and returns base type constraint
func (z *ZodUnknown[T, R]) NonOptional() *ZodUnknown[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodUnknown[T, T]{
		internals: &ZodUnknownInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
		},
	}
}

// Default preserves current constraint type R
func (z *ZodUnknown[T, R]) Default(v T) *ZodUnknown[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodUnknown[T, R]) DefaultFunc(fn func() T) *ZodUnknown[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodUnknown[T, R]) Prefault(v T) *ZodUnknown[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodUnknown[T, R]) PrefaultFunc(fn func() T) *ZodUnknown[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this unknown schema.
func (z *ZodUnknown[T, R]) Meta(meta core.GlobalMeta) *ZodUnknown[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Overwrite applies data transformation while preserving type structure
// This allows for data cleaning, normalization, and preprocessing operations
// Unknown type's Overwrite is particularly useful for sanitizing unpredictable data
func (z *ZodUnknown[T, R]) Overwrite(transform func(T) T, params ...any) *ZodUnknown[T, R] {
	// Create transformation function that wraps the user's transform function
	transformFunc := checks.NewZodCheckOverwrite(func(value any) any {
		// Convert any to T, apply transformation, and return as any
		return convertToUnknownType[T, R](transform(extractUnknownValue[T, R](convertToUnknownConstraintType[T, R](value))))
	}, params...)

	// Clone current internals and add the overwrite check
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(transformFunc)

	return z.withInternals(newInternals)
}

// Refine applies type-safe validation with constraint type R
func (z *ZodUnknown[T, R]) Refine(fn func(R) bool, args ...any) *ZodUnknown[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToUnknownConstraintValue[T, R](v); ok {
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
func (z *ZodUnknown[T, R]) RefineAny(fn func(any) bool, args ...any) *ZodUnknown[T, R] {
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

// Transform creates type-safe transformation pipeline using WrapFn pattern
func (z *ZodUnknown[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		unknownValue := extractUnknownValue[T, R](input)
		return fn(unknownValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates validation pipeline to another schema using WrapFn pattern
func (z *ZodUnknown[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		return target.Parse(input, ctx)
	}
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodUnknown[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodUnknown[T, *T] {
	return &ZodUnknown[T, *T]{internals: &ZodUnknownInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

func (z *ZodUnknown[T, R]) withInternals(in *core.ZodTypeInternals) *ZodUnknown[T, R] {
	return &ZodUnknown[T, R]{internals: &ZodUnknownInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodUnknown[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodUnknown[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToUnknownConstraintType converts a base type T to constraint type R
func convertToUnknownConstraintType[T any, R any](value any) R {
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

// extractUnknownValue extracts base type T from constraint type R
func extractUnknownValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *any:
		if v != nil {
			return any(*v).(T)
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToUnknownConstraintValue converts any value to constraint type R if possible
func convertToUnknownConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok {
		return r, true
	}

	// Handle pointer conversion for unknown types
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

// convertToUnknownType helper function for Overwrite transformations
// Converts transformed result back to any type for further processing in Unknown validation
func convertToUnknownType[T any, R any](value T) any {
	return any(value)
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodUnknownFromDef constructs new ZodUnknown from definition
func newZodUnknownFromDef[T any, R any](def *ZodUnknownDef) *ZodUnknown[T, R] {
	internals := &ZodUnknownInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		unknownDef := &ZodUnknownDef{
			ZodTypeDef: *newDef,
		}
		return any(newZodUnknownFromDef[T, R](unknownDef)).(core.ZodType[any])
	}

	schema := &ZodUnknown[T, R]{internals: internals}

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

// Unknown creates unknown schema that accepts any value - returns value constraint
func Unknown(params ...any) *ZodUnknown[any, any] {
	return UnknownTyped[any, any](params...)
}

// UnknownPtr creates unknown schema that accepts any value - returns pointer constraint
func UnknownPtr(params ...any) *ZodUnknown[any, *any] {
	return UnknownTyped[any, *any](params...)
}

// UnknownTyped creates typed unknown schema with generic constraints
func UnknownTyped[T any, R any](params ...any) *ZodUnknown[T, R] {
	param := utils.GetFirstParam(params...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodUnknownDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeUnknown,
			Checks: []core.ZodCheck{},
		},
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodUnknownFromDef[T, R](def)
}
