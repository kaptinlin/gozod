package core

import (
	"errors"
	"slices"
)

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
	issues   []ZodIssue
}

// ParsePayload contains the value and validation issues during parsing.
type ParsePayload struct {
	value  any
	issues []ZodRawIssue
	path   []any
	ctx    *ParseContext
}

// NewRefinementContext creates a refinement context for the given value and parse context.
func NewRefinementContext(ctx *ParseContext, value any) *RefinementContext {
	rc := &RefinementContext{
		ParseContext: ctx,
		Value:        value,
		issues:       make([]ZodIssue, 0, 1),
	}
	rc.AddIssue = func(issue ZodIssue) {
		rc.issues = append(rc.issues, issue)
	}
	return rc
}

// Issues returns a copy of the issues added via AddIssue.
func (ctx *RefinementContext) Issues() []ZodIssue {
	return slices.Clone(ctx.issues)
}

// Err returns a combined error for the issues added via AddIssue.
func (ctx *RefinementContext) Err() error {
	if len(ctx.issues) == 0 {
		return nil
	}

	errs := make([]error, 0, len(ctx.issues))
	for _, issue := range ctx.issues {
		errs = append(errs, issue)
	}
	return errors.Join(errs...)
}

func getOrCreateContext(ctx ...*ParseContext) *ParseContext {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return &ParseContext{}
}
