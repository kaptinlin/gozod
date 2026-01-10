package issues

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// MESSAGE FORMATTING SYSTEM
// =============================================================================

// MessageFormatter provides a unified interface for formatting validation error messages
// Compatible with TypeScript Zod v4 error formatting patterns
//
// Note: Error structure formatting (tree, flat, pretty) is handled by standalone functions
// in the errors.go file, following TypeScript Zod v4's functional approach:
// - TreeifyError() / TreeifyErrorWithMapper()
// - FlattenError() / FlattenErrorWithMapper()
// - PrettifyError() / PrettifyErrorWithFormatter()
// - FormatError() / FormatErrorWithMapper()
type MessageFormatter interface {
	FormatMessage(raw core.ZodRawIssue) string
}

// DefaultMessageFormatter implements the default English message formatting
type DefaultMessageFormatter struct{}

var defaultFormatter = &DefaultMessageFormatter{}

// =============================================================================
// SHARED FORMATTER UTILITIES
// =============================================================================

// SizingInfo represents sizing terminology for different types
type SizingInfo struct {
	Unit string // The unit name (e.g., "characters", "items")
	Verb string // The verb to use (e.g., "to have")
}

// Sizable maps type names to their sizing terminology using mapx for better type safety
var Sizable = func() map[string]SizingInfo {
	sizableData := map[string]SizingInfo{
		"string": {Unit: "characters", Verb: "to have"},
		"file":   {Unit: "bytes", Verb: "to have"},
		"array":  {Unit: "items", Verb: "to have"},
		"slice":  {Unit: "items", Verb: "to have"},
		"set":    {Unit: "items", Verb: "to have"},
		"object": {Unit: "keys", Verb: "to have"},
		"map":    {Unit: "keys", Verb: "to have"},
	}
	return sizableData
}()

// GetSizing returns the appropriate sizing information for a given type
func GetSizing(origin string) *SizingInfo {
	if info, exists := Sizable[origin]; exists {
		return &info
	}
	return nil
}

