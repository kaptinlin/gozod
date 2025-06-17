package gozod

import (
	"errors"
	"fmt"
)

//////////////////////////
////   RECORD TYPES  ////
//////////////////////////

// ZodRecordDef defines the configuration for record validation
type ZodRecordDef struct {
	ZodTypeDef
	Type      string            // "record"
	KeyType   ZodType[any, any] // Schema for validating keys
	ValueType ZodType[any, any] // Schema for validating values
}

// ZodRecordInternals contains record validator internal state
type ZodRecordInternals struct {
	ZodTypeInternals
	Def       *ZodRecordDef          // Schema definition
	KeyType   ZodType[any, any]      // Key validation schema
	ValueType ZodType[any, any]      // Value validation schema
	Isst      ZodIssueInvalidType    // Invalid type issue template
	Bag       map[string]interface{} // Additional metadata
}

// ZodRecord represents a record validation schema for key-value pairs
type ZodRecord struct {
	internals *ZodRecordInternals
}

// GetInternals returns the internal state of the schema
func (z *ZodRecord) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// Coerce implements Coercible interface for record type conversion
func (z *ZodRecord) Coerce(input interface{}) (interface{}, bool) {
	// Use generic convertToMap function for coercion
	if mapped := convertToMap(input); mapped != nil {
		// Convert to map[interface{}]interface{}
		result := make(map[interface{}]interface{})
		for k, v := range mapped {
			result[k] = v
		}
		return result, true
	}
	return input, false
}

// Parse validates input with smart type inference
func (z *ZodRecord) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	// 1. Unified nil handling
	if input == nil {
		if !z.internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, "object", "null")
			finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		return (*map[interface{}]interface{})(nil), nil
	}

	// 2. Smart type inference: check pointer type matching
	if ptr, ok := input.(*map[interface{}]interface{}); ok {
		if ptr == nil {
			if !z.internals.Nilable {
				rawIssue := CreateInvalidTypeIssue(input, "object", "null")
				finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
			return (*map[interface{}]interface{})(nil), nil
		}
		// Validate directly and return modified value
		result, err := z.validateRecordAndRunChecks(*ptr)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	// 3. Smart type inference: check direct type matching and conversion
	var recordMap map[interface{}]interface{}
	var ok bool

	switch val := input.(type) {
	case map[interface{}]interface{}:
		recordMap = val
		ok = true
	case map[string]interface{}:
		// Convert map[string]interface{} to map[interface{}]interface{}
		recordMap = make(map[interface{}]interface{})
		for k, value := range val {
			recordMap[k] = value
		}
		ok = true
	default:
		// Try coercion
		if shouldCoerce(z.internals.Bag) {
			if mapped := convertToMap(input); mapped != nil {
				recordMap = make(map[interface{}]interface{})
				for k, value := range mapped {
					recordMap[k] = value
				}
				ok = true
			}
		}
	}

	if !ok {
		rawIssue := CreateInvalidTypeIssue(input, "object", string(GetParsedType(input)))
		finalIssue := FinalizeIssue(rawIssue, parseCtx, GetConfig())
		return nil, NewZodError([]ZodIssue{finalIssue})
	}

	// 4. Validate and return modified value
	result, err := z.validateRecordAndRunChecks(recordMap)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MustParse validates the input value and panics on failure
func (z *ZodRecord) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// Transform provides type-safe record transformation with smart dereferencing support
func (z *ZodRecord) Transform(fn func(map[interface{}]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		recordMap, isNil, err := extractRecordPointerValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilRecord
		}
		return fn(recordMap, ctx)
	})
}

// TransformAny flexible version of transformation
func (z *ZodRecord) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),         // Type conversion
		out: any(transform).(ZodType[any, any]), // Type conversion
		def: ZodTypeDef{Type: "pipe"},
	}
}

///////////////////////////
////   RECORD WRAPPERS ////
///////////////////////////

