package gozod

import (
	"errors"
	"reflect"
)

// =============================================================================
// OBJECT MODE CONSTANTS
// =============================================================================

// ObjectMode defines how to handle unknown keys in object validation
type ObjectMode string

const (
	OBJECT_STRICT_MODE ObjectMode = "strict" // Error on unknown keys
	OBJECT_STRIP_MODE  ObjectMode = "strip"  // Strip unknown keys (default)
	OBJECT_LOOSE_MODE  ObjectMode = "loose"  // Allow unknown keys
)

// =============================================================================
// OBJECT TYPE DEFINITIONS (Three-Layer Architecture)
// =============================================================================

// ZodObjectDef defines object validation configuration
type ZodObjectDef struct {
	ZodTypeDef
	Type        string            // "object"
	Shape       ObjectSchema      // Field definitions
	Catchall    ZodType[any, any] // Catchall schema for unknown keys
	UnknownKeys ObjectMode        // How to handle unknown keys
}

// ZodObjectInternals contains object validator internal state
type ZodObjectInternals struct {
	ZodTypeInternals
	Def   *ZodObjectDef          // Schema definition
	Shape ObjectSchema           // Field definitions map
	Bag   map[string]interface{} // Runtime configuration
}

// ZodObject represents fixed-field object validation
type ZodObject struct {
	internals *ZodObjectInternals
}

// Shape provides access to the internal field schemas
func (z *ZodObject) Shape() ObjectSchema {
	return z.internals.Shape
}

// =============================================================================
// CONSTRUCTOR FUNCTIONS
// =============================================================================

// Object creates object schema for fixed-field validation
func Object(shape ObjectSchema, params ...SchemaParams) *ZodObject {
	def := &ZodObjectDef{
		ZodTypeDef: ZodTypeDef{
			Type:   "object",
			Checks: make([]ZodCheck, 0),
		},
		Type:        "object",
		Shape:       shape,
		Catchall:    nil,
		UnknownKeys: OBJECT_STRIP_MODE, // Default mode
	}

	schema := createZodObjectFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]

		// Store coerce flag in bag
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
		}

		// Handle schema-level error configuration
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		// Store additional parameters
		if param.Params != nil {
			for key, value := range param.Params {
				schema.internals.Bag[key] = value
			}
		}
	}

	return schema
}

// StrictObject creates strict object (disallows unknown keys)
func StrictObject(shape ObjectSchema, params ...SchemaParams) *ZodObject {
	schema := Object(shape, params...)
	schema.internals.Def.UnknownKeys = OBJECT_STRICT_MODE
	return schema
}

// LooseObject creates loose object (allows unknown keys)
func LooseObject(shape ObjectSchema, params ...SchemaParams) *ZodObject {
	schema := Object(shape, params...)
	schema.internals.Def.UnknownKeys = OBJECT_LOOSE_MODE
	return schema
}

// NewZodObject creates a new object schema with full configuration
func NewZodObject(shape ObjectSchema, params ...SchemaParams) *ZodObject {
	return Object(shape, params...)
}

// =============================================================================
// OBJECT UTILITY FUNCTIONS
// =============================================================================

// extractObjectValue extracts object value handling nil cases
func extractObjectValue(input any) (interface{}, bool, error) {
	if input == nil {
		return nil, true, nil
	}

	// Handle pointer types
	if reflect.TypeOf(input).Kind() == reflect.Ptr {
		v := reflect.ValueOf(input)
		if v.IsNil() {
			return nil, true, nil
		}
		elem := v.Elem()
		// Check if it's a map or struct
		if elem.Kind() == reflect.Map || elem.Kind() == reflect.Struct {
			return elem.Interface(), false, nil
		}
	}

	// Handle direct object types (map or struct)
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Map || v.Kind() == reflect.Struct {
		return input, false, nil
	}

	rawIssue := CreateInvalidTypeIssue(input, "object", string(GetParsedType(input)))
	finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
	return nil, false, NewZodError([]ZodIssue{finalIssue})
}

// convertObjectToMap converts various object types to map[string]interface{}
func convertObjectToMap(input interface{}) map[string]interface{} {
	// Use existing convertToMap from utils.go for consistent handling
	return convertToMap(input)
}

