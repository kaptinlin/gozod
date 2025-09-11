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

// Static error variables
var (
	ErrInternalMapType             = errors.New("internal error: T is not a map type")
	ErrInternalMapKeyNotString     = errors.New("internal error: map key is not string")
	ErrInternalCannotConvertValue  = errors.New("internal error: cannot convert value type")
	ErrInternalCannotConvertRecord = errors.New("internal error: cannot convert validated record back to T")
	ErrNilPointerToRecord          = errors.New("nil pointer to record")
	ErrNonStringKeyInMap           = errors.New("non-string key found in map, records require string keys")
	ErrExpectedMapStringAny        = errors.New("expected map[string]any")
	ErrValueValidationFailed       = errors.New("value validation failed for key")
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodRecordDef defines the schema definition for record validation
type ZodRecordDef struct {
	core.ZodTypeDef
	KeyType   any // The key schema (type-erased for flexibility)
	ValueType any // The value schema (type-erased for flexibility)
}

// ZodRecordInternals contains the internal state for record schema
type ZodRecordInternals struct {
	core.ZodTypeInternals
	Def       *ZodRecordDef // Schema definition reference
	KeyType   any           // Key schema for runtime validation
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

// KeyType returns the key schema for this record, which is always a string.
func (z *ZodRecord[T, R]) KeyType() any {
	return z.internals.KeyType
}

// ValueType returns the value schema for JSON Schema conversion
func (z *ZodRecord[T, R]) ValueType() any {
	return z.internals.ValueType
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodRecord[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodRecord[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Parse validates input using unified ParseComplex API
func (z *ZodRecord[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	result, err := engine.ParseComplex[T](
		input,
		&z.internals.ZodTypeInternals,
		core.ZodTypeRecord,
		z.extractRecordType,
		z.extractRecordPtr,
		z.validateRecordValue,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	// Convert result to constraint type R
	if convertedResult, ok := convertToRecordConstraintValue[T, R](result); ok {
		return convertedResult, nil
	}

	// Fallback conversion
	var zero R
	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}
	return zero, issues.CreateTypeConversionError(
		fmt.Sprintf("%T", result),
		fmt.Sprintf("%T", zero),
		result,
		parseCtx,
	)
}

// MustParse validates the input value and panics on failure
func (z *ZodRecord[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
func (z *ZodRecord[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
	constraintInput, ok := convertToRecordConstraintValue[T, R](input)
	if !ok {
		var zero R
		if len(ctx) == 0 {
			ctx = []*core.ParseContext{core.NewParseContext()}
		}
		return zero, issues.CreateTypeConversionError(
			fmt.Sprintf("%T", input),
			"record constraint type",
			any(input),
			ctx[0],
		)
	}

	result, err := engine.ParseComplexStrict[T, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeRecord,
		z.extractRecordType,
		z.extractRecordPtr,
		z.validateRecordValue,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}

// MustStrictParse is the variant that panics on error
func (z *ZodRecord[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
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
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values and returns pointer constraint
func (z *ZodRecord[T, R]) Nilable() *ZodRecord[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodRecord[T, R]) Nullish() *ZodRecord[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces record presence.
func (z *ZodRecord[T, R]) NonOptional() *ZodRecord[T, T] {
	in := z.internals.Clone()
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
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current constraint type R
func (z *ZodRecord[T, R]) DefaultFunc(fn func() T) *ZodRecord[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodRecord[T, R]) Prefault(v T) *ZodRecord[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodRecord[T, R]) PrefaultFunc(fn func() T) *ZodRecord[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any { return fn() })
	return z.withInternals(in)
}

// Meta stores metadata for this record schema.
func (z *ZodRecord[T, R]) Meta(meta core.GlobalMeta) *ZodRecord[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum number of entries
func (z *ZodRecord[T, R]) Min(minLen int, params ...any) *ZodRecord[T, R] {
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
func (z *ZodRecord[T, R]) Max(maxLen int, params ...any) *ZodRecord[T, R] {
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
func (z *ZodRecord[T, R]) Size(exactLen int, params ...any) *ZodRecord[T, R] {
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
	return core.NewZodPipe[R, any](z, target, targetFn)
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
	newInternals := z.internals.Clone()
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
func (z *ZodRecord[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodRecord[T, R] {
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

func (z *ZodRecord[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodRecord[T, *T] {
	return &ZodRecord[T, *T]{internals: &ZodRecordInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
		ValueType:        z.internals.ValueType,
	}}
}

func (z *ZodRecord[T, R]) withInternals(in *core.ZodTypeInternals) *ZodRecord[T, R] {
	return &ZodRecord[T, R]{internals: &ZodRecordInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		KeyType:          z.internals.KeyType,
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
// EXTRACTION AND VALIDATION HELPERS
// =============================================================================

// extractRecordType extracts T value from input for ParseComplex
func (z *ZodRecord[T, R]) extractRecordType(input any) (T, bool) {
	var zero T
	recordValue, err := z.extractRecord(input)
	if err != nil {
		return zero, false
	}

	// Convert map[string]any to T
	if converted, ok := any(recordValue).(T); ok {
		return converted, true
	}

	// Try to convert using reflection for different map types
	if reflectx.IsMap(any(zero)) {
		zeroValue := reflect.ValueOf(zero)
		zeroType := zeroValue.Type()

		// Create a new map of the target type
		newMap := reflect.MakeMap(zeroType)

		// Convert each value to the target value type
		valueType := zeroType.Elem()
		for k, v := range recordValue {
			keyValue := reflect.ValueOf(k)
			valValue := reflect.ValueOf(v)

			// Convert value to target type if needed
			if valValue.Type().ConvertibleTo(valueType) {
				convertedVal := valValue.Convert(valueType)
				newMap.SetMapIndex(keyValue, convertedVal)
			} else {
				// If conversion fails, return false
				return zero, false
			}
		}

		// Convert the result back to T
		if typedResult, ok := newMap.Interface().(T); ok {
			return typedResult, true
		}
	}

	return zero, false
}

// extractRecordPtr extracts *T from input for ParseComplex
func (z *ZodRecord[T, R]) extractRecordPtr(input any) (*T, bool) {
	if ptr, ok := input.(*T); ok {
		return ptr, true
	}
	return nil, false
}

// validateRecordValue validates T value with checks for ParseComplex
func (z *ZodRecord[T, R]) validateRecordValue(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// Convert T back to map[string]any for validation using reflection
	var recordValue map[string]any
	if converted, ok := any(value).(map[string]any); ok {
		recordValue = converted
	} else {
		// Use reflection to convert T to map[string]any
		valueReflect := reflect.ValueOf(value)
		if valueReflect.Kind() != reflect.Map {
			return value, fmt.Errorf("%w", ErrInternalMapType)
		}

		recordValue = make(map[string]any)
		for _, key := range valueReflect.MapKeys() {
			keyStr, ok := key.Interface().(string)
			if !ok {
				return value, fmt.Errorf("%w", ErrInternalMapKeyNotString)
			}
			recordValue[keyStr] = valueReflect.MapIndex(key).Interface()
		}
	}

	// Validate the record content
	transformedRecord, err := z.validateRecord(recordValue, checks, ctx)
	if err != nil {
		return value, err
	}

	// Convert back to T using reflection
	if result, ok := any(transformedRecord).(T); ok {
		return result, nil
	} else {
		// Use reflection to convert map[string]any back to T
		targetType := reflect.TypeOf(value)
		if targetType.Kind() != reflect.Map {
			return value, fmt.Errorf("%w", ErrInternalMapType)
		}

		newMap := reflect.MakeMap(targetType)
		valueType := targetType.Elem()

		for k, v := range transformedRecord {
			keyValue := reflect.ValueOf(k)
			valValue := reflect.ValueOf(v)

			// Convert value to target type if needed
			if valValue.Type() != valueType {
				if valValue.CanConvert(valueType) {
					valValue = valValue.Convert(valueType)
				} else {
					return value, fmt.Errorf("%w", ErrInternalCannotConvertValue)
				}
			}

			newMap.SetMapIndex(keyValue, valValue)
		}

		if result, ok := newMap.Interface().(T); ok {
			return result, nil
		}
		return value, fmt.Errorf("%w", ErrInternalCannotConvertRecord)
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

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

	// Handle nil values for pointer types
	if value == nil {
		if _, ok := any(zero).(*map[string]any); ok {
			return any((*map[string]any)(nil)).(R), true
		}
	}

	// Direct type match
	if r, ok := any(value).(R); ok { //nolint:unconvert // Required for generic type constraint conversion
		return r, true
	}

	// Handle pointer conversion for record types
	if _, ok := any(zero).(*map[string]any); ok {
		// Need to convert map[string]any to *map[string]any
		if recordVal, ok := value.(map[string]any); ok {
			recordCopy := recordVal
			return any(&recordCopy).(R), true
		}
		// Handle *map[string]any to *map[string]any conversion
		if recordPtr, ok := value.(*map[string]any); ok {
			return any(recordPtr).(R), true
		}
	} else {
		// Handle non-pointer constraint type (R = map[string]any)
		// Need to convert *map[string]any to map[string]any
		if recordPtr, ok := value.(*map[string]any); ok && recordPtr != nil {
			return any(*recordPtr).(R), true
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
		return nil, fmt.Errorf("%w", ErrNilPointerToRecord)
	}

	// Handle map[any]any and convert to map[string]any
	if mapVal, ok := value.(map[any]any); ok {
		result := make(map[string]any)
		for k, v := range mapVal {
			if strKey, ok := k.(string); ok {
				result[strKey] = v
			} else {
				// Non-string key found, invalid for record
				return nil, ErrNonStringKeyInMap
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
					return nil, ErrNonStringKeyInMap
				}
			}
			return result, nil
		}
	}

	return nil, ErrExpectedMapStringAny
}

// validateRecord validates record entries using value schema
func (z *ZodRecord[T, R]) validateRecord(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	// First, apply standard checks like Min, Max, Size, and Overwrite transformations.
	transformedValue, err := engine.ApplyChecks[any](value, checks, ctx)
	if err != nil {
		return nil, err // These checks return a finalized ZodError.
	}

	// Re-assign 'value' if it was transformed by Overwrite.
	if v, ok := transformedValue.(map[string]any); ok {
		value = v
	} else if v, ok := transformedValue.(*map[string]any); ok && v != nil {
		value = *v
	}

	var rawIssues []core.ZodRawIssue
	allowedKeys, isExhaustive := tryGetExpectedKeys(z.internals.KeyType)
	isPartial, _ := z.internals.Bag["partial"].(bool)

	// --- Key Validation ---
	if isExhaustive {
		seenKeys, unrecognizedKeys := make(map[string]bool), []string{}
		for key := range value {
			if !stringInSlice(key, allowedKeys) {
				unrecognizedKeys = append(unrecognizedKeys, key)
			}
			seenKeys[key] = true
		}

		if len(unrecognizedKeys) > 0 {
			rawIssue := issues.NewRawIssue(core.UnrecognizedKeys, value, issues.WithKeys(unrecognizedKeys))
			rawIssues = append(rawIssues, rawIssue)
		}

		if !isPartial { // Exhaustiveness check for non-partial records.
			valueTypeName := core.ZodTypeAny
			if valType, ok := z.internals.ValueType.(core.ZodType[any]); ok {
				valueTypeName = valType.GetInternals().Type
			}
			for _, k := range allowedKeys {
				if !seenKeys[k] {
					issue := issues.NewRawIssue(core.InvalidType, nil, issues.WithExpected(string(valueTypeName)), issues.WithPath([]any{k}))
					rawIssues = append(rawIssues, issue)
				}
			}
		}
	} else if z.internals.KeyType != nil {
		// Non-exhaustive key validation (e.g., string with pattern).
		if keySchema, ok := z.internals.KeyType.(core.ZodType[any]); ok {
			for key := range value {
				if _, keyErr := keySchema.Parse(key); keyErr != nil {
					// For non-ZodError errors, propagate immediately to match strict validation behavior.
					var zodErr *issues.ZodError
					if errors.As(keyErr, &zodErr) {
						for _, issue := range zodErr.Issues {
							rawIssues = append(rawIssues, issues.ConvertZodIssueToRaw(issue))
						}
					} else {
						return nil, keyErr
					}
				}
			}
		}
	}

	// --- Value Validation ---
	if z.internals.ValueType != nil {
		for key, val := range value {
			// Pre-emptive check for obviously wrong types for numeric schemas to prevent panics.
			if vs, ok := z.internals.ValueType.(core.ZodType[any]); ok {
				internals := vs.GetInternals()
				if (internals.Type == core.ZodTypeInt || internals.Type == core.ZodTypeFloat) && !reflectx.IsNumber(val) {
					return nil, issues.CreateInvalidTypeError(core.ZodTypeFloat, val, ctx)
				}
			}

			// Use generic validateValue helper to leverage existing reflection logic
			if err := z.validateValue(val, z.internals.ValueType, ctx, key); err != nil {
				return nil, err
			}
		}
	}

	if len(rawIssues) > 0 {
		finalizedIssues := make([]core.ZodIssue, len(rawIssues))
		config := core.GetConfig()
		for i, raw := range rawIssues {
			finalizedIssues[i] = issues.FinalizeIssue(raw, ctx, config)
		}
		return nil, issues.NewZodError(finalizedIssues)
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
				return fmt.Errorf("%w '%s': %w", ErrValueValidationFailed, key, err)
			}
		}
	}

	return nil
}

// tryGetExpectedKeys attempts to extract expected keys from an enum/literal schema via reflection.
func tryGetExpectedKeys(schema any) ([]string, bool) {
	if schema == nil {
		return nil, false
	}

	v := reflect.ValueOf(schema)
	if !v.IsValid() || v.IsNil() {
		return nil, false
	}

	// Enumerate common methods: Options() []string or EnumValues() []string etc.
	if method := v.MethodByName("Options"); method.IsValid() {
		res := method.Call(nil)
		if len(res) == 1 {
			if arr, ok := res[0].Interface().([]string); ok {
				return arr, true
			}
		}
	}
	if method := v.MethodByName("EnumValues"); method.IsValid() {
		res := method.Call(nil)
		if len(res) == 1 {
			if arr, ok := res[0].Interface().([]string); ok {
				return arr, true
			}
		}
	}
	if method := v.MethodByName("Values"); method.IsValid() {
		res := method.Call(nil)
		if len(res) == 1 {
			if arr, ok := res[0].Interface().([]string); ok {
				return arr, true
			}
		}
	}
	return nil, false
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
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
		KeyType:   def.KeyType,
		ValueType: def.ValueType,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		recordDef := &ZodRecordDef{
			ZodTypeDef: *newDef,
			KeyType:    def.KeyType,
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

// Record creates record schema with key schema and value schema - returns value constraint
func Record(keySchema, valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, map[string]any] {
	return RecordTyped[map[string]any, map[string]any](keySchema, valueSchema, paramArgs...)
}

// RecordPtr creates record schema returning pointer constraint
func RecordPtr(keySchema, valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, *map[string]any] {
	return RecordTyped[map[string]any, *map[string]any](keySchema, valueSchema, paramArgs...)
}

// PartialRecord creates a record schema that skips exhaustive key checks (equivalent to Zod's partialRecord)
func PartialRecord(keySchema, valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, map[string]any] {
	schema := Record(keySchema, valueSchema, paramArgs...)
	if schema.internals.Bag == nil {
		schema.internals.Bag = make(map[string]any)
	}
	schema.internals.Bag["partial"] = true
	return schema
}

// PartialRecordPtr variant returning pointer constraint
func PartialRecordPtr(keySchema, valueSchema any, paramArgs ...any) *ZodRecord[map[string]any, *map[string]any] {
	schema := RecordPtr(keySchema, valueSchema, paramArgs...)
	if schema.internals.Bag == nil {
		schema.internals.Bag = make(map[string]any)
	}
	schema.internals.Bag["partial"] = true
	return schema
}

// RecordTyped creates typed record schema with generic constraints
func RecordTyped[T any, R any](keySchema, valueSchema any, paramArgs ...any) *ZodRecord[T, R] {
	param := utils.GetFirstParam(paramArgs...)
	normalizedParams := utils.NormalizeParams(param)

	def := &ZodRecordDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeRecord,
			Checks: []core.ZodCheck{},
		},
		KeyType:   keySchema,
		ValueType: valueSchema,
	}

	// Apply normalized parameters to schema definition
	if normalizedParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, normalizedParams)
	}

	recordSchema := newZodRecordFromDef[T, R](def)

	// Ensure validator is called when key or value schema exists
	if keySchema != nil || valueSchema != nil {
		alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
		recordSchema.internals.AddCheck(alwaysPassCheck)
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

	check := checks.NewCustom[any](wrapper, utils.NormalizeCustomParams(params...))
	newInternals := z.internals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}
