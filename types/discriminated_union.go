package types

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodDiscriminatedUnionDef defines the configuration for discriminated union validation
type ZodDiscriminatedUnionDef struct {
	core.ZodTypeDef
	Discriminator string           // The discriminator field name
	Options       []core.ZodSchema // Union member schemas using unified interface
}

// ZodDiscriminatedUnionInternals contains discriminated union validator internal state
type ZodDiscriminatedUnionInternals struct {
	core.ZodTypeInternals
	Def           *ZodDiscriminatedUnionDef // Schema definition reference
	Discriminator string                    // The discriminator field name
	Options       []core.ZodSchema          // Union member schemas for runtime validation
	DiscMap       map[any]core.ZodSchema    // Discriminator value to schema mapping
}

// ZodDiscriminatedUnion represents a discriminated union validation schema with dual generic parameters
// T = base type (any), R = constraint type (any or *any)
type ZodDiscriminatedUnion[T any, R any] struct {
	internals *ZodDiscriminatedUnionInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodDiscriminatedUnion[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodDiscriminatedUnion[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodDiscriminatedUnion[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates input using discriminated union logic with direct validation approach
func (z *ZodDiscriminatedUnion[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	// Check for delayed construction error
	if constructionError, exists := z.internals.Bag["construction_error"]; exists {
		if errMsg, ok := constructionError.(string); ok {
			return *new(R), fmt.Errorf("%s", errMsg)
		}
	}

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
		// If not optional/nilable, fall through to try discriminated union members
	}

	// Handle default value
	if input == nil && (z.internals.DefaultValue != nil || z.internals.DefaultFunc != nil) {
		if z.internals.DefaultFunc != nil {
			input = z.internals.DefaultFunc()
		} else {
			input = z.internals.DefaultValue
		}
	}

	// Ensure input is a map
	inputMap, ok := input.(map[string]any)
	if !ok {
		// Try prefault if available
		if z.internals.PrefaultValue != nil || z.internals.PrefaultFunc != nil {
			var prefaultValue any
			if z.internals.PrefaultFunc != nil {
				prefaultValue = z.internals.PrefaultFunc()
			} else {
				prefaultValue = z.internals.PrefaultValue
			}
			return z.Parse(prefaultValue, parseCtx)
		}
		return *new(R), fmt.Errorf("expected map[string]any, got %T", input)
	}

	// Get the value of the discriminator key
	discriminatorValue, exists := inputMap[z.internals.Discriminator]
	if !exists {
		return *new(R), fmt.Errorf("missing discriminator field '%s'", z.internals.Discriminator)
	}

	// Find the corresponding schema based on the discriminator value
	targetSchema, found := z.internals.DiscMap[discriminatorValue]
	var result any
	var err error
	if found {
		// Fast path: discriminator table hit
		result, err = targetSchema.ParseAny(inputMap, parseCtx)
	} else {
		// Fallback: sequentially try all options (for schemas without explicit discriminator literals)
		var allErrs []error
		for _, option := range z.internals.Options {
			if option == nil {
				continue
			}
			if res, e := option.ParseAny(inputMap, parseCtx); e == nil {
				result = res
				err = nil
				break
			} else {
				allErrs = append(allErrs, e)
			}
		}
		if result == nil {
			return *new(R), fmt.Errorf("no schema matched discriminator value: %v, errors: %v", discriminatorValue, allErrs)
		}
	}

	if err != nil {
		return *new(R), err
	}

	// Apply any custom checks on the discriminated union itself and capture transformed result
	if len(z.internals.Checks) > 0 {
		transformedResult, validationErr := engine.ApplyChecks[any](result, z.internals.Checks, parseCtx)
		if validationErr != nil {
			return *new(R), validationErr
		}
		result = transformedResult
	}

	// Convert result to constraint type R
	return convertToDiscriminatedUnionConstraintType[T, R](result), nil
}

// MustParse validates the input value and panics on failure
func (z *ZodDiscriminatedUnion[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodDiscriminatedUnion[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional discriminated union schema that returns pointer constraint
func (z *ZodDiscriminatedUnion[T, R]) Optional() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint
func (z *ZodDiscriminatedUnion[T, R]) Nilable() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodDiscriminatedUnion[T, R]) Nullish() *ZodDiscriminatedUnion[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Default preserves current constraint type R
func (z *ZodDiscriminatedUnion[T, R]) Default(v T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodDiscriminatedUnion[T, R]) DefaultFunc(fn func() T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodDiscriminatedUnion[T, R]) Prefault(v T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodDiscriminatedUnion[T, R]) PrefaultFunc(fn func() T) *ZodDiscriminatedUnion[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Discriminator returns the discriminator field name
func (z *ZodDiscriminatedUnion[T, R]) Discriminator() string {
	return z.internals.Discriminator
}

// Options returns all union member schemas
func (z *ZodDiscriminatedUnion[T, R]) Options() []core.ZodSchema {
	result := make([]core.ZodSchema, len(z.internals.Options))
	copy(result, z.internals.Options)
	return result
}

// DiscriminatorMap returns the discriminator value to schema mapping
func (z *ZodDiscriminatedUnion[T, R]) DiscriminatorMap() map[any]core.ZodSchema {
	result := make(map[any]core.ZodSchema, len(z.internals.DiscMap))
	for k, v := range z.internals.DiscMap {
		result[k] = v
	}
	return result
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates type-safe transformation pipeline using WrapFn pattern
func (z *ZodDiscriminatedUnion[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractDiscriminatedUnionValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates validation pipeline to another schema using WrapFn pattern
func (z *ZodDiscriminatedUnion[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractDiscriminatedUnionValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with constraint type R
func (z *ZodDiscriminatedUnion[T, R]) Refine(fn func(R) bool, params ...any) *ZodDiscriminatedUnion[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToDiscriminatedUnionConstraintValue[T, R](v); ok {
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
func (z *ZodDiscriminatedUnion[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodDiscriminatedUnion[T, R] {
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

// CloneFrom copies configuration from another schema
func (z *ZodDiscriminatedUnion[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodDiscriminatedUnion[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// CONVERSION HELPERS
// =============================================================================

// convertToDiscriminatedUnionConstraintType converts a base type T to constraint type R.
func convertToDiscriminatedUnionConstraintType[T any, R any](value any) R {
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

// extractDiscriminatedUnionValue extracts base type T from constraint type R
func extractDiscriminatedUnionValue[T any, R any](value R) T {
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

// convertToDiscriminatedUnionConstraintValue converts any value to constraint type R if possible
func convertToDiscriminatedUnionConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok {
		return r, true
	}

	// Handle pointer conversion for discriminated union types
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
// DISCRIMINATOR VALUE EXTRACTION
// =============================================================================

// getDiscriminatorValues extracts discriminator values from a schema
func getDiscriminatorValues(schema core.ZodSchema, discriminatorField string) ([]any, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	// Get schema internals
	internals := schema.GetInternals()
	if internals == nil {
		return nil, fmt.Errorf("schema internals is nil")
	}

	// Try to extract discriminator values from internals
	values := extractDiscriminatorFromInternals(internals, discriminatorField)
	if len(values) > 0 {
		return values, nil
	}

	// Try to get discriminator values from the schema directly using type assertion
	return getDiscriminatorValuesFromAnySchema(schema, discriminatorField)
}

// getDiscriminatorValuesFromAnySchema extracts discriminator values from any schema using type assertion
func getDiscriminatorValuesFromAnySchema(schema any, discriminatorField string) ([]any, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	// Try to get internals directly
	if internalsSchema, ok := schema.(interface {
		GetInternals() *core.ZodTypeInternals
	}); ok {
		if internals := internalsSchema.GetInternals(); internals != nil {
			values := extractDiscriminatorFromInternals(internals, discriminatorField)
			if len(values) > 0 {
				return values, nil
			}
		}
	}

	// Try to check if it has a Shape method (for object/struct schemas)
	if objectSchema, ok := schema.(interface {
		Shape() core.ObjectSchema
	}); ok {
		shape := objectSchema.Shape()
		if fieldSchema, exists := shape[discriminatorField]; exists {
			if fieldInternals := fieldSchema.GetInternals(); fieldInternals != nil {
				values := extractDiscriminatorFromInternals(fieldInternals, discriminatorField)
				if len(values) > 0 {
					return values, nil
				}
			}
		}
	}

	// Try to check if it has a Shape method for struct schemas
	if structSchema, ok := schema.(interface {
		Shape() core.StructSchema
	}); ok {
		shape := structSchema.Shape()
		if fieldSchema, exists := shape[discriminatorField]; exists {
			if fieldInternals := fieldSchema.GetInternals(); fieldInternals != nil {
				values := extractDiscriminatorFromInternals(fieldInternals, discriminatorField)
				if len(values) > 0 {
					return values, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no discriminator values found for field '%s' in schema", discriminatorField)
}

// extractDiscriminatorFromInternals extracts discriminator values from schema internals
func extractDiscriminatorFromInternals(internals *core.ZodTypeInternals, discriminatorField string) []any {
	if internals == nil {
		return nil
	}

	var values []any

	// Check for Values field (for Enum/Literal types)
	if len(internals.Values) > 0 {
		for value := range internals.Values {
			values = append(values, value)
		}
		return values
	}

	// Check for Bag data (alternative storage)
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

	// Check type-specific patterns
	//nolint:exhaustive // Only interested in Literal and Enum types here
	switch internals.Type {
	case core.ZodTypeLiteral:
		// For literal types, check if there's a literal value
		if literalValue, exists := internals.Bag["literal"]; exists {
			values = append(values, literalValue)
		}
	case core.ZodTypeEnum:
		// For enum types, values should be in Values field (already checked above)
		// Fallback: check if there are enum values in Bag
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

// buildDiscriminatorMap builds the discriminator value to schema mapping with improved error handling
func buildDiscriminatorMap(discriminator string, options []core.ZodSchema) (map[any]core.ZodSchema, error) {
	discMap := make(map[any]core.ZodSchema)
	var allErrors []error

	for i, option := range options {
		if option == nil {
			allErrors = append(allErrors, fmt.Errorf("option %d is nil", i))
			continue
		}

		discriminatorValues, err := getDiscriminatorValues(option, discriminator)
		if err != nil {
			// Following TypeScript Zod's approach: require discriminator field in all options
			allErrors = append(allErrors, fmt.Errorf("option %d: %w", i, err))
			continue
		}

		// Add values to discriminator map
		for _, value := range discriminatorValues {
			if _, exists := discMap[value]; exists {
				return nil, fmt.Errorf("duplicate discriminator value: %v", value)
			}
			discMap[value] = option
		}
	}

	// If we have errors and no valid discriminator values, return error
	if len(discMap) == 0 && len(allErrors) > 0 {
		return nil, fmt.Errorf("failed to build discriminator map: %v", allErrors)
	}

	if len(discMap) == 0 {
		return nil, fmt.Errorf("no valid discriminator values found for field: %s", discriminator)
	}

	return discMap, nil
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodDiscriminatedUnionFromDef constructs new ZodDiscriminatedUnion from definition
// Uses delayed error strategy: construction errors are deferred to parse-time for graceful handling
func newZodDiscriminatedUnionFromDef[T any, R any](def *ZodDiscriminatedUnionDef) *ZodDiscriminatedUnion[T, R] {
	// Build discriminator map with graceful error handling
	discMap, err := buildDiscriminatorMap(def.Discriminator, def.Options)

	internals := &ZodDiscriminatedUnionInternals{
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Discriminator:    def.Discriminator,
		Options:          def.Options,
		DiscMap:          discMap,
	}

	// Store construction error for delayed reporting
	if err != nil {
		internals.Bag["construction_error"] = fmt.Sprintf("DiscriminatedUnion construction error: %v", err)
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		discriminatedUnionDef := &ZodDiscriminatedUnionDef{
			ZodTypeDef:    *newDef,
			Discriminator: def.Discriminator,
			Options:       def.Options,
		}
		return any(newZodDiscriminatedUnionFromDef[T, R](discriminatedUnionDef)).(core.ZodType[any])
	}

	schema := &ZodDiscriminatedUnion[T, R]{internals: internals}

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

// DiscriminatedUnion creates discriminated union schema that accepts one of the specified object types - returns value constraint
func DiscriminatedUnion(discriminator string, options []any, args ...any) *ZodDiscriminatedUnion[any, any] {
	return DiscriminatedUnionTyped[any, any](discriminator, options, args...)
}

// DiscriminatedUnionPtr creates discriminated union schema that accepts one of the specified object types - returns pointer constraint
func DiscriminatedUnionPtr(discriminator string, options []any, args ...any) *ZodDiscriminatedUnion[any, *any] {
	return DiscriminatedUnionTyped[any, *any](discriminator, options, args...)
}

// DiscriminatedUnionTyped creates typed discriminated union schema with generic constraints
func DiscriminatedUnionTyped[T any, R any](discriminator string, options []any, args ...any) *ZodDiscriminatedUnion[T, R] {
	param := utils.GetFirstParam(args...)
	normalizedParams := utils.NormalizeParams(param)

	// Convert all options to core.ZodSchema using direct type assertion
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

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	return newZodDiscriminatedUnionFromDef[T, R](def)
}
