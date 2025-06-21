package mapx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// =============================================================================
// MAP TYPE CHECKING
// =============================================================================

// Is checks if the value is a map type
func Is(v any) bool {
	return reflectx.IsMap(v)
}

// IsStringKey checks if the value is a map with string keys
func IsStringKey(v any) bool {
	if !Is(v) {
		return false
	}

	switch v.(type) {
	case map[string]any:
		return true
	default:
		return false
	}
}

// =============================================================================
// BASIC PROPERTY OPERATIONS
// =============================================================================

// Get safely gets a property from the properties map
func Get(props map[string]any, key string) (any, bool) {
	if props == nil {
		return nil, false
	}
	value, exists := props[key]
	return value, exists
}

// Set safely sets a property in the properties map
func Set(props map[string]any, key string, value any) {
	if props != nil {
		props[key] = value
	}
}

// Has checks if a property exists in the properties map
func Has(props map[string]any, key string) bool {
	if props == nil {
		return false
	}
	_, exists := props[key]
	return exists
}

// Count returns the number of properties in the map
func Count(props map[string]any) int {
	if props == nil {
		return 0
	}
	return len(props)
}

// Copy creates a shallow copy of the properties map
func Copy(props map[string]any) map[string]any {
	if props == nil {
		return nil
	}

	result := make(map[string]any)
	for key, value := range props {
		result[key] = value
	}
	return result
}

// Merge merges two properties maps, with the second taking precedence
func Merge(props1, props2 map[string]any) map[string]any {
	if props1 == nil && props2 == nil {
		return nil
	}
	if props1 == nil {
		return Copy(props2)
	}
	if props2 == nil {
		return Copy(props1)
	}

	result := Copy(props1)
	for key, value := range props2 {
		result[key] = value
	}
	return result
}

// =============================================================================
// TYPE-SAFE PROPERTY ACCESSORS
// =============================================================================

// GetString safely gets a string property from the properties map
func GetString(props map[string]any, key string) (string, bool) {
	if props == nil {
		return "", false
	}
	if value, ok := props[key].(string); ok {
		return value, true
	}
	return "", false
}

// GetBool safely gets a bool property from the properties map
func GetBool(props map[string]any, key string) (bool, bool) {
	if props == nil {
		return false, false
	}
	if value, ok := props[key].(bool); ok {
		return value, true
	}
	return false, false
}

// GetInt safely gets an int property from the properties map
func GetInt(props map[string]any, key string) (int, bool) {
	if props == nil {
		return 0, false
	}

	switch value := props[key].(type) {
	case int:
		return value, true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case float64:
		return int(value), true
	default:
		return 0, false
	}
}

// GetFloat64 safely gets a float64 property from the properties map
func GetFloat64(props map[string]any, key string) (float64, bool) {
	if props == nil {
		return 0, false
	}

	switch value := props[key].(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int32:
		return float64(value), true
	case int64:
		return float64(value), true
	default:
		return 0, false
	}
}

// GetStrings safely gets a []string property from the properties map
func GetStrings(props map[string]any, key string) ([]string, bool) {
	if props == nil {
		return nil, false
	}
	if value, ok := props[key].([]string); ok {
		return value, true
	}
	return nil, false
}

// GetAnySlice safely gets a []any property from the properties map
func GetAnySlice(props map[string]any, key string) ([]any, bool) {
	if props == nil {
		return nil, false
	}
	if value, ok := props[key].([]any); ok {
		return value, true
	}
	return nil, false
}

// GetMap safely gets a map[string]any property from the properties map
func GetMap(props map[string]any, key string) (map[string]any, bool) {
	if props == nil {
		return nil, false
	}
	if value, ok := props[key].(map[string]any); ok {
		return value, true
	}
	return nil, false
}

// =============================================================================
// DEFAULT VALUE PROPERTY ACCESSORS
// =============================================================================

// GetStringDefault returns a string property or default value
func GetStringDefault(props map[string]any, key, defaultValue string) string {
	if value, ok := GetString(props, key); ok {
		return value
	}
	return defaultValue
}

// GetBoolDefault returns a bool property or default value
func GetBoolDefault(props map[string]any, key string, defaultValue bool) bool {
	if value, ok := GetBool(props, key); ok {
		return value
	}
	return defaultValue
}

// GetIntDefault returns an int property or default value
func GetIntDefault(props map[string]any, key string, defaultValue int) int {
	if value, ok := GetInt(props, key); ok {
		return value
	}
	return defaultValue
}

