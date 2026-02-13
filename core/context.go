package core

import (
	"slices"
)

// NewParseContext creates a new parse context with default values.
func NewParseContext() *ParseContext {
	return &ParseContext{}
}

// NewParsePayload creates a new parsing payload with given value.
func NewParsePayload(value any) *ParsePayload {
	return &ParsePayload{
		value:  value,
		issues: make([]ZodRawIssue, 0, 2),
		path:   make([]any, 0, 4),
	}
}

// NewParsePayloadWithPath creates a new parsing payload with value and path.
func NewParsePayloadWithPath(value any, path []any) *ParsePayload {
	return &ParsePayload{
		value:  value,
		issues: make([]ZodRawIssue, 0, 2),
		path:   slices.Clone(path),
	}
}

// AddIssue appends a validation issue, prepending the current path.
func (p *ParsePayload) AddIssue(issue ZodRawIssue) {
	if len(p.path) > 0 {
		issue.Path = slices.Concat(p.path, issue.Path)
	}
	p.issues = append(p.issues, issue)
}

// AddIssueWithPath adds a validation issue with a specific path override.
func (p *ParsePayload) AddIssueWithPath(issue ZodRawIssue, path []any) {
	issue.Path = slices.Clone(path)
	p.issues = append(p.issues, issue)
}

// AddIssueWithMessage adds a custom issue with the provided message.
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
	if len(issues) == 0 {
		return
	}
	p.issues = append(p.issues, issues...)
}

// HasIssues reports whether the payload has any validation issues.
func (p *ParsePayload) HasIssues() bool {
	return len(p.issues) > 0
}

// IssueCount returns the number of validation issues.
func (p *ParsePayload) IssueCount() int {
	return len(p.issues)
}

// Clone creates a deep copy of the payload.
func (p *ParsePayload) Clone() *ParsePayload {
	return &ParsePayload{
		value:  p.value,
		issues: slices.Clone(p.issues),
		path:   slices.Clone(p.path),
	}
}

// PushPath adds a path element to the current validation path.
func (p *ParsePayload) PushPath(element any) {
	p.path = append(p.path, element)
}

// Path returns a copy of the current validation path.
func (p *ParsePayload) Path() []any {
	return slices.Clone(p.path)
}

// WithCleanIssues creates a new payload with same value and path but no issues.
func (p *ParsePayload) WithCleanIssues() *ParsePayload {
	return &ParsePayload{
		value:  p.value,
		issues: make([]ZodRawIssue, 0, 2),
		path:   slices.Clone(p.path),
	}
}

// WithCustomError creates a new context with a custom error mapping function.
func (ctx *ParseContext) WithCustomError(errorMap ZodErrorMap) *ParseContext {
	return &ParseContext{
		Error:       errorMap,
		ReportInput: ctx.ReportInput,
	}
}

// WithReportInput creates a new context with input reporting enabled/disabled.
func (ctx *ParseContext) WithReportInput(report bool) *ParseContext {
	return &ParseContext{
		Error:       ctx.Error,
		ReportInput: report,
	}
}

// Clone creates a copy of the parse context.
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

// SetIssues replaces the issues slice.
func (p *ParsePayload) SetIssues(issues []ZodRawIssue) {
	p.issues = issues
}

// Value returns the current value being validated.
func (p *ParsePayload) Value() any {
	return p.value
}

// Issues returns a copy of the current validation issues.
func (p *ParsePayload) Issues() []ZodRawIssue {
	return slices.Clone(p.issues)
}
