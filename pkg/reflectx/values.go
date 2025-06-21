package reflectx

import "reflect"

// =============================================================================
// VALUE EXTRACTION - Get*/Extract* functions
// =============================================================================

// GetValue safely gets the reflect.Value of a value
func GetValue(v any) (reflect.Value, bool) {
	if v == nil {
		return reflect.Value{}, false
	}
	rv := reflect.ValueOf(v)
	return rv, rv.IsValid()
}

// GetType safely gets the reflect.Type of a value
func GetType(v any) (reflect.Type, bool) {
	if v == nil {
		return nil, false
	}
	return reflect.TypeOf(v), true
}

// GetKind gets the reflect.Kind of a value
func GetKind(v any) reflect.Kind {
	if v == nil {
		return reflect.Invalid
	}
	return reflect.TypeOf(v).Kind()
}

// ExtractBool extracts a bool value (direct type only)
func ExtractBool(v any) (bool, bool) {
	if v == nil {
		return false, false
	}
	if b, ok := v.(bool); ok {
		return b, true
	}
	return false, false
}

// ExtractString extracts a string value (direct type only)
func ExtractString(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

// ExtractInt extracts an integer value as int64 (all int types unified)
func ExtractInt(v any) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	default:
		return 0, false
	}
}

// ExtractUint extracts an unsigned integer value as uint64 (all uint types unified)
func ExtractUint(v any) (uint64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case uint:
		return uint64(val), true
	case uint8:
		return uint64(val), true
	case uint16:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	case uintptr:
		return uint64(val), true
	default:
		return 0, false
	}
}

// ExtractFloat extracts a float value as float64 (all float types unified)
func ExtractFloat(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// ExtractComplex extracts a complex value as complex128 (all complex types unified)
func ExtractComplex(v any) (complex128, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case complex64:
		return complex128(val), true
	case complex128:
		return val, true
	default:
		return 0, false
	}
}

// ExtractSlice extracts a slice as []any
func ExtractSlice(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}

	// Fast path for []any
	if slice, ok := v.([]any); ok {
		return slice, true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	result := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = rv.Index(i).Interface()
	}
	return result, true
}

// ExtractMap extracts a map as map[any]any
func ExtractMap(v any) (map[any]any, bool) {
	if v == nil {
		return nil, false
	}

	// Fast path for map[any]any
	if m, ok := v.(map[any]any); ok {
		return m, true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, false
	}

	result := make(map[any]any, rv.Len())
	for _, key := range rv.MapKeys() {
		result[key.Interface()] = rv.MapIndex(key).Interface()
	}
	return result, true
}

// ExtractStruct extracts a struct as map[string]any
func ExtractStruct(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return nil, false
	}

	rt := rv.Type()
	result := make(map[string]any, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.IsExported() {
			result[field.Name] = rv.Field(i).Interface()
		}
	}
	return result, true
}

// =============================================================================
// VALUE PROPERTIES - Has* functions
// =============================================================================

// HasLength checks if a value has a length property
func HasLength(v any) bool {
	if v == nil {
		return false
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.String, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

// HasSize checks if a value has a size property
func HasSize(v any) bool {
	if v == nil {
		return false
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Map, reflect.Chan, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

// HasFields checks if a value has fields (is a struct)
func HasFields(v any) bool {
	return IsStruct(v)
}

// HasMethods checks if a value has methods
func HasMethods(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).NumMethod() > 0
}

// GetLength gets the length of a value
func GetLength(v any) (int, bool) {
	if !HasLength(v) {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	return rv.Len(), true
}

// GetSize gets the size of a value
func GetSize(v any) (int, bool) {
	if !HasSize(v) {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	return rv.Len(), true
}

// GetFieldCount gets the number of fields in a struct
func GetFieldCount(v any) (int, bool) {
	if !IsStruct(v) {
		return 0, false
	}
	return reflect.TypeOf(v).NumField(), true
}

// GetMethodCount gets the number of methods of a value
func GetMethodCount(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	return reflect.TypeOf(v).NumMethod(), true
}
