package issues

import (
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// CreateIssue creates a new ZodRawIssue with mapx for safer property handling
func CreateIssue(code core.IssueCode, message string, properties map[string]any, input any) core.ZodRawIssue {
	safeProps := mapx.Copy(properties)
	if safeProps == nil {
		safeProps = make(map[string]any)
	}

	return core.ZodRawIssue{
		Code:       code,
		Message:    message,
		Properties: safeProps,
		Input:      input,
		Path:       []any{},
	}
}

// CreateInvalidTypeIssue creates an invalid type issue
func CreateInvalidTypeIssue(expected core.ZodTypeCode, input any) core.ZodRawIssue {
	properties := map[string]any{
		"expected": string(expected),
		"received": string(reflectx.ParsedType(input)),
	}

	return CreateIssue(core.InvalidType, "", properties, input)
}

// CreateInvalidTypeIssueFromCode creates an invalid type issue using ZodTypeCode
func CreateInvalidTypeIssueFromCode(expected core.ZodTypeCode, input any) core.ZodRawIssue {
	properties := map[string]any{
		"expected": string(expected),
		"received": string(reflectx.ParsedType(input)),
	}

	return CreateIssue(core.InvalidType, "", properties, input)
}

// CreateInvalidValueIssue creates an invalid value issue
func CreateInvalidValueIssue(validValues []any, input any) core.ZodRawIssue {
	// Use slicex to get unique values
	values := validValues
	if !slicex.IsEmpty(validValues) {
		if uniqueValues, err := slicex.Unique(validValues); err == nil {
			if uniqueSlice, ok := uniqueValues.([]any); ok {
				values = uniqueSlice
			}
		}
	}

	properties := map[string]any{
		"values": values,
	}

	return CreateIssue(core.InvalidValue, "", properties, input)
}

// CreateTooBigIssue creates a "too big" issue
func CreateTooBigIssue(maximum any, inclusive bool, origin string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"maximum":   maximum,
		"inclusive": inclusive,
		"origin":    origin,
	}

	return CreateIssue(core.TooBig, "", properties, input)
}

// CreateTooSmallIssue creates a "too small" issue
func CreateTooSmallIssue(minimum any, inclusive bool, origin string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"minimum":   minimum,
		"inclusive": inclusive,
		"origin":    origin,
	}

	return CreateIssue(core.TooSmall, "", properties, input)
}

// CreateInvalidFormatIssue creates an invalid format issue
func CreateInvalidFormatIssue(format string, input any, additionalProps map[string]any) core.ZodRawIssue {
	properties := map[string]any{
		"format": format,
	}

	// Use mapx for safer property merging
	if additionalProps != nil {
		properties = mapx.Merge(properties, additionalProps)
	}

	return CreateIssue(core.InvalidFormat, "", properties, input)
}

// CreateNotMultipleOfIssue creates a "not multiple of" issue
func CreateNotMultipleOfIssue(divisor any, origin string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"divisor": divisor,
		"origin":  origin,
	}

	return CreateIssue(core.NotMultipleOf, "", properties, input)
}

// CreateUnrecognizedKeysIssue creates an unrecognized keys issue
func CreateUnrecognizedKeysIssue(keys []string, input any) core.ZodRawIssue {
	// Use slicex to get unique keys
	processedKeys := keys
	if !slicex.IsEmpty(keys) {
		if uniqueKeys, err := slicex.Unique(keys); err == nil {
			if uniqueSlice, ok := uniqueKeys.([]string); ok {
				processedKeys = uniqueSlice
			}
		}
	}

	properties := map[string]any{
		"keys": processedKeys,
	}

	return CreateIssue(core.UnrecognizedKeys, "", properties, input)
}

// CreateInvalidKeyIssue creates an invalid key issue
func CreateInvalidKeyIssue(key string, origin string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"key":    key,
		"origin": origin,
	}

	return CreateIssue(core.InvalidKey, "", properties, input)
}

