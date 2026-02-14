// Package utils provides internal utility functions for the gozod validation library,
// including error map conversion, parameter normalization, value origin detection,
// value comparison, and error path formatting.
package utils

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/reflectx"
)

// Error conversion

// ToErrorMap converts various error representations to a ZodErrorMap.
// It accepts string, ZodErrorMap, *ZodErrorMap, or func(ZodRawIssue) string.
// The second return value reports whether the conversion succeeded.
func ToErrorMap(err any) (*core.ZodErrorMap, bool) {
	switch v := err.(type) {
	case string:
		return new(core.ZodErrorMap(func(core.ZodRawIssue) string { return v })), true
	case core.ZodErrorMap:
		return new(v), true
	case *core.ZodErrorMap:
		return v, true
	case func(core.ZodRawIssue) string:
		return new(core.ZodErrorMap(v)), true
	default:
		return nil, false
	}
}

// Parameter helpers

// FirstParam returns the first element of params, or nil if params is empty.
func FirstParam(params ...any) any {
	if len(params) == 0 {
		return nil
	}
	return params[0]
}

// NormalizeParams converts the first variadic argument into a SchemaParams.
// It accepts nil, string, SchemaParams, or *SchemaParams.
// Unrecognized types yield an empty SchemaParams.
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
		return new(v)
	case *core.SchemaParams:
		if v == nil {
			return &core.SchemaParams{}
		}
		cp := *v
		return &cp
	default:
		return &core.SchemaParams{}
	}
}

// NormalizeCustomParams converts the first variadic argument into a CustomParams.
// It accepts nil, string, CustomParams, or *CustomParams.
// Any other type is stored as the Error field directly.
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
		return new(v)
	case *core.CustomParams:
		if v == nil {
			return &core.CustomParams{}
		}
		cp := *v
		return &cp
	default:
		return &core.CustomParams{Error: v}
	}
}

// ApplySchemaParams sets the error map on def from params.
// It is a no-op when params or params.Error is nil.
func ApplySchemaParams(def *core.ZodTypeDef, params *core.SchemaParams) {
	if params == nil || params.Error == nil {
		return
	}
	if m, ok := ToErrorMap(params.Error); ok {
		def.Error = m
	}
}

// Value origin detection

// OriginFromValue returns a human-readable origin label for value,
// used in error messages. Possible results: "number", "string", "array",
// "object", or "unknown".
func OriginFromValue(value any) string {
	switch {
	case reflectx.IsNumeric(value):
		return "number"
	case reflectx.IsString(value):
		return "string"
	case reflectx.HasLength(value):
		return "array"
	case reflectx.IsMap(value):
		return "object"
	default:
		return "unknown"
	}
}

// NumericOrigin classifies value into a numeric origin category.
// Possible results: "nil", "integer", "number", "bigint", "string", or "unknown".
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
		pt := reflectx.ParsedType(value)
		switch pt { //nolint:exhaustive // only bigint and string need special handling
		case core.ParsedTypeBigint:
			return "bigint"
		case core.ParsedTypeString:
			return "string"
		default:
			return "unknown"
		}
	}
}

// SizableOrigin classifies value into a sizable origin category.
// Possible results: "nil", "string", "slice", "array", "map", "struct",
// "file", or "unknown".
func SizableOrigin(value any) string {
	if value == nil {
		return "nil"
	}

	switch {
	case reflectx.IsString(value):
		return "string"
	case reflectx.IsSlice(value):
		return "slice"
	case reflectx.IsArray(value):
		return "array"
	case reflectx.IsMap(value):
		return "map"
	case reflectx.IsStruct(value):
		return "struct"
	default:
		pt := reflectx.ParsedType(value)
		if pt == core.ParsedTypeFile {
			return "file"
		}
		return "unknown"
	}
}

// LengthableOrigin classifies value into a lengthable origin category.
// Possible results: "nil", "string", "slice", "array", or "unknown".
func LengthableOrigin(value any) string {
	if value == nil {
		return "nil"
	}

	switch {
	case reflectx.IsString(value):
		return "string"
	case reflectx.IsSlice(value):
		return "slice"
	case reflectx.IsArray(value):
		return "array"
	default:
		return "unknown"
	}
}

// Value comparison

// CompareValues compares a and b after dereferencing all pointers.
// It returns -1, 0, or 1. Mismatched or unsupported types return 0.
func CompareValues(a, b any) int {
	da := reflectx.DerefAll(a)
	db := reflectx.DerefAll(b)

	if da == nil && db == nil {
		return 0
	}
	if da == nil {
		return -1
	}
	if db == nil {
		return 1
	}

	switch va := da.(type) {
	case int:
		if vb, ok := db.(int); ok {
			return cmp.Compare(va, vb)
		}
	case int64:
		if vb, ok := db.(int64); ok {
			return cmp.Compare(va, vb)
		}
	case float64:
		if vb, ok := db.(float64); ok {
			return cmp.Compare(va, vb)
		}
	case float32:
		if vb, ok := db.(float32); ok {
			return cmp.Compare(va, vb)
		}
	case string:
		if vb, ok := db.(string); ok {
			return cmp.Compare(va, vb)
		}
	}

	return 0
}

// Path formatting

// ToDotPath converts an error path to dot notation.
// Integer segments use bracket notation (e.g., [0]), string segments use dot
// notation unless they contain non-identifier characters, in which case bracket
// notation with quotes is used. Compatible with TypeScript Zod v4 path formatting.
func ToDotPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(path) * 8)

	for i, seg := range path {
		switch v := seg.(type) {
		case int:
			b.WriteByte('[')
			b.WriteString(strconv.Itoa(v))
			b.WriteByte(']')
		case string:
			switch {
			case i == 0:
				b.WriteString(v)
			case needsBracketNotation(v):
				b.WriteString(`["`)
				b.WriteString(v)
				b.WriteString(`"]`)
			default:
				b.WriteByte('.')
				b.WriteString(v)
			}
		default:
			fmt.Fprintf(&b, "[%v]", v)
		}
	}

	return b.String()
}

// FormatErrorPath formats an error path for display using the given style.
// Supported styles are "bracket" (all bracket notation) and "dot" (default).
func FormatErrorPath(path []any, style string) string {
	if style == "bracket" {
		return formatBracketPath(path)
	}
	return ToDotPath(path)
}

// formatBracketPath formats every segment of path using bracket notation.
func formatBracketPath(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(path) * 8)

	for _, seg := range path {
		switch v := seg.(type) {
		case int:
			b.WriteByte('[')
			b.WriteString(strconv.Itoa(v))
			b.WriteByte(']')
		case string:
			b.WriteString(`["`)
			b.WriteString(v)
			b.WriteString(`"]`)
		default:
			fmt.Fprintf(&b, "[%v]", v)
		}
	}

	return b.String()
}

// needsBracketNotation reports whether s requires bracket notation in a dot path.
// A key needs brackets if it starts with a digit or contains non-identifier characters.
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

// isIdentChar reports whether c is a valid identifier character (letter, digit, or underscore).
func isIdentChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}
