package issues

import (
	"errors"
	"fmt"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/kaptinlin/gozod/pkg/reflectx"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// CreateIssue creates a new ZodRawIssue with safely copied properties.
func CreateIssue(code core.IssueCode, message string, properties map[string]any, input any) core.ZodRawIssue {
	var props map[string]any
	if len(properties) == 0 {
		props = make(map[string]any)
	} else {
		props = mapx.Copy(properties)
	}

	return core.ZodRawIssue{
		Code:       code,
		Message:    message,
		Properties: props,
		Input:      input,
		Path:       []any{},
	}
}

// CreateInvalidTypeIssue creates an invalid type issue.
func CreateInvalidTypeIssue(expected core.ZodTypeCode, input any) core.ZodRawIssue {
	properties := make(map[string]any, 2)
	properties["expected"] = string(expected)
	properties["received"] = string(reflectx.ParsedType(input))

	return CreateIssue(core.InvalidType, "", properties, input)
}

// CreateInvalidValueIssue creates an invalid value issue with deduplicated values.
func CreateInvalidValueIssue(validValues []any, input any) core.ZodRawIssue {
	values := validValues
	if len(validValues) > 1 && len(validValues) <= 10 {
		if uniqueValues, err := slicex.Unique(validValues); err == nil {
			if uniqueSlice, ok := uniqueValues.([]any); ok {
				values = uniqueSlice
			}
		}
	} else if len(validValues) > 10 {
		values = validValues
	}

	properties := make(map[string]any, 1)
	properties["values"] = values

	return CreateIssue(core.InvalidValue, "", properties, input)
}

// CreateTooBigIssue creates a "too big" issue.
func CreateTooBigIssue(maximum any, inclusive bool, origin string, input any) core.ZodRawIssue {
	properties := make(map[string]any, 3)
	properties["maximum"] = maximum
	properties["inclusive"] = inclusive
	properties["origin"] = origin

	return CreateIssue(core.TooBig, "", properties, input)
}

// CreateTooSmallIssue creates a "too small" issue.
func CreateTooSmallIssue(minimum any, inclusive bool, origin string, input any) core.ZodRawIssue {
	properties := make(map[string]any, 3)
	properties["minimum"] = minimum
	properties["inclusive"] = inclusive
	properties["origin"] = origin

	return CreateIssue(core.TooSmall, "", properties, input)
}

// CreateFixedLengthArrayIssue creates a size constraint issue for fixed-length arrays.
// Sets both minimum and maximum to enable "expected exactly N" formatting.
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

// CreateInvalidFormatIssue creates an invalid format issue.
func CreateInvalidFormatIssue(format string, input any, additionalProps map[string]any) core.ZodRawIssue {
	size := 1
	if additionalProps != nil {
		size += len(additionalProps)
	}
	properties := make(map[string]any, size)
	properties["format"] = format

	if additionalProps != nil {
		properties = mapx.Merge(properties, additionalProps)
	}

	return CreateIssue(core.InvalidFormat, "", properties, input)
}

// CreateNotMultipleOfIssue creates a "not multiple of" issue.
func CreateNotMultipleOfIssue(divisor any, origin string, input any) core.ZodRawIssue {
	properties := make(map[string]any, 2)
	properties["divisor"] = divisor
	properties["origin"] = origin

	return CreateIssue(core.NotMultipleOf, "", properties, input)
}

// CreateUnrecognizedKeysIssue creates an unrecognized keys issue with deduplicated keys.
func CreateUnrecognizedKeysIssue(keys []string, input any) core.ZodRawIssue {
	deduped := keys
	if len(keys) > 1 && len(keys) <= 5 {
		if uniqueKeys, err := slicex.Unique(keys); err == nil {
			if uniqueSlice, ok := uniqueKeys.([]string); ok {
				deduped = uniqueSlice
			}
		}
	}

	properties := map[string]any{
		"keys": deduped,
	}

	return CreateIssue(core.UnrecognizedKeys, "", properties, input)
}

// CreateInvalidKeyIssue creates an invalid key issue.
func CreateInvalidKeyIssue(key string, origin string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"key":    key,
		"origin": origin,
	}

	return CreateIssue(core.InvalidKey, "", properties, input)
}

// CreateInvalidUnionIssue creates an invalid union issue.
func CreateInvalidUnionIssue(unionErrors []core.ZodRawIssue, input any) core.ZodRawIssue {
	properties := map[string]any{
		"union_errors": unionErrors,
	}

	if !slicex.IsEmpty(unionErrors) {
		properties["error_count"] = len(unionErrors)
	}

	return CreateIssue(core.InvalidUnion, "", properties, input)
}

