package reflectx

import (
	"fmt"
	"reflect"
)

// =============================================================================
// COLLECTION OPERATIONS
// =============================================================================

// Length gets the length of a value (0 if not applicable)
func Length(v any) int {
	if length, ok := GetLength(v); ok {
		return length
	}
	return 0
}

// Size gets the size of a value (0 if not applicable)
func Size(v any) int {
	if size, ok := GetSize(v); ok {
		return size
	}
	return 0
}

// Capacity gets the capacity of a value (0 if not applicable)
func Capacity(v any) int {
	if v == nil {
		return 0
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Chan:
		return rv.Cap()
	default:
		return 0
	}
}

// IsEmpty checks if a value is empty
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// IsCollection checks if a value is a collection type
func IsCollection(v any) bool {
	if v == nil {
		return false
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

// =============================================================================
// COLLECTION CONVERSION
// =============================================================================

// ToSliceOf converts a value to a slice of specified type (using coerce)
// Note: This is a placeholder - actual implementation should use coerce package
func ToSliceOf[T any](v any) ([]T, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot convert nil to slice")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("value is not a slice or array")
	}

	result := make([]T, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if converted, ok := elem.(T); ok {
			result[i] = converted
		} else {
			return nil, fmt.Errorf("cannot convert element at index %d to target type", i)
		}
	}
	return result, nil
}

// ToMapOf converts a value to a map of specified types (using coerce)
// Note: This is a placeholder - actual implementation should use coerce package
func ToMapOf[K comparable, V any](v any) (map[K]V, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot convert nil to map")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("value is not a map")
	}

	result := make(map[K]V, rv.Len())
	for _, key := range rv.MapKeys() {
		keyInterface := key.Interface()
		valueInterface := rv.MapIndex(key).Interface()

		convertedKey, keyOk := keyInterface.(K)
		convertedValue, valueOk := valueInterface.(V)

		if !keyOk {
			return nil, fmt.Errorf("cannot convert key to target type")
		}
		if !valueOk {
			return nil, fmt.Errorf("cannot convert value to target type")
		}

		result[convertedKey] = convertedValue
	}
	return result, nil
}

// =============================================================================
// COLLECTION OPERATIONS
// =============================================================================

// ForEach iterates over a collection and calls the function for each element
func ForEach(v any, fn func(index int, value any) bool) error {
	if v == nil {
		return fmt.Errorf("cannot iterate over nil value")
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if !fn(i, rv.Index(i).Interface()) {
				break
			}
		}
		return nil
	case reflect.Map:
		i := 0
		for _, key := range rv.MapKeys() {
			if !fn(i, rv.MapIndex(key).Interface()) {
				break
			}
			i++
		}
		return nil
	case reflect.String:
		str := v.(string)
		for i, r := range str {
			if !fn(i, string(r)) {
				break
			}
		}
		return nil
	default:
		return fmt.Errorf("value is not iterable")
	}
}

// Map applies a function to each element and returns a new collection
func Map(v any, fn func(any) any) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot map over nil value")
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = fn(rv.Index(i).Interface())
		}
		return result, nil
	case reflect.Map:
		result := make(map[any]any, rv.Len())
		for _, key := range rv.MapKeys() {
			keyInterface := key.Interface()
			valueInterface := rv.MapIndex(key).Interface()
			result[keyInterface] = fn(valueInterface)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("value is not mappable")
	}
}

// Filter filters elements based on a predicate function
func Filter(v any, fn func(any) bool) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot filter nil value")
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		var result []any
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			if fn(elem) {
				result = append(result, elem)
			}
		}
		return result, nil
	case reflect.Map:
		result := make(map[any]any)
		for _, key := range rv.MapKeys() {
			keyInterface := key.Interface()
			valueInterface := rv.MapIndex(key).Interface()
			if fn(valueInterface) {
				result[keyInterface] = valueInterface
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("value is not filterable")
	}
}