// CreateInvalidUnionIssue creates an invalid union issue
func CreateInvalidUnionIssue(unionErrors []core.ZodRawIssue, input any) core.ZodRawIssue {
	properties := map[string]any{
		"union_errors": unionErrors,
	}

	// Use slicex for better error analysis
	if !slicex.IsEmpty(unionErrors) {
		properties["error_count"] = len(unionErrors)
	}

	return CreateIssue(core.InvalidUnion, "", properties, input)
}

// ConvertZodIssueToRaw converts a ZodIssue to ZodRawIssue efficiently
// This is a helper for common ZodError -> RawIssue conversion patterns
func ConvertZodIssueToRaw(issue core.ZodIssue) core.ZodRawIssue {
	// Preserve the original code rather than creating a custom issue
	return core.ZodRawIssue{
		Code:       issue.Code,
		Message:    issue.Message,
		Input:      issue.Input,
		Path:       []any{}, // Path will be set by caller
		Properties: make(map[string]any),
	}
}

// CreateInvalidElementIssue creates an invalid element issue
func CreateInvalidElementIssue(index int, origin string, input any, elementError core.ZodRawIssue) core.ZodRawIssue {
	properties := map[string]any{
		"index":         index,
		"origin":        origin,
		"element_error": elementError,
	}

	return CreateIssue(core.InvalidElement, "", properties, input)
}

// CreateCustomIssue creates a custom issue
func CreateCustomIssue(message string, properties map[string]any, input any) core.ZodRawIssue {
	// Use mapx for safer property handling
	safeProps := mapx.Copy(properties)
	if safeProps == nil {
		safeProps = make(map[string]any)
	}

	return CreateIssue(core.Custom, message, safeProps, input)
}

// CreateMissingKeyIssue creates a missing key issue
func CreateMissingKeyIssue(key string, options ...func(*core.ZodRawIssue)) core.ZodRawIssue {
	properties := map[string]any{
		"missing_key": key,
		"expected":    "required",
	}

	issue := CreateIssue(core.InvalidType, "", properties, nil)

	// Apply options if provided
	for _, option := range options {
		option(&issue)
	}

	return issue
}

// CreateErrorMap creates an error map from various input types
func CreateErrorMap(errorInput any) *core.ZodErrorMap {
	if errorInput == nil {
		return nil
	}

	switch e := errorInput.(type) {
	case string:
		errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return e
		})
		return &errorMap
	case core.ZodErrorMap:
		return &e
	case *core.ZodErrorMap:
		return e
	case func(core.ZodRawIssue) string:
		errorMap := core.ZodErrorMap(e)
		return &errorMap
	default:
		// Try to convert to string
		if str := fmt.Sprintf("%v", e); str != "" {
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return str
			})
			return &errorMap
		}
	}

	return nil
}

// CreateInvalidTypeWithMsg creates an invalid type issue with custom message
// This function supports type-safe expected types using core.ZodTypeCode
func CreateInvalidTypeWithMsg(expected core.ZodTypeCode, message string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"expected": string(expected),
		"received": string(reflectx.ParsedType(input)),
	}

	return CreateIssue(core.InvalidType, message, properties, input)
}

// CreateInvalidUnionIssueWithResults creates an invalid union issue with results array
// This maintains TypeScript-compatible structure with "errors" property for union validation
func CreateInvalidUnionIssueWithResults(unionErrors []core.ZodRawIssue, input any, continueOnError ...bool) core.ZodRawIssue {
	properties := map[string]any{
		"errors": unionErrors, // TypeScript-compatible property name
	}

	// Use slicex for better error analysis
	if !slicex.IsEmpty(unionErrors) {
		properties["error_count"] = len(unionErrors)
	}

	// Add continue flag if specified
	if len(continueOnError) > 0 && continueOnError[0] {
		properties["continue"] = true
	}

	return CreateIssue(core.InvalidUnion, "", properties, input)
}

// CreateNonOptionalIssue returns invalid_type issue with expected "nonoptional" – used by .NonOptional()
func CreateNonOptionalIssue(input any) core.ZodRawIssue {
	props := map[string]any{
		"expected": "nonoptional",
		"received": string(reflectx.ParsedType(input)),
	}
	return CreateIssue(core.InvalidType, "", props, input)
}
