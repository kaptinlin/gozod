package reflectx

import (
	"reflect"
	"slices"
)

// lengthKinds are reflect.Kinds that support len() semantics (string, array, slice).
var lengthKinds = [...]reflect.Kind{reflect.String, reflect.Array, reflect.Slice}

// sizeKinds are reflect.Kinds that support size semantics (map, chan, slice, array).
var sizeKinds = [...]reflect.Kind{reflect.Map, reflect.Chan, reflect.Slice, reflect.Array}

// StringVal returns the string value if v is a string.
// For nil or non-string values it returns ("", false).
func StringVal(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// HasLength reports whether v supports len() (string, array, or slice).
func HasLength(v any) bool {
	return isKindIn(v, lengthKinds[:])
}

// HasSize reports whether v has size semantics (map, chan, slice, or array).
func HasSize(v any) bool {
	return isKindIn(v, sizeKinds[:])
}

// Length returns len(v) for strings, arrays, and slices.
// It returns (0, false) when v does not support len().
func Length(v any) (int, bool) {
	return lenByKinds(v, lengthKinds[:])
}

// Size returns len(v) for maps, channels, slices, and arrays.
// It returns (0, false) when v does not support size semantics.
func Size(v any) (int, bool) {
	return lenByKinds(v, sizeKinds[:])
}

// isKindIn reports whether v's reflect.Kind is in the given set.
func isKindIn(v any, kinds []reflect.Kind) bool {
	if v == nil {
		return false
	}
	return slices.Contains(kinds, reflect.TypeOf(v).Kind())
}

// lenByKinds returns reflect.ValueOf(v).Len() when v's Kind is in the given set.
func lenByKinds(v any, kinds []reflect.Kind) (int, bool) {
	if v == nil {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	if slices.Contains(kinds, rv.Kind()) {
		return rv.Len(), true
	}
	return 0, false
}
