package gozod

import (
	"errors"
	"fmt"
	"reflect"
)

// =============================================================================
// MAP TYPE DEFINITION (Go map[K]V validation)
// =============================================================================

// ZodMapDef defines the configuration for map validation
type ZodMapDef struct {
	ZodTypeDef
	Type      string            // "map"
	KeyType   ZodType[any, any] // Schema for validating keys
	ValueType ZodType[any, any] // Schema for validating values
}

// ZodMapInternals contains map validator internal state
type ZodMapInternals struct {
	ZodTypeInternals
	Def       *ZodMapDef             // Schema definition
	KeyType   ZodType[any, any]      // Key validation schema
	ValueType ZodType[any, any]      // Value validation schema
	Isst      ZodIssueInvalidType    // Invalid type issue template
	Bag       map[string]interface{} // Additional metadata
}

// ZodMap represents a map validation schema for Go map[K]V types
type ZodMap struct {
	internals *ZodMapInternals
}

// =============================================================================
// MAP TYPE UTILITY FUNCTIONS
// =============================================================================

// extractMapValue extracts map value handling nil cases
func extractMapValue(input any) (interface{}, bool, error) {
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
		if elem.Kind() == reflect.Map {
			return elem.Interface(), false, nil
		}
	}

	// Handle direct map types
	if reflect.TypeOf(input).Kind() == reflect.Map {
		return input, false, nil
	}

	rawIssue := CreateInvalidTypeIssue(input, "map", string(GetParsedType(input)))
	finalIssue := FinalizeIssue(rawIssue, nil, GetConfig())
	return nil, false, NewZodError([]ZodIssue{finalIssue})
}

// convertMapToGeneric converts various map types to map[interface{}]interface{}
func convertMapToGeneric(input interface{}) map[interface{}]interface{} {
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Map {
		return nil
	}

	result := make(map[interface{}]interface{})
	for _, key := range val.MapKeys() {
		result[key.Interface()] = val.MapIndex(key).Interface()
	}
	return result
}

// extractMapValueForRefine extract map value for refine
func extractMapValueForRefine(input any) (interface{}, bool, error) {
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
	case map[string]string:
		result := make(map[interface{}]interface{})
		for k, val := range v {
			result[k] = val
		}
		return result, false, nil
	case map[string]int:
		result := make(map[interface{}]interface{})
		for k, val := range v {
			result[k] = val
		}
		return result, false, nil
	default:
		rv := reflect.ValueOf(input)
		if rv.Kind() == reflect.Map {
			result := make(map[interface{}]interface{})
			for _, key := range rv.MapKeys() {
				result[key.Interface()] = rv.MapIndex(key).Interface()
			}
			return result, false, nil
		}
		return nil, false, fmt.Errorf("%w, got %T", ErrExpectedMap, input)
	}
}

