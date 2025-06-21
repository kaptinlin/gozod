package issues

import (
	"fmt"
	"math"
	"strconv"

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
	if slicex.IsEmpty(values) {
		return ""
	}

	// Use slicex.Map to transform values to strings
	stringValues, err := slicex.Map(values, func(v any) any {
		return StringifyPrimitive(v)
	})
	if err != nil {
		// Fallback to manual processing
		quoted := make([]string, len(values))
		for i, v := range values {
			quoted[i] = StringifyPrimitive(v)
		}
		return slicex.Join(quoted, separator)
	}

	// Convert []any to []string using slicex
	if strings, err := slicex.ToTyped[string](stringValues); err == nil {
		return slicex.Join(strings, separator)
	}

	// Final fallback
	return slicex.Join(stringValues, separator)
}

// ParsedTypeToString converts input to parsed type string
// Enhanced with better float handling for NaN and Infinity
func ParsedTypeToString(input any) string {
	parsedType := reflectx.ParsedType(input)

	// Handle special cases to match TypeScript patterns
	switch parsedType {
	case core.ParsedTypeNaN:
		return "NaN"
	case core.ParsedTypeNil:
		return "null"
	case core.ParsedTypeSlice:
		return "array" // TypeScript doesn't differentiate slice/array
	case core.ParsedTypeArray:
		return "array"
	case core.ParsedTypeMap, core.ParsedTypeObject:
		return "object"
	case core.ParsedTypeStruct:
		return "object" // Structs are object-like in error messages
	case core.ParsedTypeComplex:
		return "complex" // Complex number type
	case core.ParsedTypeEnum:
		return "enum" // Enumeration type
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
		return "boolean"
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
		"uuid":             "UUID",
		"uuidv4":           "UUIDv4",
		"uuidv6":           "UUIDv6",
		"nanoid":           "nanoid",
		"guid":             "GUID",
		"cuid":             "cuid",
		"cuid2":            "cuid2",
		"ulid":             "ULID",
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
		received := ParsedTypeToString(raw.Input)
		return fmt.Sprintf("Invalid input: expected %s, received %s", expected, received)

	case core.InvalidValue:
		values := mapx.GetAnySliceDefault(raw.Properties, "values", nil)
		if slicex.IsEmpty(values) {
			return "Invalid value"
		}
		if len(values) == 1 {
			return fmt.Sprintf("Invalid input: expected %s", StringifyPrimitive(values[0]))
		}
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
		return "Invalid input"

	case core.InvalidElement:
		origin := mapx.GetStringDefault(raw.Properties, "origin", "")
		if origin == "" {
			return "Invalid element"
		}
		return fmt.Sprintf("Invalid value in %s", origin)

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

	inclusive := mapx.GetBoolDefault(raw.Properties, "inclusive", true)
	adj := GetComparisonOperator(inclusive, isTooSmall)
	sizing := GetSizing(origin)
	thresholdStr := FormatThreshold(threshold)

	if sizing != nil {
		if isTooSmall {
			return fmt.Sprintf("Too small: expected %s to have %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		} else {
			return fmt.Sprintf("Too big: expected %s to have %s%s %s", origin, adj, thresholdStr, sizing.Unit)
		}
	}

	if isTooSmall {
		return fmt.Sprintf("Too small: expected %s to be %s%s", origin, adj, thresholdStr)
	} else {
		return fmt.Sprintf("Too big: expected %s to be %s%s", origin, adj, thresholdStr)
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
