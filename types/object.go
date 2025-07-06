package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/utils"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ObjectMode defines how to handle unknown keys in object validation
type ObjectMode string

const (
	ObjectModeStrict      ObjectMode = "strict"      // Reject unknown keys
	ObjectModeStrip       ObjectMode = "strip"       // Remove unknown keys
	ObjectModePassthrough ObjectMode = "passthrough" // Allow unknown keys
)

// ZodObjectDef defines the schema definition for object validation
type ZodObjectDef struct {
	core.ZodTypeDef
	Shape       core.ObjectSchema // Field schemas
	Catchall    core.ZodSchema    // Schema for unrecognized keys
	UnknownKeys ObjectMode        // How to handle unknown keys
}

// ZodObjectInternals contains the internal state for object schema
type ZodObjectInternals struct {
	core.ZodTypeInternals
	Def               *ZodObjectDef     // Schema definition reference
	Shape             core.ObjectSchema // Field schemas for runtime validation
	Catchall          core.ZodSchema    // Catchall schema for unknown fields
	UnknownKeys       ObjectMode        // Validation mode
	IsPartial         bool              // Whether this is a partial object (all fields optional)
	PartialExceptions map[string]bool   // Fields that should remain required in partial mode
}

// ZodObject represents a type-safe object validation schema
type ZodObject[T any, R any] struct {
	internals *ZodObjectInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// GetInternals exposes internal state for framework usage
func (z *ZodObject[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodObject[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodObject[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Parse validates input using object-specific parsing logic
func (z *ZodObject[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil input for optional/nilable schemas
	if input == nil {
		var zero R
		if z.internals.Optional || z.internals.Nilable {
			return zero, nil
		}
		return zero, fmt.Errorf("object value cannot be nil")
	}

	// Extract object from input
	objectValue, err := z.extractObject(input)
	if err != nil {
		var zero R
		return zero, err
	}

	// Validate the object
	transformedObject, err := z.validateObject(objectValue, z.internals.Checks, parseCtx)
	if err != nil {
		var zero R
		return zero, err
	}

	// Convert to constraint type R
	return convertToObjectConstraintType[T, R](any(transformedObject).(T)), nil
}

// MustParse validates the input value and panics on failure
func (z *ZodObject[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns untyped result for runtime scenarios.
// Zero-overhead wrapper around Parse to eliminate reflection calls.
func (z *ZodObject[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional creates optional object schema
func (z *ZodObject[T, R]) Optional() *ZodObject[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable allows nil values
func (z *ZodObject[T, R]) Nilable() *ZodObject[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers
func (z *ZodObject[T, R]) Nullish() *ZodObject[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes the optional flag and enforces non-nil value type.
// It returns a schema whose constraint type is the base value (T), mirroring
// the behaviour of .Optional().NonOptional() chain in TypeScript Zod.
func (z *ZodObject[T, R]) NonOptional() *ZodObject[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodObject[T, T]{
		internals: &ZodObjectInternals{
			ZodTypeInternals:  *in,
			Def:               z.internals.Def,
			Shape:             z.internals.Shape,
			Catchall:          z.internals.Catchall,
			UnknownKeys:       z.internals.UnknownKeys,
			IsPartial:         z.internals.IsPartial,
			PartialExceptions: z.internals.PartialExceptions,
		},
	}
}

// Default preserves current type
func (z *ZodObject[T, R]) Default(v T) *ZodObject[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc preserves current type
func (z *ZodObject[T, R]) DefaultFunc(fn func() T) *ZodObject[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodObject[T, R]) Prefault(v T) *ZodObject[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodObject[T, R]) PrefaultFunc(fn func() T) *ZodObject[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// Min sets minimum number of fields
func (z *ZodObject[T, R]) Min(minLen int, params ...any) *ZodObject[T, R] {
	check := checks.MinSize(minLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Max sets maximum number of fields
func (z *ZodObject[T, R]) Max(maxLen int, params ...any) *ZodObject[T, R] {
	check := checks.MaxSize(maxLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Size sets exact number of fields
func (z *ZodObject[T, R]) Size(exactLen int, params ...any) *ZodObject[T, R] {
	check := checks.Size(exactLen, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Property validates a specific property using the provided schema
func (z *ZodObject[T, R]) Property(key string, schema core.ZodSchema, params ...any) *ZodObject[T, R] {
	check := checks.NewProperty(key, schema, params...)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check.GetZod())
	return z.withInternals(newInternals)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Shape returns the object shape (field schemas)
func (z *ZodObject[T, R]) Shape() core.ObjectSchema {
	result := make(core.ObjectSchema)
	for k, v := range z.internals.Shape {
		result[k] = v
	}
	return result
}

// Pick creates a new object with only specified keys
// Non-existent keys are silently ignored to maintain fluent interface design
func (z *ZodObject[T, R]) Pick(keys []string, params ...any) *ZodObject[T, R] {
	newShape := make(core.ObjectSchema)
	for _, key := range keys {
		if schema, exists := z.internals.Shape[key]; exists {
			newShape[key] = schema
		}
		// Silently ignore non-existent keys for chainability
	}
	return ObjectTyped[T, R](newShape, params...)
}

// Omit creates a new object excluding specified keys
// Non-existent keys are silently ignored to maintain fluent interface design
func (z *ZodObject[T, R]) Omit(keys []string, params ...any) *ZodObject[T, R] {
	excludeSet := make(map[string]bool)
	for _, key := range keys {
		// Silently ignore non-existent keys for chainability
		excludeSet[key] = true
	}

	newShape := make(core.ObjectSchema)
	for key, schema := range z.internals.Shape {
		if !excludeSet[key] {
			newShape[key] = schema
		}
	}
	return ObjectTyped[T, R](newShape, params...)
}

// Extend creates a new object by extending with additional fields
func (z *ZodObject[T, R]) Extend(augmentation core.ObjectSchema, params ...any) *ZodObject[T, R] {
	// Create new shape combining existing + extension fields
	newShape := make(core.ObjectSchema)

	// Copy existing shape
	for k, v := range z.internals.Shape {
		newShape[k] = v
	}

	// Add augmentation fields
	for k, schema := range augmentation {
		newShape[k] = schema
	}

	return ObjectTyped[T, R](newShape, params...)
}

// Merge combines two object schemas
func (z *ZodObject[T, R]) Merge(other *ZodObject[T, R], params ...any) *ZodObject[T, R] {
	return z.Extend(other.internals.Shape, params...)
}

// Partial makes all fields optional
func (z *ZodObject[T, R]) Partial(keys ...[]string) *ZodObject[T, R] {
	newInternals := z.internals.ZodTypeInternals.Clone()

	var partialExceptions map[string]bool
	if len(keys) > 0 && len(keys[0]) > 0 {
		// Specific keys provided - these are the ones to make optional
		// All other fields remain required
		partialExceptions = make(map[string]bool)
		for fieldName := range z.internals.Shape {
			partialExceptions[fieldName] = true // Mark all as exceptions initially
		}
		// Remove the keys that should be made optional from exceptions
		for _, key := range keys[0] {
			delete(partialExceptions, key)
		}
	}

	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:  *newInternals,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		Catchall:          z.internals.Catchall,
		UnknownKeys:       z.internals.UnknownKeys,
		IsPartial:         true,
		PartialExceptions: partialExceptions,
	}}
}

// Required makes all fields required (opposite of Partial)
func (z *ZodObject[T, R]) Required(fields ...[]string) *ZodObject[T, R] {
	newInternals := z.internals.ZodTypeInternals.Clone()

	var partialExceptions map[string]bool
	if len(fields) > 0 && len(fields[0]) > 0 {
		// Specific fields provided - these become required
		partialExceptions = make(map[string]bool)
		for _, fieldName := range fields[0] {
			partialExceptions[fieldName] = true
		}
	}

	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:  *newInternals,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		Catchall:          z.internals.Catchall,
		UnknownKeys:       z.internals.UnknownKeys,
		IsPartial:         true,              // Keep as partial, but with specific required fields
		PartialExceptions: partialExceptions, // Fields in this map are required
	}}
}

// Strict sets strict mode (no unknown keys allowed)
func (z *ZodObject[T, R]) Strict() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModeStrict)
}

// Strip sets strip mode (unknown keys are removed)
func (z *ZodObject[T, R]) Strip() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModeStrip)
}

// Passthrough sets passthrough mode (unknown keys are allowed)
func (z *ZodObject[T, R]) Passthrough() *ZodObject[T, R] {
	return z.withUnknownKeys(ObjectModePassthrough)
}

// Catchall sets a schema for unknown keys
func (z *ZodObject[T, R]) Catchall(catchallSchema core.ZodSchema) *ZodObject[T, R] {
	newInternals := z.internals.ZodTypeInternals.Clone()
	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals: *newInternals,
		Def:              z.internals.Def,
		Shape:            z.internals.Shape,
		Catchall:         catchallSchema,
		UnknownKeys:      z.internals.UnknownKeys,
	}}
}

// Keyof returns a string literal schema of all keys
func (z *ZodObject[T, R]) Keyof() *ZodEnum[string, string] {
	keys := make([]string, 0, len(z.internals.Shape))
	for key := range z.internals.Shape {
		keys = append(keys, key)
	}
	return EnumSlice(keys)
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodObject[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		baseValue := extractObjectValue[T, R](input)
		return fn(baseValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
func (z *ZodObject[T, R]) Overwrite(transform func(R) R, params ...any) *ZodObject[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToObjectType[T, R](input)
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

// Pipe creates a pipeline using WrapFn pattern
func (z *ZodObject[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		baseValue := extractObjectValue[T, R](input)
		return target.Parse(baseValue, ctx)
	}
	return core.NewZodPipe[R, any](z, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine applies type-safe validation using constraint type
func (z *ZodObject[T, R]) Refine(fn func(R) bool, params ...any) *ZodObject[T, R] {
	wrapper := func(v any) bool {
		if v == nil {
			return true // Skip refinement for nil values
		}
		// Convert any to constraint type R for type-safe refinement
		if constraintVal, ok := convertToConstraintValue[T, R](v); ok {
			return fn(constraintVal)
		}
		return false
	}

	checkParams := checks.NormalizeCheckParams(params...)
	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// RefineAny provides flexible validation without type conversion
func (z *ZodObject[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodObject[T, R] {
	checkParams := checks.NormalizeCheckParams(params...)
	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// HELPER AND PRIVATE METHODS
// =============================================================================

// withPtrInternals creates new instance with pointer constraint type
func (z *ZodObject[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodObject[T, *T] {
	return &ZodObject[T, *T]{internals: &ZodObjectInternals{
		ZodTypeInternals:  *in,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		Catchall:          z.internals.Catchall,
		UnknownKeys:       z.internals.UnknownKeys,
		IsPartial:         z.internals.IsPartial,
		PartialExceptions: z.internals.PartialExceptions,
	}}
}

// withInternals creates new instance preserving type
func (z *ZodObject[T, R]) withInternals(in *core.ZodTypeInternals) *ZodObject[T, R] {
	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:  *in,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		Catchall:          z.internals.Catchall,
		UnknownKeys:       z.internals.UnknownKeys,
		IsPartial:         z.internals.IsPartial,
		PartialExceptions: z.internals.PartialExceptions,
	}}
}

// withUnknownKeys creates new instance with specified unknown keys handling
func (z *ZodObject[T, R]) withUnknownKeys(mode ObjectMode) *ZodObject[T, R] {
	newInternals := z.internals.ZodTypeInternals.Clone()
	return &ZodObject[T, R]{internals: &ZodObjectInternals{
		ZodTypeInternals:  *newInternals,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		Catchall:          z.internals.Catchall,
		UnknownKeys:       mode,
		IsPartial:         z.internals.IsPartial,
		PartialExceptions: z.internals.PartialExceptions,
	}}
}

// CloneFrom copies configuration from another schema
func (z *ZodObject[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodObject[T, R]); ok {
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// convertToObjectConstraintType converts base type T to constraint type R
func convertToObjectConstraintType[T any, R any](value T) R {
	// For value types, R should be T
	// For pointer constraints, R should be *T
	var result R
	resultVal := reflect.ValueOf(&result).Elem()

	// Check if R is a pointer to T
	if resultVal.Kind() == reflect.Pointer {
		// Create a new instance of T and set it
		if reflect.TypeOf(value).AssignableTo(resultVal.Type().Elem()) {
			newVal := reflect.New(resultVal.Type().Elem())
			newVal.Elem().Set(reflect.ValueOf(value))
			resultVal.Set(newVal)
		}
	} else {
		// Direct assignment for value types
		if reflect.TypeOf(value).AssignableTo(resultVal.Type()) {
			resultVal.Set(reflect.ValueOf(value))
		}
	}

	return result
}

// extractObjectValue extracts the base value from constraint type R
func extractObjectValue[T any, R any](value R) T {
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

// convertToObjectType converts any value to the object constraint type R with strict type checking
func convertToObjectType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Extract object value from input
	var objectValue map[string]any
	var isValid bool

	switch val := v.(type) {
	case map[string]any:
		objectValue, isValid = val, true
	case *map[string]any:
		if val != nil {
			objectValue, isValid = *val, true
		}
	case map[any]any:
		// Convert map[any]any to map[string]any
		objectValue = make(map[string]any)
		for k, v := range val {
			if strKey, ok := k.(string); ok {
				objectValue[strKey] = v
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
		if converted, ok := any(&objectValue).(R); ok {
			return converted, true
		}
	} else {
		// R is map[string]any
		if converted, ok := any(objectValue).(R); ok {
			return converted, true
		}
	}

	return zero, false
}

// extractObject extracts map[string]any from value with proper conversion
func (z *ZodObject[T, R]) extractObject(value any) (map[string]any, error) {
	// Handle direct map[string]any
	if objectVal, ok := value.(map[string]any); ok {
		return objectVal, nil
	}

	// Handle pointer to map[string]any
	if objectPtr, ok := value.(*map[string]any); ok && objectPtr != nil {
		return *objectPtr, nil
	}

	// Handle struct conversion using reflection
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("nil pointer cannot be converted to object")
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		result := make(map[string]any)
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.IsExported() {
				// Use json tag if present, otherwise use field name
				name := field.Name
				if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
					if commaIdx := strings.Index(tag, ","); commaIdx > 0 {
						name = tag[:commaIdx]
					} else {
						name = tag
					}
				}
				result[name] = rv.Field(i).Interface()
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("cannot convert %T to map[string]any", value)
}

// validateObject validates object fields using field schemas
func (z *ZodObject[T, R]) validateObject(value map[string]any, checks []core.ZodCheck, ctx *core.ParseContext) (map[string]any, error) {
	// Validate known fields first
	for fieldName, fieldSchema := range z.internals.Shape {
		fieldValue, exists := value[fieldName]

		if !exists {
			if !z.isFieldOptional(fieldSchema, fieldName) {
				return nil, fmt.Errorf("missing required field: %s", fieldName)
			}
			continue
		}

		if err := z.validateField(fieldValue, fieldSchema, ctx, fieldName); err != nil {
			return nil, fmt.Errorf("field '%s' validation failed: %w", fieldName, err)
		}
	}

	// Separate Overwrite checks from other checks
	var overwriteChecks []core.ZodCheck
	var otherChecks []core.ZodCheck

	for _, check := range checks {
		// Check if this is an Overwrite check by checking the check name
		if checkInternals := check.GetZod(); checkInternals != nil && checkInternals.Def != nil && checkInternals.Def.Check == "overwrite" {
			overwriteChecks = append(overwriteChecks, check)
		} else {
			otherChecks = append(otherChecks, check)
		}
	}

	// Apply Overwrite transformations before unknown field handling (preserves added fields)
	if len(overwriteChecks) > 0 {
		transformedValue, err := engine.ApplyChecks[map[string]any](value, overwriteChecks, ctx)
		if err != nil {
			return nil, err
		}
		value = transformedValue
	}

	// Handle unknown fields based on mode
	var fieldsToRemove []string
	for fieldName, fieldValue := range value {
		if _, isKnown := z.internals.Shape[fieldName]; !isKnown {
			switch z.internals.UnknownKeys {
			case ObjectModeStrict:
				return nil, fmt.Errorf("unknown field not allowed in strict mode: %s", fieldName)
			case ObjectModeStrip:
				fieldsToRemove = append(fieldsToRemove, fieldName)
			case ObjectModePassthrough:
				if z.internals.Catchall != nil {
					if err := z.validateField(fieldValue, z.internals.Catchall, ctx, fieldName); err != nil {
						return nil, fmt.Errorf("catchall validation failed for field '%s': %w", fieldName, err)
					}
				}
			}
		}
	}

	// Remove unknown fields for strip mode (except for Overwrite-added fields)
	if len(overwriteChecks) == 0 {
		for _, fieldName := range fieldsToRemove {
			delete(value, fieldName)
		}
	}

	// Apply other checks (Size, Min, Max, etc.) after field processing
	if len(otherChecks) > 0 {
		finalValue, err := engine.ApplyChecks[map[string]any](value, otherChecks, ctx)
		if err != nil {
			return nil, err
		}
		return finalValue, nil
	}

	return value, nil
}

// validateField validates a single field using reflection (same as struct.go)
func (z *ZodObject[T, R]) validateField(element any, schema any, ctx *core.ParseContext, fieldName string) error {
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
	var firstArg reflect.Value
	if element == nil {
		// If element is nil, use zero value of the method's first argument type
		firstArg = reflect.Zero(methodType.In(0))
	} else {
		firstArg = reflect.ValueOf(element)
	}
	args := []reflect.Value{firstArg}
	if methodType.NumIn() > 1 && methodType.In(1).String() == "*core.ParseContext" {
		// Add context parameter if expected
		if ctx == nil {
			args = append(args, reflect.Zero(methodType.In(1)))
		} else {
			args = append(args, reflect.ValueOf(ctx))
		}
	}

	// Call Parse method
	results := parseMethod.Call(args)
	if len(results) >= 2 {
		// Check if there's an error (second return value)
		if errInterface := results[1].Interface(); errInterface != nil {
			if err, ok := errInterface.(error); ok {
				return err
			}
		}
	}

	return nil
}

// isFieldOptional checks if a field schema is optional using reflection or partial state
func (z *ZodObject[T, R]) isFieldOptional(schema any, fieldName string) bool {
	if schema == nil {
		return true
	}

	// Check if this object is in partial mode and this field should be optional in partial mode
	if z.internals.IsPartial {
		// If there are no exceptions, all fields are optional
		if z.internals.PartialExceptions == nil {
			return true
		}
		// If this field is not in the exceptions list, it's optional
		if !z.internals.PartialExceptions[fieldName] {
			return true
		}
	}

	schemaValue := reflect.ValueOf(schema)
	if !schemaValue.IsValid() || schemaValue.IsNil() {
		return true
	}

	// Try to get internals to check if optional
	if internalsMethod := schemaValue.MethodByName("GetInternals"); internalsMethod.IsValid() {
		results := internalsMethod.Call(nil)
		if len(results) > 0 {
			if internals, ok := results[0].Interface().(*core.ZodTypeInternals); ok {
				return internals.Optional
			}
		}
	}

	return false
}

// newZodObjectFromDef constructs new ZodObject from definition
func newZodObjectFromDef[T any, R any](def *ZodObjectDef) *ZodObject[T, R] {
	internals := &ZodObjectInternals{
		ZodTypeInternals:  engine.NewBaseZodTypeInternals(def.Type),
		Def:               def,
		Shape:             def.Shape,
		Catchall:          def.Catchall,
		UnknownKeys:       def.UnknownKeys,
		IsPartial:         false,
		PartialExceptions: nil,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		objectDef := &ZodObjectDef{
			ZodTypeDef:  *newDef,
			Shape:       def.Shape,
			Catchall:    def.Catchall,
			UnknownKeys: def.UnknownKeys,
		}
		return any(newZodObjectFromDef[T, R](objectDef)).(core.ZodType[any])
	}

	schema := &ZodObject[T, R]{internals: internals}

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
// CONSTRUCTORS AND FACTORY FUNCTIONS
// =============================================================================

// ObjectTyped creates a typed object schema with explicit type parameters
func ObjectTyped[T any, R any](shape core.ObjectSchema, params ...any) *ZodObject[T, R] {
	schemaParams := utils.NormalizeParams(params...)

	def := &ZodObjectDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeObject,
			Checks: []core.ZodCheck{},
		},
		Shape:       make(core.ObjectSchema),
		Catchall:    nil,
		UnknownKeys: ObjectModeStrip, // Default mode
	}

	// Copy the shape directly since core.ObjectSchema contains ZodSchema types
	for key, schema := range shape {
		def.Shape[key] = schema
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	objectSchema := newZodObjectFromDef[T, R](def)

	// Add a minimal check to trigger field validation
	alwaysPassCheck := checks.NewCustom[any](func(v any) bool { return true }, core.SchemaParams{})
	objectSchema.internals.ZodTypeInternals.AddCheck(alwaysPassCheck)

	return objectSchema
}

// Object creates object schema with default types (map[string]any, map[string]any)
func Object(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return ObjectTyped[map[string]any, map[string]any](shape, params...)
}

// ObjectPtr creates object schema with pointer constraint
func ObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...)
}

// StrictObject creates strict object schema with default types
func StrictObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return Object(shape, params...).Strict()
}

// LooseObject creates loose object schema with default types
func LooseObject(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, map[string]any] {
	return Object(shape, params...).Passthrough()
}

// StrictObjectPtr creates strict object schema with pointer constraint
func StrictObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...).Strict()
}

// LooseObjectPtr creates loose object schema with pointer constraint
func LooseObjectPtr(shape core.ObjectSchema, params ...any) *ZodObject[map[string]any, *map[string]any] {
	return ObjectTyped[map[string]any, *map[string]any](shape, params...).Passthrough()
}

// Check adds a custom validation function that can report multiple issues for object schema.
func (z *ZodObject[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodObject[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// Direct assertion attempt.
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Pointer/value mismatch handling.
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
