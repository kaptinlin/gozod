package reflectx

import (
	"fmt"
	"reflect"
)

// =============================================================================
// POINTER OPERATIONS
// =============================================================================

// IsNilPointer checks if a value is a nil pointer
func IsNilPointer(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

// IsPointerTo checks if a value is a pointer to a specific type
func IsPointerTo(v any, target reflect.Type) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return false
	}
	if rv.IsNil() {
		return false
	}
	return rv.Elem().Type() == target
}

// Deref dereferences a pointer once
func Deref(v any) (any, bool) {
	if v == nil {
		return nil, false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return v, true // Not a pointer, return as-is
	}
	if rv.IsNil() {
		return nil, false
	}
	return rv.Elem().Interface(), true
}

// DerefAll recursively dereferences all pointers
func DerefAll(v any) any {
	if v == nil {
		return nil
	}

	for {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			break
		}
		if rv.IsNil() {
			return nil
		}
		v = rv.Elem().Interface()
	}
	return v
}

// DerefTo dereferences a pointer to a specific type
func DerefTo[T any](v any) (T, bool) {
	var zero T
	if v == nil {
		return zero, false
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		// Try direct type assertion
		if result, ok := v.(T); ok {
			return result, true
		}
		return zero, false
	}

	if rv.IsNil() {
		return zero, false
	}

	elem := rv.Elem().Interface()
	if result, ok := elem.(T); ok {
		return result, true
	}
	return zero, false
}

// ToPointer creates a pointer to a value
func ToPointer(v any) any {
	if v == nil {
		return nil
	}

	rv := reflect.ValueOf(v)
	ptr := reflect.New(rv.Type())
	ptr.Elem().Set(rv)
	return ptr.Interface()
}

// ToPointerOf creates a pointer to a value of specific type
func ToPointerOf[T any](v T) *T {
	return &v
}

// PointerTo creates a pointer to a value of target type
func PointerTo(v any, target reflect.Type) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot create pointer to nil value")
	}

	rv := reflect.ValueOf(v)
	if rv.Type() != target {
		return nil, fmt.Errorf("value type %v does not match target type %v", rv.Type(), target)
	}

	ptr := reflect.New(target)
	ptr.Elem().Set(rv)
	return ptr.Interface(), nil
}

// PointerFrom gets value from a pointer
func PointerFrom(ptr any) (any, error) {
	if ptr == nil {
		return nil, fmt.Errorf("cannot get value from nil pointer")
	}

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("value is not a pointer")
	}

	if rv.IsNil() {
		return nil, fmt.Errorf("pointer is nil")
	}

	return rv.Elem().Interface(), nil
}

// =============================================================================
// POINTER UTILITIES
// =============================================================================

// PointerDepth returns the depth of pointer indirection
func PointerDepth(v any) int {
	if v == nil {
		return 0
	}

	depth := 0
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		depth++
		t = t.Elem()
	}
	return depth
}

// PointerChain returns the chain of types through pointer indirection
func PointerChain(v any) []reflect.Type {
	if v == nil {
		return nil
	}

	var chain []reflect.Type
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		chain = append(chain, t)
		t = t.Elem()
	}
	chain = append(chain, t) // Add the final non-pointer type
	return chain
}

// UltimateType returns the final non-pointer type
func UltimateType(v any) reflect.Type {
	if v == nil {
		return nil
	}

	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// UltimateValue returns the final non-pointer value
func UltimateValue(v any) (any, bool) {
	if v == nil {
		return nil, false
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}
	return rv.Interface(), true
}
