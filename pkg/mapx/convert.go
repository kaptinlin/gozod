package mapx

import (
	"maps"
	"reflect"
	"slices"

	"github.com/kaptinlin/gozod/pkg/structx"
)

// Keys returns the string keys of an object (map or struct).
// For map[any]any, only string keys are included.
func Keys(v any) []string {
	if v == nil {
		return nil
	}

	switch x := v.(type) {
	case map[string]any:
		return slices.Collect(maps.Keys(x))
	case map[any]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			if s, ok := k.(string); ok {
				keys = append(keys, s)
			}
		}
		return keys
	default:
		return structKeys(v)
	}
}

// ToGeneric converts any map type to map[any]any.
// It returns ErrInputNotMap for non-map inputs.
func ToGeneric(v any) (map[any]any, error) {
	if v == nil {
		return nil, nil
	}

	switch x := v.(type) {
	case map[any]any:
		return x, nil
	case map[string]any:
		return convertMap(x), nil
	case map[string]string:
		return convertMap(x), nil
	case map[string]int:
		return convertMap(x), nil
	case map[int]any:
		return convertMap(x), nil
	default:
		return toGenericReflect(v)
	}
}

// convertMap converts a typed map to map[any]any.
func convertMap[K comparable, V any](m map[K]V) map[any]any {
	result := make(map[any]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// structKeys extracts exported field names from a struct.
func structKeys(v any) []string {
	m := structx.Marshal(v)
	if m == nil {
		return nil
	}
	return slices.Collect(maps.Keys(m))
}

// toGenericReflect converts an arbitrary map to map[any]any via reflection.
func toGenericReflect(v any) (map[any]any, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, ErrInputNotMap
	}
	result := make(map[any]any, rv.Len())
	for _, key := range rv.MapKeys() {
		result[key.Interface()] = rv.MapIndex(key).Interface()
	}
	return result, nil
}
