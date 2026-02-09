package mapx

import (
	"iter"
	"maps"
	"reflect"
	"slices"

	"github.com/kaptinlin/gozod/pkg/structx"
)

// Keys returns the string keys of an object (map or struct).
// For map[any]any, only string keys are included.
func Keys(input any) []string {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case map[string]any:
		return slices.Collect(mapStringKeys(v))
	case map[any]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			if s, ok := k.(string); ok {
				keys = append(keys, s)
			}
		}
		return keys
	default:
		return structKeys(input)
	}
}

// ToGeneric converts any map type to map[any]any.
// It returns ErrInputNotMap for non-map inputs.
func ToGeneric(input any) (map[any]any, error) {
	if input == nil {
		return nil, nil
	}

	switch v := input.(type) {
	case map[any]any:
		return v, nil
	case map[string]any:
		return convertMap(v), nil
	case map[string]string:
		return convertMap(v), nil
	case map[string]int:
		return convertMap(v), nil
	case map[int]any:
		return convertMap(v), nil
	default:
		return toGenericReflect(input)
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
func structKeys(input any) []string {
	m := structx.Marshal(input)
	if m == nil {
		return nil
	}
	return slices.Collect(mapStringKeys(m))
}

// mapStringKeys returns an iterator over the keys of a map[string]any.
func mapStringKeys(m map[string]any) iter.Seq[string] {
	return maps.Keys(m)
}

// toGenericReflect converts an arbitrary map to map[any]any via reflection.
func toGenericReflect(input any) (map[any]any, error) {
	rv := reflect.ValueOf(input)
	if rv.Kind() != reflect.Map {
		return nil, ErrInputNotMap
	}
	result := make(map[any]any, rv.Len())
	for _, key := range rv.MapKeys() {
		result[key.Interface()] = rv.MapIndex(key).Interface()
	}
	return result, nil
}
