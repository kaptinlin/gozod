package gozod

import (
	"errors"
	"fmt"
	"reflect"
)

//////////////////////////////////////////
//////////////////////////////////////////

// =============================================================================
// CORE TYPE DEFINITIONS
// =============================================================================

// ZodStructDef defines the configuration for struct/object validation
type ZodStructDef struct {
	ZodTypeDef
	Type     string            // "struct"
	Shape    ObjectSchema      // Field schemas
	Catchall ZodType[any, any] // Schema for unrecognized keys
	Mode     string            // "strict", "strip", "loose"
}

// ZodStructInternals contains struct validator internal state
type ZodStructInternals struct {
	ZodTypeInternals
	Def      *ZodStructDef          // Schema definition
	Shape    ObjectSchema           // Field schemas
	Mode     string                 // Validation mode
	Catchall ZodType[any, any]      // Catchall schema
	Bag      map[string]interface{} // Runtime configuration and custom checks
}

// ZodStruct represents a struct/object validation schema with type safety
type ZodStruct struct {
	internals *ZodStructInternals
}

// Mode constants for struct validation
const (
	STRICT_MODE = "strict" // Error on unknown keys
	STRIP_MODE  = "strip"  // Strip unknown keys (default)
	LOOSE_MODE  = "loose"  // Allow unknown keys
)

// =============================================================================
// VALIDATION METHODS
// =============================================================================

// GetInternals returns the internal state of the schema
func (z *ZodStruct) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the struct-specific internals for framework usage
func (z *ZodStruct) GetZod() *ZodStructInternals {
	return z.internals
}

// Shape provides access to the internal field schemas
func (z *ZodStruct) Shape() ObjectSchema {
	return z.internals.Shape
}

// Coerce implements Coercible interface
func (z *ZodStruct) Coerce(input interface{}) (interface{}, bool) {
	if mapped := convertToMap(input); mapped != nil {
		return mapped, true
	}
	return input, false
}

// Parse validates and parses with smart type inference
func (z *ZodStruct) Parse(input any, ctx ...*ParseContext) (any, error) {
	// handle nil input
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "struct", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*map[string]interface{})(nil), nil // return struct type nil pointer
	}

	// smart type inference: input type determines output type
	inputType := reflect.TypeOf(input)
	var objectValue interface{}
	var isNil bool
	var err error

	// Extract and validate the struct/object using existing utility
	objectValue, isNil, err = extractStructValue(input)
	if err != nil {
		return nil, err
	}
	if isNil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "struct", "null")
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*map[string]interface{})(nil), nil
	}

	// Convert to map for unified validation
	objectMap := objectValue.(map[string]interface{})

	// Validate the object using unified parsing infrastructure
	payload := &ParsePayload{
		Value:  objectMap,
		Issues: make([]ZodRawIssue, 0),
		Path:   make([]interface{}, 0),
	}

	var parseCtx *ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	result := z.internals.Parse(payload, parseCtx)
	if len(result.Issues) > 0 {
		return nil, &ZodError{Issues: convertRawIssuesToIssues(result.Issues, parseCtx)}
	}

	// Run checks on the validated object
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

	// smart type inference: keep input type characteristics
	validatedMap := result.Value.(map[string]interface{})

	// if input was a struct, convert back to struct
	if IsStructType(input) {
		if inputType.Kind() == reflect.Ptr {
			// Handle pointer to struct
			elemType := inputType.Elem()
			if structResult, err := mapToStruct(validatedMap, elemType); err == nil {
				// Create pointer to struct
				ptrValue := reflect.New(elemType)
				ptrValue.Elem().Set(reflect.ValueOf(structResult))
				return ptrValue.Interface(), nil
			}
		} else {
			// Handle direct struct
			if structResult, err := mapToStruct(validatedMap, inputType); err == nil {
				return structResult, nil
			}
		}
	}

	// If input was a pointer to map, maintain pointer characteristics
	if inputType.Kind() == reflect.Ptr {
		if elemType := inputType.Elem(); elemType.Kind() == reflect.Map && elemType.Key().Kind() == reflect.String {
			// Convert to specific map type and return as pointer
			result := reflect.MakeMap(elemType)
			for k, v := range validatedMap {
				key := reflect.ValueOf(k)
				value := reflect.ValueOf(v)
				if value.Type().ConvertibleTo(elemType.Elem()) {
					result.SetMapIndex(key, value.Convert(elemType.Elem()))
				} else {
					result.SetMapIndex(key, value)
				}
			}
			ptr := reflect.New(elemType)
			ptr.Elem().Set(result)
			return ptr.Interface(), nil
		}
		// Default for pointer: return pointer to map[string]interface{}
		return &validatedMap, nil
	}

	// If input was a map, try to convert to original map type
	if inputType.Kind() == reflect.Map && inputType.Key().Kind() == reflect.String {
		result := reflect.MakeMap(inputType)
		for k, v := range validatedMap {
			key := reflect.ValueOf(k)
			value := reflect.ValueOf(v)
			if value.Type().ConvertibleTo(inputType.Elem()) {
				result.SetMapIndex(key, value.Convert(inputType.Elem()))
			} else {
				result.SetMapIndex(key, value)
			}
		}
		return result.Interface(), nil
	}

	// Default: return as map[string]interface{}
	return validatedMap, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodStruct) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodStruct) CloneFrom(source any) {
	if src, ok := source.(*ZodStruct); ok {
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}
		z.internals.Shape = src.internals.Shape
		z.internals.Mode = src.internals.Mode
		z.internals.Catchall = src.internals.Catchall
	}
}

