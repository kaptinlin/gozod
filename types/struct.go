package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
	"github.com/kaptinlin/gozod/internal/engine"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/internal/utils"
)

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

// GetInternals exposes internal state for framework usage
func (z *ZodStruct[T, R]) GetInternals() *core.ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// IsOptional returns true if this schema accepts undefined/missing values
func (z *ZodStruct[T, R]) IsOptional() bool {
	return z.internals.ZodTypeInternals.IsOptional()
}

// IsNilable returns true if this schema accepts nil values
func (z *ZodStruct[T, R]) IsNilable() bool {
	return z.internals.ZodTypeInternals.IsNilable()
}

// Shape returns the struct field schemas for JSON Schema conversion
func (z *ZodStruct[T, R]) Shape() core.StructSchema {
	return z.internals.Shape
}

// GetUnknownKeys returns the unknown keys handling mode for JSON Schema conversion
// Structs are strict by default (don't allow additional properties)
func (z *ZodStruct[T, R]) GetUnknownKeys() string {
	return "strict"
}

// GetCatchall returns nil since structs don't support catchall by default
func (z *ZodStruct[T, R]) GetCatchall() core.ZodSchema {
	return nil
}

// Parse validates input using struct-specific parsing logic with engine.ParseComplex
func (z *ZodStruct[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	parseCtx := core.NewParseContext()
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	}

	// Check if constraint type R is a pointer type
	var zeroR R
	isPointerConstraint := reflect.TypeOf(zeroR).Kind() == reflect.Ptr

	// Temporarily enable Optional flag for pointer constraint types to ensure pointer identity preservation in ParseComplex
	// But only if no other modifiers (Optional, Nilable, Prefault) are set
	originalInternals := &z.internals.ZodTypeInternals
	if isPointerConstraint &&
		!originalInternals.Optional &&
		!originalInternals.Nilable &&
		originalInternals.PrefaultValue == nil &&
		originalInternals.PrefaultFunc == nil {
		// Create a copy of internals with Optional flag temporarily enabled
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
		var zero R
		// Check if this is a generic struct type error that we can improve
		if errStr := err.Error(); strings.Contains(errStr, "Invalid input: expected struct, received") {
			// Generate more specific error with actual type information
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
			var zero R
			return zero, nil
		}
		return convertToStructConstraintType[T, R](*structPtr), nil
	}

	// Handle nil result for optional/nilable schemas
	if result == nil {
		var zero R
		return zero, nil
	}

	// This should not happen in well-formed schemas
	var zero R
	return zero, issues.CreateTypeConversionError(fmt.Sprintf("%T", result), "struct", input, parseCtx)
}

