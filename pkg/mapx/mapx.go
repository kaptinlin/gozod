package mapx

import (
	"errors"
	"maps"
)

// ErrInputNotMap indicates that the provided input is not of map type.
var ErrInputNotMap = errors.New("input is not a map type")

// =============================================================================
// GENERIC ACCESSORS
// =============================================================================

// ValueOf performs a type assertion on the map value for key.
// It returns the typed value and true if the assertion succeeds.
// Reading from a nil map is safe and returns the zero value.
func ValueOf[T any](props map[string]any, key string) (T, bool) {
	v, ok := props[key].(T)
	return v, ok
}

// ValueOrDefault returns the typed value for key, or defaultValue
// when the key is missing or the type assertion fails.
func ValueOrDefault[T any](props map[string]any, key string, defaultValue T) T {
	if v, ok := props[key].(T); ok {
		return v
	}
	return defaultValue
}

// =============================================================================
// BASIC OPERATIONS
// =============================================================================

// Get returns the value for key and whether it was found.
func Get(props map[string]any, key string) (any, bool) {
	v, ok := props[key]
	return v, ok
}

// Set assigns value to key. It is a no-op on a nil map.
func Set(props map[string]any, key string, value any) {
	if props != nil {
		props[key] = value
	}
}

// Has reports whether key exists in the map.
func Has(props map[string]any, key string) bool {
	_, ok := props[key]
	return ok
}

// Count returns the number of entries in the map.
func Count(props map[string]any) int {
	return len(props)
}

// Copy returns a shallow clone of the map, or nil if props is nil.
func Copy(props map[string]any) map[string]any {
	if props == nil {
		return nil
	}
	return maps.Clone(props)
}