// convertMapToObjectType converts map back to the original object type
func convertMapToObjectType(validatedMap map[string]interface{}, originalType reflect.Type) interface{} {
	switch originalType.Kind() {
	case reflect.Invalid:
		return validatedMap // invalid types return as-is
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice, reflect.String,
		reflect.UnsafePointer:
		return validatedMap // non-convertible types return as-is
	case reflect.Ptr:
		// Handle pointer to map/struct
		elemType := originalType.Elem()
		if elemType.Kind() == reflect.Map && elemType.Key().Kind() == reflect.String {
			// Convert to original map type
			result := reflect.MakeMap(elemType)
			for k, v := range validatedMap {
				key := reflect.ValueOf(k)
				if v == nil {
					// Handle nil values explicitly
					result.SetMapIndex(key, reflect.Zero(elemType.Elem()))
				} else {
					value := reflect.ValueOf(v)
					if value.Type().ConvertibleTo(elemType.Elem()) {
						result.SetMapIndex(key, value.Convert(elemType.Elem()))
					} else {
						result.SetMapIndex(key, value)
					}
				}
			}
			ptr := reflect.New(elemType)
			ptr.Elem().Set(result)
			return ptr.Interface()
		} else if elemType.Kind() == reflect.Struct {
			// Convert to struct
			if structResult, err := mapToStruct(validatedMap, elemType); err == nil {
				ptr := reflect.New(elemType)
				ptr.Elem().Set(reflect.ValueOf(structResult))
				return ptr.Interface()
			}
		}
	case reflect.Map:
		if originalType.Key().Kind() == reflect.String {
			// Handle direct map
			result := reflect.MakeMap(originalType)
			for k, v := range validatedMap {
				key := reflect.ValueOf(k)
				if v == nil {
					// Handle nil values explicitly
					result.SetMapIndex(key, reflect.Zero(originalType.Elem()))
				} else {
					value := reflect.ValueOf(v)
					if value.Type().ConvertibleTo(originalType.Elem()) {
						result.SetMapIndex(key, value.Convert(originalType.Elem()))
					} else {
						result.SetMapIndex(key, value)
					}
				}
			}
			return result.Interface()
		}
	case reflect.Struct:
		// Handle direct struct
		if structResult, err := mapToStruct(validatedMap, originalType); err == nil {
			return structResult
		}
	}
	return validatedMap
}

// =============================================================================
// OBJECT CREATION HELPER
// =============================================================================

// createZodObjectFromDef creates a ZodObject from definition following the unified pattern
func createZodObjectFromDef(def *ZodObjectDef, params ...SchemaParams) *ZodObject {
	// Create internals with modern pattern
	internals := &ZodObjectInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Shape:            def.Shape,
		Bag:              make(map[string]interface{}),
	}

	// Apply schema parameters following unified pattern
	for _, param := range params {
		if param.Coerce {
			internals.Bag["coerce"] = true
		}
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				internals.Error = errorMap
			}
		}
		if param.Description != "" {
			internals.Bag["description"] = param.Description
		}
		if param.Abort {
			internals.Bag["abort"] = true
		}
		if len(param.Path) > 0 {
			internals.Bag["path"] = param.Path
		}
		if len(param.Params) > 0 {
			internals.Bag["params"] = param.Params
		}
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		objectDef := &ZodObjectDef{
			ZodTypeDef:  *newDef,
			Type:        "object",
			Shape:       def.Shape,
			Catchall:    def.Catchall,
			UnknownKeys: def.UnknownKeys,
		}
		newSchema := createZodObjectFromDef(objectDef)
		return any(newSchema).(ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodObject{internals: internals}
		result, err := schema.Parse(payload.Value, ctx)
		if err != nil {
			var zodErr *ZodError
			if errors.As(err, &zodErr) {
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Input:   issue.Input,
						Path:    issue.Path,
						Message: issue.Message,
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			}
			return payload
		}
		payload.Value = result
		return payload
	}

	schema := &ZodObject{internals: internals}

	// Use unified infrastructure for initialization
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