// MustParse validates the input value and panics on failure
func (z *ZodStruct[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// StrictParse provides compile-time type safety by requiring exact type matching.
// This eliminates runtime type checking overhead for maximum performance.
// The input must exactly match the schema's constraint type T.
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

// MustStrictParse validates input with compile-time type safety and panics on failure.
// This method provides zero-overhead abstraction with strict type constraints.
func (z *ZodStruct[T, R]) MustStrictParse(input T, ctx ...*core.ParseContext) R {
	result, err := z.StrictParse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseAny validates the input value and returns any type (for runtime interface)
func (z *ZodStruct[T, R]) ParseAny(input any, ctx ...*core.ParseContext) (any, error) {
	return z.Parse(input, ctx...)
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional always returns *T constraint because the optional value may be nil.
func (z *ZodStruct[T, R]) Optional() *ZodStruct[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	return z.withPtrInternals(in)
}

// Nilable always returns *T constraint because the value may be nil.
func (z *ZodStruct[T, R]) Nilable() *ZodStruct[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// Nullish combines optional and nilable modifiers for maximum flexibility
func (z *ZodStruct[T, R]) Nullish() *ZodStruct[T, *T] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetOptional(true)
	in.SetNilable(true)
	return z.withPtrInternals(in)
}

// NonOptional removes Optional flag and enforces non-nil struct value (T).
func (z *ZodStruct[T, R]) NonOptional() *ZodStruct[T, T] {
	in := z.internals.ZodTypeInternals.Clone()
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

// Default keeps the current generic constraint type R.
func (z *ZodStruct[T, R]) Default(v T) *ZodStruct[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultValue(v)
	return z.withInternals(in)
}

// DefaultFunc keeps the current generic constraint type R.
func (z *ZodStruct[T, R]) DefaultFunc(fn func() T) *ZodStruct[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetDefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Prefault provides fallback values on validation failure
func (z *ZodStruct[T, R]) Prefault(v T) *ZodStruct[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultValue(v)
	return z.withInternals(in)
}

// PrefaultFunc provides dynamic fallback values
func (z *ZodStruct[T, R]) PrefaultFunc(fn func() T) *ZodStruct[T, R] {
	in := z.internals.ZodTypeInternals.Clone()
	in.SetPrefaultFunc(func() any {
		return fn()
	})
	return z.withInternals(in)
}

// Meta stores metadata for this struct schema.
func (z *ZodStruct[T, R]) Meta(meta core.GlobalMeta) *ZodStruct[T, R] {
	core.GlobalRegistry.Add(z, meta)
	return z
}

// =============================================================================
// TRANSFORMATION AND PIPELINE METHODS
// =============================================================================

// Transform creates a type-safe transformation using WrapFn pattern
func (z *ZodStruct[T, R]) Transform(fn func(T, *core.RefinementContext) (any, error)) *core.ZodTransform[R, any] {
	wrapperFn := func(input R, ctx *core.RefinementContext) (any, error) {
		structValue := extractStructValue[T, R](input)
		return fn(structValue, ctx)
	}
	return core.NewZodTransform[R, any](z, wrapperFn)
}

// Overwrite transforms the input value while preserving the original type.
// Unlike Transform, this method doesn't change the inferred type and returns an instance of the original class.
// The transformation function is stored as a check, so it doesn't modify the inferred type.
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
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// Pipe creates a pipeline using WrapFn pattern
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

func (z *ZodStruct[T, R]) Refine(fn func(R) bool, params ...any) *ZodStruct[T, R] {
	// Wrapper converts the raw value to R before calling fn
	wrapper := func(v any) bool {
		// Convert value to constraint type R and call the refinement function
		if constraintValue, ok := convertToConstraintValue[T, R](v); ok {
			return fn(constraintValue)
		}
		return false
	}

	checkParams := checks.NormalizeCheckParams(params...)
	check := checks.NewCustom[any](wrapper, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

func (z *ZodStruct[T, R]) RefineAny(fn func(any) bool, params ...any) *ZodStruct[T, R] {
	checkParams := checks.NormalizeCheckParams(params...)
	check := checks.NewCustom[any](fn, checkParams)
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
}

// =============================================================================
// TYPE-SPECIFIC METHODS
// =============================================================================

// Extend creates a new struct by extending with additional field schemas
func (z *ZodStruct[T, R]) Extend(augmentation core.StructSchema, params ...any) *ZodStruct[T, R] {
	// Create new shape combining existing + extension fields
	newShape := make(core.StructSchema)

	// Copy existing shape
	for k, v := range z.internals.Shape {
		newShape[k] = v
	}

	// Add augmentation fields
	for k, schema := range augmentation {
		newShape[k] = schema
	}

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

// Partial makes all fields optional by allowing zero values
func (z *ZodStruct[T, R]) Partial(keys ...[]string) *ZodStruct[T, R] {
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

	return &ZodStruct[T, R]{internals: &ZodStructInternals{
		ZodTypeInternals:  *newInternals,
		Def:               z.internals.Def,
		Shape:             z.internals.Shape,
		IsPartial:         true,
		PartialExceptions: partialExceptions,
	}}
}

// Required makes all fields required (opposite of Partial)
func (z *ZodStruct[T, R]) Required(fields ...[]string) *ZodStruct[T, R] {
	newInternals := z.internals.ZodTypeInternals.Clone()

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

// CloneFrom allows copying configuration from a source
func (z *ZodStruct[T, R]) CloneFrom(source any) {
	if src, ok := source.(*ZodStruct[T, R]); ok {
		originalChecks := z.internals.ZodTypeInternals.Checks
		*z.internals = *src.internals
		z.internals.ZodTypeInternals.Checks = originalChecks
	}
}

// =============================================================================
// TYPE CONVERSION HELPERS
// =============================================================================

// convertToStructConstraintType converts a base type T to constraint type R
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

// extractStructValue extracts the base type T from constraint type R
func extractStructValue[T any, R any](value R) T {
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

// convertToConstraintValue converts any value to constraint type R if possible
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

// convertToStructType converts any value to the struct constraint type R with strict type checking
func convertToStructType[T any, R any](v any) (R, bool) {
	var zero R

	if v == nil {
		// Handle nil values for pointer types
		zeroType := reflect.TypeOf((*R)(nil)).Elem()
		if zeroType.Kind() == reflect.Ptr {
			return zero, true // zero value for pointer types is nil
		}
		return zero, false // nil not allowed for value types
	}

	// Try direct conversion first
	if converted, ok := any(v).(R); ok { //nolint:unconvert
		return converted, true
	}

	// Handle pointer conversion for different types
	zeroType := reflect.TypeOf((*R)(nil)).Elem()
	if zeroType.Kind() == reflect.Ptr {
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

// extractStructForEngine extracts struct value T from input for engine processing
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
	if val := reflect.ValueOf(input); val.Kind() == reflect.Ptr && !val.IsNil() {
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

// createStructTypeError creates a specific struct type error with concrete type information
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

// extractStructPtrForEngine extracts pointer to struct *T from input for engine processing
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
	if val := reflect.ValueOf(input); val.Kind() == reflect.Ptr && !val.IsNil() {
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

// validateStructForEngine validates struct value using schema checks
func (z *ZodStruct[T, R]) validateStructForEngine(value T, checks []core.ZodCheck, ctx *core.ParseContext) (T, error) {
	// Validate struct fields if schema is defined
	if z.internals.Shape != nil && len(z.internals.Shape) > 0 {
		if err := z.validateStructFields(any(value), ctx); err != nil {
			return value, err
		}
	}

	// Apply validation checks
	if len(checks) > 0 {
		transformedValue, err := engine.ApplyChecks(value, checks, ctx)
		if err != nil {
			return value, err
		}

		// Use reflection to check if transformed value can be converted to T
		transformedVal := reflect.ValueOf(transformedValue)
		if transformedVal.IsValid() && !transformedVal.IsZero() {
			if transformedVal.Type().AssignableTo(reflect.TypeOf(value)) {
				return transformedVal.Interface().(T), nil
			}
		}
	}

	return value, nil
}

// =============================================================================
// VALIDATION LOGIC
// =============================================================================

// validateStructFields validates struct fields against the defined schema with multiple error collection (TypeScript Zod v4 behavior adapted for Go structs)
func (z *ZodStruct[T, R]) validateStructFields(input any, ctx *core.ParseContext) error {
	if z.internals.Shape == nil || len(z.internals.Shape) == 0 {
		return nil // No field schemas defined
	}

	var collectedIssues []core.ZodRawIssue

	// Use reflection to access struct fields
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return issues.CreateInvalidTypeError(core.ZodTypeStruct, input, ctx)
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return issues.CreateInvalidTypeError(core.ZodTypeStruct, input, ctx)
	}

	structType := val.Type()

	// Validate each field defined in the schema and collect all errors (TypeScript Zod v4 behavior)
	for fieldName, fieldSchema := range z.internals.Shape {
		if fieldSchema == nil {
			continue // Skip nil schemas
		}

		// Find the struct field (check both field name and json tag)
		fieldValue, found := z.getStructFieldValue(val, structType, fieldName)
		if !found {
			// Field not found in struct
			if !z.isFieldOptional(fieldSchema, fieldName) {
				// Create missing required field issue
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

		// Validate field directly to preserve original error codes
		if err := z.validateFieldDirect(fieldValue.Interface(), fieldSchema, ctx); err != nil {
			// Collect field validation errors with path prefix (TypeScript Zod v4 behavior adapted for Go structs)
			var zodErr *issues.ZodError
			if errors.As(err, &zodErr) {
				// Propagate all issues from field validation with path prefix
				for _, fieldIssue := range zodErr.Issues {
					// Create raw issue preserving original code and essential properties
					rawIssue := core.ZodRawIssue{
						Code:       fieldIssue.Code,
						Message:    fieldIssue.Message,
						Input:      fieldIssue.Input,
						Path:       append([]any{fieldName}, fieldIssue.Path...), // Prepend field name to path
						Properties: make(map[string]any),
					}
					// Copy essential properties from ZodIssue to ZodRawIssue
					if fieldIssue.Minimum != nil {
						rawIssue.Properties["minimum"] = fieldIssue.Minimum
					}
					if fieldIssue.Maximum != nil {
						rawIssue.Properties["maximum"] = fieldIssue.Maximum
					}
					if fieldIssue.Expected != "" {
						rawIssue.Properties["expected"] = fieldIssue.Expected
					}
					if fieldIssue.Received != "" {
						rawIssue.Properties["received"] = fieldIssue.Received
					}
					rawIssue.Properties["inclusive"] = fieldIssue.Inclusive
					collectedIssues = append(collectedIssues, rawIssue)
				}
			} else {
				// Handle non-ZodError by creating a raw issue with field path
				rawIssue := issues.CreateIssue(core.Custom, err.Error(), nil, fieldValue.Interface())
				rawIssue.Path = []any{fieldName}
				collectedIssues = append(collectedIssues, rawIssue)
			}
		}
	}

	// If we collected any issues, return them as a combined error
	if len(collectedIssues) > 0 {
		return issues.CreateArrayValidationIssues(collectedIssues)
	}

	return nil
}

// getStructFieldValue gets the value of a struct field by name or json tag
func (z *ZodStruct[T, R]) getStructFieldValue(val reflect.Value, structType reflect.Type, fieldName string) (reflect.Value, bool) {
	// First, try to find by exact field name
	if field := val.FieldByName(fieldName); field.IsValid() {
		return field, true
	}

	// Then, try to find by json tag
	for i := 0; i < val.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check json tag
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			tagName := tag
			if commaIdx := strings.Index(tag, ","); commaIdx > 0 {
				tagName = tag[:commaIdx]
			}
			if tagName == fieldName {
				return val.Field(i), true
			}
		}
	}

	return reflect.Value{}, false
}

// validateFieldDirect validates a single field using its schema without wrapping errors (preserves original error codes)
func (z *ZodStruct[T, R]) validateFieldDirect(element any, schema any, ctx *core.ParseContext) error {
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
	args := []reflect.Value{reflect.ValueOf(element)}
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
	case reflect.Ptr, reflect.Interface:
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
		if val, ok := payload.GetValue().(R); ok {
			fn(val, payload)
			return
		}

		// Handle pointer/value mismatch scenarios.
		var zero R
		zeroTyp := reflect.TypeOf(zero)
		if zeroTyp != nil && zeroTyp.Kind() == reflect.Ptr {
			// zeroTyp is *T, so elemTyp should be T
			elemTyp := zeroTyp.Elem()
			valRV := reflect.ValueOf(payload.GetValue())
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
	newInternals := z.internals.ZodTypeInternals.Clone()
	newInternals.AddCheck(check)
	return z.withInternals(newInternals)
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

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		key := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName != "" && tagName != "-" {
				key = tagName
			}
		}

		val, ok := data[key]
		if !ok {
			// missing field
			return zero, false
		}

		rv := reflect.ValueOf(val)
		if !rv.IsValid() {
			return zero, false
		}

		if rv.Type().AssignableTo(field.Type) {
			v.Field(i).Set(rv)
		} else if rv.Type().ConvertibleTo(field.Type) {
			v.Field(i).Set(rv.Convert(field.Type))
		} else {
			return zero, false
		}
	}

	return v.Interface().(T), true
}
