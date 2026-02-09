package reflectx

import "reflect"

// ExtractString returns the string value if v is a string.
// For nil or non-string values it returns ("", false).
func ExtractString(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// HasLength reports whether v supports len() (string, array, or slice).
func HasLength(v any) bool {
	//nolint:exhaustive // default handles all other cases
	switch kindOf(v) {
	case reflect.String, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

// HasSize reports whether v has size semantics (map, chan, slice, or array).
func HasSize(v any) bool {
	//nolint:exhaustive // default handles all other cases
	switch kindOf(v) {
	case reflect.Map, reflect.Chan, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

// Length returns len(v) for strings, arrays, and slices.
// It returns (0, false) when v does not support len().
func Length(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	//nolint:exhaustive // default handles all other cases
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice:
		return rv.Len(), true
	default:
		return 0, false
	}
}

// Size returns len(v) for maps, channels, slices, and arrays.
// It returns (0, false) when v does not support size semantics.
func Size(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	//nolint:exhaustive // default handles all other cases
	switch rv.Kind() {
	case reflect.Map, reflect.Chan, reflect.Slice, reflect.Array:
		return rv.Len(), true
	default:
		return 0, false
	}
}
