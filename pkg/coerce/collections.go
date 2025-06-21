package coerce

import (
	"fmt"
	"reflect"

	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
	"github.com/kaptinlin/gozod/pkg/structx"
)

// =============================================================================
// COLLECTION CONVERSION - DELEGATE TO SPECIALIZED PACKAGES
// =============================================================================

// ToSlice converts input to []any using slicex package
func ToSlice(input any) ([]any, error) {
	if input == nil {
		return nil, NewUnsupportedError("nil", "slice")
	}

	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(input)
	if !ok {
		return nil, fmt.Errorf("cannot convert nil pointer to slice: %w", ErrNilPointer)
	}

	// Delegate to slicex package
	result, err := slicex.ToAny(derefed)
	if err != nil {
		return nil, NewUnsupportedError(fmt.Sprintf("%T", derefed), "slice")
	}

	return result, nil
}

// ToMap converts input to map[any]any using mapx package
func ToMap(input any) (map[any]any, error) {
	if input == nil {
		return nil, NewUnsupportedError("nil", "map[any]any")
	}

	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(input)
	if !ok {
		return nil, fmt.Errorf("cannot convert nil pointer to map: %w", ErrNilPointer)
	}

	// If the underlying value is a struct, attempt to marshal via structx
	if reflectx.IsStruct(derefed) {
		if objMap := structx.Marshal(derefed); objMap != nil {
			// Convert to map[any]any
			result := make(map[any]any)
			for k, v := range objMap {
				result[k] = v
			}
			return result, nil
		}
	}

	// Use reflection for other map types
	rv := reflect.ValueOf(derefed)
	if rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("input cannot be converted to map: %T", input)
	}

	result := make(map[any]any)
	for _, key := range rv.MapKeys() {
		result[key.Interface()] = rv.MapIndex(key).Interface()
	}
	return result, nil
}

// ToObject converts input to map[string]any using mapx package
func ToObject(input any) (map[string]any, error) {
	if input == nil {
		return nil, NewUnsupportedError("nil", "map[string]any")
	}

	// Use unified pointer dereferencing from reflectx
	derefed, ok := reflectx.Deref(input)
	if !ok {
		return nil, fmt.Errorf("cannot convert nil pointer to object: %w", ErrNilPointer)
	}

	// Delegate to mapx package
	result, err := mapx.ToStringKey(derefed)
	if err != nil {
		return nil, fmt.Errorf("cannot convert %T to object: %w", derefed, err)
	}

	return result, nil
}
