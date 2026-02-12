package reflectx

import (
	"math/big"
	"reflect"

	"github.com/kaptinlin/gozod/core"
)

// IsNil reports whether v is nil, including typed nils (nil pointers,
// nil slices, nil maps, nil channels, nil functions, and nil interfaces).
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	//nolint:exhaustive // default handles all other cases
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

// IsBool reports whether v is a bool.
func IsBool(v any) bool {
	_, ok := v.(bool)
	return ok
}

// IsString reports whether v is a string.
func IsString(v any) bool {
	_, ok := v.(string)
	return ok
}

// IsNumeric reports whether v is any numeric type
// (signed/unsigned integer, float, complex, or [big.Int]).
func IsNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64,
		complex64, complex128,
		*big.Int, big.Int:
		return true
	default:
		return false
	}
}

// IsArray reports whether v is an array.
func IsArray(v any) bool {
	return v != nil && reflect.TypeOf(v).Kind() == reflect.Array
}

// IsSlice reports whether v is a slice.
func IsSlice(v any) bool {
	return v != nil && reflect.TypeOf(v).Kind() == reflect.Slice
}

// IsMap reports whether v is a map.
func IsMap(v any) bool {
	return v != nil && reflect.TypeOf(v).Kind() == reflect.Map
}

// IsStruct reports whether v is a struct.
func IsStruct(v any) bool {
	return v != nil && reflect.TypeOf(v).Kind() == reflect.Struct
}

// ParsedType returns the [core.ParsedType] for runtime type detection.
// It corresponds to Zod v4's getParsedType() function.
// Pointers and interfaces are dereferenced recursively; nil yields
// [core.ParsedTypeNil].
//
// See: .reference/zod/packages/zod/src/v4/core/util.ts:412
func ParsedType(v any) core.ParsedType {
	if v == nil {
		return core.ParsedTypeNil
	}

	// Handle *big.Int before the general type switch.
	if _, ok := v.(*big.Int); ok {
		return core.ParsedTypeBigint
	}

	switch v.(type) {
	case bool:
		return core.ParsedTypeBool
	case string:
		return core.ParsedTypeString
	case int, int8, int16, int32, int64:
		return core.ParsedTypeNumber
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return core.ParsedTypeNumber
	case float32, float64:
		return core.ParsedTypeFloat
	case complex64, complex128:
		return core.ParsedTypeComplex
	default:
		return parsedTypeByKind(v)
	}
}

// parsedTypeByKind resolves composite and pointer kinds via reflection.
func parsedTypeByKind(v any) core.ParsedType {
	//nolint:exhaustive // default handles all other cases
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
		return core.ParsedTypeArray
	case reflect.Slice:
		return core.ParsedTypeSlice
	case reflect.Map:
		return core.ParsedTypeMap
	case reflect.Struct:
		return core.ParsedTypeStruct
	case reflect.Func:
		return core.ParsedTypeFunction
	case reflect.Pointer, reflect.Interface:
		if IsNil(v) {
			return core.ParsedTypeNil
		}
		return ParsedType(reflect.ValueOf(v).Elem().Interface())
	default:
		return core.ParsedTypeUnknown
	}
}

// ParsedCategory returns a broad human-readable category for v:
// "nil", "bool", "string", "number", "array", "object", or "unknown".
func ParsedCategory(v any) string {
	return parsedTypeCategory(ParsedType(v))
}

// parsedTypeCategory maps a [core.ParsedType] to its broad category string.
func parsedTypeCategory(pt core.ParsedType) string {
	//nolint:exhaustive // default handles all other cases
	switch pt {
	case core.ParsedTypeNil:
		return "nil"
	case core.ParsedTypeBool:
		return "bool"
	case core.ParsedTypeString:
		return "string"
	case core.ParsedTypeNumber, core.ParsedTypeFloat, core.ParsedTypeComplex, core.ParsedTypeBigint:
		return "number"
	case core.ParsedTypeArray, core.ParsedTypeSlice:
		return "array"
	case core.ParsedTypeMap, core.ParsedTypeStruct:
		return "object"
	default:
		return "unknown"
	}
}
