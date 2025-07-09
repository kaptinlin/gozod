package types

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
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
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodMap[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates input using direct validation approach
func (z *ZodMap[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
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
		return *new(R), fmt.Errorf("input cannot be nil")
	}

	// Extract map value using proper conversion
	mapValue, err := z.extractMap(input)
	if err != nil {
		return *new(R), err
	}

	// Validate the map content
	transformedMap, err := z.validateMap(mapValue, z.internals.Checks, parseCtx)
	if err != nil {
		return *new(R), err
	}

	// Convert to constraint type R using safe conversion
	return convertMapFromGeneric[T, R](transformedMap), nil
}

// MustParse validates the input value and panics on failure
func (z *ZodMap[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint
func (z *ZodMap[T, R]) Nilable() *ZodMap[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodMap[T, R]) Nullish() *ZodMap[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag to require map value and returns base constraint type.
func (z *ZodMap[T, R]) NonOptional() *ZodMap[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
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
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodMap[T, R]) DefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodMap[T, R]) Prefault(v T) *ZodMap[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodMap[T, R]) PrefaultFunc(fn func() T) *ZodMap[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
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

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum number of entries
func (z *ZodMap[T, R]) Min(minLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}
	check := checks.MinSize(minLen, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of entries
func (z *ZodMap[T, R]) Max(maxLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}
	check := checks.MaxSize(maxLen, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size sets exact number of entries
func (z *ZodMap[T, R]) Size(exactLen int, params ...any) *ZodMap[T, R] {
	schemaParams := utils.NormalizeParams(params...)
	var checkParams any
	if schemaParams.Error != nil {
		checkParams = schemaParams
	}
	check := checks.Size(exactLen, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
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
	newInternals := z.internals.ZodTypeInternals.Clone()
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
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToMapConstraintValue[T, R](v); ok {
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
func (z *ZodMap[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodMap[T, R] {
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
			return any(&mapCopy).(R) //nolint:unconvert
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
	if r, ok := any(value).(R); ok {
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
func (z *ZodMap[T, R]) extractMap(value any) (map[any]any, error) {
	// Handle direct map[any]any
	if mapVal, ok := value.(map[any]any); ok {
		return mapVal, nil
	}

	// Handle pointer to map
	if mapPtr, ok := value.(*map[any]any); ok {
		if mapPtr != nil {
			return *mapPtr, nil
		}
		return nil, fmt.Errorf("nil pointer to map")
	}

	// Try to convert using mapx
	if reflectx.IsMap(value) {
		if converted, err := mapx.ToGeneric(value); err == nil && converted != nil {
			return converted, nil
		}
	}

	return nil, fmt.Errorf("expected map[any]any, got %T", value)
}

// validateMap validates map entries using key and value schemas
func (z *ZodMap[T, R]) validateMap(value map[any]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[any]any, error) {
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

	// Always validate each key-value pair if schemas are provided - this is what makes Map different from map[any]any
	for key, val := range value {
		// Validate key if key schema is provided
		if z.internals.KeyType != nil {
			if err := z.validateValue(key, z.internals.KeyType, ctx, "key"); err != nil {
				return nil, err
			}
		}

		// Validate value if value schema is provided
		if z.internals.ValueType != nil {
			if err := z.validateValue(val, z.internals.ValueType, ctx, "value"); err != nil {
				return nil, err
			}
		}
	}

	return value, nil
}

// validateValue validates a single value using the provided schema
func (z *ZodMap[T, R]) validateValue(value any, schema any, ctx *core.ParseContext, valueType string) error {
	if schema == nil {
		return nil
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
				return fmt.Errorf("%s validation failed: %w", valueType, err)
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
		mapSchema.internals.ZodTypeInternals.AddCheck(alwaysPassCheck)
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
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}
