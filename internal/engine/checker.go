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

	var pc *core.ParseContext
	if len(ctx) > 0 {
		pc = ctx[0]
	}

	return executeChecks(payload.Value(), checks, payload, pc)
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

	var pc *core.ParseContext
	if len(ctx) > 0 {
		pc = ctx[0]
	}

	return executeChecks(value, checks, payload, pc)
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
		if ci.When != nil {
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
			for j := range len(iss) {
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
	if len(iss) == 0 || start >= len(iss) {
		return false
	}
	for i := start; i < len(iss); i++ {
		if !iss[i].Continue {
			return true
		}
	}
	return false
}
