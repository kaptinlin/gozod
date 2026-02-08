package core

import (
	"slices"
	"strconv"
	"strings"
)

// =============================================================================
// CONTEXT CREATION FUNCTIONS
// =============================================================================

// NewParseContext creates a new parse context with default values.
func NewParseContext() *ParseContext {
	return &ParseContext{}
}

// NewParsePayload creates a new parsing payload with given value
func NewParsePayload(value any) *ParsePayload {
	return &ParsePayload{
		value:  value,
		issues: make([]ZodRawIssue, 0, 2),
		path:   make([]any, 0, 4),
	}
}

// NewParsePayloadWithPath creates a new parsing payload with value and path
func NewParsePayloadWithPath(value any, path []any) *ParsePayload {
	pathCopy := slices.Clone(path)

	return &ParsePayload{
		value:  value,
		issues: make([]ZodRawIssue, 0, 2),
		path:   pathCopy,
	}
}

// =============================================================================
// PAYLOAD MANIPULATION METHODS
// =============================================================================

// AddIssue adds a validation issue to the payload
func (p *ParsePayload) AddIssue(issue ZodRawIssue) {
	if len(p.path) > 0 {
		issue.Path = slices.Concat(p.path, issue.Path)
	}

	p.issues = append(p.issues, issue)
}

// AddIssueWithPath adds a validation issue with a specific path override
func (p *ParsePayload) AddIssueWithPath(issue ZodRawIssue, path []any) {
	issue.Path = slices.Clone(path)

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

// AddIssues adds multiple issues to the payload
func (p *ParsePayload) AddIssues(issues ...ZodRawIssue) {
	if len(issues) == 0 {
		return
	}

	if p.issues == nil {
		p.issues = make([]ZodRawIssue, 0, len(issues)+2)
	}

	p.issues = slices.Concat(p.issues, issues)
}

// HasIssues checks if the payload has any validation issues
func (p *ParsePayload) HasIssues() bool {
	return len(p.issues) > 0
}

// GetIssueCount returns the number of validation issues
func (p *ParsePayload) GetIssueCount() int {
	return len(p.issues)
}

// Clone creates a deep copy of the payload
func (p *ParsePayload) Clone() *ParsePayload {
	return &ParsePayload{
		value:  p.value,
		issues: slices.Clone(p.issues),
		path:   slices.Clone(p.path),
	}
}

// =============================================================================
// PATH MANIPULATION METHODS
// =============================================================================

// PushPath adds a path element to the current validation path
func (p *ParsePayload) PushPath(element any) {
	p.path = append(p.path, element)
}

// GetPath returns a copy of the current validation path
func (p *ParsePayload) GetPath() []any {
	return slices.Clone(p.path)
}

func estimateIntLen(n int) int {
	if n == 0 {
		return 1
	}
	count := 0
	if n < 0 {
		count = 1
		n = -n
	}
	for n > 0 {
		count++
		n /= 10
	}
	return count
}

// GetPathString returns a string representation of the current path
func (p *ParsePayload) GetPathString() string {
	if len(p.path) == 0 {
		return "root"
	}

	capacity := 4
	for _, element := range p.path {
		switch elem := element.(type) {
		case string:
			capacity += 1 + len(elem)
		case int:
			capacity += 2 + estimateIntLen(elem)
		default:
			capacity += 10
		}
	}

	var b strings.Builder
	b.Grow(capacity)

	for i, element := range p.path {
		if i == 0 {
			b.WriteString("root")
		}

		switch elem := element.(type) {
		case string:
			b.WriteByte('.')
			b.WriteString(elem)
		case int:
			b.WriteByte('[')
			b.WriteString(strconv.Itoa(elem))
			b.WriteByte(']')
		default:
			b.WriteByte('.')
			b.WriteString(strconv.Itoa(i))
		}
	}

	return b.String()
}

// =============================================================================
// VALIDATION HELPER METHODS
// =============================================================================

// WithCleanIssues creates a new payload with same value and path but no issues
func (p *ParsePayload) WithCleanIssues() *ParsePayload {
	return &ParsePayload{
		value:  p.value,
		issues: make([]ZodRawIssue, 0, 2),
		path:   slices.Clone(p.path),
	}
}

// =============================================================================
// ADVANCED CONTEXT METHODS
// =============================================================================

// WithCustomError creates a new context with a custom error mapping function
func (ctx *ParseContext) WithCustomError(errorMap ZodErrorMap) *ParseContext {
	return &ParseContext{
		Error:       errorMap,
		ReportInput: ctx.ReportInput,
	}
}

// WithReportInput creates a new context with input reporting enabled/disabled
func (ctx *ParseContext) WithReportInput(report bool) *ParseContext {
	return &ParseContext{
		Error:       ctx.Error,
		ReportInput: report,
	}
}

// Clone creates a copy of the parse context
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

// SetIssues replaces the issues slice with a new one
func (p *ParsePayload) SetIssues(issues []ZodRawIssue) {
	p.issues = issues
}

// GetValue returns the current value being validated
func (p *ParsePayload) GetValue() any {
	return p.value
}

// GetIssues returns a copy of the current validation issues
func (p *ParsePayload) GetIssues() []ZodRawIssue {
	return slices.Clone(p.issues)
}
