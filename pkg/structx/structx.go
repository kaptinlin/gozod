package structx

import (
	"errors"
	"reflect"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

// Sentinel errors for structx operations.
var (
	// ErrInvalidStructInput indicates the input is not a struct or is nil.
	ErrInvalidStructInput = errors.New("input is not a struct or is nil")
	// ErrTargetTypeMustBeStruct indicates the target type is not a struct.
	ErrTargetTypeMustBeStruct = errors.New("target type must be struct")
)

// ToMap converts a struct to map[string]any using the named field-name tag
// (e.g. "json", "yaml", "toml") for field names. It returns
// [ErrInvalidStructInput] if input is nil, a nil pointer, or not a struct type.
func ToMap(fieldNameTag string, input any) (map[string]any, error) {
	v, ok := structValue(input)
	if !ok {
		return nil, ErrInvalidStructInput
	}

	m := make(map[string]any, v.NumField())

	for f, value := range v.Fields() {
		if !f.IsExported() || !value.CanInterface() {
			continue
		}

		fieldKey := tagparser.FieldName(fieldNameTag, f)
		if fieldKey.Skip {
			continue
		}

		m[fieldKey.Name] = value.Interface()
	}

	return m, nil
}

// FromMap converts map[string]any to a struct of the given type using the named
// field-name tag. It returns [ErrTargetTypeMustBeStruct] if target is not a
// struct type. Fields are matched by tag name, falling back to the Go field name.
func FromMap(fieldNameTag string, data map[string]any, target reflect.Type) (any, error) {
	return Unmarshal(fieldNameTag, data, target)
}

// Marshal converts a struct to map[string]any using the named field-name tag for
// field names. Marshal returns nil if input is nil, a nil pointer, or not a struct.
func Marshal(fieldNameTag string, input any) map[string]any {
	m, err := ToMap(fieldNameTag, input)
	if err != nil {
		return nil
	}
	return m
}

// Unmarshal converts map[string]any to a struct of the given type using the
// named field-name tag. It returns [ErrTargetTypeMustBeStruct] if typ is not a
// struct type. Fields are matched by tag name, falling back to the Go field name.
func Unmarshal(fieldNameTag string, data map[string]any, typ reflect.Type) (any, error) {
	if typ.Kind() == reflect.Pointer {
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

		name := fieldName(fieldNameTag, f)
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
	if v.Kind() == reflect.Pointer {
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
// AssignableTo is checked first as the cheaper, more common path.
func setField(dst reflect.Value, src reflect.Value, targetType reflect.Type) {
	switch {
	case src.Type().AssignableTo(targetType):
		dst.Set(src)
	case src.Type().ConvertibleTo(targetType):
		dst.Set(src.Convert(targetType))
	}
}

// fieldName returns the map key for a struct field based on the named
// field-name tag. It returns an empty string for fields that should be skipped.
func fieldName(fieldNameTag string, field reflect.StructField) string {
	fieldKey := tagparser.FieldName(fieldNameTag, field)
	if fieldKey.Skip {
		return ""
	}
	return fieldKey.Name
}
