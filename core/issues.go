package core

import (
	"fmt"
)

// ZodErrorMap represents a function that maps raw issues to error messages.
type ZodErrorMap func(ZodRawIssue) string

// ZodIssueBase represents the base structure for all validation issues.
type ZodIssueBase struct {
	Code    IssueCode `json:"code,omitempty"`  // Issue type identifier
	Input   any       `json:"input,omitempty"` // The input that caused the issue
	Path    []any     `json:"path"`            // Path to the problematic field
	Message string    `json:"message"`         // Human-readable error message
}

// ZodRawIssue represents a raw issue before finalization.
type ZodRawIssue struct {
	Code       IssueCode      `json:"code"`                 // Issue type code
	Input      any            `json:"input,omitempty"`      // Input value that failed
	Path       []any          `json:"path,omitempty"`       // Field path in nested structures
	Message    string         `json:"message,omitempty"`    // Error message
	Properties map[string]any `json:"properties,omitempty"` // Additional issue properties
	Continue   bool           `json:"-"`                    // Whether to continue validation
	Inst       any            `json:"-"`                    // Instance that generated the issue
}

// ZodIssue represents a finalized validation issue.
type ZodIssue struct {
	ZodIssueBase
	Expected  ZodTypeCode    `json:"expected,omitempty"`  // Expected type or value
	Received  ZodTypeCode    `json:"received,omitempty"`  // Actual type or value
	Minimum   any            `json:"minimum,omitempty"`   // Minimum value for range checks
	Maximum   any            `json:"maximum,omitempty"`   // Maximum value for range checks
	Inclusive bool           `json:"inclusive,omitempty"` // Whether range bounds are inclusive
	Keys      []string       `json:"keys,omitempty"`      // Keys for unrecognized_keys errors
	Options   []any          `json:"options,omitempty"`   // Valid options for literal errors
	Errors    [][]ZodIssue   `json:"errors,omitempty"`    // Nested errors for union types
	Issues    []ZodIssue     `json:"issues,omitempty"`    // Sub-issues for complex validations
	Format    string         `json:"format,omitempty"`    // Expected format for format validation
	Divisor   any            `json:"divisor,omitempty"`   // Divisor for multiple_of validation
	Pattern   string         `json:"pattern,omitempty"`   // Regex pattern for string validation
	Includes  string         `json:"includes,omitempty"`  // Substring for includes validation
	Prefix    string         `json:"prefix,omitempty"`    // Prefix for starts_with validation
	Suffix    string         `json:"suffix,omitempty"`    // Suffix for ends_with validation
	Values    []any          `json:"values,omitempty"`    // Valid values for enum validation
	Algorithm string         `json:"algorithm,omitempty"` // Algorithm for JWT validation
	Origin    string         `json:"origin,omitempty"`    // Origin type for size validation
	Key       any            `json:"key,omitempty"`       // Key for invalid element errors
	Params    map[string]any `json:"params,omitempty"`    // Custom parameters for validation
}

// ZodIssueInvalidType represents an invalid type error.
type ZodIssueInvalidType struct {
	ZodIssueBase
	Expected ZodTypeCode `json:"expected"` // Expected type name
	Received ZodTypeCode `json:"received"` // Actual type name
}

// ZodIssueInvalidValue represents an invalid value error.
type ZodIssueInvalidValue struct {
	ZodIssueBase
	Options []any `json:"options"` // List of valid options
}

// MinValue returns the minimum value and whether it is present.
func (z *ZodIssue) MinValue() (any, bool) {
	return z.Minimum, z.Minimum != nil
}

// MaxValue returns the maximum value and whether it is present.
func (z *ZodIssue) MaxValue() (any, bool) {
	return z.Maximum, z.Maximum != nil
}

// Error implements the error interface.
func (z ZodIssue) Error() string {
	return z.Message
}

// String returns a string representation for debugging.
func (z ZodIssue) String() string {
	return fmt.Sprintf("ZodIssue{Code: %s, Message: %s, Path: %v}",
		z.Code, z.Message, z.Path)
}

