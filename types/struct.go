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

// ZodStructDef defines the schema definition for struct validation
type ZodStructDef struct {
	core.ZodTypeDef
	Shape core.StructSchema // Field schemas for struct validation
}

// ZodStructInternals contains the internal state for struct schema
type ZodStructInternals struct {
	core.ZodTypeInternals
	Def   *ZodStructDef     // Schema definition reference
	Shape core.StructSchema // Field schemas for runtime validation
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

// Parse validates input using struct-specific parsing logic
func (z *ZodStruct[T, R]) Parse(input any, ctx ...*core.ParseContext) (R, error) {
	var parseCtx *core.ParseContext
	if len(ctx) > 0 && ctx[0] != nil {
		parseCtx = ctx[0]
	} else {
		parseCtx = &core.ParseContext{}
	}

	// Handle nil input for optional/nilable schemas or pointer constraint types
	if input == nil {
		var zero R
		// Check if R is a pointer type (like *T)
		var zeroR R
		_, isConstraintPtr := any(zeroR).(*T)

		if z.internals.Optional || z.internals.Nilable || isConstraintPtr {
			return zero, nil
		}
		return zero, fmt.Errorf("struct value cannot be nil")
	}

	return z.parseGoStruct(input, z.internals.Checks, parseCtx)
}

// MustParse validates the input value and panics on failure
func (z *ZodStruct[T, R]) MustParse(input any, ctx ...*core.ParseContext) R {
	result, err := z.Parse(input, ctx...)
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
// HELPER METHODS
// =============================================================================

func (z *ZodStruct[T, R]) withPtrInternals(in *core.ZodTypeInternals) *ZodStruct[T, *T] {
	return &ZodStruct[T, *T]{internals: &ZodStructInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Shape:            z.internals.Shape, // Preserve shape
	}}
}

func (z *ZodStruct[T, R]) withInternals(in *core.ZodTypeInternals) *ZodStruct[T, R] {
	return &ZodStruct[T, R]{internals: &ZodStructInternals{
		ZodTypeInternals: *in,
		Def:              z.internals.Def,
		Shape:            z.internals.Shape, // Preserve shape
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
// VALIDATION LOGIC
// =============================================================================

// parseGoStruct handles direct Go struct validation
func (z *ZodStruct[T, R]) parseGoStruct(input any, checks []core.ZodCheck, ctx *core.ParseContext) (R, error) {
	// Allow map inputs by converting them to struct T when possible
	switch m := input.(type) {
	case map[string]any:
		if converted, ok := convertMapToStructStrict[T](m); ok {
			input = converted
		}
	case map[any]any:
		strMap := make(map[string]any)
		for k, v := range m {
			if ks, ok := k.(string); ok {
				strMap[ks] = v
			}
		}
		if converted, ok := convertMapToStructStrict[T](strMap); ok {
			input = converted
		}
	}

	var zero R

	// Check if the constraint type R is a pointer type
	var isConstraintPtr bool
	var zeroR R
	_, isConstraintPtr = any(zeroR).(*T)

	if isConstraintPtr {
		// R is *T, so we expect pointer input or we can accept T and convert to *T
		if input == nil {
			return zero, nil // Return nil pointer
		}

		// If input is a pointer, validate it
		val := reflect.ValueOf(input)
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return zero, nil // Return nil pointer
			}

			// Get the underlying struct value for validation
			structValue := val.Elem().Interface()
			if !z.validateGoStructType(structValue) {
				return zero, fmt.Errorf("expected pointer to struct of type %T, got %T", *new(T), input)
			}

			// Validate struct fields if schema is defined
			if z.internals.Shape != nil && len(z.internals.Shape) > 0 {
				if err := z.validateStructFields(structValue, ctx); err != nil {
					return zero, err
				}
			}

			// Run validation checks on the underlying struct
			if len(checks) > 0 {
				transformedValue, err := engine.ApplyChecks(structValue, checks, ctx)
				if err != nil {
					return zero, err
				}
				// For Go struct types, we need to handle transformation carefully
				if transformedStruct, ok := transformedValue.(T); ok {
					structValue = transformedStruct
					// Create new pointer for the transformed struct
					return any(&structValue).(R), nil //nolint:unconvert
				}
			}

			// Return the original pointer
			return any(input).(R), nil //nolint:unconvert
		} else {
			// Input is not a pointer, validate as T and convert to *T
			if !z.validateGoStructType(input) {
				return zero, fmt.Errorf("expected struct of type %T, got %T", *new(T), input)
			}

			// Validate struct fields if schema is defined
			if z.internals.Shape != nil && len(z.internals.Shape) > 0 {
				if err := z.validateStructFields(input, ctx); err != nil {
					return zero, err
				}
			}

			// Run validation checks
			structValue := input
			if len(checks) > 0 {
				transformedValue, err := engine.ApplyChecks(structValue, checks, ctx)
				if err != nil {
					return zero, err
				}
				if transformedStruct, ok := transformedValue.(T); ok {
					structValue = transformedStruct
				}
			}

			// Convert to pointer
			return convertToStructConstraintType[T, R](any(structValue).(T)), nil //nolint:unconvert
		}
	} else {
		// R is T, so we expect direct struct input or pointer that we can dereference
		if input == nil {
			return zero, fmt.Errorf("struct cannot be nil")
		}

		// Handle pointer input - dereference to get T
		val := reflect.ValueOf(input)
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return zero, fmt.Errorf("struct cannot be nil")
			}
			input = val.Elem().Interface()
		}

		// Validate that input is of the expected type T
		if !z.validateGoStructType(input) {
			return zero, fmt.Errorf("expected struct of type %T, got %T", *new(T), input)
		}

		// Validate struct fields if schema is defined
		if z.internals.Shape != nil && len(z.internals.Shape) > 0 {
			if err := z.validateStructFields(input, ctx); err != nil {
				return zero, err
			}
		}

		// Run validation checks using the engine
		if len(checks) > 0 {
			transformedValue, err := engine.ApplyChecks(input, checks, ctx)
			if err != nil {
				return zero, err
			}
			// For Go struct types, use the transformed value if it's the right type
			if transformedStruct, ok := transformedValue.(T); ok {
				input = transformedStruct
			}
		}

		// Convert to constraint type R
		return convertToStructConstraintType[T, R](any(input).(T)), nil //nolint:unconvert
	}
}