// callMapRefineFunc call refine function for map
func callMapRefineFunc(fn any, mapValue interface{}) bool {
	if fn == nil {
		return true
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func || fnType.NumIn() != 1 || fnType.NumOut() != 1 {
		return false
	}

	if fnType.Out(0) != reflect.TypeOf(true) {
		return false
	}

	paramType := fnType.In(0)

	switch {
	case paramType == reflect.TypeOf(map[interface{}]interface{}{}):
		genericMap, isNil, err := extractMapValueForRefine(mapValue)
		if err != nil || isNil {
			return false
		}
		result := fnValue.Call([]reflect.Value{reflect.ValueOf(genericMap)})
		return result[0].Bool()

	case paramType.Kind() == reflect.Map:
		if convertedMap := convertMapValueToType(mapValue, paramType); convertedMap != nil {
			result := fnValue.Call([]reflect.Value{reflect.ValueOf(convertedMap)})
			return result[0].Bool()
		}
		return false

	default:
		return false
	}
}

// convertMapValueToType convert map value to specific type
func convertMapValueToType(mapValue interface{}, targetType reflect.Type) interface{} {
	if targetType.Kind() != reflect.Map {
		return nil
	}

	genericMap, isNil, err := extractMapValueForRefine(mapValue)
	if err != nil || isNil {
		return nil
	}

	sourceMap, ok := genericMap.(map[interface{}]interface{})
	if !ok {
		return nil
	}

	if targetType == reflect.TypeOf(map[interface{}]interface{}{}) {
		return sourceMap
	}

	result := reflect.MakeMap(targetType)
	keyType := targetType.Key()
	elemType := targetType.Elem()

	for key, value := range sourceMap {
		convertedKey, err := convertInterfaceToType(key, keyType)
		if err != nil {
			return nil
		}

		convertedVal, err := convertInterfaceToType(value, elemType)
		if err != nil {
			return nil
		}

		result.SetMapIndex(convertedKey, convertedVal)
	}

	return result.Interface()
}

// convertInterfaceToType convert interface{} value to target type
func convertInterfaceToType(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	sourceValue := reflect.ValueOf(value)
	sourceType := sourceValue.Type()

	if sourceType == targetType {
		return sourceValue, nil
	}

	if sourceType.ConvertibleTo(targetType) {
		return sourceValue.Convert(targetType), nil
	}

	if sourceType.Kind() == reflect.Interface {
		actualValue := sourceValue.Elem()
		if actualValue.IsValid() && actualValue.Type().ConvertibleTo(targetType) {
			return actualValue.Convert(targetType), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("%w %v to %v", ErrCannotConvertType, sourceType, targetType)
}

// callMapTransformFunc call transform function for map
func callMapTransformFunc(fn any, mapValue interface{}, ctx *RefinementContext) (any, error) {
	if fn == nil {
		return mapValue, nil
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func || fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		return nil, ErrTransformFunctionSignature
	}

	ctxType := reflect.TypeOf((*RefinementContext)(nil))
	if fnType.In(1) != ctxType {
		return nil, ErrTransformFunctionParameter
	}

	if fnType.Out(0) != reflect.TypeOf((*interface{})(nil)).Elem() || fnType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return nil, ErrTransformFunctionReturn
	}

	paramType := fnType.In(0)

	switch {
	case paramType == reflect.TypeOf(map[interface{}]interface{}{}):
		genericMap, isNil, err := extractMapValueForRefine(mapValue)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilMap
		}
		results := fnValue.Call([]reflect.Value{reflect.ValueOf(genericMap), reflect.ValueOf(ctx)})
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil

	case paramType.Kind() == reflect.Map:
		if convertedMap := convertMapValueToType(mapValue, paramType); convertedMap != nil {
			results := fnValue.Call([]reflect.Value{reflect.ValueOf(convertedMap), reflect.ValueOf(ctx)})
			if !results[1].IsNil() {
				return nil, results[1].Interface().(error)
			}
			return results[0].Interface(), nil
		}
		return nil, fmt.Errorf("%w %v", ErrFailedToConvertMapToRequired, paramType)

	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedTransformParameter, paramType)
	}
}

// convertGenericToMapType converts map[interface{}]interface{} back to the original map type
func convertGenericToMapType(genericMap map[interface{}]interface{}, originalType reflect.Type) interface{} {
	if originalType == nil {
		return genericMap
	}

	if originalType.Kind() == reflect.Ptr {
		elemType := originalType.Elem()
		if elemType.Kind() == reflect.Map {
			result := reflect.MakeMap(elemType)
			for k, v := range genericMap {
				key := reflect.ValueOf(k)
				value := reflect.ValueOf(v)
				if key.Type().ConvertibleTo(elemType.Key()) && value.Type().ConvertibleTo(elemType.Elem()) {
					result.SetMapIndex(key.Convert(elemType.Key()), value.Convert(elemType.Elem()))
				}
			}
			ptr := reflect.New(elemType)
			ptr.Elem().Set(result)
			return ptr.Interface()
		}
	} else if originalType.Kind() == reflect.Map {
		result := reflect.MakeMap(originalType)
		for k, v := range genericMap {
			key := reflect.ValueOf(k)
			value := reflect.ValueOf(v)
			if key.Type().ConvertibleTo(originalType.Key()) && value.Type().ConvertibleTo(originalType.Elem()) {
				result.SetMapIndex(key.Convert(originalType.Key()), value.Convert(originalType.Elem()))
			}
		}
		return result.Interface()
	}
	return genericMap
}

