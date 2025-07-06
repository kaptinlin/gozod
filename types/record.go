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

// ZodRecordDef defines the schema definition for record validation
type ZodRecordDef struct {
	core.ZodTypeDef
	ValueType any // The value schema (type-erased for flexibility)
}

// ZodRecordInternals contains the internal state for record schema
type ZodRecordInternals struct {
	core.ZodTypeInternals
	Def       *ZodRecordDef // Schema definition reference
	ValueType any           // Value schema for runtime validation
}

// ZodRecord represents a type-safe record validation schema with dual generic parameters
// T = base type (map[string]any), R = constraint type (map[string]any or *map[string]any)
type ZodRecord[T any, R any] struct {
	internals *ZodRecordInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodRecord[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodRecord[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodRecord[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates input using direct validation approach
func (z *ZodRecord[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
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

	// Extract record value using proper conversion
	recordValue, err := z.extractRecord(input)
	if err != nil {
		return *new(R), err
	}

	// Validate the record content
	transformedRecord, err := z.validateRecord(recordValue, z.internals.Checks, parseCtx)
	if err != nil {
		return *new(R), err
	}

	// Convert to constraint type R using safe conversion
	return convertRecordFromGeneric[T, R](transformedRecord), nil
}

// MustParse validates the input value and panics on failure
func (z *ZodRecord[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodRecord[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional record schema that returns pointer constraint
func (z *ZodRecord[T, R]) Optional() *ZodRecord[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint
func (z *ZodRecord[T, R]) Nilable() *ZodRecord[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodRecord[T, R]) Nullish() *ZodRecord[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces record presence.
func (z *ZodRecord[T, R]) NonOptional() *ZodRecord[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodRecord[T, T]{
		internals: &ZodRecordInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			ValueType:        z.internals.ValueType,
		},
	}
}

// Default preserves current constraint type R
func (z *ZodRecord[T, R]) Default(v T) *ZodRecord[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodRecord[T, R]) DefaultFunc(fn func() T) *ZodRecord[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodRecord[T, R]) Prefault(v T) *ZodRecord[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodRecord[T, R]) PrefaultFunc(fn func() T) *ZodRecord[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum number of entries
func (z *ZodRecord[T, R]) Min(minLen int, params ...any) *ZodRecord[T, R] {
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
func (z *ZodRecord[T, R]) Max(maxLen int, params ...any) *ZodRecord[T, R] {
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
func (z *ZodRecord[T, R]) Size(exactLen int, params ...any) *ZodRecord[T, R] {
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
// Record types implement direct extraction of T values for transformation.
func (z *ZodRecord[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	// WrapFn Pattern: Create wrapper function for type-safe extraction
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		recordValue := extractRecordValue[T, R](input) // Use existing extraction logic
		return fn(recordValue, ctx)
	}

	// Use the new factory function for ZodTransform
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Pipe creates a pipeline using the WrapFn pattern.
// Instead of using adapter structures, this creates a target function that handles type conversion.
func (z *ZodRecord[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	// WrapFn Pattern: Create target function for type conversion and validation
	targetFn := func(input R, ctx *core.ParseContext) (any, error) {
		// Extract T value from constraint type R
		recordValue := extractRecordValue[T, R](input)
		// Apply target schema to the extracted T
		return target.Parse(recordValue, ctx)
	}

	// Use the new factory function for ZodPipe
	return core.NewZodPipe[R, any](z, targetFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodRecord[T, R]) Overwrite(transform func(R) R, params ...any) *ZodRecord[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToRecordType[T, R](input)
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

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation with constraint type R
func (z *ZodRecord[T, R]) Refine(fn func(R) bool, params ...any) *ZodRecord[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToRecordConstraintValue[T, R](v); ok {
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
func (z *ZodRecord[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodRecord[T, R] {
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

func (z *ZodRecord[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodRecord[T, *T] {
	return &ZodRecord[T, *T]{internals: &ZodRecordInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

func (z *ZodRecord[T, R]) withInternals(in *core.ZodTypeInternals) *ZodRecord[T, R] {
	return &ZodRecord[T, R]{internals: &ZodRecordInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		ValueType:        z.internals.ValueType,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodRecord[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodRecord[T, R]); ok {
		z.internals = src.internals
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertRecordFromGeneric safely converts map[string]any to constraint type R
func convertRecordFromGeneric[T any, R any](recordValue map[string]any) R {
	// First convert to base type T
	var baseValue T
	switch any(baseValue).(type) {
	case map[string]int:
		// Convert map[string]any to map[string]int
		converted := make(map[string]int)
		for k, v := range recordValue {
			if intVal, ok := v.(int); ok {
				converted[k] = intVal
			}
		}
		baseValue = any(converted).(T)
	case map[string]string:
		// Convert map[string]any to map[string]string
		converted := make(map[string]string)
		for k, v := range recordValue {
			if strVal, ok := v.(string); ok {
				converted[k] = strVal
			}
		}
		baseValue = any(converted).(T)
	case map[string]any:
		// Direct assignment for map[string]any
		baseValue = any(recordValue).(T)
	default:
		// For other types, try direct conversion
		baseValue = any(recordValue).(T)
	}

	// Then convert base type to constraint type
	return convertToRecordConstraintType[T, R](baseValue)
}

// convertToRecordConstraintType converts a base type T to constraint type R
func convertToRecordConstraintType[T any, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *map[string]any:
		// Need to return *map[string]any from map[string]any
		if recordVal, ok := any(value).(map[string]any); ok {
			recordCopy := recordVal
			return any(&recordCopy).(R)
		}
		return any((*map[string]any)(nil)).(R)
	default:
		// Return T directly
		return any(value).(R)
	}
}

// extractRecordValue extracts base type T from constraint type R
func extractRecordValue[T any, R any](value R) T {
	switch v := any(value).(type) {
	case *map[string]any:
		if v != nil {
			return any(*v).(T)
		}
		var zero T
		return zero
	default:
		return any(value).(T)
	}
}

// convertToRecordConstraintValue converts any value to constraint type R if possible
func convertToRecordConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Direct type match
	if r, ok := any(value).(R); ok {
		return r, true
	}

	// Handle pointer conversion for record types
	if _, ok := any(zero).(*map[string]any); ok {
		// Need to convert map[string]any to *map[string]any
		if recordVal, ok := value.(map[string]any); ok {
			recordCopy := recordVal
			return any(&recordCopy).(R), true
		}
	}

	return zero, false
}

// convertToRecordType converts any value to the record constraint type R with strict type checking
func convertToRecordType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Extract record value from input
	var recordValue map[string]any
	var isValid bool

	switch val := v.(type) {
	case map[string]any:
		recordValue, isValid = val, true
	case *map[string]any:
		if val != nil {
			recordValue, isValid = *val, true
		}
	case map[any]any:
		// Convert map[any]any to map[string]any
		recordValue = make(map[string]any)
		for k, v := range val {
			if strKey, ok := k.(string); ok {
				recordValue[strKey] = v
			} else {
				return zero, false // Non-string key found
			}
		}
		isValid = true
	default:
		return zero, false // Reject all non-map types
	}

	if !isValid {
		return zero, false
	}

	// Convert to target constraint type R
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
		// R is *map[string]any
		if converted, ok := any(&recordValue).(R); ok {
			return converted, true
		}
	} else {
		// R is map[string]any
		if converted, ok := any(recordValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// =============================================================================
// EXTRACTION AND VALIDATION
// =============================================================================

// extractRecord converts input to map[string]any
func (z *ZodRecord[T, R]) extractRecord(value any) (map[string]any, error) {
	// Handle direct map[string]any
	if recordVal, ok := value.(map[string]any); ok {
		return recordVal, nil
	}

	// Handle pointer to map
	if recordPtr, ok := value.(*map[string]any); ok {
		if recordPtr != nil {
			return *recordPtr, nil
		}
		return nil, fmt.Errorf("nil pointer to record")
	}

	// Handle map[any]any and convert to map[string]any
	if mapVal, ok := value.(map[any]any); ok {
		result := make(map[string]any)
		for k, v := range mapVal {
			if strKey, ok := k.(string); ok {
				result[strKey] = v
			} else {
				// Non-string key found, invalid for record
				return nil, fmt.Errorf("non-string key found in map: %T, records require string keys", k)
			}
		}
		return result, nil
	}

	// Try to convert using mapx
	if reflectx.IsMap(value) {
		if converted, err := mapx.ToGeneric(value); err == nil && converted != nil {
			// Convert to map[string]any
			result := make(map[string]any)
			for k, v := range converted {
				if strKey, ok := k.(string); ok {
					result[strKey] = v
				} else {
					// Non-string key found, invalid for record
					return nil, fmt.Errorf("non-string key found in map: %T, records require string keys", k)
				}
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("expected map[string]any, got %T", value)
}

// validateRecord validates record entries using value schema
func (z *ZodRecord[T, R]) validateRecord(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	// First apply checks (including Overwrite transformations) to get the transformed value
	transformedValue, err := engine.ApplyChecks[any](value, checks, ctx)
	if err != nil {
		return nil, err
	}

	// Handle potential pointer type from Overwrite transformations
	switch v := transformedValue.(type) {
	case map[string]any:
		value = v
	case *map[string]any:
		if v != nil {
			value = *v
		}
	default:
		// If transformation returned unexpected type, keep original value
		// This handles the case where no transformation was applied
	}

	// Always validate each value if value schema is provided - this is what makes Record different from map[string]any
	if z.internals.ValueType != nil {
		for key, val := range value {
			if err := z.validateValue(val, z.internals.ValueType, ctx, key); err != nil {
				return nil, err
			}
		}
	}

	return value, nil
}

// validateValue validates a single value using the provided schema
func (z *ZodRecord[T, R]) validateValue(value any, schema any, ctx *core.ParseContext, key string) error {
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
				return fmt.Errorf("value validation failed for key '%s': %w", key, err)
			}
		}
	}

	return nil
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodRecordFromDef constructs new ZodRecord from definition
func newZodRecordFromDef[T any, R any](def *ZodRecordDef) *ZodRecord[T, R] {
	internals := &ZodRecordInternals{
		ZodTypeInternals: core.ZodTypeInternals{
			Type:   def.Type,
			Checks: def.Checks,
			Coerce: def.Coerce,
			Bag:    make(map[string]any),
		},
		Def:       def,
		ValueType: def.ValueType,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		recordDef := &ZodRecordDef{
			ZodTypeDef: *newDef,
			ValueType:  def.ValueType,
		}
		return any(newZodRecordFromDef[T, R](recordDef)).(core.ZodType[any])
	}

	if def.Error != nil {
		internals.Error = def.Error
	}

	return &ZodRecord[T, R]{internals: internals}
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// Record creates record schema with string keys and typed values - returns value constraint
func Record(valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, map[string]any] {
	return RecordTyped[map[string]any, map[string]any](valueSchema, paramArgs...)
}

// RecordPtr creates record schema with string keys and typed values - returns pointer constraint
func RecordPtr(valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, *map[string]any] {
	return RecordTyped[map[string]any, *map[string]any](valueSchema, paramArgs...)
}

// RecordTyped creates typed record schema with generic constraints
func RecordTyped[T any, R any](valueSchema any, paramArgs ...any) *ZodRecord[T, R] {
	param := utils.GetFirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodRecordDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeRecord,
			Checks: []core.ZodCheck{},
		},
		ValueType: valueSchema,
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	recordSchema := newZodRecordFromDef[T, R](def)

	// Ensure validator is called when value schema exists
	// Add a minimal check that always passes to trigger validation
	if valueSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		recordSchema.internals.ZodTypeInternals.AddCheck(alwaysPassCheck)
	}

	return recordSchema
}

// Check adds a custom validation function that can report multiple issues for record schema.
func (z *ZodRecord[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodRecord[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// Try direct assertion.
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Pointer/value mismatch adaptation.
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

	check := checks.NewCustom[R](wrapper, utils.GetFirstParam(params...))
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}