// CreateInvalidElementIssue creates an invalid element issue with proper path.
func CreateInvalidElementIssue(index int, origin string, input any, elementError core.ZodRawIssue) core.ZodRawIssue {
	properties := map[string]any{
		"index":         index,
		"origin":        origin,
		"element_error": elementError,
	}

	rawIssue := CreateIssue(core.InvalidElement, "", properties, input)
	rawIssue.Path = []any{index}
	return rawIssue
}

// CreateCustomIssue creates a custom issue.
func CreateCustomIssue(message string, properties map[string]any, input any) core.ZodRawIssue {
	props := mapx.Copy(properties)
	if props == nil {
		props = make(map[string]any)
	}

	return CreateIssue(core.Custom, message, props, input)
}

// CreateMissingRequiredIssue creates a missing required field issue.
func CreateMissingRequiredIssue(fieldName string, fieldType string) core.ZodRawIssue {
	properties := map[string]any{"field_name": fieldName, "field_type": fieldType}
	return CreateIssue(core.MissingRequired, "", properties, nil)
}

// CreateTypeConversionIssue creates a type conversion failure issue.
func CreateTypeConversionIssue(fromType, toType string, input any) core.ZodRawIssue {
	properties := map[string]any{"from_type": fromType, "to_type": toType}
	return CreateIssue(core.TypeConversion, "", properties, input)
}

// CreateInvalidSchemaIssue creates an invalid schema issue.
func CreateInvalidSchemaIssue(reason string, input any, additionalProps ...map[string]any) core.ZodRawIssue {
	properties := map[string]any{
		"reason": reason,
	}

	if len(additionalProps) > 0 && additionalProps[0] != nil {
		properties = mapx.Merge(properties, additionalProps[0])
	}

	return CreateIssue(core.InvalidSchema, "", properties, input)
}

// CreateIncompatibleTypesIssue creates an incompatible types issue.
func CreateIncompatibleTypesIssue(conflictType string, value1, value2 any, input any) core.ZodRawIssue {
	properties := map[string]any{"conflict_type": conflictType, "value1": value1, "value2": value2}
	return CreateIssue(core.IncompatibleTypes, "", properties, input)
}

// CreateMissingKeyIssue creates a missing key issue.
func CreateMissingKeyIssue(key string, options ...func(*core.ZodRawIssue)) core.ZodRawIssue {
	properties := map[string]any{
		"missing_key": key,
		"expected":    "required",
	}

	issue := CreateIssue(core.InvalidType, "", properties, nil)

	for _, option := range options {
		option(&issue)
	}

	return issue
}

// CreateInvalidTypeWithMsg creates an invalid type issue with a custom message.
func CreateInvalidTypeWithMsg(expected core.ZodTypeCode, message string, input any) core.ZodRawIssue {
	properties := map[string]any{
		"expected": string(expected),
		"received": string(reflectx.ParsedType(input)),
	}

	return CreateIssue(core.InvalidType, message, properties, input)
}

// CreateInvalidUnionIssueWithResults creates an invalid union issue with TypeScript-compatible "errors" property.
func CreateInvalidUnionIssueWithResults(unionErrors []core.ZodRawIssue, input any, continueOnError ...bool) core.ZodRawIssue {
	properties := map[string]any{
		"errors": unionErrors,
	}

	if !slicex.IsEmpty(unionErrors) {
		properties["error_count"] = len(unionErrors)
	}

	if len(continueOnError) > 0 && continueOnError[0] {
		properties["continue"] = true
	}

	return CreateIssue(core.InvalidUnion, "", properties, input)
}

// CreateNonOptionalIssue returns an invalid_type issue with expected "nonoptional".
func CreateNonOptionalIssue(input any) core.ZodRawIssue {
	props := map[string]any{
		"expected": "nonoptional",
		"received": string(reflectx.ParsedType(input)),
	}
	return CreateIssue(core.InvalidType, "", props, input)
}

// ConvertZodIssueToRaw converts a ZodIssue to ZodRawIssue.
func ConvertZodIssueToRaw(issue core.ZodIssue) core.ZodRawIssue {
	return core.ZodRawIssue{
		Code:       issue.Code,
		Message:    issue.Message,
		Input:      issue.Input,
		Path:       []any{},
		Properties: make(map[string]any),
	}
}

