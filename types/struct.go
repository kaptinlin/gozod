package types

import (
	"errors"
	"fmt"
	"maps"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
	"github.com/kaptinlin/gozod/pkg/tagparser"
)

// Static error variables
var (
	ErrFieldNotFoundOrNotSettable = errors.New("field not found or not settable")
	ErrCannotAssignToField        = errors.New("cannot assign value to field of type")
)

// anyType is the reflect.Type for interface{}/any, cached to avoid repeated allocation.
var anyType = reflect.TypeFor[any]()

// jsonTagName extracts the field name from a JSON struct tag, ignoring options like omitempty.
func jsonTagName(tag string) string {
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ZodStructDef defines the schema definition for struct validation
type ZodStructDef struct {
	core.ZodTypeDef
	Shape core.StructSchema // Field schemas for struct validation
}

// ZodStructInternals contains the internal state for struct schema
type ZodStructInternals struct {
	core.ZodTypeInternals
	Def               *ZodStructDef     // Schema definition reference
	Shape             core.StructSchema // Field schemas for runtime validation
	IsPartial         bool              // Whether this is a partial struct (fields can be zero values)
	PartialExceptions map[string]bool   // Fields that should remain required in partial mode
}

// ZodStruct represents a type-safe struct validation schema with dual generic parameters
// T is the base comparable type, R is the constraint type (T | *T)
type ZodStruct[T any, R any] struct {
	internals *ZodStructInternals
}

// =============================================================================
// CORE METHODS
// =============================================================================

// Internals returns the internal state of the schema.
func (z *ZodStruct[T, R]) Internals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional reports whether this schema accepts undefined/missing values.
func (z *ZodStruct[T, R]) IsOptional() bool {
	return z.internals.IsOptional()
}

// IsNilable reports whether this schema accepts nil values.
func (z *ZodStruct[T, R]) IsNilable() bool {
	return z.internals.IsNilable()
}

// Shape returns the struct field schemas.
func (z *ZodStruct[T, R]) Shape() core.StructSchema {
	return z.internals.Shape
}

// UnknownKeys returns the unknown keys handling mode.
func (z *ZodStruct[T, R]) UnknownKeys() string {
	return "strict"
}

// Catchall returns nil because struct schemas are always strict.
func (z *ZodStruct[T, R]) Catchall() core.ZodSchema {
	return nil
}

// Parse validates input and returns a value of type R.
func (z *ZodStruct[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var zero R

	parseCtx := core.NewParseContext()
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	}

	// Check if constraint type R is a pointer type
	isPointerConstraint := reflect.TypeFor[R]().Kind() == reflect.Pointer

	// Temporarily enable Optional flag for pointer constraint types to ensure pointer identity preservation in ParseComplex
	// But only if no other modifiers (Optional, Nilable, Prefault) are set
	originalInternals := &z.internals.ZodTypeInternals
	if isPointerConstraint &&
		!originalInternals.Optional &&
		!originalInternals.Nilable &&
		originalInternals.PrefaultValue == nil &&
		originalInternals.PrefaultFunc == nil {
		modifiedInternals := *originalInternals
		modifiedInternals.Optional = true
		originalInternals = &modifiedInternals
	}

	result, err := engine.ParseComplex[T](
		input,
		originalInternals,
		core.ZodTypeStruct,
		z.extractStructForEngine,
		z.extractStructPtrForEngine,
		z.validateStructForEngine,
		parseCtx,
	)
	if err != nil {
		// Check if this is a generic struct type error that we can improve
		if errStr := err.Error(); strings.Contains(errStr, "Invalid input: expected struct, received") {
			return zero, z.createStructTypeError(input, parseCtx)
		}
		return zero, err
	}

	// Convert result to constraint type R
	if structVal, ok := result.(T); ok {
		return convertToStructConstraintType[T, R](structVal), nil
	}

	// Handle pointer to T
	if structPtr, ok := result.(*T); ok {
		if structPtr == nil {
			return zero, nil
		}
		return convertToStructConstraintType[T, R](*structPtr), nil
	}

	// Handle nil result for optional/nilable schemas
	if result == nil {
		return zero, nil
	}

	return zero, issues.CreateTypeConversionError(fmt.Sprintf("%T", result), "struct", input, parseCtx)
}

// MustParse panics on validation failure.
func (z *ZodStruct[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse validates input with compile-time type safety.
func (z *ZodStruct[T, R]) StrictParse(input T, ctx ...*core.ParseContext) (R, error) {
	// Convert T to R for ParseComplexStrict
	constraintInput := convertToStructConstraintType[T, R](input)

	result, err := engine.ParseComplexStrict[T, R](
		constraintInput,
		&z.internals.ZodTypeInternals,
		core.ZodTypeStruct,
		z.extractStructForEngine,
		z.extractStructPtrForEngine,
		z.validateStructForEngine,
		ctx...,
	)
	if err != nil {
		var zero R
		return zero, err
	}

	return result, nil
}

// MustStrictParse panics on validation failure with compile-time type safety.
func (z *ZodStruct[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates input and returns an untyped result.
func (z *ZodStruct[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional returns a schema that accepts nil values with pointer constraint.
func (z *ZodStruct[T, R]) Optional() *ZodStruct[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable returns a schema that accepts nil values with pointer constraint.
func (z *ZodStruct[T, R]) Nilable() *ZodStruct[T, *T] {
	in := z.internals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers.
func (z *ZodStruct[T, R]) Nullish() *ZodStruct[T, *T] {
	in := z.internals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional enforces non-nil struct value.
func (z *ZodStruct[T, R]) NonOptional() *ZodStruct[T, T] {
	in := z.internals.Clone()
	in.SetOptional(false)
	in.SetNonOptional(true)

	return &ZodStruct[T, T]{
		internals: &ZodStructInternals{
			ZodTypeInternals: *in,
			Def:              z.internals.Def,
			Shape:            z.internals.Shape,
		},
	}
}

// Default sets the default value.
func (z *ZodStruct[T, R]) Default(v T) *ZodStruct[T, R] {
	in := z.internals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc sets the default value function.
func (z *ZodStruct[T, R]) DefaultFunc(fn func() T) *ZodStruct[T, R] {
	in := z.internals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault sets the prefault value.
func (z *ZodStruct[T, R]) Prefault(v T) *ZodStruct[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc sets the prefault value function.
func (z *ZodStruct[T, R]) PrefaultFunc(fn func() T) *ZodStruct[T, R] {
	in := z.internals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata.
func (z *ZodStruct[T, R]) Meta(meta core.GlobalMeta) *ZodStruct[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// Describe registers a description in the global registry.
func (z *ZodStruct[T, R]) Describe(description string) *ZodStruct[T, R] {
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
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation.
func (z *ZodStruct[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		structValue := extractStructValue[T, R](input)
		return fn(structValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
func (z *ZodStruct[T, R]) Overwrite(transform func(R) R, params ...any) *ZodStruct[T, R] {
	// Create a transformation function that works with the constraint type R
	transformAny := func(input any) any {
		// Try to convert input to constraint type R
		converted, ok := convertToStructType[T, R](input)
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

// Pipe creates a pipeline.
func (z *ZodStruct[T, R]) Pipe(target core.ZodType[any]) *core.ZodPipe[R, any] {
	wrapperFn := func(input R, ctx *core.ParseContext) (any, error) {
		structValue := extractStructValue[T, R](input)
		return target.Parse(structValue, ctx)
	}
	return core.NewZodPipe[R, any](z, target, wrapperFn)
}

// =============================================================================
// REFINEMENT METHODS
// =============================================================================

// Refine adds a custom validation function.
func (z *ZodStruct[T, R]) Refine(fn func(R) bool, params ...any) *ZodStruct[T, R] {
	wrapper := func(v any) bool {
		if constraintValue, ok := convertToConstraintValue[T, R](v); ok {
			return fn(constraintValue)
		}
		return false
	}
	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](wrapper, checks.NormalizeCheckParams(params...)))
	return z.withInternals(in)
}

// RefineAny adds a custom validation function for any type.
func (z *ZodStruct[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodStruct[T, R] {
	in := z.internals.Clone()
	in.AddCheck(checks.NewCustom[any](fn, checks.NormalizeCheckParams(params...)))
	return z.withInternals(in)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Extend creates a new struct with additional fields.
func (z *ZodStruct[T, R]) Extend(augmentation core.StructSchema, params ...any) *ZodStruct[T, R] {
	// Create new shape combining existing + extension fields
	newShape := make(core.StructSchema)
	maps.Copy(newShape, z.internals.Shape)
	maps.Copy(newShape, augmentation)

	// Create new definition with extended shape
	schemaParams := utils.NormalizeParams(params...)
	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: []core.ZodCheck{},
		},
		Shape: newShape,
	}

	// Apply schema parameters if provided
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodStructFromDef[T, R](def)
}

// Partial makes all fields optional.
func (z *ZodStruct[T, R]) Partial(keys ...[]string) *ZodStruct[T, R] {
	newInternals := z.internals.Clone()

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

	return &ZodStruct[T, R]{internals: &ZodStructInternals{
		ZodTypeInternals:  *newInternals,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		IsPartial:         true,
		PartialExceptions: partialExceptions,
	}}
}

// Required makes all fields required.
func (z *ZodStruct[T, R]) Required(fields ...[]string) *ZodStruct[T, R] {
	newInternals := z.internals.Clone()

	var partialExceptions map[string]bool
	if len(fields) > 0 && len(fields[0]) > 0 {
		// Specific fields provided - these become required
		partialExceptions = make(map[string]bool)
		for _, fieldName := range fields[0] {
			partialExceptions[fieldName] = true
		}
	}

	return &ZodStruct[T, R]{internals: &ZodStructInternals{
		ZodTypeInternals:  *newInternals,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		IsPartial:         true,              // Keep as partial, but with specific required fields
		PartialExceptions: partialExceptions, // Fields in this map are required
	}}
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (z *ZodStruct[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodStruct[T, *T] {
	return &ZodStruct[T, *T]{internals: &ZodStructInternals{
		ZodTypeInternals:  *in,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		IsPartial:         z.internals.IsPartial,
		PartialExceptions: z.internals.PartialExceptions,
	}}
}

func (z *ZodStruct[T, R]) withInternals(in *core.ZodTypeInternals) *ZodStruct[T, R] {
	return &ZodStruct[T, R]{internals: &ZodStructInternals{
		ZodTypeInternals:  *in,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		IsPartial:         z.internals.IsPartial,
		PartialExceptions: z.internals.PartialExceptions,
	}}
}

// CloneFrom copies configuration from a source.
func (z *ZodStruct[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodStruct[T, R]); ok {
		originalChecks := z.internals.Checks
		*z.internals = *src.internals
		z.internals.Checks = originalChecks
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToStructConstraintType converts T to constraint type R.
func convertToStructConstraintType[T any, R any](value T) R {
	var zero R
	switch any(zero).(type) {
	case *T:
		// Need to return *T from T
		valueCopy := value
		return any(&valueCopy).(R)
	default:
		// Return T directly
		return any(value).(R)
	}
}

// extractStructValue extracts T from constraint type R.
func extractStructValue[T any, R any](value R) T {
	// Handle direct assignment (when T == R)
	if directValue, ok := any(value).(T); ok {
		return directValue
	}

	// Handle pointer dereferencing
	if ptrValue := reflect.ValueOf(value); ptrValue.Kind() == reflect.Pointer && !ptrValue.IsNil() {
		if derefValue, ok := ptrValue.Elem().Interface().(T); ok {
			return derefValue
		}
	}

	// Fallback to zero value
	var zero T
	return zero
}

// convertToConstraintValue converts any value to constraint type R.
func convertToConstraintValue[T any, R any](value any) (R, bool) {
	var zero R

	// Handle nil values for pointer types
	if value == nil {
		if _, ok := any(zero).(*T); ok {
			// R is *T, return nil pointer
			return zero, true
		}
		return zero, false
	}

	// Direct type match
	if r, ok := any(value).(R); ok { //nolint:unconvert
		return r, true
	}

	// Handle pointer conversion for struct types
	if _, ok := any(zero).(*T); ok {
		// Need to convert T to *T
		if structVal, ok := value.(T); ok {
			structCopy := structVal
			return any(&structCopy).(R), true
		}
	}

	return zero, false
}

// convertToStructType converts any value to constraint type R.
func convertToStructType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeFor[R]()
		if zeroType.Kind() == reflect.Pointer {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Try direct conversion first
	if converted, ok := any(v).(R); ok { //nolint:unconvert
		return converted, true
	}

	// Handle pointer conversion for different types
	zeroType := reflect.TypeFor[R]()
	if zeroType.Kind() == reflect.Pointer {
		// R is a pointer type
		elemType := zeroType.Elem()

		// Check if v can be converted to the element type
		if vType := reflect.TypeOf(v); vType != nil {
			if vType == elemType || vType.ConvertibleTo(elemType) {
				// Create pointer to the value
				vValue := reflect.ValueOf(v)
				if vType.ConvertibleTo(elemType) {
					converted := vValue.Convert(elemType)
					ptrVal := reflect.New(elemType)
					ptrVal.Elem().Set(converted)
					if result, ok := ptrVal.Interface().(R); ok {
						return result, true
					}
				} else {
					// Direct pointer creation
					if result, ok := any(&v).(R); ok {
						return result, true
					}
				}
			}
		}
		return zero, false
	}

	return zero, false
}

// =============================================================================
// ENGINE INTEGRATION METHODS
// =============================================================================

// extractStructForEngine extracts T from input.
func (z *ZodStruct[T, R]) extractStructForEngine(input any) (T, bool) {
	var zero T

	// Handle nil input
	if input == nil {
		return zero, false
	}

	// Handle direct type T
	if structVal, ok := input.(T); ok {
		return structVal, true
	}

	// Handle pointer to T
	if val := reflect.ValueOf(input); val.Kind() == reflect.Pointer && !val.IsNil() {
		if structVal, ok := val.Elem().Interface().(T); ok {
			return structVal, true
		}
	}

	// Handle map conversion
	switch m := input.(type) {
	case map[string]any:
		if converted, ok := convertMapToStructStrict[T](m); ok {
			return converted, true
		}
	case map[any]any:
		strMap := make(map[string]any)
		for k, v := range m {
			if ks, ok := k.(string); ok {
				strMap[ks] = v
			}
		}
		if converted, ok := convertMapToStructStrict[T](strMap); ok {
			return converted, true
		}
	}

	return zero, false
}

// createStructTypeError creates a struct type error.
func (z *ZodStruct[T, R]) createStructTypeError(input any, ctx *core.ParseContext) error {
	var zero T
	expectedType := reflect.TypeOf(zero)
	inputType := reflect.TypeOf(input)

	// Ensure we have a valid context for the error
	if ctx == nil {
		ctx = core.NewParseContext()
	}

	if expectedType != nil && inputType != nil {
		message := fmt.Sprintf("Invalid input: expected struct of type %s, got %s", expectedType, inputType)
		return issues.CreateCustomError(message, map[string]any{
			"expected": expectedType.String(),
			"received": inputType.String(),
		}, input, ctx)
	}

	// Fallback to generic error
	return issues.CreateInvalidTypeError(core.ZodTypeStruct, input, ctx)
}

// extractStructPtrForEngine extracts *T from input.
func (z *ZodStruct[T, R]) extractStructPtrForEngine(input any) (*T, bool) {
	// Handle nil input
	if input == nil {
		return nil, true
	}

	// Handle direct *T
	if structPtr, ok := input.(*T); ok {
		return structPtr, true
	}

	// Handle T and convert to *T
	if structVal, ok := input.(T); ok {
		return &structVal, true
	}

	// Handle pointer to T via reflection
	if val := reflect.ValueOf(input); val.Kind() == reflect.Pointer && !val.IsNil() {
		if structVal, ok := val.Elem().Interface().(T); ok {
			structCopy := structVal
			return &structCopy, true
		}
	}

	// Handle map conversion
	switch m := input.(type) {
	case map[string]any:
		if converted, ok := convertMapToStructStrict[T](m); ok {
			return &converted, true
		}
	case map[any]any:
		strMap := make(map[string]any)
		for k, v := range m {
			if ks, ok := k.(string); ok {
				strMap[ks] = v
			}
		}
		if converted, ok := convertMapToStructStrict[T](strMap); ok {
			return &converted, true
		}
	}

	return nil, false
}

// validateStructForEngine validates struct value.
func (z *ZodStruct[T, R]) validateStructForEngine(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// Apply field defaults and validate struct fields if schema is defined
	transformedValue := value
	if len(z.internals.Shape) > 0 {
		if transformed, err := z.parseStructWithDefaults(any(value), ctx); err != nil {
			return value, err
		} else if convertedValue, ok := transformed.(T); ok {
			transformedValue = convertedValue
		}
	}

	// Apply validation checks
	if len(checks) > 0 {
		finalValue, err := engine.ApplyChecks(transformedValue, checks, ctx)
		if err != nil {
			return transformedValue, err
		}

		// Use reflection to check if transformed value can be converted to T
		transformedVal := reflect.ValueOf(finalValue)
		if transformedVal.IsValid() && !transformedVal.IsZero() {
			if transformedVal.Type().AssignableTo(reflect.TypeOf(transformedValue)) {
				return transformedVal.Interface().(T), nil
			}
		}
	}

	return transformedValue, nil
}

// =============================================================================
// VALIDATION LOGIC
// =============================================================================

// parseStructWithDefaults parses struct fields with defaults.
func (z *ZodStruct[T, R]) parseStructWithDefaults(input any, ctx *core.ParseContext) (any, error) {
	if len(z.internals.Shape) == 0 {
		return input, nil // No field schemas defined, return input as-is
	}

	// Use reflection to access struct fields
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, issues.CreateInvalidTypeError(core.ZodTypeStruct, input, ctx)
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, issues.CreateInvalidTypeError(core.ZodTypeStruct, input, ctx)
	}

	structType := val.Type()

	// Create a new struct instance to hold transformed values
	newStruct := reflect.New(structType).Elem()

	// Copy original values first
	newStruct.Set(val)

	var collectedIssues []core.ZodRawIssue

	// Process each field defined in the schema
	for fieldName, fieldSchema := range z.internals.Shape {
		if fieldSchema == nil {
			continue // Skip nil schemas
		}

		// Find the struct field (check both field name and json tag)
		fieldValue, found := z.getStructFieldValue(val, structType, fieldName)
		if !found {
			// Field not found in struct - handle missing required fields
			if !z.isFieldOptional(fieldSchema, fieldName) {
				rawIssue := issues.CreateIssue(core.InvalidType, fmt.Sprintf("Missing required struct field: %s", fieldName), map[string]any{
					"expected": "nonoptional",
					"received": "undefined",
				}, nil)
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
			continue
		}

		// Check if this field should be skipped in partial mode
		if z.shouldSkipFieldInPartialMode(fieldValue.Interface(), fieldName) {
			continue
		}

		// Parse the field value with its schema (this applies defaults and transformations)
		parsedFieldValue, err := z.parseFieldWithSchema(fieldValue.Interface(), fieldSchema, ctx)
		if err != nil {
			// Collect field validation errors with path prefix
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				for _, fieldIssue := range zodErr.Issues {
					rawIssue := issues.ConvertZodIssueToRawWithPrependedPath(fieldIssue, []any{fieldName})
					collectedIssues = append(collectedIssues, rawIssue)
				}
			} else {
				rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, fieldValue.Interface())
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
		} else {
			// Set the parsed (potentially transformed) value back to the new struct
			if err := z.setStructFieldValue(newStruct, structType, fieldName, parsedFieldValue); err != nil {
				// Failed to set field value
				rawIssue := issues.CreateIssue(core.Custom, fmt.Sprintf("Failed to set field %s: %v", fieldName, err), nil, parsedFieldValue)
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
		}
	}

	// If we collected any issues, return them as a combined error
	if len(collectedIssues) > 0 {
		return nil, issues.CreateArrayValidationIssues(collectedIssues)
	}

	return newStruct.Interface(), nil
}

// parseFieldWithSchema parses a field value.
func (z *ZodStruct[T, R]) parseFieldWithSchema(fieldValue any, fieldSchema any, ctx *core.ParseContext) (any, error) {
	if fieldSchema == nil {
		return fieldValue, nil
	}

	// Use reflection to call Parse method - this handles all schema types
	schemaVal := reflect.ValueOf(fieldSchema)
	parseMethod := schemaVal.MethodByName("Parse")

	if !parseMethod.IsValid() {
		return fieldValue, nil // Schema doesn't have Parse method, return original value
	}

	// Call Parse(fieldValue, ctx)
	args := []reflect.Value{
		reflect.ValueOf(fieldValue),
		reflect.ValueOf(ctx),
	}

	results := parseMethod.Call(args)
	if len(results) != 2 {
		return fieldValue, nil // Unexpected return signature
	}

	// Check for error (second return value)
	if !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok {
			return nil, err
		}
	}

	// Return the parsed value (first return value)
	return results[0].Interface(), nil
}

// setStructFieldValue sets a field value.
func (z *ZodStruct[T, R]) setStructFieldValue(structVal reflect.Value, structType reflect.Type, fieldName string, value any) error {
	// First try to find by field name
	fieldVal := structVal.FieldByName(fieldName)
	if fieldVal.IsValid() && fieldVal.CanSet() {
		return z.setReflectFieldValue(fieldVal, value)
	}

	// Then try to find by json tag
	for i := range structType.NumField() {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			// Parse json tag (handle omitempty, etc.)
			jsonName := jsonTagName(jsonTag)
			if jsonName == fieldName {
				fieldVal := structVal.Field(i)
				if fieldVal.CanSet() {
					return z.setReflectFieldValue(fieldVal, value)
				}
			}
		}
	}

	return ErrFieldNotFoundOrNotSettable
}

// setReflectFieldValue sets a reflect.Value field.
func (z *ZodStruct[T, R]) setReflectFieldValue(fieldVal reflect.Value, value any) error {
	if value == nil {
		// Set to zero value if nil
		fieldVal.Set(reflect.Zero(fieldVal.Type()))
		return nil
	}

	valueVal := reflect.ValueOf(value)
	if valueVal.Type().AssignableTo(fieldVal.Type()) {
		fieldVal.Set(valueVal)
		return nil
	}

	// Handle map type conversions (e.g., map[any]any to map[string]string)
	if fieldVal.Type().Kind() == reflect.Map && valueVal.Type().Kind() == reflect.Map {
		if convertedMap := z.convertMapTypes(value, fieldVal.Type()); convertedMap != nil {
			fieldVal.Set(reflect.ValueOf(convertedMap))
			return nil
		}
	}

	// Handle pointer type conversions
	if fieldVal.Type().Kind() == reflect.Pointer && valueVal.Type().Kind() != reflect.Pointer {
		// Field is pointer, value is not - create pointer to value
		if valueVal.Type().AssignableTo(fieldVal.Type().Elem()) {
			ptrVal := reflect.New(fieldVal.Type().Elem())
			ptrVal.Elem().Set(valueVal)
			fieldVal.Set(ptrVal)
			return nil
		}
		// Handle pointer to map with conversion
		if fieldVal.Type().Elem().Kind() == reflect.Map && valueVal.Type().Kind() == reflect.Map {
			if convertedMap := z.convertMapTypes(value, fieldVal.Type().Elem()); convertedMap != nil {
				ptrVal := reflect.New(fieldVal.Type().Elem())
				ptrVal.Elem().Set(reflect.ValueOf(convertedMap))
				fieldVal.Set(ptrVal)
				return nil
			}
		}
		// Handle generic type conversions for pointer fields
		if converted := z.convertValue(value, fieldVal.Type().Elem()); converted != nil {
			ptrVal := reflect.New(fieldVal.Type().Elem())
			ptrVal.Elem().Set(reflect.ValueOf(converted))
			fieldVal.Set(ptrVal)
			return nil
		}
	} else if fieldVal.Type().Kind() != reflect.Pointer && valueVal.Type().Kind() == reflect.Pointer {
		// Field is not pointer, value is pointer - dereference value
		if !valueVal.IsNil() && valueVal.Elem().Type().AssignableTo(fieldVal.Type()) {
			fieldVal.Set(valueVal.Elem())
			return nil
		}
		// Handle dereferenced pointer to map with conversion
		if !valueVal.IsNil() && fieldVal.Type().Kind() == reflect.Map && valueVal.Elem().Type().Kind() == reflect.Map {
			if convertedMap := z.convertMapTypes(valueVal.Elem().Interface(), fieldVal.Type()); convertedMap != nil {
				fieldVal.Set(reflect.ValueOf(convertedMap))
				return nil
			}
		}
		// Handle generic type conversions for dereferenced pointer values
		if !valueVal.IsNil() {
			if converted := z.convertValue(valueVal.Elem().Interface(), fieldVal.Type()); converted != nil {
				fieldVal.Set(reflect.ValueOf(converted))
				return nil
			}
		}
	}

	// Last resort: try generic conversion
	if converted := z.convertValue(value, fieldVal.Type()); converted != nil {
		fieldVal.Set(reflect.ValueOf(converted))
		return nil
	}

	return ErrCannotAssignToField
}

// convertMapTypes converts between map types.
func (z *ZodStruct[T, R]) convertMapTypes(sourceValue any, targetType reflect.Type) any {
	sourceVal := reflect.ValueOf(sourceValue)
	if sourceVal.Kind() != reflect.Map || targetType.Kind() != reflect.Map {
		return nil
	}

	targetKeyType := targetType.Key()
	targetValueType := targetType.Elem()

	// Create new map of target type
	newMap := reflect.MakeMap(targetType)

	// Convert each key-value pair
	for _, key := range sourceVal.MapKeys() {
		sourceKey := key.Interface()
		sourceMapValue := sourceVal.MapIndex(key).Interface()

		// Convert key to target key type
		convertedKey := z.convertValue(sourceKey, targetKeyType)
		if convertedKey == nil {
			return nil // Failed to convert key
		}

		// Convert value to target value type
		convertedValue := z.convertValue(sourceMapValue, targetValueType)
		if convertedValue == nil {
			return nil // Failed to convert value
		}

		newMap.SetMapIndex(reflect.ValueOf(convertedKey), reflect.ValueOf(convertedValue))
	}

	return newMap.Interface()
}

// convertValue converts a value to the target type.
func (z *ZodStruct[T, R]) convertValue(value any, targetType reflect.Type) any {
	if value == nil {
		return reflect.Zero(targetType).Interface()
	}

	valueVal := reflect.ValueOf(value)
	valueType := valueVal.Type()

	// Direct assignment if types match
	if valueType.AssignableTo(targetType) {
		return value
	}

	// Handle conversions between compatible types
	//nolint:exhaustive // Only handling specific conversion cases
	switch targetType.Kind() {
	case reflect.String:
		if valueType.Kind() == reflect.String {
			return value.(string)
		}
		// Convert interface{} to string if it contains a string
		if valueType == anyType {
			if str, ok := value.(string); ok {
				return str
			}
		}
	case reflect.Int:
		//nolint:exhaustive // Only handling specific conversion cases
		switch valueType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(valueVal.Int())
		case reflect.Float32, reflect.Float64:
			return int(valueVal.Float())
		}
		// Convert interface{} to int if it contains a number
		if valueType == anyType {
			if intVal, ok := value.(int); ok {
				return intVal
			}
			if floatVal, ok := value.(float64); ok {
				return int(floatVal)
			}
		}
	case reflect.Float64:
		//nolint:exhaustive // Only handling specific conversion cases
		switch valueType.Kind() {
		case reflect.Float32, reflect.Float64:
			return valueVal.Float()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(valueVal.Int())
		}
		// Convert interface{} to float64 if it contains a number
		if valueType == anyType {
			if floatVal, ok := value.(float64); ok {
				return floatVal
			}
			if intVal, ok := value.(int); ok {
				return float64(intVal)
			}
		}
	case reflect.Interface:
		// interface{} can accept any value
		if targetType == anyType {
			return value
		}
	case reflect.Slice:
		// Handle slice conversions (e.g., []interface{} to []SpecificType)
		if valueVal.Type().Kind() == reflect.Slice {
			return z.convertSliceTypes(value, targetType)
		}
	case reflect.Struct:
		// Handle struct conversions from maps
		if valueVal.Type().Kind() == reflect.Map {
			return z.convertMapToStruct(value, targetType)
		}
		// Handle interface{} containing map that should become struct
		if valueVal.Type() == anyType {
			if actualMap, ok := value.(map[string]any); ok {
				return z.convertMapToStruct(actualMap, targetType)
			}
		}
	default:
		// No specific conversion logic for other types
	}

	// Attempt direct conversion if possible
	if valueVal.Type().ConvertibleTo(targetType) {
		return valueVal.Convert(targetType).Interface()
	}

	return nil
}

// convertSliceTypes converts between different slice types (e.g., []any to []SpecificType)
func (z *ZodStruct[T, R]) convertSliceTypes(sourceValue any, targetType reflect.Type) any {
	sourceVal := reflect.ValueOf(sourceValue)
	if sourceVal.Kind() != reflect.Slice || targetType.Kind() != reflect.Slice {
		return nil
	}

	targetElemType := targetType.Elem()
	newSlice := reflect.MakeSlice(targetType, sourceVal.Len(), sourceVal.Len())

	for i := range sourceVal.Len() {
		sourceElem := sourceVal.Index(i).Interface()
		convertedElem := z.convertValue(sourceElem, targetElemType)
		if convertedElem == nil {
			return nil // Failed to convert element
		}
		newSlice.Index(i).Set(reflect.ValueOf(convertedElem))
	}

	return newSlice.Interface()
}

// convertMapToStruct converts a map to a struct using field matching
func (z *ZodStruct[T, R]) convertMapToStruct(sourceValue any, targetType reflect.Type) any {
	sourceVal := reflect.ValueOf(sourceValue)
	if sourceVal.Kind() != reflect.Map || targetType.Kind() != reflect.Struct {
		return nil
	}

	// Create new struct instance
	newStruct := reflect.New(targetType).Elem()

	// Convert map to struct by matching field names
	for i := range targetType.NumField() {
		field := targetType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Get field name, check json tag first, then field name
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if tagName := jsonTagName(jsonTag); tagName != "" && tagName != "-" {
				fieldName = tagName
			}
		}

		// Look for value in map by exact field name match
		var mapValue any
		var found bool

		for _, key := range sourceVal.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			if keyStr == fieldName || keyStr == field.Name {
				mapValue = sourceVal.MapIndex(key).Interface()
				found = true
				break
			}
		}

		if found && mapValue != nil {
			// Try direct assignment first
			mapValueVal := reflect.ValueOf(mapValue)
			if mapValueVal.Type().AssignableTo(field.Type) {
				newStruct.Field(i).Set(mapValueVal)
			} else {
				// Try type conversion
				convertedValue := z.convertValue(mapValue, field.Type)
				if convertedValue != nil {
					newStruct.Field(i).Set(reflect.ValueOf(convertedValue))
				}
			}
		}
	}

	return newStruct.Interface()
}

// getStructFieldValue gets the value of a struct field by name or json tag
func (z *ZodStruct[T, R]) getStructFieldValue(val reflect.Value, structType reflect.Type, fieldName string) (reflect.Value, bool) {
	// First, try to find by exact field name
	if field := val.FieldByName(fieldName); field.IsValid() {
		return field, true
	}

	// Then, try to find by json tag
	for i := range val.NumField() {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check json tag
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if jsonTagName(tag) == fieldName {
				return val.Field(i), true
			}
		}
	}

	return reflect.Value{}, false
}

// shouldSkipFieldInPartialMode checks if a field should be skipped in partial mode
func (z *ZodStruct[T, R]) shouldSkipFieldInPartialMode(fieldValue any, fieldName string) bool {
	// Only apply partial logic if we're in partial mode
	if !z.internals.IsPartial {
		return false
	}

	// If there are partial exceptions, check if this field is excepted (required)
	if z.internals.PartialExceptions != nil {
		// If field is in exceptions, it's required and should not be skipped
		if z.internals.PartialExceptions[fieldName] {
			return false
		}
	}

	// Check if the field value is a zero value
	return isZeroValue(fieldValue)
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return val.IsNil()
	case reflect.Struct:
		// For structs, compare with zero value of the same type
		zeroVal := reflect.Zero(val.Type())
		return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	default:
		// For other types, use reflect.Zero comparison
		zeroVal := reflect.Zero(val.Type())
		return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
	}
}

// isFieldOptional checks if a field schema is optional using reflection or partial state
func (z *ZodStruct[T, R]) isFieldOptional(schema any, fieldName string) bool {
	if schema == nil {
		return true
	}

	// Check if this struct is in partial mode and this field should be optional in partial mode
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
	if internalsMethod := schemaValue.MethodByName("Internals"); internalsMethod.IsValid() {
		results := internalsMethod.Call(nil)
		if len(results) > 0 {
			if internals, ok := results[0].Interface().(*core.ZodTypeInternals); ok {
				return internals.Optional
			}
		}
	}

	return false
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// newZodStructFromDef creates a new ZodStruct from definition
func newZodStructFromDef[T any, R any](def *ZodStructDef) *ZodStruct[T, R] {
	internals := &ZodStructInternals{
		ZodTypeInternals:  engine.NewBaseZodTypeInternals(def.Type),
		Def:               def,
		Shape:             def.Shape,
		IsPartial:         false,
		PartialExceptions: nil,
	}

	// Provide constructor for AddCheck functionality
	internals.Constructor = func(newDef *core.ZodTypeDef) core.ZodType[any] {
		structDef := &ZodStructDef{
			ZodTypeDef: *newDef,
			Shape:      def.Shape, // Preserve shape in constructor
		}
		return any(newZodStructFromDef[T, R](structDef)).(core.ZodType[any])
	}

	schema := &ZodStruct[T, R]{internals: internals}

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

// Struct creates a struct schema for validating Go struct instances directly
// Usage: Struct[User]() validates User struct instances
// Usage: Struct[User](core.StructSchema{...}) validates User with field schemas
func Struct[T any](params ...any) *ZodStruct[T, T] {
	// Parse first parameter as StructSchema if provided
	var shape core.StructSchema
	var remainingParams []any

	if len(params) > 0 {
		if structSchema, ok := params[0].(core.StructSchema); ok {
			shape = structSchema
			remainingParams = params[1:]
		} else {
			remainingParams = params
		}
	}

	schemaParams := utils.NormalizeParams(remainingParams...)

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: []core.ZodCheck{},
		},
		Shape: shape,
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodStructFromDef[T, T](def)
}

// StructPtr creates a struct schema for validating pointer to Go struct instances
// Usage: StructPtr[User]() validates *User (can be nil)
// Usage: StructPtr[User](core.StructSchema{...}) validates *User with field schemas
func StructPtr[T any](params ...any) *ZodStruct[T, *T] {
	// Parse first parameter as StructSchema if provided
	var shape core.StructSchema
	var remainingParams []any

	if len(params) > 0 {
		if structSchema, ok := params[0].(core.StructSchema); ok {
			shape = structSchema
			remainingParams = params[1:]
		} else {
			remainingParams = params
		}
	}

	schemaParams := utils.NormalizeParams(remainingParams...)

	def := &ZodStructDef{
		ZodTypeDef: core.ZodTypeDef{
			Type:   core.ZodTypeStruct,
			Checks: []core.ZodCheck{},
		},
		Shape: shape,
	}

	// Use utils.ApplySchemaParams for consistent parameter handling
	if schemaParams != nil {
		utils.ApplySchemaParams(&def.ZodTypeDef, schemaParams)
	}

	return newZodStructFromDef[T, *T](def)
}

// Check adds a custom validation function that can report multiple issues for struct schema.
func (z *ZodStruct[T, R]) Check(fn func(value R, payload *core.ParsePayload), params ...any) *ZodStruct[T, R] {
	wrapper := func(payload *core.ParsePayload) {
		// Attempt direct type assertion first.
		if val, ok := payload.Value().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch scenarios.
		zeroTyp := reflect.TypeFor[R]()
		if zeroTyp != nil && zeroTyp.Kind() == reflect.Pointer {
			// zeroTyp is *T, so elemTyp should be T
			elemTyp := zeroTyp.Elem()
			valRV := reflect.ValueOf(payload.Value())
			if valRV.IsValid() && valRV.Type() == elemTyp {
				// Create a new pointer to the value so it matches R
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
func (z *ZodStruct[T, R]) With(fn func(value R, payload *core.ParsePayload), params ...any) *ZodStruct[T, R] {
	return z.Check(fn, params...)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// convertMapToStructStrict converts a map to a struct with strict field matching.
func convertMapToStructStrict[T any](data map[string]any) (T, bool) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() != reflect.Struct {
		return zero, false
	}

	v := reflect.New(t).Elem()

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		key := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			tagName, _, _ := strings.Cut(tag, ",")
			if tagName != "" && tagName != "-" {
				key = tagName
			}
		}

		val, ok := data[key]
		if !ok {
			return zero, false
		}

		rv := reflect.ValueOf(val)
		if !rv.IsValid() {
			return zero, false
		}

		switch {
		case rv.Type().AssignableTo(field.Type):
			v.Field(i).Set(rv)
		case rv.Type().ConvertibleTo(field.Type):
			v.Field(i).Set(rv.Convert(field.Type))
		default:
			return zero, false
		}
	}

	return v.Interface().(T), true
}

// =============================================================================
// STRUCT TAG SUPPORT
// =============================================================================

// FromStruct creates a ZodStruct schema from struct tags
// This is a convenience function that uses the tag parsing infrastructure
func FromStruct[T any]() *ZodStruct[T, T] {
	// For now, create basic struct with minimal tag parsing
	var zero T
	structType := reflect.TypeOf(zero)

	// Check if struct has any gozod tags
	if !hasGozodTags(structType) {
		return Struct[T]()
	}

	// Parse struct tags and create schema with field validation
	fieldSchemas := parseStructTagsToSchemas(structType)
	if len(fieldSchemas) == 0 {
		return Struct[T]()
	}

	return Struct[T](fieldSchemas)
}

// FromStructPtr creates a ZodStruct schema for pointer types from struct tags
func FromStructPtr[T any]() *ZodStruct[T, *T] {
	var zero T
	structType := reflect.TypeOf(zero)

	// Check if struct has any gozod tags
	if !hasGozodTags(structType) {
		return StructPtr[T]()
	}

	// Parse struct tags and create schema with field validation
	fieldSchemas := parseStructTagsToSchemas(structType)
	if len(fieldSchemas) == 0 {
		return StructPtr[T]()
	}

	return StructPtr[T](fieldSchemas)
}

// hasGozodTags checks if a struct type has any gozod tags
func hasGozodTags(structType reflect.Type) bool {
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return false
	}

	for i := range structType.NumField() {
		field := structType.Field(i)
		if _, exists := field.Tag.Lookup("gozod"); exists {
			return true
		}
	}

	return false
}

// parseStructTagsToSchemas converts struct tags to field schemas
func parseStructTagsToSchemas(structType reflect.Type) core.StructSchema {
	// Initialize cycle detection context
	visited := make(map[reflect.Type]bool)
	return parseStructTagsToSchemasWithCycleDetection(structType, visited)
}

// parseStructTagsToSchemasWithCycleDetection parses struct tags with cycle detection
func parseStructTagsToSchemasWithCycleDetection(structType reflect.Type, visited map[reflect.Type]bool) core.StructSchema {
	schemas := make(core.StructSchema)

	// Use tagparser to parse struct tags
	parser := tagparser.New()
	fields, err := parser.ParseStructTags(structType)
	if err != nil {
		return schemas
	}

	// Mark this type as visited to detect cycles
	visited[structType] = true
	defer func() {
		// Clean up after processing
		delete(visited, structType)
	}()

	// Convert parsed fields to schemas
	for _, field := range fields {
		// Skip fields without gozod tags
		if len(field.Rules) == 0 && !field.Required {
			continue
		}

		// Create basic schema based on field type
		// Pass the field info and visited map for cycle detection
		schema := createSchemaFromTypeWithCycleDetection(field.Type, field, visited)
		if schema != nil {
			// Apply parsed tag rules
			schema = applyParsedTagRules(schema, field)
			schemas[field.JSONName] = schema
		}
	}

	return schemas
}

// createSchemaFromTypeWithCycleDetection creates a schema with cycle detection
func createSchemaFromTypeWithCycleDetection(fieldType reflect.Type, fieldInfo tagparser.FieldInfo, visited map[reflect.Type]bool) core.ZodSchema {
	// Check for circular reference
	actualType := fieldType
	isPointer := actualType.Kind() == reflect.Pointer
	if isPointer {
		actualType = actualType.Elem()
	}

	isSlice := actualType.Kind() == reflect.Slice || actualType.Kind() == reflect.Array
	if isSlice {
		actualType = actualType.Elem()
		if actualType.Kind() == reflect.Pointer {
			actualType = actualType.Elem()
		}
	}

	// If this is a struct type that we're already visiting, it's a circular reference
	if actualType.Kind() == reflect.Struct && visited[actualType] {
		// Create a lazy schema for circular reference
		return createLazySchemaForType(fieldType, fieldInfo)
	}

	// For nested structs (not circular), create schema with cycle detection
	if actualType.Kind() == reflect.Struct && actualType != reflect.TypeFor[time.Time]() {
		// Check if coercion is enabled
		hasCoerce := false
		for _, rule := range fieldInfo.Rules {
			if rule.Name == "coerce" {
				hasCoerce = true
				break
			}
		}

		if hasCoerce && actualType == reflect.TypeFor[time.Time]() {
			// Handle time.Time with coercion
			if isPointer {
				return CoercedTimePtr()
			}
			return CoercedTime()
		}

		// Create nested struct schema with cycle detection
		var schema core.ZodSchema
		if hasGozodTags(actualType) {
			// Parse nested struct tags recursively with cycle detection
			fieldSchemas := parseStructTagsToSchemasWithCycleDetection(actualType, visited)
			if len(fieldSchemas) > 0 {
				schema = Object(fieldSchemas)
			} else {
				schema = Any()
			}
		} else {
			schema = Any()
		}

		// Handle pointer/slice wrappers
		if isSlice {
			// For slices, we need to wrap the schema in a slice validator
			// Use Slice[any] for dynamic types
			sliceSchema := Slice[any](schema)
			schema = sliceSchema
		}
		if isPointer && schema != nil {
			// Make nested struct nilable if it's a pointer and not required
			if !fieldInfo.Required {
				if nilableSchema, ok := schema.(interface{ Nilable() core.ZodSchema }); ok {
					schema = nilableSchema.Nilable()
				}
			}
		}

		// Apply parsed tag rules
		return applyParsedTagRules(schema, fieldInfo)
	}

	// Otherwise, create schema normally for non-struct types
	return createSchemaFromTypeWithInfo(fieldType, fieldInfo)
}

// createLazySchemaForType creates a lazy schema for circular reference types
func createLazySchemaForType(fieldType reflect.Type, fieldInfo tagparser.FieldInfo) core.ZodSchema {
	// Store the original type for later reference
	capturedType := fieldType
	capturedInfo := fieldInfo

	// Check if this is a slice/array of circular types
	actualType := capturedType
	isPointer := actualType.Kind() == reflect.Pointer
	if isPointer {
		actualType = actualType.Elem()
	}

	isSlice := actualType.Kind() == reflect.Slice || actualType.Kind() == reflect.Array
	if isSlice {
		// For slices, we need to create a slice of lazy schemas
		elementType := actualType.Elem()
		isElementPointer := elementType.Kind() == reflect.Pointer
		if isElementPointer {
			elementType = elementType.Elem()
		}

		// Create lazy schema for the element
		lazyElementSchema := Lazy(func() core.ZodSchema {
			// Create the nested struct schema
			// The cache in parseStructTagsToSchemas will prevent infinite recursion
			var schema core.ZodSchema
			if hasGozodTags(elementType) {
				// This will use cached schemas on subsequent calls
				fieldSchemas := parseStructTagsToSchemas(elementType)
				if len(fieldSchemas) > 0 {
					schema = Object(fieldSchemas)
				} else {
					schema = Any()
				}
			} else {
				schema = Any()
			}

			// If slice element is a pointer, handle that
			if isElementPointer {
				if nilableSchema, ok := schema.(interface{ Nilable() core.ZodSchema }); ok {
					schema = nilableSchema.Nilable()
				}
			}

			return schema
		})

		// Create slice of lazy elements
		sliceSchema := Slice[any](lazyElementSchema)

		// Apply parsed tag rules
		return applyParsedTagRules(sliceSchema, capturedInfo)
	}

	// For non-slice types, create a single lazy schema
	lazySchema := Lazy(func() core.ZodSchema {
		// Create the nested struct schema
		// The cache in parseStructTagsToSchemas will prevent infinite recursion
		var schema core.ZodSchema
		if hasGozodTags(actualType) {
			// This will use cached schemas on subsequent calls
			fieldSchemas := parseStructTagsToSchemas(actualType)
			if len(fieldSchemas) > 0 {
				schema = Object(fieldSchemas)
			} else {
				schema = Any()
			}
		} else {
			schema = Any()
		}

		// Handle pointer wrapper if needed
		if isPointer && !capturedInfo.Required {
			if nilableSchema, ok := schema.(interface{ Nilable() core.ZodSchema }); ok {
				schema = nilableSchema.Nilable()
			}
		}

		// Apply parsed tag rules
		return applyParsedTagRules(schema, capturedInfo)
	})

	return lazySchema
}

// createSchemaFromTypeWithInfo creates a basic schema based on Go type with field info
func createSchemaFromTypeWithInfo(fieldType reflect.Type, fieldInfo tagparser.FieldInfo) core.ZodSchema {
	// Check if coercion is enabled
	hasCoerce := false
	for _, rule := range fieldInfo.Rules {
		if rule.Name == "coerce" {
			hasCoerce = true
			break
		}
	}

	// Handle pointer types
	isPointer := fieldType.Kind() == reflect.Pointer
	if isPointer {
		fieldType = fieldType.Elem()
	}

	var schema core.ZodSchema
	switch fieldType.Kind() {
	case reflect.String:
		if hasCoerce {
			// Use coerced constructors when coerce tag is present
			if isPointer {
				schema = CoercedStringPtr()
			} else {
				schema = CoercedString()
			}
		} else {
			if isPointer {
				schema = StringPtr()
			} else {
				schema = String()
			}
		}
	case reflect.Int:
		if hasCoerce {
			if isPointer {
				schema = CoercedIntPtr()
			} else {
				schema = CoercedInt()
			}
		} else {
			if isPointer {
				// Use pointer constructor for pointer fields
				// IntPtr() already handles nil values appropriately
				schema = IntPtr()
			} else {
				schema = Int()
			}
		}
	case reflect.Int8:
		if hasCoerce {
			if isPointer {
				schema = CoercedInt8Ptr()
			} else {
				schema = CoercedInt8()
			}
		} else {
			if isPointer {
				schema = Int8Ptr()
			} else {
				schema = Int8()
			}
		}
	case reflect.Int16:
		if hasCoerce {
			if isPointer {
				schema = CoercedInt16Ptr()
			} else {
				schema = CoercedInt16()
			}
		} else {
			if isPointer {
				schema = Int16Ptr()
			} else {
				schema = Int16()
			}
		}
	case reflect.Int32:
		if hasCoerce {
			if isPointer {
				schema = CoercedInt32Ptr()
			} else {
				schema = CoercedInt32()
			}
		} else {
			if isPointer {
				schema = Int32Ptr()
			} else {
				schema = Int32()
			}
		}
	case reflect.Int64:
		if hasCoerce {
			if isPointer {
				schema = CoercedInt64Ptr()
			} else {
				schema = CoercedInt64()
			}
		} else {
			if isPointer {
				schema = Int64Ptr()
			} else {
				schema = Int64()
			}
		}
	case reflect.Uint:
		if hasCoerce {
			if isPointer {
				schema = CoercedUintPtr()
			} else {
				schema = CoercedUint()
			}
		} else {
			if isPointer {
				schema = UintPtr()
			} else {
				schema = Uint()
			}
		}
	case reflect.Uint8:
		if hasCoerce {
			if isPointer {
				schema = CoercedUint8Ptr()
			} else {
				schema = CoercedUint8()
			}
		} else {
			if isPointer {
				schema = Uint8Ptr()
			} else {
				schema = Uint8()
			}
		}
	case reflect.Uint16:
		if hasCoerce {
			if isPointer {
				schema = CoercedUint16Ptr()
			} else {
				schema = CoercedUint16()
			}
		} else {
			if isPointer {
				schema = Uint16Ptr()
			} else {
				schema = Uint16()
			}
		}
	case reflect.Uint32:
		if hasCoerce {
			if isPointer {
				schema = CoercedUint32Ptr()
			} else {
				schema = CoercedUint32()
			}
		} else {
			if isPointer {
				schema = Uint32Ptr()
			} else {
				schema = Uint32()
			}
		}
	case reflect.Uint64:
		if hasCoerce {
			if isPointer {
				schema = CoercedUint64Ptr()
			} else {
				schema = CoercedUint64()
			}
		} else {
			if isPointer {
				schema = Uint64Ptr()
			} else {
				schema = Uint64()
			}
		}
	case reflect.Float32:
		if hasCoerce {
			if isPointer {
				schema = CoercedFloat32Ptr()
			} else {
				schema = CoercedFloat32()
			}
		} else {
			if isPointer {
				schema = Float32Ptr()
			} else {
				schema = Float32()
			}
		}
	case reflect.Float64:
		if hasCoerce {
			if isPointer {
				schema = CoercedFloat64Ptr()
			} else {
				schema = CoercedFloat64()
			}
		} else {
			if isPointer {
				schema = Float64Ptr()
			} else {
				schema = Float64()
			}
		}
	case reflect.Bool:
		if hasCoerce {
			if isPointer {
				schema = CoercedBoolPtr()
			} else {
				schema = CoercedBool()
			}
		} else {
			if isPointer {
				schema = BoolPtr()
			} else {
				schema = Bool()
			}
		}
	case reflect.Interface:
		schema = Any()
	case reflect.Slice, reflect.Array:
		// Handle slices and arrays
		elemType := fieldType.Elem()
		elemSchema := createSchemaFromType(elemType)
		if elemSchema != nil {
			if isPointer {
				// For pointer to slice (*[]T), use SlicePtr
				schema = createSlicePtrSchema(elemSchema, elemType)
			} else {
				// For regular slice ([]T), use Slice
				schema = createSliceSchema(elemSchema, elemType)
			}
		} else {
			schema = SlicePtr[any](Any())
		}
	case reflect.Map:
		// Handle maps
		valueType := fieldType.Elem()
		valueSchema := createSchemaFromType(valueType)
		if valueSchema != nil {
			if isPointer {
				// For pointer to map (*map[K]V), use MapPtr
				schema = createMapPtrSchema(valueSchema, valueType)
			} else {
				// For regular map (map[K]V), use Map
				schema = createMapSchema(valueSchema, valueType)
			}
		} else {
			schema = MapPtr(String(), Any())
		}
	case reflect.Struct:
		// Handle nested structs (including time.Time)
		if fieldType == reflect.TypeFor[time.Time]() {
			if hasCoerce {
				if isPointer {
					schema = CoercedTimePtr()
				} else {
					schema = CoercedTime()
				}
			} else {
				if isPointer {
					if fieldInfo.Required {
						schema = TimePtr()
					} else {
						schema = Time().Nilable()
					}
				} else {
					schema = Time()
				}
			}
		} else {
			// For other structs, create a nested schema
			// This should never be reached in the cycle detection path
			// as createSchemaFromTypeWithCycleDetection handles it
			schema = createNestedStructSchema(fieldType)
			if isPointer {
				// Make nested struct nilable if it's a pointer and not required
				if !fieldInfo.Required {
					if nilableSchema, ok := schema.(interface{ Nilable() core.ZodSchema }); ok {
						schema = nilableSchema.Nilable()
					}
				}
			}
		}
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Pointer, reflect.UnsafePointer:
		// Unsupported types - fallback to Any()
		schema = Any()
	default:
		schema = Any()
	}

	return schema
}

// createSchemaFromType creates a basic schema based on Go type
func createSchemaFromType(fieldType reflect.Type) core.ZodSchema {
	// This is the original function used by other places
	// Create a dummy field info that doesn't have required flag
	dummyFieldInfo := tagparser.FieldInfo{
		Required: false,
		Optional: true,
	}
	return createSchemaFromTypeWithInfo(fieldType, dummyFieldInfo)
}

// applyParsedTagRules applies validation rules from parsed tagparser.FieldInfo
func applyParsedTagRules(schema core.ZodSchema, fieldInfo tagparser.FieldInfo) core.ZodSchema {
	// We'll apply optional at the end, after all other rules

	// Apply each parsed rule
	for _, rule := range fieldInfo.Rules {
		// Handle simple rules without parameters
		switch rule.Name {
		case "required":
			// Already handled above
		case "optional":
			// Optional is handled at struct level, not individual field level
		case "coerce":
			// Coercion is already handled in createSchemaFromTypeWithInfo
		case "nilable":
			schema = applyNilableModifier(schema)
		case "email":
			// Replace string schema with email schema
			// But preserve pointer type if it's a pointer schema
			switch schema.(type) {
			case *ZodString[string]:
				schema = Email()
			case *ZodString[*string]:
				schema = EmailPtr()
			}
		case "url":
			// Replace string schema with URL schema
			switch schema.(type) {
			case *ZodString[string]:
				schema = URL()
			case *ZodString[*string]:
				schema = URLPtr()
			}
		case "uuid":
			// Replace string schema with UUID schema
			switch schema.(type) {
			case *ZodString[string]:
				schema = Uuid()
			case *ZodString[*string]:
				schema = UuidPtr()
			}
		case "ipv4":
			switch schema.(type) {
			case *ZodString[string]:
				schema = IPv4()
			case *ZodString[*string]:
				schema = IPv4Ptr()
			}
		case "ipv6":
			switch schema.(type) {
			case *ZodString[string]:
				schema = IPv6()
			case *ZodString[*string]:
				schema = IPv6Ptr()
			}
		case "cidrv4":
			switch schema.(type) {
			case *ZodString[string]:
				schema = CIDRv4()
			case *ZodString[*string]:
				schema = CIDRv4Ptr()
			}
		case "cidrv6":
			switch schema.(type) {
			case *ZodString[string]:
				schema = CIDRv6()
			case *ZodString[*string]:
				schema = CIDRv6Ptr()
			}
		case "cuid":
			switch schema.(type) {
			case *ZodString[string]:
				schema = Cuid()
			case *ZodString[*string]:
				schema = CuidPtr()
			}
		case "cuid2":
			switch schema.(type) {
			case *ZodString[string]:
				schema = Cuid2()
			case *ZodString[*string]:
				schema = Cuid2Ptr()
			}
		case "jwt":
			switch schema.(type) {
			case *ZodString[string]:
				schema = JWT()
			case *ZodString[*string]:
				schema = JWTPtr()
			}
		case "iso_datetime":
			switch schema.(type) {
			case *ZodString[string]:
				schema = IsoDateTime()
			case *ZodString[*string]:
				schema = IsoDateTimePtr()
			}
		case "iso_date":
			switch schema.(type) {
			case *ZodString[string]:
				schema = IsoDate()
			case *ZodString[*string]:
				schema = IsoDatePtr()
			}
		case "iso_time":
			switch schema.(type) {
			case *ZodString[string]:
				schema = IsoTime()
			case *ZodString[*string]:
				schema = IsoTimePtr()
			}
		case "iso_duration":
			switch schema.(type) {
			case *ZodString[string]:
				schema = IsoDuration()
			case *ZodString[*string]:
				schema = IsoDurationPtr()
			}
		case "time":
			// Special handling for time.Time fields
			schema = Time()
		case "positive":
			schema = applyPositiveModifier(schema)
		case "negative":
			schema = applyNegativeModifier(schema)
		case "finite":
			if floatSchema, ok := schema.(*ZodFloatTyped[float64, float64]); ok {
				schema = floatSchema.Finite()
			}
			if float32Schema, ok := schema.(*ZodFloatTyped[float32, float32]); ok {
				schema = float32Schema.Finite()
			}
		case "nonempty":
			if stringSchema, ok := schema.(*ZodString[string]); ok {
				schema = stringSchema.Min(1)
			} else if sliceSchema, ok := schema.(*ZodSlice[string, []string]); ok {
				schema = sliceSchema.Min(1)
			} else if sliceIntSchema, ok := schema.(*ZodSlice[int, []int]); ok {
				schema = sliceIntSchema.Min(1)
			} else if sliceAnySchema, ok := schema.(*ZodSlice[any, []any]); ok {
				schema = sliceAnySchema.Min(1)
			} else if mapStrSchema, ok := schema.(*ZodMap[map[string]string, map[string]string]); ok {
				schema = mapStrSchema.Min(1)
			} else if mapIntSchema, ok := schema.(*ZodMap[map[string]int, map[string]int]); ok {
				schema = mapIntSchema.Min(1)
			} else if mapAnySchema, ok := schema.(*ZodMap[map[string]any, map[string]any]); ok {
				schema = mapAnySchema.Min(1)
			}
		case "enum":
			// Handle enum with all parameters
			if len(rule.Params) > 0 {
				schema = applyEnumConstraint(schema, rule.Params)
			}
		case "literal":
			// Handle literal with single parameter
			if len(rule.Params) > 0 {
				schema = applyLiteralConstraint(schema, rule.Params[0])
			}
		case "default":
			// Handle default values
			if len(rule.Params) > 0 {
				schema = applyDefaultValue(schema, rule.Params[0])
			}
		case "prefault":
			// Handle prefault values
			if len(rule.Params) > 0 {
				schema = applyPrefaultValue(schema, rule.Params[0])
			}
		default:
			// Handle parameterized rules (pass only first parameter)
			if len(rule.Params) > 0 && rule.Name != "enum" && rule.Name != "literal" {
				schema = applyParameterizedRule(schema, rule.Name, rule.Params[0])
			}
		}
	}

	// Apply optional/required for pointer fields AFTER all other rules
	if fieldInfo.Type.Kind() == reflect.Pointer {
		if !fieldInfo.Required {
			// Make pointer fields optional (accept nil) unless marked as required
			// Use type switch to handle each schema type's Optional() method
			schema = applyOptionalToSchema(schema)
		}
		// If required, leave as is - pointer constructors by default don't accept nil for Parse
	}

	return schema
}

// applyNilableModifier applies nilable modifier to compatible schema types
func applyNilableModifier(schema core.ZodSchema) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodString[string]:
		return s.Nilable()
	case *ZodIntegerTyped[int, int]:
		return s.Nilable()
	case *ZodIntegerTyped[int64, int64]:
		return s.Nilable()
	case *ZodFloatTyped[float64, float64]:
		return s.Nilable()
	case *ZodFloatTyped[float32, float32]:
		return s.Nilable()
	case *ZodBool[bool]:
		return s.Nilable()
	default:
		// Try generic interface approach
		if nilableSchema, ok := schema.(interface{ Nilable() core.ZodSchema }); ok {
			return nilableSchema.Nilable()
		}
	}
	return schema
}

// applyPositiveModifier applies positive constraint to numeric types
func applyPositiveModifier(schema core.ZodSchema) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.Positive()
	case *ZodIntegerTyped[int64, int64]:
		return s.Positive()
	case *ZodFloatTyped[float64, float64]:
		return s.Positive()
	case *ZodFloatTyped[float32, float32]:
		return s.Positive()
	}
	return schema
}

// applyNegativeModifier applies negative constraint to numeric types
func applyNegativeModifier(schema core.ZodSchema) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.Negative()
	case *ZodIntegerTyped[int64, int64]:
		return s.Negative()
	case *ZodFloatTyped[float64, float64]:
		return s.Negative()
	case *ZodFloatTyped[float32, float32]:
		return s.Negative()
	}
	return schema
}

// applyParameterizedRule applies rules that have parameters
func applyParameterizedRule(schema core.ZodSchema, ruleName, param string) core.ZodSchema {
	switch ruleName {
	case "min":
		if value, err := strconv.Atoi(param); err == nil {
			schema = applyMinConstraint(schema, value)
		}
	case "max":
		if value, err := strconv.Atoi(param); err == nil {
			schema = applyMaxConstraint(schema, value)
		}
	case "length":
		if value, err := strconv.Atoi(param); err == nil {
			if stringSchema, ok := schema.(*ZodString[string]); ok {
				schema = stringSchema.Length(value)
			} else if sliceSchema, ok := schema.(*ZodSlice[string, []string]); ok {
				schema = sliceSchema.Length(value)
			} else if sliceIntSchema, ok := schema.(*ZodSlice[int, []int]); ok {
				schema = sliceIntSchema.Length(value)
			} else if sliceAnySchema, ok := schema.(*ZodSlice[any, []any]); ok {
				schema = sliceAnySchema.Length(value)
			}
		}
	case "gt":
		if value, err := strconv.ParseFloat(param, 64); err == nil {
			schema = applyGtConstraint(schema, value)
		}
	case "gte":
		if value, err := strconv.ParseFloat(param, 64); err == nil {
			schema = applyGteConstraint(schema, value)
		}
	case "lt":
		if value, err := strconv.ParseFloat(param, 64); err == nil {
			schema = applyLtConstraint(schema, value)
		}
	case "lte":
		if value, err := strconv.ParseFloat(param, 64); err == nil {
			schema = applyLteConstraint(schema, value)
		}
	case "regex":
		if stringSchema, ok := schema.(*ZodString[string]); ok {
			schema = stringSchema.RegexString(param)
		}
	case "includes":
		if stringSchema, ok := schema.(*ZodString[string]); ok {
			schema = stringSchema.Includes(param)
		}
	case "startswith":
		if stringSchema, ok := schema.(*ZodString[string]); ok {
			schema = stringSchema.StartsWith(param)
		}
	case "endswith":
		if stringSchema, ok := schema.(*ZodString[string]); ok {
			schema = stringSchema.EndsWith(param)
		}
	case "default":
		schema = applyDefaultValue(schema, param)
	case "prefault":
		schema = applyPrefaultValue(schema, param)
	case "multipleof":
		if value, err := strconv.ParseFloat(param, 64); err == nil {
			schema = applyMultipleOfConstraint(schema, value)
		}
	}

	return schema
}

// Helper functions for applying constraints
func applyMinConstraint(schema core.ZodSchema, value int) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodString[string]:
		return s.Min(value)
	case *ZodIntegerTyped[int, int]:
		return s.Min(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.Min(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.Min(float64(value))
	case *ZodFloatTyped[float32, float32]:
		return s.Min(float64(value))
	case *ZodSlice[string, []string]:
		return s.Min(value)
	case *ZodSlice[int, []int]:
		return s.Min(value)
	case *ZodSlice[any, []any]:
		return s.Min(value)
	case *ZodMap[map[string]string, map[string]string]:
		return s.Min(value)
	case *ZodMap[map[string]int, map[string]int]:
		return s.Min(value)
	case *ZodMap[map[string]any, map[string]any]:
		return s.Min(value)
	}
	return schema
}

func applyMaxConstraint(schema core.ZodSchema, value int) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodString[string]:
		return s.Max(value)
	case *ZodIntegerTyped[int, int]:
		return s.Max(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.Max(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.Max(float64(value))
	case *ZodFloatTyped[float32, float32]:
		return s.Max(float64(value))
	case *ZodSlice[string, []string]:
		return s.Max(value)
	case *ZodSlice[int, []int]:
		return s.Max(value)
	case *ZodSlice[any, []any]:
		return s.Max(value)
	case *ZodMap[map[string]string, map[string]string]:
		return s.Max(value)
	case *ZodMap[map[string]int, map[string]int]:
		return s.Max(value)
	case *ZodMap[map[string]any, map[string]any]:
		return s.Max(value)
	}
	return schema
}

func applyGtConstraint(schema core.ZodSchema, value float64) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.Gt(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.Gt(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.Gt(value)
	case *ZodFloatTyped[float32, float32]:
		return s.Gt(value)
	}
	return schema
}

func applyGteConstraint(schema core.ZodSchema, value float64) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.Gte(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.Gte(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.Gte(value)
	case *ZodFloatTyped[float32, float32]:
		return s.Gte(value)
	}
	return schema
}

func applyLtConstraint(schema core.ZodSchema, value float64) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.Lt(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.Lt(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.Lt(value)
	case *ZodFloatTyped[float32, float32]:
		return s.Lt(value)
	}
	return schema
}

func applyLteConstraint(schema core.ZodSchema, value float64) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.Lte(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.Lte(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.Lte(value)
	case *ZodFloatTyped[float32, float32]:
		return s.Lte(value)
	}
	return schema
}

func applyEnumConstraint(schema core.ZodSchema, values []string) core.ZodSchema {
	// For enums, we need to replace the schema with an enum schema
	switch schema.(type) {
	case *ZodString[string]:
		return EnumSlice(values)
	case *ZodIntegerTyped[int, int]:
		// Try to parse as integers
		intValues := make([]int, 0, len(values))
		for _, v := range values {
			if intVal, err := strconv.Atoi(v); err == nil {
				intValues = append(intValues, intVal)
			}
		}
		if len(intValues) > 0 {
			return EnumSlice(intValues)
		}
	}
	return schema
}

func applyLiteralConstraint(schema core.ZodSchema, value string) core.ZodSchema {
	// For literals, replace with a literal schema
	switch schema.(type) {
	case *ZodString[string]:
		return Literal(value)
	case *ZodIntegerTyped[int, int]:
		if intVal, err := strconv.Atoi(value); err == nil {
			return Literal(intVal)
		}
	case *ZodBool[bool]:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return Literal(boolVal)
		}
	}
	return schema
}

func applyMultipleOfConstraint(schema core.ZodSchema, value float64) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodIntegerTyped[int, int]:
		return s.MultipleOf(int64(value))
	case *ZodIntegerTyped[int64, int64]:
		return s.MultipleOf(int64(value))
	case *ZodFloatTyped[float64, float64]:
		return s.MultipleOf(value)
	case *ZodFloatTyped[float32, float32]:
		return s.MultipleOf(value)
	}
	return schema
}

func applyDefaultValue(schema core.ZodSchema, value string) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodString[string]:
		return s.Default(value)
	case *ZodString[*string]:
		return s.Default(value)
	case *ZodIntegerTyped[int, int]:
		if intVal, err := strconv.Atoi(value); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[int, *int]:
		if intVal, err := strconv.Atoi(value); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[int8, int8]:
		if intVal, err := strconv.ParseInt(value, 10, 8); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int8, *int8]:
		if intVal, err := strconv.ParseInt(value, 10, 8); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int16, int16]:
		if intVal, err := strconv.ParseInt(value, 10, 16); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int16, *int16]:
		if intVal, err := strconv.ParseInt(value, 10, 16); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int32, int32]:
		if intVal, err := strconv.ParseInt(value, 10, 32); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int32, *int32]:
		if intVal, err := strconv.ParseInt(value, 10, 32); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int64, int64]:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[int64, *int64]:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return s.Default(intVal)
		}
	case *ZodIntegerTyped[uint, uint]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Default(int64(intVal))
			}
		}
	case *ZodIntegerTyped[uint, *uint]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Default(int64(intVal))
			}
		}
	case *ZodIntegerTyped[uint8, uint8]:
		if intVal, err := strconv.ParseUint(value, 10, 8); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[uint8, *uint8]:
		if intVal, err := strconv.ParseUint(value, 10, 8); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[uint16, uint16]:
		if intVal, err := strconv.ParseUint(value, 10, 16); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[uint16, *uint16]:
		if intVal, err := strconv.ParseUint(value, 10, 16); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[uint32, uint32]:
		if intVal, err := strconv.ParseUint(value, 10, 32); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[uint32, *uint32]:
		if intVal, err := strconv.ParseUint(value, 10, 32); err == nil {
			return s.Default(int64(intVal))
		}
	case *ZodIntegerTyped[uint64, uint64]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Default(int64(intVal))
			}
		}
	case *ZodIntegerTyped[uint64, *uint64]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Default(int64(intVal))
			}
		}
	case *ZodFloatTyped[float64, float64]:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return s.Default(floatVal)
		}
	case *ZodFloatTyped[float64, *float64]:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return s.Default(floatVal)
		}
	case *ZodFloatTyped[float32, float32]:
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return s.Default(floatVal)
		}
	case *ZodFloatTyped[float32, *float32]:
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return s.Default(floatVal)
		}
	case *ZodBool[bool]:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return s.Default(boolVal)
		}
	case *ZodBool[*bool]:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return s.Default(boolVal)
		}
	case *ZodSlice[string, []string]:
		// Parse array-like default value: ['val1','val2'] or ["val1","val2"]
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			// Convert interface{} slice to string slice
			stringArray := make([]string, 0, len(defaultArray))
			for _, item := range defaultArray {
				if str, ok := item.(string); ok {
					stringArray = append(stringArray, str)
				}
			}
			if len(stringArray) > 0 {
				return s.Default(stringArray)
			}
		}
	case *ZodSlice[int, []int]:
		// Parse array of integers
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			intArray := make([]int, 0, len(defaultArray))
			for _, item := range defaultArray {
				switch v := item.(type) {
				case float64: // JSON numbers are float64
					intArray = append(intArray, int(v))
				case int:
					intArray = append(intArray, v)
				case string:
					if intVal, err := strconv.Atoi(v); err == nil {
						intArray = append(intArray, intVal)
					}
				}
			}
			if len(intArray) > 0 {
				return s.Default(intArray)
			}
		}
	case *ZodSlice[float64, []float64]:
		// Parse array of floats
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			floatArray := make([]float64, 0, len(defaultArray))
			for _, item := range defaultArray {
				switch v := item.(type) {
				case float64:
					floatArray = append(floatArray, v)
				case int:
					floatArray = append(floatArray, float64(v))
				case string:
					if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
						floatArray = append(floatArray, floatVal)
					}
				}
			}
			if len(floatArray) > 0 {
				return s.Default(floatArray)
			}
		}
	case *ZodSlice[bool, []bool]:
		// Parse array of booleans
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			boolArray := make([]bool, 0, len(defaultArray))
			for _, item := range defaultArray {
				switch v := item.(type) {
				case bool:
					boolArray = append(boolArray, v)
				case string:
					if boolVal, err := strconv.ParseBool(v); err == nil {
						boolArray = append(boolArray, boolVal)
					}
				}
			}
			if len(boolArray) > 0 {
				return s.Default(boolArray)
			}
		}
	case *ZodMap[map[string]string, map[string]string]:
		// Parse map-like default value: {"key":"val"}
		if defaultMap := parseMapDefault(value); len(defaultMap) > 0 {
			// Convert interface{} map to string map
			stringMap := make(map[string]string)
			for k, v := range defaultMap {
				if str, ok := v.(string); ok {
					stringMap[k] = str
				}
			}
			if len(stringMap) > 0 {
				return s.Default(stringMap)
			}
		}
	case *ZodMap[map[string]any, map[string]any]:
		// Parse map with any values: {"key":"val", "count":42}
		if defaultMap := parseMapDefault(value); len(defaultMap) > 0 {
			return s.Default(defaultMap)
		}

	// Pointer slice types
	case *ZodSlice[string, *[]string]:
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			stringArray := make([]string, 0, len(defaultArray))
			for _, item := range defaultArray {
				if str, ok := item.(string); ok {
					stringArray = append(stringArray, str)
				}
			}
			if len(stringArray) > 0 {
				return s.Default(stringArray)
			}
		}
	case *ZodSlice[int, *[]int]:
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			intArray := make([]int, 0, len(defaultArray))
			for _, item := range defaultArray {
				switch v := item.(type) {
				case float64:
					intArray = append(intArray, int(v))
				case int:
					intArray = append(intArray, v)
				case string:
					if intVal, err := strconv.Atoi(v); err == nil {
						intArray = append(intArray, intVal)
					}
				}
			}
			if len(intArray) > 0 {
				return s.Default(intArray)
			}
		}
	case *ZodSlice[float64, *[]float64]:
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			floatArray := make([]float64, 0, len(defaultArray))
			for _, item := range defaultArray {
				switch v := item.(type) {
				case float64:
					floatArray = append(floatArray, v)
				case int:
					floatArray = append(floatArray, float64(v))
				case string:
					if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
						floatArray = append(floatArray, floatVal)
					}
				}
			}
			if len(floatArray) > 0 {
				return s.Default(floatArray)
			}
		}
	case *ZodSlice[bool, *[]bool]:
		if defaultArray := parseArrayDefault(value); len(defaultArray) > 0 {
			boolArray := make([]bool, 0, len(defaultArray))
			for _, item := range defaultArray {
				switch v := item.(type) {
				case bool:
					boolArray = append(boolArray, v)
				case string:
					if boolVal, err := strconv.ParseBool(v); err == nil {
						boolArray = append(boolArray, boolVal)
					}
				}
			}
			if len(boolArray) > 0 {
				return s.Default(boolArray)
			}
		}

	// Pointer map types
	case *ZodMap[map[string]string, *map[string]string]:
		if defaultMap := parseMapDefault(value); len(defaultMap) > 0 {
			stringMap := make(map[string]string)
			for k, v := range defaultMap {
				if str, ok := v.(string); ok {
					stringMap[k] = str
				}
			}
			if len(stringMap) > 0 {
				return s.Default(stringMap)
			}
		}
	case *ZodMap[map[string]any, *map[string]any]:
		if defaultMap := parseMapDefault(value); len(defaultMap) > 0 {
			return s.Default(defaultMap)
		}

	// Record types for map[string]any fields
	case *ZodRecord[map[string]any, *map[string]any]:
		if defaultMap := parseMapDefault(value); len(defaultMap) > 0 {
			return s.Default(defaultMap)
		}
	}
	return schema
}

