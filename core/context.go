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
		value:  value,
		issues: make([]ZodRawIssue, 0), // Pre-allocate empty slice
		path:   make([]any, 0),         // Initialize empty path
	}
}

// NewParsePayloadWithPath creates a new parsing payload with value and path
// Used when validating nested structures where path context is important
func NewParsePayloadWithPath(value any, path []any) *ParsePayload {
	// Copy path to avoid mutation issues
	pathCopy := make([]any, len(path))
	copy(pathCopy, path)

	return &ParsePayload{
		value:  value,
		issues: make([]ZodRawIssue, 0),
		path:   pathCopy,
	}
}

// =============================================================================
// PAYLOAD MANIPULATION METHODS
// =============================================================================

// AddIssue adds a validation issue to the payload
// The issue's path will be updated with the current payload path
func (p *ParsePayload) AddIssue(issue ZodRawIssue) {
	// Update issue path with current payload path
	if len(p.path) > 0 {
		// Combine payload path with issue path
		combinedPath := make([]any, len(p.path)+len(issue.Path))
		copy(combinedPath, p.path)
		copy(combinedPath[len(p.path):], issue.Path)
		issue.Path = combinedPath
	}

	p.issues = append(p.issues, issue)
}

// AddIssueWithPath adds a validation issue with a specific path override
// Use this when the issue path should be different from the current payload path
func (p *ParsePayload) AddIssueWithPath(issue ZodRawIssue, path []any) {
	// Set the specific path for this issue
	pathCopy := make([]any, len(path))
	copy(pathCopy, path)
	issue.Path = pathCopy

	p.issues = append(p.issues, issue)
}

// AddIssueWithMessage adds a custom issue with the provided message using IssueCode Custom.
func (p *ParsePayload) AddIssueWithMessage(message string) {
	p.AddIssue(ZodRawIssue{
		Code:    Custom,
		Message: message,
	})
}

// AddIssueWithCode adds an issue with the specified IssueCode and message.
func (p *ParsePayload) AddIssueWithCode(code IssueCode, message string) {
	p.AddIssue(ZodRawIssue{
		Code:    code,
		Message: message,
	})
}

// AddIssues adds multiple issues to the payload.
func (p *ParsePayload) AddIssues(issues ...ZodRawIssue) {
	if p.issues == nil {
		p.issues = make([]ZodRawIssue, 0)
	}
	p.issues = append(p.issues, issues...)
}

// HasIssues checks if the payload has any validation issues
// Returns true if validation has failed and issues are present
func (p *ParsePayload) HasIssues() bool {
	return len(p.issues) > 0
}

// GetIssueCount returns the number of validation issues
// Useful for checking severity of validation failures
func (p *ParsePayload) GetIssueCount() int {
	return len(p.issues)
}

// Clone creates a deep copy of the payload
// Useful when branching validation logic or preserving state
func (p *ParsePayload) Clone() *ParsePayload {
	// Copy issues slice
	issuesCopy := make([]ZodRawIssue, len(p.issues))
	copy(issuesCopy, p.issues)

	// Copy path slice
	pathCopy := make([]any, len(p.path))
	copy(pathCopy, p.path)

	return &ParsePayload{
		value:  p.value,
		issues: issuesCopy,
		path:   pathCopy,
	}
}

// =============================================================================
// PATH MANIPULATION METHODS
// =============================================================================

// PushPath adds a path element to the current validation path
// Used when entering nested validation contexts (object fields, array elements)
func (p *ParsePayload) PushPath(element any) {
	p.path = append(p.path, element)
}

// GetPath returns a copy of the current validation path
// Safe to modify without affecting the original payload
func (p *ParsePayload) GetPath() []any {
	pathCopy := make([]any, len(p.path))
	copy(pathCopy, p.path)
	return pathCopy
}

// GetPathString returns a string representation of the current path
// Useful for debugging and error messages
func (p *ParsePayload) GetPathString() string {
	if len(p.path) == 0 {
		return "root"
	}

	var result string
	for i, element := range p.path {
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

// WithCleanIssues creates a new payload with same value and path but no issues
// Used when starting fresh validation on a value with preserved context
func (p *ParsePayload) WithCleanIssues() *ParsePayload {
	// Copy path to avoid mutation
	pathCopy := make([]any, len(p.path))
	copy(pathCopy, p.path)

	return &ParsePayload{
		value:  p.value,
		issues: make([]ZodRawIssue, 0),
		path:   pathCopy,
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

// SetValue assigns a new value to the payload.
func (p *ParsePayload) SetValue(v any) {
	p.value = v
}

// SetIssues replaces the issues slice with a new one.
// Use when bulk merging or overwriting issues.
func (p *ParsePayload) SetIssues(issues []ZodRawIssue) {
	p.issues = issues
}

// GetValue returns the current value being validated
func (p *ParsePayload) GetValue() any {
	return p.value
}

// GetIssues returns a copy of the current validation issues
// Safe to modify without affecting the original payload
func (p *ParsePayload) GetIssues() []ZodRawIssue {
	issuesCopy := make([]ZodRawIssue, len(p.issues))
	copy(issuesCopy, p.issues)
	return issuesCopy
}

// ClonePath returns a copy of the current validation path
// Alias for GetPath() for better semantic clarity
func (p *ParsePayload) ClonePath() []any {
	return p.GetPath()
}
