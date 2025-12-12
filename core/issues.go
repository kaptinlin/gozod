package core

import (
	"fmt"
)

// =============================================================================
// ERROR HANDLING TYPES
// =============================================================================

// ZodErrorMap represents a function that maps raw issues to error messages
// Allows customization of error messages based on validation context
type ZodErrorMap func(ZodRawIssue) string

// =============================================================================
// ISSUE BASE TYPES
// =============================================================================

// ZodIssueBase represents the base structure for all validation issues
// Contains common fields that all validation issues share
type ZodIssueBase struct {
	Code    IssueCode `json:"code,omitempty"`  // Issue type identifier
	Input   any       `json:"input,omitempty"` // The input that caused the issue
	Path    []any     `json:"path"`            // Path to the problematic field
	Message string    `json:"message"`         // Human-readable error message
}

// ZodRawIssue represents a raw issue before finalization
// Used internally during validation before converting to final ZodIssue
type ZodRawIssue struct {
	Code       IssueCode      `json:"code"`                 // Issue type code
	Input      any            `json:"input,omitempty"`      // Input value that failed
	Path       []any          `json:"path,omitempty"`       // Field path in nested structures
	Message    string         `json:"message,omitempty"`    // Error message
	Properties map[string]any `json:"properties,omitempty"` // Additional issue properties
	Continue   bool           `json:"-"`                    // Whether to continue validation
	Inst       any            `json:"-"`                    // Instance that generated the issue
}

// ZodIssue represents a finalized validation issue
// This is the final form of validation issues after processing
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

// =============================================================================
// SPECIALIZED ISSUE TYPES
// =============================================================================

// ZodIssueInvalidType represents invalid type error
// Used when the input type doesn't match the expected type
type ZodIssueInvalidType struct {
	ZodIssueBase
	Expected ZodTypeCode `json:"expected"` // Expected type name
	Received ZodTypeCode `json:"received"` // Actual type name
}

// ZodIssueInvalidValue represents invalid value error
// Used when the input value is not in the allowed set
type ZodIssueInvalidValue struct {
	ZodIssueBase
	Options []any `json:"options"` // List of valid options
}

// =============================================================================
// ZODISSUE METHODS
// =============================================================================

// GetMinimum returns the minimum value if present
func (z *ZodIssue) GetMinimum() (any, bool) {
	return z.Minimum, z.Minimum != nil
}

// GetMaximum returns the maximum value if present
func (z *ZodIssue) GetMaximum() (any, bool) {
	return z.Maximum, z.Maximum != nil
}

// Error implements the error interface for ZodIssue
func (z ZodIssue) Error() string {
	return z.Message
}

// String returns string representation of ZodIssue for debugging
func (z ZodIssue) String() string {
	return fmt.Sprintf("ZodIssue{Code: %s, Message: %s, Path: %v}", z.Code, z.Message, z.Path)
}

// GetExpected returns the expected type for invalid_type issues
func (i ZodIssue) GetExpected() (ZodTypeCode, bool) {
	if i.Code != InvalidType {
		return "", false
	}
	return i.Expected, i.Expected != ""
}

// GetReceived returns the received type for invalid_type issues
func (i ZodIssue) GetReceived() (ZodTypeCode, bool) {
	if i.Code != InvalidType {
		return "", false
	}
	return i.Received, i.Received != ""
}

// GetFormat returns the format for invalid_format issues
func (i ZodIssue) GetFormat() (string, bool) {
	if i.Code != InvalidFormat {
		return "", false
	}
	return i.Format, i.Format != ""
}

// GetDivisor returns the divisor for not_multiple_of issues
func (i ZodIssue) GetDivisor() (any, bool) {
	if i.Code != NotMultipleOf {
		return nil, false
	}
	return i.Divisor, i.Divisor != nil
}

// =============================================================================
// ZODRAWISSUE ACCESSOR METHODS
// =============================================================================

// getStringProperty safely gets a string property from the properties map
func getStringProperty(properties map[string]any, key string) string {
	if properties == nil {
		return ""
	}
	if value, ok := properties[key].(string); ok {
		return value
	}
	return ""
}

// getZodTypeCodeProperty safely gets a ZodTypeCode property from the properties map
func getZodTypeCodeProperty(properties map[string]any, key string) ZodTypeCode {
	if properties == nil {
		return ""
	}
	if value, ok := properties[key].(ZodTypeCode); ok {
		return value
	}
	if value, ok := properties[key].(string); ok {
		return ZodTypeCode(value)
	}
	return ""
}

// GetExpected returns the expected value from properties map
func (r ZodRawIssue) GetExpected() ZodTypeCode {
	return getZodTypeCodeProperty(r.Properties, "expected")
}

// GetReceived returns the received value from properties map
func (r ZodRawIssue) GetReceived() ZodTypeCode {
	return getZodTypeCodeProperty(r.Properties, "received")
}

// GetOrigin returns the origin value from properties map
func (r ZodRawIssue) GetOrigin() string {
	return getStringProperty(r.Properties, "origin")
}

// GetFormat returns the format value from properties map
func (r ZodRawIssue) GetFormat() string {
	return getStringProperty(r.Properties, "format")
}

// GetPattern returns the pattern value from properties map
func (r ZodRawIssue) GetPattern() string {
	return getStringProperty(r.Properties, "pattern")
}

// GetPrefix returns the prefix value from properties map
func (r ZodRawIssue) GetPrefix() string {
	return getStringProperty(r.Properties, "prefix")
}

// GetSuffix returns the suffix value from properties map
func (r ZodRawIssue) GetSuffix() string {
	return getStringProperty(r.Properties, "suffix")
}

// GetIncludes returns the includes value from properties map
func (r ZodRawIssue) GetIncludes() string {
	return getStringProperty(r.Properties, "includes")
}

// getProperty safely gets a property from the properties map
func (r ZodRawIssue) getProperty(key string) any {
	if r.Properties == nil {
		return nil
	}
	return r.Properties[key]
}

// GetMinimum returns the minimum value from properties map
func (r ZodRawIssue) GetMinimum() any { return r.getProperty("minimum") }

// GetMaximum returns the maximum value from properties map
func (r ZodRawIssue) GetMaximum() any { return r.getProperty("maximum") }

// GetInclusive returns the inclusive value from properties map
func (r ZodRawIssue) GetInclusive() bool {
	if val, ok := r.getProperty("inclusive").(bool); ok {
		return val
	}
	return false
}

// GetDivisor returns the divisor value from properties map
func (r ZodRawIssue) GetDivisor() any { return r.getProperty("divisor") }

// GetKeys returns the keys value from properties map
func (r ZodRawIssue) GetKeys() []string {
	if val, ok := r.getProperty("keys").([]string); ok {
		return val
	}
	return nil
}

// GetValues returns the values from properties map
func (r ZodRawIssue) GetValues() []any {
	if val, ok := r.getProperty("values").([]any); ok {
		return val
	}
	return nil
}