// Helper function to parse array-like default values
func parseArrayDefault(value string) []any {
	// Handle both ['val1','val2'] and JSON array format
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil
	}

	// Try JSON parsing first for complex arrays
	var jsonResult []any
	if err := json.Unmarshal([]byte(value), &jsonResult); err == nil {
		return jsonResult
	}

	// Fallback to simple parsing for backward compatibility
	value = value[1 : len(value)-1] // Remove brackets
	if value == "" {
		return []any{}
	}

	// Parse comma-separated values
	var result []any
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, char := range value {
		switch {
		case char == '\'' || char == '"':
			switch {
			case !inQuote:
				inQuote = true
				quoteChar = char
			case char == quoteChar:
				inQuote = false
				quoteChar = 0
			default:
				current.WriteRune(char)
			}
		case char == ',' && !inQuote:
			if str := strings.TrimSpace(current.String()); str != "" {
				result = append(result, str)
			}
			current.Reset()
		default:
			if inQuote || !unicode.IsSpace(char) {
				current.WriteRune(char)
			}
		}
	}

	// Add final value
	if str := strings.TrimSpace(current.String()); str != "" {
		result = append(result, str)
	}

	return result
}

// Helper function to parse map-like default values
func parseMapDefault(value string) map[string]any {
	// Handle JSON object format: {"key":"value","count":42}
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
		return nil
	}

	// Try to parse as JSON
	var result map[string]any
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return make(map[string]any)
	}

	return result
}

