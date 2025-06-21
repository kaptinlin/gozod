package structx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// =============================================================================
// STRUCT TYPE CHECKING
// =============================================================================

// Is checks if the value is a struct type
func Is(v any) bool {
	return reflectx.IsStruct(v)
}

// IsPointer checks if the value is a pointer to a struct
func IsPointer(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct
}

// =============================================================================
// STRUCT CONVERSION FUNCTIONS
// =============================================================================

// ToMap converts a struct to map[string]any with error handling
func ToMap(input any) (map[string]any, error) {
	result := Marshal(input)
	if result == nil {
		return nil, errors.New("input is not a struct or is nil")
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
		return nil, errors.New("target type must be struct")
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
// STRUCT FIELD OPERATIONS
// =============================================================================

// Fields returns all field names of a struct
func Fields(input any) []string {
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
	fields := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := getFieldName(field)
		if fieldName != "-" {
			fields = append(fields, fieldName)
		}
	}

	return fields
}

// GetField gets a field value from a struct by field name
func GetField(input any, fieldName string) (any, bool) {
	if input == nil {
		return nil, false
	}

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, false
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if getFieldName(field) == fieldName {
			return v.Field(i).Interface(), true
		}
	}

	return nil, false
}

// SetField sets a field value in a struct by field name (requires pointer)
func SetField(input any, fieldName string, value any) error {
	if input == nil {
		return errors.New("input is nil")
	}

	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Ptr {
		return errors.New("input must be a pointer to struct")
	}

	if v.IsNil() {
		return errors.New("input pointer is nil")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("input must be a pointer to struct")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if getFieldName(field) == fieldName {
			fieldValue := reflect.ValueOf(value)
			targetField := v.Field(i)

			if fieldValue.Type().ConvertibleTo(field.Type) {
				targetField.Set(fieldValue.Convert(field.Type))
				return nil
			} else if fieldValue.Type().AssignableTo(field.Type) {
				targetField.Set(fieldValue)
				return nil
			}
			return errors.New("value type not compatible with field type")
		}
	}

	return errors.New("field not found")
}

// HasField checks if a struct has a specific field
func HasField(input any, fieldName string) bool {
	_, exists := GetField(input, fieldName)
	return exists
}

// =============================================================================
// STRUCT EXTRACTION FUNCTIONS
// =============================================================================

// Extract extracts a struct and converts it to map[string]any
// Returns the extracted map and whether the extraction was successful
func Extract(input any) (map[string]any, bool) {
	result := Marshal(input)
	return result, result != nil
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

// Count returns the number of exported fields in a struct
func Count(input any) int {
	fields := Fields(input)
	if fields == nil {
		return 0
	}
	return len(fields)
}

// Clone creates a deep copy of a struct through map conversion
func Clone(input any) (any, error) {
	if !Is(input) {
		return nil, errors.New("input is not a struct")
	}

	// Convert to map
	m := Marshal(input)
	if m == nil {
		return nil, errors.New("failed to marshal struct")
	}

	// Get the type
	t := reflect.TypeOf(input)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Convert back to struct
	return Unmarshal(m, t)
}
