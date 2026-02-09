package reflectx

import "reflect"

// Deref dereferences a single pointer level. For non-pointer concrete values
// it returns (v, true). For nil, typed nil pointers, or nil pointer values
// it returns (nil, false).
func Deref(v any) (any, bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return v, true // already a concrete value
	}
	if rv.IsNil() {
		return nil, false
	}
	return rv.Elem().Interface(), true
}

// DerefAll recursively follows pointer chains until a non-pointer value is
// reached. A nil input or nil pointer at any level yields nil.
func DerefAll(v any) any {
	for v != nil {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Pointer {
			return v
		}
		if rv.IsNil() {
			return nil
		}
		v = rv.Elem().Interface()
	}
	return nil
}
