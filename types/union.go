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

// Parse validates input using engine.ParseComplex for unified Default/Prefault handling
func (z *ZodUnion[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Use engine.ParseComplex for unified Default/Prefault handling
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
		return *new(R), err
	}
	return convertToUnionConstraintType[T, R](result), nil
}

// extractUnionType extracts the union type from input
func (z *ZodUnion[T, R]) extractUnionType(input any) (any, bool) {
	return input, true
}

// extractUnionPtr extracts pointer from union input
func (z *ZodUnion[T, R]) extractUnionPtr(input any) (*any, bool) {
	if input == nil {
		return nil, true
	}
	return &input, true
}

// validateUnionValue validates the union value using try-each logic
func (z *ZodUnion[T, R]) validateUnionValue(input any, checks []core.ZodCheck, parseCtx *core.ParseContext) (any, error) {
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
			if len(checks) > 0 {
				transformedResult, validationErr := engine.ApplyChecks[any](result, checks, parseCtx)
				if validationErr != nil {
					// Treat failed check as parse failure, collect error and continue
					allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, validationErr))
					continue
				}
				result = transformedResult
			}

			// Prefer the schema whose result type matches the original input type
			if inputType != nil && reflect.TypeOf(result) == inputType {
				return result, nil
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
		return firstSuccess, nil
	}

	// No union member matched - create appropriate error
	if len(allErrors) == 0 {
		return nil, issues.CreateInvalidSchemaError("no union options provided", input, parseCtx)
	}

	// Create union-specific error
	return nil, issues.CreateInvalidUnionError(allErrors, input, parseCtx)
}

// MustParse validates the input value and panics on failure
func (z *ZodUnion[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety and enhanced performance
// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodUnion[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
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

	result, err := engine.ParseComplexStrict[any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeUnion,
		z.extractUnionType,
		z.extractUnionPtr,
		z.validateUnionValue,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}

// MustStrictParse validates input with compile-time type safety and panics on failure
func (z *ZodUnion[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
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

	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message
	}

	check := checks.NewCustom[any](wrapper, errorMessage)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodUnion[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodUnion[T, R] {
	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message
	}

	check := checks.NewCustom[any](fn, errorMessage)
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
	// Handle nil value
	if value == nil {
		// Get the type of R to determine if it's a pointer type
		rType := reflect.TypeOf((*R)(nil)).Elem()
		if rType.Kind() == reflect.Ptr {
			// R is a pointer type, return nil pointer
			return any((*any)(nil)).(R)
		}
		return *new(R)
	}

	// Get the type of R to determine if it's a pointer type
	rType := reflect.TypeOf((*R)(nil)).Elem()

	// Check if R is *any (pointer to interface{})
	// We need to check if R is a pointer to interface{}
	if rType.Kind() == reflect.Ptr {
		// R is some kind of pointer type
		// Check if value is already a pointer
		if reflect.TypeOf(value).Kind() == reflect.Ptr {
			return any(value).(R)
		}
		// For Optional schemas, if the value comes from Default/Prefault processing
		// and the original input was nil, we should return nil instead of wrapping
		// This is handled by the engine's processModifiersInternal logic
		// Convert value to pointer
		valueCopy := value
		return any(&valueCopy).(R)
	} else {
		// R is not a pointer type (like any, string, int, etc.)
		// Check if value is a pointer and R is not a pointer type
		// This handles Prefault values that come as *any but need to be converted to any
		if reflect.TypeOf(value).Kind() == reflect.Ptr {
			// Dereference the pointer
			valueReflect := reflect.ValueOf(value)
			if !valueReflect.IsNil() {
				dereferencedValue := valueReflect.Elem().Interface()
				return any(dereferencedValue).(R)
			}
			var zero R
			return zero
		}

		// For non-pointer values, try direct conversion (for Default values)
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
