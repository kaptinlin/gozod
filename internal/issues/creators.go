package issues

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// CreateIssue creates a new ZodRawIssue with mapx for safer property handling
func CreateIssue(code core.IssueCode, message string, properties map[string]any, input any) core.ZodRawIssue {
	// Skip expensive copy operation when properties map is empty
	var safeProps map[string]any
	if len(properties) == 0 {
		safeProps = make(map[string]any)
	} else {
		safeProps = mapx.Copy(properties)
	}

	return core.ZodRawIssue{
		Code:       code,
		Message:    message,
		Properties: safeProps,
		Input:      input,
		Path:       []any{},
	}
}

// =============================================================================
// LOW-LEVEL ISSUE CREATION (Internal Use Only)
// =============================================================================

// CreateInvalidTypeIssue creates an invalid type issue
func CreateInvalidTypeIssue(expected core.ZodTypeCode, input any) core.ZodRawIssue {
	// Pre-allocate map capacity to avoid rehashing during insertion
	properties := make(map[string]any, 2)
	properties["expected"] = string(expected)
	properties["received"] = string(reflectx.ParsedType(input))

	return CreateIssue(core.InvalidType, "", properties, input)
}

// CreateInvalidValueIssue creates an invalid value issue
func CreateInvalidValueIssue(validValues []any, input any) core.ZodRawIssue {
	// Apply unique operation only for small slices to balance deduplication vs performance
	values := validValues
	if len(validValues) > 1 && len(validValues) <= 10 {
		// For small slices, deduplication cost is minimal
		if uniqueValues, err := slicex.Unique(validValues); err == nil {
			if uniqueSlice, ok := uniqueValues.([]any); ok {
				values = uniqueSlice
			}
		}
	} else if len(validValues) > 10 {
		// For larger slices, keep original to avoid O(n²) unique operation
		values = validValues
	}

	// Single-property map creation
	properties := make(map[string]any, 1)
	properties["values"] = values

	return CreateIssue(core.InvalidValue, "", properties, input)
}

// CreateTooBigIssue creates a "too big" issue
func CreateTooBigIssue(maximum any, inclusive bool, origin string, input any) core.ZodRawIssue {
	// Pre-allocate map capacity for three known properties
	properties := make(map[string]any, 3)
	properties["maximum"] = maximum
	properties["inclusive"] = inclusive
	properties["origin"] = origin

	return CreateIssue(core.TooBig, "", properties, input)
}

// CreateTooSmallIssue creates a "too small" issue
func CreateTooSmallIssue(minimum any, inclusive bool, origin string, input any) core.ZodRawIssue {
	// Pre-allocate map capacity for three known properties
	properties := make(map[string]any, 3)
	properties["minimum"] = minimum
	properties["inclusive"] = inclusive
	properties["origin"] = origin

	return CreateIssue(core.TooSmall, "", properties, input)
}

// CreateFixedLengthArrayIssue creates a size constraint issue for fixed-length arrays
// This function sets both minimum and maximum properties to enable "expected exactly N" formatting
func CreateFixedLengthArrayIssue(expectedLength any, actualLength int, input any, isTooSmall bool) core.ZodRawIssue {
	properties := map[string]any{
		"minimum":   expectedLength,
		"maximum":   expectedLength,
		"inclusive": true,
		"origin":    "array",
	}

	code := core.TooSmall
	if !isTooSmall {
		code = core.TooBig
	}

	return CreateIssue(code, "", properties, input)
}

// CreateInvalidFormatIssue creates an invalid format issue
func CreateInvalidFormatIssue(format string, input any, additionalProps map[string]any) core.ZodRawIssue {
	// Calculate map capacity based on format field plus any additional properties
	size := 1
	if additionalProps != nil {
		size += len(additionalProps)
	}
	properties := make(map[string]any, size)
	properties["format"] = format

	// Use mapx for safer property merging
	if additionalProps != nil {
		properties = mapx.Merge(properties, additionalProps)
	}

	return CreateIssue(core.InvalidFormat, "", properties, input)
}

// CreateNotMultipleOfIssue creates a "not multiple of" issue
func CreateNotMultipleOfIssue(divisor any, origin string, input any) core.ZodRawIssue {
	// Pre-allocate map capacity for two known properties
	properties := make(map[string]any, 2)
	properties["divisor"] = divisor
	properties["origin"] = origin

	return CreateIssue(core.NotMultipleOf, "", properties, input)
}