// validateMapKeysAndValues validates each key-value pair in the map
func validateMapKeysAndValues(inputMap map[interface{}]interface{}, keySchema, valueSchema ZodType[any, any], ctx *ParseContext) (map[interface{}]interface{}, error) {
	result := make(map[interface{}]interface{})
	var allErrors []ZodIssue

	for key, value := range inputMap {
		validatedKey, keyErr := keySchema.Parse(key, ctx)
		if keyErr != nil {
			var zodErr *ZodError
			if errors.As(keyErr, &zodErr) {
				for _, issue := range zodErr.Issues {
					issue.Path = append([]interface{}{key}, issue.Path...)
				}
				allErrors = append(allErrors, zodErr.Issues...)
			} else {
				rawIssue := CreateInvalidTypeIssue(key, "key", string(GetParsedType(key)))
				rawIssue.Path = []interface{}{key}
				finalIssue := FinalizeIssue(rawIssue, ctx, GetConfig())
				allErrors = append(allErrors, finalIssue)
			}
			continue
		}

		validatedValue, valueErr := valueSchema.Parse(value, ctx)
		if valueErr != nil {
			var zodErr *ZodError
			if errors.As(valueErr, &zodErr) {
				for _, issue := range zodErr.Issues {
					issue.Path = append([]interface{}{key}, issue.Path...)
				}
				allErrors = append(allErrors, zodErr.Issues...)
			} else {
				rawIssue := CreateInvalidTypeIssue(value, "value", string(GetParsedType(value)))
				rawIssue.Path = []interface{}{key}
				finalIssue := FinalizeIssue(rawIssue, ctx, GetConfig())
				allErrors = append(allErrors, finalIssue)
			}
			continue
		}

		result[validatedKey] = validatedValue
	}

	if len(allErrors) > 0 {
		return nil, NewZodError(allErrors)
	}

	return result, nil
}

// createZodMapFromDef creates a ZodMap instance from a definition following the unified pattern
func createZodMapFromDef(def *ZodMapDef, params ...SchemaParams) *ZodMap {
	// Create internals with modern pattern
	internals := &ZodMapInternals{
		ZodTypeInternals: newBaseZodTypeInternals("map"),
		Def:              def,
		KeyType:          def.KeyType,
		ValueType:        def.ValueType,
		Isst:             ZodIssueInvalidType{Expected: "map"},
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
		mapDef := &ZodMapDef{
			ZodTypeDef: *newDef,
			Type:       "map",
			KeyType:    def.KeyType,
			ValueType:  def.ValueType,
		}
		return any(createZodMapFromDef(mapDef)).(ZodType[any, any])
	}

	// Set up parse function
	internals.Parse = func(payload *ParsePayload, ctx *ParseContext) *ParsePayload {
		schema := &ZodMap{internals: internals}
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

	schema := &ZodMap{internals: internals}

	// Use unified infrastructure for initialization
	initZodType(any(schema).(ZodType[any, any]), &def.ZodTypeDef)

	return schema
}

// Map creates a new map schema with the given key and value schemas
func Map(keySchema, valueSchema ZodType[any, any], params ...SchemaParams) *ZodMap {
	// Create map definition
	def := &ZodMapDef{
		ZodTypeDef: ZodTypeDef{Type: "map"},
		Type:       "map",
		KeyType:    keySchema,
		ValueType:  valueSchema,
	}

	return createZodMapFromDef(def, params...)
}

// NewZodMap creates a new map schema with given key and value types
func NewZodMap(keySchema, valueSchema ZodType[any, any], params ...SchemaParams) *ZodMap {
	return Map(keySchema, valueSchema, params...)
}

// =============================================================================
// MAP TYPE METHODS
// =============================================================================

// GetInternals returns the internal state of the map schema
func (z *ZodMap) GetInternals() *ZodTypeInternals {
	return &z.internals.ZodTypeInternals
}

// GetZod returns the map-specific internals for framework usage
func (z *ZodMap) GetZod() *ZodMapInternals {
	return z.internals
}

// Coerce coerce input to map[interface{}]interface{}
func (z *ZodMap) Coerce(input interface{}) (interface{}, bool) {
	if coerced := tryCoerceToMap(input); coerced != nil {
		return coerced, true
	}
	return input, false
}

// Parse parse map with smart type inference
func (z *ZodMap) Parse(input any, ctx ...*ParseContext) (any, error) {
	parseCtx := (*ParseContext)(nil)
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	result, err := parseType[map[interface{}]interface{}](
		input,
		&z.internals.ZodTypeInternals,
		"map",
		func(v any) (map[interface{}]interface{}, bool) {
			mapValue, isNil, err := extractMapValue(v)
			if err != nil {
				return nil, false
			}
			if isNil {
				return nil, false
			}
			genericMap := convertMapToGeneric(mapValue)
			return genericMap, genericMap != nil
		},
		func(v any) (*map[interface{}]interface{}, bool) {
			if ptr, ok := v.(*map[interface{}]interface{}); ok {
				return ptr, true
			}
			return nil, false
		},
		func(value map[interface{}]interface{}, checks []ZodCheck, ctx *ParseContext) error {
			if len(checks) > 0 {
				payload := &ParsePayload{
					Value:  value,
					Issues: make([]ZodRawIssue, 0),
				}
				runChecksOnValue(value, checks, payload, ctx)
				if len(payload.Issues) > 0 {
					return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
				}
			}

			validatedMap, err := validateMapKeysAndValues(value, z.internals.KeyType, z.internals.ValueType, ctx)
			if err != nil {
				return err
			}

			for k := range value {
				delete(value, k)
			}
			for k, v := range validatedMap {
				value[k] = v
			}

			return nil
		},
		coerceToMap,
		parseCtx,
	)

	if err != nil {
		return nil, err
	}

	if result == nil {
		return result, nil
	}

	if genericMap, ok := result.(map[interface{}]interface{}); ok {
		originalType := reflect.TypeOf(input)
		return convertGenericToMapType(genericMap, originalType), nil
	}

	return result, nil
}

// MustParse parses the input and panics on validation failure
func (z *ZodMap) MustParse(input any, ctx ...*ParseContext) any {
	result, err := z.Parse(input, ctx...)
	if err != nil {
		panic(err)
	}
	return result
}

// tryCoerceToMap attempts to coerce input to a map with enhanced logic
func tryCoerceToMap(input interface{}) interface{} {
	// Handle nil input - create empty map
	if input == nil {
		return make(map[interface{}]interface{})
	}

	// Check if already a map and try to convert to general map type
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Map {
		// Convert any map type to map[interface{}]interface{}
		result := make(map[interface{}]interface{})
		for _, key := range rv.MapKeys() {
			keyInterface := key.Interface()
			valueInterface := rv.MapIndex(key).Interface()
			result[keyInterface] = valueInterface
		}
		return result
	}

	// Try struct-to-map conversion for struct types
	if IsStructType(input) {
		if mapped := convertToMapWithTags(input, []string{"json", "yaml"}); mapped != nil {
			// Convert map[string]interface{} to map[interface{}]interface{}
			result := make(map[interface{}]interface{})
			for k, v := range mapped {
				result[k] = v
			}
			return result
		}
	}

	// Use convertToMap from utils.go for other object-like types
	if mapped := convertToMap(input); mapped != nil {
		// Convert map[string]interface{} to map[interface{}]interface{} for general use
		result := make(map[interface{}]interface{})
		for k, v := range mapped {
			result[k] = v
		}
		return result
	}

	// Use getObjectKeys and getObjectValue for additional object-like types
	keys := getObjectKeys(input)
	if len(keys) > 0 {
		result := make(map[interface{}]interface{})
		for _, key := range keys {
			if value, exists := getObjectValue(input, key); exists {
				result[key] = value
			}
		}
		return result
	}

	// Return nil if coercion is not possible
	return nil
}

// Transform transform map with given function
func (z *ZodMap) Transform(fn any) ZodType[any, any] {
	return z.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		mapValue, isNil, err := extractMapValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilMap
		}

		return callMapTransformFunc(fn, mapValue, ctx)
	})
}