// Helper function to create slice schema based on element type
func createSliceSchema(elemSchema core.ZodSchema, elemType reflect.Type) core.ZodSchema {
	// Create appropriate slice schema based on element type
	switch elemType.Kind() {
	case reflect.String:
		return Slice[string](elemSchema)
	case reflect.Int:
		return Slice[int](elemSchema)
	case reflect.Int64:
		return Slice[int64](elemSchema)
	case reflect.Float64:
		return Slice[float64](elemSchema)
	case reflect.Bool:
		return Slice[bool](elemSchema)
	case reflect.Invalid, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice,
		reflect.Struct, reflect.UnsafePointer:
		return SlicePtr[any](elemSchema)
	}
	return SlicePtr[any](elemSchema) // Default fallback
}

// Helper function to create pointer slice schema based on element type
func createSlicePtrSchema(elemSchema core.ZodSchema, elemType reflect.Type) core.ZodSchema {
	// Create appropriate slice pointer schema based on element type
	switch elemType.Kind() {
	case reflect.String:
		return SlicePtr[string](elemSchema)
	case reflect.Int:
		return SlicePtr[int](elemSchema)
	case reflect.Int64:
		return SlicePtr[int64](elemSchema)
	case reflect.Float64:
		return SlicePtr[float64](elemSchema)
	case reflect.Bool:
		return SlicePtr[bool](elemSchema)
	case reflect.Invalid, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice,
		reflect.Struct, reflect.UnsafePointer:
		return SlicePtr[any](elemSchema)
	}
	return SlicePtr[any](elemSchema) // Default fallback
}

