package structx

import (
	"errors"
	"reflect"
	"strings"
)

// Sentinel errors for structx package.
var (
	ErrInvalidStructInput     = errors.New("input is not a struct or is nil")
	ErrTargetTypeMustBeStruct = errors.New("target type must be struct")
)

// =============================================================================
// STRUCT CONVERSION FUNCTIONS
// =============================================================================

// ToMap converts a struct to map[string]any with error handling
func ToMap(input any) (map[string]any, error) {
	result := Marshal(input)
	if result == nil {
		return nil, ErrInvalidStructInput
	}
	return result, nil
}

// FromMap converts map[string]any to struct using reflection
func FromMap(data map[string]any, target reflect.Type) (any, error) {
	return Unmarshal(data, target)
}

// Marshal converts a struct to map[string]any using reflection (no error return)
// Atomic operation - pure struct marshaling without external dependencies
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

	result := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := getFieldName(field)
		if fieldName == "-" {
			continue // Skip fields with json:"-"
		}

		fieldValue := v.Field(i).Interface()
		result[fieldName] = fieldValue
	}

	return result
}

// Unmarshal converts map[string]any to struct using reflection
// Atomic operation - pure struct unmarshaling without external dependencies
func Unmarshal(data map[string]any, structType reflect.Type) (any, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, ErrTargetTypeMustBeStruct
	}

	result := reflect.New(structType).Elem()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := getFieldName(field)
		if fieldName == "-" {
			continue // Skip fields with json:"-"
		}

		if value, exists := data[fieldName]; exists && value != nil {
			fieldValue := reflect.ValueOf(value)
			targetField := result.Field(i)

			if fieldValue.Type().ConvertibleTo(field.Type) {
				targetField.Set(fieldValue.Convert(field.Type))
			} else if fieldValue.Type().AssignableTo(field.Type) {
				targetField.Set(fieldValue)
			}
		}
	}

	return result.Interface(), nil
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// getFieldName extracts the field name considering json tags
func getFieldName(field reflect.StructField) string {
	// Check for json tag
	if tag := field.Tag.Get("json"); tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	return field.Name
}
