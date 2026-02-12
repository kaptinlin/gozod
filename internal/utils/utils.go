package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// ToErrorMap converts various error representations to ZodErrorMap.
// Supports: string, ZodErrorMap, *ZodErrorMap, func(ZodRawIssue) string.
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

// FirstParam extracts the first parameter from variadic arguments.
// Returns nil if no parameters provided.
func FirstParam(params ...any) any {
	if len(params) == 0 {
		return nil
	}
	return params[0]
}

// OriginFromValue determines the origin of a value for error messages.
func OriginFromValue(value any) string {
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

// NumericOrigin determines the numeric origin type for error messages.
func NumericOrigin(value any) string {
	if value == nil {
		return "nil"
	}

	switch value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	default:
		parsedType := reflectx.ParsedType(value)
		switch parsedType { //nolint:exhaustive // only bigint and string need special handling
		case core.ParsedTypeBigint:
			return "bigint"
		case core.ParsedTypeString:
			return "string"
		default:
			return "unknown"
		}
	}
}

// SizableOrigin determines the origin type for sizable values.
func SizableOrigin(value any) string {
	if value == nil {
		return "nil"
	}
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

	parsedType := reflectx.ParsedType(value)
	switch parsedType { //nolint:exhaustive // only file type needs special handling
	case core.ParsedTypeFile:
		return "file"
	default:
		return "unknown"
	}
}

// LengthableOrigin returns the origin type for lengthable values.
func LengthableOrigin(value any) string {
	if value == nil {
		return "nil"
	}
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

// CompareValues compares two values, returns -1, 0, or 1.
// Automatically dereferences pointers before comparison.
func CompareValues(a, b any) int {
	derefA := reflectx.DerefAll(a)
	derefB := reflectx.DerefAll(b)

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
			return cmpOrdered(va, vb)
		}
	case int64:
		if vb, ok := derefB.(int64); ok {
			return cmpOrdered(va, vb)
		}
	case float64:
		if vb, ok := derefB.(float64); ok {
			return cmpOrdered(va, vb)
		}
	case float32:
		if vb, ok := derefB.(float32); ok {
			return cmpOrdered(va, vb)
		}
	case string:
		if vb, ok := derefB.(string); ok {
			return cmpOrdered(va, vb)
		}
	}

	return 0
}

// cmpOrdered returns -1, 0, or 1 for ordered types.
func cmpOrdered[T ~int | ~int64 | ~float64 | ~float32 | ~string](a, b T) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// NormalizeParams normalizes input parameters into a SchemaParams struct.
// Supports: nil, string, SchemaParams, *SchemaParams.
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
		return &core.SchemaParams{Error: v}
	case core.SchemaParams:
		return &v
	case *core.SchemaParams:
		if v == nil {
			return &core.SchemaParams{}
		}
		result := *v
		return &result
	default:
		return &core.SchemaParams{}
	}
}

// NormalizeCustomParams normalizes input parameters into a CustomParams struct.
// Supports: nil, string, CustomParams, *CustomParams, any (as error).
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
		return &core.CustomParams{Error: v}
	case core.CustomParams:
		return &v
	case *core.CustomParams:
		if v == nil {
			return &core.CustomParams{}
		}
		result := *v
		return &result
	default:
		return &core.CustomParams{Error: v}
	}
}

// ApplySchemaParams applies SchemaParams to a type definition.
func ApplySchemaParams(def *core.ZodTypeDef, params *core.SchemaParams) {
	if params == nil {
		return
	}
	if params.Error != nil {
		if err, ok := ToErrorMap(params.Error); ok {
			def.Error = err
		}
	}
}

// ToDotPath converts an error path to dot notation string.
// Compatible with TypeScript Zod v4 path formatting.
func ToDotPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var result strings.Builder
	// Estimate capacity: average ~8 chars per segment
	result.Grow(len(path) * 8)

	for i, segment := range path {
		switch v := segment.(type) {
		case int:
			result.WriteByte('[')
			result.WriteString(strconv.Itoa(v))
			result.WriteByte(']')
		case string:
			switch {
			case i == 0:
				result.WriteString(v)
			case needsBracketNotation(v):
				result.WriteString(`["`)
				result.WriteString(v)
				result.WriteString(`"]`)
			default:
				result.WriteByte('.')
				result.WriteString(v)
			}
		default:
			fmt.Fprintf(&result, "[%v]", v)
		}
	}

	return result.String()
}

// needsBracketNotation checks if a string key needs bracket notation.
func needsBracketNotation(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] >= '0' && s[0] <= '9' {
		return true
	}
	for _, c := range s {
		if !isIdentChar(c) {
			return true
		}
	}
	return false
}

// isIdentChar reports whether c is a valid identifier character.
func isIdentChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}

// FormatErrorPath formats an error path for display.
func FormatErrorPath(path []any, style string) string {
	if style == "bracket" {
		return formatBracketPath(path)
	}
	return ToDotPath(path)
}

// formatBracketPath formats path using bracket notation only.
func formatBracketPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var result strings.Builder
	result.Grow(len(path) * 8)

	for _, segment := range path {
		switch v := segment.(type) {
		case int:
			fmt.Fprintf(&result, "[%d]", v)
		case string:
			fmt.Fprintf(&result, `["%s"]`, v)
		default:
			fmt.Fprintf(&result, "[%v]", v)
		}
	}

	return result.String()
}