// ConvertZodIssueToRawWithProperties converts a ZodIssue to ZodRawIssue with essential properties.
// The pathPrefix is set directly as the path (for slice/set where elements have simple index paths).
func ConvertZodIssueToRawWithProperties(issue core.ZodIssue, pathPrefix []any) core.ZodRawIssue {
	return convertZodIssueToRawWithPath(issue, pathPrefix)
}

// ConvertZodIssueToRawWithPrependedPath converts a ZodIssue to ZodRawIssue,
// prepending pathPrefix to the issue's existing path.
func ConvertZodIssueToRawWithPrependedPath(issue core.ZodIssue, pathPrefix []any) core.ZodRawIssue {
	fullPath := make([]any, 0, len(pathPrefix)+len(issue.Path))
	fullPath = append(fullPath, pathPrefix...)
	fullPath = append(fullPath, issue.Path...)
	return convertZodIssueToRawWithPath(issue, fullPath)
}

// convertZodIssueToRawWithPath creates a raw issue with the given path and copies essential properties.
func convertZodIssueToRawWithPath(issue core.ZodIssue, path []any) core.ZodRawIssue {
	rawIssue := core.ZodRawIssue{
		Code:       issue.Code,
		Message:    issue.Message,
		Input:      issue.Input,
		Path:       path,
		Properties: make(map[string]any),
	}

	if issue.Minimum != nil {
		rawIssue.Properties["minimum"] = issue.Minimum
	}
	if issue.Maximum != nil {
		rawIssue.Properties["maximum"] = issue.Maximum
	}
	if issue.Expected != "" {
		rawIssue.Properties["expected"] = issue.Expected
	}
	if issue.Received != "" {
		rawIssue.Properties["received"] = issue.Received
	}
	rawIssue.Properties["inclusive"] = issue.Inclusive

	return rawIssue
}

// CreateErrorMap creates an error map from various input types.
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
		if str := fmt.Sprintf("%v", e); str != "" {
			errorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
				return str
			})
			return &errorMap
		}
	}

	return nil
}

// CreateFinalError creates a finalized ZodError from issue parameters.
func CreateFinalError(code core.IssueCode, message string, properties map[string]any, input any, ctx *core.ParseContext, config *core.ZodConfig) error {
	raw := CreateIssue(code, message, properties, input)
	final := FinalizeIssue(raw, ctx, config)
	return NewZodError([]core.ZodIssue{final})
}

