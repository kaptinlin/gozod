package mapx

// =============================================================================
// DEFAULT VALUE ACCESSORS
// =============================================================================

// GetStringDefault returns the string value for key, or defaultValue.
func GetStringDefault(props map[string]any, key, defaultValue string) string {
	return ValueOrDefault(props, key, defaultValue)
}

// GetBoolDefault returns the bool value for key, or defaultValue.
func GetBoolDefault(props map[string]any, key string, defaultValue bool) bool {
	return ValueOrDefault(props, key, defaultValue)
}

// GetIntDefault returns the int value for key (with numeric coercion),
// or defaultValue.
func GetIntDefault(props map[string]any, key string, defaultValue int) int {
	if v, ok := GetInt(props, key); ok {
		return v
	}
	return defaultValue
}

// GetFloat64Default returns the float64 value for key (with numeric
// coercion), or defaultValue.
func GetFloat64Default(props map[string]any, key string, defaultValue float64) float64 {
	if v, ok := GetFloat64(props, key); ok {
		return v
	}
	return defaultValue
}

// GetAnyDefault returns the value for key, or defaultValue.
func GetAnyDefault(props map[string]any, key string, defaultValue any) any {
	if v, ok := Get(props, key); ok {
		return v
	}
	return defaultValue
}

// GetStringsDefault returns the []string value for key, or defaultValue.
func GetStringsDefault(props map[string]any, key string, defaultValue []string) []string {
	return ValueOrDefault(props, key, defaultValue)
}

// GetAnySliceDefault returns the []any value for key, or defaultValue.
func GetAnySliceDefault(props map[string]any, key string, defaultValue []any) []any {
	return ValueOrDefault(props, key, defaultValue)
}

// GetMapDefault returns the map[string]any value for key, or defaultValue.
func GetMapDefault(props map[string]any, key string, defaultValue map[string]any) map[string]any {
	return ValueOrDefault(props, key, defaultValue)
}
