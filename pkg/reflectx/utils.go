package reflectx

import (
	"fmt"
	"reflect"
)

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// ValueOf safely gets reflect.Value
func ValueOf(v any) reflect.Value {
	if v == nil {
		return reflect.Value{}
	}
	return reflect.ValueOf(v)
}

// TypeOf safely gets reflect.Type
func TypeOf(v any) reflect.Type {
	if v == nil {
		return nil
	}
	return reflect.TypeOf(v)
}

// KindOf gets reflect.Kind
func KindOf(v any) reflect.Kind {
	if v == nil {
		return reflect.Invalid
	}
	return reflect.TypeOf(v).Kind()
}

// =============================================================================
// COMPARISON OPERATIONS
// =============================================================================

// Equal performs deep equality comparison
func Equal(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// Compare compares two values (-1, 0, 1)
func Compare(a, b any) (int, error) {
	if a == nil && b == nil {
		return 0, nil
	}
	if a == nil {
		return -1, nil
	}
	if b == nil {
		return 1, nil
	}

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Type() != vb.Type() {
		return 0, fmt.Errorf("cannot compare different types: %v vs %v", va.Type(), vb.Type())
	}

	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		aVal := va.Int()
		bVal := vb.Int()
		if aVal < bVal {
			return -1, nil
		} else if aVal > bVal {
			return 1, nil
		}
		return 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		aVal := va.Uint()
		bVal := vb.Uint()
		if aVal < bVal {
			return -1, nil
		} else if aVal > bVal {
			return 1, nil
		}
		return 0, nil
	case reflect.Float32, reflect.Float64:
		aVal := va.Float()
		bVal := vb.Float()
		if aVal < bVal {
			return -1, nil
		} else if aVal > bVal {
			return 1, nil
		}
		return 0, nil
	case reflect.String:
		aVal := va.String()
		bVal := vb.String()
		if aVal < bVal {
			return -1, nil
		} else if aVal > bVal {
			return 1, nil
		}
		return 0, nil
	case reflect.Bool:
		aVal := va.Bool()
		bVal := vb.Bool()
		if !aVal && bVal {
			return -1, nil
		} else if aVal && !bVal {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot compare values of type %v", va.Type())
	}
}

// =============================================================================
// ZERO VALUE OPERATIONS
// =============================================================================

// Zero creates a zero value of the given type
func Zero(t reflect.Type) any {
	if t == nil {
		return nil
	}
	return reflect.Zero(t).Interface()
}

// ZeroOf creates a zero value of the generic type
func ZeroOf[T any]() T {
	var zero T
	return zero
}

// IsZeroValue checks if a value is the zero value
func IsZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.IsZero()
}

// =============================================================================
// TYPE INFORMATION
// =============================================================================

// TypeName gets the name of a type
func TypeName(v any) string {
	if v == nil {
		return ""
	}
	t := reflect.TypeOf(v)
	return t.Name()
}

// PackagePath gets the package path of a type
func PackagePath(v any) string {
	if v == nil {
		return ""
	}
	t := reflect.TypeOf(v)
	return t.PkgPath()
}

// FullTypeName gets the full type name including package
func FullTypeName(v any) string {
	if v == nil {
		return ""
	}
	t := reflect.TypeOf(v)
	if t.PkgPath() == "" {
		return t.Name()
	}
	return t.PkgPath() + "." + t.Name()
}

// =============================================================================
// METHOD OPERATIONS
// =============================================================================

// HasMethod checks if a value has a method with the given name
func HasMethod(v any, name string) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	_, ok := t.MethodByName(name)
	return ok
}

// CallMethod calls a method on a value
func CallMethod(v any, name string, args ...any) ([]any, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot call method on nil value")
	}

	rv := reflect.ValueOf(v)
	method := rv.MethodByName(name)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", name)
	}

	// Convert args to reflect.Value
	argVals := make([]reflect.Value, len(args))
	for i, arg := range args {
		argVals[i] = reflect.ValueOf(arg)
	}

	// Call the method
	results := method.Call(argVals)

	// Convert results back to any
	resultVals := make([]any, len(results))
	for i, result := range results {
		resultVals[i] = result.Interface()
	}

	return resultVals, nil
}

// GetMethod gets a method by name
func GetMethod(v any, name string) (reflect.Method, bool) {
	if v == nil {
		return reflect.Method{}, false
	}
	t := reflect.TypeOf(v)
	return t.MethodByName(name)
}

// =============================================================================
// SAFE TYPE ASSERTIONS AND CONVERSIONS
// =============================================================================