// =============================================================================
// TRANSFORM METHODS
// =============================================================================

// Refine adds type-safe custom validation logic to the struct schema
func (z *ZodStruct) Refine(fn func(map[string]interface{}) bool, params ...SchemaParams) *ZodStruct {
	result := z.RefineAny(func(v any) bool {
		structVal, isNil, err := extractStructValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true
		}
		return fn(structVal)
	}, params...)
	return result.(*ZodStruct)
}

// RefineAny adds flexible custom validation logic to the struct schema
func (z *ZodStruct) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

// TransformAny creates a transform that modifies the value
func (z *ZodStruct) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipeline with another schema
func (z *ZodStruct) Pipe(schema ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: schema,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// =============================================================================
// MODIFIER METHODS
// =============================================================================

// Optional makes the struct optional
func (z *ZodStruct) Optional() ZodType[any, any] {
	return Optional(any(z).(ZodType[any, any]))
}

// Nilable modifier: only changes nil handling, not type inference logic
func (z *ZodStruct) Nilable() ZodType[any, any] {
	return Clone(z, func(def *ZodTypeDef) {}).(*ZodStruct).setNilable()
}

// setNilable sets the Nilable flag - internal method
func (z *ZodStruct) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Nullish makes the struct both optional and nilable
func (z *ZodStruct) Nullish() ZodType[any, any] {
	return Nullish(any(z).(ZodType[any, any]))
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodStruct) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// =============================================================================
// WRAPPER TYPES & CHAIN METHODS
// =============================================================================

// ZodStructDefault is the Default wrapper for Struct type
type ZodStructDefault struct {
	*ZodDefault[*ZodStruct] // Embed concrete pointer for method promotion
}

// ZodStructPrefault is the Prefault wrapper for Struct type
type ZodStructPrefault struct {
	*ZodPrefault[*ZodStruct] // Embed concrete pointer for method promotion
}

// Default adds a default value to the struct schema, returns ZodStructDefault
// For structs, we use map[string]interface{} as the canonical form
func (z *ZodStruct) Default(value map[string]interface{}) ZodStructDefault {
	return ZodStructDefault{
		&ZodDefault[*ZodStruct]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the struct schema, returns ZodStructDefault
func (z *ZodStruct) DefaultFunc(fn func() map[string]interface{}) ZodStructDefault {
	genericFn := func() any { return fn() }
	return ZodStructDefault{
		&ZodDefault[*ZodStruct]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault adds a prefault value to the struct schema, returns ZodStructPrefault
func (z *ZodStruct) Prefault(value map[string]interface{}) ZodStructPrefault {
	return ZodStructPrefault{
		&ZodPrefault[*ZodStruct]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the struct schema, returns ZodStructPrefault
func (z *ZodStruct) PrefaultFunc(fn func() map[string]interface{}) ZodStructPrefault {
	genericFn := func() any { return fn() }
	return ZodStructPrefault{
		&ZodPrefault[*ZodStruct]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

//////////////////////////////////
////   STRUCTPREFAULT CHAIN METHODS ////
//////////////////////////////////

// Pick pick specified keys, return ZodStructPrefault
func (s ZodStructPrefault) Pick(keys []string) ZodStructPrefault {
	newInner := s.innerType.Pick(keys)
	return ZodStructPrefault{
		&ZodPrefault[*ZodStruct]{
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Omit omit specified keys, return ZodStructPrefault
func (s ZodStructPrefault) Omit(keys []string) ZodStructPrefault {
	newInner := s.innerType.Omit(keys)
	return ZodStructPrefault{
		&ZodPrefault[*ZodStruct]{
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Partial make all fields optional, return ZodStructPrefault
func (s ZodStructPrefault) Partial(keys ...[]string) ZodStructPrefault {
	newInner := s.innerType.Partial(keys...)
	return ZodStructPrefault{
		&ZodPrefault[*ZodStruct]{
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Refine refine the struct with given function, return ZodStructPrefault
func (s ZodStructPrefault) Refine(fn func(map[string]interface{}) bool, params ...SchemaParams) ZodStructPrefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodStructPrefault{
		&ZodPrefault[*ZodStruct]{
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Transform transform the struct with given function, return ZodType
func (s ZodStructPrefault) Transform(fn func(interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		structVal, isNil, err := extractStructValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilStruct
		}
		return fn(structVal, ctx)
	})
}

// Optional make the struct optional
func (s ZodStructPrefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable make the struct nilable
func (s ZodStructPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// Pick creates a new struct with only the specified keys
func (z *ZodStruct) Pick(keys []string) *ZodStruct {
	newShape := make(map[string]ZodType[any, any])
	for _, key := range keys {
		if schema, exists := z.internals.Shape[key]; exists {
			newShape[key] = schema
		} else {
			panic(fmt.Sprintf("Cannot pick key '%s': it does not exist on the schema", key))
		}
	}

	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      newShape,
		Catchall:   z.internals.Def.Catchall,
		Mode:       z.internals.Def.Mode,
	}
	return createZodStructFromDef(newDef)
}

// Omit creates a new struct schema without the specified properties
func (z *ZodStruct) Omit(keys []string) *ZodStruct {
	newShape := make(map[string]ZodType[any, any])
	keySet := make(map[string]struct{})

	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	for key, schema := range z.internals.Shape {
		if _, shouldOmit := keySet[key]; !shouldOmit {
			newShape[key] = schema
		}
	}

	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      newShape,
		Catchall:   z.internals.Def.Catchall,
		Mode:       z.internals.Def.Mode,
	}
	return createZodStructFromDef(newDef)
}

// Extend extends this struct schema with additional properties
func (z *ZodStruct) Extend(augmentation ObjectSchema) *ZodStruct {
	newShape := make(map[string]ZodType[any, any])

	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	for key, schema := range augmentation {
		newShape[key] = schema
	}

	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      newShape,
		Catchall:   z.internals.Def.Catchall,
		Mode:       z.internals.Def.Mode,
	}
	return createZodStructFromDef(newDef)
}

// Partial makes all or specified fields optional
func (z *ZodStruct) Partial(keys ...[]string) *ZodStruct {
	newShape := make(map[string]ZodType[any, any])

	if len(keys) > 0 && keys[0] != nil {
		keySet := make(map[string]struct{})
		for _, key := range keys[0] {
			keySet[key] = struct{}{}
		}

		for key, schema := range z.internals.Shape {
			if _, shouldMakeOptional := keySet[key]; shouldMakeOptional {
				newShape[key] = Optional(schema)
			} else {
				newShape[key] = schema
			}
		}
	} else {
		for key, schema := range z.internals.Shape {
			newShape[key] = Optional(schema)
		}
	}

	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      newShape,
		Catchall:   z.internals.Def.Catchall,
		Mode:       z.internals.Def.Mode,
	}
	return createZodStructFromDef(newDef)
}

// Merge merges this struct schema with another struct schema
func (z *ZodStruct) Merge(other *ZodStruct) *ZodStruct {
	newShape := make(map[string]ZodType[any, any])

	for key, schema := range z.internals.Shape {
		newShape[key] = schema
	}

	for key, schema := range other.internals.Shape {
		newShape[key] = schema
	}

	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      newShape,
		Catchall:   other.internals.Def.Catchall,
		Mode:       other.internals.Def.Mode,
	}
	return createZodStructFromDef(newDef)
}

// Catchall sets a schema to validate unknown keys
func (z *ZodStruct) Catchall(catchallSchema ZodType[any, any]) *ZodStruct {
	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      z.internals.Shape,
		Catchall:   catchallSchema,
		Mode:       LOOSE_MODE,
	}
	return createZodStructFromDef(newDef)
}

// Passthrough allows unknown keys to pass through
func (z *ZodStruct) Passthrough() *ZodStruct {
	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      z.internals.Shape,
		Catchall:   Unknown(),
		Mode:       LOOSE_MODE,
	}
	return createZodStructFromDef(newDef)
}

// Strict sets strict mode (rejects unknown keys)
func (z *ZodStruct) Strict() *ZodStruct {
	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      z.internals.Shape,
		Catchall:   nil,
		Mode:       STRICT_MODE,
	}
	return createZodStructFromDef(newDef)
}

// Strip sets strip mode (removes unknown keys)
func (z *ZodStruct) Strip() *ZodStruct {
	newDef := &ZodStructDef{
		ZodTypeDef: z.internals.Def.ZodTypeDef,
		Type:       z.internals.Def.Type,
		Shape:      z.internals.Shape,
		Catchall:   nil,
		Mode:       STRIP_MODE,
	}
	return createZodStructFromDef(newDef)
}

// Keyof creates an enum schema from the struct keys
func (z *ZodStruct) Keyof() ZodType[any, any] {
	keys := make([]interface{}, 0, len(z.internals.Shape))
	for key := range z.internals.Shape {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return Never()
	}

	return EnumSlice(keys)
}

// =============================================================================
// SECTION 6: CONSTRUCTOR FUNCTIONS
// =============================================================================

// createZodStructFromDef creates a ZodStruct from definition
func createZodStructFromDef(def *ZodStructDef) *ZodStruct {
	internals := &ZodStructInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		Shape:            def.Shape,
		Mode:             def.Mode,
		Catchall:         def.Catchall,
		Bag:              make(map[string]interface{}),
	}

	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		structDef := &ZodStructDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeStruct,
			Shape:      def.Shape,
			Catchall:   def.Catchall,
			Mode:       def.Mode,
		}
		return any(createZodStructFromDef(structDef)).(ZodType[any, any])
	}

	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		parseCtx := (*ParseContext)(nil)
		if ctx != nil {
			parseCtx = ctx
		}

		// Use parseType template, including smart type inference and struct validation
		typeChecker := func(v any) (map[string]interface{}, bool) {
			// Check if v is map[string]interface{}
			if m, ok := v.(map[string]interface{}); ok {
				return m, true
			}
			return nil, false
		}

		coercer := func(v any) (map[string]interface{}, bool) {
			// Try to coerce to map[string]interface{}
			if mapped := convertToMap(v); mapped != nil {
				return mapped, true
			}
			return nil, false
		}

		validator := func(value map[string]interface{}, checks []ZodCheck, ctx *ParseContext) error {
			// Run base checks
			if len(checks) > 0 {
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			// Validate struct fields
			result := make(map[string]interface{})
			processedKeys := make(map[string]struct{})

			// Validate defined fields
			for schemaKey, fieldSchema := range internals.Shape {
				processedKeys[schemaKey] = struct{}{}

				// Get field value
				fieldValue, exists := value[schemaKey]
				if !exists {
					// Field does not exist, check if it is an optional field
					if isOptionalField(fieldSchema) {
						// For optional fields, try to parse nil to get default value
						fieldResultValue, fieldErr := fieldSchema.Parse(nil, ctx)
						if fieldErr == nil && fieldResultValue != nil {
							// Provided default value, include in result
							result[schemaKey] = fieldResultValue
						}
						// If no default value or parsing fails, skip the field (Optional behavior)
						continue
					} else {
						// Missing required field
						issue := createMissingKeyIssue(schemaKey, func(issue *ZodRawIssue) {
							issue.Path = payload.Path
							issue.Inst = internals
						})
						payload.Issues = append(payload.Issues, issue)
						return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
					}
				}

				// Validate field value
				fieldPayload := &ParsePayload{
					Value:  fieldValue,
					Path:   append(payload.Path, schemaKey),
					Issues: make([]ZodRawIssue, 0),
				}

				fieldResult := fieldSchema.GetInternals().Parse(fieldPayload, ctx)

				// Process field validation results
				if len(fieldResult.Issues) > 0 {
					fieldIssues := processFieldPath(fieldResult.Issues, schemaKey)
					payload.Issues = append(payload.Issues, fieldIssues...)
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				} else {
					result[schemaKey] = fieldResult.Value
				}
			}

			// Process unrecognized fields
			var unrecognizedFields []string
			for key := range value {
				if _, processed := processedKeys[key]; !processed {
					unrecognizedFields = append(unrecognizedFields, key)
				}
			}

			if len(unrecognizedFields) > 0 {
				switch internals.Mode {
				case STRICT_MODE:
					issue := createUnrecognizedKeysIssue(unrecognizedFields, func(issue *ZodRawIssue) {
						issue.Path = payload.Path
						issue.Inst = internals
					})
					payload.Issues = append(payload.Issues, issue)
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}

				case LOOSE_MODE:
					// Process unrecognized fields
					for _, key := range unrecognizedFields {
						fieldValue := value[key]

						// If there is a Catchall validator, use it to validate unrecognized fields
						if internals.Catchall != nil {
							fieldPayload := &ParsePayload{
								Value:  fieldValue,
								Path:   append(payload.Path, key),
								Issues: make([]ZodRawIssue, 0),
							}

							fieldResult := internals.Catchall.GetInternals().Parse(fieldPayload, ctx)

							if len(fieldResult.Issues) > 0 {
								fieldIssues := processFieldPath(fieldResult.Issues, key)
								payload.Issues = append(payload.Issues, fieldIssues...)
								return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
							} else {
								result[key] = fieldResult.Value
							}
						} else {
							// No Catchall validator, just copy
							result[key] = fieldValue
						}
					}

				case STRIP_MODE:
					// Do not process unrecognized fields (strip)
				}
			}

			// Update payload value
			payload.Value = result
			return nil
		}

		// Directly call validator, not using parseType template
		// Because Struct's validator needs to modify payload.Value
		if value, ok := typeChecker(payload.Value); ok {
			err := validator(value, internals.Checks, parseCtx)
			if err != nil {
				// Convert error to ParsePayload format
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
			}
			// validator has modified payload.Value
			return payload
		}

		// Try coercion
		if shouldCoerce(internals.Bag) {
			if coerced, ok := coercer(payload.Value); ok {
				err := validator(coerced, internals.Checks, parseCtx)
				if err != nil {
					// Convert error to ParsePayload format
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
				}
				// validator has modified payload.Value
				return payload
			}
		}

		// Type mismatch
		receivedType := string(GetParsedType(payload.Value))
		issue := CreateInvalidTypeIssue(payload.Value, "struct", receivedType, func(issue *ZodRawIssue) {
			issue.Inst = internals
		})
		payload.Issues = append(payload.Issues, issue)
		return payload
	}

	schema := &ZodStruct{internals: internals}
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)
	return schema
}

// NewZodStruct creates a new struct schema
func NewZodStruct(shape ObjectSchema, params ...SchemaParams) *ZodStruct {
	def := &ZodStructDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeStruct,
			Checks: make([]ZodCheck, 0),
		},
		Type:     ZodTypeStruct,
		Shape:    shape,
		Catchall: nil,
		Mode:     STRIP_MODE,
	}

	schema := createZodStructFromDef(def)

	if len(params) > 0 {
		param := params[0]
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
		}
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}
		if param.Description != "" {
			schema.internals.Bag["description"] = param.Description
		}
		if param.Abort {
			schema.internals.Bag["abort"] = true
		}
		if len(param.Path) > 0 {
			schema.internals.Bag["path"] = param.Path
		}
		if len(param.Params) > 0 {
			schema.internals.Bag["params"] = param.Params
		}
	}

	return schema
}

// Struct creates a new struct schema (public constructor)
func Struct(shape ObjectSchema, params ...SchemaParams) *ZodStruct {
	return NewZodStruct(shape, params...)
}

// StrictStruct creates a struct schema in strict mode
func StrictStruct(shape ObjectSchema, params ...SchemaParams) *ZodStruct {
	schema := NewZodStruct(shape, params...)
	schema.internals.Mode = STRICT_MODE
	schema.internals.Def.Mode = STRICT_MODE
	return schema
}

// LooseStruct creates a struct schema in loose mode
func LooseStruct(shape ObjectSchema, params ...SchemaParams) *ZodStruct {
	schema := NewZodStruct(shape, params...)
	schema.internals.Mode = LOOSE_MODE
	schema.internals.Def.Mode = LOOSE_MODE
	return schema
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// extractStructValue intelligently extracts struct values
func extractStructValue(input any) (map[string]interface{}, bool, error) {
	switch v := input.(type) {
	case map[string]interface{}:
		return v, false, nil
	case *map[string]interface{}:
		if v == nil {
			return nil, true, nil
		}
		return *v, false, nil
	default:
		structValue := reflect.ValueOf(input)
		if structValue.Kind() == reflect.Ptr {
			if structValue.IsNil() {
				return nil, true, nil
			}
			structValue = structValue.Elem()
		}

		if structValue.Kind() == reflect.Struct {
			structMap := make(map[string]interface{})
			structType := structValue.Type()

			for i := 0; i < structValue.NumField(); i++ {
				field := structType.Field(i)
				if field.IsExported() {
					fieldName := field.Name
					if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
						fieldName = tag
					}
					value := structValue.Field(i)
					if value.CanInterface() {
						structMap[fieldName] = value.Interface()
					}
				}
			}
			return structMap, false, nil
		}

		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedStruct, input)
	}
}
