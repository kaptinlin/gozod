package utils

import (
	"fmt"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// ToErrorMap converts various error representations to ZodErrorMap
// Supports: string, ZodErrorMap, *ZodErrorMap, func(ZodRawIssue) string
func ToErrorMap(err any) (*core.ZodErrorMap, bool) {
	switch v := err.(type) {
	case string:
		errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string { return v })
		return &errorMap, true
	case core.ZodErrorMap:
		return &v, true
	case *core.ZodErrorMap:
		return v, true
	case func(core.ZodRawIssue) string:
		errorMap := core.ZodErrorMap(v)
		return &errorMap, true
	}
	return nil, false
}

// =============================================================================
// PARAMETER UTILITIES
// =============================================================================

// GetFirstParam extracts the first parameter from variadic arguments
// Provides convenience for Go's variadic parameter style while maintaining
// compatibility with Zod TypeScript v4's single parameter pattern
// Returns nil if no parameters provided
func GetFirstParam(params ...any) any {
	if len(params) == 0 {
		return nil
	}
	return params[0]
}

// IsPrimitiveType checks if a schema type is a primitive type that supports coercion
// Only primitive types should support coercion according to TypeScript Zod v4 alignment
func IsPrimitiveType(typeName core.ZodTypeCode) bool {
	switch typeName {
	case core.ZodTypeString, core.ZodTypeBool, core.ZodTypeTime:
		return true
	case core.ZodTypeInt, core.ZodTypeInt8, core.ZodTypeInt16, core.ZodTypeInt32, core.ZodTypeInt64:
		return true
	case core.ZodTypeUint, core.ZodTypeUint8, core.ZodTypeUint16, core.ZodTypeUint32, core.ZodTypeUint64:
		return true
	case core.ZodTypeFloat32, core.ZodTypeFloat64:
		return true
	case core.ZodTypeComplex64, core.ZodTypeComplex128:
		return true
	case core.ZodTypeBigInt:
		return true
	case core.ZodTypeNumber, core.ZodTypeNaN, core.ZodTypeInteger, core.ZodTypeDate, core.ZodTypeNil, core.ZodTypeAny, core.ZodTypeUnknown, core.ZodTypeNever, core.ZodTypeArray, core.ZodTypeSlice, core.ZodTypeObject, core.ZodTypeStruct, core.ZodTypeRecord, core.ZodTypeMap, core.ZodTypeUnion, core.ZodTypeDiscriminated, core.ZodTypeIntersection, core.ZodTypeStringBool, core.ZodTypeFunction, core.ZodTypeLazy, core.ZodTypeLiteral, core.ZodTypeEnum, core.ZodTypeOptional, core.ZodTypeNilable, core.ZodTypeDefault, core.ZodTypePrefault, core.ZodTypePipeline, core.ZodTypeTransform, core.ZodTypePipe, core.ZodTypeCustom, core.ZodTypeCheck, core.ZodTypeRefine, core.ZodTypeIPv4, core.ZodTypeIPv6, core.ZodTypeCIDRv4, core.ZodTypeCIDRv6, core.ZodTypeEmail, core.ZodTypeURL, core.ZodTypeIso, core.ZodTypeISODateTime, core.ZodTypeISODate, core.ZodTypeISOTime, core.ZodTypeISODuration, core.ZodTypeFile, core.ZodTypeFloat, core.ZodTypeUintptr, core.ZodTypeNonOptional:
		return false
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
		case core.ParsedTypeNumber, core.ParsedTypeBool, core.ParsedTypeFloat, core.ParsedTypeObject, core.ParsedTypeFunction, core.ParsedTypeFile, core.ParsedTypeDate, core.ParsedTypeArray, core.ParsedTypeSlice, core.ParsedTypeMap, core.ParsedTypeNaN, core.ParsedTypeNil, core.ParsedTypeComplex, core.ParsedTypeStruct, core.ParsedTypeEnum, core.ParsedTypeUnknown:
			return "unknown"
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
	case core.ParsedTypeString, core.ParsedTypeNumber, core.ParsedTypeBigint, core.ParsedTypeBool, core.ParsedTypeFloat, core.ParsedTypeObject, core.ParsedTypeFunction, core.ParsedTypeDate, core.ParsedTypeArray, core.ParsedTypeSlice, core.ParsedTypeMap, core.ParsedTypeNaN, core.ParsedTypeNil, core.ParsedTypeComplex, core.ParsedTypeStruct, core.ParsedTypeEnum, core.ParsedTypeUnknown:
		return "unknown"
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

// =============================================================================
// PARAMETER NORMALIZATION (moved from engine to avoid circular dependencies)
// =============================================================================

// NormalizeParams normalizes input parameters into a standard SchemaParams struct
// Supports variadic arguments where the first parameter is used:
// - nil -> empty SchemaParams
// - string -> { Error: string }
// - SchemaParams -> normalized copy
// - *SchemaParams -> normalized copy
func NormalizeParams(params ...any) *core.SchemaParams {
	if len(params) == 0 {
		return &core.SchemaParams{}
	}

	param := params[0]
	if param == nil {
		return &core.SchemaParams{}
	}

	switch v := param.(type) {
	case string:
		// String shorthand for error message
		return &core.SchemaParams{Error: v}

	case core.SchemaParams:
		// Copy to avoid mutation
		return &v

	case *core.SchemaParams:
		if v == nil {
			return &core.SchemaParams{}
		}
		// Copy to avoid mutation
		result := *v
		return &result

	default:
		// Unsupported types return empty params
		return &core.SchemaParams{}
	}
}

// NormalizeCustomParams normalizes input parameters into a standard CustomParams struct
// Supports variadic arguments where the first parameter is used:
// - nil -> empty CustomParams
// - string -> { Error: string }
// - CustomParams -> normalized copy
// - *CustomParams -> normalized copy
// - any -> { Error: any }
func NormalizeCustomParams(params ...any) *core.CustomParams {
	if len(params) == 0 {
		return &core.CustomParams{}
	}

	param := params[0]
	if param == nil {
		return &core.CustomParams{}
	}

	switch v := param.(type) {
	case string:
		// String shorthand for error message
		return &core.CustomParams{Error: v}

	case core.CustomParams:
		// Copy to avoid mutation
		return &v

	case *core.CustomParams:
		if v == nil {
			return &core.CustomParams{}
		}
		// Copy to avoid mutation
		result := *v
		return &result

	default:
		// For any other type, use it as error message
		return &core.CustomParams{Error: v}
	}
}

// ApplySchemaParams applies SchemaParams to a type definition
// Updates the definition with normalized parameters
func ApplySchemaParams(def *core.ZodTypeDef, params *core.SchemaParams) {
	if params == nil {
		return
	}

	// Apply error configuration
	if params.Error != nil {
		if err, ok := ToErrorMap(params.Error); ok {
			def.Error = err
		}
	}

	// Other parameters can be applied to def as needed
}

// =============================================================================
// PATH UTILITIES FOR ERROR FORMATTING
// =============================================================================

// ToDotPath converts an error path to dot notation string
// Compatible with TypeScript Zod v4 path formatting
func ToDotPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var parts []string
	for i, segment := range path {
		switch v := segment.(type) {
		case int:
			parts = append(parts, fmt.Sprintf("[%d]", v))
		case string:
			if i == 0 {
				parts = append(parts, v)
			} else {
				// Check if string contains special characters that need bracket notation
				if needsBracketNotation(v) {
					parts = append(parts, fmt.Sprintf(`["%s"]`, v))
				} else {
					parts = append(parts, "."+v)
				}
			}
		default:
			parts = append(parts, fmt.Sprintf("[%v]", v))
		}
	}

	return strings.Join(parts, "")
}

// needsBracketNotation checks if a string needs bracket notation in path formatting
func needsBracketNotation(s string) bool {
	// Check for spaces, hyphens, dots, or non-alphanumeric characters
	for _, char := range s {
		if char == ' ' || char == '-' || char == '.' || (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return true
		}
	}
	return false
}

// =============================================================================
// ERROR FORMATTING UTILITIES
// =============================================================================

// FormatErrorPath formats an error path for display
func FormatErrorPath(path []any, style string) string {
	switch style {
	case "dot":
		return ToDotPath(path)
	case "bracket":
		return formatBracketPath(path)
	default:
		return ToDotPath(path) // Default to dot notation
	}
}

// formatBracketPath formats path using bracket notation only
func formatBracketPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var parts []string
	for _, segment := range path {
		switch v := segment.(type) {
		case int:
			parts = append(parts, fmt.Sprintf("[%d]", v))
		case string:
			parts = append(parts, fmt.Sprintf(`["%s"]`, v))
		default:
			parts = append(parts, fmt.Sprintf("[%v]", v))
		}
	}

	return strings.Join(parts, "")
}

// =============================================================================
// VALIDATION CONTEXT UTILITIES
// =============================================================================

// CreateErrorContext creates error context for formatting
func CreateErrorContext(code core.IssueCode, path []any, input any) map[string]any {
	return map[string]any{
		"code":  code,
		"path":  path,
		"input": input,
	}
}

// ExtractErrorInfo extracts error information from a raw issue
func ExtractErrorInfo(issue core.ZodRawIssue) map[string]any {
	return map[string]any{
		"code":       issue.Code,
		"path":       issue.Path,
		"input":      issue.Input,
		"message":    issue.Message,
		"properties": issue.Properties,
	}
}
