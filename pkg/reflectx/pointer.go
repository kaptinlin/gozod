package reflectx

import "reflect"

// Deref returns the value pointed to by v. If v is nil, not a pointer, or the
// pointer is nil, it returns (nil, false). When v is not a pointer the value is
// returned as-is with ok = true.
func Deref(v any) (val any, ok bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return v, true // already a concrete value
	}
	if rv.IsNil() {
		return nil, false
	}
	return rv.Elem().Interface(), true
}

// DerefAll recursively follows pointer chains until a non-pointer value (or
// nil) is reached. A nil input or nil pointer yields nil.
func DerefAll(v any) any {
	for v != nil {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return v
		}
		if rv.IsNil() {
			return nil
		}
		v = rv.Elem().Interface()
	}
	return nil
}
