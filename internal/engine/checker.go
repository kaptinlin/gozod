package engine

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/pkg/slicex"
)

// =============================================================================
// CHECK EXECUTION ENGINE
// =============================================================================

// runChecksOnValue executes all validation checks on a given value
// This is the central validation execution engine that applies all checks
// Enhanced version using slicex for safer slice operations
func RunChecksOnValue(value any, checkList []core.ZodCheck, payload *core.ParsePayload, ctx *core.ParseContext) {
	// Use slicex to safely check for empty
	if slicex.IsEmpty(checkList) {
		return
	}

	// Pre-allocate Issues slice capacity based on number of checks
	// This reduces memory allocations during validation
	if cap(payload.Issues) < len(checkList) {
		// Allocate with some extra capacity (2x) to handle multiple issues per check
		newCapacity := len(checkList) * 2
		if newCapacity < 4 {
			newCapacity = 4 // Minimum reasonable capacity
		}
		payload.Issues = make([]core.ZodRawIssue, len(payload.Issues), newCapacity)
	}

	// Use slicex.Filter to preprocess valid checks
	validChecks, _ := slicex.Filter(checkList, func(item any) bool {
		if check, ok := item.(core.ZodCheck); ok {
			return check != nil && check.GetZod() != nil
		}
		return false
	})

	// Convert to typed slice for safer iteration
	if checks, err := slicex.ToTyped[core.ZodCheck](validChecks); err == nil {
		for _, check := range checks {
			// Get check internals and validate they exist
			if checkInternals := check.GetZod(); checkInternals != nil && checkInternals.Check != nil {
				// Create independent payload for each check to avoid interference
				checkPayload := &core.ParsePayload{
					Value:  value,
					Path:   payload.Path, // Preserve path context
					Issues: make([]core.ZodRawIssue, 0),
				}

				// Execute the check function
				checkInternals.Check(checkPayload)

				// If the check has custom error mapping, apply it to all produced issues
				if checkInternals.Def != nil && checkInternals.Def.Error != nil {
					for i := range checkPayload.Issues {
						checkPayload.Issues[i].Message = (*checkInternals.Def.Error)(checkPayload.Issues[i])
						checkPayload.Issues[i].Inst = checkInternals // attach for downstream resolution
					}
				}

				// Merge any issues found into the main payload
				if !slicex.IsEmpty(checkPayload.Issues) {
					// Use slicex.Merge to safely combine issue slices
					if mergedIssues, err := slicex.Merge(payload.Issues, checkPayload.Issues); err == nil {
						if typedIssues, err := slicex.ToTyped[core.ZodRawIssue](mergedIssues); err == nil {
							payload.Issues = typedIssues
						}
					}
				}

				// Support for early exit on Abort flag
				if !slicex.IsEmpty(checkPayload.Issues) && checkInternals.Def.Abort {
					break
				}
			}
		}
	}
}

// RunChecks executes all checks on a payload synchronously
// This is the core validation engine that processes all validation checks
func RunChecks(payload *core.ParsePayload, checks []core.ZodCheck, ctx *core.ParseContext) *core.ParsePayload {
	RunChecksOnValue(payload.Value, checks, payload, ctx)
	return payload
}

// CheckAborted checks if parsing should be aborted
// Examines issues starting from a specific index to determine if validation should stop
func CheckAborted(x core.ParsePayload, startIndex int) bool {
	// Use slicex to safely handle slice bounds
	if slicex.IsEmpty(x.Issues) || startIndex >= len(x.Issues) {
		return false
	}

	// Check issues from startIndex onwards
	for i := startIndex; i < len(x.Issues); i++ {
		if !x.Issues[i].Continue {
			return true
		}
	}
	return false
}
