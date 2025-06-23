package issues

import (
	"github.com/kaptinlin/gozod/core"
)

// =============================================================================
// RE-EXPORTED TYPES FROM CORE
// =============================================================================

// IssueCode represents validation issue types - re-exported from core
type IssueCode = core.IssueCode

// Re-export issue code constants for convenient access
const (
	InvalidType      = core.InvalidType
	InvalidValue     = core.InvalidValue
	InvalidFormat    = core.InvalidFormat
	InvalidUnion     = core.InvalidUnion
	InvalidKey       = core.InvalidKey
	InvalidElement   = core.InvalidElement
	TooBig           = core.TooBig
	TooSmall         = core.TooSmall
	NotMultipleOf    = core.NotMultipleOf
	UnrecognizedKeys = core.UnrecognizedKeys
	Custom           = core.Custom
)

// Core issue types
type ParseParams = core.ParseParams
type ZodErrorMap = core.ZodErrorMap
type ZodIssueBase = core.ZodIssueBase
type ZodRawIssue = core.ZodRawIssue
type ZodIssue = core.ZodIssue

// =============================================================================
// ISSUE SUBTYPES
// =============================================================================

// ZodIssueInvalidType represents a type validation error
type ZodIssueInvalidType struct {
	ZodIssueBase
	// TODO: consider switching to core.ParsedType for better type-safety
	Expected core.ZodTypeCode `json:"expected"`
	Received core.ZodTypeCode `json:"received"`
}

// ZodIssueTooBig represents a value exceeding maximum constraint error
type ZodIssueTooBig struct {
	ZodIssueBase
	Origin    string `json:"origin"`
	Maximum   any    `json:"maximum"`
	Inclusive bool   `json:"inclusive,omitempty"`
}

// ZodIssueTooSmall represents a value below minimum constraint error
type ZodIssueTooSmall struct {
	ZodIssueBase
	Origin    string `json:"origin"`
	Minimum   any    `json:"minimum"`
	Inclusive bool   `json:"inclusive,omitempty"`
}

// ZodIssueInvalidStringFormat represents an invalid string format error
type ZodIssueInvalidStringFormat struct {
	ZodIssueBase
	Format  string `json:"format"`
	Pattern string `json:"pattern,omitempty"`
}

// ZodIssueNotMultipleOf represents a value not being a multiple of expected divisor
type ZodIssueNotMultipleOf struct {
	ZodIssueBase
	Divisor any `json:"divisor"`
}

// ZodIssueUnrecognizedKeys represents unrecognized object keys error
type ZodIssueUnrecognizedKeys struct {
	ZodIssueBase
	Keys []string `json:"keys"`
}

// ZodIssueInvalidUnion represents failure to match any union schemas
type ZodIssueInvalidUnion struct {
	ZodIssueBase
	Errors [][]ZodIssue `json:"errors"`
}

// ZodIssueInvalidKey represents invalid key in a map or record
type ZodIssueInvalidKey struct {
	ZodIssueBase
	Origin string     `json:"origin"`
	Issues []ZodIssue `json:"issues"`
}

// ZodIssueInvalidElement represents invalid element in a collection
type ZodIssueInvalidElement struct {
	ZodIssueBase
	Origin string     `json:"origin"`
	Key    any        `json:"key"`
	Issues []ZodIssue `json:"issues"`
}

// ZodIssueInvalidValue represents a value not matching expected values
type ZodIssueInvalidValue struct {
	ZodIssueBase
	Values []any `json:"values"`
}

// ZodIssueCustom represents a custom validation error
type ZodIssueCustom struct {
	ZodIssueBase
	Params map[string]any `json:"params,omitempty"`
}

// =============================================================================
// STRING FORMAT ISSUES
// =============================================================================

// ZodIssueStringCommonFormats represents common string format validation errors
type ZodIssueStringCommonFormats struct {
	ZodIssueInvalidStringFormat
}

// ZodIssueStringInvalidRegex represents regex pattern validation error
type ZodIssueStringInvalidRegex struct {
	ZodIssueInvalidStringFormat
	Pattern string `json:"pattern"`
}

// ZodIssueStringInvalidJWT represents JWT validation error
type ZodIssueStringInvalidJWT struct {
	ZodIssueInvalidStringFormat
	Algorithm string `json:"algorithm,omitempty"`
}

// ZodIssueStringStartsWith represents string prefix validation error
type ZodIssueStringStartsWith struct {
	ZodIssueInvalidStringFormat
	Prefix string `json:"prefix"`
}

// ZodIssueStringEndsWith represents string suffix validation error
type ZodIssueStringEndsWith struct {
	ZodIssueInvalidStringFormat
	Suffix string `json:"suffix"`
}

// ZodIssueStringIncludes represents string inclusion validation error
type ZodIssueStringIncludes struct {
	ZodIssueInvalidStringFormat
	Includes string `json:"includes"`
}
