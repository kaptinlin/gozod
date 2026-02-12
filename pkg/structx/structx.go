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

// ToMap converts a struct to map[string]any.
// It returns [ErrInvalidStructInput] if input is nil, a nil pointer,
// or not a struct.
func ToMap(input any) (map[string]any, error) {
	m := Marshal(input)
	if m == nil {
		return nil, ErrInvalidStructInput
	}
	return m, nil
}

// FromMap converts map[string]any to a struct of the given type.
// It returns [ErrTargetTypeMustBeStruct] if target is not a struct type.
func FromMap(data map[string]any, target reflect.Type) (any, error) {
	return Unmarshal(data, target)
}

// Marshal converts a struct to map[string]any using json tag field names.
// Returns nil if input is nil, a nil pointer, or not a struct.
func Marshal(input any) map[string]any {
	if input == nil {
		return nil
	}

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
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

		rv := reflect.ValueOf(val)
		dst := result.Field(i)

		if rv.Type().ConvertibleTo(f.Type) {
			dst.Set(rv.Convert(f.Type))
		} else if rv.Type().AssignableTo(f.Type) {
			dst.Set(rv)
		}
	}

	return result.Interface(), nil
}

// fieldName returns the map key for a struct field based on its json tag.
// Returns empty string for fields that should be skipped (json:"-").
func fieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "-" {
		return ""
	}
	if name == "" {
		return field.Name
	}
	return name
}