// TransformAny creates a transform with given function
func (z *ZodMap) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	transform := NewZodTransform[any, any](fn)

	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: any(transform).(ZodType[any, any]),
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Pipe creates a pipe with another schema
func (z *ZodMap) Pipe(next ZodType[any, any]) ZodType[any, any] {
	return &ZodPipe[any, any]{
		in:  any(z).(ZodType[any, any]),
		out: next,
		def: ZodTypeDef{Type: "pipe"},
	}
}

// Optional makes the map optional
func (z *ZodMap) Optional() ZodType[any, any] {
	return any(Optional(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Nilable make the map nilable
func (z *ZodMap) Nilable() ZodType[any, any] {
	return Clone(z, func(def *ZodTypeDef) {
	}).(*ZodMap).setNilable()
}

// setNilable set nilable flag
func (z *ZodMap) setNilable() ZodType[any, any] {
	z.internals.Nilable = true
	return z
}

// Nullish makes the map both optional and nilable
func (z *ZodMap) Nullish() ZodType[any, any] {
	return any(Nullish(any(z).(ZodType[any, any]))).(ZodType[any, any])
}

// Refine adds a flexible validation function to the map - supports any function type
func (z *ZodMap) Refine(fn any, params ...SchemaParams) *ZodMap {
	result := z.RefineAny(func(v any) bool {
		mapValue, isNil, err := extractMapValue(v)
		if err != nil {
			return false
		}
		if isNil {
			return true
		}

		return callMapRefineFunc(fn, mapValue)
	}, params...)
	return result.(*ZodMap)
}

// RefineAny adds a custom validation function to the map
func (z *ZodMap) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	check := NewCustom[interface{}](fn, params...)
	return AddCheck(z, check)
}

// Check adds modern validation using direct payload access
func (z *ZodMap) Check(fn CheckFn) *ZodMap {
	check := NewCustom[map[interface{}]interface{}](fn, SchemaParams{})
	result := AddCheck(z, check)
	return result.(*ZodMap)
}

// Length adds a size validation check to the map
func (z *ZodMap) Length(size int, params ...SchemaParams) *ZodMap {
	check := NewZodCheckMapSize(size, params...)
	result := AddCheck(z, check)
	return result.(*ZodMap)
}

// Min adds a minimum size validation check to the map
func (z *ZodMap) Min(minimum int, params ...SchemaParams) *ZodMap {
	check := NewZodCheckMapMinSize(minimum, params...)
	result := AddCheck(z, check)
	return result.(*ZodMap)
}

// Max adds a maximum size validation check to the map
func (z *ZodMap) Max(maximum int, params ...SchemaParams) *ZodMap {
	check := NewZodCheckMapMaxSize(maximum, params...)
	result := AddCheck(z, check)
	return result.(*ZodMap)
}

// ZodMapDefault is a Default wrapper for map type
type ZodMapDefault struct {
	*ZodDefault[*ZodMap]
}

type ZodMapPrefault struct {
	*ZodPrefault[*ZodMap] // embed specific pointer, allow method promotion
}

// Default adds a default value to the map
func (z *ZodMap) Default(value any) ZodMapDefault {
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    z,
			defaultValue: value,
			isFunction:   false,
		},
	}
}

