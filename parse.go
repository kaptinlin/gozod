package gozod

import (
	"fmt"
	"math/big"
	"reflect"
)

// =============================================================================
// CORE TYPES
// =============================================================================

// ParseContext contains parsing context information for schema validation
type ParseContext struct {
	// Error customizes error messages
	Error ZodErrorMap
	// ReportInput includes the input field in issue objects, default false
	ReportInput bool
	// Note: jitless field omitted - Go is compiled and doesn't use JIT optimization
}

// ParsePayload contains value and validation issues during parsing
type ParsePayload struct {
	// Value being validated
	Value interface{}
	// Issues collected during validation
	Issues []ZodRawIssue
	// Path is the current validation path
	Path []interface{}
}

// =============================================================================
// CORE PARSING FUNCTIONS
// =============================================================================

// Parse validates a value against a schema and returns result with error
func Parse[In, Out any](schema ZodType[In, Out], value any, ctx *ParseContext) (Out, error) {
	// Create initial payload
	payload := &ParsePayload{
		Value:  value,
		Issues: make([]ZodRawIssue, 0),
	}

	// Run schema validation
	internals := schema.GetInternals()
	if internals == nil {
		var zero Out
		return zero, ErrSchemaNoInternals
	}
	if internals.Parse == nil {
		var zero Out
		return zero, ErrSchemaNoParseFunction
	}

	// Parse the value
	result := internals.Parse(payload, ctx)
	if result == nil {
		var zero Out
		return zero, ErrParseReturnedNil
	}

	// Note: Checks are already executed in type-specific parse functions
	// This avoids duplicate check execution

	// Note: No Promise check needed since Go is synchronous

	// Check for validation issues
	if len(result.Issues) > 0 {
		config := GetConfig()
		finalizedIssues := make([]ZodIssue, len(result.Issues))
		for i, rawIssue := range result.Issues {
			finalizedIssues[i] = FinalizeIssue(rawIssue, ctx, config)
		}
		var zero Out
		return zero, NewZodError(finalizedIssues)
	}

	// Return the validated value with type assertion
	// Type-safe result conversion: ensure return value matches Out type
	if result.Value == nil {
		var zero Out
		return zero, nil
	}

	// Safe type assertion with better error information
	if converted, ok := result.Value.(Out); ok {
		return converted, nil
	}

	// If type assertion fails, return zero value and detailed error
	var zero Out
	return zero, fmt.Errorf("%w: expected %T, got %T", ErrTypeAssertionFailed, zero, result.Value)
}

// MustParse validates a value against a schema and panics on failure
func MustParse[In, Out any](schema ZodType[In, Out], value any, ctx *ParseContext) Out {
	result, err := Parse[In, Out](schema, value, ctx)
	if err != nil {
		// Provide better panic information for debugging
		panic(fmt.Sprintf("MustParse failed: %v", err))
	}
	return result
}

// shouldCoerce checks if coercion is enabled in the Bag
// This reduces repetitive coercion checking code across type files
func shouldCoerce(bag map[string]interface{}) bool {
	if bag == nil {
		return false
	}
	if coerceFlag, exists := bag["coerce"].(bool); exists {
		return coerceFlag
	}
	return false
}

// tryApplyCoercion attempts to apply coercion to input if enabled
// Elegant coercion interface design: each type handles its own coerce logic
// Avoids large switch statements in core parsing layer, follows open-closed principle
func tryApplyCoercion(schema ZodType[any, any], input any) (any, error) {
	internals := schema.GetInternals()
	if internals == nil {
		return input, nil
	}

	// Check if coerce is enabled - use unified shouldCoerce function
	if !shouldCoerce(internals.Bag) {
		return input, nil
	}

	// Elegant interface call - types handle their own coerce logic
	if coercible, ok := schema.(Coercible); ok {
		if coercedValue, success := coercible.Coerce(input); success {
			return coercedValue, nil
		}
	}

	// coerce failed or not implemented, return original value for subsequent validation
	return input, nil
}

///////////////////////////
////   HELPER FUNCTIONS ////
///////////////////////////

// NewParseContext creates a new parse context with default values
func NewParseContext() *ParseContext {
	return &ParseContext{
		Error:       nil,
		ReportInput: false,
	}
}

// NewParsePayload creates a new parsing payload with given value
func NewParsePayload(value interface{}) *ParsePayload {
	return &ParsePayload{
		Value:  value,
		Issues: make([]ZodRawIssue, 0),
	}
}

// AddIssue adds a validation issue to the payload
func (p *ParsePayload) AddIssue(issue ZodRawIssue) {
	p.Issues = append(p.Issues, issue)
}