// SafeConvert performs safe generic type assertion
func SafeConvert[T any](v any) (result T, ok bool) {
	if v == nil {
		return result, false
	}
	result, ok = v.(T)
	return result, ok
}

// ConvertTo converts a value to a specific reflect.Type
func ConvertTo(v any, target reflect.Type) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot convert nil value")
	}
	if target == nil {
		return nil, fmt.Errorf("target type cannot be nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Type() == target {
		return v, nil
	}

	if rv.Type().ConvertibleTo(target) {
		converted := rv.Convert(target)
		return converted.Interface(), nil
	}

	return nil, fmt.Errorf("cannot convert %v to %v", rv.Type(), target)
}

// ToReflectValue converts a value to reflect.Value
func ToReflectValue(v any) reflect.Value {
	return reflect.ValueOf(v)
}

// FromReflectValue gets the any value from reflect.Value
func FromReflectValue(rv reflect.Value) any {
	if !rv.IsValid() {
		return nil
	}
	return rv.Interface()
}

// ToType gets the reflect.Type of a generic type
func ToType[T any]() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}

// =============================================================================
// REFLECTION UTILITIES
// =============================================================================

// CanSet checks if a reflect.Value can be set
func CanSet(rv reflect.Value) bool {
	return rv.IsValid() && rv.CanSet()
}

// CanAddr checks if a reflect.Value can be addressed
func CanAddr(rv reflect.Value) bool {
	return rv.IsValid() && rv.CanAddr()
}

// CanInterface checks if a reflect.Value can be converted to any
func CanInterface(rv reflect.Value) bool {
	return rv.IsValid() && rv.CanInterface()
}

// IsExported checks if a struct field is exported
func IsExported(field reflect.StructField) bool {
	return field.IsExported()
}

// GetTag gets a struct field tag
func GetTag(field reflect.StructField, key string) string {
	return field.Tag.Get(key)
}

// =============================================================================
// STRING CONVERSION
// =============================================================================

// ToString converts any value to string using reflection fallback
// This handles types not covered by fast-path conversions in coerce package
func ToString(v any) (string, error) {
	if v == nil {
		return "", fmt.Errorf("cannot convert nil to string")
	}

	rv := reflect.ValueOf(v)

	// Handle pointers by dereferencing (additional safety check)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return "", fmt.Errorf("cannot convert nil pointer to string")
		}
		rv = rv.Elem()
	}

	// Try to convert using fmt.Sprintf as fallback
	str := fmt.Sprintf("%v", rv.Interface())
	return str, nil
}

// =============================================================================
// GENERIC CONVERSION
// =============================================================================

// ConvertToGeneric is a generic conversion function using reflection
// This delegates to appropriate specialized functions or uses reflection
func ConvertToGeneric[T any](v any) (T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)

	if v == nil {
		return zero, fmt.Errorf("cannot convert nil value")
	}

	sourceValue := reflect.ValueOf(v)
	sourceType := sourceValue.Type()

	// Direct assignability check
	if sourceType.AssignableTo(targetType) {
		return v.(T), nil
	}

	// Convertibility check
	if sourceType.ConvertibleTo(targetType) {
		converted := sourceValue.Convert(targetType)
		return converted.Interface().(T), nil
	}

	// Handle specific numeric conversions through intermediate types
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		// For smaller integer types, use bounds checking
		if sourceType.Kind() >= reflect.Int && sourceType.Kind() <= reflect.Uint64 {
			converted := sourceValue.Convert(targetType)
			return converted.Interface().(T), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// For unsigned types
		if sourceType.Kind() >= reflect.Int && sourceType.Kind() <= reflect.Uint64 {
			converted := sourceValue.Convert(targetType)
			return converted.Interface().(T), nil
		}
	case reflect.Float32:
		// Convert through float64
		if sourceType.Kind() >= reflect.Int && sourceType.Kind() <= reflect.Float64 {
			converted := sourceValue.Convert(targetType)
			return converted.Interface().(T), nil
		}
	case reflect.Complex64, reflect.Complex128:
		// Try converting real numbers to complex
		if sourceType.Kind() >= reflect.Int && sourceType.Kind() <= reflect.Float64 {
			if targetType.Kind() == reflect.Complex64 {
				real := sourceValue.Convert(reflect.TypeOf(float32(0))).Float()
				result := complex(float32(real), 0)
				return any(result).(T), nil
			} else {
				real := sourceValue.Convert(reflect.TypeOf(float64(0))).Float()
				result := complex(real, 0)
				return any(result).(T), nil
			}
		}
	}

	return zero, fmt.Errorf("unsupported conversion from %T to %v", v, targetType)
}