// Optional makes the record optional
func (z *ZodRecord) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable makes the record nilable
func (z *ZodRecord) Nilable() ZodType[any, any] {
	return z.setNilable()
}

func (z *ZodRecord) setNilable() ZodType[any, any] {
	cloned := Clone(any(z).(ZodType[any, any]), func(def *ZodTypeDef) {
		// setNilable only changes nil handling, not other logic
	})
	cloned.(*ZodRecord).internals.Nilable = true
	return cloned
}

// Nullish makes the record both optional and nilable
func (z *ZodRecord) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Pipe creates a validation pipeline
func (z *ZodRecord) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: out,
		def: ZodTypeDef{Type: "pipe"},
	}
}

////////////////////////////
////   TYPE-SAFE PREFAULT   ////
////////////////////////////

// ZodRecordPrefault is a Prefault wrapper for record type
// Provides perfect type safety and chainable method support
type ZodRecordPrefault struct {
	*ZodPrefault[*ZodRecord] // Embed generic wrapper
}

////////////////////////////
////   PREFAULT METHODS   ////
////////////////////////////

// Prefault adds a prefault value to the record schema
// Compile-time type safety: Record().Prefault(invalidValue) will fail to compile
func (z *ZodRecord) Prefault(value map[interface{}]interface{}) ZodRecordPrefault {
	return ZodRecordPrefault{
		&ZodPrefault[*ZodRecord]{
			innerType:     z,
			prefaultValue: value,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the record schema
func (z *ZodRecord) PrefaultFunc(fn func() map[interface{}]interface{}) ZodRecordPrefault {
	genericFn := func() any { return fn() }
	return ZodRecordPrefault{
		&ZodPrefault[*ZodRecord]{
			innerType:    z,
			prefaultFunc: genericFn,
			isFunction:   true,
		},
	}
}

////////////////////////////
////   RECORDPREFAULT CHAINING METHODS ////
////////////////////////////

// Refine adds a flexible validation function to the record schema, returns ZodRecordPrefault
func (r ZodRecordPrefault) Refine(fn func(map[interface{}]interface{}) bool, params ...SchemaParams) ZodRecordPrefault {
	newInner := r.innerType.Refine(fn, params...)
	return ZodRecordPrefault{
		&ZodPrefault[*ZodRecord]{
			innerType:     newInner,
			prefaultValue: r.prefaultValue,
			prefaultFunc:  r.prefaultFunc,
			isFunction:    r.isFunction,
		},
	}
}

// Transform adds data transformation, returns generic ZodType to support transformation pipeline
func (r ZodRecordPrefault) Transform(fn func(map[interface{}]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodPrefault's TransformAny method
	return r.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of record values
		recordVal, isNil, err := extractRecordPointerValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilRecord
		}
		return fn(recordVal, ctx)
	})
}

// Optional makes the record optional
func (r ZodRecordPrefault) Optional() ZodType[any, any] {
	return Optional(any(r).(ZodType[any, any]))
}

// Nilable makes the record nilable
func (r ZodRecordPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(r).(ZodType[any, any]))
}

// GetZod returns the internal state for framework usage
func (z *ZodRecord) GetZod() *ZodRecordInternals {
	return z.internals
}

// Refine provides type-safe record validation with smart dereferencing support
func (z *ZodRecord) Refine(fn func(map[interface{}]interface{}) bool, params ...SchemaParams) *ZodRecord {
	result := z.RefineAny(func(v any) bool {
		recordMap, isNil, err := extractRecordPointerValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true // Let upper logic decide nil handling
		}
		return fn(recordMap)
	}, params...)
	return result.(*ZodRecord)
}

// RefineAny flexible version of validation that accepts any type
func (z *ZodRecord) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[any](fn, params...)
	return AddCheck(any(z).(ZodType[any, any]), check)
}

