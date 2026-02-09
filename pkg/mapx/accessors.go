package mapx

import "maps"

// Merge returns a new map containing all entries from both maps.
// Entries in props2 take precedence over entries in props1.
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
	maps.Copy(result, props2)
	return result
}

// =============================================================================
// TYPE-SAFE ACCESSORS (backward-compatible wrappers over ValueOf)
// =============================================================================

// GetString returns the string value for key via type assertion.
func GetString(props map[string]any, key string) (string, bool) {
	return ValueOf[string](props, key)
}

// GetBool returns the bool value for key via type assertion.
func GetBool(props map[string]any, key string) (bool, bool) {
	return ValueOf[bool](props, key)
}

// GetInt returns the int value for key, with numeric type coercion
// from int32, int64, and float64.
func GetInt(props map[string]any, key string) (int, bool) {
	switch v := props[key].(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetFloat64 returns the float64 value for key, with numeric type
// coercion from float32, int, int32, and int64.
func GetFloat64(props map[string]any, key string) (float64, bool) {
	switch v := props[key].(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetStrings returns the []string value for key via type assertion.
func GetStrings(props map[string]any, key string) ([]string, bool) {
	return ValueOf[[]string](props, key)
}

// GetAnySlice returns the []any value for key via type assertion.
func GetAnySlice(props map[string]any, key string) ([]any, bool) {
	return ValueOf[[]any](props, key)
}

// GetMap returns the map[string]any value for key via type assertion.
func GetMap(props map[string]any, key string) (map[string]any, bool) {
	return ValueOf[map[string]any](props, key)
}