// Helper function to create map schema based on value type
func createMapSchema(valueSchema core.ZodSchema, valueType reflect.Type) core.ZodSchema {
	// Maps in Go always have string keys for JSON unmarshaling
	switch valueType.Kind() {
	case reflect.String:
		return Map(String(), valueSchema)
	case reflect.Int:
		return Map(String(), valueSchema)
	case reflect.Interface:
		return Map(String(), Any())
	case reflect.Invalid, reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice,
		reflect.Struct, reflect.UnsafePointer:
		return Map(String(), valueSchema)
	}
	return Map(String(), valueSchema) // Default fallback
}

// Helper function to create pointer map schema based on value type
func createMapPtrSchema(valueSchema core.ZodSchema, valueType reflect.Type) core.ZodSchema {
	// Maps in Go always have string keys for JSON unmarshaling
	switch valueType.Kind() {
	case reflect.String:
		return MapPtr(String(), valueSchema)
	case reflect.Int:
		return MapPtr(String(), valueSchema)
	case reflect.Interface:
		// For any values, use Record instead of Map for better type compatibility
		// Record uses map[string]any internally which matches Go's map[string]any
		return RecordTyped[map[string]any, *map[string]any](String(), Any())
	case reflect.Invalid, reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice,
		reflect.Struct, reflect.UnsafePointer:
		return MapPtr(String(), valueSchema)
	}
	return MapPtr(String(), valueSchema) // Default fallback
}

