package engine

import (
	"slices"

	"github.com/kaptinlin/gozod/core"
)

// RunChecks executes all validation checks on a payload's value.
func RunChecks(checks []core.ZodCheck, payload *core.ParsePayload, ctx ...*core.ParseContext) *core.ParsePayload {
	if payload == nil || len(checks) == 0 {
		return payload
	}

	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return executeChecks(payload.GetValue(), checks, payload, parseCtx)
}

// RunChecksOnValue executes all validation checks on a specific value.
func RunChecksOnValue(value any, checks []core.ZodCheck, payload *core.ParsePayload, ctx ...*core.ParseContext) *core.ParsePayload {
	if payload == nil || len(checks) == 0 {
		return payload
	}

	var parseCtx *core.ParseContext
	if len(ctx) > 0 {
		parseCtx = ctx[0]
	}

	return executeChecks(value, checks, payload, parseCtx)
}

// executeChecks runs all checks sequentially, collecting issues and applying overwrites.
func executeChecks(value any, checks []core.ZodCheck, payload *core.ParsePayload, _ *core.ParseContext) *core.ParsePayload {
	checksLen := len(checks)
	if checksLen == 0 {
		return payload
	}

	// Pre-allocate capacity for issues if needed
	currentIssues := payload.GetIssues()
	if cap(currentIssues) < len(currentIssues)+checksLen {
		payload.SetIssues(slices.Grow(currentIssues, checksLen))
	}

	payloadPath := payload.Path()
	currentValue := value

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
			whenPayload := core.NewParsePayloadWithPath(currentValue, payload.Path())
			if !checkInternals.When(whenPayload) {
				continue
			}
		}

		checkPayload := core.NewParsePayloadWithPath(currentValue, payloadPath)
		checkInternals.Check(checkPayload)
		currentValue = checkPayload.GetValue()

		checkIssues := checkPayload.GetIssues()
		if len(checkIssues) == 0 {
			continue
		}

		// Apply custom error mapping if configured
		if checkInternals.Def != nil && checkInternals.Def.Error != nil {
			errorFn := *checkInternals.Def.Error
			for j := range len(checkIssues) {
				checkIssues[j].Message = errorFn(checkIssues[j])
				checkIssues[j].Inst = checkInternals
			}
		}

		payload.AddIssues(checkIssues...)

		if checkInternals.Def.Abort {
			break
		}
	}

	payload.SetValue(currentValue)
	return payload
}

// CheckAborted reports whether any issue from startIndex onwards signals an abort.
func CheckAborted(x core.ParsePayload, startIndex int) bool {
	issues := x.GetIssues()
	if len(issues) == 0 || startIndex >= len(issues) {
		return false
	}

	for i := startIndex; i < len(issues); i++ {
		if !issues[i].Continue {
			return true
		}
	}
	return false
}
