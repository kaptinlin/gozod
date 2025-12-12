package core

// =============================================================================
// PARSING CONTEXT & PAYLOAD
// =============================================================================
//
// This file contains the core data structures used during the validation
// process. When a schema's .Parse() method is called, a `ParsePayload` is
// created to track the value and any validation issues, while a `ParseContext`
// holds configuration for the parsing run.

// -----------------------------------------------------------------------------
// Parse-Time Configuration
// -----------------------------------------------------------------------------

// ParseParams represents parameters for a single parse-time configuration.
// It allows overriding global or schema-level settings for one validation run.
type ParseParams struct {
	Error       ZodErrorMap // Per-parse error customization
	ReportInput bool        // Whether to include input in errors
}

// ParseContext contains the full configuration and state for a validation run.
// It is passed through the validation pipeline.
type ParseContext struct {
	// Error customizes error messages during validation.
	// When provided, this function will be called to generate custom error messages.
	Error ZodErrorMap

	// ReportInput includes the input field in issue objects.
	// When true, validation issues will include the original input value.
	ReportInput bool

	// IsPrefaultContext indicates if this parsing is for a prefault value
	// When true, validators should allow prefault values to proceed to refinement
	IsPrefaultContext bool
}

// -----------------------------------------------------------------------------
// Refinement Context
// -----------------------------------------------------------------------------

// RefinementContext provides context for refinement and transformation operations.
// It is passed to the user-provided function in `.transform()`.
type RefinementContext struct {
	*ParseContext                      // Embeds the base parse context
	Value         any                  // The value being refined/transformed
	AddIssue      func(issue ZodIssue) // Function to add validation issues
}

// -----------------------------------------------------------------------------
// Parse-Time State
// -----------------------------------------------------------------------------

// ParsePayload contains the value and validation issues during parsing.
// This structure is passed through the validation pipeline to collect results.
type ParsePayload struct {
	// value being validated - the current value in the validation pipeline.
	// Private field - use GetValue()/SetValue() methods for access.
	value any

	// issues collected during validation - accumulates all validation failures.
	// Private field - use GetIssues()/SetIssues() methods for access.
	issues []ZodRawIssue

	// path is the current validation path - tracks location in nested structures.
	// Private field - use GetPath() method for access.
	path []any
}

// -----------------------------------------------------------------------------
// Context Utilities
// -----------------------------------------------------------------------------

// getOrCreateContext ensures a valid ParseContext exists for an operation.
// If one is not provided, it creates a default, empty context.
func getOrCreateContext(ctx ...*ParseContext) *ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return &ParseContext{}
}
