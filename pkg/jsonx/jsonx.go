package jsonx

import (
	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/gozod/pkg/coerce"
)

// IsValid performs comprehensive JSON validation using go-json-experiment
func IsValid(value any) bool {
	// Convert value to string using coerce
	str, err := coerce.ToString(value)
	if err != nil {
		return false
	}

	return isValidJSON(str)
}

// IsValidString performs JSON string validation specifically for string inputs
func IsValidString(s string) bool {
	return isValidJSON(s)
}

// IsPrimitive checks if the JSON represents a primitive value (string, number, boolean, null)
func IsPrimitive(value any) bool {
	// Convert value to string using coerce
	str, err := coerce.ToString(value)
	if err != nil {
		return false
	}

	return isPrimitiveJSON(str)
}

// IsNumber checks if the JSON represents a number
func IsNumber(value any) bool {
	// Convert value to string using coerce
	str, err := coerce.ToString(value)
	if err != nil {
		return false
	}

	return isNumberJSON(str)
}

// isValidJSON internal helper for JSON validation
func isValidJSON(s string) bool {
	if s == "" {
		return false
	}

	var v any
	return json.Unmarshal([]byte(s), &v) == nil
}

// isPrimitiveJSON checks if string contains primitive JSON
func isPrimitiveJSON(s string) bool {
	if !isValidJSON(s) {
		return false
	}

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return false
	}

	// Check if it's a primitive type (not object or array)
	switch v.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

// isNumberJSON checks if string contains a JSON number
func isNumberJSON(s string) bool {
	if !isValidJSON(s) {
		return false
	}

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return false
	}

	// Check if it's a number type
	switch v.(type) {
	case float64, int, int64:
		return true
	default:
		return false
	}
}