// StringifyPrimitive converts a primitive value to its string representation with quotes
// Enhanced with better type handling using modern Go practices
func StringifyPrimitive(value any) string {
	switch v := value.(type) {
	case string:
		return `"` + v + `"`
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		f, err := coerce.ToFloat64(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		if math.IsNaN(f) {
			return "NaN"
		}
		if math.IsInf(f, 0) {
			if f > 0 {
				return "Infinity"
			}
			return "-Infinity"
		}
		// Format without unnecessary decimals
		if f == float64(int64(f)) {
			return fmt.Sprintf("%.0f", f)
		}
		return fmt.Sprintf("%g", f)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

// JoinValuesWithSeparator formats an array of values with a custom separator using slicex
// Enhanced with slicex for better slice handling and type safety
func JoinValuesWithSeparator(values []any, separator string) string {
	if len(values) == 0 {
		return ""
	}

	// Handle single value case directly without string building overhead
	if len(values) == 1 {
		return StringifyPrimitive(values[0])
	}

	// Pre-allocate builder capacity based on estimated string lengths
	var builder strings.Builder
	builder.Grow(len(values) * 10) // Rough estimate

	for i, v := range values {
		if i > 0 {
			builder.WriteString(separator)
		}
		builder.WriteString(StringifyPrimitive(v))
	}

	return builder.String()
}

// ParsedTypeToString converts input to parsed type string
// Enhanced with better float handling for NaN and Infinity
// Matches TypeScript Zod v4 reference implementation behavior
func ParsedTypeToString(input any) string {
	parsedType := reflectx.ParsedType(input)

	// Handle special cases to match Go language semantics
	switch parsedType {
	case core.ParsedTypeNaN:
		return "NaN"
	case core.ParsedTypeNil:
		return "nil" // Go language null value representation
	case core.ParsedTypeSlice:
		return "slice" // Go language slice type
	case core.ParsedTypeArray:
		return "array"
	case core.ParsedTypeMap:
		return "map" // Go language map type
	case core.ParsedTypeObject:
		return "object"
	case core.ParsedTypeStruct:
		return "struct" // Go language struct type
	case core.ParsedTypeComplex:
		return "complex" // Complex number type in Go
	case core.ParsedTypeEnum:
		return "enum" // Enumeration type maps to enum in Go semantic types
	case core.ParsedTypeTuple:
		return "tuple" // Tuple type
	case core.ParsedTypeFloat:
		// Check for special float values
		switch v := input.(type) {
		case float32:
			f64 := float64(v)
			if math.IsNaN(f64) {
				return "NaN"
			}
			if math.IsInf(f64, 0) {
				return "Infinity"
			}
		case float64:
			if math.IsNaN(v) {
				return "NaN"
			}
			if math.IsInf(v, 0) {
				return "Infinity"
			}
		}
		return "number"
	case core.ParsedTypeNumber:
		return "number"
	case core.ParsedTypeBigint:
		return "bigint"
	case core.ParsedTypeBool:
		return "bool" // Go language boolean type
	case core.ParsedTypeString:
		return "string"
	case core.ParsedTypeFunction:
		return "function"
	case core.ParsedTypeFile:
		return "File"
	case core.ParsedTypeDate:
		return "Date"
	case core.ParsedTypeUnknown:
		return "unknown"
	default:
		return string(parsedType)
	}
}

// FormatThreshold converts a threshold value to string for consistent formatting
func FormatThreshold(threshold any) string {
	switch v := threshold.(type) {
	case int:
		return strconv.Itoa(v)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int(v)) {
			return strconv.Itoa(int(v))
		}
		return fmt.Sprintf("%.1f", v)
	case float32:
		f := float64(v)
		if f == float64(int(f)) {
			return strconv.Itoa(int(f))
		}
		return fmt.Sprintf("%.1f", f)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetComparisonOperator returns the comparison operator string based on inclusivity
func GetComparisonOperator(isInclusive bool, isGreaterThan bool) string {
	if isGreaterThan {
		if isInclusive {
			return ">="
		}
		return ">"
	} else {
		if isInclusive {
			return "<="
		}
		return "<"
	}
}

// GetFriendlyComparisonText returns user-friendly comparison text instead of mathematical operators
func GetFriendlyComparisonText(isInclusive bool, isTooSmall bool) string {
	if isTooSmall {
		if isInclusive {
			return "at least "
		}
		return "more than "
	} else {
		if isInclusive {
			return "at most "
		}
		return "less than "
	}
}

// =============================================================================
// FORMAT NAME MAPPINGS - Enhanced with mapx
// =============================================================================

// FormatNouns maps format names to human-readable descriptions
// Enhanced initialization using modern Go patterns
var FormatNouns = func() map[string]string {
	return map[string]string{
		"regex":            "input",
		"email":            "email address",
		"url":              "URL",
		"emoji":            "emoji",
		"uuid":             "uuid",
		"uuidv4":           "uuid",
		"uuidv6":           "uuid",
		"nanoid":           "nanoid",
		"guid":             "guid",
		"cuid":             "cuid",
		"cuid2":            "cuid2",
		"ulid":             "ulid",
		"xid":              "XID",
		"ksuid":            "KSUID",
		"datetime":         "ISO datetime",
		"date":             "ISO date",
		"time":             "ISO time",
		"duration":         "ISO duration",
		"ipv4":             "IPv4 address",
		"ipv6":             "IPv6 address",
		"cidrv4":           "IPv4 range",
		"cidrv6":           "IPv6 range",
		"base64":           "base64-encoded string",
		"base64url":        "base64url-encoded string",
		"json_string":      "JSON string",
		"e164":             "E.164 number",
		"jwt":              "JWT",
		"template_literal": "input",
		// ISO formats
		"iso_date":     "ISO date format",
		"iso_time":     "ISO time format",
		"iso_datetime": "ISO datetime format",
		"iso_duration": "ISO duration",
		// Go-specific formats
		"int8":       "8-bit integer",
		"int16":      "16-bit integer",
		"int32":      "32-bit integer",
		"int64":      "64-bit integer",
		"uint8":      "8-bit unsigned integer",
		"uint16":     "16-bit unsigned integer",
		"uint32":     "32-bit unsigned integer",
		"uint64":     "64-bit unsigned integer",
		"float32":    "32-bit float",
		"float64":    "64-bit float",
		"complex64":  "64-bit complex number",
		"complex128": "128-bit complex number",
	}
}()

// GetFormatNoun returns the human-readable noun for a format name
// Enhanced with mapx for safer access
func GetFormatNoun(format string) string {
	if noun, exists := FormatNouns[format]; exists {
		return noun
	}
	return format
}

// =============================================================================
// MAIN FORMATTING LOGIC - Enhanced with all utility packages
// =============================================================================

// FormatMessage generates error messages using enhanced utilities
func (f *DefaultMessageFormatter) FormatMessage(raw core.ZodRawIssue) string {
	code := raw.Code

	switch code {
	case core.InvalidType:
		expected := mapx.GetStringDefault(raw.Properties, "expected", "")
		// Special handling for StringBool type to display "boolean" instead of "stringbool"
		if expected == "stringbool" {
			expected = "boolean"
		}
		// Special handling for complex types to display "complex" instead of specific types
		if expected == "complex64" || expected == "complex128" {
			expected = "complex"
		}
		received := ParsedTypeToString(raw.Input)

		// Special handling for object type conversion errors
		if expected == "object" && (received == "string" || received == "map") {
			return fmt.Sprintf("Type conversion failed: cannot convert %s to %s", received, expected)
		}

		return fmt.Sprintf("Invalid input: expected %s, received %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if slicex.IsEmpty(values) {
			return "Invalid value"
		}
		// TypeScript always uses "Invalid option: expected one of" format, even for single values
		return fmt.Sprintf("Invalid option: expected one of %s", JoinValuesWithSeparator(values, "|"))

	case core.TooBig:
		return f.formatSizeConstraint(raw, false)

	case core.TooSmall:
		return f.formatSizeConstraint(raw, true)

	case core.InvalidFormat:
		format := mapx.GetStringDefault(raw.Properties, "format", "")
		if format == "" {
			return "Invalid format"
		}
		return f.formatStringValidation(raw, format)

	case core.NotMultipleOf:
		divisor := mapx.GetAnyDefault(raw.Properties, "divisor", nil)
		if divisor == nil {
			return "Invalid number: must be a multiple of divisor"
		}
		return fmt.Sprintf("Invalid number: must be a multiple of %v", divisor)

	case core.UnrecognizedKeys:
		keys := mapx.GetStringsDefault(raw.Properties, "keys", nil)
		if slicex.IsEmpty(keys) {
			return "Unrecognized key(s) in object"
		}
		keyStr := "key"
		if len(keys) > 1 {
			keyStr = "keys"
		}
		// Convert to []any for consistent handling
		if keysAny, err := slicex.ToAny(keys); err == nil {
			return fmt.Sprintf("Unrecognized %s: %s", keyStr, JoinValuesWithSeparator(keysAny, ", "))
		}
		return fmt.Sprintf("Unrecognized %s in object", keyStr)

	case core.InvalidKey:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Invalid key"
		}
		return fmt.Sprintf("Invalid key in %s", origin)

	case core.InvalidUnion:
		return "Invalid input: no union member matched"

	case core.InvalidElement:
		// Get origin and index information first
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		index := mapx.GetAnyDefault(raw.Properties, "index", nil)

		// Try to extract the original element error message
		if elementError, exists := raw.Properties["element_error"]; exists {
			if elementRaw, ok := elementError.(core.ZodRawIssue); ok {
				// Format the original element error and add index information
				elementMessage := f.FormatMessage(elementRaw)
				if index != nil {
					if origin == "array rest" {
						return fmt.Sprintf("%s (rest element at index %v)", elementMessage, index)
					}
					return fmt.Sprintf("%s (element at index %v)", elementMessage, index)
				}
				return elementMessage
			}
		}
		// Fallback to generic message if element_error is not available
		if origin == "" {
			return "Invalid element"
		}
		if index != nil {
			if origin == "array rest" {
				return fmt.Sprintf("Invalid value in %s: rest element at index %v", origin, index)
			}
			return fmt.Sprintf("Invalid value in %s: element at index %v", origin, index)
		}
		return fmt.Sprintf("Invalid value in %s", origin)

	case core.MissingRequired:
		fieldName := mapx.GetStringDefault(raw.Properties, "field_name", "")
		fieldType := mapx.GetStringDefault(raw.Properties, "field_type", "field")
		if fieldName == "" {
			return fmt.Sprintf("Missing required %s", fieldType)
		}
		return fmt.Sprintf("Missing required %s: %s", fieldType, fieldName)

	case core.TypeConversion:
		fromType := mapx.GetStringDefault(raw.Properties, "from_type", "unknown")
		toType := mapx.GetStringDefault(raw.Properties, "to_type", "unknown")
		return fmt.Sprintf("Type conversion failed: cannot convert %s to %s", fromType, toType)

	case core.InvalidSchema:
		// Prefer reason from properties if provided
		reason := mapx.GetStringDefault(raw.Properties, "reason", "")
		if reason != "" {
			return fmt.Sprintf("Invalid schema: %s", reason)
		}
		return "Invalid schema definition"

	case core.InvalidDiscriminator:
		field := mapx.GetStringDefault(raw.Properties, "field", "discriminator")
		return fmt.Sprintf("Invalid or missing discriminator field: %s", field)

	case core.IncompatibleTypes:
		conflictType := mapx.GetStringDefault(raw.Properties, "conflict_type", "values")
		return fmt.Sprintf("Cannot merge %s: incompatible types", conflictType)

	case core.NilPointer:
		return "Nil pointer encountered"

	case core.Custom:
		// Prefer explicit message field if provided
		if raw.Message != "" {
			return raw.Message
		}
		// Fallback to properties.message for backward compatibility
		message := mapx.GetStringDefault(raw.Properties, "message", "")
		if message != "" {
			return message
		}
		return "Invalid input"

	default:
		return "Invalid input"
	}
}

// formatSizeConstraint formats size constraint messages using enhanced utilities
// Provides user-friendly messages that match TypeScript Zod v4 format
func (f *DefaultMessageFormatter) formatSizeConstraint(raw core.ZodRawIssue, isTooSmall bool) string {
	origin := mapx.GetStringDefault(raw.Properties, "origin", "value")

	var threshold any
	if isTooSmall {
		threshold = mapx.GetAnyDefault(raw.Properties, "minimum", nil)
	} else {
		threshold = mapx.GetAnyDefault(raw.Properties, "maximum", nil)
	}

	if threshold == nil {
		if isTooSmall {
			return "Too small"
		}
		return "Too big"
	}

	// Special handling for arrays
	if origin == "array" {
		minimum := mapx.GetAnyDefault(raw.Properties, "minimum", nil)
		maximum := mapx.GetAnyDefault(raw.Properties, "maximum", nil)

		// Check if this is a rest parameter array
		if isRestParam, ok := raw.Properties["is_rest_param"].(bool); ok && isRestParam {
			return fmt.Sprintf("expected at least %s", FormatThreshold(minimum))
		}

		// Fixed-length arrays: when both minimum and maximum are present and equal
		if minimum != nil && maximum != nil {
			// Check if minimum and maximum are equal (fixed length)
			minStr := FormatThreshold(minimum)
			maxStr := FormatThreshold(maximum)
			if minStr == maxStr {
				return fmt.Sprintf("expected exactly %s", minStr)
			}
		}
	}

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	sizing := GetSizing(origin)
	thresholdStr := FormatThreshold(threshold)

	// Special handling for file size validation to match expected format
	if origin == "file" {
		if isTooSmall {
			return fmt.Sprintf("File size must be at least %s bytes", thresholdStr)
		} else {
			return fmt.Sprintf("File size must be at most %s bytes", thresholdStr)
		}
	}

	// Use consistent "Too small/Too big: expected X to be/have Y" format for other types
	if sizing != nil {
		// For sized types (strings, arrays, etc.), use "have" with sizing info
		adj := GetFriendlyComparisonText(inclusive, isTooSmall)
		if isTooSmall {
			return fmt.Sprintf("Too small: expected %s to have %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		} else {
			return fmt.Sprintf("Too big: expected %s to have %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
	} else {
		// For numeric and other types, use "be" format
		adj := GetFriendlyComparisonText(inclusive, isTooSmall)
		if isTooSmall {
			return fmt.Sprintf("Too small: expected %s to be %s%s", origin, adj, thresholdStr)
		} else {
			return fmt.Sprintf("Too big: expected %s to be %s%s", origin, adj, thresholdStr)
		}
	}
}

// formatStringValidation handles string format validation messages using enhanced utilities
func (f *DefaultMessageFormatter) formatStringValidation(raw core.ZodRawIssue, format string) string {
	switch format {
	case "starts_with":
		prefix := mapx.GetStringDefault(raw.Properties, "prefix", "")
		if prefix == "" {
			return "Invalid string: must start with specified prefix"
		}
		return fmt.Sprintf("Invalid string: must start with %s", StringifyPrimitive(prefix))
	case "ends_with":
		suffix := mapx.GetStringDefault(raw.Properties, "suffix", "")
		if suffix == "" {
			return "Invalid string: must end with specified suffix"
		}
		return fmt.Sprintf("Invalid string: must end with %s", StringifyPrimitive(suffix))
	case "includes":
		includes := mapx.GetStringDefault(raw.Properties, "includes", "")
		if includes == "" {
			return "Invalid string: must include specified substring"
		}
		return fmt.Sprintf("Invalid string: must include %s", StringifyPrimitive(includes))
	case "regex":
		pattern := mapx.GetStringDefault(raw.Properties, "pattern", "")
		if pattern == "" {
			return "Invalid string: must match pattern"
		}
		return fmt.Sprintf("Invalid string: must match pattern %s", pattern)
	default:
		noun := GetFormatNoun(format)
		return fmt.Sprintf("Invalid %s", noun)
	}
}

// =============================================================================
// CONVENIENCE FUNCTIONS - Enhanced with all utilities
// =============================================================================

// GenerateDefaultMessage generates a default error message for an issue using enhanced utilities
func GenerateDefaultMessage(raw core.ZodRawIssue) string {
	return defaultFormatter.FormatMessage(raw)
}

// FormatMessage formats a single issue using the default formatter
func FormatMessage(raw core.ZodRawIssue) string {
	return GenerateDefaultMessage(raw)
}

// FormatMessageWithFormatter formats a message using a custom formatter
func FormatMessageWithFormatter(raw core.ZodRawIssue, formatter MessageFormatter) string {
	return formatter.FormatMessage(raw)
}
