package types

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodUnionDef defines the schema definition for union validation
type ZodUnionDef struct {
	core.ZodTypeDef
	Options []core.ZodSchema // Union member schemas using unified interface
}

// ZodUnionInternals contains the internal state for union schema
type ZodUnionInternals struct {
	core.ZodTypeInternals
	Def     *ZodUnionDef     // Schema definition reference
	Options []core.ZodSchema // Union member schemas for runtime validation
}

// ZodUnion represents a union validation schema with dual generic parameters
// T = base type (any), R = constraint type (any or *any)
type ZodUnion[T any, R any] struct {
	internals *ZodUnionInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodUnion[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodUnion[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodUnion[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates input using direct validation approach with union try-each logic
func (z *ZodUnion[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil values for optional/nilable cases
	if input == nil {
		if z.internals.Nilable || z.internals.Optional {
			var zero R
			return zero, nil
		}
		// If not optional/nilable, fall through to try union members
	}

	// Handle default value
	if input == nil && (z.internals.DefaultValue != nil || z.internals.DefaultFunc != nil) {
		if z.internals.DefaultFunc != nil {
			input = z.internals.DefaultFunc()
		} else {
			input = z.internals.DefaultValue
		}
	}

	// Try each union member schema but capture the first successful match
	var (
		firstSuccess any
		successFound bool
		allErrors    []error
	)

	inputType := reflect.TypeOf(input)

	for i, option := range z.internals.Options {
		if option == nil {
			continue // Skip nil schemas gracefully
		}

		if result, err := option.ParseAny(input, parseCtx); err == nil {
			// Apply any custom checks on the union itself
			if len(z.internals.Checks) > 0 {
				transformedResult, validationErr := engine.ApplyChecks[any](result, z.internals.Checks, parseCtx)
				if validationErr != nil {
					// Treat failed check as parse failure, collect error and continue
					allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, validationErr))
					continue
				}
				result = transformedResult
			}

			// Prefer the schema whose result type matches the original input type
			if inputType != nil && reflect.TypeOf(result) == inputType {
				return convertToUnionConstraintType[T, R](result), nil
			}

			if !successFound {
				firstSuccess = result
				successFound = true
			}
		} else {
			// Collect error for reporting
			allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
		}
	}

	// If we had at least one successful parse, return the first success
	if successFound {
		return convertToUnionConstraintType[T, R](firstSuccess), nil
	}

	// No union member matched - try prefault if available
	if z.internals.PrefaultValue != nil || z.internals.PrefaultFunc != nil {
		var prefaultValue any
		if z.internals.PrefaultFunc != nil {
			prefaultValue = z.internals.PrefaultFunc()
		} else {
			prefaultValue = z.internals.PrefaultValue
		}
		// Try to parse the prefault value
		return z.Parse(prefaultValue, parseCtx)
	}

	// No union member matched and no prefault - create appropriate error
	if len(allErrors) == 0 {
		return *new(R), fmt.Errorf("no union options provided")
	}

	// Create union-specific error
	return *new(R), fmt.Errorf("no union member matched. Tried %d options: %v", len(allErrors), allErrors)
}

// MustParse validates the input value and panics on failure
func (z *ZodUnion[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodUnion[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional union schema that returns pointer constraint
func (z *ZodUnion[T, R]) Optional() *ZodUnion[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint
func (z *ZodUnion[T, R]) Nilable() *ZodUnion[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodUnion[T, R]) Nullish() *ZodUnion[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil value (T).
func (z *ZodUnion[T, R]) NonOptional() *ZodUnion[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
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

// Default preserves current constraint type R
func (z *ZodUnion[T, R]) Default(v T) *ZodUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodUnion[T, R]) DefaultFunc(fn func() T) *ZodUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodUnion[T, R]) Prefault(v T) *ZodUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc keeps current generic type R.
func (z *ZodUnion[T, R]) PrefaultFunc(fn func() T) *ZodUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this union schema.
func (z *ZodUnion[T, R]) Meta(meta core.GlobalMeta) *ZodUnion[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Options returns all union member schemas
func (z *ZodUnion[T, R]) Options() []core.ZodSchema {
	result := make([]core.ZodSchema, len(z.internals.Options))
	copy(result, z.internals.Options)
	return result
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation pipeline using WrapFn pattern
func (z *ZodUnion[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		unionValue := extractUnionValue[T, R](input)
		return fn(unionValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates validation pipeline to another schema using WrapFn pattern
func (z *ZodUnion[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		unionValue := extractUnionValue[T, R](input)
		return target.Parse(unionValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with constraint type R
func (z *ZodUnion[T, R]) Refine(fn func(R) bool, params ...any) *ZodUnion[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToUnionConstraintValue[T, R](v); ok {
			return fn(constraintValue)
		}
		return false
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodUnion[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodUnion[T, R] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}

	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
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

// CloneFrom copies configuration from another schema
func (z *ZodUnion[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodUnion[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToUnionConstraintType converts a base type T to constraint type R
func convertToUnionConstraintType[T any, R any](value any) R {
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

// extractUnionValue extracts base type T from constraint type R
func extractUnionValue[T any, R any](value R) T {
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

// convertToUnionConstraintValue converts any value to constraint type R if possible
func convertToUnionConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok {
		return r, true
	}

	// Handle pointer conversion for union types
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

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodUnionFromDef constructs new ZodUnion from definition
func newZodUnionFromDef[T any, R any](def *ZodUnionDef) *ZodUnion[T, R] {
	internals := &ZodUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Options:          def.Options,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		unionDef := &ZodUnionDef{
			ZodTypeDef: *newDef,
			Options:    def.Options,
		}
		return any(newZodUnionFromDef[T, R](unionDef)).(core.ZodType[any])
	}

	schema := &ZodUnion[T, R]{internals: internals}

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

// Union creates union schema that accepts one of multiple types - returns value constraint
func Union(options []any, args ...any) *ZodUnion[any, any] {
	return UnionTyped[any, any](options, args...)
}

// UnionPtr creates union schema that accepts one of multiple types - returns pointer constraint
func UnionPtr(options []any, args ...any) *ZodUnion[any, *any] {
	return UnionTyped[any, *any](options, args...)
}

// UnionOf creates union schema from variadic arguments - returns value constraint
func UnionOf(schemas ...any) *ZodUnion[any, any] {
	return Union(schemas)
}

// UnionOfPtr creates union schema from variadic arguments - returns pointer constraint
func UnionOfPtr(schemas ...any) *ZodUnion[any, *any] {
	return UnionPtr(schemas)
}

// UnionTyped creates typed union schema with generic constraints
func UnionTyped[T any, R any](options []any, args ...any) *ZodUnion[T, R] {
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	// Convert all options to ZodSchema using direct type assertion, skip nils gracefully
	wrappedOptions := make([]core.ZodSchema, len(options))
	for i, option := range options {
		if option == nil {
			wrappedOptions[i] = nil
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

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	unionSchema := newZodUnionFromDef[T, R](def)

	// Add a minimal check to trigger union validation
	alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
	unionSchema.internals.ZodTypeInternals.AddCheck(alwaysPassCheck)

	return unionSchema
}
