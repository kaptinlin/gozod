package engine

import (
	"slices"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// CHECK EXECUTION ENGINE
// =============================================================================

// RunChecks executes all validation checks on a payload's value
// This is the primary method for running checks within the parsing pipeline
// Returns the payload for method chaining and consistent API
func RunChecks(checks []core.ZodCheck, payload *core.ParsePayload, ctx ...*core.ParseContext) *core.ParsePayload {
	if payload == nil || slicex.IsEmpty(checks) {
		return payload
	}

	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return executeChecks(payload.GetValue(), checks, payload, parseCtx)
}

// RunChecksOnValue executes all validation checks on a specific value
// Returns the payload for consistent API and error checking
func RunChecksOnValue(value any, checks []core.ZodCheck, payload *core.ParsePayload, ctx ...*core.ParseContext) *core.ParsePayload {
	if payload == nil || slicex.IsEmpty(checks) {
		return payload
	}

	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return executeChecks(value, checks, payload, parseCtx)
}

// executeChecks is the core implementation that runs all checks
// Manages memory allocation and executes validation logic
// Uses Go 1.22+ optimizations for slice operations
func executeChecks(value any, checks []core.ZodCheck, payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
	checksLen := len(checks)
	if checksLen == 0 {
		return payload
	}

	// Prepare issues slice with precise pre-allocation
	currentIssues := payload.GetIssues()
	currentLen := len(currentIssues)

	// Pre-allocate capacity based on common patterns:
	// Most checks produce 0-1 issues, with occasional bursts
	if cap(currentIssues) < currentLen+checksLen {
		// Use Go 1.22+ slices.Grow for optimal capacity management
		newIssues := slices.Grow(currentIssues, checksLen)
		payload.SetIssues(newIssues)
	}

	// Store payload path to avoid repeated calls
	payloadPath := payload.GetPath()

	// Track the current value as it may be modified by overwrite checks
	currentValue := value

	// Execute each check sequentially (Go 1.22+ range over int)
	for i := range checksLen {
		check := checks[i]
		if check == nil {
			continue
		}

		checkInternals := check.GetZod()
		if checkInternals == nil || checkInternals.Check == nil {
			continue
		}

		// Evaluate `when` predicate: skip this check if it returns false.
		if checkInternals.When != nil {
			// Create independent payload for when check with current value
			whenPayload := core.NewParsePayloadWithPath(currentValue, payload.GetPath())
			if !checkInternals.When(whenPayload) {
				continue
			}
		}

		// Create independent payload for each check with current value
		checkPayload := core.NewParsePayloadWithPath(currentValue, payloadPath)

		// Execute the check function
		checkInternals.Check(checkPayload)

		// Update current value in case it was modified by the check (e.g., overwrite)
		currentValue = checkPayload.GetValue()

		// Process check results
		checkIssues := checkPayload.GetIssues()
		checkIssuesLen := len(checkIssues)

		if checkIssuesLen == 0 {
			continue // No issues found, continue to next check
		}

		// Apply custom error mapping if configured
		if checkInternals.Def != nil && checkInternals.Def.Error != nil {
			errorFn := *checkInternals.Def.Error
			for j := 0; j < checkIssuesLen; j++ {
				checkIssues[j].Message = errorFn(checkIssues[j])
				checkIssues[j].Inst = checkInternals
			}
		}

		// Merge issues into main payload using efficient batch operation
		payload.AddIssues(checkIssues...)

		// Stop execution if abort flag is set
		if checkInternals.Def.Abort {
			break
		}
	}

	// Update the main payload value with the potentially modified value
	payload.SetValue(currentValue)

	return payload
}

// CheckAborted checks if parsing should be aborted
// Returns true if any issue from startIndex onwards has Continue set to false
func CheckAborted(x core.ParsePayload, startIndex int) bool {
	issues := x.GetIssues()
	issuesLen := len(issues)

	if issuesLen == 0 || startIndex >= issuesLen {
		return false
	}

	// Check issues from startIndex onwards for abort signals
	for i := startIndex; i < issuesLen; i++ {
		if !issues[i].Continue {
			return true
		}
	}
	return false
}