// CreateNonOptionalError creates a non-optional error with proper context.
func CreateNonOptionalError(ctx *core.ParseContext) error {
	raw := CreateNonOptionalIssue(nil)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidTypeError creates an invalid type error with proper context.
func CreateInvalidTypeError(expectedType core.ZodTypeCode, input any, ctx *core.ParseContext) error {
	raw := CreateInvalidTypeIssue(expectedType, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidTypeErrorWithInst creates an invalid type error with schema internals.
func CreateInvalidTypeErrorWithInst(expectedType core.ZodTypeCode, input any, ctx *core.ParseContext, inst any) error {
	raw := CreateInvalidTypeIssue(expectedType, input)
	raw.Inst = inst
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidValueError creates an invalid value error with proper context.
func CreateInvalidValueError(validValues []any, input any, ctx *core.ParseContext) error {
	raw := CreateInvalidValueIssue(validValues, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateTooBigError creates a "too big" error with proper context.
func CreateTooBigError(maximum any, inclusive bool, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateTooBigIssue(maximum, inclusive, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateTooSmallError creates a "too small" error with proper context.
func CreateTooSmallError(minimum any, inclusive bool, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateTooSmallIssue(minimum, inclusive, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateRestParameterTooSmallError creates a TooSmall error for rest parameter arrays.
func CreateRestParameterTooSmallError(minimum any, inclusive bool, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateTooSmallIssue(minimum, inclusive, origin, input)
	raw.Properties["is_rest_param"] = true
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateFixedLengthArrayError creates an error for fixed-length array validation.
func CreateFixedLengthArrayError(expectedLength any, actualLength int, input any, isTooSmall bool, ctx *core.ParseContext) error {
	raw := CreateFixedLengthArrayIssue(expectedLength, actualLength, input, isTooSmall)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidFormatError creates an invalid format error with proper context.
func CreateInvalidFormatError(format string, input any, ctx *core.ParseContext, additionalProps ...map[string]any) error {
	var props map[string]any
	if len(additionalProps) > 0 {
		props = additionalProps[0]
	}
	raw := CreateInvalidFormatIssue(format, input, props)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateCustomError creates a custom error with proper context.
func CreateCustomError(message string, properties map[string]any, input any, ctx *core.ParseContext) error {
	raw := CreateCustomIssue(message, properties, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateUnrecognizedKeysError creates an unrecognized keys error with proper context.
func CreateUnrecognizedKeysError(keys []string, input any, ctx *core.ParseContext) error {
	raw := CreateUnrecognizedKeysIssue(keys, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidKeyError creates an invalid key error with proper context.
func CreateInvalidKeyError(key string, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateInvalidKeyIssue(key, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateElementValidationIssue creates a raw issue for invalid element validation.
func CreateElementValidationIssue(index int, origin string, element any, elementError error) core.ZodRawIssue {
	var raw core.ZodRawIssue
	var zodErr *ZodError
	if errors.As(elementError, &zodErr) && len(zodErr.Issues) > 0 {
		raw = ConvertZodIssueToRaw(zodErr.Issues[0])
	} else {
		raw = CreateIssue(core.InvalidElement, elementError.Error(), nil, element)
	}

	return CreateInvalidElementIssue(index, origin, element, raw)
}

// CreateArrayValidationIssues creates a ZodError from multiple array validation issues.
func CreateArrayValidationIssues(issues []core.ZodRawIssue) error {
	if len(issues) == 0 {
		return nil
	}

	finalizedIssues := make([]core.ZodIssue, len(issues))
	for i, rawIssue := range issues {
		finalizedIssues[i] = FinalizeIssue(rawIssue, core.NewParseContext(), nil)
	}

	return NewZodError(finalizedIssues)
}

// CreateInvalidElementError creates an invalid element error with proper context.
func CreateInvalidElementError(index int, origin string, input any, elementError error, ctx *core.ParseContext) error {
	var raw core.ZodRawIssue
	var zodErr *ZodError
	if errors.As(elementError, &zodErr) && len(zodErr.Issues) > 0 {
		raw = ConvertZodIssueToRaw(zodErr.Issues[0])
	} else {
		raw = CreateCustomIssue(elementError.Error(), nil, input)
	}
	issue := CreateInvalidElementIssue(index, origin, input, raw)
	final := FinalizeIssue(issue, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateNotMultipleOfError creates a "not multiple of" error with proper context.
func CreateNotMultipleOfError(divisor any, origin string, input any, ctx *core.ParseContext) error {
	raw := CreateNotMultipleOfIssue(divisor, origin, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidUnionError creates an invalid union error with proper context.
func CreateInvalidUnionError(unionErrors []error, input any, ctx *core.ParseContext) error {
	raws := make([]core.ZodRawIssue, len(unionErrors))
	for i, err := range unionErrors {
		var zodErr *ZodError
		if errors.As(err, &zodErr) && len(zodErr.Issues) > 0 {
			raws[i] = ConvertZodIssueToRaw(zodErr.Issues[0])
		} else {
			raws[i] = CreateCustomIssue(err.Error(), nil, input)
		}
	}
	raw := CreateInvalidUnionIssue(raws, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidXorError creates an exclusive union error when multiple options match.
// Uses InvalidUnion code with inclusive=false to indicate xor failure.
// See: .reference/zod/packages/zod/src/v4/core/schemas.ts:2185-2192
func CreateInvalidXorError(matchCount int, input any, ctx *core.ParseContext) error {
	properties := map[string]any{
		"errors":      []core.ZodRawIssue{},
		"inclusive":   false,
		"match_count": matchCount,
	}
	raw := CreateIssue(core.InvalidUnion, "", properties, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateMissingRequiredError creates a missing required field error with proper context.
func CreateMissingRequiredError(fieldName string, fieldType string, input any, ctx *core.ParseContext) error {
	raw := CreateMissingRequiredIssue(fieldName, fieldType)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateTypeConversionError creates a type conversion error with proper context.
func CreateTypeConversionError(fromType, toType string, input any, ctx *core.ParseContext) error {
	raw := CreateTypeConversionIssue(fromType, toType, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateIncompatibleTypesError creates an incompatible types error with proper context.
func CreateIncompatibleTypesError(conflictType string, value1, value2 any, input any, ctx *core.ParseContext) error {
	raw := CreateIncompatibleTypesIssue(conflictType, value1, value2, input)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}

// CreateInvalidSchemaError creates an invalid schema error with proper context.
func CreateInvalidSchemaError(reason string, input any, ctx *core.ParseContext, additionalProps ...map[string]any) error {
	var props map[string]any
	if len(additionalProps) > 0 {
		props = additionalProps[0]
	}
	raw := CreateInvalidSchemaIssue(reason, input, props)
	final := FinalizeIssue(raw, ctx, nil)
	return NewZodError([]core.ZodIssue{final})
}