// HasIssues checks if the payload has any validation issues
func (p *ParsePayload) HasIssues() bool {
	return len(p.Issues) > 0
}

//////////////////////////////////////
////   parseType TEMPLATE SYSTEM ////
//////////////////////////////////////

// parseType provides unified type parsing logic with smart type inference
// T: target type (e.g. string, []interface{}, map[interface{}]interface{})
func parseType[T any](
	input any,
	internals *ZodTypeInternals,
	expectedType string,
	typeChecker func(any) (T, bool),
	pointerChecker func(any) (*T, bool),
	validator func(T, []ZodCheck, *ParseContext) error,
	coercer func(any) (T, bool),
	ctx *ParseContext,
) (any, error) {
	// 1. Unified nil handling
	if input == nil {
		if !internals.Nilable {
			rawIssue := CreateInvalidTypeIssue(input, expectedType, "null")
			finalIssue := FinalizeIssue(rawIssue, ctx, GetConfig())
			return nil, NewZodError([]ZodIssue{finalIssue})
		}
		// For pointer types, return nil; for non-pointer types, return (*T)(nil)
		var zero T
		if reflect.TypeOf(zero).Kind() == reflect.Ptr {
			return zero, nil
		}
		return (*T)(nil), nil
	}

	// 2. Try type coercion (if enabled) - prioritize over pointer type inference
	if shouldCoerce(internals.Bag) {
		// First try direct coercion
		if coerced, ok := coercer(input); ok {
			if err := validator(coerced, internals.Checks, ctx); err != nil {
				return nil, err
			}
			return coerced, nil // coercion always returns base type
		}
		// Try dereferencing pointer then coercion (but don't handle nil pointers)
		if ptr, ok := pointerChecker(input); ok && ptr != nil {
			if coerced, ok := coercer(*ptr); ok {
				if err := validator(coerced, internals.Checks, ctx); err != nil {
					return nil, err
				}
				return coerced, nil // coercion always returns base type
			}
		}
		// If coercion fails, continue to next step (don't return error directly)
	}

	// 3. Smart type inference: check pointer type matching
	if ptr, ok := pointerChecker(input); ok {
		if ptr == nil {
			if !internals.Nilable {
				rawIssue := CreateInvalidTypeIssue(input, expectedType, "null")
				finalIssue := FinalizeIssue(rawIssue, ctx, GetConfig())
				return nil, NewZodError([]ZodIssue{finalIssue})
			}
			// For pointer types, return nil; for non-pointer types, return (*T)(nil)
			var zero T
			if reflect.TypeOf(zero).Kind() == reflect.Ptr {
				return zero, nil
			}
			return (*T)(nil), nil
		}
		if err := validator(*ptr, internals.Checks, ctx); err != nil {
			return nil, err
		}
		// Return same pointer, no dereferencing
		return ptr, nil // *T → *T (keep same pointer)
	}

	// 4. Smart type inference: check direct type matching
	if value, ok := typeChecker(input); ok {
		if err := validator(value, internals.Checks, ctx); err != nil {
			return nil, err
		}
		return value, nil // T → T (keep original type)
	}

	// 5. Unified error creation
	rawIssue := CreateInvalidTypeIssue(input, expectedType, string(GetParsedType(input)))
	finalIssue := FinalizeIssue(rawIssue, ctx, GetConfig())
	return nil, NewZodError([]ZodIssue{finalIssue})
}

//////////////////////////////////////
////   TYPE-SPECIFIC VALIDATORS   ////
//////////////////////////////////////

// validateString string validator
func validateString(value string, checks []ZodCheck, ctx *ParseContext) error {
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}
	return nil
}

// validateBool boolean validator
func validateBool(value bool, checks []ZodCheck, ctx *ParseContext) error {
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}
	return nil
}

// validateInteger integer validator (generic)
func validateInteger[T ZodIntegerConstraint](value T, checks []ZodCheck, ctx *ParseContext) error {
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}
	return nil
}

// validateFloat float validator (generic)
func validateFloat[T ZodFloatConstraint](value T, checks []ZodCheck, ctx *ParseContext) error {
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}
	return nil
}

// validateBigInt big integer validator
func validateBigInt(value *big.Int, checks []ZodCheck, ctx *ParseContext) error {
	if len(checks) > 0 {
		payload := &ParsePayload{
			Value:  value,
			Issues: make([]ZodRawIssue, 0),
		}
		runChecksOnValue(value, checks, payload, ctx)
		if len(payload.Issues) > 0 {
			return &ZodError{Issues: convertRawIssuesToIssues(payload.Issues, ctx)}
		}
	}
	return nil
}