// validateGoStructType validates that the input is of the expected Go struct type
func (z *ZodStruct[T, R]) validateGoStructType(input any) bool {
	if input == nil {
		return false
	}

	// Get the type of T
	var zeroT T
	expectedType := reflect.TypeOf(zeroT)
	actualType := reflect.TypeOf(input)

	// Check if types match
	return actualType == expectedType
}

// validateStructFields validates struct fields against the defined schema
func (z *ZodStruct[T, R]) validateStructFields(input any, ctx *core.ParseContext) error {
	if z.internals.Shape == nil || len(z.internals.Shape) == 0 {
		return nil // No field schemas defined
	}

	// Use reflection to access struct fields
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("cannot validate fields of nil struct")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("input is not a struct, got %T", input)
	}

	structType := val.Type()

	// Validate each field defined in the schema
	for fieldName, fieldSchema := range z.internals.Shape {
		if fieldSchema == nil {
			continue // Skip nil schemas
		}

		// Find the struct field (check both field name and json tag)
		fieldValue, found := z.getStructFieldValue(val, structType, fieldName)
		if !found {
			// Field not found in struct
			if !z.isFieldOptional(fieldSchema) {
				return fmt.Errorf("required field '%s' not found in struct", fieldName)
			}
			continue
		}

		// Validate the field value using its schema
		if err := z.validateField(fieldValue.Interface(), fieldSchema, ctx, fieldName); err != nil {
			return fmt.Errorf("field '%s' validation failed: %w", fieldName, err)
		}
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

// validateField validates a single field using its schema (similar to object.go)
func (z *ZodStruct[T, R]) validateField(element any, schema any, ctx *core.ParseContext, fieldName string) error {
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
				return err
			}
		}
	}

	return nil
}

// isFieldOptional checks if a field schema is optional using reflection
func (z *ZodStruct[T, R]) isFieldOptional(schema any) bool {
	if schema == nil {
		return true
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
		ZodTypeInternals: engine.NewBaseZodTypeInternals(def.Type),
		Def:              def,
		Shape:            def.Shape, // Copy shape to internals
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