////////////////////////////
////   INTERNAL METHODS ////
////////////////////////////

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodRecord) CloneFrom(source any) {
	if src, ok := source.(*ZodRecord); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy key and value schemas
		z.internals.KeyType = src.internals.KeyType
		z.internals.ValueType = src.internals.ValueType
	}
}

////////////////////////////
////   RECORD DEFAULT WRAPPER ////
////////////////////////////

// ZodRecordDefault is a Default wrapper for record type
// Provides perfect type safety and chainable method support
type ZodRecordDefault struct {
	*ZodDefault[*ZodRecord] // Embed concrete pointer to enable method promotion
}

// Parse ensures correct validation call to inner type
func (s ZodRecordDefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodDefault.Parse(input, ctx...)
}

////////////////////////////
////   DEFAULT METHODS   ////
////////////////////////////

// Default default value for record
func (z *ZodRecord) Default(value map[interface{}]interface{}) ZodRecordDefault {
	return ZodRecordDefault{
		&ZodDefault[*ZodRecord]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc default value for record
func (z *ZodRecord) DefaultFunc(fn func() map[interface{}]interface{}) ZodRecordDefault {
	return ZodRecordDefault{
		&ZodDefault[*ZodRecord]{
			innerType:   z,
			defaultFunc: func() any { return fn() },
			isFunction:  true,
		},
	}
}

////////////////////////////
////   RECORDDEFAULT CHAINING METHODS ////
////////////////////////////

// Refine refine the record with given function
func (s ZodRecordDefault) Refine(fn func(map[interface{}]interface{}) bool, params ...SchemaParams) ZodRecordDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodRecordDefault{
		&ZodDefault[*ZodRecord]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds data transformation, returns generic ZodType to support transformation pipeline
func (s ZodRecordDefault) Transform(fn func(map[interface{}]interface{}, *RefinementContext) (any, error)) ZodType[any, any] {
	// Use embedded ZodDefault's TransformAny method
	return s.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		// Smart handling of record values
		recordVal, isNil, err := extractRecordPointerValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilRecord
		}
		return fn(recordVal, ctx)
	})
}

