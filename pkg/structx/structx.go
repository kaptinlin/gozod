package structx

import (
	"errors"
	"reflect"
	"strings"
)

// Sentinel errors for structx operations.
var (
	// ErrInvalidStructInput indicates the input is not a struct or is nil.
	ErrInvalidStructInput = errors.New("input is not a struct or is nil")
	// ErrTargetTypeMustBeStruct indicates the target type is not a struct.
	ErrTargetTypeMustBeStruct = errors.New("target type must be struct")
)

// ToMap converts a struct to map[string]any using json tag field names.
// It returns [ErrInvalidStructInput] if input is nil, a nil pointer,
// or not a struct type.
func ToMap(input any) (map[string]any, error) {
	v, ok := structValue(input)
	if !ok {
		return nil, ErrInvalidStructInput
	}

	t := v.Type()
	m := make(map[string]any, t.NumField())

	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := fieldName(f)
		if name == "" {
			continue
		}

		m[name] = v.Field(i).Interface()
	}

	return m, nil
}

// FromMap converts map[string]any to a struct of the given type.
// It returns [ErrTargetTypeMustBeStruct] if target is not a struct type.
// Fields are matched by json tag name, falling back to the Go field name.
func FromMap(data map[string]any, target reflect.Type) (any, error) {
	return Unmarshal(data, target)
}

// Marshal converts a struct to map[string]any using json tag field names.
// Marshal returns nil if input is nil, a nil pointer, or not a struct.
func Marshal(input any) map[string]any {
	m, err := ToMap(input)
	if err != nil {
		return nil
	}
	return m
}

// Unmarshal converts map[string]any to a struct of the given type.
// It returns [ErrTargetTypeMustBeStruct] if typ is not a struct type.
// Fields are matched by json tag name, falling back to the Go field name.
func Unmarshal(data map[string]any, typ reflect.Type) (any, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, ErrTargetTypeMustBeStruct
	}

	result := reflect.New(typ).Elem()

	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}

		name := fieldName(f)
		if name == "" {
			continue
		}

		val, ok := data[name]
		if !ok || val == nil {
			continue
		}

		setField(result.Field(i), reflect.ValueOf(val), f.Type)
	}

	return result.Interface(), nil
}

// structValue extracts the underlying struct reflect.Value from input.
// It dereferences pointers and returns false if input is nil, a nil pointer,
// or not a struct.
func structValue(input any) (reflect.Value, bool) {
	if input == nil {
		return reflect.Value{}, false
	}

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	return v, true
}

// setField assigns src to dst, converting types when possible.
func setField(dst reflect.Value, src reflect.Value, targetType reflect.Type) {
	switch {
	case src.Type().ConvertibleTo(targetType):
		dst.Set(src.Convert(targetType))
	case src.Type().AssignableTo(targetType):
		dst.Set(src)
	}
}

// fieldName returns the map key for a struct field based on its json tag.
// It returns an empty string for fields that should be skipped (json:"-").
func fieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")

	switch name {
	case "-":
		return ""
	case "":
		return field.Name
	default:
		return name
	}
}