// GetFloat64Default returns a float64 property or default value
func GetFloat64Default(props map[string]any, key string, defaultValue float64) float64 {
	if value, ok := GetFloat64(props, key); ok {
		return value
	}
	return defaultValue
}

// GetAnyDefault returns an any property or default value
func GetAnyDefault(props map[string]any, key string, defaultValue any) any {
	if value, ok := Get(props, key); ok {
		return value
	}
	return defaultValue
}

// GetStringsDefault returns a []string property or default value
func GetStringsDefault(props map[string]any, key string, defaultValue []string) []string {
	if value, ok := GetStrings(props, key); ok {
		return value
	}
	return defaultValue
}

// GetAnySliceDefault returns a []any property or default value
func GetAnySliceDefault(props map[string]any, key string, defaultValue []any) []any {
	if value, ok := GetAnySlice(props, key); ok {
		return value
	}
	return defaultValue
}

// GetMapDefault returns a map[string]any property or default value
func GetMapDefault(props map[string]any, key string, defaultValue map[string]any) map[string]any {
	if value, ok := GetMap(props, key); ok {
		return value
	}
	return defaultValue
}

// =============================================================================
// OBJECT OPERATIONS
// =============================================================================

// Keys returns the keys of an object (struct or map)
func Keys(input any) []string {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		return keys
	case map[any]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			if keyStr, ok := key.(string); ok {
				keys = append(keys, keyStr)
			}
		}
		return keys
	default:
		// Try to get keys from struct
		return getStructKeys(input)
	}
}

// Value gets a value from an object by key
func Value(input any, key string) (any, bool) {
	if input == nil {
		return nil, false
	}

	switch v := input.(type) {
	case map[string]any:
		value, exists := v[key]
		return value, exists
	case map[any]any:
		value, exists := v[key]
		return value, exists
	default:
		// Try to get value from struct
		return getStructValue(input, key)
	}
}

// =============================================================================
// MAP CONVERSION FUNCTIONS
// =============================================================================

// ToGeneric converts any map to map[any]any
func ToGeneric(input any) (map[any]any, error) {
	if input == nil {
		return nil, nil
	}

	switch v := input.(type) {
	case map[any]any:
		return v, nil
	case map[string]any:
		result := make(map[any]any)
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	case map[string]string:
		result := make(map[any]any)
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	case map[string]int:
		result := make(map[any]any)
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	case map[int]any:
		result := make(map[any]any)
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	default:
		// Use reflection for other map types
		rv := reflect.ValueOf(input)
		if rv.Kind() != reflect.Map {
			return nil, errors.New("input is not a map type")
		}

		result := make(map[any]any)
		for _, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			result[key.Interface()] = value.Interface()
		}
		return result, nil
	}
}

// ToStringKey converts any map to map[string]any
func ToStringKey(input any) (map[string]any, error) {
	if input == nil {
		return nil, nil
	}

	switch v := input.(type) {
	case map[string]any:
		return v, nil
	case map[string]string:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	case map[string]int:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	case map[any]any:
		result := make(map[string]any)
		for k, val := range v {
			if keyStr, ok := k.(string); ok {
				result[keyStr] = val
			}
		}
		return result, nil
	default:
		// Use reflection for other map types with string keys
		rv := reflect.ValueOf(input)
		if rv.Kind() != reflect.Map {
			// Try to convert struct to map
			if structMap := structx.Marshal(input); structMap != nil {
				return structMap, nil
			}
			return nil, errors.New("input cannot be converted to map[string]any")
		}

		// Check if keys are strings
		if rv.Type().Key().Kind() != reflect.String {
			return nil, errors.New("map keys are not strings")
		}

		result := make(map[string]any)
		for _, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			result[key.String()] = value.Interface()
		}
		return result, nil
	}
}

// FromAny converts various types to map[string]any
func FromAny(input any) map[string]any {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case map[string]any:
		return v
	case map[any]any:
		// Convert map[any]any to map[string]any
		result := make(map[string]any)
		for k, val := range v {
			if keyStr, ok := k.(string); ok {
				result[keyStr] = val
			}
		}
		return result
	default:
		// Try to convert struct to map using reflection
		return structx.Marshal(input)
	}
}

// =============================================================================
// MAP MERGE FUNCTIONS
// =============================================================================