// Optional makes the record optional
func (s ZodRecordDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable makes the record nilable
func (s ZodRecordDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// createZodRecordFromDef creates a ZodRecord from definition
func createZodRecordFromDef(def *ZodRecordDef) *ZodRecord {
	internals := &ZodRecordInternals{
		ZodTypeInternals: newBaseZodTypeInternals(def.Type),
		Def:              def,
		KeyType:          def.KeyType,
		ValueType:        def.ValueType,
		Isst:             ZodIssueInvalidType{Expected: "record"},
		Bag:              make(map[string]interface{}),
	}

	// Set up simplified constructor for cloning
	internals.Constructor = func(newDef *ZodTypeDef) ZodType[any, any] {
		recordDef := &ZodRecordDef{
			ZodTypeDef: *newDef,
			Type:       ZodTypeRecord,
			KeyType:    def.KeyType,   // Preserve original key type
			ValueType:  def.ValueType, // Preserve original value type
		}
		return createZodRecordFromDef(recordDef)
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodRecord{internals: internals}
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

	schema := &ZodRecord{internals: internals}

	// Initialize the schema with proper error handling support
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

// NewZodRecord creates a new record schema
func NewZodRecord(keyType, valueType ZodType[any, any], params ...SchemaParams) *ZodRecord {
	def := &ZodRecordDef{
		ZodTypeDef: ZodTypeDef{
			Type:   ZodTypeRecord,
			Checks: make([]ZodCheck, 0),
		},
		Type:      ZodTypeRecord,
		KeyType:   keyType,
		ValueType: valueType,
	}

	schema := createZodRecordFromDef(def)

	// Apply schema parameters
	if len(params) > 0 {
		param := params[0]

		// Set coercion flag in dual-layer Bag as per guidelines
		if param.Coerce {
			schema.internals.Bag["coerce"] = true
			// Critical: Also set in ZodTypeInternals (dual-layer Bag setting)
			schema.internals.ZodTypeInternals.Bag["coerce"] = true
		}

		// Handle schema-level error mapping using utility function
		if param.Error != nil {
			errorMap := createErrorMap(param.Error)
			if errorMap != nil {
				def.Error = errorMap
				schema.internals.Error = errorMap
			}
		}

		// Handle additional parameters following unified pattern
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

////////////////////////////
////   HELPER FUNCTIONS ////
////////////////////////////

// extractRecordPointerValue extract record pointer value
func extractRecordPointerValue(input any) (map[interface{}]interface{}, bool, error) {
	switch v := input.(type) {
	case map[interface{}]interface{}:
		return v, false, nil
	case *map[interface{}]interface{}:
		if v == nil {
			return nil, true, nil
		}
		return *v, false, nil
	case map[string]interface{}:
		result := make(map[interface{}]interface{})
		for k, val := range v {
			result[k] = val
		}
		return result, false, nil
	case *map[string]interface{}:
		if v == nil {
			return nil, true, nil
		}
		result := make(map[interface{}]interface{})
		for k, val := range *v {
			result[k] = val
		}
		return result, false, nil
	default:
		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedRecord, input)
	}
}

// validateRecordAndRunChecks validate record and run checks
func (z *ZodRecord) validateRecordAndRunChecks(recordMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	if len(z.internals.Checks) > 0 {
		payload := &ParsePayload{
			Value:  recordMap,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(recordMap, z.internals.Checks, payload, nil)
		if len(payload.Issues) > 0 {
			return nil, &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, nil)}
		}
	}

	// Validate keys and values
	result := make(map[interface{}]interface{})
	for key, value := range recordMap {
		// Validate key - use tryApplyCoercion as per guidelines
		var validatedKey any
		var err error

		// Try to coerce key first (if child schema has coercion enabled)
		coercedKey, coerceErr := tryApplyCoercion(z.internals.KeyType, key)
		if coerceErr == nil {
			validatedKey, err = z.internals.KeyType.Parse(coercedKey)
		} else {
			validatedKey, err = z.internals.KeyType.Parse(key)
		}

		if err != nil {
			rawIssue := createInvalidKeyIssue(recordMap, key, []ZodRawIssue{
				{Code: "invalid_key", Message: err.Error(), Input: key},
			})
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}

		// Validate value - use tryApplyCoercion as per guidelines
		var validatedValue any

		// Try to coerce value first (if child schema has coercion enabled)
		coercedValue, coerceErr := tryApplyCoercion(z.internals.ValueType, value)
		if coerceErr == nil {
			validatedValue, err = z.internals.ValueType.Parse(coercedValue)
		} else {
			validatedValue, err = z.internals.ValueType.Parse(value)
		}

		if err != nil {
			rawIssue := NewRawIssue("invalid_value", recordMap)
			rawIssue.Message = err.Error()
			rawIssue.Path = []interface{}{key}
			finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}

		result[validatedKey] = validatedValue
	}
	return result, nil
}

// Record creates a new record validation schema
func Record(keyType, valueType ZodType[any, any], params ...SchemaParams) *ZodRecord {
	return NewZodRecord(keyType, valueType, params...)
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodRecord) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

////////////////////////////
////   PARTIAL RECORD   ////
////////////////////////////

// PartialRecord creates a partial record schema where keys are optional
func PartialRecord(keyType, valueType ZodType[any, any], params ...SchemaParams) *ZodRecord {
	// Create union([keyType, never()]) as key type
	// This makes keys optional, because never() will never match any value
	// But allows original keyType values to pass validation
	unionKeyType := Union([]ZodType[any, any]{keyType, Never()})

	// Use union key type to create Record
	// The resulting Record will accept keys conforming to keyType, but not require all possible keys to exist
	return NewZodRecord(unionKeyType, valueType, params...)
}
