package core

// =============================================================================
// CONTEXT CREATION FUNCTIONS
// =============================================================================

// NewParseContext creates a new parse context with default values
// Provides sensible defaults for typical validation scenarios
func NewParseContext() *ParseContext {
	return &ParseContext{
		Error:       nil,   // Use default error messages
		ReportInput: false, // Don't include input in issues by default
	}
}

// NewParsePayload creates a new parsing payload with given value
// Initializes empty issues slice and empty path for validation
func NewParsePayload(value any) *ParsePayload {
	return &ParsePayload{
		Value:  value,
		Issues: make([]ZodRawIssue, 0), // Pre-allocate empty slice
		Path:   make([]any, 0),         // Initialize empty path
	}
}

// NewParsePayloadWithPath creates a new parsing payload with value and path
// Used when validating nested structures where path context is important
func NewParsePayloadWithPath(value any, path []any) *ParsePayload {
	// Copy path to avoid mutation issues
	pathCopy := make([]any, len(path))
	copy(pathCopy, path)

	return &ParsePayload{
		Value:  value,
		Issues: make([]ZodRawIssue, 0),
		Path:   pathCopy,
	}
}

// =============================================================================
// PAYLOAD MANIPULATION METHODS
// =============================================================================

// AddIssue adds a validation issue to the payload
// The issue's path will be updated with the current payload path
func (p *ParsePayload) AddIssue(issue ZodRawIssue) {
	// Update issue path with current payload path
	if len(p.Path) > 0 {
		// Combine payload path with issue path
		combinedPath := make([]any, len(p.Path)+len(issue.Path))
		copy(combinedPath, p.Path)
		copy(combinedPath[len(p.Path):], issue.Path)
		issue.Path = combinedPath
	}

	p.Issues = append(p.Issues, issue)
}

// AddIssueWithPath adds a validation issue with a specific path override
// Use this when the issue path should be different from the current payload path
func (p *ParsePayload) AddIssueWithPath(issue ZodRawIssue, path []any) {
	// Set the specific path for this issue
	pathCopy := make([]any, len(path))
	copy(pathCopy, path)
	issue.Path = pathCopy

	p.Issues = append(p.Issues, issue)
}

// HasIssues checks if the payload has any validation issues
// Returns true if validation has failed and issues are present
func (p *ParsePayload) HasIssues() bool {
	return len(p.Issues) > 0
}

// GetIssueCount returns the number of validation issues
// Useful for checking severity of validation failures
func (p *ParsePayload) GetIssueCount() int {
	return len(p.Issues)
}

// ClearIssues removes all validation issues from the payload
// Use with caution - typically only for recovery scenarios
func (p *ParsePayload) ClearIssues() {
	p.Issues = p.Issues[:0] // Reuse underlying slice capacity
}

// Clone creates a deep copy of the payload
// Useful when branching validation logic or preserving state
func (p *ParsePayload) Clone() *ParsePayload {
	// Copy issues slice
	issuesCopy := make([]ZodRawIssue, len(p.Issues))
	copy(issuesCopy, p.Issues)

	// Copy path slice
	pathCopy := make([]any, len(p.Path))
	copy(pathCopy, p.Path)

	return &ParsePayload{
		Value:  p.Value,
		Issues: issuesCopy,
		Path:   pathCopy,
	}
}

// =============================================================================
// PATH MANIPULATION METHODS
// =============================================================================

// PushPath adds a path element to the current validation path
// Used when entering nested validation contexts (object fields, array elements)
func (p *ParsePayload) PushPath(element any) {
	p.Path = append(p.Path, element)
}

// PopPath removes the last path element from the current validation path
// Used when exiting nested validation contexts
func (p *ParsePayload) PopPath() any {
	if len(p.Path) == 0 {
		return nil
	}

	last := p.Path[len(p.Path)-1]
	p.Path = p.Path[:len(p.Path)-1]
	return last
}

// SetPath sets the validation path to a specific value
// Use when jumping to a different validation context
func (p *ParsePayload) SetPath(path []any) {
	// Create copy to avoid external mutation
	pathCopy := make([]any, len(path))
	copy(pathCopy, path)
	p.Path = pathCopy
}

// GetPath returns a copy of the current validation path
// Safe to modify without affecting the original payload
func (p *ParsePayload) GetPath() []any {
	pathCopy := make([]any, len(p.Path))
	copy(pathCopy, p.Path)
	return pathCopy
}

// GetPathString returns a string representation of the current path
// Useful for debugging and error messages
func (p *ParsePayload) GetPathString() string {
	if len(p.Path) == 0 {
		return "root"
	}

	var result string
	for i, element := range p.Path {
		if i == 0 {
			result += "root"
		}

		switch elem := element.(type) {
		case string:
			result += "." + elem
		case int:
			result += "[" + string(rune(elem)) + "]"
		default:
			result += "." + string(rune(i))
		}
	}

	return result
}

// =============================================================================
// VALIDATION HELPER METHODS
// =============================================================================

// WithValue creates a new payload with a different value but same context
// Preserves path and issues while updating the value being validated
func (p *ParsePayload) WithValue(value any) *ParsePayload {
	return &ParsePayload{
		Value:  value,
		Issues: p.Issues, // Share the same issues slice
		Path:   p.Path,   // Share the same path slice
	}
}

// WithCleanIssues creates a new payload with same value and path but no issues
// Used when starting fresh validation on a value with preserved context
func (p *ParsePayload) WithCleanIssues() *ParsePayload {
	// Copy path to avoid mutation
	pathCopy := make([]any, len(p.Path))
	copy(pathCopy, p.Path)

	return &ParsePayload{
		Value:  p.Value,
		Issues: make([]ZodRawIssue, 0),
		Path:   pathCopy,
	}
}

// =============================================================================
// ADVANCED CONTEXT METHODS
// =============================================================================

// WithCustomError creates a new context with a custom error mapping function
// Allows per-validation customization of error messages
func (ctx *ParseContext) WithCustomError(errorMap ZodErrorMap) *ParseContext {
	return &ParseContext{
		Error:       errorMap,
		ReportInput: ctx.ReportInput,
	}
}

// WithReportInput creates a new context with input reporting enabled/disabled
// Controls whether validation errors include the original input value
func (ctx *ParseContext) WithReportInput(report bool) *ParseContext {
	return &ParseContext{
		Error:       ctx.Error,
		ReportInput: report,
	}
}

// Clone creates a copy of the parse context
// Useful for modifying context for specific validation branches
func (ctx *ParseContext) Clone() *ParseContext {
	return &ParseContext{
		Error:       ctx.Error,
		ReportInput: ctx.ReportInput,
	}
}
