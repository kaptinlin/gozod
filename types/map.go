package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodMapDef defines the schema definition for map validation
type ZodMapDef struct {
	core.ZodTypeDef
	KeyType   any // The key schema (type-erased for flexibility)
	ValueType any // The value schema (type-erased for flexibility)
}

// ZodMapInternals contains the internal state for map schema
type ZodMapInternals struct {
	core.ZodTypeInternals
	Def       *ZodMapDef // Schema definition reference
	KeyType   any        // Key schema for runtime validation
	ValueType any        // Value schema for runtime validation
}

// ZodMap represents a type-safe map validation schema with dual generic parameters
// T = base type (map[any]any), R = constraint type (map[any]any or *map[any]any)
type ZodMap[T any, R any] struct {
	internals *ZodMapInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodMap[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodMap[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodMap[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using direct validation approach
// Parse validates input using map-specific parsing logic with engine.ParseComplex
func (z *ZodMap[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex[map[any]any](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMap,
		z.extractMapForEngine,
		z.extractMapPtrForEngine,
		z.validateMapForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	// Convert result to constraint type R
	if mapValue, ok := result.(map[any]any); ok {
		return convertMapFromGeneric[T, R](mapValue), nil
	}

	// Handle pointer to map[any]any
	if mapPtr, ok := result.(*map[any]any); ok {
		if mapPtr == nil {
			var zero R
			return zero, nil
		}
		return convertMapFromGeneric[T, R](*mapPtr), nil
	}

	// Handle nil result for optional/nilable schemas
	if result == nil {
		var zero R
		return zero, nil
	}

	// This should not happen in well-formed schemas
	var zero R
	parseCtx := core.NewParseContext()
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	}
	return zero, issues.CreateTypeConversionError(
		fmt.Sprintf("%T", result),
		"map",
		input,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodMap[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety
func (z *ZodMap[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
	constraintInput, ok := convertToMapConstraintValue[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"map constraint type",
			any(input),
			ctx[0],
		)
	}

	result, err := engine.ParseComplexStrict[map[any]any, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeMap,
		z.extractMapForEngine,
		z.extractMapPtrForEngine,
		z.validateMapForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}

// MustStrictParse validates input with compile-time type safety and panics on failure
func (z *ZodMap[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodMap[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// KeyType returns the key schema for this map.
func (z *ZodMap[T, R]) KeyType() any {
	return z.internals.KeyType
}

// ValueType returns the value schema for this map.
func (z *ZodMap[T, R]) ValueType() any {
	return z.internals.ValueType
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional map schema that returns pointer constraint
func (z *ZodMap[T, R]) Optional() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint
func (z *ZodMap[T, R]) Nilable() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodMap[T, R]) Nullish() *ZodMap[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag to require map value and returns base constraint type.
func (z *ZodMap[T, R]) NonOptional() *ZodMap[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodMap[T, T]{
		internals: &ZodMapInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			KeyType:          z.internals.KeyType,
			ValueType:        z.internals.ValueType,
		},
	}
}

// Default preserves current constraint type R
func (z *ZodMap[T, R]) Default(v T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodMap[T, R]) DefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodMap[T, R]) Prefault(v T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodMap[T, R]) PrefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this map schema.
func (z *ZodMap[T, R]) Meta(meta core.GlobalMeta) *ZodMap[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
// TypeScript Zod v4 equivalent: schema.describe(description)
func (z *ZodMap[T, R]) Describe(description string) *ZodMap[T, R] {
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

// Min sets minimum number of entries
func (z *ZodMap[T, R]) Min(minLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}
	check := checks.MinSize(minLen, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of entries
func (z *ZodMap[T, R]) Max(maxLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}
	check := checks.MaxSize(maxLen, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size sets exact number of entries
func (z *ZodMap[T, R]) Size(exactLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var errorMessage any
	if schemaParams.Error != nil {
		errorMessage = schemaParams.Error // Pass the actual error message, not the SchemaParams
	}
	check := checks.Size(exactLen, errorMessage)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// NonEmpty ensures map has at least one entry.
// This is a convenience method equivalent to Min(1).
func (z *ZodMap[T, R]) NonEmpty(params ...any) *ZodMap[T, R] {
	return z.Min(1, params...)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform applies a transformation function using the WrapFn pattern.
// Map types implement direct extraction of T values for transformation.
func (z *ZodMap[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	// WrapFn Pattern: Create wrapper function for type-safe extraction
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		mapValue := extractMapValue[T, R](input) // Use existing extraction logic
		return fn(mapValue, ctx)
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodMap[T, R]) Overwrite(transform func(R) R, params ...any) *ZodMap[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToMapType[T, R](input)
		if !ok {
			// If conversion fails, return original value
			return input
		}

		// Apply transformation directly on constraint type R
		return transform(converted)
	}

	check := checks.NewZodCheckOverwrite(transformAny, params...)
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using the WrapFn pattern.
// Instead of using adapter structures, this creates a target function that handles type conversion.
// Map types implement direct extraction of T values for transformation.
func (z *ZodMap[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	// WrapFn Pattern: Create target function for type conversion and validation
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		// Extract T value from constraint type R
		mapValue := extractMapValue[T, R](input)
		// Apply target schema to the extracted T
		return target.Parse(mapValue, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[R, any](z, target, targetFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with constraint type R
func (z *ZodMap[T, R]) Refine(fn func(R) bool, params ...any) *ZodMap[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Handle nilable case: check if R is a pointer type and v is nil
		var zero R
		switch any(zero).(type) {
		case *map[any]any:
			// For pointer types, allow nil to pass refinement
			if v == nil {
				return true
			}
			// Convert non-nil value to constraint type and call fn
			if constraintValue, ok := convertToMapConstraintValue[T, R](v); ok {
				return fn(constraintValue)
			}
			return false
		default:
			// For non-pointer types, convert value to constraint type R and call fn
			if constraintValue, ok := convertToMapConstraintValue[T, R](v); ok {
				return fn(constraintValue)
			}
			return false
		}
	}

	// Use unified parameter handling
	schemaParams := utils.NormalizeParams(params...)

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
func (z *ZodMap[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodMap[T, R] {
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
// HELPER METHODS
// =============================================================================

func (z *ZodMap[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodMap[T, *T] {
	return &ZodMap[T, *T]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

func (z *ZodMap[T, R]) withInternals(in *core.ZodTypeInternals) *ZodMap[T, R] {
	return &ZodMap[T, R]{internals: &ZodMapInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodMap[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodMap[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertMapFromGeneric safely converts map[any]any to constraint type R
func convertMapFromGeneric[T any, R any](mapValue map[any]any) R {
	// First convert to base type T
	var baseValue T
	switch any(baseValue).(type) {
	case map[string]int:
		// Convert map[any]any to map[string]int
		converted := make(map[string]int)
		for k, v := range mapValue {
			if strKey, ok := k.(string); ok {
				if intVal, ok := v.(int); ok {
					converted[strKey] = intVal
				}
			}
		}
		baseValue = any(converted).(T)
	case map[string]any:
		// Convert map[any]any to map[string]any
		converted := make(map[string]any)
		for k, v := range mapValue {
			if strKey, ok := k.(string); ok {
				converted[strKey] = v
			}
		}
		baseValue = any(converted).(T)
	case map[any]any:
		// Direct assignment for map[any]any
		baseValue = any(mapValue).(T)
	default:
		// For other types, try direct conversion
		baseValue = any(mapValue).(T)
	}

	// Then convert base type to constraint type
	return convertToMapConstraintType[T, R](baseValue)
}

// convertToMapConstraintType converts a base type T to constraint type R
func convertToMapConstraintType[T any, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *map[any]any:
		// Need to return *map[any]any from map[any]any
		if mapVal, ok := any(value).(map[any]any); ok {
			mapCopy := mapVal
			return any(&mapCopy).(R)
		}
		return any((*map[any]any)(nil)).(R)
	default:
		// Return T directly
		return any(value).(R)
	}
}

// extractMapValue extracts the base type T from constraint type R
func extractMapValue[T any, R any](value R) T {
	// Handle direct assignment (when T == R)
	if directValue, ok := any(value).(T); ok {
		return directValue
	}

	// Handle pointer dereferencing
	if ptrValue := reflect.ValueOf(value); ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
		if derefValue, ok := ptrValue.Elem().Interface().(T); ok {
			return derefValue
		}
	}

	// Fallback to zero value
	var zero T
	return zero
}

// convertToMapType converts any value to the map constraint type R with strict type checking
func convertToMapType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Extract map value from input
	var mapValue map[any]any
	var isValid bool

	switch val := v.(type) {
	case map[any]any:
		mapValue, isValid = val, true
	case *map[any]any:
		if val != nil {
			mapValue, isValid = *val, true
		}
	case map[string]any:
		// Convert map[string]any to map[any]any
		mapValue = make(map[any]any)
		for k, v := range val {
			mapValue[k] = v
		}
		isValid = true
	case *map[string]any:
		if val != nil {
			mapValue = make(map[any]any)
			for k, v := range *val {
				mapValue[k] = v
			}
			isValid = true
		}
	default:
		// Try to extract using mapx if available
		return zero, false // Reject all non-map types
	}

	if !isValid {
		return zero, false
	}

	// Convert to target constraint type R
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
		// R is *map[any]any
		if converted, ok := any(&mapValue).(R); ok {
			return converted, true
		}
	} else {
		// R is map[any]any
		if converted, ok := any(mapValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// convertToMapConstraintValue converts any value to constraint type R if possible
func convertToMapConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	// Handle pointer conversion for map types
	if _, ok := any(zero).(*map[any]any); ok {
		// Need to convert map[any]any to *map[any]any
		if mapVal, ok := value.(map[any]any); ok {
			mapCopy := mapVal
			return any(&mapCopy).(R), true
		}
	}

	return zero, false
}

// =============================================================================
// EXTRACTION AND VALIDATION
// =============================================================================

// extractMap converts input to map[any]any
func (z *ZodMap[T, R]) extractMap(value any, ctx *core.ParseContext) (map[any]any, error) {
	// Handle direct map[any]any
	if mapVal, ok := value.(map[any]any); ok {
		return mapVal, nil
	}

	// Handle pointer to map
	if mapPtr, ok := value.(*map[any]any); ok {
		if mapPtr != nil {
			return *mapPtr, nil
		}
		return nil, issues.CreateNonOptionalError(ctx)
	}

	// Try to convert using mapx
	if reflectx.IsMap(value) {
		if converted, err := mapx.ToGeneric(value); err == nil && converted != nil {
			return converted, nil
		}
	}

	return nil, issues.CreateInvalidTypeError(core.ZodTypeMap, value, ctx)
}

// validateMap validates map entries using key and value schemas with multiple error collection (TypeScript Zod v4 behavior adapted for Go maps)
func (z *ZodMap[T, R]) validateMap(value map[any]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
	var collectedIssues []core.ZodRawIssue

	// First apply checks (including Overwrite transformations) to get the transformed value
	transformedValue, err := engine.ApplyChecks[any](value, checks, ctx)
	if err != nil {
		return nil, err
	}

	// Handle potential pointer type from Overwrite transformations
	switch v := transformedValue.(type) {
	case map[any]any:
		value = v
	case *map[any]any:
		if v != nil {
			value = *v
		}
	default:
		// If transformation returned unexpected type, keep original value
		// This handles the case where no transformation was applied
	}

	// Always validate each key-value pair if schemas are provided and collect all errors (TypeScript Zod v4 behavior)
	for key, val := range value {
		// Validate key if key schema is provided
		if z.internals.KeyType != nil {
			if err := z.validateValueDirect(key, z.internals.KeyType, ctx); err != nil {
				// Collect key validation errors with path prefix (TypeScript Zod v4 behavior adapted for Go maps)
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					// Propagate all issues from key validation with path prefix
					for _, keyIssue := range zodErr.Issues {
						// Create raw issue preserving original code and essential properties
						rawIssue := core.ZodRawIssue{
							Code:       keyIssue.Code,
							Message:    keyIssue.Message,
							Input:      keyIssue.Input,
							Path:       append([]any{key}, keyIssue.Path...), // Prepend map key to path
							Properties: make(map[string]any),
						}
						// Copy essential properties from ZodIssue to ZodRawIssue
						if keyIssue.Minimum != nil {
							rawIssue.Properties["minimum"] = keyIssue.Minimum
						}
						if keyIssue.Maximum != nil {
							rawIssue.Properties["maximum"] = keyIssue.Maximum
						}
						if keyIssue.Expected != "" {
							rawIssue.Properties["expected"] = keyIssue.Expected
						}
						if keyIssue.Received != "" {
							rawIssue.Properties["received"] = keyIssue.Received
						}
						rawIssue.Properties["inclusive"] = keyIssue.Inclusive
						collectedIssues = append(collectedIssues, rawIssue)
					}
				} else {
					// Handle non-ZodError by creating a raw issue with key path
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, key)
					rawIssue.Path = []any{key}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			}
		}

		// Validate value if value schema is provided
		if z.internals.ValueType != nil {
			if err := z.validateValueDirect(val, z.internals.ValueType, ctx); err != nil {
				// Collect value validation errors with path prefix (TypeScript Zod v4 behavior adapted for Go maps)
				var zodErr *issues.ZodError
				if errors.As(err, &zodErr) {
					// Propagate all issues from value validation with path prefix
					for _, valueIssue := range zodErr.Issues {
						// Create raw issue preserving original code and essential properties
						rawIssue := core.ZodRawIssue{
							Code:       valueIssue.Code,
							Message:    valueIssue.Message,
							Input:      valueIssue.Input,
							Path:       append([]any{key}, valueIssue.Path...), // Prepend map key to path
							Properties: make(map[string]any),
						}
						// Copy essential properties from ZodIssue to ZodRawIssue
						if valueIssue.Minimum != nil {
							rawIssue.Properties["minimum"] = valueIssue.Minimum
						}
						if valueIssue.Maximum != nil {
							rawIssue.Properties["maximum"] = valueIssue.Maximum
						}
						if valueIssue.Expected != "" {
							rawIssue.Properties["expected"] = valueIssue.Expected
						}
						if valueIssue.Received != "" {
							rawIssue.Properties["received"] = valueIssue.Received
						}
						rawIssue.Properties["inclusive"] = valueIssue.Inclusive
						collectedIssues = append(collectedIssues, rawIssue)
					}
				} else {
					// Handle non-ZodError by creating a raw issue with key path
					rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, val)
					rawIssue.Path = []any{key}
					collectedIssues = append(collectedIssues, rawIssue)
				}
			}
		}
	}

	// If we collected any issues, return them as a combined error
	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return value, nil
}

// validateValueDirect validates a single value using the provided schema without wrapping errors (preserves original error codes)
func (z *ZodMap[T, R]) validateValueDirect(value any, schema any, ctx *core.ParseContext) error {
	if schema == nil {
		return nil
	}

	// Ensure we have a valid context
	if ctx == nil {
		ctx = core.NewParseContext()
	}

	// Try using reflection to call Parse method - this handles all schema types
	schemaValue := reflect.ValueOf(schema)
	if !schemaValue.IsValid() || schemaValue.IsNil() {
		return nil
	}

	parseMethod := schemaValue.MethodByName("Parse")
	if !parseMethod.IsValid() {
		return nil
	}

	methodType := parseMethod.Type()
	if methodType.NumIn() < 1 {
		return nil
	}

	// Build arguments for Parse call
	args := []reflect.Value{reflect.ValueOf(value)}
	if methodType.NumIn() > 1 && methodType.In(1).String() == "*core.ParseContext" {
		// Add context parameter if expected
		args = append(args, reflect.ValueOf(ctx))
	}

	// Call Parse method
	results := parseMethod.Call(args)
	if len(results) >= 2 {
		// Check if there's an error (second return value)
		if errInterface := results[1].Interface(); errInterface != nil {
			if err, ok := errInterface.(error); ok {
				return err // Return the error directly without wrapping
			}
		}
	}

	return nil
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodMapFromDef constructs new ZodMap from definition
func newZodMapFromDef[T any, R any](def *ZodMapDef) *ZodMap[T, R] {
	internals := &ZodMapInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:       def,
		KeyType:   def.KeyType,
		ValueType: def.ValueType,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		mapDef := &ZodMapDef{
			ZodTypeDef: *newDef,
			KeyType:    def.KeyType,
			ValueType:  def.ValueType,
		}
		return any(newZodMapFromDef[T, R](mapDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodMap[T, R]{internals: internals}
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// Map creates map schema with key and value validation - returns value constraint
func Map(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, map[any]any] {
	return MapTyped[map[any]any, map[any]any](keySchema, valueSchema, paramArgs...)
}

// MapPtr creates map schema with key and value validation - returns pointer constraint
func MapPtr(keySchema, valueSchema any, paramArgs ...any) *ZodMap[map[any]any, *map[any]any] {
	return MapTyped[map[any]any, *map[any]any](keySchema, valueSchema, paramArgs...)
}

// MapTyped creates typed map schema with generic constraints
func MapTyped[T any, R any](keySchema, valueSchema any, paramArgs ...any) *ZodMap[T, R] {
	param := utils.GetFirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodMapDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeMap,
			Checks: []core.ZodCheck{},
		},
		KeyType:   keySchema,
		ValueType: valueSchema,
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	mapSchema := newZodMapFromDef[T, R](def)

	// Ensure validator is called when key/value schemas exist
	// Add a minimal check that always passes to trigger validation
	if keySchema != nil || valueSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		mapSchema.internals.AddCheck(alwaysPassCheck)
	}

	return mapSchema
}

// Check adds a custom validation function that can report multiple issues for map schema.
func (z *ZodMap[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodMap[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// Direct assertion to generic type R.
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch when R is a pointer type (*map) but the value is its base type.
		var zero R
		zeroTyp := reflect.TypeOf(zero)
		if zeroTyp != nil && zeroTyp.Kind() == reflect.Ptr {
			elemTyp := zeroTyp.Elem()
			valRV := reflect.ValueOf(payload.GetValue())
			if valRV.IsValid() && valRV.Type() == elemTyp {
				ptr := reflect.New(elemTyp)
				ptr.Elem().Set(valRV)
				if casted, ok := ptr.Interface().(R); ok {
					fn(casted, payload)
				}
			}
		}
	}

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// With is an alias for Check - adds a custom validation function.
// TypeScript Zod v4 equivalent: schema.with(...)
func (z *ZodMap[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodMap[T, R] {
	return z.Check(fn, params...)
}

// extractMapForEngine extracts map[any]any from input for engine.ParseComplex
func (z *ZodMap[T, R]) extractMapForEngine(input any) (map[any]any, bool) {
	result, err := z.extractMap(input, core.NewParseContext())
	if err != nil {
		return nil, false
	}
	return result, true
}

// extractMapPtrForEngine extracts pointer to map[any]any from input for engine.ParseComplex
func (z *ZodMap[T, R]) extractMapPtrForEngine(input any) (*map[any]any, bool) {
	// Try direct pointer extraction
	if ptr, ok := input.(*map[any]any); ok {
		return ptr, true
	}

	// Try extracting map and return pointer to it
	result, err := z.extractMap(input, core.NewParseContext())
	if err != nil {
		return nil, false
	}
	return &result, true
}

// validateMapForEngine validates map[any]any for engine.ParseComplex
func (z *ZodMap[T, R]) validateMapForEngine(value map[any]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
	return z.validateMap(value, checks, ctx)
}
