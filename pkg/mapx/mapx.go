package mapx

import (
	"errors"
	"reflect"

	"github.com/kaptinlin/gozod/pkg/structx"
)

// ErrInputNotMap indicates that the provided input is not of map type.
var ErrInputNotMap = errors.New("input is not a map type")

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
			return nil, ErrInputNotMap
		}

		result := make(map[any]any)
		for _, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			result[key.Interface()] = value.Interface()
		}
		return result, nil
	}
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
