package mapx

// String returns the string value for key.
func String(m map[string]any, key string) (string, bool) {
	return ValueOf[string](m, key)
}

// Bool returns the bool value for key.
func Bool(m map[string]any, key string) (bool, bool) {
	return ValueOf[bool](m, key)
}

// IntCoerce returns the int value for key, with numeric
// coercion from int32, int64, and float64.
func IntCoerce(m map[string]any, key string) (int, bool) {
	switch v := m[key].(type) {
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

// Int returns the int value for key.
//
// Deprecated: use [IntCoerce] for explicit numeric coercion.
func Int(m map[string]any, key string) (int, bool) {
	return IntCoerce(m, key)
}

// Float64Coerce returns the float64 value for key, with
// numeric coercion from float32, int, int32, and int64.
func Float64Coerce(m map[string]any, key string) (float64, bool) {
	switch v := m[key].(type) {
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

// Float64 returns the float64 value for key.
//
// Deprecated: use [Float64Coerce] for explicit numeric coercion.
func Float64(m map[string]any, key string) (float64, bool) {
	return Float64Coerce(m, key)
}

// Strings returns the []string value for key.
func Strings(m map[string]any, key string) ([]string, bool) {
	return ValueOf[[]string](m, key)
}

// AnySlice returns the []any value for key.
func AnySlice(m map[string]any, key string) ([]any, bool) {
	return ValueOf[[]any](m, key)
}

// Map returns the map[string]any value for key.
func Map(m map[string]any, key string) (map[string]any, bool) {
	return ValueOf[map[string]any](m, key)
}
