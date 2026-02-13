package engine

import (
	"slices"

	"github.com/kaptinlin/gozod/core"
)

// ----------------------------------------------------------------------------
// Public check execution API
// ----------------------------------------------------------------------------

// RunChecks executes all validation checks on a payload's value.
func RunChecks(
	checks []core.ZodCheck,
	payload *core.ParsePayload,
	ctx ...*core.ParseContext,
) *core.ParsePayload {
	if payload == nil || len(checks) == 0 {
		return payload
	}
	return executeChecks(payload.Value(), checks, payload, firstContext(ctx))
}

// RunChecksOnValue executes all validation checks on a specific value.
func RunChecksOnValue(
	value any,
	checks []core.ZodCheck,
	payload *core.ParsePayload,
	ctx ...*core.ParseContext,
) *core.ParsePayload {
	if payload == nil || len(checks) == 0 {
		return payload
	}
	return executeChecks(value, checks, payload, firstContext(ctx))
}

// firstContext returns the first non-nil context from the variadic slice, or nil.
func firstContext(ctx []*core.ParseContext) *core.ParseContext {
	if len(ctx) > 0 {
		return ctx[0]
	}
	return nil
}

// ----------------------------------------------------------------------------
// Internal check execution
// ----------------------------------------------------------------------------

// executeChecks runs all checks sequentially, collecting issues and applying overwrites.
func executeChecks(
	value any,
	checks []core.ZodCheck,
	payload *core.ParsePayload,
	_ *core.ParseContext,
) *core.ParsePayload {
	n := len(checks)
	if n == 0 {
		return payload
	}

	cur := payload.Issues()
	if cap(cur) < len(cur)+n {
		payload.SetIssues(slices.Grow(cur, n))
	}

	path := payload.Path()
	val := value

	for i := range n {
		c := checks[i]
		if c == nil {
			continue
		}

		ci := c.Zod()
		if ci == nil || ci.Check == nil {
			continue
		}

		// Skip check if "when" predicate returns false.
		// Check for explicit abort before evaluating When — a prior check
		// with Abort:true should prevent subsequent checks from running,
		// even if they have a When predicate (Zod v4 fix: 5b574501).
		if ci.When != nil {
			if CheckAborted(*payload, 0) {
				continue
			}
			wp := core.NewParsePayloadWithPath(val, payload.Path())
			if !ci.When(wp) {
				continue
			}
		}

		cp := core.NewParsePayloadWithPath(val, path)
		ci.Check(cp)
		val = cp.Value()

		iss := cp.Issues()
		if len(iss) == 0 {
			continue
		}

		if ci.Def != nil && ci.Def.Error != nil {
			errFn := *ci.Def.Error
			for j := range iss {
				iss[j].Message = errFn(iss[j])
				iss[j].Inst = ci
			}
		}

		payload.AddIssues(iss...)

		if ci.Def.Abort {
			break
		}
	}

	payload.SetValue(val)
	return payload
}

// CheckAborted reports whether any issue from start onwards signals an abort.
func CheckAborted(x core.ParsePayload, start int) bool {
	iss := x.Issues()
	if start >= len(iss) {
		return false
	}
	for _, issue := range iss[start:] {
		if !issue.Continue {
			return true
		}
	}
	return false
}
