package mapx

// StringOr returns the string value for key, or def.
func StringOr(m map[string]any, key, def string) string {
	return ValueOr(m, key, def)
}

// BoolOr returns the bool value for key, or def.
func BoolOr(m map[string]any, key string, def bool) bool {
	return ValueOr(m, key, def)
}

// IntOr returns the int value for key (with numeric coercion),
// or def.
func IntOr(m map[string]any, key string, def int) int {
	if v, ok := Int(m, key); ok {
		return v
	}
	return def
}

// FloatOr returns the float64 value for key (with numeric
// coercion), or def.
func FloatOr(m map[string]any, key string, def float64) float64 {
	if v, ok := Float64(m, key); ok {
		return v
	}
	return def
}

// AnyOr returns the value for key, or def.
func AnyOr(m map[string]any, key string, def any) any {
	if v, ok := Get(m, key); ok {
		return v
	}
	return def
}

// StringsOr returns the []string value for key, or def.
func StringsOr(m map[string]any, key string, def []string) []string {
	return ValueOr(m, key, def)
}

// AnySliceOr returns the []any value for key, or def.
func AnySliceOr(m map[string]any, key string, def []any) []any {
	return ValueOr(m, key, def)
}

// MapOr returns the map[string]any value for key, or def.
func MapOr(m map[string]any, key string, def map[string]any) map[string]any {
	return ValueOr(m, key, def)
}