// DefaultFunc adds a default function to the map
func (z *ZodMap) DefaultFunc(fn func() any) ZodMapDefault {
	genericFn := func() any { return fn() }
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:   z,
			defaultFunc: genericFn,
			isFunction:  true,
		},
	}
}

// Prefault adds a prefault value to the map
func (z *ZodMap) Prefault(value any) ZodMapPrefault {
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     internals,
			innerType:     z,
			prefaultValue: value,
			prefaultFunc:  nil,
			isFunction:    false,
		},
	}
}

// PrefaultFunc adds a prefault function to the map
func (z *ZodMap) PrefaultFunc(fn func() any) ZodMapPrefault {
	baseInternals := z.GetInternals()
	internals := &ZodTypeInternals{
		Version:     baseInternals.Version,
		Type:        ZodTypePrefault,
		Checks:      baseInternals.Checks,
		Coerce:      baseInternals.Coerce,
		Optional:    baseInternals.Optional,
		Nilable:     baseInternals.Nilable,
		Constructor: baseInternals.Constructor,
		Values:      baseInternals.Values,
		Pattern:     baseInternals.Pattern,
		Error:       baseInternals.Error,
		Bag:         baseInternals.Bag,
	}

	genericFn := func() any { return fn() }
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     internals,
			innerType:     z,
			prefaultValue: nil,
			prefaultFunc:  genericFn,
			isFunction:    true,
		},
	}
}

// CloneFrom implements Cloneable interface for type-specific state copying
func (z *ZodMap) CloneFrom(source any) {
	if src, ok := source.(*ZodMap); ok {
		// Copy type-specific fields from Bag
		if src.internals.Bag != nil {
			if z.internals.Bag == nil {
				z.internals.Bag = make(map[string]interface{})
			}
			for k, v := range src.internals.Bag {
				z.internals.Bag[k] = v
			}
		}

		// Copy map-specific fields
		z.internals.KeyType = src.internals.KeyType
		z.internals.ValueType = src.internals.ValueType
	}
}

// Unwrap returns the inner type (for basic types, returns self)
func (z *ZodMap) Unwrap() ZodType[any, any] {
	return any(z).(ZodType[any, any])
}

// ==================== ZodMapDefault chain call methods ====================

// Length adds a size validation check to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Length(size int, params ...SchemaParams) ZodMapDefault {
	newInner := s.innerType.Length(size, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Min adds a minimum size validation check to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Min(minimum int, params ...SchemaParams) ZodMapDefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Max adds a maximum size validation check to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Max(maximum int, params ...SchemaParams) ZodMapDefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Refine(fn any, params ...SchemaParams) ZodMapDefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Transform adds a data transformation function to the map, returns ZodType support transform pipeline
func (s ZodMapDefault) Transform(fn any) ZodType[any, any] {
	return s.ZodDefault.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		mapValue, isNil, err := extractMapValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilMap
		}
		return callMapTransformFunc(fn, mapValue, ctx)
	})
}