// Helper function to create nested struct schema
func createNestedStructSchema(structType reflect.Type) core.ZodSchema {
	// Check if struct has any gozod tags
	if hasGozodTags(structType) {
		// Parse nested struct tags recursively
		fieldSchemas := parseStructTagsToSchemas(structType)
		if len(fieldSchemas) > 0 {
			return Object(fieldSchemas)
		}
	}
	// For structs without tags, use Any() for now
	return Any()
}

// applyOptionalToSchema applies the Optional() method to the schema using a type switch
// This handles the type compatibility issue where each schema's Optional() method returns its specific type
// Optional() makes schemas accept nil values - it doesn't change the output type for pointer schemas
func applyOptionalToSchema(schema core.ZodSchema) core.ZodSchema {
	switch s := schema.(type) {
	// String types
	case *ZodString[string]:
		return s.Optional() // Converts to *ZodString[*string] that accepts nil
	case *ZodString[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodString[*string]
	case *ZodEmail[string]:
		return s.Optional() // Converts to *ZodEmail[*string] that accepts nil
	case *ZodEmail[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodEmail[*string]
	case *ZodURL[string]:
		return s.Optional() // Converts to *ZodURL[*string] that accepts nil
	case *ZodURL[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodURL[*string]
	case *ZodIPv4[string]:
		return s.Optional() // Converts to *ZodIPv4[*string] that accepts nil
	case *ZodIPv4[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodIPv4[*string]
	case *ZodIPv6[string]:
		return s.Optional() // Converts to *ZodIPv6[*string] that accepts nil
	case *ZodIPv6[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodIPv6[*string]
	case *ZodCIDRv4[string]:
		return s.Optional() // Converts to *ZodCIDRv4[*string] that accepts nil
	case *ZodCIDRv4[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodCIDRv4[*string]
	case *ZodCIDRv6[string]:
		return s.Optional() // Converts to *ZodCIDRv6[*string] that accepts nil
	case *ZodCIDRv6[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodCIDRv6[*string]
	case *ZodIso[string]:
		return s.Optional() // Converts to *ZodIso[*string] that accepts nil
	case *ZodIso[*string]:
		return s.Optional() // Makes it accept nil, still returns *ZodIso[*string]

	// Numeric types
	case *ZodIntegerTyped[int, int]:
		return s.Optional()
	case *ZodIntegerTyped[int, *int]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[int8, int8]:
		return s.Optional()
	case *ZodIntegerTyped[int8, *int8]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[int16, int16]:
		return s.Optional()
	case *ZodIntegerTyped[int16, *int16]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[int32, int32]:
		return s.Optional()
	case *ZodIntegerTyped[int32, *int32]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[int64, int64]:
		return s.Optional()
	case *ZodIntegerTyped[int64, *int64]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[uint, uint]:
		return s.Optional()
	case *ZodIntegerTyped[uint, *uint]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[uint8, uint8]:
		return s.Optional()
	case *ZodIntegerTyped[uint8, *uint8]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[uint16, uint16]:
		return s.Optional()
	case *ZodIntegerTyped[uint16, *uint16]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[uint32, uint32]:
		return s.Optional()
	case *ZodIntegerTyped[uint32, *uint32]:
		return s.Optional() // Makes it accept nil
	case *ZodIntegerTyped[uint64, uint64]:
		return s.Optional()
	case *ZodIntegerTyped[uint64, *uint64]:
		return s.Optional() // Makes it accept nil

	// Float types
	case *ZodFloatTyped[float32, float32]:
		return s.Optional()
	case *ZodFloatTyped[float32, *float32]:
		return s.Optional() // Makes it accept nil
	case *ZodFloatTyped[float64, float64]:
		return s.Optional()
	case *ZodFloatTyped[float64, *float64]:
		return s.Optional() // Makes it accept nil

	// Other primitive types
	case *ZodBool[bool]:
		return s.Optional()
	case *ZodBool[*bool]:
		return s.Optional() // Makes it accept nil
	case *ZodTime[time.Time]:
		return s.Optional()
	case *ZodTime[*time.Time]:
		return s.Optional() // Makes it accept nil
	case *ZodBigInt[*big.Int]:
		return s.Optional()
	case *ZodBigInt[**big.Int]:
		return s.Optional() // Makes it accept nil
	case *ZodComplex[complex64]:
		return s.Optional()
	case *ZodComplex[complex128]:
		return s.Optional()
	case *ZodComplex[*complex64]:
		return s.Optional() // Makes it accept nil
	case *ZodComplex[*complex128]:
		return s.Optional() // Makes it accept nil
	case *ZodStringBool[bool]:
		return s.Optional()
	case *ZodStringBool[*bool]:
		return s.Optional() // Makes it accept nil

	// Collection types
	case *ZodSlice[any, []any]:
		return s.Optional()
	case *ZodSlice[any, *[]any]:
		return s.Optional() // Makes it accept nil
	case *ZodArray[any, any]:
		return s.Optional()
	case *ZodArray[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodMap[any, any]:
		return s.Optional()
	case *ZodMap[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodRecord[any, any]:
		return s.Optional()
	case *ZodRecord[any, *any]:
		return s.Optional() // Makes it accept nil

	// Object/Struct types
	case *ZodObject[any, any]:
		return s.Optional()
	case *ZodObject[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodStruct[any, any]:
		return s.Optional()
	case *ZodStruct[any, *any]:
		return s.Optional() // Makes it accept nil

	// Composite types
	case *ZodUnion[any, any]:
		return s.Optional()
	case *ZodUnion[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodIntersection[any, any]:
		return s.Optional()
	case *ZodIntersection[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodDiscriminatedUnion[any, any]:
		return s.Optional()
	case *ZodDiscriminatedUnion[any, *any]:
		return s.Optional() // Makes it accept nil

	// Other types
	case *ZodEnum[any, any]:
		return s.Optional()
	case *ZodEnum[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodLiteral[any, any]:
		return s.Optional()
	case *ZodLiteral[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodAny[any, any]:
		return s.Optional()
	case *ZodAny[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodUnknown[any, any]:
		return s.Optional()
	case *ZodUnknown[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodNever[any, any]:
		return s.Optional()
	case *ZodNever[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodNil[any, any]:
		return s.Optional()
	case *ZodNil[any, *any]:
		return s.Optional() // Makes it accept nil
	case *ZodLazy[any]:
		return s.Optional()
	case *ZodLazy[*any]:
		return s.Optional() // Makes it accept nil
	case *ZodFunction[any]:
		return s.Optional()
	case *ZodFunction[*any]:
		return s.Optional() // Makes it accept nil
	case *ZodFile[any, any]:
		return s.Optional()
	case *ZodFile[any, *any]:
		return s.Optional() // Makes it accept nil

	default:
		// For any unknown types, return as-is
		// This ensures we don't break on custom schemas
		return s
	}
}

// applyPrefaultValue applies prefault values to schema types (pre-parse default with full validation)
func applyPrefaultValue(schema core.ZodSchema, value string) core.ZodSchema {
	switch s := schema.(type) {
	case *ZodString[string]:
		return s.Prefault(value)
	case *ZodString[*string]:
		return s.Prefault(value)
	case *ZodIntegerTyped[int, int]:
		if intVal, err := strconv.Atoi(value); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[int, *int]:
		if intVal, err := strconv.Atoi(value); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[int8, int8]:
		if intVal, err := strconv.ParseInt(value, 10, 8); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int8, *int8]:
		if intVal, err := strconv.ParseInt(value, 10, 8); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int16, int16]:
		if intVal, err := strconv.ParseInt(value, 10, 16); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int16, *int16]:
		if intVal, err := strconv.ParseInt(value, 10, 16); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int32, int32]:
		if intVal, err := strconv.ParseInt(value, 10, 32); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int32, *int32]:
		if intVal, err := strconv.ParseInt(value, 10, 32); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int64, int64]:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[int64, *int64]:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return s.Prefault(intVal)
		}
	case *ZodIntegerTyped[uint, uint]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Prefault(int64(intVal))
			}
		}
	case *ZodIntegerTyped[uint, *uint]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Prefault(int64(intVal))
			}
		}
	case *ZodIntegerTyped[uint8, uint8]:
		if intVal, err := strconv.ParseUint(value, 10, 8); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[uint8, *uint8]:
		if intVal, err := strconv.ParseUint(value, 10, 8); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[uint16, uint16]:
		if intVal, err := strconv.ParseUint(value, 10, 16); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[uint16, *uint16]:
		if intVal, err := strconv.ParseUint(value, 10, 16); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[uint32, uint32]:
		if intVal, err := strconv.ParseUint(value, 10, 32); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[uint32, *uint32]:
		if intVal, err := strconv.ParseUint(value, 10, 32); err == nil {
			return s.Prefault(int64(intVal))
		}
	case *ZodIntegerTyped[uint64, uint64]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Prefault(int64(intVal))
			}
		}
	case *ZodIntegerTyped[uint64, *uint64]:
		if intVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			if intVal <= 9223372036854775807 { // max int64 value
				return s.Prefault(int64(intVal))
			}
		}
	case *ZodFloatTyped[float64, float64]:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return s.Prefault(floatVal)
		}
	case *ZodFloatTyped[float64, *float64]:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return s.Prefault(floatVal)
		}
	case *ZodFloatTyped[float32, float32]:
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return s.Prefault(floatVal)
		}
	case *ZodFloatTyped[float32, *float32]:
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return s.Prefault(floatVal)
		}
	case *ZodBool[bool]:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return s.Prefault(boolVal)
		}
	case *ZodBool[*bool]:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return s.Prefault(boolVal)
		}
	case *ZodSlice[string, []string]:
		// Parse array-like prefault value: ['val1','val2'] or ["val1","val2"]
		if prefaultArray := parseArrayDefault(value); len(prefaultArray) > 0 {
			// Convert interface{} slice to string slice
			stringArray := make([]string, 0, len(prefaultArray))
			for _, item := range prefaultArray {
				if str, ok := item.(string); ok {
					stringArray = append(stringArray, str)
				}
			}
			if len(stringArray) > 0 {
				return s.Prefault(stringArray)
			}
		}
	case *ZodSlice[int, []int]:
		// Parse array of integers
		if prefaultArray := parseArrayDefault(value); len(prefaultArray) > 0 {
			intArray := make([]int, 0, len(prefaultArray))
			for _, item := range prefaultArray {
				switch v := item.(type) {
				case float64: // JSON numbers are float64
					intArray = append(intArray, int(v))
				case int:
					intArray = append(intArray, v)
				case string:
					if intVal, err := strconv.Atoi(v); err == nil {
						intArray = append(intArray, intVal)
					}
				}
			}
			if len(intArray) > 0 {
				return s.Prefault(intArray)
			}
		}
	case *ZodMap[map[string]string, map[string]string]:
		// Parse map-like prefault value: {"key":"val"}
		if prefaultMap := parseMapDefault(value); len(prefaultMap) > 0 {
			// Convert interface{} map to string map
			stringMap := make(map[string]string)
			for k, v := range prefaultMap {
				if str, ok := v.(string); ok {
					stringMap[k] = str
				}
			}
			if len(stringMap) > 0 {
				return s.Prefault(stringMap)
			}
		}
	}
	return schema
}
