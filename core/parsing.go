package core

// ParseParams represents parameters for a single parse-time configuration.
type ParseParams struct {
	Error       ZodErrorMap // Per-parse error customization
	ReportInput bool        // Whether to include input in errors
}

// ParseContext contains the configuration and state for a validation run.
type ParseContext struct {
	Error             ZodErrorMap // Custom error message generator
	ReportInput       bool        // Include original input in issues
	IsPrefaultContext bool        // Whether parsing a prefault value
}

// RefinementContext provides context for refinement and transformation operations.
type RefinementContext struct {
	*ParseContext
	Value    any                  // The value being refined/transformed
	AddIssue func(issue ZodIssue) // Function to add validation issues
}

// ParsePayload contains the value and validation issues during parsing.
type ParsePayload struct {
	value  any
	issues []ZodRawIssue
	path   []any
}

// getOrCreateContext ensures a valid ParseContext exists.
func getOrCreateContext(ctx ...*ParseContext) *ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return &ParseContext{}
}
