package structx

import (
	"errors"
	"reflect"
	"strings"
)

// Sentinel errors for structx operations.
var (
	// ErrInvalidStructInput is returned when the input is not a struct or is nil.
	ErrInvalidStructInput = errors.New("input is not a struct or is nil")
	// ErrTargetTypeMustBeStruct is returned when the target type is not a struct.
	ErrTargetTypeMustBeStruct = errors.New("target type must be struct")
)

// ToMap converts a struct to map[string]any.
// It returns ErrInvalidStructInput if input is nil, a nil pointer, or not a struct.
func ToMap(input any) (map[string]any, error) {
	result := Marshal(input)
	if result == nil {
		return nil, ErrInvalidStructInput
	}
	return result, nil
}

// FromMap converts map[string]any to a struct of the given type.
// It returns ErrTargetTypeMustBeStruct if target is not a struct type.
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
	result := make(map[string]any, t.NumField())

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := fieldName(field)
		if name == "" {
			continue
		}

		result[name] = v.Field(i).Interface()
	}

	return result
}

// Unmarshal converts map[string]any to a struct of the given type.
// It returns ErrTargetTypeMustBeStruct if structType is not a struct type.
// Fields are matched by json tag name, falling back to the Go field name.
func Unmarshal(data map[string]any, structType reflect.Type) (any, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, ErrTargetTypeMustBeStruct
	}

	result := reflect.New(structType).Elem()

	for i := range structType.NumField() {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		name := fieldName(field)
		if name == "" {
			continue
		}

		value, exists := data[name]
		if !exists || value == nil {
			continue
		}

		fieldValue := reflect.ValueOf(value)
		targetField := result.Field(i)

		if fieldValue.Type().ConvertibleTo(field.Type) {
			targetField.Set(fieldValue.Convert(field.Type))
		} else if fieldValue.Type().AssignableTo(field.Type) {
			targetField.Set(fieldValue)
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
