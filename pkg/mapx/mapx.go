package mapx

import (
	"errors"
	"maps"
)

// ErrInputNotMap indicates that the input is not a map type.
var ErrInputNotMap = errors.New("input is not a map type")

// ValueOf performs a type assertion on m[key].
// Reading from a nil map is safe and returns the zero value.
func ValueOf[T any](m map[string]any, key string) (T, bool) {
	v, ok := m[key].(T)
	return v, ok
}

// ValueOr returns the typed value for key, or def when the key
// is missing or the type assertion fails.
func ValueOr[T any](m map[string]any, key string, def T) T {
	if v, ok := m[key].(T); ok {
		return v
	}
	return def
}

// Get returns the value for key and whether it was found.
func Get(m map[string]any, key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

// Set assigns value to key. It is a no-op on a nil map.
func Set(m map[string]any, key string, value any) {
	if m != nil {
		m[key] = value
	}
}

// Has reports whether key exists in the map.
func Has(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}

// Count returns the number of entries.
func Count(m map[string]any) int {
	return len(m)
}

// Copy returns a shallow clone, or nil if m is nil.
func Copy(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	return maps.Clone(m)
}

// Merge returns a new map containing all entries from both maps.
// Entries in b take precedence over entries in a.
func Merge(a, b map[string]any) map[string]any {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return Copy(b)
	}
	if b == nil {
		return Copy(a)
	}
	result := Copy(a)
	maps.Copy(result, b)
	return result
}
