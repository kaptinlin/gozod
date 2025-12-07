package reflectx

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/kaptinlin/gozod/core"
)

// =============================================================================
// TYPE CHECKING
// =============================================================================

// IsNil checks if a value is nil (including typed nil)
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	//nolint:exhaustive // default handles all other cases
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

// IsZero checks if a value is the zero value for its type
func IsZero(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.IsZero()
}

// IsValid checks if a reflect.Value is valid
func IsValid(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	return rv.IsValid()
}

// IsBool checks if a value is a bool type
func IsBool(v any) bool {
	if v == nil {
		return false
	}
	_, ok := v.(bool)
	return ok
}

// IsString checks if a value is a string type
func IsString(v any) bool {
	if v == nil {
		return false
	}
	_, ok := v.(string)
	return ok
}

// IsInt checks if a value is any integer type
func IsInt(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return true
	default:
		return false
	}
}

// IsUint checks if a value is any unsigned integer type
func IsUint(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return true
	default:
		return false
	}
}

// IsFloat checks if a value is any float type
func IsFloat(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

// IsComplex checks if a value is any complex type
func IsComplex(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case complex64, complex128:
		return true
	default:
		return false
	}
}

// IsBigInt checks if a value is a big.Int type
func IsBigInt(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case *big.Int, big.Int:
		return true
	default:
		return false
	}
}

// IsNumber checks if a value is any numeric type (alias for IsNumeric)
func IsNumber(v any) bool {
	return IsNumeric(v)
}

// IsNumeric checks if a value is any numeric type
func IsNumeric(v any) bool {
	return IsInt(v) || IsUint(v) || IsFloat(v) || IsComplex(v) || IsBigInt(v)
}

// IsArray checks if a value is an array
func IsArray(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Array
}

// IsSlice checks if a value is a slice
func IsSlice(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

// IsMap checks if a value is a map
func IsMap(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Map
}

// IsStruct checks if a value is a struct
func IsStruct(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Struct
}

// IsInterface checks if a value is an interface
func IsInterface(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Interface
}

// IsFunc checks if a value is a function
func IsFunc(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Func
}

// IsChan checks if a value is a channel
func IsChan(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Chan
}

// IsPointer checks if a value is a pointer
func IsPointer(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Ptr
}

// IsError checks if a value implements the error interface
func IsError(v any) bool {
	if v == nil {
		return false
	}
	_, ok := v.(error)
	return ok
}

// IsStringer checks if a value implements the fmt.Stringer interface
func IsStringer(v any) bool {
	if v == nil {
		return false
	}
	_, ok := v.(fmt.Stringer)
	return ok
}

// IsComparable checks if a value's type is comparable
func IsComparable(v any) bool {
	if v == nil {
		return true
	}
	return reflect.TypeOf(v).Comparable()
}

// IsIterable checks if a value can be iterated (array, slice, map, string)
func IsIterable(v any) bool {
	if v == nil {
		return false
	}
	//nolint:exhaustive // default handles all other cases
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return true
	default:
		return false
	}
}

// =============================================================================
// TYPE PARSING
// =============================================================================

// ParsedType returns the parsed type for runtime type detection
// This corresponds to Zod v4's getParsedType() function which returns ParsedTypes
// See: .reference/zod/packages/zod/src/v4/core/util.ts:412
func ParsedType(v any) core.ParsedType {
	if v == nil {
		return core.ParsedTypeNil
	}

	// Handle big.Int types first (before general type checking)
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
		case reflect.Ptr:
			if IsNil(v) {
				return core.ParsedTypeNil
			}
			return ParsedType(reflect.ValueOf(v).Elem().Interface())
		case reflect.Interface:
			if IsNil(v) {
				return core.ParsedTypeNil
			}
			return ParsedType(reflect.ValueOf(v).Elem().Interface())
		default:
			return core.ParsedTypeUnknown
		}
	}
}

// ParsedCategory returns the general category of a type
func ParsedCategory(v any) string {
	if v == nil {
		return "null"
	}

	if IsNumeric(v) {
		return "number"
	}
	if IsString(v) {
		return "string"
	}
	if IsBool(v) {
		return "boolean"
	}
	if IsArray(v) || IsSlice(v) {
		return "array"
	}
	if IsMap(v) || IsStruct(v) {
		return "object"
	}

	return "unknown"
}

// ParsedKind returns the reflect.Kind of a value
func ParsedKind(v any) reflect.Kind {
	if v == nil {
		return reflect.Invalid
	}
	return reflect.TypeOf(v).Kind()
}