// =============================================================================
// CORE INTERFACE METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodObject) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Parse validates the input value using smart type inference
func (z *ZodObject) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// 1. nil handling
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "object", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*map[string]interface{})(nil), nil
	}

	// 2. smart type inference: check direct type matching
	objectValue, isNil, err := extractObjectValue(input)
	if err != nil {
		return nil, err
	}
	if isNil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "object", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*map[string]interface{})(nil), nil
	}

	// 3. convert to generic format and perform full validation (including field filtering)
	objectMap := convertObjectToMap(objectValue)
	if objectMap == nil {
		// 4. try type coercion (if enabled)
		if shouldCoerce(z.internals.Bag) {
			if coerced, ok := coerceToObject(input); ok {
				objectMap = coerced
			} else {
				rawIssue := CreateInvalidTypeIssue(input, "object", string(GetParsedType(input)))
				finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
		} else {
			rawIssue := CreateInvalidTypeIssue(input, "object", string(GetParsedType(input)))
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
	}

	// 5. use parseObjectCore for full object validation and field filtering
	payload := &ParsePayload{
		Value:  objectMap,
		Issues: make([]ZodRawIssue, 0),
		Path:   make([]interface{}, 0),
	}

	result := parseObjectCore(payload, z.internals, parseCtx)
	if len(result.Issues) > 0 {
		return nil, &ZodError{Issues: convertRawIssuesToIssues(result.Issues, parseCtx)}
	}

	// 6. run additional checks
	if len(z.internals.Checks) > 0 {
		checksPayload := &ParsePayload{
			Value:  result.Value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(result.Value, z.internals.Checks, checksPayload, parseCtx)
		if len(checksPayload.Issues) > 0 {
			return nil, &ZodError{Issues: convertRawIssuesToIssues(checksPayload.Issues, parseCtx)}
		}
	}

	// 7. for non-nil results, need to convert back to original type to maintain smart type inference
	if validatedMap, ok := result.Value.(map[string]interface{}); ok {
		originalType := reflect.TypeOf(input)
		return convertMapToObjectType(validatedMap, originalType), nil
	}

	return result.Value, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodObject) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Transform provides type-safe object transformation, supporting smart dereferencing
// Automatically handles input of map[string]interface{}, struct, *struct, and nil pointer
func (z *ZodObject) Transform(fn func(map[string]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		obj, isNil, err := extractObjectValue(input)

		if err != nil {
			return nil, err
		}

		if isNil {
			return nil, ErrTransformNilObject
		}

		// Convert to map[string]interface{} for consistent processing
		objMap := convertToMap(obj)
		if objMap == nil {
			return nil, ErrCannotConvertObject
		}

		return fn(objMap, ctx)
	})
}

// TransformAny flexible version of Transform - same implementation as Transform, providing backward compatibility
// Implements ZodType[any, any] interface: TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any]
func (z *ZodObject) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a new schema by piping the output of this schema to another
func (z *ZodObject) Pipe(target ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: target,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// =============================================================================
// OBJECT OPERATIONS
// =============================================================================

// Pick creates a new object schema with only the specified fields
func (z *ZodObject) Pick(keys []string) *ZodObject {
	newShape := make(ObjectSchema)
	for _, key := range keys {
		if schema, exists := z.internals.Shape[key]; exists {
			newShape[key] = schema
		}
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		Catchall:    z.internals.Def.Catchall,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Omit creates a new object schema without the specified fields
func (z *ZodObject) Omit(keys []string) *ZodObject {
	omitSet := make(map[string]struct{})
	for _, key := range keys {
		omitSet[key] = struct{}{}
	}

	newShape := make(ObjectSchema)
	for key, schema := range z.internals.Shape {
		if _, shouldOmit := omitSet[key]; !shouldOmit {
			newShape[key] = schema
		}
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		Catchall:    z.internals.Def.Catchall,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Extend creates a new object schema with additional fields
func (z *ZodObject) Extend(extension ObjectSchema) *ZodObject {
	newShape := make(ObjectSchema)

	// Copy existing fields
	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	// Add extension fields (overrides existing ones)
	for key, schema := range extension {
		newShape[key] = schema
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		Catchall:    z.internals.Def.Catchall,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Partial makes all fields optional
func (z *ZodObject) Partial() *ZodObject {
	newShape := make(ObjectSchema)
	for key, schema := range z.internals.Shape {
		newShape[key] = Optional(schema)
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		Catchall:    z.internals.Def.Catchall,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Required makes all fields required
func (z *ZodObject) Required(fields ...[]string) *ZodObject {
	newShape := make(ObjectSchema)

	// If specific fields provided, only make those required
	var targetFields map[string]struct{}
	if len(fields) > 0 && len(fields[0]) > 0 {
		targetFields = make(map[string]struct{})
		for _, field := range fields[0] {
			targetFields[field] = struct{}{}
		}
	}

	for key, schema := range z.internals.Shape {
		if targetFields != nil {
			// Only make specified fields required
			if _, shouldRequire := targetFields[key]; shouldRequire {
				// Remove optional wrapper if present
				if optionalType, ok := schema.(*ZodOptional[ZodType[any, any]]); ok {
					newShape[key] = optionalType.Unwrap()
				} else {
					newShape[key] = schema
				}
			} else {
				newShape[key] = schema
			}
		} else {
			// Make all fields required
			if optionalType, ok := schema.(*ZodOptional[ZodType[any, any]]); ok {
				newShape[key] = optionalType.Unwrap()
			} else {
				newShape[key] = schema
			}
		}
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		Catchall:    z.internals.Def.Catchall,
		UnknownKeys: z.internals.Def.UnknownKeys,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Merge combines this object schema with another
func (z *ZodObject) Merge(other *ZodObject) *ZodObject {
	newShape := make(ObjectSchema)

	// Copy fields from this schema
	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	// Override with fields from other schema
	for key, schema := range other.internals.Shape {
		newShape[key] = schema
	}

	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       newShape,
		Catchall:    other.internals.Def.Catchall,    // Use other's catchall
		UnknownKeys: other.internals.Def.UnknownKeys, // Use other's mode
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from both schemas (other takes precedence)
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}
	for key, value := range other.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Catchall sets a schema to validate unknown keys
func (z *ZodObject) Catchall(catchallSchema ZodType[any, any]) *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		Catchall:    catchallSchema,
		UnknownKeys: OBJECT_LOOSE_MODE, // Catchall implies loose mode
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Passthrough allows unknown keys to pass through (alias for loose mode)
func (z *ZodObject) Passthrough() *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		Catchall:    Unknown(), // Use Unknown for passthrough
		UnknownKeys: OBJECT_LOOSE_MODE,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Strict sets strict mode (rejects unknown keys)
func (z *ZodObject) Strict() *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		Catchall:    nil,
		UnknownKeys: OBJECT_STRICT_MODE,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Strip sets strip mode (removes unknown keys) - default behavior
func (z *ZodObject) Strip() *ZodObject {
	newDef := &ZodObjectDef{
		ZodTypeDef:  z.internals.Def.ZodTypeDef,
		Type:        z.internals.Def.Type,
		Shape:       z.internals.Shape,
		Catchall:    nil,
		UnknownKeys: OBJECT_STRIP_MODE,
	}

	newSchema := createZodObjectFromDef(newDef)

	// Copy runtime configuration from original schema
	for key, value := range z.internals.Bag {
		newSchema.internals.Bag[key] = value
	}

	return newSchema
}

// Keyof creates an enum schema from the object keys
func (z *ZodObject) Keyof() ZodType[any, any] {
	keys := make([]string, 0, len(z.internals.Shape))
	for key := range z.internals.Shape {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return Never() // Empty object has no keys
	}

	return Enum(keys...)
}

// =============================================================================
// WRAPPER METHODS
// =============================================================================

//////////////////////////////////////////
//////////   Utility Methods     //////////
//////////////////////////////////////////

// Refine adds a type-safe validation function to the object
func (z *ZodObject) Refine(fn func(map[string]interface{}) bool, params ...SchemaParams) *ZodObject {
	result := z.RefineAny(func(v any) bool {
		objectValue, isNil, err := extractObjectValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // let the upper logic decide
		}
		objectMap := convertObjectToMap(objectValue)
		if objectMap == nil {
			return false
		}
		return fn(objectMap)
	}, params...)
	return result.(*ZodObject)
}

// RefineAny adds flexible custom validation logic to the object schema
func (z *ZodObject) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[interface{}](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// Optional makes the object optional
func (z *ZodObject) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the object nilable
func (z *ZodObject) Nilable() ZodType[any, any] {
	return Clone(any(z).(ZodType[any, any]), func(def *ZodTypeDef) {
	}).(*ZodObject).setNilable()
}

func (z *ZodObject) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return any(z).(ZodType[any, any])
}

// Check adds modern validation using direct payload access
func (z *ZodObject) Check(fn CheckFn) *ZodObject {
	check := NewCustom[map[string]interface{}](fn, SchemaParams{})
	result := AddCheck(z, check)
	return result.(*ZodObject)
}

////////////////////////////
////   OBJECT DEFAULT WRAPPER ////
////////////////////////////

type ZodObjectDefault struct {
	*ZodDefault[*ZodObject] // embed pointer to allow method promotion
}

func (s ZodObjectDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

////////////////////////////
////   DEFAULT method   ////
////////////////////////////

// Default sets the default value for the object
func (z *ZodObject) Default(value map[string]interface{}) ZodObjectDefault {
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc sets the default function for the object
func (z *ZodObject) DefaultFunc(fn func() map[string]interface{}) ZodObjectDefault {
	genericFn := func() any { return fn() }
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

////////////////////////////
////   OBJECTDEFAULT CHAIN METHODS ////
////////////////////////////

// Pick creates a new object schema with only the specified fields, return ZodObjectDefault
func (s ZodObjectDefault) Pick(keys []string) ZodObjectDefault {
	newInner := s.innerType.Pick(keys)

	// filter the default value, only keep the fields of Pick
	var newDefaultValue any
	var newDefaultFunc func() any

	if s.isFunction && s.defaultFunc != nil {
		// if it is a function default value, create a new function to filter the result
		newDefaultFunc = func() any {
			originalValue := s.defaultFunc()
			if originalMap, ok := originalValue.(map[string]interface{}); ok {
				filteredMap := make(map[string]interface{})
				for _, key := range keys {
					if value, exists := originalMap[key]; exists {
						filteredMap[key] = value
					}
				}
				return filteredMap
			}
			return originalValue
		}
	} else {
		// if it is a static default value, filter directly
		if originalMap, ok := s.defaultValue.(map[string]interface{}); ok {
			filteredMap := make(map[string]interface{})
			for _, key := range keys {
				if value, exists := originalMap[key]; exists {
					filteredMap[key] = value
				}
			}
			newDefaultValue = filteredMap
		} else {
			newDefaultValue = s.defaultValue
		}
	}

	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: newDefaultValue,
			defaultFunc:  newDefaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Omit creates a new object schema with the specified fields omitted, return ZodObjectDefault
func (s ZodObjectDefault) Omit(keys []string) ZodObjectDefault {
	newInner := s.innerType.Omit(keys)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Extend extends the object schema, return ZodObjectDefault
func (s ZodObjectDefault) Extend(extension ObjectSchema) ZodObjectDefault {
	newInner := s.innerType.Extend(extension)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Partial makes all fields optional, return ZodObjectDefault
func (s ZodObjectDefault) Partial() ZodObjectDefault {
	newInner := s.innerType.Partial()
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Required makes the specified fields required, return ZodObjectDefault
func (s ZodObjectDefault) Required(fields ...[]string) ZodObjectDefault {
	newInner := s.innerType.Required(fields...)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Merge merges another object schema, return ZodObjectDefault
func (s ZodObjectDefault) Merge(other *ZodObject) ZodObjectDefault {
	newInner := s.innerType.Merge(other)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the object schema, return ZodObjectDefault
func (s ZodObjectDefault) Refine(fn func(map[string]interface{}) bool, params ...SchemaParams) ZodObjectDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodObjectDefault{
		&ZodDefault[*ZodObject]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, return a generic ZodType
func (s ZodObjectDefault) Transform(fn func(map[string]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	// use the TransformAny method of the embedded ZodDefault
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// smartly handle object value
		objVal, isNil, err := extractObjectValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilObject
		}

		// convert to map[string]interface{}
		objMap := convertToMap(objVal)
		if objMap == nil {
			return nil, ErrCannotConvertObject
		}

		return fn(objMap, ctx)
	})
}

// Optional modifier - correctly wrap Default wrapper
func (s ZodObjectDefault) Optional() ZodType[any, any] {
	// wrap the current ZodObjectDefault instance, keep Default logic
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable modifier - correctly wrap Default wrapper
func (s ZodObjectDefault) Nilable() ZodType[any, any] {
	// wrap the current ZodObjectDefault instance, keep Default logic
	return Nilable(any(s).(ZodType[any, any]))
}

////////////////////////////
////   OBJECT PREFAULT WRAPPER ////
////////////////////////////

// ZodObjectPrefault is the Prefault wrapper for object type
// provide perfect type safety and chain call support
type ZodObjectPrefault struct {
	*ZodPrefault[*ZodObject] // embed pointer to allow method promotion
}

////////////////////////////
////   PREFAULT method   ////
////////////////////////////

// Prefault adds a prefault value to the object schema, return ZodObjectPrefault
func (z *ZodObject) Prefault(value map[string]interface{}) ZodObjectPrefault {
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the object schema, return ZodObjectPrefault
func (z *ZodObject) PrefaultFunc(fn func() map[string]interface{}) ZodObjectPrefault {
	genericFn := func() any { return fn() }
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

////////////////////////////
////   OBJECTPREFAULT chain methods ////
////////////////////////////

// Pick selects the specified fields, return ZodObjectPrefault
func (o ZodObjectPrefault) Pick(keys []string) ZodObjectPrefault {
	newInner := o.innerType.Pick(keys)

	// filter prefaultValue, only keep the fields of Pick
	var newPrefaultValue any
	var newPrefaultFunc func() any

	if o.isFunction && o.prefaultFunc != nil {
		// if it is a function default value, create a new function to filter the result
		newPrefaultFunc = func() any {
			originalValue := o.prefaultFunc()
			if originalMap, ok := originalValue.(map[string]interface{}); ok {
				filteredMap := make(map[string]interface{})
				for _, key := range keys {
					if value, exists := originalMap[key]; exists {
						filteredMap[key] = value
					}
				}
				return filteredMap
			}
			return originalValue
		}
	} else {
		// if it is a static default value, filter directly
		if originalMap, ok := o.prefaultValue.(map[string]interface{}); ok {
			filteredMap := make(map[string]interface{})
			for _, key := range keys {
				if value, exists := originalMap[key]; exists {
					filteredMap[key] = value
				}
			}
			newPrefaultValue = filteredMap
		} else {
			newPrefaultValue = o.prefaultValue
		}
	}

	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: newPrefaultValue,
			prefaultFunc:  newPrefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Omit excludes the specified fields, return ZodObjectPrefault
func (o ZodObjectPrefault) Omit(keys []string) ZodObjectPrefault {
	newInner := o.innerType.Omit(keys)

	// create the omit set
	omitSet := make(map[string]struct{})
	for _, key := range keys {
		omitSet[key] = struct{}{}
	}

	// filter prefaultValue, exclude the fields of Omit
	var newPrefaultValue any
	var newPrefaultFunc func() any

	if o.isFunction && o.prefaultFunc != nil {
		// if it is a function default value, create a new function to filter the result
		newPrefaultFunc = func() any {
			originalValue := o.prefaultFunc()
			if originalMap, ok := originalValue.(map[string]interface{}); ok {
				filteredMap := make(map[string]interface{})
				for key, value := range originalMap {
					if _, shouldOmit := omitSet[key]; !shouldOmit {
						filteredMap[key] = value
					}
				}
				return filteredMap
			}
			return originalValue
		}
	} else {
		// if it is a static default value, filter directly
		if originalMap, ok := o.prefaultValue.(map[string]interface{}); ok {
			filteredMap := make(map[string]interface{})
			for key, value := range originalMap {
				if _, shouldOmit := omitSet[key]; !shouldOmit {
					filteredMap[key] = value
				}
			}
			newPrefaultValue = filteredMap
		} else {
			newPrefaultValue = o.prefaultValue
		}
	}

	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: newPrefaultValue,
			prefaultFunc:  newPrefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Partial makes all fields optional, return ZodObjectPrefault
func (o ZodObjectPrefault) Partial() ZodObjectPrefault {
	newInner := o.innerType.Partial()
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: o.prefaultValue,
			prefaultFunc:  o.prefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the object schema, return ZodObjectPrefault
func (o ZodObjectPrefault) Refine(fn func(map[string]interface{}) bool, params ...SchemaParams) ZodObjectPrefault {
	newInner := o.innerType.Refine(fn, params...)
	return ZodObjectPrefault{
		&ZodPrefault[*ZodObject]{
			innerType:     newInner,
			prefaultValue: o.prefaultValue,
			prefaultFunc:  o.prefaultFunc,
			isFunction:    o.isFunction,
		},
	}
}

// Transform adds data transformation, return a generic ZodType
func (o ZodObjectPrefault) Transform(fn func(map[string]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	// use the TransformAny method of the embedded ZodPrefault
	return o.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// smartly handle object value
		objVal, isNil, err := extractObjectValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilObject
		}

		// convert to map[string]interface{}
		objMap := convertToMap(objVal)
		if objMap == nil {
			return nil, ErrCannotConvertObject
		}

		return fn(objMap, ctx)
	})
}

// Optional modifier - correctly wrap Prefault wrapper
func (o ZodObjectPrefault) Optional() ZodType[any, any] {
	// wrap the current ZodObjectPrefault instance, keep Prefault logic
	return Optional(any(o).(ZodType[any, any]))
}

// Nilable modifier - correctly wrap Prefault wrapper
func (o ZodObjectPrefault) Nilable() ZodType[any, any] {
	// wrap the current ZodObjectPrefault instance, keep Prefault logic
	return Nilable(any(o).(ZodType[any, any]))
}

// =============================================================================
// INTERNAL CHECK MANAGEMENT
// =============================================================================

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodObject) CloneFrom(source any) {
	if src, ok := source.(*ZodObject); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy shape and other object-specific fields
		if src.internals.Shape != nil {
			z.internals.Shape = make(ObjectSchema)
			for k, v := range src.internals.Shape {
				z.internals.Shape[k] = v
			}
		}

		// Copy object definition fields
		if src.internals.Def != nil {
			z.internals.Def.Catchall = src.internals.Def.Catchall
			z.internals.Def.UnknownKeys = src.internals.Def.UnknownKeys
		}
	}
}

// =============================================================================
// CORE PARSING LOGIC
// =============================================================================

// parseObjectCore handles object-specific parsing logic
func parseObjectCore(payload *ParsePayload, internals *ZodObjectInternals, ctx *ParseContext) *ParsePayload {
	// 1. type check - only accept map[string]interface{}
	objectData, ok := payload.Value.(map[string]interface{})
	if !ok {
		// check if coercion is enabled
		if shouldCoerce(internals.Bag) {
			if coerced := convertToMap(payload.Value); coerced != nil {
				objectData = coerced
			} else {
				receivedType := string(GetParsedType(payload.Value))
				issue := CreateInvalidTypeIssue(payload.Value, "object", receivedType, func(issue *ZodRawIssue) {
					issue.Inst = internals
				})
				payload.Issues = append(payload.Issues, issue)
				return payload
			}
		} else {
			receivedType := string(GetParsedType(payload.Value))
			issue := CreateInvalidTypeIssue(payload.Value, "object", receivedType, func(issue *ZodRawIssue) {
				issue.Inst = internals
			})
			payload.Issues = append(payload.Issues, issue)
			return payload
		}
	}

	// 2. field validation - validate each field according to Shape definition
	result := make(map[string]interface{})
	processedKeys := make(map[string]struct{})

	// validate defined fields
	for fieldName, fieldSchema := range internals.Shape {
		fieldPath := make([]interface{}, 0, len(payload.Path)+1)
		fieldPath = append(fieldPath, payload.Path...)
		fieldPath = append(fieldPath, fieldName)
		fieldValue, exists := objectData[fieldName]

		if !exists {
			// check if field is optional
			if isOptionalField(fieldSchema) {
				// for optional fields, try parsing nil to get default values
				// this handles Default types that should provide default values

				// use the schema's Parse method directly, not GetInternals().Parse
				// this ensures wrapper types like Default work correctly
				fieldResultValue, fieldErr := fieldSchema.Parse(nil, ctx)
				if fieldErr == nil && fieldResultValue != nil {
					// default value provided, include it in result
					result[fieldName] = fieldResultValue
				}
				// if no default value or parsing failed, skip the field (Optional behavior)
				continue
			} else {
				// missing required field
				issue := createMissingKeyIssue(fieldName, func(issue *ZodRawIssue) {
					issue.Path = fieldPath
					issue.Inst = internals
				})
				payload.Issues = append(payload.Issues, issue)
				continue
			}
		}

		// if object-level coercion is enabled, enable it for field schemas too
		actualFieldSchema := fieldSchema
		if shouldCoerce(internals.Bag) {
			actualFieldSchema = enableCoercionForFieldType(fieldSchema)
		}

		// fix nil pointer error: for wrapper types, use schema.Parse method directly
		if actualFieldSchema == nil {
			// if schema is nil, create an error
			issue := CreateInvalidTypeIssue(fieldValue, "unknown", string(GetParsedType(fieldValue)), func(issue *ZodRawIssue) {
				issue.Path = fieldPath
				issue.Inst = internals
			})
			payload.Issues = append(payload.Issues, issue)
			continue
		}

		// use the schema's Parse method directly, this ensures wrapper types (like Prefault, Default) work correctly
		fieldResultValue, fieldErr := actualFieldSchema.Parse(fieldValue, ctx)
		if fieldErr != nil {
			// when parsing a field, convert the error to payload format
			var zodErr *ZodError
			if errors.As(fieldErr, &zodErr) {
				// convert ZodIssue to ZodRawIssue and update the error path
				for _, issue := range zodErr.Issues {
					rawIssue := ZodRawIssue{
						Code:    issue.Code,
						Message: issue.Message,
						Path:    append(fieldPath, issue.Path...),
						Input:   issue.Input,
						Inst:    internals,
					}
					// copy other fields to Properties
					if rawIssue.Properties == nil {
						rawIssue.Properties = make(map[string]interface{})
					}
					if issue.Expected != "" {
						rawIssue.Properties["expected"] = issue.Expected
					}
					if issue.Received != "" {
						rawIssue.Properties["received"] = issue.Received
					}
					if issue.Format != "" {
						rawIssue.Properties["format"] = issue.Format
					}
					payload.Issues = append(payload.Issues, rawIssue)
				}
			} else {
				// if not ZodError, create a generic error
				issue := CreateInvalidTypeIssue(fieldValue, "unknown", string(GetParsedType(fieldValue)), func(issue *ZodRawIssue) {
					issue.Path = fieldPath
					issue.Inst = internals
				})
				payload.Issues = append(payload.Issues, issue)
			}
		} else {
			result[fieldName] = fieldResultValue
		}

		processedKeys[fieldName] = struct{}{}
	}

	// 3. handle unknown keys according to UnknownKeys mode
	for key, value := range objectData {
		if _, processed := processedKeys[key]; !processed {
			// Unknown key found
			switch internals.Def.UnknownKeys {
			case OBJECT_STRICT_MODE:
				// strict mode: error on unknown keys
				issue := createUnrecognizedKeysIssue([]string{key}, func(issue *ZodRawIssue) {
					issuePath := make([]interface{}, 0, len(payload.Path)+1)
					issuePath = append(issuePath, payload.Path...)
					issuePath = append(issuePath, key)
					issue.Path = issuePath
					issue.Inst = internals
				})
				payload.Issues = append(payload.Issues, issue)

			case OBJECT_LOOSE_MODE:
				// loose mode: allow unknown keys to pass through
				if internals.Def.Catchall != nil {
					// validate with catchall schema
					catchallPayload := &ParsePayload{
						Value:  value,
						Path:   append(payload.Path, key),
						Issues: make([]ZodRawIssue, 0),
					}
					catchallResult := internals.Def.Catchall.GetInternals().Parse(catchallPayload, ctx)
					if len(catchallResult.Issues) > 0 {
						payload.Issues = append(payload.Issues, catchallResult.Issues...)
					} else {
						result[key] = catchallResult.Value
					}
				} else {
					// no catchall, pass through as-is
					result[key] = value
				}

			case OBJECT_STRIP_MODE:
				// strip mode: ignore unknown keys (default behavior)
				// do nothing - key is stripped
			}
		}
	}

	// 4. run custom checks if object validation succeeded
	if len(payload.Issues) == 0 {
		if customChecks, exists := internals.Bag["customChecks"].([]ZodCheck); exists {
			runChecksOnValue(result, customChecks, payload, ctx)
		}
	}

	// update payload with validated object
	payload.Value = result
	return payload
}

// enableCoercionForFieldType enable coercion for field type
func enableCoercionForFieldType(schema ZodType[any, any]) ZodType[any, any] {
	if schema == nil {
		return schema
	}

	internals := schema.GetInternals()
	if internals == nil {
		return schema
	}

	// if coercion is already enabled, return directly
	if shouldCoerce(internals.Bag) {
		return schema
	}

	// set coercion flag
	if internals.Bag == nil {
		internals.Bag = make(map[string]interface{})
	}
	internals.Bag["coerce"] = true

	return schema
}

// Unwrap returns the inner type (for basic types, return self)
func (z *ZodObject) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}
