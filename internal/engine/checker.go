package engine

import (
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
func executeChecks(value any, checks []core.ZodCheck, payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
	checksLen := len(checks)
	if checksLen == 0 {
		return payload
	}

	// Prepare issues slice with appropriate capacity
	currentIssues := payload.GetIssues()
	currentLen := len(currentIssues)

	// Expand capacity if needed
	if cap(currentIssues) < currentLen+checksLen {
		// Calculate new capacity based on current needs
		newCapacity := currentLen + checksLen
		if newCapacity < 8 {
			newCapacity = 8 // Reasonable minimum
		} else if newCapacity > 64 {
			newCapacity = currentLen + checksLen/2 + 8 // Moderate growth for large slices
		}

		newIssues := make([]core.ZodRawIssue, currentLen, newCapacity)
		copy(newIssues, currentIssues)
		payload.SetIssues(newIssues)
	}

	// Store payload path to avoid repeated calls
	payloadPath := payload.GetPath()

	// Execute each check sequentially
	for i := 0; i < checksLen; i++ {
		check := checks[i]
		if check == nil {
			continue
		}

		checkInternals := check.GetZod()
		if checkInternals == nil || checkInternals.Check == nil {
			continue
		}

		// Create independent payload for each check
		checkPayload := core.NewParsePayloadWithPath(value, payloadPath)

		// Execute the check function
		checkInternals.Check(checkPayload)

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
			checkPayload.SetIssues(checkIssues)
		}

		// Merge issues into main payload
		mainIssues := payload.GetIssues()
		mainIssuesLen := len(mainIssues)

		// Append issues to main payload
		if cap(mainIssues) >= mainIssuesLen+checkIssuesLen {
			// Sufficient capacity, extend existing slice
			newIssues := mainIssues[:mainIssuesLen+checkIssuesLen]
			copy(newIssues[mainIssuesLen:], checkIssues)
			payload.SetIssues(newIssues)
		} else {
			// Need to allocate new slice
			mergedIssues := make([]core.ZodRawIssue, mainIssuesLen+checkIssuesLen)
			copy(mergedIssues, mainIssues)
			copy(mergedIssues[mainIssuesLen:], checkIssues)
			payload.SetIssues(mergedIssues)
		}

		// Stop execution if abort flag is set
		if checkInternals.Def.Abort {
			break
		}
	}

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