// CreateUnrecognizedKeysIssue creates an unrecognized keys issue
func CreateUnrecognizedKeysIssue(keys []string, input any) core.ZodRawIssue {
	// Apply deduplication only for small key sets where performance cost is negligible
	processedKeys := keys
	if len(keys) > 1 && len(keys) <= 5 {
		// For small key sets, deduplication improves error message clarity
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

// CreateInvalidElementIssue creates an invalid element issue with proper path
func CreateInvalidElementIssue(index int, origin string, input any, elementError core.ZodRawIssue) core.ZodRawIssue {
	properties := map[string]any{
		"index":         index,
		"origin":        origin,
		"element_error": elementError,
	}

	// Create the raw issue and set the path to [index]
	rawIssue := CreateIssue(core.InvalidElement, "", properties, input)
	rawIssue.Path = []any{index}
	return rawIssue
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

// CreateMissingRequiredIssue creates a missing required field issue
func CreateMissingRequiredIssue(fieldName string, fieldType string) core.ZodRawIssue {
	properties := map[string]any{"field_name": fieldName, "field_type": fieldType}
	return CreateIssue(core.MissingRequired, "", properties, nil)
}

// CreateTypeConversionIssue creates a type conversion failure issue
func CreateTypeConversionIssue(fromType, toType string, input any) core.ZodRawIssue {
	properties := map[string]any{"from_type": fromType, "to_type": toType}
	return CreateIssue(core.TypeConversion, "", properties, input)
}

// CreateInvalidSchemaIssue creates an invalid schema issue
func CreateInvalidSchemaIssue(reason string, input any, additionalProps ...map[string]any) core.ZodRawIssue {
	properties := map[string]any{
		"reason": reason,
	}

	// Use mapx for safer property merging
	if len(additionalProps) > 0 && additionalProps[0] != nil {
		properties = mapx.Merge(properties, additionalProps[0])
	}

	return CreateIssue(core.InvalidSchema, "", properties, input)
}

// CreateIncompatibleTypesIssue creates an incompatible types issue
func CreateIncompatibleTypesIssue(conflictType string, value1, value2 any, input any) core.ZodRawIssue {
	properties := map[string]any{"conflict_type": conflictType, "value1": value1, "value2": value2}
	return CreateIssue(core.IncompatibleTypes, "", properties, input)
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

// =============================================================================
// HIGH-LEVEL ERROR CREATION API
// =============================================================================

// CreateFinalError directly creates a final ZodError, skipping intermediate steps
// This is the recommended way to create errors in most cases
func CreateFinalError(code core.IssueCode, message string, properties map[string]any, input any, ctx *core.ParseContext, config *core.ZodConfig) error {
	raw := CreateIssue(code, message, properties, input)
	final := FinalizeIssue(raw, ctx, config)
	return NewZodError([]core.ZodIssue{final})
}

// CreateNonOptionalError creates a standardized non-optional error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateNonOptionalError(ctx *core.ParseContext) error {
	raw := CreateNonOptionalIssue(nil)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidTypeError creates a standardized invalid type error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateInvalidTypeError(expectedType core.ZodTypeCode, input any, ctx *core.ParseContext) error {
	raw := CreateInvalidTypeIssue(expectedType, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidTypeErrorWithInst creates an invalid type error with schema internals
func CreateInvalidTypeErrorWithInst(expectedType core.ZodTypeCode, input any, ctx *core.ParseContext, inst any) error {
	raw := CreateInvalidTypeIssue(expectedType, input)
	raw.Inst = inst
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidValueError creates a standardized invalid value error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateInvalidValueError(validValues []any, input any, ctx *core.ParseContext) error {
	raw := CreateInvalidValueIssue(validValues, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateTooBigError creates a standardized "too big" error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateTooBigError(maximum any, inclusive bool, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateTooBigIssue(maximum, inclusive, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateTooSmallError creates a standardized "too small" error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateTooSmallError(minimum any, inclusive bool, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateTooSmallIssue(minimum, inclusive, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateRestParameterTooSmallError creates a TooSmall error specifically for rest parameter arrays
func CreateRestParameterTooSmallError(minimum any, inclusive bool, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateTooSmallIssue(minimum, inclusive, origin, input)
	// Add rest parameter flag to properties
	raw.Properties["is_rest_param"] = true
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateFixedLengthArrayError creates an error for fixed-length array validation
func CreateFixedLengthArrayError(expectedLength any, actualLength int, input any, isTooSmall bool, ctx *core.ParseContext) error {
	raw := CreateFixedLengthArrayIssue(expectedLength, actualLength, input, isTooSmall)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidFormatError creates a standardized invalid format error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateInvalidFormatError(format string, input any, ctx *core.ParseContext, additionalProps ...map[string]any) error {
	var props map[string]any
	if len(additionalProps) > 0 {
		props = additionalProps[0]
	}
	raw := CreateInvalidFormatIssue(format, input, props)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateCustomError creates a standardized custom error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateCustomError(message string, properties map[string]any, input any, ctx *core.ParseContext) error {
	raw := CreateCustomIssue(message, properties, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateUnrecognizedKeysError creates a standardized unrecognized keys error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateUnrecognizedKeysError(keys []string, input any, ctx *core.ParseContext) error {
	raw := CreateUnrecognizedKeysIssue(keys, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidKeyError creates a standardized invalid key error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateInvalidKeyError(key string, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateInvalidKeyIssue(key, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateElementValidationIssue creates a raw issue for invalid element validation
func CreateElementValidationIssue(index int, origin string, element any, elementError error) core.ZodRawIssue {
	// Convert error to raw issue if it's a ZodError
	var elementRawIssue core.ZodRawIssue
	var zodErr *ZodError
	if errors.As(elementError, &zodErr) && len(zodErr.Issues) > 0 {
		// Extract the raw issue from the ZodError
		elementRawIssue = ConvertZodIssueToRaw(zodErr.Issues[0])
	} else {
		// Create a basic issue for non-ZodError
		elementRawIssue = CreateIssue(core.InvalidElement, elementError.Error(), nil, element)
	}

	// Use the existing CreateInvalidElementIssue function
	return CreateInvalidElementIssue(index, origin, element, elementRawIssue)
}

// CreateArrayValidationIssues creates multiple issues for array validation, collecting all validation errors
func CreateArrayValidationIssues(issues []core.ZodRawIssue) error {
	if len(issues) == 0 {
		return nil
	}

	// Convert raw issues to finalized issues - each issue should already have proper path set
	finalizedIssues := make([]core.ZodIssue, len(issues))
	for i, rawIssue := range issues {
		finalizedIssues[i] = FinalizeIssue(rawIssue, core.NewParseContext(), nil)
	}

	return NewZodError(finalizedIssues)
}

// CreateInvalidElementError creates a standardized invalid element error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateInvalidElementError(index int, origin string, input any, elementError error, ctx *core.ParseContext) error {
	// Convert error to raw issue if it's a ZodError
	var elementRawIssue core.ZodRawIssue
	var zodErr *ZodError
	if errors.As(elementError, &zodErr) && len(zodErr.Issues) > 0 {
		elementRawIssue = ConvertZodIssueToRaw(zodErr.Issues[0])
	} else {
		// Create a custom raw issue from the error
		elementRawIssue = CreateCustomIssue(elementError.Error(), nil, input)
	}
	raw := CreateInvalidElementIssue(index, origin, input, elementRawIssue)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateNotMultipleOfError creates a standardized "not multiple of" error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateNotMultipleOfError(divisor any, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateNotMultipleOfIssue(divisor, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidUnionError creates a standardized invalid union error with proper context
// Simplified version that directly returns error instead of requiring multiple steps
func CreateInvalidUnionError(unionErrors []error, input any, ctx *core.ParseContext) error {
	// Convert errors to raw issues
	unionRawIssues := make([]core.ZodRawIssue, len(unionErrors))
	for i, err := range unionErrors {
		var zodErr *ZodError
		if errors.As(err, &zodErr) && len(zodErr.Issues) > 0 {
			unionRawIssues[i] = ConvertZodIssueToRaw(zodErr.Issues[0])
		} else {
			// Create a custom raw issue from the error
			unionRawIssues[i] = CreateCustomIssue(err.Error(), nil, input)
		}
	}
	raw := CreateInvalidUnionIssue(unionRawIssues, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidXorError creates a standardized exclusive union error when multiple options match
// Zod v4 xor semantics: exactly one option must match
// Uses InvalidUnion code with inclusive=false to indicate exclusive union (xor) failure
// See: .reference/zod/packages/zod/src/v4/core/schemas.ts:2185-2192
func CreateInvalidXorError(matchCount int, input any, ctx *core.ParseContext) error {
	properties := map[string]any{
		"errors":      []core.ZodRawIssue{}, // Empty errors array for multiple match case
		"inclusive":   false,                // Zod v4: inclusive=false indicates xor (exclusive union)
		"match_count": matchCount,           // Additional context for debugging
	}
	raw := CreateIssue(core.InvalidUnion, "", properties, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateMissingRequiredError creates a standardized missing required field error with proper context
func CreateMissingRequiredError(fieldName string, fieldType string, input any, ctx *core.ParseContext) error {
	raw := CreateMissingRequiredIssue(fieldName, fieldType)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateTypeConversionError creates a standardized type conversion error with proper context
func CreateTypeConversionError(fromType, toType string, input any, ctx *core.ParseContext) error {
	raw := CreateTypeConversionIssue(fromType, toType, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateIncompatibleTypesError creates a standardized incompatible types error with proper context
func CreateIncompatibleTypesError(conflictType string, value1, value2 any, input any, ctx *core.ParseContext) error {
	raw := CreateIncompatibleTypesIssue(conflictType, value1, value2, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidSchemaError creates an invalid schema error
func CreateInvalidSchemaError(reason string, input any, ctx *core.ParseContext, additionalProps ...map[string]any) error {
	var props map[string]any
	if len(additionalProps) > 0 {
		props = additionalProps[0]
	}
	raw := CreateInvalidSchemaIssue(reason, input, props)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}