// MergeMaps merges two maps of any type
func MergeMaps(a, b any) (any, error) {
	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		return b, nil
	}
	if b == nil {
		return a, nil
	}

	// Handle map[string]any
	if mapA, okA := a.(map[string]any); okA {
		if mapB, okB := b.(map[string]any); okB {
			return mergeStringMaps(mapA, mapB), nil
		}
	}

	// Handle map[any]any
	if mapA, okA := a.(map[any]any); okA {
		if mapB, okB := b.(map[any]any); okB {
			return mergeGenericMaps(mapA, mapB), nil
		}
	}

	// Try reflection-based merge for other map types
	return mergeReflectionMaps(a, b)
}

// =============================================================================
// MAP EXTRACTION FUNCTIONS
// =============================================================================

// Extract extracts map from input, returns the map and whether extraction was successful
func Extract(input any) (any, bool) {
	if input == nil {
		return nil, false
	}

	switch v := input.(type) {
	case map[string]any, map[any]any:
		return v, true
	default:
		return nil, false
	}
}

// ExtractRecord extracts a record (map[any]any) from input
func ExtractRecord(input any) (map[any]any, bool) {
	if input == nil {
		return nil, false
	}

	// Fast-path: already generic map types we support
	switch v := input.(type) {
	case map[any]any:
		return v, true
	case map[string]any:
		result := make(map[any]any, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result, true
	}

	// ---------------------------------------------------------------------
	// Reflection-based fallback – handle arbitrary map key types (e.g.
	// map[int]string, map[custom]Struct). We convert the map into a generic
	// map[any]any to align with Zod record/map processing, preserving original
	// key values as-is. This broadens support beyond string-key maps which the
	// original implementation was limited to.
	// ---------------------------------------------------------------------
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Map {
		result := make(map[any]any, rv.Len())
		for _, key := range rv.MapKeys() {
			result[key.Interface()] = rv.MapIndex(key).Interface()
		}
		return result, true
	}

	return nil, false
}

// =============================================================================
// VALIDATION FUNCTIONS
// =============================================================================

// ValidateType checks if a property has the expected type
func ValidateType(props map[string]any, key string, expectedType string) bool {
	if props == nil {
		return false
	}

	value, exists := props[key]
	if !exists {
		return false
	}

	actualType := reflectx.ParsedType(value)
	return string(actualType) == expectedType
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// getStructKeys extracts keys from a struct using reflection
func getStructKeys(input any) []string {
	structMap := structx.Marshal(input)
	if structMap == nil {
		return nil
	}

	keys := make([]string, 0, len(structMap))
	for key := range structMap {
		keys = append(keys, key)
	}
	return keys
}

// getStructValue gets a value from a struct by field name or json tag
func getStructValue(input any, key string) (any, bool) {
	if input == nil {
		return nil, false
	}

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, false
	}

	t := v.Type()

	// First, try to find by json tag
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check json tag
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			if parts[0] == key {
				return v.Field(i).Interface(), true
			}
		}
	}

	// Then, try to find by field name
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Name == key {
			return v.Field(i).Interface(), true
		}
	}

	return nil, false
}

// mergeStringMaps merges two map[string]any
func mergeStringMaps(a, b map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy from first map
	for k, v := range a {
		result[k] = v
	}

	// Merge from second map (overwrites)
	for k, v := range b {
		result[k] = v
	}

	return result
}

// mergeGenericMaps merges two map[any]any
func mergeGenericMaps(a, b map[any]any) map[any]any {
	result := make(map[any]any)

	// Copy from first map
	for k, v := range a {
		result[k] = v
	}

	// Merge from second map (overwrites)
	for k, v := range b {
		result[k] = v
	}

	return result
}

// mergeReflectionMaps merges maps using reflection
func mergeReflectionMaps(a, b any) (any, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Kind() != reflect.Map || vb.Kind() != reflect.Map {
		return nil, errors.New("both values must be maps")
	}

	if va.Type() != vb.Type() {
		return nil, errors.New("maps must have the same type")
	}

	result := reflect.MakeMap(va.Type())

	// Copy from first map
	for _, key := range va.MapKeys() {
		result.SetMapIndex(key, va.MapIndex(key))
	}

	// Merge from second map (overwrites)
	for _, key := range vb.MapKeys() {
		result.SetMapIndex(key, vb.MapIndex(key))
	}

	return result.Interface(), nil
}