// ExpectedType returns the expected type for invalid_type issues.
func (z ZodIssue) ExpectedType() (ZodTypeCode, bool) {
	if z.Code != InvalidType {
		return "", false
	}
	return z.Expected, z.Expected != ""
}

// ReceivedType returns the received type for invalid_type issues.
func (z ZodIssue) ReceivedType() (ZodTypeCode, bool) {
	if z.Code != InvalidType {
		return "", false
	}
	return z.Received, z.Received != ""
}

// FormatName returns the format for invalid_format issues.
func (z ZodIssue) FormatName() (string, bool) {
	if z.Code != InvalidFormat {
		return "", false
	}
	return z.Format, z.Format != ""
}

// DivisorValue returns the divisor for not_multiple_of issues.
func (z ZodIssue) DivisorValue() (any, bool) {
	if z.Code != NotMultipleOf {
		return nil, false
	}
	return z.Divisor, z.Divisor != nil
}

// typedProperty safely retrieves a typed value from the properties map.
// This is an unexported helper function for internal use.
func typedProperty[T any](properties map[string]any, key string) (val T) {
	if properties == nil {
		return val
	}
	v, _ := properties[key].(T)
	return v
}

// zodTypeCodeProperty safely gets a ZodTypeCode from the properties map.
// It accepts both ZodTypeCode and plain string values.
// This is an unexported helper function for internal use.
func zodTypeCodeProperty(properties map[string]any, key string) ZodTypeCode {
	if properties == nil {
		return ""
	}
	switch v := properties[key].(type) {
	case ZodTypeCode:
		return v
	case string:
		return ZodTypeCode(v)
	default:
		return ""
	}
}

// property safely gets a property from the properties map.
// This is an unexported helper method for internal use.
func (r ZodRawIssue) property(key string) any {
	if r.Properties == nil {
		return nil
	}
	return r.Properties[key]
}

// Expected returns the expected value from properties.
func (r ZodRawIssue) Expected() ZodTypeCode {
	return zodTypeCodeProperty(r.Properties, "expected")
}

// Received returns the received value from properties.
func (r ZodRawIssue) Received() ZodTypeCode {
	return zodTypeCodeProperty(r.Properties, "received")
}

// Origin returns the origin value from properties.
func (r ZodRawIssue) Origin() string {
	return typedProperty[string](r.Properties, "origin")
}

// Format returns the format value from properties.
func (r ZodRawIssue) Format() string {
	return typedProperty[string](r.Properties, "format")
}

// Pattern returns the pattern value from properties.
func (r ZodRawIssue) Pattern() string {
	return typedProperty[string](r.Properties, "pattern")
}

// Prefix returns the prefix value from properties.
func (r ZodRawIssue) Prefix() string {
	return typedProperty[string](r.Properties, "prefix")
}

// Suffix returns the suffix value from properties.
func (r ZodRawIssue) Suffix() string {
	return typedProperty[string](r.Properties, "suffix")
}

// Includes returns the includes value from properties.
func (r ZodRawIssue) Includes() string {
	return typedProperty[string](r.Properties, "includes")
}

// Minimum returns the minimum value from properties.
func (r ZodRawIssue) Minimum() any { return r.property("minimum") }

// Maximum returns the maximum value from properties.
func (r ZodRawIssue) Maximum() any { return r.property("maximum") }

// Inclusive returns the inclusive flag from properties.
func (r ZodRawIssue) Inclusive() bool {
	return typedProperty[bool](r.Properties, "inclusive")
}

// Divisor returns the divisor value from properties.
func (r ZodRawIssue) Divisor() any { return r.property("divisor") }

// Keys returns the keys from properties.
func (r ZodRawIssue) Keys() []string {
	return typedProperty[[]string](r.Properties, "keys")
}

// Values returns the values from properties.
func (r ZodRawIssue) Values() []any {
	return typedProperty[[]any](r.Properties, "values")
}
