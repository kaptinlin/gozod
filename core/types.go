package core

import (
	"regexp"
)

// =============================================================================
// SCHEMA DEFINITION TYPES
// =============================================================================

// ZodTypeInternals represents internal type state
// Contains all the internal configuration and state for a schema type
type ZodTypeInternals struct {
	Version string                                                       // Library version for compatibility
	Type    string                                                       // Type identifier (e.g., "string", "number")
	Checks  []ZodCheck                                                   // List of validation checks to apply
	Parse   func(payload *ParsePayload, ctx *ParseContext) *ParsePayload // Parse function

	// Core validation flags
	Coerce   bool // Whether to enable type coercion
	Optional bool // Whether the field is optional
	Nilable  bool // Whether nil values are allowed

	// Optionality configuration
	OptIn  string // Optionality mode input
	OptOut string // Optionality mode output

	// Constructor and configuration
	Constructor func(def *ZodTypeDef) ZodType[any, any] // Factory function
	Values      map[any]struct{}                        // Valid values for literal types
	Pattern     *regexp.Regexp                          // Regex pattern for string validation
	Error       *ZodErrorMap                            // Custom error mapping
	Bag         map[string]any                          // Additional configuration storage
}

// ZodTypeDef represents type definition
// Basic configuration for creating schema types
type ZodTypeDef struct {
	Type     string       // Type name
	Coerce   bool         // Enable coercion
	Required bool         // Field is required
	Error    *ZodErrorMap // Custom error handler
	Checks   []ZodCheck   // Validation checks
}

// =============================================================================
// SCHEMA PARAMETER TYPES
// =============================================================================

// SchemaParams contains optional parameters for schema creation
// Used to configure schema behavior during construction
type SchemaParams struct {
	Description   string         // Human-readable description
	Error         any            // Error message or error map (string or ZodErrorMap)
	Abort         bool           // Abort on first validation failure
	Path          []string       // Path for nested validation
	Params        map[string]any // Additional parameters
	UnionFallback bool           // Use as fallback in union types
}

// ApplySchemaParams applies parameters to a check definition
func ApplySchemaParams(def *ZodCheckDef, params ...SchemaParams) {
	for _, param := range params {
		if param.Error != nil {
			// Handle string error messages by converting to ZodErrorMap
			if errStr, ok := param.Error.(string); ok {
				errorMap := ZodErrorMap(func(issue ZodRawIssue) string {
					return errStr
				})
				def.Error = &errorMap
			} else if errorMap, ok := param.Error.(ZodErrorMap); ok {
				def.Error = &errorMap
			} else if errorMapPtr, ok := param.Error.(*ZodErrorMap); ok {
				def.Error = errorMapPtr
			} else if errorFunc, ok := param.Error.(func(ZodRawIssue) string); ok {
				errorMap := ZodErrorMap(errorFunc)
				def.Error = &errorMap
			}
		}
		if param.Abort {
			def.Abort = true
		}
	}
}

// =============================================================================
// PARSE CONTEXT DEFINITIONS
// =============================================================================

// ParseContext contains parsing context information for schema validation
// This structure provides configuration and context for validation operations
type ParseContext struct {
	// Error customizes error messages during validation
	// When provided, this function will be called to generate custom error messages
	Error ZodErrorMap

	// ReportInput includes the input field in issue objects
	// When true, validation issues will include the original input value
	// Default is false to reduce output size
	ReportInput bool

	// Note: jitless field omitted - Go is compiled and doesn't use JIT optimization
	// Unlike JavaScript Zod, Go code is compiled ahead of time
}

// ParsePayload contains value and validation issues during parsing
// This structure is passed through the validation pipeline to collect results
type ParsePayload struct {
	// Value being validated - the current value in the validation pipeline
	Value any

	// Issues collected during validation - accumulates all validation failures
	Issues []ZodRawIssue

	// Path is the current validation path - tracks location in nested structures
	// Each element represents a step in the path (object key, array index, etc.)
	Path []any
}

// =============================================================================
// CONTEXT TYPES
// =============================================================================

// CheckContext provides context for validation checks
// Contains information about the current validation state
type CheckContext struct {
	Value any   // The value being validated
	Path  []any // Path to the current field
}

// RefinementContext provides context for refinement operations
// Used during refinement and transformation operations
type RefinementContext struct {
	*ParseContext                      // Base parse context
	Value         any                  // Value being refined/transformed
	AddIssue      func(issue ZodIssue) // Function to add validation issues
}

// =============================================================================
// PARSE PARAMETERS
// =============================================================================

// ParseParams represents parameters for parse-time configuration
type ParseParams struct {
	Error       ZodErrorMap // Per-parse error customization
	ReportInput bool        // Whether to include input in errors
}

// =============================================================================
// GENERIC TYPE CONSTRAINTS
// =============================================================================

// ZodIntegerConstraint defines the constraint for integer types
// Used in generic integer validation functions and type definitions
type ZodIntegerConstraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// ZodFloatConstraint defines the constraint for float types
// Used in generic float validation functions and type definitions
type ZodFloatConstraint interface {
	~float32 | ~float64
}

// ZodNumericConstraint defines the constraint for all numeric types
// Combines integer and float constraints for comprehensive numeric handling
type ZodNumericConstraint interface {
	ZodIntegerConstraint | ZodFloatConstraint
}