// Check adds a modern validation function to the map, returns ZodMapDefault support chain call
func (s ZodMapDefault) Check(fn CheckFn) ZodMapDefault {
	newInner := s.innerType.Check(fn)
	return ZodMapDefault{
		&ZodDefault[*ZodMap]{
			innerType:    newInner,
			defaultValue: s.defaultValue,
			defaultFunc:  s.defaultFunc,
			isFunction:   s.isFunction,
		},
	}
}

// Optional adds an optional check to the map, returns ZodType support chain call
func (s ZodMapDefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the map, returns ZodType support chain call
func (s ZodMapDefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ==================== ZodMapPrefault chain call methods ====================

// Length adds a size validation check to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Length(size int, params ...SchemaParams) ZodMapPrefault {
	newInner := s.innerType.Length(size, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Min adds a minimum size validation check to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Min(minimum int, params ...SchemaParams) ZodMapPrefault {
	newInner := s.innerType.Min(minimum, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Max adds a maximum size validation check to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Max(maximum int, params ...SchemaParams) ZodMapPrefault {
	newInner := s.innerType.Max(maximum, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Refine adds a flexible validation function to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Refine(fn any, params ...SchemaParams) ZodMapPrefault {
	newInner := s.innerType.Refine(fn, params...)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Transform adds a data transformation function to the map, returns ZodType support transform pipeline
func (s ZodMapPrefault) Transform(fn any) ZodType[any, any] {
	return s.ZodPrefault.TransformAny(func(input any, ctx *RefinementContext) (any, error) {
		mapValue, isNil, err := extractMapValue(input)
		if err != nil {
			return nil, err
		}
		if isNil {
			return nil, ErrTransformNilMap
		}
		return callMapTransformFunc(fn, mapValue, ctx)
	})
}

// TransformAny adds a generic data transformation function to the map, returns ZodType support transform pipeline
func (s ZodMapPrefault) TransformAny(fn func(any, *RefinementContext) (any, error)) ZodType[any, any] {
	return s.ZodPrefault.TransformAny(fn)
}

// Check adds a modern validation function to the map, returns ZodMapPrefault support chain call
func (s ZodMapPrefault) Check(fn CheckFn) ZodMapPrefault {
	newInner := s.innerType.Check(fn)
	return ZodMapPrefault{
		&ZodPrefault[*ZodMap]{
			internals:     s.internals,
			innerType:     newInner,
			prefaultValue: s.prefaultValue,
			prefaultFunc:  s.prefaultFunc,
			isFunction:    s.isFunction,
		},
	}
}

// Optional adds an optional check to the map, returns ZodType support chain call
func (s ZodMapPrefault) Optional() ZodType[any, any] {
	return Optional(any(s).(ZodType[any, any]))
}

// Nilable adds a nilable check to the map, returns ZodType support chain call
func (s ZodMapPrefault) Nilable() ZodType[any, any] {
	return Nilable(any(s).(ZodType[any, any]))
}

// ==================== ZodMapPrefault core interface methods ====================

// Parse implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Parse(input any, ctx ...*ParseContext) (any, error) {
	return s.ZodPrefault.Parse(input, ctx...)
}

// MustParse implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) MustParse(input any, ctx ...*ParseContext) any {
	return s.ZodPrefault.MustParse(input, ctx...)
}

// GetInternals implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) GetInternals() *ZodTypeInternals {
	return s.ZodPrefault.GetInternals()
}

// Pipe implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Pipe(out ZodType[any, any]) ZodType[any, any] {
	return s.ZodPrefault.Pipe(out)
}

// RefineAny implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) RefineAny(fn func(any) bool, params ...SchemaParams) ZodType[any, any] {
	return s.ZodPrefault.RefineAny(fn, params...)
}

// Unwrap implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Unwrap() ZodType[any, any] {
	return s.ZodPrefault.Unwrap()
}

// Prefault implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) Prefault(value any) ZodType[any, any] {
	return s.ZodPrefault.Prefault(value)
}

// PrefaultFunc implements ZodType interface, delegates to embedded ZodPrefault
func (s ZodMapPrefault) PrefaultFunc(fn func() any) ZodType[any, any] {
	return s.ZodPrefault.PrefaultFunc(fn)
}
