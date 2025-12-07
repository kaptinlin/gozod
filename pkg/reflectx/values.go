package reflectx

import "reflect"

// ExtractString returns the string value if v is a string. Otherwise ok=false.
func ExtractString(v any) (str string, ok bool) {
	if v == nil {
		return "", false
	}
	str, ok = v.(string)
	return
}

// HasLength reports whether v supports the built-in len() function (string,
// array, slice).
func HasLength(v any) bool {
	if v == nil {
		return false
	}
	//nolint:exhaustive // default handles all other cases
	switch reflect.TypeOf(v).Kind() {
	case reflect.String, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

// HasSize reports whether v has size semantics (map, chan, slice, array).
func HasSize(v any) bool {
	if v == nil {
		return false
	}
	//nolint:exhaustive // default handles all other cases
	switch reflect.TypeOf(v).Kind() {
	case reflect.Map, reflect.Chan, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

// GetLength returns len(v) and ok=true when HasLength(v) is true.
func GetLength(v any) (length int, ok bool) {
	if !HasLength(v) {
		return 0, false
	}
	return reflect.ValueOf(v).Len(), true
}

// GetSize returns len(v) and ok=true when HasSize(v) is true.
func GetSize(v any) (size int, ok bool) {
	if !HasSize(v) {
		return 0, false
	}
	return reflect.ValueOf(v).Len(), true
}
