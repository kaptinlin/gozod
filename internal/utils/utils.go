package utils

import (
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// IsPrimitiveType checks if a schema type is a primitive type that supports coercion
// Only primitive types should support coercion according to TypeScript Zod v4 alignment
func IsPrimitiveType(typeName string) bool {
	switch typeName {
	case "string", "bool", "boolean":
		return true
	case "int", "int8", "int16", "int32", "int64":
		return true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "float32", "float64", "number":
		return true
	case "complex64", "complex128":
		return true
	case "bigint":
		return true
	default:
		return false
	}
}

// =============================================================================
// ORIGIN TYPE FUNCTIONS FOR ERROR MESSAGES
// =============================================================================

// GetOriginFromValue smartly determines the origin of a value (general purpose)
func GetOriginFromValue(value any) string {
	if reflectx.IsNumeric(value) {
		return "number"
	}
	if reflectx.IsString(value) {
		return "string"
	}
	if reflectx.HasLength(value) {
		return "array"
	}
	if reflectx.IsMap(value) {
		return "object"
	}
	return "unknown"
}

// GetNumericOrigin determines the numeric origin type for error messages
func GetNumericOrigin(value any) string {
	if value == nil {
		return "nil"
	}

	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	default:
		// Check for big.Int and string through reflection
		parsedType := reflectx.ParsedType(value)
		switch parsedType {
		case core.ParsedTypeBigint:
			return "bigint"
		case core.ParsedTypeString:
			return "string"
		default:
			return "unknown"
		}
	}
}

// GetSizableOrigin determines the origin type for sizable values
func GetSizableOrigin(value any) string {
	if value == nil {
		return "nil"
	}

	// Use reflectx for type categorization
	if reflectx.IsString(value) {
		return "string"
	}
	if reflectx.IsSlice(value) {
		return "slice"
	}
	if reflectx.IsArray(value) {
		return "array"
	}
	if reflectx.IsMap(value) {
		return "map"
	}
	if reflectx.IsStruct(value) {
		return "struct"
	}

	// Check parsed type for other cases
	parsedType := reflectx.ParsedType(value)
	switch parsedType {
	case core.ParsedTypeFile:
		return "file"
	default:
		return "unknown"
	}
}

// GetLengthableOrigin returns the origin type for lengthable values
func GetLengthableOrigin(value any) string {
	if value == nil {
		return "nil"
	}

	// Use reflectx for type checking
	if reflectx.IsString(value) {
		return "string"
	}
	if reflectx.IsSlice(value) {
		return "slice"
	}
	if reflectx.IsArray(value) {
		return "array"
	}

	return "unknown"
}

// =============================================================================
// STRING OPERATIONS
// =============================================================================

// EscapeRegex escapes special characters in a string for use in regex
func EscapeRegex(str string) string {
	// Characters that need escaping in regex
	specialChars := []string{"\\", "^", "$", ".", "[", "]", "|", "(", ")", "?", "*", "+", "{", "}"}

	result := str
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}

// =============================================================================
// COMPARISON FUNCTIONS
// =============================================================================

// CompareValues compares two values, returns -1, 0, or 1
// Automatically dereferences pointers before comparison
func CompareValues(a, b any) int {
	// Dereference pointers using reflectx
	derefA := reflectx.DerefAll(a)
	derefB := reflectx.DerefAll(b)

	// Handle nil cases
	if derefA == nil && derefB == nil {
		return 0
	}
	if derefA == nil {
		return -1
	}
	if derefB == nil {
		return 1
	}

	switch va := derefA.(type) {
	case int:
		if vb, ok := derefB.(int); ok {
			if va < vb {
				return -1
			}
			if va > vb {
				return 1
			}
			return 0
		}
	case int64:
		if vb, ok := derefB.(int64); ok {
			if va < vb {
				return -1
			}
			if va > vb {
				return 1
			}
			return 0
		}
	case float64:
		if vb, ok := derefB.(float64); ok {
			if va < vb {
				return -1
			}
			if va > vb {
				return 1
			}
			return 0
		}
	case float32:
		if vb, ok := derefB.(float32); ok {
			if va < vb {
				return -1
			}
			if va > vb {
				return 1
			}
			return 0
		}
	case string:
		if vb, ok := derefB.(string); ok {
			if va < vb {
				return -1
			}
			if va > vb {
				return 1
			}
			return 0
		}
	}

	// Equal or incomparable
	return 0
}
